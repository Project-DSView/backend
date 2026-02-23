package handler

import (
	"github.com/Project-DSView/backend/go/internal/application/services"
	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SubmissionHandler struct {
	submissionService *services.SubmissionService
	userService       *services.UserService
	enrollmentService *services.EnrollmentService
	db                *gorm.DB
}

func NewSubmissionHandler(
	subSvc *services.SubmissionService,
	userSvc *services.UserService,
	enrollSvc *services.EnrollmentService,
	db *gorm.DB,
) *SubmissionHandler {
	return &SubmissionHandler{
		submissionService: subSvc,
		userService:       userSvc,
		enrollmentService: enrollSvc,
		db:                db,
	}
}

// SubmitExercise godoc
// @Summary Submit exercise (DEPRECATED)
// @Description This endpoint is deprecated. Please use POST /api/course-materials/{id}/submit instead.
// @Tags submissions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Exercise ID"
// @Success 410 {object} object{success=bool,error=string} "Gone - Use course materials API instead"
// @Router /api/exercises/{id}/submit [post]
func (h *SubmissionHandler) SubmitExercise(c *fiber.Ctx) error {
	return response.SendError(c, fiber.StatusGone, "Exercise submissions are deprecated. Please use course materials API instead.")
}

// ListExerciseSubmissions godoc
// @Summary List exercise submissions (DEPRECATED)
// @Description This endpoint is deprecated. Please use course materials API instead.
// @Tags submissions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Exercise ID"
// @Success 410 {object} object{success=bool,error=string} "Gone - Use course materials API instead"
// @Router /api/exercises/{id}/submissions [get]
func (h *SubmissionHandler) ListExerciseSubmissions(c *fiber.Ctx) error {
	return response.SendError(c, fiber.StatusGone, "Exercise submissions are deprecated. Please use course materials API instead.")
}

// GetSubmission godoc
// @Summary Get submission detail
// @Description ดูรายละเอียดการส่งงาน (เจ้าของ, Teachers, หรือ TAs ที่ enroll ในคอร์สเท่านั้น)
// @Tags submissions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Submission ID"
// @Success 200 {object} object{success=bool,message=string,data=object{submissionId=string,userId=string,exerciseId=string,code=string,passedCount=int,failedCount=int,totalScore=int,status=string,errorMessage=string,submittedAt=string,results=[]object{resultId=string,testCaseId=string,status=string,actualOutput=object,errorMessage=string}}}
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 403 {object} object{success=bool,error=string} "Forbidden"
// @Failure 404 {object} object{success=bool,error=string} "Submission not found"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/submissions/{id} [get]
func (h *SubmissionHandler) GetSubmission(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	subID := c.Params("id")
	if subID == "" {
		return response.SendBadRequest(c, "Submission ID is required")
	}

	sub, err := h.submissionService.GetSubmissionByID(subID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get submission: "+err.Error())
	}
	if sub == nil {
		return response.SendNotFound(c, "Submission not found")
	}

	curUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get user: "+err.Error())
	}

	// Allow owner to view their own submission
	if sub.UserID == claims.UserID {
		return h.sendSubmissionResponse(c, sub)
	}

	// For others, check if they can view submissions for this material
	if sub.MaterialID != "" {
		canView, err := h.canViewMaterialSubmissions(claims.UserID, sub.MaterialID, curUser.IsTeacher)
		if err != nil {
			return response.SendInternalError(c, "Failed to check permissions: "+err.Error())
		}
		if !canView {
			return response.SendError(c, fiber.StatusForbidden, "You don't have permission to view this submission")
		}
	} else {
		// Legacy exercise submission - deny access (exercises are deprecated)
		return response.SendError(c, fiber.StatusForbidden, "Exercise submissions are no longer accessible. Please use course materials.")
	}

	return h.sendSubmissionResponse(c, sub)
}

// Helper method สำหรับตรวจสอบสิทธิ์การดู material submissions
func (h *SubmissionHandler) canViewMaterialSubmissions(userID, materialID string, isTeacher bool) (bool, error) {
	// Teachers สามารถดู submissions ทั้งหมดได้
	if isTeacher {
		return true, nil
	}

	// สำหรับ non-teachers ต้องเช็คว่าเป็น TA ในคอร์สที่มี material นี้หรือไม่
	// ดึง course ที่มี material นี้
	var courseID string
	err := h.db.Model(&models.CourseMaterial{}).
		Select("course_id").
		Where("material_id = ?", materialID).
		Scan(&courseID).Error
	if err != nil {
		return false, err
	}

	// เช็คว่า user เป็น TA ในคอร์สนี้หรือไม่
	enrollment, err := h.enrollmentService.GetUserEnrollmentInCourse(courseID, userID)
	if err != nil {
		return false, err
	}
	if enrollment != nil && enrollment.Role == enums.EnrollmentRoleTA {
		return true, nil // เป็น TA ใน course ที่มี material นี้
	}

	return false, nil // ไม่ใช่ teacher และไม่ใช่ TA ในคอร์สที่เกี่ยวข้อง
}

// Helper method สำหรับส่ง submission response
func (h *SubmissionHandler) sendSubmissionResponse(c *fiber.Ctx, sub *models.Submission) error {
	responseData := fiber.Map{
		"submission_id":      sub.SubmissionID,
		"user_id":            sub.UserID,
		"material_id":        sub.MaterialID,
		"code":               sub.Code,
		"passed_count":       sub.PassedCount,
		"failed_count":       sub.FailedCount,
		"total_score":        sub.TotalScore,
		"status":             sub.Status,
		"error_message":      sub.ErrorMessage,
		"submitted_at":       sub.SubmittedAt,
		"results":            sub.Results,
		"file_url":           sub.FileURL,
		"file_name":          sub.FileName,
		"file_size":          sub.FileSize,
		"mime_type":          sub.MimeType,
		"is_late_submission": sub.IsLateSubmission,
		"feedback":           sub.Feedback,
		"feedback_file_url":  sub.FeedbackFileURL,
		"graded_at":          sub.GradedAt,
		"graded_by":          sub.GradedBy,
		"review_status":      sub.ReviewStatus,
		"queue_job_id":       sub.QueueJobID,
	}

	// Get graded_by_user information if graded_by exists
	if sub.GradedBy != "" {
		var gradedByUser models.User
		if err := h.db.First(&gradedByUser, "user_id = ?", sub.GradedBy).Error; err == nil {
			responseData["graded_by_user"] = fiber.Map{
				"user_id":   gradedByUser.UserID,
				"firstname": gradedByUser.FirstName,
				"lastname":  gradedByUser.LastName,
				"email":     gradedByUser.Email,
			}
		}
	}

	// Get queue job status and processed_by_user if queue_job_id exists
	if sub.QueueJobID != nil && *sub.QueueJobID != "" {
		var queueJob models.QueueJob
		if err := h.db.First(&queueJob, "id = ?", *sub.QueueJobID).Error; err == nil {
			responseData["queue_status"] = string(queueJob.Status)

			// Load processed_by_user if exists
			if queueJob.ProcessedBy != nil && *queueJob.ProcessedBy != "" {
				var processedByUser models.User
				if err := h.db.First(&processedByUser, "user_id = ?", *queueJob.ProcessedBy).Error; err == nil {
					responseData["queue_processed_by_user"] = fiber.Map{
						"user_id":   processedByUser.UserID,
						"firstname": processedByUser.FirstName,
						"lastname":  processedByUser.LastName,
						"email":     processedByUser.Email,
					}
				}
			}

			// Calculate queue position (number of pending jobs before this one)
			// Only calculate if status is pending
			if queueJob.Status == enums.QueueStatusPending {
				var queuePosition int64
				// Count pending jobs with same material_id and course_id that were created before this job
				positionQuery := h.db.Model(&models.QueueJob{}).
					Where("status = ?", enums.QueueStatusPending).
					Where("created_at < ?", queueJob.CreatedAt)

				// Filter by same material_id if available
				if queueJob.MaterialID != nil && *queueJob.MaterialID != "" {
					positionQuery = positionQuery.Where("material_id = ?", *queueJob.MaterialID)
				} else if queueJob.CourseID != nil && *queueJob.CourseID != "" {
					// Fallback to course_id if material_id is not available
					positionQuery = positionQuery.Where("course_id = ?", *queueJob.CourseID)
				}

				if err := positionQuery.Count(&queuePosition).Error; err == nil {
					// Position is 1-based (1 = first in queue)
					responseData["queue_position"] = int(queuePosition) + 1
				}
			}
		}
	}

	return response.SendSuccess(c, "Submission retrieved successfully", responseData)
}

// SubmitMaterialExercise godoc
// @Summary Submit material exercise
// @Description Submit code for a material-based exercise (new unified system)
// @Tags submissions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Material ID"
// @Param request body object{code=string} true "Code submission"
// @Success 201 {object} object{success=bool,message=string,data=object{submissionId=string,passedCount=int,failedCount=int,totalScore=int,results=[]object{resultId=string,testCaseId=string,status=string,actualOutput=object,errorMessage=string}}}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 401 {object} object{success=bool,error=string}
// @Failure 500 {object} object{success=bool,error=string}
// @Router /api/course-materials/{id}/submit [post]
func (h *SubmissionHandler) SubmitMaterialExercise(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	materialID := c.Params("id")
	if materialID == "" {
		return response.SendBadRequest(c, "Material ID is required")
	}

	var req struct {
		Code string `json:"code" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}

	if req.Code == "" {
		return response.SendBadRequest(c, "Code is required")
	}

	// Submit material exercise
	result, err := h.submissionService.SubmitMaterialExercise(claims.UserID, materialID, req.Code)
	if err != nil {
		return response.SendInternalError(c, "Failed to submit material exercise: "+err.Error())
	}

	// Extract submission data from result
	submission := result.Submission.(*models.Submission)

	return response.SendSuccess(c, "Material exercise submitted successfully", fiber.Map{
		"submission_id": submission.SubmissionID,
		"status":        submission.Status,
		"passed_count":  submission.PassedCount,
		"failed_count":  submission.FailedCount,
		"total_score":   submission.TotalScore,
		"results":       result.Results,
	})
}

// GetMyMaterialSubmission godoc
// @Summary Get my material submission
// @Description Get the current user's submission for a specific material
// @Tags submissions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Material ID"
// @Success 200 {object} object{success=bool,message=string,data=object{submission_id=string,user_id=string,material_id=string,code=string,passed_count=int,failed_count=int,total_score=int,status=string,error_message=string,submitted_at=string,results=[]object{result_id=string,test_case_id=string,status=string,actual_output=object,error_message=string}}}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 401 {object} object{success=bool,error=string}
// @Failure 500 {object} object{success=bool,error=string}
// @Router /api/course-materials/{id}/submissions/me [get]
func (h *SubmissionHandler) GetMyMaterialSubmission(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	materialID := c.Params("id")
	if materialID == "" {
		return response.SendBadRequest(c, "Material ID is required")
	}

	// Get user's latest submission for this material
	var sub models.Submission
	err := h.db.Preload("Results").
		Where("user_id = ? AND material_id = ?", claims.UserID, materialID).
		Order("submitted_at DESC").
		First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.SendSuccess(c, "No submission found", nil)
		}
		return response.SendInternalError(c, "Failed to get submission: "+err.Error())
	}

	return h.sendSubmissionResponse(c, &sub)
}

// SubmitPDFExercise godoc
// @Summary Submit PDF exercise
// @Description Submit a PDF file for a PDF exercise material
// @Tags submissions
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Material ID"
// @Param file formData file true "PDF file to upload"
// @Success 200 {object} object{success=bool,message=string,data=object{submission_id=string,file_url=string,file_name=string,file_size=int64,status=string,submitted_at=string}}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 401 {object} object{success=bool,error=string}
// @Failure 413 {object} object{success=bool,error=string}
// @Failure 415 {object} object{success=bool,error=string}
// @Failure 500 {object} object{success=bool,error=string}
// @Router /api/course-materials/{id}/submit-pdf [post]
func (h *SubmissionHandler) SubmitPDFExercise(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	materialID := c.Params("id")
	if materialID == "" {
		return response.SendBadRequest(c, "Material ID is required")
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return response.SendBadRequest(c, "File is required: "+err.Error())
	}

	// Validate file type
	if file.Header.Get("Content-Type") != "application/pdf" {
		return response.SendError(c, fiber.StatusUnsupportedMediaType, "Only PDF files are allowed")
	}

	// Validate file size (10MB limit)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if file.Size > maxFileSize {
		return response.SendError(c, fiber.StatusRequestEntityTooLarge, "File size exceeds 10MB limit")
	}

	// Submit PDF exercise
	submission, err := h.submissionService.SubmitPDFExercise(claims.UserID, materialID, file)
	if err != nil {
		return response.SendInternalError(c, "Failed to submit PDF exercise: "+err.Error())
	}

	return response.SendSuccess(c, "PDF exercise submitted successfully", fiber.Map{
		"submission_id": submission.SubmissionID,
		"file_url":      submission.FileURL,
		"file_name":     submission.FileName,
		"file_size":     submission.FileSize,
		"status":        submission.Status,
		"submitted_at":  submission.SubmittedAt,
	})
}
