package rater

import (
	"sync"
	"time"
)

type SlidingWindowCounter struct {
	mu           sync.Mutex
	window       time.Duration
	maxRequests  int
	prevCount    int
	currentCount int
	windowStart  time.Time
}

func NewSlidingWindowCounter(window time.Duration, maxRequest int) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		window:       window,
		maxRequests:  maxRequest,
		prevCount:    0,
		currentCount: 0,
		windowStart:  time.Time{},
	}
}

func (sc *SlidingWindowCounter) Allow(userKey string) bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := time.Now()

	elapsed := now.Sub(sc.windowStart)

	if elapsed >= 2*sc.window {
		sc.prevCount = 0
		sc.currentCount = 0
		sc.windowStart = now
	} else if elapsed >= sc.window {
		sc.prevCount = sc.currentCount
		sc.currentCount = 0
		sc.windowStart = sc.windowStart.Add(sc.window)
	}

	elapsed = now.Sub(sc.windowStart)
	weight := float64(sc.window-elapsed) / float64(sc.window)
	estimatedCount := (sc.prevCount * int(weight)) + sc.currentCount

	if estimatedCount < sc.maxRequests {
		sc.currentCount++
		return true
	}

	return false
}
