package validation

import (
	"regexp"
	"strings"
	"time"

	"github.com/Project-DSView/backend/go/pkg/errors"
)

// Common validation patterns
var (
	EmailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	UUIDRegex     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	UsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)
)

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid bool
	Errors  []*errors.AppError
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		IsValid: true,
		Errors:  make([]*errors.AppError, 0),
	}
}

// AddError adds an error to the validation result
func (vr *ValidationResult) AddError(err *errors.AppError) {
	vr.IsValid = false
	vr.Errors = append(vr.Errors, err)
}

// AddFieldError adds a field-specific error
func (vr *ValidationResult) AddFieldError(field, message string) {
	vr.AddError(errors.NewValidationError(message, field))
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// GetFirstError returns the first validation error
func (vr *ValidationResult) GetFirstError() *errors.AppError {
	if len(vr.Errors) > 0 {
		return vr.Errors[0]
	}
	return nil
}

// Common validation functions

// ValidateEmail validates an email address
func ValidateEmail(email string) *errors.AppError {
	if email == "" {
		return errors.NewValidationError("Email is required", "email")
	}
	if !EmailRegex.MatchString(email) {
		return errors.NewValidationError("Invalid email format", "email")
	}
	return nil
}

// ValidateUUID validates a UUID string
func ValidateUUID(id, fieldName string) *errors.AppError {
	if id == "" {
		return errors.NewValidationError(fieldName+" is required", fieldName)
	}
	if !UUIDRegex.MatchString(id) {
		return errors.NewValidationError("Invalid "+fieldName+" format", fieldName)
	}
	return nil
}

// ValidateRequired validates that a field is not empty
func ValidateRequired(value, fieldName string) *errors.AppError {
	if strings.TrimSpace(value) == "" {
		return errors.NewValidationError(fieldName+" is required", fieldName)
	}
	return nil
}

// ValidateLength validates string length
func ValidateLength(value, fieldName string, min, max int) *errors.AppError {
	length := len(strings.TrimSpace(value))
	if length < min {
		return errors.NewValidationError(fieldName+" must be at least "+string(rune(min))+" characters", fieldName)
	}
	if length > max {
		return errors.NewValidationError(fieldName+" must be no more than "+string(rune(max))+" characters", fieldName)
	}
	return nil
}

// ValidateRange validates numeric range
func ValidateRange(value int, fieldName string, min, max int) *errors.AppError {
	if value < min {
		return errors.NewValidationError(fieldName+" must be at least "+string(rune(min)), fieldName)
	}
	if value > max {
		return errors.NewValidationError(fieldName+" must be no more than "+string(rune(max)), fieldName)
	}
	return nil
}

// ValidateDate validates date format and range
func ValidateDate(date time.Time, fieldName string, notBefore, notAfter *time.Time) *errors.AppError {
	if date.IsZero() {
		return errors.NewValidationError(fieldName+" is required", fieldName)
	}

	if notBefore != nil && date.Before(*notBefore) {
		return errors.NewValidationError(fieldName+" cannot be before "+notBefore.Format("2006-01-02"), fieldName)
	}

	if notAfter != nil && date.After(*notAfter) {
		return errors.NewValidationError(fieldName+" cannot be after "+notAfter.Format("2006-01-02"), fieldName)
	}

	return nil
}

// ValidateUserUpdate validates user update data
func ValidateUserUpdate(firstName, lastName string) *ValidationResult {
	result := NewValidationResult()

	// Validate first name
	if firstName != "" {
		if err := ValidateLength(firstName, "firstname", 1, 255); err != nil {
			result.AddError(err)
		}
	}

	// Validate last name
	if lastName != "" {
		if err := ValidateLength(lastName, "lastname", 1, 255); err != nil {
			result.AddError(err)
		}
	}

	return result
}

// SanitizeUserInput sanitizes user input
func SanitizeUserInput(input string) string {
	// Remove leading/trailing whitespace
	input = strings.TrimSpace(input)

	// Remove potentially dangerous characters
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#x27;")

	return input
}

// ParseFullName parses full name into first and last name
func ParseFullName(fullName, givenName, familyName string) (string, string) {
	firstName := givenName
	lastName := familyName

	// If given name and family name are not provided, try to parse from full name
	if firstName == "" && lastName == "" && fullName != "" {
		parts := strings.Fields(strings.TrimSpace(fullName))
		if len(parts) >= 2 {
			firstName = parts[0]
			lastName = strings.Join(parts[1:], " ")
		} else if len(parts) == 1 {
			firstName = parts[0]
			lastName = "User"
		}
	}

	// Fallback values
	if firstName == "" {
		firstName = "Unknown"
	}
	if lastName == "" {
		lastName = "User"
	}

	return firstName, lastName
}

// BuildGoogleUserUpdates builds update map for Google user data
func BuildGoogleUserUpdates(firstName, lastName, profileImg string) map[string]interface{} {
	updates := make(map[string]interface{})

	if firstName != "" {
		updates["first_name"] = SanitizeUserInput(firstName)
	}
	if lastName != "" {
		updates["last_name"] = SanitizeUserInput(lastName)
	}
	if profileImg != "" {
		updates["profile_img"] = profileImg
	}

	return updates
}
