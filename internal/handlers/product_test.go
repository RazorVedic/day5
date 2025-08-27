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

func setupProductTestRouter() (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	
	// Setup test environment
	testutils.SetupTestConfig()
	db := testutils.SetupTestDB(&testing.T{})
	
	router := gin.New()
	handler := NewProductHandler()
	
	router.POST("/product", handler.CreateProduct)
	router.GET("/products", handler.GetProducts)
	router.GET("/product/:id", handler.GetProduct)
	router.PUT("/product/:id", handler.UpdateProduct)
	
	return router, db
}

func TestCreateProduct(t *testing.T) {
	router, db := setupProductTestRouter()
	defer testutils.CleanupTestDB(db)

	tests := []struct {
		name         string
		payload      interface{}
		expectedCode int
		checkDB      bool
	}{
		{
			name: "Valid product creation",
			payload: models.ProductRequest{
				ProductName: "iPhone 15",
				Price:       999.99,
				Quantity:    10,
			},
			expectedCode: http.StatusCreated,
			checkDB:      true,
		},
		{
			name: "Invalid payload - missing name",
			payload: models.ProductRequest{
				Price:    999.99,
				Quantity: 10,
			},
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
		{
			name: "Invalid payload - negative price",
			payload: models.ProductRequest{
				ProductName: "iPhone 15",
				Price:       -10.0,
				Quantity:    10,
			},
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
		{
			name: "Invalid payload - negative quantity",
			payload: models.ProductRequest{
				ProductName: "iPhone 15",
				Price:       999.99,
				Quantity:    -1,
			},
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
		{
			name:         "Invalid JSON",
			payload:      "invalid json",
			expectedCode: http.StatusBadRequest,
			checkDB:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectedCode, w.Code)
			
			if tt.checkDB && w.Code == http.StatusCreated {
				var response models.ProductResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Contains(t, response.ID, "PROD")
				assert.Equal(t, "iPhone 15", response.ProductName)
				
				// Verify in database
				var product models.Product
				err = db.First(&product, "id = ?", response.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, "iPhone 15", product.ProductName)
			}
		})
	}
}

func TestGetProducts(t *testing.T) {
	router, db := setupProductTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test products
	testutils.CreateTestProduct(db, "iPhone 15", 999.99, 10)
	testutils.CreateTestProduct(db, "MacBook Pro", 1999.99, 5)

	req, _ := http.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["count"])
	
	products := response["products"].([]interface{})
	assert.Len(t, products, 2)
}

func TestGetProduct(t *testing.T) {
	router, db := setupProductTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test product
	product := testutils.CreateTestProduct(db, "iPhone 15", 999.99, 10)

	tests := []struct {
		name         string
		productID    string
		expectedCode int
	}{
		{
			name:         "Valid product ID",
			productID:    product.ID,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid product ID",
			productID:    "INVALID",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/product/"+tt.productID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if w.Code == http.StatusOK {
				var response models.Product
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, product.ID, response.ID)
				assert.Equal(t, "iPhone 15", response.ProductName)
			}
		})
	}
}

func TestUpdateProduct(t *testing.T) {
	router, db := setupProductTestRouter()
	defer testutils.CleanupTestDB(db)

	// Create test product
	product := testutils.CreateTestProduct(db, "iPhone 15", 999.99, 10)

	tests := []struct {
		name         string
		productID    string
		payload      interface{}
		expectedCode int
		checkUpdate  bool
	}{
		{
			name:      "Valid price update",
			productID: product.ID,
			payload: models.ProductUpdateRequest{
				Price: &[]float64{899.99}[0],
			},
			expectedCode: http.StatusOK,
			checkUpdate:  true,
		},
		{
			name:      "Valid quantity update",
			productID: product.ID,
			payload: models.ProductUpdateRequest{
				Quantity: &[]int{20}[0],
			},
			expectedCode: http.StatusOK,
			checkUpdate:  true,
		},
		{
			name:      "Valid price and quantity update",
			productID: product.ID,
			payload: models.ProductUpdateRequest{
				Price:    &[]float64{849.99}[0],
				Quantity: &[]int{25}[0],
			},
			expectedCode: http.StatusOK,
			checkUpdate:  true,
		},
		{
			name:      "Invalid product ID",
			productID: "INVALID",
			payload: models.ProductUpdateRequest{
				Price: &[]float64{899.99}[0],
			},
			expectedCode: http.StatusNotFound,
			checkUpdate:  false,
		},
		{
			name:      "Invalid price (negative)",
			productID: product.ID,
			payload: models.ProductUpdateRequest{
				Price: &[]float64{-100.0}[0],
			},
			expectedCode: http.StatusBadRequest,
			checkUpdate:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("PUT", "/product/"+tt.productID, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectedCode, w.Code)
			
			if tt.checkUpdate && w.Code == http.StatusOK {
				var response models.ProductResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Product successfully updated", response.Message)
			}
		})
	}
}
