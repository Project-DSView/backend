package security

import (
	"strings"

	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/gofiber/fiber/v2"
)

// APIKeyAuth middleware validates API key from header
func APIKeyAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get API key from header
		apiKey := c.Get(cfg.APIKey.APIKeyName)
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":  "API key is required",
				"header": cfg.APIKey.APIKeyName,
			})
		}

		// Validate API key
		if apiKey != cfg.APIKey.APIKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}

		// Store API key info in context for potential use
		c.Locals("api_key", apiKey)
		c.Locals("api_key_name", cfg.APIKey.APIKeyName)

		return c.Next()
	}
}

// APIKeyAuthOptional middleware validates API key if provided, but doesn't require it
func APIKeyAuthOptional(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get API key from header
		apiKey := c.Get(cfg.APIKey.APIKeyName)

		// If API key is provided, validate it
		if apiKey != "" {
			if apiKey != cfg.APIKey.APIKey {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid API key",
				})
			}

			// Store API key info in context
			c.Locals("api_key", apiKey)
			c.Locals("api_key_name", cfg.APIKey.APIKeyName)
		}

		return c.Next()
	}
}

// APIKeyAndJWTAuth middleware requires both API key and JWT authentication
func APIKeyAndJWTAuth(cfg *config.Config, jwtService *services.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check for API key first
		apiKey := c.Get(cfg.APIKey.APIKeyName)
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":          "API key is required",
				"api_key_header": cfg.APIKey.APIKeyName,
			})
		}

		// Validate API key
		if apiKey != cfg.APIKey.APIKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}

		// Store API key info
		c.Locals("api_key", apiKey)
		c.Locals("api_key_name", cfg.APIKey.APIKeyName)

		// Check for JWT token
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":      "JWT token is required",
				"jwt_header": "Authorization: Bearer <token>",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
		}

		// Validate JWT token
		claims, err := jwtService.ValidateToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT token: " + err.Error(),
			})
		}

		// Store JWT info (compatible with existing JWTAuth middleware)
		c.Locals("claims", claims)
		c.Locals("auth_type", "api_key_and_jwt")
		c.Locals("jwt_token", parts[1])
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_name", claims.Name)

		return c.Next()
	}
}
