package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"day5/internal/models"
	"day5/internal/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupOrderTestRouter() (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// Setup test environment
	testutils.SetupTestConfig()
	db := testutils.SetupTestDB(&testing.T{})

	router := gin.New()
	handler := NewOrderHandler()

	router.POST("/order", handler.PlaceOrder)
	router.GET("/orders/customer/:customer_id", handler.GetOrderHistory)
	router.GET("/orders", handler.GetAllOrders)

	return router, db
}

func TestPlaceOrder(t *testing.T) {
	router, db := setupOrderTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test data
	customer := testutils.CreateTestCustomer(db, "John Doe", "john@example.com", "+1234567890")
	product := testutils.CreateTestProduct(db, "iPhone 15", 999.99, 10)

	tests := []struct {
		name         string
		payload      any
		expectedCode int
		checkDB      bool
	}{
		{
			name: "Valid order",
			payload: models.OrderRequest{
				CustomerID: customer.ID,
				ProductID:  product.ID,
				Quantity:   2,
			},
			expectedCode: http.StatusCreated,
			checkDB:      true,
		},
		{
			name: "Invalid customer ID",
			payload: models.OrderRequest{
				CustomerID: "INVALID",
				ProductID:  product.ID,
				Quantity:   1,
			},
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
		{
			name: "Invalid product ID",
			payload: models.OrderRequest{
				CustomerID: customer.ID,
				ProductID:  "INVALID",
				Quantity:   1,
			},
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
		{
			name: "Insufficient quantity",
			payload: models.OrderRequest{
				CustomerID: customer.ID,
				ProductID:  product.ID,
				Quantity:   20, // More than available (10)
			},
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
		{
			name: "Invalid quantity (zero)",
			payload: models.OrderRequest{
				CustomerID: customer.ID,
				ProductID:  product.ID,
				Quantity:   0,
			},
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean cooldown for fresh test
			db.Delete(&models.CustomerCooldown{}, "customer_id = ?", customer.ID)

			jsonData, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/order", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.checkDB && w.Code == http.StatusCreated {
				var response models.OrderResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Contains(t, response.ID, "ORD")
				assert.Equal(t, "Order successfully placed", response.Message)

				// Verify order in database
				var order models.Order
				err = db.First(&order, "id = ?", response.ID).Error
				assert.NoError(t, err)

				// Verify transaction was created
				var transaction models.Transaction
				err = db.First(&transaction, "order_id = ?", response.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, models.TransactionTypeOrder, transaction.Type)

				// Verify cooldown was set
				var cooldown models.CustomerCooldown
				err = db.First(&cooldown, "customer_id = ?", customer.ID).Error
				assert.NoError(t, err)
				assert.WithinDuration(t, time.Now(), cooldown.LastOrderTime, time.Second)

				// Verify product quantity was updated
				var updatedProduct models.Product
				err = db.First(&updatedProduct, "id = ?", product.ID).Error
				assert.NoError(t, err)
				expectedQuantity := product.Quantity - tt.payload.(models.OrderRequest).Quantity
				assert.Equal(t, expectedQuantity, updatedProduct.Quantity)
			}
		})
	}
}

func TestCooldownMechanism(t *testing.T) {
	router, db := setupOrderTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test data
	customer := testutils.CreateTestCustomer(db, "John Doe", "john@example.com", "+1234567890")
	product := testutils.CreateTestProduct(db, "iPhone 15", 999.99, 10)

	// Place first order
	orderPayload := models.OrderRequest{
		CustomerID: customer.ID,
		ProductID:  product.ID,
		Quantity:   1,
	}

	jsonData, _ := json.Marshal(orderPayload)
	req, _ := http.NewRequest("POST", "/order", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Try to place second order immediately (should fail)
	req2, _ := http.NewRequest("POST", "/order", bytes.NewBuffer(jsonData))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	var errorResponse map[string]any
	err := json.Unmarshal(w2.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Customer is in cooldown period", errorResponse["error"])
	assert.Contains(t, errorResponse, "cooldown_remaining_seconds")
	assert.Contains(t, errorResponse, "cooldown_remaining_minutes")
}

func TestOrderAfterCooldownExpires(t *testing.T) {
	router, db := setupOrderTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test data
	customer := testutils.CreateTestCustomer(db, "John Doe", "john@example.com", "+1234567890")
	product := testutils.CreateTestProduct(db, "iPhone 15", 999.99, 10)

	// Set cooldown manually with old timestamp (cooldown expired)
	cooldown := models.CustomerCooldown{
		CustomerID:    customer.ID,
		LastOrderTime: time.Now().Add(-10 * time.Minute), // 10 minutes ago
	}
	db.Create(&cooldown)

	// Place order (should succeed as cooldown has expired)
	orderPayload := models.OrderRequest{
		CustomerID: customer.ID,
		ProductID:  product.ID,
		Quantity:   1,
	}

	jsonData, _ := json.Marshal(orderPayload)
	req, _ := http.NewRequest("POST", "/order", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order successfully placed", response.Message)
}

func TestGetOrderHistory(t *testing.T) {
	router, db := setupOrderTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test data
	customer := testutils.CreateTestCustomer(db, "John Doe", "john@example.com", "+1234567890")
	product := testutils.CreateTestProduct(db, "iPhone 15", 999.99, 10)
	testutils.CreateTestOrder(db, customer.ID, product.ID, 1, 999.99)
	testutils.CreateTestOrder(db, customer.ID, product.ID, 2, 999.99)

	tests := []struct {
		name          string
		customerID    string
		expectedCode  int
		expectedCount int
	}{
		{
			name:          "Valid customer with orders",
			customerID:    customer.ID,
			expectedCode:  http.StatusOK,
			expectedCount: 2,
		},
		{
			name:          "Invalid customer ID",
			customerID:    "INVALID",
			expectedCode:  http.StatusNotFound,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/orders/customer/"+tt.customerID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if w.Code == http.StatusOK {
				var response models.OrderHistoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, response.Count)
				assert.Len(t, response.Orders, tt.expectedCount)

				// Verify order details
				if len(response.Orders) > 0 {
					assert.Equal(t, customer.ID, response.Orders[0].CustomerID)
					assert.Equal(t, product.ID, response.Orders[0].ProductID)
				}
			}
		})
	}
}

func TestGetAllOrders(t *testing.T) {
	router, db := setupOrderTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test data
	customer1 := testutils.CreateTestCustomer(db, "John Doe", "john@example.com", "+1234567890")
	customer2 := testutils.CreateTestCustomer(db, "Jane Smith", "jane@example.com", "+0987654321")
	product := testutils.CreateTestProduct(db, "iPhone 15", 999.99, 10)

	// Create test orders for different customers
	testutils.CreateTestOrder(db, customer1.ID, product.ID, 1, 999.99)
	testutils.CreateTestOrder(db, customer2.ID, product.ID, 2, 999.99)

	req, _ := http.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.OrderHistoryResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.Count)
	assert.Len(t, response.Orders, 2)
}
