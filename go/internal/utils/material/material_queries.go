package material

import (
	"gorm.io/gorm"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
)

// BuildMaterialQuery builds a query for all material types
func BuildMaterialQuery(db *gorm.DB, courseID string, week *int, materialType *string) *gorm.DB {
	var query *gorm.DB

	// Build UNION query for all material types
	if materialType != nil {
		// Query specific type
		switch enums.MaterialType(*materialType) {
		case enums.MaterialTypeVideo:
			query = db.Model(&models.Video{}).Where("course_id = ?", courseID)
		case enums.MaterialTypeDocument:
			query = db.Model(&models.Document{}).Where("course_id = ?", courseID)
		case enums.MaterialTypeCodeExercise:
			query = db.Model(&models.CodeExercise{}).Where("course_id = ?", courseID)
		case enums.MaterialTypePDFExercise:
			query = db.Model(&models.PDFExercise{}).Where("course_id = ?", courseID)
		default:
			// Return empty query
			return db.Where("1 = 0")
		}
	} else {
		// For all types, we'll need to query separately and combine
		// This is a placeholder - actual implementation will query each type
		query = db.Model(&models.Video{}).Where("course_id = ?", courseID)
	}

	// Filter by week if specified
	if week != nil {
		query = query.Where("week = ?", *week)
	}

	return query
}

// GetAllMaterialsByCourse queries all material types for a course
func GetAllMaterialsByCourse(db *gorm.DB, courseID string, week *int, materialType *string, limit, offset int) ([]models.Material, int64, error) {
	var allMaterials []models.Material
	var total int64

	// If specific type is requested, query only that type
	if materialType != nil {
		switch enums.MaterialType(*materialType) {
		case enums.MaterialTypeVideo:
			var videos []models.Video
			query := db.Model(&models.Video{}).Where("course_id = ?", courseID)
			if week != nil {
				query = query.Where("week = ?", *week)
			}
			if err := query.Count(&total).Error; err != nil {
				return nil, 0, err
			}
			if err := query.Preload("Creator").Order("week ASC, created_at DESC").
				Limit(limit).Offset(offset).Find(&videos).Error; err != nil {
				return nil, 0, err
			}
			for i := range videos {
				allMaterials = append(allMaterials, &videos[i])
			}

		case enums.MaterialTypeDocument:
			var documents []models.Document
			query := db.Model(&models.Document{}).Where("course_id = ?", courseID)
			if week != nil {
				query = query.Where("week = ?", *week)
			}
			if err := query.Count(&total).Error; err != nil {
				return nil, 0, err
			}
			if err := query.Preload("Creator").Order("week ASC, created_at DESC").
				Limit(limit).Offset(offset).Find(&documents).Error; err != nil {
				return nil, 0, err
			}
			for i := range documents {
				allMaterials = append(allMaterials, &documents[i])
			}

		case enums.MaterialTypeCodeExercise:
			var exercises []models.CodeExercise
			query := db.Model(&models.CodeExercise{}).Where("course_id = ?", courseID)
			if week != nil {
				query = query.Where("week = ?", *week)
			}
			if err := query.Count(&total).Error; err != nil {
				return nil, 0, err
			}
			if err := query.Preload("Creator").Order("week ASC, created_at DESC").
				Limit(limit).Offset(offset).Find(&exercises).Error; err != nil {
				return nil, 0, err
			}
			for i := range exercises {
				allMaterials = append(allMaterials, &exercises[i])
			}

		case enums.MaterialTypePDFExercise:
			var exercises []models.PDFExercise
			query := db.Model(&models.PDFExercise{}).Where("course_id = ?", courseID)
			if week != nil {
				query = query.Where("week = ?", *week)
			}
			if err := query.Count(&total).Error; err != nil {
				return nil, 0, err
			}
			if err := query.Preload("Creator").Order("week ASC, created_at DESC").
				Limit(limit).Offset(offset).Find(&exercises).Error; err != nil {
				return nil, 0, err
			}
			for i := range exercises {
				allMaterials = append(allMaterials, &exercises[i])
			}
		}
	} else {
		// Query all types and combine
		var videos []models.Video
		var documents []models.Document
		var codeExercises []models.CodeExercise
		var pdfExercises []models.PDFExercise

		// Query each type
		videoQuery := db.Model(&models.Video{}).Where("course_id = ?", courseID)
		docQuery := db.Model(&models.Document{}).Where("course_id = ?", courseID)
		codeExQuery := db.Model(&models.CodeExercise{}).Where("course_id = ?", courseID)
		pdfExQuery := db.Model(&models.PDFExercise{}).Where("course_id = ?", courseID)

		if week != nil {
			videoQuery = videoQuery.Where("week = ?", *week)
			docQuery = docQuery.Where("week = ?", *week)
			codeExQuery = codeExQuery.Where("week = ?", *week)
			pdfExQuery = pdfExQuery.Where("week = ?", *week)
		}

		// Count totals
		var videoCount, docCount, codeExCount, pdfExCount int64
		videoQuery.Count(&videoCount)
		docQuery.Count(&docCount)
		codeExQuery.Count(&codeExCount)
		pdfExQuery.Count(&pdfExCount)
		total = videoCount + docCount + codeExCount + pdfExCount

		// Fetch all (we'll need to sort and paginate in memory or use a more complex query)
		// For now, fetch with limit/offset per type
		videoQuery.Preload("Creator").Order("week ASC, created_at DESC").Find(&videos)
		docQuery.Preload("Creator").Order("week ASC, created_at DESC").Find(&documents)
		codeExQuery.Preload("Creator").Order("week ASC, created_at DESC").Find(&codeExercises)
		pdfExQuery.Preload("Creator").Order("week ASC, created_at DESC").Find(&pdfExercises)

		// Convert to Material interface
		for i := range videos {
			allMaterials = append(allMaterials, &videos[i])
		}
		for i := range documents {
			allMaterials = append(allMaterials, &documents[i])
		}
		for i := range codeExercises {
			allMaterials = append(allMaterials, &codeExercises[i])
		}
		for i := range pdfExercises {
			allMaterials = append(allMaterials, &pdfExercises[i])
		}
	}

	return allMaterials, total, nil
}


















