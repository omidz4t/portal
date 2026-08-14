package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a fixed-window counter keyed by string (in-process only).
type Limiter struct {
	mu   sync.Mutex
	max  int
	win  time.Duration
	hits map[string][]time.Time
}

func New(max int, window time.Duration) *Limiter {
	if max < 1 {
		max = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &Limiter{max: max, win: window, hits: make(map[string][]time.Time)}
}

// Allow reports whether key may proceed and records the hit when allowed.
func (l *Limiter) Allow(key string) bool {
	if l == nil {
		return true
	}
	now := time.Now()
	cut := now.Add(-l.win)
	l.mu.Lock()
	defer l.mu.Unlock()
	stamps := l.hits[key]
	n := 0
	for _, t := range stamps {
		if t.After(cut) {
			stamps[n] = t
			n++
		}
	}
	stamps = stamps[:n]
	if len(stamps) >= l.max {
		l.hits[key] = stamps
		return false
	}
	l.hits[key] = append(stamps, now)
	if len(l.hits) > 10_000 {
		l.pruneLocked(cut)
	}
	return true
}

func (l *Limiter) pruneLocked(cut time.Time) {
	for k, stamps := range l.hits {
		keep := stamps[:0]
		for _, t := range stamps {
			if t.After(cut) {
				keep = append(keep, t)
			}
		}
		if len(keep) == 0 {
			delete(l.hits, k)
		} else {
			l.hits[k] = keep
		}
	}
}

// Set holds the named limiters for a single tgportal process.
type Set struct {
	PairTG    *Limiter
	ClaimTG   *Limiter
	ClaimDC   *Limiter
	BridgeTG  *Limiter
	PairBotTG *Limiter
}

func Defaults() *Set {
	return &Set{
		PairTG:    New(5, 15*time.Minute),
		ClaimTG:   New(8, 15*time.Minute),
		ClaimDC:   New(8, 15*time.Minute),
		BridgeTG:  New(30, time.Minute),
		PairBotTG: New(3, time.Hour),
	}
}
