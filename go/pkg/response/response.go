package response

import (
	"github.com/gofiber/fiber/v2"
)

// StandardResponse represents a standard API response structure
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// LoginResponse represents the response structure for login
type LoginResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	User    UserResponse `json:"user"`
	Token   string       `json:"token,omitempty"`
}

// AuthURLResponse represents the response structure for auth URL requests
type AuthURLResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

// SendSuccess sends a successful response
func SendSuccess(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendError sends an error response with a specific status code
func SendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(StandardResponse{
		Success: false,
		Message: message,
	})
}

// SendBadRequest sends a 400 Bad Request response
func SendBadRequest(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusBadRequest, message)
}

// SendUnauthorized sends a 401 Unauthorized response
func SendUnauthorized(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusUnauthorized, message)
}

// SendNotFound sends a 404 Not Found response
func SendNotFound(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusNotFound, message)
}

// SendInternalError sends a 500 Internal Server Error response
func SendInternalError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(StandardResponse{
		Success: false,
		Message: message,
	})
}

// SendValidationError sends a validation error response
func SendValidationError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(StandardResponse{
		Success: false,
		Message: "Validation error",
		Error:   message,
	})
}

// SendConflictError sends a conflict error response
func SendConflictError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusConflict).JSON(StandardResponse{
		Success: false,
		Message: "Conflict",
		Error:   message,
	})
}

// SendAuthURLResponse sends an auth URL response
func SendAuthURLResponse(c *fiber.Ctx, authURL, state string) error {
	return c.JSON(AuthURLResponse{
		AuthURL: authURL,
		State:   state,
	})
}

// SendLoginResponse sends a login response
func SendLoginResponse(c *fiber.Ctx, userResp UserResponse, token string) error {
	return c.JSON(LoginResponse{
		Success: true,
		Message: "Login successful",
		User:    userResp,
		Token:   token,
	})
}

// SendLogoutResponse sends a logout response
func SendLogoutResponse(c *fiber.Ctx) error {
	return SendSuccess(c, "Logged out successfully", nil)
}

// SendProfileResponse sends a user profile response
func SendProfileResponse(c *fiber.Ctx, userResp UserResponse) error {
	return c.JSON(userResp)
}

// SendTokenRefreshResponse sends a token refresh response
func SendTokenRefreshResponse(c *fiber.Ctx, userResp UserResponse) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Token refreshed successfully",
		"user":    userResp,
	})
}

// RedirectToFrontend redirects to frontend with query parameters
func RedirectToFrontend(c *fiber.Ctx, url string) error {
	if url == "" {
		return SendInternalError(c, "Frontend URL not configured")
	}
	return c.Redirect(url)
}

// ErrorResponse sends an error response with status code, message and error details
func ErrorResponse(c *fiber.Ctx, statusCode int, message string, errorDetails interface{}) error {
	return c.Status(statusCode).JSON(StandardResponse{
		Success: false,
		Message: message,
		Error:   errorDetails,
	})
}

// SuccessResponse sends a success response with status code, message and data
func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}
