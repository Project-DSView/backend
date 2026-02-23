package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
)

func SetupTestCaseRoutes(
	app *fiber.App,
	cfg *config.Config,
	jwtService *services.JWTService,
	testCaseService *services.TestCaseService,
	userService *services.UserService,
) {
	// Create test case handler
	testCaseHandler := handler.NewTestCaseHandler(testCaseService, userService)

	// Test case routes group
	testCaseGroup := app.Group("/api/test-cases")

	// Protected routes (JWT or API key authentication required)
	testCaseGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// Test case management routes
	testCaseGroup.Get("/exercises/:exercise_id", testCaseHandler.GetTestCases)    // GET /api/test-cases/exercises/:exercise_id
	testCaseGroup.Post("/exercises/:exercise_id", testCaseHandler.CreateTestCase) // POST /api/test-cases/exercises/:exercise_id
	testCaseGroup.Put("/:id", testCaseHandler.UpdateTestCase)                     // PUT /api/test-cases/:id
	testCaseGroup.Delete("/:id", testCaseHandler.DeleteTestCase)                  // DELETE /api/test-cases/:id
}
