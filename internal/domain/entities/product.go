package entities

import (
	"fmt"
	"time"
)

// Product represents the core product entity
// Domain entities contain business logic but no external dependencies
type Product struct {
	ID          string    `json:"id"`
	ProductName string    `json:"product_name"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Business logic methods on the entity

// IsAvailable checks if the product has sufficient quantity
func (p *Product) IsAvailable(requestedQuantity int) bool {
	return p.Quantity >= requestedQuantity && requestedQuantity > 0
}

// ReduceQuantity reduces the product quantity (for order processing)
func (p *Product) ReduceQuantity(amount int) error {
	if !p.IsAvailable(amount) {
		return fmt.Errorf("insufficient quantity: available=%d, requested=%d",
			p.Quantity, amount)
	}
	p.Quantity -= amount
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdatePrice updates the product price with validation
func (p *Product) UpdatePrice(newPrice float64) error {
	if newPrice <= 0 {
		return fmt.Errorf("price must be greater than zero: %f", newPrice)
	}
	p.Price = newPrice
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateQuantity updates the product quantity with validation
func (p *Product) UpdateQuantity(newQuantity int) error {
	if newQuantity < 0 {
		return fmt.Errorf("quantity cannot be negative: %d", newQuantity)
	}
	p.Quantity = newQuantity
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// CalculateValue calculates the total value of the product inventory
func (p *Product) CalculateValue() float64 {
	return p.Price * float64(p.Quantity)
}

// Validate performs business rule validation
func (p *Product) Validate() error {
	if p.ProductName == "" {
		return fmt.Errorf("product name is required")
	}
	if p.Price <= 0 {
		return fmt.Errorf("price must be greater than zero")
	}
	if p.Quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}
	return nil
}
