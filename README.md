# DSView Backend Services

## 🚀 Quick Start

### 1. Setup Environment
```bash
# Copy environment file
cp env.example .env

# Edit with your Docker Hub username
nano .env
```

### 2. Docker Hub Configuration
```bash
# Edit .env file
DOCKER_HUB_USERNAME=your-username
DOCKER_HUB_REPO=dsview-backend
VERSION=latest
```

### 3. Start All Services
```bash
# Start all services with Traefik
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

## 🌐 Service URLs

### **Monitoring & Logging**
- **Grafana Dashboard**: http://localhost:3001 (admin/admin)
- **Traefik Dashboard**: http://localhost:8080
- **Loki API**: http://localhost:3100
- **Prometheus Metrics**: http://localhost:8080/metrics

### **API Services**
- **FastAPI**: http://api.fastapi.localhost หรือ http://localhost/fastapi
- **Go API**: http://api.go.localhost หรือ http://localhost/go

### **Database & Storage**
- **PostgreSQL**: localhost:5432
- **MinIO Console**: http://localhost:9001
- **RabbitMQ Management**: http://localhost:15672

## 🔧 Configuration

### **Environment Variables**
```bash
# Database
POSTGRES_DB=dsview_db
POSTGRES_USER=postgres
POSTGRES_PASSWORD=1234

# MinIO
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin

# RabbitMQ
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=admin
```

### **Service Routing**
- **Path-based**: `/fastapi/*` → FastAPI service
- **Path-based**: `/go/*` → Go service
- **Host-based**: `api.fastapi.localhost` → FastAPI service
- **Host-based**: `api.go.localhost` → Go service

## 🐛 Debugging

### **View Logs**
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f fastapi
docker-compose logs -f go-app
docker-compose logs -f traefik

# Structured logs (JSON format)
tail -f fastapi/logs/fastapi_structured.log
tail -f go/logs/go_app.log

# Access Grafana for centralized log viewing
# http://localhost:3001 (admin/admin)
```

### **Health Checks**
```bash
# FastAPI health
curl http://localhost/fastapi/health

# Go health
curl http://localhost/go/health

# Traefik dashboard
curl http://localhost:8080
```

## 📁 Project Structure
```
backend/
├── docker-compose.yml      # Main compose file
├── .env                    # Environment variables
├── env.example            # Environment template
├── traefik/
│   └── logs/              # Traefik logs
├── fastapi/               # FastAPI service
├── go/                    # Go service
└── README.md              # This file
```

## 🔄 Development Workflow

### **Start Development**
```bash
# Start all services
make dev

# View real-time logs
make logs
```

### **Docker Hub Management**
```bash
# Build images
make build

# Push to Docker Hub
make push

# Build and push
make build-push

# Pull from Docker Hub
make pull
```

### **Production Deployment**
```bash
# Start production environment
make prod

# Stop all services
make stop
```

### **Restart Service**
```bash
# Restart development
make dev-restart

# Restart production
make prod-restart
```

## 📊 Monitoring & Logging

### **Enhanced Observability**
- ✅ **Centralized Logging**: All logs in Grafana UI
- ✅ **Request Tracing**: Track requests across services
- ✅ **Real-time Dashboards**: Beautiful monitoring dashboards
- ✅ **Structured Logs**: JSON format with correlation IDs
- ✅ **Performance Metrics**: Response times, error rates
- ✅ **Security Monitoring**: Rate limiting, security headers

### **Quick Log Queries**
```bash
# View all logs
{job=~".+"}

# View error logs
{job=~".+"} |= "ERROR"

# View slow requests
{service=~"fastapi|go-app"} | json | duration > 1.0

# View requests by user
{service=~"fastapi|go-app"} | json | user_id != ""
```

### **Access Monitoring**
1. **Grafana**: http://localhost:3001 (admin/admin)
2. **Navigate to "Dashboards"** for pre-configured views
3. **Use "Explore"** for custom LogQL queries

## 🎯 Benefits of Enhanced System

- ✅ **Production-ready observability**
- ✅ **Low resource usage** (~500MB-1GB RAM)
- ✅ **Easy to scale** (add more Promtail instances)
- ✅ **Beautiful dashboards** out of the box
- ✅ **Security & Performance** middleware
- ✅ **Request tracing** across services