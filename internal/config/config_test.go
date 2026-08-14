package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPairingAndDBKeyFromEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("TGPORTAL_DB_KEY", "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseKey == "" {
		t.Fatal("expected key from env")
	}
	if cfg.Pairing.CodeLength != 8 || cfg.Pairing.PendingTTLSec != 1800 {
		t.Fatalf("pairing defaults: %+v", cfg.Pairing)
	}
	if cfg.Persona.MaxBotsPerOwner != 3 || cfg.Persona.MaxGhostsPerBot != 200 {
		t.Fatalf("persona caps: %+v", cfg.Persona)
	}
	if !cfg.DatabaseEncrypt {
		t.Fatal("database_encrypt must default on for public hosts")
	}
	if cfg.Persona.AllowGroups {
		t.Fatal("allow_groups must default off")
	}
	if strings.Contains(cfg.Name, cfg.DatabaseKey) {
		t.Fatal("key leaked into name")
	}
}

func TestLoadDatabaseKeyFromYAML(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("TGPORTAL_DB_KEY", "")
	dir := t.TempDir()
	path := filepath.Join(dir, "c.yml")
	if err := os.WriteFile(path, []byte("database_key: aabbccddeeff00112233445566778899aabbccddeeff0011223344556677\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseKey == "" {
		t.Fatal("expected yaml key")
	}
}
