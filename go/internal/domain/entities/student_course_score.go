package models

import (
	"time"

	"gorm.io/gorm"
)

type StudentCourseScore struct {
	UserID      string    `json:"user_id" gorm:"type:varchar(36);primaryKey;index"`
	CourseID    string    `json:"course_id" gorm:"type:varchar(36);primaryKey;index"`
	TotalScore  int       `json:"total_score" gorm:"default:0;not null"`
	LastUpdated time.Time `json:"last_updated" gorm:"autoUpdateTime"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relations
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
	Course Course `json:"course,omitempty" gorm:"foreignKey:CourseID;references:CourseID"`
}

func (scs *StudentCourseScore) BeforeCreate(tx *gorm.DB) error {
	return nil
}

func (StudentCourseScore) TableName() string {
	return "student_course_scores"
}

func (scs *StudentCourseScore) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"user_id":      scs.UserID,
		"course_id":    scs.CourseID,
		"total_score":  scs.TotalScore,
		"last_updated": scs.LastUpdated,
		"created_at":   scs.CreatedAt,
	}
}
