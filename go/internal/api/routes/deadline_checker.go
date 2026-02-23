package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
)

func SetupDeadlineCheckerRoutes(
	app *fiber.App,
	cfg *config.Config,
	deadlineHandler *handler.DeadlineCheckerHandler,
	jwtService *services.JWTService,
) {
	// Deadline checker routes group
	deadlineGroup := app.Group("/api/materials")

	// Public routes (no authentication required)
	deadlineGroup.Get("/check-deadline", deadlineHandler.CheckMaterialDeadline)            // GET /api/materials/check-deadline?material_id=xxx
	deadlineGroup.Get("/by-deadline-status", deadlineHandler.GetMaterialsByDeadlineStatus) // GET /api/materials/by-deadline-status?course_id=xxx
	deadlineGroup.Get("/upcoming-deadlines", deadlineHandler.GetUpcomingDeadlines)         // GET /api/materials/upcoming-deadlines?course_id=xxx&hours=24
	deadlineGroup.Get("/deadline-stats", deadlineHandler.GetDeadlineStats)                 // GET /api/materials/deadline-stats?course_id=xxx

	// Protected routes (JWT or API key authentication required)
	deadlineGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))
	deadlineGroup.Get("/available", deadlineHandler.GetAvailableMaterials) // GET /api/materials/available?course_id=xxx
	deadlineGroup.Get("/expired", deadlineHandler.GetExpiredMaterials)     // GET /api/materials/expired?course_id=xxx
	deadlineGroup.Get("/can-submit", deadlineHandler.CanSubmitExercise)    // GET /api/materials/can-submit?exercise_id=xxx
}
