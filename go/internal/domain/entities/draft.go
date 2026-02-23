// internal/models/exercise_draft.go (ไฟล์ใหม่)
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExerciseDraft struct {
	DraftID     string    `json:"draft_id" gorm:"primaryKey;type:varchar(36)"`
	UserID      string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	MaterialID  string    `json:"material_id" gorm:"type:varchar(36);not null;index"`
	MaterialType string   `json:"material_type,omitempty" gorm:"type:varchar(20);index"` // Polymorphic: code_exercise (only code exercises have drafts)
	Code       string    `json:"code" gorm:"type:text;not null"`
	FileName   string    `json:"file_name" gorm:"type:varchar(255);not null"`
	FilePath   string    `json:"file_path" gorm:"type:text;not null"`
	FileSize   int64     `json:"file_size" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations (optional for joins)
	User           *User           `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
	CourseMaterial *CourseMaterial `json:"course_material,omitempty" gorm:"foreignKey:MaterialID;references:MaterialID"`
}

func (d *ExerciseDraft) BeforeCreate(tx *gorm.DB) error {
	if d.DraftID == "" {
		d.DraftID = uuid.New().String()
	}
	return nil
}

func (ExerciseDraft) TableName() string {
	return "exercise_drafts"
}

func (d *ExerciseDraft) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"draft_id":    d.DraftID,
		"user_id":     d.UserID,
		"material_id": d.MaterialID,
		"code":        d.Code,
		"file_name":   d.FileName,
		"file_path":   d.FilePath,
		"file_size":   d.FileSize,
		"created_at":  d.CreatedAt,
		"updated_at":  d.UpdatedAt,
	}

	// Add user relation if present
	if d.User != nil {
		result["user"] = d.User.ToJSON()
	}

	// Add course material relation if present
	if d.CourseMaterial != nil {
		result["course_material"] = d.CourseMaterial.ToJSON()
	}

	return result
}
