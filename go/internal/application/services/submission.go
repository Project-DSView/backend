package services

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/external"
	"github.com/Project-DSView/backend/go/pkg/logger"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"gorm.io/gorm"
)

type SubmissionService struct {
	db                 *gorm.DB
	testCaseService    *TestCaseService
	userService        *UserService
	deadlineService    *DeadlineCheckerService
	courseScoreService *CourseScoreService
	materialService    *CourseMaterialService
	exec               *external.DockerExecutor
	storageService     storage.StorageService
	queueService       *QueueService
}

func NewSubmissionService(
	db *gorm.DB,
	testCaseSvc *TestCaseService,
	userSvc *UserService,
	deadlineSvc *DeadlineCheckerService,
	courseScoreSvc *CourseScoreService,
	materialSvc *CourseMaterialService,
	exec *external.DockerExecutor,
	storageSvc storage.StorageService,
	queueSvc *QueueService,
) *SubmissionService {
	return &SubmissionService{
		db:                 db,
		testCaseService:    testCaseSvc,
		userService:        userSvc,
		deadlineService:    deadlineSvc,
		courseScoreService: courseScoreSvc,
		materialService:    materialSvc,
		exec:               exec,
		storageService:     storageSvc,
		queueService:       queueSvc,
	}
}

// SubmitResult is now defined in internal/types/services.go

// getValueType returns a human-readable type description
func getValueType(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64:
		if v == float64(int64(v)) {
			return "integer"
		}
		return "number"
	case string:
		return "string"
	case []interface{}:
		return fmt.Sprintf("array (length: %d)", len(v))
	case map[string]interface{}:
		return fmt.Sprintf("object (keys: %d)", len(v))
	default:
		return fmt.Sprintf("unknown (%T)", value)
	}
}

// analyzeFailureReason provides detailed analysis of why a test case failed
func analyzeFailureReason(expected, actual interface{}) string {
	expectedType := getValueType(expected)
	actualType := getValueType(actual)

	if expectedType != actualType {
		return fmt.Sprintf("Type mismatch: expected %s, got %s", expectedType, actualType)
	}

	switch exp := expected.(type) {
	case []interface{}:
		if act, ok := actual.([]interface{}); ok {
			if len(exp) != len(act) {
				return fmt.Sprintf("Array length mismatch: expected %d elements, got %d elements", len(exp), len(act))
			}
			// Find first differing element
			for i := 0; i < len(exp) && i < len(act); i++ {
				if !external.CompareJSON(exp[i], act[i]) {
					return fmt.Sprintf("Array element at index %d differs: expected %v, got %v",
						i, exp[i], act[i])
				}
			}
		}
	case map[string]interface{}:
		if act, ok := actual.(map[string]interface{}); ok {
			// Check for missing keys
			for key := range exp {
				if _, exists := act[key]; !exists {
					return fmt.Sprintf("Missing key '%s' in output object", key)
				}
			}
			// Check for extra keys
			for key := range act {
				if _, exists := exp[key]; !exists {
					return fmt.Sprintf("Unexpected key '%s' in output object", key)
				}
			}
			// Check for different values
			for key, expectedVal := range exp {
				if actualVal, exists := act[key]; exists {
					if !external.CompareJSON(expectedVal, actualVal) {
						return fmt.Sprintf("Value mismatch for key '%s': expected %v, got %v",
							key, expectedVal, actualVal)
					}
				}
			}
		}
	case string:
		if act, ok := actual.(string); ok {
			if len(exp) != len(act) {
				return fmt.Sprintf("String length mismatch: expected %d characters, got %d characters",
					len(exp), len(act))
			}
			return "String content differs"
		}
	case float64:
		if act, ok := actual.(float64); ok {
			diff := exp - act
			if diff < 0 {
				diff = -diff
			}
			return fmt.Sprintf("Number mismatch: difference of %.6f", diff)
		}
	}

	return "Values do not match"
}

func (s *SubmissionService) SubmitExercise(
	userID, exerciseID, code string,
) (*types.SubmitResult, error) {
	return nil, fmt.Errorf("legacy exercise submission is deprecated. Please use course materials API instead")
}

// SubmissionFilter is now defined in internal/types/services.go

func (s *SubmissionService) ListExerciseSubmissions(exerciseID string, f types.SubmissionFilter) ([]models.Submission, int, error) {
	return nil, 0, fmt.Errorf("legacy exercise submissions are deprecated. Please use course materials API instead")
}

func (s *SubmissionService) GetSubmissionByID(id string) (*models.Submission, error) {
	// Cleanup old submissions before querying (async, don't wait for result)
	go func() {
		if err := s.CleanupOldSubmissions(); err != nil {
			logger.Warnf("Failed to cleanup old submissions: %v", err)
		}
	}()

	var sub models.Submission
	if err := s.db.Preload("Results").
		Where("submission_id = ?", id).
		First(&sub).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get submission: %w", err)
	}
	return &sub, nil
}

// SubmitMaterialExercise submits code for a material-based exercise (new system)
func (s *SubmissionService) SubmitMaterialExercise(
	userID, materialID, code string,
) (*types.SubmitResult, error) {
	// Get material with test cases
	material, testCases, err := s.materialService.GetCourseMaterialWithTestCases(materialID)
	if err != nil {
		return nil, fmt.Errorf("get material: %w", err)
	}

	// Validate material is a code exercise
	if !material.IsCodeExercise() {
		return nil, fmt.Errorf("material is not a code exercise")
	}

	// Check if material has test cases (only code exercises have test cases)
	if len(testCases) == 0 {
		return nil, fmt.Errorf("code exercise must have test cases")
	}

	// Get code exercise details for deadline and course info
	var codeExercise models.CodeExercise
	if err := s.db.First(&codeExercise, "material_id = ?", materialID).Error; err != nil {
		return nil, fmt.Errorf("failed to get code exercise details: %w", err)
	}

	// Get course info for storage upload
	var course models.Course
	if err := s.db.First(&course, "course_id = ?", material.CourseID).Error; err != nil {
		return nil, fmt.Errorf("failed to get course details: %w", err)
	}

	// Check deadline (if material has deadline)
	var isLateSubmission bool
	if codeExercise.Deadline != nil {
		canSubmit, message, err := s.deadlineService.CanSubmitMaterial(userID, materialID)
		if err != nil {
			return nil, fmt.Errorf("check deadline: %w", err)
		}
		if !canSubmit {
			return nil, fmt.Errorf("cannot submit: %s", message)
		}
		// Mark as late submission if message indicates it
		if message != "" {
			isLateSubmission = true
		}
	}

	// Cancel pending/processing queue jobs and reset progress status for resubmission
	// Keep old submissions for history
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Cancel pending and processing queue jobs for this material (if any)
		// This allows resubmission to work properly
		if err := tx.Model(&models.QueueJob{}).
			Where("user_id = ? AND material_id = ? AND status IN ?", userID, materialID, []string{string(enums.QueueStatusPending), string(enums.QueueStatusProcessing)}).
			Updates(map[string]interface{}{
				"status":       enums.QueueStatusCancelled,
				"completed_at": time.Now(),
			}).Error; err != nil {
			// Log error but continue - don't fail the transaction
			logger.Warnf("Failed to cancel queue jobs: %v", err)
		}

		// Reset progress status to in_progress if it was waiting_approval or completed
		// This allows the student to request approval again after resubmission
		if err := tx.Model(&models.StudentProgress{}).
			Where("user_id = ? AND material_id = ? AND status IN ?", userID, materialID, []string{string(enums.ProgressWaitingApproval), string(enums.ProgressCompleted)}).
			Update("status", enums.ProgressInProgress).Error; err != nil {
			// Log error but continue - don't fail the transaction
			logger.Warnf("Failed to reset progress status: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cleanup for resubmission: %w", err)
	}

	// Get user information for MinIO upload
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Upload code to MinIO
	codeReader := strings.NewReader(code)
	fileName := "submission.py"
	fileURL, err := s.storageService.UploadStudentCodeSubmission(
		context.Background(),
		material.CourseID,
		course.Name,
		material.Week,
		user.Email,
		codeReader,
		fileName,
		"text/x-python",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload code to storage: %w", err)
	}

	// Create submission
	sub := &models.Submission{
		UserID:           userID,
		MaterialID:       materialID,
		Code:             code,
		FileURL:          fileURL,
		FileName:         fileName,
		FileSize:         int64(len(code)),
		MimeType:         "text/x-python",
		Status:           enums.SubmissionRunning,
		IsLateSubmission: isLateSubmission,
		SubmittedAt:      time.Now(),
	}

	if err := s.db.Create(sub).Error; err != nil {
		return nil, fmt.Errorf("create submission: %w", err)
	}

	// Update submission status to running
	if err := s.db.Model(&models.Submission{}).
		Where("submission_id = ?", sub.SubmissionID).
		Update("status", enums.SubmissionRunning).Error; err != nil {
		return nil, fmt.Errorf("update submission status: %w", err)
	}

	// Submit to code execution queue (async processing)
	if s.queueService != nil {
		_, err := s.queueService.SubmitCodeExecutionJob(
			context.Background(),
			userID,
			materialID,
			sub.SubmissionID,
			code,
			material.CourseID, // Pass courseID for course-specific queue
		)
		if err != nil {
			// If queue submission fails, fall back to synchronous execution
			logger.Warnf("Failed to submit to queue, falling back to synchronous execution: %v", err)
			if err := s.ExecuteCodeSubmission(sub.SubmissionID, code, materialID); err != nil {
				return nil, fmt.Errorf("code execution failed: %w", err)
			}
		}
	} else {
		// No queue available, execute synchronously
		if err := s.ExecuteCodeSubmission(sub.SubmissionID, code, materialID); err != nil {
			return nil, fmt.Errorf("code execution failed: %w", err)
		}
	}

	// reload submission with results
	var saved models.Submission
	if err := s.db.Preload("Results").
		Where("submission_id = ?", sub.SubmissionID).
		First(&saved).Error; err != nil {
		return nil, fmt.Errorf("reload submission: %w", err)
	}

	// No auto-queue creation - student must manually request review
	// Progress status is set to in_progress when all tests pass, ready for manual review request

	return &types.SubmitResult{
		Submission: &saved,
		Results:    saved.Results,
	}, nil
}

// ExecuteCodeSubmission executes code and runs test cases for a submission (called by queue worker)
func (s *SubmissionService) ExecuteCodeSubmission(submissionID, code, materialID string) error {
	// Get submission
	var sub models.Submission
	if err := s.db.Where("submission_id = ?", submissionID).First(&sub).Error; err != nil {
		return fmt.Errorf("get submission: %w", err)
	}

	// Get material with test cases
	material, testCases, err := s.materialService.GetCourseMaterialWithTestCases(materialID)
	if err != nil {
		return fmt.Errorf("get material: %w", err)
	}

	// Validate material is a code exercise
	if !material.IsCodeExercise() {
		return fmt.Errorf("material is not a code exercise")
	}

	// Get code exercise details for total points
	var codeExercise models.CodeExercise
	if err := s.db.First(&codeExercise, "material_id = ?", materialID).Error; err != nil {
		return fmt.Errorf("failed to get code exercise details: %w", err)
	}

	// Run test cases
	passed := 0
	results := make([]models.SubmissionResult, 0, len(testCases))
	totalPoints := 0
	if codeExercise.TotalPoints != nil {
		totalPoints = *codeExercise.TotalPoints
	}

	for i, tc := range testCases {

		// input JSON for STDIN
		stdinBytes, _ := json.Marshal(tc.InputData)

		execRes, runErr := s.exec.RunPython(code, string(stdinBytes))

		result := models.SubmissionResult{
			SubmissionID: sub.SubmissionID,
			TestCaseID:   tc.TestCaseID,
		}

		if runErr != nil {
			result.Status = "error"
			result.ErrorMessage = fmt.Sprintf("Execution error: %s", runErr.Error())

		} else if execRes.TimedOut {
			result.Status = "error"
			result.ErrorMessage = "Code execution timed out. Your program may have an infinite loop or is taking too long to execute."

		} else if execRes.ExitCode != 0 {
			result.Status = "error"
			errorMsg := "Your code encountered an error during execution."
			if execRes.Stderr != "" {
				errorMsg = fmt.Sprintf("Runtime error: %s", execRes.Stderr)
			} else {
				errorMsg = fmt.Sprintf("Your code exited with error code %d", execRes.ExitCode)
			}
			result.ErrorMessage = errorMsg

		} else {
			// parse student's stdout as JSON

			var actual interface{}
			if err := json.Unmarshal([]byte(execRes.Stdout), &actual); err != nil {
				result.Status = "error"
				result.ErrorMessage = fmt.Sprintf("Your code output is not valid JSON: %s\nOutput received: %s",
					err.Error(), execRes.Stdout)

			} else {
				result.ActualOutput = types.JSONData{}
				_ = json.Unmarshal([]byte(execRes.Stdout), &result.ActualOutput)

				var expected interface{}
				expected = tc.ExpectedOutput

				if external.CompareJSON(expected, actual) {
					result.Status = "passed"
					passed++

				} else {
					result.Status = "failed"

					// Create detailed error message
					inputJSON, _ := json.Marshal(tc.InputData)
					expectedJSON, _ := json.Marshal(expected)
					actualJSON, _ := json.Marshal(actual)

					failureReason := analyzeFailureReason(expected, actual)

					result.ErrorMessage = fmt.Sprintf(
						"Test case %d failed\n\n"+
							"Input:\n%s\n\n"+
							"Expected output:\n%s\n\n"+
							"Your output:\n%s\n\n"+
							"Reason: %s",
						i+1,
						string(inputJSON),
						string(expectedJSON),
						string(actualJSON),
						failureReason,
					)

				}
			}
		}

		results = append(results, result)
	}

	failed := len(testCases) - passed
	score := external.ScoreFromCounts(totalPoints, passed, len(testCases))

	// persist results
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		// save results
		if len(results) > 0 {
			if err := tx.Create(&results).Error; err != nil {
				return fmt.Errorf("create results: %w", err)
			}
		}
		// update submission
		update := map[string]interface{}{
			"passed_count": passed,
			"failed_count": failed,
			"total_score":  score,
			"status":       enums.SubmissionPending,
		}
		if err := tx.Model(&models.Submission{}).
			Where("submission_id = ?", sub.SubmissionID).
			Updates(update).Error; err != nil {
			return fmt.Errorf("update submission: %w", err)
		}

		// upsert student progress
		var prog models.StudentProgress
		err := tx.Where("user_id = ? AND material_id = ?", sub.UserID, materialID).
			First(&prog).Error
		now := time.Now()
		// กำหนดสถานะตามการผ่าน test case
		var newStatus enums.ProgressStatus
		if passed == len(testCases) && len(testCases) > 0 {
			// ผ่านทุก test case -> พร้อมขอ review
			newStatus = enums.ProgressInProgress
		} else {
			// ไม่ผ่านหรือไม่มี test case
			newStatus = enums.ProgressNotStarted
		}

		if err == gorm.ErrRecordNotFound {
			prog = models.StudentProgress{
				UserID:          sub.UserID,
				MaterialID:      materialID,
				Status:          newStatus,
				Score:           score,
				LastSubmittedAt: &now,
			}
			if err := tx.Create(&prog).Error; err != nil {
				return fmt.Errorf("create progress: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("get progress: %w", err)
		} else {
			// อัพเดท progress ที่มีอยู่
			if score > prog.Score {
				prog.Score = score
			}
			// อัพเดทสถานะถ้าผ่านทุก test case
			if passed == len(testCases) && len(testCases) > 0 {
				prog.Status = enums.ProgressInProgress
			}
			prog.LastSubmittedAt = &now
			if err := tx.Where("user_id = ? AND material_id = ?", sub.UserID, materialID).Updates(&prog).Error; err != nil {
				return fmt.Errorf("update progress: %w", err)
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("persist results: %w", err)
	}

	return nil
}

// ProcessFile processes uploaded files (PDF, images, videos) - called by queue worker
func (s *SubmissionService) ProcessFile(jobData types.QueueJobData) error {
	// Get submission if submission ID is provided
	if jobData.SubmissionID == "" {
		return fmt.Errorf("submission ID is required")
	}

	var submission models.Submission
	if err := s.db.Where("submission_id = ?", jobData.SubmissionID).First(&submission).Error; err != nil {
		return fmt.Errorf("get submission: %w", err)
	}

	// Process based on file type
	switch jobData.SubmissionType {
	case "pdf":
		return s.processPDFFile(submission, jobData)
	case "image":
		return s.processImageFile(submission, jobData)
	case "video":
		return s.processVideoFile(submission, jobData)
	default:
		// For unknown types, just mark as processed
		return nil
	}
}

// processPDFFile processes PDF files: extract text, get page count, generate thumbnail
func (s *SubmissionService) processPDFFile(submission models.Submission, jobData types.QueueJobData) error {
	// TODO: Implement PDF processing
	// - Extract text from PDF
	// - Get page count
	// - Generate thumbnail for first page
	// - Store metadata in submission or separate table

	// For now, just log that processing would happen

	// Update submission with processing status (optional metadata field)
	// This is a placeholder - actual implementation would extract and store data
	return nil
}

// processImageFile processes image files: resize, optimize, generate thumbnails
func (s *SubmissionService) processImageFile(submission models.Submission, jobData types.QueueJobData) error {
	// TODO: Implement image processing
	// - Resize to standard sizes
	// - Optimize file size
	// - Generate thumbnails
	// - Extract metadata (dimensions, format, etc.)

	// For now, just log that processing would happen

	return nil
}

// processVideoFile processes video files: transcode, generate preview frames
func (s *SubmissionService) processVideoFile(submission models.Submission, jobData types.QueueJobData) error {
	// TODO: Implement video processing
	// - Extract metadata (duration, resolution, format)
	// - Generate preview frames/thumbnails
	// - Transcode to web-friendly format (optional)

	// For now, just log that processing would happen

	return nil
}

// SubmitPDFExercise handles PDF file submission for PDF exercises
func (s *SubmissionService) SubmitPDFExercise(userID, materialID string, file *multipart.FileHeader) (*models.Submission, error) {
	// Validate material exists and is PDF exercise
	var material models.CourseMaterial
	if err := s.db.Where("material_id = ? AND type = ?", materialID, enums.MaterialTypePDFExercise).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("PDF exercise material not found")
		}
		return nil, fmt.Errorf("get material: %w", err)
	}

	// Check if user is enrolled in the course
	var enrollment models.Enrollment
	if err := s.db.Where("course_id = ? AND user_id = ?", material.CourseID, userID).First(&enrollment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not enrolled in course")
		}
		return nil, fmt.Errorf("check enrollment: %w", err)
	}

	var submission *models.Submission
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Delete old submission if exists
		if err := s.deleteOldSubmission(tx, userID, materialID); err != nil {
			return fmt.Errorf("delete old submission: %w", err)
		}

		// Upload file to MinIO
		fileURL, err := s.uploadPDFFile(file, userID, materialID)
		if err != nil {
			return fmt.Errorf("upload file: %w", err)
		}

		// Create submission record
		submission = &models.Submission{
			UserID:     userID,
			MaterialID: materialID,
			FileURL:    fileURL,
			FileName:   file.Filename,
			FileSize:   file.Size,
			MimeType:   "application/pdf",
			Status:     enums.SubmissionPending,
		}

		if err := tx.Create(submission).Error; err != nil {
			return fmt.Errorf("create submission: %w", err)
		}

		// Create or update student progress
		var prog models.StudentProgress
		if err := tx.Where("user_id = ? AND material_id = ?", userID, materialID).First(&prog).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new progress
				prog = models.StudentProgress{
					UserID:     userID,
					MaterialID: materialID,
					Status:     enums.ProgressInProgress,
					Score:      0,
				}
				if err := tx.Create(&prog).Error; err != nil {
					return fmt.Errorf("create progress: %w", err)
				}
			} else {
				return fmt.Errorf("get progress: %w", err)
			}
		} else {
			// Update existing progress
			prog.Status = enums.ProgressInProgress
			prog.Score = 0 // Reset score for new submission
			if err := tx.Where("user_id = ? AND material_id = ?", userID, materialID).Updates(&prog).Error; err != nil {
				return fmt.Errorf("update progress: %w", err)
			}
		}

		// Create queue job for file processing (extract text, generate thumbnails, etc.)
		fileProcessingJobData := types.QueueJobData{
			FileURL:        fileURL,
			FileName:       file.Filename,
			FileSize:       file.Size,
			SubmissionType: "pdf",
			MaterialID:     materialID,
			CourseID:       material.CourseID,
			SubmissionID:   submission.SubmissionID,
		}

		fileProcessingDataJSON, _ := json.Marshal(fileProcessingJobData)
		fileProcessingJob := &models.QueueJob{
			Type:       enums.QueueTypeFileProcessing,
			Status:     enums.QueueStatusPending,
			UserID:     userID,
			MaterialID: &materialID,
			CourseID:   &material.CourseID,
			Data:       string(fileProcessingDataJSON),
		}

		if err := tx.Create(fileProcessingJob).Error; err != nil {
			return fmt.Errorf("create file processing job: %w", err)
		}

		// Publish file processing message to RabbitMQ (after transaction)
		// Note: Publish after transaction to avoid issues
		go func() {
			if s.queueService != nil {
				fileProcessingMessage := &external.QueueMessage{
					ID:   fileProcessingJob.ID,
					Type: string(enums.QueueTypeFileProcessing),
					Data: map[string]interface{}{"job_id": fileProcessingJob.ID},
				}
				// Use a method to publish with courseID
				if err := s.queueService.PublishFileProcessingJob(context.Background(), material.CourseID, fileProcessingMessage); err != nil {
					logger.Warnf("Failed to publish file processing message: %v", err)
				}
			}
		}()

		// Create queue job for PDF review
		queueJobData := types.QueueJobData{
			FileURL:        fileURL,
			FileName:       file.Filename,
			FileSize:       file.Size,
			SubmissionType: "pdf",
			MaterialID:     materialID,
			CourseID:       material.CourseID,
			SubmissionID:   submission.SubmissionID,
		}

		queueJob, err := s.queueService.SubmitReviewJobWithLabTable(
			userID,
			materialID,
			material.CourseID,
			submission.SubmissionID,
			"", // lab_room - will be set when student requests approval
			"", // table_number - will be set when student requests approval
			queueJobData,
		)
		if err != nil {
			return fmt.Errorf("create queue job: %w", err)
		}

		// Update submission with queue job info
		submission.QueueJobID = &queueJob.ID

		return nil
	})

	if err != nil {
		return nil, err
	}

	return submission, nil
}

// CleanupOldSubmissions deletes submissions where the material deadline has passed
func (s *SubmissionService) CleanupOldSubmissions() error {
	now := time.Now().Format(time.RFC3339)

	// Find submissions where material deadline has passed
	// Join with course_materials to check deadline
	var oldSubmissions []models.Submission
	if err := s.db.Table("submissions s").
		Joins("INNER JOIN course_materials cm ON s.material_id = cm.material_id").
		Where("cm.deadline IS NOT NULL AND cm.deadline != '' AND cm.deadline <= ?", now).
		Select("s.*").
		Find(&oldSubmissions).Error; err != nil {
		return fmt.Errorf("failed to find old submissions: %w", err)
	}

	deletedCount := 0
	for _, submission := range oldSubmissions {
		// Delete file from MinIO if exists
		if submission.FileURL != "" {
			if err := s.storageService.DeleteFile(context.Background(), submission.FileURL); err != nil {
				// Log error but continue
				logger.Warnf("Failed to delete file from MinIO for submission %s: %v", submission.SubmissionID, err)
			}
		}

		// Delete submission results
		if err := s.db.Where("submission_id = ?", submission.SubmissionID).Delete(&models.SubmissionResult{}).Error; err != nil {
			logger.Warnf("Failed to delete submission results for %s: %v", submission.SubmissionID, err)
			continue
		}

		// Delete submission
		if err := s.db.Delete(&submission).Error; err != nil {
			logger.Warnf("Failed to delete submission %s: %v", submission.SubmissionID, err)
			continue
		}

		deletedCount++
	}

	if deletedCount > 0 {
		logger.Infof("Cleaned up %d old submissions (past deadline)", deletedCount)
	}

	return nil
}

// deleteOldSubmission deletes old submission and its file from MinIO
func (s *SubmissionService) deleteOldSubmission(tx *gorm.DB, userID, materialID string) error {
	var oldSubmission models.Submission
	if err := tx.Where("user_id = ? AND material_id = ?", userID, materialID).First(&oldSubmission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No old submission to delete
		}
		return fmt.Errorf("get old submission: %w", err)
	}

	// Cancel pending queue jobs for this material (if any)
	// This allows resubmission to work properly
	if err := tx.Model(&models.QueueJob{}).
		Where("user_id = ? AND material_id = ? AND status = ?", userID, materialID, enums.QueueStatusPending).
		Updates(map[string]interface{}{
			"status":       enums.QueueStatusCancelled,
			"completed_at": time.Now(),
		}).Error; err != nil {
		// Log error but continue - don't fail the transaction
		fmt.Printf("Warning: Failed to cancel pending queue jobs: %v\n", err)
	}

	// Reset progress status to in_progress if it was waiting_approval
	// This allows the student to request approval again after resubmission
	if err := tx.Model(&models.StudentProgress{}).
		Where("user_id = ? AND material_id = ? AND status = ?", userID, materialID, enums.ProgressWaitingApproval).
		Update("status", enums.ProgressInProgress).Error; err != nil {
		// Log error but continue - don't fail the transaction
		fmt.Printf("Warning: Failed to reset progress status: %v\n", err)
	}

	// Delete file from MinIO
	if oldSubmission.FileURL != "" {
		if err := s.storageService.DeleteFile(context.Background(), oldSubmission.FileURL); err != nil {
			// Log error but continue - don't fail the transaction
			fmt.Printf("Warning: Failed to delete old file from MinIO: %v\n", err)
		}
	}

	// Delete submission results
	if err := tx.Where("submission_id = ?", oldSubmission.SubmissionID).Delete(&models.SubmissionResult{}).Error; err != nil {
		return fmt.Errorf("delete submission results: %w", err)
	}

	// Delete submission
	if err := tx.Delete(&oldSubmission).Error; err != nil {
		return fmt.Errorf("delete submission: %w", err)
	}

	return nil
}

// uploadPDFFile uploads PDF file to MinIO and returns the file URL
func (s *SubmissionService) uploadPDFFile(file *multipart.FileHeader, userID, materialID string) (string, error) {
	// Open file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer src.Close()

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s_%s_%d.pdf", userID, materialID, timestamp)

	// Upload to MinIO
	fileURL, err := s.storageService.UploadFile(context.Background(), filename, src, "application/pdf")
	if err != nil {
		return "", fmt.Errorf("upload to storage: %w", err)
	}

	return fileURL, nil
}
