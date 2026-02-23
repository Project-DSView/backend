package services

import (
	"errors"
	"fmt"
	"time"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"gorm.io/gorm"
)

type AnnouncementService struct {
	db *gorm.DB
}

func NewAnnouncementService(db *gorm.DB) *AnnouncementService {
	return &AnnouncementService{db: db}
}

// CreateAnnouncement creates a new announcement
func (s *AnnouncementService) CreateAnnouncement(announcement *models.Announcement) error {
	// Check if course exists
	var course models.Course
	if err := s.db.First(&course, "course_id = ?", announcement.CourseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("course not found")
		}
		return err
	}

	// Check if creator exists and is a teacher
	var creator models.User
	if err := s.db.First(&creator, "user_id = ?", announcement.CreatedBy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("creator not found")
		}
		return err
	}

	if !creator.IsTeacher {
		return errors.New("only teachers can create announcements")
	}

	// Create announcement
	if err := s.db.Create(announcement).Error; err != nil {
		return fmt.Errorf("failed to create announcement: %w", err)
	}

	return nil
}

// GetAnnouncementsByCourse retrieves announcements for a specific course
func (s *AnnouncementService) GetAnnouncementsByCourse(courseID string, week *int, limit, offset int) ([]models.Announcement, int64, error) {
	var announcements []models.Announcement
	var total int64

	// Count total with optimized query
	countQuery := s.db.Model(&models.Announcement{}).Where("course_id = ?", courseID)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get announcements with optimized query (simplified - no week or pin)
	query := s.db.Table("announcements a").
		Select(`
			a.material_id,
			a.course_id,
			a.title,
			a.description,
			a.week,
			a.is_public,
			a.content,
			a.created_at,
			a.updated_at,
			a.created_by,
			CONCAT(u.first_name, ' ', u.last_name) as creator_name,
			u.email as creator_email
		`).
		Joins("LEFT JOIN users u ON a.created_by = u.user_id").
		Where("a.course_id = ?", courseID)

	// Execute query with optimized pagination
	if err := query.
		Order("a.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&announcements).Error; err != nil {
		return nil, 0, err
	}

	return announcements, total, nil
}

// GetAnnouncementsByCourseCursor retrieves announcements with cursor-based pagination
func (s *AnnouncementService) GetAnnouncementsByCourseCursor(courseID string, week *int, limit int, cursor string) ([]models.Announcement, string, bool, error) {
	var announcements []models.Announcement

	// Parse cursor as timestamp
	var cursorTime time.Time
	var err error
	if cursor != "" {
		cursorTime, err = time.Parse(time.RFC3339, cursor)
		if err != nil {
			return nil, "", false, fmt.Errorf("invalid cursor format: %w", err)
		}
	}

	// Build query with cursor-based pagination (simplified - no week or pin)
	query := s.db.Table("announcements a").
		Select(`
			a.material_id,
			a.course_id,
			a.title,
			a.description,
			a.week,
			a.is_public,
			a.content,
			a.created_at,
			a.updated_at,
			a.created_by,
			CONCAT(u.first_name, ' ', u.last_name) as creator_name,
			u.email as creator_email
		`).
		Joins("LEFT JOIN users u ON a.created_by = u.user_id").
		Where("a.course_id = ?", courseID)

	// Apply cursor-based pagination
	if cursor != "" {
		query = query.Where("a.created_at < ?", cursorTime)
	}

	// Execute query with limit + 1 to check if there are more pages
	if err := query.
		Order("a.created_at DESC").
		Limit(limit + 1).
		Find(&announcements).Error; err != nil {
		return nil, "", false, err
	}

	// Check if there are more pages
	hasNext := len(announcements) > limit
	if hasNext {
		announcements = announcements[:limit] // Remove the extra item
	}

	// Get next cursor
	var nextCursor string
	if hasNext && len(announcements) > 0 {
		nextCursor = announcements[len(announcements)-1].CreatedAt.Format(time.RFC3339)
	}

	return announcements, nextCursor, hasNext, nil
}

// GetAnnouncementByID retrieves a specific announcement
func (s *AnnouncementService) GetAnnouncementByID(announcementID string) (*models.Announcement, error) {
	var announcement models.Announcement
	if err := s.db.Preload("Creator").First(&announcement, "material_id = ?", announcementID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("announcement not found")
		}
		return nil, err
	}

	return &announcement, nil
}

// UpdateAnnouncement updates an existing announcement
func (s *AnnouncementService) UpdateAnnouncement(announcementID string, userID string, updates map[string]interface{}) error {
	// Check if announcement exists and user is the creator
	var announcement models.Announcement
	if err := s.db.First(&announcement, "material_id = ?", announcementID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("announcement not found")
		}
		return err
	}

	// Check if user is the creator
	if announcement.CreatedBy != userID {
		return errors.New("only the creator can update this announcement")
	}

	// Update announcement
	if err := s.db.Model(&announcement).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update announcement: %w", err)
	}

	return nil
}

// DeleteAnnouncement deletes an announcement
func (s *AnnouncementService) DeleteAnnouncement(announcementID string, userID string) error {
	// Check if announcement exists and user is the creator
	var announcement models.Announcement
	if err := s.db.First(&announcement, "material_id = ?", announcementID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("announcement not found")
		}
		return err
	}

	// Check if user is the creator
	if announcement.CreatedBy != userID {
		return errors.New("only the creator can delete this announcement")
	}

	// Delete announcement
	if err := s.db.Delete(&announcement).Error; err != nil {
		return fmt.Errorf("failed to delete announcement: %w", err)
	}

	return nil
}

// GetRecentAnnouncements retrieves recent announcements across all courses for a user
func (s *AnnouncementService) GetRecentAnnouncements(userID string, limit int) ([]models.Announcement, error) {
	var announcements []models.Announcement

	// Get user's enrolled courses
	var enrollments []models.Enrollment
	if err := s.db.Where("user_id = ?", userID).Find(&enrollments).Error; err != nil {
		return nil, err
	}

	if len(enrollments) == 0 {
		return announcements, nil
	}

	// Extract course IDs
	courseIDs := make([]string, len(enrollments))
	for i, enrollment := range enrollments {
		courseIDs[i] = enrollment.CourseID
	}

	// Get recent announcements from enrolled courses
	if err := s.db.Where("course_id IN ?", courseIDs).
		Preload("Creator").
		Preload("Course").
		Order("created_at DESC").
		Limit(limit).
		Find(&announcements).Error; err != nil {
		return nil, err
	}

	return announcements, nil
}

// GetAnnouncementStats retrieves statistics for announcements
func (s *AnnouncementService) GetAnnouncementStats(courseID string) (map[string]interface{}, error) {
	var stats struct {
		TotalAnnouncements int64
		ThisWeekCount      int64
	}

	// Total announcements
	if err := s.db.Model(&models.Announcement{}).Where("course_id = ?", courseID).Count(&stats.TotalAnnouncements).Error; err != nil {
		return nil, err
	}

	// This week's announcements
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	if err := s.db.Model(&models.Announcement{}).Where("course_id = ? AND created_at >= ?", courseID, weekStart).Count(&stats.ThisWeekCount).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_announcements": stats.TotalAnnouncements,
		"this_week_count":     stats.ThisWeekCount,
	}, nil
}
