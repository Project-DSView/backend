package container

import (
	"context"
	"fmt"

	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/domain/repositories"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/Project-DSView/backend/go/internal/infrastructure/database"
	gorm_repositories "github.com/Project-DSView/backend/go/internal/infrastructure/repositories"
	"github.com/Project-DSView/backend/go/pkg/logger"

	"gorm.io/gorm"
)

// Container holds all dependencies
type Container struct {
	// Database
	DB *gorm.DB

	// Repositories
	CourseScoreRepo     repositories.CourseScoreRepository
	StudentProgressRepo repositories.StudentProgressRepository

	// Services
	CourseScoreService *services.CourseScoreService
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config) (*Container, error) {
	container := &Container{}

	// Initialize database
	if err := container.initDatabase(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize repositories
	container.initRepositories()

	// Initialize services
	container.initServices()

	// Create database indexes for optimization
	if err := container.createIndexes(); err != nil {
		logger.Warnf("Failed to create database indexes: %v", err)
	}

	return container, nil
}

// initDatabase initializes the database connection
func (c *Container) initDatabase(cfg *config.Config) error {
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		return err
	}
	c.DB = db.GetDB()
	return nil
}

// initRepositories initializes all repositories
func (c *Container) initRepositories() {
	c.CourseScoreRepo = gorm_repositories.NewGormCourseScoreRepository(c.DB)
	c.StudentProgressRepo = gorm_repositories.NewGormStudentProgressRepository(c.DB)
}

// initServices initializes all services
func (c *Container) initServices() {
	c.CourseScoreService = services.NewCourseScoreService(
		c.CourseScoreRepo,
		c.StudentProgressRepo,
	)
}

// createIndexes creates database indexes for performance optimization
func (c *Container) createIndexes() error {
	return database.CreateOptimizedIndexes(c.DB)
}

// HealthCheck performs health checks on all dependencies
func (c *Container) HealthCheck(ctx context.Context) error {
	// Check database
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Close closes all connections
func (c *Container) Close() error {
	var err error

	// Close database
	if sqlDB, dbErr := c.DB.DB(); dbErr == nil {
		if closeErr := sqlDB.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}

	return err
}
