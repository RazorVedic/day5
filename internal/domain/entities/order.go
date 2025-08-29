package entities

import (
	"fmt"
	"time"
)

// Order represents the core order entity
type Order struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	TotalAmount float64   `json:"total_amount"`
	OrderDate   time.Time `json:"order_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Navigation properties (not persisted, used for responses)
	Customer *Customer `json:"customer,omitempty"`
	Product  *Product  `json:"product,omitempty"`
}

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusCompleted OrderStatus = "completed"
)

// Business logic methods

// Validate performs business rule validation for orders
func (o *Order) Validate() error {
	if o.CustomerID == "" {
		return fmt.Errorf("customer ID is required")
	}

	if o.ProductID == "" {
		return fmt.Errorf("product ID is required")
	}

	if o.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero: %d", o.Quantity)
	}

	if o.UnitPrice <= 0 {
		return fmt.Errorf("unit price must be greater than zero: %f", o.UnitPrice)
	}

	// Validate total amount calculation
	expectedTotal := float64(o.Quantity) * o.UnitPrice
	if abs(o.TotalAmount-expectedTotal) > 0.01 { // Allow for floating point precision
		return fmt.Errorf("total amount mismatch: expected %f, got %f",
			expectedTotal, o.TotalAmount)
	}

	return nil
}

// CalculateTotal calculates and sets the total amount
func (o *Order) CalculateTotal() {
	o.TotalAmount = float64(o.Quantity) * o.UnitPrice
}

// SetOrderDate sets the order date to current time
func (o *Order) SetOrderDate() {
	o.OrderDate = time.Now().UTC()
}

// CanBeCancelled checks if the order can be cancelled
func (o *Order) CanBeCancelled() bool {
	// Business rule: orders can be cancelled within 30 minutes
	return time.Since(o.OrderDate) <= 30*time.Minute
}

// GetOrderSummary returns a summary of the order
func (o *Order) GetOrderSummary() map[string]any {
	return map[string]any{
		"id":           o.ID,
		"customer_id":  o.CustomerID,
		"product_id":   o.ProductID,
		"quantity":     o.Quantity,
		"unit_price":   o.UnitPrice,
		"total_amount": o.TotalAmount,
		"order_date":   o.OrderDate,
		"can_cancel":   o.CanBeCancelled(),
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
