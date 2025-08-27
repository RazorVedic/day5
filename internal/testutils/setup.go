package testutils

import (
	"os"
	"testing"

	"day5/internal/config"
	"day5/internal/database"
	"day5/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	// Set test environment
	os.Setenv("ENV", "test")

	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Set the global database for the application
	database.DB = db

	// Run migrations
	err = db.AutoMigrate(
		&models.Product{},
		&models.Customer{},
		&models.Order{},
		&models.Transaction{},
		&models.CustomerCooldown{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(db *gorm.DB) {
	// Clean all tables
	db.Exec("DELETE FROM transactions")
	db.Exec("DELETE FROM customer_cooldowns")
	db.Exec("DELETE FROM orders")
	db.Exec("DELETE FROM customers")
	db.Exec("DELETE FROM products")
}

// SetupTestConfig sets up test configuration
func SetupTestConfig() {
	config.AppConfig = &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
			Env:  "test",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "3306",
			User:     "test",
			Password: "test",
			DBName:   "test",
			DSN:      "test",
		},
	}
}

// CreateTestProduct creates a test product
func CreateTestProduct(db *gorm.DB, name string, price float64, quantity int) *models.Product {
	product := &models.Product{
		ProductName: name,
		Price:       price,
		Quantity:    quantity,
	}
	db.Create(product)
	return product
}

// CreateTestCustomer creates a test customer
func CreateTestCustomer(db *gorm.DB, name, email, phone string) *models.Customer {
	customer := &models.Customer{
		Name:  name,
		Email: email,
		Phone: phone,
	}
	db.Create(customer)
	return customer
}

// CreateTestOrder creates a test order
func CreateTestOrder(db *gorm.DB, customerID, productID string, quantity int, unitPrice float64) *models.Order {
	order := &models.Order{
		CustomerID: customerID,
		ProductID:  productID,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		Status:     models.OrderStatusCompleted,
	}
	db.Create(order)
	return order
}
