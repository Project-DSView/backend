package config

import (
	"fmt"

	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/go-playground/validator/v10"
)

// Validate is the global validator instance
var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

// ValidateConfig validates required configuration values
func ValidateConfig(cfg *config.Config) error {
	if cfg.Google.ClientID == "" {
		return fmt.Errorf("google client id is required")
	}
	if cfg.Google.ClientSecret == "" {
		return fmt.Errorf("google client secret is required")
	}
	if cfg.Google.RedirectURL == "" {
		return fmt.Errorf("google redirect url is required")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}
	if cfg.Frontend.BaseURL == "" {
		return fmt.Errorf("frontend base url is required")
	}
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	if cfg.APIKey.APIKey == "" {
		return fmt.Errorf("api key is required")
	}
	return nil
}
