# DSView Backend API

Backend API สำหรับระบบ DSView - Stateless Architecture ที่ให้บริการ Code Execution พร้อม Step-by-step Visualization

## 🚀 Features

- **Playground API**: Code execution with step-by-step visualization (No Auth Required)
- **Data Structure Support**: 
  - Stack
  - Singly Linked List
  - Doubly Linked List
  - Binary Search Tree
  - Undirected Graph
  - Directed Graph
- **Stateless Architecture**: No database storage, all data processed in memory
- **Rate Limiting**: Built-in protection against abuse
- **Security**: API key authentication and code validation
- **Docker Support**: Ready for containerized deployment

## 📋 Prerequisites

- Python 3.12+
- Docker (optional)
- pip

## 🛠️ Installation

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd backend/fastapi
   ```

2. **Create virtual environment**
   ```bash
   python -m venv venv
   # Windows
   venv\Scripts\activate
   # Linux/Mac
   source venv/bin/activate
   ```

3. **Install dependencies**
   ```bash
   make install
   # หรือ
   pip install -r requirements.txt
   ```

4. **Setup environment variables**
   ```bash
   cp env.example .env
   # แก้ไขค่าใน .env ตามต้องการ
   ```

5. **Run the application**
   ```bash
   # Development mode (with auto-reload)
   make dev
   
   # Production mode
   make run
   ```

### Docker Deployment

1. **Build Docker image**
   ```bash
   make docker-build
   ```

2. **Run with Docker Compose**
   ```bash
   docker-compose -f docker-compose.dev.yml up
   ```

## ⚙️ Configuration

สร้างไฟล์ `.env` จาก `env.example` และปรับแต่งค่าต่างๆ:

```env
# API
API_KEY_NAME=dsview-api-key
API_KEY=change_me_secure_key

# CORS (comma-separated)
ALLOW_ORIGINS=http://127.0.0.1:5500,http://localhost:3000,http://127.0.0.1:8080,http://localhost:8080,https://localhost:3000

# Execution Settings
MAX_CODE_LENGTH=10000
EXECUTION_TIMEOUT=30
MAX_LOOPS=15
MAX_FOR_LOOPS=20
MAX_FUNCTIONS=30

# Rate Limiting
RATE_LIMIT_PER_MINUTE=10
RATE_LIMIT_PER_SECOND=2
```

## 📚 API Documentation

เมื่อรันแอปพลิเคชันแล้ว สามารถเข้าถึง API documentation ได้ที่:

- **Swagger UI**: http://localhost:8000/docs
- **ReDoc**: http://localhost:8000/redoc
- **OpenAPI JSON**: http://localhost:8000/openapi.json

## 🔌 API Endpoints

### Playground API

#### POST `/api/playground/run`

Execute code และรับ step-by-step visualization

**Headers:**
```
X-API-Key: your-api-key
Content-Type: application/json
```

**Request Body:**
```json
{
  "code": "class Stack:\n    def __init__(self):\n        self.items = []\n    def push(self, item):\n        self.items.append(item)\n    def pop(self):\n        return self.items.pop()\n\ns = Stack()\ns.push(1)\ns.push(2)\nprint(s.pop())",
  "dataType": "stack"
}
```

**Response:**
```json
{
  "executionId": "uuid-string",
  "code": "executed code",
  "dataType": "stack",
  "steps": [
    {
      "stepNumber": 1,
      "line": 1,
      "code": "class Stack:",
      "state": {}
    }
  ],
  "totalSteps": 10,
  "status": "success",
  "errorMessage": null,
  "executedAt": "2024-01-01T00:00:00Z",
  "createdAt": "2024-01-01T00:00:00Z"
}
```

### Health Check

#### GET `/health`
ตรวจสอบสถานะของ API

#### GET `/`
ข้อมูลพื้นฐานของ API

## 🛡️ Security Features

- **API Key Authentication**: ต้องใช้ API key ในการเรียกใช้
- **Code Validation**: ตรวจสอบและป้องกัน dangerous code patterns
- **Rate Limiting**: จำกัดจำนวน request ต่อนาที/วินาที
- **Input Sanitization**: ทำความสะอาด input ก่อนประมวลผล
- **Timeout Protection**: ป้องกัน infinite loops และ long-running code

## 📁 Project Structure

```
fastapi/
├── app/
│   ├── api/
│   │   ├── controllers/     # Business logic controllers
│   │   ├── routes/         # API route definitions
│   │   └── health.py       # Health check endpoints
│   ├── core/
│   │   ├── config.py       # Configuration management
│   │   ├── security.py     # Security utilities
│   │   └── startup.py      # App startup events
│   ├── examples/
│   │   └── exercises/      # Example code for each data structure
│   ├── middleware/
│   │   ├── cors.py         # CORS configuration
│   │   ├── rate_limiting.py # Rate limiting middleware
│   │   └── request_id.py   # Request ID middleware
│   ├── schemas/
│   │   └── playground.py   # Pydantic models
│   ├── services/
│   │   └── simulators/     # Data structure simulators
│   └── utils/
│       └── execution_helpers.py # Code execution utilities
├── dockerfile
├── docker-compose.dev.yml
├── requirements.txt
├── Makefile
└── README.md
```

## 🧪 Testing

```bash
# Run basic syntax check
make lint

# Run tests (when implemented)
make test
```

## 🚀 Deployment

### Production with Docker

1. **Build production image**
   ```bash
   docker build -t dsview-backend:latest -f dockerfile .
   ```

2. **Run container**
   ```bash
   docker run -d \
     --name dsview-backend \
     -p 8000:8000 \
     --env-file .env \
     --restart unless-stopped \
     dsview-backend:latest
   ```

### Docker Compose

```bash
docker-compose -f docker-compose.dev.yml up -d
```

## 📊 Performance

- **Gunicorn**: Multi-worker WSGI server
- **Uvicorn Workers**: Async-capable workers
- **Worker Configuration**: 4 workers, optimized for production
- **Memory Management**: Stateless design for better scalability

## 🔧 Available Commands

```bash
make help          # Show all available commands
make install       # Install dependencies
make dev           # Run in development mode
make run           # Run in production mode
make lint          # Run syntax check
make test          # Run tests
make docker-build  # Build Docker image
make docker-run    # Run Docker container
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is part of the DSView system.

## 🆘 Support

หากมีปัญหาหรือคำถาม สามารถติดต่อได้ผ่าน:
- GitHub Issues
- Project documentation

---

**Version**: 0.0.5-alpha  
**Python**: 3.12+  
**FastAPI**: 0.116.1
