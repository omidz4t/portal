package persona

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/omidz4t/portal/internal/dc"
	"github.com/omidz4t/portal/internal/store"
)

// PortalHooks implements telegram.PersonaHooks using Manager + DC peer lookup.
type PortalHooks struct {
	M  *Manager
	DC *dc.Session
	St *store.Store
}

// RegisterToken implements telegram.PersonaHooks.
func (h *PortalHooks) RegisterToken(ownerTG int64, ownerUsername string, token, botURL string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" || !strings.Contains(token, ":") {
		return "", fmt.Errorf("invalid bot token format")
	}
	if h.M != nil && IsPortalBotToken(token, h.M.cfg.TelegramToken) {
		return "", fmt.Errorf("that token is the portal bot — create a new bot with @BotFather")
	}
	pair, err := h.St.GetActiveByTelegram(ownerTG)
	if err != nil {
		return "", err
	}
	if pair == nil {
		return "", fmt.Errorf("pair with Delta Chat first (/pair)")
	}

	// Ensure we have the owner's public key (vcard) from mode-1 pairing.
	vcard, err := h.M.EnsureOwnerVcardForOwner(ownerTG, pair.DCAccountID, pair.DCChatID)
	if err != nil {
		return "", fmt.Errorf("could not export your DC public key — re-run /pair and finish pairing, then try again: %v", err)
	}

	ownerAddr := ""
	if peer, err := h.DC.PeerContact(pair.DCAccountID, pair.DCChatID); err == nil {
		ownerAddr = peer.Address
	}

	if strings.TrimSpace(h.M.cfg.Persona.AccountQR) == "" {
		return "", fmt.Errorf("persona account_qr is not set — add PERSONA_ACCOUNT_QR=dcaccount:… to .env and restart")
	}
	client, err := NewTGClient(token, h.M.cfg, h.M.log)
	if err != nil {
		return "", fmt.Errorf("telegram rejected token (check BotFather token)")
	}
	uname := client.api.Self.UserName
	botUserID := int64(client.api.Self.ID)
	if botURL == "" && uname != "" {
		botURL = "https://t.me/" + uname
	}

	b := &store.PersonaBot{
		OwnerTelegramUserID: ownerTG,
		OwnerDCAccountID:    pair.DCAccountID,
		OwnerDCChatID:       pair.DCChatID,
		OwnerDCAddress:      ownerAddr,
		OwnerVcard:          vcard,
		BotToken:            token,
		BotUserID:           botUserID,
		BotUsername:         uname,
		BotURL:              botURL,
		Status:              store.PersonaBotActive,
	}
	_, err = h.M.RegisterBot(b, h.M.StartPoller)
	if err != nil {
		return uname, err
	}
	_ = ownerUsername
	return uname, nil
}

// Unregister implements telegram.PersonaHooks.
func (h *PortalHooks) Unregister(ownerTG int64, botRef string) (int64, error) {
	botRef = strings.TrimSpace(botRef)
	var id int64
	var username string
	if botRef != "" {
		if strings.HasPrefix(botRef, "@") || !isAllDigits(botRef) {
			username = strings.TrimPrefix(botRef, "@")
		} else {
			id, _ = strconv.ParseInt(botRef, 10, 64)
		}
	}
	bots, _ := h.St.ListPersonaBotsByOwner(ownerTG)
	for _, b := range bots {
		if b.Status != store.PersonaBotActive {
			continue
		}
		match := botRef == "" || b.ID == id || strings.EqualFold(b.BotUsername, username)
		if match {
			h.M.DetachBot(b.ID)
		}
	}
	return h.St.DisablePersonaBotByOwner(ownerTG, id, username)
}

// ListBots implements telegram.PersonaHooks.
func (h *PortalHooks) ListBots(ownerTG int64) (string, error) {
	bots, err := h.St.ListPersonaBotsByOwner(ownerTG)
	if err != nil {
		return "", err
	}
	if len(bots) == 0 {
		return "No persona bots. After /pair, use:\n/pair-bot <TOKEN> [https://t.me/YourBot]", nil
	}
	var b strings.Builder
	b.WriteString("Your persona bots:\n")
	for _, bot := range bots {
		b.WriteString(fmt.Sprintf("\n• id=%d @%s — %s", bot.ID, bot.BotUsername, bot.Status))
		if bot.BotURL != "" {
			b.WriteString("\n  " + bot.BotURL)
		}
		if strings.TrimSpace(bot.OwnerVcard) == "" {
			b.WriteString("\n  ⚠ missing owner key (re-pair portal)")
		} else {
			b.WriteString("\n  owner key: ok")
		}
	}
	b.WriteString("\n\n/unpair-bot [id|@user] to disable")
	return b.String(), nil
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
