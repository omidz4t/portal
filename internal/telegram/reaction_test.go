package telegram

import "testing"

func TestTrimReaction(t *testing.T) {
	if got := trimReaction("  ✅  "); got != "✅" {
		t.Fatalf("%q", got)
	}
	if got := trimReaction(""); got != "" {
		t.Fatalf("%q", got)
	}
}
