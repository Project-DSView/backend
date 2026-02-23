package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"gorm.io/gorm"
)

func SetupSubmissionRoutes(
	app *fiber.App,
	cfg *config.Config,
	jwtService *services.JWTService,
	submissionService *services.SubmissionService,
	progressService *services.ProgressService,
	courseService *services.CourseService,
	enrollmentService *services.EnrollmentService,
	userService *services.UserService,
	draftService *services.DraftService,
	materialService *services.CourseMaterialService,
	db *gorm.DB,
) {
	// Create handlers
	submissionHandler := handler.NewSubmissionHandler(submissionService, userService, enrollmentService, db)
	progressHandler := handler.NewProgressHandler(progressService, userService, enrollmentService, courseService)

	// Submission routes group
	submissionGroup := app.Group("/api/submissions")

	// Protected routes (JWT or API key authentication required)
	submissionGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// Submission management routes
	submissionGroup.Post("/exercises/:id", submissionHandler.SubmitExercise)         // POST /api/submissions/exercises/:id
	submissionGroup.Get("/exercises/:id", submissionHandler.ListExerciseSubmissions) // GET /api/submissions/exercises/:id
	submissionGroup.Get("/:id", submissionHandler.GetSubmission)                     // GET /api/submissions/:id

	// Course materials submission routes
	courseMaterialGroup := app.Group("/api/course-materials")
	courseMaterialGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))
	courseMaterialGroup.Post("/:id/submit", submissionHandler.SubmitMaterialExercise)         // POST /api/course-materials/:id/submit
	courseMaterialGroup.Post("/:id/submit-pdf", submissionHandler.SubmitPDFExercise)          // POST /api/course-materials/:id/submit-pdf
	courseMaterialGroup.Get("/:id/submissions/me", submissionHandler.GetMyMaterialSubmission) // GET /api/course-materials/:id/submissions/me

	// Progress routes group
	progressGroup := app.Group("/api/progress")

	// Protected routes (JWT or API key authentication required)
	progressGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// Progress management routes
	progressGroup.Post("/:id/verify", progressHandler.VerifyProgress)                     // POST /api/progress/:id/verify
	progressGroup.Get("/:id/logs", progressHandler.GetVerificationLogs)                   // GET /api/progress/:id/logs
	progressGroup.Post("/:material_id/request-approval", progressHandler.RequestApproval) // POST /api/progress/:material_id/request-approval

	// Students progress route (separate group to match swagger docs)
	studentsGroup := app.Group("/api/students")
	studentsGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))
	studentsGroup.Get("/progress", progressHandler.GetSelfProgress) // GET /api/students/progress

	// Course progress route
	coursesGroup := app.Group("/api/courses")
	coursesGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))
	coursesGroup.Get("/:id/progress", progressHandler.GetCourseProgress) // GET /api/courses/:id/progress
}
