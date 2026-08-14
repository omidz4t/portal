package telegram

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestIsPrivatePairingChat(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		typ  string
		want bool
	}{
		{"private", "private", true},
		{"group", "group", false},
		{"supergroup", "supergroup", false},
		{"channel", "channel", false},
		{"empty", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: -100, Type: tc.typ}}
			if tc.typ == "private" {
				msg.Chat.ID = 42
			}
			if got := IsPrivatePairingChat(msg); got != tc.want {
				t.Fatalf("type %q: got %v want %v", tc.typ, got, tc.want)
			}
		})
	}
	if IsPrivatePairingChat(nil) {
		t.Fatal("nil message must not be private")
	}
}
