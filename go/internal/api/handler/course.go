package handler

import (
	"fmt"
	"time"

	"github.com/Project-DSView/backend/go/internal/application/services"
	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/auth"
	"github.com/Project-DSView/backend/go/pkg/errors"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"github.com/Project-DSView/backend/go/pkg/validation"
	"github.com/gofiber/fiber/v2"
)

type CourseHandler struct {
	courseService         *services.CourseService
	courseMaterialService *services.CourseMaterialService
	userService           *services.UserService
	enrollmentService     *services.EnrollmentService
	queueService          *services.QueueService
	storageService        storage.StorageService
}

func NewCourseHandler(courseService *services.CourseService, courseMaterialService *services.CourseMaterialService, userService *services.UserService, enrollmentService *services.EnrollmentService, queueService *services.QueueService, storageService storage.StorageService) *CourseHandler {
	return &CourseHandler{
		courseService:         courseService,
		courseMaterialService: courseMaterialService,
		userService:           userService,
		enrollmentService:     enrollmentService,
		queueService:          queueService,
		storageService:        storageService,
	}
}

// GetCourses godoc
// @Summary List courses
// @Description Get courses with pagination and filtering. Requires both API key and JWT authentication.
// @Tags courses
// @Security BearerAuth
// @Security ApiKeyAuth
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20)"
// @Param status query string false "Filter by status" Enums(active,archived)
// @Param search query string false "Search in name and description"
// @Success 200 {object} object{success=bool,data=object{courses=[]object,pagination=object}} "List of courses"
// @Failure 401 {object} map[string]string "Unauthorized - Both API key and JWT token required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses [get]
func (h *CourseHandler) GetCourses(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorizedError(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil {
		return response.SendGenericError(c, err)
	}
	if currentUser == nil {
		return response.SendNotFoundError(c, "User")
	}

	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	status := c.Query("status")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get courses using isTeacher instead of role
	courses, total, err := h.courseService.GetCoursesWithFilters(page, limit, status, search, claims.UserID, currentUser.IsTeacher)
	if err != nil {
		return response.SendGenericError(c, errors.Wrap(err, "Failed to fetch courses"))
	}

	// Convert to response format
	courseData := make([]response.CourseResponse, len(courses))
	for i, courseModel := range courses {
		// Teachers can see enroll key
		includeEnrollKey := currentUser.IsTeacher || courseModel.CreatedBy == claims.UserID
		courseData[i] = response.ConvertToCourseResponse(&courseModel, includeEnrollKey)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"courses": courseData,
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": (total + limit - 1) / limit,
			},
		},
	})
}

// CreateCourse godoc
// @Summary Create a new course
// @Description Create a new course with optional image upload (Teacher and Admin only). Requires both API key and JWT authentication.
// @Tags courses
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Course name"
// @Param description formData string true "Course description"
// @Param enroll_key formData string false "Enrollment key"
// @Param image formData file false "Course image (JPEG/PNG/WebP)"
// @Success 201 {object} object{success=bool,message=string,data=object} "Course created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized - Both API key and JWT token required"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses [post]
func (h *CourseHandler) CreateCourse(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	// Check permissions - only teachers can create courses
	if !currentUser.IsTeacher {
		return response.SendError(c, fiber.StatusForbidden, "Only teachers can create courses")
	}

	// Parse form data
	name := c.FormValue("name")
	description := c.FormValue("description")
	enrollKey := c.FormValue("enroll_key")

	// Validate required fields
	if name == "" || description == "" {
		return response.SendBadRequest(c, "Name and description are required")
	}

	// Validate input
	if err := validation.ValidateCourseCreation(name, description, enrollKey); err != nil {
		return response.SendValidationError(c, err.Error())
	}

	// Create course
	courseModel := models.Course{
		Name:        validation.SanitizeInput(name),
		Description: validation.SanitizeInput(description),
		CreatedBy:   claims.UserID,
	}

	if enrollKey != "" {
		courseModel.EnrollKey = enrollKey
	}

	if err := h.courseService.CreateCourse(&courseModel); err != nil {
		return response.SendInternalError(c, "Failed to create course: "+err.Error())
	}

	// Handle image upload if provided
	imageFile, err := c.FormFile("image")
	if err == nil && imageFile != nil {
		// Validate file type
		contentType := imageFile.Header.Get("Content-Type")
		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
			return response.SendBadRequest(c, "Image must be JPEG, PNG, or WebP format")
		}

		// Validate file size (10MB limit)
		const maxFileSize = 10 * 1024 * 1024
		if imageFile.Size > maxFileSize {
			return response.SendBadRequest(c, "Image file too large. Maximum size is 10MB")
		}

		// Open file
		src, err := imageFile.Open()
		if err != nil {
			return response.SendInternalError(c, "Failed to open image file: "+err.Error())
		}
		defer src.Close()

		// Upload image
		imageURL, err := h.storageService.UploadCourseImage(
			c.Context(),
			courseModel.CourseID,
			src,
			imageFile.Filename,
			contentType,
		)
		if err != nil {
			return response.SendValidationError(c, "Failed to upload image: "+err.Error())
		}

		// Update course with image URL
		updates := map[string]interface{}{
			"image_url": imageURL,
		}
		if err := h.courseService.UpdateCourse(courseModel.CourseID, updates); err != nil {
			return response.SendInternalError(c, "Failed to update course with image: "+err.Error())
		}
	}

	createdCourse, _ := h.courseService.GetCourseByID(courseModel.CourseID)
	courseResp := response.ConvertToCourseResponse(createdCourse, true)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Course created successfully",
		"data":    courseResp,
	})
}

// GetCourse godoc
// @Summary Get course by ID
// @Description Get detailed course information. Requires both API key and JWT authentication.
// @Tags courses
// @Security BearerAuth
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} object{success=bool,message=string,data=object} "Course details"
// @Failure 401 {object} map[string]string "Unauthorized - Both API key and JWT token required"
// @Failure 404 {object} map[string]string "Course not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id} [get]
func (h *CourseHandler) GetCourse(c *fiber.Ctx) error {
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

	courseModel, err := h.courseService.GetCourseByID(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to fetch course: "+err.Error())
	}

	if courseModel == nil {
		return response.SendNotFound(c, "Course not found")
	}

	// Check if user can see enroll key
	includeEnrollKey := auth.CanEditCourse(currentUser.IsTeacher, courseModel.CreatedBy, claims.UserID)
	courseResp := response.ConvertToCourseResponse(courseModel, includeEnrollKey)

	return response.SendSuccess(c, "Course retrieved successfully", courseResp)
}

// UpdateCourse godoc
// @Summary Update course
// @Description Update course information (Teacher/Admin only, teachers can only update their own). Requires both API key and JWT authentication.
// @Tags courses
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Course ID"
// @Param course body object{name=string,description=string,status=string} true "Course update data"
// @Success 200 {object} object{success=bool,message=string,data=object} "Updated course"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized - Both API key and JWT token required"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Course not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id} [put]
func (h *CourseHandler) UpdateCourse(c *fiber.Ctx) error {
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

	// Check permissions
	canModify, err := h.courseService.CanUserModifyCourse(claims.UserID, courseID, currentUser.IsTeacher)
	if err != nil {
		return response.SendInternalError(c, "Failed to check permissions: "+err.Error())
	}
	if !canModify {
		return response.SendError(c, fiber.StatusForbidden, "You don't have permission to modify this course")
	}

	// Parse request body
	var req struct {
		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
		Status      *string `json:"status,omitempty"`
		EnrollKey   *string `json:"enroll_key,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}

	// Validate input
	if err := validation.ValidateCourseUpdate(req.Name, req.Description, req.EnrollKey); err != nil {
		return response.SendValidationError(c, err.Error())
	}

	// Build updates map
	updates := response.BuildCourseUpdates(req.Name, req.Description, req.Status, req.EnrollKey)

	if len(updates) == 0 {
		return response.SendBadRequest(c, "No valid updates provided")
	}

	// Update course
	if err := h.courseService.UpdateCourse(courseID, updates); err != nil {
		return response.SendInternalError(c, "Failed to update course: "+err.Error())
	}

	// Get updated course
	updatedCourse, err := h.courseService.GetCourseByID(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to fetch updated course")
	}

	courseResp := response.ConvertToCourseResponse(updatedCourse, true) // Include enroll key for updater
	return response.SendSuccess(c, "Course updated successfully", courseResp)
}

// DeleteCourse godoc
// @Summary Delete course
// @Description Delete a course (Teacher/Admin only, teachers can only delete their own). Requires both API key and JWT authentication.
// @Tags courses
// @Security BearerAuth
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} object{success=bool,message=string,data=interface{}} "Course deleted successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized - Both API key and JWT token required"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Course not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id} [delete]
func (h *CourseHandler) DeleteCourse(c *fiber.Ctx) error {
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

	// Check permissions
	canModify, err := h.courseService.CanUserModifyCourse(claims.UserID, courseID, currentUser.IsTeacher)
	if err != nil {
		return response.SendInternalError(c, "Failed to check permissions: "+err.Error())
	}
	if !canModify {
		return response.SendError(c, fiber.StatusForbidden, "You don't have permission to delete this course")
	}

	// Delete course
	if err := h.courseService.DeleteCourse(courseID); err != nil {
		return response.SendInternalError(c, "Failed to delete course: "+err.Error())
	}

	return response.SendSuccess(c, "Course deleted successfully", nil)
}

// GetCourseReportForTeacher godoc
// @Summary Get course report for teacher
// @Description Get comprehensive course report including enrollment count, today's queue jobs, and materials count
// @Tags courses
// @Security BearerAuth
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} object{success=bool,message=string,data=object{enrollment_count=int,today_queue_jobs=[]object,exercises_created=int,current_materials_count=int}}
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Course not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id}/report/teacher [get]
func (h *CourseHandler) GetCourseReportForTeacher(c *fiber.Ctx) error {
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

	// Check if course exists
	courseModel, err := h.courseService.GetCourseByID(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to fetch course: "+err.Error())
	}
	if courseModel == nil {
		return response.SendNotFound(c, "Course not found")
	}

	// Check permissions - only course creator or teachers can access
	if !currentUser.IsTeacher && courseModel.CreatedBy != claims.UserID {
		return response.SendError(c, fiber.StatusForbidden, "Only course creator or teachers can access this report")
	}

	// Get enrollment count
	enrollmentCount, err := h.courseService.GetCourseEnrollmentCount(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get enrollment count: "+err.Error())
	}

	// Get materials count
	materialsCount, err := h.courseService.GetMaterialsCountByCourse(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get materials count: "+err.Error())
	}

	// Get today's queue jobs
	today := time.Now().Format("2006-01-02")
	todayStart := today + " 00:00:00"
	todayEnd := today + " 23:59:59"

	todayQueueJobs, err := h.queueService.GetQueueJobsByDateRange(courseID, todayStart, todayEnd)
	if err != nil {
		return response.SendInternalError(c, "Failed to get today's queue jobs: "+err.Error())
	}

	// Convert queue jobs to response format
	queueJobData := make([]map[string]interface{}, len(todayQueueJobs))
	for i, job := range todayQueueJobs {
		queueJobData[i] = job.ToJSON()
	}

	reportData := fiber.Map{
		"enrollment_count":        enrollmentCount,
		"today_queue_jobs":        queueJobData,
		"exercises_created":       materialsCount, // Same as current materials count
		"current_materials_count": materialsCount,
	}

	return response.SendSuccess(c, "Course report retrieved successfully", reportData)
}

// GetCourseReportForTA godoc
// @Summary Get course report for TA
// @Description Get course report for TA including today's queue jobs
// @Tags courses
// @Security BearerAuth
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} object{success=bool,message=string,data=object{today_queue_jobs=[]object}}
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Course not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id}/report/ta [get]
func (h *CourseHandler) GetCourseReportForTA(c *fiber.Ctx) error {
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

	// Check if course exists
	courseModel, err := h.courseService.GetCourseByID(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to fetch course: "+err.Error())
	}
	if courseModel == nil {
		return response.SendNotFound(c, "Course not found")
	}

	// Check permissions - only TAs enrolled in this course can access
	isEnrolled, enrollmentRole, err := h.courseService.IsUserEnrolledInCourse(courseID, claims.UserID)
	if err != nil {
		return response.SendInternalError(c, "Failed to check enrollment: "+err.Error())
	}
	if !isEnrolled || enrollmentRole != enums.EnrollmentRoleTA {
		return response.SendError(c, fiber.StatusForbidden, "Only TAs enrolled in this course can access this report")
	}

	// Get today's queue jobs
	today := time.Now().Format("2006-01-02")
	todayStart := today + " 00:00:00"
	todayEnd := today + " 23:59:59"

	todayQueueJobs, err := h.queueService.GetQueueJobsByDateRange(courseID, todayStart, todayEnd)
	if err != nil {
		return response.SendInternalError(c, "Failed to get today's queue jobs: "+err.Error())
	}

	// Convert queue jobs to response format
	queueJobData := make([]map[string]interface{}, len(todayQueueJobs))
	for i, job := range todayQueueJobs {
		queueJobData[i] = job.ToJSON()
	}

	reportData := fiber.Map{
		"today_queue_jobs": queueJobData,
	}

	return response.SendSuccess(c, "Course report retrieved successfully", reportData)
}

// GetCourseExercises godoc
// @Summary Get exercises in course
// @Description Get exercises in a specific course with role-based filtering
// @Tags courses
// @Security BearerAuth
// @Produce json
// @Param id path string true "Course ID" example("c0653a00-1382-4170-ac3a-c0924fdf5ec2")
// @Param page query int false "Page number" minimum(1) default(1) example(1)
// @Param limit query int false "Items per page" minimum(1) maximum(100) default(20) example(20)
// @Param status query string false "Filter by status (Teachers/Admins only)" Enums(draft,published,archived) example("published")
// @Success 200 {object} object{success=bool,data=object{exercises=[]object,pagination=object,course_info=object,user_permissions=object}} "Course exercises"
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 403 {object} object{success=bool,error=string} "Forbidden - not enrolled or insufficient permissions"
// @Failure 404 {object} object{success=bool,error=string} "Course not found"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/courses/{id}/exercises [get]
func (h *CourseHandler) GetCourseExercises(c *fiber.Ctx) error {
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

	// Check if course exists
	courseModel, err := h.courseService.GetCourseByID(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to fetch course: "+err.Error())
	}
	if courseModel == nil {
		return response.SendNotFound(c, "Course not found")
	}

	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	statusFilter := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Check access permissions and get filtered materials
	materials, total, userPermissions, err := h.getCourseMaterialsWithPermissions(
		courseID, claims.UserID, currentUser.IsTeacher, page, limit, statusFilter)
	if err != nil {
		if err.Error() == "access_denied" {
			return response.SendError(c, fiber.StatusForbidden, "You don't have access to view materials in this course. Please enroll first.")
		}
		return response.SendInternalError(c, "Failed to fetch course materials: "+err.Error())
	}

	// Convert materials to JSON
	materialData := make([]map[string]interface{}, len(materials))
	for i, material := range materials {
		materialData[i] = material.CourseMaterial.ToJSON()
	}

	// Prepare course info
	courseInfo := map[string]interface{}{
		"course_id":   courseModel.CourseID,
		"name":        courseModel.Name,
		"description": courseModel.Description,
		"status":      courseModel.Status,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"materials": materialData,
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": (total + limit - 1) / limit,
			},
			"course_info":      courseInfo,
			"user_permissions": userPermissions,
		},
	})
}

// Helper method for getting materials with permissions
func (h *CourseHandler) getCourseMaterialsWithPermissions(courseID, userID string, isTeacher bool, page, limit int, statusFilter string) ([]services.CourseMaterialWithWeek, int, map[string]interface{}, error) {
	// Check user permissions
	canViewDrafts := isTeacher

	// For students, check enrollment
	var isEnrolled bool
	var enrollmentRole string

	if !isTeacher {
		enrolled, err := h.enrollmentService.IsUserEnrolled(courseID, userID)
		if err != nil {
			return nil, 0, nil, fmt.Errorf("failed to check enrollment: %w", err)
		}
		if !enrolled {
			return nil, 0, nil, fmt.Errorf("access_denied")
		}
		isEnrolled = true

		// Get enrollment role for additional permissions
		enrollments, err := h.enrollmentService.GetCourseEnrollments(courseID)
		if err == nil {
			for _, enrollment := range enrollments {
				if enrollment.UserID == userID {
					enrollmentRole = string(enrollment.Role)
					// TAs might have some additional permissions
					if enrollment.Role == enums.EnrollmentRoleTA {
						// TAs can see more but still can't see drafts
					}
					break
				}
			}
		}
	} else {
		// Teachers always have access
		isEnrolled = true
		enrollmentRole = "instructor"
	}

	// Build status filter based on permissions
	var allowedStatuses []string
	if canViewDrafts {
		// Teachers can see all statuses
		if statusFilter != "" && enums.IsValidExerciseStatus(statusFilter) {
			allowedStatuses = []string{statusFilter}
		} else {
			allowedStatuses = []string{
				string(enums.ExerciseStatusDraft),
				string(enums.ExerciseStatusPublished),
				string(enums.ExerciseStatusArchived),
			}
		}
	} else {
		// Students can only see published
		allowedStatuses = []string{string(enums.ExerciseStatusPublished)}
		if statusFilter != "" && statusFilter != string(enums.ExerciseStatusPublished) {
			// Invalid status filter for students
			allowedStatuses = []string{} // Return empty results
		}
	}

	// Get course materials with filtering
	materials, total, err := h.courseService.GetCourseMaterialsWithFilters(courseID, allowedStatuses, page, limit)
	if err != nil {
		return nil, 0, nil, err
	}

	// Prepare user permissions info
	userPermissions := map[string]interface{}{
		"is_enrolled":       isEnrolled,
		"enrollment_role":   enrollmentRole,
		"can_view_drafts":   canViewDrafts,
		"can_view_archived": canViewDrafts,
		"is_teacher":        isTeacher,
		"allowed_statuses":  allowedStatuses,
		"user_type": func() string {
			if isTeacher {
				return "teacher"
			}
			return "student"
		}(),
	}

	return materials, total, userPermissions, nil
}

// UploadCourseImage godoc
// @Summary Upload course image
// @Description Upload an image for a course (Teacher/Admin only)
// @Tags courses
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Course ID"
// @Param image formData file true "Course image (JPEG/PNG/WebP)"
// @Success 200 {object} object{success=bool,message=string,data=object{image_url=string}} "Image uploaded successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Course not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id}/image [post]
func (h *CourseHandler) UploadCourseImage(c *fiber.Ctx) error {
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

	// Check permissions
	canModify, err := h.courseService.CanUserModifyCourse(claims.UserID, courseID, currentUser.IsTeacher)
	if err != nil {
		return response.SendInternalError(c, "Failed to check permissions: "+err.Error())
	}
	if !canModify {
		return response.SendError(c, fiber.StatusForbidden, "You don't have permission to modify this course")
	}

	// Get course to check if it exists
	courseModel, err := h.courseService.GetCourseByID(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to fetch course: "+err.Error())
	}
	if courseModel == nil {
		return response.SendNotFound(c, "Course not found")
	}

	// Get uploaded file
	file, err := c.FormFile("image")
	if err != nil {
		return response.SendBadRequest(c, "No image file uploaded")
	}

	// Validate file size
	const maxFileSize = 5 * 1024 * 1024 // 5MB
	if file.Size > maxFileSize {
		return response.SendValidationError(c, "File size too large. Maximum size is 5MB")
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if !h.isAllowedImageType(contentType) {
		return response.SendValidationError(c, "Invalid file type. Only JPEG, PNG, and WebP are allowed")
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return response.SendInternalError(c, "Failed to open file")
	}
	defer src.Close()

	// Delete old image if exists
	if courseModel.ImageURL != "" {
		_ = h.storageService.DeleteFile(c.Context(), courseModel.ImageURL)
	}

	// Upload to storage
	imageURL, err := h.storageService.UploadCourseImage(
		c.Context(),
		courseID,
		src,
		file.Filename,
		contentType,
	)
	if err != nil {
		return response.SendValidationError(c, "Failed to upload image: "+err.Error())
	}

	// Update course with new image URL
	updates := map[string]interface{}{
		"image_url": imageURL,
	}
	if err := h.courseService.UpdateCourse(courseID, updates); err != nil {
		return response.SendInternalError(c, "Failed to update course: "+err.Error())
	}

	return response.SendSuccess(c, "Course image uploaded successfully", fiber.Map{
		"image_url": imageURL,
	})
}

// DeleteCourseImage godoc
// @Summary Delete course image
// @Description Delete the image of a course (Teacher/Admin only)
// @Tags courses
// @Security BearerAuth
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} object{success=bool,message=string} "Image deleted successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Course not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/courses/{id}/image [delete]
func (h *CourseHandler) DeleteCourseImage(c *fiber.Ctx) error {
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

	// Check permissions
	canModify, err := h.courseService.CanUserModifyCourse(claims.UserID, courseID, currentUser.IsTeacher)
	if err != nil {
		return response.SendInternalError(c, "Failed to check permissions: "+err.Error())
	}
	if !canModify {
		return response.SendError(c, fiber.StatusForbidden, "You don't have permission to modify this course")
	}

	// Get course
	courseModel, err := h.courseService.GetCourseByID(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to fetch course: "+err.Error())
	}
	if courseModel == nil {
		return response.SendNotFound(c, "Course not found")
	}

	// Delete image from storage if exists
	if courseModel.ImageURL != "" {
		if err := h.storageService.DeleteFile(c.Context(), courseModel.ImageURL); err != nil {
			// Log error but don't fail the operation
			// log.Printf("Failed to delete image from storage: %v", err)
		}
	}

	// Update course to remove image URL
	updates := map[string]interface{}{
		"image_url": "",
	}
	if err := h.courseService.UpdateCourse(courseID, updates); err != nil {
		return response.SendInternalError(c, "Failed to update course: "+err.Error())
	}

	return response.SendSuccess(c, "Course image deleted successfully", nil)
}

// Helper method to validate image types
func (h *CourseHandler) isAllowedImageType(contentType string) bool {
	allowedTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}
