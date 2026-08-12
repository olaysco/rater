package rater

type Limiter interface {
	Allow(key string) bool
}
