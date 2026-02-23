package handlers

import (
	"github.com/Project-DSView/backend/go/internal/application/services"
	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// GetClaimsFromContext extracts and validates claims from Fiber context
func GetClaimsFromContext(c *fiber.Ctx) (*types.Claims, error) {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid authentication")
	}
	return claims, nil
}

// GetCurrentUser retrieves the current user from claims and validates it exists
func GetCurrentUser(c *fiber.Ctx, userService *services.UserService) (*models.User, error) {
	claims, err := GetClaimsFromContext(c)
	if err != nil {
		return nil, err
	}

	user, err := userService.GetUserByID(claims.UserID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to get user: "+err.Error())
	}
	if user == nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	return user, nil
}

// RequireTeacher checks if the current user is a teacher, returns error if not
func RequireTeacher(c *fiber.Ctx, userService *services.UserService) (*models.User, error) {
	user, err := GetCurrentUser(c, userService)
	if err != nil {
		return nil, err
	}

	if !user.IsTeacher {
		return nil, fiber.NewError(fiber.StatusForbidden, "Only teachers can perform this action")
	}

	return user, nil
}

// RequireTeacherOrSelf checks if the current user is a teacher or the target user themselves
func RequireTeacherOrSelf(c *fiber.Ctx, userService *services.UserService, targetUserID string) (*models.User, error) {
	user, err := GetCurrentUser(c, userService)
	if err != nil {
		return nil, err
	}

	if !user.IsTeacher && user.UserID != targetUserID {
		return nil, fiber.NewError(fiber.StatusForbidden, "You can only access your own resources")
	}

	return user, nil
}

// GetClaimsAndUser is a convenience function that gets both claims and user
func GetClaimsAndUser(c *fiber.Ctx, userService *services.UserService) (*types.Claims, *models.User, error) {
	claims, err := GetClaimsFromContext(c)
	if err != nil {
		return nil, nil, err
	}

	user, err := GetCurrentUser(c, userService)
	if err != nil {
		return nil, nil, err
	}

	return claims, user, nil
}

// HandleAuthError handles authentication errors with appropriate response
func HandleAuthError(c *fiber.Ctx, err error) error {
	if fiberErr, ok := err.(*fiber.Error); ok {
		switch fiberErr.Code {
		case fiber.StatusUnauthorized:
			return response.SendUnauthorized(c, fiberErr.Message)
		case fiber.StatusForbidden:
			return response.SendError(c, fiber.StatusForbidden, fiberErr.Message)
		case fiber.StatusNotFound:
			return response.SendNotFound(c, fiberErr.Message)
		default:
			return response.SendInternalError(c, fiberErr.Message)
		}
	}
	return response.SendInternalError(c, err.Error())
}
