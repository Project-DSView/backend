package repositories

import (
	"context"

	entities "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/internal/domain/repositories"
	"gorm.io/gorm"
)

// GormStudentProgressRepository implements StudentProgressRepository using GORM
type GormStudentProgressRepository struct {
	db *gorm.DB
}

// NewGormStudentProgressRepository creates a new GORM student progress repository
func NewGormStudentProgressRepository(db *gorm.DB) repositories.StudentProgressRepository {
	return &GormStudentProgressRepository{db: db}
}

// GetByUserAndCourse gets all student progress for a user in a specific course
func (r *GormStudentProgressRepository) GetByUserAndCourse(ctx context.Context, userID, courseID string) ([]entities.StudentProgress, error) {
	var models []entities.StudentProgress

	// Optimized query with JOIN to avoid N+1 problem
	err := r.db.WithContext(ctx).
		Table("student_progress sp").
		Joins("INNER JOIN course_materials cm ON sp.material_id = cm.material_id").
		Where("sp.user_id = ? AND cm.course_id = ?", userID, courseID).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	entities := make([]entities.StudentProgress, len(models))
	for i, model := range models {
		entities[i] = *r.modelToEntity(&model)
	}

	return entities, nil
}

// GetByUserAndMaterial gets student progress for a specific user and material
func (r *GormStudentProgressRepository) GetByUserAndMaterial(ctx context.Context, userID, materialID string) (*entities.StudentProgress, error) {
	var model entities.StudentProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND material_id = ?", userID, materialID).
		First(&model).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.modelToEntity(&model), nil
}

// Create creates a new student progress record
func (r *GormStudentProgressRepository) Create(ctx context.Context, progress *entities.StudentProgress) error {
	model := r.entityToModel(progress)
	return r.db.WithContext(ctx).Create(model).Error
}

// Update updates an existing student progress record
func (r *GormStudentProgressRepository) Update(ctx context.Context, progress *entities.StudentProgress) error {
	model := r.entityToModel(progress)
	return r.db.WithContext(ctx).Save(model).Error
}

// BatchGetByUserAndMaterials gets multiple student progress records efficiently
func (r *GormStudentProgressRepository) BatchGetByUserAndMaterials(ctx context.Context, userID string, materialIDs []string) ([]entities.StudentProgress, error) {
	var models []entities.StudentProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND material_id IN ?", userID, materialIDs).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	entities := make([]entities.StudentProgress, len(models))
	for i, model := range models {
		entities[i] = *r.modelToEntity(&model)
	}

	return entities, nil
}

// entityToModel converts domain entity to GORM model
func (r *GormStudentProgressRepository) entityToModel(entity *entities.StudentProgress) *entities.StudentProgress {
	return &entities.StudentProgress{
		ProgressID:      entity.ProgressID,
		UserID:          entity.UserID,
		MaterialID:      entity.MaterialID,
		Status:          enums.ProgressStatus(entity.Status),
		Score:           entity.Score,
		SeatNumber:      entity.SeatNumber,
		LastSubmittedAt: entity.LastSubmittedAt,
		CreatedAt:       entity.CreatedAt,
		UpdatedAt:       entity.UpdatedAt,
	}
}

// modelToEntity converts GORM model to domain entity
func (r *GormStudentProgressRepository) modelToEntity(model *entities.StudentProgress) *entities.StudentProgress {
	return &entities.StudentProgress{
		ProgressID:      model.ProgressID,
		UserID:          model.UserID,
		MaterialID:      model.MaterialID,
		Status:          enums.ProgressStatus(model.Status),
		Score:           model.Score,
		SeatNumber:      model.SeatNumber,
		LastSubmittedAt: model.LastSubmittedAt,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}
