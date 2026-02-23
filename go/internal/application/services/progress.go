package services

import (
	"fmt"
	"time"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"gorm.io/gorm"
)

type ProgressService struct {
	db            *gorm.DB
	userService   *UserService
	courseService *CourseService
}

func NewProgressService(db *gorm.DB, userSvc *UserService, courseSvc *CourseService) *ProgressService {
	return &ProgressService{
		db:            db,
		userService:   userSvc,
		courseService: courseSvc,
	}
}

func (s *ProgressService) GetSelfProgress(userID, courseID string) ([]map[string]interface{}, error) {
	var results []struct {
		ProgressID      string     `gorm:"column:progress_id"`
		UserID          string     `gorm:"column:user_id"`
		MaterialID      string     `gorm:"column:material_id"`
		MaterialTitle   string     `gorm:"column:material_title"`
		Status          string     `gorm:"column:status"`
		Score           int        `gorm:"column:score"`
		SeatNumber      string     `gorm:"column:seat_number"`
		LastSubmittedAt *time.Time `gorm:"column:last_submitted_at"`
		ReviewStatus    *string    `gorm:"column:review_status"`
	}

	// Single optimized query with JOIN to avoid N+1 problem
	// Get title from the actual material tables using COALESCE and CASE
	// Get review_status from latest submission using subquery
	query := s.db.Table("student_progress sp").
		Select(`
			sp.progress_id,
			sp.user_id,
			sp.material_id,
			COALESCE(
				CASE WHEN cm.reference_type = 'code_exercise' THEN ce.title END,
				CASE WHEN cm.reference_type = 'pdf_exercise' THEN pe.title END,
				CASE WHEN cm.reference_type = 'document' THEN d.title END,
				CASE WHEN cm.reference_type = 'video' THEN v.title END,
				CASE WHEN cm.reference_type = 'announcement' THEN a.title END,
				'ไม่ระบุ'
			) as material_title,
			sp.status,
			sp.score,
			sp.seat_number,
			sp.last_submitted_at,
			(
				SELECT review_status 
				FROM submissions 
				WHERE submissions.user_id = sp.user_id 
				AND submissions.material_id = sp.material_id 
				ORDER BY submitted_at DESC 
				LIMIT 1
			) as review_status
		`).
		Joins("LEFT JOIN course_materials cm ON sp.material_id = cm.material_id").
		Joins("LEFT JOIN code_exercises ce ON cm.reference_id = ce.material_id AND cm.reference_type = 'code_exercise'").
		Joins("LEFT JOIN pdf_exercises pe ON cm.reference_id = pe.material_id AND cm.reference_type = 'pdf_exercise'").
		Joins("LEFT JOIN documents d ON cm.reference_id = d.material_id AND cm.reference_type = 'document'").
		Joins("LEFT JOIN videos v ON cm.reference_id = v.material_id AND cm.reference_type = 'video'").
		Joins("LEFT JOIN announcements a ON cm.reference_id = a.material_id AND cm.reference_type = 'announcement'").
		Where("sp.user_id = ?", userID)

	if courseID != "" {
		query = query.Where("cm.course_id = ?", courseID)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("find progress: %w", err)
	}

	out := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		result := map[string]interface{}{
			"progress_id":       r.ProgressID,
			"user_id":           r.UserID,
			"material_id":       r.MaterialID,
			"material_title":    r.MaterialTitle,
			"status":            r.Status,
			"score":             r.Score,
			"seat_number":       r.SeatNumber,
			"last_submitted_at": r.LastSubmittedAt,
		}
		// Add review_status if it exists
		if r.ReviewStatus != nil {
			result["review_status"] = *r.ReviewStatus
		}
		out = append(out, result)
	}
	return out, nil
}

// CourseProgressRow is now defined in internal/types/services.go

func (s *ProgressService) GetCourseProgress(courseID string) ([]map[string]interface{}, error) {
	// Single optimized query to get all data at once
	var results []struct {
		UserID         string     `gorm:"column:user_id"`
		UserName       string     `gorm:"column:user_name"`
		UserEmail      string     `gorm:"column:user_email"`
		SeatNumber     string     `gorm:"column:seat_number"`
		Completed      int        `gorm:"column:completed"`
		ScoreSum       int        `gorm:"column:score_sum"`
		ProgressCount  int        `gorm:"column:progress_count"`
		LastActivity   *time.Time `gorm:"column:last_activity"`
		TotalMaterials int64      `gorm:"column:total_materials"`
	}

	// Single query with aggregation to get all course progress data
	err := s.db.Table("enrollments e").
		Select(`
			e.user_id,
			u.name as user_name,
			u.email as user_email,
			e.seat_number,
			COALESCE(COUNT(CASE WHEN sp.status = ? THEN 1 END), 0) as completed,
			COALESCE(SUM(sp.score), 0) as score_sum,
			COALESCE(COUNT(sp.progress_id), 0) as progress_count,
			MAX(sp.last_submitted_at) as last_activity,
			(SELECT COUNT(*) FROM course_materials cm WHERE cm.course_id = ? AND cm.type IN ('code_exercise', 'pdf_exercise')) as total_materials
		`, enums.ProgressCompleted, courseID).
		Joins("LEFT JOIN users u ON e.user_id = u.user_id").
		Joins("LEFT JOIN student_progress sp ON e.user_id = sp.user_id").
		Joins("LEFT JOIN course_materials cm ON sp.material_id = cm.material_id AND cm.course_id = ?", courseID).
		Where("e.course_id = ?", courseID).
		Group("e.user_id, u.name, u.email, e.seat_number").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("get course progress: %w", err)
	}

	// Convert results to output format
	out := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		avg := 0.0
		if r.ProgressCount > 0 {
			avg = float64(r.ScoreSum) / float64(r.ProgressCount)
		}

		out = append(out, map[string]interface{}{
			"user_id":         r.UserID,
			"user_name":       r.UserName,
			"user_email":      r.UserEmail,
			"seat_number":     r.SeatNumber,
			"completed":       r.Completed,
			"total_materials": r.TotalMaterials,
			"average_score":   avg,
			"last_activity":   r.LastActivity,
		})
	}

	return out, nil
}

func (s *ProgressService) VerifyProgress(progressID, verifiedBy string, status enums.VerificationStatus, comment string) (*models.VerificationLog, error) {
	if status != enums.VerificationApproved && status != enums.VerificationRejected {
		return nil, fmt.Errorf("invalid verification status")
	}
	log := &models.VerificationLog{
		ProgressID: progressID,
		VerifiedBy: verifiedBy,
		Status:     status,
		Comment:    comment,
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(log).Error; err != nil {
			return fmt.Errorf("create log: %w", err)
		}
		// update progress status
		var newStatus enums.ProgressStatus = enums.ProgressInProgress
		if status == enums.VerificationApproved {
			newStatus = enums.ProgressCompleted
		}
		if err := tx.Model(&models.StudentProgress{}).
			Where("progress_id = ?", progressID).
			Update("status", newStatus).Error; err != nil {
			return fmt.Errorf("update progress: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return log, nil
}

func (s *ProgressService) GetVerificationLogs(progressID string) ([]models.VerificationLog, error) {
	var logs []models.VerificationLog
	if err := s.db.Where("progress_id = ?", progressID).
		Order("verified_at DESC").
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("get logs: %w", err)
	}
	return logs, nil
}

func (s *ProgressService) CanRequestReview(userID, exerciseID string) (bool, error) {
	// ตรวจสอบว่ามี StudentProgress และสถานะ in_progress
	var prog models.StudentProgress
	err := s.db.Where("user_id = ? AND exercise_id = ?", userID, exerciseID).
		First(&prog).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("get progress: %w", err)
	}

	// ต้องเป็น in_progress และยังไม่มี verification log ที่ pending
	if prog.Status != enums.ProgressInProgress {
		return false, nil
	}

	// ตรวจสอบว่าไม่มี pending verification
	var count int64
	if err := s.db.Model(&models.VerificationLog{}).
		Where("progress_id = ? AND status = ?", prog.ProgressID, enums.VerificationPending).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check existing verification: %w", err)
	}

	return count == 0, nil
}

// CanRequestApproval checks if student can request approval for a material
func (s *ProgressService) CanRequestApproval(userID, materialID string) (bool, error) {
	// Check if student has progress with in_progress status (all tests passed)
	var prog models.StudentProgress
	err := s.db.Where("user_id = ? AND material_id = ?", userID, materialID).
		First(&prog).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("get progress: %w", err)
	}

	// Must be in_progress status (all tests passed) and not already waiting for approval
	if prog.Status != enums.ProgressInProgress {
		return false, nil
	}

	// Check if there's already a pending queue job for this material
	var count int64
	if err := s.db.Model(&models.QueueJob{}).
		Where("user_id = ? AND material_id = ? AND status = ?", userID, materialID, enums.QueueStatusPending).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check existing queue job: %w", err)
	}

	return count == 0, nil
}

func (s *ProgressService) RequestReview(userID, exerciseID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// หา progress
		var prog models.StudentProgress
		if err := tx.Where("user_id = ? AND exercise_id = ?", userID, exerciseID).
			First(&prog).Error; err != nil {
			return fmt.Errorf("get progress: %w", err)
		}

		// สร้าง verification log
		log := models.VerificationLog{
			ProgressID: prog.ProgressID,
			VerifiedBy: "", // จะถูกกำหนดเมื่อ teacher review

			Status:  enums.VerificationPending,
			Comment: "Review requested by student",
		}

		if err := tx.Create(&log).Error; err != nil {
			return fmt.Errorf("create verification log: %w", err)
		}

		return nil
	})
}

// RequestApproval creates a queue job for TA approval with lab/table selection
func (s *ProgressService) RequestApproval(userID, materialID, labRoom, tableNumber, notes string) (*models.QueueJob, error) {
	var queueJob *models.QueueJob
	var err error

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Get progress
		var prog models.StudentProgress
		if err := tx.Where("user_id = ? AND material_id = ?", userID, materialID).
			First(&prog).Error; err != nil {
			return fmt.Errorf("get progress: %w", err)
		}

		// Get course ID from material
		var material models.CourseMaterial
		if err := tx.Where("material_id = ?", materialID).First(&material).Error; err != nil {
			return fmt.Errorf("get material: %w", err)
		}

		// Get latest submission for this material
		var submission models.Submission
		if err := tx.Where("user_id = ? AND material_id = ?", userID, materialID).
			Order("submitted_at DESC").First(&submission).Error; err != nil {
			return fmt.Errorf("get submission: %w", err)
		}

		// Create queue job with lab/table selection
		// Notes is optional (can be empty string)
		queueJob = &models.QueueJob{
			Type:         enums.QueueTypeReview,
			Status:       enums.QueueStatusPending,
			UserID:       userID,
			MaterialID:   &materialID,
			CourseID:     &material.CourseID,
			SubmissionID: &submission.SubmissionID,
			LabRoom:      &labRoom,
			TableNumber:  &tableNumber,
			Data:         fmt.Sprintf(`{"review_notes":"%s","lab_room":"%s","table_number":"%s"}`, notes, labRoom, tableNumber),
		}

		if err := tx.Create(queueJob).Error; err != nil {
			return fmt.Errorf("create queue job: %w", err)
		}

		// Update submission with queue_job_id
		if err := tx.Model(&models.Submission{}).
			Where("submission_id = ?", submission.SubmissionID).
			Update("queue_job_id", queueJob.ID).Error; err != nil {
			return fmt.Errorf("update submission queue_job_id: %w", err)
		}

		// Update progress status to waiting_approval
		if err := tx.Model(&models.StudentProgress{}).
			Where("progress_id = ?", prog.ProgressID).
			Update("status", enums.ProgressWaitingApproval).Error; err != nil {
			return fmt.Errorf("update progress status: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return queueJob, nil
}

func (s *ProgressService) GetCourseIDFromProgress(progressID string) (string, error) {
	var progress models.StudentProgress
	if err := s.db.Where("progress_id = ?", progressID).First(&progress).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("progress not found")
		}
		return "", fmt.Errorf("failed to get progress: %w", err)
	}

	// Get course ID through material -> course relationship
	var courseMaterial models.CourseMaterial
	if err := s.db.Where("material_id = ?", progress.MaterialID).First(&courseMaterial).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // Material not found
		}
		return "", fmt.Errorf("failed to get course from material: %w", err)
	}

	return courseMaterial.CourseID, nil
}
