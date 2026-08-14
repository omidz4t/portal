package persona

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/omidz4t/portal/internal/bridge"
	"github.com/omidz4t/portal/internal/config"
	"github.com/omidz4t/portal/internal/dc"
	"github.com/omidz4t/portal/internal/safemedia"
	"github.com/omidz4t/portal/internal/store"
)

// Manager owns persona bot pollers and ghost provisioning.
type Manager struct {
	log   *zap.SugaredLogger
	cfg   config.Config
	dc    *dc.Session
	store *store.Store

	mu      sync.Mutex
	pollers map[int64]*BotRuntime // persona_bot.id → runtime

	ghostMu    sync.Mutex
	ghostLocks map[string]*sync.Mutex
}

// BotRuntime is a running poller for one user-owned Telegram bot.
type BotRuntime struct {
	Bot     store.PersonaBot
	API     TelegramAPI
	stop    chan struct{}
	stopped chan struct{}
}

// TelegramAPI is the subset of Bot API used by persona pollers.
type TelegramAPI interface {
	SendText(chatID int64, text string) error
	SendMediaFile(chatID int64, path, filename string, asPhoto, asVideo, asAnimation, asDocument bool) error
	GetMeUsername() string
	Token() string
	// DownloadProfilePhoto saves the user's current profile photo; empty path if none.
	DownloadProfilePhoto(userID int64) (localPath, fileID string, err error)
}

// New creates a manager (call Start after DC session is ready).
func New(log *zap.SugaredLogger, cfg config.Config, sess *dc.Session, st *store.Store) *Manager {
	return &Manager{
		log:        log.With("component", "persona"),
		cfg:        cfg,
		dc:         sess,
		store:      st,
		pollers:    make(map[int64]*BotRuntime),
		ghostLocks: make(map[string]*sync.Mutex),
	}
}

// Start loads active persona bots and begins polling each.
func (m *Manager) Start(startPoller func(bot store.PersonaBot) (*BotRuntime, error)) error {
	if !m.cfg.PersonaEnabled() {
		m.log.Info("persona mode disabled")
		return nil
	}
	if m.cfg.Persona.AccountQR == "" {
		m.log.Warn("persona enabled but account_qr / PERSONA_ACCOUNT_QR empty — ghost creation will fail until set")
	}
	// Backfill owner vcards for existing pairs/bots (pairs created before key capture).
	// Do not flip disabled → active: /unpair-bot must survive restart.
	m.backfillOwnerVcards()

	bots, err := m.store.ListActivePersonaBots()
	if err != nil {
		return err
	}
	m.log.Infof("starting %d active persona bot(s)", len(bots))
	for _, b := range bots {
		// Refresh in-memory copy after backfill.
		if fresh, _ := m.store.GetPersonaBot(b.ID); fresh != nil {
			b = *fresh
		}
		if err := m.AttachBot(b, startPoller); err != nil {
			m.log.Errorf("persona bot id=%d @%s: %v", b.ID, b.BotUsername, err)
			_ = m.store.SetPersonaBotStatus(b.ID, store.PersonaBotError)
		}
	}
	return nil
}

// backfillOwnerVcards exports peer vcards for active pairs missing them and
// copies into persona_bots. Does not change persona bot status. Safe to call repeatedly.
func (m *Manager) backfillOwnerVcards() {
	pairs, err := m.store.ListActivePairs()
	if err != nil {
		m.log.Warnf("list active pairs: %v", err)
		return
	}
	for _, p := range pairs {
		if _, err := m.EnsureOwnerVcardForOwner(p.TelegramUserID, p.DCAccountID, p.DCChatID); err != nil {
			m.log.Warnf("backfill vcard owner_tg=%d chat=%d: %v", p.TelegramUserID, p.DCChatID, err)
		}
	}
	// Sync vcards onto any persona bots (including disabled) that still have empty owner_vcard.
	bots, err := m.store.ListPersonaBotsByOwnerAny()
	if err != nil {
		return
	}
	for _, b := range bots {
		if strings.TrimSpace(b.OwnerVcard) != "" {
			continue
		}
		if _, err := m.EnsureOwnerVcardForOwner(b.OwnerTelegramUserID, b.OwnerDCAccountID, b.OwnerDCChatID); err != nil {
			m.log.Warnf("backfill bot id=%d: %v", b.ID, err)
		}
	}
}

// EnsureOwnerVcardForOwner captures/stores owner vcard from the portal pair chat.
func (m *Manager) EnsureOwnerVcardForOwner(ownerTG int64, fallbackAcc, fallbackChat uint32) (string, error) {
	pair, err := m.store.GetActiveByTelegram(ownerTG)
	if err != nil {
		return "", err
	}
	acc, chat := fallbackAcc, fallbackChat
	if pair != nil {
		if strings.TrimSpace(pair.OwnerVcard) != "" {
			if n, _ := m.store.UpdatePersonaOwnerVcardForOwner(ownerTG, pair.OwnerVcard); n > 0 {
				m.log.Infof("synced owner vcard to %d persona bot(s) (from pair)", n)
			}
			return pair.OwnerVcard, nil
		}
		if pair.DCAccountID != 0 && pair.DCChatID != 0 {
			acc, chat = pair.DCAccountID, pair.DCChatID
		}
	}
	if acc == 0 || chat == 0 {
		return "", fmt.Errorf("no portal pair chat to export owner key from (owner must /pair first)")
	}
	vcard, peer, err := m.dc.ExportPeerVcard(acc, chat)
	if err != nil {
		return "", fmt.Errorf("export peer vcard chat=%d: %w", chat, err)
	}
	if pair != nil {
		_ = m.store.SetPairOwnerVcard(pair.ID, vcard)
	}
	if n, _ := m.store.UpdatePersonaOwnerVcardForOwner(ownerTG, vcard); n > 0 {
		m.log.Infof("captured owner vcard (%d bytes, peer=%s) → %d persona bot(s)", len(vcard), peer.Address, n)
	} else {
		m.log.Infof("captured owner vcard (%d bytes, peer=%s)", len(vcard), peer.Address)
	}
	return vcard, nil
}

// EnsureOwnerVcard returns a non-empty owner vcard for a persona bot, capturing if needed.
func (m *Manager) EnsureOwnerVcard(bot *store.PersonaBot) (string, error) {
	if bot == nil {
		return "", fmt.Errorf("nil bot")
	}
	if v := strings.TrimSpace(bot.OwnerVcard); v != "" {
		return v, nil
	}
	v, err := m.EnsureOwnerVcardForOwner(bot.OwnerTelegramUserID, bot.OwnerDCAccountID, bot.OwnerDCChatID)
	if err != nil {
		return "", err
	}
	bot.OwnerVcard = v
	_ = m.store.UpdatePersonaOwnerVcard(bot.ID, v)
	return v, nil
}

// AttachBot starts polling for an already-persisted bot.
func (m *Manager) AttachBot(b store.PersonaBot, startPoller func(bot store.PersonaBot) (*BotRuntime, error)) error {
	m.mu.Lock()
	if _, ok := m.pollers[b.ID]; ok {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	rt, err := startPoller(b)
	if err != nil {
		return err
	}

	m.mu.Lock()
	if _, ok := m.pollers[b.ID]; ok {
		m.mu.Unlock()
		if rt != nil && rt.stop != nil {
			close(rt.stop) // pollLoop exits; do not wait (StartPoller may not have started it yet)
		}
		return nil
	}
	m.pollers[b.ID] = rt
	m.mu.Unlock()
	m.log.Infof("persona bot polling @%s (id=%d)", b.BotUsername, b.ID)
	return nil
}

// dropPoller removes a poller entry without waiting. Call from the poll loop
// itself (e.g. Unauthorized) so DetachBot cannot deadlock on stopped.
func (m *Manager) dropPoller(id int64) {
	m.mu.Lock()
	delete(m.pollers, id)
	m.mu.Unlock()
}

// DetachBot stops a poller without deleting ghosts.
func (m *Manager) DetachBot(id int64) {
	m.mu.Lock()
	rt, ok := m.pollers[id]
	if ok {
		delete(m.pollers, id)
	}
	m.mu.Unlock()
	if !ok || rt == nil {
		return
	}
	close(rt.stop)
	<-rt.stopped
}

// Stop all pollers.
func (m *Manager) Stop() {
	m.mu.Lock()
	ids := make([]int64, 0, len(m.pollers))
	for id := range m.pollers {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	for _, id := range ids {
		m.DetachBot(id)
	}
}

// Runtime returns a running bot runtime by persona bot id.
func (m *Manager) Runtime(id int64) *BotRuntime {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pollers[id]
}

func (m *Manager) ghostKey(botID, tgUser int64) string {
	return fmt.Sprintf("%d:%d", botID, tgUser)
}

func (m *Manager) lockGhost(botID, tgUser int64) func() {
	key := m.ghostKey(botID, tgUser)
	m.ghostMu.Lock()
	lk, ok := m.ghostLocks[key]
	if !ok {
		lk = &sync.Mutex{}
		m.ghostLocks[key] = lk
	}
	m.ghostMu.Unlock()
	lk.Lock()
	return lk.Unlock
}

// AvatarDownloader optionally supplies Telegram profile photos during ghost create.
type AvatarDownloader interface {
	DownloadProfilePhoto(userID int64) (localPath, fileID string, err error)
}

// GetOrCreateGhost returns the unique DC ghost for (personaBot, tgUser), creating if needed.
// New ghosts: configure account, set name+avatar, import owner's vcard (public key), open 1:1.
func (m *Manager) GetOrCreateGhost(bot *store.PersonaBot, tgUserID int64, username, displayName string, av AvatarDownloader) (*store.GhostAccount, error) {
	unlock := m.lockGhost(bot.ID, tgUserID)
	defer unlock()

	existing, err := m.store.GetGhostByTG(bot.ID, tgUserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if displayName != "" && (existing.DisplayName != displayName || existing.TelegramUsername != username) {
			_ = m.store.UpdateGhostProfile(existing.ID, username, displayName)
			existing.DisplayName = displayName
			existing.TelegramUsername = username
			_ = m.dc.SetDisplayName(existing.DCAccountID, displayName)
		}
		// Always ensure owner key-contact chat (fixes old ghosts with missing keys).
		if chatID, err := m.ensureOwnerChat(existing, bot); err != nil {
			m.log.Warnf("ensure owner chat ghost=%d: %v", existing.ID, err)
		} else {
			existing.OwnerChatID = chatID
		}
		if av != nil {
			m.maybeSyncAvatar(existing, av)
		}
		return existing, nil
	}

	if _, err := m.EnsureOwnerVcard(bot); err != nil {
		return nil, fmt.Errorf("owner public key: %w", err)
	}
	if strings.TrimSpace(m.cfg.Persona.AccountQR) == "" {
		return nil, fmt.Errorf("PERSONA_ACCOUNT_QR / persona.account_qr is empty")
	}

	n, err := m.store.CountGhostAccounts()
	if err != nil {
		return nil, err
	}
	if n >= m.cfg.Persona.MaxGhosts {
		return nil, fmt.Errorf("ghost account limit reached (%d)", m.cfg.Persona.MaxGhosts)
	}
	perBot, err := m.store.CountGhostAccountsByBot(bot.ID)
	if err != nil {
		return nil, err
	}
	if perBot >= m.cfg.Persona.MaxGhostsPerBot {
		return nil, fmt.Errorf("ghost account limit reached for this bot (%d)", m.cfg.Persona.MaxGhostsPerBot)
	}

	if displayName == "" {
		if username != "" {
			displayName = username
		} else {
			displayName = fmt.Sprintf("TG %d", tgUserID)
		}
	}

	var g *store.GhostAccount
	accID, err := withNewGhostAccount(m.dc, func(accID uint32) error {
		if err := m.dc.ConfigureAccountFromQR(accID, m.cfg.Persona.AccountQR); err != nil {
			return fmt.Errorf("configure ghost: %w", err)
		}
		_ = m.dc.SetDisplayName(accID, displayName)

		var avatarFileID string
		if av != nil {
			if path, fid, err := av.DownloadProfilePhoto(tgUserID); err == nil && path != "" {
				if err := m.dc.SetSelfAvatar(accID, path); err != nil {
					m.log.Warnf("ghost avatar: %v", err)
				} else {
					avatarFileID = fid
				}
				_ = os.Remove(path)
			}
		}

		chatID, _, err := m.dc.ImportOwnerAsKeyContact(accID, bot.OwnerVcard)
		if err != nil {
			return fmt.Errorf("import owner key: %w", err)
		}
		addr, err := m.dc.AccountAddress(accID)
		if err != nil {
			m.log.Warnf("ghost acc %d address: %v", accID, err)
			addr = ""
		}
		created, err := m.store.InsertGhostAccount(&store.GhostAccount{
			PersonaBotID:     bot.ID,
			TelegramUserID:   tgUserID,
			TelegramUsername: username,
			DisplayName:      displayName,
			DCAccountID:      accID,
			DCAddress:        addr,
			OwnerChatID:      chatID,
			AvatarFileID:     avatarFileID,
		})
		if err != nil {
			return err
		}
		g = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	m.log.Infof("created ghost dc_account=%d for tg_user=%d persona_bot=%d owner_chat=%d",
		accID, tgUserID, bot.ID, g.OwnerChatID)
	return g, nil
}

// ensureOwnerChat imports the owner's vcard into the ghost and opens/refreshes the 1:1 chat.
// Always re-imports so keys are present (idempotent).
func (m *Manager) ensureOwnerChat(ghost *store.GhostAccount, bot *store.PersonaBot) (uint32, error) {
	// Never replace a known owner 1:1 with "first encrypted chat" (wrong recipient).
	if ghost.OwnerChatID != 0 {
		return ghost.OwnerChatID, nil
	}
	vcard, err := m.EnsureOwnerVcard(bot)
	if err != nil {
		return 0, err
	}
	if chatID, err := m.dc.FindEncryptedDirectChat(ghost.DCAccountID); err == nil && chatID != 0 {
		_ = m.store.UpdateGhostOwnerChat(ghost.ID, chatID)
		ghost.OwnerChatID = chatID
		return chatID, nil
	}
	chatID, _, err := m.dc.ImportOwnerAsKeyContact(ghost.DCAccountID, vcard)
	if err != nil {
		return 0, fmt.Errorf("import owner key: %w", err)
	}
	// After import, prefer encrypted chat if core created one.
	if enc, err := m.dc.FindEncryptedDirectChat(ghost.DCAccountID); err == nil && enc != 0 {
		chatID = enc
	}
	_ = m.store.UpdateGhostOwnerChat(ghost.ID, chatID)
	ghost.OwnerChatID = chatID
	return chatID, nil
}

func (m *Manager) maybeSyncAvatar(ghost *store.GhostAccount, av AvatarDownloader) {
	path, fid, err := av.DownloadProfilePhoto(ghost.TelegramUserID)
	if err != nil || path == "" {
		return
	}
	defer os.Remove(path)
	if fid != "" && fid == ghost.AvatarFileID {
		return
	}
	if err := m.dc.SetSelfAvatar(ghost.DCAccountID, path); err != nil {
		m.log.Warnf("resync avatar: %v", err)
		return
	}
	_ = m.store.UpdateGhostAvatar(ghost.ID, fid)
	ghost.AvatarFileID = fid
}

// RegisterBot validates caps and persists a new persona bot, then starts its poller.
func (m *Manager) RegisterBot(b *store.PersonaBot, startPoller func(bot store.PersonaBot) (*BotRuntime, error)) (*store.PersonaBot, error) {
	if !m.cfg.PersonaEnabled() {
		return nil, fmt.Errorf("persona mode is disabled (set mode: persona or both)")
	}
	if strings.TrimSpace(m.cfg.Persona.AccountQR) == "" {
		return nil, fmt.Errorf("PERSONA_ACCOUNT_QR / persona.account_qr is empty")
	}
	if strings.TrimSpace(b.OwnerVcard) == "" {
		return nil, fmt.Errorf("owner DC key not available — complete /pair on the portal bot first (re-pair if you paired before this feature)")
	}
	if IsPortalBotToken(b.BotToken, m.cfg.TelegramToken) {
		return nil, fmt.Errorf("that token is the portal bot — create a new bot with @BotFather")
	}
	n, err := m.store.CountPersonaBots(true)
	if err != nil {
		return nil, err
	}
	if n >= m.cfg.Persona.MaxBots {
		return nil, fmt.Errorf("persona bot limit reached (%d)", m.cfg.Persona.MaxBots)
	}
	owned, err := m.store.CountPersonaBotsByOwner(b.OwnerTelegramUserID, true)
	if err != nil {
		return nil, err
	}
	if owned >= m.cfg.Persona.MaxBotsPerOwner {
		return nil, fmt.Errorf("persona bot limit reached for this owner (%d)", m.cfg.Persona.MaxBotsPerOwner)
	}
	if existing, _ := m.store.GetPersonaBotByBotUserID(b.BotUserID); existing != nil {
		if existing.Status == store.PersonaBotActive {
			return nil, fmt.Errorf("this Telegram bot is already registered")
		}
		if existing.OwnerTelegramUserID != b.OwnerTelegramUserID {
			return nil, fmt.Errorf("this Telegram bot is already registered")
		}
		saved, err := m.store.ReactivatePersonaBot(existing.ID, b)
		if err != nil {
			return nil, err
		}
		if err := m.AttachBot(*saved, startPoller); err != nil {
			_ = m.store.SetPersonaBotStatus(saved.ID, store.PersonaBotError)
			return saved, err
		}
		return saved, nil
	}
	saved, err := m.store.InsertPersonaBot(b)
	if err != nil {
		return nil, err
	}
	if err := m.AttachBot(*saved, startPoller); err != nil {
		_ = m.store.SetPersonaBotStatus(saved.ID, store.PersonaBotError)
		return saved, err
	}
	return saved, nil
}

// HandleDCMessage routes messages arriving on ghost accounts back to Telegram.
// fromID is the Delta Chat contact that sent the message on this ghost account.
//
// Only messages from the **persona bot owner** (the paired human) are sent to Telegram.
// Messages from other ghosts or third parties are ignored (prevents echo loops in groups).
func (m *Manager) HandleDCMessage(accID, chatID, fromID uint32, text string, hasFile bool, filePath, fileName string, sendTG func(botID int64, tgChatID int64, text, path, name string) error) (bool, error) {
	ghost, err := m.store.GetGhostByDCAccount(accID)
	if err != nil || ghost == nil {
		return false, err
	}
	bot, err := m.store.GetPersonaBot(ghost.PersonaBotID)
	if err != nil || bot == nil || bot.Status != store.PersonaBotActive {
		return true, fmt.Errorf("persona bot missing for ghost")
	}

	// Only the owner of the persona bot may cause outbound Telegram traffic.
	if !m.isOwnerSender(accID, ghost, bot, fromID) {
		m.log.Debugf("ignore non-owner DC sender from=%d acc=%d chat=%d", fromID, accID, chatID)
		return true, nil
	}

	// Group chat on this ghost? Only the coordinator account posts to TG (avoid N duplicates).
	if gg, _ := m.store.GetGhostGroupByMemberDCChat(chatID, ghost.ID); gg != nil {
		if ghost.DCAccountID != gg.CoordinatorDCAccountID {
			return true, nil
		}
		return true, m.sendOwnerToTG(bot.ID, gg.TelegramChatID, text, hasFile, filePath, fileName, sendTG)
	}
	// Also match coordinator chat id if member row missing.
	if gg, _ := m.store.GetGhostGroupByDCChat(accID, chatID); gg != nil {
		return true, m.sendOwnerToTG(bot.ID, gg.TelegramChatID, text, hasFile, filePath, fileName, sendTG)
	}

	// 1:1 only when this is the stored owner chat — do not DM on unknown groups.
	if !isOwnerDirectChat(ghost.OwnerChatID, chatID) {
		m.log.Debugf("ignore owner DC message on unmatched chat=%d ghost=%d", chatID, ghost.ID)
		return true, nil
	}
	return true, m.sendOwnerToTG(bot.ID, ghost.TelegramUserID, text, hasFile, filePath, fileName, sendTG)
}

func (m *Manager) sendOwnerToTG(botID int64, tgChatID int64, text string, hasFile bool, filePath, fileName string, sendTG func(botID int64, tgChatID int64, text, path, name string) error) error {
	if hasFile && filePath != "" {
		role := safemedia.RoleFromKind(string(bridge.KindFromFilename(fileName)))
		if role == safemedia.RoleFile || fileName == "" {
			role = safemedia.RoleFile
		}
		if err := safemedia.ValidateFile(filePath, role, safemedia.DefaultMaxBytes); err != nil {
			return err
		}
		return sendTG(botID, tgChatID, text, filePath, fileName)
	}
	if text != "" {
		return sendTG(botID, tgChatID, text, "", "")
	}
	return nil
}

// isOwnerSender reports whether fromID on this ghost account is the persona bot owner.
// The owner is the peer of the ghost's 1:1 owner chat (imported via owner vcard).
func (m *Manager) isOwnerSender(accID uint32, ghost *store.GhostAccount, bot *store.PersonaBot, fromID uint32) bool {
	if fromID == 0 {
		return false
	}
	// Ensure we have an owner 1:1 chat so we can resolve the owner contact id.
	ownerChat := ghost.OwnerChatID
	if ownerChat == 0 {
		id, err := m.ensureOwnerChat(ghost, bot)
		if err != nil {
			m.log.Warnf("isOwnerSender ensure owner chat: %v", err)
			return false
		}
		ownerChat = id
	}
	ownerContactID, err := m.dc.FirstPeerContactID(accID, ownerChat)
	if err != nil {
		m.log.Warnf("isOwnerSender owner contact: %v", err)
		return false
	}
	return fromID == ownerContactID
}

// SendFromGhost delivers DC→TG via the persona bot API.
func (m *Manager) SendFromGhost(botID int64, tgChatID int64, text, path, name string) error {
	rt := m.Runtime(botID)
	if rt == nil || rt.API == nil {
		bot, err := m.store.GetPersonaBot(botID)
		if err != nil || bot == nil {
			return fmt.Errorf("persona bot %d not running", botID)
		}
		client, err := NewTGClient(bot.BotToken, m.cfg, m.log)
		if err != nil {
			return err
		}
		return sendVia(client, tgChatID, text, path, name)
	}
	return sendVia(rt.API, tgChatID, text, path, name)
}
