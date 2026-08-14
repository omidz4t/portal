package persona

import (
	"fmt"
	"strings"
	"sync"

	"github.com/omidz4t/portal/internal/store"
)

// groupLocks serialize create/join per (personaBot, tgChat).
var (
	groupMu    sync.Mutex
	groupLocks = map[string]*sync.Mutex{}
)

func groupKey(botID, tgChat int64) string {
	return fmt.Sprintf("%d:%d", botID, tgChat)
}

func lockGroup(botID, tgChat int64) func() {
	key := groupKey(botID, tgChat)
	groupMu.Lock()
	lk, ok := groupLocks[key]
	if !ok {
		lk = &sync.Mutex{}
		groupLocks[key] = lk
	}
	groupMu.Unlock()
	lk.Lock()
	return lk.Unlock
}

// EnsureGroup ensures a mirrored DC group exists for a Telegram group and that
// the sender ghost (and owner) are members. Returns the group chat id on the
// *sender's* account for SendMsg.
//
// Model (design §5.3):
//   - First speaker creates an encrypted group, adds the owner (key-contact), becomes coordinator.
//   - Later speakers SecureJoin the group invite (after the group is promoted by a real message).
//   - Each speaker sends into the group as their own ghost account.
func (m *Manager) EnsureGroup(bot *store.PersonaBot, ghost *store.GhostAccount, tgChatID int64, title string) (memberChatID uint32, err error) {
	unlock := lockGroup(bot.ID, tgChatID)
	defer unlock()

	if title == "" {
		title = fmt.Sprintf("TG group %d", tgChatID)
	}
	// Keep original title; prefix only if not already tagged.
	if !strings.HasPrefix(title, "TG: ") && !strings.HasPrefix(title, "TG ") {
		title = "TG: " + title
	}

	gg, err := m.store.GetGhostGroup(bot.ID, tgChatID)
	if err != nil {
		return 0, err
	}

	// Owner must be a key-contact on this ghost (for AddContactToChat into encrypted group).
	ownerChat, err := m.ensureOwnerChat(ghost, bot)
	if err != nil {
		return 0, fmt.Errorf("owner chat for group: %w", err)
	}
	ownerContactID, err := m.dc.FirstPeerContactID(ghost.DCAccountID, ownerChat)
	if err != nil {
		return 0, fmt.Errorf("owner contact: %w", err)
	}

	if gg == nil {
		return m.createMirroredGroup(bot, ghost, tgChatID, title, ownerContactID)
	}

	// Already a member with a known local chat id.
	mem, err := m.store.GetGhostGroupMember(gg.ID, ghost.TelegramUserID)
	if err != nil {
		return 0, err
	}
	if mem != nil && mem.MemberDCChatID != 0 {
		if title != gg.Title {
			_ = m.store.UpdateGhostGroupMeta(gg.ID, title, gg.InviteQR, gg.CoordinatorDCAccountID, gg.DCChatID)
		}
		return mem.MemberDCChatID, nil
	}

	// Join existing mirrored group.
	joined, err := m.joinMirroredGroup(bot, ghost, gg)
	if err != nil {
		return 0, err
	}
	_ = m.store.UpsertGhostGroupMember(&store.GhostGroupMember{
		GhostGroupID:   gg.ID,
		TelegramUserID: ghost.TelegramUserID,
		GhostAccountID: ghost.ID,
		MemberDCChatID: joined,
	})
	return joined, nil
}

func (m *Manager) createMirroredGroup(bot *store.PersonaBot, ghost *store.GhostAccount, tgChatID int64, title string, ownerContactID uint32) (uint32, error) {
	dcChatID, err := m.dc.CreateGroupChat(ghost.DCAccountID, title)
	if err != nil {
		return 0, fmt.Errorf("create group: %w", err)
	}
	if err := m.dc.AddContactIDToChat(ghost.DCAccountID, dcChatID, ownerContactID); err != nil {
		return 0, fmt.Errorf("add owner to group: %w", err)
	}
	// Invite QR is often only useful after the group is promoted (first message).
	// Store empty for now; RefreshGroupInvite after first successful send.
	gg, err := m.store.InsertGhostGroup(&store.GhostGroup{
		PersonaBotID:           bot.ID,
		TelegramChatID:         tgChatID,
		Title:                  title,
		CoordinatorDCAccountID: ghost.DCAccountID,
		DCChatID:               dcChatID,
		InviteQR:               "",
	})
	if err != nil {
		return 0, err
	}
	_ = m.store.UpsertGhostGroupMember(&store.GhostGroupMember{
		GhostGroupID:   gg.ID,
		TelegramUserID: ghost.TelegramUserID,
		GhostAccountID: ghost.ID,
		MemberDCChatID: dcChatID,
	})
	m.log.Infof("created mirrored group tg=%d dc_chat=%d coord_acc=%d title=%q",
		tgChatID, dcChatID, ghost.DCAccountID, title)
	return dcChatID, nil
}

func (m *Manager) joinMirroredGroup(bot *store.PersonaBot, ghost *store.GhostAccount, gg *store.GhostGroup) (uint32, error) {
	if ghost.DCAccountID == gg.CoordinatorDCAccountID {
		return gg.DCChatID, nil
	}

	// Prefer SecureJoin with group invite (after promotion).
	if strings.TrimSpace(gg.InviteQR) != "" {
		id, err := m.dc.SecureJoin(ghost.DCAccountID, gg.InviteQR)
		if err == nil && id != 0 {
			m.log.Infof("ghost %d joined group via SecureJoin → chat %d", ghost.DCAccountID, id)
			return id, nil
		}
		m.log.Warnf("securejoin group: %v", err)
	}

	// Coordinator tries to add this ghost by address, then refresh invite and SecureJoin again.
	if ghost.DCAddress != "" && gg.CoordinatorDCAccountID != 0 && gg.DCChatID != 0 {
		if err := m.dc.AddContactToChatByAddress(gg.CoordinatorDCAccountID, gg.DCChatID, ghost.DCAddress, ghost.DisplayName); err != nil {
			m.log.Warnf("coordinator add member: %v", err)
		}
		// Refresh invite from coordinator after membership change.
		if inv, err := m.dc.GetGroupSecurejoinQR(gg.CoordinatorDCAccountID, gg.DCChatID); err == nil && inv != "" {
			_ = m.store.UpdateGhostGroupMeta(gg.ID, gg.Title, inv, gg.CoordinatorDCAccountID, gg.DCChatID)
			gg.InviteQR = inv
			if id, err := m.dc.SecureJoin(ghost.DCAccountID, inv); err == nil && id != 0 {
				return id, nil
			}
		}
	}

	// Last resort: import coordinator's self? fail clearly.
	return 0, fmt.Errorf("could not join mirrored group %q (tg %d): invite missing or SecureJoin failed — wait for first speaker message to promote the group, then retry",
		gg.Title, gg.TelegramChatID)
}

// RefreshGroupInvite stores a SecureJoin QR after the group is promoted (first real message).
func (m *Manager) RefreshGroupInvite(bot *store.PersonaBot, ghost *store.GhostAccount, tgChatID int64, memberChatID uint32) {
	gg, err := m.store.GetGhostGroup(bot.ID, tgChatID)
	if err != nil || gg == nil {
		return
	}
	// Only coordinator should refresh the canonical invite.
	if ghost.DCAccountID != gg.CoordinatorDCAccountID {
		return
	}
	inv, err := m.dc.GetGroupSecurejoinQR(ghost.DCAccountID, memberChatID)
	if err != nil || inv == "" {
		return
	}
	if inv == gg.InviteQR {
		return
	}
	_ = m.store.UpdateGhostGroupMeta(gg.ID, gg.Title, inv, gg.CoordinatorDCAccountID, gg.DCChatID)
	m.log.Infof("refreshed group invite tg=%d", tgChatID)
}

// IsTelegramGroup reports whether a Telegram chat is a group/supergroup/channel.
// Group and supergroup IDs are always negative in the Bot API.
func IsTelegramGroup(chatID int64, chatType string) bool {
	if chatID < 0 {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(chatType)) {
	case "group", "supergroup", "channel":
		return true
	default:
		return false
	}
}
