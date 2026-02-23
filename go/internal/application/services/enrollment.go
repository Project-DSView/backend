package services

import (
	"fmt"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/pkg/logger"
	"gorm.io/gorm"
)

type EnrollmentService struct {
	db          *gorm.DB
	userService *UserService
}

func NewEnrollmentService(db *gorm.DB, userService *UserService) *EnrollmentService {
	return &EnrollmentService{
		db:          db,
		userService: userService,
	}
}

func (s *EnrollmentService) EnrollUser(courseID, userID, enrollKey string, role enums.EnrollmentRole) (*models.Enrollment, error) {
	// Verify course exists and enrollKey is correct
	var courseModel models.Course
	if err := s.db.Where("course_id = ? AND enroll_key = ?", courseID, enrollKey).First(&courseModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid course or enrollment key")
		}
		return nil, fmt.Errorf("failed to verify course: %w", err)
	}

	// Check if already enrolled
	var existingEnrollment models.Enrollment
	err := s.db.Where("course_id = ? AND user_id = ?", courseID, userID).First(&existingEnrollment).Error
	if err == nil {
		return nil, fmt.Errorf("already enrolled in this course")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing enrollment: %w", err)
	}

	// Create enrollment
	enrollment := models.Enrollment{
		CourseID: courseID,
		UserID:   userID,
		Role:     role,
	}

	if err := s.db.Create(&enrollment).Error; err != nil {
		return nil, fmt.Errorf("failed to create enrollment: %w", err)
	}

	//  ดึงข้อมูล user มาใส่ใน enrollment
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		logger.Warnf("Could not get user info for enrollment %s: %v", enrollment.EnrollmentID, err)
	} else if user != nil {
		enrollment.UserInfo = user
	}

	return &enrollment, nil
}

func (s *EnrollmentService) GetCourseEnrollments(courseID string) ([]models.Enrollment, error) {
	var enrollments []models.Enrollment
	if err := s.db.Where("course_id = ?", courseID).Order("enrolled_at ASC").Find(&enrollments).Error; err != nil {
		return nil, fmt.Errorf("failed to get enrollments: %w", err)
	}

	// Populate user info for each enrollment
	for i := range enrollments {
		if enrollments[i].UserID != "" {
			user, err := s.userService.GetUserByID(enrollments[i].UserID)
			if err != nil {
				logger.Warnf("Could not get user info for enrollment %s: %v", enrollments[i].EnrollmentID, err)
			} else if user != nil {
				enrollments[i].UserInfo = user
			}
		}
	}

	return enrollments, nil
}

func (s *EnrollmentService) UnenrollUser(courseID, userID string) error {
	result := s.db.Where("course_id = ? AND user_id = ?", courseID, userID).Delete(&models.Enrollment{})
	if result.Error != nil {
		return fmt.Errorf("failed to unenroll user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("enrollment not found")
	}
	return nil
}

func (s *EnrollmentService) IsUserEnrolled(courseID, userID string) (bool, error) {
	var count int64
	db := s.db.Model(&models.Enrollment{}).Where("course_id = ? AND user_id = ?", courseID, userID).Count(&count)
	return count > 0, db.Error
}

func (s *EnrollmentService) GetEnrollmentCount(courseID string) (int, error) {
	var count int64
	if err := s.db.Model(&models.Enrollment{}).Where("course_id = ?", courseID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count enrollments: %w", err)
	}
	return int(count), nil
}

func (s *EnrollmentService) GetUserEnrollmentInCourse(courseID, userID string) (*models.Enrollment, error) {
	var enrollment models.Enrollment
	err := s.db.Where("course_id = ? AND user_id = ?", courseID, userID).First(&enrollment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not enrolled
		}
		return nil, fmt.Errorf("failed to get user enrollment: %w", err)
	}
	return &enrollment, nil
}

func (s *EnrollmentService) UpdateEnrollmentRole(courseID, userID string, newRole enums.EnrollmentRole) error {
	result := s.db.Model(&models.Enrollment{}).
		Where("course_id = ? AND user_id = ?", courseID, userID).
		Update("role", newRole)

	if result.Error != nil {
		return fmt.Errorf("failed to update enrollment role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("enrollment not found")
	}
	return nil
}
