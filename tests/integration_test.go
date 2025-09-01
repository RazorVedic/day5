package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"day5/internal/application/usecases"
	"day5/internal/config"
	"day5/internal/infrastructure/container"
	httpHandlers "day5/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupIntegrationTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// Create test configuration
	cfg := &config.AppConfig{
		Database: config.DatabaseSettings{
			Dialect: "sqlite",
			Name:    ":memory:",
		},
		App: config.AppSettings{
			Name: "day5-test",
		},
		Server: config.ServerSettings{
			Port: 8080,
		},
		Business: config.BusinessSettings{
			CooldownPeriodMinutes: 5,
			DefaultCurrency:       "USD",
		},
		Security: config.SecuritySettings{
			JWTSecret: "test-secret",
		},
	}

	// Create DI container and initialize with config
	diContainer := container.NewContainer()
	err := diContainer.Initialize(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize container: %v", err)
	}

	// Setup router
	router := httpHandlers.NewRouter(diContainer)
	appRouter := router.SetupRoutes()

	return appRouter, diContainer.GetDatabase().GetDB()
}

func TestCompleteRetailerWorkflow(t *testing.T) {
	appRouter, db := setupIntegrationTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	var productIDs []string

	// 1. Add products (Retailer)
	t.Run("Add Products", func(t *testing.T) {
		products := []usecases.CreateProductRequest{
			{ProductName: "iPhone 15 Pro", Price: 999.99, Quantity: 25},
			{ProductName: "MacBook Air M3", Price: 1199.99, Quantity: 10},
			{ProductName: "AirPods Pro", Price: 249.99, Quantity: 50},
		}

		for _, product := range products {
			jsonData, _ := json.Marshal(product)
			req, _ := http.NewRequest("POST", "/api/v1/product", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			appRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response httpHandlers.ProductResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.ID)
			productIDs = append(productIDs, response.ID)
		}
	})

	var customerIDs []string
	t.Run("Register Customers", func(t *testing.T) {
		customers := []usecases.CreateCustomerRequest{
			{Name: "Alice Johnson", Email: "alice@example.com", Phone: "+1234567890"},
			{Name: "Bob Smith", Email: "bob@example.com", Phone: "+1987654321"},
			{Name: "Charlie Brown", Email: "charlie@example.com", Phone: "+1555555555"},
		}

		for _, customer := range customers {
			jsonData, _ := json.Marshal(customer)
			req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			appRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response httpHandlers.CustomerResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.ID)
			customerIDs = append(customerIDs, response.ID)
		}
	})

	// 3. View all products (Customer)
	t.Run("View Products", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/products", nil)
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response httpHandlers.ProductListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 3, response.Count)
		assert.Len(t, response.Products, 3)
	})

	var orderIDs []string

	// 4. Place orders
	t.Run("Place Orders", func(t *testing.T) {
		orders := []usecases.PlaceOrderRequest{
			{CustomerID: customerIDs[0], ProductID: productIDs[0], Quantity: 1}, // Alice orders iPhone
			{CustomerID: customerIDs[1], ProductID: productIDs[2], Quantity: 2}, // Bob orders AirPods
		}

		for _, order := range orders {
			jsonData, _ := json.Marshal(order)
			req, _ := http.NewRequest("POST", "/api/v1/order", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			appRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response httpHandlers.OrderResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.ID)
			orderIDs = append(orderIDs, response.ID)
			assert.Greater(t, response.TotalAmount, 0.0)
		}
	})

	// 5. Test cooldown period
	t.Run("Test Cooldown", func(t *testing.T) {
		// Alice tries to place another order immediately (should fail)
		order := usecases.PlaceOrderRequest{
			CustomerID: customerIDs[0],
			ProductID:  productIDs[1], // Different product
			Quantity:   1,
		}

		jsonData, _ := json.Marshal(order)
		req, _ := http.NewRequest("POST", "/api/v1/order", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "cooldown_remaining_seconds")
	})

	// 6. View customer order history
	t.Run("Customer Order History", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/orders/customer/"+customerIDs[0], nil)
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response httpHandlers.OrderHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 1, response.Count) // Alice placed 1 order
		assert.Len(t, response.Orders, 1)
	})

	// 7. View all orders (Retailer)
	t.Run("All Orders", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/orders", nil)
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response httpHandlers.OrderHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, response.Count) // Total orders placed
		assert.Len(t, response.Orders, 2)
	})

	// 8. View business transactions (Retailer)
	t.Run("Business Transactions", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response httpHandlers.TransactionHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, response.Count) // Should have 2 transactions
		assert.Greater(t, response.TotalAmount, 0.0)
	})

	// 9. View business analytics
	t.Run("Business Analytics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/transactions/stats", nil)
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "total_revenue")
		assert.Contains(t, response, "order_count")
		assert.Contains(t, response, "top_selling_products")
	})

	// 10. Update product (Retailer)
	t.Run("Update Product", func(t *testing.T) {
		price := 899.99
		quantity := 30
		updateReq := usecases.UpdateProductRequest{
			Price:    &price,
			Quantity: &quantity,
		}

		jsonData, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", "/api/v1/product/"+productIDs[0], bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response httpHandlers.ProductResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 899.99, response.Price)
		assert.Equal(t, 30, response.Quantity)
	})
}

func TestErrorScenarios(t *testing.T) {
	appRouter, db := setupIntegrationTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	t.Run("Order with Non-existent Customer", func(t *testing.T) {
		// Create a product first
		productReq := usecases.CreateProductRequest{
			ProductName: "Test Product",
			Price:       99.99,
			Quantity:    10,
		}
		jsonData, _ := json.Marshal(productReq)
		req, _ := http.NewRequest("POST", "/api/v1/product", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		var productResponse httpHandlers.ProductResponse
		json.Unmarshal(w.Body.Bytes(), &productResponse)

		order := usecases.PlaceOrderRequest{
			CustomerID: "CUST99999",
			ProductID:  productResponse.ID,
			Quantity:   1,
		}

		jsonData, _ = json.Marshal(order)
		req, _ = http.NewRequest("POST", "/api/v1/order", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Order with Non-existent Product", func(t *testing.T) {
		// Create a customer first
		customerReq := usecases.CreateCustomerRequest{
			Name:  "Test User",
			Email: "test@example.com",
			Phone: "+1234567890",
		}
		jsonData, _ := json.Marshal(customerReq)
		req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		var customerResponse httpHandlers.CustomerResponse
		json.Unmarshal(w.Body.Bytes(), &customerResponse)

		order := usecases.PlaceOrderRequest{
			CustomerID: customerResponse.ID,
			ProductID:  "PROD99999",
			Quantity:   1,
		}

		jsonData, _ = json.Marshal(order)
		req, _ = http.NewRequest("POST", "/api/v1/order", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Order with Insufficient Quantity", func(t *testing.T) {
		// Create customer via HTTP request with unique email
		customerReq := usecases.CreateCustomerRequest{
			Name:  "Test User",
			Email: "test-insufficient-qty@example.com",
			Phone: "+1234567890",
		}
		jsonData, _ := json.Marshal(customerReq)
		req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		var customerResponse httpHandlers.CustomerResponse
		json.Unmarshal(w.Body.Bytes(), &customerResponse)

		// Create product via HTTP request
		productReq := usecases.CreateProductRequest{
			ProductName: "Limited Product",
			Price:       99.99,
			Quantity:    2,
		}
		jsonData, _ = json.Marshal(productReq)
		req, _ = http.NewRequest("POST", "/api/v1/product", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		var productResponse httpHandlers.ProductResponse
		json.Unmarshal(w.Body.Bytes(), &productResponse)

		// Create order with insufficient quantity
		order := usecases.PlaceOrderRequest{
			CustomerID: customerResponse.ID,
			ProductID:  productResponse.ID,
			Quantity:   5, // More than available
		}

		jsonData, _ = json.Marshal(order)
		req, _ = http.NewRequest("POST", "/api/v1/order", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "available_quantity")
	})
}

func TestCooldownExpired(t *testing.T) {
	appRouter, db := setupIntegrationTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Create customer and product via API
	customerReq := usecases.CreateCustomerRequest{
		Name:  "Test User",
		Email: "test-cooldown@example.com",
		Phone: "+1234567890",
	}
	jsonData, _ := json.Marshal(customerReq)
	req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	appRouter.ServeHTTP(w, req)

	var customerResponse httpHandlers.CustomerResponse
	json.Unmarshal(w.Body.Bytes(), &customerResponse)

	productReq := usecases.CreateProductRequest{
		ProductName: "Test Product",
		Price:       99.99,
		Quantity:    10,
	}
	jsonData, _ = json.Marshal(productReq)
	req, _ = http.NewRequest("POST", "/api/v1/product", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	appRouter.ServeHTTP(w, req)

	var productResponse httpHandlers.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &productResponse)

	// Manually set cooldown that has already expired (by directly updating database)
	// This simulates a customer who waited longer than the cooldown period
	db.Exec("INSERT INTO customer_cooldowns (customer_id, last_order_time, created_at, updated_at) VALUES (?, ?, ?, ?)",
		customerResponse.ID, time.Now().Add(-10*time.Minute), time.Now(), time.Now())

	// Should be able to place order (cooldown expired)
	order := usecases.PlaceOrderRequest{
		CustomerID: customerResponse.ID,
		ProductID:  productResponse.ID,
		Quantity:   1,
	}

	jsonData, _ = json.Marshal(order)
	req, _ = http.NewRequest("POST", "/api/v1/order", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	appRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response httpHandlers.OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
}
