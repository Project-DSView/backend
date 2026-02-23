package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CourseWeek represents a week configuration for a specific course
type CourseWeek struct {
	CourseWeekID string    `json:"course_week_id" gorm:"primaryKey;type:varchar(36)"`
	CourseID     string    `json:"course_id" gorm:"type:varchar(36);not null;index"`
	WeekNumber   int       `json:"week_number" gorm:"type:int;not null;index"`
	Title        string    `json:"title" gorm:"type:varchar(255);not null"`
	Description  string    `json:"description" gorm:"type:text"`
	CreatedBy    string    `json:"created_by" gorm:"type:varchar(36);not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Course  Course `json:"course,omitempty" gorm:"foreignKey:CourseID;references:CourseID"`
	Creator User   `json:"creator,omitempty" gorm:"foreignKey:CreatedBy;references:UserID"`
}

// BeforeCreate hook to set CourseWeekID if not provided
func (cw *CourseWeek) BeforeCreate(tx *gorm.DB) error {
	if cw.CourseWeekID == "" {
		cw.CourseWeekID = uuid.New().String()
	}
	return nil
}

// TableName returns the table name for CourseWeek
func (CourseWeek) TableName() string {
	return "course_weeks"
}

// ToJSON converts CourseWeek to JSON map
func (cw *CourseWeek) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"course_week_id": cw.CourseWeekID,
		"course_id":      cw.CourseID,
		"week_number":    cw.WeekNumber,
		"title":          cw.Title,
		"description":    cw.Description,
		"created_by":     cw.CreatedBy,
		"created_at":     cw.CreatedAt,
		"updated_at":     cw.UpdatedAt,
	}

	if cw.Creator.UserID != "" {
		result["creator"] = cw.Creator.ToJSON()
	}

	return result
}

// GetDefaultTitle returns a default title for a week number
func GetDefaultTitle(weekNumber int) string {
	return fmt.Sprintf("สัปดาห์ที่ %d", weekNumber)
}

// NewCourseWeek creates a new CourseWeek with default title
func NewCourseWeek(courseID string, weekNumber int, createdBy string) *CourseWeek {
	return &CourseWeek{
		CourseWeekID: uuid.New().String(),
		CourseID:     courseID,
		WeekNumber:   weekNumber,
		Title:        GetDefaultTitle(weekNumber),
		CreatedBy:    createdBy,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// NewCourseWeekWithTitle creates a new CourseWeek with custom title
func NewCourseWeekWithTitle(courseID string, weekNumber int, title, description, createdBy string) *CourseWeek {
	return &CourseWeek{
		CourseWeekID: uuid.New().String(),
		CourseID:     courseID,
		WeekNumber:   weekNumber,
		Title:        title,
		Description:  description,
		CreatedBy:    createdBy,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
