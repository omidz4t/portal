package telegram

import (
	"fmt"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/omidz4t/portal/internal/bridge"
)

// SendText sends a plain text message to a Telegram chat (DC → TG).
func (b *Bot) SendText(tgChatID int64, text string) error {
	if tgChatID == 0 {
		return fmt.Errorf("telegram chat id is 0")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	_, err := b.api.Send(tgbotapi.NewMessage(tgChatID, text))
	return err
}

// SendMedia sends a local file to Telegram (no caption).
func (b *Bot) SendMedia(tgChatID int64, path, filename string, kind bridge.Kind) error {
	return b.SendMediaCaption(tgChatID, path, filename, "", kind)
}

// SendMediaCaption sends a local file to Telegram with optional caption
// (Delta Chat file+text → one Telegram media message).
//
// ArcaneChat often omits filenames on stickers; we sniff magic bytes and rename
// so sendSticker receives .webp / .tgs / .webm as required by Bot API.
func (b *Bot) SendMediaCaption(tgChatID int64, path, filename, caption string, kind bridge.Kind) error {
	if tgChatID == 0 {
		return fmt.Errorf("telegram chat id is 0")
	}
	if path == "" {
		return fmt.Errorf("empty file path")
	}
	if filename == "" {
		filename = filepath.Base(path)
	}
	caption = strings.TrimSpace(caption)
	if len(caption) > telegramCaptionMax {
		caption = truncateCaption(caption, telegramCaptionMax)
	}

	uploadPath, uploadName, cleanup, err := ensureUploadName(path, filename, string(kind))
	if err != nil {
		return fmt.Errorf("prepare upload name: %w", err)
	}
	defer cleanup()

	ext := strings.ToLower(filepath.Ext(uploadName))
	file := tgbotapi.FilePath(uploadPath)

	// Prefer kind, fall back to extension.
	switch kind {
	case bridge.KindSticker, bridge.KindCustomEmoji, bridge.KindLottie:
		// Stickers cannot carry captions in Bot API; send sticker then optional text.
		if err := b.sendAsTelegramSticker(tgChatID, file, ext); err != nil {
			return err
		}
		if caption != "" {
			return b.SendText(tgChatID, caption)
		}
		return nil
	case bridge.KindVideoSticker:
		if ext == ".webm" {
			if _, err := b.api.Send(tgbotapi.NewSticker(tgChatID, file)); err == nil {
				if caption != "" {
					return b.SendText(tgChatID, caption)
				}
				return nil
			}
			b.log.Warnf("sendSticker webm failed, trying video: %v", err)
		}
		v := tgbotapi.NewVideo(tgChatID, file)
		v.Caption = caption
		_, err := b.api.Send(v)
		return err
	case bridge.KindGif:
		if ext == ".gif" || ext == ".mp4" {
			a := tgbotapi.NewAnimation(tgChatID, file)
			a.Caption = caption
			_, err := b.api.Send(a)
			return err
		}
		v := tgbotapi.NewVideo(tgChatID, file)
		v.Caption = caption
		_, err := b.api.Send(v)
		return err
	case bridge.KindImage:
		p := tgbotapi.NewPhoto(tgChatID, file)
		p.Caption = caption
		_, err := b.api.Send(p)
		return err
	case bridge.KindVideo:
		v := tgbotapi.NewVideo(tgChatID, file)
		v.Caption = caption
		_, err := b.api.Send(v)
		return err
	}

	// Generic by extension
	switch ext {
	case ".webp", ".tgs", ".webm":
		if err := b.sendAsTelegramSticker(tgChatID, file, ext); err != nil {
			return err
		}
		if caption != "" {
			return b.SendText(tgChatID, caption)
		}
		return nil
	case ".gif", ".mp4":
		a := tgbotapi.NewAnimation(tgChatID, file)
		a.Caption = caption
		_, err := b.api.Send(a)
		return err
	case ".mkv", ".mov":
		v := tgbotapi.NewVideo(tgChatID, file)
		v.Caption = caption
		_, err := b.api.Send(v)
		return err
	case ".jpg", ".jpeg", ".png":
		p := tgbotapi.NewPhoto(tgChatID, file)
		p.Caption = caption
		_, err := b.api.Send(p)
		return err
	default:
		d := tgbotapi.NewDocument(tgChatID, file)
		d.Caption = caption
		_, err := b.api.Send(d)
		return err
	}
}

// sendAsTelegramSticker tries Bot API sendSticker with format-aware fallbacks.
func (b *Bot) sendAsTelegramSticker(tgChatID int64, file tgbotapi.RequestFileData, ext string) error {
	// Telegram Bot API: static WEBP/PNG, animated TGS, video WEBM.
	switch ext {
	case ".webp", ".tgs", ".webm", ".png":
		if _, err := b.api.Send(tgbotapi.NewSticker(tgChatID, file)); err == nil {
			return nil
		} else {
			b.log.Warnf("sendSticker (%s) failed: %v — falling back", ext, err)
		}
	}
	if ext == ".tgs" {
		// Still animated Lottie; document keeps the file even if sticker upload fails.
		_, err := b.api.Send(tgbotapi.NewDocument(tgChatID, file))
		return err
	}
	if ext == ".gif" {
		_, err := b.api.Send(tgbotapi.NewAnimation(tgChatID, file))
		return err
	}
	if ext == ".webm" {
		_, err := b.api.Send(tgbotapi.NewVideo(tgChatID, file))
		return err
	}
	// Static sticker that Telegram rejected as sticker → photo.
	if _, err := b.api.Send(tgbotapi.NewPhoto(tgChatID, file)); err == nil {
		return nil
	}
	_, err := b.api.Send(tgbotapi.NewDocument(tgChatID, file))
	return err
}
