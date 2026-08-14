package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPairingFlow(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p, err := s.CreatePendingPair(42, "alice", 99)
	if err != nil {
		t.Fatal(err)
	}
	if p.Code == "" || p.Status != StatusPending {
		t.Fatalf("bad pair: %+v", p)
	}

	got, err := s.GetPendingByCode(stringsToLower(p.Code))
	if err != nil || got == nil {
		t.Fatalf("pending: %v %+v", err, got)
	}

	act, err := s.ActivatePair(p.Code, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if act.Status != StatusActive || act.DCChatID != 10 {
		t.Fatalf("activate: %+v", act)
	}

	byTG, err := s.GetActiveByTelegram(42)
	if err != nil || byTG == nil {
		t.Fatal(err)
	}
	if byTG.DCAccountID != 1 {
		t.Fatalf("dc account: %v", byTG.DCAccountID)
	}
}

func TestDCInitiatedPairingFlow(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "test-dc.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p, err := s.CreatePendingFromDC(1, 55)
	if err != nil {
		t.Fatal(err)
	}
	if p.DCChatID != 55 || p.TelegramUserID != 0 {
		t.Fatalf("bad dc pending: %+v", p)
	}

	// Pasting DC-initiated code on DC should fail (must open Telegram link).
	if _, err := s.ActivatePair(p.Code, 1, 55); err == nil {
		t.Fatal("expected error activating DC-initiated code from DC side")
	}

	act, err := s.ActivatePairFromTelegram(p.Code, 77, "bob", 77)
	if err != nil {
		t.Fatal(err)
	}
	if act.Status != StatusActive || act.TelegramUserID != 77 || act.DCChatID != 55 {
		t.Fatalf("activate from tg: %+v", act)
	}

	byTG, err := s.GetActiveByTelegram(77)
	if err != nil || byTG == nil || byTG.DCChatID != 55 {
		t.Fatalf("by tg: %v %+v", err, byTG)
	}
}

func TestActivatePairRefusesAlreadyPairedDCChat(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "hijack.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	first, err := s.CreatePendingPair(1, "alice", 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ActivatePair(first.Code, 1, 10); err != nil {
		t.Fatal(err)
	}

	second, err := s.CreatePendingPair(2, "mallory", 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ActivatePair(second.Code, 1, 10); err == nil {
		t.Fatal("expected reject when DC chat is already paired")
	}

	got, err := s.GetActiveByDCChat(1, 10)
	if err != nil || got == nil || got.TelegramUserID != 1 {
		t.Fatalf("original pair must remain: %v %+v", err, got)
	}
	pend, err := s.GetPendingByCode(second.Code)
	if err != nil || pend == nil || pend.Status != StatusPending {
		t.Fatalf("mallory pending must stay pending: %v %+v", err, pend)
	}
}

func TestCodeHMACUniqueWhenEncrypted(t *testing.T) {
	key := testKey(t)
	s, err := OpenOpts(filepath.Join(t.TempDir(), "hmac-uid.db"), Options{Key: key})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	a, err := s.CreatePendingPair(1, "a", 1)
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreatePendingPair(2, "b", 2)
	if err != nil {
		t.Fatal(err)
	}
	if a.Code == b.Code {
		t.Fatal("codes must differ")
	}
	ha := s.crypt.HMACCode(a.Code)
	hb := s.crypt.HMACCode(b.Code)
	if ha == "" || ha == hb {
		t.Fatalf("hmac a=%s b=%s", ha, hb)
	}

	_, err = s.db.Exec(
		`INSERT INTO pairs (code, code_hmac, telegram_user_id, telegram_username, telegram_chat_id, status, created_at)
		 VALUES (?, ?, 3, 'c', 3, ?, 0)`,
		"enc:v1:collision-probe", ha, StatusPending,
	)
	if err == nil {
		t.Fatal("duplicate code_hmac must be rejected")
	}
	if !strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
		t.Fatalf("want UNIQUE, got %v", err)
	}

	got, err := s.GetPendingByCode(a.Code)
	if err != nil || got == nil || got.ID != a.ID {
		t.Fatalf("original lookup: %v %+v", err, got)
	}
}

func TestCodeHMACEmptyNotUniqueWithoutEncryption(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "hmac-empty.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.CreatePendingPair(1, "a", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreatePendingPair(2, "b", 2); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM pairs WHERE code_hmac = ''`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("plaintext store keeps empty hmac: %d", n)
	}
}

func TestActivatePairDisconnectsOtherTGActives(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "split.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	old, err := s.CreatePendingPair(7, "alice", 7)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ActivatePair(old.Code, 1, 10); err != nil {
		t.Fatal(err)
	}
	neu, err := s.CreatePendingPair(7, "alice", 7)
	if err != nil {
		t.Fatal(err)
	}
	act, err := s.ActivatePair(neu.Code, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if act.DCChatID != 20 {
		t.Fatalf("new pair: %+v", act)
	}
	if p, _ := s.GetActiveByDCChat(1, 10); p != nil {
		t.Fatalf("old DC pair must be disconnected: %+v", p)
	}
	if p, _ := s.GetActiveByTelegram(7); p == nil || p.DCChatID != 20 {
		t.Fatalf("telegram must point at new chat: %+v", p)
	}
}

func TestCreatePendingPairCodeLength(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "len.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p, err := s.CreatePendingPair(1, "u", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Code) != DefaultCodeLength {
		t.Fatalf("len=%d code=%s", len(p.Code), p.Code)
	}
	if !LooksLikeCode(p.Code) {
		t.Fatalf("LooksLikeCode %s", p.Code)
	}
}

func TestPendingTTLExpires(t *testing.T) {
	s, err := OpenOpts(filepath.Join(t.TempDir(), "ttl.db"), Options{PendingTTL: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p, err := s.CreatePendingPair(1, "u", 1)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	got, err := s.GetPendingByCode(p.Code)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected expired, got %+v", got)
	}
	if _, err := s.ActivatePair(p.Code, 1, 2); err == nil {
		t.Fatal("activate after TTL should fail")
	}
}

func TestEncryptRequiredNoKey(t *testing.T) {
	_, err := OpenOpts(filepath.Join(t.TempDir(), "req.db"), Options{EncryptRequired: true})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEncryptedPairingHMACLookup(t *testing.T) {
	key := testKey(t)
	path := filepath.Join(t.TempDir(), "enc.db")
	s, err := OpenOpts(path, Options{Key: key})
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.CreatePendingPair(9, "z", 9)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetPendingByCode(stringsToLower(p.Code))
	if err != nil || got == nil || got.Code != p.Code {
		t.Fatalf("hmac lookup: %v %+v", err, got)
	}
	if _, err := s.GetPendingByCode("ZZZZZZZZ"); err != nil {
		t.Fatal(err)
	}
	s.Close()

	// Wrong key cannot open existing ciphertext.
	bad := make([]byte, 32)
	if _, err := OpenOpts(path, Options{Key: bad}); err == nil {
		t.Fatal("wrong key should fail after encrypted rows exist")
	}
}

func TestMigratePlaintextSecrets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plain.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.CreatePendingPair(3, "m", 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertPersonaBot(&PersonaBot{
		OwnerTelegramUserID: 3,
		BotToken:            "123:plaintext-token",
		BotUserID:           1,
		OwnerVcard:          "BEGIN:VCARD\nEND:VCARD",
	}); err != nil {
		t.Fatal(err)
	}
	s.Close()

	s2, err := OpenOpts(path, Options{Key: testKey(t)})
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	got, err := s2.GetPendingByCode(p.Code)
	if err != nil || got == nil {
		t.Fatalf("after migrate: %v %+v", err, got)
	}
	bots, err := s2.ListPersonaBotsByOwner(3)
	if err != nil || len(bots) != 1 || bots[0].BotToken != "123:plaintext-token" {
		t.Fatalf("token: %v %+v", err, bots)
	}
}

func TestNewDBFileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mode.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("perm %o", st.Mode().Perm())
	}
}

func TestLooksLikeCode(t *testing.T) {
	if !LooksLikeCode("ABCD2345") || LooksLikeCode("") || LooksLikeCode("ab!") {
		t.Fatal("LooksLikeCode")
	}
}

func stringsToLower(s string) string {
	// exercise case-insensitive lookup via normalize in GetPendingByCode
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
