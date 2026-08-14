package persona

import "testing"

func TestIsOwnerDirectChat(t *testing.T) {
	t.Parallel()
	if !isOwnerDirectChat(9, 9) {
		t.Fatal("same id")
	}
	if isOwnerDirectChat(9, 10) {
		t.Fatal("unknown group/chat must not be treated as owner 1:1")
	}
	if isOwnerDirectChat(0, 10) {
		t.Fatal("unset owner chat must not fall back to DM")
	}
}
