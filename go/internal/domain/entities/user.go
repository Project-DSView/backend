package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UserID         string     `json:"user_id" gorm:"primaryKey;type:varchar(36)"`
	FirstName      string     `json:"firstname" gorm:"column:first_name;type:varchar(255);not null"`
	LastName       string     `json:"lastname" gorm:"column:last_name;type:varchar(255);not null"`
	Email          string     `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	IsTeacher      bool       `json:"is_teacher" gorm:"default:false;not null"`
	ProfileImg     string     `json:"profile_img" gorm:"type:text"`
	RefreshToken   string     `json:"-" gorm:"type:text"` // Hide from JSON
	RefreshTokenAt *time.Time `json:"-" gorm:"type:timestamp"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UserID == "" {
		u.UserID = uuid.New().String()
	}
	// Ensure names are not empty
	if u.FirstName == "" {
		u.FirstName = "Unknown"
	}
	if u.LastName == "" {
		u.LastName = "User"
	}
	return nil
}

// TableName overrides the table name used by User to `users`
func (User) TableName() string {
	return "users"
}

// ToJSON returns a map representation suitable for JSON responses
func (u *User) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"user_id":     u.UserID,
		"firstname":   u.FirstName,
		"lastname":    u.LastName,
		"email":       u.Email,
		"is_teacher":  u.IsTeacher,
		"profile_img": u.ProfileImg,
		"created_at":  u.CreatedAt,
		"updated_at":  u.UpdatedAt,
	}
}
