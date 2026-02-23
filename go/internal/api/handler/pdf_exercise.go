package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Project-DSView/backend/go/internal/api/types"
	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/pkg/config"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type PDFExerciseHandler struct {
	pdfSubmissionService *services.PDFExerciseSubmissionService
}

func NewPDFExerciseHandler(pdfSubmissionService *services.PDFExerciseSubmissionService) *PDFExerciseHandler {
	return &PDFExerciseHandler{
		pdfSubmissionService: pdfSubmissionService,
	}
}

// SubmitPDFExercise godoc
// @Summary Submit PDF exercise
// @Description Submit a PDF file for an exercise
// @Tags pdf-exercises
// @Accept multipart/form-data
// @Produce json
// @Param material_id path string true "Material ID"
// @Param file formData file true "PDF file"
// @Success 201 {object} response.StandardResponse{data=models.Submission}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/materials/{material_id}/submit [post]
// @Security BearerAuth
func (h *PDFExerciseHandler) SubmitPDFExercise(c *fiber.Ctx) error {
	materialID := c.Params("material_id")
	if materialID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Material ID is required", nil)
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "No file uploaded", err.Error())
	}

	// Validate file type
	if file.Header.Get("Content-Type") != "application/pdf" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Only PDF files are allowed", nil)
	}

	// Validate file size (max 10MB)
	const maxFileSize = 10 * 1024 * 1024
	if file.Size > maxFileSize {
		return response.ErrorResponse(c, http.StatusBadRequest, "File size too large. Maximum size is 10MB", nil)
	}

	// Open file for reading
	fileReader, err := file.Open()
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Failed to open file", err.Error())
	}
	defer fileReader.Close()

	// Submit PDF exercise
	submission, err := h.pdfSubmissionService.SubmitPDFExercise(
		userID,
		materialID,
		fileReader,
		file.Filename,
		file.Size,
		file.Header.Get("Content-Type"),
	)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit PDF exercise", err.Error())
	}

	return response.SuccessResponse(c, http.StatusCreated, "PDF exercise submitted successfully", submission)
}

// ApprovePDFSubmission godoc
// @Summary Approve PDF submission
// @Description Approve a PDF submission and assign score (Teachers/TAs only). Supports optional feedback file upload.
// @Tags pdf-exercises
// @Accept multipart/form-data
// @Produce json
// @Param submission_id path string true "Submission ID"
// @Param score formData int true "Score (0-100)"
// @Param comment formData string true "Comment/Feedback"
// @Param feedback_file formData file false "Optional feedback PDF file"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/submissions/{submission_id}/approve [post]
// @Security BearerAuth
func (h *PDFExerciseHandler) ApprovePDFSubmission(c *fiber.Ctx) error {
	submissionID := c.Params("submission_id")
	if submissionID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Submission ID is required", nil)
	}

	// Parse form data
	scoreStr := c.FormValue("score")
	maxScoreStr := c.FormValue("maxScore")
	comment := c.FormValue("comment")

	if scoreStr == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Score is required", nil)
	}
	if comment == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Comment is required", nil)
	}

	// Parse score
	var score int
	if _, err := fmt.Sscanf(scoreStr, "%d", &score); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid score format", nil)
	}

	// Parse maxScore (optional, defaults to 100 if not provided)
	var maxScore int = 100
	if maxScoreStr != "" {
		if _, err := fmt.Sscanf(maxScoreStr, "%d", &maxScore); err != nil {
			return response.ErrorResponse(c, http.StatusBadRequest, "Invalid maxScore format", nil)
		}
		if maxScore <= 0 {
			return response.ErrorResponse(c, http.StatusBadRequest, "MaxScore must be greater than 0", nil)
		}
	}

	// Validate score range
	if score < 0 {
		return response.ErrorResponse(c, http.StatusBadRequest, "Score must be greater than or equal to 0", nil)
	}
	if score > maxScore {
		return response.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Score must be between 0 and %d", maxScore), nil)
	}

	// Get optional feedback file
	var feedbackFileReader io.Reader
	var feedbackFileName string
	var feedbackFileSize int64
	var feedbackFileMimeType string

	file, err := c.FormFile("feedback_file")
	if err == nil && file != nil {
		// Validate file type
		if file.Header.Get("Content-Type") != "application/pdf" {
			return response.ErrorResponse(c, http.StatusBadRequest, "Feedback file must be a PDF", nil)
		}

		// Validate file size (max 10MB)
		const maxFileSize = 10 * 1024 * 1024
		if file.Size > maxFileSize {
			return response.ErrorResponse(c, http.StatusBadRequest, "Feedback file size too large. Maximum size is 10MB", nil)
		}

		feedbackFileReader, err = file.Open()
		if err != nil {
			return response.ErrorResponse(c, http.StatusBadRequest, "Failed to open feedback file", err.Error())
		}
		defer feedbackFileReader.(io.Closer).Close()
		feedbackFileName = file.Filename
		feedbackFileSize = file.Size
		feedbackFileMimeType = file.Header.Get("Content-Type")
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	// Approve submission
	if err := h.pdfSubmissionService.ApprovePDFSubmissionWithFile(
		submissionID,
		userID,
		score,
		comment,
		feedbackFileReader,
		feedbackFileName,
		feedbackFileSize,
		feedbackFileMimeType,
	); err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to approve submission", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Submission approved successfully", nil)
}

// RejectPDFSubmission godoc
// @Summary Reject PDF submission
// @Description Reject a PDF submission (Teachers/TAs only)
// @Tags pdf-exercises
// @Accept json
// @Produce json
// @Param submission_id path string true "Submission ID"
// @Param request body types.RejectPDFSubmissionRequest true "Rejection data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/submissions/{submission_id}/reject [post]
// @Security BearerAuth
func (h *PDFExerciseHandler) RejectPDFSubmission(c *fiber.Ctx) error {
	submissionID := c.Params("submission_id")
	if submissionID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Submission ID is required", nil)
	}

	var req types.RejectPDFSubmissionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validate request
	if err := config.Validate.Struct(req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	// Reject submission
	if err := h.pdfSubmissionService.RejectPDFSubmission(submissionID, userID, req.Comment); err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to reject submission", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Submission rejected successfully", nil)
}

// GetPDFSubmissions godoc
// @Summary Get PDF submissions
// @Description Get PDF submissions for a material (Teachers/TAs only)
// @Tags pdf-exercises
// @Produce json
// @Param material_id path string true "Material ID"
// @Success 200 {object} response.StandardResponse{data=[]models.Submission}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/materials/{material_id}/submissions [get]
// @Security BearerAuth
func (h *PDFExerciseHandler) GetPDFSubmissions(c *fiber.Ctx) error {
	materialID := c.Params("material_id")
	if materialID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Material ID is required", nil)
	}

	// Get submissions
	submissions, err := h.pdfSubmissionService.GetPDFSubmissions(materialID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get submissions", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Submissions retrieved successfully", submissions)
}

// GetPDFSubmission godoc
// @Summary Get PDF submission
// @Description Get a specific PDF submission
// @Tags pdf-exercises
// @Produce json
// @Param submission_id path string true "Submission ID"
// @Success 200 {object} response.StandardResponse{data=models.Submission}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/submissions/{submission_id} [get]
// @Security BearerAuth
func (h *PDFExerciseHandler) GetPDFSubmission(c *fiber.Ctx) error {
	submissionID := c.Params("submission_id")
	if submissionID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Submission ID is required", nil)
	}

	// Get submission
	submission, err := h.pdfSubmissionService.GetPDFSubmission(submissionID)
	if err != nil {
		if err.Error() == "submission not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Submission not found", nil)
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get submission", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Submission retrieved successfully", submission)
}

// DownloadPDFSubmission godoc
// @Summary Download PDF submission
// @Description Stream a PDF submission file directly (Teachers/TAs only)
// @Tags pdf-exercises
// @Produce application/pdf
// @Param submission_id path string true "Submission ID"
// @Success 200 {file} file "PDF file"
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/submissions/{submission_id}/download [get]
// @Security BearerAuth
func (h *PDFExerciseHandler) DownloadPDFSubmission(c *fiber.Ctx) error {
	submissionID := c.Params("submission_id")
	if submissionID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Submission ID is required", nil)
	}

	// Stream file directly from MinIO
	reader, contentType, filename, size, err := h.pdfSubmissionService.StreamPDFSubmission(submissionID)
	if err != nil {
		if err.Error() == "submission not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Submission not found", nil)
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to stream file", err.Error())
	}
	defer func() {
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	}()

	// Get origin from request for CORS
	origin := c.Get("Origin")
	if origin == "" {
		// Try to extract from Referer header
		referer := c.Get("Referer")
		if referer != "" {
			// Extract origin from referer URL (e.g., "https://localhost:3000" from "https://localhost:3000/page")
			if strings.HasPrefix(referer, "https://localhost:3000") {
				origin = "https://localhost:3000"
			} else if strings.HasPrefix(referer, "http://localhost:3000") {
				origin = "http://localhost:3000"
			}
		}
	}

	// Set CORS headers explicitly (must be set before streaming)
	// Note: Cannot use "*" with Allow-Credentials, so we must use specific origin
	if origin != "" {
		c.Set("Access-Control-Allow-Origin", origin)
		c.Set("Access-Control-Allow-Credentials", "true")
	} else {
		// If no origin, allow all (but cannot use credentials)
		c.Set("Access-Control-Allow-Origin", "*")
	}
	c.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, Cookie, X-Requested-With, X-CSRF-Token, dsview-api-key")

	// Set response headers for file download
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Set("Content-Length", fmt.Sprintf("%d", size))

	// Stream the file using io.Copy for better CORS compatibility
	c.Status(200)
	_, copyErr := io.Copy(c.Response().BodyWriter(), reader)
	if copyErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to stream file: "+copyErr.Error())
	}
	return nil
}

// CancelPDFSubmission godoc
// @Summary Cancel PDF submission
// @Description Cancel a PDF submission (only by the submitter)
// @Tags pdf-exercises
// @Produce json
// @Param submission_id path string true "Submission ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/submissions/{submission_id}/cancel [delete]
// @Security BearerAuth
func (h *PDFExerciseHandler) CancelPDFSubmission(c *fiber.Ctx) error {
	submissionID := c.Params("submission_id")
	if submissionID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Submission ID is required", nil)
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	// Cancel submission
	if err := h.pdfSubmissionService.CancelPDFSubmission(submissionID, userID); err != nil {
		if err.Error() == "submission not found or not owned by user" {
			return response.ErrorResponse(c, http.StatusNotFound, "Submission not found or not owned by user", nil)
		}
		if err.Error() == "only pending submissions can be cancelled" {
			return response.ErrorResponse(c, http.StatusBadRequest, "Only pending submissions can be cancelled", nil)
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to cancel submission", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Submission cancelled successfully", nil)
}

// GetMyPDFSubmission godoc
// @Summary Get my PDF submission
// @Description Get the current user's submission for a specific material
// @Tags pdf-exercises
// @Produce json
// @Param material_id path string true "Material ID"
// @Success 200 {object} response.StandardResponse{data=models.Submission}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/materials/{material_id}/submissions/me [get]
// @Security BearerAuth
func (h *PDFExerciseHandler) GetMyPDFSubmission(c *fiber.Ctx) error {
	materialID := c.Params("material_id")
	if materialID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Material ID is required", nil)
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	// Get user's submission
	submission, err := h.pdfSubmissionService.GetUserSubmissionForMaterial(userID, materialID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get submission", err.Error())
	}

	if submission == nil {
		return response.SuccessResponse(c, http.StatusOK, "No submission found", nil)
	}

	return response.SuccessResponse(c, http.StatusOK, "Submission retrieved successfully", submission)
}

// GetCoursePDFSubmissions godoc
// @Summary Get PDF submissions for a course
// @Description Get all PDF submissions for a course (Teachers/TAs only)
// @Tags pdf-exercises
// @Produce json
// @Param course_id path string true "Course ID"
// @Success 200 {object} response.StandardResponse{data=[]services.CoursePDFSubmission}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/courses/{course_id}/pdf-submissions [get]
// @Security BearerAuth
func (h *PDFExerciseHandler) GetCoursePDFSubmissions(c *fiber.Ctx) error {
	courseID := c.Params("course_id")
	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	// Get submissions
	submissions, err := h.pdfSubmissionService.GetPDFSubmissionsByCourse(courseID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get submissions", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Submissions retrieved successfully", submissions)
}

// DownloadFeedbackFile godoc
// @Summary Download feedback file
// @Description Stream a feedback PDF file directly (Students only, for their own submissions)
// @Tags pdf-exercises
// @Produce application/pdf
// @Param submission_id path string true "Submission ID"
// @Success 200 {file} file "PDF file"
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/submissions/{submission_id}/feedback/download [get]
// @Security BearerAuth
func (h *PDFExerciseHandler) DownloadFeedbackFile(c *fiber.Ctx) error {
	submissionID := c.Params("submission_id")
	if submissionID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Submission ID is required", nil)
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	// Stream feedback file directly from MinIO
	reader, contentType, filename, size, err := h.pdfSubmissionService.StreamFeedbackFile(submissionID, userID)
	if err != nil {
		if err.Error() == "submission not found" || err.Error() == "feedback file not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Feedback file not found", nil)
		}
		if err.Error() == "unauthorized" {
			return response.ErrorResponse(c, http.StatusForbidden, "You don't have permission to access this feedback file", nil)
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to stream file", err.Error())
	}
	defer func() {
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	}()

	// Get origin from request for CORS
	origin := c.Get("Origin")
	if origin == "" {
		// Try to extract from Referer header
		referer := c.Get("Referer")
		if referer != "" {
			if strings.HasPrefix(referer, "https://localhost:3000") {
				origin = "https://localhost:3000"
			} else if strings.HasPrefix(referer, "http://localhost:3000") {
				origin = "http://localhost:3000"
			}
		}
	}

	// Set CORS headers explicitly
	if origin != "" {
		c.Set("Access-Control-Allow-Origin", origin)
		c.Set("Access-Control-Allow-Credentials", "true")
	} else {
		c.Set("Access-Control-Allow-Origin", "*")
	}
	c.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, Cookie, X-Requested-With, X-CSRF-Token, dsview-api-key")

	// Set response headers for file download
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Set("Content-Length", fmt.Sprintf("%d", size))

	// Stream the file
	c.Status(200)
	_, copyErr := io.Copy(c.Response().BodyWriter(), reader)
	if copyErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to stream file: "+copyErr.Error())
	}
	return nil
}
