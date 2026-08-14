package telegram

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureUploadNameConcurrentNoClash(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "blob.bin")
	// RIFF/WEBP magic so sniff forces .webp while the path stays .bin.
	webp := []byte("RIFF\x00\x00\x00\x00WEBP")
	if err := os.WriteFile(src, webp, 0o600); err != nil {
		t.Fatal(err)
	}

	p1, n1, c1, err := ensureUploadName(src, "file.bin", "sticker")
	if err != nil {
		t.Fatal(err)
	}
	defer c1()
	p2, n2, c2, err := ensureUploadName(src, "file.bin", "sticker")
	if err != nil {
		t.Fatal(err)
	}
	defer c2()

	if p1 == p2 {
		t.Fatalf("paths must be unique: %s", p1)
	}
	if n1 != "sticker.webp" || n2 != "sticker.webp" {
		t.Fatalf("telegram names: %q %q", n1, n2)
	}
	if _, err := os.Stat(p1); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p2); err != nil {
		t.Fatal(err)
	}

	c1()
	if _, err := os.Stat(p1); !os.IsNotExist(err) {
		t.Fatal("cleanup must remove only first upload")
	}
	if _, err := os.Stat(p2); err != nil {
		t.Fatal("second upload must survive first cleanup")
	}
}

func TestEnsureUploadNameKeepsMatchingExt(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "image.jpg")
	if err := os.WriteFile(src, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	p, name, cleanup, err := ensureUploadName(src, "image.jpg", "image")
	if err != nil {
		t.Fatal(err)
	}
	cleanup()
	if p != src {
		t.Fatalf("same ext should reuse path: %s", p)
	}
	if name != "image.jpg" {
		t.Fatalf("name %q", name)
	}
}
