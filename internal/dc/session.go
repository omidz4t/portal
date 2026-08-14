package dc

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chatmail/rpc-client-go/v2/deltachat"

	"github.com/omidz4t/portal/internal/safemedia"
)

// Session serializes application-level Delta Chat RPC used by TGPORTAL.
//
// Context (from rpc-client-go Bot.Run):
//   - Events are pulled on one goroutine and handlers run inline on that loop.
//   - jrpc2.Client is concurrent-safe, so multiple Call/CallResult can multiplex.
//   - In practice, concurrent SendMsg / SecureJoin / GetMessage from many
//     Telegram workers + DC handlers races with long retries and can overload
//     a single-account core session.
//
// Session keeps a single mutex around *our* RPC ops (not GetNextEvent, which
// the library runs separately). Handlers must not block Bot.Run — call Session
// methods from worker goroutines instead.
type Session struct {
	Bot *deltachat.Bot
	mu  sync.Mutex
}

// NewSession wraps bot for concurrent-safe outbound RPC.
func NewSession(bot *deltachat.Bot) *Session {
	return &Session{Bot: bot}
}

// Do runs fn while holding the RPC mutex (one app-level DC call at a time).
func (s *Session) Do(fn func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn()
}

// FirstConfiguredAccount returns the first configured account id.
func (s *Session) FirstConfiguredAccount() (uint32, error) {
	var accID uint32
	err := s.Do(func() error {
		accounts, err := s.Bot.Rpc.GetAllAccountIds()
		if err != nil {
			return err
		}
		for _, id := range accounts {
			ok, err := s.Bot.Rpc.IsConfigured(id)
			if err != nil || !ok {
				continue
			}
			accID = id
			return nil
		}
		return fmt.Errorf("no configured Delta Chat account")
	})
	return accID, err
}

// GetMessage fetches a message under the session lock.
func (s *Session) GetMessage(accID, msgID uint32) (deltachat.Message, error) {
	var msg deltachat.Message
	err := s.Do(func() error {
		m, err := s.Bot.Rpc.GetMessage(accID, msgID)
		if err != nil {
			return err
		}
		msg = m
		return nil
	})
	return msg, err
}

// GetChatSecurejoinQrCode returns the bot invite link.
func (s *Session) GetChatSecurejoinQrCode(accID uint32) (string, error) {
	var link string
	err := s.Do(func() error {
		var err error
		link, err = s.Bot.Rpc.GetChatSecurejoinQrCode(accID, nil)
		return err
	})
	return link, err
}

// SetConfig sets an account config key.
func (s *Session) SetConfig(accID uint32, key string, value *string) error {
	return s.Do(func() error {
		return s.Bot.Rpc.SetConfig(accID, key, value)
	})
}

// OpenChatFromInvite prefers SecureJoin; falls back to address contact.
func (s *Session) OpenChatFromInvite(accID uint32, inviteURL string) (uint32, error) {
	var chatID uint32
	err := s.Do(func() error {
		if inviteURL == "" {
			return fmt.Errorf("invite URL is empty")
		}
		if id, err := s.Bot.Rpc.SecureJoin(accID, inviteURL); err == nil && id != 0 {
			chatID = id
			return nil
		}
		addr, err := EmailFromInviteURL(inviteURL)
		if err != nil {
			return fmt.Errorf("secure join failed and could not parse invite addr: %w", err)
		}
		contactID, err := s.Bot.Rpc.CreateContact(accID, addr, nil)
		if err != nil {
			return err
		}
		id, err := s.Bot.Rpc.CreateChatByContactId(accID, contactID)
		if err != nil {
			return err
		}
		chatID = id
		return nil
	})
	return chatID, err
}

// SendTextWithRetry sends text, retrying while encryption keys settle.
// The mutex is released between attempts so other work can proceed.
func (s *Session) SendTextWithRetry(accID, chatID uint32, text string, attempts int) error {
	msg := text
	return s.sendWithRetry(accID, chatID, deltachat.MessageData{Text: &msg}, attempts)
}

// SendFileWithRetry sends a file with optional caption (empty caption = media only).
func (s *Session) SendFileWithRetry(accID, chatID uint32, path, filename, caption string, view *deltachat.Viewtype, attempts int) error {
	data := deltachat.MessageData{
		File:     &path,
		Viewtype: view,
	}
	if filename != "" {
		data.Filename = &filename
	}
	if caption != "" {
		data.Text = &caption
	}
	return s.sendWithRetry(accID, chatID, data, attempts)
}

func (s *Session) sendWithRetry(accID, chatID uint32, data deltachat.MessageData, attempts int) error {
	if attempts < 1 {
		attempts = 1
	}
	var last error
	for i := 1; i <= attempts; i++ {
		var sentID uint32
		err := s.Do(func() error {
			id, err := s.Bot.Rpc.SendMsg(accID, chatID, data)
			sentID = id
			return err
		})
		if err == nil {
			if data.File != nil {
				s.ForgetMessageFile(accID, sentID)
			}
			return nil
		}
		last = err
		if i < attempts {
			// Unlock already released; yield so other goroutines can use DC RPC.
			time.Sleep(500 * time.Millisecond)
		}
	}
	return fmt.Errorf("after retries: %w", last)
}

// AllConfiguredAccounts lists configured account IDs.
func (s *Session) AllConfiguredAccounts() ([]uint32, error) {
	var out []uint32
	err := s.Do(func() error {
		accounts, err := s.Bot.Rpc.GetAllAccountIds()
		if err != nil {
			return err
		}
		for _, id := range accounts {
			ok, err := s.Bot.Rpc.IsConfigured(id)
			if err != nil || !ok {
				continue
			}
			out = append(out, id)
		}
		return nil
	})
	return out, err
}

// SaveMsgFile writes a message attachment to path (creates parent dirs if needed).
func (s *Session) SaveMsgFile(accID, msgID uint32, path string) error {
	return s.Do(func() error {
		return s.Bot.Rpc.SaveMsgFile(accID, msgID, path)
	})
}

// DownloadFullMessage requests a full download when the message is only a placeholder.
func (s *Session) DownloadFullMessage(accID, msgID uint32) error {
	return s.Do(func() error {
		return s.Bot.Rpc.DownloadFullMessage(accID, msgID)
	})
}

// PeerInChat returns the non-self contact address/display name for a 1:1 chat.
// Used by /status to show the Delta Chat identity on the other side of the bridge.
func (s *Session) PeerInChat(accID, chatID uint32) (address, displayName string, err error) {
	c, err := s.PeerContact(accID, chatID)
	if err != nil {
		return "", "", err
	}
	return c.Address, c.DisplayName, nil
}

// PeerContactInfo is the non-self contact in a 1:1 chat.
type PeerContactInfo struct {
	ContactID    uint32
	Address      string
	DisplayName  string
	IsKeyContact bool
}

// PeerContact returns the best non-special contact in a chat (prefer e2ee key-contacts).
func (s *Session) PeerContact(accID, chatID uint32) (PeerContactInfo, error) {
	var out PeerContactInfo
	var found bool
	err := s.Do(func() error {
		ids, err := s.Bot.Rpc.GetChatContacts(accID, chatID)
		if err != nil {
			return err
		}
		var fallback PeerContactInfo
		var haveFallback bool
		for _, id := range ids {
			if id <= deltachat.ContactLastSpecial {
				continue
			}
			c, err := s.Bot.Rpc.GetContact(accID, id)
			if err != nil {
				continue
			}
			if c.Address == "" && !c.IsKeyContact {
				continue
			}
			info := PeerContactInfo{
				ContactID:    id,
				Address:      c.Address,
				IsKeyContact: c.IsKeyContact,
				DisplayName:  c.DisplayName,
			}
			if info.DisplayName == "" {
				info.DisplayName = c.Name
			}
			if info.DisplayName == "" {
				info.DisplayName = c.AuthName
			}
			// Prefer contacts that already have e2e keys for vcard export.
			if c.E2eeAvail {
				out = info
				found = true
				return nil
			}
			if c.IsKeyContact && !found {
				out = info
				found = true
				// keep scanning for E2eeAvail
				continue
			}
			if !haveFallback {
				fallback = info
				haveFallback = true
			}
		}
		if found {
			return nil
		}
		if haveFallback {
			out = fallback
			found = true
			return nil
		}
		return fmt.Errorf("no peer contact in chat %d", chatID)
	})
	return out, err
}

// MakeVcardForContacts exports contacts as a vCard string (includes public keys when known).
func (s *Session) MakeVcardForContacts(accID uint32, contactIDs []uint32) (string, error) {
	var vcard string
	err := s.Do(func() error {
		var err error
		vcard, err = s.Bot.Rpc.MakeVcard(accID, contactIDs)
		return err
	})
	return vcard, err
}

// ImportVcardContents imports a vCard into an account and returns created/modified contact ids.
func (s *Session) ImportVcardContents(accID uint32, vcard string) ([]uint32, error) {
	var ids []uint32
	err := s.Do(func() error {
		var err error
		ids, err = s.Bot.Rpc.ImportVcardContents(accID, vcard)
		return err
	})
	return ids, err
}

// CreateChatByContactID opens (or reuses) the 1:1 chat for a contact id.
func (s *Session) CreateChatByContactID(accID, contactID uint32) (uint32, error) {
	var chatID uint32
	err := s.Do(func() error {
		id, err := s.Bot.Rpc.CreateChatByContactId(accID, contactID)
		if err != nil {
			return err
		}
		chatID = id
		return nil
	})
	return chatID, err
}

// ExportPeerVcard exports the peer of a 1:1 chat as vCard (for owner key capture).
func (s *Session) ExportPeerVcard(accID, chatID uint32) (vcard string, peer PeerContactInfo, err error) {
	peer, err = s.PeerContact(accID, chatID)
	if err != nil {
		return "", peer, err
	}
	vcard, err = s.MakeVcardForContacts(accID, []uint32{peer.ContactID})
	if err != nil {
		return "", peer, err
	}
	if strings.TrimSpace(vcard) == "" {
		return "", peer, fmt.Errorf("empty vcard for peer contact %d", peer.ContactID)
	}
	return vcard, peer, nil
}

// ImportOwnerAsKeyContact imports the owner's vCard into a ghost account and opens a 1:1 chat.
func (s *Session) ImportOwnerAsKeyContact(ghostAccID uint32, ownerVcard string) (chatID uint32, contactID uint32, err error) {
	ids, err := s.ImportVcardContents(ghostAccID, ownerVcard)
	if err != nil {
		return 0, 0, fmt.Errorf("import owner vcard: %w", err)
	}
	if len(ids) == 0 {
		return 0, 0, fmt.Errorf("import owner vcard: no contacts created")
	}
	// Prefer a contact that has e2e available after import.
	contactID = ids[0]
	err = s.Do(func() error {
		for _, id := range ids {
			c, err := s.Bot.Rpc.GetContact(ghostAccID, id)
			if err != nil {
				continue
			}
			if c.E2eeAvail || c.IsKeyContact {
				contactID = id
				if c.E2eeAvail {
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return 0, 0, err
	}
	// Verify key presence before opening chat.
	var e2ee bool
	_ = s.Do(func() error {
		c, err := s.Bot.Rpc.GetContact(ghostAccID, contactID)
		if err != nil {
			return err
		}
		e2ee = c.E2eeAvail
		return nil
	})
	if !e2ee {
		// Still try — some cores set e2ee only after chat open; report for logging.
	}
	chatID, err = s.CreateChatByContactID(ghostAccID, contactID)
	if err != nil {
		return 0, contactID, fmt.Errorf("create chat with owner: %w", err)
	}
	return chatID, contactID, nil
}

// ContactE2EE reports whether encryption is available for a contact.
func (s *Session) ContactE2EE(accID, contactID uint32) (bool, error) {
	var ok bool
	err := s.Do(func() error {
		c, err := s.Bot.Rpc.GetContact(accID, contactID)
		if err != nil {
			return err
		}
		ok = c.E2eeAvail
		return nil
	})
	return ok, err
}

// AccountAddress returns the bot account's configured email address.
func (s *Session) AccountAddress(accID uint32) (string, error) {
	var addr string
	err := s.Do(func() error {
		// Prefer configured_addr; fall back to addr.
		for _, key := range []string{"configured_addr", "addr"} {
			v, err := s.Bot.Rpc.GetConfig(accID, key)
			if err != nil || v == nil || *v == "" {
				continue
			}
			addr = *v
			return nil
		}
		return fmt.Errorf("account %d has no configured address", accID)
	})
	return addr, err
}

// AddAccount creates a new unconfigured core account and returns its id.
func (s *Session) AddAccount() (uint32, error) {
	var id uint32
	err := s.Do(func() error {
		var err error
		id, err = s.Bot.Rpc.AddAccount()
		return err
	})
	return id, err
}

// RemoveAccount stops IO and deletes a core account (failed ghost create cleanup).
func (s *Session) RemoveAccount(accID uint32) error {
	if accID == 0 {
		return nil
	}
	return s.Do(func() error {
		_ = s.Bot.Rpc.StopIo(accID)
		return s.Bot.Rpc.RemoveAccount(accID)
	})
}

// ConfigureAccountFromQR provisions transport from a dcaccount:/dclogin: QR and starts IO.
// Long-running: does not hold Session.mu so other sends/receives can proceed
// (jrpc2 is concurrent-safe on the wire).
func (s *Session) ConfigureAccountFromQR(accID uint32, accountQR string) error {
	if strings.TrimSpace(accountQR) == "" {
		return fmt.Errorf("persona account_qr is empty (set persona.account_qr or PERSONA_ACCOUNT_QR)")
	}
	if err := s.Bot.Rpc.AddTransportFromQr(accID, accountQR); err != nil {
		return fmt.Errorf("add_transport_from_qr: %w", err)
	}
	if err := s.Bot.Rpc.StartIo(accID); err != nil {
		return fmt.Errorf("start_io: %w", err)
	}
	return nil
}

// SetDisplayName sets the account displayname.
func (s *Session) SetDisplayName(accID uint32, name string) error {
	return s.Do(func() error {
		v := name
		return s.Bot.Rpc.SetConfig(accID, "displayname", &v)
	})
}

// SetSelfAvatar sets the account selfavatar from a local file path (empty clears).
func (s *Session) SetSelfAvatar(accID uint32, path string) error {
	if path != "" {
		if err := safemedia.ValidateFile(path, safemedia.RoleAvatar, safemedia.AvatarMaxBytes); err != nil {
			return fmt.Errorf("avatar: %w", err)
		}
	}
	return s.Do(func() error {
		var p *string
		if path != "" {
			p = &path
		} else {
			empty := ""
			p = &empty
		}
		return s.Bot.Rpc.SetConfig(accID, "selfavatar", p)
	})
}

// StartIO starts background IO for one account.
func (s *Session) StartIO(accID uint32) error {
	return s.Do(func() error {
		return s.Bot.Rpc.StartIo(accID)
	})
}

// CreateChatWithAddress creates (or reuses) a 1:1 chat with the given email address.
// Note: chatmail requires e2e keys; address-only contacts often cannot send until SecureJoin.
func (s *Session) CreateChatWithAddress(accID uint32, address, name string) (uint32, error) {
	var chatID uint32
	err := s.Do(func() error {
		var namePtr *string
		if name != "" {
			n := name
			namePtr = &n
		}
		contactID, err := s.Bot.Rpc.CreateContact(accID, address, namePtr)
		if err != nil {
			return err
		}
		id, err := s.Bot.Rpc.CreateChatByContactId(accID, contactID)
		if err != nil {
			return err
		}
		chatID = id
		return nil
	})
	return chatID, err
}

// GetSetupContactQR returns the account's setup-contact invite (SecureJoin QR text).
func (s *Session) GetSetupContactQR(accID uint32) (string, error) {
	var link string
	err := s.Do(func() error {
		var err error
		link, err = s.Bot.Rpc.GetChatSecurejoinQrCode(accID, nil)
		return err
	})
	return link, err
}

// ChatE2EEAvailable reports whether a chat is encrypted (key-contacts ready).
func (s *Session) ChatE2EEAvailable(accID, chatID uint32) (bool, error) {
	var ok bool
	err := s.Do(func() error {
		info, err := s.Bot.Rpc.GetBasicChatInfo(accID, chatID)
		if err != nil {
			return err
		}
		ok = info.IsEncrypted && !info.IsDeviceChat && !info.IsSelfTalk
		return nil
	})
	return ok, err
}

// IsDirectChat reports whether chat is a 1:1 conversation (pairable).
func (s *Session) IsDirectChat(accID, chatID uint32) (bool, error) {
	var ok bool
	err := s.Do(func() error {
		info, err := s.Bot.Rpc.GetBasicChatInfo(accID, chatID)
		if err != nil {
			return err
		}
		ok = IsPairableChatType(info.ChatType)
		return nil
	})
	return ok, err
}

// AcceptChat accepts a contact-request chat if needed.
func (s *Session) AcceptChat(accID, chatID uint32) error {
	return s.Do(func() error {
		return s.Bot.Rpc.AcceptChat(accID, chatID)
	})
}

// FindEncryptedDirectChat returns the first encrypted 1:1 chat on the account
// (used after the owner SecureJoins a ghost via the ghost's invite QR).
func (s *Session) FindEncryptedDirectChat(accID uint32) (uint32, error) {
	var found uint32
	err := s.Do(func() error {
		entries, err := s.Bot.Rpc.GetChatlistEntries(accID, nil, nil, nil)
		if err != nil {
			return err
		}
		for _, chatID := range entries {
			if chatID == 0 {
				continue
			}
			info, err := s.Bot.Rpc.GetBasicChatInfo(accID, chatID)
			if err != nil {
				continue
			}
			if info.IsDeviceChat || info.IsSelfTalk {
				continue
			}
			if info.ChatType != deltachat.ChatTypeSingle {
				continue
			}
			if !info.IsEncrypted {
				continue
			}
			if info.IsContactRequest {
				_ = s.Bot.Rpc.AcceptChat(accID, chatID)
			}
			found = chatID
			return nil
		}
		return nil
	})
	return found, err
}

// CreateGroupChat creates an unpromoted encrypted group and returns its chat id.
func (s *Session) CreateGroupChat(accID uint32, name string) (uint32, error) {
	var chatID uint32
	err := s.Do(func() error {
		id, err := s.Bot.Rpc.CreateGroupChat(accID, name, false)
		if err != nil {
			return err
		}
		chatID = id
		return nil
	})
	return chatID, err
}

// AddContactToChat adds a contact (by address) to a group chat on accID.
func (s *Session) AddContactToChatByAddress(accID, chatID uint32, address, name string) error {
	return s.Do(func() error {
		var namePtr *string
		if name != "" {
			n := name
			namePtr = &n
		}
		contactID, err := s.Bot.Rpc.CreateContact(accID, address, namePtr)
		if err != nil {
			return err
		}
		return s.Bot.Rpc.AddContactToChat(accID, chatID, contactID)
	})
}

// AddContactIDToChat adds an existing contact to a group.
func (s *Session) AddContactIDToChat(accID, chatID, contactID uint32) error {
	return s.Do(func() error {
		return s.Bot.Rpc.AddContactToChat(accID, chatID, contactID)
	})
}

// FirstPeerContactID returns the first non-special contact in a chat.
func (s *Session) FirstPeerContactID(accID, chatID uint32) (uint32, error) {
	var cid uint32
	err := s.Do(func() error {
		ids, err := s.Bot.Rpc.GetChatContacts(accID, chatID)
		if err != nil {
			return err
		}
		for _, id := range ids {
			if id > deltachat.ContactLastSpecial {
				cid = id
				return nil
			}
		}
		return fmt.Errorf("no peer in chat %d", chatID)
	})
	return cid, err
}

// GetGroupSecurejoinQR returns a SecureJoin QR for a group chat.
func (s *Session) GetGroupSecurejoinQR(accID, chatID uint32) (string, error) {
	var link string
	err := s.Do(func() error {
		id := chatID
		var err error
		link, err = s.Bot.Rpc.GetChatSecurejoinQrCode(accID, &id)
		return err
	})
	return link, err
}

// SecureJoin joins a setup-contact or group invite from the given account.
func (s *Session) SecureJoin(accID uint32, qr string) (uint32, error) {
	var chatID uint32
	err := s.Do(func() error {
		id, err := s.Bot.Rpc.SecureJoin(accID, qr)
		if err != nil {
			return err
		}
		chatID = id
		return nil
	})
	return chatID, err
}

// IsConfigured reports whether the account is configured.
func (s *Session) IsConfigured(accID uint32) (bool, error) {
	var ok bool
	err := s.Do(func() error {
		var err error
		ok, err = s.Bot.Rpc.IsConfigured(accID)
		return err
	})
	return ok, err
}

// ApplyProxy sets core proxy_url / proxy_enabled for all configured accounts
// and restarts IO so the change takes effect.
// proxyURL empty + enabled false clears/disables proxy.
// Config writes stay serialized; IO restart is outside Session.mu so bridging
// is not frozen for the whole multi-account loop.
func (s *Session) ApplyProxy(proxyURL string, enabled bool) error {
	accounts, err := s.AllConfiguredAccounts()
	if err != nil {
		return err
	}
	en := "0"
	if enabled && proxyURL != "" {
		en = "1"
	}
	var urlPtr *string
	if proxyURL != "" {
		u := proxyURL
		urlPtr = &u
	} else {
		empty := ""
		urlPtr = &empty
	}
	for _, id := range accounts {
		err := s.Do(func() error {
			if err := s.Bot.Rpc.SetConfig(id, "proxy_url", urlPtr); err != nil {
				return fmt.Errorf("account %d proxy_url: %w", id, err)
			}
			enCopy := en
			if err := s.Bot.Rpc.SetConfig(id, "proxy_enabled", &enCopy); err != nil {
				return fmt.Errorf("account %d proxy_enabled: %w", id, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
		if err := s.restartAccountIO(id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) restartAccountIO(accID uint32) error {
	_ = s.Bot.Rpc.StopIo(accID)
	if err := s.Bot.Rpc.StartIo(accID); err != nil {
		return fmt.Errorf("account %d start_io: %w", accID, err)
	}
	return nil
}
