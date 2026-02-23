package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
)

func SetupCourseMaterialRoutes(
	app *fiber.App,
	cfg *config.Config,
	materialHandler *handler.CourseMaterialHandler,
	submissionHandler *handler.SubmissionHandler,
	jwtService *services.JWTService,
) {
	// Course material routes group
	materialGroup := app.Group("/api/course-materials")

	// All course material routes now require authentication for enrollment validation
	materialGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// GET routes with enrollment validation
	materialGroup.Get("/", materialHandler.GetCourseMaterials)   // GET /api/course-materials?course_id=xxx
	materialGroup.Get("/:id", materialHandler.GetCourseMaterial) // GET /api/course-materials/:id

	// POST/PUT/DELETE routes (teachers only)
	materialGroup.Post("/", materialHandler.CreateCourseMaterial)           // POST /api/course-materials
	materialGroup.Post("/upload", materialHandler.UploadCourseMaterialFile) // POST /api/course-materials/upload
	materialGroup.Put("/:id", materialHandler.UpdateCourseMaterial)         // PUT /api/course-materials/:id
	materialGroup.Delete("/:id", materialHandler.DeleteCourseMaterial)      // DELETE /api/course-materials/:id

	// Material-based exercise routes
	materialGroup.Post("/:id/submit", submissionHandler.SubmitMaterialExercise)         // POST /api/course-materials/:id/submit
	materialGroup.Get("/:id/submissions/me", submissionHandler.GetMyMaterialSubmission) // GET /api/course-materials/:id/submissions/me
	materialGroup.Get("/:id/test-cases", materialHandler.GetTestCases)                  // GET /api/course-materials/:id/test-cases
	materialGroup.Post("/:id/test-cases", materialHandler.AddTestCase)                  // POST /api/course-materials/:id/test-cases
	materialGroup.Put("/test-cases/:test_case_id", materialHandler.UpdateTestCase)      // PUT /api/course-materials/test-cases/:test_case_id
	materialGroup.Delete("/test-cases/:test_case_id", materialHandler.DeleteTestCase)   // DELETE /api/course-materials/test-cases/:test_case_id

	// Problem image management routes
	materialGroup.Post("/:id/images", materialHandler.UploadProblemImage) // POST /api/course-materials/:id/images
}
