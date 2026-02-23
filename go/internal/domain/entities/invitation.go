package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CourseInvitation represents an invitation link for a course
type CourseInvitation struct {
	InvitationID string    `json:"invitation_id" gorm:"primaryKey;type:varchar(36)"`
	CourseID     string    `json:"course_id" gorm:"type:varchar(36);not null;index"`
	Token        string    `json:"token" gorm:"type:varchar(64);uniqueIndex;not null"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"type:timestamp;not null;index"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Course Course `json:"course,omitempty" gorm:"foreignKey:CourseID;references:CourseID"`
}

func (i *CourseInvitation) BeforeCreate(tx *gorm.DB) error {
	if i.InvitationID == "" {
		i.InvitationID = uuid.New().String()
	}
	return nil
}

func (CourseInvitation) TableName() string {
	return "course_invitations"
}

func (i *CourseInvitation) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"invitation_id": i.InvitationID,
		"course_id":     i.CourseID,
		"token":         i.Token,
		"expires_at":    i.ExpiresAt,
		"created_at":    i.CreatedAt,
		"updated_at":    i.UpdatedAt,
	}
}

// IsExpired checks if the invitation has expired
func (i *CourseInvitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

