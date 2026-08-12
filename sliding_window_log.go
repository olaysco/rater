package rater

import (
	"sync"
	"time"
)

type SlidingWindowLog struct {
	mu          sync.Mutex
	window      time.Duration
	maxRequests int
	logs        map[string][]time.Time
}

func NewSlidingWindowLog(window time.Duration, maxRequests int) *SlidingWindowLog {
	return &SlidingWindowLog{
		window:      window,
		maxRequests: maxRequests,
		logs:        make(map[string][]time.Time),
	}
}

func (sw *SlidingWindowLog) Allow(userKey string) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	trimFrom := 0

	userLog, ok := sw.logs[userKey]
	if !ok {
		userLog = make([]time.Time, 0)
	}

	windowStart := now.Add(-sw.window)

	for i, t := range userLog {
		if t.Before(windowStart) {
			trimFrom = i + 1
		}

		if t.After(windowStart) {
			break
		}
	}
	userLog = userLog[trimFrom:]

	if len(userLog) < sw.maxRequests {
		userLog = append(userLog, now)
		sw.logs[userKey] = userLog
		return true
	}

	sw.logs[userKey] = userLog
	return false
}
