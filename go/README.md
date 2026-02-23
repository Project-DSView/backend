# DSView Backend API - Go Service

Backend API สำหรับระบบ DSView - Full-featured Learning Management System with Authentication, Exercise Management, และ Code Execution

## 🚀 Features

- **Authentication & Authorization**: Google OAuth 2.0, JWT tokens, Role-based access control
- **User Management**: Student, TA, Teacher, Admin roles
- **Exercise Management**: Create, edit, delete exercises with test cases
- **Course Management**: Course creation, enrollment, materials, announcements
- **Code Execution**: Integration with FastAPI playground service
- **File Storage**: MinIO/S3 integration for file uploads
- **Progress Tracking**: Student progress monitoring and verification
- **Deadline Management**: Exercise deadline checking and notifications
- **Queue System**: RabbitMQ integration for background tasks
- **API Documentation**: Swagger/OpenAPI documentation
- **Database**: PostgreSQL with GORM ORM

## 📋 Prerequisites

- Go 1.24.6+
- PostgreSQL 12+
- MinIO (or AWS S3)
- RabbitMQ (optional)
- Docker (optional)

## 🛠️ Installation

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd backend/go
   ```

2. **Install dependencies**
   ```bash
   make deps
   # หรือ
   go mod tidy
   go mod download
   ```

3. **Setup configuration**
   ```bash
   make setup
   # หรือ
   cp configs/development.yaml.example configs/development.yaml
   ```

4. **Configure environment**
   แก้ไขไฟล์ `configs/development.yaml` ตามต้องการ:
   ```yaml
   server:
     host: "127.0.0.1"
     port: "8080"
   
   database:
     host: "localhost"
     port: 5432
     user: "postgres"
     password: "your_password"
     dbname: "dsview_db"
   
   google:
     client_id: "your_google_client_id"
     client_secret: "your_google_client_secret"
   
   jwt:
     secret: "your_jwt_secret"
   ```

5. **Start infrastructure services**
   ```bash
   make infra-up
   # หรือ
   docker-compose up -d postgres minio rabbitmq
   ```

6. **Run the application**
   ```bash
   make run
   # หรือ
   go run cmd/server/main.go
   ```

### Docker Deployment

1. **Build Docker image**
   ```bash
   make docker-build
   # หรือ
   docker build -t dsview-backend .
   ```

2. **Run with Docker Compose**
   ```bash
   docker-compose up -d
   ```

## ⚙️ Configuration

### Environment Variables

```bash
APP_ENV=development  # development, production
```

### Configuration Files

- `configs/development.yaml` - Development configuration
- `configs/production.yaml` - Production configuration

### Key Configuration Sections

```yaml
server:
  host: "127.0.0.1"
  port: "8080"

database:
  host: "postgres"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "dsview_db"

google:
  client_id: "your_google_client_id"
  client_secret: "your_google_client_secret"
  redirect_url: "http://127.0.0.1:8080/auth/google/callback"

jwt:
  secret: "your_jwt_secret"
  expires_in: 24h

minio:
  endpoint: "minio:9000"
  access_key_id: "minioadmin"
  secret_access_key: "minioadmin"
  bucket_name: "dsview"

rabbitmq:
  host: "rabbitmq"
  port: 5672
  username: "admin"
  password: "admin"
```

## 📚 API Documentation

เมื่อรันแอปพลิเคชันแล้ว สามารถเข้าถึง API documentation ได้ที่:

- **Swagger UI**: http://localhost:8080/docs/
- **OpenAPI JSON**: http://localhost:8080/docs/doc.json
- **API Info**: http://localhost:8080/ (ต้องใช้ API key)

## 🔌 API Endpoints

### Authentication

#### GET `/api/auth/google`
เริ่มต้น Google OAuth flow

#### GET `/api/auth/google/callback`
Google OAuth callback

#### POST `/api/auth/logout`
ออกจากระบบ

#### POST `/api/auth/refresh`
Refresh JWT token

### User Management

#### GET `/api/profile`
ดูข้อมูล profile ของผู้ใช้ปัจจุบัน

#### PUT `/api/profile`
อัปเดตข้อมูล profile

#### GET `/api/users` (Admin only)
ดูรายการผู้ใช้ทั้งหมด

### Exercise Management

#### GET `/api/exercises`
ดูรายการ exercises

#### POST `/api/exercises`
สร้าง exercise ใหม่

#### GET `/api/exercises/:id`
ดูรายละเอียด exercise

#### PUT `/api/exercises/:id`
อัปเดต exercise

#### DELETE `/api/exercises/:id`
ลบ exercise

### Course Management

#### GET `/api/courses`
ดูรายการ courses

#### POST `/api/courses`
สร้าง course ใหม่

#### GET `/api/courses/:id`
ดูรายละเอียด course

#### POST `/api/courses/:id/enroll`
สมัครเรียน course

### Code Execution

#### POST `/api/exec/run`
ส่งโค้ดไปประมวลผลที่ FastAPI service

#### GET `/api/exec/:source`
ดูประวัติการประมวลผล

### File Upload

#### POST `/api/course-materials/upload`
อัปโหลดไฟล์ course materials

## 🛡️ Security Features

- **JWT Authentication**: Stateless authentication with JWT tokens
- **Google OAuth 2.0**: Social login integration
- **Role-based Access Control**: Student, TA, Teacher, Admin roles
- **API Key Protection**: API key authentication for secure endpoints
- **CORS Configuration**: Configurable CORS settings
- **Input Validation**: Request validation and sanitization
- **Rate Limiting**: Built-in rate limiting (via middleware)

## 📁 Project Structure

```
go/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handler/             # HTTP handlers
│   │   ├── middleware/          # Middleware functions
│   │   └── routes/              # Route definitions
│   ├── application/
│   │   └── services/            # Business logic services
│   ├── domain/
│   │   ├── entities/            # Domain models
│   │   ├── enums/               # Enumerations
│   │   └── repositories/        # Repository interfaces
│   ├── infrastructure/
│   │   ├── config/              # Configuration management
│   │   ├── database/            # Database setup
│   │   └── repositories/        # Repository implementations
│   └── types/                   # Common types
├── pkg/
│   ├── auth/                    # Authentication utilities
│   ├── handlers/                # Common handlers
│   ├── response/                # Response utilities
│   ├── storage/                 # Storage interfaces
│   └── validation/              # Validation utilities
├── docs/                        # Swagger documentation
├── configs/                     # Configuration files
├── html/                        # Test HTML files
├── go.mod
├── go.sum
├── Dockerfile
├── Makefile
└── README.md
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run short tests
make test-short

# Run benchmarks
make benchmark
```

## 🔧 Available Commands

```bash
make help              # Show all available commands
make setup             # Initial setup for development
make deps              # Install dependencies
make run               # Run in development mode
make run-dev           # Run with development environment
make run-prod          # Run with production environment
make build             # Build the application
make build-linux       # Build for Linux
make build-windows     # Build for Windows
make test              # Run tests
make test-coverage     # Run tests with coverage
make fmt               # Format Go code
make lint              # Run linter
make vet               # Run go vet
make check             # Run all code quality checks
make swagger           # Generate Swagger documentation
make docker-build      # Build Docker image
make docker-run        # Run Docker container
make infra-up          # Start infrastructure services
make infra-down        # Stop infrastructure services
make clean             # Clean all generated files
```

## 🚀 Deployment

### Production with Docker

1. **Build production image**
   ```bash
   make docker-build
   ```

2. **Run with Docker Compose**
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

### Manual Deployment

1. **Build binary**
   ```bash
   make build-linux
   ```

2. **Copy files to server**
   ```bash
   scp bin/dsview-backend-linux user@server:/opt/dsview/
   scp configs/production.yaml user@server:/opt/dsview/
   ```

3. **Run on server**
   ```bash
   ./dsview-backend-linux
   ```

## 📊 Performance

- **Fiber Framework**: High-performance HTTP framework
- **GORM**: Efficient ORM with connection pooling
- **PostgreSQL**: Robust relational database
- **MinIO/S3**: Scalable object storage
- **RabbitMQ**: Message queue for background tasks
- **JWT**: Stateless authentication for scalability

## 🔄 Integration

### FastAPI Service Integration

Go service integrates with FastAPI playground service for code execution:

```yaml
fastapi:
  base_url: "http://localhost:8000"
  timeout: 30s
  retry_count: 3
  health_check: true
```

### External Services

- **Google OAuth 2.0**: User authentication
- **PostgreSQL**: Primary database
- **MinIO/S3**: File storage
- **RabbitMQ**: Message queue
- **FastAPI**: Code execution service

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run `make check` to ensure code quality
6. Submit a pull request

## 📄 License

This project is part of the DSView system.

## 🆘 Support

หากมีปัญหาหรือคำถาม สามารถติดต่อได้ผ่าน:
- GitHub Issues
- Project documentation
- Email: 65070209@kmitl.ac.th

---

**Version**: 1.0.0-alpha  
**Go Version**: 1.24.6  
**Fiber**: v2.52.9  
**GORM**: v1.30.1
