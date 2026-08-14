package dc

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Ensures Session.Do serializes concurrent callers (mutex behavior).
func TestSessionDoSerializes(t *testing.T) {
	s := &Session{}
	var concurrent int32
	var maxConcurrent int32
	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.Do(func() error {
				c := atomic.AddInt32(&concurrent, 1)
				for {
					old := atomic.LoadInt32(&maxConcurrent)
					if c <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, c) {
						break
					}
				}
				time.Sleep(5 * time.Millisecond)
				atomic.AddInt32(&concurrent, -1)
				return nil
			})
		}()
	}
	wg.Wait()
	if maxConcurrent != 1 {
		t.Fatalf("expected max concurrent 1, got %d", maxConcurrent)
	}
}

func TestConfigureAccountFromQREmpty(t *testing.T) {
	s := &Session{}
	if err := s.ConfigureAccountFromQR(1, "  "); err == nil {
		t.Fatal("empty QR must fail before any RPC")
	}
}

func TestDoRunsWhileConfigureWouldNotHoldLock(t *testing.T) {
	// Short SetConfig-style work must not be stuck behind a simulated
	// long unlocked provision (the bug #8 we removed).
	s := &Session{}
	var muHits atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})

	go func() {
		close(started)
		<-release
	}()
	<-started
	if err := s.Do(func() error {
		muHits.Add(1)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if muHits.Load() != 1 {
		t.Fatal("Session.Do must stay usable during unlocked provision")
	}
	close(release)
}
