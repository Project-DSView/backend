package repositories

import (
	"context"

	entities "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/repositories"
	"gorm.io/gorm"
)

// GormCourseScoreRepository implements CourseScoreRepository using GORM
type GormCourseScoreRepository struct {
	db *gorm.DB
}

// NewGormCourseScoreRepository creates a new GORM course score repository
func NewGormCourseScoreRepository(db *gorm.DB) repositories.CourseScoreRepository {
	return &GormCourseScoreRepository{db: db}
}

// Create creates a new course score
func (r *GormCourseScoreRepository) Create(ctx context.Context, courseScore *entities.CourseScore) error {
	model := r.entityToModel(courseScore)
	return r.db.WithContext(ctx).Create(model).Error
}

// Update updates an existing course score
func (r *GormCourseScoreRepository) Update(ctx context.Context, courseScore *entities.CourseScore) error {
	model := r.entityToModel(courseScore)
	return r.db.WithContext(ctx).Save(model).Error
}

// GetByUserAndCourse gets a course score by user ID and course ID
func (r *GormCourseScoreRepository) GetByUserAndCourse(ctx context.Context, userID, courseID string) (*entities.CourseScore, error) {
	var model entities.StudentCourseScore
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		First(&model).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.modelToEntity(&model), nil
}

// GetCourseScoreStats gets statistics for scores in a course
func (r *GormCourseScoreRepository) GetCourseScoreStats(ctx context.Context, courseID string) (*repositories.CourseScoreStats, error) {
	var stats struct {
		TotalStudents     int64
		AverageScore      float64
		AveragePercentage float64
		PassedStudents    int64
	}

	// Use a single query to get all statistics
	err := r.db.WithContext(ctx).
		Model(&entities.StudentCourseScore{}).
		Where("course_id = ?", courseID).
		Select(`
			COUNT(*) as total_students,
			AVG(total_score) as average_score,
			AVG(total_score) as average_percentage,
			COUNT(CASE WHEN total_score >= 60 THEN 1 END) as passed_students
		`).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// Get score distribution based on actual scores (not percentage)
	scoreDistribution := make(map[string]int64)
	scoreRanges := []string{"0-20", "21-40", "41-60", "61-80", "81-100"}

	var scoreCounts []struct {
		ScoreRange string
		Count      int64
	}

	err = r.db.WithContext(ctx).
		Model(&entities.StudentCourseScore{}).
		Where("course_id = ?", courseID).
		Select(`
			CASE 
				WHEN total_score <= 20 THEN '0-20'
				WHEN total_score <= 40 THEN '21-40'
				WHEN total_score <= 60 THEN '41-60'
				WHEN total_score <= 80 THEN '61-80'
				ELSE '81-100'
			END as score_range,
			COUNT(*) as count
		`).
		Group("score_range").
		Find(&scoreCounts).Error

	if err != nil {
		return nil, err
	}

	for _, sc := range scoreCounts {
		scoreDistribution[sc.ScoreRange] = sc.Count
	}

	// Ensure all score ranges are present
	for _, range_ := range scoreRanges {
		if _, exists := scoreDistribution[range_]; !exists {
			scoreDistribution[range_] = 0
		}
	}

	passRate := 0.0
	if stats.TotalStudents > 0 {
		passRate = float64(stats.PassedStudents) / float64(stats.TotalStudents) * 100
	}

	return &repositories.CourseScoreStats{
		TotalStudents:     stats.TotalStudents,
		AverageScore:      stats.AverageScore,
		AveragePercentage: stats.AveragePercentage,
		PassedStudents:    stats.PassedStudents,
		PassRate:          passRate,
		GradeDistribution: scoreDistribution,
	}, nil
}

// GetStudentScoreStats gets statistics for a student's scores across all courses
func (r *GormCourseScoreRepository) GetStudentScoreStats(ctx context.Context, userID string) (*repositories.StudentScoreStats, error) {
	var stats struct {
		TotalCourses      int64
		AveragePercentage float64
		TotalScore        int64
		TotalMaxScore     int64
	}

	err := r.db.WithContext(ctx).
		Model(&entities.StudentCourseScore{}).
		Where("user_id = ?", userID).
		Select(`
			COUNT(*) as total_courses,
			AVG(total_score) as average_percentage,
			SUM(total_score) as total_score,
			COUNT(*) as total_max_score
		`).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return &repositories.StudentScoreStats{
		TotalCourses:      stats.TotalCourses,
		AveragePercentage: stats.AveragePercentage,
		TotalScore:        stats.TotalScore,
		TotalMaxScore:     stats.TotalMaxScore,
	}, nil
}

// BatchGetByUserAndCourses gets multiple course scores by user ID and course IDs
func (r *GormCourseScoreRepository) BatchGetByUserAndCourses(ctx context.Context, userID string, courseIDs []string) ([]entities.CourseScore, error) {
	var models []entities.StudentCourseScore
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND course_id IN ?", userID, courseIDs).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	entities := make([]entities.CourseScore, len(models))
	for i, model := range models {
		entities[i] = *r.modelToEntity(&model)
	}

	return entities, nil
}

// entityToModel converts domain entity to GORM model
func (r *GormCourseScoreRepository) entityToModel(entity *entities.CourseScore) *entities.StudentCourseScore {
	return &entities.StudentCourseScore{
		UserID:      entity.UserID,
		CourseID:    entity.CourseID,
		TotalScore:  entity.TotalScore,
		LastUpdated: entity.LastUpdated,
		CreatedAt:   entity.CreatedAt,
	}
}

// modelToEntity converts GORM model to domain entity
func (r *GormCourseScoreRepository) modelToEntity(model *entities.StudentCourseScore) *entities.CourseScore {
	return &entities.CourseScore{
		UserID:      model.UserID,
		CourseID:    model.CourseID,
		TotalScore:  model.TotalScore,
		LastUpdated: model.LastUpdated,
		CreatedAt:   model.CreatedAt,
	}
}
