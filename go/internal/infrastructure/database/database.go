package database

import (
	"fmt"
	"time"

	entities "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/Project-DSView/backend/go/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Database represents the single unified database connection
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates a new database connection to DSView_DB
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := cfg.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Error),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Minute)

	logger.Infof("Database connection pool configured: MaxOpenConns=%d, MaxIdleConns=%d, ConnMaxLifetime=%dm, ConnMaxIdleTime=%dm",
		cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime, cfg.ConnMaxIdleTime)

	// Create database with foreign key constraints temporarily disabled
	dbWithoutFK, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   gormlogger.Default.LogMode(gormlogger.Error),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection for migration: %w", err)
	}

	// Auto-migrate models without foreign key constraints first
	// Note: CourseMaterial is now a central table with polymorphic references
	// Actual data is stored in separate tables: videos, documents, code_exercises, pdf_exercises, announcements
	logger.Info("Starting database migration...")
	if err := dbWithoutFK.AutoMigrate(
		&entities.User{},
		&entities.Course{},
		&entities.Enrollment{},
		&entities.CourseInvitation{},
		// Material tables (separate tables for each type)
		&entities.Video{},
		&entities.Document{},
		&entities.CodeExercise{},
		&entities.PDFExercise{},
		&entities.Announcement{},
		// Central reference table
		&entities.CourseMaterial{},
		&entities.TestCase{},
		&entities.Submission{},
		&entities.SubmissionResult{},
		&entities.StudentProgress{},
		&entities.VerificationLog{},
		&entities.ExerciseDraft{},
		&entities.CourseWeek{},
		&entities.StudentCourseScore{},
		&entities.QueueJob{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate tables: %w", err)
	}
	logger.Info("Database migration completed successfully")

	// Now create foreign key constraints (excluding QueueJob and TestCase to avoid constraint issues)
	// Note: TestCase is excluded because its FK points FROM test_cases TO course_materials,
	// and GORM would try to create an inverse FK on test_cases(material_id) which is not unique.
	logger.Info("Creating foreign key constraints...")
	if err := db.AutoMigrate(
		&entities.Enrollment{},
		// Material tables
		&entities.Video{},
		&entities.Document{},
		&entities.CodeExercise{},
		&entities.PDFExercise{},
		&entities.Announcement{},
		// Central reference table
		&entities.CourseMaterial{},
		&entities.CourseWeek{},
		&entities.StudentCourseScore{},
	); err != nil {
		logger.Warnf("Could not create foreign key constraints: %v", err)
	}

	// Create optimized indexes
	logger.Info("Creating database indexes...")
	if err := CreateOptimizedIndexes(db); err != nil {
		logger.Warnf("Failed to create optimized indexes: %v", err)
	}

	logger.Info("Database connected and migrated successfully")

	database := &Database{DB: db}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the database instance for use in services
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// Health check method
func (d *Database) HealthCheck() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// CreateOptimizedIndexes creates database indexes for performance optimization
func CreateOptimizedIndexes(db *gorm.DB) error {
	indexes := []string{
		// Student Progress indexes
		"CREATE INDEX IF NOT EXISTS idx_student_progress_user_material ON student_progress(user_id, material_id)",
		"CREATE INDEX IF NOT EXISTS idx_student_progress_status ON student_progress(status)",
		"CREATE INDEX IF NOT EXISTS idx_student_progress_last_submitted ON student_progress(last_submitted_at)",

		// Enrollments indexes
		"CREATE INDEX IF NOT EXISTS idx_enrollments_user_course ON enrollments(user_id, course_id)",
		"CREATE INDEX IF NOT EXISTS idx_enrollments_course ON enrollments(course_id)",

		// Videos indexes
		"CREATE INDEX IF NOT EXISTS idx_videos_course_id ON videos(course_id)",
		"CREATE INDEX IF NOT EXISTS idx_videos_week ON videos(week)",
		"CREATE INDEX IF NOT EXISTS idx_videos_created_by ON videos(created_by)",

		// Documents indexes
		"CREATE INDEX IF NOT EXISTS idx_documents_course_id ON documents(course_id)",
		"CREATE INDEX IF NOT EXISTS idx_documents_week ON documents(week)",
		"CREATE INDEX IF NOT EXISTS idx_documents_created_by ON documents(created_by)",

		// Code Exercises indexes
		"CREATE INDEX IF NOT EXISTS idx_code_exercises_course_id ON code_exercises(course_id)",
		"CREATE INDEX IF NOT EXISTS idx_code_exercises_week ON code_exercises(week)",
		"CREATE INDEX IF NOT EXISTS idx_code_exercises_created_by ON code_exercises(created_by)",

		// PDF Exercises indexes
		"CREATE INDEX IF NOT EXISTS idx_pdf_exercises_course_id ON pdf_exercises(course_id)",
		"CREATE INDEX IF NOT EXISTS idx_pdf_exercises_week ON pdf_exercises(week)",
		"CREATE INDEX IF NOT EXISTS idx_pdf_exercises_created_by ON pdf_exercises(created_by)",

		// Announcements indexes
		"CREATE INDEX IF NOT EXISTS idx_announcements_course_id ON announcements(course_id)",
		"CREATE INDEX IF NOT EXISTS idx_announcements_week ON announcements(week)",
		"CREATE INDEX IF NOT EXISTS idx_announcements_created_by ON announcements(created_by)",
		"CREATE INDEX IF NOT EXISTS idx_announcements_created_at ON announcements(created_at)",

		// Course Scores indexes
		"CREATE INDEX IF NOT EXISTS idx_student_course_scores_user_course ON student_course_scores(user_id, course_id)",
		"CREATE INDEX IF NOT EXISTS idx_student_course_scores_course ON student_course_scores(course_id)",
		"CREATE INDEX IF NOT EXISTS idx_student_course_scores_total_score ON student_course_scores(total_score DESC)",

		// Submissions indexes
		"CREATE INDEX IF NOT EXISTS idx_submissions_user_material ON submissions(user_id, material_id)",
		"CREATE INDEX IF NOT EXISTS idx_submissions_material ON submissions(material_id)",
		"CREATE INDEX IF NOT EXISTS idx_submissions_submitted_at ON submissions(submitted_at)",

		// Users indexes
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_is_teacher ON users(is_teacher)",

		// Course Materials indexes (central reference table)
		"CREATE INDEX IF NOT EXISTS idx_course_materials_course ON course_materials(course_id)",
		"CREATE INDEX IF NOT EXISTS idx_course_materials_course_week ON course_materials(course_id, week)",
		"CREATE INDEX IF NOT EXISTS idx_course_materials_type ON course_materials(type)",
		"CREATE INDEX IF NOT EXISTS idx_course_materials_reference_id ON course_materials(reference_id)",
		"CREATE INDEX IF NOT EXISTS idx_course_materials_reference_type ON course_materials(reference_type)",

		// Course Weeks indexes
		"CREATE INDEX IF NOT EXISTS idx_course_weeks_course ON course_weeks(course_id)",
		"CREATE INDEX IF NOT EXISTS idx_course_weeks_course_week ON course_weeks(course_id, week_number)",
		"CREATE INDEX IF NOT EXISTS idx_course_weeks_week_number ON course_weeks(week_number)",

		// Exercise Drafts indexes
		"CREATE INDEX IF NOT EXISTS idx_exercise_drafts_user_material ON exercise_drafts(user_id, material_id)",
		"CREATE INDEX IF NOT EXISTS idx_exercise_drafts_user ON exercise_drafts(user_id)",

		// Queue Jobs indexes
		"CREATE INDEX IF NOT EXISTS idx_queue_jobs_status ON queue_jobs(status)",
		"CREATE INDEX IF NOT EXISTS idx_queue_jobs_created_at ON queue_jobs(created_at)",

		// Test Cases indexes
		"CREATE INDEX IF NOT EXISTS idx_test_cases_material ON test_cases(material_id)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			logger.Warnf("Failed to create index: %s, error: %v", indexSQL, err)
		}
	}

	// Create additional specialized indexes
	createCourseScoreIndexes(db)
	createStudentProgressIndexes(db)
	createSubmissionIndexes(db)

	logger.Info("Database indexes created successfully")
	return nil
}
