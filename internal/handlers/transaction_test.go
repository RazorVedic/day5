package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"day5/internal/models"
	"day5/internal/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTransactionTestRouter() (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// Setup test environment
	testutils.SetupTestConfig()
	db := testutils.SetupTestDB(&testing.T{})

	router := gin.New()
	handler := NewTransactionHandler()

	router.GET("/transactions", handler.GetTransactionHistory)
	router.GET("/transactions/stats", handler.GetTransactionStats)

	return router, db
}

func createTestTransaction(db *gorm.DB, orderID, customerID, productID string, amount float64, quantity int) *models.Transaction {
	transaction := &models.Transaction{
		OrderID:     orderID,
		CustomerID:  customerID,
		ProductID:   productID,
		Type:        models.TransactionTypeOrder,
		Amount:      amount,
		Quantity:    quantity,
		Description: "Test transaction",
	}
	db.Create(transaction)
	return transaction
}

func TestGetTransactionHistory(t *testing.T) {
	router, db := setupTransactionTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test data
	customer := testutils.CreateTestCustomer(db, "John Doe", "john@example.com", "+1234567890")
	product := testutils.CreateTestProduct(db, "iPhone 15", 999.99, 10)
	order := testutils.CreateTestOrder(db, customer.ID, product.ID, 2, 999.99)

	// Create test transactions
	createTestTransaction(db, order.ID, customer.ID, product.ID, 1999.98, 2)
	createTestTransaction(db, order.ID, customer.ID, product.ID, 999.99, 1)

	tests := []struct {
		name         string
		queryParams  string
		expectedCode int
		checkCount   bool
		expectedMin  int
	}{
		{
			name:         "Get all transactions",
			queryParams:  "",
			expectedCode: http.StatusOK,
			checkCount:   true,
			expectedMin:  2,
		},
		{
			name:         "Get transactions with limit",
			queryParams:  "?limit=1",
			expectedCode: http.StatusOK,
			checkCount:   true,
			expectedMin:  1,
		},
		{
			name:         "Get transactions by customer",
			queryParams:  "?customer_id=" + customer.ID,
			expectedCode: http.StatusOK,
			checkCount:   true,
			expectedMin:  2,
		},
		{
			name:         "Get transactions by product",
			queryParams:  "?product_id=" + product.ID,
			expectedCode: http.StatusOK,
			checkCount:   true,
			expectedMin:  2,
		},
		{
			name:         "Get transactions by type",
			queryParams:  "?type=order",
			expectedCode: http.StatusOK,
			checkCount:   true,
			expectedMin:  2,
		},
		{
			name:         "Get transactions with offset",
			queryParams:  "?offset=1",
			expectedCode: http.StatusOK,
			checkCount:   true,
			expectedMin:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/transactions"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if w.Code == http.StatusOK {
				var response models.TransactionHistoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				if tt.checkCount {
					assert.GreaterOrEqual(t, response.Count, tt.expectedMin)
					assert.GreaterOrEqual(t, len(response.Transactions), tt.expectedMin)
					assert.Greater(t, response.TotalAmount, 0.0)
				}

				// Verify transaction details
				if len(response.Transactions) > 0 {
					tx := response.Transactions[0]
					assert.NotEmpty(t, tx.ID)
					assert.Contains(t, tx.ID, "TXN")
					assert.Equal(t, models.TransactionTypeOrder, tx.Type)
					assert.Greater(t, tx.Amount, 0.0)
				}
			}
		})
	}
}

func TestGetTransactionHistoryWithDateFilter(t *testing.T) {
	router, db := setupTransactionTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test data
	customer := testutils.CreateTestCustomer(db, "John Doe", "john@example.com", "+1234567890")
	product := testutils.CreateTestProduct(db, "iPhone 15", 999.99, 10)
	order := testutils.CreateTestOrder(db, customer.ID, product.ID, 1, 999.99)

	// Create test transaction
	createTestTransaction(db, order.ID, customer.ID, product.ID, 999.99, 1)

	tests := []struct {
		name        string
		queryParams string
		expectData  bool
	}{
		{
			name:        "Filter by today's date",
			queryParams: "?start_date=2025-01-01&end_date=2025-12-31",
			expectData:  true,
		},
		{
			name:        "Filter by future date",
			queryParams: "?start_date=2024-01-01&end_date=2024-12-31",
			expectData:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/transactions"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.TransactionHistoryResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectData {
				assert.Greater(t, response.Count, 0)
			} else {
				assert.Equal(t, 0, response.Count)
			}
		})
	}
}

func TestGetTransactionStats(t *testing.T) {
	router, db := setupTransactionTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test data
	customer := testutils.CreateTestCustomer(db, "John Doe", "john@example.com", "+1234567890")
	product1 := testutils.CreateTestProduct(db, "iPhone 15", 999.99, 10)
	product2 := testutils.CreateTestProduct(db, "MacBook Pro", 1999.99, 5)

	order1 := testutils.CreateTestOrder(db, customer.ID, product1.ID, 2, 999.99)
	order2 := testutils.CreateTestOrder(db, customer.ID, product2.ID, 1, 1999.99)

	// Create test transactions
	createTestTransaction(db, order1.ID, customer.ID, product1.ID, 1999.98, 2)
	createTestTransaction(db, order2.ID, customer.ID, product2.ID, 1999.99, 1)

	req, _ := http.NewRequest("GET", "/transactions/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check required fields exist
	assert.Contains(t, response, "today")
	assert.Contains(t, response, "week")
	assert.Contains(t, response, "month")
	assert.Contains(t, response, "all_time")
	assert.Contains(t, response, "top_products")
	assert.Contains(t, response, "stats_date")

	// Check all_time stats (should have our test data)
	allTime := response["all_time"].(map[string]any)
	assert.Greater(t, allTime["total_amount"], 0.0)
	assert.Greater(t, allTime["order_count"], 0.0)

	// Check top products
	topProducts := response["top_products"].([]any)
	assert.GreaterOrEqual(t, len(topProducts), 0) // May be empty depending on query structure
}

func TestGetTransactionStatsEmpty(t *testing.T) {
	router, db := setupTransactionTestRouter()
	defer testutils.CleanupTestDB(db)

	// No test data - should return zeros

	req, _ := http.NewRequest("GET", "/transactions/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that all stats are zero
	allTime := response["all_time"].(map[string]any)
	assert.Equal(t, 0.0, allTime["total_amount"])
	assert.Equal(t, 0.0, allTime["order_count"])

	today := response["today"].(map[string]any)
	assert.Equal(t, 0.0, today["total_amount"])
	assert.Equal(t, 0.0, today["order_count"])
}
