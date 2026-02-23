package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/Project-DSView/backend/go/pkg/storage"
)

func SetupAuthRoutes(
	app *fiber.App,
	cfg *config.Config,
	oauthService *services.OAuthService,
	jwtService *services.JWTService,
	userService *services.UserService,
	storageService storage.StorageService,
) {
	// Create auth handler
	authHandler := handler.NewAuthHandler(oauthService, jwtService, userService, &cfg.Frontend, storageService)

	// Test routes group
	testGroup := app.Group("/test")
	// Test token endpoints
	testGroup.Get("/token", func(c *fiber.Ctx) error {
		return handler.TestToken(c, jwtService, cfg)
	})
	testGroup.Post("/token", func(c *fiber.Ctx) error {
		return handler.TestTokenPostWithDB(c, jwtService, userService, cfg)
	})

	// Auth routes group with rate limiting
	authGroup := app.Group("/api/auth")
	authGroup.Use(security.AuthRateLimit()) // Apply stricter rate limiting for auth endpoints

	// Public auth routes
	authGroup.Get("/google", authHandler.GoogleLogin)             // GET /api/auth/google
	authGroup.Get("/google/callback", authHandler.GoogleCallback) // GET /api/auth/google/callback
	authGroup.Post("/logout", authHandler.Logout)                 // POST /api/auth/logout

	// Refresh token endpoint (allows expired tokens)
	authGroup.Post("/refresh", security.JWTAuthAllowExpired(jwtService), authHandler.RefreshToken) // POST /api/auth/refresh

	// Profile routes group
	profileGroup := app.Group("/api/profile")

	// Protected routes (JWT authentication required)
	profileGroup.Use(security.JWTAuth(jwtService))

	// User profile management routes
	profileGroup.Get("/", authHandler.Profile)   // GET /api/profile
	profileGroup.Get("/me", authHandler.Profile) // GET /api/profile/me (Alias for profile)
}
