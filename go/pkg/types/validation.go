package types

// Validation Types
type ValidationResult struct {
	IsValid bool                   `json:"is_valid"`
	Errors  []ValidationError      `json:"errors,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type FileValidationError struct {
	Field    string `json:"field"`
	Message  string `json:"message"`
	Code     string `json:"code,omitempty"`
	FileName string `json:"file_name,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
	FileType string `json:"file_type,omitempty"`
}

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	TokenType string `json:"token_type"`
}
