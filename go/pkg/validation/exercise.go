package validation

import (
	"fmt"
	"time"

	"github.com/Project-DSView/backend/go/internal/api/types"
	"github.com/gofiber/fiber/v2"
)

// ValidateCreateExerciseRequest validates exercise creation request
func ValidateCreateExerciseRequest(req *types.CreateExerciseRequest) error {
	if req.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Title is required")
	}
	if len(req.Title) > 255 {
		return fiber.NewError(fiber.StatusBadRequest, "Title must be less than 255 characters")
	}
	if req.Description == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Description is required")
	}
	if len(req.Description) > 5000 {
		return fiber.NewError(fiber.StatusBadRequest, "Description must be less than 5000 characters")
	}
	if req.TotalPoints == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Total points is required")
	}
	if req.DataTypeAllowed < 0 || req.DataTypeAllowed > 2 {
		return fiber.NewError(fiber.StatusBadRequest, "Data type allowed must be 0 (Any), 1 (Numbers), or 2 (Strings)")
	}

	// Validate deadline format if provided
	if req.Deadline != nil && *req.Deadline != "" {
		if _, err := time.Parse(time.RFC3339, *req.Deadline); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid deadline format. Use RFC3339 format (e.g., 2025-12-31T23:59:59Z)")
		}

		// Check if deadline is in the future
		if deadline, _ := time.Parse(time.RFC3339, *req.Deadline); deadline.Before(time.Now()) {
			return fiber.NewError(fiber.StatusBadRequest, "Deadline must be in the future")
		}
	}

	return nil
}

// CourseAccessValidator interface for validating course access
type CourseAccessValidator interface {
	GetUserAccessibleCourses(userID string, isTeacher bool) ([]CourseInfo, error)
}

// CourseInfo represents course information for access validation
type CourseInfo struct {
	CourseID string
}

// ValidateCourseAccess validates if user has access to specified courses
func ValidateCourseAccess(courseIDs []string, userID string, isTeacher bool, validator CourseAccessValidator) *fiber.Error {
	accessibleCourses, err := validator.GetUserAccessibleCourses(userID, isTeacher)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError,
			"Failed to check course access: "+err.Error())
	}

	// Create map of accessible course IDs
	accessibleMap := make(map[string]bool, len(accessibleCourses))
	for _, course := range accessibleCourses {
		accessibleMap[course.CourseID] = true
	}

	// Collect invalid courses
	var invalidCourses []string
	for _, courseID := range courseIDs {
		if !accessibleMap[courseID] {
			invalidCourses = append(invalidCourses, courseID)
		}
	}

	// Validate
	if n := len(invalidCourses); n > 0 {
		msg := fmt.Sprintf("You don't have access to course: %s", invalidCourses[0])
		if n > 1 {
			msg = fmt.Sprintf("You don't have access to courses: %s and %d others", invalidCourses[0], n-1)
		}
		return fiber.NewError(fiber.StatusForbidden, msg)
	}

	return nil
}

// ValidateUpdateExerciseRequest validates exercise update request
func ValidateUpdateExerciseRequest(req *types.UpdateExerciseRequest) error {
	if req.Title != nil {
		if *req.Title == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Title cannot be empty")
		}
		if len(*req.Title) > 255 {
			return fiber.NewError(fiber.StatusBadRequest, "Title must be less than 255 characters")
		}
	}

	if req.Description != nil {
		if *req.Description == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Description cannot be empty")
		}
		if len(*req.Description) > 5000 {
			return fiber.NewError(fiber.StatusBadRequest, "Description must be less than 5000 characters")
		}
	}

	if req.DataTypeAllowed != nil {
		if *req.DataTypeAllowed < 0 || *req.DataTypeAllowed > 2 {
			return fiber.NewError(fiber.StatusBadRequest, "Data type allowed must be 0 (Any), 1 (Numbers), or 2 (Strings)")
		}
	}

	if req.TotalPoints != nil && *req.TotalPoints == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Total points cannot be empty")
	}

	// Validate deadline format if provided
	if req.Deadline != nil && *req.Deadline != "" {
		if _, err := time.Parse(time.RFC3339, *req.Deadline); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid deadline format. Use RFC3339 format (e.g., 2025-12-31T23:59:59Z)")
		}

		// Check if deadline is in the future
		if deadline, _ := time.Parse(time.RFC3339, *req.Deadline); deadline.Before(time.Now()) {
			return fiber.NewError(fiber.StatusBadRequest, "Deadline must be in the future")
		}
	}

	return nil
}

// ValidateCreateTestCaseRequest validates test case creation request
func ValidateCreateTestCaseRequest(req *types.CreateTestCaseRequest) error {
	if req.InputData == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Input data is required")
	}
	if req.ExpectedOutput == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Expected output is required")
	}
	return nil
}
