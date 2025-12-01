package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerSecond float64
	Burst             int
	CleanupInterval   time.Duration
}

// DefaultRateLimiterConfig returns default rate limiter configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 10,              // 10 requests per second
		Burst:             20,              // Allow burst of 20 requests
		CleanupInterval:   5 * time.Minute, // Cleanup old limiters every 5 minutes
	}
}

// IPRateLimiter manages rate limiters for each IP
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	r        float64
	b        int
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(r float64, b int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}

	// Start cleanup goroutine
	go limiter.cleanupOldLimiters(5 * time.Minute)

	return limiter
}

// GetLimiter returns the rate limiter for the given IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(i.r), i.b)
		i.limiters[ip] = limiter
	}

	return limiter
}

// cleanupOldLimiters removes limiters that haven't been used
func (i *IPRateLimiter) cleanupOldLimiters(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		for ip, limiter := range i.limiters {
			// Remove limiter if it has available tokens (not being used)
			if limiter.Tokens() == float64(i.b) {
				delete(i.limiters, ip)
			}
		}
		i.mu.Unlock()
	}
}

// RateLimiter returns a middleware that limits requests per IP
func RateLimiter(config ...RateLimiterConfig) gin.HandlerFunc {
	cfg := DefaultRateLimiterConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	limiter := NewIPRateLimiter(cfg.RequestsPerSecond, cfg.Burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := limiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": "please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StrictRateLimiter for sensitive endpoints like login
func StrictRateLimiter() gin.HandlerFunc {
	return RateLimiter(RateLimiterConfig{
		RequestsPerSecond: 3, // 3 requests per second
		Burst:             5, // Allow burst of 5
		CleanupInterval:   5 * time.Minute,
	})
}

// AuthRateLimiter specifically for authentication endpoints
func AuthRateLimiter() gin.HandlerFunc {
	limiter := NewIPRateLimiter(5, 10) // 5 req/sec, burst 10

	return func(c *gin.Context) {
		ip := c.ClientIP()
		rateLimiter := limiter.GetLimiter(ip)

		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "too many authentication attempts",
				"message": "please wait before trying again",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserBasedRateLimiter limits requests per authenticated user
type UserBasedRateLimiter struct {
	limiters map[uint]*rate.Limiter
	mu       sync.RWMutex
	r        float64
	b        int
}

// NewUserBasedRateLimiter creates a new user-based rate limiter
func NewUserBasedRateLimiter(r float64, b int) *UserBasedRateLimiter {
	limiter := &UserBasedRateLimiter{
		limiters: make(map[uint]*rate.Limiter),
		r:        r,
		b:        b,
	}

	go limiter.cleanupOldLimiters(5 * time.Minute)

	return limiter
}

// GetLimiter returns the rate limiter for the given user ID
func (u *UserBasedRateLimiter) GetLimiter(userID uint) *rate.Limiter {
	u.mu.Lock()
	defer u.mu.Unlock()

	limiter, exists := u.limiters[userID]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(u.r), u.b)
		u.limiters[userID] = limiter
	}

	return limiter
}

func (u *UserBasedRateLimiter) cleanupOldLimiters(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		u.mu.Lock()
		for userID, limiter := range u.limiters {
			if limiter.Tokens() == float64(u.b) {
				delete(u.limiters, userID)
			}
		}
		u.mu.Unlock()
	}
}

// UserRateLimiter limits requests per authenticated user (use after AuthMiddleware)
func UserRateLimiter(config ...RateLimiterConfig) gin.HandlerFunc {
	cfg := DefaultRateLimiterConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	limiter := NewUserBasedRateLimiter(cfg.RequestsPerSecond, cfg.Burst)

	return func(c *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			// Fallback to IP-based if user not authenticated
			c.Next()
			return
		}

		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
			c.Abort()
			return
		}

		rateLimiter := limiter.GetLimiter(uid)

		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "you are making requests too quickly",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
