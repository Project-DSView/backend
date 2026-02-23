package handler

import (
	"encoding/json"

	"github.com/Project-DSView/backend/go/internal/application/services"
	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/errors"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type TestCaseHandler struct {
	testCaseService *services.TestCaseService
	userService     *services.UserService
}

func NewTestCaseHandler(testCaseService *services.TestCaseService, userService *services.UserService) *TestCaseHandler {
	return &TestCaseHandler{
		testCaseService: testCaseService,
		userService:     userService,
	}
}

// GetTestCases godoc
// @Summary Get test cases by exercise ID
// @Description Get all test cases for a specific exercise
// @Tags test-cases
// @Security BearerAuth
// @Produce json
// @Param exercise_id path string true "Exercise ID"
// @Success 200 {object} object{success=bool,data=[]object} "List of test cases"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/exercises/{exercise_id}/test-cases [get]
func (h *TestCaseHandler) GetTestCases(c *fiber.Ctx) error {
	// Get current user
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorizedError(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil {
		return response.SendGenericError(c, err)
	}
	if currentUser == nil {
		return response.SendNotFoundError(c, "User")
	}

	exerciseID := c.Params("exercise_id")
	if exerciseID == "" {
		return response.SendValidationErrorWithField(c, "Exercise ID is required", "exercise_id")
	}

	// Check permissions - for now, allow all authenticated users to view test cases
	// TODO: Implement proper material-based permission checking
	testCases, err := h.testCaseService.GetTestCasesByMaterialID(exerciseID)
	if err != nil {
		return response.SendGenericError(c, errors.Wrap(err, "Failed to fetch test cases"))
	}

	testCaseData := make([]map[string]interface{}, len(testCases))
	for i, testCase := range testCases {
		testCaseData[i] = testCase.ToJSON()
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    testCaseData,
	})
}

// CreateTestCase godoc
// @Summary Create a new test case
// @Description Create a new test case for an exercise
// @Tags test-cases
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param exercise_id path string true "Exercise ID"
// @Param testcase body object{input_data=object,expected_output=object} true "Test case data"
// @Success 201 {object} object{success=bool,message=string,data=object} "Test case created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/exercises/{exercise_id}/test-cases [post]
func (h *TestCaseHandler) CreateTestCase(c *fiber.Ctx) error {
	// Get current user
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	exerciseID := c.Params("exercise_id")
	if exerciseID == "" {
		return response.SendBadRequest(c, "Exercise ID is required")
	}

	// Check permissions - for now, allow all authenticated users to create test cases
	// TODO: Implement proper material-based permission checking

	var req struct {
		InputData      map[string]interface{} `json:"input_data"`
		ExpectedOutput map[string]interface{} `json:"expected_output"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}

	// Convert maps to JSON bytes
	inputDataBytes, _ := json.Marshal(req.InputData)
	expectedOutputBytes, _ := json.Marshal(req.ExpectedOutput)

	testCase := models.TestCase{
		MaterialID:     &exerciseID, // Using exerciseID as materialID for now
		InputData:      types.JSONData(inputDataBytes),
		ExpectedOutput: types.JSONData(expectedOutputBytes),
	}

	if err := h.testCaseService.CreateTestCase(&testCase); err != nil {
		return response.SendInternalError(c, "Failed to create test case: "+err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Test case created successfully",
		"data":    testCase.ToJSON(),
	})
}

// UpdateTestCase godoc
// @Summary Update test case
// @Description Update test case information
// @Tags test-cases
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Test Case ID"
// @Param testcase body object{input_data=object,expected_output=object} true "Test case update data"
// @Success 200 {object} object{success=bool,message=string,data=object} "Updated test case"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Test case not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/test-cases/{id} [put]
func (h *TestCaseHandler) UpdateTestCase(c *fiber.Ctx) error {
	// Get current user
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	testCaseID := c.Params("id")
	if testCaseID == "" {
		return response.SendBadRequest(c, "Test case ID is required")
	}

	// Parse request body
	var req struct {
		InputData      map[string]interface{} `json:"input_data,omitempty"`
		ExpectedOutput map[string]interface{} `json:"expected_output,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.SendBadRequest(c, "Invalid request body: "+err.Error())
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.InputData != nil {
		updates["input_data"] = req.InputData
	}
	if req.ExpectedOutput != nil {
		updates["expected_output"] = req.ExpectedOutput
	}

	if len(updates) == 0 {
		return response.SendBadRequest(c, "No valid updates provided")
	}

	// Update test case
	if err := h.testCaseService.UpdateTestCase(testCaseID, updates); err != nil {
		return response.SendInternalError(c, "Failed to update test case: "+err.Error())
	}

	return response.SendSuccess(c, "Test case updated successfully", nil)
}

// DeleteTestCase godoc
// @Summary Delete test case
// @Description Delete a test case
// @Tags test-cases
// @Security BearerAuth
// @Produce json
// @Param id path string true "Test Case ID"
// @Success 200 {object} object{success=bool,message=string,data=interface{}} "Test case deleted successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Test case not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/test-cases/{id} [delete]
func (h *TestCaseHandler) DeleteTestCase(c *fiber.Ctx) error {
	// Get current user
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Invalid authentication")
	}

	currentUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || currentUser == nil {
		return response.SendUnauthorized(c, "User not found")
	}

	testCaseID := c.Params("id")
	if testCaseID == "" {
		return response.SendBadRequest(c, "Test case ID is required")
	}

	// Delete test case
	if err := h.testCaseService.DeleteTestCase(testCaseID); err != nil {
		return response.SendInternalError(c, "Failed to delete test case: "+err.Error())
	}

	return response.SendSuccess(c, "Test case deleted successfully", nil)
}
