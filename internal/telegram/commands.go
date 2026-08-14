package telegram

import (
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/omidz4t/portal/internal/config"
)

// botCommand is the Telegram BotCommand object for setMyCommands.
type botCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// supportedCommands is the public bot command menu (Telegram UI).
var supportedCommands = []botCommand{
	{Command: "start", Description: "Welcome + branding"},
	{Command: "pair", Description: "Get Delta Chat invite, QR, and pairing code"},
	{Command: "connect", Description: "Same as /pair — invite link + QR + code"},
	{Command: "disconnect", Description: "Unlink Telegram from Delta Chat"},
	{Command: "pair_bot", Description: "Register your BotFather bot (persona mode)"},
	{Command: "unpair_bot", Description: "Disable your persona bot"},
	{Command: "bots", Description: "List your persona bots"},
	{Command: "send_pack", Description: "Reply to a sticker — send full pack to Delta Chat"},
	{Command: "status", Description: "Show pairing status"},
	{Command: "help", Description: "List commands and how to use the bridge"},
}

// registerBotCommands publishes the command menu via setMyCommands.
// https://core.telegram.org/bots/api#setmycommands
func (b *Bot) registerBotCommands() error {
	payload, err := json.Marshal(supportedCommands)
	if err != nil {
		return err
	}
	_, err = b.api.MakeRequest("setMyCommands", tgbotapi.Params{
		"commands": string(payload),
	})
	if err != nil {
		return fmt.Errorf("setMyCommands: %w", err)
	}
	b.log.Infof("registered %d Telegram bot commands", len(supportedCommands))
	return nil
}

func commandsHelpText(cfg config.Config) string {
	s := "Delta ↔️ TG commands\n\n" +
		"/start — welcome (or /start CODE from a Delta Chat link)\n" +
		"/pair or /connect — Delta Chat invite link + QR + pairing code\n" +
		"/disconnect — unlink this Telegram from Delta Chat\n"
	if cfg.PersonaEnabled() {
		s += "/pair-bot <TOKEN> [url] — register your own bot (persona mode)\n" +
			"/unpair-bot [id|@user] — disable persona bot(s)\n" +
			"/bots — list your persona bots\n"
	}
	s += "/send_pack — reply to a sticker, send the whole pack to Delta Chat\n" +
		"/status — pairing status\n" +
		"/help — this list\n\n" +
		"After pairing: text, photos, short videos, stickers, custom emoji, GIFs either way.\n" +
		"(Short video max duration/size: see config.yml bridge.limits.)"
	if cfg.PersonaEnabled() {
		s += "\n\nPersona mode: after /pair-bot, people who message your bot appear as Delta Chat contacts."
	}
	return s
}
