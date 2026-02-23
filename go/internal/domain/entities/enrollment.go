package models

import (
	"time"

	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Enrollment struct {
	EnrollmentID string               `json:"enrollment_id" gorm:"primaryKey;type:varchar(36)"`
	CourseID     string               `json:"course_id" gorm:"type:varchar(36);not null;index"`
	UserID       string               `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Role         enums.EnrollmentRole `json:"role" gorm:"type:varchar(20);check:role IN ('student','ta','teacher');default:'student';not null"`
	EnrolledAt   time.Time            `json:"enrolled_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time            `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations - Remove the problematic constraints
	Course Course `json:"course,omitempty" gorm:"foreignKey:CourseID;references:CourseID"`
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`

	// User info for response - populated manually
	UserInfo *User `json:"user_info,omitempty" gorm:"-"`
}

func (e *Enrollment) BeforeCreate(tx *gorm.DB) error {
	if e.EnrollmentID == "" {
		e.EnrollmentID = uuid.New().String()
	}
	if e.Role == "" {
		e.Role = enums.EnrollmentRoleStudent
	}
	return nil
}

func (Enrollment) TableName() string {
	return "enrollments"
}

func (e *Enrollment) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"enrollment_id": e.EnrollmentID,
		"course_id":     e.CourseID,
		"user_id":       e.UserID,
		"role":          e.Role,
		"enrolled_at":   e.EnrolledAt,
		"updated_at":    e.UpdatedAt,
	}

	if e.UserInfo != nil {
		result["firstname"] = e.UserInfo.FirstName
		result["lastname"] = e.UserInfo.LastName
		result["email"] = e.UserInfo.Email
	}

	return result
}

func IsValidEnrollmentRole(role string) bool {
	switch enums.EnrollmentRole(role) {
	case enums.EnrollmentRoleStudent, enums.EnrollmentRoleTA, enums.EnrollmentRoleTeacher:
		return true
	default:
		return false
	}
}
