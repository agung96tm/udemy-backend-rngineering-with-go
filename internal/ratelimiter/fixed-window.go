package ratelimiter

import (
	"sync"
	"time"
)

type FixedWindowRateLimiter struct {
	mu      sync.Mutex
	clients map[string]int
	limit   int
	window  time.Duration
}

func NewFixedWindowRateLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		clients: make(map[string]int),
		limit:   limit,
		window:  window,
	}
}

func (rl *FixedWindowRateLimiter) Allow(ip string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	count, exists := rl.clients[ip]
	if !exists || count < rl.limit {
		if !exists {
			go rl.resetCount(ip)
		}
		rl.clients[ip]++
		return true, 0
	}

	return false, rl.window
}

func (rl *FixedWindowRateLimiter) resetCount(ip string) {
	time.Sleep(rl.window)
	rl.mu.Lock()
	delete(rl.clients, ip)
	rl.mu.Unlock()
}
