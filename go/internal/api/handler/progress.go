package handler

import (
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type ProgressHandler struct {
	progressService   *services.ProgressService
	userService       *services.UserService
	enrollmentService *services.EnrollmentService
	courseService     *services.CourseService
}

func NewProgressHandler(
	progressSvc *services.ProgressService,
	userSvc *services.UserService,
	enrollmentSvc *services.EnrollmentService,
	courseSvc *services.CourseService,
) *ProgressHandler {
	return &ProgressHandler{
		progressService:   progressSvc,
		userService:       userSvc,
		enrollmentService: enrollmentSvc,
		courseService:     courseSvc,
	}
}

// GetSelfProgress godoc
// @Summary Get own progress
// @Description ดูความก้าวหน้าของตนเอง โดยกรองตาม courseId ได้
// @Tags progress
// @Security BearerAuth
// @Produce json
// @Param courseId query string false "Filter by Course ID"
// @Success 200 {object} object{success=bool,message=string,data=object{progress=[]object{progressId=string,userId=string,exerciseId=string,exerciseTitle=string,status=string,score=int,seatNumber=string,lastSubmittedAt=string}}}
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/students/progress [get]
func (h *ProgressHandler) GetSelfProgress(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}
	courseID := c.Query("courseId")
	items, err := h.progressService.GetSelfProgress(claims.UserID, courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to fetch progress: "+err.Error())
	}
	return response.SendSuccess(c, "Progress retrieved successfully", fiber.Map{
		"progress": items,
	})
}

// GetCourseProgress godoc
// @Summary Get course progress (Teachers/TAs only)
// @Description ดูความก้าวหน้าของนักเรียนในรายวิชา
// @Tags progress
// @Security BearerAuth
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} object{success=bool,message=string,data=object{courseProgress=[]object{userId=string,studentName=string,completedExercises=int,totalExercises=int,averageScore=number,lastActivity=string}}}
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 403 {object} object{success=bool,error=string} "Forbidden"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/courses/{id}/progress [get]
func (h *ProgressHandler) GetCourseProgress(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	curUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || curUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	courseID := c.Params("id")
	if courseID == "" {
		return response.SendBadRequest(c, "Course ID is required")
	}

	// Check permissions: must be teacher or TA in this course
	canView, err := h.canViewCourseProgress(claims.UserID, courseID, curUser.IsTeacher)
	if err != nil {
		return response.SendInternalError(c, "Failed to check permissions: "+err.Error())
	}
	if !canView {
		return response.SendError(c, fiber.StatusForbidden, "Only teachers or TAs can view course progress")
	}

	items, err := h.progressService.GetCourseProgress(courseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to fetch course progress: "+err.Error())
	}
	return response.SendSuccess(c, "Course progress retrieved successfully", fiber.Map{
		"courseProgress": items,
	})
}

// VerifyProgress godoc
// @Summary Verify progress (Teachers/TAs only)
// @Description ตรวจสอบและอนุมัติ/ปฏิเสธความก้าวหน้า หาก approved จะเปลี่ยนสถานะ progress เป็น completed
// @Tags progress
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Progress ID"
// @Success 200 {object} object{success=bool,message=string,data=object{logId=string,progressId=string,verifiedBy=string,status=string,comment=string,verifiedAt=string}}
// @Failure 400 {object} object{success=bool,error=string} "Bad request"
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 403 {object} object{success=bool,error=string} "Forbidden"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/progress/{id}/verify [post]
func (h *ProgressHandler) VerifyProgress(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	curUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || curUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	progressID := c.Params("progressId")
	if progressID == "" {
		progressID = c.Params("id")
	}
	if progressID == "" {
		return response.SendBadRequest(c, "Progress ID is required")
	}

	// Check permissions: must be teacher or TA in the course related to this progress
	canVerify, err := h.canVerifyProgress(claims.UserID, progressID, curUser.IsTeacher)
	if err != nil {
		return response.SendInternalError(c, "Failed to check permissions: "+err.Error())
	}
	if !canVerify {
		return response.SendError(c, fiber.StatusForbidden, "Only teachers or TAs can verify progress")
	}

	var req struct {
		Status  string `json:"status"`
		Comment string `json:"comment"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}

	var vStatus enums.VerificationStatus
	switch req.Status {
	case "approved":
		vStatus = enums.VerificationApproved
	case "rejected":
		vStatus = enums.VerificationRejected
	default:
		return response.SendBadRequest(c, "status must be 'approved' or 'rejected'")
	}

	log, err := h.progressService.VerifyProgress(progressID, claims.UserID, vStatus, req.Comment)
	if err != nil {
		return response.SendInternalError(c, "Failed to verify progress: "+err.Error())
	}

	return response.SendSuccess(c, "Progress verified successfully", fiber.Map{
		"logId":      log.LogID,
		"progressId": log.ProgressID,
		"verifiedBy": log.VerifiedBy,
		"status":     log.Status,
		"comment":    log.Comment,
		"verifiedAt": log.VerifiedAt,
	})
}

// GetVerificationLogs godoc
// @Summary Get verification logs
// @Description ดูประวัติการตรวจสอบความก้าวหน้า
// @Tags progress
// @Security BearerAuth
// @Produce json
// @Param id path string true "Progress ID"
// @Success 200 {object} object{success=bool,message=string,data=object{logs=[]object{logId=string,progressId=string,verifiedBy=string,status=string,comment=string,verifiedAt=string}}}
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 403 {object} object{success=bool,error=string} "Forbidden"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/progress/{id}/logs [get]
func (h *ProgressHandler) GetVerificationLogs(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	curUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || curUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	progressID := c.Params("progressId")
	if progressID == "" {
		progressID = c.Params("id")
	}
	if progressID == "" {
		return response.SendBadRequest(c, "Progress ID is required")
	}

	// Check permissions: must be teacher or TA in the course related to this progress
	canView, err := h.canVerifyProgress(claims.UserID, progressID, curUser.IsTeacher)
	if err != nil {
		return response.SendInternalError(c, "Failed to check permissions: "+err.Error())
	}
	if !canView {
		return response.SendError(c, fiber.StatusForbidden, "Only teachers or TAs can view verification logs")
	}

	logs, err := h.progressService.GetVerificationLogs(progressID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get logs: "+err.Error())
	}

	return response.SendSuccess(c, "Verification logs retrieved successfully", fiber.Map{
		"logs": logs,
	})
}


// RequestApproval godoc
// @Summary Request approval for material completion
// @Description Student requests TA approval for completed material with lab and table selection
// @Tags progress
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param material_id path string true "Material ID"
// @Param request body object{lab_room=string,table_number=string,notes=string} false "Approval request with lab/table selection (notes is optional)"
// @Success 200 {object} object{success=bool,message=string,data=object{queue_job_id=string,lab_room=string,table_number=string}}
// @Failure 400 {object} object{success=bool,error=string} "Bad request"
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 403 {object} object{success=bool,error=string} "Not eligible for approval"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/progress/{material_id}/request-approval [post]
func (h *ProgressHandler) RequestApproval(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	materialID := c.Params("material_id")
	if materialID == "" {
		return response.SendBadRequest(c, "Material ID is required")
	}

	var req struct {
		LabRoom     string `json:"lab_room" validate:"required"`
		TableNumber string `json:"table_number" validate:"required"`
		Notes       string `json:"notes"` // Optional field, can be empty
	}

	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}

	if req.LabRoom == "" || req.TableNumber == "" {
		return response.SendBadRequest(c, "Lab room and table number are required")
	}

	// Check if student is eligible for approval request
	canRequest, err := h.progressService.CanRequestApproval(claims.UserID, materialID)
	if err != nil {
		return response.SendInternalError(c, "Failed to check approval eligibility: "+err.Error())
	}
	if !canRequest {
		return response.SendError(c, fiber.StatusForbidden, "Material not completed or already under review")
	}

	// Create approval request with lab/table selection
	queueJob, err := h.progressService.RequestApproval(claims.UserID, materialID, req.LabRoom, req.TableNumber, req.Notes)
	if err != nil {
		return response.SendInternalError(c, "Failed to request approval: "+err.Error())
	}

	return response.SendSuccess(c, "Approval requested successfully", fiber.Map{
		"queue_job_id": queueJob.ID,
		"lab_room":     req.LabRoom,
		"table_number": req.TableNumber,
		"status":       "pending",
	})
}

// Helper methods for permission checking

// canViewCourseProgress checks if user can view course progress
func (h *ProgressHandler) canViewCourseProgress(userID, courseID string, isTeacher bool) (bool, error) {
	// Teachers can view any course progress
	if isTeacher {
		return true, nil
	}

	// Check if user is enrolled as TA in this course
	enrollment, err := h.enrollmentService.GetUserEnrollmentInCourse(courseID, userID)
	if err != nil {
		return false, err
	}
	if enrollment == nil {
		return false, nil // Not enrolled
	}

	// TAs can view course progress
	return enrollment.Role == enums.EnrollmentRoleTA, nil
}

// canVerifyProgress checks if user can verify progress
func (h *ProgressHandler) canVerifyProgress(userID, progressID string, isTeacher bool) (bool, error) {
	// Teachers can verify any progress
	if isTeacher {
		return true, nil
	}

	// Get the course related to this progress
	courseID, err := h.getCourseIDFromProgressID(progressID)
	if err != nil {
		return false, err
	}
	if courseID == "" {
		return false, nil
	}

	// Check if user is enrolled as TA in the related course
	enrollment, err := h.enrollmentService.GetUserEnrollmentInCourse(courseID, userID)
	if err != nil {
		return false, err
	}
	if enrollment == nil {
		return false, nil // Not enrolled
	}

	// TAs can verify progress
	return enrollment.Role == enums.EnrollmentRoleTA, nil
}

// getCourseIDFromProgressID gets course ID from progress ID
func (h *ProgressHandler) getCourseIDFromProgressID(progressID string) (string, error) {
	// This requires a new method in ProgressService to get progress details
	// or we can get it through exercise -> course relationship
	return h.progressService.GetCourseIDFromProgress(progressID)
}
