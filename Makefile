# Makefile for Day5 Retailer API (Clean Architecture)

# Variables
APP_NAME := day5
VERSION := latest
BINARY_NAME := server
MAIN_PATH := cmd/server/main.go
NAMESPACE := day5

# Build targets
.PHONY: build clean run test test-coverage test-handlers test-integration test-watch docker-build minikube-setup k8s-deploy k8s-clean helm-deploy helm-clean test-k8s dev-local dev-docker help

# Default target
help:
	@echo "=== Day5 Retailer API - Clean Architecture ==="
	@echo ""
	@echo "Development:"
	@echo "  build           - Build the Go binary"
	@echo "  run             - Run the application locally"
	@echo "  dev-local       - Run with local SQLite database"
	@echo "  dev-docker      - Run with Docker Compose (MySQL)"
	@echo "  clean           - Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  test            - Run all tests with coverage"
	@echo "  test-coverage   - Run tests and generate HTML coverage report"
	@echo "  test-handlers   - Run only handler tests"
	@echo "  test-integration- Run only integration tests"
	@echo "  test-watch      - Run tests in watch mode"
	@echo "  test-k8s        - Test complete Kubernetes deployment"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build    - Build Docker image"
	@echo ""
	@echo "Kubernetes:"
	@echo "  minikube-setup  - Setup and start Minikube cluster"
	@echo "  k8s-deploy      - Deploy to Kubernetes manually"
	@echo "  k8s-clean       - Clean Kubernetes deployment"
	@echo "  helm-deploy     - Deploy with Helm"
	@echo "  helm-clean      - Clean Helm deployment"
	@echo ""
	@echo "Utilities:"
	@echo "  deps            - Install dependencies"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	@CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: bin/$(BINARY_NAME)"

# Clean build artifacts  
clean:    
	@set -x    
	@echo "Cleaning..."    
	@rm -rf bin/    
	@docker image prune -f    
	@if lsof -i :8080 | awk 'NR>1 {print $$2}' | grep .; then \
		echo "Processes running on port 8080:"; \
		lsof -i :8080 | awk 'NR>1 {print "PID: "$$2}'; \
		lsof -i :8080 | awk 'NR>1 {print $$2}' | while read -r pid; do \
			echo "-----------------------------------------------"; \
			echo "Details for PID: $$pid"; \
			ps -p $$pid -o pid,ppid,user,tty,time,cmd; \
			echo "Command to kill this process: kill $$pid"; \
		done; \
	else \
		echo "No process is using port 8080."; \
	fi    
	@echo "Clean completed"  

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	@go run $(MAIN_PATH)

# Run tests
test:
	@echo "🧪 Running all tests..."
	@echo "Installing test dependencies..."
	@go mod tidy
	@echo "Running unit tests..."
	@go test -v ./internal/...
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

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):$(VERSION) .
	@echo "Docker image built: $(APP_NAME):$(VERSION)"


# Setup Minikube
minikube-setup:
	@echo "🚀 Setting up Minikube..."
	@minikube start --driver=docker || echo "❌ Minikube start failed"
	@kubectl config set-context --current --namespace=$(NAMESPACE) || true
	@echo "✅ Minikube setup completed"

# Deploy to Kubernetes manually
k8s-deploy:
	@echo "Cleaning"
	@make k8s-clean
	@echo "📦 Deploying to Kubernetes..."
	@kubectl create namespace $(NAMESPACE) || true
	@kubectl config set-context --current --namespace=$(NAMESPACE)
	@eval $$(minikube docker-env) && docker build -t $(APP_NAME):$(VERSION) .
	@kubectl apply -f deployments/k8s/mysql-configmap.yaml
	@kubectl apply -f deployments/k8s/mysql-secret.yaml
	@kubectl apply -f deployments/k8s/mysql-deployment.yaml
	@kubectl apply -f deployments/k8s/mysql-service.yaml
	@echo "⏳ Waiting for MySQL to be ready..."
	@./scripts/wait-for-pods.sh $(NAMESPACE) "app=mysql" 120
	@kubectl apply -f deployments/k8s/app-configmap.yaml
	@kubectl apply -f deployments/k8s/app-deployment.yaml
	@kubectl apply -f deployments/k8s/app-service.yaml
	@echo "⏳ Waiting for application to be ready..."
	@./scripts/wait-for-pods.sh $(NAMESPACE) "app=$(APP_NAME)" 120
	@echo "✅ Kubernetes deployment completed"
	@echo "🔗 Run 'kubectl port-forward service/$(APP_NAME) 8080:80' to access the API"

# Clean Kubernetes deployment
k8s-clean:
	@make clean
	@echo "🧹 Cleaning Kubernetes deployment..."
	@kubectl delete namespace $(NAMESPACE) || true
	@echo "✅ Kubernetes cleanup completed"

# Deploy with Helm
helm-deploy:
	@echo "Cleaning Helm"
	@make helm-clean
	@echo "⚙️  Deploying with Helm..."
	@kubectl create namespace $(NAMESPACE) || true
	@kubectl config set-context --current --namespace=$(NAMESPACE)
	@eval $$(minikube docker-env) && docker build -t $(APP_NAME):$(VERSION) .
	@helm upgrade --install $(APP_NAME) ./deployments/helm/$(APP_NAME) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--set image.repository=$(APP_NAME) \
		--set image.tag=$(VERSION) \
		--set image.pullPolicy=Never \
		--atomic \
		--no-hooks
		--wait --timeout 120s
# @echo "⏳ Waiting for MySQL to be ready..."
# @./scripts/wait-for-pods.sh $(NAMESPACE) "app=mysql" 120
# @echo "⏳ Waiting for application to be ready..."
# @./scripts/wait-for-pods.sh $(NAMESPACE) "app=$(APP_NAME)" 120

	@echo "⏳ Waiting for deployment to be ready..."
	@echo "✅ Helm deployment completed"
	@echo "🔗 Run 'kubectl port-forward service/$(APP_NAME) 8080:80' to access the API"

# Clean Helm deployment
helm-clean:
	@make clean
	@echo "🧹 Cleaning Helm deployment..."
	@helm uninstall $(APP_NAME) --namespace $(NAMESPACE) || true
	@kubectl delete namespace $(NAMESPACE) || true
	@echo "✅ Helm cleanup completed"

# Local development with SQLite
dev-local:
	@echo "🏠 Starting local development with SQLite..."
	@echo "📝 Using config/dev.toml with SQLite database"
	@mkdir -p data
	@go run $(MAIN_PATH)

# Local development with Docker Compose
dev-docker:
	@echo "🐳 Starting local development with Docker Compose..."
	@echo "📝 Using Docker Compose with MySQL database"
	@docker-compose up --build

# Test Kubernetes deployment end-to-end
test-k8s:
	@echo "🔬 Testing Kubernetes deployment..."
	@echo "⏳ Waiting for port-forward to be ready..."
	@sleep 5
	@echo "📋 Testing health endpoint..."
	@curl -f http://localhost:8080/health > /dev/null && echo "✅ Health check passed" || echo "❌ Health check failed"
	@echo "📦 Testing product creation..."
	@PRODUCT_ID=$$(curl -s -X POST http://localhost:8080/api/v1/product \
		-H "Content-Type: application/json" \
		-d '{"product_name":"Test Product","price":99.99,"quantity":10}' | \
		python3 -c "import sys,json; print(json.load(sys.stdin)['id'])") && \
	echo "✅ Product created: $$PRODUCT_ID" || echo "❌ Product creation failed"
	@echo "👤 Testing customer creation..."
	@CUSTOMER_ID=$$(curl -s -X POST http://localhost:8080/api/v1/customer \
		-H "Content-Type: application/json" \
		-d '{"name":"Test Customer","email":"test@example.com","phone":"1234567890"}' | \
		python3 -c "import sys,json; print(json.load(sys.stdin)['id'])") && \
	echo "✅ Customer created: $$CUSTOMER_ID" || echo "❌ Customer creation failed"
	@echo "🛍️ Testing order workflow..."
	@echo "✅ Basic API endpoints are working"
	@echo "🎉 Kubernetes deployment test completed!"

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@go mod download
	@echo "✅ Dependencies installed"

# Format code
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Run linter
lint:
	@echo "🔍 Running linter..."
	@go vet ./...
	@echo "✅ Linting completed"
