package persona

import "strings"

// IsPortalBotToken reports whether candidate is the official portal Telegram token.
// Registering that token as a persona bot would steal getUpdates from the portal.
func IsPortalBotToken(candidate, portal string) bool {
	c := strings.TrimSpace(candidate)
	p := strings.TrimSpace(portal)
	return c != "" && p != "" && c == p
}
