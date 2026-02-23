package handler

import (
	"github.com/Project-DSView/backend/go/internal/application/services"
	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type QueueHandler struct {
	queueService *services.QueueService
	userService  *services.UserService
}

func NewQueueHandler(queueService *services.QueueService, userService *services.UserService) *QueueHandler {
	return &QueueHandler{
		queueService: queueService,
		userService:  userService,
	}
}

// GetQueueJobs godoc
// @Summary Get queue jobs
// @Description Get queue jobs with filtering and pagination (Teachers and TAs can see jobs by course, students see only their own)
// @Tags queue
// @Security BearerAuth
// @Produce json
// @Param type query string false "Queue type filter" Enums(code_execution,review)
// @Param status query string false "Status filter" Enums(pending,processing,completed,failed,cancelled)
// @Param course_id query string false "Course ID filter (Teachers and TAs only)"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20)"
// @Success 200 {object} object{success=bool,data=object{jobs=[]object,pagination=object}} "Queue jobs retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/queue/jobs [get]
// @Security BearerAuth
func (h *QueueHandler) GetQueueJobs(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	// Parse query parameters
	queueType := c.Query("type")
	status := c.Query("status")
	courseID := c.Query("course_id")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Check permissions for course_id filter
	// Students can view queue jobs in courses they're enrolled in (read-only)
	// Teachers and TAs can view and manage queue jobs
	if courseID != "" {
		if !currentUser.IsTeacher {
			// Check if user is TA in the specified course
			isTA, err := h.queueService.IsUserTAInCourse(courseID, claims.UserID)
			if err != nil {
				return response.SendInternalError(c, "Failed to check TA status: "+err.Error())
			}
			// If not TA, check if user is enrolled in the course (students can view)
			if !isTA {
				// Check enrollment - students can view queue jobs in courses they're enrolled in
				var enrollmentCount int64
				if err := h.queueService.GetDB().Model(&models.Enrollment{}).
					Where("course_id = ? AND user_id = ?", courseID, claims.UserID).
					Count(&enrollmentCount).Error; err != nil {
					return response.SendInternalError(c, "Failed to check enrollment: "+err.Error())
				}
				if enrollmentCount == 0 {
					return response.SendError(c, fiber.StatusForbidden, "You are not enrolled in this course")
				}
			}
		}
	}

	// Get queue jobs
	jobs, total, err := h.queueService.GetQueueJobsWithDateFilter(queueType, status, courseID, claims.UserID, currentUser.IsTeacher, page, limit, fromDate, toDate)
	if err != nil {
		return response.SendInternalError(c, "Failed to get queue jobs: "+err.Error())
	}

	// Convert to response format
	jobData := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		jobData[i] = job.ToJSON()
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"jobs": jobData,
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": (total + limit - 1) / limit,
			},
		},
	})
}

// GetQueueJob godoc
// @Summary Get queue job by ID
// @Description Get detailed information about a specific queue job
// @Tags queue
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} object{success=bool,data=object} "Queue job retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Job not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/queue/jobs/{id} [get]
// @Security BearerAuth
func (h *QueueHandler) GetQueueJob(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	jobID := c.Params("id")
	if jobID == "" {
		return response.SendBadRequest(c, "Job ID is required")
	}

	// Get queue job
	job, err := h.queueService.GetQueueJobByID(jobID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get queue job: "+err.Error())
	}
	if job == nil {
		return response.SendNotFound(c, "Job not found")
	}

	// Check permissions - students can only see their own jobs
	if !currentUser.IsTeacher && job.UserID != claims.UserID {
		return response.SendError(c, fiber.StatusForbidden, "You can only view your own jobs")
	}

	return response.SendSuccess(c, "Queue job retrieved successfully", job.ToJSON())
}

// CancelQueueJob godoc
// @Summary Cancel a queue job
// @Description Cancel a pending queue job (students can cancel their own, teachers can cancel any)
// @Tags queue
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} object{success=bool,message=string} "Job cancelled successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Job not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/queue/jobs/{id}/cancel [post]
// @Security BearerAuth
func (h *QueueHandler) CancelQueueJob(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	jobID := c.Params("id")
	if jobID == "" {
		return response.SendBadRequest(c, "Job ID is required")
	}

	// Cancel job
	if err := h.queueService.CancelJob(jobID, claims.UserID, currentUser.IsTeacher); err != nil {
		if err.Error() == "job not found" {
			return response.SendNotFound(c, "Job not found")
		}
		if err.Error() == "permission denied" {
			return response.SendError(c, fiber.StatusForbidden, "You can only cancel your own jobs")
		}
		if err.Error() == "can only cancel pending jobs" {
			return response.SendBadRequest(c, "Can only cancel pending jobs")
		}
		return response.SendInternalError(c, "Failed to cancel job: "+err.Error())
	}

	return response.SendSuccess(c, "Job cancelled successfully", nil)
}

// GetQueueStats godoc
// @Summary Get queue statistics
// @Description Get statistics about the queue (Teachers and Admins only)
// @Tags queue
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{success=bool,data=object} "Queue statistics retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/queue/stats [get]
func (h *QueueHandler) GetQueueStats(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	// Check permissions - only teachers can view queue statistics
	if !currentUser.IsTeacher {
		return response.SendError(c, fiber.StatusForbidden, "Only teachers can view queue statistics")
	}

	// Get queue statistics
	stats, err := h.queueService.GetQueueStats()
	if err != nil {
		return response.SendInternalError(c, "Failed to get queue statistics: "+err.Error())
	}

	return response.SendSuccess(c, "Queue statistics retrieved successfully", stats)
}

// SubmitCodeExecution godoc
// @Summary Submit code for execution (DEPRECATED)
// @Description This endpoint is deprecated. Code execution is now handled automatically in course materials submission.
// @Tags queue
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 410 {object} map[string]string "Gone - Use course materials API instead"
// @Router /api/queue/execute [post]
func (h *QueueHandler) SubmitCodeExecution(c *fiber.Ctx) error {
	return response.SendError(c, fiber.StatusGone, "Code execution is now handled automatically in course materials submission. Please use POST /api/course-materials/{id}/submit instead.")
}

// SubmitCodeReview godoc
// @Summary Submit code for review
// @Description Submit code for review in the queue
// @Tags queue
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param review body object{exercise_id=string,course_id=string,review_notes=string} true "Code review data"
// @Success 200 {object} object{success=bool,message=string,data=object{job_id=string}} "Code submitted for review"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/queue/review [post]
func (h *QueueHandler) SubmitCodeReview(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	var req struct {
		ExerciseID  string `json:"exercise_id"`
		CourseID    string `json:"course_id"`
		ReviewNotes string `json:"review_notes"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}

	if req.ExerciseID == "" || req.CourseID == "" {
		return response.SendBadRequest(c, "Exercise ID and Course ID are required")
	}

	// Submit job
	job, err := h.queueService.SubmitCodeReviewJob(c.Context(), claims.UserID, req.ExerciseID, req.CourseID, req.ReviewNotes)
	if err != nil {
		return response.SendInternalError(c, "Failed to submit code review: "+err.Error())
	}

	return response.SendSuccess(c, "Code submitted for review", fiber.Map{
		"job_id": job.ID,
	})
}

// ProcessQueueJob godoc
// @Summary Process a queue job (Teacher/TA only)
// @Description Manually process a queue job (Teachers and TAs only)
// @Tags queue
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} object{success=bool,message=string} "Job processed successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Job not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/queue/jobs/{id}/process [post]
func (h *QueueHandler) ProcessQueueJob(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	// Check permissions - only teachers and TAs can process jobs
	if !currentUser.IsTeacher {
		return response.SendError(c, fiber.StatusForbidden, "Only teachers and TAs can process jobs")
	}

	jobID := c.Params("id")
	if jobID == "" {
		return response.SendBadRequest(c, "Job ID is required")
	}

	// Get job
	job, err := h.queueService.GetQueueJobByID(jobID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get job: "+err.Error())
	}
	if job == nil {
		return response.SendNotFound(c, "Job not found")
	}

	// Check if job can be processed
	if job.Status != enums.QueueStatusPending {
		return response.SendBadRequest(c, "Job is not in pending status")
	}

	// Update job status to processing
	if err := h.queueService.UpdateJobStatus(jobID, enums.QueueStatusProcessing, claims.UserID, nil, ""); err != nil {
		return response.SendInternalError(c, "Failed to update job status: "+err.Error())
	}

	// TODO: Implement actual processing logic here
	// For now, just simulate processing
	// In a real implementation, this would trigger the actual code execution or review

	return response.SendSuccess(c, "Job processing started", nil)
}

// ClaimQueueJob godoc
// @Summary Claim a queue job (TA/Teacher only)
// @Description TA or Teacher claims a pending queue job for review
// @Tags queue
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} object{success=bool,message=string,data=object{job_id=string,claimed_by=string,claimed_at=string}}
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Job not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/queue/jobs/{id}/claim [post]
func (h *QueueHandler) ClaimQueueJob(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	jobID := c.Params("id")
	if jobID == "" {
		return response.SendBadRequest(c, "Job ID is required")
	}

	// Get job to check course
	job, err := h.queueService.GetQueueJobByID(jobID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get job: "+err.Error())
	}
	if job == nil {
		return response.SendNotFound(c, "Job not found")
	}

	// Check permissions - only teachers and TAs can claim jobs
	if !currentUser.IsTeacher {
		// Check if user is TA in the course
		if job.CourseID == nil {
			return response.SendError(c, fiber.StatusForbidden, "Job does not have a course ID")
		}
		isTA, err := h.queueService.IsUserTAInCourse(*job.CourseID, claims.UserID)
		if err != nil {
			return response.SendInternalError(c, "Failed to check TA status: "+err.Error())
		}
		if !isTA {
			return response.SendError(c, fiber.StatusForbidden, "Only teachers and TAs can claim jobs")
		}
	}

	// Claim job
	job, err = h.queueService.ClaimJob(jobID, claims.UserID)
	if err != nil {
		if err.Error() == "job not found" {
			return response.SendNotFound(c, "Job not found")
		}
		if err.Error() == "job already claimed" {
			return response.SendBadRequest(c, "Job is already claimed by another TA")
		}
		if err.Error() == "job not pending" {
			return response.SendBadRequest(c, "Job is not in pending status")
		}
		return response.SendInternalError(c, "Failed to claim job: "+err.Error())
	}

	return response.SendSuccess(c, "Job claimed successfully", fiber.Map{
		"job_id":       job.ID,
		"claimed_by":   claims.UserID,
		"claimed_at":   job.ClaimedAt,
		"lab_room":     job.LabRoom,
		"table_number": job.TableNumber,
	})
}

// CompleteReview godoc
// @Summary Complete review (TA/Teacher only)
// @Description TA or Teacher completes review with approval/rejection decision
// @Tags queue
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param review body object{status=string,comment=string} true "Review decision"
// @Success 200 {object} object{success=bool,message=string,data=object{job_id=string,status=string,reviewed_by=string}}
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Job not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/queue/jobs/{id}/complete [post]
func (h *QueueHandler) CompleteReview(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	jobID := c.Params("id")
	if jobID == "" {
		return response.SendBadRequest(c, "Job ID is required")
	}

	// Get job to check course
	job, err := h.queueService.GetQueueJobByID(jobID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get job: "+err.Error())
	}
	if job == nil {
		return response.SendNotFound(c, "Job not found")
	}

	// Check permissions - only teachers and TAs can complete reviews
	if !currentUser.IsTeacher {
		// Check if user is TA in the course
		if job.CourseID == nil {
			return response.SendError(c, fiber.StatusForbidden, "Job does not have a course ID")
		}
		isTA, err := h.queueService.IsUserTAInCourse(*job.CourseID, claims.UserID)
		if err != nil {
			return response.SendInternalError(c, "Failed to check TA status: "+err.Error())
		}
		if !isTA {
			return response.SendError(c, fiber.StatusForbidden, "Only teachers and TAs can complete reviews")
		}
	}

	var req struct {
		Status  string `json:"status" validate:"required,oneof=approved rejected"`
		Comment string `json:"comment"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}

	if req.Status != "approved" && req.Status != "rejected" {
		return response.SendBadRequest(c, "Status must be 'approved' or 'rejected'")
	}

	// Complete review
	job, err = h.queueService.CompleteReview(jobID, claims.UserID, req.Status, req.Comment)
	if err != nil {
		if err.Error() == "job not found" {
			return response.SendNotFound(c, "Job not found")
		}
		if err.Error() == "job not claimed by user" {
			return response.SendBadRequest(c, "Job is not claimed by you")
		}
		if err.Error() == "job not in processing status" {
			return response.SendBadRequest(c, "Job is not in processing status")
		}
		return response.SendInternalError(c, "Failed to complete review: "+err.Error())
	}

	return response.SendSuccess(c, "Review completed successfully", fiber.Map{
		"job_id":       job.ID,
		"status":       req.Status,
		"reviewed_by":  claims.UserID,
		"comment":      req.Comment,
		"completed_at": job.CompletedAt,
	})
}

// RetryQueueJob godoc
// @Summary Retry a queue job
// @Description Students can retry a queue job if the original submission is more than 1 day old
// @Tags queue
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} object{success=bool,message=string,data=object{job_id=string,retry_job_id=string}}
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Job not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/queue/jobs/{id}/retry [post]
func (h *QueueHandler) RetryQueueJob(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	jobID := c.Params("id")
	if jobID == "" {
		return response.SendBadRequest(c, "Job ID is required")
	}

	// Create retry job
	retryJob, err := h.queueService.CreateRetryQueueJob(jobID, claims.UserID)
	if err != nil {
		if err.Error() == "queue job not found" {
			return response.SendNotFound(c, "Queue job not found")
		}
		if err.Error() == "queue job does not belong to user" {
			return response.SendError(c, fiber.StatusForbidden, "You can only retry your own queue jobs")
		}
		if err.Error() == "submission is less than 1 day old, cannot retry yet" || err.Error() == "queue job is less than 1 day old, cannot retry yet" {
			return response.SendBadRequest(c, "Cannot retry yet. Please wait at least 1 day from the original submission.")
		}
		return response.SendInternalError(c, "Failed to create retry job: "+err.Error())
	}

	return response.SendSuccess(c, "Queue job retry created successfully", fiber.Map{
		"job_id":       jobID,
		"retry_job_id": retryJob.ID,
		"status":       retryJob.Status,
		"created_at":   retryJob.CreatedAt,
	})
}
