package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	PersonaBotActive   = "active"
	PersonaBotDisabled = "disabled"
	PersonaBotError    = "error"
)

// PersonaBot is a user-owned Telegram bot registered via /pair-bot.
type PersonaBot struct {
	ID                  int64
	OwnerTelegramUserID int64
	OwnerDCAccountID    uint32
	OwnerDCChatID       uint32
	OwnerDCAddress      string
	// OwnerVcard is the owner's contact+public key (from mode-1 pairing), imported into each ghost.
	OwnerVcard  string
	BotToken    string
	BotUserID   int64
	BotUsername string
	BotURL      string
	Status      string
	CreatedAt   time.Time
}

// GhostAccount is a Delta Chat account bound to one Telegram user for a persona bot.
type GhostAccount struct {
	ID               int64
	PersonaBotID     int64
	TelegramUserID   int64
	TelegramUsername string
	DisplayName      string
	DCAccountID      uint32
	DCAddress        string
	// OwnerChatID is the 1:1 chat with the persona owner on this ghost account (via key-contact import).
	OwnerChatID uint32
	// AvatarFileID is the last Telegram profile photo file_id we applied (empty if none).
	AvatarFileID string
	CreatedAt    time.Time
}

// GhostGroup mirrors a Telegram group as a Delta Chat group.
type GhostGroup struct {
	ID             int64
	PersonaBotID   int64
	TelegramChatID int64
	Title          string
	// CoordinatorDCAccountID is the ghost account that created the DC group.
	CoordinatorDCAccountID uint32
	// DCChatID is the group chat id on the coordinator account.
	DCChatID  uint32
	InviteQR  string
	CreatedAt time.Time
}

// GhostGroupMember links a TG user ghost into a mirrored group.
type GhostGroupMember struct {
	GhostGroupID   int64
	TelegramUserID int64
	GhostAccountID int64
	// MemberDCChatID is the group chat id as seen from this member's account (0 if unknown).
	MemberDCChatID uint32
	JoinedAt       time.Time
}

func (s *Store) migratePersona() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS persona_bots (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  owner_telegram_user_id INTEGER NOT NULL,
  owner_dc_account_id INTEGER NOT NULL DEFAULT 0,
  owner_dc_chat_id INTEGER NOT NULL DEFAULT 0,
  owner_dc_address TEXT NOT NULL DEFAULT '',
  owner_vcard TEXT NOT NULL DEFAULT '',
  bot_token TEXT NOT NULL,
  bot_user_id INTEGER NOT NULL DEFAULT 0,
  bot_username TEXT NOT NULL DEFAULT '',
  bot_url TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  created_at INTEGER NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_persona_bots_bot_user ON persona_bots(bot_user_id) WHERE bot_user_id != 0;
CREATE INDEX IF NOT EXISTS idx_persona_bots_owner ON persona_bots(owner_telegram_user_id);
CREATE INDEX IF NOT EXISTS idx_persona_bots_status ON persona_bots(status);

CREATE TABLE IF NOT EXISTS ghost_accounts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  persona_bot_id INTEGER NOT NULL,
  telegram_user_id INTEGER NOT NULL,
  telegram_username TEXT NOT NULL DEFAULT '',
  display_name TEXT NOT NULL DEFAULT '',
  dc_account_id INTEGER NOT NULL,
  dc_address TEXT NOT NULL DEFAULT '',
  owner_chat_id INTEGER NOT NULL DEFAULT 0,
  avatar_file_id TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL,
  UNIQUE(persona_bot_id, telegram_user_id)
);
CREATE INDEX IF NOT EXISTS idx_ghost_dc_account ON ghost_accounts(dc_account_id);
CREATE INDEX IF NOT EXISTS idx_ghost_persona ON ghost_accounts(persona_bot_id);

CREATE TABLE IF NOT EXISTS ghost_groups (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  persona_bot_id INTEGER NOT NULL,
  telegram_chat_id INTEGER NOT NULL,
  title TEXT NOT NULL DEFAULT '',
  coordinator_dc_account_id INTEGER NOT NULL DEFAULT 0,
  dc_chat_id INTEGER NOT NULL DEFAULT 0,
  invite_qr TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL,
  UNIQUE(persona_bot_id, telegram_chat_id)
);

CREATE TABLE IF NOT EXISTS ghost_group_members (
  ghost_group_id INTEGER NOT NULL,
  telegram_user_id INTEGER NOT NULL,
  ghost_account_id INTEGER NOT NULL,
  member_dc_chat_id INTEGER NOT NULL DEFAULT 0,
  joined_at INTEGER NOT NULL,
  PRIMARY KEY (ghost_group_id, telegram_user_id)
);
`)
	if err != nil {
		return err
	}
	// Additive migrations for DBs created before these columns existed.
	for _, q := range []string{
		`ALTER TABLE persona_bots ADD COLUMN owner_vcard TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE persona_bots ADD COLUMN owner_invite_url TEXT NOT NULL DEFAULT ''`, // legacy; unused
		`ALTER TABLE ghost_accounts ADD COLUMN owner_chat_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE ghost_accounts ADD COLUMN avatar_file_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE ghost_accounts ADD COLUMN invite_qr TEXT NOT NULL DEFAULT ''`,         // legacy
		`ALTER TABLE ghost_accounts ADD COLUMN invite_notified INTEGER NOT NULL DEFAULT 0`, // legacy
	} {
		_, _ = s.db.Exec(q) // ignore "duplicate column"
	}
	return nil
}

// InsertPersonaBot registers a new active persona bot.
func (s *Store) InsertPersonaBot(b *PersonaBot) (*PersonaBot, error) {
	now := time.Now().Unix()
	if b.Status == "" {
		b.Status = PersonaBotActive
	}
	vcard, err := s.seal(b.OwnerVcard)
	if err != nil {
		return nil, err
	}
	tok, err := s.seal(b.BotToken)
	if err != nil {
		return nil, err
	}
	res, err := s.db.Exec(
		`INSERT INTO persona_bots (
		  owner_telegram_user_id, owner_dc_account_id, owner_dc_chat_id, owner_dc_address,
		  owner_vcard, bot_token, bot_user_id, bot_username, bot_url, status, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.OwnerTelegramUserID, b.OwnerDCAccountID, b.OwnerDCChatID, b.OwnerDCAddress,
		vcard, tok, b.BotUserID, b.BotUsername, b.BotURL, b.Status, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return nil, fmt.Errorf("this Telegram bot is already registered")
		}
		return nil, err
	}
	id, _ := res.LastInsertId()
	b.ID = id
	b.CreatedAt = time.Unix(now, 0)
	return b, nil
}

// CountPersonaBots returns the number of bots (any status) or only active when activeOnly.
func (s *Store) CountPersonaBots(activeOnly bool) (int, error) {
	q := `SELECT COUNT(*) FROM persona_bots`
	if activeOnly {
		q += ` WHERE status = 'active'`
	}
	var n int
	err := s.db.QueryRow(q).Scan(&n)
	return n, err
}

// CountGhostAccounts returns total ghost rows.
func (s *Store) CountGhostAccounts() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM ghost_accounts`).Scan(&n)
	return n, err
}

// CountPersonaBotsByOwner counts bots for one owner (any status when activeOnly is false).
func (s *Store) CountPersonaBotsByOwner(ownerTG int64, activeOnly bool) (int, error) {
	q := `SELECT COUNT(*) FROM persona_bots WHERE owner_telegram_user_id = ?`
	if activeOnly {
		q += ` AND status = 'active'`
	}
	var n int
	err := s.db.QueryRow(q, ownerTG).Scan(&n)
	return n, err
}

// CountGhostAccountsByBot counts ghosts bound to one persona bot.
func (s *Store) CountGhostAccountsByBot(personaBotID int64) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM ghost_accounts WHERE persona_bot_id = ?`, personaBotID).Scan(&n)
	return n, err
}

// ListActivePersonaBots returns all active persona bots.
func (s *Store) ListActivePersonaBots() ([]PersonaBot, error) {
	rows, err := s.db.Query(
		`SELECT id, owner_telegram_user_id, owner_dc_account_id, owner_dc_chat_id, owner_dc_address,
		        owner_vcard, bot_token, bot_user_id, bot_username, bot_url, status, created_at
		 FROM persona_bots WHERE status = ? ORDER BY id`,
		PersonaBotActive,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanPersonaBots(rows)
}

// ListPersonaBotsByOwnerAny returns all persona bots for any owner (all statuses).
func (s *Store) ListPersonaBotsByOwnerAny() ([]PersonaBot, error) {
	rows, err := s.db.Query(
		`SELECT id, owner_telegram_user_id, owner_dc_account_id, owner_dc_chat_id, owner_dc_address,
		        owner_vcard, bot_token, bot_user_id, bot_username, bot_url, status, created_at
		 FROM persona_bots ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanPersonaBots(rows)
}

// ListPersonaBotsByOwner returns bots for an owner TG user.
func (s *Store) ListPersonaBotsByOwner(ownerTG int64) ([]PersonaBot, error) {
	rows, err := s.db.Query(
		`SELECT id, owner_telegram_user_id, owner_dc_account_id, owner_dc_chat_id, owner_dc_address,
		        owner_vcard, bot_token, bot_user_id, bot_username, bot_url, status, created_at
		 FROM persona_bots WHERE owner_telegram_user_id = ? ORDER BY id`,
		ownerTG,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanPersonaBots(rows)
}

// GetPersonaBot returns a bot by id.
func (s *Store) GetPersonaBot(id int64) (*PersonaBot, error) {
	row := s.db.QueryRow(
		`SELECT id, owner_telegram_user_id, owner_dc_account_id, owner_dc_chat_id, owner_dc_address,
		        owner_vcard, bot_token, bot_user_id, bot_username, bot_url, status, created_at
		 FROM persona_bots WHERE id = ?`, id,
	)
	return s.scanPersonaBot(row)
}

// GetPersonaBotByBotUserID looks up by Telegram bot user id.
func (s *Store) GetPersonaBotByBotUserID(botUserID int64) (*PersonaBot, error) {
	row := s.db.QueryRow(
		`SELECT id, owner_telegram_user_id, owner_dc_account_id, owner_dc_chat_id, owner_dc_address,
		        owner_vcard, bot_token, bot_user_id, bot_username, bot_url, status, created_at
		 FROM persona_bots WHERE bot_user_id = ? LIMIT 1`, botUserID,
	)
	return s.scanPersonaBot(row)
}

// SetPersonaBotStatus updates status (active/disabled/error).
func (s *Store) SetPersonaBotStatus(id int64, status string) error {
	_, err := s.db.Exec(`UPDATE persona_bots SET status = ? WHERE id = ?`, status, id)
	return err
}

// ReactivatePersonaBot updates registration fields on an existing row and sets status active.
// Used when the owner /pair-bot again after /unpair-bot (same Telegram bot user id).
func (s *Store) ReactivatePersonaBot(id int64, b *PersonaBot) (*PersonaBot, error) {
	vcard, err := s.seal(b.OwnerVcard)
	if err != nil {
		return nil, err
	}
	tok, err := s.seal(b.BotToken)
	if err != nil {
		return nil, err
	}
	if b.Status == "" {
		b.Status = PersonaBotActive
	}
	_, err = s.db.Exec(
		`UPDATE persona_bots SET
		   owner_telegram_user_id = ?, owner_dc_account_id = ?, owner_dc_chat_id = ?,
		   owner_dc_address = ?, owner_vcard = ?, bot_token = ?, bot_username = ?,
		   bot_url = ?, status = ?
		 WHERE id = ?`,
		b.OwnerTelegramUserID, b.OwnerDCAccountID, b.OwnerDCChatID,
		b.OwnerDCAddress, vcard, tok, b.BotUsername,
		b.BotURL, PersonaBotActive, id,
	)
	if err != nil {
		return nil, err
	}
	return s.GetPersonaBot(id)
}

// UpdatePersonaOwnerVcard stores the owner's vcard/public key on a persona bot.
func (s *Store) UpdatePersonaOwnerVcard(id int64, vcard string) error {
	stored, err := s.seal(vcard)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE persona_bots SET owner_vcard = ? WHERE id = ?`, stored, id)
	return err
}

// UpdatePersonaOwnerVcardForOwner updates all bots owned by a TG user.
func (s *Store) UpdatePersonaOwnerVcardForOwner(ownerTG int64, vcard string) (int64, error) {
	stored, err := s.seal(vcard)
	if err != nil {
		return 0, err
	}
	res, err := s.db.Exec(
		`UPDATE persona_bots SET owner_vcard = ? WHERE owner_telegram_user_id = ?`,
		stored, ownerTG,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DisablePersonaBotByOwner disables a bot owned by ownerTG (by id or username).
func (s *Store) DisablePersonaBotByOwner(ownerTG int64, botID int64, username string) (int64, error) {
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	if botID > 0 {
		res, err := s.db.Exec(
			`UPDATE persona_bots SET status = ? WHERE id = ? AND owner_telegram_user_id = ?`,
			PersonaBotDisabled, botID, ownerTG,
		)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	}
	if username != "" {
		res, err := s.db.Exec(
			`UPDATE persona_bots SET status = ? WHERE owner_telegram_user_id = ? AND lower(bot_username) = lower(?)`,
			PersonaBotDisabled, ownerTG, username,
		)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	}
	// Disable all active for owner
	res, err := s.db.Exec(
		`UPDATE persona_bots SET status = ? WHERE owner_telegram_user_id = ? AND status = ?`,
		PersonaBotDisabled, ownerTG, PersonaBotActive,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

const ghostSelect = `id, persona_bot_id, telegram_user_id, telegram_username, display_name,
		        dc_account_id, dc_address, owner_chat_id, avatar_file_id, created_at`

// ListGhostsByPersonaBot returns ghosts bound to a persona bot.
func (s *Store) ListGhostsByPersonaBot(personaBotID int64) ([]GhostAccount, error) {
	rows, err := s.db.Query(
		`SELECT `+ghostSelect+` FROM ghost_accounts WHERE persona_bot_id = ?`,
		personaBotID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGhosts(rows)
}

// ListGhostsByTelegramUser returns ghost rows for a Telegram user (any persona bot).
func (s *Store) ListGhostsByTelegramUser(tgUserID int64) ([]GhostAccount, error) {
	rows, err := s.db.Query(
		`SELECT `+ghostSelect+` FROM ghost_accounts WHERE telegram_user_id = ?`,
		tgUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGhosts(rows)
}

// GhostAccountsForPurge lists unique ghost accounts to remove for a Telegram user:
// ghosts they own via persona bots, plus ghosts that represent them on others' bots.
func (s *Store) GhostAccountsForPurge(tgUserID int64) ([]GhostAccount, error) {
	seen := map[int64]struct{}{}
	var out []GhostAccount
	add := func(list []GhostAccount) {
		for _, g := range list {
			if _, ok := seen[g.ID]; ok {
				continue
			}
			seen[g.ID] = struct{}{}
			out = append(out, g)
		}
	}
	bots, err := s.ListPersonaBotsByOwner(tgUserID)
	if err != nil {
		return nil, err
	}
	for _, b := range bots {
		gs, err := s.ListGhostsByPersonaBot(b.ID)
		if err != nil {
			return nil, err
		}
		add(gs)
	}
	own, err := s.ListGhostsByTelegramUser(tgUserID)
	if err != nil {
		return nil, err
	}
	add(own)
	return out, nil
}

// PurgeTelegramUser deletes pairing, persona bots, ghosts, and group rows for a Telegram user.
func (s *Store) PurgeTelegramUser(tgUserID int64) error {
	bots, err := s.ListPersonaBotsByOwner(tgUserID)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, b := range bots {
		if _, err := tx.Exec(
			`DELETE FROM ghost_group_members WHERE ghost_group_id IN (SELECT id FROM ghost_groups WHERE persona_bot_id = ?)`,
			b.ID,
		); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM ghost_groups WHERE persona_bot_id = ?`, b.ID); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM ghost_accounts WHERE persona_bot_id = ?`, b.ID); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM persona_bots WHERE id = ?`, b.ID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM ghost_group_members WHERE telegram_user_id = ?`, tgUserID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM ghost_accounts WHERE telegram_user_id = ?`, tgUserID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM pairs WHERE telegram_user_id = ?`, tgUserID); err != nil {
		return err
	}
	return tx.Commit()
}

// PurgeDCChatPairs deletes pair rows for a Delta Chat conversation.
func (s *Store) PurgeDCChatPairs(accID, chatID uint32) error {
	_, err := s.db.Exec(
		`DELETE FROM pairs WHERE dc_account_id = ? AND dc_chat_id = ?`,
		accID, chatID,
	)
	return err
}

// GetGhostByTG returns the ghost for (persona bot, telegram user).
func (s *Store) GetGhostByTG(personaBotID, tgUserID int64) (*GhostAccount, error) {
	row := s.db.QueryRow(
		`SELECT id, persona_bot_id, telegram_user_id, telegram_username, display_name,
		        dc_account_id, dc_address, owner_chat_id, avatar_file_id, created_at
		 FROM ghost_accounts WHERE persona_bot_id = ? AND telegram_user_id = ?`,
		personaBotID, tgUserID,
	)
	return scanGhost(row)
}

// GetGhostByDCAccount looks up a ghost by its Delta Chat account id.
func (s *Store) GetGhostByDCAccount(dcAccountID uint32) (*GhostAccount, error) {
	row := s.db.QueryRow(
		`SELECT id, persona_bot_id, telegram_user_id, telegram_username, display_name,
		        dc_account_id, dc_address, owner_chat_id, avatar_file_id, created_at
		 FROM ghost_accounts WHERE dc_account_id = ? LIMIT 1`,
		dcAccountID,
	)
	return scanGhost(row)
}

// InsertGhostAccount stores a new ghost binding.
func (s *Store) InsertGhostAccount(g *GhostAccount) (*GhostAccount, error) {
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`INSERT INTO ghost_accounts (
		  persona_bot_id, telegram_user_id, telegram_username, display_name,
		  dc_account_id, dc_address, owner_chat_id, avatar_file_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		g.PersonaBotID, g.TelegramUserID, g.TelegramUsername, g.DisplayName,
		g.DCAccountID, g.DCAddress, g.OwnerChatID, g.AvatarFileID, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return s.GetGhostByTG(g.PersonaBotID, g.TelegramUserID)
		}
		return nil, err
	}
	id, _ := res.LastInsertId()
	g.ID = id
	g.CreatedAt = time.Unix(now, 0)
	return g, nil
}

// UpdateGhostProfile updates display name / username for an existing ghost.
func (s *Store) UpdateGhostProfile(id int64, username, displayName string) error {
	_, err := s.db.Exec(
		`UPDATE ghost_accounts SET telegram_username = ?, display_name = ? WHERE id = ?`,
		username, displayName, id,
	)
	return err
}

// UpdateGhostOwnerChat stores the 1:1 chat id with the owner on the ghost account.
func (s *Store) UpdateGhostOwnerChat(id int64, ownerChatID uint32) error {
	_, err := s.db.Exec(`UPDATE ghost_accounts SET owner_chat_id = ? WHERE id = ?`, ownerChatID, id)
	return err
}

// UpdateGhostAvatar records the last Telegram profile photo file_id applied.
func (s *Store) UpdateGhostAvatar(id int64, fileID string) error {
	_, err := s.db.Exec(`UPDATE ghost_accounts SET avatar_file_id = ? WHERE id = ?`, fileID, id)
	return err
}

// GetGhostGroup returns mirrored group for (bot, tg chat).
func (s *Store) GetGhostGroup(personaBotID, tgChatID int64) (*GhostGroup, error) {
	row := s.db.QueryRow(
		`SELECT id, persona_bot_id, telegram_chat_id, title, coordinator_dc_account_id,
		        dc_chat_id, invite_qr, created_at
		 FROM ghost_groups WHERE persona_bot_id = ? AND telegram_chat_id = ?`,
		personaBotID, tgChatID,
	)
	return s.scanGhostGroup(row)
}

// GetGhostGroupByDCChat finds a group by coordinator account + chat id.
func (s *Store) GetGhostGroupByDCChat(coordAcc, dcChatID uint32) (*GhostGroup, error) {
	row := s.db.QueryRow(
		`SELECT id, persona_bot_id, telegram_chat_id, title, coordinator_dc_account_id,
		        dc_chat_id, invite_qr, created_at
		 FROM ghost_groups WHERE coordinator_dc_account_id = ? AND dc_chat_id = ? LIMIT 1`,
		coordAcc, dcChatID,
	)
	return s.scanGhostGroup(row)
}

// GetGhostGroupByMemberDCChat finds a group when a member account received a message.
func (s *Store) GetGhostGroupByMemberDCChat(memberDCChatID uint32, ghostAccountID int64) (*GhostGroup, error) {
	row := s.db.QueryRow(
		`SELECT g.id, g.persona_bot_id, g.telegram_chat_id, g.title, g.coordinator_dc_account_id,
		        g.dc_chat_id, g.invite_qr, g.created_at
		 FROM ghost_groups g
		 JOIN ghost_group_members m ON m.ghost_group_id = g.id
		 WHERE m.member_dc_chat_id = ? AND m.ghost_account_id = ? LIMIT 1`,
		memberDCChatID, ghostAccountID,
	)
	return s.scanGhostGroup(row)
}

// InsertGhostGroup creates a mirrored group row.
func (s *Store) InsertGhostGroup(g *GhostGroup) (*GhostGroup, error) {
	now := time.Now().Unix()
	inviteQR, err := s.seal(g.InviteQR)
	if err != nil {
		return nil, err
	}
	res, err := s.db.Exec(
		`INSERT INTO ghost_groups (
		  persona_bot_id, telegram_chat_id, title, coordinator_dc_account_id,
		  dc_chat_id, invite_qr, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		g.PersonaBotID, g.TelegramChatID, g.Title, g.CoordinatorDCAccountID,
		g.DCChatID, inviteQR, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return s.GetGhostGroup(g.PersonaBotID, g.TelegramChatID)
		}
		return nil, err
	}
	id, _ := res.LastInsertId()
	g.ID = id
	g.CreatedAt = time.Unix(now, 0)
	return g, nil
}

// UpdateGhostGroupMeta updates title / invite / coordinator chat ids.
func (s *Store) UpdateGhostGroupMeta(id int64, title, inviteQR string, coordAcc, dcChatID uint32) error {
	stored, err := s.seal(inviteQR)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`UPDATE ghost_groups SET title = ?, invite_qr = ?, coordinator_dc_account_id = ?, dc_chat_id = ? WHERE id = ?`,
		title, stored, coordAcc, dcChatID, id,
	)
	return err
}

// UpsertGhostGroupMember records membership; updates member_dc_chat_id when non-zero.
func (s *Store) UpsertGhostGroupMember(m *GhostGroupMember) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT INTO ghost_group_members (ghost_group_id, telegram_user_id, ghost_account_id, member_dc_chat_id, joined_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(ghost_group_id, telegram_user_id) DO UPDATE SET
		   ghost_account_id = excluded.ghost_account_id,
		   member_dc_chat_id = CASE WHEN excluded.member_dc_chat_id != 0 THEN excluded.member_dc_chat_id ELSE ghost_group_members.member_dc_chat_id END`,
		m.GhostGroupID, m.TelegramUserID, m.GhostAccountID, m.MemberDCChatID, now,
	)
	return err
}

// GetGhostGroupMember returns one member row.
func (s *Store) GetGhostGroupMember(groupID, tgUserID int64) (*GhostGroupMember, error) {
	row := s.db.QueryRow(
		`SELECT ghost_group_id, telegram_user_id, ghost_account_id, member_dc_chat_id, joined_at
		 FROM ghost_group_members WHERE ghost_group_id = ? AND telegram_user_id = ?`,
		groupID, tgUserID,
	)
	var m GhostGroupMember
	var joined int64
	err := row.Scan(&m.GhostGroupID, &m.TelegramUserID, &m.GhostAccountID, &m.MemberDCChatID, &joined)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	m.JoinedAt = time.Unix(joined, 0)
	return &m, nil
}

func (s *Store) scanPersonaBots(rows *sql.Rows) ([]PersonaBot, error) {
	var out []PersonaBot
	for rows.Next() {
		var b PersonaBot
		var created int64
		if err := rows.Scan(
			&b.ID, &b.OwnerTelegramUserID, &b.OwnerDCAccountID, &b.OwnerDCChatID, &b.OwnerDCAddress,
			&b.OwnerVcard, &b.BotToken, &b.BotUserID, &b.BotUsername, &b.BotURL, &b.Status, &created,
		); err != nil {
			return nil, err
		}
		if err := s.unsealPersona(&b); err != nil {
			return nil, err
		}
		b.CreatedAt = time.Unix(created, 0)
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) scanPersonaBot(row *sql.Row) (*PersonaBot, error) {
	var b PersonaBot
	var created int64
	err := row.Scan(
		&b.ID, &b.OwnerTelegramUserID, &b.OwnerDCAccountID, &b.OwnerDCChatID, &b.OwnerDCAddress,
		&b.OwnerVcard, &b.BotToken, &b.BotUserID, &b.BotUsername, &b.BotURL, &b.Status, &created,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := s.unsealPersona(&b); err != nil {
		return nil, err
	}
	b.CreatedAt = time.Unix(created, 0)
	return &b, nil
}

func (s *Store) unsealPersona(b *PersonaBot) error {
	tok, err := s.unseal(b.BotToken)
	if err != nil {
		return err
	}
	vc, err := s.unseal(b.OwnerVcard)
	if err != nil {
		return err
	}
	b.BotToken = tok
	b.OwnerVcard = vc
	return nil
}

func scanGhost(row *sql.Row) (*GhostAccount, error) {
	return scanGhostScanner(row)
}

func scanGhosts(rows *sql.Rows) ([]GhostAccount, error) {
	var out []GhostAccount
	for rows.Next() {
		g, err := scanGhostScanner(rows)
		if err != nil {
			return nil, err
		}
		if g != nil {
			out = append(out, *g)
		}
	}
	return out, rows.Err()
}

func scanGhostScanner(row pairScanner) (*GhostAccount, error) {
	var g GhostAccount
	var created int64
	err := row.Scan(
		&g.ID, &g.PersonaBotID, &g.TelegramUserID, &g.TelegramUsername, &g.DisplayName,
		&g.DCAccountID, &g.DCAddress, &g.OwnerChatID, &g.AvatarFileID, &created,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	g.CreatedAt = time.Unix(created, 0)
	return &g, nil
}

func (s *Store) scanGhostGroup(row *sql.Row) (*GhostGroup, error) {
	var g GhostGroup
	var created int64
	err := row.Scan(
		&g.ID, &g.PersonaBotID, &g.TelegramChatID, &g.Title, &g.CoordinatorDCAccountID,
		&g.DCChatID, &g.InviteQR, &created,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	qr, err := s.unseal(g.InviteQR)
	if err != nil {
		return nil, err
	}
	g.InviteQR = qr
	g.CreatedAt = time.Unix(created, 0)
	return &g, nil
}
