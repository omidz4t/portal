package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/hkdf"
)

const (
	encPrefix   = "enc:v1:"
	hkdfInfo    = "tgportal-sqlite-v1"
	aesKeyBytes = 32
)

// crypter seals strings with AES-256-GCM and HMACs pairing codes for lookup.
type crypter struct {
	aead cipher.AEAD
	mac  []byte
}

// ParseKey decodes a 32-byte database key from hex or base64. Empty input yields nil, nil.
func ParseKey(s string) ([]byte, error) {
	return parseDatabaseKey(s)
}

func parseDatabaseKey(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if b, err := hex.DecodeString(s); err == nil && len(b) == aesKeyBytes {
		return b, nil
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil && len(b) == aesKeyBytes {
		return b, nil
	}
	if b, err := base64.RawStdEncoding.DecodeString(s); err == nil && len(b) == aesKeyBytes {
		return b, nil
	}
	return nil, fmt.Errorf("database key must be 32 bytes as hex or base64")
}

func newCrypter(rawKey []byte) (*crypter, error) {
	if len(rawKey) == 0 {
		return nil, nil
	}
	r := hkdf.New(sha256.New, rawKey, nil, []byte(hkdfInfo))
	derived := make([]byte, aesKeyBytes+aesKeyBytes)
	if _, err := io.ReadFull(r, derived); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(derived[:aesKeyBytes])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &crypter{aead: aead, mac: derived[aesKeyBytes:]}, nil
}

func isEncrypted(s string) bool {
	return strings.HasPrefix(s, encPrefix)
}

func (c *crypter) Seal(plain string) (string, error) {
	if c == nil {
		return plain, nil
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	out := c.aead.Seal(nonce, nonce, []byte(plain), nil)
	return encPrefix + base64.StdEncoding.EncodeToString(out), nil
}

func (c *crypter) Open(stored string) (string, error) {
	if stored == "" {
		return "", nil
	}
	if !isEncrypted(stored) {
		return stored, nil
	}
	if c == nil {
		return "", fmt.Errorf("encrypted value in database but no TGPORTAL_DB_KEY")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, encPrefix))
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	ns := c.aead.NonceSize()
	if len(raw) < ns {
		return "", fmt.Errorf("decrypt: ciphertext too short")
	}
	plain, err := c.aead.Open(nil, raw[:ns], raw[ns:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: wrong key or corrupt data")
	}
	return string(plain), nil
}

func (c *crypter) HMACCode(code string) string {
	if c == nil {
		return ""
	}
	mac := hmac.New(sha256.New, c.mac)
	_, _ = mac.Write([]byte(normalizeCode(code)))
	return hex.EncodeToString(mac.Sum(nil))
}
