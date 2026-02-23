package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/external"
	"github.com/Project-DSView/backend/go/pkg/logger"
	"gorm.io/gorm"
)

type QueueService struct {
	db                *gorm.DB
	rabbitMQ          *external.RabbitMQService
	userService       *UserService
	submissionService *SubmissionService // Optional: set after initialization to avoid circular dependency
}

// QueueJobData, QueueJobResult, and TestResult are now defined in internal/types/services.go

func NewQueueService(db *gorm.DB, rabbitMQ *external.RabbitMQService, userService *UserService) *QueueService {
	return &QueueService{
		db:          db,
		rabbitMQ:    rabbitMQ,
		userService: userService,
	}
}

// SetSubmissionService sets the submission service (called after initialization to avoid circular dependency)
func (s *QueueService) SetSubmissionService(submissionService *SubmissionService) {
	s.submissionService = submissionService
}

// GetDB returns the database instance (for handler access)
func (s *QueueService) GetDB() *gorm.DB {
	return s.db
}

// SubmitCodeExecutionJob submits a code execution job to the queue
func (s *QueueService) SubmitCodeExecutionJob(ctx context.Context, userID, materialID, submissionID, code, courseID string) (*models.QueueJob, error) {
	// Create job data
	jobData := types.QueueJobData{
		Code:         code,
		MaterialID:   materialID,
		SubmissionID: submissionID,
		FileName:     "submission.py",
		CourseID:     courseID,
	}

	dataJSON, err := json.Marshal(jobData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job data: %w", err)
	}

	// Create queue job record
	queueJob := &models.QueueJob{
		Type:       enums.QueueTypeCodeExecution,
		Status:     enums.QueueStatusPending,
		UserID:     userID,
		MaterialID: &materialID,
		CourseID:   &courseID,
		Data:       string(dataJSON),
	}

	if err := s.db.Create(queueJob).Error; err != nil {
		return nil, fmt.Errorf("failed to create queue job: %w", err)
	}

	// Publish to RabbitMQ using course-specific queue
	message := &external.QueueMessage{
		ID:   queueJob.ID,
		Type: string(enums.QueueTypeCodeExecution),
		Data: map[string]interface{}{"job_id": queueJob.ID},
	}

	if err := s.rabbitMQ.PublishMessage(context.Background(), string(enums.QueueTypeCodeExecution), courseID, message); err != nil {
		// Update job status to failed
		s.db.Model(&models.QueueJob{}).Where("id = ?", queueJob.ID).Updates(map[string]interface{}{
			"status": enums.QueueStatusFailed,
			"error":  fmt.Sprintf("Failed to publish to queue: %v", err),
		})
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	return queueJob, nil
}

// PublishFileProcessingJob publishes a file processing job message to RabbitMQ
func (s *QueueService) PublishFileProcessingJob(ctx context.Context, courseID string, message *external.QueueMessage) error {
	if s.rabbitMQ == nil {
		return fmt.Errorf("RabbitMQ service not available")
	}
	return s.rabbitMQ.PublishMessage(ctx, string(enums.QueueTypeFileProcessing), courseID, message)
}

// SubmitCodeReviewJob submits a code review job to the queue
func (s *QueueService) SubmitCodeReviewJob(ctx context.Context, userID, materialID, courseID, reviewNotes string) (*models.QueueJob, error) {
	// Create job data
	jobData := types.QueueJobData{
		MaterialID:  materialID,
		CourseID:    courseID,
		ReviewNotes: reviewNotes,
	}

	dataJSON, err := json.Marshal(jobData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job data: %w", err)
	}

	// Create queue job record
	queueJob := &models.QueueJob{
		Type:       enums.QueueTypeReview,
		Status:     enums.QueueStatusPending,
		UserID:     userID,
		MaterialID: &materialID,
		CourseID:   &courseID,
		Data:       string(dataJSON),
	}

	if err := s.db.Create(queueJob).Error; err != nil {
		return nil, fmt.Errorf("failed to create queue job: %w", err)
	}

	// Publish to RabbitMQ using course-specific queue
	message := &external.QueueMessage{
		ID:   queueJob.ID,
		Type: string(enums.QueueTypeReview),
		Data: map[string]interface{}{"job_id": queueJob.ID},
	}

	if err := s.rabbitMQ.PublishMessage(context.Background(), string(enums.QueueTypeReview), courseID, message); err != nil {
		// Update job status to failed
		s.db.Model(&models.QueueJob{}).Where("id = ?", queueJob.ID).Updates(map[string]interface{}{
			"status": enums.QueueStatusFailed,
			"error":  fmt.Sprintf("Failed to publish to queue: %v", err),
		})
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	return queueJob, nil
}

// SubmitReviewJobWithLabTable submits a review job with lab and table selection (for both code and PDF)
func (s *QueueService) SubmitReviewJobWithLabTable(userID, materialID, courseID, submissionID, labRoom, tableNumber string, jobData types.QueueJobData) (*models.QueueJob, error) {
	// Set additional fields in job data
	jobData.MaterialID = materialID
	jobData.CourseID = courseID
	jobData.SubmissionID = submissionID
	jobData.LabRoom = labRoom
	jobData.TableNumber = tableNumber

	dataJSON, err := json.Marshal(jobData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job data: %w", err)
	}

	// Create queue job record
	queueJob := &models.QueueJob{
		Type:         enums.QueueTypeReview,
		Status:       enums.QueueStatusPending,
		UserID:       userID,
		MaterialID:   &materialID,
		CourseID:     &courseID,
		SubmissionID: &submissionID,
		LabRoom:      &labRoom,
		TableNumber:  &tableNumber,
		Data:         string(dataJSON),
	}

	if err := s.db.Create(queueJob).Error; err != nil {
		return nil, fmt.Errorf("failed to create queue job: %w", err)
	}

	// Publish to RabbitMQ
	message := &external.QueueMessage{
		ID:   queueJob.ID,
		Type: string(enums.QueueTypeReview),
		Data: map[string]interface{}{"job_id": queueJob.ID},
	}

	if err := s.rabbitMQ.PublishMessage(context.Background(), string(enums.QueueTypeReview), courseID, message); err != nil {
		// Update job status to failed
		s.db.Model(&models.QueueJob{}).Where("id = ?", queueJob.ID).Updates(map[string]interface{}{
			"status": enums.QueueStatusFailed,
			"error":  fmt.Sprintf("Failed to publish to queue: %v", err),
		})
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	return queueJob, nil
}

// GetQueueJobs retrieves queue jobs with filtering
func (s *QueueService) GetQueueJobs(queueType string, status string, courseID string, userID string, isTeacher bool, page, limit int) ([]models.QueueJob, int, error) {
	return s.GetQueueJobsWithDateFilter(queueType, status, courseID, userID, isTeacher, page, limit, "", "")
}

// GetQueueJobsWithDateFilter retrieves queue jobs with filtering including date range
func (s *QueueService) GetQueueJobsWithDateFilter(queueType string, status string, courseID string, userID string, isTeacher bool, page, limit int, fromDate, toDate string) ([]models.QueueJob, int, error) {
	// Cleanup old queue jobs before querying (async, don't wait for result)
	go func() {
		if err := s.CleanupOldQueueJobs(); err != nil {
			logger.Warnf("Failed to cleanup old queue jobs: %v", err)
		}
	}()

	var jobs []models.QueueJob
	var total int64

	query := s.db.Model(models.QueueJob{})

	// Filter by queue type
	if queueType != "" && models.IsValidQueueType(queueType) {
		query = query.Where("type = ?", queueType)
	}

	// Filter by status
	if status != "" && models.IsValidQueueStatus(status) {
		query = query.Where("status = ?", status)
	}

	// Filter by today's date (based on created_at) - always apply
	// Use Thailand timezone (UTC+7)
	thailandLocation, _ := time.LoadLocation("Asia/Bangkok")
	today := time.Now().In(thailandLocation)
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, thailandLocation)
	todayEnd := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, thailandLocation)
	query = query.Where("queue_jobs.created_at >= ? AND queue_jobs.created_at <= ?", todayStart, todayEnd)

	// Filter by code exercise only - join with course_materials to check material type
	query = query.Joins("LEFT JOIN course_materials cm ON queue_jobs.material_id = cm.material_id")
	query = query.Where("cm.type = ?", string(enums.MaterialTypeCodeExercise))

	// Apply additional date filtering if provided (for submission date)
	if fromDate != "" || toDate != "" {
		// Join with submissions table to filter by submission date
		query = query.Joins("LEFT JOIN submissions ON queue_jobs.submission_id = submissions.submission_id")

		if fromDate != "" {
			query = query.Where("submissions.submitted_at >= ?", fromDate)
		}
		if toDate != "" {
			query = query.Where("submissions.submitted_at <= ?", toDate)
		}
	}

	// Apply role-based filtering
	if courseID != "" {
		// If specific course_id is provided, use it
		query = query.Where("queue_jobs.course_id = ?", courseID)
	} else {
		// Apply role-based default filtering
		if isTeacher {
			// Teachers see only jobs from courses they created
			courseIDs, err := s.GetCoursesByCreator(userID)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to get courses by creator: %w", err)
			}
			if len(courseIDs) > 0 {
				query = query.Where("queue_jobs.course_id IN ?", courseIDs)
			} else {
				// No courses created by this teacher, return empty result
				return []models.QueueJob{}, 0, nil
			}
		} else {
			// Check if user is TA in any courses
			taCourseIDs, err := s.GetCoursesByTA(userID)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to get courses by TA: %w", err)
			}
			if len(taCourseIDs) > 0 {
				// TAs see jobs from courses they're enrolled as TA + their own jobs
				query = query.Where("queue_jobs.course_id IN ? OR queue_jobs.user_id = ?", taCourseIDs, userID)
			} else {
				// Not a TA, only see own jobs
				query = query.Where("queue_jobs.user_id = ?", userID)
			}
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count queue jobs: %w", err)
	}

	// Apply pagination
	offset := (page - 1) * limit
	if err := query.
		Order("queue_jobs.created_at ASC").
		Offset(offset).Limit(limit).Find(&jobs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get queue jobs: %w", err)
	}

	// Load relations manually (since they use gorm:"-")
	for i := range jobs {
		// Load User
		if jobs[i].UserID != "" {
			var user models.User
			if err := s.db.First(&user, "user_id = ?", jobs[i].UserID).Error; err == nil {
				jobs[i].User = &user
			}
		}

		// Load CourseMaterial
		if jobs[i].MaterialID != nil && *jobs[i].MaterialID != "" {
			var material models.CourseMaterial
			if err := s.db.First(&material, "material_id = ?", *jobs[i].MaterialID).Error; err == nil {
				jobs[i].CourseMaterial = &material
			}
		}

		// Load Course
		if jobs[i].CourseID != nil && *jobs[i].CourseID != "" {
			var course models.Course
			if err := s.db.First(&course, "course_id = ?", *jobs[i].CourseID).Error; err == nil {
				jobs[i].Course = &course
			}
		}

		// Load ProcessedByUser
		if jobs[i].ProcessedBy != nil && *jobs[i].ProcessedBy != "" {
			var processedByUser models.User
			if err := s.db.First(&processedByUser, "user_id = ?", *jobs[i].ProcessedBy).Error; err == nil {
				jobs[i].ProcessedByUser = &processedByUser
			}
		}

		// Load review_status from submission if submission_id exists
		if jobs[i].SubmissionID != nil && *jobs[i].SubmissionID != "" {
			var submission models.Submission
			if err := s.db.First(&submission, "submission_id = ?", *jobs[i].SubmissionID).Error; err == nil {
				if submission.ReviewStatus != "" {
					reviewStatus := submission.ReviewStatus
					jobs[i].ReviewStatus = &reviewStatus
				}
			}
		}
	}

	return jobs, int(total), nil
}

// GetQueueJobsByDateRange retrieves queue jobs for a specific course within a date range
func (s *QueueService) GetQueueJobsByDateRange(courseID, startDate, endDate string) ([]models.QueueJob, error) {
	var jobs []models.QueueJob

	query := s.db.Model(models.QueueJob{}).
		Joins("LEFT JOIN submissions ON queue_jobs.submission_id = submissions.submission_id").
		Where("queue_jobs.course_id = ?", courseID)

	if startDate != "" {
		query = query.Where("submissions.submitted_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("submissions.submitted_at <= ?", endDate)
	}

	if err := query.
		Order("queue_jobs.created_at ASC").Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("failed to get queue jobs by date range: %w", err)
	}

	// Load relations manually (since they use gorm:"-")
	for i := range jobs {
		// Load User
		if jobs[i].UserID != "" {
			var user models.User
			if err := s.db.First(&user, "user_id = ?", jobs[i].UserID).Error; err == nil {
				jobs[i].User = &user
			}
		}

		// Load CourseMaterial
		if jobs[i].MaterialID != nil && *jobs[i].MaterialID != "" {
			var material models.CourseMaterial
			if err := s.db.First(&material, "material_id = ?", *jobs[i].MaterialID).Error; err == nil {
				jobs[i].CourseMaterial = &material
			}
		}

		// Load Course
		if jobs[i].CourseID != nil && *jobs[i].CourseID != "" {
			var course models.Course
			if err := s.db.First(&course, "course_id = ?", *jobs[i].CourseID).Error; err == nil {
				jobs[i].Course = &course
			}
		}

		// Load ProcessedByUser
		if jobs[i].ProcessedBy != nil && *jobs[i].ProcessedBy != "" {
			var processedByUser models.User
			if err := s.db.First(&processedByUser, "user_id = ?", *jobs[i].ProcessedBy).Error; err == nil {
				jobs[i].ProcessedByUser = &processedByUser
			}
		}
	}

	return jobs, nil
}

// GetQueueJobByID retrieves a specific queue job
func (s *QueueService) GetQueueJobByID(jobID string) (*models.QueueJob, error) {
	var job models.QueueJob
	if err := s.db.First(&job, "id = ?", jobID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get queue job: %w", err)
	}

	// Load relations manually (since they use gorm:"-")
	// Load User
	if job.UserID != "" {
		var user models.User
		if err := s.db.First(&user, "user_id = ?", job.UserID).Error; err == nil {
			job.User = &user
		}
	}

	// Load CourseMaterial
	if job.MaterialID != nil && *job.MaterialID != "" {
		var material models.CourseMaterial
		if err := s.db.First(&material, "material_id = ?", *job.MaterialID).Error; err == nil {
			job.CourseMaterial = &material
		}
	}

	// Load Course
	if job.CourseID != nil && *job.CourseID != "" {
		var course models.Course
		if err := s.db.First(&course, "course_id = ?", *job.CourseID).Error; err == nil {
			job.Course = &course
		}
	}

	// Load ProcessedByUser
	if job.ProcessedBy != nil && *job.ProcessedBy != "" {
		var processedByUser models.User
		if err := s.db.First(&processedByUser, "user_id = ?", *job.ProcessedBy).Error; err == nil {
			job.ProcessedByUser = &processedByUser
		}
	}

	return &job, nil
}

// UpdateJobStatus updates the status of a queue job
func (s *QueueService) UpdateJobStatus(jobID string, status enums.QueueStatus, processedBy string, result *types.QueueJobResult, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if processedBy != "" {
		updates["processed_by"] = processedBy
	}

	if result != nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		updates["result"] = string(resultJSON)
	}

	if errorMsg != "" {
		updates["error"] = errorMsg
	}

	now := time.Now()
	switch status {
	case enums.QueueStatusProcessing:
		updates["started_at"] = &now
	case enums.QueueStatusCompleted, enums.QueueStatusFailed, enums.QueueStatusCancelled:
		updates["completed_at"] = &now
	}

	if err := s.db.Model(&models.QueueJob{}).Where("id = ?", jobID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// IsUserTAInCourse checks if a user is a TA in a specific course
func (s *QueueService) IsUserTAInCourse(courseID, userID string) (bool, error) {
	var count int64
	err := s.db.Model(&models.Enrollment{}).
		Where("course_id = ? AND user_id = ? AND role = ?", courseID, userID, enums.EnrollmentRoleTA).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check TA status: %w", err)
	}
	return count > 0, nil
}

// CancelJob cancels a pending job
func (s *QueueService) CancelJob(jobID, userID string, isTeacher bool) error {
	var job models.QueueJob
	if err := s.db.First(&job, "id = ?", jobID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("job not found")
		}
		return fmt.Errorf("failed to get job: %w", err)
	}

	// Check permissions
	if !isTeacher && job.UserID != userID {
		return fmt.Errorf("permission denied")
	}

	// Only allow cancelling pending jobs
	if job.Status != enums.QueueStatusPending {
		return fmt.Errorf("can only cancel pending jobs")
	}

	return s.UpdateJobStatus(jobID, enums.QueueStatusCancelled, "", nil, "")
}

// GetQueueStats returns statistics about the queue
func (s *QueueService) GetQueueStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count by status
	var statusCounts []struct {
		Status string
		Count  int
	}
	if err := s.db.Model(&models.QueueJob{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	statusMap := make(map[string]int)
	for _, sc := range statusCounts {
		statusMap[sc.Status] = sc.Count
	}
	stats["by_status"] = statusMap

	// Count by type
	var typeCounts []struct {
		Type  string
		Count int
	}
	if err := s.db.Model(&models.QueueJob{}).
		Select("type, count(*) as count").
		Group("type").
		Scan(&typeCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get type counts: %w", err)
	}

	typeMap := make(map[string]int)
	for _, tc := range typeCounts {
		typeMap[tc.Type] = tc.Count
	}
	stats["by_type"] = typeMap

	// Total count
	var total int64
	if err := s.db.Model(&models.QueueJob{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total"] = total

	return stats, nil
}

// StartQueueConsumer starts consuming messages from course-specific queues
func (s *QueueService) StartQueueConsumer(ctx context.Context) error {
	// Get all active course IDs
	courseIDs, err := s.GetActiveCourseIDs()
	if err != nil {
		return fmt.Errorf("failed to get active course IDs: %w", err)
	}

	if len(courseIDs) == 0 {
		logger.Info("No active courses found, queue consumers will not be started")
		return nil
	}

	// Start consumers for each course
	for _, courseID := range courseIDs {
		// Start code execution consumer for this course
		if err := s.rabbitMQ.ConsumeMessages(ctx, string(enums.QueueTypeCodeExecution), courseID, s.handleCodeExecutionMessage); err != nil {
			logger.Warnf("Failed to start code execution consumer for course %s: %v", courseID, err)
			continue
		}

		// Start code review consumer for this course
		if err := s.rabbitMQ.ConsumeMessages(ctx, string(enums.QueueTypeReview), courseID, s.handleCodeReviewMessage); err != nil {
			logger.Warnf("Failed to start code review consumer for course %s: %v", courseID, err)
			continue
		}

		// Start file processing consumer for this course
		if err := s.rabbitMQ.ConsumeMessages(ctx, string(enums.QueueTypeFileProcessing), courseID, s.handleFileProcessingMessage); err != nil {
			logger.Warnf("Failed to start file processing consumer for course %s: %v", courseID, err)
			continue
		}

		logger.Infof("Started queue consumers for course: %s", courseID)
	}

	logger.Infof("Started queue consumers for %d courses", len(courseIDs))
	return nil
}

// handleCodeExecutionMessage processes code execution messages
func (s *QueueService) handleCodeExecutionMessage(msg *external.QueueMessage) error {
	jobID, ok := msg.Data["job_id"].(string)
	if !ok {
		return fmt.Errorf("invalid job_id in message")
	}

	// Get job from database
	job, err := s.GetQueueJobByID(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Update status to processing
	if err := s.UpdateJobStatus(jobID, enums.QueueStatusProcessing, "", nil, ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Parse job data
	var jobData types.QueueJobData
	if err := json.Unmarshal([]byte(job.Data), &jobData); err != nil {
		s.UpdateJobStatus(jobID, enums.QueueStatusFailed, "", nil, fmt.Sprintf("Failed to parse job data: %v", err))
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	// Check if submission service is available
	if s.submissionService == nil {
		s.UpdateJobStatus(jobID, enums.QueueStatusFailed, "", nil, "Submission service not available")
		return fmt.Errorf("submission service not available")
	}

	// Execute code submission
	if err := s.submissionService.ExecuteCodeSubmission(jobData.SubmissionID, jobData.Code, jobData.MaterialID); err != nil {
		s.UpdateJobStatus(jobID, enums.QueueStatusFailed, "", nil, fmt.Sprintf("Code execution failed: %v", err))
		return fmt.Errorf("code execution failed: %w", err)
	}

	// Update job status to completed
	if err := s.UpdateJobStatus(jobID, enums.QueueStatusCompleted, "", nil, ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// handleCodeReviewMessage processes code review messages
func (s *QueueService) handleCodeReviewMessage(msg *external.QueueMessage) error {
	jobID, ok := msg.Data["job_id"].(string)
	if !ok {
		return fmt.Errorf("invalid job_id in message")
	}

	// Get job from database
	job, err := s.GetQueueJobByID(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Update status to processing
	if err := s.UpdateJobStatus(jobID, enums.QueueStatusProcessing, "", nil, ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Parse job data
	var jobData types.QueueJobData
	if err := json.Unmarshal([]byte(job.Data), &jobData); err != nil {
		s.UpdateJobStatus(jobID, enums.QueueStatusFailed, "", nil, fmt.Sprintf("Failed to parse job data: %v", err))
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	// TODO: Implement actual code review logic here
	// For now, simulate processing
	time.Sleep(1 * time.Second)

	// Create mock result
	result := &types.QueueJobResult{
		Success: true,
		Output:  "Code review completed",
		Metadata: map[string]interface{}{
			"review_notes": jobData.ReviewNotes,
			"reviewer":     "system",
		},
	}

	// Update job status to completed
	if err := s.UpdateJobStatus(jobID, enums.QueueStatusCompleted, "", result, ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// handleFileProcessingMessage processes file processing messages
func (s *QueueService) handleFileProcessingMessage(msg *external.QueueMessage) error {
	jobID, ok := msg.Data["job_id"].(string)
	if !ok {
		return fmt.Errorf("invalid job_id in message")
	}

	// Get job from database
	job, err := s.GetQueueJobByID(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Update status to processing
	if err := s.UpdateJobStatus(jobID, enums.QueueStatusProcessing, "", nil, ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Parse job data
	var jobData types.QueueJobData
	if err := json.Unmarshal([]byte(job.Data), &jobData); err != nil {
		s.UpdateJobStatus(jobID, enums.QueueStatusFailed, "", nil, fmt.Sprintf("Failed to parse job data: %v", err))
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	// Check if submission service is available
	if s.submissionService == nil {
		s.UpdateJobStatus(jobID, enums.QueueStatusFailed, "", nil, "Submission service not available")
		return fmt.Errorf("submission service not available")
	}

	// Process file based on type
	if err := s.submissionService.ProcessFile(jobData); err != nil {
		s.UpdateJobStatus(jobID, enums.QueueStatusFailed, "", nil, fmt.Sprintf("File processing failed: %v", err))
		return fmt.Errorf("file processing failed: %w", err)
	}

	// Update job status to completed
	if err := s.UpdateJobStatus(jobID, enums.QueueStatusCompleted, "", nil, ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// ClaimJob claims a pending queue job for a TA/Teacher
func (s *QueueService) ClaimJob(jobID, userID string) (*models.QueueJob, error) {
	var job models.QueueJob
	if err := s.db.First(&job, "id = ?", jobID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Check if job is pending
	if job.Status != enums.QueueStatusPending {
		return nil, fmt.Errorf("job not pending")
	}

	// Check if job is already claimed
	if job.ProcessedBy != nil && *job.ProcessedBy != "" {
		return nil, fmt.Errorf("job already claimed")
	}

	// Claim the job
	now := time.Now()
	updates := map[string]interface{}{
		"status":       enums.QueueStatusProcessing,
		"processed_by": userID,
		"claimed_at":   &now,
	}

	if err := s.db.Model(&models.QueueJob{}).Where("id = ?", jobID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to claim job: %w", err)
	}

	// Reload job with updated data
	if err := s.db.First(&job, "id = ?", jobID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload job: %w", err)
	}

	return &job, nil
}

// CompleteReview completes a review with approval/rejection decision
func (s *QueueService) CompleteReview(jobID, userID, status, comment string) (*models.QueueJob, error) {
	var job models.QueueJob
	if err := s.db.First(&job, "id = ?", jobID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Check if job is claimed by the user
	if job.ProcessedBy == nil || *job.ProcessedBy != userID {
		return nil, fmt.Errorf("job not claimed by user")
	}

	// Check if job is in processing status
	if job.Status != enums.QueueStatusProcessing {
		return nil, fmt.Errorf("job not in processing status")
	}

	// Complete the review
	now := time.Now()
	updates := map[string]interface{}{
		"status":       enums.QueueStatusCompleted,
		"completed_at": &now,
	}

	if err := s.db.Model(&models.QueueJob{}).Where("id = ?", jobID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to complete job: %w", err)
	}

	// Update progress based on review decision
	if err := s.updateProgressFromReview(&job, status, comment, userID); err != nil {
		// Log error but don't fail the job completion
		logger.Warnf("Failed to update progress from review: %v", err)
	}

	// Reload job with updated data
	if err := s.db.First(&job, "id = ?", jobID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload job: %w", err)
	}

	return &job, nil
}

// updateProgressFromReview updates student progress based on review decision
func (s *QueueService) updateProgressFromReview(job *models.QueueJob, status, comment, reviewerID string) error {
	if job.MaterialID == nil || job.UserID == "" {
		return fmt.Errorf("missing material ID or user ID")
	}

	// Update submission feedback if submission_id exists
	if job.SubmissionID != nil && *job.SubmissionID != "" {
		now := time.Now()
		// Store review_status as 'approved' or 'rejected'
		reviewStatus := "approved"
		if status == "rejected" {
			reviewStatus = "rejected"
		}
		updates := map[string]interface{}{
			"feedback":      comment,
			"graded_by":     reviewerID,
			"graded_at":     &now,
			"review_status": reviewStatus,
		}
		if err := s.db.Model(&models.Submission{}).
			Where("submission_id = ?", *job.SubmissionID).
			Updates(updates).Error; err != nil {
			// Log error but don't fail the review
			logger.Warnf("Failed to update submission feedback: %v", err)
		}
	}

	// Get progress
	var progress models.StudentProgress
	if err := s.db.Where("user_id = ? AND material_id = ?", job.UserID, *job.MaterialID).
		First(&progress).Error; err != nil {
		return fmt.Errorf("get progress: %w", err)
	}

	// Create verification log
	verificationStatus := enums.VerificationApproved
	if status == "rejected" {
		verificationStatus = enums.VerificationRejected
	}

	log := models.VerificationLog{
		ProgressID: progress.ProgressID,
		VerifiedBy: reviewerID,
		Status:     verificationStatus,
		Comment:    comment,
	}

	if err := s.db.Create(&log).Error; err != nil {
		return fmt.Errorf("create verification log: %w", err)
	}

	// Update progress status
	var newStatus enums.ProgressStatus
	if status == "approved" {
		newStatus = enums.ProgressCompleted
	} else {
		// Rejected - reset to not_started but keep submission history
		newStatus = enums.ProgressNotStarted
		// Reset score to 0
		if err := s.db.Model(&models.StudentProgress{}).
			Where("progress_id = ?", progress.ProgressID).
			Updates(map[string]interface{}{
				"status": newStatus,
				"score":  0,
			}).Error; err != nil {
			return fmt.Errorf("reset progress: %w", err)
		}
		return nil
	}

	// Update progress for approval
	if err := s.db.Model(&models.StudentProgress{}).
		Where("progress_id = ?", progress.ProgressID).
		Update("status", newStatus).Error; err != nil {
		return fmt.Errorf("update progress: %w", err)
	}

	return nil
}

// GetActiveCourseIDs returns all active course IDs from the database
func (s *QueueService) GetActiveCourseIDs() ([]string, error) {
	var courseIDs []string
	err := s.db.Model(&models.Course{}).
		Where("status = ?", "active").
		Pluck("course_id", &courseIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active course IDs: %w", err)
	}
	return courseIDs, nil
}

// GetCoursesByCreator returns course IDs created by the specified user
func (s *QueueService) GetCoursesByCreator(userID string) ([]string, error) {
	var courseIDs []string
	err := s.db.Model(&models.Course{}).
		Where("created_by = ?", userID).
		Pluck("course_id", &courseIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get courses by creator: %w", err)
	}
	return courseIDs, nil
}

// GetCoursesByTA returns course IDs where the user is enrolled as TA
func (s *QueueService) GetCoursesByTA(userID string) ([]string, error) {
	var courseIDs []string
	err := s.db.Model(&models.Enrollment{}).
		Where("user_id = ? AND role = ?", userID, enums.EnrollmentRoleTA).
		Pluck("course_id", &courseIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get courses by TA: %w", err)
	}
	return courseIDs, nil
}

// CanRetryQueueJob checks if a queue job can be retried (must be >1 day old and belong to user)
func (s *QueueService) CanRetryQueueJob(jobID, userID string) (bool, error) {
	// Get the original job
	job, err := s.GetQueueJobByID(jobID)
	if err != nil {
		return false, fmt.Errorf("failed to get queue job: %w", err)
	}
	if job == nil {
		return false, fmt.Errorf("queue job not found")
	}

	// Check if job belongs to user
	if job.UserID != userID {
		return false, fmt.Errorf("queue job does not belong to user")
	}

	// Get the submission to check submission date
	var submission models.Submission
	if job.SubmissionID != nil {
		err := s.db.Where("submission_id = ?", *job.SubmissionID).First(&submission).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return false, fmt.Errorf("submission not found")
			}
			return false, fmt.Errorf("failed to get submission: %w", err)
		}

		// Check if submission is more than 1 day old
		oneDayAgo := time.Now().Add(-24 * time.Hour)
		if submission.SubmittedAt.After(oneDayAgo) {
			return false, fmt.Errorf("submission is less than 1 day old, cannot retry yet")
		}
	} else {
		// If no submission, check queue job creation date
		oneDayAgo := time.Now().Add(-24 * time.Hour)
		if job.CreatedAt.After(oneDayAgo) {
			return false, fmt.Errorf("queue job is less than 1 day old, cannot retry yet")
		}
	}

	return true, nil
}

// CreateRetryQueueJob creates a new queue job based on the original job
func (s *QueueService) CreateRetryQueueJob(originalJobID, userID string) (*models.QueueJob, error) {
	// Validate retry eligibility
	canRetry, err := s.CanRetryQueueJob(originalJobID, userID)
	if err != nil {
		return nil, fmt.Errorf("retry validation failed: %w", err)
	}
	if !canRetry {
		return nil, fmt.Errorf("queue job cannot be retried")
	}

	// Get the original job
	originalJob, err := s.GetQueueJobByID(originalJobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original queue job: %w", err)
	}
	if originalJob == nil {
		return nil, fmt.Errorf("original queue job not found")
	}

	// Create new queue job with same data
	newJob := &models.QueueJob{
		Type:         originalJob.Type,
		Status:       enums.QueueStatusPending,
		UserID:       userID,
		MaterialID:   originalJob.MaterialID,
		CourseID:     originalJob.CourseID,
		SubmissionID: originalJob.SubmissionID,
		LabRoom:      originalJob.LabRoom,
		TableNumber:  originalJob.TableNumber,
		Data:         originalJob.Data,
		Result:       "",  // Reset result
		Error:        "",  // Reset error
		ProcessedBy:  nil, // Reset processed by
		ClaimedAt:    nil, // Reset claimed at
		StartedAt:    nil, // Reset started at
		CompletedAt:  nil, // Reset completed at
	}

	// Save new job
	if err := s.db.Create(newJob).Error; err != nil {
		return nil, fmt.Errorf("failed to create retry queue job: %w", err)
	}

	// Submit to RabbitMQ if configured
	if s.rabbitMQ != nil && newJob.CourseID != nil {
		jobData := map[string]interface{}{
			"job_id": newJob.ID,
			"type":   string(newJob.Type),
		}

		message := &external.QueueMessage{
			ID:        newJob.ID,
			Type:      string(newJob.Type),
			Data:      jobData,
			CreatedAt: time.Now(),
		}

		if err := s.rabbitMQ.PublishMessage(context.Background(), string(newJob.Type), *newJob.CourseID, message); err != nil {
			// Log error but don't fail the operation
			logger.Warnf("Failed to publish retry job to RabbitMQ: %v", err)
		}
	}

	return newJob, nil
}

// CleanupCourseQueues deletes all queues for a specific course
// This should be called when a course is deleted or archived
func (s *QueueService) CleanupCourseQueues(courseID string) error {
	if s.rabbitMQ == nil {
		return fmt.Errorf("RabbitMQ service not available")
	}

	// Check if there are any pending jobs for this course
	var pendingJobsCount int64
	if err := s.db.Model(&models.QueueJob{}).
		Where("course_id = ? AND status IN ?", courseID, []string{"pending", "processing"}).
		Count(&pendingJobsCount).Error; err != nil {
		return fmt.Errorf("failed to check pending jobs: %w", err)
	}

	if pendingJobsCount > 0 {
		logger.Warnf("Cannot cleanup queues for course %s: %d pending/processing jobs exist", courseID, pendingJobsCount)
		return fmt.Errorf("cannot cleanup queues: %d pending/processing jobs exist", pendingJobsCount)
	}

	// Delete all queues for this course
	if err := s.rabbitMQ.DeleteCourseQueues(courseID); err != nil {
		return fmt.Errorf("failed to delete course queues: %w", err)
	}

	logger.Infof("Cleaned up queues for course: %s", courseID)
	return nil
}

// CleanupOrphanedQueues finds and deletes queues that don't have corresponding active courses
func (s *QueueService) CleanupOrphanedQueues() error {
	if s.rabbitMQ == nil {
		return fmt.Errorf("RabbitMQ service not available")
	}

	// Get all active course IDs
	activeCourseIDs, err := s.GetActiveCourseIDs()
	if err != nil {
		return fmt.Errorf("failed to get active course IDs: %w", err)
	}

	// Create a map for quick lookup
	activeCourseMap := make(map[string]bool)
	for _, courseID := range activeCourseIDs {
		activeCourseMap[courseID] = true
	}

	// Get all queue jobs with course IDs
	var jobsWithCourses []struct {
		CourseID string
	}
	if err := s.db.Model(&models.QueueJob{}).
		Select("DISTINCT course_id").
		Where("course_id IS NOT NULL").
		Scan(&jobsWithCourses).Error; err != nil {
		return fmt.Errorf("failed to get course IDs from queue jobs: %w", err)
	}

	// Find orphaned courses (courses that have queue jobs but are not active)
	orphanedCourseIDs := make(map[string]bool)
	for _, job := range jobsWithCourses {
		if job.CourseID != "" && !activeCourseMap[job.CourseID] {
			// Check if there are any pending/processing jobs
			var pendingCount int64
			if err := s.db.Model(&models.QueueJob{}).
				Where("course_id = ? AND status IN ?", job.CourseID, []string{"pending", "processing"}).
				Count(&pendingCount).Error; err != nil {
				logger.Warnf("Failed to check pending jobs for course %s: %v", job.CourseID, err)
				continue
			}

			if pendingCount == 0 {
				orphanedCourseIDs[job.CourseID] = true
			}
		}
	}

	// Delete queues for orphaned courses
	deletedCount := 0
	for courseID := range orphanedCourseIDs {
		if err := s.rabbitMQ.DeleteCourseQueues(courseID); err != nil {
			logger.Warnf("Failed to delete queues for orphaned course %s: %v", courseID, err)
			continue
		}
		deletedCount++
		logger.Infof("Deleted orphaned queues for course: %s", courseID)
	}

	logger.Infof("Cleanup completed: deleted queues for %d orphaned courses", deletedCount)
	return nil
}

// CleanupOldQueueJobs deletes queue jobs that are older than today (based on created_at)
func (s *QueueService) CleanupOldQueueJobs() error {
	// Use Thailand timezone (UTC+7)
	thailandLocation, _ := time.LoadLocation("Asia/Bangkok")
	today := time.Now().In(thailandLocation)
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, thailandLocation)

	// Find old queue jobs (created_at < today)
	// Only delete completed, failed, or cancelled jobs (not pending/processing)
	var oldJobs []models.QueueJob
	if err := s.db.Where("created_at < ? AND status IN ?", todayStart, []string{
		string(enums.QueueStatusCompleted),
		string(enums.QueueStatusFailed),
		string(enums.QueueStatusCancelled),
	}).Find(&oldJobs).Error; err != nil {
		return fmt.Errorf("failed to find old queue jobs: %w", err)
	}

	deletedCount := 0
	for _, job := range oldJobs {
		// Delete queue job
		if err := s.db.Delete(&job).Error; err != nil {
			logger.Warnf("Failed to delete queue job %s: %v", job.ID, err)
			continue
		}

		deletedCount++
	}

	if deletedCount > 0 {
		logger.Infof("Cleaned up %d old queue jobs", deletedCount)
	}

	return nil
}
