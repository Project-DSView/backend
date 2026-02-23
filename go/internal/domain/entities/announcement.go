package models

import (
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/pkg/week"
	"gorm.io/gorm"
)

// Announcement represents an announcement material
// Uses MaterialBase for common fields and adds Content field
type Announcement struct {
	MaterialBase
	Content string `json:"content" gorm:"type:text;not null"` // Announcement content
}

// TableName returns the table name
func (Announcement) TableName() string {
	return "announcements"
}

// GetMaterialType returns the material type
func (a *Announcement) GetMaterialType() string {
	return string(enums.MaterialTypeAnnouncement)
}

// ToJSON converts Announcement to JSON map
func (a *Announcement) ToJSON() map[string]interface{} {
	result := a.MaterialBase.ToJSONBase()
	result["type"] = enums.MaterialTypeAnnouncement
	result["content"] = a.Content

	if a.Creator.UserID != "" {
		result["creator"] = a.Creator.ToJSON()
	}

	return result
}

// BeforeCreate sets the material ID if not already set
func (a *Announcement) BeforeCreate(tx *gorm.DB) error {
	return a.MaterialBase.BeforeCreate(tx)
}

// WeekBasedEntity interface implementation
func (a *Announcement) GetWeek() int {
	return a.MaterialBase.Week
}

func (a *Announcement) SetWeek(week int) {
	a.MaterialBase.Week = week
}

func (a *Announcement) GetTableName() string {
	return "announcements"
}

func (a *Announcement) GetCourseID() string {
	return a.MaterialBase.CourseID
}

// Ensure Announcement implements WeekBasedEntity interface
var _ week.WeekBasedEntity = (*Announcement)(nil)
