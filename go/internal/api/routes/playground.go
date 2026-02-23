package routes

import (
	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/gofiber/fiber/v2"
)

// SetupPlaygroundRoutes sets up playground routes for code execution
func SetupPlaygroundRoutes(app *fiber.App, cfg *config.Config, jwtService *services.JWTService) {
	// Create playground handler
	playgroundHandler := handler.NewPlaygroundHandler(cfg)

	// Playground routes group
	playgroundGroup := app.Group("/api/playground")

	// Public playground routes (no authentication required)
	playgroundGroup.Post("/run", playgroundHandler.RunCodeGateway)
	playgroundGroup.Get("/health", playgroundHandler.HealthCheck)

	// Complexity routes group
	complexityGroup := app.Group("/api/complexity")

	// Public complexity routes (no authentication required)
	complexityGroup.Post("/performance", playgroundHandler.ComplexityPerformance)

	// Apply JWT authentication middleware for sensitive routes
	complexityGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// Authenticated complexity routes
	complexityGroup.Post("/llm", playgroundHandler.ComplexityLLM)
}
