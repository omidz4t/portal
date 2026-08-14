package telegram

import "fmt"

// tryConvertTGSToGIF is intentionally disabled. Running lottie/ffmpeg on
// attacker-controlled stickers is a host RCE/DoS surface.
func tryConvertTGSToGIF(tgsPath, tmpdir string) (string, error) {
	_ = tgsPath
	_ = tmpdir
	return "", fmt.Errorf("TGS conversion disabled (unsafe on untrusted files)")
}
