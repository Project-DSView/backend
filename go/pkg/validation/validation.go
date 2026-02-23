package validation

import (
	"regexp"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ValidateCourseCreation validates course creation data
func ValidateCourseCreation(name, description, enrollKey string) error {
	if strings.TrimSpace(name) == "" {
		return &ValidationError{Field: "name", Message: "Course name is required"}
	}

	if len(name) > 255 {
		return &ValidationError{Field: "name", Message: "Course name must be less than 255 characters"}
	}

	if len(description) > 2000 {
		return &ValidationError{Field: "description", Message: "Description must be less than 2000 characters"}
	}

	if enrollKey != "" {
		if err := ValidateEnrollKey(enrollKey); err != nil {
			return err
		}
	}

	return nil
}

// ValidateEnrollKey validates enrollment key format
func ValidateEnrollKey(enrollKey string) error {
	if len(enrollKey) < 4 || len(enrollKey) > 20 {
		return &ValidationError{Field: "enroll_key", Message: "Enrollment key must be between 4 and 20 characters"}
	}

	// Allow only alphanumeric characters and hyphens
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-_]+$`, enrollKey)
	if !matched {
		return &ValidationError{Field: "enroll_key", Message: "Enrollment key can only contain letters, numbers, hyphens, and underscores"}
	}

	return nil
}

// ValidateCourseUpdate validates course update data
func ValidateCourseUpdate(name, description, enrollKey *string) error {
	if name != nil {
		if strings.TrimSpace(*name) == "" {
			return &ValidationError{Field: "name", Message: "Course name cannot be empty"}
		}
		if len(*name) > 255 {
			return &ValidationError{Field: "name", Message: "Course name must be less than 255 characters"}
		}
	}

	if description != nil && len(*description) > 2000 {
		return &ValidationError{Field: "description", Message: "Description must be less than 2000 characters"}
	}

	if enrollKey != nil {
		if strings.TrimSpace(*enrollKey) == "" {
			return &ValidationError{Field: "enroll_key", Message: "Enrollment key cannot be empty"}
		}
		if err := ValidateEnrollKey(*enrollKey); err != nil {
			return err
		}
	}

	return nil
}

// SanitizeInput sanitizes user input
func SanitizeInput(input string) string {
	return strings.TrimSpace(input)
}
