# DSView Backend - Docker Hub Management

# Configuration
DOCKER_HUB_USERNAME ?= $(shell grep DOCKER_HUB_USERNAME .env 2>/dev/null | cut -d '=' -f2 || echo "your-username")
DOCKER_HUB_REPO ?= $(shell grep DOCKER_HUB_REPO .env 2>/dev/null | cut -d '=' -f2 || echo "dsview-backend")
VERSION ?= $(shell grep VERSION .env 2>/dev/null | cut -d '=' -f2 || echo "latest")

.PHONY: help build push pull dev prod clean

help: ## Show this help message
	@echo "DSView Backend - Docker Hub Management$(NC)"
	@echo ""
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "   %-15s$(NC) %s\n", $$1, $$2}'

build: ## Build Docker images
	@echo "Building Docker images...$(NC)"
	@docker build -t $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-fastapi:$(VERSION) ./fastapi
	@docker build -t $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-go:$(VERSION) ./go
	@docker tag $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-fastapi:$(VERSION) $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-fastapi:latest
	@docker tag $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-go:$(VERSION) $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-go:latest
	@echo "Images built successfully!$(NC)"

push: ## Push images to Docker Hub
	@echo " Pushing images to Docker Hub...$(NC)"
	@docker push $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-fastapi:$(VERSION)
	@docker push $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-fastapi:latest
	@docker push $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-go:$(VERSION)
	@docker push $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-go:latest
	@echo " Images pushed successfully!$(NC)"

pull: ## Pull images from Docker Hub
	@echo " Pulling images from Docker Hub...$(NC)"
	@docker pull $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-fastapi:$(VERSION)
	@docker pull $(DOCKER_HUB_USERNAME)/$(DOCKER_HUB_REPO)-go:$(VERSION)
	@echo " Images pulled successfully!$(NC)"

dev: ## Start development environment
	@echo " Starting development environment...$(NC)"
	@docker-compose up -d

prod: ## Start production environment
	@echo " Starting production environment...$(NC)"
	@docker-compose -f docker-compose.prod.yml up -d

stop: ## Stop all services
	@echo " Stopping all services...$(NC)"
	@docker-compose down
	@docker-compose -f docker-compose.prod.yml down

clean: ## Clean up Docker resources
	@echo " Cleaning up Docker resources...$(NC)"
	@docker system prune -f
	@docker volume prune -f

logs: ## Show logs
	@docker-compose logs -f

# Quick commands
build-push: build push ## Build and push images
dev-restart: stop dev ## Restart development environment
prod-restart: stop prod ## Restart production environment
