package material

import (
	"errors"
	"fmt"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/utils/material"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"gorm.io/gorm"
)

// VideoService handles video material operations
type VideoService struct {
	*BaseMaterialService
}

// NewVideoService creates a new video service
func NewVideoService(db *gorm.DB, storageService storage.StorageInterface) *VideoService {
	return &VideoService{
		BaseMaterialService: NewBaseMaterialService(db, storageService),
	}
}

// CreateVideo creates a new video material
func (s *VideoService) CreateVideo(video *models.Video) error {
	// Validate course exists
	if err := s.ValidateCourseExists(video.CourseID); err != nil {
		return err
	}

	// Validate creator is teacher
	if err := s.ValidateCreatorIsTeacher(video.CreatedBy); err != nil {
		return err
	}

	// Validate video fields
	if err := material.ValidateVideoFields(video.VideoURL); err != nil {
		return err
	}

	// Create video
	if err := s.db.Create(video).Error; err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}

	return nil
}

// GetVideoByID retrieves a video by ID
func (s *VideoService) GetVideoByID(videoID string) (*models.Video, error) {
	var video models.Video
	if err := s.db.Preload("Creator").Preload("Course").First(&video, "material_id = ?", videoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("video not found")
		}
		return nil, err
	}
	return &video, nil
}

// GetVideosByCourse retrieves videos for a specific course
func (s *VideoService) GetVideosByCourse(courseID string, week *int, limit, offset int) ([]models.Video, int64, error) {
	var videos []models.Video
	var total int64

	query := s.db.Model(&models.Video{}).Where("course_id = ?", courseID)

	if week != nil {
		query = query.Where("week = ?", *week)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Creator").
		Order("week ASC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// UpdateVideo updates an existing video
func (s *VideoService) UpdateVideo(videoID string, userID string, updates map[string]interface{}) error {
	// Validate ownership
	if err := s.ValidateMaterialOwnership(videoID, userID, "video"); err != nil {
		return err
	}

	// Update video
	if err := s.db.Model(&models.Video{}).Where("material_id = ?", videoID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	return nil
}

// DeleteVideo deletes a video
func (s *VideoService) DeleteVideo(videoID string, userID string) error {
	// Validate ownership
	if err := s.ValidateMaterialOwnership(videoID, userID, "video"); err != nil {
		return err
	}

	// Delete video
	if err := s.db.Delete(&models.Video{}, "material_id = ?", videoID).Error; err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	return nil
}

