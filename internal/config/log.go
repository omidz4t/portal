package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

// LogTarget is the `log` config field.
// Default is off (false). Enable with "stderr", "stdout", or a file path.
//
// YAML examples:
//
//	log: false
//	log: stderr
//	log: ./data/tgportal.log
//	log: true          # same as stderr
type LogTarget struct {
	// Mode is "off", "stderr", "stdout", or "file".
	Mode string
	// Path is set when Mode == "file".
	Path string
}

// UnmarshalYAML accepts bool false/true or a string destination.
func (t *LogTarget) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Tag == "!!null" || value.Value == "" && value.Kind == yaml.ScalarNode {
		*t = LogTarget{Mode: "off"}
		return nil
	}
	var b bool
	if err := value.Decode(&b); err == nil {
		if b {
			*t = LogTarget{Mode: "stderr"}
		} else {
			*t = LogTarget{Mode: "off"}
		}
		return nil
	}
	var s string
	if err := value.Decode(&s); err != nil {
		return fmt.Errorf("log: expected false, stderr, stdout, or file path: %w", err)
	}
	return t.parse(s)
}

func (t *LogTarget) parse(s string) error {
	s = strings.TrimSpace(s)
	switch strings.ToLower(s) {
	case "", "false", "off", "none", "disabled", "0", "no":
		*t = LogTarget{Mode: "off"}
	case "true", "1", "yes", "on", "stderr", "err", "console":
		*t = LogTarget{Mode: "stderr"}
	case "stdout", "out":
		*t = LogTarget{Mode: "stdout"}
	default:
		// Treat as filesystem path for log file.
		*t = LogTarget{Mode: "file", Path: s}
	}
	return nil
}

// Enabled reports whether application logging is on.
func (t LogTarget) Enabled() bool {
	return t.Mode != "" && t.Mode != "off"
}

// String for diagnostics.
func (t LogTarget) String() string {
	switch t.Mode {
	case "file":
		return t.Path
	case "":
		return "off"
	default:
		return t.Mode
	}
}

// NewLogger builds a zap sugared logger from log / log_level config.
// The returned sync function should be deferred (may be a no-op).
func NewLogger(cfg Config) (*zap.SugaredLogger, func(), error) {
	level, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		return nil, func() {}, err
	}

	if !cfg.Log.Enabled() {
		nop := zap.NewNop().Sugar()
		return nop, func() { _ = nop.Sync() }, nil
	}

	encCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var ws zapcore.WriteSyncer
	switch cfg.Log.Mode {
	case "stdout":
		ws = zapcore.AddSync(os.Stdout)
	case "stderr":
		ws = zapcore.AddSync(os.Stderr)
	case "file":
		path := cfg.Log.Path
		if path == "" {
			return nil, func() {}, fmt.Errorf("log: empty file path")
		}
		if !filepath.IsAbs(path) {
			// Relative paths stay relative to process cwd (same as other config paths).
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
			return nil, func() {}, fmt.Errorf("log: create dir for %q: %w", path, err)
		}
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, func() {}, fmt.Errorf("log: open %q: %w", path, err)
		}
		ws = zapcore.AddSync(f)
	default:
		return nil, func() {}, fmt.Errorf("log: unknown mode %q", cfg.Log.Mode)
	}

	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encCfg), ws, level)
	logger := zap.New(core, zap.AddCaller()).Sugar()
	syncFn := func() { _ = logger.Sync() }
	return logger, syncFn, nil
}

func parseLogLevel(s string) (zapcore.Level, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return zapcore.InfoLevel, nil
	}
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(s)); err != nil {
		return zapcore.InfoLevel, fmt.Errorf("log_level: %w (use debug, info, warn, error)", err)
	}
	return lvl, nil
}
