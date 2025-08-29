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

// OrderRepositoryImpl implements the OrderRepository interface
type OrderRepositoryImpl struct {
	db *gorm.DB
	mu sync.RWMutex
}

// NewOrderRepository creates a new order repository implementation
func NewOrderRepository(db *gorm.DB) repositories.OrderRepository {
	return &OrderRepositoryImpl{
		db: db,
	}
}

// Create creates a new order with thread safety
func (r *OrderRepositoryImpl) Create(ctx context.Context, order *entities.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	model := persistence.OrderToModel(order)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	persistence.ModelToOrder(model, order)
	return nil
}

// GetByID retrieves an order by ID
func (r *OrderRepositoryImpl) GetByID(ctx context.Context, id string) (*entities.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var model persistence.Order
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Product").
		First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	order := &entities.Order{}
	persistence.ModelToOrder(&model, order)
	return order, nil
}

// GetAll retrieves all orders with pagination
func (r *OrderRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Order
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Product").
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	return persistence.ModelsToOrders(models), nil
}

// Update updates an order
func (r *OrderRepositoryImpl) Update(ctx context.Context, order *entities.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	model := persistence.OrderToModel(order)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	persistence.ModelToOrder(model, order)
	return nil
}

// Delete deletes an order
func (r *OrderRepositoryImpl) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.db.WithContext(ctx).Delete(&persistence.Order{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("order with ID %s not found", id)
	}

	return nil
}

// GetByCustomerID gets orders for a specific customer
func (r *OrderRepositoryImpl) GetByCustomerID(ctx context.Context, customerID string, limit, offset int) ([]*entities.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Order
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Product").
		Where("customer_id = ?", customerID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get orders by customer: %w", err)
	}

	return persistence.ModelsToOrders(models), nil
}

// GetCustomerOrderCount gets the total number of orders for a customer
func (r *OrderRepositoryImpl) GetCustomerOrderCount(ctx context.Context, customerID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	if err := r.db.WithContext(ctx).Model(&persistence.Order{}).
		Where("customer_id = ?", customerID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count customer orders: %w", err)
	}

	return int(count), nil
}

// GetByProductID gets orders for a specific product
func (r *OrderRepositoryImpl) GetByProductID(ctx context.Context, productID string, limit, offset int) ([]*entities.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Order
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Product").
		Where("product_id = ?", productID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get orders by product: %w", err)
	}

	return persistence.ModelsToOrders(models), nil
}

// GetByDateRange gets orders within a date range
func (r *OrderRepositoryImpl) GetByDateRange(ctx context.Context, start, end time.Time) ([]*entities.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Order
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Product").
		Where("order_date BETWEEN ? AND ?", start, end).
		Order("order_date DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get orders by date range: %w", err)
	}

	return persistence.ModelsToOrders(models), nil
}

// GetTodaysOrders gets all orders placed today
func (r *OrderRepositoryImpl) GetTodaysOrders(ctx context.Context) ([]*entities.Order, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	return r.GetByDateRange(ctx, startOfDay, endOfDay)
}

// GetRecentOrders gets orders from the last N hours
func (r *OrderRepositoryImpl) GetRecentOrders(ctx context.Context, hours int) ([]*entities.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	var models []persistence.Order
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Product").
		Where("order_date >= ?", since).
		Order("order_date DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent orders: %w", err)
	}

	return persistence.ModelsToOrders(models), nil
}

// GetOrdersWithDetails gets orders with full customer and product details
func (r *OrderRepositoryImpl) GetOrdersWithDetails(ctx context.Context, limit, offset int) ([]*entities.Order, error) {
	return r.GetAll(ctx, limit, offset)
}

// GetTotalRevenue calculates total revenue for a period
func (r *OrderRepositoryImpl) GetTotalRevenue(ctx context.Context, start, end *time.Time) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var totalRevenue float64
	query := r.db.WithContext(ctx).Model(&persistence.Order{}).
		Select("SUM(total_amount)")
	
	if start != nil && end != nil {
		query = query.Where("order_date BETWEEN ? AND ?", *start, *end)
	}

	if err := query.Scan(&totalRevenue).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate total revenue: %w", err)
	}

	return totalRevenue, nil
}

// GetOrderCountByPeriod gets order count for a specific period
func (r *OrderRepositoryImpl) GetOrderCountByPeriod(ctx context.Context, start, end time.Time) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	if err := r.db.WithContext(ctx).Model(&persistence.Order{}).
		Where("order_date BETWEEN ? AND ?", start, end).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count orders by period: %w", err)
	}

	return int(count), nil
}

// Count returns the total number of orders
func (r *OrderRepositoryImpl) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	if err := r.db.WithContext(ctx).Model(&persistence.Order{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}

	return int(count), nil
}

// GetAverageOrderValue calculates the average order value
func (r *OrderRepositoryImpl) GetAverageOrderValue(ctx context.Context) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var avgValue float64
	if err := r.db.WithContext(ctx).Model(&persistence.Order{}).
		Select("AVG(total_amount)").Scan(&avgValue).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate average order value: %w", err)
	}

	return avgValue, nil
}
