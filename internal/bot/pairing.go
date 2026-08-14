package bot

import "github.com/omidz4t/portal/internal/dc"

// shouldTreatAsPairingCode is false when this DC chat is already paired so a
// pasted/forwarded code cannot hijack the link (it is bridged as text instead).
func shouldTreatAsPairingCode(alreadyPaired bool) bool {
	return !alreadyPaired
}

// shouldAttemptPairing is true only when a real pending code exists.
// Unpaired "ORDER42"-style text is not treated as a failed pairing attempt.
func shouldAttemptPairing(alreadyPaired, pendingExists bool) bool {
	return shouldTreatAsPairingCode(alreadyPaired) && pendingExists
}

const dcPairingDirectOnly = "Pairing only works in a private 1:1 chat with this bot, not in a group."

const dcPairingChatUnknown = "Could not check this chat. Pair only in a private 1:1 chat with the bot."

// decideDCPairing is the policy for finding 1: lookup result → allow pairing.
func decideDCPairing(isDirect bool, lookupErr error) (ok bool, userMsg string) {
	if lookupErr != nil {
		return false, dcPairingChatUnknown
	}
	if !isDirect {
		return false, dcPairingDirectOnly
	}
	return true, ""
}

func dcPairingAllowed(sess *dc.Session, accID, chatID uint32) (bool, string) {
	if sess == nil {
		return false, dcPairingChatUnknown
	}
	ok, err := sess.IsDirectChat(accID, chatID)
	return decideDCPairing(ok, err)
}
