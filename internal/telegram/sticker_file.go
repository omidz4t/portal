package telegram

import (
	"fmt"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/omidz4t/portal/internal/bridge"
	"github.com/omidz4t/portal/internal/safemedia"
)

// bridgeStickerFile downloads a sticker-like file by FileID and forwards to Delta Chat.
//
// Animated Telegram stickers (.tgs) are gzip-compressed Lottie JSON. ArcaneChat
// renders them when the message is Viewtype Sticker and the attachment is still
// a valid .tgs (see context/android-main LottieDecoder: GZIPInputStream + Lottie).
// Sending as File or converting to GIF makes them open as documents / static GIFs.
func (b *Bot) bridgeStickerFile(tgUserID int64, fileID string, isAnimated, isVideo bool) error {
	fileMeta, err := b.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return err
	}
	remote := strings.ToLower(fileMeta.FilePath)

	var kind bridge.Kind
	var name string
	media := bridge.Media{}
	switch {
	case isAnimated || strings.HasSuffix(remote, ".tgs"):
		kind = bridge.KindLottie
		// Keep Telegram's .tgs extension and Sticker viewtype for ArcaneChat.
		name = "sticker.tgs"
		media.Viewtype = bridge.ViewSticker()
	case isVideo || strings.HasSuffix(remote, ".webm"):
		kind = bridge.KindVideoSticker
		name = "video.webm"
		media.Viewtype = bridge.ViewVideo()
	default:
		kind = bridge.KindSticker
		name = "image.webp"
		media.Viewtype = bridge.ViewSticker()
	}

	// Custom emoji may be enabled while a specific sticker subtype is off.
	if !b.bridge.AllowsKind(kind) && b.bridge.AllowsKind(bridge.KindCustomEmoji) {
		kind = bridge.KindCustomEmoji
	}
	if !b.bridge.AllowsKind(kind) {
		return fmt.Errorf("%s bridge disabled in config", kind)
	}

	media.Kind = kind
	media.Filename = name

	path, err := b.downloadFile(fileID, name)
	if err != nil {
		return err
	}
	defer os.Remove(path)
	if err := safemedia.ValidateFile(path, safemedia.RoleFromKind(string(kind)), b.cfg.Bridge.Limits.FileMaxBytes); err != nil {
		return err
	}

	// Native TGS for ArcaneChat — never run host converters (ffmpeg/lottie).
	media.Path = path
	return b.bridge.ForwardToDelta(tgUserID, media)
}
