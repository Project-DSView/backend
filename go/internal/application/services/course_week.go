package services

import (
	"fmt"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/repositories"
)

// CourseWeekService defines the interface for course week business logic
type CourseWeekService interface {
	CreateCourseWeek(courseID string, weekNumber int, title, description, createdBy string) (*models.CourseWeek, error)
	GetCourseWeek(courseID string, weekNumber int) (*models.CourseWeek, error)
	GetCourseWeeks(courseID string) ([]models.CourseWeek, error)
	UpdateCourseWeek(courseID string, weekNumber int, title, description string) (*models.CourseWeek, error)
	DeleteCourseWeek(courseID string, weekNumber int) error
	GetOrCreateCourseWeek(courseID string, weekNumber int, createdBy string) (*models.CourseWeek, error)
	GetWeekTitle(courseID string, weekNumber int) string
}

// courseWeekService implements CourseWeekService interface
type courseWeekService struct {
	courseWeekRepo repositories.CourseWeekRepository
}

// NewCourseWeekService creates a new course week service
func NewCourseWeekService(courseWeekRepo repositories.CourseWeekRepository) CourseWeekService {
	return &courseWeekService{
		courseWeekRepo: courseWeekRepo,
	}
}

// CreateCourseWeek creates a new course week
func (s *courseWeekService) CreateCourseWeek(courseID string, weekNumber int, title, description, createdBy string) (*models.CourseWeek, error) {
	// Validate week number
	if weekNumber < 1 || weekNumber > 52 {
		return nil, fmt.Errorf("week number must be between 1 and 52")
	}

	// Check if week already exists
	existingWeek, err := s.courseWeekRepo.GetByCourseAndWeek(courseID, weekNumber)
	if err == nil && existingWeek != nil {
		return nil, fmt.Errorf("week %d already exists for this course", weekNumber)
	}

	// Create new course week
	courseWeek := models.NewCourseWeekWithTitle(courseID, weekNumber, title, description, createdBy)
	if err := s.courseWeekRepo.Create(courseWeek); err != nil {
		return nil, fmt.Errorf("failed to create course week: %w", err)
	}

	return courseWeek, nil
}

// GetCourseWeek retrieves a course week by course ID and week number
func (s *courseWeekService) GetCourseWeek(courseID string, weekNumber int) (*models.CourseWeek, error) {
	courseWeek, err := s.courseWeekRepo.GetByCourseAndWeek(courseID, weekNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get course week: %w", err)
	}
	return courseWeek, nil
}

// GetCourseWeeks retrieves all course weeks for a specific course
func (s *courseWeekService) GetCourseWeeks(courseID string) ([]models.CourseWeek, error) {
	courseWeeks, err := s.courseWeekRepo.GetByCourseID(courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course weeks: %w", err)
	}
	return courseWeeks, nil
}

// UpdateCourseWeek updates an existing course week
func (s *courseWeekService) UpdateCourseWeek(courseID string, weekNumber int, title, description string) (*models.CourseWeek, error) {
	courseWeek, err := s.courseWeekRepo.GetByCourseAndWeek(courseID, weekNumber)
	if err != nil {
		return nil, fmt.Errorf("course week not found: %w", err)
	}

	// Update fields
	courseWeek.Title = title
	courseWeek.Description = description

	if err := s.courseWeekRepo.Update(courseWeek); err != nil {
		return nil, fmt.Errorf("failed to update course week: %w", err)
	}

	return courseWeek, nil
}

// DeleteCourseWeek deletes a course week
func (s *courseWeekService) DeleteCourseWeek(courseID string, weekNumber int) error {
	if err := s.courseWeekRepo.DeleteByCourseAndWeek(courseID, weekNumber); err != nil {
		return fmt.Errorf("failed to delete course week: %w", err)
	}
	return nil
}

// GetOrCreateCourseWeek retrieves an existing course week or creates a new one
func (s *courseWeekService) GetOrCreateCourseWeek(courseID string, weekNumber int, createdBy string) (*models.CourseWeek, error) {
	courseWeek, err := s.courseWeekRepo.GetOrCreate(courseID, weekNumber, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create course week: %w", err)
	}
	return courseWeek, nil
}

// GetWeekTitle returns the title for a specific week, or default title if not found
func (s *courseWeekService) GetWeekTitle(courseID string, weekNumber int) string {
	courseWeek, err := s.courseWeekRepo.GetByCourseAndWeek(courseID, weekNumber)
	if err != nil || courseWeek == nil {
		return models.GetDefaultTitle(weekNumber)
	}
	return courseWeek.Title
}
