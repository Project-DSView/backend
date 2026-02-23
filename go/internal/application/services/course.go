package services

import (
	"fmt"
	"strings"
	"time"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/pkg/logger"
	"gorm.io/gorm"
)

type CourseService struct {
	db                *gorm.DB
	userService       *UserService
	enrollmentService *EnrollmentService
}

func NewCourseService(db *gorm.DB, userService *UserService, enrollmentService *EnrollmentService) *CourseService {
	return &CourseService{
		db:                db,
		userService:       userService,
		enrollmentService: enrollmentService,
	}
}

func (s *CourseService) CreateCourse(courseData *models.Course) error {
	return s.db.Create(courseData).Error
}

func (s *CourseService) GetCourseByID(courseID string) (*models.Course, error) {
	var courseModel models.Course
	if err := s.db.Preload("Enrollments").Where("course_id = ?", courseID).First(&courseModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get course: %w", err)
	}

	// Populate creator info
	if courseModel.CreatedBy != "" {
		creator, err := s.userService.GetUserByID(courseModel.CreatedBy)
		if err != nil {
			logger.Warnf("Could not get creator info for course %s: %v", courseID, err)
		} else if creator != nil {
			courseModel.CreatorInfo = creator
		}
	}

	// Set computed fields
	courseModel.EnrollmentCount = len(courseModel.Enrollments)

	// Get material count (replaces exercise count)
	var materialCount int64
	s.db.Table("course_materials").Where("course_id = ?", courseID).Count(&materialCount)
	courseModel.MaterialCount = int(materialCount)

	return &courseModel, nil
}

func (s *CourseService) GetCoursesWithFilters(page, limit int, status, search, userID string, isTeacher bool) ([]models.Course, int, error) {
	var courses []models.Course
	var total int64

	query := s.db.Model(&models.Course{})

	// Apply status filter
	if status != "" && enums.IsValidCourseStatus(status) {
		query = query.Where("status = ?", status)
	}

	// Apply search filter
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	// For students, show all active courses (not just enrolled ones)
	// For teachers, show all courses including archived ones
	if !isTeacher {
		query = query.Where("status = ?", enums.CourseStatusActive)
	}
	// Teachers can see all courses (active, archived, etc.)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count courses: %w", err)
	}

	// Apply pagination and get courses
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&courses).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch courses: %w", err)
	}

	// Populate additional info for each course (same as before)
	for i := range courses {
		if courses[i].CreatedBy != "" {
			creator, err := s.userService.GetUserByID(courses[i].CreatedBy)
			if err != nil {
				logger.Warnf("Could not get creator info for course %s: %v", courses[i].CourseID, err)
			} else if creator != nil {
				courses[i].CreatorInfo = creator
			}
		}

		if s.enrollmentService != nil {
			enrollmentCount, err := s.enrollmentService.GetEnrollmentCount(courses[i].CourseID)
			if err != nil {
				logger.Warnf("Could not get enrollment count for course %s: %v", courses[i].CourseID, err)
			} else {
				courses[i].EnrollmentCount = enrollmentCount
			}
		}

		var materialCount int64
		s.db.Table("course_materials").Where("course_id = ?", courses[i].CourseID).Count(&materialCount)
		courses[i].MaterialCount = int(materialCount)
	}

	return courses, int(total), nil
}

func (s *CourseService) UpdateCourse(courseID string, updates map[string]interface{}) error {
	result := s.db.Model(&models.Course{}).Where("course_id = ?", courseID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update course: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("course not found")
	}
	return nil
}

func (s *CourseService) DeleteCourse(courseID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {

		enrollmentCount, err := s.enrollmentService.GetEnrollmentCount(courseID)
		if err != nil {
			return fmt.Errorf("failed to check enrollments: %w", err)
		}

		if enrollmentCount > 0 {
			return fmt.Errorf("cannot delete course with active enrollments. Please archive first")
		}
		// Delete course materials instead of course exercises
		if err := tx.Where("course_id = ?", courseID).Delete(&models.CourseMaterial{}).Error; err != nil {
			return fmt.Errorf("failed to delete course materials: %w", err)
		}

		// Delete enrollments
		if err := tx.Where("course_id = ?", courseID).Delete(&models.Enrollment{}).Error; err != nil {
			return fmt.Errorf("failed to delete enrollments: %w", err)
		}

		// Delete the course
		result := tx.Where("course_id = ?", courseID).Delete(&models.Course{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete course: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("course not found")
		}

		return s.db.Delete(&models.Course{}, "course_id = ?", courseID).Error
	})
}

func (s *CourseService) AddMaterialToCourse(courseID, materialID, addedBy string) error {
	// This method is no longer needed as materials are created directly with course_id
	return fmt.Errorf("use CourseMaterialService.CreateCourseMaterial instead")
}

func (s *CourseService) RemoveMaterialFromCourse(courseID, materialID string) error {
	result := s.db.Where("course_id = ? AND material_id = ?", courseID, materialID).Delete(&models.CourseMaterial{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove material from course: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("course-material relationship not found")
	}
	return nil
}

func (s *CourseService) GetCourseMaterials(courseID string) ([]models.CourseMaterial, error) {
	var materials []models.CourseMaterial
	if err := s.db.Where("course_id = ?", courseID).
		Order("week ASC, created_at ASC").
		Find(&materials).Error; err != nil {
		return nil, fmt.Errorf("failed to get course materials: %w", err)

	}
	return materials, nil
}

func (s *CourseService) GetCourseStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total courses
	var totalCourses int64
	if err := s.db.Model(&models.Course{}).Count(&totalCourses).Error; err != nil {
		return nil, fmt.Errorf("failed to count courses: %w", err)
	}
	stats["total_courses"] = totalCourses

	// Courses by status
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	if err := s.db.Model(&models.Course{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get status statistics: %w", err)
	}

	statusMap := make(map[string]int64)
	for _, stat := range statusStats {
		statusMap[stat.Status] = stat.Count
	}
	stats["courses_by_status"] = statusMap

	// Total enrollments
	var totalEnrollments int64
	if err := s.db.Model(&models.Enrollment{}).Count(&totalEnrollments).Error; err != nil {
		return nil, fmt.Errorf("failed to count enrollments: %w", err)
	}
	stats["total_enrollments"] = totalEnrollments

	// Recent courses (last 7 days)
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	var recentCourses int64
	if err := s.db.Model(&models.Course{}).
		Where("created_at >= ?", weekAgo).
		Count(&recentCourses).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent courses: %w", err)
	}
	stats["recent_courses_7d"] = recentCourses

	stats["last_updated"] = time.Now()
	return stats, nil
}

func (s *CourseService) CanUserModifyCourse(userID, courseID string, isTeacher bool) (bool, error) {
	if isTeacher {
		return true, nil
	}
	return false, nil
}

func (s *CourseService) GetCourseByEnrollKey(enrollKey string) (*models.Course, error) {
	var courseModel models.Course
	if err := s.db.Where("enroll_key = ?", enrollKey).First(&courseModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get course by enroll key: %w", err)
	}
	return &courseModel, nil
}

// CourseMaterialWithWeek represents a material with its week information
type CourseMaterialWithWeek struct {
	models.CourseMaterial
	Week int `json:"week"`
}

func (s *CourseService) GetCourseMaterialsWithFilters(courseID string, allowedTypes []string, page, limit int) ([]CourseMaterialWithWeek, int, error) {
	var materials []CourseMaterialWithWeek
	var total int64

	// Base query
	query := s.db.Table("course_materials").
		Select("course_materials.*, course_materials.week").
		Where("course_materials.course_id = ?", courseID)

	// Apply type filter
	if len(allowedTypes) > 0 {
		query = query.Where("course_materials.type IN ?", allowedTypes)
	} else {
		// No allowed types = no results
		return materials, 0, nil
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count course materials: %w", err)
	}

	// Apply pagination and ordering
	offset := (page - 1) * limit
	if err := query.Order("course_materials.week ASC, course_materials.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&materials).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch course materials: %w", err)
	}

	return materials, int(total), nil
}

// IsUserEnrolledInCourse checks if user is enrolled in course (helper method)
func (s *CourseService) IsUserEnrolledInCourse(courseID, userID string) (bool, enums.EnrollmentRole, error) {
	var enrollment models.Enrollment
	err := s.db.Where("course_id = ? AND user_id = ?", courseID, userID).First(&enrollment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check enrollment: %w", err)
	}
	return true, enrollment.Role, nil
}

// GetCourseEnrollmentCount returns the total number of enrollments in a course
func (s *CourseService) GetCourseEnrollmentCount(courseID string) (int, error) {
	var count int64
	err := s.db.Model(&models.Enrollment{}).Where("course_id = ?", courseID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count enrollments: %w", err)
	}
	return int(count), nil
}

// GetMaterialsCountByCourse returns the total number of materials in a course
func (s *CourseService) GetMaterialsCountByCourse(courseID string) (int, error) {
	var count int64
	err := s.db.Model(&models.CourseMaterial{}).Where("course_id = ?", courseID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count materials: %w", err)
	}
	return int(count), nil
}
