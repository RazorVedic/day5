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

// ProductUseCase encapsulates business logic for product operations
type ProductUseCase struct {
	productRepo repositories.ProductRepository
}

// NewProductUseCase creates a new product use case
func NewProductUseCase(productRepo repositories.ProductRepository) *ProductUseCase {
	return &ProductUseCase{
		productRepo: productRepo,
	}
}

// CreateProductRequest represents the request to create a product
type CreateProductRequest struct {
	ProductName string  `json:"product_name" binding:"required"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Quantity    int     `json:"quantity" binding:"required,gte=0"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	Price    *float64 `json:"price,omitempty" binding:"omitempty,gt=0"`
	Quantity *int     `json:"quantity,omitempty" binding:"omitempty,gte=0"`
}

// CreateProduct creates a new product
func (uc *ProductUseCase) CreateProduct(ctx context.Context, req *CreateProductRequest) (*entities.Product, error) {
	// Generate unique product ID
	id, err := generateProductID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate product ID: %w", err)
	}

	// Create product entity
	product := &entities.Product{
		ID:          id,
		ProductName: req.ProductName,
		Price:       req.Price,
		Quantity:    req.Quantity,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Validate business rules
	if err := product.Validate(); err != nil {
		return nil, fmt.Errorf("product validation failed: %w", err)
	}

	// Save to repository
	if err := uc.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// GetProduct retrieves a product by ID
func (uc *ProductUseCase) GetProduct(ctx context.Context, id string) (*entities.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("product ID is required")
	}

	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// GetAllProducts retrieves all products with pagination
func (uc *ProductUseCase) GetAllProducts(ctx context.Context, limit, offset int) ([]*entities.Product, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	products, err := uc.productRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	return products, nil
}

// UpdateProduct updates a product's price and/or quantity
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, id string, req *UpdateProductRequest) (*entities.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("product ID is required")
	}

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Update fields if provided
	if req.Price != nil {
		if err := product.UpdatePrice(*req.Price); err != nil {
			return nil, fmt.Errorf("failed to update price: %w", err)
		}
	}

	if req.Quantity != nil {
		if err := product.UpdateQuantity(*req.Quantity); err != nil {
			return nil, fmt.Errorf("failed to update quantity: %w", err)
		}
	}

	// Validate after updates
	if err := product.Validate(); err != nil {
		return nil, fmt.Errorf("product validation failed: %w", err)
	}

	// Save to repository
	if err := uc.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// GetAvailableProducts gets products that have quantity > 0
func (uc *ProductUseCase) GetAvailableProducts(ctx context.Context) ([]*entities.Product, error) {
	products, err := uc.productRepo.GetAvailableProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available products: %w", err)
	}

	return products, nil
}

// CheckProductAvailability checks if a product has sufficient quantity
func (uc *ProductUseCase) CheckProductAvailability(ctx context.Context, productID string, requestedQuantity int) (*entities.Product, error) {
	product, err := uc.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if !product.IsAvailable(requestedQuantity) {
		return product, fmt.Errorf("insufficient quantity: available=%d, requested=%d",
			product.Quantity, requestedQuantity)
	}

	return product, nil
}

// SearchProducts searches products by name
func (uc *ProductUseCase) SearchProducts(ctx context.Context, name string) ([]*entities.Product, error) {
	if name == "" {
		return nil, fmt.Errorf("search name is required")
	}

	products, err := uc.productRepo.SearchByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return products, nil
}

// GetLowStockProducts gets products with quantity below threshold
func (uc *ProductUseCase) GetLowStockProducts(ctx context.Context, threshold int) ([]*entities.Product, error) {
	if threshold < 0 {
		threshold = 5 // Default threshold
	}

	products, err := uc.productRepo.GetLowStockProducts(ctx, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	return products, nil
}

// generateProductID generates a unique product ID in format PROD12345
func generateProductID() (string, error) {
	max := big.NewInt(99999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	number := n.Int64() + 10000
	if number > 99999 {
		number = number%90000 + 10000
	}

	return fmt.Sprintf("PROD%05d", number), nil
}
