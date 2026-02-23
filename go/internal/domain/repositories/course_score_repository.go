package repositories

import (
	"context"

	entities "github.com/Project-DSView/backend/go/internal/domain/entities"
)

// CourseScoreRepository defines the interface for course score data operations
type CourseScoreRepository interface {
	// Create creates a new course score
	Create(ctx context.Context, courseScore *entities.CourseScore) error

	// Update updates an existing course score
	Update(ctx context.Context, courseScore *entities.CourseScore) error

	// GetByUserAndCourse gets a course score by user ID and course ID
	GetByUserAndCourse(ctx context.Context, userID, courseID string) (*entities.CourseScore, error)

	// GetCourseScoreStats gets statistics for scores in a course
	GetCourseScoreStats(ctx context.Context, courseID string) (*CourseScoreStats, error)

	// GetStudentScoreStats gets statistics for a student's scores across all courses
	GetStudentScoreStats(ctx context.Context, userID string) (*StudentScoreStats, error)

	// BatchGetByUserAndCourses gets multiple course scores by user ID and course IDs
	BatchGetByUserAndCourses(ctx context.Context, userID string, courseIDs []string) ([]entities.CourseScore, error)
}

// CourseScoreStats represents statistics for course scores
type CourseScoreStats struct {
	TotalStudents     int64
	AverageScore      float64
	AveragePercentage float64
	PassedStudents    int64
	PassRate          float64
	GradeDistribution map[string]int64
}

// StudentScoreStats represents statistics for student scores
type StudentScoreStats struct {
	TotalCourses      int64
	AveragePercentage float64
	TotalScore        int64
	TotalMaxScore     int64
}
