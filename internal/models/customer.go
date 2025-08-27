package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	ID        string    `json:"id" gorm:"type:varchar(20);primaryKey;not null"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null" binding:"required"`
	Email     string    `json:"email" gorm:"type:varchar(255);unique;not null" binding:"required,email"`
	Phone     string    `json:"phone" gorm:"type:varchar(20)" binding:"required"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationship
	Orders []Order `json:"-" gorm:"foreignKey:CustomerID"`
}

type CustomerRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Phone string `json:"phone" binding:"required"`
}

type CustomerResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message,omitempty"`
}

// BeforeCreate hook to generate custom ID before creating record
func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		id, err := generateCustomerID()
		if err != nil {
			return err
		}
		c.ID = id
	}
	return nil
}

// generateCustomerID generates a unique customer ID in format CUST12345
func generateCustomerID() (string, error) {
	max := big.NewInt(99999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	number := n.Int64() + 10000
	if number > 99999 {
		number = number%90000 + 10000
	}

	return fmt.Sprintf("CUST%05d", number), nil
}

// ToResponse converts Customer model to CustomerResponse
func (c *Customer) ToResponse(message string) CustomerResponse {
	return CustomerResponse{
		ID:        c.ID,
		Name:      c.Name,
		Email:     c.Email,
		Phone:     c.Phone,
		CreatedAt: c.CreatedAt,
		Message:   message,
	}
}
