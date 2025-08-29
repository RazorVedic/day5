package repositories

import (
	"context"
	"day5/internal/domain/entities"
)

// ProductRepository defines the contract for product data operations
// This interface belongs to the domain layer but is implemented in infrastructure
type ProductRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, product *entities.Product) error
	GetByID(ctx context.Context, id string) (*entities.Product, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Product, error)
	Update(ctx context.Context, product *entities.Product) error
	Delete(ctx context.Context, id string) error

	// Business-specific queries
	GetAvailableProducts(ctx context.Context) ([]*entities.Product, error)
	GetByPriceRange(ctx context.Context, minPrice, maxPrice float64) ([]*entities.Product, error)
	GetLowStockProducts(ctx context.Context, threshold int) ([]*entities.Product, error)

	// Inventory operations
	ReduceQuantity(ctx context.Context, productID string, quantity int) error
	IncreaseQuantity(ctx context.Context, productID string, quantity int) error

	// Search operations
	SearchByName(ctx context.Context, name string) ([]*entities.Product, error)

	// Statistics
	GetTotalValue(ctx context.Context) (float64, error)
	Count(ctx context.Context) (int, error)
}
