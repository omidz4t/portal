package telegram

import (
	"encoding/json"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// setMessageReaction sets a Telegram emoji reaction on a message (Bot API 7+).
// https://core.telegram.org/bots/api#setmessagereaction
func (b *Bot) setMessageReaction(chatID int64, messageID int, emoji string) error {
	emoji = trimReaction(emoji)
	if emoji == "" || messageID == 0 {
		return nil
	}

	reaction, err := json.Marshal([]map[string]string{
		{"type": "emoji", "emoji": emoji},
	})
	if err != nil {
		return err
	}

	params := tgbotapi.Params{
		"chat_id":    strconv.FormatInt(chatID, 10),
		"message_id": strconv.Itoa(messageID),
		"reaction":   string(reaction),
	}
	_, err = b.api.MakeRequest("setMessageReaction", params)
	if err != nil {
		return fmt.Errorf("setMessageReaction: %w", err)
	}
	return nil
}

// reactOK reacts to a successfully bridged Telegram message (best-effort).
func (b *Bot) reactOK(msg *tgbotapi.Message) {
	if msg == nil {
		return
	}
	emoji := b.cfg.Telegram.Reaction
	if emoji == "" {
		emoji = defaultReaction
	}
	if emoji == "-" || emoji == "off" || emoji == "none" {
		return
	}
	if err := b.setMessageReaction(msg.Chat.ID, msg.MessageID, emoji); err != nil {
		b.log.Warnf("reaction failed chat=%d msg=%d: %v", msg.Chat.ID, msg.MessageID, err)
	}
}

const defaultReaction = "✅"

func trimReaction(s string) string {
	// keep multi-rune emoji; only strip spaces
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
