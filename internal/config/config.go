package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"

	"github.com/omidz4t/portal/assets"
	"github.com/omidz4t/portal/internal/proxy"
)

// DefaultFolder is the bot data directory when neither --folder nor config.yml sets one.
const DefaultFolder = "./data"

// DefaultName is the Delta Chat display name applied when config omits name.
const DefaultName = "Delta ↔️ TG"

// DefaultImage is the Delta Chat profile avatar path when config omits image.
const DefaultImage = "./assets/logo.jpg"

// DefaultTelegramLogo is shown on /start (Telegram photo).
const DefaultTelegramLogo = "./assets/logo.jpg"

// DefaultStartAnimation is the /start video/animation (Telegram treats as GIF/animation).
const DefaultStartAnimation = "./assets/start_black_hole.mp4"

// App modes: personal (portal pair bridge), persona (user-owned bots + ghost DC accounts), both.
const (
	ModePersonal = "personal"
	ModePersona  = "persona"
	ModeBoth     = "both"
)

// Config holds bot settings from YAML + environment.
type Config struct {
	// Mode selects operating modes: personal | persona | both (default both).
	Mode string `yaml:"mode"`

	// Reply is the text sent in response to plain DC messages.
	Reply string `yaml:"reply"`

	// BootMessage is sent to INVITE_URL when the bot starts serving.
	BootMessage string `yaml:"boot_message"`

	// Name is the bot display name (Delta Chat "displayname").
	Name string `yaml:"name"`

	// Image is a path to the profile avatar (Delta Chat "selfavatar").
	Image string `yaml:"image"`

	// Folder is the bot data directory (maps to --folder).
	Folder string `yaml:"folder"`

	// DatabasePath is the SQLite file for pairing (relative to folder unless absolute).
	DatabasePath string `yaml:"database"`

	// DatabaseKey encrypts secrets in SQLite (prefer TGPORTAL_DB_KEY in .env).
	DatabaseKey string `yaml:"database_key"`

	// DatabaseEncrypt when true refuses to open the store without a key.
	DatabaseEncrypt bool `yaml:"database_encrypt"`

	// Pairing codes and pending TTL.
	Pairing Pairing `yaml:"pairing"`

	// Account limits operations to this account ID (maps to --account). 0 means all.
	Account uint32 `yaml:"account"`

	// Log controls application log destination.
	// Default false (off). Set to "stderr", "stdout", or a file path (e.g. ./data/tgportal.log).
	Log LogTarget `yaml:"log"`

	// LogLevel is debug|info|warn|error. Default "info" when logging is enabled.
	LogLevel string `yaml:"log_level"`

	// Proxy is the shared default proxy (SOCKS5/HTTP) for Telegram and Delta Chat.
	// Per-side telegram.proxy / deltachat.proxy override this.
	Proxy proxy.Config `yaml:"proxy"`

	// Telegram bot settings (token from env only).
	Telegram Telegram `yaml:"telegram"`

	// Deltachat holds Delta Chat–specific options (proxy, etc.).
	Deltachat Deltachat `yaml:"deltachat"`

	// Persona is mode-2 settings (user-owned bots + per-TG-user ghost DC accounts).
	Persona Persona `yaml:"persona"`

	// Bridge toggles for Telegram → Delta Chat media.
	Bridge Bridge `yaml:"bridge"`

	// InviteURL is optional admin DC invite from .env (INVITE_URL).
	InviteURL string `yaml:"-"`

	// TelegramToken from .env (TELEGRAM_BOT_TOKEN).
	TelegramToken string `yaml:"-"`
}

// Persona configures user-owned Telegram bots and ghost Delta Chat accounts.
type Persona struct {
	// Enabled is true when mode is persona or both (also set from mode on load).
	Enabled bool `yaml:"enabled"`

	// AccountQR is a dcaccount:/dclogin: URI used to provision ghost accounts.
	// Prefer PERSONA_ACCOUNT_QR in .env for production.
	AccountQR string `yaml:"account_qr"`

	// MaxGhosts caps total ghost DC accounts in the store.
	MaxGhosts int `yaml:"max_ghosts"`

	// MaxBots caps registered user-owned Telegram bots.
	MaxBots int `yaml:"max_bots"`

	// MaxBotsPerOwner caps active persona bots for one Telegram owner.
	MaxBotsPerOwner int `yaml:"max_bots_per_owner"`

	// MaxGhostsPerBot caps ghost accounts created for one persona bot.
	MaxGhostsPerBot int `yaml:"max_ghosts_per_bot"`

	// AllowRegisterFromTG allows /pair-bot on the portal Telegram bot.
	AllowRegisterFromTG bool `yaml:"allow_register_from_tg"`

	// AllowGroups mirrors Telegram groups. Default false (DMs only) for public hosts.
	AllowGroups bool `yaml:"allow_groups"`
}

// Pairing controls pairing-code length and lifetime.
type Pairing struct {
	CodeLength    int `yaml:"code_length"`
	PendingTTLSec int `yaml:"pending_ttl_sec"`
}

// Telegram controls the Telegram bot side of Portal.
type Telegram struct {
	Enabled        bool         `yaml:"enabled"`
	BotURL         string       `yaml:"bot_url"`
	Logo           string       `yaml:"logo"`
	StartAnimation string       `yaml:"start_animation"`
	AllowedUserIDs []int64      `yaml:"allowed_user_ids"`
	Proxy          proxy.Config `yaml:"proxy"`

	// Reaction is the emoji used on Telegram after a successful bridge
	// (setMessageReaction). Default "✅". Use "off" or empty with explicit
	// disable via "none"/"off"/"-" to disable. Empty string falls back to default.
	Reaction string `yaml:"reaction"`
}

// Deltachat holds Delta Chat core options applied via set_config.
type Deltachat struct {
	Proxy proxy.Config `yaml:"proxy"`
}

// Bridge selects which Telegram media types are forwarded to Delta Chat.
type Bridge struct {
	// Text bridges plain text messages both ways.
	Text bool `yaml:"text"`
	// Images bridges photos (JPG/PNG, etc.).
	Images bool `yaml:"images"`
	// Videos bridges short video messages (subject to Limits).
	Videos bool `yaml:"videos"`

	Stickers      bool `yaml:"stickers"`
	Lottie        bool `yaml:"lottie"`
	VideoStickers bool `yaml:"video_stickers"`
	Gif           bool `yaml:"gif"`
	// CustomEmoji bridges Telegram custom/premium emojis (getCustomEmojiStickers).
	CustomEmoji bool `yaml:"custom_emoji"`
	// StickerPacks enables /send_pack (full pack from a quoted sticker).
	StickerPacks bool `yaml:"sticker_packs"`

	// Limits caps media size/duration (0 = unlimited for that field).
	Limits BridgeLimits `yaml:"limits"`
}

// BridgeLimits size/duration caps for media (bytes / seconds).
// Zero means no limit for that field (except sticker_pack_max uses a default).
type BridgeLimits struct {
	// VideoMaxDurationSec is max video length in seconds (short videos). Default 60.
	VideoMaxDurationSec int `yaml:"video_max_duration_sec"`
	// VideoMaxBytes is max video file size in bytes. Default 20 MiB.
	VideoMaxBytes int64 `yaml:"video_max_bytes"`
	// ImageMaxBytes is max photo/image size in bytes. Default 10 MiB.
	ImageMaxBytes int64 `yaml:"image_max_bytes"`
	// FileMaxBytes is max other file size in bytes. Default 20 MiB.
	FileMaxBytes int64 `yaml:"file_max_bytes"`
	// StickerPackMax is max stickers to send from one /send_pack. Default 120.
	StickerPackMax int `yaml:"sticker_pack_max"`
}

// Default bridge limits.
const (
	DefaultVideoMaxDurationSec = 60
	DefaultVideoMaxBytes       = 20 * 1024 * 1024 // 20 MiB
	DefaultImageMaxBytes       = 10 * 1024 * 1024 // 10 MiB
	DefaultFileMaxBytes        = 20 * 1024 * 1024 // 20 MiB
	DefaultStickerPackMax      = 120
)

func defaults() Config {
	return Config{
		Mode:            ModePersonal,
		Reply:           "hi",
		BootMessage:     "hi — bot is online",
		Name:            DefaultName,
		Image:           DefaultImage,
		Folder:          DefaultFolder,
		DatabaseEncrypt: true,
		Log:             LogTarget{Mode: "off"},
		LogLevel:        "info",
		Telegram: Telegram{
			Enabled:        true,
			BotURL:         "https://t.me/tgdeltabridgebot",
			Logo:           DefaultTelegramLogo,
			StartAnimation: DefaultStartAnimation,
			Reaction:       "✅",
		},
		Persona: Persona{
			Enabled:             true,
			MaxGhosts:           200,
			MaxBots:             20,
			MaxBotsPerOwner:     3,
			MaxGhostsPerBot:     200,
			AllowRegisterFromTG: true,
		},
		Pairing: Pairing{
			CodeLength:    8,
			PendingTTLSec: 1800,
		},
		Bridge: Bridge{
			Text:          true,
			Images:        true,
			Videos:        true,
			Stickers:      true,
			Lottie:        true,
			VideoStickers: true,
			Gif:           true,
			CustomEmoji:   true,
			StickerPacks:  true,
			Limits: BridgeLimits{
				VideoMaxDurationSec: DefaultVideoMaxDurationSec,
				VideoMaxBytes:       DefaultVideoMaxBytes,
				ImageMaxBytes:       DefaultImageMaxBytes,
				FileMaxBytes:        DefaultFileMaxBytes,
				StickerPackMax:      DefaultStickerPackMax,
			},
		},
	}
}

// LoadDotEnv loads .env from the working directory if present.
func LoadDotEnv() {
	_ = godotenv.Load()
}

// Load reads YAML path and merges secrets from the environment.
func Load(path string) (Config, error) {
	LoadDotEnv()

	cfg := defaults()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return cfg, fmt.Errorf("read config %q: %w", path, err)
			}
		} else if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parse config %q: %w", path, err)
		}
	}

	cfg.Mode = strings.ToLower(strings.TrimSpace(cfg.Mode))
	if cfg.Mode == "" {
		cfg.Mode = ModePersonal
	}
	switch cfg.Mode {
	case ModePersonal, ModePersona, ModeBoth:
	default:
		return cfg, fmt.Errorf("config mode %q: want personal, persona, or both", cfg.Mode)
	}
	// Mode is authoritative for whether persona runs.
	cfg.Persona.Enabled = cfg.Mode == ModePersona || cfg.Mode == ModeBoth
	if cfg.Persona.MaxGhosts <= 0 {
		cfg.Persona.MaxGhosts = 200
	}
	if cfg.Persona.MaxBots <= 0 {
		cfg.Persona.MaxBots = 20
	}
	if cfg.Persona.MaxBotsPerOwner <= 0 {
		cfg.Persona.MaxBotsPerOwner = 3
	}
	if cfg.Persona.MaxGhostsPerBot <= 0 {
		cfg.Persona.MaxGhostsPerBot = 200
	}
	if cfg.Pairing.CodeLength < 4 || cfg.Pairing.CodeLength > 12 {
		cfg.Pairing.CodeLength = 8
	}
	if cfg.Pairing.PendingTTLSec <= 0 {
		cfg.Pairing.PendingTTLSec = 1800
	}
	if qr := strings.TrimSpace(os.Getenv("PERSONA_ACCOUNT_QR")); qr != "" {
		cfg.Persona.AccountQR = qr
	}
	cfg.Persona.AccountQR = strings.TrimSpace(cfg.Persona.AccountQR)

	if cfg.Reply == "" {
		cfg.Reply = defaults().Reply
	}
	if cfg.BootMessage == "" {
		cfg.BootMessage = defaults().BootMessage
	}
	if cfg.Name == "" {
		cfg.Name = DefaultName
	}
	if cfg.Folder == "" {
		cfg.Folder = DefaultFolder
	}
	if cfg.Telegram.BotURL == "" {
		cfg.Telegram.BotURL = "https://t.me/tgdeltabridgebot"
	}
	if cfg.Telegram.Logo == "" {
		cfg.Telegram.Logo = DefaultTelegramLogo
	}
	if cfg.Telegram.StartAnimation == "" {
		cfg.Telegram.StartAnimation = DefaultStartAnimation
	}
	if cfg.Telegram.Reaction == "" {
		cfg.Telegram.Reaction = "✅"
	}
	if cfg.Image == "" {
		cfg.Image = DefaultImage
	}
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = "tgportal.db"
	}
	if cfg.Log.Mode == "" {
		cfg.Log = LogTarget{Mode: "off"}
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	// Apply default limits only when left at zero (omit or empty means use defaults).
	// To disable a limit explicitly, set a very large value or -1 is treated as unlimited (0 after load).
	if cfg.Bridge.Limits.VideoMaxDurationSec == 0 {
		cfg.Bridge.Limits.VideoMaxDurationSec = DefaultVideoMaxDurationSec
	}
	if cfg.Bridge.Limits.VideoMaxDurationSec < 0 {
		cfg.Bridge.Limits.VideoMaxDurationSec = 0 // unlimited
	}
	if cfg.Bridge.Limits.VideoMaxBytes == 0 {
		cfg.Bridge.Limits.VideoMaxBytes = DefaultVideoMaxBytes
	}
	if cfg.Bridge.Limits.VideoMaxBytes < 0 {
		cfg.Bridge.Limits.VideoMaxBytes = 0
	}
	if cfg.Bridge.Limits.ImageMaxBytes == 0 {
		cfg.Bridge.Limits.ImageMaxBytes = DefaultImageMaxBytes
	}
	if cfg.Bridge.Limits.ImageMaxBytes < 0 {
		cfg.Bridge.Limits.ImageMaxBytes = 0
	}
	if cfg.Bridge.Limits.FileMaxBytes == 0 {
		cfg.Bridge.Limits.FileMaxBytes = DefaultFileMaxBytes
	}
	if cfg.Bridge.Limits.FileMaxBytes < 0 {
		cfg.Bridge.Limits.FileMaxBytes = 0
	}
	if cfg.Bridge.Limits.StickerPackMax == 0 {
		cfg.Bridge.Limits.StickerPackMax = DefaultStickerPackMax
	}
	if cfg.Bridge.Limits.StickerPackMax < 0 {
		cfg.Bridge.Limits.StickerPackMax = DefaultStickerPackMax
	}

	cfg.InviteURL = strings.TrimSpace(os.Getenv("INVITE_URL"))
	cfg.TelegramToken = strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if k := strings.TrimSpace(os.Getenv("TGPORTAL_DB_KEY")); k != "" {
		cfg.DatabaseKey = k
	}
	cfg.DatabaseKey = strings.TrimSpace(cfg.DatabaseKey)

	if raw := strings.TrimSpace(os.Getenv("TELEGRAM_ALLOWED_USER_IDS")); raw != "" {
		ids, err := parseInt64List(raw)
		if err != nil {
			return cfg, fmt.Errorf("TELEGRAM_ALLOWED_USER_IDS: %w", err)
		}
		cfg.Telegram.AllowedUserIDs = ids
	}

	// Resolve brand assets: prefer on-disk paths; otherwise unpack embeds (single-binary release).
	if err := resolveEmbeddedAssets(&cfg); err != nil {
		return cfg, err
	}

	// Proxy env overrides (highest priority for URLs).
	// PROXY_URL / ALL_PROXY → shared default
	// TELEGRAM_PROXY_URL, DELTACHAT_PROXY_URL → per side
	if u := firstEnv("PROXY_URL", "ALL_PROXY"); u != "" {
		cfg.Proxy.URL = u
	}
	if u := strings.TrimSpace(os.Getenv("TELEGRAM_PROXY_URL")); u != "" {
		cfg.Telegram.Proxy.URL = u
	}
	if u := strings.TrimSpace(os.Getenv("DELTACHAT_PROXY_URL")); u != "" {
		cfg.Deltachat.Proxy.URL = u
	}
	if v := strings.TrimSpace(os.Getenv("PROXY_ENABLED")); v != "" {
		b := v == "1" || strings.EqualFold(v, "true") || v == "yes"
		cfg.Proxy.Enabled = &b
	}

	if err := cfg.TelegramProxy().Validate(); err != nil {
		return cfg, fmt.Errorf("telegram proxy: %w", err)
	}
	if err := cfg.DeltachatProxy().Validate(); err != nil {
		return cfg, fmt.Errorf("deltachat proxy: %w", err)
	}

	return cfg, nil
}

// TelegramProxy returns effective Telegram proxy (side override + shared default).
func (c Config) TelegramProxy() proxy.Config {
	return proxy.Merge(c.Proxy, c.Telegram.Proxy)
}

// DeltachatProxy returns effective Delta Chat proxy.
func (c Config) DeltachatProxy() proxy.Config {
	return proxy.Merge(c.Proxy, c.Deltachat.Proxy)
}

// PersonalEnabled reports whether the classic portal pairing bridge should run.
func (c Config) PersonalEnabled() bool {
	return c.Mode == ModePersonal || c.Mode == ModeBoth
}

// PersonaEnabled reports whether user-owned persona bots / ghost accounts are active.
func (c Config) PersonaEnabled() bool {
	return c.Mode == ModePersona || c.Mode == ModeBoth
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

// resolveEmbeddedAssets fills logo/avatar/animation paths from the binary embed
// when the configured path is missing (all-in-one release binary).
func resolveEmbeddedAssets(cfg *Config) error {
	dest := filepath.Join(cfg.Folder, "assets")
	var err error
	if cfg.Image != "" {
		if cfg.Image, err = assets.ResolvePath(cfg.Image, assets.LogoJPG, dest); err != nil {
			// image optional if empty after failure — keep soft for custom missing paths
			if p, e2 := assets.Ensure(assets.LogoJPG, dest); e2 == nil {
				cfg.Image = p
			}
		}
	}
	if cfg.Telegram.Logo != "" {
		if cfg.Telegram.Logo, err = assets.ResolvePath(cfg.Telegram.Logo, assets.LogoJPG, dest); err != nil {
			if p, e2 := assets.Ensure(assets.LogoJPG, dest); e2 == nil {
				cfg.Telegram.Logo = p
			}
		}
	}
	if cfg.Telegram.StartAnimation != "" {
		if cfg.Telegram.StartAnimation, err = assets.ResolvePath(cfg.Telegram.StartAnimation, assets.StartBlackHoleMP4, dest); err != nil {
			if p, e2 := assets.Ensure(assets.StartBlackHoleMP4, dest); e2 == nil {
				cfg.Telegram.StartAnimation = p
			}
		}
	}
	return nil
}

func parseInt64List(raw string) ([]int64, error) {
	parts := strings.Split(raw, ",")
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}
