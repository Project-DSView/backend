package handlers

import (
	"strconv"
	"time"

	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// ParsePaginationParams parses pagination parameters from query string
func ParsePaginationParams(c *fiber.Ctx) (page, limit int) {
	page = c.QueryInt("page", 1)
	limit = c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return page, limit
}

// ParseIntParam parses an integer parameter from URL path
func ParseIntParam(c *fiber.Ctx, paramName string) (int, error) {
	paramStr := c.Params(paramName)
	if paramStr == "" {
		return 0, fiber.NewError(fiber.StatusBadRequest, paramName+" is required")
	}

	value, err := strconv.Atoi(paramStr)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "Invalid "+paramName+": must be a number")
	}

	return value, nil
}

// ParseStringParam parses a string parameter from URL path
func ParseStringParam(c *fiber.Ctx, paramName string) (string, error) {
	param := c.Params(paramName)
	if param == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, paramName+" is required")
	}
	return param, nil
}

// ParseOptionalStringParam parses an optional string parameter from URL path
func ParseOptionalStringParam(c *fiber.Ctx, paramName string) string {
	return c.Params(paramName)
}

// ParseQueryParam parses a query parameter
func ParseQueryParam(c *fiber.Ctx, paramName string) string {
	return c.Query(paramName)
}

// ParseOptionalIntParam parses an optional integer parameter from URL path
func ParseOptionalIntParam(c *fiber.Ctx, paramName string) (int, error) {
	paramStr := c.Params(paramName)
	if paramStr == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(paramStr)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "Invalid "+paramName+": must be a number")
	}

	return value, nil
}

// ParseTimeParam parses a time parameter in RFC3339 format
func ParseTimeParam(timeStr string, fieldName string) (*time.Time, error) {
	if timeStr == "" {
		return nil, nil
	}

	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid "+fieldName+" format. Use RFC3339 format (e.g., 2025-12-31T23:59:59Z)")
	}

	return &parsedTime, nil
}

// ValidateRequestBody validates and parses request body into target struct
func ValidateRequestBody(c *fiber.Ctx, target interface{}) error {
	if err := c.BodyParser(target); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}
	return nil
}

// ValidateRequiredField checks if a required field is not empty
func ValidateRequiredField(value, fieldName string) error {
	if value == "" {
		return fiber.NewError(fiber.StatusBadRequest, fieldName+" is required")
	}
	return nil
}

// ValidatePositiveInt validates that an integer is positive
func ValidatePositiveInt(value int, fieldName string) error {
	if value <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, fieldName+" must be positive")
	}
	return nil
}

// ValidateIntRange validates that an integer is within a range
func ValidateIntRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return fiber.NewError(fiber.StatusBadRequest, fieldName+" must be between "+strconv.Itoa(min)+" and "+strconv.Itoa(max))
	}
	return nil
}

// ValidateStringLength validates that a string is within length limits
func ValidateStringLength(value string, min, max int, fieldName string) error {
	length := len(value)
	if length < min || length > max {
		return fiber.NewError(fiber.StatusBadRequest, fieldName+" must be between "+strconv.Itoa(min)+" and "+strconv.Itoa(max)+" characters")
	}
	return nil
}

// ValidateOneOf validates that a value is one of the allowed values
func ValidateOneOf(value string, allowed []string, fieldName string) error {
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return nil
		}
	}
	return fiber.NewError(fiber.StatusBadRequest, fieldName+" must be one of: "+joinStrings(allowed, ", "))
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
