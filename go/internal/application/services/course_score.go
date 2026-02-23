package services

import (
	"context"
	"fmt"

	entities "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/repositories"
)

// CourseScoreService handles course score business logic
type CourseScoreService struct {
	courseScoreRepo     repositories.CourseScoreRepository
	studentProgressRepo repositories.StudentProgressRepository
}

// NewCourseScoreService creates a new course score service
func NewCourseScoreService(
	courseScoreRepo repositories.CourseScoreRepository,
	studentProgressRepo repositories.StudentProgressRepository,
) *CourseScoreService {
	return &CourseScoreService{
		courseScoreRepo:     courseScoreRepo,
		studentProgressRepo: studentProgressRepo,
	}
}

// UpdateCourseScore updates the total score for a student in a course
// This method is optimized to avoid N+1 queries by using batch operations
func (s *CourseScoreService) UpdateCourseScore(ctx context.Context, userID, courseID string) error {
	// Get all student progress for the student in this course (single query)
	progressList, err := s.studentProgressRepo.GetByUserAndCourse(ctx, userID, courseID)
	if err != nil {
		return fmt.Errorf("failed to get student progress: %w", err)
	}

	if len(progressList) == 0 {
		return nil // No progress to calculate
	}

	// Calculate total score directly from progress (no need for exercise lookup)

	// Calculate total score
	totalScore := 0
	for _, progress := range progressList {
		totalScore += progress.Score
	}

	// Create or update course score
	courseScore := entities.NewCourseScore(userID, courseID, totalScore)

	// Check if course score already exists
	existingScore, err := s.courseScoreRepo.GetByUserAndCourse(ctx, userID, courseID)
	if err != nil {
		return fmt.Errorf("failed to check existing course score: %w", err)
	}

	if existingScore == nil {
		// Create new course score
		if err := s.courseScoreRepo.Create(ctx, courseScore); err != nil {
			return fmt.Errorf("failed to create course score: %w", err)
		}
	} else {
		// Update existing course score
		existingScore.UpdateScore(totalScore)
		if err := s.courseScoreRepo.Update(ctx, existingScore); err != nil {
			return fmt.Errorf("failed to update course score: %w", err)
		}
	}

	return nil
}

// GetStudentCourseScore gets the total score for a student in a course
func (s *CourseScoreService) GetStudentCourseScore(ctx context.Context, userID, courseID string) (*entities.CourseScore, error) {
	// Get from database
	courseScorePtr, err := s.courseScoreRepo.GetByUserAndCourse(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}

	if courseScorePtr == nil {
		return nil, nil
	}

	return courseScorePtr, nil
}


// GetCourseScoreStats gets statistics for scores in a course
func (s *CourseScoreService) GetCourseScoreStats(ctx context.Context, courseID string) (*repositories.CourseScoreStats, error) {
	// Get from database
	statsPtr, err := s.courseScoreRepo.GetCourseScoreStats(ctx, courseID)
	if err != nil {
		return nil, err
	}

	return statsPtr, nil
}

// GetStudentScoreStats gets statistics for a student's scores across all courses
func (s *CourseScoreService) GetStudentScoreStats(ctx context.Context, userID string) (*repositories.StudentScoreStats, error) {
	// Get from database
	statsPtr, err := s.courseScoreRepo.GetStudentScoreStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	return statsPtr, nil
}

// BatchUpdateCourseScores updates course scores for multiple students efficiently
func (s *CourseScoreService) BatchUpdateCourseScores(ctx context.Context, userIDs []string, courseID string) error {
	// Get all student progress for all users in this course (single query)
	progressMap := make(map[string][]entities.StudentProgress)

	for _, userID := range userIDs {
		progressList, err := s.studentProgressRepo.GetByUserAndCourse(ctx, userID, courseID)
		if err != nil {
			return fmt.Errorf("failed to get student progress for user %s: %w", userID, err)
		}
		progressMap[userID] = progressList
	}

	// Calculate scores directly from progress (no exercise lookup needed)

	// Calculate and update course scores for each user
	for _, userID := range userIDs {
		progressList := progressMap[userID]

		totalScore := 0
		for _, progress := range progressList {
			totalScore += progress.Score
		}

		courseScore := entities.NewCourseScore(userID, courseID, totalScore)

		// Check if course score already exists
		existingScore, err := s.courseScoreRepo.GetByUserAndCourse(ctx, userID, courseID)
		if err != nil {
			return fmt.Errorf("failed to check existing course score for user %s: %w", userID, err)
		}

		if existingScore == nil {
			// Create new course score
			if err := s.courseScoreRepo.Create(ctx, courseScore); err != nil {
				return fmt.Errorf("failed to create course score for user %s: %w", userID, err)
			}
		} else {
			// Update existing course score
			existingScore.UpdateScore(totalScore)
			if err := s.courseScoreRepo.Update(ctx, existingScore); err != nil {
				return fmt.Errorf("failed to update course score for user %s: %w", userID, err)
			}
		}

	}

	return nil
}
