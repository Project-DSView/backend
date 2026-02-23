package security

import (
	"strings"

	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/gofiber/fiber/v2"
)

func JWTAuth(jwtService *services.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip JWT for playground routes
		if strings.HasPrefix(c.Path(), "/api/playground") {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No authorization header provided",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
		}

		token := parts[1]
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}

// Optional middleware to allow expired tokens for refresh endpoint
func JWTAuthAllowExpired(jwtService *services.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No authorization header provided",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
		}

		token := parts[1]
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No access token provided",
			})
		}

		// Validate token (allow expired for refresh)
		claims, err := jwtService.ValidateTokenAllowExpired(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Store claims in context
		c.Locals("claims", claims)
		return c.Next()
	}
}
