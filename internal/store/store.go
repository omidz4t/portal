package store

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	StatusPending = "pending"
	StatusActive  = "active"

	DefaultCodeLength    = 8
	DefaultPendingTTLSec = 1800
)

// Pair links a Telegram user to a Delta Chat conversation.
type Pair struct {
	ID               int64
	Code             string
	TelegramUserID   int64
	TelegramUsername string
	TelegramChatID   int64
	DCAccountID      uint32
	DCChatID         uint32
	// OwnerVcard is the peer's vCard (with public key) captured at pairing for persona ghosts.
	OwnerVcard string
	Status     string
	CreatedAt  time.Time
	PairedAt   *time.Time
}

// Options configure store Open.
type Options struct {
	PendingTTL      time.Duration
	CodeLength      int
	Key             []byte
	EncryptRequired bool
}

// Store is a SQLite-backed pairing database.
type Store struct {
	db         *sql.DB
	crypt      *crypter
	pendingTTL time.Duration
	codeLen    int
}

// Open opens (or creates) the SQLite database at path and migrates schema.
func Open(path string) (*Store, error) {
	return OpenOpts(path, Options{})
}

// OpenOpts opens the store with pairing TTL, code length, and optional encryption.
func OpenOpts(path string, opts Options) (*Store, error) {
	if opts.EncryptRequired && len(opts.Key) == 0 {
		return nil, fmt.Errorf("database_encrypt is set but TGPORTAL_DB_KEY / database_key is empty")
	}
	dir := filepath.Dir(path)
	if st, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, err
		}
	} else if !st.IsDir() {
		return nil, fmt.Errorf("database parent is not a directory: %s", dir)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	ttl := opts.PendingTTL
	if ttl <= 0 {
		ttl = time.Duration(DefaultPendingTTLSec) * time.Second
	}
	n := opts.CodeLength
	if n < 4 || n > 12 {
		n = DefaultCodeLength
	}
	cr, err := newCrypter(opts.Key)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	s := &Store{db: db, crypt: cr, pendingTTL: ttl, codeLen: n}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.verifyAndMigrateSecrets(); err != nil {
		_ = db.Close()
		return nil, err
	}
	_ = os.Chmod(path, 0o600)
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// Vacuum rebuilds the SQLite file so deleted rows are not left in free pages.
func (s *Store) Vacuum() error {
	_, err := s.db.Exec(`VACUUM`)
	return err
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS pairs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  code TEXT NOT NULL UNIQUE,
  telegram_user_id INTEGER NOT NULL,
  telegram_username TEXT NOT NULL DEFAULT '',
  telegram_chat_id INTEGER NOT NULL DEFAULT 0,
  dc_account_id INTEGER NOT NULL DEFAULT 0,
  dc_chat_id INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  paired_at INTEGER
);
CREATE INDEX IF NOT EXISTS idx_pairs_tg_user ON pairs(telegram_user_id);
CREATE INDEX IF NOT EXISTS idx_pairs_status ON pairs(status);
CREATE INDEX IF NOT EXISTS idx_pairs_code ON pairs(code);
`)
	if err != nil {
		return err
	}
	// Additive: owner vcard for persona key import (mode-1 pairing capture).
	_, _ = s.db.Exec(`ALTER TABLE pairs ADD COLUMN owner_vcard TEXT NOT NULL DEFAULT ''`)
	_, _ = s.db.Exec(`ALTER TABLE pairs ADD COLUMN code_hmac TEXT NOT NULL DEFAULT ''`)
	// Replace the old non-unique hmac index so two rows cannot share a code when encrypted.
	_, _ = s.db.Exec(`DROP INDEX IF EXISTS idx_pairs_code_hmac`)
	_, err = s.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_pairs_code_hmac_uid ON pairs(code_hmac) WHERE code_hmac != ''`)
	if err != nil {
		return err
	}
	return s.migratePersona()
}

func (s *Store) seal(plain string) (string, error) {
	return s.crypt.Seal(plain)
}

func (s *Store) unseal(stored string) (string, error) {
	return s.crypt.Open(stored)
}

func (s *Store) storePairCode(plain string) (stored, hmac string, err error) {
	stored, err = s.seal(plain)
	if err != nil {
		return "", "", err
	}
	if s.crypt != nil {
		hmac = s.crypt.HMACCode(plain)
	}
	return stored, hmac, nil
}

func (s *Store) expireIfStale(p *Pair) (*Pair, error) {
	if p == nil || p.Status != StatusPending {
		return p, nil
	}
	if time.Since(p.CreatedAt) <= s.pendingTTL {
		return p, nil
	}
	_, _ = s.db.Exec(`UPDATE pairs SET status = 'expired' WHERE id = ? AND status = ?`, p.ID, StatusPending)
	return nil, nil
}

func (s *Store) verifyAndMigrateSecrets() error {
	if s.crypt == nil {
		var n int
		_ = s.db.QueryRow(`SELECT COUNT(*) FROM pairs WHERE owner_vcard LIKE 'enc:v1:%' OR code LIKE 'enc:v1:%'`).Scan(&n)
		if n > 0 {
			return fmt.Errorf("database contains encrypted rows; set TGPORTAL_DB_KEY")
		}
		_ = s.db.QueryRow(`SELECT COUNT(*) FROM persona_bots WHERE bot_token LIKE 'enc:v1:%' OR owner_vcard LIKE 'enc:v1:%'`).Scan(&n)
		if n > 0 {
			return fmt.Errorf("database contains encrypted rows; set TGPORTAL_DB_KEY")
		}
		return nil
	}
	// Probe any existing ciphertext with this key.
	var sample string
	_ = s.db.QueryRow(`SELECT owner_vcard FROM pairs WHERE owner_vcard LIKE 'enc:v1:%' LIMIT 1`).Scan(&sample)
	if sample == "" {
		_ = s.db.QueryRow(`SELECT bot_token FROM persona_bots WHERE bot_token LIKE 'enc:v1:%' LIMIT 1`).Scan(&sample)
	}
	if sample == "" {
		_ = s.db.QueryRow(`SELECT code FROM pairs WHERE code LIKE 'enc:v1:%' LIMIT 1`).Scan(&sample)
	}
	if sample != "" {
		if _, err := s.unseal(sample); err != nil {
			return err
		}
	}
	return s.encryptPlaintextSecrets()
}

func (s *Store) encryptPlaintextSecrets() error {
	if s.crypt == nil {
		return nil
	}
	rows, err := s.db.Query(`SELECT id, code, owner_vcard, code_hmac FROM pairs`)
	if err != nil {
		return err
	}
	var pairs []struct {
		id                int64
		code, vcard, hmac string
	}
	for rows.Next() {
		var id int64
		var code, vcard, hmac string
		if err := rows.Scan(&id, &code, &vcard, &hmac); err != nil {
			rows.Close()
			return err
		}
		pairs = append(pairs, struct {
			id                int64
			code, vcard, hmac string
		}{id, code, vcard, hmac})
	}
	rows.Close()
	for _, p := range pairs {
		code := p.code
		vcard := p.vcard
		hmac := p.hmac
		changed := false
		if code != "" && !isEncrypted(code) {
			plain := code
			sealed, err := s.seal(plain)
			if err != nil {
				return err
			}
			code = sealed
			hmac = s.crypt.HMACCode(plain)
			changed = true
		}
		if vcard != "" && !isEncrypted(vcard) {
			sealed, err := s.seal(vcard)
			if err != nil {
				return err
			}
			vcard = sealed
			changed = true
		}
		if changed {
			if _, err := s.db.Exec(`UPDATE pairs SET code = ?, owner_vcard = ?, code_hmac = ? WHERE id = ?`,
				code, vcard, hmac, p.id); err != nil {
				return err
			}
		}
	}

	prows, err := s.db.Query(`SELECT id, bot_token, owner_vcard FROM persona_bots`)
	if err != nil {
		return err
	}
	type pb struct {
		id           int64
		token, vcard string
	}
	var bots []pb
	for prows.Next() {
		var r pb
		if err := prows.Scan(&r.id, &r.token, &r.vcard); err != nil {
			prows.Close()
			return err
		}
		bots = append(bots, r)
	}
	prows.Close()
	for _, r := range bots {
		tok, vc := r.token, r.vcard
		changed := false
		if tok != "" && !isEncrypted(tok) {
			sld, err := s.seal(tok)
			if err != nil {
				return err
			}
			tok = sld
			changed = true
		}
		if vc != "" && !isEncrypted(vc) {
			sld, err := s.seal(vc)
			if err != nil {
				return err
			}
			vc = sld
			changed = true
		}
		if changed {
			if _, err := s.db.Exec(`UPDATE persona_bots SET bot_token = ?, owner_vcard = ? WHERE id = ?`, tok, vc, r.id); err != nil {
				return err
			}
		}
	}

	grows, err := s.db.Query(`SELECT id, invite_qr FROM ghost_groups`)
	if err != nil {
		return err
	}
	var groups []struct {
		id int64
		qr string
	}
	for grows.Next() {
		var id int64
		var qr string
		if err := grows.Scan(&id, &qr); err != nil {
			grows.Close()
			return err
		}
		groups = append(groups, struct {
			id int64
			qr string
		}{id, qr})
	}
	grows.Close()
	for _, g := range groups {
		if g.qr == "" || isEncrypted(g.qr) {
			continue
		}
		sld, err := s.seal(g.qr)
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(`UPDATE ghost_groups SET invite_qr = ? WHERE id = ?`, sld, g.id); err != nil {
			return err
		}
	}
	return nil
}

// CreatePendingPair creates a new pending pairing code for a Telegram user.
// Previous pending pairs for the same user are cancelled.
// (Telegram-initiated: TG filled, DC set later when the code is sent in Delta Chat.)
func (s *Store) CreatePendingPair(tgUserID int64, username string, tgChatID int64) (*Pair, error) {
	if _, err := s.db.Exec(
		`UPDATE pairs SET status = 'expired' WHERE telegram_user_id = ? AND status = ?`,
		tgUserID, StatusPending,
	); err != nil {
		return nil, err
	}

	for i := 0; i < 12; i++ {
		c, err := randomCode(s.codeLen)
		if err != nil {
			return nil, err
		}
		stored, hmac, err := s.storePairCode(c)
		if err != nil {
			return nil, err
		}
		now := time.Now().Unix()
		res, err := s.db.Exec(
			`INSERT INTO pairs (code, code_hmac, telegram_user_id, telegram_username, telegram_chat_id, status, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			stored, hmac, tgUserID, username, tgChatID, StatusPending, now,
		)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				continue
			}
			return nil, err
		}
		id, _ := res.LastInsertId()
		return &Pair{
			ID:               id,
			Code:             c,
			TelegramUserID:   tgUserID,
			TelegramUsername: username,
			TelegramChatID:   tgChatID,
			Status:           StatusPending,
			CreatedAt:        time.Unix(now, 0),
		}, nil
	}
	return nil, fmt.Errorf("could not allocate unique pairing code")
}

// CreatePendingFromDC creates a pending code for a Delta Chat conversation.
// Previous pending pairs for the same DC chat are cancelled.
// (DC-initiated: DC filled, Telegram claims via t.me/bot?start=CODE.)
func (s *Store) CreatePendingFromDC(accID, chatID uint32) (*Pair, error) {
	if _, err := s.db.Exec(
		`UPDATE pairs SET status = 'expired'
		 WHERE dc_account_id = ? AND dc_chat_id = ? AND status = ? AND telegram_user_id = 0`,
		accID, chatID, StatusPending,
	); err != nil {
		return nil, err
	}

	for i := 0; i < 12; i++ {
		c, err := randomCode(s.codeLen)
		if err != nil {
			return nil, err
		}
		stored, hmac, err := s.storePairCode(c)
		if err != nil {
			return nil, err
		}
		now := time.Now().Unix()
		res, err := s.db.Exec(
			`INSERT INTO pairs (code, code_hmac, telegram_user_id, telegram_username, telegram_chat_id,
			                   dc_account_id, dc_chat_id, status, created_at)
			 VALUES (?, ?, 0, '', 0, ?, ?, ?, ?)`,
			stored, hmac, accID, chatID, StatusPending, now,
		)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				continue
			}
			return nil, err
		}
		id, _ := res.LastInsertId()
		return &Pair{
			ID:          id,
			Code:        c,
			DCAccountID: accID,
			DCChatID:    chatID,
			Status:      StatusPending,
			CreatedAt:   time.Unix(now, 0),
		}, nil
	}
	return nil, fmt.Errorf("could not allocate unique pairing code")
}

// GetPendingByDCChat returns a DC-initiated pending pair (telegram_user_id = 0) for a chat.
func (s *Store) GetPendingByDCChat(accID, chatID uint32) (*Pair, error) {
	p, err := s.scanOne(
		`SELECT id, code, telegram_user_id, telegram_username, telegram_chat_id,
		        dc_account_id, dc_chat_id, owner_vcard, status, created_at, paired_at
		 FROM pairs WHERE dc_account_id = ? AND dc_chat_id = ? AND status = ?
		   AND telegram_user_id = 0 ORDER BY id DESC LIMIT 1`,
		accID, chatID, StatusPending,
	)
	if err != nil {
		return nil, err
	}
	return s.expireIfStale(p)
}

// ActivatePairFromTelegram claims a DC-initiated pending code for a Telegram user.
func (s *Store) ActivatePairFromTelegram(code string, tgUserID int64, username string, tgChatID int64) (*Pair, error) {
	code = normalizeCode(code)
	pending, err := s.GetPendingByCode(code)
	if err != nil {
		return nil, err
	}
	if pending == nil {
		return nil, fmt.Errorf("invalid or expired pairing code")
	}
	// Already has a different Telegram user (TG-initiated code for someone else).
	if pending.TelegramUserID != 0 && pending.TelegramUserID != tgUserID {
		return nil, fmt.Errorf("this code belongs to another Telegram account")
	}
	// TG-initiated pending (DC not bound yet): just attach/update this TG user and wait for DC.
	if pending.DCAccountID == 0 || pending.DCChatID == 0 {
		if _, err := s.db.Exec(
			`UPDATE pairs SET telegram_user_id = ?, telegram_username = ?, telegram_chat_id = ?
			 WHERE id = ? AND status = ?`,
			tgUserID, username, tgChatID, pending.ID, StatusPending,
		); err != nil {
			return nil, err
		}
		return s.GetPendingByCode(code)
	}
	// DC-initiated: activate fully.
	if _, err := s.db.Exec(
		`UPDATE pairs SET status = 'disconnected' WHERE telegram_user_id = ? AND status = ? AND id != ?`,
		tgUserID, StatusActive, pending.ID,
	); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`UPDATE pairs SET status = ?, telegram_user_id = ?, telegram_username = ?,
		                     telegram_chat_id = ?, paired_at = ?
		 WHERE id = ? AND status = ?`,
		StatusActive, tgUserID, username, tgChatID, now, pending.ID, StatusPending,
	)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, fmt.Errorf("invalid or expired pairing code")
	}
	return s.scanOne(
		`SELECT id, code, telegram_user_id, telegram_username, telegram_chat_id,
		        dc_account_id, dc_chat_id, owner_vcard, status, created_at, paired_at
		 FROM pairs WHERE id = ? LIMIT 1`,
		pending.ID,
	)
}

// GetActiveByTelegram returns the active pair for a Telegram user, if any.
func (s *Store) GetActiveByTelegram(tgUserID int64) (*Pair, error) {
	return s.scanOne(
		`SELECT id, code, telegram_user_id, telegram_username, telegram_chat_id,
		        dc_account_id, dc_chat_id, owner_vcard, status, created_at, paired_at
		 FROM pairs WHERE telegram_user_id = ? AND status = ? ORDER BY id DESC LIMIT 1`,
		tgUserID, StatusActive,
	)
}

// GetPendingByCode returns a pending pair by code (case-insensitive).
func (s *Store) GetPendingByCode(code string) (*Pair, error) {
	code = normalizeCode(code)
	var p *Pair
	var err error
	if s.crypt != nil {
		p, err = s.scanOne(
			`SELECT id, code, telegram_user_id, telegram_username, telegram_chat_id,
			        dc_account_id, dc_chat_id, owner_vcard, status, created_at, paired_at
			 FROM pairs WHERE code_hmac = ? AND status = ? LIMIT 1`,
			s.crypt.HMACCode(code), StatusPending,
		)
	} else {
		p, err = s.scanOne(
			`SELECT id, code, telegram_user_id, telegram_username, telegram_chat_id,
			        dc_account_id, dc_chat_id, owner_vcard, status, created_at, paired_at
			 FROM pairs WHERE code = ? AND status = ? LIMIT 1`,
			code, StatusPending,
		)
	}
	if err != nil {
		return nil, err
	}
	return s.expireIfStale(p)
}

// SetPairOwnerVcard stores the peer vCard (public key) for an active pair.
func (s *Store) SetPairOwnerVcard(pairID int64, vcard string) error {
	stored, err := s.seal(vcard)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE pairs SET owner_vcard = ? WHERE id = ?`, stored, pairID)
	return err
}

// ListActivePairs returns all active mode-1 pairs.
func (s *Store) ListActivePairs() ([]Pair, error) {
	rows, err := s.db.Query(
		`SELECT id, code, telegram_user_id, telegram_username, telegram_chat_id,
		        dc_account_id, dc_chat_id, owner_vcard, status, created_at, paired_at
		 FROM pairs WHERE status = ? ORDER BY id`,
		StatusActive,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Pair
	for rows.Next() {
		var p Pair
		var created, paired sql.NullInt64
		var username, vcard string
		if err := rows.Scan(
			&p.ID, &p.Code, &p.TelegramUserID, &username, &p.TelegramChatID,
			&p.DCAccountID, &p.DCChatID, &vcard, &p.Status, &created, &paired,
		); err != nil {
			return nil, err
		}
		p.TelegramUsername = username
		plainCode, err := s.unseal(p.Code)
		if err != nil {
			return nil, err
		}
		p.Code = plainCode
		plainV, err := s.unseal(vcard)
		if err != nil {
			return nil, err
		}
		p.OwnerVcard = plainV
		if created.Valid {
			p.CreatedAt = time.Unix(created.Int64, 0)
		}
		if paired.Valid {
			t := time.Unix(paired.Int64, 0)
			p.PairedAt = &t
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ActivatePair binds a pending code to a Delta Chat account/chat (TG-initiated flow).
// Rejects DC-initiated codes that already have DC bound (those are claimed via Telegram).
func (s *Store) ActivatePair(code string, accID, chatID uint32) (*Pair, error) {
	code = normalizeCode(code)
	pending, err := s.GetPendingByCode(code)
	if err != nil {
		return nil, err
	}
	if pending == nil {
		return nil, fmt.Errorf("invalid or expired pairing code")
	}
	// DC-initiated code: user must open the Telegram start link, not paste on DC again.
	if pending.DCAccountID != 0 && pending.DCChatID != 0 && pending.TelegramUserID == 0 {
		return nil, fmt.Errorf("open the Telegram link for this code (t.me/…?start=%s)", code)
	}
	if pending.TelegramUserID == 0 {
		return nil, fmt.Errorf("invalid or expired pairing code")
	}
	// Refuse hijack: a new code must not steal an already-paired DC chat.
	if existing, err := s.GetActiveByDCChat(accID, chatID); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, fmt.Errorf("this Delta Chat is already paired")
	}
	// One Telegram user → one active pair (avoid split-brain with an old DC chat).
	if _, err := s.db.Exec(
		`UPDATE pairs SET status = 'disconnected' WHERE telegram_user_id = ? AND status = ? AND id != ?`,
		pending.TelegramUserID, StatusActive, pending.ID,
	); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`UPDATE pairs SET status = ?, dc_account_id = ?, dc_chat_id = ?, paired_at = ?
		 WHERE id = ? AND status = ?`,
		StatusActive, accID, chatID, now, pending.ID, StatusPending,
	)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, fmt.Errorf("invalid or expired pairing code")
	}
	return s.scanOne(
		`SELECT id, code, telegram_user_id, telegram_username, telegram_chat_id,
		        dc_account_id, dc_chat_id, owner_vcard, status, created_at, paired_at
		 FROM pairs WHERE id = ? LIMIT 1`,
		pending.ID,
	)
}

// GetActiveByDCChat returns active pair for a DC chat (if any).
func (s *Store) GetActiveByDCChat(accID, chatID uint32) (*Pair, error) {
	return s.scanOne(
		`SELECT id, code, telegram_user_id, telegram_username, telegram_chat_id,
		        dc_account_id, dc_chat_id, owner_vcard, status, created_at, paired_at
		 FROM pairs WHERE dc_account_id = ? AND dc_chat_id = ? AND status = ? LIMIT 1`,
		accID, chatID, StatusActive,
	)
}

// DisconnectByTelegram ends active (and pending) pairs for a Telegram user.
func (s *Store) DisconnectByTelegram(tgUserID int64) (int64, error) {
	res, err := s.db.Exec(
		`UPDATE pairs SET status = 'disconnected' WHERE telegram_user_id = ? AND status IN (?, ?)`,
		tgUserID, StatusPending, StatusActive,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DisconnectByDCChat ends the active pair for a Delta Chat conversation.
func (s *Store) DisconnectByDCChat(accID, chatID uint32) (int64, error) {
	res, err := s.db.Exec(
		`UPDATE pairs SET status = 'disconnected' WHERE dc_account_id = ? AND dc_chat_id = ? AND status = ?`,
		accID, chatID, StatusActive,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ListPairsByTelegram returns all pair rows for a Telegram user (any status).
func (s *Store) ListPairsByTelegram(tgUserID int64) ([]Pair, error) {
	rows, err := s.db.Query(
		`SELECT id, code, telegram_user_id, telegram_username, telegram_chat_id,
		        dc_account_id, dc_chat_id, owner_vcard, status, created_at, paired_at
		 FROM pairs WHERE telegram_user_id = ?`,
		tgUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanMany(rows)
}

// ListPairsByDCChat returns pair rows for a Delta Chat conversation (any status).
func (s *Store) ListPairsByDCChat(accID, chatID uint32) ([]Pair, error) {
	rows, err := s.db.Query(
		`SELECT id, code, telegram_user_id, telegram_username, telegram_chat_id,
		        dc_account_id, dc_chat_id, owner_vcard, status, created_at, paired_at
		 FROM pairs WHERE dc_account_id = ? AND dc_chat_id = ?`,
		accID, chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanMany(rows)
}

func (s *Store) scanMany(rows *sql.Rows) ([]Pair, error) {
	var out []Pair
	for rows.Next() {
		p, err := s.scanPair(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

type pairScanner interface {
	Scan(dest ...any) error
}

func (s *Store) scanOne(query string, args ...any) (*Pair, error) {
	return s.scanPair(s.db.QueryRow(query, args...))
}

func (s *Store) scanPair(row pairScanner) (*Pair, error) {
	var p Pair
	var created, paired sql.NullInt64
	var username string
	var vcard string
	err := row.Scan(
		&p.ID, &p.Code, &p.TelegramUserID, &username, &p.TelegramChatID,
		&p.DCAccountID, &p.DCChatID, &vcard, &p.Status, &created, &paired,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.TelegramUsername = username
	plainCode, err := s.unseal(p.Code)
	if err != nil {
		return nil, err
	}
	p.Code = plainCode
	plainV, err := s.unseal(vcard)
	if err != nil {
		return nil, err
	}
	p.OwnerVcard = plainV
	if created.Valid {
		p.CreatedAt = time.Unix(created.Int64, 0)
	}
	if paired.Valid {
		t := time.Unix(paired.Int64, 0)
		p.PairedAt = &t
	}
	return &p, nil
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// LooksLikeCode reports whether text is a plausible pairing code.
func LooksLikeCode(text string) bool {
	c := normalizeCode(text)
	if len(c) < 4 || len(c) > 12 {
		return false
	}
	for _, r := range c {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O/1/I

func randomCode(n int) (string, error) {
	var b strings.Builder
	max := big.NewInt(int64(len(codeAlphabet)))
	for i := 0; i < n; i++ {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(codeAlphabet[v.Int64()])
	}
	return b.String(), nil
}
