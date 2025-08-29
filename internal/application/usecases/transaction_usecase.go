package usecases

import (
	"context"
	"fmt"
	"time"

	"day5/internal/domain/entities"
	"day5/internal/domain/repositories"
)

// TransactionUseCase encapsulates business logic for transaction and analytics operations
type TransactionUseCase struct {
	transactionRepo repositories.TransactionRepository
	customerRepo    repositories.CustomerRepository
	productRepo     repositories.ProductRepository
}

// NewTransactionUseCase creates a new transaction use case
func NewTransactionUseCase(
	transactionRepo repositories.TransactionRepository,
	customerRepo repositories.CustomerRepository,
	productRepo repositories.ProductRepository,
) *TransactionUseCase {
	return &TransactionUseCase{
		transactionRepo: transactionRepo,
		customerRepo:    customerRepo,
		productRepo:     productRepo,
	}
}

// GetTransactionHistory gets transaction history with optional filters
func (uc *TransactionUseCase) GetTransactionHistory(ctx context.Context, filters TransactionFilters) ([]*entities.Transaction, error) {
	limit := filters.Limit
	if limit <= 0 {
		limit = 50
	}

	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	var transactions []*entities.Transaction
	var err error

	// Apply filters
	switch {
	case filters.CustomerID != "":
		transactions, err = uc.transactionRepo.GetByCustomerID(ctx, filters.CustomerID, limit, offset)
	case filters.ProductID != "":
		transactions, err = uc.transactionRepo.GetByProductID(ctx, filters.ProductID, limit, offset)
	case filters.Type != "":
		transactionType := entities.TransactionType(filters.Type)
		transactions, err = uc.transactionRepo.GetByType(ctx, transactionType, limit, offset)
	case filters.StartDate != nil && filters.EndDate != nil:
		transactions, err = uc.transactionRepo.GetByDateRange(ctx, *filters.StartDate, *filters.EndDate, limit, offset)
	default:
		transactions, err = uc.transactionRepo.GetAll(ctx, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}

	// Enrich transactions with related data
	for _, transaction := range transactions {
		if err := uc.enrichTransaction(ctx, transaction); err != nil {
			// Log error but don't fail the entire operation
			continue
		}
	}

	return transactions, nil
}

// GetBusinessStats gets comprehensive business statistics
func (uc *TransactionUseCase) GetBusinessStats(ctx context.Context, period StatsPeriod) (map[string]any, error) {
	var start, end *time.Time
	now := time.Now().UTC()

	// Calculate time ranges based on period
	switch period {
	case StatsPeriodToday:
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		start = &startOfDay
		end = &now
	case StatsPeriodThisWeek:
		weekStart := now.AddDate(0, 0, -int(now.Weekday()))
		weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
		start = &weekStart
		end = &now
	case StatsPeriodThisMonth:
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		start = &monthStart
		end = &now
	case StatsPeriodAllTime:
		// No time filter for all-time stats
	}

	// Get business stats from repository
	stats, err := uc.transactionRepo.GetBusinessStats(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get business stats: %w", err)
	}

	// Get additional metrics
	topProducts, err := uc.transactionRepo.GetTopSellingProducts(ctx, 5, start, end)
	if err == nil && len(topProducts) > 0 {
		stats.TopSellingProducts = make([]entities.ProductSales, len(topProducts))
		for i, product := range topProducts {
			stats.TopSellingProducts[i] = *product
		}
	}

	// Format response
	response := map[string]any{
		"total_revenue":       stats.TotalRevenue,
		"order_count":         stats.OrderCount,
		"average_order_value": stats.AverageOrderValue,
		"total_quantity_sold": stats.TotalQuantitySold,
		"unique_customers":    stats.UniqueCustomers,
	}

	if len(stats.TopSellingProducts) > 0 {
		response["top_selling_products"] = stats.TopSellingProducts
	}

	return response, nil
}

// GetComprehensiveStats gets stats for multiple periods
func (uc *TransactionUseCase) GetComprehensiveStats(ctx context.Context) (map[string]any, error) {
	// Get stats for different periods
	allTimeStats, err := uc.GetBusinessStats(ctx, StatsPeriodAllTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get all-time stats: %w", err)
	}

	todayStats, err := uc.GetBusinessStats(ctx, StatsPeriodToday)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's stats: %w", err)
	}

	thisWeekStats, err := uc.GetBusinessStats(ctx, StatsPeriodThisWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get this week's stats: %w", err)
	}

	thisMonthStats, err := uc.GetBusinessStats(ctx, StatsPeriodThisMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get this month's stats: %w", err)
	}

	return map[string]any{
		"all_time":   allTimeStats,
		"today":      todayStats,
		"this_week":  thisWeekStats,
		"this_month": thisMonthStats,
	}, nil
}

// GetCustomerTransactionSummary gets transaction summary for a specific customer
func (uc *TransactionUseCase) GetCustomerTransactionSummary(ctx context.Context, customerID string) (map[string]any, error) {
	if customerID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}

	// Verify customer exists
	customer, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Get transaction summary
	summary, err := uc.transactionRepo.GetCustomerTransactionSummary(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer transaction summary: %w", err)
	}

	// Add customer details
	summary["customer_name"] = customer.Name
	summary["customer_email"] = customer.Email
	summary["customer_since"] = customer.CreatedAt

	return summary, nil
}

// GetRevenueAnalytics gets detailed revenue analytics
func (uc *TransactionUseCase) GetRevenueAnalytics(ctx context.Context, days int) (map[string]any, error) {
	if days <= 0 {
		days = 30 // Default to last 30 days
	}

	// Get daily revenue
	dailyRevenue, err := uc.transactionRepo.GetDailyRevenue(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily revenue: %w", err)
	}

	// Get monthly revenue
	monthlyRevenue, err := uc.transactionRepo.GetMonthlyRevenue(ctx, 12) // Last 12 months
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly revenue: %w", err)
	}

	// Get revenue growth
	growth, err := uc.transactionRepo.GetRevenueGrowth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue growth: %w", err)
	}

	return map[string]any{
		"daily_revenue":   dailyRevenue,
		"monthly_revenue": monthlyRevenue,
		"growth":          growth,
	}, nil
}

// enrichTransaction adds related customer and product data to transaction
func (uc *TransactionUseCase) enrichTransaction(ctx context.Context, transaction *entities.Transaction) error {
	// Get customer data
	if transaction.CustomerID != "" {
		customer, err := uc.customerRepo.GetByID(ctx, transaction.CustomerID)
		if err == nil {
			transaction.Customer = customer
		}
	}

	// Get product data
	if transaction.ProductID != "" {
		product, err := uc.productRepo.GetByID(ctx, transaction.ProductID)
		if err == nil {
			transaction.Product = product
		}
	}

	return nil
}

// TransactionFilters represents filters for transaction queries
type TransactionFilters struct {
	CustomerID string     `json:"customer_id,omitempty"`
	ProductID  string     `json:"product_id,omitempty"`
	Type       string     `json:"type,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Offset     int        `json:"offset,omitempty"`
}

// StatsPeriod represents different time periods for statistics
type StatsPeriod string

const (
	StatsPeriodToday     StatsPeriod = "today"
	StatsPeriodThisWeek  StatsPeriod = "this_week"
	StatsPeriodThisMonth StatsPeriod = "this_month"
	StatsPeriodAllTime   StatsPeriod = "all_time"
)
