package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/gofiber/fiber/v2"
)

// PlaygroundHandler handles playground code execution
type PlaygroundHandler struct {
	cfg *config.Config
}

// NewPlaygroundHandler creates a new playground handler
func NewPlaygroundHandler(cfg *config.Config) *PlaygroundHandler {
	return &PlaygroundHandler{cfg: cfg}
}

// ExecutionRequest represents the request body for code execution
type ExecutionRequest struct {
	Code     string `json:"code"`
	DataType string `json:"dataType"`
}

// ComplexityRequest represents the request body for complexity analysis
type ComplexityRequest struct {
	Code  string `json:"code"`
	Model string `json:"model,omitempty"`
}

// FastAPIExecutionResponse represents the response from FastAPI
type FastAPIExecutionResponse struct {
	ExecutionID  string      `json:"executionId"`
	Code         string      `json:"code"`
	DataType     string      `json:"dataType"`
	Steps        interface{} `json:"steps"`
	TotalSteps   int         `json:"totalSteps"`
	Status       string      `json:"status"`
	ErrorMessage *string     `json:"errorMessage,omitempty"`
	ExecutedAt   string      `json:"executedAt"`
	CreatedAt    string      `json:"createdAt"`
	Complexity   interface{} `json:"complexity,omitempty"`
}

// ExecutionResponse represents the response sent to client
type ExecutionResponse struct {
	Success     bool        `json:"success"`
	ExecutionID string      `json:"executionId"`
	Steps       interface{} `json:"steps"`
	Error       string      `json:"error,omitempty"`
	Complexity  interface{} `json:"complexity,omitempty"`
}

// RunCodeGateway handles code execution requests
func (h *PlaygroundHandler) RunCodeGateway(c *fiber.Ctx) error {
	var req ExecutionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate request
	if req.Code == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Code cannot be empty",
		})
	}

	if req.DataType == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "DataType is required",
		})
	}

	// Forward request to FastAPI playground service
	fastAPIURL := h.cfg.Fastapi.BaseURL + "/api/playground/run"

	// Create request to FastAPI
	client := &http.Client{Timeout: 30 * time.Second}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to marshal request",
		})
	}

	httpReq, err := http.NewRequest("POST", fastAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create request",
		})
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(h.cfg.APIKey.APIKeyName, h.cfg.APIKey.APIKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to execute code",
		})
	}
	defer resp.Body.Close()

	// Decode FastAPI response
	var fastAPIResp FastAPIExecutionResponse
	if err := json.NewDecoder(resp.Body).Decode(&fastAPIResp); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to decode response",
		})
	}

	// Map FastAPI response to client response format
	clientResp := ExecutionResponse{
		Success:     fastAPIResp.Status == "success" || fastAPIResp.Status == "waiting",
		ExecutionID: fastAPIResp.ExecutionID,
		Steps:       fastAPIResp.Steps,
		Complexity:  fastAPIResp.Complexity,
	}

	// If there's an error message, include it
	if fastAPIResp.ErrorMessage != nil && *fastAPIResp.ErrorMessage != "" {
		clientResp.Error = *fastAPIResp.ErrorMessage
	} else if fastAPIResp.Status == "error" {
		// If status is error but no errorMessage, provide default message
		clientResp.Error = "Code execution failed"
	}

	return c.JSON(clientResp)
}

// HealthCheck returns playground service health status
func (h *PlaygroundHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Playground service is healthy",
		"service": "go-playground-gateway",
	})
}

// ComplexityLLM handles LLM complexity analysis requests
func (h *PlaygroundHandler) ComplexityLLM(c *fiber.Ctx) error {
	var req ComplexityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate request
	if req.Code == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Code cannot be empty",
		})
	}

	// Forward request to FastAPI complexity LLM service
	fastAPIURL := h.cfg.Fastapi.BaseURL + "/api/complexity/llm"

	// Create request to FastAPI
	client := &http.Client{Timeout: 60 * time.Second} // Longer timeout for LLM

	reqBody, err := json.Marshal(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to marshal request",
		})
	}

	httpReq, err := http.NewRequest("POST", fastAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create request",
		})
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(h.cfg.APIKey.APIKeyName, h.cfg.APIKey.APIKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to analyze complexity",
		})
	}
	defer resp.Body.Close()

	// Forward FastAPI response directly
	// We read the body to ensure we can forward it exactly as is, or decode generic map
	var fastAPIResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&fastAPIResp); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to decode response from analysis service",
		})
	}

	return c.JSON(fastAPIResp)
}

// ComplexityPerformance handles performance complexity analysis requests
func (h *PlaygroundHandler) ComplexityPerformance(c *fiber.Ctx) error {
	var req ComplexityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate request
	if req.Code == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Code cannot be empty",
		})
	}

	// Forward request to FastAPI complexity performance service
	fastAPIURL := h.cfg.Fastapi.BaseURL + "/api/complexity/performance"

	// Create request to FastAPI
	client := &http.Client{Timeout: 30 * time.Second}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to marshal request",
		})
	}

	httpReq, err := http.NewRequest("POST", fastAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create request",
		})
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(h.cfg.APIKey.APIKeyName, h.cfg.APIKey.APIKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to analyze complexity",
		})
	}
	defer resp.Body.Close()

	// Forward FastAPI response directly
	var fastAPIResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&fastAPIResp); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to decode response from analysis service",
		})
	}

	return c.JSON(fastAPIResp)
}
