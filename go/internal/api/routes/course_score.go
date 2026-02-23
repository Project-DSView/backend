package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
)

func SetupCourseScoreRoutes(
	app *fiber.App,
	cfg *config.Config,
	scoreHandler *handler.CourseScoreHandler,
	jwtService *services.JWTService,
) {
	// Course score routes group
	scoreGroup := app.Group("/api/course-scores")

	// Protected routes (JWT or API key authentication required)
	scoreGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))
	scoreGroup.Get("/course", scoreHandler.GetStudentCourseScore) // GET /api/course-scores/course?course_id=xxx
}
