package handler

import (
	"net/http"
	"strconv"

	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type DeadlineCheckerHandler struct {
	deadlineService *services.DeadlineCheckerService
}

func NewDeadlineCheckerHandler(deadlineService *services.DeadlineCheckerService) *DeadlineCheckerHandler {
	return &DeadlineCheckerHandler{
		deadlineService: deadlineService,
	}
}

// GetAvailableMaterials gets materials available for submission
// @Summary Get available materials
// @Description Get materials that are available for submission (not past deadline)
// @Tags deadline-checker
// @Produce json
// @Param course_id query string false "Filter by course ID"
// @Success 200 {object} response.StandardResponse{data=[]models.CourseMaterial}
// @Failure 500 {object} response.StandardResponse
// @Router /api/materials/available [get]
// @Security BearerAuth
func (h *DeadlineCheckerHandler) GetAvailableMaterials(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	courseID := c.Query("course_id")

	materials, err := h.deadlineService.GetAvailableMaterialsForUser(userID, courseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get available materials", err.Error())
	}

	// Convert to JSON
	materialsJSON := make([]map[string]interface{}, len(materials))
	for i, material := range materials {
		materialsJSON[i] = material.ToJSON()
	}

	return response.SuccessResponse(c, http.StatusOK, "Available materials retrieved successfully", materialsJSON)
}

// GetExpiredMaterials gets materials that are past deadline
// @Summary Get expired materials
// @Description Get materials that are past their submission deadline
// @Tags deadline-checker
// @Produce json
// @Param course_id query string false "Filter by course ID"
// @Success 200 {object} response.StandardResponse{data=[]models.CourseMaterial}
// @Failure 500 {object} response.StandardResponse
// @Router /api/materials/expired [get]
// @Security BearerAuth
func (h *DeadlineCheckerHandler) GetExpiredMaterials(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	courseID := c.Query("course_id")

	materials, err := h.deadlineService.GetExpiredMaterialsForUser(userID, courseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get expired materials", err.Error())
	}

	// Convert to JSON
	materialsJSON := make([]map[string]interface{}, len(materials))
	for i, material := range materials {
		materialsJSON[i] = material.ToJSON()
	}

	return response.SuccessResponse(c, http.StatusOK, "Expired materials retrieved successfully", materialsJSON)
}

// CheckMaterialDeadline checks if a material is past its deadline
// @Summary Check material deadline
// @Description Check if a material is past its submission deadline
// @Tags deadline-checker
// @Produce json
// @Param material_id query string true "Material ID"
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/materials/check-deadline [get]
// @Security BearerAuth
func (h *DeadlineCheckerHandler) CheckMaterialDeadline(c *fiber.Ctx) error {
	materialID := c.Query("material_id")
	if materialID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Material ID is required", nil)
	}

	isExpired, err := h.deadlineService.CheckMaterialDeadline(materialID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to check deadline", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Deadline check completed", map[string]interface{}{
		"material_id": materialID,
		"is_expired":  isExpired,
	})
}

// CanSubmitExercise checks if a user can submit an exercise
// @Summary Check if can submit exercise
// @Description Check if a user can submit an exercise (not past deadline)
// @Tags deadline-checker
// @Produce json
// @Param exercise_id query string true "Exercise ID"
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/exercises/can-submit [get]
// @Security BearerAuth
func (h *DeadlineCheckerHandler) CanSubmitExercise(c *fiber.Ctx) error {
	exerciseID := c.Query("exercise_id")
	if exerciseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Exercise ID is required", nil)
	}

	userID := c.Locals("user_id").(string)

	canSubmit, message, err := h.deadlineService.CanSubmitExercise(userID, exerciseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to check submission eligibility", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Submission eligibility checked", map[string]interface{}{
		"exercise_id": exerciseID,
		"can_submit":  canSubmit,
		"message":     message,
	})
}

// GetMaterialsByDeadlineStatus gets materials grouped by deadline status
// @Summary Get materials by deadline status
// @Description Get materials grouped by available and expired status
// @Tags deadline-checker
// @Produce json
// @Param course_id query string false "Filter by course ID"
// @Success 200 {object} response.StandardResponse{data=map[string][]models.CourseMaterial}
// @Failure 500 {object} response.StandardResponse
// @Router /api/materials/by-deadline-status [get]
// @Security BearerAuth
func (h *DeadlineCheckerHandler) GetMaterialsByDeadlineStatus(c *fiber.Ctx) error {
	courseID := c.Query("course_id")

	materialsByStatus, err := h.deadlineService.GetMaterialsByDeadlineStatus(courseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get materials by deadline status", err.Error())
	}

	// Convert to JSON
	result := make(map[string][]map[string]interface{})
	for status, materials := range materialsByStatus {
		materialsJSON := make([]map[string]interface{}, len(materials))
		for i, material := range materials {
			materialsJSON[i] = material.ToJSON()
		}
		result[status] = materialsJSON
	}

	return response.SuccessResponse(c, http.StatusOK, "Materials by deadline status retrieved successfully", result)
}

// GetUpcomingDeadlines gets exercises with deadlines approaching
// @Summary Get upcoming deadlines
// @Description Get exercises with deadlines approaching within specified hours
// @Tags deadline-checker
// @Produce json
// @Param course_id query string false "Filter by course ID"
// @Param hours query int false "Hours ahead to check" default(24)
// @Success 200 {object} response.StandardResponse{data=[]models.CourseMaterial}
// @Failure 500 {object} response.StandardResponse
// @Router /api/exercises/upcoming-deadlines [get]
// @Security BearerAuth
func (h *DeadlineCheckerHandler) GetUpcomingDeadlines(c *fiber.Ctx) error {
	courseID := c.Query("course_id")
	hours := 24

	if hoursStr := c.Query("hours"); hoursStr != "" {
		if parsed, err := strconv.Atoi(hoursStr); err == nil && parsed > 0 {
			hours = parsed
		}
	}

	exercises, err := h.deadlineService.GetUpcomingDeadlines(courseID, hours)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get upcoming deadlines", err.Error())
	}

	// Convert to JSON
	exercisesJSON := make([]map[string]interface{}, len(exercises))
	for i, exercise := range exercises {
		exercisesJSON[i] = exercise.ToJSON()
	}

	return response.SuccessResponse(c, http.StatusOK, "Upcoming deadlines retrieved successfully", exercisesJSON)
}

// GetDeadlineStats gets statistics about deadlines
// @Summary Get deadline statistics
// @Description Get statistics about deadlines in a course
// @Tags deadline-checker
// @Produce json
// @Param course_id query string false "Filter by course ID"
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 500 {object} response.StandardResponse
// @Router /api/exercises/deadline-stats [get]
// @Security BearerAuth
func (h *DeadlineCheckerHandler) GetDeadlineStats(c *fiber.Ctx) error {
	courseID := c.Query("course_id")

	stats, err := h.deadlineService.GetDeadlineStats(courseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get deadline statistics", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Deadline statistics retrieved successfully", stats)
}
