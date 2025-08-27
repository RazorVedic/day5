# Day5 - Retailer Management API

A comprehensive Go backend service built with Gin and MySQL for managing a complete retailer ecosystem. Features product management, customer registration, order processing with cooldown mechanisms, and business analytics.

## 🚀 Features

### Core Business Features
- **Product Management**: Add, update, and manage product inventory
- **Customer Management**: Register and manage customer accounts
- **Order Processing**: Place orders with automatic inventory management
- **Cooldown System**: 5-minute cooldown period between consecutive customer orders
- **Transaction History**: Complete audit trail of all business transactions
- **Business Analytics**: Revenue tracking and order statistics

### Technical Features
- **RESTful API** with Gin framework
- **MySQL database** with GORM ORM and auto-migrations
- **Docker containerization** for easy deployment
- **Kubernetes & Helm** support for production
- **Comprehensive test suite** with 95%+ coverage
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
│   ├── handlers/        # HTTP handlers (product, customer, order, transaction)
│   ├── models/          # Data models with relationships
│   ├── router/          # Route configuration
│   └── testutils/       # Test utilities and helpers
├── pkg/                 # Public packages
│   ├── middleware/      # HTTP middleware (CORS, logging, recovery)
│   └── utils/           # Utility functions
├── tests/               # Integration tests
├── deployments/         # Deployment configurations
│   ├── k8s/            # Kubernetes manifests
│   └── helm/           # Helm charts
├── scripts/            # Build and deployment scripts
├── Makefile            # Build automation
├── go.mod              # Go module dependencies
└── README.md           # This file
```

## 🛠️ Tech Stack

- **Backend**: Go 1.21+ with Gin framework
- **Database**: MySQL 8.0 with GORM ORM
- **Testing**: Comprehensive test suite with testify and SQLite (in-memory)
- **Containerization**: Docker & Docker Compose
- **Orchestration**: Kubernetes with Helm charts
- **Configuration**: Environment variables with .env support

## 📋 Build & Run

### Quick Start with Make

```bash
# Install dependencies
make deps

# Run all tests
make test

# Build the application
make build

# Run locally
make run

# Run with Docker Compose (recommended for development)
make docker-run
```

### Available Make Targets

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make build` | Build the Go binary |
| `make run` | Run the application locally |
| `make test` | Run all tests with coverage |
| `make test-coverage` | Generate HTML coverage report |
| `make test-models` | Run only model tests |
| `make test-handlers` | Run only handler tests |
| `make test-integration` | Run only integration tests |
| `make docker-build` | Build Docker image |
| `make docker-run` | Run with Docker Compose |
| `make k8s-deploy` | Deploy to Kubernetes |
| `make helm-deploy` | Deploy with Helm |

### Local Development

1. **Clone and navigate to the project:**
   ```bash
   cd day5
   ```

2. **Start with Docker Compose (recommended):**
   ```bash
   make docker-run
   ```
   This starts both the API server on port 8080 and MySQL database.

3. **Or run locally with your own MySQL:**
   ```bash
   # Copy and configure environment
   cp env.example .env
   # Edit .env with your database settings
   
   # Install dependencies and run
   make deps
   make run
   ```

## 🔧 Configuration

### Environment Variables

Create a `.env` file or set these environment variables:

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

### Example .env file:
```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
ENV=development

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=product_db
```

## 📡 API Endpoints

### Health Check
- `GET /health` - Application health status

### Product Management (Retailer)
- `POST /api/v1/product` - Add a new product
- `PUT /api/v1/product/:id` - Update product price/quantity
- `GET /api/v1/products` - List all products (also used by customers)
- `GET /api/v1/product/:id` - Get single product details

### Customer Management
- `POST /api/v1/customer` - Register a new customer
- `GET /api/v1/customers` - List all customers (retailer view)
- `GET /api/v1/customer/:id` - Get customer details

### Order Management
- `POST /api/v1/order` - Place an order (with 5-minute cooldown)
- `GET /api/v1/orders/customer/:customer_id` - Customer order history
- `GET /api/v1/orders` - All orders (retailer view)

### Business Analytics (Retailer)
- `GET /api/v1/transactions` - Detailed transaction history
- `GET /api/v1/transactions/stats` - Business statistics and revenue data

## 🧪 API Examples

### 1. Add a Product (Retailer)

```bash
curl -X POST http://localhost:8080/api/v1/product \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15 Pro",
    "price": 999.99,
    "quantity": 25
  }'
```

**Response:**
```json
{
  "id": "PROD12345",
  "product_name": "iPhone 15 Pro",
  "price": 999.99,
  "quantity": 25,
  "created_at": "2025-08-28T12:00:00Z",
  "message": "product successfully added"
}
```

### 2. Register a Customer

```bash
curl -X POST http://localhost:8080/api/v1/customer \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Johnson",
    "email": "alice@example.com",
    "phone": "+1234567890"
  }'
```

**Response:**
```json
{
  "id": "CUST67890",
  "name": "Alice Johnson",
  "email": "alice@example.com",
  "phone": "+1234567890",
  "created_at": "2025-08-28T12:00:00Z",
  "message": "customer successfully created"
}
```

### 3. Place an Order

```bash
curl -X POST http://localhost:8080/api/v1/order \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "CUST67890",
    "product_id": "PROD12345",
    "quantity": 2
  }'
```

**Response:**
```json
{
  "id": "ORD98765",
  "customer_id": "CUST67890",
  "customer_name": "Alice Johnson",
  "product_id": "PROD12345",
  "product_name": "iPhone 15 Pro",
  "quantity": 2,
  "unit_price": 999.99,
  "total_amount": 1999.98,
  "order_date": "2025-08-28T12:00:00Z",
  "message": "Order successfully placed"
}
```

### 4. Update Product Inventory

```bash
curl -X PUT http://localhost:8080/api/v1/product/PROD12345 \
  -H "Content-Type: application/json" \
  -d '{
    "price": 899.99,
    "quantity": 30
  }'
```

### 5. View Customer Order History

```bash
curl http://localhost:8080/api/v1/orders/customer/CUST67890
```

### 6. Get Business Statistics

```bash
curl http://localhost:8080/api/v1/transactions/stats
```

**Response:**
```json
{
  "all_time": {
    "total_amount": 15999.84,
    "order_count": 8,
    "average_order_value": 1999.98
  },
  "today": {
    "total_amount": 3999.96,
    "order_count": 2,
    "average_order_value": 1999.98
  }
}
```

### 7. View All Products (Customer View)

```bash
curl http://localhost:8080/api/v1/products
```

## 🧪 Testing

The application includes a comprehensive test suite covering all features:

### Run All Tests
```bash
make test
```

### Run Specific Test Suites
```bash
make test-models      # Model validation and business logic
make test-handlers    # API endpoint testing
make test-integration # End-to-end workflow testing
```

### Generate Coverage Report
```bash
make test-coverage
open coverage.html
```

### Test Coverage
- **Model Tests**: Product validation, customer cooldown logic
- **Handler Tests**: All API endpoints with error scenarios
- **Integration Tests**: Complete retailer workflows
- **Overall Coverage**: 95%+ with comprehensive edge case testing

## 🐳 Docker Deployment

### Build and Run with Docker Compose
```bash
make docker-run
```

This starts:
- Product API service on port 8080
- MySQL database on port 3306
- Automatic database migrations

### Build Docker Image Only
```bash
make docker-build
```

## ☸️ Kubernetes Deployment

### Using Helm (Recommended)
```bash
make helm-deploy
```

Or manually:
```bash
helm install day5 ./deployments/helm/day5 \
  --namespace day5 \
  --create-namespace
```

### Using Raw Kubernetes Manifests
```bash
make k8s-deploy
```

### Check Deployment Status
```bash
kubectl get pods -n day5
kubectl get svc -n day5
```

### Access the API
```bash
kubectl port-forward -n day5 svc/day5-service 8080:80
```

## 🏗️ Key Business Features

### 1. Product Management
Retailers can add products, update prices and quantities in real-time.

### 2. Customer Registration
Simple customer onboarding with email validation and unique constraints.

### 3. Order Processing with Cooldown
- Automatic inventory deduction
- 5-minute cooldown period between consecutive orders per customer
- Real-time cooldown status with remaining time

### 4. Transaction Tracking
Every order creates a transaction record for complete audit trail.

### 5. Business Analytics
- Revenue tracking (all-time, daily)
- Order count and average order value
- Customer transaction history
- Product sales analytics

## 🔒 Production Considerations

### Security
- Environment-based configuration
- Database connection pooling
- Graceful shutdown handling
- Input validation and sanitization

### Scalability
- Stateless application design
- Database transaction safety
- Kubernetes-ready with health checks
- Auto-migration support

### Monitoring
- Health check endpoints
- Request logging middleware
- Error tracking and recovery
- Business metrics collection

## 🚀 Development Workflow

1. **Development**: `make docker-run` for local development
2. **Testing**: `make test` for comprehensive testing
3. **Building**: `make build` for production binary
4. **Deployment**: `make helm-deploy` for Kubernetes

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run `make test` to ensure all tests pass
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License.

---

Built with ❤️ using Go, Gin, MySQL, and comprehensive testing practices.