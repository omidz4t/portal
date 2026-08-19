package erasure

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/omidz4t/portal/internal/store"
	"go.uber.org/zap"
)

func TestPendingConsumeExpires(t *testing.T) {
	s := New(zap.NewNop().Sugar(), nil, nil, nil)
	s.Request("tg:1")
	if !s.Consume("tg:1") {
		t.Fatal("expected pending")
	}
	if s.Consume("tg:1") {
		t.Fatal("should be one-shot")
	}
	s.Request("tg:2")
	s.mu.Lock()
	s.pending["tg:2"] = time.Now().Add(-time.Second)
	s.mu.Unlock()
	if s.Consume("tg:2") {
		t.Fatal("expired should not consume")
	}
}

func TestKeys(t *testing.T) {
	if TelegramKey(42) != "tg:42" {
		t.Fatalf("tg key: %s", TelegramKey(42))
	}
	if DCKey(1, 9) != "dc:1:9" {
		t.Fatalf("dc key: %s", DCKey(1, 9))
	}
}

func TestConsumeNilService(t *testing.T) {
	var s *Service
	s.Request("x")
	if s.Consume("x") {
		t.Fatal("nil service must not consume")
	}
}

func TestApproveRequiresPriorRequest(t *testing.T) {
	s := New(zap.NewNop().Sugar(), nil, nil, nil)
	if s.Consume(TelegramKey(1)) {
		t.Fatal("approve without request")
	}
	s.Request(TelegramKey(1))
	s.Request(DCKey(2, 3))
	if !s.Consume(TelegramKey(1)) {
		t.Fatal("tg approve")
	}
	if s.Consume(TelegramKey(1)) {
		t.Fatal("tg approve reused")
	}
	if !s.Consume(DCKey(2, 3)) {
		t.Fatal("dc approve")
	}
}

func TestPurgeTelegramUserRemovesRows(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	p, err := st.CreatePendingPair(7, "u", 8)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.ActivatePair(p.Code, 1, 2); err != nil {
		t.Fatal(err)
	}
	bot, err := st.InsertPersonaBot(&store.PersonaBot{
		OwnerTelegramUserID: 7,
		BotToken:            "tok",
		BotUserID:           99,
		BotUsername:         "mybot",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.InsertGhostAccount(&store.GhostAccount{
		PersonaBotID:   bot.ID,
		TelegramUserID: 55,
		DCAccountID:    40,
	}); err != nil {
		t.Fatal(err)
	}
	keep, err := st.CreatePendingPair(8, "other", 9)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.ActivatePair(keep.Code, 1, 99); err != nil {
		t.Fatal(err)
	}

	svc := New(zap.NewNop().Sugar(), st, nil, nil)
	if err := svc.PurgeTelegramUser(7); err != nil {
		t.Fatal(err)
	}
	got, err := st.GetActiveByTelegram(7)
	if err != nil || got != nil {
		t.Fatalf("pair leftover: %v %+v", err, got)
	}
	bots, err := st.ListPersonaBotsByOwner(7)
	if err != nil || len(bots) != 0 {
		t.Fatalf("bots leftover: %v %d", err, len(bots))
	}
	if g, err := st.GetGhostByDCAccount(40); err != nil || g != nil {
		t.Fatalf("ghost leftover: %v %+v", err, g)
	}
	if other, err := st.GetActiveByTelegram(8); err != nil || other == nil {
		t.Fatalf("other user pair: %v %+v", err, other)
	}
}

func TestPurgeFromDCChatFollowsTelegramUser(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "dc.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	p, err := st.CreatePendingPair(7, "u", 8)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.ActivatePair(p.Code, 4, 5); err != nil {
		t.Fatal(err)
	}
	svc := New(zap.NewNop().Sugar(), st, nil, nil)
	if err := svc.PurgeFromDCChat(4, 5); err != nil {
		t.Fatal(err)
	}
	if got, err := st.GetActiveByTelegram(7); err != nil || got != nil {
		t.Fatalf("should resolve tg user from dc chat: %v %+v", err, got)
	}
}

func TestPurgeFromDCChatPendingOnly(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "dc-pending.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if _, err := st.CreatePendingFromDC(1, 2); err != nil {
		t.Fatal(err)
	}
	keep, err := st.CreatePendingFromDC(1, 3)
	if err != nil {
		t.Fatal(err)
	}
	svc := New(zap.NewNop().Sugar(), st, nil, nil)
	if err := svc.PurgeFromDCChat(1, 2); err != nil {
		t.Fatal(err)
	}
	if left, err := st.ListPairsByDCChat(1, 2); err != nil || len(left) != 0 {
		t.Fatalf("pending leftover: %v n=%d", err, len(left))
	}
	if still, err := st.ListPairsByDCChat(1, 3); err != nil || len(still) != 1 || still[0].ID != keep.ID {
		t.Fatalf("other pending: %v %+v", err, still)
	}
}
