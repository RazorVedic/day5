# Retailer Service API Documentation

This service provides a complete solution for a hypothetical retailer with product management, customer management, order processing, and business analytics.

## Base URL
```
http://localhost:8080/api/v1
```

## Features Implemented

✅ **Product Management** - Add products, update price/quantity  
✅ **Customer Management** - Register customers  
✅ **Order Processing** - Place orders with inventory management  
✅ **Order History** - Customer and retailer order views  
✅ **Transaction History** - Detailed business analytics  
✅ **Cooldown Mechanism** - 5-minute cooldown between customer orders  
✅ **Business Dashboard** - Sales statistics and top products  

---

## 🛍️ Product Management (Retailer)

### Add a Product
```http
POST /api/v1/product
Content-Type: application/json

{
  "product_name": "iPhone 15",
  "price": 799.99,
  "quantity": 50
}
```

**Response:**
```json
{
  "id": "PROD12345",
  "product_name": "iPhone 15",
  "price": 799.99,
  "quantity": 50,
  "message": "product successfully added"
}
```

### Update Product Price/Quantity
```http
PUT /api/v1/product/PROD12345
Content-Type: application/json

{
  "price": 749.99,
  "quantity": 45
}
```

### View All Products
```http
GET /api/v1/products
```

**Response:**
```json
{
  "products": [
    {
      "id": "PROD12345",
      "product_name": "iPhone 15",
      "price": 749.99,
      "quantity": 45,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ],
  "count": 1
}
```

---

## 👥 Customer Management

### Register a Customer
```http
POST /api/v1/customer
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+1234567890"
}
```

**Response:**
```json
{
  "id": "CUST12345",
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "created_at": "2024-01-15T10:30:00Z",
  "message": "customer successfully created"
}
```

### View All Customers (Retailer)
```http
GET /api/v1/customers
```

---

## 🛒 Order Management

### Place an Order (Customer)
```http
POST /api/v1/order
Content-Type: application/json

{
  "customer_id": "CUST12345",
  "product_id": "PROD12345",
  "quantity": 2
}
```

**Success Response:**
```json
{
  "id": "ORD12345",
  "customer_id": "CUST12345",
  "customer_name": "John Doe",
  "product_id": "PROD12345",
  "product_name": "iPhone 15",
  "quantity": 2,
  "unit_price": 749.99,
  "total_price": 1499.98,
  "status": "completed",
  "created_at": "2024-01-15T12:00:00Z",
  "message": "order successfully placed"
}
```

**Cooldown Error Response:**
```json
{
  "error": "Customer is in cooldown period",
  "cooldown_remaining_seconds": 180,
  "cooldown_remaining_minutes": "3.0"
}
```

### View Customer Order History
```http
GET /api/v1/orders/customer/CUST12345
```

### View All Orders (Retailer)
```http
GET /api/v1/orders
```

---

## 📊 Transaction History & Analytics (Retailer)

### Detailed Transaction History
```http
GET /api/v1/transactions
```

**Query Parameters:**
- `limit` - Number of transactions to return (default: 100)
- `offset` - Pagination offset (default: 0)
- `start_date` - Filter from date (format: YYYY-MM-DD)
- `end_date` - Filter to date (format: YYYY-MM-DD)
- `type` - Filter by transaction type (order, refund, adjustment)
- `customer_id` - Filter by customer
- `product_id` - Filter by product

**Example:**
```http
GET /api/v1/transactions?limit=50&start_date=2024-01-01&end_date=2024-01-31&type=order
```

**Response:**
```json
{
  "transactions": [
    {
      "id": "TXN12345",
      "order_id": "ORD12345",
      "customer_id": "CUST12345",
      "customer_name": "John Doe",
      "product_id": "PROD12345",
      "product_name": "iPhone 15",
      "type": "order",
      "amount": 1499.98,
      "quantity": 2,
      "description": "Order for iPhone 15 (x2)",
      "created_at": "2024-01-15T12:00:00Z"
    }
  ],
  "count": 1,
  "total_amount": 1499.98
}
```

### Business Statistics Dashboard
```http
GET /api/v1/transactions/stats
```

**Response:**
```json
{
  "today": {
    "total_amount": 1499.98,
    "order_count": 1
  },
  "week": {
    "total_amount": 5299.92,
    "order_count": 4
  },
  "month": {
    "total_amount": 15749.85,
    "order_count": 12
  },
  "all_time": {
    "total_amount": 45299.55,
    "order_count": 35
  },
  "top_products": [
    {
      "product_id": "PROD12345",
      "product_name": "iPhone 15",
      "total_sold": 25,
      "revenue": 18749.75
    }
  ],
  "stats_date": "2024-01-15 15:30:45"
}
```

---

## 🚀 Deployment

The service is ready to deploy with your existing methods:

### Docker Compose (Recommended for development)
```bash
make docker-run
```

### Kubernetes
```bash
make k8s-deploy
```

### Helm (Production)
```bash
make helm-deploy
```

---

## 🔒 Cooldown Mechanism

- **5-minute cooldown** between orders per customer
- Prevents spam orders and manages inventory better
- Returns remaining cooldown time in error response
- Tracked per customer ID in `customer_cooldowns` table

---

## 🗄️ Database Schema

The service automatically creates these tables:

1. **products** - Product catalog with inventory
2. **customers** - Customer information  
3. **orders** - Order records with relationships
4. **transactions** - Complete business transaction log
5. **customer_cooldowns** - Cooldown tracking per customer

All tables use auto-generated IDs with prefixes:
- Products: `PROD12345`
- Customers: `CUST12345`  
- Orders: `ORD12345`
- Transactions: `TXN12345`

---

## 🎯 Business Use Cases Covered

✅ **a)** Retailer can add products and update price/quantity  
✅ **b)** Customers can view all available products via REST API  
✅ **c)** Customers can place orders and view order history  
✅ **d)** Retailer can view detailed transaction history of all business  
✅ **e)** 5-minute cooldown period between customer orders  

The implementation is production-ready and works with your existing Docker and Kubernetes deployment infrastructure.
