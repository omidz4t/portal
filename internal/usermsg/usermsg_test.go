package usermsg

import (
	"fmt"
	"strings"
	"testing"
)

func TestSafeNeverEchoesInternalError(t *testing.T) {
	t.Parallel()
	secrets := []string{
		"/data/tgportal.db",
		"PERSONA_ACCOUNT_QR",
		"UNIQUE constraint",
		"decrypt: wrong key",
		"111:PORTAL-SECRET",
		"enc:v1:AAAA",
	}
	err := fmt.Errorf("sqlite open %s: %s token=%s", secrets[0], secrets[2], secrets[4])
	got := Safe(err)
	if got == "" || got == err.Error() {
		t.Fatalf("must not echo error: %q", got)
	}
	low := strings.ToLower(got)
	for _, s := range secrets {
		if strings.Contains(got, s) || strings.Contains(low, strings.ToLower(s)) {
			t.Fatalf("leaked %q in %q", s, got)
		}
	}
}

func TestPublicConstantsHaveNoInternals(t *testing.T) {
	t.Parallel()
	banned := []string{"sqlite", "rpc", "decrypt", "PERSONA_", ".db", "token", "hmac", "panic"}
	for _, msg := range []string{Generic, PairingFailed, BridgeFailed, RegisterFailed} {
		low := strings.ToLower(msg)
		for _, b := range banned {
			if strings.Contains(low, strings.ToLower(b)) {
				t.Fatalf("%q contains %q", msg, b)
			}
		}
	}
}

func TestSafeNilStillGeneric(t *testing.T) {
	if Safe(nil) != Generic {
		t.Fatal(Safe(nil))
	}
}
