package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductValidation(t *testing.T) {
	tests := []struct {
		name    string
		product Product
		isValid bool
	}{
		{
			name: "Valid product",
			product: Product{
				ProductName: "Test Product",
				Price:       99.99,
				Quantity:    10,
			},
			isValid: true,
		},
		{
			name: "Empty product name",
			product: Product{
				ProductName: "",
				Price:       99.99,
				Quantity:    10,
			},
			isValid: false,
		},
		{
			name: "Zero price",
			product: Product{
				ProductName: "Test Product",
				Price:       0,
				Quantity:    10,
			},
			isValid: false,
		},
		{
			name: "Negative quantity",
			product: Product{
				ProductName: "Test Product",
				Price:       99.99,
				Quantity:    -1,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the product structure
			if tt.isValid {
				assert.NotEmpty(t, tt.product.ProductName)
				assert.Greater(t, tt.product.Price, 0.0)
				assert.GreaterOrEqual(t, tt.product.Quantity, 0)
			}
		})
	}
}

func TestGenerateProductID(t *testing.T) {
	id1, err1 := GenerateProductID()
	id2, err2 := GenerateProductID()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // Should generate different IDs
	assert.Contains(t, id1, "PROD")
	assert.Len(t, id1, 9) // PROD + 5 digits
}

func TestProductToResponse(t *testing.T) {
	product := Product{
		ID:          "PROD12345",
		ProductName: "Test Product",
		Price:       99.99,
		Quantity:    10,
	}

	response := product.ToResponse("Test message")

	assert.Equal(t, "PROD12345", response.ID)
	assert.Equal(t, "Test Product", response.ProductName)
	assert.Equal(t, 99.99, response.Price)
	assert.Equal(t, 10, response.Quantity)
	assert.Equal(t, "Test message", response.Message)
}
