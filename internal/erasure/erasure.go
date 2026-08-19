package erasure

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/omidz4t/portal/internal/dc"
	"github.com/omidz4t/portal/internal/persona"
	"github.com/omidz4t/portal/internal/store"
)

const PendingTTL = 10 * time.Minute

const (
	WarnText = "This permanently erases Portal data for you: pairing records, " +
		"persona bots you registered, ghost Delta Chat accounts created for you, " +
		"and the local chat with this bot.\n\n" +
		"This cannot be undone.\n\n" +
		"If you are sure, send:\n/delete_my_data_approve\n\n" +
		"That confirmation expires in 10 minutes."
	NeedRequestText = "Nothing to confirm. Send /delete_my_data first, then /delete_my_data_approve within 10 minutes."
	DoneText        = "Your Portal data has been deleted. Pairing, stored records, and related ghost accounts are gone.\n\nYou can /pair again later if you want."
	PrivateOnly     = "Data deletion only works in a private 1:1 chat with this bot."
)

// Service holds two-step deletion confirmations and runs the purge.
type Service struct {
	log *zap.SugaredLogger
	st  *store.Store
	dc  *dc.Session
	pm  *persona.Manager

	mu      sync.Mutex
	pending map[string]time.Time
}

func New(log *zap.SugaredLogger, st *store.Store, sess *dc.Session, pm *persona.Manager) *Service {
	return &Service{
		log:     log.With("component", "erasure"),
		st:      st,
		dc:      sess,
		pm:      pm,
		pending: make(map[string]time.Time),
	}
}

func TelegramKey(tgUserID int64) string {
	return fmt.Sprintf("tg:%d", tgUserID)
}

func DCKey(accID, chatID uint32) string {
	return fmt.Sprintf("dc:%d:%d", accID, chatID)
}

// Request records a pending confirmation for key.
func (s *Service) Request(key string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweepLocked(time.Now())
	s.pending[key] = time.Now().Add(PendingTTL)
}

// Consume returns whether a valid pending request existed (and removes it).
func (s *Service) Consume(key string) bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.sweepLocked(now)
	exp, ok := s.pending[key]
	if !ok || now.After(exp) {
		return false
	}
	delete(s.pending, key)
	return true
}

func (s *Service) sweepLocked(now time.Time) {
	for k, exp := range s.pending {
		if now.After(exp) {
			delete(s.pending, k)
		}
	}
}

// PurgeTelegramUser removes collected Portal data for a Telegram user.
func (s *Service) PurgeTelegramUser(tgUserID int64) error {
	if s == nil || s.st == nil {
		return fmt.Errorf("erasure not configured")
	}
	ghosts, err := s.st.GhostAccountsForPurge(tgUserID)
	if err != nil {
		return err
	}
	pairs, err := s.st.ListPairsByTelegram(tgUserID)
	if err != nil {
		return err
	}
	bots, err := s.st.ListPersonaBotsByOwner(tgUserID)
	if err != nil {
		return err
	}
	if s.pm != nil {
		for _, b := range bots {
			s.pm.DetachBot(b.ID)
		}
	}
	for _, g := range ghosts {
		if g.DCAccountID == 0 {
			continue
		}
		if s.dc != nil {
			if err := s.dc.RemoveAccount(g.DCAccountID); err != nil {
				s.log.Warnf("remove ghost account: %v", err)
			}
		}
	}
	if err := s.st.PurgeTelegramUser(tgUserID); err != nil {
		return err
	}
	// Delete the portal-bot chat after rows are gone so the confirmation can still send.
	if s.dc != nil {
		for _, p := range pairs {
			if p.DCAccountID != 0 && p.DCChatID != 0 {
				if err := s.dc.ForgetChat(p.DCAccountID, p.DCChatID); err != nil {
					s.log.Warnf("forget paired chat: %v", err)
				}
			}
		}
	}
	if err := s.st.Vacuum(); err != nil {
		s.log.Warnf("vacuum store: %v", err)
	}
	return nil
}

// PurgeFromDCChat erases data for whoever is paired to this DC chat, or the chat's pair rows.
func (s *Service) PurgeFromDCChat(accID, chatID uint32) error {
	if s == nil || s.st == nil {
		return fmt.Errorf("erasure not configured")
	}
	pairs, err := s.st.ListPairsByDCChat(accID, chatID)
	if err != nil {
		return err
	}
	var tgID int64
	for _, p := range pairs {
		if p.TelegramUserID != 0 {
			tgID = p.TelegramUserID
			break
		}
	}
	if tgID != 0 {
		return s.PurgeTelegramUser(tgID)
	}
	if err := s.st.PurgeDCChatPairs(accID, chatID); err != nil {
		return err
	}
	if s.dc != nil {
		if err := s.dc.ForgetChat(accID, chatID); err != nil {
			s.log.Warnf("forget unpaired chat: %v", err)
		}
	}
	if err := s.st.Vacuum(); err != nil {
		s.log.Warnf("vacuum store: %v", err)
	}
	return nil
}
