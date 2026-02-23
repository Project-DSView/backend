package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"gorm.io/gorm"
)

type InvitationService struct {
	db                *gorm.DB
	courseService     *CourseService
	enrollmentService *EnrollmentService
}

func NewInvitationService(db *gorm.DB, courseService *CourseService, enrollmentService *EnrollmentService) *InvitationService {
	return &InvitationService{
		db:                db,
		courseService:     courseService,
		enrollmentService: enrollmentService,
	}
}

// generateToken generates a random 32-character token
func generateToken() (string, error) {
	bytes := make([]byte, 16) // 16 bytes = 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateInvitation creates a new invitation link for a course
func (s *InvitationService) CreateInvitation(courseID, userID string) (*models.CourseInvitation, error) {
	// Verify course exists
	course, err := s.courseService.GetCourseByID(courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course: %w", err)
	}
	if course == nil {
		return nil, errors.New("course not found")
	}

	// Verify user is the creator
	if course.CreatedBy != userID {
		return nil, errors.New("only course creator can create invitations")
	}

	// Generate token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// Set expiration to 100 years from now (permanent link)
	expiresAt := time.Now().Add(100 * 365 * 24 * time.Hour)

	// Create invitation
	invitation := models.CourseInvitation{
		CourseID:  courseID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := s.db.Create(&invitation).Error; err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	return &invitation, nil
}

// GetInvitationByToken retrieves an invitation by token
func (s *InvitationService) GetInvitationByToken(token string) (*models.CourseInvitation, error) {
	var invitation models.CourseInvitation
	if err := s.db.Where("token = ?", token).First(&invitation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("invitation not found")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}
	return &invitation, nil
}

// GetCourseInvitations retrieves all invitations for a course
func (s *InvitationService) GetCourseInvitations(courseID, userID string) ([]models.CourseInvitation, error) {
	// Verify course exists and user is creator
	course, err := s.courseService.GetCourseByID(courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course: %w", err)
	}
	if course == nil {
		return nil, errors.New("course not found")
	}

	if course.CreatedBy != userID {
		return nil, errors.New("only course creator can view invitations")
	}

	var invitations []models.CourseInvitation
	if err := s.db.Where("course_id = ?", courseID).Order("created_at DESC").Find(&invitations).Error; err != nil {
		return nil, fmt.Errorf("failed to get invitations: %w", err)
	}

	return invitations, nil
}

// ValidateInvitation validates an invitation token and checks expiration
func (s *InvitationService) ValidateInvitation(token string) (*models.CourseInvitation, error) {
	invitation, err := s.GetInvitationByToken(token)
	if err != nil {
		return nil, err
	}

	// Check if expired
	if invitation.IsExpired() {
		return nil, errors.New("invitation has expired")
	}

	return invitation, nil
}

// EnrollViaInvitation enrolls a user in a course using an invitation token
func (s *InvitationService) EnrollViaInvitation(token, userID string) (*models.Enrollment, error) {
	// Validate invitation
	invitation, err := s.ValidateInvitation(token)
	if err != nil {
		return nil, err
	}

	// Verify course exists
	course, err := s.courseService.GetCourseByID(invitation.CourseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course: %w", err)
	}
	if course == nil {
		return nil, errors.New("course not found")
	}

	// Check if already enrolled
	isEnrolled, err := s.enrollmentService.IsUserEnrolled(invitation.CourseID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check enrollment: %w", err)
	}
	if isEnrolled {
		return nil, errors.New("already enrolled in this course")
	}

	// Create enrollment directly without checking enroll_key (invitation bypasses this)
	enrollment := models.Enrollment{
		CourseID: invitation.CourseID,
		UserID:   userID,
		Role:     enums.EnrollmentRoleStudent,
	}

	if err := s.db.Create(&enrollment).Error; err != nil {
		return nil, fmt.Errorf("failed to create enrollment: %w", err)
	}

	// Populate user info
	userService := NewUserService(s.db)
	user, err := userService.GetUserByID(userID)
	if err == nil && user != nil {
		enrollment.UserInfo = user
	}

	return &enrollment, nil
}
