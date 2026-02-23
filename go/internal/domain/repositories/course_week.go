package repositories

import (
	"fmt"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"gorm.io/gorm"
)

// CourseWeekRepository defines the interface for course week operations
type CourseWeekRepository interface {
	Create(courseWeek *models.CourseWeek) error
	GetByID(courseWeekID string) (*models.CourseWeek, error)
	GetByCourseAndWeek(courseID string, weekNumber int) (*models.CourseWeek, error)
	GetByCourseID(courseID string) ([]models.CourseWeek, error)
	Update(courseWeek *models.CourseWeek) error
	Delete(courseWeekID string) error
	DeleteByCourseAndWeek(courseID string, weekNumber int) error
	GetOrCreate(courseID string, weekNumber int, createdBy string) (*models.CourseWeek, error)
	GetOrCreateWithTitle(courseID string, weekNumber int, title, description, createdBy string) (*models.CourseWeek, error)
}

// courseWeekRepository implements CourseWeekRepository interface
type courseWeekRepository struct {
	db *gorm.DB
}

// NewCourseWeekRepository creates a new course week repository
func NewCourseWeekRepository(db *gorm.DB) CourseWeekRepository {
	return &courseWeekRepository{db: db}
}

// Create creates a new course week
func (r *courseWeekRepository) Create(courseWeek *models.CourseWeek) error {
	if err := r.db.Create(courseWeek).Error; err != nil {
		return fmt.Errorf("failed to create course week: %w", err)
	}
	return nil
}

// GetByID retrieves a course week by ID
func (r *courseWeekRepository) GetByID(courseWeekID string) (*models.CourseWeek, error) {
	var courseWeek models.CourseWeek
	if err := r.db.Where("course_week_id = ?", courseWeekID).First(&courseWeek).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("course week not found")
		}
		return nil, fmt.Errorf("failed to get course week: %w", err)
	}
	return &courseWeek, nil
}

// GetByCourseAndWeek retrieves a course week by course ID and week number
func (r *courseWeekRepository) GetByCourseAndWeek(courseID string, weekNumber int) (*models.CourseWeek, error) {
	var courseWeek models.CourseWeek
	if err := r.db.Where("course_id = ? AND week_number = ?", courseID, weekNumber).First(&courseWeek).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("course week not found")
		}
		return nil, fmt.Errorf("failed to get course week: %w", err)
	}
	return &courseWeek, nil
}

// GetByCourseID retrieves all course weeks for a specific course
func (r *courseWeekRepository) GetByCourseID(courseID string) ([]models.CourseWeek, error) {
	var courseWeeks []models.CourseWeek
	if err := r.db.Where("course_id = ?", courseID).Order("week_number ASC").Find(&courseWeeks).Error; err != nil {
		return nil, fmt.Errorf("failed to get course weeks: %w", err)
	}
	return courseWeeks, nil
}

// Update updates an existing course week
func (r *courseWeekRepository) Update(courseWeek *models.CourseWeek) error {
	if err := r.db.Save(courseWeek).Error; err != nil {
		return fmt.Errorf("failed to update course week: %w", err)
	}
	return nil
}

// Delete deletes a course week by ID
func (r *courseWeekRepository) Delete(courseWeekID string) error {
	if err := r.db.Where("course_week_id = ?", courseWeekID).Delete(&models.CourseWeek{}).Error; err != nil {
		return fmt.Errorf("failed to delete course week: %w", err)
	}
	return nil
}

// DeleteByCourseAndWeek deletes a course week by course ID and week number
func (r *courseWeekRepository) DeleteByCourseAndWeek(courseID string, weekNumber int) error {
	if err := r.db.Where("course_id = ? AND week_number = ?", courseID, weekNumber).Delete(&models.CourseWeek{}).Error; err != nil {
		return fmt.Errorf("failed to delete course week: %w", err)
	}
	return nil
}

// GetOrCreate retrieves an existing course week or creates a new one with default title
func (r *courseWeekRepository) GetOrCreate(courseID string, weekNumber int, createdBy string) (*models.CourseWeek, error) {
	courseWeek, err := r.GetByCourseAndWeek(courseID, weekNumber)
	if err == nil {
		return courseWeek, nil
	}

	// Create new course week with default title
	newCourseWeek := models.NewCourseWeek(courseID, weekNumber, createdBy)
	if err := r.Create(newCourseWeek); err != nil {
		return nil, err
	}

	return newCourseWeek, nil
}

// GetOrCreateWithTitle retrieves an existing course week or creates a new one with custom title
func (r *courseWeekRepository) GetOrCreateWithTitle(courseID string, weekNumber int, title, description, createdBy string) (*models.CourseWeek, error) {
	courseWeek, err := r.GetByCourseAndWeek(courseID, weekNumber)
	if err == nil {
		// Update existing course week with new title/description
		courseWeek.Title = title
		courseWeek.Description = description
		if err := r.Update(courseWeek); err != nil {
			return nil, err
		}
		return courseWeek, nil
	}

	// Create new course week with custom title
	newCourseWeek := models.NewCourseWeekWithTitle(courseID, weekNumber, title, description, createdBy)
	if err := r.Create(newCourseWeek); err != nil {
		return nil, err
	}

	return newCourseWeek, nil
}
