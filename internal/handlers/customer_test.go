package handlers

import (
	"bytes"
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

func setupCustomerTestRouter() (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// Setup test environment
	testutils.SetupTestConfig()
	db := testutils.SetupTestDB(&testing.T{})

	router := gin.New()
	handler := NewCustomerHandler()

	router.POST("/customer", handler.CreateCustomer)
	router.GET("/customers", handler.GetCustomers)
	router.GET("/customer/:id", handler.GetCustomer)

	return router, db
}

func TestCreateCustomer(t *testing.T) {
	router, db := setupCustomerTestRouter()
	defer testutils.CleanupTestDB(db)

	tests := []struct {
		name         string
		payload      interface{}
		expectedCode int
		checkDB      bool
	}{
		{
			name: "Valid customer creation",
			payload: models.CustomerRequest{
				Name:  "John Doe",
				Email: "john@example.com",
				Phone: "+1234567890",
			},
			expectedCode: http.StatusCreated,
			checkDB:      true,
		},
		{
			name: "Invalid email",
			payload: models.CustomerRequest{
				Name:  "John Doe",
				Email: "invalid-email",
				Phone: "+1234567890",
			},
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
		{
			name: "Missing name",
			payload: models.CustomerRequest{
				Email: "john@example.com",
				Phone: "+1234567890",
			},
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
		{
			name: "Missing email",
			payload: models.CustomerRequest{
				Name:  "John Doe",
				Phone: "+1234567890",
			},
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
		{
			name: "Missing phone",
			payload: models.CustomerRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/customer", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.checkDB && w.Code == http.StatusCreated {
				var response models.CustomerResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Contains(t, response.ID, "CUST")
				assert.Equal(t, "John Doe", response.Name)
				assert.Equal(t, "customer successfully created", response.Message)

				// Verify in database
				var customer models.Customer
				err = db.First(&customer, "id = ?", response.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, "John Doe", customer.Name)
				assert.Equal(t, "john@example.com", customer.Email)
			}
		})
	}
}

func TestCreateCustomerDuplicateEmail(t *testing.T) {
	router, db := setupCustomerTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create first customer
	customer1 := models.CustomerRequest{
		Name:  "John Doe",
		Email: "john@example.com",
		Phone: "+1234567890",
	}

	jsonData, _ := json.Marshal(customer1)
	req, _ := http.NewRequest("POST", "/customer", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Try to create second customer with same email
	customer2 := models.CustomerRequest{
		Name:  "Jane Doe",
		Email: "john@example.com", // Same email
		Phone: "+0987654321",
	}

	jsonData2, _ := json.Marshal(customer2)
	req2, _ := http.NewRequest("POST", "/customer", bytes.NewBuffer(jsonData2))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusInternalServerError, w2.Code) // Database constraint violation
}

func TestGetCustomers(t *testing.T) {
	router, db := setupCustomerTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test customers
	testutils.CreateTestCustomer(db, "John Doe", "john@example.com", "+1234567890")
	testutils.CreateTestCustomer(db, "Jane Smith", "jane@example.com", "+0987654321")

	req, _ := http.NewRequest("GET", "/customers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["count"])

	customers := response["customers"].([]interface{})
	assert.Len(t, customers, 2)
}

func TestGetCustomer(t *testing.T) {
	router, db := setupCustomerTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test customer
	customer := testutils.CreateTestCustomer(db, "John Doe", "john@example.com", "+1234567890")

	tests := []struct {
		name         string
		customerID   string
		expectedCode int
	}{
		{
			name:         "Valid customer ID",
			customerID:   customer.ID,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid customer ID",
			customerID:   "INVALID",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/customer/"+tt.customerID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if w.Code == http.StatusOK {
				var response models.CustomerResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, customer.ID, response.ID)
				assert.Equal(t, "John Doe", response.Name)
				assert.Equal(t, "john@example.com", response.Email)
			}
		})
	}
}
