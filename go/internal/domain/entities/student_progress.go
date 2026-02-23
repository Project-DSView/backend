package models

import (
	"time"

	"github.com/Project-DSView/backend/go/internal/domain/enums"
)

// StudentProgress represents the domain entity for student progress
type StudentProgress struct {
	ProgressID      string                `json:"progress_id" gorm:"primaryKey;type:varchar(36)"`
	UserID          string                `json:"user_id" gorm:"type:varchar(36);index;not null"`
	MaterialID      string                `json:"material_id" gorm:"type:varchar(36);index;not null"` // สำหรับ course materials
	MaterialType    string                `json:"material_type,omitempty" gorm:"type:varchar(20);index"` // Polymorphic: video, document, code_exercise, pdf_exercise
	Status          enums.ProgressStatus  `json:"status" gorm:"type:varchar(20);not null;default:'not_started'"`
	Score           int                   `json:"score" gorm:"default:0;not null"`
	SeatNumber      string                `json:"seat_number" gorm:"type:varchar(50)"`
	LastSubmittedAt *time.Time            `json:"last_submitted_at" gorm:"type:timestamp"`
	CreatedAt       time.Time             `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time             `json:"updated_at" gorm:"autoUpdateTime"`
}

// NewStudentProgress creates a new StudentProgress entity for materials
func NewStudentProgress(userID, materialID string) *StudentProgress {
	return &StudentProgress{
		UserID:     userID,
		MaterialID: materialID,
		Status:     enums.ProgressNotStarted,
		Score:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// UpdateScore updates the score and status
func (sp *StudentProgress) UpdateScore(score int) {
	sp.Score = score
	sp.Status = enums.ProgressCompleted
	sp.UpdatedAt = time.Now()
}

// IsCompleted returns true if the progress is completed
func (sp *StudentProgress) IsCompleted() bool {
	return sp.Status == enums.ProgressCompleted
}

// GetReferenceID returns the material ID
func (sp *StudentProgress) GetReferenceID() string {
	return sp.MaterialID
}

// IsMaterialBased returns true (always true now)
func (sp *StudentProgress) IsMaterialBased() bool {
	return true
}
