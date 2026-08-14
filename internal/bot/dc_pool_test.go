package bot

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDCWorkPoolEnqueueAndRun(t *testing.T) {
	var n atomic.Int32
	var wg sync.WaitGroup
	wg.Add(3)
	p := newDCWorkPool(2, 8, func(accID, msgID uint32) {
		n.Add(1)
		wg.Done()
	})
	if !p.tryEnqueue(1, 10) || !p.tryEnqueue(1, 11) || !p.tryEnqueue(2, 12) {
		t.Fatal("expected enqueue")
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("workers did not run jobs")
	}
	if n.Load() != 3 {
		t.Fatalf("handled %d", n.Load())
	}
}

func TestDCWorkPoolQueueFull(t *testing.T) {
	started := make(chan struct{})
	block := make(chan struct{})
	p := newDCWorkPool(1, 2, func(uint32, uint32) {
		select {
		case <-started:
		default:
			close(started)
		}
		<-block
	})
	if !p.tryEnqueue(1, 1) {
		t.Fatal("first")
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not start")
	}
	if !p.tryEnqueue(1, 2) || !p.tryEnqueue(1, 3) {
		t.Fatal("queue slots")
	}
	if p.tryEnqueue(1, 4) {
		t.Fatal("must drop when full")
	}
	close(block)
}

func TestDCWorkPoolNilRejects(t *testing.T) {
	var p *dcWorkPool
	if p.tryEnqueue(1, 1) {
		t.Fatal("nil pool")
	}
}

func TestDCWorkPoolPanicDoesNotKillWorker(t *testing.T) {
	var n atomic.Int32
	p := newDCWorkPool(1, 4, func(_ uint32, msgID uint32) {
		if msgID == 1 {
			panic("boom")
		}
		n.Add(1)
	})
	if !p.tryEnqueue(1, 1) || !p.tryEnqueue(1, 2) {
		t.Fatal("enqueue")
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if n.Load() == 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("worker should keep running after panic")
}
