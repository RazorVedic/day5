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

// TransactionRepositoryImpl implements the TransactionRepository interface
type TransactionRepositoryImpl struct {
	db *gorm.DB
	mu sync.RWMutex // Thread safety
}

// NewTransactionRepository creates a new transaction repository implementation
func NewTransactionRepository(db *gorm.DB) repositories.TransactionRepository {
	return &TransactionRepositoryImpl{
		db: db,
	}
}

// Create creates a new transaction with thread safety
func (r *TransactionRepositoryImpl) Create(ctx context.Context, transaction *entities.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	model := persistence.TransactionToModel(transaction)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update the entity with generated fields
	persistence.ModelToTransaction(model, transaction)
	return nil
}

// GetByID retrieves a transaction by ID with thread safety
func (r *TransactionRepositoryImpl) GetByID(ctx context.Context, id string) (*entities.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var model persistence.Transaction
	if err := r.db.WithContext(ctx).Preload("Order").Preload("Customer").Preload("Product").
		First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	transaction := &entities.Transaction{}
	persistence.ModelToTransaction(&model, transaction)
	return transaction, nil
}

// GetAll retrieves all transactions with pagination and thread safety
func (r *TransactionRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Transaction
	query := r.db.WithContext(ctx).Preload("Order").Preload("Customer").Preload("Product").
		Order("transaction_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	transactions := make([]*entities.Transaction, len(models))
	for i, model := range models {
		transactions[i] = &entities.Transaction{}
		persistence.ModelToTransaction(&model, transactions[i])
	}

	return transactions, nil
}

// Update updates a transaction with thread safety
func (r *TransactionRepositoryImpl) Update(ctx context.Context, transaction *entities.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	model := persistence.TransactionToModel(transaction)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	persistence.ModelToTransaction(model, transaction)
	return nil
}

// Delete deletes a transaction with thread safety
func (r *TransactionRepositoryImpl) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.db.WithContext(ctx).Delete(&persistence.Transaction{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete transaction: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("transaction with ID %s not found", id)
	}

	return nil
}

// GetByCustomerID retrieves transactions by customer ID
func (r *TransactionRepositoryImpl) GetByCustomerID(ctx context.Context, customerID string, limit, offset int) ([]*entities.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Transaction
	query := r.db.WithContext(ctx).Preload("Order").Preload("Customer").Preload("Product").
		Where("customer_id = ?", customerID).Order("transaction_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get transactions by customer ID: %w", err)
	}

	transactions := make([]*entities.Transaction, len(models))
	for i, model := range models {
		transactions[i] = &entities.Transaction{}
		persistence.ModelToTransaction(&model, transactions[i])
	}

	return transactions, nil
}

// GetByProductID retrieves transactions by product ID
func (r *TransactionRepositoryImpl) GetByProductID(ctx context.Context, productID string, limit, offset int) ([]*entities.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Transaction
	query := r.db.WithContext(ctx).Preload("Order").Preload("Customer").Preload("Product").
		Where("product_id = ?", productID).Order("transaction_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get transactions by product ID: %w", err)
	}

	transactions := make([]*entities.Transaction, len(models))
	for i, model := range models {
		transactions[i] = &entities.Transaction{}
		persistence.ModelToTransaction(&model, transactions[i])
	}

	return transactions, nil
}

// GetByOrderID retrieves transaction by order ID
func (r *TransactionRepositoryImpl) GetByOrderID(ctx context.Context, orderID string) (*entities.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var model persistence.Transaction
	if err := r.db.WithContext(ctx).Preload("Order").Preload("Customer").Preload("Product").
		First(&model, "order_id = ?", orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction with order ID %s not found", orderID)
		}
		return nil, fmt.Errorf("failed to get transaction by order ID: %w", err)
	}

	transaction := &entities.Transaction{}
	persistence.ModelToTransaction(&model, transaction)
	return transaction, nil
}

// GetByType retrieves transactions by type
func (r *TransactionRepositoryImpl) GetByType(ctx context.Context, transactionType entities.TransactionType, limit, offset int) ([]*entities.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Transaction
	query := r.db.WithContext(ctx).Preload("Order").Preload("Customer").Preload("Product").
		Where("type = ?", string(transactionType)).Order("transaction_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get transactions by type: %w", err)
	}

	transactions := make([]*entities.Transaction, len(models))
	for i, model := range models {
		transactions[i] = &entities.Transaction{}
		persistence.ModelToTransaction(&model, transactions[i])
	}

	return transactions, nil
}

// GetByDateRange retrieves transactions within a date range
func (r *TransactionRepositoryImpl) GetByDateRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*entities.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Transaction
	query := r.db.WithContext(ctx).Preload("Order").Preload("Customer").Preload("Product").
		Where("transaction_at BETWEEN ? AND ?", start, end).Order("transaction_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get transactions by date range: %w", err)
	}

	transactions := make([]*entities.Transaction, len(models))
	for i, model := range models {
		transactions[i] = &entities.Transaction{}
		persistence.ModelToTransaction(&model, transactions[i])
	}

	return transactions, nil
}

// GetTodaysTransactions retrieves today's transactions
func (r *TransactionRepositoryImpl) GetTodaysTransactions(ctx context.Context) ([]*entities.Transaction, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return r.GetByDateRange(ctx, startOfDay, endOfDay, 0, 0)
}

// GetTransactionsByPeriod retrieves transactions for a specific period
func (r *TransactionRepositoryImpl) GetTransactionsByPeriod(ctx context.Context, start, end time.Time) ([]*entities.Transaction, error) {
	return r.GetByDateRange(ctx, start, end, 0, 0)
}

// GetBusinessStats calculates business statistics
func (r *TransactionRepositoryImpl) GetBusinessStats(ctx context.Context, start, end *time.Time) (*entities.BusinessStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var stats entities.BusinessStats
	query := r.db.WithContext(ctx).Model(&persistence.Transaction{}).Where("type = ?", "order")

	if start != nil && end != nil {
		query = query.Where("transaction_at BETWEEN ? AND ?", *start, *end)
	}

	// Get total revenue and order count
	var totalRevenue float64
	var orderCount int64

	if err := query.Select("COALESCE(SUM(amount), 0)").Scan(&totalRevenue).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total revenue: %w", err)
	}

	if err := query.Count(&orderCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	stats.TotalRevenue = totalRevenue
	stats.OrderCount = int(orderCount)
	stats.CalculateAverageOrderValue()

	// Get total quantity sold
	var totalQuantity int64
	if err := query.Select("COALESCE(SUM(quantity), 0)").Scan(&totalQuantity).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total quantity: %w", err)
	}
	stats.TotalQuantitySold = int(totalQuantity)

	// Get unique customers
	var uniqueCustomers int64
	if err := query.Select("COUNT(DISTINCT customer_id)").Scan(&uniqueCustomers).Error; err != nil {
		return nil, fmt.Errorf("failed to count unique customers: %w", err)
	}
	stats.UniqueCustomers = int(uniqueCustomers)

	return &stats, nil
}

// GetRevenueByPeriod calculates revenue for a specific period
func (r *TransactionRepositoryImpl) GetRevenueByPeriod(ctx context.Context, start, end time.Time) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var revenue float64
	if err := r.db.WithContext(ctx).Model(&persistence.Transaction{}).
		Where("type = ? AND transaction_at BETWEEN ? AND ?", "order", start, end).
		Select("COALESCE(SUM(amount), 0)").Scan(&revenue).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate revenue: %w", err)
	}

	return revenue, nil
}

// GetTopSellingProducts gets top selling products
func (r *TransactionRepositoryImpl) GetTopSellingProducts(ctx context.Context, limit int, start, end *time.Time) ([]*entities.ProductSales, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := r.db.WithContext(ctx).Table("transactions t").
		Select("t.product_id, p.product_name, SUM(t.quantity) as quantity_sold, SUM(t.amount) as total_revenue").
		Joins("JOIN products p ON t.product_id = p.id").
		Where("t.type = ?", "order").
		Group("t.product_id, p.product_name").
		Order("quantity_sold DESC")

	if start != nil && end != nil {
		query = query.Where("t.transaction_at BETWEEN ? AND ?", *start, *end)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	type productSalesResult struct {
		ProductID    string  `json:"product_id"`
		ProductName  string  `json:"product_name"`
		QuantitySold int     `json:"quantity_sold"`
		TotalRevenue float64 `json:"total_revenue"`
	}

	var results []productSalesResult
	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get top selling products: %w", err)
	}

	productSales := make([]*entities.ProductSales, len(results))
	for i, result := range results {
		productSales[i] = &entities.ProductSales{
			ProductID:    result.ProductID,
			ProductName:  result.ProductName,
			QuantitySold: result.QuantitySold,
			TotalRevenue: result.TotalRevenue,
		}
	}

	return productSales, nil
}

// GetCustomerTransactionSummary gets transaction summary for a customer
func (r *TransactionRepositoryImpl) GetCustomerTransactionSummary(ctx context.Context, customerID string) (map[string]any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summary := make(map[string]any)

	// Total transactions
	var totalTransactions int64
	if err := r.db.WithContext(ctx).Model(&persistence.Transaction{}).
		Where("customer_id = ?", customerID).Count(&totalTransactions).Error; err != nil {
		return nil, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Total amount spent
	var totalSpent float64
	if err := r.db.WithContext(ctx).Model(&persistence.Transaction{}).
		Where("customer_id = ? AND type = ?", customerID, "order").
		Select("COALESCE(SUM(amount), 0)").Scan(&totalSpent).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total spent: %w", err)
	}

	// First transaction date
	var firstTransaction time.Time
	if err := r.db.WithContext(ctx).Model(&persistence.Transaction{}).
		Where("customer_id = ?", customerID).
		Select("MIN(transaction_at)").Scan(&firstTransaction).Error; err != nil {
		return nil, fmt.Errorf("failed to get first transaction: %w", err)
	}

	// Last transaction date
	var lastTransaction time.Time
	if err := r.db.WithContext(ctx).Model(&persistence.Transaction{}).
		Where("customer_id = ?", customerID).
		Select("MAX(transaction_at)").Scan(&lastTransaction).Error; err != nil {
		return nil, fmt.Errorf("failed to get last transaction: %w", err)
	}

	summary["total_transactions"] = totalTransactions
	summary["total_spent"] = totalSpent
	summary["first_transaction"] = firstTransaction
	summary["last_transaction"] = lastTransaction

	return summary, nil
}

// Count returns the total number of transactions
func (r *TransactionRepositoryImpl) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	if err := r.db.WithContext(ctx).Model(&persistence.Transaction{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	return int(count), nil
}

// GetTotalRevenue calculates total revenue
func (r *TransactionRepositoryImpl) GetTotalRevenue(ctx context.Context) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var totalRevenue float64
	if err := r.db.WithContext(ctx).Model(&persistence.Transaction{}).
		Where("type = ?", "order").
		Select("COALESCE(SUM(amount), 0)").Scan(&totalRevenue).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate total revenue: %w", err)
	}

	return totalRevenue, nil
}

// GetTransactionCountByType returns count of transactions by type
func (r *TransactionRepositoryImpl) GetTransactionCountByType(ctx context.Context, transactionType entities.TransactionType) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	if err := r.db.WithContext(ctx).Model(&persistence.Transaction{}).
		Where("type = ?", string(transactionType)).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count transactions by type: %w", err)
	}

	return int(count), nil
}

// GetDailyRevenue gets daily revenue for the last N days
func (r *TransactionRepositoryImpl) GetDailyRevenue(ctx context.Context, days int) ([]map[string]any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	startDate := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)

	var results []map[string]any
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT DATE(transaction_at) as date, COALESCE(SUM(amount), 0) as revenue
		FROM transactions 
		WHERE type = 'order' AND transaction_at >= ?
		GROUP BY DATE(transaction_at)
		ORDER BY date DESC
	`, startDate).Rows()

	if err != nil {
		return nil, fmt.Errorf("failed to get daily revenue: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var date time.Time
		var revenue float64
		if err := rows.Scan(&date, &revenue); err != nil {
			return nil, fmt.Errorf("failed to scan daily revenue: %w", err)
		}
		results = append(results, map[string]any{
			"date":    date.Format("2006-01-02"),
			"revenue": revenue,
		})
	}

	return results, nil
}

// GetMonthlyRevenue gets monthly revenue for the last N months
func (r *TransactionRepositoryImpl) GetMonthlyRevenue(ctx context.Context, months int) ([]map[string]any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	startDate := time.Now().AddDate(0, -months, 0)

	var results []map[string]any
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT DATE_FORMAT(transaction_at, '%Y-%m') as month, COALESCE(SUM(amount), 0) as revenue
		FROM transactions 
		WHERE type = 'order' AND transaction_at >= ?
		GROUP BY DATE_FORMAT(transaction_at, '%Y-%m')
		ORDER BY month DESC
	`, startDate).Rows()

	if err != nil {
		return nil, fmt.Errorf("failed to get monthly revenue: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var month string
		var revenue float64
		if err := rows.Scan(&month, &revenue); err != nil {
			return nil, fmt.Errorf("failed to scan monthly revenue: %w", err)
		}
		results = append(results, map[string]any{
			"month":   month,
			"revenue": revenue,
		})
	}

	return results, nil
}

// GetRevenueGrowth calculates revenue growth
func (r *TransactionRepositoryImpl) GetRevenueGrowth(ctx context.Context) (map[string]any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get current month revenue
	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	currentMonthEnd := currentMonthStart.AddDate(0, 1, 0)

	currentRevenue, err := r.GetRevenueByPeriod(ctx, currentMonthStart, currentMonthEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get current month revenue: %w", err)
	}

	// Get previous month revenue
	previousMonthStart := currentMonthStart.AddDate(0, -1, 0)
	previousMonthEnd := currentMonthStart

	previousRevenue, err := r.GetRevenueByPeriod(ctx, previousMonthStart, previousMonthEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous month revenue: %w", err)
	}

	// Calculate growth percentage
	var growthPercentage float64
	if previousRevenue > 0 {
		growthPercentage = ((currentRevenue - previousRevenue) / previousRevenue) * 100
	}

	return map[string]any{
		"current_month_revenue":  currentRevenue,
		"previous_month_revenue": previousRevenue,
		"growth_percentage":      growthPercentage,
	}, nil
}
