package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeOrder      TransactionType = "order"
	TransactionTypeRefund     TransactionType = "refund"
	TransactionTypeAdjustment TransactionType = "adjustment"
)

type Transaction struct {
	ID          string          `json:"id" gorm:"type:varchar(20);primaryKey;not null"`
	OrderID     string          `json:"order_id" gorm:"type:varchar(20);not null"`
	CustomerID  string          `json:"customer_id" gorm:"type:varchar(20);not null"`
	ProductID   string          `json:"product_id" gorm:"type:varchar(20);not null"`
	Type        TransactionType `json:"type" gorm:"type:varchar(20);not null"`
	Amount      float64         `json:"amount" gorm:"type:decimal(10,2);not null"`
	Quantity    int             `json:"quantity" gorm:"type:int;not null"`
	Description string          `json:"description" gorm:"type:text"`
	CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Order    Order    `json:"order" gorm:"foreignKey:OrderID"`
	Customer Customer `json:"customer" gorm:"foreignKey:CustomerID"`
	Product  Product  `json:"product" gorm:"foreignKey:ProductID"`
}

type TransactionResponse struct {
	ID           string          `json:"id"`
	OrderID      string          `json:"order_id"`
	CustomerID   string          `json:"customer_id"`
	CustomerName string          `json:"customer_name"`
	ProductID    string          `json:"product_id"`
	ProductName  string          `json:"product_name"`
	Type         TransactionType `json:"type"`
	Amount       float64         `json:"amount"`
	Quantity     int             `json:"quantity"`
	Description  string          `json:"description"`
	CreatedAt    time.Time       `json:"created_at"`
}

type TransactionHistoryResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Count        int                   `json:"count"`
	TotalAmount  float64               `json:"total_amount"`
}

// BeforeCreate hook to generate custom ID before creating record
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		id, err := generateTransactionID()
		if err != nil {
			return err
		}
		t.ID = id
	}
	return nil
}

// generateTransactionID generates a unique transaction ID in format TXN12345
func generateTransactionID() (string, error) {
	max := big.NewInt(99999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	number := n.Int64() + 10000
	if number > 99999 {
		number = number%90000 + 10000
	}

	return fmt.Sprintf("TXN%05d", number), nil
}

// ToResponse converts Transaction model to TransactionResponse
func (t *Transaction) ToResponse() TransactionResponse {
	return TransactionResponse{
		ID:           t.ID,
		OrderID:      t.OrderID,
		CustomerID:   t.CustomerID,
		CustomerName: t.Customer.Name,
		ProductID:    t.ProductID,
		ProductName:  t.Product.ProductName,
		Type:         t.Type,
		Amount:       t.Amount,
		Quantity:     t.Quantity,
		Description:  t.Description,
		CreatedAt:    t.CreatedAt,
	}
}
