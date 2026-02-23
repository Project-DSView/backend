package handler

import (
	"io"

	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"github.com/Project-DSView/backend/go/pkg/validation"
	"github.com/gofiber/fiber/v2"
)

type DraftHandler struct {
	draftService   *services.DraftService
	userService    *services.UserService
	storageService storage.StorageService
}

func NewDraftHandler(
	draftService *services.DraftService,
	userService *services.UserService,
	storageService storage.StorageService,
) *DraftHandler {
	return &DraftHandler{
		draftService:   draftService,
		userService:    userService,
		storageService: storageService,
	}
}

// UploadPythonFile godoc
// @Summary Upload Python file
// @Description Upload a Python (.py) file for an exercise
// @Tags drafts
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param exercise_id path string true "Exercise ID"
// @Param file formData file true "Python file (.py)"
// @Success 200 {object} object{success=bool,message=string,data=object{draft_id=string,file_name=string,file_size=int,code_preview=string}} "File uploaded successfully"
// @Failure 400 {object} object{success=bool,error=string} "Bad request - invalid file"
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 413 {object} object{success=bool,error=string} "File too large"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/exercises/{exercise_id}/upload [post]
func (h *DraftHandler) UploadPythonFile(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	exerciseID := c.Params("exercise_id")
	if exerciseID == "" {
		return response.SendBadRequest(c, "Exercise ID is required")
	}

	// TODO: Implement material validation
	// For now, skip exercise validation since exercises are deprecated

	// รับไฟล์จาก form
	file, err := c.FormFile("file")
	if err != nil {
		return response.SendBadRequest(c, "No file uploaded")
	}

	// เปิดไฟล์และอ่านเนื้อหา
	fileContent, err := file.Open()
	if err != nil {
		return response.SendInternalError(c, "Failed to open uploaded file")
	}
	defer fileContent.Close()

	content, err := io.ReadAll(fileContent)
	if err != nil {
		return response.SendInternalError(c, "Failed to read file content")
	}

	// ตรวจสอบไฟล์
	if err := validation.ValidatePythonFile(file.Filename, content); err != nil {
		return response.SendValidationError(c, err.Error())
	}

	// ทำความสะอาดชื่อไฟล์
	sanitizedFilename := validation.SanitizeFilename(file.Filename)

	// Upload to object storage (MinIO)
	filePath, err := h.storageService.UploadCodeFile(
		c.Context(),
		claims.UserID,
		exerciseID,
		fileContent,
		file.Filename,
		file.Header.Get("Content-Type"),
	)
	if err != nil {
		return response.SendInternalError(c, "Failed to save file: "+err.Error())
	}

	// บันทึก draft (store metadata; file path is stored separately if needed)
	draft, err := h.draftService.SaveDraft(
		claims.UserID,
		exerciseID, // Using exerciseID as materialID for now
		string(content),
		sanitizedFilename,
		file.Size,
	)
	if err != nil {
		return response.SendInternalError(c, "Failed to save draft: "+err.Error())
	}

	// สร้าง preview ของโค้ด (100 ตัวอักษรแรก)
	codePreview := string(content)
	if len(codePreview) > 100 {
		codePreview = codePreview[:100] + "..."
	}

	return response.SendSuccess(c, "File uploaded successfully", fiber.Map{
		"draft_id":     draft.DraftID,
		"file_name":    draft.FileName,
		"file_size":    draft.FileSize,
		"code_preview": codePreview,
		"file_path":    filePath,
		"material_id":  exerciseID, // Using exerciseID as materialID for now
	})
}

// SaveDraft godoc
// @Summary Save code draft
// @Description Save code as draft without submitting
// @Tags drafts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param exercise_id path string true "Exercise ID"
// @Param draft body object{code=string,file_name=string} true "Draft data"
// @Success 200 {object} object{success=bool,message=string,data=object{draft_id=string,saved_at=string}} "Draft saved successfully"
// @Failure 400 {object} object{success=bool,error=string} "Bad request"
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/exercises/{exercise_id}/draft [post]
func (h *DraftHandler) SaveDraft(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	exerciseID := c.Params("exercise_id")
	if exerciseID == "" {
		return response.SendBadRequest(c, "Exercise ID is required")
	}

	var req struct {
		Code     string `json:"code"`
		FileName string `json:"file_name,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}

	// ตรวจสอบเนื้อหาโค้ด
	if err := validation.ValidateCodeContent(req.Code); err != nil {
		return response.SendValidationError(c, err.Error())
	}

	// ตั้งชื่อไฟล์ default ถ้าไม่มี
	if req.FileName == "" {
		req.FileName = "solution.py"
	}

	// บันทึก draft
	draft, err := h.draftService.SaveDraft(
		claims.UserID,
		exerciseID,
		req.Code,
		req.FileName,
		int64(len(req.Code)),
	)
	if err != nil {
		return response.SendInternalError(c, "Failed to save draft: "+err.Error())
	}

	return response.SendSuccess(c, "Draft saved successfully", fiber.Map{
		"draft_id": draft.DraftID,
		"saved_at": draft.UpdatedAt,
	})
}

// GetDraft godoc
// @Summary Get code draft
// @Description Retrieve saved code draft for an exercise
// @Tags drafts
// @Security BearerAuth
// @Produce json
// @Param exercise_id path string true "Exercise ID"
// @Success 200 {object} object{success=bool,data=object{draft_id=string,code=string,file_name=string,file_size=int,updated_at=string}} "Draft retrieved successfully"
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 404 {object} object{success=bool,error=string} "Draft not found"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/exercises/{exercise_id}/draft [get]
func (h *DraftHandler) GetDraft(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	exerciseID := c.Params("exercise_id")
	if exerciseID == "" {
		return response.SendBadRequest(c, "Exercise ID is required")
	}

	draft, err := h.draftService.GetDraft(claims.UserID, exerciseID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get draft: "+err.Error())
	}

	if draft == nil {
		return response.SendNotFound(c, "No draft found for this exercise")
	}

	return response.SendSuccess(c, "Draft retrieved successfully", fiber.Map{
		"draft_id":   draft.DraftID,
		"code":       draft.Code,
		"file_name":  draft.FileName,
		"file_size":  draft.FileSize,
		"updated_at": draft.UpdatedAt,
	})
}

// GetMyDrafts godoc
// @Summary Get all user drafts
// @Description Get all code drafts created by current user
// @Tags drafts
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{success=bool,data=[]object{draft_id=string,exercise_id=string,exercise_title=string,file_name=string,file_size=int,updated_at=string}} "User drafts retrieved successfully"
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/my-drafts [get]
func (h *DraftHandler) GetMyDrafts(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	drafts, err := h.draftService.GetUserDrafts(claims.UserID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get drafts: "+err.Error())
	}

	// เติมข้อมูล material title
	draftData := make([]map[string]interface{}, len(drafts))
	for i, draft := range drafts {
		materialTitle := ""
		// TODO: Implement material title lookup
		// For now, use material_id as title

		draftData[i] = map[string]interface{}{
			"draft_id":       draft.DraftID,
			"material_id":    draft.MaterialID,
			"material_title": materialTitle,
			"file_name":      draft.FileName,
			"file_size":      draft.FileSize,
			"updated_at":     draft.UpdatedAt,
		}
	}

	return response.SendSuccess(c, "User drafts retrieved successfully", draftData)
}

// DeleteDraft godoc
// @Summary Delete code draft
// @Description Delete saved code draft for an exercise
// @Tags drafts
// @Security BearerAuth
// @Produce json
// @Param exercise_id path string true "Exercise ID"
// @Success 200 {object} object{success=bool,message=string} "Draft deleted successfully"
// @Failure 401 {object} object{success=bool,error=string} "Unauthorized"
// @Failure 500 {object} object{success=bool,error=string} "Internal server error"
// @Router /api/exercises/{exercise_id}/draft [delete]
func (h *DraftHandler) DeleteDraft(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	exerciseID := c.Params("exercise_id")
	if exerciseID == "" {
		return response.SendBadRequest(c, "Exercise ID is required")
	}

	if err := h.draftService.DeleteDraft(claims.UserID, exerciseID); err != nil {
		return response.SendInternalError(c, "Failed to delete draft: "+err.Error())
	}

	return response.SendSuccess(c, "Draft deleted successfully", nil)
}
