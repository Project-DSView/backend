package services

import (
	"fmt"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"gorm.io/gorm"
)

type DraftService struct {
	db *gorm.DB
}

func NewDraftService(db *gorm.DB) *DraftService {
	return &DraftService{db: db}
}

func (s *DraftService) SaveDraft(userID, materialID, code, fileName string, fileSize int64) (*models.ExerciseDraft, error) {
	// หา draft ที่มีอยู่หรือสร้างใหม่
	var draft models.ExerciseDraft
	err := s.db.Where("user_id = ? AND material_id = ?", userID, materialID).First(&draft).Error

	if err == gorm.ErrRecordNotFound {
		// สร้างใหม่
		draft = models.ExerciseDraft{
			UserID:     userID,
			MaterialID: materialID,
			Code:       code,
			FileName:   fileName,
			FileSize:   fileSize,
		}
		if err := s.db.Create(&draft).Error; err != nil {
			return nil, fmt.Errorf("failed to create draft: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to query draft: %w", err)
	} else {
		// อัพเดทที่มีอยู่
		updates := map[string]interface{}{
			"code":      code,
			"file_name": fileName,
			"file_size": fileSize,
		}
		if err := s.db.Model(&draft).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update draft: %w", err)
		}
	}

	return &draft, nil
}

func (s *DraftService) GetDraft(userID, materialID string) (*models.ExerciseDraft, error) {
	var draft models.ExerciseDraft
	err := s.db.Where("user_id = ? AND material_id = ?", userID, materialID).First(&draft).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // ไม่มี draft
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get draft: %w", err)
	}
	return &draft, nil
}

func (s *DraftService) DeleteDraft(userID, materialID string) error {
	result := s.db.Where("user_id = ? AND material_id = ?", userID, materialID).Delete(&models.ExerciseDraft{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete draft: %w", result.Error)
	}
	return nil
}

func (s *DraftService) GetUserDrafts(userID string) ([]models.ExerciseDraft, error) {
	var drafts []models.ExerciseDraft
	if err := s.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&drafts).Error; err != nil {
		return nil, fmt.Errorf("failed to get user drafts: %w", err)
	}
	return drafts, nil
}
