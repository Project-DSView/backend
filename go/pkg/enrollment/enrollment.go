package enrollment

import (
	"fmt"

	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/gofiber/fiber/v2"
)

// EnrollmentValidator provides methods to validate course enrollment access
type EnrollmentValidator struct {
	enrollmentService *services.EnrollmentService
	courseService     *services.CourseService
}

// NewEnrollmentValidator creates a new enrollment validator
func NewEnrollmentValidator(enrollmentService *services.EnrollmentService, courseService *services.CourseService) *EnrollmentValidator {
	return &EnrollmentValidator{
		enrollmentService: enrollmentService,
		courseService:     courseService,
	}
}

// ValidateCourseAccess validates if a user has access to a course
// Returns error if access is denied, nil if access is granted
func (v *EnrollmentValidator) ValidateCourseAccess(c *fiber.Ctx, courseID string) error {
	// Get user claims from context
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return fmt.Errorf("invalid authentication claims")
	}

	// If user is a teacher, check if they are the creator
	if claims.IsTeacher {
		course, err := v.courseService.GetCourseByID(courseID)
		if err != nil {
			return fmt.Errorf("failed to get course: %w", err)
		}
		if course == nil {
			return fmt.Errorf("course not found")
		}
		// If teacher is the creator, grant access
		if course.CreatedBy == claims.UserID {
			return nil
		}
		// If teacher is not the creator, check enrollment
		isEnrolled, err := v.enrollmentService.IsUserEnrolled(courseID, claims.UserID)
		if err != nil {
			return fmt.Errorf("failed to check enrollment: %w", err)
		}
		if !isEnrolled {
			return fmt.Errorf("access denied: not enrolled in course")
		}
		return nil
	}

	// For non-teachers, check if they are enrolled in the course
	isEnrolled, err := v.enrollmentService.IsUserEnrolled(courseID, claims.UserID)
	if err != nil {
		return fmt.Errorf("failed to check enrollment: %w", err)
	}

	if !isEnrolled {
		return fmt.Errorf("access denied: not enrolled in course")
	}

	return nil
}

// ValidateCourseAccessFromClaims validates course access using provided claims
// This is useful when you already have the claims and don't want to extract from context
func (v *EnrollmentValidator) ValidateCourseAccessFromClaims(claims *types.Claims, courseID string) error {
	// If user is a teacher, check if they are the creator
	if claims.IsTeacher {
		course, err := v.courseService.GetCourseByID(courseID)
		if err != nil {
			return fmt.Errorf("failed to get course: %w", err)
		}
		if course == nil {
			return fmt.Errorf("course not found")
		}
		// If teacher is the creator, grant access
		if course.CreatedBy == claims.UserID {
			return nil
		}
		// If teacher is not the creator, check enrollment
		isEnrolled, err := v.enrollmentService.IsUserEnrolled(courseID, claims.UserID)
		if err != nil {
			return fmt.Errorf("failed to check enrollment: %w", err)
		}
		if !isEnrolled {
			return fmt.Errorf("access denied: not enrolled in course")
		}
		return nil
	}

	// For non-teachers, check if they are enrolled in the course
	isEnrolled, err := v.enrollmentService.IsUserEnrolled(courseID, claims.UserID)
	if err != nil {
		return fmt.Errorf("failed to check enrollment: %w", err)
	}

	if !isEnrolled {
		return fmt.Errorf("access denied: not enrolled in course")
	}

	return nil
}

// GetUserEnrolledCourses returns a list of course IDs that the user is enrolled in
// For teachers, this returns all courses (you may want to implement course ownership check)
func (v *EnrollmentValidator) GetUserEnrolledCourses(userID string, isTeacher bool) ([]string, error) {
	// For teachers, you might want to return courses they created/teach
	// For now, we'll return empty slice and let the calling code handle teacher access
	if isTeacher {
		return []string{}, nil
	}

	// For students, get their enrolled courses
	// This would require a new method in enrollment service to get user's courses
	// For now, we'll return empty slice and handle this in the service layer
	return []string{}, nil
}

// IsUserEnrolledInCourse checks if a user is enrolled in a specific course
func (v *EnrollmentValidator) IsUserEnrolledInCourse(userID, courseID string) (bool, error) {
	return v.enrollmentService.IsUserEnrolled(courseID, userID)
}
