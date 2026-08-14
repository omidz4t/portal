package store

import (
	"path/filepath"
	"testing"
)

func TestPersonaBotAndGhostBind(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "persona.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	b, err := s.InsertPersonaBot(&PersonaBot{
		OwnerTelegramUserID: 100,
		OwnerDCAccountID:    1,
		OwnerDCChatID:       2,
		OwnerDCAddress:      "owner@example.org",
		OwnerVcard:          "BEGIN:VCARD\nFN:Owner\nEND:VCARD",
		BotToken:            "1:token-a",
		BotUserID:           9001,
		BotUsername:         "alice_bot",
		BotURL:              "https://t.me/alice_bot",
	})
	if err != nil {
		t.Fatal(err)
	}
	if b.ID == 0 || b.Status != PersonaBotActive {
		t.Fatalf("bot: %+v", b)
	}

	if _, err := s.InsertPersonaBot(&PersonaBot{
		OwnerTelegramUserID: 101,
		BotToken:            "1:token-b",
		BotUserID:           9001,
		BotUsername:         "alice_bot",
	}); err == nil {
		t.Fatal("expected unique bot_user_id error")
	}

	g1, err := s.InsertGhostAccount(&GhostAccount{
		PersonaBotID:     b.ID,
		TelegramUserID:   55,
		TelegramUsername: "bob",
		DisplayName:      "Bob",
		DCAccountID:      10,
		DCAddress:        "ghost55@example.org",
		OwnerChatID:      42,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateGhostAvatar(g1.ID, "file_abc"); err != nil {
		t.Fatal(err)
	}

	g2, err := s.InsertGhostAccount(&GhostAccount{
		PersonaBotID:   b.ID,
		TelegramUserID: 55,
		DCAccountID:    99,
		DCAddress:      "other@example.org",
	})
	if err != nil {
		t.Fatal(err)
	}
	if g2.ID != g1.ID || g2.DCAccountID != 10 {
		t.Fatalf("reuse failed: %+v vs %+v", g1, g2)
	}

	byDC, err := s.GetGhostByDCAccount(10)
	if err != nil || byDC == nil || byDC.TelegramUserID != 55 || byDC.OwnerChatID != 42 {
		t.Fatalf("by dc: %v %+v", err, byDC)
	}
	if byDC.AvatarFileID != "file_abc" {
		t.Fatalf("avatar: %+v", byDC)
	}

	gotBot, err := s.GetPersonaBot(b.ID)
	if err != nil || gotBot.OwnerVcard == "" {
		t.Fatalf("owner vcard: %v %+v", err, gotBot)
	}

	if _, err := s.InsertGhostAccount(&GhostAccount{
		PersonaBotID:   b.ID,
		TelegramUserID: 56,
		DCAccountID:    11,
		DCAddress:      "g56@example.org",
	}); err != nil {
		t.Fatal(err)
	}

	n, err := s.CountGhostAccounts()
	if err != nil || n != 2 {
		t.Fatalf("count ghosts: %v %d", err, n)
	}

	nOwn, err := s.CountPersonaBotsByOwner(100, true)
	if err != nil || nOwn != 1 {
		t.Fatalf("count owner active: %v %d", err, nOwn)
	}
	nBotGhosts, err := s.CountGhostAccountsByBot(b.ID)
	if err != nil || nBotGhosts != 2 {
		t.Fatalf("count ghosts by bot: %v %d", err, nBotGhosts)
	}

	nDis, err := s.DisablePersonaBotByOwner(100, b.ID, "")
	if err != nil || nDis != 1 {
		t.Fatalf("disable: %v %d", err, nDis)
	}

	// /unpair-bot must stick: restart only starts ListActivePersonaBots.
	active, err := s.ListActivePersonaBots()
	if err != nil {
		t.Fatal(err)
	}
	for _, bot := range active {
		if bot.ID == b.ID {
			t.Fatal("disabled bot must not appear in active list (restart would poll it)")
		}
	}
	got, err := s.GetPersonaBot(b.ID)
	if err != nil || got == nil || got.Status != PersonaBotDisabled {
		t.Fatalf("status after disable: %v %+v", err, got)
	}

	again, err := s.ReactivatePersonaBot(b.ID, &PersonaBot{
		OwnerTelegramUserID: 100,
		OwnerDCAccountID:    1,
		OwnerDCChatID:       2,
		OwnerDCAddress:      "owner@example.org",
		OwnerVcard:          "BEGIN:VCARD\nFN:Owner\nEND:VCARD",
		BotToken:            "1:token-a-rotated",
		BotUserID:           9001,
		BotUsername:         "alice_bot",
		BotURL:              "https://t.me/alice_bot",
	})
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != b.ID {
		t.Fatalf("reactivate must keep id: got %d want %d", again.ID, b.ID)
	}
	if again.Status != PersonaBotActive || again.BotToken != "1:token-a-rotated" {
		t.Fatalf("reactivate: %+v", again)
	}
	all, err := s.ListPersonaBotsByOwner(100)
	if err != nil || len(all) != 1 {
		t.Fatalf("must not insert a second row: %v n=%d", err, len(all))
	}
}

func TestPairOwnerVcard(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "pair-vcard.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p, err := s.CreatePendingPair(42, "alice", 99)
	if err != nil {
		t.Fatal(err)
	}
	act, err := s.ActivatePair(p.Code, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetPairOwnerVcard(act.ID, "BEGIN:VCARD\nEND:VCARD"); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetActiveByTelegram(42)
	if err != nil || got == nil || got.OwnerVcard == "" {
		t.Fatalf("vcard: %v %+v", err, got)
	}
}
