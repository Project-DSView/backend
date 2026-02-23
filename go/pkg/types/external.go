package types

// External Service Types
type DockerConfig struct {
	ImageName   string `yaml:"image_name"`
	MaxMemory   string `yaml:"max_memory"`
	MaxCPUs     string `yaml:"max_cpus"`
	Timeout     int    `yaml:"timeout"`
	NetworkMode string `yaml:"network_mode"`
}

type ExecResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Duration int64  `json:"duration_ms"`
}

type RabbitMQConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	VHost    string `yaml:"vhost"`
}

type QueueMessage struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt string                 `json:"created_at"`
}

type MinIOConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl"`
	BucketName      string `yaml:"bucket_name"`
}

type S3Config struct {
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	BucketName      string `yaml:"bucket_name"`
}

type Permission int

const (
	PermissionRead Permission = iota
	PermissionWrite
	PermissionDelete
	PermissionAdmin
)

type PermissionError struct {
	Required Permission `json:"required"`
	Actual   Permission `json:"actual"`
	Message  string     `json:"message"`
}

func (e *PermissionError) Error() string {
	return e.Message
}
