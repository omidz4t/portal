package telegram

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/omidz4t/portal/internal/usermsg"
)

// cmdSendPack: reply to a sticker in the bot chat with /send_pack
// → fetch full sticker set and bridge every sticker to paired Delta Chat.
func (b *Bot) cmdSendPack(msg *tgbotapi.Message) error {
	if !b.cfg.Bridge.StickerPacks {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Sticker packs are disabled in config (bridge.sticker_packs)."))
		return err
	}

	reply := msg.ReplyToMessage
	if reply == nil || reply.Sticker == nil {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID,
			"/send_pack\n\nReply to a sticker in this chat, then send /send_pack.\n"+
				"I will send the whole pack to your paired Delta Chat."))
		return err
	}

	st := reply.Sticker
	setName := strings.TrimSpace(st.SetName)
	if setName == "" {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID,
			"That sticker has no pack name (custom/single stickers can't be expanded)."))
		return err
	}

	pair, err := b.store.GetActiveByTelegram(msg.From.ID)
	if err != nil {
		return err
	}
	if pair == nil {
		_, err := b.api.Send(tgbotapi.NewMessage(msg.Chat.ID,
			"Not paired. Use /pair first, then /send_pack again."))
		return err
	}

	// Fetch + send off the Telegram worker so getUpdates stays responsive.
	msgCopy := *msg
	launchSendPack(func() { b.runSendPack(&msgCopy, setName) })
	return nil
}

// launchSendPack starts pack work without waiting (overridable in tests).
var launchSendPack = func(work func()) { go work() }

func (b *Bot) runSendPack(msg *tgbotapi.Message, setName string) {
	set, err := b.api.GetStickerSet(tgbotapi.GetStickerSetConfig{Name: setName})
	if err != nil {
		b.log.Warnf("send_pack getStickerSet %s: %v", setName, err)
		_, _ = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, usermsg.BridgeFailed))
		return
	}
	if len(set.Stickers) == 0 {
		_, _ = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Pack is empty."))
		return
	}
	max := b.cfg.Bridge.Limits.StickerPackMax
	if max <= 0 {
		max = 120
	}
	stickers := set.Stickers
	truncated := false
	if len(stickers) > max {
		stickers = stickers[:max]
		truncated = true
	}
	_, _ = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID,
		fmt.Sprintf("Sending pack “%s” (%d stickers%s) to Delta Chat…",
			set.Title, len(stickers), map[bool]string{true: fmt.Sprintf(", max %d", max), false: ""}[truncated])))
	b.sendPackStickers(msg, setName, set.Title, set.IsAnimated, stickers, truncated)
}

func (b *Bot) sendPackStickers(msg *tgbotapi.Message, setName, title string, setAnimated bool, stickers []tgbotapi.Sticker, truncated bool) {
	sent, failed := 0, 0
	for i, s := range stickers {
		isVideo := false
		if fm, e := b.api.GetFile(tgbotapi.FileConfig{FileID: s.FileID}); e == nil {
			isVideo = strings.HasSuffix(strings.ToLower(fm.FilePath), ".webm")
		}
		if err := b.bridgeStickerFile(msg.From.ID, s.FileID, s.IsAnimated || setAnimated, isVideo); err != nil {
			failed++
			b.log.Warnf("send_pack [%d/%d] %s: %v", i+1, len(stickers), setName, err)
			continue
		}
		sent++
	}
	summary := fmt.Sprintf("Pack “%s” done: %d sent, %d failed.", title, sent, failed)
	if truncated {
		summary += fmt.Sprintf(" (pack truncated to %d by bridge.limits.sticker_pack_max)", b.cfg.Bridge.Limits.StickerPackMax)
	}
	_, _ = b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, summary))
	if sent > 0 {
		b.reactOK(msg)
	}
	b.log.Infof("send_pack user=%d set=%s sent=%d failed=%d", msg.From.ID, setName, sent, failed)
}
