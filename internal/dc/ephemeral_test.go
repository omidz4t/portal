package dc

import "testing"

func TestIsEphemeralMediaPath(t *testing.T) {
	keep := []string{
		"/var/lib/tgportal/assets/logo.jpg",
		"/var/lib/tgportal/assets/start_black_hole.mp4",
		"/etc/tgportal/config.yml",
	}
	drop := []string{
		"/var/lib/tgportal/accounts/x/dc.db-blobs/abc.webm",
		"/var/lib/tgportal/tg-cache/1_sticker.webp",
		"/var/lib/tgportal/dc-cache/1_2_video.mp4",
	}
	for _, p := range keep {
		if isEphemeralMediaPath(p) {
			t.Fatalf("must keep %s", p)
		}
	}
	for _, p := range drop {
		if !isEphemeralMediaPath(p) {
			t.Fatalf("must drop %s", p)
		}
	}
}
