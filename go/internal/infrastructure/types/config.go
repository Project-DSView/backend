package types

// Infrastructure Types
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Google   GoogleConfig   `yaml:"google"`
	JWT      JWTConfig      `yaml:"jwt"`
	Frontend FrontendConfig `yaml:"frontend"`
	Fastapi  FastapiConfig  `yaml:"fastapi"`
	Executor ExecutorConfig `yaml:"executor"`
	Storage  StorageConfig  `yaml:"storage"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	APIKey   APIKeyConfig   `yaml:"api_key"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type GoogleConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURL  string `yaml:"redirect_url"`
}

type JWTConfig struct {
	SecretKey     string `yaml:"secret_key"`
	Expiration    int    `yaml:"expiration"`
	RefreshExpiry int    `yaml:"refresh_expiry"`
}

type FrontendConfig struct {
	URL string `yaml:"url"`
}

type FastapiConfig struct {
	URL     string `yaml:"url"`
	Timeout int    `yaml:"timeout"`
}

type ExecutorConfig struct {
	MaxConcurrency int `yaml:"max_concurrency"`
	Timeout        int `yaml:"timeout"`
}

type MinIOConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl"`
	BucketName      string `yaml:"bucket_name"`
	CoursePrefix    string `yaml:"course_prefix"`
	CodePrefix      string `yaml:"code_prefix"`
}

type S3Config struct {
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	BucketName      string `yaml:"bucket_name"`
}

type StorageConfig struct {
	// Using MinIO only - local storage removed
	MinIO MinIOConfig `yaml:"minio"`
}

type RabbitMQConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	VHost    string `yaml:"vhost"`
}

type APIKeyConfig struct {
	SecretKey string `yaml:"secret_key"`
}
