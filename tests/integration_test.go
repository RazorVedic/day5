package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"day5/internal/models"
	"day5/internal/router"
	"day5/internal/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupIntegrationTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// Setup test environment
	testutils.SetupTestConfig()
	db := testutils.SetupTestDB(t)

	// Use the actual router setup from router package
	appRouter := router.SetupRouter()

	return appRouter, db
}

func TestCompleteRetailerWorkflow(t *testing.T) {
	appRouter, db := setupIntegrationTest(t)
	defer testutils.CleanupTestDB(db)

	// Test complete workflow: Add products → Register customers → Place orders → View analytics

	// 1. Add products (Retailer)
	t.Run("Add Products", func(t *testing.T) {
		products := []models.ProductRequest{
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

			var response models.ProductResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response.ID, "PROD")
		}
	})

	// 2. Register customers
	var customerIDs []string
	t.Run("Register Customers", func(t *testing.T) {
		customers := []models.CustomerRequest{
			{Name: "Alice Johnson", Email: "alice@example.com", Phone: "+1234567890"},
			{Name: "Bob Smith", Email: "bob@example.com", Phone: "+1987654321"},
		}

		for _, customer := range customers {
			jsonData, _ := json.Marshal(customer)
			req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			appRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response models.CustomerResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response.ID, "CUST")
			customerIDs = append(customerIDs, response.ID)
		}
	})

	// 3. Get products to place orders
	var productIDs []string
	t.Run("Get Products", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/products", nil)
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(3), response["count"])

		products := response["products"].([]any)
		for _, product := range products {
			productMap := product.(map[string]any)
			productIDs = append(productIDs, productMap["id"].(string))
		}
		assert.Len(t, productIDs, 3)
	})

	// 4. Place orders
	t.Run("Place Orders", func(t *testing.T) {
		orders := []models.OrderRequest{
			{CustomerID: customerIDs[0], ProductID: productIDs[0], Quantity: 1}, // Alice orders iPhone
			{CustomerID: customerIDs[1], ProductID: productIDs[2], Quantity: 2}, // Bob orders AirPods
		}

		for i, order := range orders {
			jsonData, _ := json.Marshal(order)
			req, _ := http.NewRequest("POST", "/api/v1/order", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			appRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response models.OrderResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response.ID, "ORD")
			assert.Equal(t, "Order successfully placed", response.Message)

			// Verify customer name mapping
			if i == 0 {
				assert.Equal(t, "Alice Johnson", response.CustomerName)
			} else {
				assert.Equal(t, "Bob Smith", response.CustomerName)
			}
		}
	})

	// 5. Test cooldown mechanism
	t.Run("Test Cooldown", func(t *testing.T) {
		// Alice tries to place another order immediately (should fail)
		order := models.OrderRequest{
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

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Customer is in cooldown period", response["error"])
		assert.Contains(t, response, "cooldown_remaining_seconds")
	})

	// 6. View order history
	t.Run("View Order History", func(t *testing.T) {
		// Alice's order history
		req, _ := http.NewRequest("GET", "/api/v1/orders/customer/"+customerIDs[0], nil)
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.OrderHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 1, response.Count)
		assert.Len(t, response.Orders, 1)
		assert.Equal(t, "Alice Johnson", response.Orders[0].CustomerName)
	})

	// 7. View all orders (Retailer)
	t.Run("View All Orders", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/orders", nil)
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.OrderHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, response.Count)
		assert.Len(t, response.Orders, 2)
	})

	// 8. View transaction history (Retailer)
	t.Run("View Transaction History", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.TransactionHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, response.Count)
		assert.Greater(t, response.TotalAmount, 0.0)
	})

	// 9. View business statistics
	t.Run("View Business Statistics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/transactions/stats", nil)
		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verify stats structure
		assert.Contains(t, response, "all_time")
		allTime := response["all_time"].(map[string]any)
		assert.Greater(t, allTime["total_amount"], 0.0)
		assert.Equal(t, 2.0, allTime["order_count"])
	})

	// 10. Update product (Retailer)
	t.Run("Update Product", func(t *testing.T) {
		updateReq := models.ProductUpdateRequest{
			Price:    &[]float64{899.99}[0],
			Quantity: &[]int{30}[0],
		}

		jsonData, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", "/api/v1/product/"+productIDs[0], bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ProductResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Product successfully updated", response.Message)
		assert.Equal(t, 899.99, response.Price)
		assert.Equal(t, 30, response.Quantity)
	})
}

func TestErrorScenarios(t *testing.T) {
	appRouter, db := setupIntegrationTest(t)
	defer testutils.CleanupTestDB(db)

	t.Run("Order with Non-existent Customer", func(t *testing.T) {
		product := testutils.CreateTestProduct(db, "Test Product", 99.99, 10)

		order := models.OrderRequest{
			CustomerID: "CUST99999",
			ProductID:  product.ID,
			Quantity:   1,
		}

		jsonData, _ := json.Marshal(order)
		req, _ := http.NewRequest("POST", "/api/v1/order", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"].(string), "Customer with ID CUST99999 not found")
	})

	t.Run("Order with Non-existent Product", func(t *testing.T) {
		customer := testutils.CreateTestCustomer(db, "Test User", "test@example.com", "+1234567890")

		order := models.OrderRequest{
			CustomerID: customer.ID,
			ProductID:  "PROD99999",
			Quantity:   1,
		}

		jsonData, _ := json.Marshal(order)
		req, _ := http.NewRequest("POST", "/api/v1/order", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"].(string), "Product with ID PROD99999 not found")
	})

	t.Run("Order with Insufficient Quantity", func(t *testing.T) {
		// Create customer via HTTP request with unique email
		customerReq := models.CustomerRequest{
			Name:  "Test User",
			Email: "test-insufficient-qty@example.com",
			Phone: "+1234567890",
		}
		customerData, _ := json.Marshal(customerReq)
		req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBuffer(customerData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		var customerResponse models.CustomerResponse
		json.Unmarshal(w.Body.Bytes(), &customerResponse)

		// Create product via HTTP request
		productReq := models.ProductRequest{
			ProductName: "Limited Product",
			Price:       99.99,
			Quantity:    2,
		}
		productData, _ := json.Marshal(productReq)
		req, _ = http.NewRequest("POST", "/api/v1/product", bytes.NewBuffer(productData))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		var productResponse models.ProductResponse
		json.Unmarshal(w.Body.Bytes(), &productResponse)

		// Create order with insufficient quantity
		order := models.OrderRequest{
			CustomerID: customerResponse.ID,
			ProductID:  productResponse.ID,
			Quantity:   5,
		}
		jsonData, _ := json.Marshal(order)
		req, _ = http.NewRequest("POST", "/api/v1/order", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		appRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Insufficient product quantity", response["error"])
		assert.Equal(t, float64(2), response["available_quantity"])
		assert.Equal(t, float64(5), response["requested_quantity"])
	})
}

func TestCooldownExpiryWorkflow(t *testing.T) {
	appRouter, db := setupIntegrationTest(t)
	defer testutils.CleanupTestDB(db)

	// Create test data
	customer := testutils.CreateTestCustomer(db, "Test Customer", "customer@example.com", "+1234567890")
	product := testutils.CreateTestProduct(db, "Test Product", 99.99, 10)

	// Set cooldown that has already expired
	cooldown := models.CustomerCooldown{
		CustomerID:    customer.ID,
		LastOrderTime: time.Now().Add(-10 * time.Minute), // 10 minutes ago
	}
	db.Create(&cooldown)

	// Should be able to place order (cooldown expired)
	order := models.OrderRequest{
		CustomerID: customer.ID,
		ProductID:  product.ID,
		Quantity:   1,
	}

	jsonData, _ := json.Marshal(order)
	req, _ := http.NewRequest("POST", "/api/v1/order", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	appRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order successfully placed", response.Message)

	// Verify cooldown was updated
	var updatedCooldown models.CustomerCooldown
	err = db.First(&updatedCooldown, "customer_id = ?", customer.ID).Error
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now(), updatedCooldown.LastOrderTime, time.Second)
}
