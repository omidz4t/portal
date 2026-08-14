package store

import (
	"encoding/hex"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	k, err := hex.DecodeString("00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func TestSealOpenRoundTrip(t *testing.T) {
	c, err := newCrypter(testKey(t))
	if err != nil || c == nil {
		t.Fatal(err)
	}
	s1, err := c.Seal("secret-token")
	if err != nil {
		t.Fatal(err)
	}
	s2, err := c.Seal("secret-token")
	if err != nil {
		t.Fatal(err)
	}
	if s1 == s2 {
		t.Fatal("nonces should differ")
	}
	if !isEncrypted(s1) {
		t.Fatalf("prefix: %s", s1)
	}
	got, err := c.Open(s1)
	if err != nil || got != "secret-token" {
		t.Fatalf("open: %v %q", err, got)
	}
}

func TestOpenTamper(t *testing.T) {
	c, err := newCrypter(testKey(t))
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Seal("x")
	if err != nil {
		t.Fatal(err)
	}
	s = s[:len(s)-2] + "AA"
	if _, err := c.Open(s); err == nil {
		t.Fatal("expected tamper error")
	}
}

func TestHMACStable(t *testing.T) {
	c, err := newCrypter(testKey(t))
	if err != nil {
		t.Fatal(err)
	}
	a := c.HMACCode("abc12xyz")
	b := c.HMACCode("ABC12XYZ")
	if a == "" || a != b {
		t.Fatalf("hmac %s %s", a, b)
	}
}

func TestParseKey(t *testing.T) {
	k, err := ParseKey("00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")
	if err != nil || len(k) != 32 {
		t.Fatalf("%v %d", err, len(k))
	}
	if _, err := ParseKey("short"); err == nil {
		t.Fatal("expected error")
	}
	k, err = ParseKey("")
	if err != nil || k != nil {
		t.Fatal("empty key")
	}
}
