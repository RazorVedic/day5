package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"day5/internal/domain/entities"
	"day5/internal/domain/repositories"
	"day5/internal/infrastructure/persistence"

	"gorm.io/gorm"
)

// CustomerRepositoryImpl implements the CustomerRepository interface
type CustomerRepositoryImpl struct {
	db *gorm.DB
	mu sync.RWMutex
}

// NewCustomerRepository creates a new customer repository implementation
func NewCustomerRepository(db *gorm.DB) repositories.CustomerRepository {
	return &CustomerRepositoryImpl{
		db: db,
	}
}

// Create creates a new customer with thread safety
func (r *CustomerRepositoryImpl) Create(ctx context.Context, customer *entities.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	model := persistence.CustomerToModel(customer)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	persistence.ModelToCustomer(model, customer)
	return nil
}

// GetByID retrieves a customer by ID
func (r *CustomerRepositoryImpl) GetByID(ctx context.Context, id string) (*entities.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var model persistence.Customer
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("customer with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	customer := &entities.Customer{}
	persistence.ModelToCustomer(&model, customer)
	return customer, nil
}

// GetByEmail retrieves a customer by email
func (r *CustomerRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entities.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var model persistence.Customer
	if err := r.db.WithContext(ctx).First(&model, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("customer with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	customer := &entities.Customer{}
	persistence.ModelToCustomer(&model, customer)
	return customer, nil
}

// GetAll retrieves all customers with pagination
func (r *CustomerRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Customer
	query := r.db.WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}

	return persistence.ModelsToCustomers(models), nil
}

// Update updates a customer
func (r *CustomerRepositoryImpl) Update(ctx context.Context, customer *entities.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	model := persistence.CustomerToModel(customer)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	persistence.ModelToCustomer(model, customer)
	return nil
}

// Delete deletes a customer
func (r *CustomerRepositoryImpl) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.db.WithContext(ctx).Delete(&persistence.Customer{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete customer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("customer with ID %s not found", id)
	}

	return nil
}

// SearchByName searches customers by name
func (r *CustomerRepositoryImpl) SearchByName(ctx context.Context, name string) ([]*entities.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Customer
	searchPattern := "%" + name + "%"
	if err := r.db.WithContext(ctx).Where("name LIKE ?", searchPattern).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to search customers by name: %w", err)
	}

	return persistence.ModelsToCustomers(models), nil
}

// GetRecentCustomers gets customers registered in the last N days
func (r *CustomerRepositoryImpl) GetRecentCustomers(ctx context.Context, days int) ([]*entities.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	since := time.Now().AddDate(0, 0, -days)
	var models []persistence.Customer
	if err := r.db.WithContext(ctx).Where("created_at >= ?", since).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent customers: %w", err)
	}

	return persistence.ModelsToCustomers(models), nil
}

// Count returns the total number of customers
func (r *CustomerRepositoryImpl) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	if err := r.db.WithContext(ctx).Model(&persistence.Customer{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count customers: %w", err)
	}

	return int(count), nil
}

// GetActiveCustomers returns customers who have placed orders in the last N days
func (r *CustomerRepositoryImpl) GetActiveCustomers(ctx context.Context, days int) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	since := time.Now().AddDate(0, 0, -days)
	var count int64

	subQuery := r.db.Select("DISTINCT customer_id").Table("orders").Where("created_at >= ?", since)
	if err := r.db.WithContext(ctx).Model(&persistence.Customer{}).Where("id IN (?)", subQuery).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count active customers: %w", err)
	}

	return int(count), nil
}

// CustomerCooldownRepositoryImpl implements the CustomerCooldownRepository interface
type CustomerCooldownRepositoryImpl struct {
	db *gorm.DB
	mu sync.RWMutex
}

// NewCustomerCooldownRepository creates a new customer cooldown repository
func NewCustomerCooldownRepository(db *gorm.DB) repositories.CustomerCooldownRepository {
	return &CustomerCooldownRepositoryImpl{
		db: db,
	}
}

// GetByCustomerID gets cooldown record by customer ID
func (r *CustomerCooldownRepositoryImpl) GetByCustomerID(ctx context.Context, customerID string) (*entities.CustomerCooldown, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var model persistence.CustomerCooldown
	if err := r.db.WithContext(ctx).First(&model, "customer_id = ?", customerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("cooldown record for customer %s not found", customerID)
		}
		return nil, fmt.Errorf("failed to get cooldown: %w", err)
	}

	cooldown := &entities.CustomerCooldown{}
	persistence.ModelToCooldown(&model, cooldown)
	return cooldown, nil
}

// Upsert creates or updates a cooldown record
func (r *CustomerCooldownRepositoryImpl) Upsert(ctx context.Context, cooldown *entities.CustomerCooldown) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	model := persistence.CooldownToModel(cooldown)

	// Use GORM's Clauses for proper upsert
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to upsert cooldown: %w", err)
	}

	persistence.ModelToCooldown(model, cooldown)
	return nil
}

// Delete deletes a cooldown record
func (r *CustomerCooldownRepositoryImpl) Delete(ctx context.Context, customerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.db.WithContext(ctx).Delete(&persistence.CustomerCooldown{}, "customer_id = ?", customerID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete cooldown: %w", result.Error)
	}

	return nil
}

// DeleteExpiredCooldowns deletes cooldown records older than specified hours
func (r *CustomerCooldownRepositoryImpl) DeleteExpiredCooldowns(ctx context.Context, olderThanHours int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-time.Duration(olderThanHours) * time.Hour)
	result := r.db.WithContext(ctx).Where("last_order_time < ?", cutoff).Delete(&persistence.CustomerCooldown{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete expired cooldowns: %w", result.Error)
	}

	return nil
}

// GetActiveCooldowns gets all active cooldown records
func (r *CustomerCooldownRepositoryImpl) GetActiveCooldowns(ctx context.Context) ([]*entities.CustomerCooldown, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.CustomerCooldown
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get active cooldowns: %w", err)
	}

	cooldowns := make([]*entities.CustomerCooldown, len(models))
	for i, model := range models {
		cooldowns[i] = &entities.CustomerCooldown{}
		persistence.ModelToCooldown(&model, cooldowns[i])
	}

	return cooldowns, nil
}
