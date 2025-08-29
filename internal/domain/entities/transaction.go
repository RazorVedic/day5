package entities

import (
	"fmt"
	"slices"
	"time"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeOrder  TransactionType = "order"
	TransactionTypeRefund TransactionType = "refund"
	TransactionTypeCredit TransactionType = "credit"
)

// Transaction represents the core transaction entity for business analytics
type Transaction struct {
	ID            string          `json:"id"`
	OrderID       string          `json:"order_id"`
	CustomerID    string          `json:"customer_id"`
	ProductID     string          `json:"product_id"`
	Type          TransactionType `json:"type"`
	Amount        float64         `json:"amount"`
	Quantity      int             `json:"quantity"`
	UnitPrice     float64         `json:"unit_price"`
	Description   string          `json:"description"`
	TransactionAt time.Time       `json:"transaction_at"`
	CreatedAt     time.Time       `json:"created_at"`

	// Navigation properties
	Order    *Order    `json:"order,omitempty"`
	Customer *Customer `json:"customer,omitempty"`
	Product  *Product  `json:"product,omitempty"`
}

// Business logic methods

// Validate performs business rule validation for transactions
func (t *Transaction) Validate() error {
	if t.OrderID == "" {
		return fmt.Errorf("order ID is required")
	}

	if t.CustomerID == "" {
		return fmt.Errorf("customer ID is required")
	}

	if t.ProductID == "" {
		return fmt.Errorf("product ID is required")
	}

	if !t.IsValidType() {
		return fmt.Errorf("invalid transaction type: %s", t.Type)
	}

	if t.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero: %f", t.Amount)
	}

	if t.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero: %d", t.Quantity)
	}

	return nil
}

// IsValidType checks if the transaction type is valid
func (t *Transaction) IsValidType() bool {
	validTypes := []TransactionType{
		TransactionTypeOrder,
		TransactionTypeRefund,
		TransactionTypeCredit,
	}
	return slices.Contains(validTypes, t.Type)
}

// SetTransactionTime sets the transaction time to current time
func (t *Transaction) SetTransactionTime() {
	t.TransactionAt = time.Now().UTC()
}

// IsRevenue returns true if this transaction generates revenue
func (t *Transaction) IsRevenue() bool {
	return t.Type == TransactionTypeOrder
}

// IsRefund returns true if this transaction is a refund
func (t *Transaction) IsRefund() bool {
	return t.Type == TransactionTypeRefund
}

// GetRevenueAmount returns the revenue amount (positive for orders, negative for refunds)
func (t *Transaction) GetRevenueAmount() float64 {
	switch t.Type {
	case TransactionTypeOrder:
		return t.Amount
	case TransactionTypeRefund:
		return -t.Amount
	default:
		return 0
	}
}

// CreateFromOrder creates a transaction from an order
func (t *Transaction) CreateFromOrder(order *Order) {
	t.OrderID = order.ID
	t.CustomerID = order.CustomerID
	t.ProductID = order.ProductID
	t.Type = TransactionTypeOrder
	t.Amount = order.TotalAmount
	t.Quantity = order.Quantity
	t.UnitPrice = order.UnitPrice
	t.Description = fmt.Sprintf("Order for %d units", order.Quantity)
	t.SetTransactionTime()
}

// BusinessStats represents business statistics
type BusinessStats struct {
	TotalRevenue       float64        `json:"total_revenue"`
	OrderCount         int            `json:"order_count"`
	AverageOrderValue  float64        `json:"average_order_value"`
	TotalQuantitySold  int            `json:"total_quantity_sold"`
	UniqueCustomers    int            `json:"unique_customers"`
	TopSellingProducts []ProductSales `json:"top_selling_products,omitempty"`
}

// ProductSales represents sales data for a product
type ProductSales struct {
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	QuantitySold int     `json:"quantity_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

// CalculateAverageOrderValue calculates the average order value
func (bs *BusinessStats) CalculateAverageOrderValue() {
	if bs.OrderCount > 0 {
		bs.AverageOrderValue = bs.TotalRevenue / float64(bs.OrderCount)
	} else {
		bs.AverageOrderValue = 0
	}
}
