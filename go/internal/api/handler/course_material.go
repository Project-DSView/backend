package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Project-DSView/backend/go/internal/api/types"
	"github.com/Project-DSView/backend/go/internal/application/services"
	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	internaltypes "github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/config"
	"github.com/Project-DSView/backend/go/pkg/enrollment"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"github.com/gofiber/fiber/v2"
)

type CourseMaterialHandler struct {
	materialService     *services.CourseMaterialService
	enrollmentValidator *enrollment.EnrollmentValidator
	storageService      storage.StorageService
}

func NewCourseMaterialHandler(materialService *services.CourseMaterialService, enrollmentValidator *enrollment.EnrollmentValidator, storageService storage.StorageService) *CourseMaterialHandler {
	return &CourseMaterialHandler{
		materialService:     materialService,
		enrollmentValidator: enrollmentValidator,
		storageService:      storageService,
	}
}

// Request models are now defined in internal/api/types/requests.go

// CreateCourseMaterial creates a new course material
// @Summary Create course material
// @Description Create a new course material for a course (teachers only)
// @Tags course-materials
// @Accept multipart/form-data
// @Produce json
// @Param CourseID formData string true "Course ID"
// @Param Title formData string true "Material title"
// @Param Description formData string false "Material description"
// @Param Type formData string true "Material type" Enums(pdf_exercise,code_exercise,document,video)
// @Param Week formData int false "Week number"
// @Param IsPublic formData bool false "Is public"
// @Param TotalPoints formData int false "Total points"
// @Param Deadline formData string false "Deadline (ISO 8601)"
// @Param File formData file false "File to upload"
// @Success 201 {object} response.StandardResponse{data=models.CourseMaterial}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-materials [post]
// @Security BearerAuth
func (h *CourseMaterialHandler) CreateCourseMaterial(c *fiber.Ctx) error {
	// Parse form data
	courseID := c.FormValue("CourseID")
	title := c.FormValue("Title")
	description := c.FormValue("Description")
	materialType := c.FormValue("Type")
	weekStr := c.FormValue("Week")
	isPublicStr := c.FormValue("IsPublic")
	totalPointsStr := c.FormValue("TotalPoints")
	deadline := c.FormValue("Deadline")
	videoURL := c.FormValue("VideoURL")

	// Parse problem statement fields (for code exercises)
	problemStatement := c.FormValue("ProblemStatement")
	constraints := c.FormValue("Constraints")
	hints := c.FormValue("Hints")

	// Parse announcement content
	content := c.FormValue("Content")

	// Parse test cases for code exercises
	testCasesJSON := c.FormValue("TestCases")

	// Get uploaded file (optional)
	file, err := c.FormFile("File")
	if err != nil && err.Error() != "there is no uploaded file associated with the given key" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Failed to get uploaded file", err.Error())
	}

	// Validate required fields
	if courseID == "" || title == "" || materialType == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Missing required fields", "CourseID, Title, and Type are required")
	}

	// Parse optional fields
	week := 1
	if weekStr != "" {
		if w, err := strconv.Atoi(weekStr); err == nil {
			week = w
		}
	}

	isPublic := true
	if isPublicStr == "false" {
		isPublic = false
	}

	var totalPoints *int = nil
	if totalPointsStr != "" {
		if tp, err := strconv.Atoi(totalPointsStr); err == nil {
			totalPoints = &tp
		}
	} else if materialType == "code_exercise" || materialType == "pdf_exercise" {
		// Only set default for exercises
		defaultPoints := 100
		totalPoints = &defaultPoints
	}

	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id").(string)

	// Handle deadline if provided
	var deadlinePtr *string
	if deadline != "" {
		deadlinePtr = &deadline
	}

	// Create the specific material type based on materialType
	var materialID string

	switch materialType {
	case "code_exercise":
		// Create CodeExercise
		codeExercise := &models.CodeExercise{
			MaterialBase: models.MaterialBase{
				CourseID:    courseID,
				Title:       title,
				Description: description,
				Week:        week,
				IsPublic:    isPublic,
				CreatedBy:   userID,
			},
			TotalPoints:      totalPoints,
			Deadline:         deadlinePtr,
			ProblemStatement: problemStatement,
			Constraints:      constraints,
			Hints:            hints,
		}

		// Note: File upload for problem images will be handled after material creation

		// Parse test cases from JSON string
		testCases := []models.TestCase{}
		var exampleInputs []string
		var exampleOutputs []string
		if testCasesJSON != "" {
			// Parse JSON array of test cases
			var testCasesData []map[string]interface{}
			if err := json.Unmarshal([]byte(testCasesJSON), &testCasesData); err != nil {
				return response.ErrorResponse(c, http.StatusBadRequest, "Invalid test cases JSON format", err.Error())
			}

			// Convert to models.TestCase
			for _, tcData := range testCasesData {
				// Get input_data and expected_output
				inputDataRaw, ok := tcData["input_data"]
				if !ok {
					return response.ErrorResponse(c, http.StatusBadRequest, "Missing input_data in test case", nil)
				}
				expectedOutputRaw, ok := tcData["expected_output"]
				if !ok {
					return response.ErrorResponse(c, http.StatusBadRequest, "Missing expected_output in test case", nil)
				}

				// Convert input_data to JSON
				var inputDataBytes []byte
				var inputStr string
				if inputStrVal, ok := inputDataRaw.(string); ok {
					inputStr = inputStrVal
					// Try to parse as JSON if it looks like JSON
					if len(inputStr) > 0 && (inputStr[0] == '{' || inputStr[0] == '[') {
						// Try parsing as JSON
						if json.Unmarshal([]byte(inputStr), &map[string]interface{}{}) == nil {
							inputDataBytes = []byte(inputStr)
						} else {
							// Invalid JSON, keep as is (will be wrapped as string)
							inputDataBytes, _ = json.Marshal(inputStr)
							inputStr = string(inputDataBytes)
						}
					} else {
						// Plain text, keep as is (will be wrapped as string)
						inputDataBytes, _ = json.Marshal(inputStr)
					}
				} else {
					// Already an object, marshal it
					inputDataBytes, _ = json.Marshal(inputDataRaw)
					inputStr = string(inputDataBytes)
				}

				// Convert expected_output to JSON
				var expectedOutputBytes []byte
				var outputStr string
				if outputStrVal, ok := expectedOutputRaw.(string); ok {
					outputStr = outputStrVal
					// Try to parse as JSON if it looks like JSON
					if len(outputStr) > 0 && (outputStr[0] == '{' || outputStr[0] == '[') {
						// Try parsing as JSON
						if json.Unmarshal([]byte(outputStr), &map[string]interface{}{}) == nil {
							expectedOutputBytes = []byte(outputStr)
						} else {
							// Invalid JSON, wrap as {"output": "..."}
							outputObj := map[string]interface{}{"output": outputStr}
							expectedOutputBytes, _ = json.Marshal(outputObj)
							outputStr = string(expectedOutputBytes)
						}
					} else {
						// Plain text, convert to {"output": "..."}
						outputObj := map[string]interface{}{"output": outputStr}
						expectedOutputBytes, _ = json.Marshal(outputObj)
						outputStr = string(expectedOutputBytes)
					}
				} else {
					// Already an object, marshal it
					expectedOutputBytes, _ = json.Marshal(expectedOutputRaw)
					outputStr = string(expectedOutputBytes)
				}

				// Get display_name if present
				displayName := ""
				if dn, ok := tcData["display_name"].(string); ok && dn != "" {
					displayName = dn
				}

				// All test cases are public by default, so add to examples
				exampleInputs = append(exampleInputs, inputStr)
				exampleOutputs = append(exampleOutputs, outputStr)

				testCase := models.TestCase{
					InputData:      internaltypes.JSONData(inputDataBytes),
					ExpectedOutput: internaltypes.JSONData(expectedOutputBytes),
					IsPublic:       true, // All test cases are public by default
					DisplayName:    displayName,
				}
				testCases = append(testCases, testCase)
			}
		}

		// Set example inputs/outputs from test cases
		if len(exampleInputs) > 0 {
			exampleInputsJSON, _ := json.Marshal(exampleInputs)
			codeExercise.ExampleInputs = internaltypes.JSONData(exampleInputsJSON)
		}
		if len(exampleOutputs) > 0 {
			exampleOutputsJSON, _ := json.Marshal(exampleOutputs)
			codeExercise.ExampleOutputs = internaltypes.JSONData(exampleOutputsJSON)
		}

		// Create code exercise (this will also create the CourseMaterial reference)
		if err := h.materialService.CreateCodeExercise(codeExercise, testCases); err != nil {
			return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create code exercise", err.Error())
		}
		materialID = codeExercise.MaterialID

		// Handle file upload for problem images (optional) - after material creation
		if file != nil {
			// Validate file size
			const maxFileSize = 10 * 1024 * 1024 // 10MB
			if file.Size > maxFileSize {
				return response.ErrorResponse(c, http.StatusBadRequest, "File too large", "Maximum file size is 10MB")
			}

			// Validate file type (should be image)
			contentType := file.Header.Get("Content-Type")
			if !h.isAllowedImageType(contentType) {
				return response.ErrorResponse(c, http.StatusBadRequest, "Invalid file type", "Only JPEG, PNG, WebP, and GIF images are allowed")
			}

			// Open file
			src, err := file.Open()
			if err != nil {
				return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to open file", err.Error())
			}
			defer src.Close()

			// Upload image to storage
			imageURL, err := h.storageService.UploadExerciseImage(c.Context(), courseID, materialID, src, file.Filename, contentType)
			if err != nil {
				return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload image", err.Error())
			}

			// Update code exercise with problem image
			// Get existing problem images or create new array
			var existingExercise models.CodeExercise
			if err := h.materialService.GetDB().First(&existingExercise, "material_id = ?", materialID).Error; err == nil {
				var existingImages []string
				if existingExercise.ProblemImages != nil {
					json.Unmarshal(existingExercise.ProblemImages, &existingImages)
				}
				existingImages = append(existingImages, imageURL)
				problemImagesJSON, _ := json.Marshal(existingImages)
				updates := map[string]interface{}{
					"problem_images": internaltypes.JSONData(problemImagesJSON),
				}
				if err := h.materialService.UpdateCodeExercise(materialID, userID, updates); err != nil {
					return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update problem images", err.Error())
				}
			}
		}

	case "pdf_exercise":
		// For PDF exercises, handle file upload first if provided
		var fileURL, fileName string
		var fileSize int64
		var mimeType string

		if file != nil {
			// Validate file size
			const maxFileSize = 10 * 1024 * 1024 // 10MB
			if file.Size > maxFileSize {
				return response.ErrorResponse(c, http.StatusBadRequest, "File too large", "Maximum file size is 10MB")
			}

			// Open file
			src, err := file.Open()
			if err != nil {
				return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to open file", err.Error())
			}
			defer src.Close()

			// Generate temporary ID for upload path
			tempID := "temp_" + userID
			fileURL, err = h.storageService.UploadCoursePDFFile(c.Context(), courseID, tempID, src, file.Filename, file.Header.Get("Content-Type"))
			if err != nil {
				return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload file", err.Error())
			}
			fileName = file.Filename
			fileSize = file.Size
			mimeType = file.Header.Get("Content-Type")
		} else {
			return response.ErrorResponse(c, http.StatusBadRequest, "File is required for PDF exercise", "Please upload a PDF file")
		}

		// Create PDFExercise
		pdfExercise := &models.PDFExercise{
			MaterialBase: models.MaterialBase{
				CourseID:    courseID,
				Title:       title,
				Description: description,
				Week:        week,
				IsPublic:    isPublic,
				CreatedBy:   userID,
			},
			TotalPoints: totalPoints,
			Deadline:    deadlinePtr,
			FileURL:     fileURL,
			FileName:    fileName,
			FileSize:    fileSize,
			MimeType:    mimeType,
		}

		if err := h.materialService.CreatePDFExercise(pdfExercise); err != nil {
			return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create PDF exercise", err.Error())
		}
		materialID = pdfExercise.MaterialID

	case "document":
		// For documents, handle file upload first if provided
		var fileURL, fileName string
		var fileSize int64
		var mimeType string

		if file != nil {
			// Validate file size
			const maxFileSize = 10 * 1024 * 1024 // 10MB
			if file.Size > maxFileSize {
				return response.ErrorResponse(c, http.StatusBadRequest, "File too large", "Maximum file size is 10MB")
			}

			// Open file
			src, err := file.Open()
			if err != nil {
				return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to open file", err.Error())
			}
			defer src.Close()

			// Generate temporary ID for upload path
			tempID := "temp_" + userID
			fileURL, err = h.storageService.UploadCourseDocument(c.Context(), courseID, tempID, src, file.Filename, file.Header.Get("Content-Type"))
			if err != nil {
				return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload file", err.Error())
			}
			fileName = file.Filename
			fileSize = file.Size
			mimeType = file.Header.Get("Content-Type")
		} else {
			return response.ErrorResponse(c, http.StatusBadRequest, "File is required for document", "Please upload a file")
		}

		// Create Document
		document := &models.Document{
			MaterialBase: models.MaterialBase{
				CourseID:    courseID,
				Title:       title,
				Description: description,
				Week:        week,
				IsPublic:    isPublic,
				CreatedBy:   userID,
			},
			FileURL:  fileURL,
			FileName: fileName,
			FileSize: fileSize,
			MimeType: mimeType,
		}

		if err := h.materialService.CreateDocument(document); err != nil {
			return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create document", err.Error())
		}
		materialID = document.MaterialID

	case "video":
		// For videos, validate video URL
		if videoURL == "" {
			return response.ErrorResponse(c, http.StatusBadRequest, "Video URL is required", "Please provide a video URL")
		}

		// Create Video
		video := &models.Video{
			MaterialBase: models.MaterialBase{
				CourseID:    courseID,
				Title:       title,
				Description: description,
				Week:        week,
				IsPublic:    isPublic,
				CreatedBy:   userID,
			},
			VideoURL: videoURL,
		}

		if err := h.materialService.CreateVideo(video); err != nil {
			return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create video", err.Error())
		}
		materialID = video.MaterialID

	case "announcement":
		// Validate content is provided
		if content == "" {
			return response.ErrorResponse(c, http.StatusBadRequest, "Content is required", "Please provide announcement content")
		}

		// Create Announcement
		announcement := &models.Announcement{
			MaterialBase: models.MaterialBase{
				CourseID:    courseID,
				Title:       title,
				Description: description,
				Week:        week,
				IsPublic:    isPublic,
				CreatedBy:   userID,
			},
			Content: content,
		}

		if err := h.materialService.CreateAnnouncement(announcement); err != nil {
			return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create announcement", err.Error())
		}
		materialID = announcement.MaterialID

	default:
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid material type", "Material type must be one of: code_exercise, pdf_exercise, document, video, announcement")
	}

	// Get the created material to return
	material, err := h.materialService.GetCourseMaterialByID(materialID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve created material", err.Error())
	}

	// Note: pdf_exercise, document, and video already have their files/URLs set during creation above
	// For code_exercise, file uploads (optional problem images) would be handled separately if needed
	// The file URL for code exercises would need to be stored in a separate table or handled differently

	// Material is already a map[string]interface{} from service, no need to call ToJSON()
	return response.SuccessResponse(c, http.StatusCreated, "Course material created successfully", material)
}

// UploadCourseMaterialFile uploads a file for course material
// @Summary Upload course material file
// @Description Upload a file for course material (teachers only)
// @Tags course-materials
// @Accept multipart/form-data
// @Produce json
// @Param course_id formData string true "Course ID"
// @Param material_type formData string true "Material type" Enums(document,exercise)
// @Param file formData file true "File to upload"
// @Success 200 {object} response.StandardResponse{data=map[string]string}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-materials/upload [post]
// @Security BearerAuth
func (h *CourseMaterialHandler) UploadCourseMaterialFile(c *fiber.Ctx) error {
	// Get form data
	courseID := c.FormValue("course_id")
	materialType := c.FormValue("material_type")

	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	if materialType == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Material type is required", nil)
	}

	// Validate material type
	if !models.IsValidMaterialType(materialType) {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid material type", nil)
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "No file uploaded", err.Error())
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to open file", err.Error())
	}
	defer src.Close()

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	// Upload file
	fileURL, err := h.materialService.UploadCourseMaterialFile(c.Context(), courseID, userID, materialType, src, file.Filename, file.Header.Get("Content-Type"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload file", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "File uploaded successfully", map[string]string{
		"file_url":  fileURL,
		"file_name": file.Filename,
		"file_size": strconv.FormatInt(file.Size, 10),
	})
}

// GetCourseMaterials retrieves materials for a course
// @Summary Get course materials
// @Description Get course materials for a specific course with optional filtering
// @Tags course-materials
// @Produce json
// @Param course_id query string true "Course ID"
// @Param week query int false "Filter by week"
// @Param type query string false "Filter by material type"
// @Param limit query int false "Limit results" default(20)
// @Param offset query int false "Offset results" default(0)
// @Success 200 {object} response.StandardResponse{data=[]models.CourseMaterial}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-materials [get]
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *CourseMaterialHandler) GetCourseMaterials(c *fiber.Ctx) error {
	courseID := c.Query("course_id")
	if courseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Course ID is required", nil)
	}

	// Validate course access for non-teachers
	if err := h.enrollmentValidator.ValidateCourseAccess(c, courseID); err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
	}

	// Parse optional parameters
	limit := 20
	offset := 0
	var week *int
	var materialType *string

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if weekStr := c.Query("week"); weekStr != "" {
		if parsed, err := strconv.Atoi(weekStr); err == nil && parsed >= 0 {
			week = &parsed
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		if models.IsValidMaterialType(typeStr) {
			materialType = &typeStr
		}
	}

	materials, total, err := h.materialService.GetCourseMaterialsByCourse(courseID, week, materialType, limit, offset)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get course materials", err.Error())
	}

	// Materials already contain full details from service (no need to call ToJSON())
	return response.SuccessResponse(c, http.StatusOK, "Course materials retrieved successfully", map[string]interface{}{
		"materials": materials,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetCourseMaterial retrieves a specific course material
// @Summary Get course material
// @Description Get a specific course material by ID
// @Tags course-materials
// @Produce json
// @Param id path string true "Material ID"
// @Success 200 {object} response.StandardResponse{data=models.CourseMaterial}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-materials/{id} [get]
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *CourseMaterialHandler) GetCourseMaterial(c *fiber.Ctx) error {
	materialID := c.Params("id")
	if materialID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Material ID is required", nil)
	}

	material, err := h.materialService.GetCourseMaterialByID(materialID)
	if err != nil {
		if err.Error() == "course material not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Course material not found", nil)
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get course material", err.Error())
	}

	// Material already contains full details from service (no need to call ToJSON())
	return response.SuccessResponse(c, http.StatusOK, "Course material retrieved successfully", material)
}

// UpdateCourseMaterial updates a course material
// @Summary Update course material
// @Description Update an existing course material (creator only)
// @Tags course-materials
// @Accept json
// @Produce json
// @Param id path string true "Material ID"
// @Param request body types.UpdateCourseMaterialRequest true "Update data"
// @Success 200 {object} response.StandardResponse{data=models.CourseMaterial}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-materials/{id} [put]
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *CourseMaterialHandler) UpdateCourseMaterial(c *fiber.Ctx) error {
	materialID := c.Params("id")
	if materialID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Material ID is required", nil)
	}

	// Parse body manually to handle test_cases array correctly
	bodyBytes := c.Body()
	if len(bodyBytes) == 0 {
		return response.ErrorResponse(c, http.StatusBadRequest, "Request body is empty", nil)
	}

	var rawBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawBody); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON body", err.Error())
	}

	// Extract test_cases before parsing to struct
	var testCasesRaw interface{}
	var hasTestCases bool
	if testCasesRaw, hasTestCases = rawBody["test_cases"]; hasTestCases {
		// Remove test_cases from rawBody temporarily for BodyParser
		delete(rawBody, "test_cases")
		// Re-marshal without test_cases
		bodyBytes, _ = json.Marshal(rawBody)
	}

	var req types.UpdateCourseMaterialRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Manually parse test_cases if present
	if hasTestCases && req.TestCases == nil {
		if testCasesArray, ok := testCasesRaw.([]interface{}); ok {
			testCasesMaps := make([]map[string]interface{}, 0, len(testCasesArray))
			for _, tc := range testCasesArray {
				if tcMap, ok := tc.(map[string]interface{}); ok {
					testCasesMaps = append(testCasesMaps, tcMap)
				}
			}
			req.TestCases = &testCasesMaps
		}
	}

	// Validate request (skip validation for test_cases as it's handled manually)
	reqForValidation := req
	reqForValidation.TestCases = nil
	if err := config.Validate.Struct(reqForValidation); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	// Prepare updates
	updates := make(map[string]interface{})
	if req.Title != nil && *req.Title != "" {
		updates["title"] = *req.Title
	}
	if req.Description != nil && *req.Description != "" {
		updates["description"] = *req.Description
	}
	if req.Type != nil && *req.Type != "" {
		updates["type"] = *req.Type
	}
	if req.Week != nil {
		updates["week"] = *req.Week
	}
	if req.VideoURL != nil && *req.VideoURL != "" {
		updates["video_url"] = *req.VideoURL
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	// Code exercise fields
	if req.TotalPoints != nil {
		updates["total_points"] = *req.TotalPoints
	}
	if req.Deadline != nil {
		updates["deadline"] = *req.Deadline
	}
	if req.ProblemStatement != nil {
		updates["problem_statement"] = *req.ProblemStatement
	}
	if req.Constraints != nil {
		updates["constraints"] = *req.Constraints
	}
	if req.Hints != nil {
		updates["hints"] = *req.Hints
	}
	// Note: test_cases are handled separately below

	// Handle test cases update for code exercises (before updating material)
	// Update test cases if provided (even if empty array to clear all test cases)
	if req.TestCases != nil {
		// Get material to check if it's a code exercise
		material, err := h.materialService.GetCourseMaterialByID(materialID)
		if err == nil {
			if materialType, ok := material["type"].(string); ok && materialType == "code_exercise" {
				// Delete all existing test cases
				existingTestCases, _ := h.materialService.GetTestCases(materialID)
				for _, tc := range existingTestCases {
					h.materialService.DeleteTestCase(tc.TestCaseID, userID)
				}

				// Create new test cases
				var exampleInputs []string
				var exampleOutputs []string
				testCases := []models.TestCase{}
				for _, tcData := range *req.TestCases {
					// Get input_data and expected_output
					inputDataRaw, ok := tcData["input_data"]
					if !ok {
						continue
					}
					expectedOutputRaw, ok := tcData["expected_output"]
					if !ok {
						continue
					}

					// Convert input_data to JSON
					var inputDataBytes []byte
					var inputStr string
					if inputStrVal, ok := inputDataRaw.(string); ok {
						inputStr = inputStrVal
						// Try to parse as JSON first (handles both JSON strings and plain text)
						var parsedInput interface{}
						if err := json.Unmarshal([]byte(inputStr), &parsedInput); err == nil {
							// Valid JSON, use it as is
							inputDataBytes, _ = json.Marshal(parsedInput)
							inputStr = string(inputDataBytes)
						} else {
							// Not valid JSON, treat as plain text
							inputDataBytes, _ = json.Marshal(inputStr)
							inputStr = string(inputDataBytes)
						}
					} else {
						// Already an object, marshal it
						inputDataBytes, _ = json.Marshal(inputDataRaw)
						inputStr = string(inputDataBytes)
					}

					// Convert expected_output to JSON
					var expectedOutputBytes []byte
					var outputStr string
					if outputStrVal, ok := expectedOutputRaw.(string); ok {
						outputStr = outputStrVal
						// Try to parse as JSON first
						var parsedOutput interface{}
						if err := json.Unmarshal([]byte(outputStr), &parsedOutput); err == nil {
							// Valid JSON, check if it's an object with "output" key
							if outputMap, ok := parsedOutput.(map[string]interface{}); ok {
								if _, hasOutput := outputMap["output"]; hasOutput {
									// Already has "output" key, use as is
									expectedOutputBytes, _ = json.Marshal(parsedOutput)
									outputStr = string(expectedOutputBytes)
								} else {
									// Valid JSON object but no "output" key, wrap it
									outputObj := map[string]interface{}{"output": parsedOutput}
									expectedOutputBytes, _ = json.Marshal(outputObj)
									outputStr = string(expectedOutputBytes)
								}
							} else {
								// Valid JSON but not an object (array, number, etc.), wrap it
								outputObj := map[string]interface{}{"output": parsedOutput}
								expectedOutputBytes, _ = json.Marshal(outputObj)
								outputStr = string(expectedOutputBytes)
							}
						} else {
							// Not valid JSON, treat as plain text and wrap
							outputObj := map[string]interface{}{"output": outputStr}
							expectedOutputBytes, _ = json.Marshal(outputObj)
							outputStr = string(expectedOutputBytes)
						}
					} else {
						// Already an object, check if it has "output" key
						if outputMap, ok := expectedOutputRaw.(map[string]interface{}); ok {
							if _, hasOutput := outputMap["output"]; hasOutput {
								// Already has "output" key, use as is
								expectedOutputBytes, _ = json.Marshal(expectedOutputRaw)
								outputStr = string(expectedOutputBytes)
							} else {
								// No "output" key, wrap it
								outputObj := map[string]interface{}{"output": expectedOutputRaw}
								expectedOutputBytes, _ = json.Marshal(outputObj)
								outputStr = string(expectedOutputBytes)
							}
						} else {
							// Not a map, wrap it
							outputObj := map[string]interface{}{"output": expectedOutputRaw}
							expectedOutputBytes, _ = json.Marshal(outputObj)
							outputStr = string(expectedOutputBytes)
						}
					}

					// Get display_name if present
					displayName := ""
					if dn, ok := tcData["display_name"].(string); ok && dn != "" {
						displayName = dn
					}

					exampleInputs = append(exampleInputs, inputStr)
					exampleOutputs = append(exampleOutputs, outputStr)

					testCase := models.TestCase{
						MaterialID:     &materialID,
						MaterialType:   "code_exercise",
						InputData:      internaltypes.JSONData(inputDataBytes),
						ExpectedOutput: internaltypes.JSONData(expectedOutputBytes),
						DisplayName:    displayName,
						IsPublic:       true,
					}
					testCases = append(testCases, testCase)
				}

				// Create new test cases
				for i := range testCases {
					if err := h.materialService.AddTestCase(materialID, &testCases[i]); err != nil {
						return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create test case", err.Error())
					}
				}

				// Update example inputs/outputs
				if len(exampleInputs) > 0 {
					exampleInputsJSON, _ := json.Marshal(exampleInputs)
					updates["example_inputs"] = internaltypes.JSONData(exampleInputsJSON)
				}
				if len(exampleOutputs) > 0 {
					exampleOutputsJSON, _ := json.Marshal(exampleOutputs)
					updates["example_outputs"] = internaltypes.JSONData(exampleOutputsJSON)
				}
			}
		}
	}

	if err := h.materialService.UpdateCourseMaterial(materialID, userID, updates); err != nil {
		if err.Error() == "course material not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Course material not found", nil)
		}
		if err.Error() == "only the creator can update this course material" {
			return response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update course material", err.Error())
	}

	// Get updated material
	material, err := h.materialService.GetCourseMaterialByID(materialID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get updated course material", err.Error())
	}

	// Material is already a map[string]interface{} from service, no need to call ToJSON()
	return response.SuccessResponse(c, http.StatusOK, "Course material updated successfully", material)
}

// DeleteCourseMaterial deletes a course material
// @Summary Delete course material
// @Description Delete a course material (creator only)
// @Tags course-materials
// @Produce json
// @Param id path string true "Material ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-materials/{id} [delete]
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *CourseMaterialHandler) DeleteCourseMaterial(c *fiber.Ctx) error {
	materialID := c.Params("id")
	if materialID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Material ID is required", nil)
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	if err := h.materialService.DeleteCourseMaterial(materialID, userID); err != nil {
		if err.Error() == "course material not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Course material not found", nil)
		}
		if err.Error() == "only the creator can delete this course material" {
			return response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete course material", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Course material deleted successfully", nil)
}

// Deprecated: GetCourseMaterialStats removed

// Test Case Management Methods

// GetTestCases retrieves test cases for a material
// @Summary Get test cases
// @Description Get test cases for a specific material
// @Tags course-materials
// @Produce json
// @Param id path string true "Material ID"
// @Success 200 {object} response.StandardResponse{data=[]models.TestCase}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-materials/{id}/test-cases [get]
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *CourseMaterialHandler) GetTestCases(c *fiber.Ctx) error {
	materialID := c.Params("id")
	if materialID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Material ID is required", nil)
	}

	testCases, err := h.materialService.GetTestCases(materialID)
	if err != nil {
		if err.Error() == "course material not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Course material not found", nil)
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get test cases", err.Error())
	}

	// Convert to JSON
	testCasesJSON := make([]map[string]interface{}, len(testCases))
	for i, testCase := range testCases {
		testCasesJSON[i] = testCase.ToJSON()
	}

	return response.SuccessResponse(c, http.StatusOK, "Test cases retrieved successfully", testCasesJSON)
}

// AddTestCase adds a test case to a material
// @Summary Add test case
// @Description Add a test case to a material (teachers only)
// @Tags course-materials
// @Accept json
// @Produce json
// @Param id path string true "Material ID"
// @Param request body object{input_data=object,expected_output=object} true "Test case data"
// @Success 201 {object} response.StandardResponse{data=models.TestCase}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-materials/{id}/test-cases [post]
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *CourseMaterialHandler) AddTestCase(c *fiber.Ctx) error {
	materialID := c.Params("id")
	if materialID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Material ID is required", nil)
	}

	var req struct {
		InputData      map[string]interface{} `json:"input_data" validate:"required"`
		ExpectedOutput map[string]interface{} `json:"expected_output" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validate request
	if err := config.Validate.Struct(req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
	}

	// Convert maps to JSON bytes
	inputDataBytes, _ := json.Marshal(req.InputData)
	expectedOutputBytes, _ := json.Marshal(req.ExpectedOutput)

	// Create test case
	testCase := &models.TestCase{
		InputData:      internaltypes.JSONData(inputDataBytes),
		ExpectedOutput: internaltypes.JSONData(expectedOutputBytes),
	}

	if err := h.materialService.AddTestCase(materialID, testCase); err != nil {
		if err.Error() == "course material not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Course material not found", nil)
		}
		if err.Error() == "can only add test cases to code exercises" {
			return response.ErrorResponse(c, http.StatusBadRequest, "Can only add test cases to code exercises", nil)
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to add test case", err.Error())
	}

	return response.SuccessResponse(c, http.StatusCreated, "Test case added successfully", testCase.ToJSON())
}

// UpdateTestCase updates a test case
// @Summary Update test case
// @Description Update a test case (teachers only)
// @Tags course-materials
// @Accept json
// @Produce json
// @Param test_case_id path string true "Test Case ID"
// @Param request body object{input_data=object,expected_output=object} true "Test case data"
// @Success 200 {object} response.StandardResponse{data=models.TestCase}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-materials/test-cases/{test_case_id} [put]
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *CourseMaterialHandler) UpdateTestCase(c *fiber.Ctx) error {
	testCaseID := c.Params("test_case_id")
	if testCaseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Test Case ID is required", nil)
	}

	var req struct {
		InputData      map[string]interface{} `json:"input_data,omitempty"`
		ExpectedOutput map[string]interface{} `json:"expected_output,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	// Prepare updates
	updates := make(map[string]interface{})
	if req.InputData != nil {
		updates["input_data"] = req.InputData
	}
	if req.ExpectedOutput != nil {
		updates["expected_output"] = req.ExpectedOutput
	}

	if err := h.materialService.UpdateTestCase(testCaseID, userID, updates); err != nil {
		if err.Error() == "test case not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Test case not found", nil)
		}
		if err.Error() == "only the creator can update test cases" {
			return response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update test case", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Test case updated successfully", nil)
}

// DeleteTestCase deletes a test case
// @Summary Delete test case
// @Description Delete a test case (teachers only)
// @Tags course-materials
// @Produce json
// @Param test_case_id path string true "Test Case ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-materials/test-cases/{test_case_id} [delete]
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *CourseMaterialHandler) DeleteTestCase(c *fiber.Ctx) error {
	testCaseID := c.Params("test_case_id")
	if testCaseID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Test Case ID is required", nil)
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)

	if err := h.materialService.DeleteTestCase(testCaseID, userID); err != nil {
		if err.Error() == "test case not found" {
			return response.ErrorResponse(c, http.StatusNotFound, "Test case not found", nil)
		}
		if err.Error() == "only the creator can delete test cases" {
			return response.ErrorResponse(c, http.StatusForbidden, "Access denied", err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete test case", err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, "Test case deleted successfully", nil)
}

// UploadProblemImage godoc
// @Summary Upload problem image
// @Description Upload an image for exercise problem statement
// @Tags course-materials
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Material ID"
// @Param image formData file true "Image file"
// @Success 200 {object} response.StandardResponse{data=object{image_url=string}}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/course-materials/{id}/images [post]
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *CourseMaterialHandler) UploadProblemImage(c *fiber.Ctx) error {
	materialID := c.Params("id")
	if materialID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Material ID is required", "")
	}

	// Get uploaded image file
	file, err := c.FormFile("image")
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Failed to get uploaded image", err.Error())
	}

	// Validate file size
	const maxFileSize = 5 * 1024 * 1024 // 5MB for images
	if file.Size > maxFileSize {
		return response.ErrorResponse(c, http.StatusBadRequest, "Image too large", "Maximum image size is 5MB")
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if !h.isAllowedImageType(contentType) {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid image type", "Only JPEG, PNG, WebP, and GIF images are allowed")
	}

	// Get material to verify it exists and get course ID
	material, err := h.materialService.GetCourseMaterialByID(materialID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusNotFound, "Material not found", err.Error())
	}

	// Get course_id from map
	courseIDStr, ok := material["course_id"].(string)
	if !ok {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Invalid material data", "Course ID not found")
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to open image file", err.Error())
	}
	defer src.Close()

	// Upload image to storage
	imageURL, err := h.storageService.UploadExerciseImage(c.Context(), courseIDStr, materialID, src, file.Filename, contentType)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload image", err.Error())
	}

	return response.SendSuccess(c, "Image uploaded successfully", fiber.Map{
		"image_url": imageURL,
	})
}

// Helper method to check if content type is allowed for images
func (h *CourseMaterialHandler) isAllowedImageType(contentType string) bool {
	allowedTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
		"image/gif",
	}

	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}
