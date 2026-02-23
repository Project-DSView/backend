package services

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/pkg/logger"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"gorm.io/gorm"
)

// PDFExerciseSubmissionService handles PDF exercise submissions
type PDFExerciseSubmissionService struct {
	db              *gorm.DB
	deadlineService *DeadlineCheckerService
	storageService  storage.StorageService
}

func NewPDFExerciseSubmissionService(db *gorm.DB, deadlineService *DeadlineCheckerService, storageService storage.StorageService) *PDFExerciseSubmissionService {
	return &PDFExerciseSubmissionService{
		db:              db,
		deadlineService: deadlineService,
		storageService:  storageService,
	}
}

// CleanupOldSubmissions deletes submissions where the material deadline has passed
func (s *PDFExerciseSubmissionService) CleanupOldSubmissions() error {
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

// GetUserSubmissionForMaterial gets a user's submission for a specific material
func (s *PDFExerciseSubmissionService) GetUserSubmissionForMaterial(userID, materialID string) (*models.Submission, error) {
	// Cleanup old submissions before querying (async, don't wait for result)
	go func() {
		if err := s.CleanupOldSubmissions(); err != nil {
			logger.Warnf("Failed to cleanup old submissions: %v", err)
		}
	}()

	var submission models.Submission
	err := s.db.Where("user_id = ? AND material_id = ?", userID, materialID).First(&submission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No submission found, not an error
		}
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}
	return &submission, nil
}

// SubmitPDFExercise submits a PDF file for an exercise
func (s *PDFExerciseSubmissionService) SubmitPDFExercise(
	userID, materialID string,
	file io.Reader,
	fileName string,
	fileSize int64,
	mimeType string,
) (*models.Submission, error) {
	// Get course material to validate
	var material models.CourseMaterial
	if err := s.db.Preload("Course").Where("material_id = ? AND type = ?", materialID, enums.MaterialTypePDFExercise).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("PDF exercise not found")
		}
		return nil, fmt.Errorf("failed to get material: %w", err)
	}

	// Check if user is enrolled in the course
	var enrollment models.Enrollment
	if err := s.db.Where("course_id = ? AND user_id = ?", material.CourseID, userID).First(&enrollment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not enrolled in course")
		}
		return nil, fmt.Errorf("failed to check enrollment: %w", err)
	}

	// Check deadline using deadline service
	canSubmit, message, err := s.deadlineService.CanSubmitMaterial(userID, materialID)
	if err != nil {
		return nil, fmt.Errorf("check submission eligibility: %w", err)
	}
	if !canSubmit {
		return nil, fmt.Errorf("cannot submit: %s", message)
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

	// Get user email for file organization
	var user models.User
	if err := s.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Upload file to MinIO
	fileURL, err := s.storageService.UploadStudentPDFSubmission(
		context.Background(),
		material.CourseID,
		material.Course.Name,
		material.Week,
		user.Email,
		file,
		fileName,
		mimeType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload PDF file: %w", err)
	}

	// Create submission (old submission already deleted above)
	submission := &models.Submission{
		UserID:      userID,
		MaterialID:  materialID,
		FileURL:     fileURL,
		FileName:    fileName,
		FileSize:    fileSize,
		MimeType:    mimeType,
		Status:      enums.SubmissionPending, // PDF exercises need manual review
		SubmittedAt: time.Now(),
	}

	if err := s.db.Create(submission).Error; err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	// Create or update student progress
	var progress models.StudentProgress
	err = s.db.Where("user_id = ? AND material_id = ?", userID, materialID).First(&progress).Error
	if err == gorm.ErrRecordNotFound {
		// Create new progress
		progress = models.StudentProgress{
			UserID:          userID,
			MaterialID:      materialID,
			Status:          enums.ProgressInProgress,
			Score:           0, // Will be updated after approval
			LastSubmittedAt: &submission.SubmittedAt,
		}
		if err := s.db.Create(&progress).Error; err != nil {
			return nil, fmt.Errorf("failed to create progress: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get progress: %w", err)
	} else {
		// Update existing progress
		progress.Status = enums.ProgressInProgress
		progress.LastSubmittedAt = &submission.SubmittedAt
		if err := s.db.Where("progress_id = ?", progress.ProgressID).Updates(&progress).Error; err != nil {
			return nil, fmt.Errorf("failed to update progress: %w", err)
		}
	}

	return submission, nil
}

// ApprovePDFSubmission approves a PDF submission and assigns score
func (s *PDFExerciseSubmissionService) ApprovePDFSubmission(
	submissionID, reviewerID string,
	score int,
	comment string,
) error {
	return s.ApprovePDFSubmissionWithFile(submissionID, reviewerID, score, comment, nil, "", 0, "")
}

// ApprovePDFSubmissionWithFile approves a PDF submission with optional feedback file
func (s *PDFExerciseSubmissionService) ApprovePDFSubmissionWithFile(
	submissionID, reviewerID string,
	score int,
	comment string,
	feedbackFile io.Reader,
	feedbackFileName string,
	feedbackFileSize int64,
	feedbackFileMimeType string,
) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get submission
		var submission models.Submission
		if err := tx.Where("submission_id = ?", submissionID).First(&submission).Error; err != nil {
			return fmt.Errorf("submission not found: %w", err)
		}

		// Get material to get course ID for feedback file upload
		var material models.CourseMaterial
		if err := tx.Where("material_id = ?", submission.MaterialID).First(&material).Error; err != nil {
			return fmt.Errorf("material not found: %w", err)
		}

		// Upload feedback file if provided
		feedbackFileURL := ""
		if feedbackFile != nil && feedbackFileName != "" {
			// Get student user info (the submitter) for file organization
			var student models.User
			if err := tx.Where("user_id = ?", submission.UserID).First(&student).Error; err != nil {
				return fmt.Errorf("failed to get student info: %w", err)
			}

			// Get course info for course name
			var course models.Course
			if err := tx.Where("course_id = ?", material.CourseID).First(&course).Error; err != nil {
				return fmt.Errorf("failed to get course info: %w", err)
			}

			// Upload feedback file to storage using same structure as exercise submissions
			uploadedURL, err := s.storageService.UploadStudentFeedbackFile(
				context.Background(),
				material.CourseID,
				course.Name,
				material.Week,
				student.Email,
				feedbackFile,
				feedbackFileName,
				feedbackFileMimeType,
			)
			if err != nil {
				return fmt.Errorf("failed to upload feedback file: %w", err)
			}
			feedbackFileURL = uploadedURL
		}

		// Update submission status
		now := time.Now()
		updates := map[string]interface{}{
			"status":      enums.SubmissionCompleted,
			"total_score": score,
			"feedback":    comment,
			"graded_at":   &now,
			"graded_by":   reviewerID,
		}
		if feedbackFileURL != "" {
			updates["feedback_file_url"] = feedbackFileURL
		}

		if err := tx.Model(&models.Submission{}).Where("submission_id = ?", submissionID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update submission: %w", err)
		}

		// Update student progress
		var progress models.StudentProgress
		if err := tx.Where("user_id = ? AND material_id = ?", submission.UserID, submission.MaterialID).First(&progress).Error; err != nil {
			return fmt.Errorf("progress not found: %w", err)
		}

		progress.Status = enums.ProgressCompleted
		progress.Score = score
		if err := tx.Where("user_id = ? AND material_id = ?", submission.UserID, submission.MaterialID).Updates(&progress).Error; err != nil {
			return fmt.Errorf("failed to update progress: %w", err)
		}

		// Create verification log
		verificationLog := models.VerificationLog{
			ProgressID: progress.ProgressID,
			VerifiedBy: reviewerID,
			Status:     enums.VerificationApproved,
			Comment:    comment,
		}
		if err := tx.Create(&verificationLog).Error; err != nil {
			return fmt.Errorf("failed to create verification log: %w", err)
		}

		// Update course score
		if err := s.updateCourseScore(tx, submission.UserID, material.CourseID); err != nil {
			return fmt.Errorf("failed to update course score: %w", err)
		}

		return nil
	})
}

// RejectPDFSubmission rejects a PDF submission
func (s *PDFExerciseSubmissionService) RejectPDFSubmission(
	submissionID, reviewerID, comment string,
) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get submission
		var submission models.Submission
		if err := tx.Where("submission_id = ?", submissionID).First(&submission).Error; err != nil {
			return fmt.Errorf("submission not found: %w", err)
		}

		// Update submission status
		now := time.Now()
		submission.Status = enums.SubmissionError
		submission.ErrorMessage = comment
		submission.Feedback = comment
		submission.GradedAt = &now
		submission.GradedBy = reviewerID
		if err := tx.Where("submission_id = ?", submissionID).Updates(&submission).Error; err != nil {
			return fmt.Errorf("failed to update submission: %w", err)
		}

		// Update student progress
		var progress models.StudentProgress
		if err := tx.Where("user_id = ? AND material_id = ?", submission.UserID, submission.MaterialID).First(&progress).Error; err != nil {
			return fmt.Errorf("progress not found: %w", err)
		}

		progress.Status = enums.ProgressInProgress // Keep as in_progress for resubmission
		if err := tx.Where("user_id = ? AND material_id = ?", submission.UserID, submission.MaterialID).Updates(&progress).Error; err != nil {
			return fmt.Errorf("failed to update progress: %w", err)
		}

		// Create verification log
		verificationLog := models.VerificationLog{
			ProgressID: progress.ProgressID,
			VerifiedBy: reviewerID,
			Status:     enums.VerificationRejected,
			Comment:    comment,
		}
		if err := tx.Create(&verificationLog).Error; err != nil {
			return fmt.Errorf("failed to create verification log: %w", err)
		}

		return nil
	})
}

// GetPDFSubmissions gets PDF submissions for a material (for teachers/TAs)
func (s *PDFExerciseSubmissionService) GetPDFSubmissions(materialID string) ([]models.Submission, error) {
	// Cleanup old submissions before querying (async, don't wait for result)
	go func() {
		if err := s.CleanupOldSubmissions(); err != nil {
			logger.Warnf("Failed to cleanup old submissions: %v", err)
		}
	}()

	var submissions []models.Submission
	if err := s.db.Where("material_id = ?", materialID).
		Order("submitted_at DESC").
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}
	return submissions, nil
}

// CoursePDFSubmission represents a PDF submission with enriched data for course view
type CoursePDFSubmission struct {
	models.Submission
	ExerciseTitle string `json:"exercise_title"`
	SubmitterName string `json:"submitter_name"`
}

// GetPDFSubmissionsByCourse gets all PDF submissions for a course (for teachers/TAs)
func (s *PDFExerciseSubmissionService) GetPDFSubmissionsByCourse(courseID string) ([]CoursePDFSubmission, error) {
	// Cleanup old submissions before querying (async, don't wait for result)
	go func() {
		if err := s.CleanupOldSubmissions(); err != nil {
			logger.Warnf("Failed to cleanup old submissions: %v", err)
		}
	}()

	// Get all PDF exercise materials for this course
	var pdfExercises []models.PDFExercise
	if err := s.db.Where("course_id = ?", courseID).Find(&pdfExercises).Error; err != nil {
		return nil, fmt.Errorf("failed to get PDF exercises: %w", err)
	}

	// Create a map of material ID to exercise title
	materialTitleMap := make(map[string]string)
	materialIDs := make([]string, 0, len(pdfExercises))
	for _, exercise := range pdfExercises {
		materialIDs = append(materialIDs, exercise.MaterialID)
		materialTitleMap[exercise.MaterialID] = exercise.Title
	}

	if len(materialIDs) == 0 {
		return []CoursePDFSubmission{}, nil
	}

	// Get all submissions for these materials
	var submissions []models.Submission
	if err := s.db.
		Where("material_id IN ?", materialIDs).
		Order("submitted_at DESC").
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	// Enrich submissions with user info and material title
	enrichedSubmissions := make([]CoursePDFSubmission, 0, len(submissions))
	for _, submission := range submissions {
		// Get user info
		var user models.User
		if err := s.db.Where("user_id = ?", submission.UserID).First(&user).Error; err != nil {
			// If user not found, continue with empty name
			user = models.User{}
		}

		// Get exercise title
		exerciseTitle := materialTitleMap[submission.MaterialID]
		if exerciseTitle == "" {
			exerciseTitle = "Unknown Exercise"
		}

		// Build submitter name
		submitterName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
		if submitterName == " " {
			submitterName = "Unknown User"
		}

		enrichedSubmissions = append(enrichedSubmissions, CoursePDFSubmission{
			Submission:    submission,
			ExerciseTitle: exerciseTitle,
			SubmitterName: submitterName,
		})
	}

	return enrichedSubmissions, nil
}

// GetPDFSubmission gets a specific PDF submission
func (s *PDFExerciseSubmissionService) GetPDFSubmission(submissionID string) (*models.Submission, error) {
	// Cleanup old submissions before querying (async, don't wait for result)
	go func() {
		if err := s.CleanupOldSubmissions(); err != nil {
			logger.Warnf("Failed to cleanup old submissions: %v", err)
		}
	}()

	var submission models.Submission
	if err := s.db.Where("submission_id = ?", submissionID).First(&submission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("submission not found")
		}
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}
	return &submission, nil
}

// DownloadPDFSubmission generates a presigned URL for downloading a PDF submission
func (s *PDFExerciseSubmissionService) DownloadPDFSubmission(submissionID string, expiration time.Duration) (string, error) {
	// Get submission
	var submission models.Submission
	if err := s.db.Where("submission_id = ?", submissionID).First(&submission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("submission not found")
		}
		return "", fmt.Errorf("failed to get submission: %w", err)
	}

	// Generate presigned download URL
	downloadURL, err := s.storageService.DownloadStudentPDFSubmission(
		context.Background(),
		submission.FileURL,
		expiration,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return downloadURL, nil
}

// StreamPDFSubmission streams a PDF submission file directly from MinIO
func (s *PDFExerciseSubmissionService) StreamPDFSubmission(submissionID string) (io.Reader, string, string, int64, error) {
	// Get submission
	var submission models.Submission
	if err := s.db.Where("submission_id = ?", submissionID).First(&submission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", "", 0, fmt.Errorf("submission not found")
		}
		return nil, "", "", 0, fmt.Errorf("failed to get submission: %w", err)
	}

	// Stream file from MinIO
	reader, contentType, size, err := s.storageService.StreamStudentPDFSubmission(
		context.Background(),
		submission.FileURL,
	)
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("failed to stream file: %w", err)
	}

	// Get filename from submission
	filename := submission.FileName
	if filename == "" {
		filename = "submission.pdf"
	}

	return reader, contentType, filename, size, nil
}

// StreamFeedbackFile streams a feedback PDF file directly from MinIO (only for submission owner)
func (s *PDFExerciseSubmissionService) StreamFeedbackFile(submissionID, userID string) (io.Reader, string, string, int64, error) {
	// Get submission
	var submission models.Submission
	if err := s.db.Where("submission_id = ?", submissionID).First(&submission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", "", 0, fmt.Errorf("submission not found")
		}
		return nil, "", "", 0, fmt.Errorf("failed to get submission: %w", err)
	}

	// Check if user owns this submission
	if submission.UserID != userID {
		return nil, "", "", 0, fmt.Errorf("unauthorized")
	}

	// Check if feedback file exists
	if submission.FeedbackFileURL == "" {
		return nil, "", "", 0, fmt.Errorf("feedback file not found")
	}

	// Stream feedback file from MinIO
	reader, contentType, size, err := s.storageService.StreamStudentPDFSubmission(
		context.Background(),
		submission.FeedbackFileURL,
	)
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("failed to stream file: %w", err)
	}

	// Extract filename from URL or use default
	filename := "feedback.pdf"
	if submission.FeedbackFileURL != "" {
		urlParts := strings.Split(submission.FeedbackFileURL, "/")
		if len(urlParts) > 0 {
			lastPart := urlParts[len(urlParts)-1]
			if lastPart != "" {
				filename = lastPart
			}
		}
	}

	return reader, contentType, filename, size, nil
}

// CancelPDFSubmission cancels a PDF submission (only by the submitter)
func (s *PDFExerciseSubmissionService) CancelPDFSubmission(submissionID, userID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get submission
		var submission models.Submission
		if err := tx.Where("submission_id = ? AND user_id = ?", submissionID, userID).First(&submission).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("submission not found or not owned by user")
			}
			return fmt.Errorf("failed to get submission: %w", err)
		}

		// Check if submission can be cancelled (only pending submissions)
		if submission.Status != enums.SubmissionPending {
			return fmt.Errorf("only pending submissions can be cancelled")
		}

		// Delete file from MinIO
		if submission.FileURL != "" {
			if err := s.storageService.DeleteFile(context.Background(), submission.FileURL); err != nil {
				// Log error but don't fail the transaction
				logger.Warnf("Failed to delete file from MinIO: %v", err)
			}
		}

		// Delete submission from database
		if err := tx.Delete(&submission).Error; err != nil {
			return fmt.Errorf("failed to delete submission: %w", err)
		}

		// Update student progress to in_progress
		var progress models.StudentProgress
		if err := tx.Where("user_id = ? AND material_id = ?", submission.UserID, submission.MaterialID).First(&progress).Error; err == nil {
			progress.Status = enums.ProgressInProgress
			if err := tx.Where("user_id = ? AND material_id = ?", submission.UserID, submission.MaterialID).Updates(&progress).Error; err != nil {
				return fmt.Errorf("failed to update progress: %w", err)
			}
		}

		return nil
	})
}

// updateCourseScore updates course score for a student
func (s *PDFExerciseSubmissionService) updateCourseScore(tx *gorm.DB, userID, courseID string) error {
	// Get all progress for the student in this course
	var progressList []models.StudentProgress
	if err := tx.Table("student_progress sp").
		Joins("INNER JOIN course_materials cm ON sp.material_id = cm.material_id").
		Where("sp.user_id = ? AND cm.course_id = ?", userID, courseID).
		Find(&progressList).Error; err != nil {
		return err
	}

	// Calculate total score
	totalScore := 0
	for _, progress := range progressList {
		totalScore += progress.Score
	}

	// Update or create course score
	var courseScore models.StudentCourseScore
	err := tx.Where("user_id = ? AND course_id = ?", userID, courseID).First(&courseScore).Error
	if err == gorm.ErrRecordNotFound {
		// Create new course score
		courseScore = models.StudentCourseScore{
			UserID:      userID,
			CourseID:    courseID,
			TotalScore:  totalScore,
			LastUpdated: time.Now(),
		}
		if err := tx.Create(&courseScore).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// Update existing course score
		courseScore.TotalScore = totalScore
		courseScore.LastUpdated = time.Now()
		if err := tx.Where("user_id = ? AND course_id = ?", userID, courseID).Updates(&courseScore).Error; err != nil {
			return err
		}
	}

	return nil
}
