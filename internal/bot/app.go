package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
	"github.com/deltachat-bot/deltabot-cli-go/v2/botcli"
	"github.com/spf13/cobra"

	"github.com/omidz4t/portal/internal/bridge"
	"github.com/omidz4t/portal/internal/config"
	"github.com/omidz4t/portal/internal/dc"
	"github.com/omidz4t/portal/internal/erasure"
	"github.com/omidz4t/portal/internal/persona"
	"github.com/omidz4t/portal/internal/ratelimit"
	"github.com/omidz4t/portal/internal/safemedia"
	"github.com/omidz4t/portal/internal/store"
	"github.com/omidz4t/portal/internal/telegram"
	"github.com/omidz4t/portal/internal/usermsg"
)

// Run starts the Portal deltabot CLI (init, serve, etc.).
func Run() error {
	cli := botcli.New("portal")

	setDefaultDataDir(cli, config.DefaultFolder)

	var (
		configPath string
		cfg        config.Config
		st         *store.Store
		sess       *dc.Session
		br         *bridge.Bridge
		pm         *persona.Manager
		phooks     *persona.PortalHooks
		limits     *ratelimit.Set
		erase      *erasure.Service
	)

	cli.RootCmd.PersistentFlags().StringVarP(
		&configPath,
		"config",
		"c",
		"config.yml",
		"path to YAML config file",
	)
	cli.RootCmd.AddCommand(completionCmd())
	cli.RootCmd.Long = cli.RootCmd.Short + `

Bot commands (Telegram or paired Delta Chat, private 1:1):
  /pair                     invite + pairing code
  /disconnect               unlink without deleting stored data
  /delete_my_data           request erasure of your Portal data
  /delete_my_data_approve   confirm erasure (within 10 minutes)
  /help                     list bot commands`

	cli.RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(configPath)
		if err != nil {
			return err
		}

		// Application logging: default off; config log: stderr | path + log_level.
		log, syncLog, err := config.NewLogger(cfg)
		if err != nil {
			return err
		}
		cli.Logger = log
		// Sync on process exit is best-effort; cobra may not call a global defer.
		_ = syncLog
		if cfg.Log.Enabled() {
			cli.Logger.Infof("logging enabled: output=%s level=%s", cfg.Log.String(), cfg.LogLevel)
		}

		flags := cmd.Root().PersistentFlags()
		if !flags.Changed("folder") {
			cli.AppDir = cfg.Folder
		}
		if !flags.Changed("account") && cfg.Account != 0 {
			cli.SelectedAccount = cfg.Account
		}
		return nil
	}

	cli.OnBotInit(func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		dbPath := cfg.DatabasePath
		if dbPath == "" {
			dbPath = filepath.Join(cfg.Folder, "tgportal.db")
		} else if !filepath.IsAbs(dbPath) {
			dbPath = filepath.Join(cfg.Folder, dbPath)
		}
		var err error
		key, err := store.ParseKey(cfg.DatabaseKey)
		if err != nil {
			cli.Logger.Errorf("database key: %v", err)
			return
		}
		st, err = store.OpenOpts(dbPath, store.Options{
			PendingTTL:      time.Duration(cfg.Pairing.PendingTTLSec) * time.Second,
			CodeLength:      cfg.Pairing.CodeLength,
			Key:             key,
			EncryptRequired: cfg.DatabaseEncrypt,
		})
		if err != nil {
			cli.Logger.Errorf("sqlite open: %v", err)
			return
		}
		cli.Logger.Infof("sqlite ready at %s", dbPath)

		sess = dc.NewSession(bot)
		br = bridge.New(cli.Logger, sess, cfg, st)
		if cfg.PersonaEnabled() {
			pm = persona.New(cli.Logger, cfg, sess, st)
			phooks = &persona.PortalHooks{M: pm, DC: sess, St: st}
		}
		limits = ratelimit.Defaults()
		erase = erasure.New(cli.Logger, st, sess, pm)
		registerHandlers(cli, sess, cfg, st, br, pm, limits, erase)
	})

	cli.OnBotStart(func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		cli.Logger.Infof("bot started; data dir=%q mode=%s", cli.AppDir, cfg.Mode)
		if sess == nil {
			sess = dc.NewSession(bot)
		}
		if br == nil {
			br = bridge.New(cli.Logger, sess, cfg, st)
		}
		if pm == nil && cfg.PersonaEnabled() && st != nil {
			pm = persona.New(cli.Logger, cfg, sess, st)
			phooks = &persona.PortalHooks{M: pm, DC: sess, St: st}
		}
		if erase == nil && st != nil {
			erase = erasure.New(cli.Logger, st, sess, pm)
		}
		ApplyProfile(cli, sess, cfg)
		if err := sess.ApplyShortDeviceRetention("60"); err != nil {
			cli.Logger.Errorf("device retention: %v", err)
		}

		if dcp := cfg.DeltachatProxy(); dcp.IsEnabled() {
			url := dcp.ResolvedURL()
			if err := sess.ApplyProxy(url, true); err != nil {
				cli.Logger.Errorf("deltachat proxy: %v", err)
			} else {
				cli.Logger.Infof("deltachat proxy enabled")
			}
		} else if cfg.Proxy.Enabled != nil && !*cfg.Proxy.Enabled {
			_ = sess.ApplyProxy("", false)
		}

		if cfg.InviteURL != "" && cfg.BootMessage != "" {
			NotifyInviteContact(cli, sess, cfg.InviteURL, cfg.BootMessage)
		}

		if st == nil {
			cli.Logger.Error("sqlite store not initialized; Telegram bridge disabled")
			return
		}

		var hooks telegram.PersonaHooks
		if phooks != nil {
			hooks = phooks
		}
		if limits == nil {
			limits = ratelimit.Defaults()
		}
		tgBot, err := telegram.Start(cli.Logger, cfg, br, st, hooks, limits, erase)
		if err != nil {
			cli.Logger.Errorf("telegram bridge: %v", err)
			return
		}
		if tgBot != nil {
			br.SetTelegramOut(tgBot)
			if cfg.PersonalEnabled() {
				cli.Logger.Info("bidirectional bridge ready (Telegram ↔ Delta Chat)")
			}
		}

		if pm != nil {
			if err := pm.Start(pm.StartPoller); err != nil {
				cli.Logger.Errorf("persona manager: %v", err)
			} else {
				cli.Logger.Info("persona mode ready (/pair-bot on portal bot)")
			}
		}
	})

	return cli.Start()
}

func setDefaultDataDir(cli *botcli.BotCli, dir string) {
	cli.AppDir = dir
	if f := cli.RootCmd.PersistentFlags().Lookup("folder"); f != nil {
		f.DefValue = dir
	}
}

func registerHandlers(cli *botcli.BotCli, sess *dc.Session, cfg config.Config, st *store.Store, br *bridge.Bridge, pm *persona.Manager, limits *ratelimit.Set, erase *erasure.Service) {
	bot := sess.Bot

	pool := newDCWorkPool(dcWorkers, dcQueue, func(accId, msgId uint32) {
		if err := handleIncomingDC(cli, sess, st, br, cfg, pm, limits, erase, accId, msgId); err != nil {
			cli.GetLogger(accId).Errorf("dc handler: %v", err)
		}
	})
	bot.OnNewMsg(func(bot *deltachat.Bot, accId uint32, msgId uint32) {
		// Non-blocking: a full queue must not stall Bot.Run.
		if !pool.tryEnqueue(accId, msgId) {
			cli.GetLogger(accId).Warnf("dc handler queue full; drop acc=%d msg=%d", accId, msgId)
		}
	})
}

func handleIncomingDC(cli *botcli.BotCli, sess *dc.Session, st *store.Store, br *bridge.Bridge, cfg config.Config, pm *persona.Manager, limits *ratelimit.Set, erase *erasure.Service, accId, msgId uint32) error {
	msg, err := sess.GetMessage(accId, msgId)
	if err != nil {
		return err
	}
	// Ignore self/system — prevents loops when we forward TG→DC as the bot.
	if msg.FromId <= deltachat.ContactLastSpecial {
		return nil
	}
	if msg.IsInfo {
		return nil
	}

	text := strings.TrimSpace(msg.Text)

	// Ghost account traffic (persona mode): owner replied to a TG person.
	if pm != nil {
		if ghost, _ := st.GetGhostByDCAccount(accId); ghost != nil {
			return handleGhostDCMessage(cli, sess, st, br, cfg, pm, accId, msg, text)
		}
	}

	// Slash commands on Delta Chat (both paired and unpaired).
	if text != "" && strings.HasPrefix(text, "/") {
		return handleDCCommand(cli, sess, st, br, cfg, pm, erase, accId, msg, text)
	}

	// Pairing codes only in unpaired 1:1 chats (groups leak codes; active chats
	// must not be hijacked by a pasted/forwarded pending code).
	if text != "" && store.LooksLikeCode(text) {
		active, err := st.GetActiveByDCChat(accId, msg.ChatId)
		if err != nil {
			return err
		}
		pending, perr := st.GetPendingByCode(text)
		if perr != nil {
			return perr
		}
		if shouldAttemptPairing(active != nil, pending != nil) {
			if ok, notice := dcPairingAllowed(sess, accId, msg.ChatId); !ok {
				_ = sess.SendTextWithRetry(accId, msg.ChatId, notice, 5)
				return nil
			}
			if limits != nil && limits.ClaimDC != nil {
				key := fmt.Sprintf("claim:dc:%d:%d", accId, msg.ChatId)
				if !limits.ClaimDC.Allow(key) {
					_ = sess.SendTextWithRetry(accId, msg.ChatId,
						"Too many pairing attempts. Wait a few minutes and try again.", 5)
					return nil
				}
			}
			if err := handlePairingCode(cli, sess, st, br, accId, msg.ChatId, text); err != nil {
				cli.GetLogger(accId).Warnf("pairing: %v", err)
				_ = sess.SendTextWithRetry(accId, msg.ChatId, usermsg.PairingFailed, 10)
			}
			return nil
		}
	}

	pair, err := st.GetActiveByDCChat(accId, msg.ChatId)
	if err != nil {
		return err
	}
	if pair == nil {
		// Unpaired: issue (or reuse) a pairing code + Telegram deep link.
		if !cfg.PersonalEnabled() {
			return nil
		}
		if ok, notice := dcPairingAllowed(sess, accId, msg.ChatId); !ok {
			return sess.SendTextWithRetry(accId, msg.ChatId, notice, 5)
		}
		return offerDCPairing(cli, sess, st, cfg, accId, msg.ChatId)
	}

	// Paired chat: bridge media and/or text to Telegram (no spam acks).
	if hasFile(msg) {
		if err := bridgeDCFileToTelegram(cli, sess, br, cfg, accId, msg); err != nil {
			cli.GetLogger(accId).Warnf("dc→tg media: %v", err)
			_ = sess.SendTextWithRetry(accId, msg.ChatId, usermsg.BridgeFailed, 5)
			return err
		}
		return nil
	}

	if text != "" {
		if err := br.ForwardTextToTelegram(accId, msg.ChatId, text); err != nil {
			return err
		}
	}
	return nil
}

func handleGhostDCMessage(cli *botcli.BotCli, sess *dc.Session, st *store.Store, br *bridge.Bridge, cfg config.Config, pm *persona.Manager, accId uint32, msg deltachat.Message, text string) error {
	// Download file if present
	var path, name string
	hasF := hasFile(msg)
	if hasF {
		cacheDir := filepath.Join(cfg.Folder, "dc-cache")
		_ = os.MkdirAll(cacheDir, 0o755)
		name = "file.bin"
		if msg.FileName != nil && *msg.FileName != "" {
			ext := filepath.Ext(*msg.FileName)
			if ext != "" {
				name = "file" + strings.ToLower(ext)
			}
		}
		path = filepath.Join(cacheDir, fmt.Sprintf("ghost_%d_%d_%s", accId, msg.Id, name))
		if err := sess.SaveMsgFile(accId, msg.Id, path); err != nil {
			// try full download
			_ = sess.DownloadFullMessage(accId, msg.Id)
			if err2 := sess.SaveMsgFile(accId, msg.Id, path); err2 != nil {
				cli.GetLogger(accId).Warnf("ghost save file: %v", err2)
				hasF = false
				path = ""
			}
		}
		if path != "" {
			defer os.Remove(path)
		}
	}

	handled, err := pm.HandleDCMessage(accId, msg.ChatId, msg.FromId, text, hasF && path != "", path, name, pm.SendFromGhost)
	if err != nil {
		return err
	}
	if !handled {
		return nil
	}
	_ = br
	_ = st
	return nil
}

func handleDCCommand(cli *botcli.BotCli, sess *dc.Session, st *store.Store, br *bridge.Bridge, cfg config.Config, _ *persona.Manager, erase *erasure.Service, accId uint32, msg deltachat.Message, text string) error {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return nil
	}
	cmd := strings.ToLower(fields[0])
	if i := strings.Index(cmd, "@"); i > 0 {
		cmd = cmd[:i]
	}
	chatID := msg.ChatId

	switch cmd {
	case "/start", "/help":
		return sess.SendTextWithRetry(accId, chatID, dcHelpText(cfg), 10)
	case "/pair", "/connect":
		return dcCmdPair(cli, sess, st, br, cfg, accId, chatID)
	case "/disconnect":
		return dcCmdDisconnect(sess, st, accId, chatID)
	case "/delete_my_data":
		return dcCmdDeleteMyData(sess, erase, accId, chatID)
	case "/delete_my_data_approve":
		return dcCmdDeleteMyDataApprove(sess, erase, accId, chatID)
	case "/status":
		return dcCmdStatus(sess, st, br, accId, chatID)
	default:
		return sess.SendTextWithRetry(accId, chatID,
			"Unknown command. Send /help for the list.", 5)
	}
}

func dcHelpText(cfg config.Config) string {
	tg := telegramDeepLinkBase(cfg)
	return "Delta ↔️ TG\n\n" +
		"/pair — get a Telegram link to connect this chat\n" +
		"/disconnect — unlink this chat from Telegram\n" +
		"/delete_my_data — request deletion of your Portal data\n" +
		"/delete_my_data_approve — confirm deletion (two-step)\n" +
		"/status — pairing status\n" +
		"/help — this list\n\n" +
		"After pairing, send stickers, images, GIFs, text, or short videos either way.\n" +
		"Telegram: " + tg
}

func dcCmdPair(cli *botcli.BotCli, sess *dc.Session, st *store.Store, _ *bridge.Bridge, cfg config.Config, accId, chatID uint32) error {
	if ok, notice := dcPairingAllowed(sess, accId, chatID); !ok {
		return sess.SendTextWithRetry(accId, chatID, notice, 5)
	}
	if pair, _ := st.GetActiveByDCChat(accId, chatID); pair != nil {
		return sess.SendTextWithRetry(accId, chatID,
			"Already paired ✅\nTelegram user id: "+formatInt64(pair.TelegramUserID)+"\n\n/disconnect to unlink.", 10)
	}
	return offerDCPairing(cli, sess, st, cfg, accId, chatID)
}

// offerDCPairing creates (or reuses) a DC-initiated code and sends t.me/bot?start=CODE.
func offerDCPairing(cli *botcli.BotCli, sess *dc.Session, st *store.Store, cfg config.Config, accId, chatID uint32) error {
	pair, err := st.GetPendingByDCChat(accId, chatID)
	if err != nil {
		return err
	}
	if pair == nil {
		pair, err = st.CreatePendingFromDC(accId, chatID)
		if err != nil {
			return err
		}
		cli.GetLogger(accId).Infof("issued DC pairing code for chat %d", chatID)
	}

	startURL := telegramStartURL(cfg, pair.Code)
	msg := "Delta ↔️ TG — not paired yet\n\n" +
		"Open Telegram and tap this link to connect this chat:\n" +
		startURL + "\n\n" +
		"Or open the bot and send:\n" +
		"/start " + pair.Code + "\n\n" +
		"Pairing code: " + pair.Code + "\n" +
		"(Also works if you /pair on Telegram first, then paste a code here.)"
	return sess.SendTextWithRetry(accId, chatID, msg, 10)
}

func telegramDeepLinkBase(cfg config.Config) string {
	u := strings.TrimSpace(cfg.Telegram.BotURL)
	if u == "" {
		u = "https://t.me/tgdeltabridgebot"
	}
	return strings.TrimRight(u, "/")
}

func telegramStartURL(cfg config.Config, code string) string {
	return telegramDeepLinkBase(cfg) + "?start=" + strings.ToUpper(strings.TrimSpace(code))
}

func dcCmdDisconnect(sess *dc.Session, st *store.Store, accId, chatID uint32) error {
	n, err := st.DisconnectByDCChat(accId, chatID)
	if err != nil {
		return err
	}
	if n == 0 {
		return sess.SendTextWithRetry(accId, chatID,
			"Not paired — nothing to disconnect.\nUse Telegram /pair then send the code here.", 10)
	}
	return sess.SendTextWithRetry(accId, chatID,
		"Disconnected ✅\nThis Delta Chat is no longer linked to Telegram.\n\n/pair for instructions.", 10)
}

func dcCmdDeleteMyData(sess *dc.Session, erase *erasure.Service, accId, chatID uint32) error {
	if ok, _ := dcPairingAllowed(sess, accId, chatID); !ok {
		return sess.SendTextWithRetry(accId, chatID, erasure.PrivateOnly, 5)
	}
	if erase == nil {
		return sess.SendTextWithRetry(accId, chatID, usermsg.Generic, 5)
	}
	erase.Request(erasure.DCKey(accId, chatID))
	return sess.SendTextWithRetry(accId, chatID, erasure.WarnText, 10)
}

func dcCmdDeleteMyDataApprove(sess *dc.Session, erase *erasure.Service, accId, chatID uint32) error {
	if ok, _ := dcPairingAllowed(sess, accId, chatID); !ok {
		return sess.SendTextWithRetry(accId, chatID, erasure.PrivateOnly, 5)
	}
	if erase == nil || !erase.Consume(erasure.DCKey(accId, chatID)) {
		return sess.SendTextWithRetry(accId, chatID, erasure.NeedRequestText, 10)
	}
	if err := sess.SendTextWithRetry(accId, chatID, erasure.DoneText, 10); err != nil {
		return err
	}
	if err := erase.PurgeFromDCChat(accId, chatID); err != nil {
		return sess.SendTextWithRetry(accId, chatID, usermsg.Generic, 5)
	}
	return nil
}

func dcCmdStatus(sess *dc.Session, _ *store.Store, br *bridge.Bridge, accId, chatID uint32) error {
	info, err := br.PeerInfoForDCChat(accId, chatID)
	if err != nil {
		return err
	}
	if info == nil {
		return sess.SendTextWithRetry(accId, chatID, "Not paired. /pair for instructions.", 10)
	}
	text := "Paired ✅\n\n"
	if info.DCAddress != "" {
		text += "Your Delta Chat email:\n" + info.DCAddress + "\n\n"
	}
	text += "Telegram (other side):\n" + formatInt64(info.TelegramUserID)
	if info.TelegramUsername != "" {
		text += "\n@" + info.TelegramUsername
	}
	text += "\n\nBridge is bidirectional."
	return sess.SendTextWithRetry(accId, chatID, text, 10)
}

func formatInt64(n int64) string {
	return fmt.Sprintf("%d", n)
}

func hasFile(msg deltachat.Message) bool {
	if msg.File != nil && *msg.File != "" {
		return true
	}
	switch msg.ViewType {
	case deltachat.ViewtypeImage, deltachat.ViewtypeGif, deltachat.ViewtypeSticker,
		deltachat.ViewtypeVideo, deltachat.ViewtypeFile, deltachat.ViewtypeAudio,
		deltachat.ViewtypeVoice:
		// Stickers from ArcaneChat always have a file attachment once delivered;
		// treat Sticker viewtype as media even when File path is not yet filled.
		if msg.ViewType == deltachat.ViewtypeSticker {
			return true
		}
		return msg.FileBytes > 0 || msg.DownloadState == deltachat.DownloadStateAvailable ||
			msg.DownloadState == deltachat.DownloadStateDone
	default:
		return false
	}
}

func bridgeDCFileToTelegram(cli *botcli.BotCli, sess *dc.Session, br *bridge.Bridge, cfg config.Config, accId uint32, msg deltachat.Message) error {
	// Ensure full body is present for partially downloaded messages.
	if msg.DownloadState == deltachat.DownloadStateAvailable || msg.DownloadState == deltachat.DownloadStateInProgress {
		_ = sess.DownloadFullMessage(accId, msg.Id)
		// re-fetch after download
		time.Sleep(300 * time.Millisecond)
		m2, err := sess.GetMessage(accId, msg.Id)
		if err == nil {
			msg = m2
		}
	}

	filename := "file.bin"
	if msg.FileName != nil && *msg.FileName != "" {
		filename = filepath.Base(*msg.FileName)
	} else if msg.File != nil && *msg.File != "" {
		filename = filepath.Base(*msg.File)
	}
	// ArcaneChat sticker picker often sets fileName=null; seed a typed name from viewtype.
	if filepath.Ext(filename) == "" || strings.EqualFold(filename, "file.bin") {
		switch msg.ViewType {
		case deltachat.ViewtypeSticker:
			filename = "sticker.webp" // may be rewritten after sniff
		case deltachat.ViewtypeGif:
			filename = "image.gif"
		case deltachat.ViewtypeImage:
			filename = "image.jpg"
		case deltachat.ViewtypeVideo:
			filename = "video.mp4"
		}
	}
	if msg.FileMime != nil {
		switch strings.ToLower(*msg.FileMime) {
		case "image/webp":
			filename = "sticker.webp"
		case "application/x-tgsticker", "application/gzip":
			// Telegram Lottie stickers travel as gzip; keep .tgs for Bot API.
			filename = "sticker.tgs"
		case "video/webm":
			filename = "video.webm"
		case "image/gif":
			filename = "image.gif"
		case "image/png":
			if msg.ViewType == deltachat.ViewtypeSticker {
				filename = "sticker.png"
			} else {
				filename = "image.png"
			}
		case "image/jpeg", "image/jpg":
			filename = "image.jpg"
		}
	}

	cacheDir := filepath.Join(cfg.Folder, "dc-cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}
	outPath := filepath.Join(cacheDir, fmt.Sprintf("%d_%d_%s", accId, msg.Id, sanitizeName(filename)))

	// Prefer SaveMsgFile for a stable copy; fall back to msg.File path.
	if err := sess.SaveMsgFile(accId, msg.Id, outPath); err != nil {
		if msg.File != nil && *msg.File != "" {
			outPath = *msg.File
		} else {
			return fmt.Errorf("save file: %w", err)
		}
	}
	defer func() {
		if strings.Contains(outPath, "dc-cache") {
			_ = os.Remove(outPath)
		}
		if msg.File != nil {
			dc.RemoveEphemeralFile(*msg.File)
		}
	}()

	if _, err := os.Stat(outPath); err != nil {
		return err
	}

	// Re-classify using path + filename after save (blob may carry real extension).
	kindName := filename
	if filepath.Ext(kindName) == "" || strings.EqualFold(kindName, "file.bin") {
		kindName = filepath.Base(outPath)
	}
	kind := bridge.KindFromDCViewtype(msg.ViewType, kindName)
	// If still generic but DC said sticker, force sticker kind for outbound sniffing.
	if kind == "" && msg.ViewType == deltachat.ViewtypeSticker {
		kind = bridge.KindSticker
	}
	if err := safemedia.ValidateFile(outPath, safemedia.RoleFromKind(string(kind)), cfg.Bridge.Limits.FileMaxBytes); err != nil {
		return fmt.Errorf("unsafe media: %w", err)
	}
	// Enforce short-video limits on DC → TG as well.
	if st, err := os.Stat(outPath); err == nil {
		what := string(kind)
		if what == "" {
			what = "file"
		}
		if filename != "" {
			what = fmt.Sprintf("%s (%s)", what, filename)
		}
		if msg.Duration > 0 {
			what = fmt.Sprintf("%s, %ds", what, msg.Duration)
		}
		what = "Delta Chat " + what
		switch kind {
		case bridge.KindVideo, bridge.KindVideoSticker, bridge.KindGif:
			if err := br.CheckVideoLimits(int(msg.Duration), st.Size(), what); err != nil {
				return err
			}
		case bridge.KindImage:
			if err := br.CheckImageLimits(st.Size(), what); err != nil {
				return err
			}
		}
	}
	// DC stores image/video captions as the message text alongside the file.
	caption := strings.TrimSpace(msg.Text)
	return br.ForwardFileToTelegram(accId, msg.ChatId, outPath, filename, caption, kind)
}

func sanitizeName(name string) string {
	name = filepath.Base(name)
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "file.bin"
	}
	return b.String()
}

func handlePairingCode(cli *botcli.BotCli, sess *dc.Session, st *store.Store, br *bridge.Bridge, accId, chatID uint32, code string) error {
	if existing, err := st.GetActiveByDCChat(accId, chatID); err != nil {
		return err
	} else if existing != nil {
		return errAlreadyPaired
	}
	pending, err := st.GetPendingByCode(code)
	if err != nil {
		return err
	}
	if pending == nil {
		return errInvalidCode
	}

	pair, err := st.ActivatePair(code, accId, chatID)
	if err != nil {
		return err
	}

	// Capture owner peer vcard (public key) for persona ghost key-import.
	if vcard, _, err := sess.ExportPeerVcard(accId, chatID); err != nil {
		cli.GetLogger(accId).Warnf("owner vcard capture: %v", err)
	} else if err := st.SetPairOwnerVcard(pair.ID, vcard); err != nil {
		cli.GetLogger(accId).Warnf("store owner vcard: %v", err)
	} else {
		pair.OwnerVcard = vcard
		cli.GetLogger(accId).Infof("captured owner DC vcard for persona (%d bytes)", len(vcard))
	}

	cli.GetLogger(accId).Infof("paired TG user %d (@%s) → chat %d code %s",
		pair.TelegramUserID, pair.TelegramUsername, chatID, pair.Code)

	// Success on both Delta Chat and Telegram.
	if br != nil {
		br.NotifyPairingSuccess(pair)
		return nil
	}
	return sess.SendTextWithRetry(accId, chatID, bridge.PairingSuccessMessage(pair.Code), 20)
}

var (
	errInvalidCode   = errString("invalid or expired pairing code")
	errAlreadyPaired = errString("this Delta Chat is already paired")
)

type errString string

func (e errString) Error() string { return string(e) }
