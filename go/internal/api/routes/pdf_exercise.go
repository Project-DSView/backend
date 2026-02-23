package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
)

func SetupPDFExerciseRoutes(
	app *fiber.App,
	cfg *config.Config,
	pdfExerciseHandler *handler.PDFExerciseHandler,
	jwtService *services.JWTService,
) {
	// PDF Exercise routes
	pdfGroup := app.Group("/api/materials")

	// Protected routes (JWT or API key authentication required)
	pdfGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// Submit PDF exercise (students)
	pdfGroup.Post("/:material_id/submit", pdfExerciseHandler.SubmitPDFExercise)

	// Get my PDF submission for a material (students)
	pdfGroup.Get("/:material_id/submissions/me", pdfExerciseHandler.GetMyPDFSubmission)

	// Get PDF submissions for a material (teachers/TAs)
	pdfGroup.Get("/:material_id/submissions", pdfExerciseHandler.GetPDFSubmissions)

	// Course routes for PDF submissions
	courseGroup := app.Group("/api/courses")

	// Protected routes (JWT or API key authentication required)
	courseGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// Get all PDF submissions for a course (teachers/TAs)
	courseGroup.Get("/:course_id/pdf-submissions", pdfExerciseHandler.GetCoursePDFSubmissions)

	// Submission management routes
	submissionGroup := app.Group("/api/submissions")

	// Protected routes (JWT or API key authentication required)
	submissionGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// Get specific submission
	submissionGroup.Get("/:submission_id", pdfExerciseHandler.GetPDFSubmission)

	// Approve submission (teachers/TAs)
	submissionGroup.Post("/:submission_id/approve", pdfExerciseHandler.ApprovePDFSubmission)

	// Reject submission (teachers/TAs)
	submissionGroup.Post("/:submission_id/reject", pdfExerciseHandler.RejectPDFSubmission)

	// Download submission (teachers/TAs)
	submissionGroup.Get("/:submission_id/download", pdfExerciseHandler.DownloadPDFSubmission)

	// Download feedback file (students only, for their own submissions)
	submissionGroup.Get("/:submission_id/feedback/download", pdfExerciseHandler.DownloadFeedbackFile)

	// Cancel submission (students only)
	submissionGroup.Delete("/:submission_id/cancel", pdfExerciseHandler.CancelPDFSubmission)
}
