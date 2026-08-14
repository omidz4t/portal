// Package usermsg holds replies shown in Telegram / Delta Chat.
// Never include err.Error() — it can leak paths, SQL, tokens, and config names.
package usermsg

const (
	Generic        = "Something went wrong. Try again in a moment."
	PairingFailed  = "Pairing didn't work. Send /pair for a new invite and code."
	BridgeFailed   = "Could not send that. Try again."
	RegisterFailed = "Could not register that bot. Create a new bot with @BotFather (not this portal)."
)

// Safe returns a user-visible string that never contains the internal error.
func Safe(err error) string {
	_ = err
	return Generic
}
