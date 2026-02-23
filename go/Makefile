# DSView Backend Makefile
# ======================

APP_NAME=dsview-backend
BINARY_NAME=bin/$(APP_NAME)
PORT=8080

.PHONY: help
help: ## Show this help message
	@echo "DSView Backend - Available Commands"
	@echo "=================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'

# ===================
# Development Commands
# ===================

.PHONY: run
run: deps swagger ## Run the application in development mode
	@echo "Starting DSView Backend in development mode..."
	@echo "Server will be available at: http://localhost:$(PORT)"
	@echo "API Documentation: http://localhost:$(PORT)/docs/"
	@if [ -f .env.development ]; then \
		echo "Loading environment variables from .env.development file..."; \
		export $$(cat .env.development | grep -v '^#' | xargs) && go run cmd/server/main.go; \
	else \
		echo "No .env.development file found, using system environment variables..."; \
		go run cmd/server/main.go;  \
	fi


.PHONY: run-prod
run-prod: ## Run with production environment
	@echo "Running with production environment..."
	@if [ -f .env.production ]; then \
		echo "Loading environment variables from .env.production file..."; \
		export $$(cat .env.production | grep -v '^#' | xargs) && APP_ENV=production go run cmd/server/main.go; \
	else \
		echo "No .env.production file found, using system environment variables..."; \
		APP_ENV=production go run cmd/server/main.go; \
	fi

# ===================
# Build Commands
# ===================

.PHONY: build
build: deps swagger ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(BINARY_NAME) cmd/server/main.go
	@echo "Build completed: $(BINARY_NAME)"

.PHONY: build-dev
build-dev: ## Build for development
	@echo "Building for development..."
	@mkdir -p bin
	APP_ENV=development go build -ldflags="-s -w" -o $(BINARY_NAME)-dev cmd/server/main.go
	@echo "Development build completed: $(BINARY_NAME)-dev"

.PHONY: build-prod
build-prod: ## Build for production
	@echo "Building for production..."
	@mkdir -p bin
	APP_ENV=production go build -ldflags="-s -w" -o $(BINARY_NAME)-prod cmd/server/main.go
	@echo "Production build completed: $(BINARY_NAME)-prod"

.PHONY: start
start: build ## Build and start the application
	@echo "Starting $(APP_NAME)..."
	@echo "Server will be available at: http://localhost:$(PORT)"
	@if [ -f .env.development ]; then \
		echo "Loading environment variables from .env.development file..."; \
		export $$(cat .env.development | grep -v '^#' | xargs) && ./$(BINARY_NAME); \
	else \
		echo "No .env.development file found, using system environment variables..."; \
		./$(BINARY_NAME); \
	fi

.PHONY: start-dev
start-dev: build-dev ## Build and start development version
	@echo "Starting $(APP_NAME) development version..."
	@echo "Server will be available at: http://localhost:$(PORT)"
	@if [ -f .env.development ]; then \
		echo "Loading environment variables from .env.development file..."; \
		export $$(cat .env.development | grep -v '^#' | xargs) && ./$(BINARY_NAME)-dev; \
	else \
		echo "No .env.development file found, using system environment variables..."; \
		./$(BINARY_NAME)-dev; \
	fi

.PHONY: start-prod
start-prod: build-prod ## Build and start production version
	@echo "Starting $(APP_NAME) production version..."
	@echo "Server will be available at: http://localhost:$(PORT)"
	@if [ -f .env.production ]; then \
		echo "Loading environment variables from .env.production file..."; \
		export $$(cat .env.production | grep -v '^#' | xargs) && ./$(BINARY_NAME)-prod; \
	else \
		echo "No .env.production file found, using system environment variables..."; \
		./$(BINARY_NAME)-prod; \
	fi

# ===================
# Dependencies
# ===================

.PHONY: deps
deps: ## Install and tidy dependencies
	@echo "Installing dependencies..."
	go mod tidy
	go mod download
	@echo "Dependencies installed"

# ===================
# Documentation
# ===================

.PHONY: swagger
swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go -o docs; \
		echo "Swagger documentation generated"; \
	else \
		echo "swag not found. Install it with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# ===================
# Code Quality
# ===================

.PHONY: format
format: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Code formatted"

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...
	@echo "Go vet completed"

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test -v ./...
	@echo "Tests completed"

# ===================
# Installation
# ===================

.PHONY: install
install: ## Install the application
	@echo "Installing $(APP_NAME)..."
	go install ./cmd/server
	@echo "Installation completed"

# ===================
# Cleanup
# ===================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	go clean
	rm -rf bin/
	@echo "Clean completed"

# ===================
# Environment Setup
# ===================

.PHONY: env-setup
env-setup: ## Setup environment files from examples
	@echo "Setting up environment files..."
	@if [ ! -f .env.development ]; then \
		cp env.example .env.development; \
		echo "Created .env.development file from env.example"; \
	else \
		echo ".env.development file already exists"; \
	fi
	@if [ ! -f .env.production ]; then \
		cp env.example .env.production; \
		echo "Created .env.production file from env.example"; \
		echo "Please update .env.production with production values"; \
	else \
		echo ".env.production file already exists"; \
	fi
	@echo "Environment setup completed"

.PHONY: setup
setup: deps env-setup swagger ## Complete setup for development
	@echo "Setup completed! You can now run 'make run-dev' to start the application"