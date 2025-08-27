package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID         string      `json:"id" gorm:"type:varchar(20);primaryKey;not null"`
	CustomerID string      `json:"customer_id" gorm:"type:varchar(20);not null"`
	ProductID  string      `json:"product_id" gorm:"type:varchar(20);not null"`
	Quantity   int         `json:"quantity" gorm:"type:int;not null" binding:"required,gt=0"`
	UnitPrice  float64     `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	TotalPrice float64     `json:"total_price" gorm:"type:decimal(10,2);not null"`
	Status     OrderStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	CreatedAt  time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time   `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Customer Customer `json:"customer" gorm:"foreignKey:CustomerID"`
	Product  Product  `json:"product" gorm:"foreignKey:ProductID"`
}

type OrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ProductID  string `json:"product_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,gt=0"`
}

type OrderResponse struct {
	ID           string      `json:"id"`
	CustomerID   string      `json:"customer_id"`
	CustomerName string      `json:"customer_name"`
	ProductID    string      `json:"product_id"`
	ProductName  string      `json:"product_name"`
	Quantity     int         `json:"quantity"`
	UnitPrice    float64     `json:"unit_price"`
	TotalPrice   float64     `json:"total_price"`
	Status       OrderStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	Message      string      `json:"message,omitempty"`
}

type OrderHistoryResponse struct {
	Orders []OrderResponse `json:"orders"`
	Count  int             `json:"count"`
}

// BeforeCreate hook to generate custom ID before creating record
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		id, err := generateOrderID()
		if err != nil {
			return err
		}
		o.ID = id
	}

	// Calculate total price
	o.TotalPrice = o.UnitPrice * float64(o.Quantity)

	return nil
}

// generateOrderID generates a unique order ID in format ORD12345
func generateOrderID() (string, error) {
	max := big.NewInt(99999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	number := n.Int64() + 10000
	if number > 99999 {
		number = number%90000 + 10000
	}

	return fmt.Sprintf("ORD%05d", number), nil
}

// ToResponse converts Order model to OrderResponse
func (o *Order) ToResponse(message string) OrderResponse {
	return OrderResponse{
		ID:           o.ID,
		CustomerID:   o.CustomerID,
		CustomerName: o.Customer.Name,
		ProductID:    o.ProductID,
		ProductName:  o.Product.ProductName,
		Quantity:     o.Quantity,
		UnitPrice:    o.UnitPrice,
		TotalPrice:   o.TotalPrice,
		Status:       o.Status,
		CreatedAt:    o.CreatedAt,
		Message:      message,
	}
}
