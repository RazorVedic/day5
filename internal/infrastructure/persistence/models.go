package persistence

import (
	"time"

	"gorm.io/gorm"
)

// Database models with GORM tags for persistence
// These models are separate from domain entities to maintain Clean Architecture

// Product represents the database model for products
type Product struct {
	ID          string    `gorm:"type:varchar(20);primaryKey;not null"`
	ProductName string    `gorm:"type:varchar(255);not null;index"`
	Price       float64   `gorm:"type:decimal(10,2);not null;check:price > 0"`
	Quantity    int       `gorm:"not null;check:quantity >= 0;index"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`



	// Relationships
	Orders       []Order       `gorm:"foreignKey:ProductID"`
	Transactions []Transaction `gorm:"foreignKey:ProductID"`
}

// Customer represents the database model for customers
type Customer struct {
	ID        string    `gorm:"type:varchar(20);primaryKey;not null"`
	Name      string    `gorm:"type:varchar(255);not null;index"`
	Email     string    `gorm:"type:varchar(255);unique;not null;index"`
	Phone     string    `gorm:"type:varchar(20);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relationships
	Orders       []Order          `gorm:"foreignKey:CustomerID"`
	Transactions []Transaction    `gorm:"foreignKey:CustomerID"`
	Cooldown     CustomerCooldown `gorm:"foreignKey:CustomerID"`
}

// Order represents the database model for orders
type Order struct {
	ID          string    `gorm:"type:varchar(20);primaryKey;not null"`
	CustomerID  string    `gorm:"type:varchar(20);not null;index"`
	ProductID   string    `gorm:"type:varchar(20);not null;index"`
	Quantity    int       `gorm:"not null;check:quantity > 0"`
	UnitPrice   float64   `gorm:"type:decimal(10,2);not null;check:unit_price > 0"`
	TotalAmount float64   `gorm:"type:decimal(10,2);not null;check:total_amount > 0"`
	OrderDate   time.Time `gorm:"not null;index"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`

	// Foreign key relationships
	Customer Customer `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Product  Product  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	// Relationships
	Transactions []Transaction `gorm:"foreignKey:OrderID"`
}

// Transaction represents the database model for transactions
type Transaction struct {
	ID            string    `gorm:"type:varchar(20);primaryKey;not null"`
	OrderID       string    `gorm:"type:varchar(20);not null;index"`
	CustomerID    string    `gorm:"type:varchar(20);not null;index"`
	ProductID     string    `gorm:"type:varchar(20);not null;index"`
	Type          string    `gorm:"type:varchar(20);not null;index;check:type IN ('order','refund','credit')"`
	Amount        float64   `gorm:"type:decimal(10,2);not null;check:amount > 0"`
	Quantity      int       `gorm:"not null;check:quantity > 0"`
	UnitPrice     float64   `gorm:"type:decimal(10,2);not null;check:unit_price > 0"`
	Description   string    `gorm:"type:text"`
	TransactionAt time.Time `gorm:"not null;index"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`

	// Foreign key relationships
	Order    Order    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Customer Customer `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Product  Product  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

// CustomerCooldown represents the database model for customer cooldowns
type CustomerCooldown struct {
	CustomerID    string    `gorm:"type:varchar(20);primaryKey;not null"`
	LastOrderTime time.Time `gorm:"not null;index"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`

	// Foreign key relationship
	Customer *Customer `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// TableName methods to customize table names if needed
func (Product) TableName() string          { return "products" }
func (Customer) TableName() string         { return "customers" }
func (Order) TableName() string            { return "orders" }
func (Transaction) TableName() string      { return "transactions" }
func (CustomerCooldown) TableName() string { return "customer_cooldowns" }

// BeforeCreate hooks for generating IDs if not set
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		// ID should be generated in the use case layer
		return nil
	}
	return nil
}

func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		// ID should be generated in the use case layer
		return nil
	}
	return nil
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		// ID should be generated in the use case layer
		return nil
	}
	return nil
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		// ID should be generated in the use case layer
		return nil
	}
	return nil
}

// GetModelsToMigrate returns all models that need to be migrated
func GetModelsToMigrate() []any {
	return []any{
		&Product{},
		&Customer{},
		&Order{},
		&Transaction{},
		&CustomerCooldown{},
	}
}
