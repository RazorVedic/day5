package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCustomerCooldownCanPlaceOrder(t *testing.T) {
	tests := []struct {
		name          string
		lastOrderTime time.Time
		canPlace      bool
	}{
		{
			name:          "No previous order",
			lastOrderTime: time.Time{},
			canPlace:      true,
		},
		{
			name:          "Order 6 minutes ago",
			lastOrderTime: time.Now().Add(-6 * time.Minute),
			canPlace:      true,
		},
		{
			name:          "Order 5 minutes ago",
			lastOrderTime: time.Now().Add(-5 * time.Minute),
			canPlace:      true,
		},
		{
			name:          "Order 4 minutes ago",
			lastOrderTime: time.Now().Add(-4 * time.Minute),
			canPlace:      false,
		},
		{
			name:          "Order 1 minute ago",
			lastOrderTime: time.Now().Add(-1 * time.Minute),
			canPlace:      false,
		},
		{
			name:          "Order just now",
			lastOrderTime: time.Now(),
			canPlace:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cooldown := CustomerCooldown{
				CustomerID:    "CUST12345",
				LastOrderTime: tt.lastOrderTime,
			}

			result := cooldown.CanPlaceOrder()
			assert.Equal(t, tt.canPlace, result)
		})
	}
}

func TestCustomerCooldownRemainingCooldown(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		lastOrderTime time.Time
		expectZero    bool
	}{
		{
			name:          "No cooldown remaining",
			lastOrderTime: now.Add(-6 * time.Minute),
			expectZero:    true,
		},
		{
			name:          "Some cooldown remaining",
			lastOrderTime: now.Add(-2 * time.Minute),
			expectZero:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cooldown := CustomerCooldown{
				CustomerID:    "CUST12345",
				LastOrderTime: tt.lastOrderTime,
			}

			remaining := cooldown.RemainingCooldown()
			if tt.expectZero {
				assert.Equal(t, time.Duration(0), remaining)
			} else {
				assert.Greater(t, remaining, time.Duration(0))
				assert.LessOrEqual(t, remaining, CooldownPeriod)
			}
		})
	}
}

func TestCustomerCooldownUpdateLastOrderTime(t *testing.T) {
	cooldown := CustomerCooldown{
		CustomerID:    "CUST12345",
		LastOrderTime: time.Now().Add(-1 * time.Hour),
	}

	oldTime := cooldown.LastOrderTime
	cooldown.UpdateLastOrderTime()

	assert.Greater(t, cooldown.LastOrderTime, oldTime)
	assert.WithinDuration(t, time.Now(), cooldown.LastOrderTime, time.Second)
}

func TestCooldownPeriodConstant(t *testing.T) {
	assert.Equal(t, 5*time.Minute, CooldownPeriod)
}
