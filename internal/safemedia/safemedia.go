// Package safemedia rejects files that can crash or exploit the host
// (executables, polyglots, decompression bombs). It does not scan for NSFW.
package safemedia

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	"golang.org/x/image/webp"
)

const (
	// DefaultMaxBytes is the hard cap for any download or attachment on disk.
	DefaultMaxBytes = 20 << 20 // 20 MiB
	// AvatarMaxBytes is the cap for Telegram/DC profile photos.
	AvatarMaxBytes = 5 << 20 // 5 MiB
	// MaxPixels rejects huge DecodeConfig dimensions (decompression bombs).
	MaxPixels = 16 << 20 // 16 megapixels
	// MaxSide is the max width or height.
	MaxSide = 8192
)

// Class is a sniffed container type.
type Class string

const (
	ClassUnknown Class = ""
	ClassJPEG    Class = "jpeg"
	ClassPNG     Class = "png"
	ClassGIF     Class = "gif"
	ClassWEBP    Class = "webp"
	ClassMP4     Class = "mp4"
	ClassWEBM    Class = "webm"
	ClassTGS     Class = "tgs" // gzip (Telegram Lottie)
	ClassELF     Class = "elf"
	ClassPE      Class = "pe"
	ClassMachO   Class = "macho"
	ClassZIP     Class = "zip"
	ClassHTML    Class = "html"
	ClassSVG     Class = "svg"
	ClassPDF     Class = "pdf"
	ClassScript  Class = "script"
)

// Role is what the caller intends to do with the file.
type Role string

const (
	RoleAvatar  Role = "avatar"
	RoleImage   Role = "image"
	RoleSticker Role = "sticker"
	RoleLottie  Role = "lottie"
	RoleVideo   Role = "video"
	RoleGIF     Role = "gif"
	RoleFile    Role = "file"
)

// Sniff classifies the first bytes of a file.
func Sniff(b []byte) Class {
	if len(b) < 2 {
		return ClassUnknown
	}
	if b[0] == 0x7f && len(b) >= 4 && string(b[1:4]) == "ELF" {
		return ClassELF
	}
	if b[0] == 'M' && b[1] == 'Z' {
		return ClassPE
	}
	if len(b) >= 4 {
		u := binary.BigEndian.Uint32(b[:4])
		if u == 0xFEEDFACE || u == 0xFEEDFACF || u == 0xCEFAEDFE || u == 0xCFFAEDFE || u == 0xCAFEBABE {
			return ClassMachO
		}
		if b[0] == 'P' && b[1] == 'K' && (b[2] == 3 || b[2] == 5 || b[2] == 7) {
			return ClassZIP
		}
		if string(b[:4]) == "%PDF" {
			return ClassPDF
		}
	}
	if b[0] == '#' && b[1] == '!' {
		return ClassScript
	}
	trim := bytes.TrimLeft(b, " \t\r\n")
	if len(trim) >= 5 && (bytes.HasPrefix(bytes.ToLower(trim), []byte("<html")) ||
		bytes.HasPrefix(bytes.ToLower(trim), []byte("<!doc"))) {
		return ClassHTML
	}
	if len(trim) >= 4 && bytes.HasPrefix(bytes.ToLower(trim), []byte("<svg")) {
		return ClassSVG
	}
	if b[0] == 0x1f && b[1] == 0x8b {
		return ClassTGS
	}
	if len(b) >= 3 && b[0] == 0xff && b[1] == 0xd8 && b[2] == 0xff {
		return ClassJPEG
	}
	if len(b) >= 8 && b[0] == 0x89 && string(b[1:4]) == "PNG" {
		return ClassPNG
	}
	if len(b) >= 6 && (string(b[0:6]) == "GIF87a" || string(b[0:6]) == "GIF89a") {
		return ClassGIF
	}
	if len(b) >= 12 && string(b[0:4]) == "RIFF" && string(b[8:12]) == "WEBP" {
		return ClassWEBP
	}
	if len(b) >= 4 && b[0] == 0x1a && b[1] == 0x45 && b[2] == 0xdf && b[3] == 0xa3 {
		return ClassWEBM
	}
	if len(b) >= 8 && string(b[4:8]) == "ftyp" {
		return ClassMP4
	}
	return ClassUnknown
}

func dangerous(c Class) bool {
	switch c {
	case ClassELF, ClassPE, ClassMachO, ClassZIP, ClassHTML, ClassSVG, ClassPDF, ClassScript:
		return true
	default:
		return false
	}
}

func allowed(c Class, role Role) bool {
	if dangerous(c) || c == ClassUnknown {
		return false
	}
	switch role {
	case RoleAvatar:
		return c == ClassJPEG || c == ClassPNG || c == ClassWEBP
	case RoleImage:
		return c == ClassJPEG || c == ClassPNG || c == ClassWEBP || c == ClassGIF
	case RoleSticker:
		return c == ClassWEBP || c == ClassPNG || c == ClassGIF || c == ClassJPEG
	case RoleLottie:
		return c == ClassTGS
	case RoleVideo:
		return c == ClassMP4 || c == ClassWEBM
	case RoleGIF:
		return c == ClassGIF || c == ClassMP4 || c == ClassWEBM
	case RoleFile:
		return c == ClassJPEG || c == ClassPNG || c == ClassWEBP || c == ClassGIF ||
			c == ClassMP4 || c == ClassWEBM || c == ClassTGS
	default:
		return false
	}
}

// CopyLimited copies at most max bytes. Returns an error if the source is larger.
func CopyLimited(dst io.Writer, src io.Reader, max int64) (int64, error) {
	if max <= 0 {
		max = DefaultMaxBytes
	}
	n, err := io.Copy(dst, io.LimitReader(src, max+1))
	if err != nil {
		return n, err
	}
	if n > max {
		return n, fmt.Errorf("file exceeds %d bytes", max)
	}
	return n, nil
}

// ValidateFile checks size, magic, and (for raster images) declared dimensions.
func ValidateFile(path string, role Role, maxBytes int64) error {
	if maxBytes <= 0 {
		if role == RoleAvatar {
			maxBytes = AvatarMaxBytes
		} else {
			maxBytes = DefaultMaxBytes
		}
	}
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.Size() <= 0 {
		return fmt.Errorf("empty file")
	}
	if st.Size() > maxBytes {
		return fmt.Errorf("file too large: %d bytes (max %d)", st.Size(), maxBytes)
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	hdr := make([]byte, 32)
	n, _ := f.Read(hdr)
	c := Sniff(hdr[:n])
	if !allowed(c, role) {
		if dangerous(c) {
			return fmt.Errorf("rejected dangerous file type %s", c)
		}
		return fmt.Errorf("rejected file type %q for %s", c, role)
	}
	if c == ClassJPEG || c == ClassPNG || c == ClassGIF || c == ClassWEBP {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return err
		}
		if err := checkImageDims(f, c); err != nil {
			return err
		}
	}
	return nil
}

func checkImageDims(r io.Reader, c Class) error {
	var cfg image.Config
	var err error
	switch c {
	case ClassWEBP:
		cfg, err = webp.DecodeConfig(r)
	default:
		cfg, _, err = image.DecodeConfig(r)
	}
	if err != nil {
		return fmt.Errorf("invalid image: %w", err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return fmt.Errorf("invalid image dimensions")
	}
	if cfg.Width > MaxSide || cfg.Height > MaxSide {
		return fmt.Errorf("image too large: %dx%d (max side %d)", cfg.Width, cfg.Height, MaxSide)
	}
	if int64(cfg.Width)*int64(cfg.Height) > MaxPixels {
		return fmt.Errorf("image too many pixels: %dx%d", cfg.Width, cfg.Height)
	}
	return nil
}

// RoleFromKind maps a bridge kind string to a validation role.
func RoleFromKind(kind string) Role {
	switch kind {
	case "image":
		return RoleImage
	case "sticker", "custom_emoji":
		return RoleSticker
	case "lottie":
		return RoleLottie
	case "video", "video_sticker":
		return RoleVideo
	case "gif":
		return RoleGIF
	default:
		return RoleFile
	}
}
