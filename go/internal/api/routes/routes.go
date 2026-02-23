// internal/api/routes/routes.go
package routes

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"gorm.io/gorm"

	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/Project-DSView/backend/go/internal/api/middleware/logging"
	"github.com/Project-DSView/backend/go/internal/api/middleware/security"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/Project-DSView/backend/go/pkg/enrollment"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/Project-DSView/backend/go/pkg/storage"

	_ "github.com/Project-DSView/backend/go/docs"
)

func SetupRoutes(
	cfg *config.Config,
	oauthService *services.OAuthService,
	jwtService *services.JWTService,
	userService *services.UserService,
	courseService *services.CourseService,
	testCaseService *services.TestCaseService,
	enrollmentService *services.EnrollmentService,
	invitationService *services.InvitationService,
	submissionService *services.SubmissionService,
	progressService *services.ProgressService,
	draftService *services.DraftService,
	queueService *services.QueueService,
	storageService storage.StorageService,
	deadlineChecker *services.DeadlineCheckerService,
	courseScoreService *services.CourseScoreService,
	courseMaterialService *services.CourseMaterialService,
	db *gorm.DB,
) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return response.SendError(c, code, err.Error())
		},
	})

	// Note: Structured logger is initialized in main.go before SetupRoutes is called

	// Global middleware
	app.Use(logging.StructuredLoggingMiddleware())
	app.Use(logging.ErrorLoggingMiddleware())
	app.Use(logging.PerformanceLoggingMiddleware(1.0)) // Log requests slower than 1 second
	app.Use(recover.New())

	// Security headers (must be before CORS)
	app.Use(security.SecurityHeaders())

	// Configure CORS
	// Add Swagger UI origin to allowed origins for CORS
	allowedOrigins := append(cfg.Frontend.AllowedOrigins,
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"https://localhost:8080",
		"https://127.0.0.1:8080",
		"https://127.0.0.1:8080",
		"https://localhost:3000",
		"https://127.0.0.1:3000",
		"http://go.lvh.me",
		"http://fastapi.lvh.me",
		"https://go.lvh.me",
		"https://fastapi.lvh.me",
		"http://dsview.lvh.me",
		"https://dsview.lvh.me")

	// Clean up any invalid origins (remove empty strings and trim whitespace)
	var cleanOrigins []string
	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin != "" && origin != "http://localhost/fastapi⁠" {
			cleanOrigins = append(cleanOrigins, origin)
		}
	}

	app.Use(cors.New(cors.Config{
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,Cookie,X-Requested-With,X-CSRF-Token,Cache-Control,Pragma," + cfg.APIKey.APIKeyName,
		AllowCredentials: true,
		MaxAge:           86400,
		// Enable CORS debugging in development
		AllowOriginsFunc: func(origin string) bool {
			if cfg.Server.Environment == "development" {

			}
			// Check if origin is in allowed origins
			for _, allowedOrigin := range cleanOrigins {
				if origin == allowedOrigin {
					return true
				}
			}
			// Allow GitHub Codespaces origins (*.app.github.dev)
			if strings.HasSuffix(origin, ".app.github.dev") && strings.HasPrefix(origin, "https://") {
				return true
			}
			// Allow Vercel deployments (*.vercel.app)
			if strings.HasSuffix(origin, ".vercel.app") && strings.HasPrefix(origin, "https://") {
				return true
			}
			// Allow localhost and 127.0.0.1 with any port for development
			if cfg.Server.Environment == "development" {
				if strings.HasPrefix(origin, "http://localhost:") ||
					strings.HasPrefix(origin, "https://localhost:") ||
					strings.HasPrefix(origin, "http://127.0.0.1:") ||
					strings.HasPrefix(origin, "https://127.0.0.1:") {
					return true
				}
			}
			return false
		},
	}))

	// Rate limiting for general API endpoints
	app.Use(security.RateLimit(100, 1*time.Minute)) // 100 requests per minute

	// Swagger documentation route
	if cfg.Server.Environment != "production" {
		app.Get("/docs/*", swagger.HandlerDefault)
	}

	// Initialize System Handler
	systemHandler := handler.NewSystemHandler(db, cfg)

	// Health check (public)
	app.Get("/health", systemHandler.HealthCheck)

	// API key protected health check
	app.Get("/health/secure", security.APIKeyAuth(cfg), systemHandler.HealthCheck)

	// APIInfo godoc
	// @Summary API information
	// @Description Get API information and available endpoints
	// @Tags system
	// @Produce json
	// @Success 200 {object} object{success=bool,message=string,data=object{service=string,version=string,documentation=object,endpoints=object,roles=object}} "API information"
	// @Router / [get]
	app.Get("/", security.APIKeyAuth(cfg), func(c *fiber.Ctx) error {
		docInfo := fiber.Map{}

		if cfg.Server.Environment != "production" {
			docInfo["swagger_ui"] = "/docs/"
			docInfo["openapi_json"] = "/docs/doc.json"
		}

		return response.SendSuccess(c, "DSView Backend API", fiber.Map{
			"service":       "DSView Backend API - Authentication, Exercise Management & Code Execution",
			"version":       "1.0.0",
			"documentation": docInfo,
			"endpoints": fiber.Map{
				"auth": fiber.Map{
					"login":    "GET /api/auth/google",
					"callback": "GET /api/auth/google/callback",
					"logout":   "POST /api/auth/logout",
					"refresh":  "POST /api/auth/refresh",
				},
				"user": fiber.Map{
					"profile": "GET /api/profile",
					"update":  "PUT /api/profile",
				},
				"course_materials": fiber.Map{
					"list_materials":  "GET /api/course-materials?course_id=xxx",
					"create_material": "POST /api/course-materials",
					"upload_file":     "POST /api/course-materials/upload",
					"get_material":    "GET /api/course-materials/:id",
					"update_material": "PUT /api/course-materials/:id",
					"delete_material": "DELETE /api/course-materials/:id",
					// removed: materials_by_type
					// removed: search_materials
					// removed: material_stats
					"submit_exercise":  "POST /api/course-materials/:id/submit",
					"test_cases":       "GET /api/course-materials/:id/test-cases",
					"add_test_case":    "POST /api/course-materials/:id/test-cases",
					"update_test_case": "PUT /api/course-materials/test-cases/:test_case_id",
					"delete_test_case": "DELETE /api/course-materials/test-cases/:test_case_id",
				},
				"workflow": fiber.Map{
					"create_code_exercise": []string{
						"1. POST /api/course-materials with type: 'code_exercise'",
						"2. POST /api/course-materials/{id}/test-cases to add test cases",
						"3. Students can submit via POST /api/course-materials/{id}/submit",
					},
					"create_pdf_exercise": []string{
						"1. POST /api/course-materials with type: 'pdf_exercise'",
						"2. Students upload PDF files via POST /api/course-materials/{id}/submit",
						"3. Teachers manually grade via progress verification",
					},
				},
				"courses": fiber.Map{
					"list_courses":     "GET /api/courses",
					"create_course":    "POST /api/courses",
					"get_course":       "GET /api/courses/:id",
					"update_course":    "PUT /api/courses/:id",
					"delete_course":    "DELETE /api/courses/:id",
					"course_materials": "GET /api/course-materials?course_id=xxx",
					"enroll":           "POST /api/courses/:id/enroll",
					"list_enrollments": "GET /api/courses/:id/enrollments",
					"unenroll":         "DELETE /api/courses/:id/enroll",
					"teacher_report":   "GET /api/courses/:id/report/teacher",
					"ta_report":        "GET /api/courses/:id/report/ta",
				},
				"execution": fiber.Map{
					"run_code":            "POST /api/exec/run",
					"get_executions":      "GET /api/exec/:source",
					"get_execution_by_id": "GET /api/exec/:source/:id",
					"health_check":        "GET /api/exec/health",
				},
				"test_cases": fiber.Map{
					"list_test_cases":  "GET /api/course-materials/:id/test-cases",
					"create_test_case": "POST /api/course-materials/:id/test-cases",
					"update_test_case": "PUT /api/course-materials/test-cases/:test_case_id",
					"delete_test_case": "DELETE /api/course-materials/test-cases/:test_case_id",
				},
				"submissions": fiber.Map{
					"submit":           "POST /api/course-materials/:id/submit",
					"list_submissions": "GET /api/course-materials/:id/submissions",
					"get_submission":   "GET /api/submissions/:id",
				},
				"progress": fiber.Map{
					"self_progress":     "GET /api/students/progress",
					"course_progress":   "GET /api/courses/:id/progress",
					"verify_progress":   "POST /api/progress/:id/verify",
					"verification_logs": "GET /api/progress/:id/logs",
				},
				"announcements": fiber.Map{
					"list_announcements":   "GET /api/announcements?course_id=xxx",
					"create_announcement":  "POST /api/announcements",
					"get_announcement":     "GET /api/announcements/:id",
					"update_announcement":  "PUT /api/announcements/:id",
					"delete_announcement":  "DELETE /api/announcements/:id",
					"pin_announcement":     "PUT /api/announcements/:id/pin",
					"announcement_stats":   "GET /api/announcements/stats?course_id=xxx",
					"recent_announcements": "GET /api/announcements/recent",
				},
				"course_scores": fiber.Map{
					"get_course_score": "GET /api/course-scores/course?course_id=xxx",
				},
				"deadline_checker": fiber.Map{
					"get_available":      "GET /api/course-materials/available?course_id=xxx",
					"get_expired":        "GET /api/course-materials/expired?course_id=xxx",
					"check_deadline":     "GET /api/course-materials/check-deadline?material_id=xxx",
					"can_submit":         "GET /api/course-materials/can-submit?material_id=xxx",
					"by_deadline_status": "GET /api/course-materials/by-deadline-status?course_id=xxx",
					"upcoming_deadlines": "GET /api/course-materials/upcoming-deadlines?course_id=xxx&hours=24",
					"deadline_stats":     "GET /api/course-materials/deadline-stats?course_id=xxx",
				},
				"queue": fiber.Map{
					"get_jobs":      "GET /api/queue/jobs",
					"get_job":       "GET /api/queue/jobs/:id",
					"cancel_job":    "POST /api/queue/jobs/:id/cancel",
					"process_job":   "POST /api/queue/jobs/:id/process",
					"claim_job":     "POST /api/queue/jobs/:id/claim",
					"complete_job":  "POST /api/queue/jobs/:id/complete",
					"retry_job":     "POST /api/queue/jobs/:id/retry",
					"queue_stats":   "GET /api/queue/stats",
					"submit_review": "POST /api/queue/review",
				},
			},
			"roles": fiber.Map{
				"student": "Default role, can view published exercises and execute code",
				"ta":      "Can view all users and published exercises",
				"teacher": "Can view all users, create/edit/delete own exercises",
				"admin":   "Full access to users, exercises, and all system functions",
			},
			"authentication": fiber.Map{
				"api_key": fiber.Map{
					"header_name": cfg.APIKey.APIKeyName,
					"description": "API key authentication for secure endpoints",
					"example":     "curl -H \"" + cfg.APIKey.APIKeyName + ": your-api-key\" " + cfg.Server.Host + ":" + cfg.Server.Port + "/",
				},
				"jwt": fiber.Map{
					"header_name": "Authorization",
					"description": "JWT token authentication",
					"example":     "curl -H \"Authorization: Bearer your-jwt-token\" " + cfg.Server.Host + ":" + cfg.Server.Port + "/api/profile",
				},
			},
		})
	})

	// Setup separated routes
	SetupAuthRoutes(app, cfg, oauthService, jwtService, userService, storageService)
	SetupTestCaseRoutes(app, cfg, jwtService, testCaseService, userService)
	SetupCourseRoutes(app, cfg, jwtService, courseService, courseMaterialService, userService, enrollmentService, invitationService, queueService, storageService)
	SetupSubmissionRoutes(app, cfg, jwtService, submissionService, progressService, courseService, enrollmentService, userService, draftService, courseMaterialService, db)
	SetupDraftRoutes(app, cfg, jwtService, draftService, userService, storageService)
	SetupQueueRoutes(app, cfg, jwtService, queueService, userService)
	SetupPlaygroundRoutes(app, cfg, jwtService)

	// Setup new announcement and course material routes
	announcementService := services.NewAnnouncementService(db) // Pass proper DB instance
	enrollmentValidator := enrollment.NewEnrollmentValidator(enrollmentService, courseService)
	announcementHandler := handler.NewAnnouncementHandler(announcementService, enrollmentValidator)
	courseMaterialHandler := handler.NewCourseMaterialHandler(courseMaterialService, enrollmentValidator, storageService)
	submissionHandler := handler.NewSubmissionHandler(submissionService, userService, enrollmentService, db)
	SetupAnnouncementRoutes(app, cfg, announcementHandler, jwtService)
	SetupCourseMaterialRoutes(app, cfg, courseMaterialHandler, submissionHandler, jwtService)

	// Setup course score routes
	courseScoreHandler := handler.NewCourseScoreHandler(courseScoreService, enrollmentValidator)
	SetupCourseScoreRoutes(app, cfg, courseScoreHandler, jwtService)

	// Setup deadline checker routes
	deadlineCheckerService := services.NewDeadlineCheckerService(db) // Pass proper DB instance
	deadlineCheckerHandler := handler.NewDeadlineCheckerHandler(deadlineCheckerService)
	SetupDeadlineCheckerRoutes(app, cfg, deadlineCheckerHandler, jwtService)

	// Setup PDF exercise routes
	pdfExerciseSubmissionService := services.NewPDFExerciseSubmissionService(db, deadlineChecker, storageService)
	pdfExerciseHandler := handler.NewPDFExerciseHandler(pdfExerciseSubmissionService)
	SetupPDFExerciseRoutes(app, cfg, pdfExerciseHandler, jwtService)

	app.Get("/test-public", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	return app
}
