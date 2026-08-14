package store

import (
	"database/sql"
	"path/filepath"
	"testing"
)

// Classic and SQLite-specific injection strings. They must be stored or rejected
// as data, never executed.
var sqliPayloads = []string{
	`' OR '1'='1`,
	`' OR '1'='1' --`,
	`'; DROP TABLE pairs;--`,
	`'; DROP TABLE persona_bots;--`,
	`'; DROP TABLE ghost_accounts;--`,
	`"; DROP TABLE pairs;--`,
	`1; DROP TABLE pairs;--`,
	`admin'--`,
	`' UNION SELECT 1,2,3--`,
	`' UNION SELECT sql FROM sqlite_master--`,
	`'; ATTACH DATABASE ':memory:' AS pwn;--`,
	`%'; DELETE FROM pairs;--`,
	"` OR 1=1 --",
	`0 OR 1=1`,
	`*/; DROP TABLE pairs;/*`,
	"\x00'; DROP TABLE pairs;--",
	`' OR username IS NOT NULL OR '`,
}

func openSQLiStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "sqli.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func tableExists(t *testing.T, s *Store, name string) bool {
	t.Helper()
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n == 1
}

func assertCoreSchema(t *testing.T, s *Store) {
	t.Helper()
	for _, name := range []string{"pairs", "persona_bots", "ghost_accounts", "ghost_groups", "ghost_group_members"} {
		if !tableExists(t, s, name) {
			t.Fatalf("table %s missing (possible SQLi)", name)
		}
	}
}

func countRows(t *testing.T, s *Store, table string) int {
	t.Helper()
	// table is a test-only literal, never user input
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestSQLInjectionPairingFields(t *testing.T) {
	s := openSQLiStore(t)
	for _, payload := range sqliPayloads {
		t.Run("username/"+payload, func(t *testing.T) {
			before := countRows(t, s, "pairs")
			p, err := s.CreatePendingPair(1001, payload, 2002)
			if err != nil {
				t.Fatal(err)
			}
			got, err := s.GetPendingByCode(p.Code)
			if err != nil || got == nil {
				t.Fatalf("lookup: %v %+v", err, got)
			}
			if got.TelegramUsername != payload {
				t.Fatalf("username not stored literally:\n got %q\nwant %q", got.TelegramUsername, payload)
			}
			if countRows(t, s, "pairs") != before+1 {
				t.Fatal("unexpected pairs row count")
			}
			assertCoreSchema(t, s)
		})
	}
}

func TestSQLInjectionPairingCodeLookup(t *testing.T) {
	s := openSQLiStore(t)
	legit, err := s.CreatePendingPair(7, "ok", 7)
	if err != nil {
		t.Fatal(err)
	}
	before := countRows(t, s, "pairs")
	for _, payload := range sqliPayloads {
		got, err := s.GetPendingByCode(payload)
		if err != nil {
			t.Fatalf("GetPendingByCode(%q): %v", payload, err)
		}
		if got != nil {
			t.Fatalf("injection code must not match a pair: %+v", got)
		}
		if _, err := s.ActivatePair(payload, 1, 1); err == nil {
			t.Fatalf("ActivatePair accepted injection %q", payload)
		}
		if _, err := s.ActivatePairFromTelegram(payload, 99, payload, 99); err == nil {
			t.Fatalf("ActivatePairFromTelegram accepted injection %q", payload)
		}
	}
	if countRows(t, s, "pairs") != before {
		t.Fatal("lookups must not insert or delete rows")
	}
	still, err := s.GetPendingByCode(legit.Code)
	if err != nil || still == nil || still.ID != legit.ID {
		t.Fatalf("legitimate pair damaged: %v %+v", err, still)
	}
	assertCoreSchema(t, s)
}

func TestSQLInjectionOwnerVcard(t *testing.T) {
	s := openSQLiStore(t)
	p, err := s.CreatePendingPair(8, "u", 8)
	if err != nil {
		t.Fatal(err)
	}
	act, err := s.ActivatePair(p.Code, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, payload := range sqliPayloads {
		if err := s.SetPairOwnerVcard(act.ID, payload); err != nil {
			t.Fatal(err)
		}
		got, err := s.GetActiveByTelegram(8)
		if err != nil || got == nil || got.OwnerVcard != payload {
			t.Fatalf("vcard: %v %+v", err, got)
		}
	}
	assertCoreSchema(t, s)
}

func TestSQLInjectionPersonaBotStrings(t *testing.T) {
	s := openSQLiStore(t)
	for i, payload := range sqliPayloads {
		b, err := s.InsertPersonaBot(&PersonaBot{
			OwnerTelegramUserID: int64(3000 + i),
			OwnerDCAddress:      payload,
			OwnerVcard:          payload,
			BotToken:            payload,
			BotUserID:           int64(8000 + i),
			BotUsername:         payload,
			BotURL:              payload,
		})
		if err != nil {
			t.Fatalf("insert %q: %v", payload, err)
		}
		got, err := s.GetPersonaBot(b.ID)
		if err != nil || got == nil {
			t.Fatal(err)
		}
		if got.BotToken != payload || got.OwnerVcard != payload || got.BotUsername != payload ||
			got.BotURL != payload || got.OwnerDCAddress != payload {
			t.Fatalf("persona fields not literal for %q: %+v", payload, got)
		}
		if err := s.UpdatePersonaOwnerVcard(b.ID, payload+"-upd"); err != nil {
			t.Fatal(err)
		}
		if _, err := s.UpdatePersonaOwnerVcardForOwner(b.OwnerTelegramUserID, payload); err != nil {
			t.Fatal(err)
		}
		if err := s.SetPersonaBotStatus(b.ID, payload); err != nil {
			t.Fatal(err)
		}
		n, err := s.DisablePersonaBotByOwner(b.OwnerTelegramUserID, 0, payload)
		if err != nil {
			t.Fatalf("disable by username injection: %v", err)
		}
		_ = n
		listed, err := s.ListPersonaBotsByOwner(b.OwnerTelegramUserID)
		if err != nil {
			t.Fatal(err)
		}
		if len(listed) != 1 {
			t.Fatalf("owner list: %d", len(listed))
		}
	}
	assertCoreSchema(t, s)
	if _, err := s.CountPersonaBots(false); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CountPersonaBots(true); err != nil {
		t.Fatal(err)
	}
}

func TestSQLInjectionGhostAndGroupStrings(t *testing.T) {
	s := openSQLiStore(t)
	bot, err := s.InsertPersonaBot(&PersonaBot{
		OwnerTelegramUserID: 1,
		BotToken:            "1:tok",
		BotUserID:           1,
		BotUsername:         "b",
	})
	if err != nil {
		t.Fatal(err)
	}
	for i, payload := range sqliPayloads {
		g, err := s.InsertGhostAccount(&GhostAccount{
			PersonaBotID:     bot.ID,
			TelegramUserID:   int64(10_000 + i),
			TelegramUsername: payload,
			DisplayName:      payload,
			DCAccountID:      uint32(100 + i),
			DCAddress:        payload,
			AvatarFileID:     payload,
		})
		if err != nil {
			t.Fatalf("ghost: %v", err)
		}
		if err := s.UpdateGhostProfile(g.ID, payload, payload); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateGhostAvatar(g.ID, payload); err != nil {
			t.Fatal(err)
		}
		got, err := s.GetGhostByTG(bot.ID, g.TelegramUserID)
		if err != nil || got == nil {
			t.Fatal(err)
		}
		if got.TelegramUsername != payload || got.DisplayName != payload || got.DCAddress != payload {
			t.Fatalf("ghost fields: %+v", got)
		}

		gg, err := s.InsertGhostGroup(&GhostGroup{
			PersonaBotID:   bot.ID,
			TelegramChatID: int64(-100000 - i),
			Title:          payload,
			InviteQR:       payload,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateGhostGroupMeta(gg.ID, payload, payload, 1, 1); err != nil {
			t.Fatal(err)
		}
		gg2, err := s.GetGhostGroup(bot.ID, gg.TelegramChatID)
		if err != nil || gg2 == nil || gg2.Title != payload || gg2.InviteQR != payload {
			t.Fatalf("group: %v %+v", err, gg2)
		}
	}
	assertCoreSchema(t, s)
}

func TestSQLInjectionEncryptedStore(t *testing.T) {
	key := testKey(t)
	s, err := OpenOpts(filepath.Join(t.TempDir(), "sqli-enc.db"), Options{Key: key})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	payload := `'; DROP TABLE pairs;--`
	p, err := s.CreatePendingPair(1, payload, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetPendingByCode(payload); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetPendingByCode(p.Code)
	if err != nil || got == nil || got.TelegramUsername != payload {
		t.Fatalf("enc lookup: %v %+v", err, got)
	}
	if err := s.SetPairOwnerVcard(p.ID, payload); err != nil {
		t.Fatal(err)
	}
	b, err := s.InsertPersonaBot(&PersonaBot{
		OwnerTelegramUserID: 1,
		BotToken:            payload,
		OwnerVcard:          payload,
		BotUserID:           42,
		BotUsername:         payload,
	})
	if err != nil {
		t.Fatal(err)
	}
	gb, err := s.GetPersonaBot(b.ID)
	if err != nil || gb.BotToken != payload {
		t.Fatalf("enc token: %v %+v", err, gb)
	}
	assertCoreSchema(t, s)
}

func TestSQLInjectionDoesNotCreateAttackerTables(t *testing.T) {
	s := openSQLiStore(t)
	_, err := s.CreatePendingPair(1, "'; CREATE TABLE pwned(x);--", 1)
	if err != nil {
		t.Fatal(err)
	}
	if tableExists(t, s, "pwned") {
		t.Fatal("attacker table created")
	}
	var extra int
	err = s.db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT IN
		 ('pairs','persona_bots','ghost_accounts','ghost_groups','ghost_group_members','sqlite_sequence')`,
	).Scan(&extra)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	if extra != 0 {
		t.Fatalf("unexpected tables: %d", extra)
	}
}

func TestCountQueriesUseBoundOwnerID(t *testing.T) {
	s := openSQLiStore(t)
	if _, err := s.InsertPersonaBot(&PersonaBot{OwnerTelegramUserID: 1, BotToken: "a", BotUserID: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertPersonaBot(&PersonaBot{OwnerTelegramUserID: 2, BotToken: "b", BotUserID: 2}); err != nil {
		t.Fatal(err)
	}
	n, err := s.CountPersonaBotsByOwner(1, false)
	if err != nil || n != 1 {
		t.Fatalf("owner 1: %v %d", err, n)
	}
	n, err = s.CountPersonaBotsByOwner(2, true)
	if err != nil || n != 1 {
		t.Fatalf("owner 2 active: %v %d", err, n)
	}
}
