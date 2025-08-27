package models

import (
	"time"
)

type CustomerCooldown struct {
	CustomerID    string    `json:"customer_id" gorm:"type:varchar(20);primaryKey;not null"`
	LastOrderTime time.Time `json:"last_order_time" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationship
	Customer Customer `json:"customer" gorm:"foreignKey:CustomerID"`
}

const CooldownPeriod = 5 * time.Minute

// CanPlaceOrder checks if the customer can place a new order (cooldown period has passed)
func (cc *CustomerCooldown) CanPlaceOrder() bool {
	return time.Since(cc.LastOrderTime) >= CooldownPeriod
}

// RemainingCooldown returns the remaining cooldown time
func (cc *CustomerCooldown) RemainingCooldown() time.Duration {
	elapsed := time.Since(cc.LastOrderTime)
	if elapsed >= CooldownPeriod {
		return 0
	}
	return CooldownPeriod - elapsed
}

// UpdateLastOrderTime updates the last order time to now
func (cc *CustomerCooldown) UpdateLastOrderTime() {
	cc.LastOrderTime = time.Now()
}
