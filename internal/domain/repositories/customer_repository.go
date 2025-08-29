package repositories

import (
	"context"
	"day5/internal/domain/entities"
)

// CustomerRepository defines the contract for customer data operations
type CustomerRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, customer *entities.Customer) error
	GetByID(ctx context.Context, id string) (*entities.Customer, error)
	GetByEmail(ctx context.Context, email string) (*entities.Customer, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Customer, error)
	Update(ctx context.Context, customer *entities.Customer) error
	Delete(ctx context.Context, id string) error

	// Business-specific queries
	SearchByName(ctx context.Context, name string) ([]*entities.Customer, error)
	GetRecentCustomers(ctx context.Context, days int) ([]*entities.Customer, error)

	// Statistics
	Count(ctx context.Context) (int, error)
	GetActiveCustomers(ctx context.Context, days int) (int, error)
}

// CustomerCooldownRepository defines the contract for cooldown operations
type CustomerCooldownRepository interface {
	// Cooldown operations
	GetByCustomerID(ctx context.Context, customerID string) (*entities.CustomerCooldown, error)
	Upsert(ctx context.Context, cooldown *entities.CustomerCooldown) error
	Delete(ctx context.Context, customerID string) error

	// Cleanup operations
	DeleteExpiredCooldowns(ctx context.Context, olderThan int) error

	// Statistics
	GetActiveCooldowns(ctx context.Context) ([]*entities.CustomerCooldown, error)
}
