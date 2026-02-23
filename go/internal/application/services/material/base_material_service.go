package material

import (
	"context"
	"errors"
	"fmt"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"gorm.io/gorm"
)

// BaseMaterialService provides shared functionality for all material services
type BaseMaterialService struct {
	db             *gorm.DB
	storageService storage.StorageInterface
}

// NewBaseMaterialService creates a new base material service
func NewBaseMaterialService(db *gorm.DB, storageService storage.StorageInterface) *BaseMaterialService {
	return &BaseMaterialService{
		db:             db,
		storageService: storageService,
	}
}

// ValidateCourseExists checks if a course exists
func (s *BaseMaterialService) ValidateCourseExists(courseID string) error {
	var course models.Course
	if err := s.db.First(&course, "course_id = ?", courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("course not found")
		}
		return err
	}
	return nil
}

// ValidateCreatorIsTeacher checks if the creator is a teacher
func (s *BaseMaterialService) ValidateCreatorIsTeacher(userID string) error {
	var user models.User
	if err := s.db.First(&user, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}
	if !user.IsTeacher {
		return errors.New("only teachers can create course materials")
	}
	return nil
}

// ValidateMaterialOwnership checks if a user owns a material
func (s *BaseMaterialService) ValidateMaterialOwnership(materialID, userID, materialType string) error {
	var createdBy string
	var err error

	switch materialType {
	case "video":
		var video models.Video
		if err = s.db.First(&video, "material_id = ?", materialID).Error; err != nil {
			return err
		}
		createdBy = video.CreatedBy
	case "document":
		var doc models.Document
		if err = s.db.First(&doc, "material_id = ?", materialID).Error; err != nil {
			return err
		}
		createdBy = doc.CreatedBy
	case "code_exercise":
		var ex models.CodeExercise
		if err = s.db.First(&ex, "material_id = ?", materialID).Error; err != nil {
			return err
		}
		createdBy = ex.CreatedBy
	case "pdf_exercise":
		var ex models.PDFExercise
		if err = s.db.First(&ex, "material_id = ?", materialID).Error; err != nil {
			return err
		}
		createdBy = ex.CreatedBy
	default:
		return fmt.Errorf("unknown material type: %s", materialType)
	}

	if createdBy != userID {
		return errors.New("only the creator can modify this material")
	}

	return nil
}

// GetStorageService returns the storage service
func (s *BaseMaterialService) GetStorageService() storage.StorageInterface {
	return s.storageService
}

// DeleteFile deletes a file from storage
func (s *BaseMaterialService) DeleteFile(ctx context.Context, fileURL string) error {
	if s.storageService != nil {
		return s.storageService.DeleteFile(ctx, fileURL)
	}
	return nil
}

// GetMaterialByID retrieves a material by ID and type
func (s *BaseMaterialService) GetMaterialByID(materialID, materialType string) (models.Material, error) {
	switch materialType {
	case "video":
		var video models.Video
		if err := s.db.Preload("Creator").First(&video, "material_id = ?", materialID).Error; err != nil {
			return nil, err
		}
		return &video, nil

	case "document":
		var doc models.Document
		if err := s.db.Preload("Creator").First(&doc, "material_id = ?", materialID).Error; err != nil {
			return nil, err
		}
		return &doc, nil

	case "code_exercise":
		var ex models.CodeExercise
		if err := s.db.Preload("Creator").Preload("TestCases").First(&ex, "material_id = ?", materialID).Error; err != nil {
			return nil, err
		}
		return &ex, nil

	case "pdf_exercise":
		var ex models.PDFExercise
		if err := s.db.Preload("Creator").First(&ex, "material_id = ?", materialID).Error; err != nil {
			return nil, err
		}
		return &ex, nil

	default:
		return nil, fmt.Errorf("unknown material type: %s", materialType)
	}
}

