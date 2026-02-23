package handler

import (
	"net/http"
	"strconv"

	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// CourseWeekHandler handles course week related requests
type CourseWeekHandler struct {
	courseWeekService services.CourseWeekService
}

// NewCourseWeekHandler creates a new course week handler
func NewCourseWeekHandler(courseWeekService services.CourseWeekService) *CourseWeekHandler {
	return &CourseWeekHandler{
		courseWeekService: courseWeekService,
	}
}

// CreateCourseWeekRequest represents the request body for creating a course week
type CreateCourseWeekRequest struct {
	WeekNumber  int    `json:"week_number" validate:"required,min=1,max=52"`
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description"`
}

// UpdateCourseWeekRequest represents the request body for updating a course week
type UpdateCourseWeekRequest struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description"`
}

// CreateCourseWeek creates a new course week
func (h *CourseWeekHandler) CreateCourseWeek(c *fiber.Ctx) error {
	courseID := c.Params("courseId")
	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	var req CreateCourseWeekRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
	}

	courseWeek, err := h.courseWeekService.CreateCourseWeek(
		courseID,
		req.WeekNumber,
		req.Title,
		req.Description,
		userID.(string),
	)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create course week", err.Error())
	}

	return response.SuccessResponse(c, http.StatusCreated, "Course week created successfully", courseWeek.ToJSON())
}

// GetCourseWeek retrieves a specific course week
func (h *CourseWeekHandler) GetCourseWeek(c *fiber.Ctx) error {
	courseID := c.Params("courseId")
	weekNumberStr := c.Params("weekNumber")

	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	weekNumber, err := strconv.Atoi(weekNumberStr)
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid week number", nil)
	}

	courseWeek, err := h.courseWeekService.GetCourseWeek(courseID, weekNumber)
	if err != nil {
		return response.ErrorResponse(c, http.StatusNotFound, "Course week not found", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Course week retrieved successfully", courseWeek.ToJSON())
}

// GetCourseWeeks retrieves all course weeks for a specific course
func (h *CourseWeekHandler) GetCourseWeeks(c *fiber.Ctx) error {
	courseID := c.Params("courseId")
	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	courseWeeks, err := h.courseWeekService.GetCourseWeeks(courseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get course weeks", err.Error())
	}

	// Convert to JSON format
	weeks := make([]map[string]interface{}, len(courseWeeks))
	for i, week := range courseWeeks {
		weeks[i] = week.ToJSON()
	}

	return response.SuccessResponse(c, http.StatusOK, "Course weeks retrieved successfully", map[string]interface{}{
		"weeks": weeks,
	})
}

// UpdateCourseWeek updates an existing course week
func (h *CourseWeekHandler) UpdateCourseWeek(c *fiber.Ctx) error {
	courseID := c.Params("courseId")
	weekNumberStr := c.Params("weekNumber")

	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	weekNumber, err := strconv.Atoi(weekNumberStr)
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid week number", nil)
	}

	var req UpdateCourseWeekRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}

	courseWeek, err := h.courseWeekService.UpdateCourseWeek(courseID, weekNumber, req.Title, req.Description)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update course week", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Course week updated successfully", courseWeek.ToJSON())
}

// DeleteCourseWeek deletes a course week
func (h *CourseWeekHandler) DeleteCourseWeek(c *fiber.Ctx) error {
	courseID := c.Params("courseId")
	weekNumberStr := c.Params("weekNumber")

	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	weekNumber, err := strconv.Atoi(weekNumberStr)
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid week number", nil)
	}

	if err := h.courseWeekService.DeleteCourseWeek(courseID, weekNumber); err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete course week", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Course week deleted successfully", nil)
}

// GetWeekTitle retrieves the title for a specific week
func (h *CourseWeekHandler) GetWeekTitle(c *fiber.Ctx) error {
	courseID := c.Params("courseId")
	weekNumberStr := c.Params("weekNumber")

	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	weekNumber, err := strconv.Atoi(weekNumberStr)
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid week number", nil)
	}

	title := h.courseWeekService.GetWeekTitle(courseID, weekNumber)

	return response.SuccessResponse(c, http.StatusOK, "Week title retrieved successfully", map[string]interface{}{
		"week_number": weekNumber,
		"title":       title,
	})
}
