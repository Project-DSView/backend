package security

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders middleware adds essential security headers
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Prevent clickjacking attacks
		c.Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Set("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (relaxed for Swagger UI)
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' https://www.googleapis.com http://localhost:* http://127.0.0.1:* https://*.app.github.dev https://*.vercel.app https://*.lvh.me wss://*.lvh.me;")

		// Remove server header (optional)
		c.Set("Server", "")

		return c.Next()
	}
}

// CORSHeaders middleware adds CORS headers (complement to existing CORS middleware)
func CORSHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			c.Set("Access-Control-Allow-Origin", "*")
			c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			c.Set("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,Cookie,X-Requested-With,X-CSRF-Token,dsview-api-key")
			c.Set("Access-Control-Max-Age", "86400")
			return c.SendStatus(204)
		}

		return c.Next()
	}
}
