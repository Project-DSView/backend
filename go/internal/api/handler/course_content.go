package handler

import (
	"net/http"
	"strconv"

	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// CourseContentHandler handles course content grouped by week
type CourseContentHandler struct {
	courseWeekService     services.CourseWeekService
	announcementService   services.AnnouncementService
	courseMaterialService services.CourseMaterialService
	// courseExerciseService services.CourseExerciseService // Comment out for now
}

// NewCourseContentHandler creates a new course content handler
func NewCourseContentHandler(
	courseWeekService services.CourseWeekService,
	announcementService services.AnnouncementService,
	courseMaterialService services.CourseMaterialService,
	// courseExerciseService services.CourseExerciseService, // Comment out for now
) *CourseContentHandler {
	return &CourseContentHandler{
		courseWeekService:     courseWeekService,
		announcementService:   announcementService,
		courseMaterialService: courseMaterialService,
		// courseExerciseService: courseExerciseService, // Comment out for now
	}
}

// WeekContent represents content for a specific week
type WeekContent struct {
	WeekNumber    int                      `json:"week_number"`
	Title         string                   `json:"title"`
	Announcements []map[string]interface{} `json:"announcements"`
	Exercises     []map[string]interface{} `json:"exercises"`
	Materials     []map[string]interface{} `json:"materials"`
}

// CourseContentResponse represents the response for course content grouped by week
type CourseContentResponse struct {
	PinnedAnnouncements []map[string]interface{} `json:"pinned_announcements"`
	Weeks               []WeekContent            `json:"weeks"`
}

// GetCourseContentByWeek retrieves all course content grouped by week
func (h *CourseContentHandler) GetCourseContentByWeek(c *fiber.Ctx) error {
	courseID := c.Params("courseId")
	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	// Get all course weeks
	courseWeeks, err := h.courseWeekService.GetCourseWeeks(courseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get course weeks", err.Error())
	}

	// Get pinned announcements (announcements without week)
	// Note: Since we removed week and pin from announcements, all announcements are now general
	pinnedAnnouncements := []map[string]interface{}{} // Empty for now

	// Get week content
	weeks := make([]WeekContent, len(courseWeeks))
	for i, courseWeek := range courseWeeks {
		weekContent := WeekContent{
			WeekNumber: courseWeek.WeekNumber,
			Title:      courseWeek.Title,
		}

		// Get announcements for this week
		// Note: Since we removed week from announcements, no week-specific announcements
		announcements := []map[string]interface{}{} // Empty for now
		weekContent.Announcements = announcements

		// Get materials for this week
		// Note: This method needs to be implemented in CourseMaterialService
		materials := []map[string]interface{}{} // Placeholder for now
		weekContent.Materials = materials

		// Get exercises for this week
		// Note: This method needs to be implemented in CourseExerciseService
		exercises := []map[string]interface{}{} // Placeholder for now
		weekContent.Exercises = exercises

		weeks[i] = weekContent
	}

	return response.SuccessResponse(c, http.StatusOK, "Course content retrieved successfully", CourseContentResponse{
		PinnedAnnouncements: pinnedAnnouncements,
		Weeks:               weeks,
	})
}

// GetWeekContent retrieves content for a specific week
func (h *CourseContentHandler) GetWeekContent(c *fiber.Ctx) error {
	courseID := c.Params("courseId")
	weekNumberStr := c.Params("weekNumber")

	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	weekNumber, err := strconv.Atoi(weekNumberStr)
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid week number", nil)
	}

	// Get course week info
	courseWeek, err := h.courseWeekService.GetCourseWeek(courseID, weekNumber)
	if err != nil {
		// If week doesn't exist, create it with default title
		courseWeek, err = h.courseWeekService.GetOrCreateCourseWeek(courseID, weekNumber, "system")
		if err != nil {
			return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get or create course week", err.Error())
		}
	}

	weekContent := WeekContent{
		WeekNumber: courseWeek.WeekNumber,
		Title:      courseWeek.Title,
	}

	// Get announcements for this week
	// Note: Since we removed week from announcements, no week-specific announcements
	announcements := []map[string]interface{}{} // Empty for now
	weekContent.Announcements = announcements

	// Get materials for this week
	// Note: This method needs to be implemented in CourseMaterialService
	materials := []map[string]interface{}{} // Placeholder for now
	weekContent.Materials = materials

	// Get exercises for this week
	// Note: This method needs to be implemented in CourseExerciseService
	exercises := []map[string]interface{}{} // Placeholder for now
	weekContent.Exercises = exercises

	return response.SuccessResponse(c, http.StatusOK, "Week content retrieved successfully", weekContent)
}
