package store

import (
	"path/filepath"
	"testing"
)

func TestPurgeTelegramUserRemovesAllOwnedAndLeavesOthers(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "purge.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p, err := s.CreatePendingPair(7, "alice", 70)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ActivatePair(p.Code, 1, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DisconnectByTelegram(7); err != nil {
		t.Fatal(err)
	}
	p2, err := s.CreatePendingPair(7, "alice", 70)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ActivatePair(p2.Code, 1, 11); err != nil {
		t.Fatal(err)
	}

	other, err := s.CreatePendingPair(8, "bob", 80)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ActivatePair(other.Code, 1, 20); err != nil {
		t.Fatal(err)
	}

	owned, err := s.InsertPersonaBot(&PersonaBot{
		OwnerTelegramUserID: 7,
		BotToken:            "tok-7",
		BotUserID:           9007,
		BotUsername:         "alicebot",
	})
	if err != nil {
		t.Fatal(err)
	}
	otherBot, err := s.InsertPersonaBot(&PersonaBot{
		OwnerTelegramUserID: 8,
		BotToken:            "tok-8",
		BotUserID:           9008,
		BotUsername:         "bobbot",
	})
	if err != nil {
		t.Fatal(err)
	}

	gOwnedPeer, err := s.InsertGhostAccount(&GhostAccount{
		PersonaBotID:   owned.ID,
		TelegramUserID: 55,
		DCAccountID:    100,
	})
	if err != nil {
		t.Fatal(err)
	}
	gSelfOnOther, err := s.InsertGhostAccount(&GhostAccount{
		PersonaBotID:   otherBot.ID,
		TelegramUserID: 7,
		DCAccountID:    101,
	})
	if err != nil {
		t.Fatal(err)
	}
	gOther, err := s.InsertGhostAccount(&GhostAccount{
		PersonaBotID:   otherBot.ID,
		TelegramUserID: 9,
		DCAccountID:    102,
	})
	if err != nil {
		t.Fatal(err)
	}

	grpOwned, err := s.InsertGhostGroup(&GhostGroup{
		PersonaBotID:           owned.ID,
		TelegramChatID:         -100,
		Title:                  "owned group",
		CoordinatorDCAccountID: 100,
		DCChatID:               3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertGhostGroupMember(&GhostGroupMember{
		GhostGroupID:   grpOwned.ID,
		TelegramUserID: 55,
		GhostAccountID: gOwnedPeer.ID,
	}); err != nil {
		t.Fatal(err)
	}

	grpOther, err := s.InsertGhostGroup(&GhostGroup{
		PersonaBotID:           otherBot.ID,
		TelegramChatID:         -200,
		Title:                  "bob group",
		CoordinatorDCAccountID: 102,
		DCChatID:               4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertGhostGroupMember(&GhostGroupMember{
		GhostGroupID:   grpOther.ID,
		TelegramUserID: 7,
		GhostAccountID: gSelfOnOther.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertGhostGroupMember(&GhostGroupMember{
		GhostGroupID:   grpOther.ID,
		TelegramUserID: 9,
		GhostAccountID: gOther.ID,
	}); err != nil {
		t.Fatal(err)
	}

	ghosts, err := s.GhostAccountsForPurge(7)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[int64]bool{}
	for _, g := range ghosts {
		ids[g.ID] = true
	}
	if !ids[gOwnedPeer.ID] || !ids[gSelfOnOther.ID] {
		t.Fatalf("purge list missing ghosts: %+v", ghosts)
	}
	if ids[gOther.ID] {
		t.Fatal("must not list an unrelated ghost")
	}

	if err := s.PurgeTelegramUser(7); err != nil {
		t.Fatal(err)
	}

	pairs, err := s.ListPairsByTelegram(7)
	if err != nil || len(pairs) != 0 {
		t.Fatalf("pairs leftover: %v n=%d", err, len(pairs))
	}
	if bots, err := s.ListPersonaBotsByOwner(7); err != nil || len(bots) != 0 {
		t.Fatalf("bots leftover: %v n=%d", err, len(bots))
	}
	if g, err := s.GetGhostByDCAccount(100); err != nil || g != nil {
		t.Fatalf("owned ghost leftover: %v %+v", err, g)
	}
	if g, err := s.GetGhostByDCAccount(101); err != nil || g != nil {
		t.Fatalf("self ghost leftover: %v %+v", err, g)
	}
	if gg, err := s.GetGhostGroup(owned.ID, -100); err != nil || gg != nil {
		t.Fatalf("owned group leftover: %v %+v", err, gg)
	}
	mem, err := s.GetGhostGroupMember(grpOther.ID, 7)
	if err != nil || mem != nil {
		t.Fatalf("membership leftover: %v %+v", err, mem)
	}

	keep, err := s.GetActiveByTelegram(8)
	if err != nil || keep == nil {
		t.Fatalf("other pair must remain: %v %+v", err, keep)
	}
	if bots, err := s.ListPersonaBotsByOwner(8); err != nil || len(bots) != 1 {
		t.Fatalf("other bot: %v n=%d", err, len(bots))
	}
	if g, err := s.GetGhostByDCAccount(102); err != nil || g == nil {
		t.Fatalf("other ghost: %v %+v", err, g)
	}
	if gg, err := s.GetGhostGroup(otherBot.ID, -200); err != nil || gg == nil {
		t.Fatalf("other group: %v %+v", err, gg)
	}
	if mem, err := s.GetGhostGroupMember(grpOther.ID, 9); err != nil || mem == nil {
		t.Fatalf("other member: %v %+v", err, mem)
	}

	if err := s.Vacuum(); err != nil {
		t.Fatal(err)
	}
}

func TestPurgeDCChatPairs(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "purge-dc.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p, err := s.CreatePendingFromDC(3, 4)
	if err != nil {
		t.Fatal(err)
	}
	keep, err := s.CreatePendingFromDC(3, 5)
	if err != nil {
		t.Fatal(err)
	}
	_ = p
	if err := s.PurgeDCChatPairs(3, 4); err != nil {
		t.Fatal(err)
	}
	left, err := s.ListPairsByDCChat(3, 4)
	if err != nil || len(left) != 0 {
		t.Fatalf("purged chat leftover: %v n=%d", err, len(left))
	}
	still, err := s.ListPairsByDCChat(3, 5)
	if err != nil || len(still) != 1 || still[0].ID != keep.ID {
		t.Fatalf("other dc pair: %v %+v", err, still)
	}
}
