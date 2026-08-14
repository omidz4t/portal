package bot

import (
	"errors"
	"testing"
)

func TestShouldTreatAsPairingCode(t *testing.T) {
	t.Parallel()
	if !shouldTreatAsPairingCode(false) {
		t.Fatal("unpaired chat must accept pairing codes")
	}
	if shouldTreatAsPairingCode(true) {
		t.Fatal("already-paired chat must not treat text as a pairing code")
	}
}

func TestShouldAttemptPairing(t *testing.T) {
	t.Parallel()
	if shouldAttemptPairing(false, false) {
		t.Fatal("ORDER42 with no pending row must not attempt pairing")
	}
	if !shouldAttemptPairing(false, true) {
		t.Fatal("real pending code on unpaired chat")
	}
	if shouldAttemptPairing(true, true) {
		t.Fatal("already paired")
	}
}

func TestDecideDCPairing(t *testing.T) {
	t.Parallel()
	ok, msg := decideDCPairing(true, nil)
	if !ok || msg != "" {
		t.Fatalf("1:1 must allow: ok=%v msg=%q", ok, msg)
	}
	ok, msg = decideDCPairing(false, nil)
	if ok || msg != dcPairingDirectOnly {
		t.Fatalf("group must refuse: ok=%v msg=%q", ok, msg)
	}
	ok, msg = decideDCPairing(true, errors.New("rpc down"))
	if ok || msg != dcPairingChatUnknown {
		t.Fatalf("lookup error must fail closed: ok=%v msg=%q", ok, msg)
	}
	ok, msg = decideDCPairing(false, errors.New("rpc down"))
	if ok {
		t.Fatal("error + group must not allow")
	}
}

func TestDCPairingAllowedNilSession(t *testing.T) {
	ok, msg := dcPairingAllowed(nil, 1, 2)
	if ok || msg != dcPairingChatUnknown {
		t.Fatalf("nil session: ok=%v msg=%q", ok, msg)
	}
}
