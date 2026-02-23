package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/Project-DSView/backend/go/internal/api/types"
	"github.com/Project-DSView/backend/go/internal/application/services"
	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/Project-DSView/backend/go/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Request models are now defined in internal/api/types/requests.go

// TestToken godoc
// @Summary Generate test token for API testing
// @Description Generate a test JWT token for API testing
// @Tags testing
// @Produce json
// @Param is_teacher query bool false "Is teacher user" default(false)
// @Success 200 {object} object{success=bool,message=string,data=object{token=string,user=object{user_id=string,email=string,name=string,is_teacher=bool,user_type=string},expires_at=string}}
// @Router /test/token [get]
func TestToken(c *fiber.Ctx, jwtService *services.JWTService, cfg *config.Config) error {
	// Only allow in non-production
	if cfg.Server.Environment == "production" {
		return response.SendError(c, fiber.StatusForbidden, "Test endpoints not available in production")
	}

	// Parse is_teacher parameter
	isTeacherParam := c.Query("is_teacher", "false")
	isTeacher := isTeacherParam == "true"

	// Generate test user data
	testUserID := uuid.New().String()

	userType := "student"
	if isTeacher {
		userType = "teacher"
	}

	testEmail := fmt.Sprintf("test-%s@example.com", userType)
	testName := fmt.Sprintf("Test %s", strings.ToTitle(userType))

	// Generate JWT token
	token, err := jwtService.GenerateToken(testUserID, testEmail, testName, isTeacher)
	if err != nil {
		return response.SendInternalError(c, "Failed to generate test token: "+err.Error())
	}

	return response.SendSuccess(c, "Test token generated successfully", fiber.Map{
		"token": token,
		"user": fiber.Map{
			"user_id":    testUserID,
			"email":      testEmail,
			"name":       testName,
			"is_teacher": isTeacher,
			"user_type":  userType,
		},
		"expires_at": time.Now().Add(cfg.JWT.ExpiresIn).Format(time.RFC3339),
		"usage": fiber.Map{
			"bearer_auth":  "Authorization: Bearer " + token,
			"curl_example": "curl -H \"Authorization: Bearer " + token + "\" " + cfg.Server.Host + ":" + cfg.Server.Port + "/api/profile",
		},
	})
}

// TestTokenPost godoc
// @Summary Generate test token via POST for API testing
// @Description Generate a test JWT token for API testing with custom parameters (in-memory only)
// @Tags testing
// @Accept json
// @Produce json
// @Param request body types.TestTokenRequest true "Token generation parameters"
// @Success 200 {object} object{success=bool,message=string,data=object{token=string,user=object{user_id=string,email=string,firstname=string,lastname=string,name=string,is_teacher=bool,user_type=string},expires_at=string}}
// @Failure 400 {object} object{error=string}
// @Failure 403 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /test/token [post]
func TestTokenPost(c *fiber.Ctx, jwtService *services.JWTService, cfg *config.Config) error {
	// Only allow in non-production
	if cfg.Server.Environment == "production" {
		return response.SendError(c, fiber.StatusForbidden, "Test endpoints not available in production")
	}

	var req types.TestTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid JSON format: "+err.Error())
	}

	// Generate test user data
	testUserID := uuid.New().String()

	testEmail := req.Email
	if testEmail == "" {
		userType := "student"
		if req.IsTeacher {
			userType = "teacher"
		}
		testEmail = fmt.Sprintf("test-%s-%s@example.com", userType, testUserID[:8])
	}

	// Parse name into first and last name
	var firstName, lastName string
	if req.Name != "" {
		nameParts := strings.Fields(req.Name)
		if len(nameParts) > 0 {
			firstName = nameParts[0]
		}
		if len(nameParts) > 1 {
			lastName = strings.Join(nameParts[1:], " ")
		} else {
			if req.IsTeacher {
				lastName = "Teacher"
			} else {
				lastName = "Student"
			}
		}
	} else {
		firstName = "Test"
		if req.IsTeacher {
			lastName = "Teacher"
		} else {
			lastName = "Student"
		}
	}

	// Parse custom duration if provided
	var tokenExpiry time.Time
	if req.Duration != "" {
		duration, err := time.ParseDuration(req.Duration)
		if err != nil {
			return response.SendBadRequest(c, "Invalid duration format. Use formats like '1h', '30m', '24h'")
		}
		if duration > 168*time.Hour {
			return response.SendBadRequest(c, "Duration too long. Maximum allowed is 168h (1 week)")
		}
		if duration < 1*time.Minute {
			return response.SendBadRequest(c, "Duration too short. Minimum allowed is 1m (1 minute)")
		}
		tokenExpiry = time.Now().Add(duration)
	} else {
		tokenExpiry = time.Now().Add(cfg.JWT.ExpiresIn)
	}

	// Generate JWT token
	fullName := firstName + " " + lastName
	token, err := jwtService.GenerateToken(testUserID, testEmail, fullName, req.IsTeacher)
	if err != nil {
		return response.SendInternalError(c, "Failed to generate test token: "+err.Error())
	}

	userType := "student"
	if req.IsTeacher {
		userType = "teacher"
	}

	return response.SendSuccess(c, "Test token generated successfully (in-memory only)", fiber.Map{
		"token": token,
		"user": fiber.Map{
			"user_id":    testUserID,
			"email":      testEmail,
			"firstname":  firstName,
			"lastname":   lastName,
			"name":       fullName,
			"is_teacher": req.IsTeacher,
			"user_type":  userType,
		},
		"expires_at": tokenExpiry.Format(time.RFC3339),
		"usage": fiber.Map{
			"bearer_auth":  "Authorization: Bearer " + token,
			"curl_example": "curl -H \"Authorization: Bearer " + token + "\" " + cfg.Server.Host + ":" + cfg.Server.Port + "/api/profile",
		},
		"note": "This token is for testing only. Use /auth/google for real authentication.",
	})
}

// TestTokenPostWithDB creates a test token AND saves user to database
func TestTokenPostWithDB(c *fiber.Ctx, jwtService *services.JWTService, userService *services.UserService, cfg *config.Config) error {
	// Only allow in non-production
	if cfg.Server.Environment == "production" {
		return response.SendError(c, fiber.StatusForbidden, "Test endpoints not available in production")
	}

	var req types.TestTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid JSON format: "+err.Error())
	}

	// Generate test user data
	testUserID := uuid.New().String()

	testEmail := req.Email
	if testEmail == "" {
		userType := "student"
		if req.IsTeacher {
			userType = "teacher"
		}
		testEmail = fmt.Sprintf("test-%s-%s@example.com", userType, testUserID[:8])
	}

	// Parse name into first and last name
	var firstName, lastName string
	if req.Name != "" {
		nameParts := strings.Fields(req.Name)
		if len(nameParts) > 0 {
			firstName = nameParts[0]
		}
		if len(nameParts) > 1 {
			lastName = strings.Join(nameParts[1:], " ")
		} else {
			if req.IsTeacher {
				lastName = "Teacher"
			} else {
				lastName = "Student"
			}
		}
	} else {
		firstName = "Test"
		if req.IsTeacher {
			lastName = "Teacher"
		} else {
			lastName = "Student"
		}
	}

	// Check if user with this email already exists
	existingUser, err := userService.GetUserByEmail(testEmail)
	if err != nil {
		return response.SendInternalError(c, "Failed to check existing user: "+err.Error())
	}

	if existingUser != nil {
		return response.SendBadRequest(c, "User with this email already exists. Please use a different email.")
	}

	// Create user in database using new User model with IsTeacher
	testUser := &models.User{
		UserID:    testUserID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     testEmail,
		IsTeacher: req.IsTeacher,
	}

	// Create user directly via GORM
	if err := userService.CreateUserDirect(testUser); err != nil {
		return response.SendInternalError(c, "Failed to create test user in database: "+err.Error())
	}

	// Parse custom duration if provided
	var tokenExpiry time.Time
	if req.Duration != "" {
		duration, err := time.ParseDuration(req.Duration)
		if err != nil {
			// Clean up created user on token generation error
			userService.DeleteUser(testUserID)
			return response.SendBadRequest(c, "Invalid duration format. Use formats like '1h', '30m', '24h'")
		}
		// Reasonable limits for test tokens
		if duration > 168*time.Hour { // 1 week max
			userService.DeleteUser(testUserID)
			return response.SendBadRequest(c, "Duration too long. Maximum allowed is 168h (1 week)")
		}
		if duration < 1*time.Minute { // 1 minute min
			userService.DeleteUser(testUserID)
			return response.SendBadRequest(c, "Duration too short. Minimum allowed is 1m (1 minute)")
		}
		tokenExpiry = time.Now().Add(duration)
	} else {
		tokenExpiry = time.Now().Add(cfg.JWT.ExpiresIn)
	}

	// Generate JWT token
	fullName := firstName + " " + lastName
	token, err := jwtService.GenerateToken(testUserID, testEmail, fullName, req.IsTeacher)
	if err != nil {
		// If token generation fails, clean up the created user
		userService.DeleteUser(testUserID)
		return response.SendInternalError(c, "Failed to generate test token: "+err.Error())
	}

	userType := "student"
	if req.IsTeacher {
		userType = "teacher"
	}

	return response.SendSuccess(c, "Test token generated successfully and user created in database", fiber.Map{
		"token": token,
		"user": fiber.Map{
			"user_id":    testUserID,
			"email":      testEmail,
			"firstname":  firstName,
			"lastname":   lastName,
			"name":       fullName,
			"is_teacher": req.IsTeacher,
			"user_type":  userType,
		},
		"expires_at": tokenExpiry.Format(time.RFC3339),
		"usage": fiber.Map{
			"bearer_auth":  "Authorization: Bearer " + token,
			"curl_example": "curl -H \"Authorization: Bearer " + token + "\" " + cfg.Server.Host + ":" + cfg.Server.Port + "/api/profile",
		},
		"database": fiber.Map{
			"created":     true,
			"user_id":     testUserID,
			"email":       testEmail,
			"table_name":  "users",
			"message":     "User record created and verified in database",
			"debug_query": fmt.Sprintf("SELECT * FROM users WHERE user_id = '%s' OR email = '%s'", testUserID, testEmail),
		},
	})
}
