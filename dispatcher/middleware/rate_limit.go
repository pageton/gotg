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
	// CleanupInterval controls how often stale buckets are evicted.
	// Defaults to 5 minutes.
	CleanupInterval time.Duration
}

// RateLimiter manages rate limiting for updates
type RateLimiter struct {
	config  *RateLimitConfig
	tokens  map[string]*tokenBucket
	mu      sync.RWMutex
	stopCh  chan struct{}
}

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter with periodic stale-bucket cleanup.
// Call Stop() when the limiter is no longer needed.
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
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	rl := &RateLimiter{
		config: config,
		tokens: make(map[string]*tokenBucket),
		stopCh: make(chan struct{}),
	}

	// Stale threshold: at least 1 minute, or 2x the refill period.
	staleThreshold := 2 * time.Duration(float64(time.Second)/config.RequestsPerSecond)
	if staleThreshold < time.Minute {
		staleThreshold = time.Minute
	}

	go rl.cleanupLoop(config.CleanupInterval, staleThreshold)
	return rl
}

// Stop terminates the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

func (rl *RateLimiter) cleanupLoop(interval, staleThreshold time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-rl.stopCh:
			return
		case <-ticker.C:
			rl.evictStale(staleThreshold)
		}
	}
}

func (rl *RateLimiter) evictStale(threshold time.Duration) {
	now := time.Now()
	rl.mu.Lock()
	for key, bucket := range rl.tokens {
		if now.Sub(bucket.lastRefill) > threshold {
			delete(rl.tokens, key)
		}
	}
	rl.mu.Unlock()
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
	defer rh.limiter.mu.Unlock()

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
		return dispatcher.ContinueGroups
	}

	return dispatcher.SkipCurrentGroup
}
