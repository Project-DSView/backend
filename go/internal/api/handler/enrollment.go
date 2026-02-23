package handler

import (
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/auth"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type EnrollmentHandler struct {
	enrollmentService *services.EnrollmentService
	userService       *services.UserService
	courseService     *services.CourseService
}

func NewEnrollmentHandler(enrollmentService *services.EnrollmentService, userService *services.UserService, courseService *services.CourseService) *EnrollmentHandler {
	return &EnrollmentHandler{
		enrollmentService: enrollmentService,
		userService:       userService,
		courseService:     courseService,
	}
}

// EnrollInCourse godoc
// @Summary Enroll in course
// @Description Enroll in a course using enrollment key
// @Tags enrollments
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Course ID"
// @Param enrollment body object{enroll_key=string} true "Enrollment data with enrollment key"
// @Success 200 {object} object{success=bool,message=string,data=object} "Enrolled successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Course not found"
// @Failure 409 {object} map[string]string "Already enrolled"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id}/enroll [post]
func (h *EnrollmentHandler) EnrollInCourse(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	// Check permissions - all users can enroll
	if !auth.CanEnrollInCourse() {
		return response.SendError(c, fiber.StatusForbidden, "You don't have permission to enroll in courses")
	}

	courseID := c.Params("id")
	if courseID == "" {
		return response.SendBadRequest(c, "Course ID is required")
	}

	// Get course to check if user is the creator
	course, err := h.courseService.GetCourseByID(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get course: "+err.Error())
	}
	if course == nil {
		return response.SendNotFound(c, "Course not found")
	}

	// Parse request body
	var req struct {
		EnrollKey string `json:"enroll_key" validate:"required" example:"COURSE123"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}

	// If user is a teacher and is the creator, skip enrollment entirely
	if currentUser.IsTeacher && course.CreatedBy == claims.UserID {
		return response.SendSuccess(c, "Course creator has automatic access", fiber.Map{
			"course_id": courseID,
			"user_id":   claims.UserID,
			"role":      enums.EnrollmentRoleTeacher,
		})
	}

	if req.EnrollKey == "" {
		return response.SendBadRequest(c, "Enrollment key is required")
	}

	// Determine enrollment role
	// If user is a teacher but not the creator, they enroll as a student
	// All other users enroll as students
	enrollmentRole := enums.EnrollmentRoleStudent

	// Enroll user
	enrollment, err := h.enrollmentService.EnrollUser(courseID, claims.UserID, req.EnrollKey, enrollmentRole)
	if err != nil {
		if err.Error() == "already enrolled in this course" {
			return response.SendError(c, fiber.StatusConflict, err.Error())
		}
		if err.Error() == "invalid course or enrollment key" {
			return response.SendError(c, fiber.StatusBadRequest, "Invalid enrollment key")
		}
		return response.SendInternalError(c, "Failed to enroll: "+err.Error())
	}

	// เพิ่มข้อมูล user มาใส่ใน enrollment
	enrollment.UserInfo = currentUser

	enrollmentResp := response.ConvertToEnrollmentResponse(enrollment)

	return response.SendSuccess(c, "Enrolled successfully", enrollmentResp)
}

// GetCourseEnrollments godoc
// @Summary Get course enrollments
// @Description Get list of users enrolled in a course
// @Tags enrollments
// @Security BearerAuth
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} object{success=bool,data=object{enrollments=[]object}} "List of enrollments"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Course not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id}/enrollments [get]
func (h *EnrollmentHandler) GetCourseEnrollments(c *fiber.Ctx) error {
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

	// Check permissions - only teachers can view enrollments
	if !auth.CanViewEnrollments(currentUser.IsTeacher) {
		return response.SendError(c, fiber.StatusForbidden, "Only teachers can view enrollments")
	}

	// Get enrollments
	enrollments, err := h.enrollmentService.GetCourseEnrollments(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to fetch enrollments: "+err.Error())
	}

	// Convert to response format
	enrollmentData := make([]response.EnrollmentResponse, len(enrollments))
	for i, enrollment := range enrollments {
		enrollmentData[i] = response.ConvertToEnrollmentResponse(&enrollment)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"enrollments": enrollmentData,
		},
	})
}

// UpdateEnrollmentRole godoc
// @Summary Update enrollment role
// @Description Update a user's enrollment role in a course (Teachers only)
// @Tags enrollments
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param courseId path string true "Course ID"
// @Param userId path string true "User ID"
// @Param enrollment body object{role=string} true "New enrollment role"
// @Success 200 {object} object{success=bool,message=string,data=object} "Role updated successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Enrollment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{courseId}/enrollments/{userId}/role [put]
func (h *EnrollmentHandler) UpdateEnrollmentRole(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	// Check permissions - only teachers can manage enrollment roles
	if !auth.CanManageEnrollmentRoles(currentUser.IsTeacher) {
		return response.SendError(c, fiber.StatusForbidden, "Only teachers can manage enrollment roles")
	}

	courseID := c.Params("courseId")
	userID := c.Params("userId")

	if courseID == "" || userID == "" {
		return response.SendBadRequest(c, "Course ID and User ID are required")
	}

	// Parse request body
	var req struct {
		Role string `json:"role" validate:"required" example:"ta"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}

	// Validate role
	if req.Role != "student" && req.Role != "ta" {
		return response.SendBadRequest(c, "Role must be 'student' or 'ta'")
	}

	var newRole enums.EnrollmentRole
	if req.Role == "ta" {
		newRole = enums.EnrollmentRoleTA
	} else {
		newRole = enums.EnrollmentRoleStudent
	}

	// Update enrollment role
	if err := h.enrollmentService.UpdateEnrollmentRole(courseID, userID, newRole); err != nil {
		if err.Error() == "enrollment not found" {
			return response.SendNotFound(c, "Enrollment not found")
		}
		return response.SendInternalError(c, "Failed to update enrollment role: "+err.Error())
	}

	// Get updated enrollment
	enrollments, err := h.enrollmentService.GetCourseEnrollments(courseID)
	if err == nil {
		for _, enrollment := range enrollments {
			if enrollment.UserID == userID {
				enrollmentResp := response.ConvertToEnrollmentResponse(&enrollment)
				return response.SendSuccess(c, "Enrollment role updated successfully", enrollmentResp)
			}
		}
	}

	return response.SendSuccess(c, "Enrollment role updated successfully", fiber.Map{
		"course_id": courseID,
		"user_id":   userID,
		"new_role":  req.Role,
	})
}

// UnenrollFromCourse godoc
// @Summary Unenroll from course
// @Description Leave a course
// @Tags enrollments
// @Security BearerAuth
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} object{success=bool,message=string,data=interface{}} "Unenrolled successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Enrollment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id}/enroll [delete]
func (h *EnrollmentHandler) UnenrollFromCourse(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	courseID := c.Params("id")
	if courseID == "" {
		return response.SendBadRequest(c, "Course ID is required")
	}

	// Unenroll user
	if err := h.enrollmentService.UnenrollUser(courseID, claims.UserID); err != nil {
		if err.Error() == "enrollment not found" {
			return response.SendNotFound(c, "You are not enrolled in this course")
		}
		return response.SendInternalError(c, "Failed to unenroll: "+err.Error())
	}

	return response.SendSuccess(c, "Unenrolled successfully", nil)
}

// GetMyEnrollment godoc
// @Summary Get my enrollment status
// @Description Get current user's enrollment status in a course
// @Tags enrollments
// @Security BearerAuth
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} object{success=bool,data=object{enrollment=object}} "Enrollment status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not enrolled"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id}/my-enrollment [get]
func (h *EnrollmentHandler) GetMyEnrollment(c *fiber.Ctx) error {
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

	// Get course to check if user is the creator
	course, err := h.courseService.GetCourseByID(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get course: "+err.Error())
	}
	if course == nil {
		return response.SendNotFound(c, "Course not found")
	}

	// If user is a teacher and is the creator, return virtual enrollment
	if currentUser.IsTeacher && course.CreatedBy == claims.UserID {
		enrollmentResp := response.EnrollmentResponse{
			EnrollmentID: "", // No enrollment ID for creator
			CourseID:     courseID,
			UserID:       claims.UserID,
			Role:         string(enums.EnrollmentRoleTeacher),
			FirstName:    currentUser.FirstName,
			LastName:     currentUser.LastName,
			Email:        currentUser.Email,
			EnrolledAt:   course.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"enrollment": enrollmentResp,
			},
		})
	}

	// Get user's enrollment in this course
	enrollment, err := h.enrollmentService.GetUserEnrollmentInCourse(courseID, claims.UserID)
	if err != nil {
		return response.SendInternalError(c, "Failed to check enrollment: "+err.Error())
	}

	if enrollment == nil {
		return response.SendNotFound(c, "You are not enrolled in this course")
	}

	// Populate user info
	enrollment.UserInfo, _ = h.userService.GetUserByID(claims.UserID)

	enrollmentResp := response.ConvertToEnrollmentResponse(enrollment)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"enrollment": enrollmentResp,
		},
	})
}
