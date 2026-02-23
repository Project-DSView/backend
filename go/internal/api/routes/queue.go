package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
)

func SetupQueueRoutes(
	app *fiber.App,
	cfg *config.Config,
	jwtService *services.JWTService,
	queueService *services.QueueService,
	userService *services.UserService,
) {
	// Create handler
	queueHandler := handler.NewQueueHandler(queueService, userService)

	// Queue routes group
	queueGroup := app.Group("/api/queue")

	// Protected routes (JWT or API key authentication required)
	queueGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// Queue management routes
	queueGroup.Get("/jobs", queueHandler.GetQueueJobs)                 // GET /api/queue/jobs
	queueGroup.Get("/jobs/:id", queueHandler.GetQueueJob)              // GET /api/queue/jobs/:id
	queueGroup.Post("/jobs/:id/cancel", queueHandler.CancelQueueJob)   // POST /api/queue/jobs/:id/cancel
	queueGroup.Post("/jobs/:id/process", queueHandler.ProcessQueueJob) // POST /api/queue/jobs/:id/process
	queueGroup.Post("/jobs/:id/claim", queueHandler.ClaimQueueJob)     // POST /api/queue/jobs/:id/claim
	queueGroup.Post("/jobs/:id/complete", queueHandler.CompleteReview) // POST /api/queue/jobs/:id/complete
	queueGroup.Post("/jobs/:id/retry", queueHandler.RetryQueueJob)     // POST /api/queue/jobs/:id/retry
	queueGroup.Get("/stats", queueHandler.GetQueueStats)               // GET /api/queue/stats

	queueGroup.Post("/review", queueHandler.SubmitCodeReview) // POST /api/queue/review
}
