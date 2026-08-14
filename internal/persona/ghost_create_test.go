package persona

import (
	"errors"
	"fmt"
	"testing"
)

type fakeGhostCore struct {
	nextID     uint32
	added      []uint32
	removed    []uint32
	failAdd    bool
	failStep   string
	failRemove bool
}

func (f *fakeGhostCore) AddAccount() (uint32, error) {
	if f.failAdd {
		return 0, errors.New("add denied")
	}
	f.nextID++
	id := f.nextID
	f.added = append(f.added, id)
	return id, nil
}

func (f *fakeGhostCore) ConfigureAccountFromQR(accID uint32, qr string) error {
	if f.failStep == "configure" {
		return errors.New("configure failed")
	}
	return nil
}

func (f *fakeGhostCore) SetDisplayName(uint32, string) error { return nil }
func (f *fakeGhostCore) SetSelfAvatar(uint32, string) error  { return nil }

func (f *fakeGhostCore) ImportOwnerAsKeyContact(uint32, string) (uint32, uint32, error) {
	if f.failStep == "import" {
		return 0, 0, errors.New("import failed")
	}
	return 1, 1, nil
}

func (f *fakeGhostCore) AccountAddress(uint32) (string, error) {
	return "ghost@example.org", nil
}

func (f *fakeGhostCore) RemoveAccount(accID uint32) error {
	if f.failRemove {
		return fmt.Errorf("remove %d failed", accID)
	}
	f.removed = append(f.removed, accID)
	return nil
}

func TestWithNewGhostAccountRollbackOnFailure(t *testing.T) {
	core := &fakeGhostCore{failStep: "configure"}
	id, err := withNewGhostAccount(core, func(accID uint32) error {
		return core.ConfigureAccountFromQR(accID, "qr")
	})
	if err == nil || id != 0 {
		t.Fatalf("expected fail: id=%d err=%v", id, err)
	}
	if len(core.added) != 1 || len(core.removed) != 1 || core.added[0] != core.removed[0] {
		t.Fatalf("must remove the added account: added=%v removed=%v", core.added, core.removed)
	}
}

func TestWithNewGhostAccountRollbackOnImportFailure(t *testing.T) {
	core := &fakeGhostCore{failStep: "import"}
	_, err := withNewGhostAccount(core, func(accID uint32) error {
		if err := core.ConfigureAccountFromQR(accID, "qr"); err != nil {
			return err
		}
		_, _, err := core.ImportOwnerAsKeyContact(accID, "vcard")
		return err
	})
	if err == nil {
		t.Fatal("expected import error")
	}
	if len(core.removed) != 1 || core.removed[0] != core.added[0] {
		t.Fatalf("import fail must remove acc: added=%v removed=%v", core.added, core.removed)
	}
}

func TestWithNewGhostAccountSuccessDoesNotRemove(t *testing.T) {
	core := &fakeGhostCore{}
	id, err := withNewGhostAccount(core, func(accID uint32) error {
		return core.ConfigureAccountFromQR(accID, "qr")
	})
	if err != nil || id != 1 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if len(core.removed) != 0 {
		t.Fatalf("success must keep account: removed=%v", core.removed)
	}
}

func TestWithNewGhostAccountAddFailureNoRemove(t *testing.T) {
	core := &fakeGhostCore{failAdd: true}
	id, err := withNewGhostAccount(core, func(uint32) error { return nil })
	if err == nil || id != 0 {
		t.Fatal("expected add error")
	}
	if len(core.removed) != 0 {
		t.Fatalf("nothing to remove: %v", core.removed)
	}
}

func TestWithNewGhostAccountReportsRemoveFailure(t *testing.T) {
	core := &fakeGhostCore{failStep: "configure", failRemove: true}
	_, err := withNewGhostAccount(core, func(accID uint32) error {
		return core.ConfigureAccountFromQR(accID, "qr")
	})
	if err == nil || err.Error() == "configure failed" {
		t.Fatalf("should wrap original and remove error: %v", err)
	}
}

func TestFailedCreatesDoNotAccumulateAccounts(t *testing.T) {
	core := &fakeGhostCore{failStep: "import"}
	for i := 0; i < 5; i++ {
		_, err := withNewGhostAccount(core, func(accID uint32) error {
			_, _, err := core.ImportOwnerAsKeyContact(accID, "v")
			return err
		})
		if err == nil {
			t.Fatal("expected fail")
		}
	}
	if len(core.added) != 5 || len(core.removed) != 5 {
		t.Fatalf("retries must not leak accounts: added=%d removed=%d", len(core.added), len(core.removed))
	}
	live := len(core.added) - len(core.removed)
	if live != 0 {
		t.Fatalf("leaked %d accounts (would bypass max_ghosts)", live)
	}
}
