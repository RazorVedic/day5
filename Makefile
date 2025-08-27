# Makefile for Product API

# Variables
APP_NAME := day5
VERSION := latest
BINARY_NAME := server
MAIN_PATH := cmd/server/main.go

# Build targets
.PHONY: build clean run test test-coverage test-models test-handlers test-integration test-watch docker-build docker-run help

# Default target
help:
	@echo "Available targets:"
	@echo "  build           - Build the Go binary"
	@echo "  clean           - Clean build artifacts"
	@echo "  run             - Run the application"
	@echo "  test            - Run all tests with coverage"
	@echo "  test-coverage   - Run tests and generate HTML coverage report"
	@echo "  test-models     - Run only model tests"
	@echo "  test-handlers   - Run only handler tests"
	@echo "  test-integration- Run only integration tests"
	@echo "  test-watch      - Run tests in watch mode"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run with Docker Compose"
	@echo "  k8s-deploy      - Deploy to Kubernetes"
	@echo "  helm-deploy     - Deploy with Helm"

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
	@echo "Installing test dependencies..."
	@go mod tidy
	@echo "Running unit tests..."
	@go test -v ./internal/models/...
	@echo "Running handler tests..."
	@go test -v ./internal/handlers/...
	@echo "Running integration tests..."
	@go test -v ./tests/...
	@echo "Running all tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@echo "✅ All tests completed successfully!"

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run specific test suites
test-models:
	@echo "Running model tests..."
	@go test -v ./internal/models/...

test-handlers:
	@echo "Running handler tests..."
	@go test -v ./internal/handlers/...

test-integration:
	@echo "Running integration tests..."
	@go test -v ./tests/...

# Run tests in watch mode (requires entr)
test-watch:
	@echo "Running tests in watch mode..."
	@find . -name "*.go" | entr -c make test

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
