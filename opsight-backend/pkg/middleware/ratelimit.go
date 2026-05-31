package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"opsight-backend/pkg/logger"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	timestamps []time.Time
	lastSeen   time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	limit   int
	window  time.Duration
	stopCh  chan struct{}
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		limit:   limit,
		window:  window,
		stopCh:  make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[ip]
	if !exists {
		rl.entries[ip] = &rateLimitEntry{
			timestamps: []time.Time{now},
			lastSeen:   now,
		}
		return true
	}

	cutoff := now.Add(-rl.window)
	valid := entry.timestamps[:0]
	for _, ts := range entry.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	entry.timestamps = valid
	entry.lastSeen = now

	if len(valid) >= rl.limit {
		return false
	}

	entry.timestamps = append(entry.timestamps, now)
	return true
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-rl.stopCh:
			return
		case <-ticker.C:
			rl.cleanup()
		}
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.window * 2)
	for ip, entry := range rl.entries {
		if entry.lastSeen.Before(cutoff) {
			delete(rl.entries, ip)
		}
	}
}

func loadEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			logger.Warn().
				Str("client_ip", ip).
				Str("path", c.Request.URL.Path).
				Msg("rate limit exceeded")
			response.Error(c, http.StatusTooManyRequests, 429, "too many requests, please slow down")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GeneralRateLimit() gin.HandlerFunc {
	rps := loadEnvInt("RATE_LIMIT_RPS", 100)
	return RateLimit(rps, 1*time.Minute)
}

func LoginRateLimit() gin.HandlerFunc {
	rpm := loadEnvInt("RATE_LIMIT_LOGIN_RPM", 5)
	return RateLimit(rpm, 1*time.Minute)
}
