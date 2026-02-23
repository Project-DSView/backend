package handler

import (
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type InvitationHandler struct {
	invitationService *services.InvitationService
	userService       *services.UserService
}

func NewInvitationHandler(invitationService *services.InvitationService, userService *services.UserService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		userService:       userService,
	}
}

// CreateInvitation godoc
// @Summary Create invitation link
// @Description Create a new invitation link for a course (Teacher who is creator only)
// @Tags invitations
// @Security BearerAuth
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} object{success=bool,message=string,data=object{invitation_id=string,token=string,invitation_url=string,expires_at=string}} "Invitation created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Course not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id}/invitations [post]
func (h *InvitationHandler) CreateInvitation(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	courseID := c.Params("id")
	if courseID == "" {
		return response.SendBadRequest(c, "Course ID is required")
	}

	// Create invitation
	invitation, err := h.invitationService.CreateInvitation(courseID, claims.UserID)
	if err != nil {
		if err.Error() == "course not found" {
			return response.SendNotFound(c, "Course not found")
		}
		if err.Error() == "only course creator can create invitations" {
			return response.SendError(c, fiber.StatusForbidden, "Only course creator can create invitations")
		}
		return response.SendInternalError(c, "Failed to create invitation: "+err.Error())
	}

	// Build invitation URL (assuming frontend base URL)
	invitationURL := "/course/invite/" + invitation.Token

	return response.SendSuccess(c, "Invitation created successfully", fiber.Map{
		"invitation_id":  invitation.InvitationID,
		"token":          invitation.Token,
		"invitation_url":  invitationURL,
		"expires_at":     invitation.ExpiresAt,
		"created_at":     invitation.CreatedAt,
	})
}

// GetCourseInvitations godoc
// @Summary Get course invitations
// @Description Get all invitation links for a course (Teacher who is creator only)
// @Tags invitations
// @Security BearerAuth
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} object{success=bool,data=[]object{invitation_id=string,token=string,invitation_url=string,expires_at=string,created_at=string}} "List of invitations"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Course not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id}/invitations [get]
func (h *InvitationHandler) GetCourseInvitations(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	courseID := c.Params("id")
	if courseID == "" {
		return response.SendBadRequest(c, "Course ID is required")
	}

	// Get invitations
	invitations, err := h.invitationService.GetCourseInvitations(courseID, claims.UserID)
	if err != nil {
		if err.Error() == "course not found" {
			return response.SendNotFound(c, "Course not found")
		}
		if err.Error() == "only course creator can view invitations" {
			return response.SendError(c, fiber.StatusForbidden, "Only course creator can view invitations")
		}
		return response.SendInternalError(c, "Failed to get invitations: "+err.Error())
	}

	// Convert to response format
	invitationData := make([]fiber.Map, len(invitations))
	for i, invitation := range invitations {
		invitationURL := "/course/invite/" + invitation.Token
		invitationData[i] = fiber.Map{
			"invitation_id": invitation.InvitationID,
			"token":         invitation.Token,
			"invitation_url": invitationURL,
			"expires_at":    invitation.ExpiresAt,
			"created_at":    invitation.CreatedAt,
			"is_expired":    invitation.IsExpired(),
		}
	}

	return response.SendSuccess(c, "Invitations retrieved successfully", invitationData)
}

// EnrollViaInvitation godoc
// @Summary Enroll via invitation link
// @Description Auto-enroll in a course using an invitation token (no enrollment key required)
// @Tags invitations
// @Security BearerAuth
// @Security ApiKeyAuth
// @Produce json
// @Param token path string true "Invitation token"
// @Success 200 {object} object{success=bool,message=string,data=object{enrollment_id=string,course_id=string,user_id=string,role=string}} "Enrolled successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Invitation not found or expired"
// @Failure 409 {object} map[string]string "Already enrolled"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/invite/{token} [post]
func (h *InvitationHandler) EnrollViaInvitation(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	token := c.Params("token")
	if token == "" {
		return response.SendBadRequest(c, "Invitation token is required")
	}

	// Enroll via invitation
	enrollment, err := h.invitationService.EnrollViaInvitation(token, claims.UserID)
	if err != nil {
		if err.Error() == "invitation not found" {
			return response.SendNotFound(c, "Invitation not found")
		}
		if err.Error() == "invitation has expired" {
			return response.SendError(c, fiber.StatusBadRequest, "Invitation has expired")
		}
		if err.Error() == "already enrolled in this course" {
			return response.SendError(c, fiber.StatusConflict, "Already enrolled in this course")
		}
		return response.SendInternalError(c, "Failed to enroll: "+err.Error())
	}

	enrollmentResp := response.ConvertToEnrollmentResponse(enrollment)

	return response.SendSuccess(c, "Enrolled successfully", enrollmentResp)
}

