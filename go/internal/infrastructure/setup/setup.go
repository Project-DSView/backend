package setup

import (
	"context"

	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/Project-DSView/backend/go/internal/infrastructure/database"
	repositories "github.com/Project-DSView/backend/go/internal/infrastructure/repositories"
	"github.com/Project-DSView/backend/go/pkg/external"
	"github.com/Project-DSView/backend/go/pkg/logger"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"gorm.io/gorm"
)

// Services holds all application services
type Services struct {
	DB                     *gorm.DB
	OAuthService           *services.OAuthService
	JWTService             *services.JWTService
	UserService            *services.UserService
	CourseService          *services.CourseService
	TestCaseService        *services.TestCaseService
	EnrollmentService      *services.EnrollmentService
	InvitationService      *services.InvitationService
	SubmissionService      *services.SubmissionService
	ProgressService        *services.ProgressService
	DraftService           *services.DraftService
	QueueService           *services.QueueService
	StorageService         storage.StorageService
	DeadlineCheckerService *services.DeadlineCheckerService
	CourseScoreService     *services.CourseScoreService
	CourseMaterialService  *services.CourseMaterialService
}

// SetupDatabase initializes database connection
func SetupDatabase(cfg *config.Config) (*gorm.DB, error) {
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		return nil, err
	}

	// Verify database health
	if err := db.HealthCheck(); err != nil {
		logger.Warnf("Database health check failed: %v", err)
	} else {
		logger.Info("Database health check passed")
	}

	return db.GetDB(), nil
}

// SetupExternalServices initializes external services
func SetupExternalServices(cfg *config.Config) (*external.DockerExecutor, storage.StorageService, error) {
	// Initialize executor
	exec := external.NewDockerExecutor(external.DockerConfig{
		Image:   cfg.Executor.Image,
		Timeout: cfg.Executor.Timeout,
		Memory:  cfg.Executor.Memory,
		CPUs:    cfg.Executor.CPUs,
	})

	// Initialize MinIO storage service
	minioConfig := &storage.MinIOConfig{
		Endpoint:         cfg.MinIO.Endpoint,
		PublicEndpoint:   cfg.MinIO.PublicEndpoint,
		AccessKeyID:      cfg.MinIO.AccessKeyID,
		SecretAccessKey:  cfg.MinIO.SecretAccessKey,
		BucketName:       cfg.MinIO.BucketName,
		MaxFileSizeBytes: cfg.MinIO.GetMaxFileSizeBytes(),
		UseSSL:           cfg.MinIO.UseSSL,
		PublicBucket:     cfg.MinIO.PublicBucket,
	}

	storageService, err := storage.NewMinIOService(minioConfig)
	if err != nil {
		return nil, nil, err
	}

	return exec, storageService, nil
}

// SetupCoreServices initializes core application services
func SetupCoreServices(db *gorm.DB, cfg *config.Config) (*Services, error) {
	// Initialize repositories
	courseScoreRepo := repositories.NewGormCourseScoreRepository(db)
	studentProgressRepo := repositories.NewGormStudentProgressRepository(db)

	// Initialize OAuth and JWT services
	oauthService := services.NewOAuthService(
		cfg.Google.ClientID,
		cfg.Google.ClientSecret,
		cfg.Google.RedirectURL,
		cfg.Google.Scopes,
	)
	jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpiresIn)

	// Initialize core services
	userService := services.NewUserService(db)
	enrollmentService := services.NewEnrollmentService(db, userService)
	courseService := services.NewCourseService(db, userService, enrollmentService)
	invitationService := services.NewInvitationService(db, courseService, enrollmentService)
	testCaseService := services.NewTestCaseService(db)

	// Initialize advanced services
	courseScoreService := services.NewCourseScoreService(courseScoreRepo, studentProgressRepo)
	deadlineCheckerService := services.NewDeadlineCheckerService(db)
	progressService := services.NewProgressService(db, userService, courseService)
	draftService := services.NewDraftService(db)

	// Initialize external services
	exec, storageService, err := SetupExternalServices(cfg)
	if err != nil {
		return nil, err
	}

	courseMaterialService := services.NewCourseMaterialService(db, storageService)

	// Initialize queue service with retry logic
	logger.Info("Initializing RabbitMQ service...")
	rabbitMQService, err := external.NewRabbitMQService(&external.RabbitMQConfig{
		URL:      cfg.RabbitMQ.URL,
		Exchange: cfg.RabbitMQ.Exchange,
	})
	if err != nil {
		logger.Warnf("Failed to initialize RabbitMQ service: %v", err)
		logger.Info("Application will continue without RabbitMQ functionality")
		rabbitMQService = nil
	} else {
		logger.Info("RabbitMQ service initialized successfully")
	}

	queueService := services.NewQueueService(db, rabbitMQService, userService)

	// Initialize submission service with all dependencies
	submissionService := services.NewSubmissionService(
		db, testCaseService, userService,
		deadlineCheckerService, courseScoreService, courseMaterialService, exec,
		storageService, queueService,
	)

	// Set submission service in queue service (to avoid circular dependency)
	queueService.SetSubmissionService(submissionService)

	// Start queue consumer if RabbitMQ is available
	if rabbitMQService != nil {
		ctx := context.Background()
		if err := queueService.StartQueueConsumer(ctx); err != nil {
			logger.Warnf("Failed to start queue consumer: %v", err)
		} else {
			logger.Info("Queue consumer started successfully")
		}
	}

	return &Services{
		DB:                     db,
		OAuthService:           oauthService,
		JWTService:             jwtService,
		UserService:            userService,
		CourseService:          courseService,
		TestCaseService:        testCaseService,
		EnrollmentService:      enrollmentService,
		InvitationService:      invitationService,
		SubmissionService:      submissionService,
		ProgressService:        progressService,
		DraftService:           draftService,
		QueueService:           queueService,
		StorageService:         storageService,
		DeadlineCheckerService: deadlineCheckerService,
		CourseScoreService:     courseScoreService,
		CourseMaterialService:  courseMaterialService,
	}, nil
}
