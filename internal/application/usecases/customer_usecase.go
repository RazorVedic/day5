package usecases

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"day5/internal/domain/entities"
	"day5/internal/domain/repositories"
)

// CustomerUseCase encapsulates business logic for customer operations
type CustomerUseCase struct {
	customerRepo   repositories.CustomerRepository
	cooldownRepo   repositories.CustomerCooldownRepository
	cooldownPeriod time.Duration
}

// NewCustomerUseCase creates a new customer use case
func NewCustomerUseCase(
	customerRepo repositories.CustomerRepository,
	cooldownRepo repositories.CustomerCooldownRepository,
	cooldownPeriodMinutes int,
) *CustomerUseCase {
	return &CustomerUseCase{
		customerRepo:   customerRepo,
		cooldownRepo:   cooldownRepo,
		cooldownPeriod: time.Duration(cooldownPeriodMinutes) * time.Minute,
	}
}

// CreateCustomerRequest represents the request to create a customer
type CreateCustomerRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Phone string `json:"phone" binding:"required"`
}

// CreateCustomer creates a new customer
func (uc *CustomerUseCase) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*entities.Customer, error) {
	// Check if email already exists
	existingCustomer, err := uc.customerRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingCustomer != nil {
		return nil, fmt.Errorf("customer with email %s already exists", req.Email)
	}

	// Generate unique customer ID
	id, err := generateCustomerID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate customer ID: %w", err)
	}

	// Create customer entity
	customer := &entities.Customer{
		ID:        id,
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Validate business rules
	if err := customer.Validate(); err != nil {
		return nil, fmt.Errorf("customer validation failed: %w", err)
	}

	// Save to repository
	if err := uc.customerRepo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (uc *CustomerUseCase) GetCustomer(ctx context.Context, id string) (*entities.Customer, error) {
	if id == "" {
		return nil, fmt.Errorf("customer ID is required")
	}

	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// GetAllCustomers retrieves all customers with pagination
func (uc *CustomerUseCase) GetAllCustomers(ctx context.Context, limit, offset int) ([]*entities.Customer, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	customers, err := uc.customerRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}

	return customers, nil
}

// CheckCustomerCooldown checks if a customer can place an order
func (uc *CustomerUseCase) CheckCustomerCooldown(ctx context.Context, customerID string) (*entities.CustomerCooldown, error) {
	if customerID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}

	// Verify customer exists
	_, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Get cooldown record
	cooldown, err := uc.cooldownRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		// No cooldown record means customer can order
		return &entities.CustomerCooldown{
			CustomerID: customerID,
		}, nil
	}

	return cooldown, nil
}

// CanCustomerPlaceOrder checks if customer can place an order based on cooldown
func (uc *CustomerUseCase) CanCustomerPlaceOrder(ctx context.Context, customerID string) (bool, *entities.CustomerCooldown, error) {
	cooldown, err := uc.CheckCustomerCooldown(ctx, customerID)
	if err != nil {
		return false, nil, err
	}

	canOrder := cooldown.CanPlaceOrder(uc.cooldownPeriod)
	return canOrder, cooldown, nil
}

// GetCooldownStatus gets the cooldown status for a customer
func (uc *CustomerUseCase) GetCooldownStatus(ctx context.Context, customerID string) (map[string]any, error) {
	cooldown, err := uc.CheckCustomerCooldown(ctx, customerID)
	if err != nil {
		return nil, err
	}

	status := cooldown.GetCooldownStatus(uc.cooldownPeriod)
	return status, nil
}

// UpdateCustomerCooldown updates the cooldown after an order is placed
func (uc *CustomerUseCase) UpdateCustomerCooldown(ctx context.Context, customerID string) error {
	if customerID == "" {
		return fmt.Errorf("customer ID is required")
	}

	// Create or update cooldown record
	cooldown := &entities.CustomerCooldown{
		CustomerID: customerID,
	}
	cooldown.UpdateLastOrderTime()

	if err := uc.cooldownRepo.Upsert(ctx, cooldown); err != nil {
		return fmt.Errorf("failed to update cooldown: %w", err)
	}

	return nil
}

// SearchCustomers searches customers by name
func (uc *CustomerUseCase) SearchCustomers(ctx context.Context, name string) ([]*entities.Customer, error) {
	if name == "" {
		return nil, fmt.Errorf("search name is required")
	}

	customers, err := uc.customerRepo.SearchByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}

	return customers, nil
}

// GetRecentCustomers gets customers registered in the last N days
func (uc *CustomerUseCase) GetRecentCustomers(ctx context.Context, days int) ([]*entities.Customer, error) {
	if days <= 0 {
		days = 30 // Default to last 30 days
	}

	customers, err := uc.customerRepo.GetRecentCustomers(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent customers: %w", err)
	}

	return customers, nil
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
