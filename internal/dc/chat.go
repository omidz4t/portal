package dc

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
)

// IsPairableChatType is true only for 1:1 chats. Groups leak pairing codes
// and would bind every member's mail to one Telegram user.
func IsPairableChatType(t deltachat.ChatType) bool {
	return t == deltachat.ChatTypeSingle
}

// EmailFromInviteURL extracts the contact address from i.delta.chat invite fragments.
func EmailFromInviteURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	frag := u.Fragment
	if frag == "" {
		return "", fmt.Errorf("invite URL has no fragment")
	}

	parts := strings.Split(frag, "&")
	values := url.Values{}
	for _, p := range parts {
		if !strings.Contains(p, "=") {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		k, v := kv[0], ""
		if len(kv) == 2 {
			v, _ = url.QueryUnescape(kv[1])
		}
		values.Set(k, v)
	}
	addr := values.Get("a")
	if addr == "" {
		return "", fmt.Errorf("invite fragment missing a= address")
	}
	return addr, nil
}
