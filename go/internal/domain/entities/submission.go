package models

import (
	"time"

	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Submission struct {
	SubmissionID     string                 `json:"submission_id" gorm:"primaryKey;type:varchar(36)"`
	UserID           string                 `json:"user_id" gorm:"type:varchar(36);index;not null"`
	MaterialID       string                 `json:"material_id" gorm:"type:varchar(36);index;not null"`    // สำหรับ course materials
	MaterialType     string                 `json:"material_type,omitempty" gorm:"type:varchar(20);index"` // Polymorphic: video, document, code_exercise, pdf_exercise
	Code             string                 `json:"code" gorm:"type:text"`                                 // สำหรับ code exercises
	FileURL          string                 `json:"file_url" gorm:"type:text"`                             // สำหรับ PDF exercises
	FileName         string                 `json:"file_name" gorm:"type:varchar(255)"`                    // ชื่อไฟล์ที่ส่ง
	FileSize         int64                  `json:"file_size" gorm:"type:bigint;default:0"`                // ขนาดไฟล์
	MimeType         string                 `json:"mime_type" gorm:"type:varchar(100)"`                    // ประเภทไฟล์
	PassedCount      int                    `json:"passed_count" gorm:"default:0;not null"`
	FailedCount      int                    `json:"failed_count" gorm:"default:0;not null"`
	TotalScore       int                    `json:"total_score" gorm:"default:0;not null"`
	Status           enums.SubmissionStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	IsLateSubmission bool                   `json:"is_late_submission" gorm:"default:false;not null"` // ส่งช้าหรือไม่ (สำหรับ practice)
	ErrorMessage     string                 `json:"error_message" gorm:"type:text"`
	Feedback         string                 `json:"feedback" gorm:"type:text"`             // คำติชมจากอาจารย์/TA
	FeedbackFileURL  string                 `json:"feedback_file_url" gorm:"type:text"`    // URL ของไฟล์ feedback ที่อาจารย์/TA อัปโหลด
	GradedAt         *time.Time             `json:"graded_at" gorm:"type:timestamp"`       // เวลาที่ตรวจแล้ว
	GradedBy         string                 `json:"graded_by" gorm:"type:varchar(36)"`     // ID ของอาจารย์/TA ที่ตรวจ
	ReviewStatus     string                 `json:"review_status" gorm:"type:varchar(20)"` // 'approved' or 'rejected'
	QueueJobID       *string                `json:"queue_job_id" gorm:"type:varchar(36)"`  // Link to queue job
	SubmittedAt      time.Time              `json:"submitted_at" gorm:"autoCreateTime"`

	Results []SubmissionResult `json:"results,omitempty" gorm:"foreignKey:SubmissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (s *Submission) BeforeCreate(tx *gorm.DB) error {
	if s.SubmissionID == "" {
		s.SubmissionID = uuid.New().String()
	}
	return nil
}

func (Submission) TableName() string { return "submissions" }

// Helper methods for submission types
func (s *Submission) IsCodeSubmission() bool {
	return s.Code != "" && s.FileURL == ""
}

func (s *Submission) IsFileSubmission() bool {
	return s.FileURL != "" && s.Code == ""
}

func (s *Submission) IsPDFSubmission() bool {
	return s.IsFileSubmission() && s.MimeType == "application/pdf"
}

func (s *Submission) RequiresManualGrading() bool {
	return s.IsFileSubmission() // PDF exercises need manual grading
}

func (s *Submission) IsAutoGraded() bool {
	return s.IsCodeSubmission() // Code exercises are auto-graded
}

// GetReferenceID returns the material ID
func (s *Submission) GetReferenceID() string {
	return s.MaterialID
}

// IsMaterialBased returns true (always true now)
func (s *Submission) IsMaterialBased() bool {
	return true
}

type SubmissionResult struct {
	ResultID     string         `json:"result_id" gorm:"primaryKey;type:varchar(36)"`
	SubmissionID string         `json:"submission_id" gorm:"type:varchar(36);index;not null"`
	TestCaseID   string         `json:"test_case_id" gorm:"type:varchar(36);index;not null"`
	Status       string         `json:"status" gorm:"type:varchar(10);not null"`
	ActualOutput types.JSONData `json:"actual_output" gorm:"type:jsonb"`
	ErrorMessage string         `json:"error_message" gorm:"type:text"`
	CreatedAt    time.Time
}

func (sr *SubmissionResult) BeforeCreate(tx *gorm.DB) error {
	if sr.ResultID == "" {
		sr.ResultID = uuid.New().String()
	}
	return nil
}

func (SubmissionResult) TableName() string { return "submission_results" }

// StudentProgress is defined in student_progress.go

func (sp *StudentProgress) BeforeCreate(tx *gorm.DB) error {
	if sp.ProgressID == "" {
		sp.ProgressID = uuid.New().String()
	}
	return nil
}

func (StudentProgress) TableName() string { return "student_progress" }

type VerificationLog struct {
	LogID      string                   `json:"log_id" gorm:"primaryKey;type:varchar(36)"`
	ProgressID string                   `json:"progress_id" gorm:"type:varchar(36);index;not null"`
	VerifiedBy string                   `json:"verified_by" gorm:"type:varchar(36);not null"`
	Status     enums.VerificationStatus `json:"status" gorm:"type:varchar(20);not null"`
	Comment    string                   `json:"comment" gorm:"type:text"`
	VerifiedAt time.Time                `json:"verified_at" gorm:"autoCreateTime"`
}

func (vl *VerificationLog) BeforeCreate(tx *gorm.DB) error {
	if vl.LogID == "" {
		vl.LogID = uuid.New().String()
	}
	return nil
}

func (VerificationLog) TableName() string { return "verification_logs" }
