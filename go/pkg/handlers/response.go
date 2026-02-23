package handlers

import (
	"strconv"

	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// CreatePaginationInfo creates pagination info from page, limit, and total
func CreatePaginationInfo(page, limit, total int) PaginationInfo {
	totalPages := (total + limit - 1) / limit
	return PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// SendPaginatedResponse sends a paginated response with data and pagination info
func SendPaginatedResponse(c *fiber.Ctx, message string, data interface{}, pagination PaginationInfo) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": message,
		"data": fiber.Map{
			"items":      data,
			"pagination": pagination,
		},
	})
}

// SendListResponse sends a simple list response
func SendListResponse(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// SendCreatedResponse sends a 201 Created response
func SendCreatedResponse(c *fiber.Ctx, message string, data interface{}) error {
	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// SendUpdatedResponse sends a 200 OK response for updates
func SendUpdatedResponse(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// SendDeletedResponse sends a 200 OK response for deletions
func SendDeletedResponse(c *fiber.Ctx, message string) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": message,
	})
}

// ConvertToMap converts any struct to map[string]interface{}
func ConvertToMap(data interface{}) map[string]interface{} {
	if data == nil {
		return make(map[string]interface{})
	}

	// This is a simple implementation - in production you might want to use
	// reflection or a JSON marshal/unmarshal approach for more complex types
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}

	// For now, return empty map - this can be enhanced based on needs
	return make(map[string]interface{})
}

// BuildSuccessResponse builds a standard success response
func BuildSuccessResponse(message string, data interface{}) fiber.Map {
	response := fiber.Map{
		"success": true,
		"message": message,
	}

	if data != nil {
		response["data"] = data
	}

	return response
}

// BuildErrorResponse builds a standard error response
func BuildErrorResponse(message string, errorCode string) fiber.Map {
	response := fiber.Map{
		"success": false,
		"message": message,
	}

	if errorCode != "" {
		response["error"] = errorCode
	}

	return response
}

// HandleValidationError handles validation errors with field information
func HandleValidationError(c *fiber.Ctx, field, message string) error {
	return response.SendValidationErrorWithField(c, message, field)
}

// HandleNotFoundError handles not found errors
func HandleNotFoundError(c *fiber.Ctx, resource string) error {
	return response.SendNotFound(c, resource+" not found")
}

// HandleUnauthorizedError handles unauthorized errors
func HandleUnauthorizedError(c *fiber.Ctx, message string) error {
	return response.SendUnauthorized(c, message)
}

// HandleForbiddenError handles forbidden errors
func HandleForbiddenError(c *fiber.Ctx, message string) error {
	return response.SendError(c, fiber.StatusForbidden, message)
}

// HandleInternalError handles internal server errors
func HandleInternalError(c *fiber.Ctx, message string) error {
	return response.SendInternalError(c, message)
}

// HandleBadRequestError handles bad request errors
func HandleBadRequestError(c *fiber.Ctx, message string) error {
	return response.SendBadRequest(c, message)
}

// ParseIntFromString safely parses an integer from string
func ParseIntFromString(str string) (int, error) {
	if str == "" {
		return 0, nil
	}
	return strconv.Atoi(str)
}

// ParseFloatFromString safely parses a float from string
func ParseFloatFromString(str string) (float64, error) {
	if str == "" {
		return 0, nil
	}
	return strconv.ParseFloat(str, 64)
}

// SendUnauthorizedError sends a 401 Unauthorized response
func SendUnauthorizedError(c *fiber.Ctx, message string) error {
	return response.SendUnauthorized(c, message)
}

// SendGenericError sends a generic error response
func SendGenericError(c *fiber.Ctx, err error) error {
	return response.SendInternalError(c, err.Error())
}

// SendNotFoundError sends a 404 Not Found response
func SendNotFoundError(c *fiber.Ctx, resource string) error {
	return response.SendNotFound(c, resource+" not found")
}

// SendForbiddenError sends a 403 Forbidden response
func SendForbiddenError(c *fiber.Ctx, message string) error {
	return response.SendError(c, fiber.StatusForbidden, message)
}

// SendValidationErrorWithField sends a validation error with field information
func SendValidationErrorWithField(c *fiber.Ctx, message, field string) error {
	return response.SendValidationErrorWithField(c, message, field)
}
