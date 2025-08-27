# Day5 - Go Backend API

A clean and scalable Go backend service built with Gin and MySQL for managing products. Simple REST API with Docker and Kubernetes deployment support.

## 🚀 Features

- **RESTful API** with Gin framework
- **MySQL database** with GORM ORM
- **Docker containerization** for easy local development
- **Kubernetes deployment** for production
- **Helm charts** for simplified K8s deployment
- **Environment-based configuration**
- **Health check endpoints**
- **Graceful shutdown**

## 📁 Project Structure

```
day5/
├── cmd/server/           # Application entry point
│   └── main.go
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and migrations
│   ├── handlers/        # HTTP handlers
│   └── models/          # Data models
├── pkg/                 # Public packages
│   ├── middleware/      # HTTP middleware
│   └── utils/           # Utility functions
├── deployments/         # Deployment configurations
│   ├── k8s/            # Kubernetes manifests
│   └── helm/           # Helm charts
├── scripts/            # Build and deployment scripts
├── go.mod              # Go module file
└── README.md           # This file
```

## 🛠️ Tech Stack

- **Backend**: Go 1.21+ with Gin framework
- **Database**: MySQL 8.0 with GORM ORM
- **Containerization**: Docker & Docker Compose
- **Orchestration**: Kubernetes with Helm
- **Configuration**: Environment variables with .env support

## 📋 API Endpoints

### POST /api/v1/product

Create a new product.

**Request:**
```json
{
  "product_name": "bottle",
  "price": 50,
  "quantity": 40
}
```

**Response:**
```json
{
  "id": "PROD12345",
  "product_name": "bottle",
  "price": 50,
  "quantity": 40,
  "message": "product successfully added"
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "service": "day5",
  "version": "1.0.0"
}
```

### Additional Endpoints (for testing)

- `GET /api/v1/products` - Get all products
- `GET /api/v1/product/:id` - Get product by ID

## 🏃‍♂️ Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- MySQL 8.0 (for local development without Docker)

### Local Development with Docker Compose

1. **Clone and navigate to the project:**
   ```bash
   cd /path/to/your/project
   ```

2. **Start the services:**
   ```bash
   docker-compose up --build
   ```

3. **Test the API:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/product \
     -H "Content-Type: application/json" \
     -d '{"product_name": "bottle", "price": 50, "quantity": 40}'
   ```

### Local Development without Docker

1. **Set up MySQL database:**
   ```bash
   # Create database
   mysql -u root -p -e "CREATE DATABASE product_db;"
   ```

2. **Copy environment configuration:**
   ```bash
   cp env.example .env
   # Edit .env with your database credentials
   ```

3. **Install dependencies:**
   ```bash
   go mod tidy
   ```

4. **Run the application:**
   ```bash
   go run cmd/server/main.go
   ```

## 🐳 Docker Deployment

### Build Docker Image

```bash
docker build -t day5:latest .
```

### Run with Docker Compose

```bash
docker-compose up -d
```

This will start:
- Product API service on port 8080
- MySQL database on port 3306

## ☸️ Kubernetes Deployment

### Using Raw Kubernetes Manifests

1. **Apply all manifests:**
   ```bash
   kubectl apply -f deployments/k8s/
   ```

2. **Check deployment status:**
   ```bash
   kubectl get pods -n day5
   kubectl get svc -n day5
   ```

3. **Access the API:**
   ```bash
   kubectl port-forward -n day5 svc/day5-service 8080:80
   ```

### Using Helm Charts

1. **Deploy with Helm:**
   ```bash
   ./scripts/deploy.sh
   ```

   Or manually:
   ```bash
   helm install day5 ./deployments/helm/day5 \
     --namespace day5 \
     --create-namespace
   ```

2. **Upgrade deployment:**
   ```bash
   helm upgrade day5 ./deployments/helm/day5 \
     --namespace day5
   ```

3. **Uninstall:**
   ```bash
   helm uninstall day5 --namespace day5
   ```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server bind address | `0.0.0.0` |
| `SERVER_PORT` | Server port | `8080` |
| `ENV` | Environment (development/production) | `development` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `3306` |
| `DB_USER` | Database username | `root` |
| `DB_PASSWORD` | Database password | `password` |
| `DB_NAME` | Database name | `product_db` |

### Production Configuration

For production deployment, make sure to:

1. **Update database credentials** in Kubernetes secrets
2. **Configure ingress** with proper domain and TLS
3. **Set resource limits** in Kubernetes manifests
4. **Enable autoscaling** if needed
5. **Configure monitoring and logging**

## 🧪 Testing

### Test Product Creation

```bash
curl -X POST http://localhost:8080/api/v1/product \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "test bottle",
    "price": 25.99,
    "quantity": 100
  }'
```

### Test Health Check

```bash
curl http://localhost:8080/health
```

### Test List Products

```bash
curl http://localhost:8080/api/v1/products
```

## 🚀 Deployment Scripts

### Build Script

```bash
./scripts/build.sh
```

Options:
- `VERSION=v1.0.0 ./scripts/build.sh` - Build with specific version
- `REGISTRY=your-registry.com PUSH=true ./scripts/build.sh` - Build and push to registry

### Deploy Script

```bash
./scripts/deploy.sh
```

Options:
- `NAMESPACE=prod ./scripts/deploy.sh` - Deploy to specific namespace
- `RELEASE_NAME=prod-api ./scripts/deploy.sh` - Use custom release name

## 📊 Monitoring and Health Checks

The application includes:

- **Health check endpoint** at `/health`
- **Kubernetes readiness and liveness probes**
- **Graceful shutdown** with 30-second timeout
- **Request logging** with response times
- **Error handling** with proper HTTP status codes

## 🔒 Security Considerations

For production deployments:

1. **Use proper secrets management** for database credentials
2. **Configure network policies** to restrict pod communication
3. **Enable TLS/HTTPS** for ingress
4. **Set proper resource limits** to prevent resource exhaustion
5. **Use non-root containers** (already configured in Dockerfile)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License.

## 🆘 Troubleshooting

### Common Issues

1. **Database connection fails:**
   - Check database credentials
   - Ensure database is running
   - Verify network connectivity

2. **Docker build fails:**
   - Check Go version compatibility
   - Ensure all dependencies are available
   - Verify Dockerfile syntax

3. **Kubernetes deployment fails:**
   - Check resource quotas
   - Verify RBAC permissions
   - Check node resources

### Getting Help

- Check logs: `kubectl logs -n day5 deployment/day5`
- Describe resources: `kubectl describe pod -n day5 <pod-name>`
- Check events: `kubectl get events -n day5`

---

Built with ❤️ using Go, Gin, and MySQL.
