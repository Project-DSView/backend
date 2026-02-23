package handler

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Project-DSView/backend/go/internal/application/services"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/response"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"github.com/Project-DSView/backend/go/pkg/validation"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	oauthService   *services.OAuthService
	jwtService     *services.JWTService
	userService    *services.UserService
	frontendConfig *config.FrontendConfig
	storageService storage.StorageService
	// Cache to prevent duplicate OAuth code usage
	usedCodes      map[string]time.Time
	usedCodesMutex sync.RWMutex
}

func NewAuthHandler(
	oauthService *services.OAuthService,
	jwtService *services.JWTService,
	userService *services.UserService,
	frontendConfig *config.FrontendConfig,
	storageService storage.StorageService,
) *AuthHandler {
	return &AuthHandler{
		oauthService:   oauthService,
		jwtService:     jwtService,
		userService:    userService,
		frontendConfig: frontendConfig,
		storageService: storageService,
		usedCodes:      make(map[string]time.Time),
	}
}

// GoogleLogin godoc
// @Summary Login with Google OAuth2
// @Description Generate Google OAuth2 login URL and return to client
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "Returns auth URL and state"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/auth/google [get]
func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	state := validation.GenerateState()

	url := h.oauthService.GetAuthURL(state)
	if url == "" {
		return response.SendInternalError(c, "Failed to get Google OAuth URL")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"auth_url": url,
			"state":    state,
		},
	})
}

// GoogleCallback godoc
// @Summary Google OAuth2 callback handler
// @Description Handle callback from Google OAuth2, validate state, exchange code for token, and authenticate user
// @Tags auth
// @Param state query string true "OAuth2 state parameter for CSRF protection"
// @Param code query string false "Authorization code from Google"
// @Param error query string false "OAuth error from Google"
// @Success 302 {string} string "Redirect to frontend with token or error page"
// @Failure 400 {object} map[string]string "Bad request - missing or invalid parameters"
// @Failure 401 {object} map[string]string "Unauthorized - authentication failed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	// Check if this is a direct API call (should be redirect only)
	userAgent := c.Get("User-Agent")
	referer := c.Get("Referer")

	// If it's a direct API call (not from Google), return error
	if !strings.Contains(userAgent, "Mozilla") && !strings.Contains(userAgent, "Chrome") &&
		!strings.Contains(userAgent, "Firefox") && !strings.Contains(userAgent, "Safari") {
		errorMessage := "Direct API calls not allowed. Please use OAuth flow."
		redirectURL := fmt.Sprintf("%s/error?error=%s&code=direct_api_call",
			h.frontendConfig.BaseURL,
			url.QueryEscape(errorMessage))
		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	// Check if request is from Google (should have Google referer)
	if referer != "" && !strings.Contains(referer, "accounts.google.com") {
	}
	// Get parameters from query string
	state := c.Query("state")
	code := c.Query("code")
	errorParam := c.Query("error")

	// Check for OAuth error first
	if errorParam != "" {
		errorMessage := "OAuth error: " + errorParam
		redirectURL := fmt.Sprintf("%s/error?error=%s&code=oauth_error",
			h.frontendConfig.BaseURL,
			url.QueryEscape(errorMessage))
		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	// Validate required parameters
	if code == "" {
		errorMessage := "Authorization code not provided"
		redirectURL := fmt.Sprintf("%s/error?error=%s&code=missing_code",
			h.frontendConfig.BaseURL,
			url.QueryEscape(errorMessage))
		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	if state == "" {
		errorMessage := "State parameter not provided"
		redirectURL := fmt.Sprintf("%s/error?error=%s&code=missing_state",
			h.frontendConfig.BaseURL,
			url.QueryEscape(errorMessage))
		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	// Check if code has been used before (prevent code reuse)
	h.usedCodesMutex.Lock()
	if _, exists := h.usedCodes[code]; exists {
		h.usedCodesMutex.Unlock()
		errorMessage := "Authorization code has already been used. Please try logging in again."
		redirectURL := fmt.Sprintf("%s/error?error=%s&code=code_used",
			h.frontendConfig.BaseURL,
			url.QueryEscape(errorMessage))
		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	// Mark code as used
	h.usedCodes[code] = time.Now()
	h.usedCodesMutex.Unlock()

	// Clean up old codes (older than 10 minutes)
	go h.cleanupOldCodes()

	// Exchange authorization code for token
	token, err := h.oauthService.ExchangeCode(c.Context(), code)
	if err != nil {
		// Remove the failed code from cache
		h.usedCodesMutex.Lock()
		delete(h.usedCodes, code)
		h.usedCodesMutex.Unlock()

		errorMessage := "Failed to exchange authorization code: " + err.Error()
		redirectURL := fmt.Sprintf("%s/error?error=%s&code=exchange_failed",
			h.frontendConfig.BaseURL,
			url.QueryEscape(errorMessage))
		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	// Get user information from Google
	googleUser, err := h.oauthService.GetUserInfo(c.Context(), token)
	if err != nil {
		errorMessage := "Failed to get user information from Google"
		redirectURL := fmt.Sprintf("%s/error?error=%s&code=user_info_failed",
			h.frontendConfig.BaseURL,
			url.QueryEscape(errorMessage))
		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	// Validate required user information
	if googleUser.Email == "" {
		errorMessage := "Email address not provided by Google"
		redirectURL := fmt.Sprintf("%s/error?error=%s&code=no_email",
			h.frontendConfig.BaseURL,
			url.QueryEscape(errorMessage))
		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	// Validate email domain (must be KMITL email)
	isValidDomain := strings.HasSuffix(googleUser.Email, "@kmitl.ac.th") || strings.HasSuffix(googleUser.Email, "@it.kmitl.ac.th")
	if !isValidDomain {
		errorMessage := "ไม่สามารถเข้าใช้งานได้ กรุณาใช้อีเมล KMITL (@kmitl.ac.th หรือ @it.kmitl.ac.th) เท่านั้น"
		redirectURL := fmt.Sprintf("%s/error?error=%s&code=invalid_email_domain",
			h.frontendConfig.BaseURL,
			url.QueryEscape(errorMessage))
		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	// Create or update user in database
	dbUser, err := h.userService.CreateOrUpdateUser(googleUser)
	if err != nil {
		errorMessage := "Failed to create or update user"
		redirectURL := fmt.Sprintf("%s/error?error=%s&code=user_creation_failed",
			h.frontendConfig.BaseURL,
			url.QueryEscape(errorMessage))
		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	// Store refresh token if available
	if token.RefreshToken != "" {
		if err := h.userService.StoreRefreshToken(dbUser.UserID, token.RefreshToken); err != nil {
			// Log error but don't fail the login
			// log.Printf("Failed to store refresh token: %v", err)
		}
	}

	// Generate JWT token
	jwtToken, err := h.jwtService.GenerateToken(dbUser.UserID, dbUser.Email, dbUser.FirstName+" "+dbUser.LastName, dbUser.IsTeacher)
	if err != nil {
		errorMessage := "Failed to generate JWT token"
		redirectURL := fmt.Sprintf("%s/error?error=%s&code=jwt_generation_failed",
			h.frontendConfig.BaseURL,
			url.QueryEscape(errorMessage))
		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	// Redirect to frontend with token in URL (for now)
	userResp := response.ConvertToUserResponse(dbUser)

	// URL encode the token to ensure it's properly formatted
	encodedToken := url.QueryEscape(jwtToken)
	encodedEmail := url.QueryEscape(userResp.Email)

	// Remove the used code from cache to prevent reuse
	h.usedCodesMutex.Lock()
	delete(h.usedCodes, code)
	h.usedCodesMutex.Unlock()

	redirectURL := fmt.Sprintf("%s?token=%s&user=%s&success=true",
		h.frontendConfig.BaseURL,
		encodedToken,
		encodedEmail)

	return c.Redirect(redirectURL, fiber.StatusFound)
}

// Profile godoc
// @Summary Get current user profile
// @Description Retrieve the profile information of the currently authenticated user
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.UserResponse "User profile data"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing token"
// @Failure 404 {object} map[string]string "User not found in database"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/profile [get]
func (h *AuthHandler) Profile(c *fiber.Ctx) error {
	// Get claims from context (set by middleware)
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return response.SendUnauthorized(c, "Unauthorized - invalid claims")
	}

	// Get full user data from database
	dbUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil {
		return response.SendInternalError(c, "Failed to get user profile: "+err.Error())
	}

	if dbUser == nil {
		return response.SendNotFound(c, "User not found")
	}

	// Convert model to response DTO
	userResp := response.ConvertToUserResponse(dbUser)
	return response.SendProfileResponse(c, userResp)
}

// RefreshToken godoc
// @Summary Refresh JWT access token
// @Description Generate a new JWT access token using stored Google refresh token
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{success=bool,message=string,data=object{token=string,user=object}} "New token generated successfully"
// @Failure 401 {object} map[string]string "Unauthorized - no valid session or refresh token expired"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Get current user from JWT claims (if still valid)
	claims, ok := c.Locals("claims").(*types.Claims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "No valid session found",
		})
	}

	// Get stored refresh token
	refreshToken, err := h.userService.GetRefreshToken(claims.UserID)
	if err != nil || refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "No refresh token available - please login again",
		})
	}

	// Try to refresh Google token
	newToken, err := h.oauthService.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		// Refresh token might be expired, remove it
		h.userService.RemoveRefreshToken(claims.UserID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Refresh token expired - please login again",
		})
	}

	// Store new refresh token if provided
	if newToken.RefreshToken != "" {
		h.userService.StoreRefreshToken(claims.UserID, newToken.RefreshToken)
	}

	// Generate new JWT
	jwtUser, err := h.userService.GetUserByID(claims.UserID)
	if err != nil || jwtUser == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get user information",
		})
	}

	jwtToken, err := h.jwtService.GenerateToken(jwtUser.UserID, jwtUser.Email, jwtUser.FirstName+" "+jwtUser.LastName, jwtUser.IsTeacher)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to generate new token",
		})
	}

	// Return new token and user data in JSON response
	userResp := response.ConvertToUserResponse(jwtUser)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Token refreshed successfully",
		"data": fiber.Map{
			"token": jwtToken,
			"user":  userResp,
		},
	})
}

// Logout godoc
// @Summary Logout user
// @Description Remove refresh token from database and clear OAuth codes (client should discard JWT token)
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string "Logout successful"
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Get user ID for cleanup
	if claims, ok := c.Locals("claims").(*types.Claims); ok {
		// Remove refresh token from database
		h.userService.RemoveRefreshToken(claims.UserID)
	}

	// Clear all OAuth codes to prevent reuse
	h.usedCodesMutex.Lock()
	h.usedCodes = make(map[string]time.Time)
	h.usedCodesMutex.Unlock()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logout successful",
	})
}

// cleanupOldCodes removes codes older than 10 minutes to prevent memory leaks
func (h *AuthHandler) cleanupOldCodes() {
	h.usedCodesMutex.Lock()
	defer h.usedCodesMutex.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	for code, usedTime := range h.usedCodes {
		if usedTime.Before(cutoff) {
			delete(h.usedCodes, code)
		}
	}
}
