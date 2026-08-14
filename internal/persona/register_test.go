package persona

import (
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"github.com/omidz4t/portal/internal/config"
	"github.com/omidz4t/portal/internal/store"
)

func testPersonaCfg() config.Config {
	return config.Config{
		Mode:          config.ModeBoth,
		TelegramToken: "9:PORTAL",
		Persona: config.Persona{
			Enabled:         true,
			AccountQR:       "dcaccount:test.example",
			MaxBots:         20,
			MaxBotsPerOwner: 5,
			MaxGhosts:       200,
			MaxGhostsPerBot: 50,
		},
	}
}

func nopStartPoller(b store.PersonaBot) (*BotRuntime, error) {
	return &BotRuntime{
		Bot:     b,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}, nil
}

func TestRegisterBotReactivatesDisabledSameOwner(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "re-pair.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	first, err := st.InsertPersonaBot(&store.PersonaBot{
		OwnerTelegramUserID: 42,
		OwnerVcard:          "BEGIN:VCARD\nEND:VCARD",
		BotToken:            "1:old-token",
		BotUserID:           7001,
		BotUsername:         "mybot",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.DisablePersonaBotByOwner(42, first.ID, ""); err != nil {
		t.Fatal(err)
	}

	m := New(zap.NewNop().Sugar(), testPersonaCfg(), nil, st)
	saved, err := m.RegisterBot(&store.PersonaBot{
		OwnerTelegramUserID: 42,
		OwnerVcard:          "BEGIN:VCARD\nFN:Me\nEND:VCARD",
		BotToken:            "1:new-token",
		BotUserID:           7001,
		BotUsername:         "mybot",
		BotURL:              "https://t.me/mybot",
	}, nopStartPoller)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID != first.ID {
		t.Fatalf("expected reuse id %d got %d", first.ID, saved.ID)
	}
	if saved.Status != store.PersonaBotActive || saved.BotToken != "1:new-token" {
		t.Fatalf("reactivated: %+v", saved)
	}
	list, err := st.ListPersonaBotsByOwner(42)
	if err != nil || len(list) != 1 {
		t.Fatalf("unique row: %v n=%d", err, len(list))
	}
}

func TestRegisterBotRejectsActiveDuplicate(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "dup.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if _, err := st.InsertPersonaBot(&store.PersonaBot{
		OwnerTelegramUserID: 1,
		OwnerVcard:          "v",
		BotToken:            "1:t",
		BotUserID:           5,
	}); err != nil {
		t.Fatal(err)
	}
	m := New(zap.NewNop().Sugar(), testPersonaCfg(), nil, st)
	if _, err := m.RegisterBot(&store.PersonaBot{
		OwnerTelegramUserID: 1,
		OwnerVcard:          "v",
		BotToken:            "1:t",
		BotUserID:           5,
	}, nopStartPoller); err == nil {
		t.Fatal("expected already registered")
	}
}

func TestDropPollerAllowsReattach(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "drop.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	m := New(zap.NewNop().Sugar(), testPersonaCfg(), nil, st)
	if err := m.AttachBot(store.PersonaBot{ID: 3, BotUsername: "x"}, nopStartPoller); err != nil {
		t.Fatal(err)
	}
	if m.Runtime(3) == nil {
		t.Fatal("attached")
	}
	m.dropPoller(3)
	if m.Runtime(3) != nil {
		t.Fatal("drop must remove so AttachBot can start again")
	}
	if err := m.AttachBot(store.PersonaBot{ID: 3, BotUsername: "x"}, nopStartPoller); err != nil {
		t.Fatal(err)
	}
	if m.Runtime(3) == nil {
		t.Fatal("reattach")
	}
}

func TestRegisterBotRejectsOtherOwner(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "steal.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	first, err := st.InsertPersonaBot(&store.PersonaBot{
		OwnerTelegramUserID: 1,
		OwnerVcard:          "v",
		BotToken:            "1:t",
		BotUserID:           8,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.DisablePersonaBotByOwner(1, first.ID, ""); err != nil {
		t.Fatal(err)
	}
	m := New(zap.NewNop().Sugar(), testPersonaCfg(), nil, st)
	if _, err := m.RegisterBot(&store.PersonaBot{
		OwnerTelegramUserID: 99,
		OwnerVcard:          "v",
		BotToken:            "1:t",
		BotUserID:           8,
	}, nopStartPoller); err == nil {
		t.Fatal("other owner must not take disabled bot")
	}
	got, err := st.GetPersonaBot(first.ID)
	if err != nil || got.Status != store.PersonaBotDisabled || got.OwnerTelegramUserID != 1 {
		t.Fatalf("original owner must keep row: %v %+v", err, got)
	}
}
