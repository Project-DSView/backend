package models

import (
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/pkg/week"
	"gorm.io/gorm"
)

// Video represents a video material
type Video struct {
	MaterialBase
	VideoURL string `json:"video_url" gorm:"type:text;not null"`
}

// TableName returns the table name
func (Video) TableName() string {
	return "videos"
}

// GetMaterialType returns the material type
func (v *Video) GetMaterialType() string {
	return string(enums.MaterialTypeVideo)
}

// ToJSON converts Video to JSON map
func (v *Video) ToJSON() map[string]interface{} {
	result := v.MaterialBase.ToJSONBase()
	result["type"] = enums.MaterialTypeVideo
	result["video_url"] = v.VideoURL

	if v.Creator.UserID != "" {
		result["creator"] = v.Creator.ToJSON()
	}

	return result
}

// BeforeCreate sets the material ID if not already set
func (v *Video) BeforeCreate(tx *gorm.DB) error {
	return v.MaterialBase.BeforeCreate(tx)
}

// HasVideo returns true if video URL is set
func (v *Video) HasVideo() bool {
	return v.VideoURL != ""
}

// GetDisplayURL returns the video URL
func (v *Video) GetDisplayURL() string {
	return v.VideoURL
}

// WeekBasedEntity interface implementation
func (v *Video) GetWeek() int {
	return v.MaterialBase.Week
}

func (v *Video) SetWeek(week int) {
	v.MaterialBase.Week = week
}

func (v *Video) GetTableName() string {
	return "videos"
}

func (v *Video) GetCourseID() string {
	return v.MaterialBase.CourseID
}

// Ensure Video implements WeekBasedEntity interface
var _ week.WeekBasedEntity = (*Video)(nil)


















