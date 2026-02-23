package security

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter represents a simple in-memory rate limiter
type RateLimiter struct {
	visitors map[string]*Visitor
	mutex    sync.RWMutex
	rate     int           // requests per minute
	window   time.Duration // time window
}

// Visitor represents a visitor with their request count and last seen time
type Visitor struct {
	Count    int
	LastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes old visitors periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		for ip, visitor := range rl.visitors {
			if now.Sub(visitor.LastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mutex.Unlock()
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	visitor, exists := rl.visitors[ip]

	if !exists {
		rl.visitors[ip] = &Visitor{
			Count:    1,
			LastSeen: now,
		}
		return true
	}

	// Reset if window has passed
	if now.Sub(visitor.LastSeen) > rl.window {
		visitor.Count = 1
		visitor.LastSeen = now
		return true
	}

	// Check if rate limit exceeded
	if visitor.Count >= rl.rate {
		return false
	}

	visitor.Count++
	visitor.LastSeen = now
	return true
}

// RateLimit middleware for general API endpoints
func RateLimit(rate int, window time.Duration) fiber.Handler {
	limiter := NewRateLimiter(rate, window)

	return func(c *fiber.Ctx) error {
		ip := c.IP()

		if !limiter.Allow(ip) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": int(window.Seconds()),
			})
		}

		return c.Next()
	}
}

// AuthRateLimit middleware for authentication endpoints (stricter)
func AuthRateLimit() fiber.Handler {
	limiter := NewRateLimiter(20, 30*time.Second) // 20 requests per 30 seconds for auth

	return func(c *fiber.Ctx) error {
		ip := c.IP()

		// Special handling for Google OAuth - allow more requests for OAuth flow
		path := c.Path()
		if path == "/api/auth/google" || path == "/api/auth/google/callback" {
			// Use more lenient rate limiting for OAuth endpoints
			oauthLimiter := NewRateLimiter(50, 30*time.Second) // 50 requests per 30 seconds for OAuth
			if !oauthLimiter.Allow(ip) {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error":       "OAuth rate limit exceeded",
					"message":     "Too many OAuth requests. Please try again later.",
					"retry_after": 30,
				})
			}
			return c.Next()
		}

		if !limiter.Allow(ip) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Authentication rate limit exceeded",
				"message":     "Too many authentication attempts. Please try again later.",
				"retry_after": 30,
			})
		}

		return c.Next()
	}
}

// CodeExecutionRateLimit middleware for code execution endpoints
func CodeExecutionRateLimit() fiber.Handler {
	limiter := NewRateLimiter(10, 1*time.Minute) // 10 requests per minute for code execution

	return func(c *fiber.Ctx) error {
		ip := c.IP()

		if !limiter.Allow(ip) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Code execution rate limit exceeded",
				"message":     "Too many code execution requests. Please try again later.",
				"retry_after": 60,
			})
		}

		return c.Next()
	}
}
