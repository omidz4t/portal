package persona

import "fmt"

// ghostCore is the DC account lifecycle used when creating a ghost.
type ghostCore interface {
	AddAccount() (uint32, error)
	ConfigureAccountFromQR(accID uint32, qr string) error
	SetDisplayName(accID uint32, name string) error
	SetSelfAvatar(accID uint32, path string) error
	ImportOwnerAsKeyContact(accID uint32, vcard string) (chatID, contactID uint32, err error)
	AccountAddress(accID uint32) (string, error)
	RemoveAccount(accID uint32) error
}

// withNewGhostAccount runs fn after AddAccount. If fn fails, the account is removed
// so a leaked chatmail/IO slot cannot bypass max_ghosts on retry.
func withNewGhostAccount(core ghostCore, fn func(accID uint32) error) (accID uint32, err error) {
	accID, err = core.AddAccount()
	if err != nil {
		return 0, fmt.Errorf("add account: %w", err)
	}
	if accID == 0 {
		return 0, fmt.Errorf("add account: empty id")
	}
	if err = fn(accID); err != nil {
		if remErr := core.RemoveAccount(accID); remErr != nil {
			return 0, fmt.Errorf("%w (also failed to remove account %d: %v)", err, accID, remErr)
		}
		return 0, err
	}
	return accID, nil
}
