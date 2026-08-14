package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/omidz4t/portal/internal/bridge"
	"github.com/omidz4t/portal/internal/config"
	"github.com/omidz4t/portal/internal/proxy"
	"github.com/omidz4t/portal/internal/ratelimit"
	"github.com/omidz4t/portal/internal/safemedia"
	"github.com/omidz4t/portal/internal/store"
	"github.com/omidz4t/portal/internal/usermsg"
)

// PersonaHooks is optional mode-2 registration (user-owned bots).
type PersonaHooks interface {
	// RegisterToken validates token, stores bot, starts poller. Never log the token.
	RegisterToken(ownerTG int64, ownerUsername string, token, botURL string) (username string, err error)
	// Unregister disables owner's bot(s). botRef is id, @username, or empty for all.
	Unregister(ownerTG int64, botRef string) (n int64, err error)
	// ListBots returns a short status text for the owner.
	ListBots(ownerTG int64) (string, error)
}

// Bot long-polls Telegram and forwards stickers / Lottie / GIFs to the bridge.
type Bot struct {
	api     *tgbotapi.BotAPI
	http    *http.Client
	log     *zap.SugaredLogger
	cfg     config.Config
	bridge  *bridge.Bridge
	store   *store.Store
	tmpdir  string
	persona PersonaHooks
	limits  *ratelimit.Set
}

// Start creates the API client, begins polling, and returns the Bot for DC→TG sends.
// persona may be nil when mode is personal-only.
func Start(log *zap.SugaredLogger, cfg config.Config, br *bridge.Bridge, st *store.Store, persona PersonaHooks, limits *ratelimit.Set) (*Bot, error) {
	if !cfg.Telegram.Enabled {
		log.Info("Telegram bot disabled in config")
		return nil, nil
	}
	if cfg.TelegramToken == "" {
		log.Warn("TELEGRAM_BOT_TOKEN not set; Telegram bridge inactive")
		return nil, nil
	}
	// Portal bot is needed for personal bridge and/or persona /pair-bot registration.
	if !cfg.PersonalEnabled() && !cfg.PersonaEnabled() {
		log.Info("neither personal nor persona mode enabled")
		return nil, nil
	}
	if cfg.PersonalEnabled() && !br.Enabled() {
		log.Warn("personal bridge not enabled (no media toggles)")
	}

	pc := cfg.TelegramProxy()
	httpClient, err := proxy.HTTPClient(pc, 120*time.Second)
	if err != nil {
		return nil, fmt.Errorf("telegram proxy: %w", err)
	}

	api, err := tgbotapi.NewBotAPIWithClient(cfg.TelegramToken, tgbotapi.APIEndpoint, httpClient)
	if err != nil {
		return nil, fmt.Errorf("telegram auth: %w", err)
	}

	tmpdir := filepath.Join(cfg.Folder, "tg-cache")
	if err := os.MkdirAll(tmpdir, 0o755); err != nil {
		return nil, err
	}

	b := &Bot{
		api:     api,
		http:    httpClient,
		log:     log.With("component", "telegram"),
		cfg:     cfg,
		bridge:  br,
		store:   st,
		tmpdir:  tmpdir,
		persona: persona,
		limits:  limits,
	}

	if pc.IsEnabled() {
		b.log.Infof("telegram proxy enabled: %s", redactedProxyURL(pc.ResolvedURL()))
	}
	if err := b.registerBotCommands(); err != nil {
		b.log.Warnf("setMyCommands: %v", err)
	}
	b.log.Infof("authorized as @%s (%s); bidirectional bridge active",
		api.Self.UserName, cfg.Telegram.BotURL)
	go b.poll()
	return b, nil
}

func redactedProxyURL(raw string) string {
	// hide userinfo in logs
	if i := strings.Index(raw, "://"); i >= 0 {
		rest := raw[i+3:]
		if at := strings.LastIndex(rest, "@"); at >= 0 {
			return raw[:i+3] + "***@" + rest[at+1:]
		}
	}
	return raw
}

// telegramWorkers bounds concurrent TG update handlers (download + DC send).
// DC RPC itself is further serialized in dc.Session.
const telegramWorkers = 8

func (b *Bot) poll() {
	// Custom getUpdates loop so we can read custom_emoji_id (not in library types).
	sem := make(chan struct{}, telegramWorkers)
	offset := 0

	for {
		params := tgbotapi.Params{}
		params.AddNonZero("timeout", 60)
		params.AddNonZero("offset", offset)
		params["allowed_updates"] = `["message"]`

		resp, err := b.api.MakeRequest("getUpdates", params)
		if err != nil {
			b.log.Errorf("getUpdates: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var rawUpdates []json.RawMessage
		if err := json.Unmarshal(resp.Result, &rawUpdates); err != nil {
			b.log.Errorf("decode updates: %v", err)
			time.Sleep(time.Second)
			continue
		}

		for _, raw := range rawUpdates {
			var meta struct {
				UpdateID int             `json:"update_id"`
				Message  json.RawMessage `json:"message"`
			}
			if err := json.Unmarshal(raw, &meta); err != nil {
				continue
			}
			if meta.UpdateID >= offset {
				offset = meta.UpdateID + 1
			}
			if len(meta.Message) == 0 || string(meta.Message) == "null" {
				continue
			}

			var msg tgbotapi.Message
			if err := json.Unmarshal(meta.Message, &msg); err != nil {
				b.log.Warnf("decode message: %v", err)
				continue
			}
			emojiIDs := extractCustomEmojiIDs(meta.Message)

			if !b.allowed(msg.From) {
				b.log.Warnf("ignore update from unauthorized user id=%v", userID(msg.From))
				continue
			}

			sem <- struct{}{}
			go func(msg tgbotapi.Message, emojiIDs []string) {
				defer func() { <-sem }()
				if err := b.handleMessage(&msg, emojiIDs); err != nil {
					b.log.Errorf("handle message: %v", err)
					_, _ = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, usermsg.BridgeFailed))
				}
			}(msg, emojiIDs)
		}
	}
}

func userID(from *tgbotapi.User) int64 {
	if from == nil {
		return 0
	}
	return from.ID
}

func (b *Bot) allowed(from *tgbotapi.User) bool {
	if from == nil {
		return false
	}
	ids := b.cfg.Telegram.AllowedUserIDs
	if len(ids) == 0 {
		return true
	}
	for _, id := range ids {
		if id == from.ID {
			return true
		}
	}
	return false
}

func (b *Bot) handleMessage(msg *tgbotapi.Message, customEmojiIDs []string) error {
	var (
		err error
		ok  bool // true when media was bridged successfully
	)
	switch {
	case msg.Sticker != nil:
		err = b.requirePairThen(msg, func() error { return b.handleSticker(msg) })
		ok = err == nil
	case msg.Photo != nil && len(msg.Photo) > 0:
		err = b.requirePairThen(msg, func() error { return b.handlePhoto(msg) })
		ok = err == nil
	case msg.Video != nil:
		err = b.requirePairThen(msg, func() error { return b.handleVideo(msg) })
		ok = err == nil
	case msg.Animation != nil:
		err = b.requirePairThen(msg, func() error { return b.handleAnimation(msg) })
		ok = err == nil
	case msg.Document != nil && isGifDocument(msg.Document):
		err = b.requirePairThen(msg, func() error { return b.handleGifDocument(msg) })
		ok = err == nil
	case len(customEmojiIDs) > 0:
		err = b.requirePairThen(msg, func() error { return b.handleCustomEmojis(msg, customEmojiIDs) })
		ok = err == nil
	case msg.Text != "" && strings.HasPrefix(msg.Text, "/"):
		return b.handleCommand(msg)
	case msg.Text != "":
		err = b.requirePairThen(msg, func() error { return b.handleText(msg) })
		ok = err == nil
	default:
		return nil
	}
	if ok {
		b.reactOK(msg)
	}
	return err
}

func (b *Bot) requirePairThen(msg *tgbotapi.Message, fn func() error) error {
	pair, err := b.store.GetActiveByTelegram(msg.From.ID)
	if err != nil {
		return err
	}
	if pair == nil {
		_, _ = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID,
			"You are not paired yet.\nSend /start to get a Delta Chat invite and pairing code."))
		return nil
	}
	if b.limits != nil && b.limits.BridgeTG != nil && !b.limits.BridgeTG.Allow(fmt.Sprintf("bridge:tg:%d", msg.From.ID)) {
		_, _ = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Too many messages. Please wait a minute."))
		return nil
	}
	return fn()
}

func (b *Bot) handleCommand(msg *tgbotapi.Message) error {
	cmd := strings.Fields(msg.Text)[0]
	if i := strings.Index(cmd, "@"); i > 0 {
		cmd = cmd[:i]
	}
	switch strings.ToLower(cmd) {
	case "/start":
		return b.cmdStart(msg)
	case "/pair", "/connect":
		return b.cmdPair(msg)
	case "/disconnect":
		return b.cmdDisconnect(msg)
	case "/send_pack":
		return b.requirePairThen(msg, func() error { return b.cmdSendPack(msg) })
	case "/help":
		return b.cmdHelp(msg)
	case "/status":
		return b.cmdStatus(msg)
	case "/pair-bot", "/pair_bot":
		return b.cmdPairBot(msg)
	case "/unpair-bot", "/unpair_bot":
		return b.cmdUnpairBot(msg)
	case "/bots":
		return b.cmdBots(msg)
	default:
		return nil
	}
}

func (b *Bot) cmdStart(msg *tgbotapi.Message) error {
	if !IsPrivatePairingChat(msg) {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, pairingPrivateOnly))
		return err
	}
	// /start CODE from t.me/bot?start=CODE (DC-initiated pairing).
	if code := startPayload(msg.Text); code != "" && store.LooksLikeCode(code) {
		return b.claimDCPairingCode(msg, code)
	}
	return b.cmdPair(msg)
}

func startPayload(text string) string {
	fields := strings.Fields(text)
	if len(fields) < 2 {
		return ""
	}
	// /start CODE or /start@bot CODE
	return fields[1]
}

// claimDCPairingCode activates a DC-issued code for this Telegram user.
func (b *Bot) claimDCPairingCode(msg *tgbotapi.Message, code string) error {
	if !IsPrivatePairingChat(msg) {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, pairingPrivateOnly))
		return err
	}
	if b.limits != nil && b.limits.ClaimTG != nil && !b.limits.ClaimTG.Allow(fmt.Sprintf("claim:tg:%d", msg.From.ID)) {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Too many pairing attempts. Wait a few minutes and try again."))
		return err
	}
	username := ""
	if msg.From != nil {
		username = msg.From.UserName
	}
	pair, err := b.store.ActivatePairFromTelegram(code, msg.From.ID, username, msg.Chat.ID)
	if err != nil {
		_, _ = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, usermsg.PairingFailed))
		return nil
	}
	if pair.Status == store.StatusPending {
		// TG-initiated style code claimed via start — still need DC to paste... but
		// if DC was empty, tell user to finish on Delta Chat.
		dcLink, _ := b.bridge.BotInviteLink()
		if dcLink == "" {
			dcLink = "(open the Delta Chat bot)"
		}
		caption := fmt.Sprintf(
			"Delta ↔️ TG — almost done\n\n"+
				"Code %s is linked to your Telegram.\n"+
				"Open Delta Chat and send the same code there:\n\n%s\n\n%s",
			pair.Code, dcLink, b.cfg.Telegram.BotURL,
		)
		if err := b.sendInviteQRPhoto(msg.Chat.ID, dcLink, caption); err != nil {
			_, _ = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, caption))
		}
		return nil
	}
	// Success on both Telegram and Delta Chat.
	b.bridge.NotifyPairingSuccess(pair)
	b.log.Infof("claimed DC pairing for tg user %d → dc chat %d", msg.From.ID, pair.DCChatID)
	return nil
}

func (b *Bot) cmdPair(msg *tgbotapi.Message) error {
	if !IsPrivatePairingChat(msg) {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, pairingPrivateOnly))
		return err
	}
	if active, err := b.store.GetActiveByTelegram(msg.From.ID); err == nil && active != nil {
		text := "You are already paired with Delta Chat ✅\n\n" +
			"Send stickers, custom emoji, GIFs, or text either way.\n\n" +
			"/disconnect — unlink\n/status — status\n/help — commands"
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, text))
		return err
	}
	if b.limits != nil && b.limits.PairTG != nil && !b.limits.PairTG.Allow(fmt.Sprintf("pair:tg:%d", msg.From.ID)) {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Too many pairing requests. Wait a few minutes."))
		return err
	}

	username := ""
	if msg.From != nil {
		username = msg.From.UserName
	}
	pair, err := b.store.CreatePendingPair(msg.From.ID, username, msg.Chat.ID)
	if err != nil {
		return err
	}

	dcLink, err := b.bridge.BotInviteLink()
	if err != nil {
		b.log.Warnf("could not get DC invite link: %v", err)
		dcLink = "(Delta Chat invite unavailable — is the DC bot configured?)"
	}

	// Single reply: QR image is the message body; all pairing info is the caption.
	caption := fmt.Sprintf(
		"Delta ↔️ TG — pair with Delta Chat\n\n"+
			"1) Scan the QR or open this invite:\n%s\n\n"+
			"2) Send this pairing code in that chat:\n%s\n\n"+
			"Or on Delta Chat message the bot first — it will give you a t.me link.\n"+
			"Commands: /help /status /disconnect\n\n"+
			"%s",
		dcLink, pair.Code, b.cfg.Telegram.BotURL,
	)
	if err := b.sendInviteQRPhoto(msg.Chat.ID, dcLink, caption); err != nil {
		b.log.Warnf("invite QR reply failed, text fallback: %v", err)
		if _, err2 := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, caption)); err2 != nil {
			return err2
		}
	}
	b.log.Infof("issued pairing code for tg user %d", msg.From.ID)
	return nil
}

func (b *Bot) cmdDisconnect(msg *tgbotapi.Message) error {
	n, err := b.store.DisconnectByTelegram(msg.From.ID)
	if err != nil {
		return err
	}
	text := "Not paired — nothing to disconnect.\nUse /pair to link Delta Chat."
	if n > 0 {
		text = "Disconnected ✅\nYour Telegram is no longer linked to Delta Chat.\n\n/pair to connect again."
	}
	_, err = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, text))
	return err
}

func (b *Bot) sendLogo(chatID int64) error {
	path := b.cfg.Telegram.Logo
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(path))
	// No file details in caption — branding only
	photo.Caption = "Delta ↔️ TG"
	_, err := b.api.Send(photo)
	return err
}

func (b *Bot) cmdHelp(msg *tgbotapi.Message) error {
	if err := b.sendLogo(msg.Chat.ID); err != nil {
		b.log.Warnf("help logo: %v", err)
	}
	_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, commandsHelpText(b.cfg)))
	return err
}

func (b *Bot) cmdPairBot(msg *tgbotapi.Message) error {
	if !b.cfg.PersonaEnabled() || b.persona == nil {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID,
			"Persona mode is disabled. Set mode: persona or both in config.yml."))
		return err
	}
	if b.limits != nil && b.limits.PairBotTG != nil && !b.limits.PairBotTG.Allow(fmt.Sprintf("pairbot:tg:%d", msg.From.ID)) {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Too many /pair-bot attempts. Try again later."))
		return err
	}
	if !b.cfg.Persona.AllowRegisterFromTG {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Registering bots from Telegram is disabled."))
		return err
	}
	if !IsPrivatePairingChat(msg) {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, pairingPrivateOnly))
		return err
	}
	pair, err := b.store.GetActiveByTelegram(msg.From.ID)
	if err != nil {
		return err
	}
	if pair == nil {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID,
			"Pair with Delta Chat first (/pair), then send:\n/pair-bot <BOT_TOKEN> [https://t.me/your_bot]"))
		return err
	}

	fields := strings.Fields(msg.Text)
	if len(fields) < 2 {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID,
			"Persona bot — register a BotFather token\n\n"+
				"1) Create a bot with @BotFather\n"+
				"2) Pair this portal account with Delta Chat (/pair) if you have not\n"+
				"3) Send (private chat only):\n"+
				"/pair-bot <TOKEN> [https://t.me/YourBot]\n\n"+
				"People who message your bot appear as Delta Chat contacts (ghost accounts).\n"+
				"Do not share your token. Delete the message after sending if you can."))
		return err
	}
	token := strings.TrimSpace(fields[1])
	botURL := ""
	if len(fields) >= 3 {
		botURL = strings.TrimSpace(fields[2])
	}
	username := ""
	if msg.From != nil {
		username = msg.From.UserName
	}
	uname, err := b.persona.RegisterToken(msg.From.ID, username, token, botURL)
	// Best-effort: try to delete the message that contained the token.
	_, _ = b.api.Request(tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID))
	if err != nil {
		_, err2 := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, usermsg.RegisterFailed))
		return err2
	}
	text := "Persona bot linked ✅\n"
	if uname != "" {
		text += "@" + uname + "\n"
	}
	text += "\nPeople who message this bot get a dedicated Delta Chat identity\n" +
		"(name + photo) and write to you as that contact — no extra setup.\n\n" +
		"/bots — list\n/unpair-bot — stop"
	_, err = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, text))
	return err
}

func (b *Bot) cmdUnpairBot(msg *tgbotapi.Message) error {
	if b.persona == nil {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Persona mode is disabled."))
		return err
	}
	fields := strings.Fields(msg.Text)
	ref := ""
	if len(fields) >= 2 {
		ref = fields[1]
	}
	n, err := b.persona.Unregister(msg.From.ID, ref)
	if err != nil {
		return err
	}
	text := "No active persona bots to disable."
	if n > 0 {
		text = fmt.Sprintf("Disabled %d persona bot(s). Ghost accounts are kept for reuse.", n)
	}
	_, err = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, text))
	return err
}

func (b *Bot) cmdBots(msg *tgbotapi.Message) error {
	if b.persona == nil {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Persona mode is disabled."))
		return err
	}
	text, err := b.persona.ListBots(msg.From.ID)
	if err != nil {
		return err
	}
	_, err = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, text))
	return err
}

func (b *Bot) cmdStatus(msg *tgbotapi.Message) error {
	info, err := b.bridge.PeerInfoForTelegram(msg.From.ID)
	if err != nil {
		return err
	}
	text := "Not paired. Send /start to begin."
	if info != nil {
		text = "Paired ✅\n\n"
		if info.DCAddress != "" {
			text += "Your Delta Chat (other side):\n"
			if info.DCDisplayName != "" && !strings.EqualFold(info.DCDisplayName, info.DCAddress) {
				text += info.DCDisplayName + "\n"
			}
			text += info.DCAddress + "\n\n"
		} else {
			text += "Delta Chat peer email: (unknown)\n\n"
		}
		if info.BotAddress != "" {
			text += "Bridge bot: " + info.BotAddress + "\n"
		}
		text += "Send media here to bridge both ways."
	}
	_, err = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, text))
	return err
}

func (b *Bot) handleSticker(msg *tgbotapi.Message) error {
	st := msg.Sticker
	// Detect video stickers via file path when library has no IsVideo field.
	isVideo := false
	if fileMeta, err := b.api.GetFile(tgbotapi.FileConfig{FileID: st.FileID}); err == nil {
		isVideo = strings.HasSuffix(strings.ToLower(fileMeta.FilePath), ".webm")
	}
	return b.bridgeStickerFile(msg.From.ID, st.FileID, st.IsAnimated, isVideo)
}

func (b *Bot) handleText(msg *tgbotapi.Message) error {
	return b.bridge.ForwardTextToDelta(msg.From.ID, msg.Text)
}

func (b *Bot) handlePhoto(msg *tgbotapi.Message) error {
	if !b.bridge.AllowsKind(bridge.KindImage) {
		return fmt.Errorf("image bridge disabled in config")
	}
	// Largest size is last in Photo[]
	ph := msg.Photo[len(msg.Photo)-1]
	if err := b.bridge.CheckImageLimits(int64(ph.FileSize), "Telegram photo"); err != nil {
		return err
	}
	path, err := b.downloadFile(ph.FileID, "image.jpg")
	if err != nil {
		return err
	}
	defer os.Remove(path)
	if err := safemedia.ValidateFile(path, safemedia.RoleImage, b.cfg.Bridge.Limits.ImageMaxBytes); err != nil {
		return err
	}

	// Caption is part of the same Delta Chat message as the image (not a separate bubble).
	return b.bridge.ForwardToDelta(msg.From.ID, bridge.Media{
		Kind:     bridge.KindImage,
		Path:     path,
		Filename: "image.jpg",
		Caption:  mediaCaption(msg),
		Viewtype: bridge.ViewImage(),
	})
}

// mediaCaption returns Telegram caption text when text bridge is enabled.
func mediaCaption(msg *tgbotapi.Message) string {
	if msg == nil {
		return ""
	}
	return strings.TrimSpace(msg.Caption)
}

func (b *Bot) handleVideo(msg *tgbotapi.Message) error {
	if !b.bridge.AllowsKind(bridge.KindVideo) {
		return fmt.Errorf("video bridge disabled in config")
	}
	v := msg.Video
	name := "video.mp4"
	if v.FileName != "" {
		// keep only extension for anonymity of basename
		if ext := filepath.Ext(v.FileName); ext != "" {
			name = "video" + strings.ToLower(ext)
		}
	}
	label := fmt.Sprintf("Telegram video (%s, %ds)", name, v.Duration)
	if err := b.bridge.CheckVideoLimits(v.Duration, int64(v.FileSize), label); err != nil {
		return err
	}
	path, err := b.downloadFile(v.FileID, name)
	if err != nil {
		return err
	}
	defer os.Remove(path)
	if err := safemedia.ValidateFile(path, safemedia.RoleVideo, b.cfg.Bridge.Limits.VideoMaxBytes); err != nil {
		return err
	}

	if st, err := os.Stat(path); err == nil {
		if err := b.bridge.CheckVideoLimits(v.Duration, st.Size(), label); err != nil {
			return err
		}
	}

	return b.bridge.ForwardToDelta(msg.From.ID, bridge.Media{
		Kind:     bridge.KindVideo,
		Path:     path,
		Filename: name,
		Caption:  mediaCaption(msg),
		Viewtype: bridge.ViewVideo(),
	})
}

func (b *Bot) handleAnimation(msg *tgbotapi.Message) error {
	if !b.bridge.AllowsKind(bridge.KindGif) {
		return fmt.Errorf("gif bridge disabled in config")
	}
	an := msg.Animation
	// Neutral local name only — never keep Telegram original filename in DC.
	name := "video.mp4"
	if strings.HasSuffix(strings.ToLower(an.FileName), ".gif") {
		name = "image.gif"
	}
	path, err := b.downloadFile(an.FileID, name)
	if err != nil {
		return err
	}
	defer os.Remove(path)
	if err := safemedia.ValidateFile(path, safemedia.RoleGIF, b.cfg.Bridge.Limits.FileMaxBytes); err != nil {
		return err
	}

	view := bridge.ViewVideo()
	if strings.HasSuffix(name, ".gif") {
		view = bridge.ViewGif()
	}

	return b.bridge.ForwardToDelta(msg.From.ID, bridge.Media{
		Kind:     bridge.KindGif,
		Path:     path,
		Filename: name,
		Caption:  mediaCaption(msg),
		Viewtype: view,
	})
}

func (b *Bot) handleGifDocument(msg *tgbotapi.Message) error {
	if !b.bridge.AllowsKind(bridge.KindGif) {
		return fmt.Errorf("gif bridge disabled in config")
	}
	path, err := b.downloadFile(msg.Document.FileID, "image.gif")
	if err != nil {
		return err
	}
	defer os.Remove(path)
	if err := safemedia.ValidateFile(path, safemedia.RoleGIF, b.cfg.Bridge.Limits.FileMaxBytes); err != nil {
		return err
	}

	return b.bridge.ForwardToDelta(msg.From.ID, bridge.Media{
		Kind:     bridge.KindGif,
		Path:     path,
		Filename: "image.gif",
		Caption:  mediaCaption(msg),
		Viewtype: bridge.ViewGif(),
	})
}

func isGifDocument(doc *tgbotapi.Document) bool {
	if doc == nil {
		return false
	}
	if strings.EqualFold(doc.MimeType, "image/gif") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(doc.FileName), ".gif")
}

func (b *Bot) downloadFile(fileID, preferredName string) (string, error) {
	f, err := b.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", err
	}
	url := f.Link(b.api.Token)

	ext := filepath.Ext(preferredName)
	if ext == "" {
		ext = filepath.Ext(f.FilePath)
	}
	if ext == "" {
		ext = ".bin"
	}
	base := strings.TrimSuffix(filepath.Base(preferredName), filepath.Ext(preferredName))
	base = sanitizeBase(base)
	out := filepath.Join(b.tmpdir, fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), base, ext))

	client := b.http
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Get(url) //nolint:gosec // Telegram CDN URL from official API
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download: HTTP %s", resp.Status)
	}

	w, err := os.Create(out)
	if err != nil {
		return "", err
	}
	max := b.cfg.Bridge.Limits.FileMaxBytes
	if max <= 0 {
		max = safemedia.DefaultMaxBytes
	}
	_, copyErr := safemedia.CopyLimited(w, resp.Body, max)
	_ = w.Close()
	if copyErr != nil {
		_ = os.Remove(out)
		return "", copyErr
	}
	return out, nil
}

func sanitizeBase(name string) string {
	name = filepath.Base(name)
	if name == "" || name == "." {
		return "file"
	}
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "file"
	}
	return b.String()
}
