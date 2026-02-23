package handler

import (
	"net/http"
	"strconv"

	"github.com/Project-DSView/backend/go/internal/api/types"
	"github.com/Project-DSView/backend/go/internal/application/services"
	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/pkg/config"
	"github.com/Project-DSView/backend/go/pkg/enrollment"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type AnnouncementHandler struct {
	announcementService *services.AnnouncementService
	enrollmentValidator *enrollment.EnrollmentValidator
}

func NewAnnouncementHandler(announcementService *services.AnnouncementService, enrollmentValidator *enrollment.EnrollmentValidator) *AnnouncementHandler {
	return &AnnouncementHandler{
		announcementService: announcementService,
		enrollmentValidator: enrollmentValidator,
	}
}

// Request models are now defined in internal/api/types/requests.go

// CreateAnnouncement creates a new announcement
// @Summary Create announcement
// @Description Create a new announcement for a course (teachers only)
// @Tags announcements
// @Accept json
// @Produce json
// @Param request body types.CreateAnnouncementRequest true "Announcement data"
// @Success 201 {object} response.StandardResponse{data=models.Announcement}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/announcements [post]
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *AnnouncementHandler) CreateAnnouncement(c *fiber.Ctx) error {
	var req types.CreateAnnouncementRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validate request
	if err := config.Validate.Struct(req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
	}

	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id").(string)

	// Create announcement
	announcement := &models.Announcement{
		MaterialBase: models.MaterialBase{
			CourseID:  req.CourseID,
			Title:     req.Title,
			CreatedBy: userID,
		},
		Content: req.Content,
	}

	if err := h.announcementService.CreateAnnouncement(announcement); err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create announcement", err.Error())
	}

	return response.SuccessResponse(c, http.StatusCreated, "Announcement created successfully", announcement.ToJSON())
}

// GetAnnouncements retrieves announcements for a course
// @Summary Get announcements
// @Description Get announcements for a specific course with optional filtering. Non-teachers can only see announcements from courses they are enrolled in.
// @Tags announcements
// @Produce json
// @Param course_id query string true "Course ID"
// @Param week query int false "Filter by week"
// @Param limit query int false "Limit results" default(20)
// @Param offset query int false "Offset results" default(0)
// @Success 200 {object} response.StandardResponse{data=[]models.Announcement}
// @Failure 400 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/announcements [get]
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *AnnouncementHandler) GetAnnouncements(c *fiber.Ctx) error {
	courseID := c.Query("course_id")
	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	// Validate course access for non-teachers
	if err := h.enrollmentValidator.ValidateCourseAccess(c, courseID); err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
	}

	// Parse optional parameters
	limit := 20
	offset := 0
	var week *int

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if weekStr := c.Query("week"); weekStr != "" {
		if parsed, err := strconv.Atoi(weekStr); err == nil && parsed >= 0 {
			week = &parsed
		}
	}

	announcements, total, err := h.announcementService.GetAnnouncementsByCourse(courseID, week, limit, offset)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get announcements", err.Error())
	}

	// Convert to JSON
	announcementsJSON := make([]map[string]interface{}, len(announcements))
	for i, announcement := range announcements {
		announcementsJSON[i] = announcement.ToJSON()
	}

	return response.SuccessResponse(c, http.StatusOK, "Announcements retrieved successfully", map[string]interface{}{
		"announcements": announcementsJSON,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetAnnouncement retrieves a specific announcement
// @Summary Get announcement
// @Description Get a specific announcement by ID. Non-teachers can only see announcements from courses they are enrolled in.
// @Tags announcements
// @Produce json
// @Param id path string true "Announcement ID"
// @Success 200 {object} response.StandardResponse{data=models.Announcement}
// @Failure 400 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/announcements/{id} [get]
// @Security BearerAuth
func (h *AnnouncementHandler) GetAnnouncement(c *fiber.Ctx) error {
	announcementID := c.Params("id")
	if announcementID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Announcement ID is required", nil)
	}

	announcement, err := h.announcementService.GetAnnouncementByID(announcementID)
	if err != nil {
		if err.Error() == "announcement not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Announcement not found", nil)
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get announcement", err.Error())
	}

	// Validate course access for non-teachers
	if err := h.enrollmentValidator.ValidateCourseAccess(c, announcement.CourseID); err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Announcement retrieved successfully", announcement.ToJSON())
}

// UpdateAnnouncement updates an announcement
// @Summary Update announcement
// @Description Update an existing announcement (creator only)
// @Tags announcements
// @Accept json
// @Produce json
// @Param id path string true "Announcement ID"
// @Param request body types.UpdateAnnouncementRequest true "Update data"
// @Success 200 {object} response.StandardResponse{data=models.Announcement}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/announcements/{id} [put]
// @Security BearerAuth
func (h *AnnouncementHandler) UpdateAnnouncement(c *fiber.Ctx) error {
	announcementID := c.Params("id")
	if announcementID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Announcement ID is required", nil)
	}

	var req types.UpdateAnnouncementRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validate request
	if err := config.Validate.Struct(req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	// Prepare updates
	updates := make(map[string]interface{})
	if req.Title != nil && *req.Title != "" {
		updates["title"] = *req.Title
	}
	if req.Content != nil && *req.Content != "" {
		updates["content"] = *req.Content
	}

	if err := h.announcementService.UpdateAnnouncement(announcementID, userID, updates); err != nil {
		if err.Error() == "announcement not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Announcement not found", nil)
		}
		if err.Error() == "only the creator can update this announcement" {
			return response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update announcement", err.Error())
	}

	// Get updated announcement
	announcement, err := h.announcementService.GetAnnouncementByID(announcementID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get updated announcement", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Announcement updated successfully", announcement.ToJSON())
}

// DeleteAnnouncement deletes an announcement
// @Summary Delete announcement
// @Description Delete an announcement (creator only)
// @Tags announcements
// @Produce json
// @Param id path string true "Announcement ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/announcements/{id} [delete]
// @Security BearerAuth
func (h *AnnouncementHandler) DeleteAnnouncement(c *fiber.Ctx) error {
	announcementID := c.Params("id")
	if announcementID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Announcement ID is required", nil)
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	if err := h.announcementService.DeleteAnnouncement(announcementID, userID); err != nil {
		if err.Error() == "announcement not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Announcement not found", nil)
		}
		if err.Error() == "only the creator can delete this announcement" {
			return response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete announcement", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Announcement deleted successfully", nil)
}

// GetAnnouncementStats retrieves announcement statistics
// @Summary Get announcement stats
// @Description Get statistics for announcements in a course. Non-teachers can only see stats from courses they are enrolled in.
// @Tags announcements
// @Produce json
// @Param course_id query string true "Course ID"
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 400 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/announcements/stats [get]
// @Security BearerAuth
func (h *AnnouncementHandler) GetAnnouncementStats(c *fiber.Ctx) error {
	courseID := c.Query("course_id")
	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	// Validate course access for non-teachers
	if err := h.enrollmentValidator.ValidateCourseAccess(c, courseID); err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
	}

	stats, err := h.announcementService.GetAnnouncementStats(courseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get announcement stats", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Announcement stats retrieved successfully", stats)
}

// GetRecentAnnouncements retrieves recent announcements for a user
// @Summary Get recent announcements
// @Description Get recent announcements from all enrolled courses
// @Tags announcements
// @Produce json
// @Param limit query int false "Limit results" default(10)
// @Success 200 {object} response.StandardResponse{data=[]models.Announcement}
// @Failure 401 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/announcements/recent [get]
// @Security BearerAuth
func (h *AnnouncementHandler) GetRecentAnnouncements(c *fiber.Ctx) error {
	// Parse limit parameter
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	announcements, err := h.announcementService.GetRecentAnnouncements(userID, limit)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get recent announcements", err.Error())
	}

	// Convert to JSON
	announcementsJSON := make([]map[string]interface{}, len(announcements))
	for i, announcement := range announcements {
		announcementsJSON[i] = announcement.ToJSON()
	}

	return response.SuccessResponse(c, http.StatusOK, "Recent announcements retrieved successfully", announcementsJSON)
}
