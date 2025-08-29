package repositories

import (
	"context"
	"day5/internal/domain/entities"
	"time"
)

// OrderRepository defines the contract for order data operations
type OrderRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, order *entities.Order) error
	GetByID(ctx context.Context, id string) (*entities.Order, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Order, error)
	Update(ctx context.Context, order *entities.Order) error
	Delete(ctx context.Context, id string) error
	
	// Customer-specific queries
	GetByCustomerID(ctx context.Context, customerID string, limit, offset int) ([]*entities.Order, error)
	GetCustomerOrderCount(ctx context.Context, customerID string) (int, error)
	
	// Product-specific queries
	GetByProductID(ctx context.Context, productID string, limit, offset int) ([]*entities.Order, error)
	
	// Time-based queries
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*entities.Order, error)
	GetTodaysOrders(ctx context.Context) ([]*entities.Order, error)
	GetRecentOrders(ctx context.Context, hours int) ([]*entities.Order, error)
	
	// Business analytics
	GetOrdersWithDetails(ctx context.Context, limit, offset int) ([]*entities.Order, error)
	GetTotalRevenue(ctx context.Context, start, end *time.Time) (float64, error)
	GetOrderCountByPeriod(ctx context.Context, start, end time.Time) (int, error)
	
	// Statistics
	Count(ctx context.Context) (int, error)
	GetAverageOrderValue(ctx context.Context) (float64, error)
}
