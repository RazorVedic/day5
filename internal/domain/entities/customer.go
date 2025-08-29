package entities

import (
	"fmt"
	"strings"
	"time"
)

// Customer represents the core customer entity
type Customer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CustomerCooldown represents the cooldown period for a customer
type CustomerCooldown struct {
	CustomerID    string    `json:"customer_id"`
	LastOrderTime time.Time `json:"last_order_time"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Business logic methods

// Validate performs business rule validation for customer
func (c *Customer) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("customer name is required")
	}

	if !c.IsValidEmail() {
		return fmt.Errorf("invalid email format: %s", c.Email)
	}

	if strings.TrimSpace(c.Phone) == "" {
		return fmt.Errorf("phone number is required")
	}

	return nil
}

// IsValidEmail validates email format (basic validation)
func (c *Customer) IsValidEmail() bool {
	email := strings.TrimSpace(c.Email)
	return strings.Contains(email, "@") &&
		strings.Contains(email, ".") &&
		len(email) > 5
}

// UpdateInfo updates customer information with validation
func (c *Customer) UpdateInfo(name, email, phone string) error {
	if strings.TrimSpace(name) != "" {
		c.Name = strings.TrimSpace(name)
	}

	if strings.TrimSpace(email) != "" {
		c.Email = strings.TrimSpace(email)
		if !c.IsValidEmail() {
			return fmt.Errorf("invalid email format: %s", email)
		}
	}

	if strings.TrimSpace(phone) != "" {
		c.Phone = strings.TrimSpace(phone)
	}

	c.UpdatedAt = time.Now().UTC()
	return nil
}

// CustomerCooldown business logic

// CanPlaceOrder checks if customer can place an order based on cooldown
func (cc *CustomerCooldown) CanPlaceOrder(cooldownPeriod time.Duration) bool {
	if cc.LastOrderTime.IsZero() {
		return true // First order
	}
	return time.Since(cc.LastOrderTime) >= cooldownPeriod
}

// RemainingCooldown returns the remaining cooldown time
func (cc *CustomerCooldown) RemainingCooldown(cooldownPeriod time.Duration) time.Duration {
	if cc.LastOrderTime.IsZero() {
		return 0
	}

	elapsed := time.Since(cc.LastOrderTime)
	if elapsed >= cooldownPeriod {
		return 0
	}

	return cooldownPeriod - elapsed
}

// UpdateLastOrderTime updates the last order time to now
func (cc *CustomerCooldown) UpdateLastOrderTime() {
	cc.LastOrderTime = time.Now().UTC()
	cc.UpdatedAt = time.Now().UTC()
}

// GetCooldownStatus returns cooldown information
func (cc *CustomerCooldown) GetCooldownStatus(cooldownPeriod time.Duration) map[string]any {
	remaining := cc.RemainingCooldown(cooldownPeriod)
	return map[string]any{
		"can_order":                  cc.CanPlaceOrder(cooldownPeriod),
		"cooldown_remaining_seconds": int(remaining.Seconds()),
		"cooldown_remaining_minutes": fmt.Sprintf("%.1f", remaining.Minutes()),
		"last_order_time":            cc.LastOrderTime,
	}
}
