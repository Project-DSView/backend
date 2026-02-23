package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Google   GoogleConfig
	JWT      JWTConfig
	Frontend FrontendConfig
	Fastapi  FastapiConfig
	Executor ExecutorConfig
	MinIO    MinIOConfig
	RabbitMQ RabbitMQConfig
	APIKey   APIKeyConfig
}

type ServerConfig struct {
	Port        string
	Host        string
	Environment string
}

// database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	TimeZone        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // minutes
	ConnMaxIdleTime int // minutes
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

type FrontendConfig struct {
	BaseURL         string
	AuthSuccessPath string
	AuthErrorPath   string
	AllowedOrigins  []string
}

type FastapiConfig struct {
	BaseURL     string
	Timeout     time.Duration
	RetryCount  int
	RetryDelay  time.Duration
	HealthCheck bool
}

type ExecutorConfig struct {
	Image   string
	Timeout time.Duration
	Memory  string
	CPUs    string
}

type MinIOConfig struct {
	Endpoint        string
	PublicEndpoint  string // Public endpoint for presigned URLs (accessible from browser)
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	MaxFileSize     string
	UseSSL          bool
	PublicBucket    bool
}

type RabbitMQConfig struct {
	URL      string
	Exchange string
	Username string
	Password string
	Host     string
	Port     int
	VHost    string
}

type APIKeyConfig struct {
	APIKeyName string
	APIKey     string
}

func Load(env string) (*Config, error) {
	config := &Config{}

	// Load server configuration
	config.Server = ServerConfig{
		Host:        getEnvOrDefault("SERVER_HOST", "127.0.0.1"),
		Port:        getEnvOrDefault("SERVER_PORT", "8080"),
		Environment: getEnvOrDefault("SERVER_ENV", env),
	}

	// Load database configuration
	config.Database = DatabaseConfig{
		Host:            getEnvOrDefault("DB_HOST", "postgres"),
		Port:            getEnvAsInt("DB_PORT", 5432),
		User:            getEnvOrDefault("DB_USER", "postgres"),
		Password:        getEnvOrDefault("DB_PASSWORD", ""),
		DBName:          getEnvOrDefault("DB_NAME", ""),
		SSLMode:         getEnvOrDefault("DB_SSLMODE", "disable"),
		TimeZone:        getEnvOrDefault("DB_TIMEZONE", "UTC"),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 30),
		ConnMaxIdleTime: getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 5),
	}

	// Load Google OAuth configuration
	config.Google = GoogleConfig{
		ClientID:     getEnvOrDefault("GOOGLE_CLIENT_ID", ""),
		ClientSecret: getEnvOrDefault("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  getEnvOrDefault("GOOGLE_REDIRECT_URL", ""),
		Scopes: getEnvAsStringSlice("GOOGLE_SCOPES", []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		}),
	}

	// Load JWT configuration
	config.JWT = JWTConfig{
		Secret:    getEnvOrDefault("JWT_SECRET", ""),
		ExpiresIn: getEnvAsDuration("JWT_EXPIRES_IN", 24*time.Hour),
	}

	// Load frontend configuration
	config.Frontend = FrontendConfig{
		BaseURL:         getEnvOrDefault("FRONTEND_BASE_URL", "http://localhost:3000"),
		AuthSuccessPath: getEnvOrDefault("FRONTEND_AUTH_SUCCESS_PATH", "/"),
		AuthErrorPath:   getEnvOrDefault("FRONTEND_AUTH_ERROR_PATH", "/html/error.html"),
		AllowedOrigins: getEnvAsStringSlice("FRONTEND_ALLOWED_ORIGINS", []string{
			"http://127.0.0.1:8080",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://localhost:3000",
			"https://127.0.0.1:3000",
			"https://localhost:3000",
			"http://127.0.0.1:5500",
			"http://localhost:5500",
		}),
	}

	// Load FastAPI configuration
	config.Fastapi = FastapiConfig{
		BaseURL:     getEnvOrDefault("FASTAPI_BASE_URL", "http://fastapi:8000"),
		Timeout:     getEnvAsDuration("FASTAPI_TIMEOUT", 30*time.Second),
		RetryCount:  getEnvAsInt("FASTAPI_RETRY_COUNT", 3),
		RetryDelay:  getEnvAsDuration("FASTAPI_RETRY_DELAY", 1*time.Second),
		HealthCheck: getEnvAsBool("FASTAPI_HEALTH_CHECK", true),
	}

	// Load executor configuration
	config.Executor = ExecutorConfig{
		Image:   getEnvOrDefault("EXECUTOR_IMAGE", "python:3.12-alpine"),
		Timeout: getEnvAsDuration("EXECUTOR_TIMEOUT", 15*time.Second),
		Memory:  getEnvOrDefault("EXECUTOR_MEMORY", "512m"),
		CPUs:    getEnvOrDefault("EXECUTOR_CPUS", "1.0"),
	}

	// Load MinIO configuration
	config.MinIO = MinIOConfig{
		Endpoint:        getEnvOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		PublicEndpoint:  getEnvOrDefault("MINIO_PUBLIC_ENDPOINT", ""), // Empty means use Endpoint
		AccessKeyID:     getEnvOrDefault("MINIO_ACCESS_KEY_ID", "minioadmin"),
		SecretAccessKey: getEnvOrDefault("MINIO_SECRET_ACCESS_KEY", "minioadmin"),
		BucketName:      getEnvOrDefault("MINIO_BUCKET_NAME", "dsview"),
		MaxFileSize:     getEnvOrDefault("MINIO_MAX_FILE_SIZE", "5MB"),
		UseSSL:          getEnvAsBool("MINIO_USE_SSL", false),
		PublicBucket:    getEnvAsBool("MINIO_PUBLIC_BUCKET", true),
	}

	// Storage configuration removed - using MinIO only

	// Load RabbitMQ configuration
	config.RabbitMQ = RabbitMQConfig{
		Host:     getEnvOrDefault("RABBITMQ_HOST", "rabbitmq"),
		Port:     getEnvAsInt("RABBITMQ_PORT", 5672),
		Username: getEnvOrDefault("RABBITMQ_USERNAME", "admin"),
		Password: getEnvOrDefault("RABBITMQ_PASSWORD", "admin"),
		VHost:    getEnvOrDefault("RABBITMQ_VHOST", "/"),
		Exchange: getEnvOrDefault("RABBITMQ_EXCHANGE", "dsview_exchange"),
		URL:      getEnvOrDefault("RABBITMQ_URL", ""),
	}

	// Generate RabbitMQ URL if not provided
	if config.RabbitMQ.URL == "" {
		config.RabbitMQ.URL = fmt.Sprintf("amqp://%s:%s@%s:%d%s",
			config.RabbitMQ.Username, config.RabbitMQ.Password,
			config.RabbitMQ.Host, config.RabbitMQ.Port, config.RabbitMQ.VHost)
	}

	// Load API key configuration
	config.APIKey = APIKeyConfig{
		APIKeyName: getEnvOrDefault("API_KEY_NAME", "dsview-api-key"),
		APIKey:     getEnvOrDefault("API_KEY", ""),
	}

	// Set defaults if not provided
	config.setDefaults()

	return config, nil
}

func (c *Config) setDefaults() {
	// Frontend defaults
	if c.Frontend.AuthSuccessPath == "" {
		c.Frontend.AuthSuccessPath = "/auth/success"
	}
	if c.Frontend.AuthErrorPath == "" {
		c.Frontend.AuthErrorPath = "/auth/error"
	}
	if len(c.Frontend.AllowedOrigins) == 0 {
		c.Frontend.AllowedOrigins = []string{
			"http://127.0.0.1:8080",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://localhost:3000",
			"http://127.0.0.1:5500",
			"http://localhost:5500",
		}
	}

	// Set default database name if not specified
	if c.Database.DBName == "" {
		c.Database.DBName = "DSView_DB"
	}

	// Database connection pool defaults
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25 // Maximum number of open connections
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5 // Maximum number of idle connections
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 30 // Connection max lifetime in minutes
	}
	if c.Database.ConnMaxIdleTime == 0 {
		c.Database.ConnMaxIdleTime = 5 // Connection max idle time in minutes
	}

	if c.Executor.Image == "" {
		c.Executor.Image = "python:3.11"
	}
	if c.Executor.Timeout == 0 {
		c.Executor.Timeout = 5 * time.Second
	}
	if c.Executor.Memory == "" {
		c.Executor.Memory = "256m"
	}
	if c.Executor.CPUs == "" {
		c.Executor.CPUs = "0.5"
	}

	if c.MinIO.Endpoint == "" {
		c.MinIO.Endpoint = "localhost:9000"
	}
	if c.MinIO.BucketName == "" {
		c.MinIO.BucketName = "dsview"
	}
	if c.MinIO.MaxFileSize == "" {
		c.MinIO.MaxFileSize = "10MB" // เพิ่มจาก 1MB เป็น 10MB สำหรับ PDF
	}

	// RabbitMQ defaults
	if c.RabbitMQ.Host == "" {
		c.RabbitMQ.Host = "rabbitmq"
	}
	if c.RabbitMQ.Port == 0 {
		c.RabbitMQ.Port = 5672
	}
	if c.RabbitMQ.Username == "" {
		c.RabbitMQ.Username = "admin"
	}
	if c.RabbitMQ.Password == "" {
		c.RabbitMQ.Password = "admin"
	}
	if c.RabbitMQ.VHost == "" {
		c.RabbitMQ.VHost = "/"
	}
	if c.RabbitMQ.Exchange == "" {
		c.RabbitMQ.Exchange = "dsview_exchange"
	}
	if c.RabbitMQ.URL == "" {
		c.RabbitMQ.URL = fmt.Sprintf("amqp://%s:%s@%s:%d%s",
			c.RabbitMQ.Username, c.RabbitMQ.Password, c.RabbitMQ.Host, c.RabbitMQ.Port, c.RabbitMQ.VHost)
	}

	// API Key defaults
	if c.APIKey.APIKeyName == "" {
		c.APIKey.APIKeyName = "X-API-Key"
	}
}

// Single DSN method for the unified database
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode, d.TimeZone)
}

// GetFrontendURL constructs the frontend URL for redirects
func (f *FrontendConfig) GetAuthSuccessURL() string {
	return f.BaseURL + f.AuthSuccessPath
}

// GetAuthErrorURL constructs the frontend error URL
func (f *FrontendConfig) GetAuthErrorURL() string {
	return f.BaseURL + f.AuthErrorPath
}

// Helper methods สำหรับแปลง file size
func (m *MinIOConfig) GetMaxFileSizeBytes() int64 {
	return parseFileSize(m.MaxFileSize)
}

// GetMaxFileSizeBytes is now handled by MinIOConfig
// func (s *StorageConfig) GetMaxFileSizeBytes() int64 {
//	return parseFileSize(s.MaxFileSize)
// }

func parseFileSize(size string) int64 {
	if size == "" {
		return 1024 * 1024 // 1MB default
	}

	// Simple parser for sizes like "5MB", "1GB"
	var multiplier int64 = 1
	var numStr string

	if len(size) >= 2 {
		suffix := size[len(size)-2:]
		numStr = size[:len(size)-2]

		switch suffix {
		case "KB":
			multiplier = 1024
		case "MB":
			multiplier = 1024 * 1024
		case "GB":
			multiplier = 1024 * 1024 * 1024
		default:
			// No suffix, assume bytes
			numStr = size
		}
	} else {
		numStr = size
	}

	if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
		return num * multiplier
	}

	return 1024 * 1024 // Default 1MB
}

// Helper functions for environment variable handling
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
