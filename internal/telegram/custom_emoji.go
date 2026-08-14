package telegram

import (
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/omidz4t/portal/internal/bridge"
)

// entityWire includes custom_emoji_id (missing from telegram-bot-api v5.5 types).
type entityWire struct {
	Type          string `json:"type"`
	Offset        int    `json:"offset"`
	Length        int    `json:"length"`
	CustomEmojiID string `json:"custom_emoji_id,omitempty"`
}

type messageEntitiesWire struct {
	Entities        []entityWire `json:"entities"`
	CaptionEntities []entityWire `json:"caption_entities"`
}

// extractCustomEmojiIDs collects unique custom_emoji_id values from message JSON.
func extractCustomEmojiIDs(msgJSON []byte) []string {
	var w messageEntitiesWire
	if err := json.Unmarshal(msgJSON, &w); err != nil {
		return nil
	}
	return uniqueEmojiIDs(w.Entities, w.CaptionEntities)
}

func uniqueEmojiIDs(lists ...[]entityWire) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, list := range lists {
		for _, e := range list {
			if e.Type != "custom_emoji" || e.CustomEmojiID == "" {
				continue
			}
			if _, ok := seen[e.CustomEmojiID]; ok {
				continue
			}
			seen[e.CustomEmojiID] = struct{}{}
			out = append(out, e.CustomEmojiID)
		}
	}
	return out
}

// stickerFile is a minimal sticker descriptor from getCustomEmojiStickers.
type stickerFile struct {
	FileID     string `json:"file_id"`
	IsAnimated bool   `json:"is_animated"`
	IsVideo    bool   `json:"is_video"`
	Emoji      string `json:"emoji"`
}

// getCustomEmojiStickers calls Telegram getCustomEmojiStickers.
// https://core.telegram.org/bots/api#getcustomemojistickers
func (b *Bot) getCustomEmojiStickers(ids []string) ([]stickerFile, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	payload, err := json.Marshal(ids)
	if err != nil {
		return nil, err
	}
	resp, err := b.api.MakeRequest("getCustomEmojiStickers", tgbotapi.Params{
		"custom_emoji_ids": string(payload),
	})
	if err != nil {
		return nil, err
	}
	var stickers []stickerFile
	if err := json.Unmarshal(resp.Result, &stickers); err != nil {
		return nil, fmt.Errorf("decode custom emoji stickers: %w", err)
	}
	return stickers, nil
}

// handleCustomEmojis resolves custom emoji entities to sticker files and bridges each.
func (b *Bot) handleCustomEmojis(msg *tgbotapi.Message, emojiIDs []string) error {
	if !b.bridge.AllowsKind(bridge.KindCustomEmoji) {
		return fmt.Errorf("custom_emoji bridge disabled in config")
	}
	if len(emojiIDs) == 0 {
		return nil
	}

	const chunk = 100
	var stickers []stickerFile
	for i := 0; i < len(emojiIDs); i += chunk {
		end := i + chunk
		if end > len(emojiIDs) {
			end = len(emojiIDs)
		}
		part, err := b.getCustomEmojiStickers(emojiIDs[i:end])
		if err != nil {
			return err
		}
		stickers = append(stickers, part...)
	}
	if len(stickers) == 0 {
		return fmt.Errorf("no sticker files returned for custom emoji ids")
	}

	var firstErr error
	sent := 0
	for _, st := range stickers {
		// Force KindCustomEmoji path via AllowsKind fallback inside bridgeStickerFile.
		if err := b.bridgeStickerFile(msg.From.ID, st.FileID, st.IsAnimated, st.IsVideo); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			b.log.Warnf("custom emoji bridge: %v", err)
			continue
		}
		sent++
	}
	if sent == 0 && firstErr != nil {
		return firstErr
	}
	b.log.Infof("bridged %d custom emoji(s) for tg user %d", sent, msg.From.ID)
	return nil
}
