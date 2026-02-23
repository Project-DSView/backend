package logging

import (
	"time"

	"github.com/Project-DSView/backend/go/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// StructuredLoggingMiddleware provides structured JSON logging for Fiber
func StructuredLoggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Log request
		logger.LogRequest(c, duration, map[string]interface{}{
			"query_params": string(c.Request().URI().QueryString()),
			"headers":      c.GetReqHeaders(),
		})

		return err
	}
}

// ErrorLoggingMiddleware logs errors with context
func ErrorLoggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		if err != nil {
			logger.Error("Request failed", err, map[string]interface{}{
				"method":     c.Method(),
				"path":       c.Path(),
				"client_ip":  c.IP(),
				"user_agent": c.Get("User-Agent"),
			})
		}

		return err
	}
}

// PerformanceLoggingMiddleware logs slow requests
func PerformanceLoggingMiddleware(threshold float64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()

		if duration > threshold {
			logger.LogPerformance("slow_request", duration, map[string]interface{}{
				"method":    c.Method(),
				"path":      c.Path(),
				"threshold": threshold,
			})
		}

		return err
	}
}
