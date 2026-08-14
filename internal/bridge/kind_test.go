package bridge

import (
	"testing"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
)

func TestKindFromFilenameDoesNotAssumeImage(t *testing.T) {
	t.Parallel()
	if KindFromFilename("file.mp4") != KindGif && KindFromFilename("clip.mp4") != KindGif {
		// ViewtypeFile + .mp4 maps to KindGif in KindFromDCViewtype
		if g := KindFromFilename("file.mp4"); g != KindGif {
			t.Fatalf("mp4: %q", g)
		}
	}
	if KindFromFilename("a.webp") != KindSticker {
		t.Fatalf("webp: %q", KindFromFilename("a.webp"))
	}
	if KindFromFilename("a.webm") != KindVideoSticker {
		t.Fatalf("webm: %q", KindFromFilename("a.webm"))
	}
	if KindFromFilename("a.tgs") != KindLottie {
		t.Fatalf("tgs: %q", KindFromFilename("a.tgs"))
	}
	if KindFromFilename("file.bin") != "" {
		t.Fatalf("bin must stay generic: %q", KindFromFilename("file.bin"))
	}
	_ = deltachat.ViewtypeFile
}
