package bridge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
	"go.uber.org/zap"

	"github.com/omidz4t/portal/internal/config"
	"github.com/omidz4t/portal/internal/dc"
	"github.com/omidz4t/portal/internal/safemedia"
	"github.com/omidz4t/portal/internal/store"
)

// Kind classifies Telegram media for bridge toggles.
type Kind string

const (
	KindSticker      Kind = "sticker"
	KindLottie       Kind = "lottie"
	KindVideoSticker Kind = "video_sticker"
	KindGif          Kind = "gif"
	// KindCustomEmoji is a Telegram custom/premium emoji (resolved via getCustomEmojiStickers).
	KindCustomEmoji Kind = "custom_emoji"
	KindImage       Kind = "image"
	KindVideo       Kind = "video"
	KindText        Kind = "text"
)

// Media is a downloaded Telegram asset ready for Delta Chat.
// Filenames are anonymized; Caption is sent as the message text with the file.
type Media struct {
	Kind     Kind
	Path     string
	Filename string
	Caption  string // optional message text shown with the attachment (Telegram caption)
	Viewtype *deltachat.Viewtype
}

// TelegramOut sends media/text from Delta Chat to a Telegram chat.
type TelegramOut interface {
	SendText(tgChatID int64, text string) error
	SendMedia(tgChatID int64, path, filename string, kind Kind) error
	SendMediaCaption(tgChatID int64, path, filename, caption string, kind Kind) error
}

// Bridge forwards media both ways between paired Telegram and Delta Chat chats.
type Bridge struct {
	log   *zap.SugaredLogger
	dc    *dc.Session
	cfg   config.Config
	store *store.Store
	tg    TelegramOut
}

// New creates a bridge (call SetTelegramOut after the Telegram bot starts).
func New(log *zap.SugaredLogger, sess *dc.Session, cfg config.Config, st *store.Store) *Bridge {
	return &Bridge{log: log, dc: sess, cfg: cfg, store: st}
}

// SetTelegramOut registers the Telegram sender for DC → TG direction.
func (b *Bridge) SetTelegramOut(tg TelegramOut) {
	b.tg = tg
}

// Enabled reports whether the Telegram bridge can run.
func (b *Bridge) Enabled() bool {
	if !b.cfg.Telegram.Enabled || b.cfg.TelegramToken == "" || b.store == nil || b.dc == nil {
		return false
	}
	br := b.cfg.Bridge
	return br.Text || br.Images || br.Videos ||
		br.Stickers || br.Lottie || br.VideoStickers || br.Gif || br.CustomEmoji
}

// AllowsKind checks config toggles.
func (b *Bridge) AllowsKind(k Kind) bool {
	switch k {
	case KindText:
		return b.cfg.Bridge.Text
	case KindImage:
		return b.cfg.Bridge.Images
	case KindVideo:
		return b.cfg.Bridge.Videos
	case KindSticker:
		return b.cfg.Bridge.Stickers
	case KindLottie:
		return b.cfg.Bridge.Lottie
	case KindVideoSticker:
		return b.cfg.Bridge.VideoStickers
	case KindGif:
		return b.cfg.Bridge.Gif
	case KindCustomEmoji:
		return b.cfg.Bridge.CustomEmoji ||
			b.cfg.Bridge.Stickers || b.cfg.Bridge.Lottie || b.cfg.Bridge.VideoStickers
	default:
		return false
	}
}

// limitLabel builds a human-readable name for limit errors.
func limitLabel(kind Kind, filename string) string {
	base := string(kind)
	if base == "" {
		base = "file"
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return base
	}
	return fmt.Sprintf("%s (%s)", base, filepath.Base(filename))
}

// CheckVideoLimits returns an error if duration/size exceed bridge.limits for videos.
// durationSec and sizeBytes use 0 when unknown (unknown size is allowed unless duration fails).
// what labels which item failed (e.g. "Telegram video", "animation GIF", "file.mp4").
func (b *Bridge) CheckVideoLimits(durationSec int, sizeBytes int64, what string) error {
	if what == "" {
		what = "video"
	}
	lim := b.cfg.Bridge.Limits
	if lim.VideoMaxDurationSec > 0 && durationSec > lim.VideoMaxDurationSec {
		return fmt.Errorf("%s too long: %ds (max %ds; bridge.limits.video_max_duration_sec)",
			what, durationSec, lim.VideoMaxDurationSec)
	}
	if lim.VideoMaxBytes > 0 && sizeBytes > lim.VideoMaxBytes {
		return fmt.Errorf("%s too large: %d bytes (max %d; bridge.limits.video_max_bytes)",
			what, sizeBytes, lim.VideoMaxBytes)
	}
	return nil
}

// CheckImageLimits returns an error if image size exceeds limit.
// what labels which item failed (e.g. "Telegram photo", "image.jpg").
func (b *Bridge) CheckImageLimits(sizeBytes int64, what string) error {
	if what == "" {
		what = "image"
	}
	lim := b.cfg.Bridge.Limits
	if lim.ImageMaxBytes > 0 && sizeBytes > lim.ImageMaxBytes {
		return fmt.Errorf("%s too large: %d bytes (max %d; bridge.limits.image_max_bytes)",
			what, sizeBytes, lim.ImageMaxBytes)
	}
	return nil
}

// CheckFileLimits returns an error if generic file size exceeds limit.
// what labels which item failed.
func (b *Bridge) CheckFileLimits(sizeBytes int64, what string) error {
	if what == "" {
		what = "file"
	}
	lim := b.cfg.Bridge.Limits
	if lim.FileMaxBytes > 0 && sizeBytes > lim.FileMaxBytes {
		return fmt.Errorf("%s too large: %d bytes (max %d; bridge.limits.file_max_bytes)",
			what, sizeBytes, lim.FileMaxBytes)
	}
	return nil
}

// ForwardToDelta sends media to the Delta Chat chat paired with this Telegram user.
// Safe for concurrent callers: DC Session serializes RPC; store is SQLite.
func (b *Bridge) ForwardToDelta(tgUserID int64, m Media) error {
	if !b.AllowsKind(m.Kind) {
		return fmt.Errorf("bridge disabled for %s", m.Kind)
	}
	if m.Path == "" {
		return fmt.Errorf("empty media path")
	}
	if _, err := os.Stat(m.Path); err != nil {
		return err
	}
	if err := safemedia.ValidateFile(m.Path, safemedia.RoleFromKind(string(m.Kind)), b.cfg.Bridge.Limits.FileMaxBytes); err != nil {
		return fmt.Errorf("unsafe media: %w", err)
	}

	pair, err := b.store.GetActiveByTelegram(tgUserID)
	if err != nil {
		return err
	}
	if pair == nil {
		return fmt.Errorf("not paired — send /start on Telegram and enter the code in Delta Chat")
	}

	name := anonymousFilename(m)
	caption := strings.TrimSpace(m.Caption)
	if err := b.dc.SendFileWithRetry(pair.DCAccountID, pair.DCChatID, m.Path, name, caption, m.Viewtype, 40); err != nil {
		return err
	}
	b.log.Infow("bridged to Delta Chat",
		"kind", m.Kind,
		"tg_user", tgUserID,
		"dc_chat", pair.DCChatID,
		"has_caption", caption != "",
	)
	return nil
}

// BotInviteLink returns the Delta Chat secure-join invite for the bot account.
func (b *Bridge) BotInviteLink() (string, error) {
	accID, err := b.dc.FirstConfiguredAccount()
	if err != nil {
		return "", err
	}
	return b.dc.GetChatSecurejoinQrCode(accID)
}

// PairingSuccessMessage is sent to both Telegram and Delta Chat after pairing.
func PairingSuccessMessage(code string) string {
	if code == "" {
		return "✅ Pairing complete!\n\n" +
			"Telegram and Delta Chat are linked.\n" +
			"Send stickers, text, images, and GIFs either way — Delta ↔️ TG.\n\n" +
			"/status · /disconnect · /help"
	}
	return fmt.Sprintf(
		"✅ Pairing complete!\n\n"+
			"Code: %s\n"+
			"Telegram and Delta Chat are linked.\n"+
			"Send stickers, text, images, and GIFs either way — Delta ↔️ TG.\n\n"+
			"/status · /disconnect · /help",
		code,
	)
}

// NotifyPairingSuccess sends the success notice to both sides of an active pair.
func (b *Bridge) NotifyPairingSuccess(pair *store.Pair) {
	if pair == nil || pair.Status != store.StatusActive {
		return
	}
	// Capture owner peer vcard (public key) for persona mode ghosts.
	if pair.DCAccountID != 0 && pair.DCChatID != 0 && b.store != nil {
		if vcard, _, err := b.dc.ExportPeerVcard(pair.DCAccountID, pair.DCChatID); err != nil {
			b.log.Warnf("owner vcard capture: %v", err)
		} else if err := b.store.SetPairOwnerVcard(pair.ID, vcard); err != nil {
			b.log.Warnf("store owner vcard: %v", err)
		} else {
			pair.OwnerVcard = vcard
			b.log.Infof("captured owner DC vcard (%d bytes) for pair %d", len(vcard), pair.ID)
		}
	}

	msg := PairingSuccessMessage(pair.Code)

	// Delta Chat
	if pair.DCAccountID != 0 && pair.DCChatID != 0 {
		if err := b.dc.SendTextWithRetry(pair.DCAccountID, pair.DCChatID, msg, 15); err != nil {
			b.log.Warnf("pairing success notify DC chat %d: %v", pair.DCChatID, err)
		}
	}

	// Telegram
	if b.tg == nil {
		b.log.Warnf("pairing success notify TG: telegram outbound not ready")
		return
	}
	tgChat := pair.TelegramChatID
	if tgChat == 0 {
		tgChat = pair.TelegramUserID
	}
	if tgChat == 0 {
		return
	}
	if err := b.tg.SendText(tgChat, msg); err != nil {
		b.log.Warnf("pairing success notify TG chat %d: %v", tgChat, err)
	}
}

// PairPeerInfo is identity shown on /status for the other side of a pair.
type PairPeerInfo struct {
	// DCAddress is the human peer's Delta Chat email (paired chat contact).
	DCAddress string
	// DCDisplayName is their display name if set.
	DCDisplayName string
	// BotAddress is this bot's Delta Chat account email.
	BotAddress string
	// TelegramUserID / Username from the pair record.
	TelegramUserID   int64
	TelegramUsername string
}

// PeerInfoForTelegram returns status details for a paired Telegram user
// (Delta Chat peer email on the other side).
func (b *Bridge) PeerInfoForTelegram(tgUserID int64) (*PairPeerInfo, error) {
	pair, err := b.store.GetActiveByTelegram(tgUserID)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, nil
	}
	info := &PairPeerInfo{
		TelegramUserID:   pair.TelegramUserID,
		TelegramUsername: pair.TelegramUsername,
	}
	if botAddr, err := b.dc.AccountAddress(pair.DCAccountID); err == nil {
		info.BotAddress = botAddr
	}
	if addr, name, err := b.dc.PeerInChat(pair.DCAccountID, pair.DCChatID); err == nil {
		info.DCAddress = addr
		info.DCDisplayName = name
	}
	return info, nil
}

// PeerInfoForDCChat returns status details for a paired Delta Chat conversation.
func (b *Bridge) PeerInfoForDCChat(accID, chatID uint32) (*PairPeerInfo, error) {
	pair, err := b.store.GetActiveByDCChat(accID, chatID)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, nil
	}
	info := &PairPeerInfo{
		TelegramUserID:   pair.TelegramUserID,
		TelegramUsername: pair.TelegramUsername,
	}
	if botAddr, err := b.dc.AccountAddress(accID); err == nil {
		info.BotAddress = botAddr
	}
	if addr, name, err := b.dc.PeerInChat(accID, chatID); err == nil {
		info.DCAddress = addr
		info.DCDisplayName = name
	}
	return info, nil
}

// ForwardTextToDelta sends plain text from a paired Telegram user to Delta Chat.
func (b *Bridge) ForwardTextToDelta(tgUserID int64, text string) error {
	if !b.AllowsKind(KindText) {
		return fmt.Errorf("text bridge disabled in config")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	pair, err := b.store.GetActiveByTelegram(tgUserID)
	if err != nil {
		return err
	}
	if pair == nil {
		return fmt.Errorf("not paired — send /start on Telegram and enter the code in Delta Chat")
	}
	if err := b.dc.SendTextWithRetry(pair.DCAccountID, pair.DCChatID, text, 20); err != nil {
		return err
	}
	b.log.Infow("bridged text to Delta Chat", "tg_user", tgUserID, "dc_chat", pair.DCChatID)
	return nil
}

// ForwardTextToTelegram sends plain text from a paired DC chat to Telegram.
func (b *Bridge) ForwardTextToTelegram(accID, chatID uint32, text string) error {
	if !b.AllowsKind(KindText) {
		return fmt.Errorf("text bridge disabled in config")
	}
	if b.tg == nil {
		return fmt.Errorf("telegram outbound not ready")
	}
	pair, err := b.store.GetActiveByDCChat(accID, chatID)
	if err != nil {
		return err
	}
	if pair == nil {
		return fmt.Errorf("no active pair for this Delta Chat")
	}
	tgChat := pair.TelegramChatID
	if tgChat == 0 {
		tgChat = pair.TelegramUserID // private chats: chat_id == user_id
	}
	if err := b.tg.SendText(tgChat, text); err != nil {
		return err
	}
	b.log.Infow("bridged text to Telegram", "dc_chat", chatID, "tg_chat", tgChat)
	return nil
}

// ForwardFileToTelegram sends a local file (from a DC message) to the paired Telegram chat.
// caption is optional message text (Delta Chat file+text → Telegram media caption).
func (b *Bridge) ForwardFileToTelegram(accID, chatID uint32, path, filename, caption string, kind Kind) error {
	if b.tg == nil {
		return fmt.Errorf("telegram outbound not ready")
	}
	if kind != "" && !b.AllowsKind(kind) {
		// Unknown empty kind still allowed as generic document.
		return fmt.Errorf("bridge disabled for %s", kind)
	}
	if err := safemedia.ValidateFile(path, safemedia.RoleFromKind(string(kind)), b.cfg.Bridge.Limits.FileMaxBytes); err != nil {
		return fmt.Errorf("unsafe media: %w", err)
	}
	// Enforce size limits when we can stat the file.
	if st, err := os.Stat(path); err == nil {
		what := limitLabel(kind, filename)
		switch kind {
		case KindVideo, KindVideoSticker, KindGif:
			if err := b.CheckVideoLimits(0, st.Size(), what); err != nil {
				return err
			}
		case KindImage:
			if err := b.CheckImageLimits(st.Size(), what); err != nil {
				return err
			}
		default:
			if err := b.CheckFileLimits(st.Size(), what); err != nil {
				return err
			}
		}
	}
	pair, err := b.store.GetActiveByDCChat(accID, chatID)
	if err != nil {
		return err
	}
	if pair == nil {
		return fmt.Errorf("no active pair for this Delta Chat")
	}
	tgChat := pair.TelegramChatID
	if tgChat == 0 {
		tgChat = pair.TelegramUserID
	}
	caption = strings.TrimSpace(caption)
	var sendErr error
	if caption != "" {
		sendErr = b.tg.SendMediaCaption(tgChat, path, filename, caption, kind)
	} else {
		sendErr = b.tg.SendMedia(tgChat, path, filename, kind)
	}
	if sendErr != nil {
		return sendErr
	}
	b.log.Infow("bridged file to Telegram", "dc_chat", chatID, "tg_chat", tgChat, "kind", kind, "has_caption", caption != "")
	return nil
}

// KindFromDCViewtype maps Delta Chat viewtype + filename to a bridge Kind.
// Prefer extension when viewtype is Sticker so Lottie/video stickers map correctly.
func KindFromDCViewtype(vt deltachat.Viewtype, filename string) Kind {
	ext := strings.ToLower(filepath.Ext(filename))
	switch vt {
	case deltachat.ViewtypeSticker:
		switch ext {
		case ".tgs":
			return KindLottie
		case ".webm":
			return KindVideoSticker
		default:
			// .webp, .png, missing ext (ArcaneChat often omits filename) → sticker
			return KindSticker
		}
	case deltachat.ViewtypeGif:
		return KindGif
	case deltachat.ViewtypeVideo:
		if ext == ".webm" {
			return KindVideoSticker
		}
		return KindVideo
	case deltachat.ViewtypeImage:
		if ext == ".webp" {
			return KindSticker
		}
		return KindImage
	case deltachat.ViewtypeFile:
		if ext == ".tgs" {
			return KindLottie
		}
		if ext == ".webp" {
			return KindSticker
		}
		if ext == ".gif" || ext == ".mp4" {
			return KindGif
		}
		if ext == ".webm" {
			return KindVideoSticker
		}
		return Kind("") // generic document
	default:
		if ext == ".tgs" {
			return KindLottie
		}
		if ext == ".webp" {
			return KindSticker
		}
		if ext == ".gif" || ext == ".mp4" {
			return KindGif
		}
		return Kind("")
	}
}

// KindFromFilename maps a Telegram document name to a bridge kind (never assume image).
func KindFromFilename(filename string) Kind {
	return KindFromDCViewtype(deltachat.ViewtypeFile, filename)
}

func anonymousFilename(m Media) string {
	ext := filepath.Ext(m.Path)
	if ext == "" && m.Filename != "" {
		ext = filepath.Ext(m.Filename)
	}
	switch m.Kind {
	case KindSticker, KindCustomEmoji:
		if ext == "" {
			ext = ".webp"
		}
		if ext == ".tgs" {
			return "sticker.tgs"
		}
		if ext == ".webm" {
			return "video.webm"
		}
		if ext == ".gif" {
			return "image.gif"
		}
		return "image" + ext
	case KindLottie:
		// ArcaneChat Lottie path: sticker viewtype + .tgs filename (gzip Lottie).
		if ext == "" || ext == ".tgs" {
			return "sticker.tgs"
		}
		if ext == ".gif" {
			return "image.gif"
		}
		return "sticker" + ext
	case KindVideoSticker:
		if ext == "" {
			ext = ".webm"
		}
		return "video" + ext
	case KindGif:
		if ext == "" {
			ext = ".mp4"
		}
		if ext == ".gif" {
			return "image.gif"
		}
		return "video" + ext
	case KindImage:
		if ext == "" {
			ext = ".jpg"
		}
		return "image" + ext
	case KindVideo:
		if ext == "" {
			ext = ".mp4"
		}
		return "video" + ext
	default:
		if ext == "" {
			ext = ".bin"
		}
		return "file" + ext
	}
}

func ViewImage() *deltachat.Viewtype {
	v := deltachat.ViewtypeImage
	return &v
}

func ViewGif() *deltachat.Viewtype {
	v := deltachat.ViewtypeGif
	return &v
}

func ViewSticker() *deltachat.Viewtype {
	v := deltachat.ViewtypeSticker
	return &v
}

func ViewFile() *deltachat.Viewtype {
	v := deltachat.ViewtypeFile
	return &v
}

func ViewVideo() *deltachat.Viewtype {
	v := deltachat.ViewtypeVideo
	return &v
}
