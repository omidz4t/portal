package safemedia

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 1×1 JPEG (public domain test fixture).
var jpeg1x1 = []byte{
	0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00, 0x01,
	0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xff, 0xdb, 0x00, 0x43,
	0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
	0x09, 0x08, 0x0a, 0x0c, 0x14, 0x0d, 0x0c, 0x0b, 0x0b, 0x0c, 0x19, 0x12,
	0x13, 0x0f, 0x14, 0x1d, 0x1a, 0x1f, 0x1e, 0x1d, 0x1a, 0x1c, 0x1c, 0x20,
	0x24, 0x2e, 0x27, 0x20, 0x22, 0x2c, 0x23, 0x1c, 0x1c, 0x28, 0x37, 0x29,
	0x2c, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1f, 0x27, 0x39, 0x3d, 0x38, 0x32,
	0x3c, 0x2e, 0x33, 0x34, 0x32, 0xff, 0xc0, 0x00, 0x0b, 0x08, 0x00, 0x01,
	0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xff, 0xc4, 0x00, 0x14, 0x00, 0x01,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x09, 0xff, 0xc4, 0x00, 0x14, 0x10, 0x01, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0xff, 0xda, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3f, 0x00,
	0x7f, 0x3f, 0xff, 0xd9,
}

func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSniffDangerous(t *testing.T) {
	cases := []struct {
		b    []byte
		want Class
	}{
		{[]byte{0x7f, 'E', 'L', 'F', 1, 1, 1}, ClassELF},
		{[]byte{'M', 'Z', 0x90, 0x00}, ClassPE},
		{[]byte("%PDF-1.7"), ClassPDF},
		{[]byte("#!/bin/sh\n"), ClassScript},
		{[]byte("<html><script>"), ClassHTML},
		{[]byte("  <svg xmlns="), ClassSVG},
		{[]byte{'P', 'K', 3, 4, 0, 0}, ClassZIP},
		{jpeg1x1, ClassJPEG},
		{[]byte{0x1f, 0x8b, 0x08, 0x00}, ClassTGS},
	}
	for _, tc := range cases {
		if g := Sniff(tc.b); g != tc.want {
			t.Errorf("Sniff %q = %s want %s", tc.b[:min(8, len(tc.b))], g, tc.want)
		}
	}
}

func TestValidateJPEGAvatar(t *testing.T) {
	p := writeTemp(t, "a.jpg", jpeg1x1)
	if err := ValidateFile(p, RoleAvatar, 0); err != nil {
		t.Fatal(err)
	}
}

func TestRejectELFAsAvatar(t *testing.T) {
	p := writeTemp(t, "a.jpg", []byte{0x7f, 'E', 'L', 'F', 2, 1, 1, 0})
	err := ValidateFile(p, RoleAvatar, 0)
	if err == nil || !strings.Contains(err.Error(), "dangerous") {
		t.Fatalf("got %v", err)
	}
}

func TestRejectHTMLAsImage(t *testing.T) {
	p := writeTemp(t, "x.jpg", []byte("<!DOCTYPE html><html>"))
	if err := ValidateFile(p, RoleImage, 0); err == nil {
		t.Fatal("expected reject")
	}
}

func TestRejectTGSAsAvatar(t *testing.T) {
	p := writeTemp(t, "a.jpg", []byte{0x1f, 0x8b, 0x08, 0x00})
	if err := ValidateFile(p, RoleAvatar, 0); err == nil {
		t.Fatal("gzip must not be an avatar")
	}
}

func TestAllowTGSAsLottie(t *testing.T) {
	p := writeTemp(t, "s.tgs", []byte{0x1f, 0x8b, 0x08, 0x00, 0x00})
	if err := ValidateFile(p, RoleLottie, 0); err != nil {
		t.Fatal(err)
	}
}

func TestRejectHugePNGDimensions(t *testing.T) {
	// Minimal PNG signature + IHDR with 100000×100000 (no need for valid CRC for DecodeConfig to read dims).
	var buf bytes.Buffer
	buf.Write([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a})
	// IHDR chunk
	var ihdr [13]byte
	binary.BigEndian.PutUint32(ihdr[0:4], 100000)
	binary.BigEndian.PutUint32(ihdr[4:8], 100000)
	ihdr[8] = 8
	ihdr[9] = 2
	var lenb [4]byte
	binary.BigEndian.PutUint32(lenb[:], 13)
	buf.Write(lenb[:])
	buf.WriteString("IHDR")
	buf.Write(ihdr[:])
	buf.Write([]byte{0, 0, 0, 0}) // crc ignored if decode fails; if it fails we still reject
	p := writeTemp(t, "bomb.png", buf.Bytes())
	err := ValidateFile(p, RoleImage, 0)
	if err == nil {
		t.Fatal("expected pixel bomb reject")
	}
}

func TestCopyLimited(t *testing.T) {
	var out bytes.Buffer
	n, err := CopyLimited(&out, bytes.NewReader(bytes.Repeat([]byte("a"), 100)), 50)
	if err == nil || n <= 50 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	out.Reset()
	n, err = CopyLimited(&out, bytes.NewReader([]byte("hi")), 50)
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

func TestRoleFromKind(t *testing.T) {
	if RoleFromKind("image") != RoleImage || RoleFromKind("lottie") != RoleLottie {
		t.Fatal("map")
	}
}
