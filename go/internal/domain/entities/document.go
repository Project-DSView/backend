package models

import (
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/pkg/week"
	"gorm.io/gorm"
)

// Document represents a document material
type Document struct {
	MaterialBase
	FileURL  string `json:"file_url" gorm:"type:text;not null"`
	FileName string `json:"file_name" gorm:"type:varchar(255);not null"`
	FileSize int64  `json:"file_size" gorm:"type:bigint;default:0"`
	MimeType string `json:"mime_type" gorm:"type:varchar(100)"`
}

// TableName returns the table name
func (Document) TableName() string {
	return "documents"
}

// GetMaterialType returns the material type
func (d *Document) GetMaterialType() string {
	return string(enums.MaterialTypeDocument)
}

// ToJSON converts Document to JSON map
func (d *Document) ToJSON() map[string]interface{} {
	result := d.MaterialBase.ToJSONBase()
	result["type"] = enums.MaterialTypeDocument
	result["file_url"] = d.FileURL
	result["file_name"] = d.FileName
	result["file_size"] = d.FileSize
	result["mime_type"] = d.MimeType

	if d.Creator.UserID != "" {
		result["creator"] = d.Creator.ToJSON()
	}

	return result
}

// BeforeCreate sets the material ID if not already set
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	return d.MaterialBase.BeforeCreate(tx)
}

// HasFile returns true if file URL is set
func (d *Document) HasFile() bool {
	return d.FileURL != "" && d.FileName != ""
}

// GetDisplayURL returns the file URL
func (d *Document) GetDisplayURL() string {
	return d.FileURL
}

// WeekBasedEntity interface implementation
func (d *Document) GetWeek() int {
	return d.MaterialBase.Week
}

func (d *Document) SetWeek(week int) {
	d.MaterialBase.Week = week
}

func (d *Document) GetTableName() string {
	return "documents"
}

func (d *Document) GetCourseID() string {
	return d.MaterialBase.CourseID
}

// Ensure Document implements WeekBasedEntity interface
var _ week.WeekBasedEntity = (*Document)(nil)


















