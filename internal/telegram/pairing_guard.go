package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const pairingPrivateOnly = "Pairing only works in a private chat with this bot.\nOpen a 1:1 chat and send /pair there — do not use groups."

// IsPrivatePairingChat reports whether pairing commands may run here.
// Groups, supergroups, and channels leak codes to every member.
func IsPrivatePairingChat(msg *tgbotapi.Message) bool {
	if msg == nil {
		return false
	}
	return msg.Chat.IsPrivate()
}
