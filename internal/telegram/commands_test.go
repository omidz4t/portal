package telegram

import (
	"strings"
	"testing"

	"github.com/omidz4t/portal/internal/config"
)

func TestSupportedCommandsIncludeDelete(t *testing.T) {
	want := map[string]bool{
		"delete_my_data":         false,
		"delete_my_data_approve": false,
	}
	for _, c := range supportedCommands {
		if _, ok := want[c.Command]; ok {
			want[c.Command] = true
		}
	}
	for cmd, seen := range want {
		if !seen {
			t.Fatalf("missing bot menu command %s", cmd)
		}
	}
}

func TestCommandsHelpTextIncludesTwoStepDelete(t *testing.T) {
	text := commandsHelpText(config.Config{})
	for _, s := range []string{"/delete_my_data", "/delete_my_data_approve"} {
		if !strings.Contains(text, s) {
			t.Fatalf("help missing %s:\n%s", s, text)
		}
	}
}
