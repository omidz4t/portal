package persona

import "testing"

func TestIsPortalBotToken(t *testing.T) {
	t.Parallel()
	const portal = "111:PORTAL-SECRET"
	if !IsPortalBotToken(portal, portal) {
		t.Fatal("same token must match")
	}
	if !IsPortalBotToken("  "+portal+"  ", portal) {
		t.Fatal("trim spaces")
	}
	if IsPortalBotToken("222:other", portal) {
		t.Fatal("other bot must not match")
	}
	if IsPortalBotToken(portal, "") {
		t.Fatal("empty portal token must not match")
	}
	if IsPortalBotToken("", portal) {
		t.Fatal("empty candidate must not match")
	}
}
