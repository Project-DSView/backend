package response

import (
	"github.com/Project-DSView/backend/go/pkg/errors"
	"github.com/gofiber/fiber/v2"
)

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Type    string `json:"type"`
	Code    int    `json:"code"`
	Field   string `json:"field,omitempty"`
	Details string `json:"details,omitempty"`
}

// SendAppError sends an AppError response
func SendAppError(c *fiber.Ctx, appErr *errors.AppError) error {
	errorInfo := &ErrorInfo{
		Type:    string(appErr.Type),
		Code:    appErr.Code,
		Field:   appErr.Field,
		Details: appErr.Details,
	}

	return c.Status(appErr.Code).JSON(StandardResponse{
		Success: false,
		Message: appErr.Message,
		Error:   errorInfo,
	})
}

// SendGenericError sends a generic error response
func SendGenericError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*errors.AppError); ok {
		return SendAppError(c, appErr)
	}

	// Convert generic error to AppError
	appErr := errors.Wrap(err, "An error occurred")
	return SendAppError(c, appErr)
}

// SendValidationErrorWithField sends a validation error response with field
func SendValidationErrorWithField(c *fiber.Ctx, message, field string) error {
	appErr := errors.NewValidationError(message, field)
	return SendAppError(c, appErr)
}

// SendNotFoundError sends a not found error response
func SendNotFoundError(c *fiber.Ctx, resource string) error {
	appErr := errors.NewNotFoundError(resource)
	return SendAppError(c, appErr)
}

// SendUnauthorizedError sends an unauthorized error response
func SendUnauthorizedError(c *fiber.Ctx, message string) error {
	appErr := errors.NewUnauthorizedError(message)
	return SendAppError(c, appErr)
}

// SendForbiddenError sends a forbidden error response
func SendForbiddenError(c *fiber.Ctx, message string) error {
	appErr := errors.NewForbiddenError(message)
	return SendAppError(c, appErr)
}
