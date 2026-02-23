package models

import (
	"time"

	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QueueJob struct {
	ID           string            `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Type         enums.QueueType   `json:"type" gorm:"type:varchar(50);not null"`
	Status       enums.QueueStatus `json:"status" gorm:"type:varchar(20);default:'pending';not null"`
	UserID       string            `json:"user_id" gorm:"type:varchar(36);not null"`
	MaterialID   *string           `json:"material_id" gorm:"type:varchar(36)"`
	MaterialType *string           `json:"material_type,omitempty" gorm:"type:varchar(20);index"` // Polymorphic: video, document, code_exercise, pdf_exercise
	CourseID     *string           `json:"course_id" gorm:"type:varchar(36)"`
	SubmissionID *string           `json:"submission_id" gorm:"type:varchar(36)"` // Link to submission
	LabRoom      *string           `json:"lab_room" gorm:"type:varchar(50)"`      // Lab room selection
	TableNumber  *string           `json:"table_number" gorm:"type:varchar(20)"`  // Table number selection
	Data         string            `json:"data" gorm:"type:text"`                 // JSON data
	Result       string            `json:"result" gorm:"type:text"`               // JSON result
	Error        string            `json:"error" gorm:"type:text"`
	ProcessedBy  *string           `json:"processed_by" gorm:"type:varchar(36)"` // TA/Teacher who processed
	ClaimedAt    *time.Time        `json:"claimed_at" gorm:"type:timestamp"`     // When TA claimed the job
	CreatedAt    time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	StartedAt    *time.Time        `json:"started_at"`
	CompletedAt  *time.Time        `json:"completed_at"`

	// Relations (without foreign key constraints to avoid database issues)
	User           *User           `json:"user,omitempty" gorm:"-"`
	CourseMaterial *CourseMaterial `json:"course_material,omitempty" gorm:"-"`
	Course         *Course         `json:"course,omitempty" gorm:"-"`
	ProcessedByUser *User          `json:"processed_by_user,omitempty" gorm:"-"` // User who processed/claimed the job
	ReviewStatus   *string         `json:"review_status,omitempty" gorm:"-"`     // Review status from submission ('approved' or 'rejected')
}

func (q *QueueJob) BeforeCreate(tx *gorm.DB) error {
	if q.ID == "" {
		q.ID = uuid.New().String()
	}
	if q.Status == "" {
		q.Status = enums.QueueStatusPending
	}
	return nil
}

func (QueueJob) TableName() string {
	return "queue_jobs"
}

func (q *QueueJob) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"id":            q.ID,
		"type":          q.Type,
		"status":        q.Status,
		"user_id":       q.UserID,
		"material_id":   q.MaterialID,
		"course_id":     q.CourseID,
		"submission_id": q.SubmissionID,
		"lab_room":      q.LabRoom,
		"table_number":  q.TableNumber,
		"data":          q.Data,
		"result":        q.Result,
		"error":         q.Error,
		"processed_by":  q.ProcessedBy,
		"claimed_at":    q.ClaimedAt,
		"created_at":    q.CreatedAt,
		"updated_at":    q.UpdatedAt,
		"started_at":    q.StartedAt,
		"completed_at":  q.CompletedAt,
	}

	if q.User != nil {
		userJSON := q.User.ToJSON()
		result["user"] = userJSON
	}
	if q.CourseMaterial != nil {
		result["course_material"] = q.CourseMaterial.ToJSON()
	}
	if q.Course != nil {
		courseJSON := q.Course.ToJSON()
		result["course"] = courseJSON
	}
	if q.ProcessedByUser != nil {
		processedByUserJSON := q.ProcessedByUser.ToJSON()
		result["processed_by_user"] = processedByUserJSON
	}

	if q.ReviewStatus != nil {
		result["review_status"] = *q.ReviewStatus
	}

	return result
}

func IsValidQueueStatus(status string) bool {
	switch enums.QueueStatus(status) {
	case enums.QueueStatusPending, enums.QueueStatusProcessing, enums.QueueStatusCompleted, enums.QueueStatusFailed, enums.QueueStatusCancelled:
		return true
	default:
		return false
	}
}

func IsValidQueueType(queueType string) bool {
	switch enums.QueueType(queueType) {
	case enums.QueueTypeCodeExecution, enums.QueueTypeReview:
		return true
	default:
		return false
	}
}
