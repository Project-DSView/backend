package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
)

func SetupAnnouncementRoutes(
	app *fiber.App,
	cfg *config.Config,
	announcementHandler *handler.AnnouncementHandler,
	jwtService *services.JWTService,
) {
	// Announcement routes group
	announcementGroup := app.Group("/api/announcements")

	// All announcement routes now require authentication for enrollment validation
	announcementGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// GET routes with enrollment validation
	announcementGroup.Get("/", announcementHandler.GetAnnouncements)             // GET /api/announcements?course_id=xxx
	announcementGroup.Get("/stats", announcementHandler.GetAnnouncementStats)    // GET /api/announcements/stats?course_id=xxx
	announcementGroup.Get("/:id", announcementHandler.GetAnnouncement)           // GET /api/announcements/:id
	announcementGroup.Get("/recent", announcementHandler.GetRecentAnnouncements) // GET /api/announcements/recent

	// POST/PUT/DELETE routes (teachers only)
	announcementGroup.Post("/", announcementHandler.CreateAnnouncement)      // POST /api/announcements
	announcementGroup.Put("/:id", announcementHandler.UpdateAnnouncement)    // PUT /api/announcements/:id
	announcementGroup.Delete("/:id", announcementHandler.DeleteAnnouncement) // DELETE /api/announcements/:id
}
