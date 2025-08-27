package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          string    `json:"id" gorm:"type:varchar(20);primaryKey;not null"`
	ProductName string    `json:"product_name" gorm:"type:varchar(255);not null" binding:"required"`
	Price       float64   `json:"price" gorm:"type:decimal(10,2);not null" binding:"required"`
	Quantity    int       `json:"quantity" gorm:"type:int;not null" binding:"required"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ProductRequest represents the request payload for creating a product
type ProductRequest struct {
	ProductName string  `json:"product_name" binding:"required"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Quantity    int     `json:"quantity" binding:"required,gte=0"`
}

// ProductResponse represents the response for creating a product
type ProductResponse struct {
	ID          string  `json:"id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Message     string  `json:"message"`
}

// BeforeCreate hook to generate custom ID before creating record
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		id, err := generateProductID()
		if err != nil {
			return err
		}
		p.ID = id
	}
	return nil
}

// generateProductID generates a unique product ID in format PROD12345
func generateProductID() (string, error) {
	// Generate a random 5-digit number
	max := big.NewInt(99999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	
	// Ensure it's always 5 digits by adding to 10000
	number := n.Int64() + 10000
	if number > 99999 {
		number = number % 90000 + 10000
	}
	
	return fmt.Sprintf("PROD%05d", number), nil
}

// ToResponse converts Product model to ProductResponse
func (p *Product) ToResponse(message string) ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		ProductName: p.ProductName,
		Price:       p.Price,
		Quantity:    p.Quantity,
		Message:     message,
	}
}
