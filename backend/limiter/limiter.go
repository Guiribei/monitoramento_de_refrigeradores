package limiter

import(
	"sync"
	"time"
)

type RateLimiter struct {
	mu     sync.Mutex
	last   time.Time
	window time.Duration
}

func NewRateLimiter(window time.Duration) *RateLimiter {
	return &RateLimiter{window: window}
}


func (r *RateLimiter) Allow() (bool, time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if r.last.IsZero() || now.Sub(r.last) >= r.window {
		r.last = now
		return true, 0
	}
	retry := r.window - now.Sub(r.last)
	return false, retry
}