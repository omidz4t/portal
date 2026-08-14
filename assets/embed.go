// Package assets embeds branding files into the release binary so a single
// executable can run without a separate assets/ directory on disk.
package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// FS holds embedded brand assets (logo, start animation, avatar).
//
//go:embed logo.jpg avatar.png poster.jpg start_black_hole.mp4
var FS embed.FS

// Names of known embedded files.
const (
	LogoJPG          = "logo.jpg"
	AvatarPNG        = "avatar.png"
	PosterJPG        = "poster.jpg"
	StartBlackHoleMP4 = "start_black_hole.mp4"
)

// Ensure extracts name from the embed FS into destDir if missing or empty.
// Returns the absolute path of the on-disk file.
func Ensure(name, destDir string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("assets: empty name")
	}
	if destDir == "" {
		destDir = "assets"
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	dest := filepath.Join(destDir, filepath.Base(name))
	if st, err := os.Stat(dest); err == nil && st.Size() > 0 {
		return filepath.Abs(dest)
	}

	data, err := FS.ReadFile(filepath.Base(name))
	if err != nil {
		return "", fmt.Errorf("assets: embed %q: %w", name, err)
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return "", err
	}
	return filepath.Abs(dest)
}

// MaterializeAll writes every embedded asset into destDir (skip if present).
func MaterializeAll(destDir string) error {
	return fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path == "." {
			return nil
		}
		_, err = Ensure(path, destDir)
		return err
	})
}

// ResolvePath returns an absolute path for a configured asset location.
// Order: existing disk path → extract embedded file into destDir.
func ResolvePath(configured, embedName, destDir string) (string, error) {
	if configured != "" {
		if filepath.IsAbs(configured) {
			if st, err := os.Stat(configured); err == nil && st.Size() > 0 {
				return configured, nil
			}
		} else {
			abs, err := filepath.Abs(configured)
			if err == nil {
				if st, err := os.Stat(abs); err == nil && st.Size() > 0 {
					return abs, nil
				}
			}
		}
	}
	// Fall back to embedded asset.
	name := embedName
	if name == "" {
		name = filepath.Base(configured)
	}
	return Ensure(name, destDir)
}
