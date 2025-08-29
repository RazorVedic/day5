package repositories

import (
	"context"
	"fmt"
	"sync"

	"day5/internal/domain/entities"
	"day5/internal/domain/repositories"
	"day5/internal/infrastructure/persistence"

	"gorm.io/gorm"
)

// ProductRepositoryImpl implements the ProductRepository interface
type ProductRepositoryImpl struct {
	db *gorm.DB
	mu sync.RWMutex // Thread safety
}

// NewProductRepository creates a new product repository implementation
func NewProductRepository(db *gorm.DB) repositories.ProductRepository {
	return &ProductRepositoryImpl{
		db: db,
	}
}

// Create creates a new product with thread safety
func (r *ProductRepositoryImpl) Create(ctx context.Context, product *entities.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	model := persistence.ProductToModel(product)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Update the entity with generated fields
	persistence.ModelToProduct(model, product)
	return nil
}

// GetByID retrieves a product by ID with thread safety
func (r *ProductRepositoryImpl) GetByID(ctx context.Context, id string) (*entities.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var model persistence.Product
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	product := &entities.Product{}
	persistence.ModelToProduct(&model, product)
	return product, nil
}

// GetAll retrieves all products with pagination and thread safety
func (r *ProductRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Product
	query := r.db.WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	products := make([]*entities.Product, len(models))
	for i, model := range models {
		products[i] = &entities.Product{}
		persistence.ModelToProduct(&model, products[i])
	}

	return products, nil
}

// Update updates a product with thread safety
func (r *ProductRepositoryImpl) Update(ctx context.Context, product *entities.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	model := persistence.ProductToModel(product)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	persistence.ModelToProduct(model, product)
	return nil
}

// Delete deletes a product with thread safety
func (r *ProductRepositoryImpl) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.db.WithContext(ctx).Delete(&persistence.Product{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("product with ID %s not found", id)
	}

	return nil
}

// GetAvailableProducts gets products with quantity > 0
func (r *ProductRepositoryImpl) GetAvailableProducts(ctx context.Context) ([]*entities.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Product
	if err := r.db.WithContext(ctx).Where("quantity > 0").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get available products: %w", err)
	}

	products := make([]*entities.Product, len(models))
	for i, model := range models {
		products[i] = &entities.Product{}
		persistence.ModelToProduct(&model, products[i])
	}

	return products, nil
}

// GetByPriceRange gets products within a price range
func (r *ProductRepositoryImpl) GetByPriceRange(ctx context.Context, minPrice, maxPrice float64) ([]*entities.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Product
	if err := r.db.WithContext(ctx).Where("price BETWEEN ? AND ?", minPrice, maxPrice).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get products by price range: %w", err)
	}

	products := make([]*entities.Product, len(models))
	for i, model := range models {
		products[i] = &entities.Product{}
		persistence.ModelToProduct(&model, products[i])
	}

	return products, nil
}

// GetLowStockProducts gets products with quantity below threshold
func (r *ProductRepositoryImpl) GetLowStockProducts(ctx context.Context, threshold int) ([]*entities.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Product
	if err := r.db.WithContext(ctx).Where("quantity < ?", threshold).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	products := make([]*entities.Product, len(models))
	for i, model := range models {
		products[i] = &entities.Product{}
		persistence.ModelToProduct(&model, products[i])
	}

	return products, nil
}

// ReduceQuantity reduces product quantity atomically
func (r *ProductRepositoryImpl) ReduceQuantity(ctx context.Context, productID string, quantity int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.db.WithContext(ctx).Model(&persistence.Product{}).
		Where("id = ? AND quantity >= ?", productID, quantity).
		Update("quantity", gorm.Expr("quantity - ?", quantity))

	if result.Error != nil {
		return fmt.Errorf("failed to reduce quantity: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("insufficient quantity or product not found")
	}

	return nil
}

// IncreaseQuantity increases product quantity atomically
func (r *ProductRepositoryImpl) IncreaseQuantity(ctx context.Context, productID string, quantity int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.db.WithContext(ctx).Model(&persistence.Product{}).
		Where("id = ?", productID).
		Update("quantity", gorm.Expr("quantity + ?", quantity))

	if result.Error != nil {
		return fmt.Errorf("failed to increase quantity: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// SearchByName searches products by name
func (r *ProductRepositoryImpl) SearchByName(ctx context.Context, name string) ([]*entities.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []persistence.Product
	searchPattern := "%" + name + "%"
	if err := r.db.WithContext(ctx).Where("product_name LIKE ?", searchPattern).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to search products by name: %w", err)
	}

	products := make([]*entities.Product, len(models))
	for i, model := range models {
		products[i] = &entities.Product{}
		persistence.ModelToProduct(&model, products[i])
	}

	return products, nil
}

// GetTotalValue calculates total inventory value
func (r *ProductRepositoryImpl) GetTotalValue(ctx context.Context) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var totalValue float64
	if err := r.db.WithContext(ctx).Model(&persistence.Product{}).
		Select("SUM(price * quantity)").Scan(&totalValue).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate total value: %w", err)
	}

	return totalValue, nil
}

// Count returns the total number of products
func (r *ProductRepositoryImpl) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	if err := r.db.WithContext(ctx).Model(&persistence.Product{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return int(count), nil
}
