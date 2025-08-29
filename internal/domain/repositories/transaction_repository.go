package repositories

import (
	"context"
	"day5/internal/domain/entities"
	"time"
)

// TransactionRepository defines the contract for transaction data operations
type TransactionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, transaction *entities.Transaction) error
	GetByID(ctx context.Context, id string) (*entities.Transaction, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Transaction, error)
	Update(ctx context.Context, transaction *entities.Transaction) error
	Delete(ctx context.Context, id string) error

	// Filtering operations
	GetByCustomerID(ctx context.Context, customerID string, limit, offset int) ([]*entities.Transaction, error)
	GetByProductID(ctx context.Context, productID string, limit, offset int) ([]*entities.Transaction, error)
	GetByOrderID(ctx context.Context, orderID string) (*entities.Transaction, error)
	GetByType(ctx context.Context, transactionType entities.TransactionType, limit, offset int) ([]*entities.Transaction, error)

	// Time-based queries
	GetByDateRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*entities.Transaction, error)
	GetTodaysTransactions(ctx context.Context) ([]*entities.Transaction, error)
	GetTransactionsByPeriod(ctx context.Context, start, end time.Time) ([]*entities.Transaction, error)

	// Business analytics and reporting
	GetBusinessStats(ctx context.Context, start, end *time.Time) (*entities.BusinessStats, error)
	GetRevenueByPeriod(ctx context.Context, start, end time.Time) (float64, error)
	GetTopSellingProducts(ctx context.Context, limit int, start, end *time.Time) ([]*entities.ProductSales, error)
	GetCustomerTransactionSummary(ctx context.Context, customerID string) (map[string]any, error)

	// Statistics
	Count(ctx context.Context) (int, error)
	GetTotalRevenue(ctx context.Context) (float64, error)
	GetTransactionCountByType(ctx context.Context, transactionType entities.TransactionType) (int, error)

	// Advanced analytics
	GetDailyRevenue(ctx context.Context, days int) ([]map[string]any, error)
	GetMonthlyRevenue(ctx context.Context, months int) ([]map[string]any, error)
	GetRevenueGrowth(ctx context.Context) (map[string]any, error)
}
