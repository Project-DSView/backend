package models

import (
	"time"

	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Course struct {
	CourseID    string             `json:"course_id" gorm:"primaryKey;type:varchar(36)"`
	Name        string             `json:"name" gorm:"type:varchar(255);not null"`
	Description string             `json:"description" gorm:"type:text"`
	ImageURL    string             `json:"image_url" gorm:"type:text"`
	CreatedBy   string             `json:"created_by" gorm:"type:varchar(36);not null"`
	EnrollKey   string             `json:"enroll_key" gorm:"type:varchar(255);uniqueIndex;not null"`
	Status      enums.CourseStatus `json:"status" gorm:"type:varchar(20);check:status IN ('active','archived');default:'active';not null"`
	CreatedAt   time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time          `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations - Fix the foreign key references
	Enrollments []Enrollment `json:"enrollments,omitempty" gorm:"foreignKey:CourseID;references:CourseID"`

	// Creator info - populated manually, not a GORM relation
	CreatorInfo *User `json:"creator,omitempty" gorm:"-"`

	// Computed fields
	EnrollmentCount int `json:"enrollment_count,omitempty" gorm:"-"`
	MaterialCount   int `json:"material_count,omitempty" gorm:"-"`
}

func (c *Course) BeforeCreate(tx *gorm.DB) error {
	if c.CourseID == "" {
		c.CourseID = uuid.New().String()
	}
	if c.Status == "" {
		c.Status = enums.CourseStatusActive
	}
	if c.EnrollKey == "" {
		c.EnrollKey = uuid.New().String()[:8] // Generate 8-char key
	}
	return nil
}

func (Course) TableName() string {
	return "courses"
}

func (c *Course) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"course_id":   c.CourseID,
		"name":        c.Name,
		"description": c.Description,
		"image_url":   c.ImageURL,
		"created_by":  c.CreatedBy,
		"enroll_key":  c.EnrollKey,
		"status":      c.Status,
		"created_at":  c.CreatedAt,
		"updated_at":  c.UpdatedAt,
	}

	if c.CreatorInfo != nil {
		result["creator"] = c.CreatorInfo.ToJSON()
	}

	if c.EnrollmentCount > 0 {
		result["enrollment_count"] = c.EnrollmentCount
	}

	if c.MaterialCount > 0 {
		result["material_count"] = c.MaterialCount
	}

	if len(c.Enrollments) > 0 {
		enrollments := make([]map[string]interface{}, len(c.Enrollments))
		for i, enrollment := range c.Enrollments {
			enrollments[i] = enrollment.ToJSON()
		}
		result["enrollments"] = enrollments
	}

	return result
}
