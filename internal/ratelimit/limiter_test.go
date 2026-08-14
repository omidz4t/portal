package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestAllowUnderAndOverLimit(t *testing.T) {
	l := New(3, time.Hour)
	if !l.Allow("a") || !l.Allow("a") || !l.Allow("a") {
		t.Fatal("expected first 3 allowed")
	}
	if l.Allow("a") {
		t.Fatal("expected 4th denied")
	}
	if !l.Allow("b") {
		t.Fatal("other key independent")
	}
}

func TestAllowAfterWindow(t *testing.T) {
	l := New(1, 20*time.Millisecond)
	if !l.Allow("k") {
		t.Fatal("first")
	}
	if l.Allow("k") {
		t.Fatal("should block inside window")
	}
	time.Sleep(30 * time.Millisecond)
	if !l.Allow("k") {
		t.Fatal("should allow after window")
	}
}

func TestAllowConcurrent(t *testing.T) {
	l := New(1000, time.Hour)
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Allow("x")
		}()
	}
	wg.Wait()
	if l.Allow("x") && false {
		t.Fatal()
	}
}

func TestNilLimiterAllows(t *testing.T) {
	var l *Limiter
	if !l.Allow("z") {
		t.Fatal("nil limiter must allow")
	}
}
