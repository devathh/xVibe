package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu         sync.RWMutex
	counters   map[string]int
	timestamps map[string]time.Time
	limit      int
	window     time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		mu:         sync.RWMutex{},
		counters:   map[string]int{},
		timestamps: map[string]time.Time{},
		limit:      limit,
		window:     window,
	}
}

func (rl *rateLimiter) cleanUp() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, ts := range rl.timestamps {
			if now.Sub(ts) > rl.window {
				delete(rl.counters, ip)
				delete(rl.timestamps, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	last, exists := rl.timestamps[ip]

	if !exists || now.Sub(last) > rl.window {
		rl.counters[ip] = 1
		rl.timestamps[ip] = now
		return true
	}

	if rl.counters[ip] >= rl.limit {
		return false
	}

	rl.counters[ip]++
	return true
}

func NewRateLimitMiddleware(maxTimes int, windowTime time.Duration) gin.HandlerFunc {
	rl := NewRateLimiter(maxTimes, windowTime)

	return func(ctx *gin.Context) {
		if !rl.allow(ctx.ClientIP()) {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}

		ctx.Next()
	}
}
