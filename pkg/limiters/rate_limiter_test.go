package limiters

import (
	"testing"
	"time"
)

func BenchmarkTokenBucket(b *testing.B) {
	limiter := NewTokenBucket(1000000, 100000)

	b.ResetTimer()
	b.ReportAllocs() // Tracks memory bytes & allocations per op

	for b.Loop() {
		limiter.Allow("ola")
	}
}

func BenchmarkSlidingWindowLog(b *testing.B) {
	limiter := NewSlidingWindowLog(1*time.Minute, 1000000)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		limiter.Allow("ola")
	}
}

func BenchmarkSlidingWindowCounter(b *testing.B) {
	limiter := NewSlidingWindowCounter(1*time.Minute, 1000000)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		limiter.Allow("ola")
	}
}

func TestTokenBucket(t *testing.T) {

	tests := map[string][3]float64{
		"test 1": {2, 1, 5},
	}

	for k, test := range tests {
		t.Run(k, func(t *testing.T) {
			limiter := NewTokenBucket(test[0], test[1])

			allowed := true
			for i := 0; i < int(test[2]); i++ {
				allowed = limiter.Allow("ola")
			}

			if allowed {
				t.Error("User not rate limited")
			}
		})
	}
}

func TestSlidingWindowCounter(t *testing.T) {
	tests := map[string][3]int{
		"test 1": {2, 1, 5},
	}

	for k, test := range tests {
		t.Run(k, func(t *testing.T) {
			limiter := NewSlidingWindowCounter(time.Duration(test[0]*int(time.Second)), test[1])

			allowed := true
			for i := 0; i < int(test[2]); i++ {
				allowed = limiter.Allow("ola")
			}

			if allowed {
				t.Error("User not rate limited")
			}
		})
	}
}

func TestSlidingWindowLog(t *testing.T) {
	tests := map[string][3]int{
		"test 1": {2, 1, 5},
	}

	for k, test := range tests {
		t.Run(k, func(t *testing.T) {
			limiter := NewSlidingWindowLog(time.Duration(test[0]*int(time.Second)), test[1])

			allowed := true
			for i := 0; i < int(test[2]); i++ {
				allowed = limiter.Allow("ola")
			}

			if allowed {
				t.Error("User not rate limited")
			}
		})
	}
}
