// internal/routes/draft_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/Project-DSView/backend/go/pkg/storage"
)

func SetupDraftRoutes(
	app *fiber.App,
	cfg *config.Config,
	jwtService *services.JWTService,
	draftService *services.DraftService,
	userService *services.UserService,
	storageService storage.StorageService,
) {
	// Create handler
	draftHandler := handler.NewDraftHandler(draftService, userService, storageService)

	// Draft routes group
	draftGroup := app.Group("/api/drafts")

	// Protected routes (JWT or API key authentication required)
	draftGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// File upload และ draft management routes
	draftGroup.Post("/exercises/:exercise_id/upload", draftHandler.UploadPythonFile) // POST /api/drafts/exercises/:exercise_id/upload
	draftGroup.Post("/exercises/:exercise_id", draftHandler.SaveDraft)               // POST /api/drafts/exercises/:exercise_id
	draftGroup.Get("/exercises/:exercise_id", draftHandler.GetDraft)                 // GET /api/drafts/exercises/:exercise_id
	draftGroup.Delete("/exercises/:exercise_id", draftHandler.DeleteDraft)           // DELETE /api/drafts/exercises/:exercise_id
	draftGroup.Get("/my", draftHandler.GetMyDrafts)                                  // GET /api/drafts/my
}
