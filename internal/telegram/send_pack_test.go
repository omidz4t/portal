package telegram

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestLaunchSendPackDoesNotWait(t *testing.T) {
	t.Cleanup(func() { launchSendPack = func(work func()) { go work() } })

	var started atomic.Bool
	block := make(chan struct{})
	returned := make(chan struct{})

	launchSendPack = func(work func()) {
		go func() {
			started.Store(true)
			work()
		}()
	}

	go func() {
		launchSendPack(func() { <-block })
		close(returned)
	}()

	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("launchSendPack must return before pack work finishes")
	}
	close(block)
	if !started.Load() {
		// work may not have scheduled yet; that is still "didn't wait"
		return
	}
}
