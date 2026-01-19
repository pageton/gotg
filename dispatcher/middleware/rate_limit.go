package middleware

import (
	"sync"
	"time"

	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher"
)

// RateLimitConfig configures rate limiting behavior
type RateLimitConfig struct {
	// RequestsPerSecond is the maximum number of requests per second per user
	RequestsPerSecond float64
	// Burst is the maximum number of requests allowed in a burst
	Burst int
	// KeyFunc generates a unique key for rate limiting (e.g., user ID, chat ID)
	KeyFunc func(*adapter.Update) string
}

// RateLimiter manages rate limiting for updates
type RateLimiter struct {
	config *RateLimitConfig
	tokens map[string]*tokenBucket
	mu     sync.RWMutex
}

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	if config.RequestsPerSecond <= 0 {
		config.RequestsPerSecond = 1.0
	}
	if config.Burst <= 0 {
		config.Burst = int(config.RequestsPerSecond)
	}
	if config.KeyFunc == nil {
		config.KeyFunc = func(u *adapter.Update) string {
			return "global"
		}
	}

	return &RateLimiter{
		config: config,
		tokens: make(map[string]*tokenBucket),
	}
}

// Middleware returns a dispatcher handler that implements rate limiting
func (rl *RateLimiter) Middleware() dispatcher.Handler {
	return &rateLimitHandler{
		limiter: rl,
	}
}

type rateLimitHandler struct {
	limiter *RateLimiter
}

func (rh *rateLimitHandler) CheckUpdate(ctx *adapter.Context, update *adapter.Update) error {
	key := rh.limiter.config.KeyFunc(update)

	rh.limiter.mu.Lock()
	bucket, exists := rh.limiter.tokens[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     float64(rh.limiter.config.Burst),
			lastRefill: time.Now(),
		}
		rh.limiter.tokens[key] = bucket
	}

	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	tokensToAdd := elapsed * rh.limiter.config.RequestsPerSecond
	bucket.tokens = min(float64(rh.limiter.config.Burst), bucket.tokens+tokensToAdd)
	bucket.lastRefill = now

	if bucket.tokens >= 1 {
		bucket.tokens--
		rh.limiter.mu.Unlock()
		return dispatcher.ContinueGroups
	}

	rh.limiter.mu.Unlock()
	return dispatcher.SkipCurrentGroup
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
