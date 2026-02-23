package material

import (
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"gorm.io/gorm"
)

// MaterialServiceFactory creates material services based on type
type MaterialServiceFactory struct {
	db             *gorm.DB
	storageService storage.StorageInterface
}

// NewMaterialServiceFactory creates a new material service factory
func NewMaterialServiceFactory(db *gorm.DB, storageService storage.StorageInterface) *MaterialServiceFactory {
	return &MaterialServiceFactory{
		db:             db,
		storageService: storageService,
	}
}

// GetService returns the appropriate service for a material type
func (f *MaterialServiceFactory) GetService(materialType string) interface{} {
	switch enums.MaterialType(materialType) {
	case enums.MaterialTypeVideo:
		return NewVideoService(f.db, f.storageService)
	case enums.MaterialTypeDocument:
		return NewDocumentService(f.db, f.storageService)
	case enums.MaterialTypeCodeExercise:
		return NewCodeExerciseService(f.db, f.storageService)
	case enums.MaterialTypePDFExercise:
		return NewPDFExerciseService(f.db, f.storageService)
	default:
		return nil
	}
}

// GetVideoService returns the video service
func (f *MaterialServiceFactory) GetVideoService() *VideoService {
	return NewVideoService(f.db, f.storageService)
}

// GetDocumentService returns the document service
func (f *MaterialServiceFactory) GetDocumentService() *DocumentService {
	return NewDocumentService(f.db, f.storageService)
}

// GetCodeExerciseService returns the code exercise service
func (f *MaterialServiceFactory) GetCodeExerciseService() *CodeExerciseService {
	return NewCodeExerciseService(f.db, f.storageService)
}

// GetPDFExerciseService returns the PDF exercise service
func (f *MaterialServiceFactory) GetPDFExerciseService() *PDFExerciseService {
	return NewPDFExerciseService(f.db, f.storageService)
}


















