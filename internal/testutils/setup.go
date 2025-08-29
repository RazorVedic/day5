package testutils

import (
	"testing"

	"day5/internal/config"
	"day5/internal/domain/entities"
	"day5/internal/infrastructure/persistence"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestConfig configures the application for testing
func SetupTestConfig() {
	// Load test configuration
	if err := config.LoadConfig(); err != nil {
		panic("Failed to load test configuration: " + err.Error())
	}
}

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent for tests
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(
		&persistence.Product{},
		&persistence.Customer{},
		&persistence.Order{},
		&persistence.Transaction{},
		&persistence.CustomerCooldown{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// CreateTestProduct creates a test product in the database
func CreateTestProduct(db *gorm.DB, name string, price float64, quantity int) *entities.Product {
	product := &persistence.Product{
		ID:          generateTestID("PROD"),
		ProductName: name,
		Price:       price,
		Quantity:    quantity,
	}

	if err := db.Create(product).Error; err != nil {
		panic("Failed to create test product: " + err.Error())
	}

	// Convert to domain entity
	entity := &entities.Product{}
	persistence.ModelToProduct(product, entity)
	return entity
}

// CreateTestCustomer creates a test customer in the database
func CreateTestCustomer(db *gorm.DB, name, email, phone string) *entities.Customer {
	customer := &persistence.Customer{
		ID:    generateTestID("CUST"),
		Name:  name,
		Email: email,
		Phone: phone,
	}

	if err := db.Create(customer).Error; err != nil {
		panic("Failed to create test customer: " + err.Error())
	}

	// Convert to domain entity
	entity := &entities.Customer{}
	persistence.ModelToCustomer(customer, entity)
	return entity
}

// CreateTestOrder creates a test order in the database
func CreateTestOrder(db *gorm.DB, customerID, productID string, quantity int, unitPrice float64) *entities.Order {
	order := &persistence.Order{
		ID:          generateTestID("ORD"),
		CustomerID:  customerID,
		ProductID:   productID,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		TotalAmount: float64(quantity) * unitPrice,
	}

	if err := db.Create(order).Error; err != nil {
		panic("Failed to create test order: " + err.Error())
	}

	// Convert to domain entity
	entity := &entities.Order{}
	persistence.ModelToOrder(order, entity)
	return entity
}

// generateTestID generates a test ID with a prefix
func generateTestID(prefix string) string {
	// Simple test ID generation - in real tests you might want something more sophisticated
	return prefix + "12345"
}
