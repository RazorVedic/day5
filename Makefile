# Makefile for Product API

# Variables
APP_NAME := day5
VERSION := latest
BINARY_NAME := server
MAIN_PATH := cmd/server/main.go

# Build targets
.PHONY: build clean run test docker-build docker-run help

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the Go binary"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  k8s-deploy    - Deploy to Kubernetes"
	@echo "  helm-deploy   - Deploy with Helm"

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: bin/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@docker image prune -f
	@echo "Clean completed"

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	@go run $(MAIN_PATH)

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):$(VERSION) .
	@echo "Docker image built: $(APP_NAME):$(VERSION)"

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	@docker-compose up --build

# Deploy to Kubernetes (requires proper cluster access)
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f deployments/k8s/ || echo "❌ K8s deployment failed - check cluster permissions"

# Deploy with Helm (requires proper cluster access)  
helm-deploy:
	@echo "Deploying with Helm..."
	@./scripts/deploy.sh || echo "❌ Helm deployment failed - check cluster permissions"

# Local development (recommended)
dev:
	@echo "Starting local development environment..."
	@docker-compose up --build

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download
