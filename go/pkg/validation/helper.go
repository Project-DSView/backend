package validation

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
)

// GetEnv gets an environment variable with a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GenerateState generates a cryptographically secure random state string for OAuth
func GenerateState() string {
	bytes := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based state if random generation fails
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(bytes)
}

type TokenResponse struct {
	AccessToken string       `json:"access_token"`
	ExpiresIn   int          `json:"expires_in"`
	TokenType   string       `json:"token_type"`
	User        *models.User `json:"user"`
}

// CreateUserDataString creates a URL-safe user data string (ใช้ถ้าจะ embed user ข้อมูลใน JWT claim)
func CreateUserDataString(user *models.User) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		url.QueryEscape(user.UserID),
		url.QueryEscape(user.FirstName),
		url.QueryEscape(user.LastName),
		url.QueryEscape(user.Email),
		url.QueryEscape(strconv.FormatBool(user.IsTeacher)),
		url.QueryEscape(user.ProfileImg),
	)
}

// ParseUserDataString parses the user data string
func ParseUserDataString(data string) (*models.User, error) {
	parts := strings.Split(data, "|")
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid user data format")
	}

	userID, _ := url.QueryUnescape(parts[0])
	firstName, _ := url.QueryUnescape(parts[1])
	lastName, _ := url.QueryUnescape(parts[2])
	email, _ := url.QueryUnescape(parts[3])
	isTeacherStr, _ := url.QueryUnescape(parts[4])
	profileImg, _ := url.QueryUnescape(parts[5])
	isTeacher, err := strconv.ParseBool(isTeacherStr)
	if err != nil {
		return nil, fmt.Errorf("invalid isTeacher value: %s", isTeacherStr)
	}

	return &models.User{
		UserID:     userID,
		FirstName:  firstName,
		LastName:   lastName,
		Email:      email,
		IsTeacher:  isTeacher,
		ProfileImg: profileImg,
	}, nil
}

// BuildRedirectURL builds a redirect URL with query parameters
func BuildRedirectURL(baseURL string, params map[string]string) string {
	if baseURL == "" {
		return ""
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL // Return original URL if parsing fails
	}

	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// ValidateState validates the OAuth state parameter
func ValidateState(receivedState, storedState string) error {
	if receivedState == "" {
		return fmt.Errorf("no state parameter provided")
	}
	if storedState == "" {
		return fmt.Errorf("no stored state found")
	}
	if receivedState != storedState {
		return fmt.Errorf("state parameter mismatch")
	}
	return nil
}

// BuildUserUpdates builds a map of user updates for database operations
func BuildUserUpdates(firstName, lastName string) map[string]interface{} {
	return map[string]interface{}{
		"first_name": SanitizeUserInput(firstName),
		"last_name":  SanitizeUserInput(lastName),
	}
}
