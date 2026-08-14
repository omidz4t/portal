package dc

import (
	"testing"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
)

func TestEmailFromInviteURL(t *testing.T) {
	raw := "https://i.delta.chat/#5BE3EBDFB8F5107F8B235A9904030E2852688056&v=3&i=HiTL2iKLGxgkzh9O_2Tk3m79&s=zbQ2CTJiloahSq_2DOovDMOi&a=pyqlx1a9i%40nine.testrun.org&n=omid"
	addr, err := EmailFromInviteURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := "pyqlx1a9i@nine.testrun.org"
	if addr != want {
		t.Fatalf("got %q want %q", addr, want)
	}
}

func TestIsPairableChatType(t *testing.T) {
	t.Parallel()
	if !IsPairableChatType(deltachat.ChatTypeSingle) {
		t.Fatal("Single must be pairable")
	}
	for _, typ := range []deltachat.ChatType{
		deltachat.ChatTypeGroup,
		deltachat.ChatTypeMailinglist,
		deltachat.ChatTypeOutBroadcast,
		deltachat.ChatTypeInBroadcast,
		"",
	} {
		if IsPairableChatType(typ) {
			t.Fatalf("%q must not be pairable", typ)
		}
	}
}
