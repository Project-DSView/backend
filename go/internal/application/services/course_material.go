package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/pkg/logger"
	materialpkg "github.com/Project-DSView/backend/go/pkg/material"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"gorm.io/gorm"
)

type CourseMaterialService struct {
	db             *gorm.DB
	storageService storage.StorageInterface
}

func NewCourseMaterialService(db *gorm.DB, storageService storage.StorageInterface) *CourseMaterialService {
	return &CourseMaterialService{
		db:             db,
		storageService: storageService,
	}
}

// GetDB returns the database instance (for handler access)
func (s *CourseMaterialService) GetDB() *gorm.DB {
	return s.db
}

// CreateCourseMaterial creates a new course material (central reference table entry)
func (s *CourseMaterialService) CreateCourseMaterial(material *models.CourseMaterial) error {
	// Validate material type
	if !models.IsValidMaterialType(string(material.Type)) {
		return errors.New("invalid material type")
	}

	// Check if course exists
	var course models.Course
	if err := s.db.First(&course, "course_id = ?", material.CourseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("course not found")
		}
		return err
	}

	// Validate that reference_id and reference_type are set
	if material.ReferenceID == nil || material.ReferenceType == nil {
		return errors.New("reference_id and reference_type are required")
	}

	// Verify that the referenced material exists
	if err := materialpkg.VerifyReferenceExists(s.db, *material.ReferenceID, *material.ReferenceType); err != nil {
		return err
	}

	// Create material reference
	if err := s.db.Create(material).Error; err != nil {
		return fmt.Errorf("failed to create course material: %w", err)
	}

	return nil
}

// UploadCourseMaterialFile uploads a file for course material
func (s *CourseMaterialService) UploadCourseMaterialFile(ctx context.Context, courseID, userID, materialType string, file io.Reader, filename, contentType string) (string, error) {
	// Check if user is a teacher
	var user models.User
	if err := s.db.First(&user, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("user not found")
		}
		return "", err
	}

	if !user.IsTeacher {
		return "", errors.New("only teachers can upload course materials")
	}

	// Check if course exists
	var course models.Course
	if err := s.db.First(&course, "course_id = ?", courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("course not found")
		}
		return "", err
	}

	// Generate unique filename
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".bin"
	}
	uniqueFilename := fmt.Sprintf("%s_%s_%d%s",
		courseID,
		time.Now().Format("20060102_150405"),
		time.Now().UnixNano()%1000000,
		ext)

	// Upload to storage
	key := fmt.Sprintf("course-materials/%s/%s", courseID, uniqueFilename)
	url, err := s.storageService.UploadFile(ctx, key, file, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return url, nil
}

// GetCourseMaterialsByCourse retrieves materials for a specific course with full details
func (s *CourseMaterialService) GetCourseMaterialsByCourse(courseID string, week *int, materialType *string, limit, offset int) ([]map[string]interface{}, int64, error) {
	var materials []models.CourseMaterial
	var total int64

	query := s.db.Model(&models.CourseMaterial{}).Where("course_id = ?", courseID)

	// Filter by week if specified
	if week != nil {
		query = query.Where("week = ?", *week)
	}

	// Filter by material type if specified
	if materialType != nil {
		query = query.Where("type = ?", *materialType)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get materials (CourseMaterial is a reference table, no Creator relation)
	if err := query.
		Order("week ASC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&materials).Error; err != nil {
		return nil, 0, err
	}

	// Get full details for each material
	materialsWithDetails := make([]map[string]interface{}, 0, len(materials))
	for i := range materials {
		details, err := materialpkg.GetMaterialWithDetails(s.db, &materials[i])
		if err != nil {
			// Log error but continue with other materials
			logger.Warnf("Failed to get details for material %s: %v", materials[i].MaterialID, err)
			// Fallback to basic CourseMaterial data
			details = materials[i].ToJSON()
		}
		materialsWithDetails = append(materialsWithDetails, details)
	}

	return materialsWithDetails, total, nil
}

// GetCourseMaterialByID retrieves a specific course material with full details
func (s *CourseMaterialService) GetCourseMaterialByID(materialID string) (map[string]interface{}, error) {
	var material models.CourseMaterial
	if err := s.db.First(&material, "material_id = ?", materialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("course material not found")
		}
		return nil, err
	}

	// Get full details
	details, err := materialpkg.GetMaterialWithDetails(s.db, &material)
	if err != nil {
		// Fallback to basic CourseMaterial data
		return material.ToJSON(), fmt.Errorf("failed to get material details: %w", err)
	}

	return details, nil
}

// UpdateCourseMaterial updates an existing course material
func (s *CourseMaterialService) UpdateCourseMaterial(materialID string, userID string, updates map[string]interface{}) error {
	// Check if material exists
	var material models.CourseMaterial
	if err := s.db.First(&material, "material_id = ?", materialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("course material not found")
		}
		return err
	}

	// Get the actual material to check creator
	var createdBy string
	if material.ReferenceID != nil && material.ReferenceType != nil {
		createdBy, _ = materialpkg.GetMaterialCreator(s.db, *material.ReferenceID, *material.ReferenceType)
	}

	// Check if user is the creator
	if createdBy != "" && createdBy != userID {
		return errors.New("only the creator can update this course material")
	}

	// Validate material type if provided
	if materialType, ok := updates["type"].(string); ok {
		if !models.IsValidMaterialType(materialType) {
			return errors.New("invalid material type")
		}
	}

	// Filter out fields that don't exist in CourseMaterial table
	// CourseMaterial only has: material_id, course_id, type, week, reference_id, reference_type, created_at, updated_at
	// Fields like file_url, video_url are stored in specific material tables (Document, Video, etc.)
	allowedFields := map[string]bool{
		"title":       false, // Not in CourseMaterial, stored in specific tables
		"description": false, // Not in CourseMaterial, stored in specific tables
		"type":        true,
		"week":        true,
		"is_public":   false, // Not in CourseMaterial, stored in specific tables
		"video_url":   false, // Not in CourseMaterial, stored in Video table
		"file_url":    false, // Not in CourseMaterial, stored in Document/PDFExercise tables
	}

	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowed, exists := allowedFields[key]; exists && allowed {
			filteredUpdates[key] = value
		} else if !exists {
			// Unknown field, skip it to avoid errors
			continue
		}
		// If field is not allowed (exists but false), skip it
	}

	// Only update if there are valid fields to update
	if len(filteredUpdates) > 0 {
		if err := s.db.Model(&material).Updates(filteredUpdates).Error; err != nil {
			return fmt.Errorf("failed to update course material: %w", err)
		}
	}

	// Update specific material table based on type
	if material.ReferenceID != nil && material.ReferenceType != nil {
		specificUpdates := make(map[string]interface{})

		// Collect fields that belong to specific material tables
		if title, ok := updates["title"].(string); ok && title != "" {
			specificUpdates["title"] = title
		}
		if description, ok := updates["description"].(string); ok {
			specificUpdates["description"] = description
		}
		if isPublic, ok := updates["is_public"].(bool); ok {
			specificUpdates["is_public"] = isPublic
		}
		if week, ok := updates["week"].(int); ok {
			specificUpdates["week"] = week
		}

		// Type-specific fields
		switch *material.ReferenceType {
		case "document":
			// Document doesn't have additional fields to update here
			// File updates would be handled separately if needed
		case "video":
			if videoURL, ok := updates["video_url"].(string); ok && videoURL != "" {
				specificUpdates["video_url"] = videoURL
			}
		case "announcement":
			if content, ok := updates["content"].(string); ok && content != "" {
				specificUpdates["content"] = content
			}
		case "code_exercise":
			if totalPoints, ok := updates["total_points"].(int); ok {
				specificUpdates["total_points"] = totalPoints
			}
			if deadline, ok := updates["deadline"].(string); ok {
				specificUpdates["deadline"] = deadline
			}
			if problemStatement, ok := updates["problem_statement"].(string); ok {
				specificUpdates["problem_statement"] = problemStatement
			}
			if constraints, ok := updates["constraints"].(string); ok {
				specificUpdates["constraints"] = constraints
			}
			if hints, ok := updates["hints"].(string); ok {
				specificUpdates["hints"] = hints
			}
			if exampleInputs, ok := updates["example_inputs"]; ok {
				specificUpdates["example_inputs"] = exampleInputs
			}
			if exampleOutputs, ok := updates["example_outputs"]; ok {
				specificUpdates["example_outputs"] = exampleOutputs
			}
		case "pdf_exercise":
			if totalPoints, ok := updates["total_points"].(int); ok {
				specificUpdates["total_points"] = totalPoints
			}
			if deadline, ok := updates["deadline"].(string); ok {
				specificUpdates["deadline"] = deadline
			}
		}

		// Update the specific material table if there are fields to update
		if len(specificUpdates) > 0 {
			switch *material.ReferenceType {
			case "document":
				if err := s.db.Model(&models.Document{}).Where("material_id = ?", materialID).Updates(specificUpdates).Error; err != nil {
					return fmt.Errorf("failed to update document: %w", err)
				}
			case "video":
				if err := s.db.Model(&models.Video{}).Where("material_id = ?", materialID).Updates(specificUpdates).Error; err != nil {
					return fmt.Errorf("failed to update video: %w", err)
				}
			case "announcement":
				if err := s.db.Model(&models.Announcement{}).Where("material_id = ?", materialID).Updates(specificUpdates).Error; err != nil {
					return fmt.Errorf("failed to update announcement: %w", err)
				}
			case "code_exercise":
				if err := s.db.Model(&models.CodeExercise{}).Where("material_id = ?", materialID).Updates(specificUpdates).Error; err != nil {
					return fmt.Errorf("failed to update code exercise: %w", err)
				}
			case "pdf_exercise":
				if err := s.db.Model(&models.PDFExercise{}).Where("material_id = ?", materialID).Updates(specificUpdates).Error; err != nil {
					return fmt.Errorf("failed to update PDF exercise: %w", err)
				}
			}
		}
	}

	return nil
}

// DeleteCourseMaterial deletes a course material
func (s *CourseMaterialService) DeleteCourseMaterial(materialID string, userID string) error {
	// Check if material exists
	var material models.CourseMaterial
	if err := s.db.First(&material, "material_id = ?", materialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("course material not found")
		}
		return err
	}

	// Get the actual material to check creator and get file URL
	var createdBy string
	var fileURL string
	if material.ReferenceID != nil && material.ReferenceType != nil {
		createdBy, _ = materialpkg.GetMaterialCreator(s.db, *material.ReferenceID, *material.ReferenceType)
		fileURL, _ = materialpkg.GetMaterialFileURL(s.db, *material.ReferenceID, *material.ReferenceType)
	}

	// Check if user is the creator
	if createdBy != "" && createdBy != userID {
		return errors.New("only the creator can delete this course material")
	}

	// Delete file from storage if exists
	if fileURL != "" {
		ctx := context.Background()
		if err := s.storageService.DeleteFile(ctx, fileURL); err != nil {
			// Log error but don't fail the deletion
			logger.Warnf("Failed to delete file from storage: %v", err)
		}
	}

	// Delete material from database
	if err := s.db.Delete(&material).Error; err != nil {
		return fmt.Errorf("failed to delete course material: %w", err)
	}

	return nil
}

// GetCourseMaterialsByWeek retrieves materials for a specific week
func (s *CourseMaterialService) GetCourseMaterialsByWeek(courseID string, week int) ([]models.CourseMaterial, error) {
	var materials []models.CourseMaterial
	if err := s.db.Where("course_id = ? AND week = ?", courseID, week).
		Order("type ASC, created_at DESC").
		Find(&materials).Error; err != nil {
		return nil, err
	}

	return materials, nil
}

// GetCourseMaterialsByType retrieves materials by type
func (s *CourseMaterialService) GetCourseMaterialsByType(courseID string, materialType string) ([]models.CourseMaterial, error) {
	if !models.IsValidMaterialType(materialType) {
		return nil, errors.New("invalid material type")
	}

	var materials []models.CourseMaterial
	if err := s.db.Where("course_id = ? AND type = ?", courseID, materialType).
		Order("week ASC, created_at DESC").
		Find(&materials).Error; err != nil {
		return nil, err
	}

	return materials, nil
}

// GetCourseMaterialStats retrieves statistics for course materials
func (s *CourseMaterialService) GetCourseMaterialStats(courseID string) (map[string]interface{}, error) {
	var stats struct {
		TotalMaterials int64
		ByType         map[string]int64
		ByWeek         map[int]int64
		TotalFileSize  int64
	}

	// Total materials
	if err := s.db.Model(&models.CourseMaterial{}).Where("course_id = ?", courseID).Count(&stats.TotalMaterials).Error; err != nil {
		return nil, err
	}

	// Count by type
	stats.ByType = make(map[string]int64)
	types := []string{"document", "video", "exercise", "assignment", "reference"}
	for _, materialType := range types {
		var count int64
		if err := s.db.Model(&models.CourseMaterial{}).Where("course_id = ? AND type = ?", courseID, materialType).Count(&count).Error; err != nil {
			return nil, err
		}
		stats.ByType[materialType] = count
	}

	// Count by week
	stats.ByWeek = make(map[int]int64)
	var weeks []int
	if err := s.db.Model(&models.CourseMaterial{}).Where("course_id = ?", courseID).
		Distinct("week").Pluck("week", &weeks).Error; err != nil {
		return nil, err
	}

	for _, week := range weeks {
		var count int64
		if err := s.db.Model(&models.CourseMaterial{}).Where("course_id = ? AND week = ?", courseID, week).Count(&count).Error; err != nil {
			return nil, err
		}
		stats.ByWeek[week] = count
	}

	// Total file size
	if err := s.db.Model(&models.CourseMaterial{}).Where("course_id = ?", courseID).
		Select("COALESCE(SUM(file_size), 0)").Scan(&stats.TotalFileSize).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_materials": stats.TotalMaterials,
		"by_type":         stats.ByType,
		"by_week":         stats.ByWeek,
		"total_file_size": stats.TotalFileSize,
	}, nil
}

// SearchCourseMaterials searches materials by title or description
func (s *CourseMaterialService) SearchCourseMaterials(courseID string, query string, limit, offset int) ([]models.CourseMaterial, int64, error) {
	var materials []models.CourseMaterial
	var total int64

	searchQuery := s.db.Model(&models.CourseMaterial{}).Where("course_id = ? AND (title ILIKE ? OR description ILIKE ?)",
		courseID, "%"+query+"%", "%"+query+"%")

	// Count total
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get materials
	if err := searchQuery.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&materials).Error; err != nil {
		return nil, 0, err
	}

	return materials, total, nil
}

// Test Case Management Methods

// CreateCodeExercise creates a code exercise material with test cases
// Note: This function creates the CodeExercise first, then creates the CourseMaterial reference
func (s *CourseMaterialService) CreateCodeExercise(codeExercise *models.CodeExercise, testCases []models.TestCase) error {
	// Validate that code exercise has required fields
	if codeExercise.TotalPoints == nil || *codeExercise.TotalPoints <= 0 {
		return errors.New("code exercise must have total points")
	}

	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create code exercise first
		if err := tx.Create(codeExercise).Error; err != nil {
			return fmt.Errorf("failed to create code exercise: %w", err)
		}

		// Create CourseMaterial reference
		referenceType := "code_exercise"
		courseMaterial := &models.CourseMaterial{
			MaterialID:    codeExercise.MaterialID, // Use the same ID
			CourseID:      codeExercise.CourseID,
			Type:          enums.MaterialTypeCodeExercise,
			Week:          codeExercise.Week,
			ReferenceID:   &codeExercise.MaterialID,
			ReferenceType: &referenceType,
		}
		if err := tx.Create(courseMaterial).Error; err != nil {
			return fmt.Errorf("failed to create course material reference: %w", err)
		}

		// Create test cases
		for i := range testCases {
			materialID := codeExercise.MaterialID
			testCases[i].MaterialID = &materialID
			if err := tx.Create(&testCases[i]).Error; err != nil {
				return fmt.Errorf("failed to create test case %d: %w", i+1, err)
			}
		}

		return nil
	})
}

// UpdateCodeExercise updates a code exercise
func (s *CourseMaterialService) UpdateCodeExercise(materialID string, userID string, updates map[string]interface{}) error {
	// Check if material exists and user is creator
	var material models.CourseMaterial
	if err := s.db.First(&material, "material_id = ?", materialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("course material not found")
		}
		return err
	}

	// Get the actual material to check creator
	var createdBy string
	if material.ReferenceID != nil && material.ReferenceType != nil {
		createdBy, _ = materialpkg.GetMaterialCreator(s.db, *material.ReferenceID, *material.ReferenceType)
	}

	// Check if user is the creator
	if createdBy != "" && createdBy != userID {
		return errors.New("only the creator can update this course material")
	}

	// Update code exercise
	if err := s.db.Model(&models.CodeExercise{}).Where("material_id = ?", materialID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update code exercise: %w", err)
	}

	return nil
}

// CreateDocument creates a document material with CourseMaterial reference
func (s *CourseMaterialService) CreateDocument(document *models.Document) error {
	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create document first
		if err := tx.Create(document).Error; err != nil {
			return fmt.Errorf("failed to create document: %w", err)
		}

		// Create CourseMaterial reference
		referenceType := "document"
		courseMaterial := &models.CourseMaterial{
			MaterialID:    document.MaterialID, // Use the same ID
			CourseID:      document.CourseID,
			Type:          enums.MaterialTypeDocument,
			Week:          document.Week,
			ReferenceID:   &document.MaterialID,
			ReferenceType: &referenceType,
		}
		if err := tx.Create(courseMaterial).Error; err != nil {
			return fmt.Errorf("failed to create course material reference: %w", err)
		}

		return nil
	})
}

// CreateVideo creates a video material with CourseMaterial reference
func (s *CourseMaterialService) CreateVideo(video *models.Video) error {
	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create video first
		if err := tx.Create(video).Error; err != nil {
			return fmt.Errorf("failed to create video: %w", err)
		}

		// Create CourseMaterial reference
		referenceType := "video"
		courseMaterial := &models.CourseMaterial{
			MaterialID:    video.MaterialID, // Use the same ID
			CourseID:      video.CourseID,
			Type:          enums.MaterialTypeVideo,
			Week:          video.Week,
			ReferenceID:   &video.MaterialID,
			ReferenceType: &referenceType,
		}
		if err := tx.Create(courseMaterial).Error; err != nil {
			return fmt.Errorf("failed to create course material reference: %w", err)
		}

		return nil
	})
}

// CreatePDFExercise creates a PDF exercise material with CourseMaterial reference
func (s *CourseMaterialService) CreatePDFExercise(pdfExercise *models.PDFExercise) error {
	// Validate that PDF exercise has required fields
	if pdfExercise.TotalPoints == nil || *pdfExercise.TotalPoints <= 0 {
		return errors.New("PDF exercise must have total points")
	}

	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create PDF exercise first
		if err := tx.Create(pdfExercise).Error; err != nil {
			return fmt.Errorf("failed to create PDF exercise: %w", err)
		}

		// Create CourseMaterial reference
		referenceType := "pdf_exercise"
		courseMaterial := &models.CourseMaterial{
			MaterialID:    pdfExercise.MaterialID, // Use the same ID
			CourseID:      pdfExercise.CourseID,
			Type:          enums.MaterialTypePDFExercise,
			Week:          pdfExercise.Week,
			ReferenceID:   &pdfExercise.MaterialID,
			ReferenceType: &referenceType,
		}
		if err := tx.Create(courseMaterial).Error; err != nil {
			return fmt.Errorf("failed to create course material reference: %w", err)
		}

		return nil
	})
}

// CreateAnnouncement creates an announcement material with CourseMaterial reference
func (s *CourseMaterialService) CreateAnnouncement(announcement *models.Announcement) error {
	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create announcement first
		if err := tx.Create(announcement).Error; err != nil {
			return fmt.Errorf("failed to create announcement: %w", err)
		}

		// Create CourseMaterial reference
		referenceType := "announcement"
		courseMaterial := &models.CourseMaterial{
			MaterialID:    announcement.MaterialID, // Use the same ID
			CourseID:      announcement.CourseID,
			Type:          enums.MaterialTypeAnnouncement,
			Week:          announcement.Week,
			ReferenceID:   &announcement.MaterialID,
			ReferenceType: &referenceType,
		}
		if err := tx.Create(courseMaterial).Error; err != nil {
			return fmt.Errorf("failed to create course material reference: %w", err)
		}

		return nil
	})
}

// GetTestCases retrieves test cases for a specific material (only for code exercises)
func (s *CourseMaterialService) GetTestCases(materialID string) ([]models.TestCase, error) {
	// Check if material exists and is a code exercise
	var material models.CourseMaterial
	if err := s.db.First(&material, "material_id = ?", materialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("course material not found")
		}
		return nil, err
	}

	// Only code exercises have test cases
	if !material.IsCodeExercise() {
		return nil, errors.New("test cases are only available for code exercises")
	}

	var testCases []models.TestCase
	if err := s.db.Where("material_id = ?", materialID).Find(&testCases).Error; err != nil {
		return nil, fmt.Errorf("failed to get test cases: %w", err)
	}

	return testCases, nil
}

// AddTestCase adds a test case to a material
func (s *CourseMaterialService) AddTestCase(materialID string, testCase *models.TestCase) error {
	// Check if material exists and is a code exercise
	var material models.CourseMaterial
	if err := s.db.First(&material, "material_id = ?", materialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("course material not found")
		}
		return err
	}

	if !material.IsCodeExercise() {
		return errors.New("can only add test cases to code exercises")
	}

	// Set material ID
	testCase.MaterialID = &materialID

	// Create test case
	if err := s.db.Create(testCase).Error; err != nil {
		return fmt.Errorf("failed to create test case: %w", err)
	}

	return nil
}

// UpdateTestCase updates an existing test case
func (s *CourseMaterialService) UpdateTestCase(testCaseID string, userID string, updates map[string]interface{}) error {
	// Check if test case exists
	var testCase models.TestCase
	if err := s.db.First(&testCase, "test_case_id = ?", testCaseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("test case not found")
		}
		return err
	}

	// Get the material to check creator and type
	var material models.CourseMaterial
	if err := s.db.First(&material, "material_id = ?", testCase.MaterialID).Error; err != nil {
		return errors.New("material not found")
	}

	// Only code exercises have test cases
	if !material.IsCodeExercise() {
		return errors.New("test cases can only be updated for code exercises")
	}

	// Get creator from actual material table
	var createdBy string
	if material.ReferenceID != nil && material.ReferenceType != nil {
		createdBy, _ = materialpkg.GetMaterialCreator(s.db, *material.ReferenceID, *material.ReferenceType)
	}

	// Check if user is the creator of the material
	if createdBy != "" && createdBy != userID {
		return errors.New("only the creator can update test cases")
	}

	// Update test case
	if err := s.db.Model(&testCase).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update test case: %w", err)
	}

	return nil
}

// DeleteTestCase deletes a test case
func (s *CourseMaterialService) DeleteTestCase(testCaseID string, userID string) error {
	// Check if test case exists
	var testCase models.TestCase
	if err := s.db.First(&testCase, "test_case_id = ?", testCaseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("test case not found")
		}
		return err
	}

	// Get the material to check creator and type
	var material models.CourseMaterial
	if err := s.db.First(&material, "material_id = ?", testCase.MaterialID).Error; err != nil {
		return errors.New("material not found")
	}

	// Only code exercises have test cases
	if !material.IsCodeExercise() {
		return errors.New("test cases can only be deleted for code exercises")
	}

	// Get creator from actual material table
	var createdBy string
	if material.ReferenceID != nil && material.ReferenceType != nil {
		createdBy, _ = materialpkg.GetMaterialCreator(s.db, *material.ReferenceID, *material.ReferenceType)
	}

	// Check if user is the creator of the material
	if createdBy != "" && createdBy != userID {
		return errors.New("only the creator can delete test cases")
	}

	// Delete test case
	if err := s.db.Delete(&testCase).Error; err != nil {
		return fmt.Errorf("failed to delete test case: %w", err)
	}

	return nil
}

// GetCourseMaterialWithTestCases retrieves a material with its test cases (only for code exercises)
func (s *CourseMaterialService) GetCourseMaterialWithTestCases(materialID string) (*models.CourseMaterial, []models.TestCase, error) {
	var material models.CourseMaterial
	if err := s.db.First(&material, "material_id = ?", materialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("course material not found")
		}
		return nil, nil, err
	}

	// Only code exercises have test cases
	if !material.IsCodeExercise() {
		return &material, nil, nil
	}

	// Get test cases for code exercise
	var testCases []models.TestCase
	if err := s.db.Where("material_id = ?", materialID).Find(&testCases).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get test cases: %w", err)
	}

	return &material, testCases, nil
}
