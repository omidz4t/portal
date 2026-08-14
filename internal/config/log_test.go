package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLogTargetYAML(t *testing.T) {
	cases := []struct {
		yaml string
		mode string
		path string
	}{
		{`log: false`, "off", ""},
		{`log: true`, "stderr", ""},
		{`log: stderr`, "stderr", ""},
		{`log: stdout`, "stdout", ""},
		{`log: ./data/app.log`, "file", "./data/app.log"},
		{`log: off`, "off", ""},
	}
	for _, tc := range cases {
		var wrap struct {
			Log LogTarget `yaml:"log"`
		}
		if err := yaml.Unmarshal([]byte(tc.yaml), &wrap); err != nil {
			t.Fatalf("%s: %v", tc.yaml, err)
		}
		if wrap.Log.Mode != tc.mode || wrap.Log.Path != tc.path {
			t.Fatalf("%s: got mode=%q path=%q want mode=%q path=%q",
				tc.yaml, wrap.Log.Mode, wrap.Log.Path, tc.mode, tc.path)
		}
	}
}

func TestNewLoggerFileAndNop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "t.log")

	// off
	log, sync, err := NewLogger(Config{Log: LogTarget{Mode: "off"}, LogLevel: "info"})
	if err != nil {
		t.Fatal(err)
	}
	log.Info("should not appear")
	sync()

	// file
	log2, sync2, err := NewLogger(Config{
		Log:      LogTarget{Mode: "file", Path: path},
		LogLevel: "debug",
	})
	if err != nil {
		t.Fatal(err)
	}
	log2.Info("hello file")
	sync2()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected log file content")
	}
}

func TestParseLogLevel(t *testing.T) {
	lvl, err := parseLogLevel("warn")
	if err != nil || lvl.String() != "warn" {
		t.Fatalf("warn: %v %v", lvl, err)
	}
	if _, err := parseLogLevel("nope"); err == nil {
		t.Fatal("expected error")
	}
}
