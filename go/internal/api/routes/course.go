package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/Project-DSView/backend/go/pkg/storage"
)

func SetupCourseRoutes(
	app *fiber.App,
	cfg *config.Config,
	jwtService *services.JWTService,
	courseService *services.CourseService,
	courseMaterialService *services.CourseMaterialService,
	userService *services.UserService,
	enrollmentService *services.EnrollmentService,
	invitationService *services.InvitationService,
	queueService *services.QueueService,
	storageService storage.StorageService,
) {
	// Create handlers
	courseHandler := handler.NewCourseHandler(courseService, courseMaterialService, userService, enrollmentService, queueService, storageService)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentService, userService, courseService)
	invitationHandler := handler.NewInvitationHandler(invitationService, userService)

	// Course routes group
	courseGroup := app.Group("/api/courses")

	// Protected routes (Both API key and JWT authentication required)
	courseGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// Course management routes
	courseGroup.Get("/", courseHandler.GetCourses)         // GET /api/courses
	courseGroup.Post("/", courseHandler.CreateCourse)      // POST /api/courses
	courseGroup.Get("/:id", courseHandler.GetCourse)       // GET /api/courses/:id
	courseGroup.Put("/:id", courseHandler.UpdateCourse)    // PUT /api/courses/:id
	courseGroup.Delete("/:id", courseHandler.DeleteCourse) // DELETE /api/courses/:id

	// Course exercise routes
	courseGroup.Get("/:id/exercises", courseHandler.GetCourseExercises) // GET /api/courses/:id/exercises

	// Course report routes
	courseGroup.Get("/:id/report/teacher", courseHandler.GetCourseReportForTeacher) // GET /api/courses/:id/report/teacher
	courseGroup.Get("/:id/report/ta", courseHandler.GetCourseReportForTA)           // GET /api/courses/:id/report/ta

	// Course image management routes
	courseGroup.Delete("/:id/image", courseHandler.DeleteCourseImage) // DELETE /api/courses/:id/image

	// Course enrollment routes (matching API documentation)
	courseGroup.Post("/:id/enroll", enrollmentHandler.EnrollInCourse)                              // POST /api/courses/:id/enroll
	courseGroup.Get("/:id/enrollments", enrollmentHandler.GetCourseEnrollments)                    // GET /api/courses/:id/enrollments
	courseGroup.Get("/:id/my-enrollment", enrollmentHandler.GetMyEnrollment)                       // GET /api/courses/:id/my-enrollment
	courseGroup.Delete("/:id/enroll", enrollmentHandler.UnenrollFromCourse)                        // DELETE /api/courses/:id/enroll
	courseGroup.Put("/:courseId/enrollments/:userId/role", enrollmentHandler.UpdateEnrollmentRole) // PUT /api/courses/:courseId/enrollments/:userId/role

	// Course invitation routes
	courseGroup.Post("/:id/invitations", invitationHandler.CreateInvitation)    // POST /api/courses/:id/invitations
	courseGroup.Get("/:id/invitations", invitationHandler.GetCourseInvitations) // GET /api/courses/:id/invitations

	// Enrollment routes group
	enrollmentGroup := app.Group("/api/enrollments")

	// Protected routes (Both API key and JWT authentication required)
	enrollmentGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))

	// Enrollment management routes
	enrollmentGroup.Post("/courses/:id", enrollmentHandler.EnrollInCourse)       // POST /api/enrollments/courses/:id
	enrollmentGroup.Get("/courses/:id", enrollmentHandler.GetCourseEnrollments)  // GET /api/enrollments/courses/:id
	enrollmentGroup.Delete("/courses/:id", enrollmentHandler.UnenrollFromCourse) // DELETE /api/enrollments/courses/:id

	// Invitation routes (separate group for public invitation endpoint)
	invitationGroup := app.Group("/api/courses/invite")
	invitationGroup.Use(security.APIKeyAndJWTAuth(cfg, jwtService))
	invitationGroup.Post("/:token", invitationHandler.EnrollViaInvitation) // POST /api/courses/invite/:token
}
