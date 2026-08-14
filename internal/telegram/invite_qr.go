package telegram

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	qrcode "github.com/skip2/go-qrcode"
)

// telegramCaptionMax is the Bot API limit for photo captions.
const telegramCaptionMax = 1024

// sendInviteQRPhoto sends one photo: PNG QR of the invite link with caption as the
// full /pair|/connect|/start pairing answer (not a second message).
func (b *Bot) sendInviteQRPhoto(chatID int64, inviteLink, caption string) error {
	inviteLink = strings.TrimSpace(inviteLink)
	if inviteLink == "" || strings.HasPrefix(inviteLink, "(") {
		// No usable link — text-only fallback
		_, err := b.api.Send(tgbotapi.NewMessage(chatID, caption))
		return err
	}

	dir := b.tmpdir
	if dir == "" {
		dir = os.TempDir()
	}
	path := filepath.Join(dir, fmt.Sprintf("dc-invite-qr-%d.png", chatID))
	if err := qrcode.WriteFile(inviteLink, qrcode.Medium, 512, path); err != nil {
		return fmt.Errorf("encode invite QR: %w", err)
	}
	defer os.Remove(path)

	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(path))
	photo.Caption = truncateCaption(caption, telegramCaptionMax)
	_, err := b.api.Send(photo)
	return err
}

func truncateCaption(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	if max < 2 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}
