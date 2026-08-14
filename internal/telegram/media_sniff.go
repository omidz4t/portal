package telegram

import (
	"os"
	"path/filepath"
	"strings"
)

// sniffMediaExt returns a Telegram-friendly extension from file magic bytes.
// ArcaneChat often sends stickers with a null filename; SaveMsgFile then loses the type.
func sniffMediaExt(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	var hdr [16]byte
	n, _ := f.Read(hdr[:])
	if n < 4 {
		return ""
	}
	b := hdr[:n]

	// gzip → Telegram animated sticker (.tgs is gzip Lottie)
	if n >= 2 && b[0] == 0x1f && b[1] == 0x8b {
		return ".tgs"
	}
	// RIFF....WEBP
	if n >= 12 && string(b[0:4]) == "RIFF" && string(b[8:12]) == "WEBP" {
		return ".webp"
	}
	// WebM / Matroska
	if n >= 4 && b[0] == 0x1a && b[1] == 0x45 && b[2] == 0xdf && b[3] == 0xa3 {
		return ".webm"
	}
	// PNG
	if n >= 8 && b[0] == 0x89 && string(b[1:4]) == "PNG" {
		return ".png"
	}
	// JPEG
	if n >= 3 && b[0] == 0xff && b[1] == 0xd8 && b[2] == 0xff {
		return ".jpg"
	}
	// GIF
	if n >= 6 && (string(b[0:6]) == "GIF87a" || string(b[0:6]) == "GIF89a") {
		return ".gif"
	}
	// MP4 / ISO BMFF (ftyp)
	if n >= 8 && string(b[4:8]) == "ftyp" {
		return ".mp4"
	}
	return ""
}

// ensureUploadName returns a path whose basename ends with a Telegram-safe extension.
// If needed, hard-links/copies into the same directory under a corrected name.
// The caller must remove the returned path if it differs from the original (use cleanup).
func ensureUploadName(path, preferredName string, kindHint string) (uploadPath string, uploadName string, cleanup func(), err error) {
	cleanup = func() {}
	ext := strings.ToLower(filepath.Ext(preferredName))
	if ext == "" {
		ext = strings.ToLower(filepath.Ext(path))
	}
	if sniff := sniffMediaExt(path); sniff != "" {
		// Prefer sniffed type over a misleading name (e.g. file.bin).
		if ext == "" || ext == ".bin" || ext == ".dat" || ext == ".tmp" {
			ext = sniff
		}
		// Stickers: force tgs/webp/webm from magic when viewtype says sticker.
		if kindHint == "sticker" || kindHint == "lottie" || kindHint == "video_sticker" {
			switch sniff {
			case ".tgs", ".webp", ".webm", ".png", ".gif":
				ext = sniff
			}
		}
	}
	if ext == "" {
		switch kindHint {
		case "sticker":
			ext = ".webp"
		case "lottie":
			ext = ".tgs"
		case "video_sticker":
			ext = ".webm"
		case "gif":
			ext = ".mp4"
		case "image":
			ext = ".jpg"
		case "video":
			ext = ".mp4"
		default:
			ext = ".bin"
		}
	}

	base := strings.TrimSuffix(filepath.Base(preferredName), filepath.Ext(preferredName))
	if base == "" || base == "file" || base == "file.bin" {
		base = "sticker"
		if kindHint == "image" {
			base = "image"
		}
		if kindHint == "video" || kindHint == "gif" {
			base = "video"
		}
	}
	uploadName = base + ext
	if strings.EqualFold(filepath.Ext(path), ext) {
		return path, uploadName, cleanup, nil
	}

	// Unique name per call so concurrent DC→TG sends do not share upload_sticker.webp.
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "upload_*"+ext)
	if err != nil {
		return "", "", cleanup, err
	}
	uploadPath = tmp.Name()
	_ = tmp.Close()
	if err := os.Remove(uploadPath); err != nil {
		return "", "", cleanup, err
	}
	if err := os.Link(path, uploadPath); err != nil {
		in, err2 := os.ReadFile(path)
		if err2 != nil {
			return "", "", cleanup, err2
		}
		if err2 = os.WriteFile(uploadPath, in, 0o600); err2 != nil {
			return "", "", cleanup, err2
		}
	}
	cleanup = func() { _ = os.Remove(uploadPath) }
	return uploadPath, uploadName, cleanup, nil
}
