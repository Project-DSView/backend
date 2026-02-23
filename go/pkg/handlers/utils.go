package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Common utility functions for handlers

// SanitizeString removes leading/trailing whitespace and normalizes string
func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

// IsEmptyString checks if a string is empty or contains only whitespace
func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

// FormatTime formats time to RFC3339 string
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseTime parses RFC3339 time string
func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}

// GetCurrentTimestamp returns current time in RFC3339 format
func GetCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// GetQueryParamWithDefault gets query parameter with default value
func GetQueryParamWithDefault(c *fiber.Ctx, key, defaultValue string) string {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetQueryIntWithDefault gets query integer parameter with default value
func GetQueryIntWithDefault(c *fiber.Ctx, key string, defaultValue int) int {
	value := c.QueryInt(key, defaultValue)
	return value
}

// GetQueryBoolWithDefault gets query boolean parameter with default value
func GetQueryBoolWithDefault(c *fiber.Ctx, key string, defaultValue bool) bool {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

// Contains checks if a slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// FilterEmptyStrings removes empty strings from a slice
func FilterEmptyStrings(slice []string) []string {
	result := []string{}

	for _, item := range slice {
		if !IsEmptyString(item) {
			result = append(result, item)
		}
	}

	return result
}

// GetClientIP gets the client IP address from the request
func GetClientIP(c *fiber.Ctx) string {
	// Try to get IP from X-Forwarded-For header first
	ip := c.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Try X-Real-IP header
	ip = c.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Fall back to remote IP
	return c.IP()
}

// GetUserAgent gets the user agent from the request
func GetUserAgent(c *fiber.Ctx) string {
	return c.Get("User-Agent")
}

// GetRequestID gets the request ID from context (if set by middleware)
func GetRequestID(c *fiber.Ctx) string {
	return c.Get("X-Request-ID")
}

// IsValidUUID checks if a string is a valid UUID format
func IsValidUUID(uuid string) bool {
	// Simple UUID format check (8-4-4-4-12 characters)
	parts := strings.Split(uuid, "-")
	if len(parts) != 5 {
		return false
	}

	expectedLengths := []int{8, 4, 4, 4, 12}
	for i, part := range parts {
		if len(part) != expectedLengths[i] {
			return false
		}
		// Check if all characters are hexadecimal
		for _, char := range part {
			if !((char >= '0' && char <= '9') ||
				(char >= 'a' && char <= 'f') ||
				(char >= 'A' && char <= 'F')) {
				return false
			}
		}
	}

	return true
}

// TruncateString truncates a string to specified length and adds ellipsis
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// GenerateSlug creates a URL-friendly slug from a string
func GenerateSlug(s string) string {
	// Convert to lowercase
	slug := strings.ToLower(s)

	// Replace spaces and special characters with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove multiple consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}

// ValidateEmailFormat performs basic email format validation
func ValidateEmailFormat(email string) bool {
	// Basic email format check
	if !strings.Contains(email, "@") {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	// Check local and domain parts are not empty
	if IsEmptyString(parts[0]) || IsEmptyString(parts[1]) {
		return false
	}

	// Check domain has at least one dot
	if !strings.Contains(parts[1], ".") {
		return false
	}

	return true
}

// GetContentType gets the content type from the request
func GetContentType(c *fiber.Ctx) string {
	return c.Get("Content-Type")
}

// IsJSONRequest checks if the request is JSON
func IsJSONRequest(c *fiber.Ctx) bool {
	contentType := GetContentType(c)
	return strings.Contains(contentType, "application/json")
}

// IsFormRequest checks if the request is form data
func IsFormRequest(c *fiber.Ctx) bool {
	contentType := GetContentType(c)
	return strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data")
}
