package persona

import (
	"fmt"
	"strings"

	"github.com/chatmail/rpc-client-go/v2/deltachat"

	"github.com/omidz4t/portal/internal/store"
)

// TGUser is the remote Telegram person (not the bot).
type TGUser struct {
	ID          int64
	Username    string
	DisplayName string
}

// Incoming is a message from a persona Telegram bot to bridge into Delta Chat.
type Incoming struct {
	Bot        *store.PersonaBot
	From       TGUser
	ChatID     int64
	IsGroup    bool
	GroupTitle string
	Text       string
	FilePath   string
	FileName   string
	Viewtype   *deltachat.Viewtype
	// Avatar is used only when creating/updating a ghost (optional).
	Avatar AvatarDownloader
}

// BridgeToDelta implements pure impersonation delivery (design §5.3):
//  1. GetOrCreateGhost (name + photo + import owner key)
//  2. Send as ghost to owner 1:1, verbatim — no labels
func (m *Manager) BridgeToDelta(in Incoming) error {
	if in.Bot == nil {
		return fmt.Errorf("nil persona bot")
	}
	if in.From.ID == 0 {
		return fmt.Errorf("missing from user")
	}
	if in.From.ID == in.Bot.BotUserID {
		return nil
	}

	ghost, err := m.GetOrCreateGhost(in.Bot, in.From.ID, in.From.Username, in.From.DisplayName, in.Avatar)
	if err != nil {
		return err
	}
	m.log.Infow("persona inbound",
		"persona_bot", in.Bot.BotUsername,
		"tg_user", in.From.ID,
		"tg_chat", in.ChatID,
		"ghost_dc_account", ghost.DCAccountID,
		"ghost_addr", ghost.DCAddress,
		"owner_chat", ghost.OwnerChatID,
		"is_group", in.IsGroup,
		"group_title", in.GroupTitle,
	)

	if in.IsGroup {
		return m.bridgeGroup(in, ghost)
	}
	return m.bridgeDM(in, ghost)
}

func (m *Manager) bridgeDM(in Incoming, ghost *store.GhostAccount) error {
	chatID, err := m.ensureOwnerChat(ghost, in.Bot)
	if err != nil {
		return fmt.Errorf("owner chat: %w", err)
	}
	ghost.OwnerChatID = chatID

	err = m.sendAsGhost(ghost.DCAccountID, chatID, in)
	if err != nil && isKeyMissingErr(err) {
		// Stale unencrypted chat (e.g. old CreateContact) — force re-import and retry once.
		m.log.Warnf("key missing on chat %d — reimport owner key and retry", chatID)
		_ = m.store.UpdateGhostOwnerChat(ghost.ID, 0)
		ghost.OwnerChatID = 0
		chatID, err = m.ensureOwnerChat(ghost, in.Bot)
		if err != nil {
			return fmt.Errorf("owner chat after reimport: %w", err)
		}
		err = m.sendAsGhost(ghost.DCAccountID, chatID, in)
	}
	if err != nil {
		return fmt.Errorf("send as ghost: %w", err)
	}
	m.log.Infow("delivered ghost→owner 1:1",
		"ghost_dc_account", ghost.DCAccountID,
		"owner_chat", chatID,
		"tg_user", in.From.ID,
	)
	return nil
}

func isKeyMissingErr(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "key is missing") ||
		strings.Contains(s, "e2e encryption unavailable") ||
		strings.Contains(s, "encryption unavailable")
}

func (m *Manager) bridgeGroup(in Incoming, ghost *store.GhostAccount) error {
	chatID, err := m.EnsureGroup(in.Bot, ghost, in.ChatID, in.GroupTitle)
	if err != nil {
		return fmt.Errorf("ensure group: %w", err)
	}
	if err := m.sendAsGhost(ghost.DCAccountID, chatID, in); err != nil {
		return fmt.Errorf("group send as ghost (dc_chat=%d): %w", chatID, err)
	}
	// After first real message the group is promoted — publish invite for later joiners.
	m.RefreshGroupInvite(in.Bot, ghost, in.ChatID, chatID)
	m.log.Infow("delivered ghost→owner group",
		"ghost_dc_account", ghost.DCAccountID,
		"dc_group_chat", chatID,
		"tg_chat", in.ChatID,
		"tg_user", in.From.ID,
		"title", in.GroupTitle,
	)
	return nil
}

func (m *Manager) sendAsGhost(accID, chatID uint32, in Incoming) error {
	_ = m.dc.AcceptChat(accID, chatID)
	if in.FilePath != "" {
		// Verbatim caption only — no bridge labels.
		return m.dc.SendFileWithRetry(accID, chatID, in.FilePath, in.FileName, strings.TrimSpace(in.Text), in.Viewtype, 15)
	}
	text := strings.TrimSpace(in.Text)
	if text == "" {
		return nil
	}
	return m.dc.SendTextWithRetry(accID, chatID, text, 15)
}
