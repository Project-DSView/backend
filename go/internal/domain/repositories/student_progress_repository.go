package repositories

import (
	"context"

	entities "github.com/Project-DSView/backend/go/internal/domain/entities"
)

// StudentProgressRepository defines the interface for student progress data operations
type StudentProgressRepository interface {
	// GetByUserAndCourse gets all student progress for a user in a specific course
	GetByUserAndCourse(ctx context.Context, userID, courseID string) ([]entities.StudentProgress, error)

	// GetByUserAndMaterial gets student progress for a specific user and material
	GetByUserAndMaterial(ctx context.Context, userID, materialID string) (*entities.StudentProgress, error)

	// Create creates a new student progress record
	Create(ctx context.Context, progress *entities.StudentProgress) error

	// Update updates an existing student progress record
	Update(ctx context.Context, progress *entities.StudentProgress) error

	// BatchGetByUserAndMaterials gets multiple student progress records efficiently
	BatchGetByUserAndMaterials(ctx context.Context, userID string, materialIDs []string) ([]entities.StudentProgress, error)
}
