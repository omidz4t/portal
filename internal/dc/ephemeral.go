package dc

import (
	"os"
	"strings"
)

// RemoveEphemeralFile deletes a local cache/blob copy after it has been
// forwarded. Never touches branding assets or unrelated paths.
func RemoveEphemeralFile(path string) {
	if path == "" {
		return
	}
	if !isEphemeralMediaPath(path) {
		return
	}
	_ = os.Remove(path)
}

func isEphemeralMediaPath(path string) bool {
	p := strings.ReplaceAll(path, "\\", "/")
	switch {
	case strings.Contains(p, "/dc.db-blobs/"):
		return true
	case strings.Contains(p, "/tg-cache/"):
		return true
	case strings.Contains(p, "/dc-cache/"):
		return true
	default:
		return false
	}
}

// ForgetMessageFile drops the on-disk attachment for a DC message after
// the bytes have already been sent (Telegram or SMTP).
func (s *Session) ForgetMessageFile(accID, msgID uint32) {
	msg, err := s.GetMessage(accID, msgID)
	if err != nil || msg.File == nil {
		return
	}
	RemoveEphemeralFile(*msg.File)
}

// ApplyShortDeviceRetention asks core to drop local messages quickly so
// stickers/videos do not accumulate under dc.db-blobs.
func (s *Session) ApplyShortDeviceRetention(seconds string) error {
	if seconds == "" {
		seconds = "60"
	}
	ids, err := s.AllConfiguredAccounts()
	if err != nil {
		return err
	}
	for _, id := range ids {
		sec := seconds
		if err := s.SetConfig(id, "delete_device_after", &sec); err != nil {
			return err
		}
	}
	return nil
}

// IsDCBlobPath reports whether path is a disposable cache/blob file.
func IsDCBlobPath(path string) bool {
	return isEphemeralMediaPath(path)
}
