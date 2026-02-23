package handler

import (
	"net/http"

	"github.com/Project-DSView/backend/go/internal/api/types"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/pkg/enrollment"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// CourseScoreHexagonalHandler handles course score requests using hexagonal architecture
type CourseScoreHandler struct {
	courseScoreService  *services.CourseScoreService
	enrollmentValidator *enrollment.EnrollmentValidator
}

// NewCourseScoreHexagonalHandler creates a new course score handler
func NewCourseScoreHandler(courseScoreService *services.CourseScoreService, enrollmentValidator *enrollment.EnrollmentValidator) *CourseScoreHandler {
	return &CourseScoreHandler{
		courseScoreService:  courseScoreService,
		enrollmentValidator: enrollmentValidator,
	}
}

// GetStudentCourseScore gets total score for a student in a course
// @Summary Get student course score
// @Description Get total score for a student in a specific course
// @Tags course-scores
// @Produce json
// @Param course_id query string true "Course ID"
// @Success 200 {object} response.StandardResponse{data=object}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-scores/course [get]
// @Security BearerAuth
func (h *CourseScoreHandler) GetStudentCourseScore(c *fiber.Ctx) error {
	courseID := c.Query("course_id")
	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	// Validate course access for non-teachers
	if err := h.enrollmentValidator.ValidateCourseAccess(c, courseID); err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
	}

	userID := c.Locals("user_id").(string)

	courseScore, err := h.courseScoreService.GetStudentCourseScore(c.Context(), userID, courseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get course score", err.Error())
	}

	if courseScore == nil {
		return response.SuccessResponse(c, http.StatusOK, "No course score found", nil)
	}

	return response.SuccessResponse(c, http.StatusOK, "Course score retrieved successfully", courseScore)
}

// UpdateCourseScore updates the total score for a student in a course
// @Summary Update course score
// @Description Update the total score for a student in a specific course
// @Tags course-scores
// @Produce json
// @Param course_id query string true "Course ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-scores/update [post]
// @Security BearerAuth
func (h *CourseScoreHandler) UpdateCourseScore(c *fiber.Ctx) error {
	courseID := c.Query("course_id")
	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	userID := c.Locals("user_id").(string)

	err := h.courseScoreService.UpdateCourseScore(c.Context(), userID, courseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update course score", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Course score updated successfully", nil)
}

// BatchUpdateCourseScores updates course scores for multiple students
// @Summary Batch update course scores
// @Description Update course scores for multiple students efficiently
// @Tags course-scores
// @Accept json
// @Produce json
// @Param request body types.CourseScoreBatchUpdateRequest true "Batch update request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-scores/batch-update [post]
// @Security BearerAuth
func (h *CourseScoreHandler) BatchUpdateCourseScores(c *fiber.Ctx) error {
	var req types.CourseScoreBatchUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}

	if req.CourseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	if len(req.UserIDs) == 0 {
		return response.ErrorResponse(c, http.StatusBadRequest, "User IDs are required", nil)
	}

	err := h.courseScoreService.BatchUpdateCourseScores(c.Context(), req.UserIDs, req.CourseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to batch update course scores", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Course scores updated successfully", nil)
}

// Request models are now defined in internal/api/types/requests.go
