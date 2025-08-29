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

// OrderUseCase encapsulates business logic for order operations
type OrderUseCase struct {
	orderRepo       repositories.OrderRepository
	customerUseCase *CustomerUseCase
	productUseCase  *ProductUseCase
	transactionRepo repositories.TransactionRepository
}

// NewOrderUseCase creates a new order use case
func NewOrderUseCase(
	orderRepo repositories.OrderRepository,
	customerUseCase *CustomerUseCase,
	productUseCase *ProductUseCase,
	transactionRepo repositories.TransactionRepository,
) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:       orderRepo,
		customerUseCase: customerUseCase,
		productUseCase:  productUseCase,
		transactionRepo: transactionRepo,
	}
}

// PlaceOrderRequest represents the request to place an order
type PlaceOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ProductID  string `json:"product_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,gt=0"`
}

// OrderResponse represents the response after placing an order
type OrderResponse struct {
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	ProductID    string    `json:"product_id"`
	ProductName  string    `json:"product_name"`
	Quantity     int       `json:"quantity"`
	UnitPrice    float64   `json:"unit_price"`
	TotalAmount  float64   `json:"total_amount"`
	OrderDate    time.Time `json:"order_date"`
	Message      string    `json:"message"`
}

// PlaceOrder places a new order with complete business logic validation
func (uc *OrderUseCase) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*OrderResponse, error) {
	// Step 1: Validate customer cooldown
	canOrder, cooldown, err := uc.customerUseCase.CanCustomerPlaceOrder(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check customer cooldown: %w", err)
	}

	if !canOrder {
		remaining := cooldown.RemainingCooldown(uc.customerUseCase.cooldownPeriod)
		return nil, &CooldownError{
			CustomerID:     req.CustomerID,
			RemainingTime:  remaining,
			CooldownStatus: cooldown.GetCooldownStatus(uc.customerUseCase.cooldownPeriod),
		}
	}

	// Step 2: Check product availability
	product, err := uc.productUseCase.CheckProductAvailability(ctx, req.ProductID, req.Quantity)
	if err != nil {
		return nil, fmt.Errorf("product availability check failed: %w", err)
	}

	// Step 3: Get customer details
	customer, err := uc.customerUseCase.GetCustomer(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Step 4: Create order entity
	orderID, err := generateOrderID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate order ID: %w", err)
	}

	order := &entities.Order{
		ID:         orderID,
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		UnitPrice:  product.Price,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		Customer:   customer,
		Product:    product,
	}

	order.CalculateTotal()
	order.SetOrderDate()

	// Validate order business rules
	if err := order.Validate(); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// Step 5: Execute transaction (all or nothing)
	if err := uc.executeOrderTransaction(ctx, order, product, req.Quantity); err != nil {
		return nil, fmt.Errorf("failed to execute order transaction: %w", err)
	}

	// Step 6: Create response
	response := &OrderResponse{
		ID:           order.ID,
		CustomerID:   order.CustomerID,
		CustomerName: customer.Name,
		ProductID:    order.ProductID,
		ProductName:  product.ProductName,
		Quantity:     order.Quantity,
		UnitPrice:    order.UnitPrice,
		TotalAmount:  order.TotalAmount,
		OrderDate:    order.OrderDate,
		Message:      "Order successfully placed",
	}

	return response, nil
}

// executeOrderTransaction handles the complete order transaction
func (uc *OrderUseCase) executeOrderTransaction(ctx context.Context, order *entities.Order, product *entities.Product, quantity int) error {
	// This should be wrapped in a database transaction
	// For now, we'll handle the sequence manually

	// 1. Reduce product quantity
	if err := product.ReduceQuantity(quantity); err != nil {
		return fmt.Errorf("failed to reduce product quantity: %w", err)
	}

	// 2. Save order
	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// 3. Update product in repository
	if err := uc.productUseCase.productRepo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to update product quantity: %w", err)
	}

	// 4. Create transaction record
	transactionID, err := generateTransactionID()
	if err != nil {
		return fmt.Errorf("failed to generate transaction ID: %w", err)
	}

	transaction := &entities.Transaction{
		ID:        transactionID,
		CreatedAt: time.Now().UTC(),
	}
	transaction.CreateFromOrder(order)

	if err := uc.transactionRepo.Create(ctx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// 5. Update customer cooldown
	if err := uc.customerUseCase.UpdateCustomerCooldown(ctx, order.CustomerID); err != nil {
		return fmt.Errorf("failed to update customer cooldown: %w", err)
	}

	return nil
}

// GetOrderHistory gets order history for a customer
func (uc *OrderUseCase) GetOrderHistory(ctx context.Context, customerID string, limit, offset int) ([]*entities.Order, error) {
	if customerID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}

	// Verify customer exists
	_, err := uc.customerUseCase.GetCustomer(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := uc.orderRepo.GetByCustomerID(ctx, customerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get order history: %w", err)
	}

	return orders, nil
}

// GetAllOrders gets all orders with pagination (for admin/retailer view)
func (uc *OrderUseCase) GetAllOrders(ctx context.Context, limit, offset int) ([]*entities.Order, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := uc.orderRepo.GetOrdersWithDetails(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all orders: %w", err)
	}

	return orders, nil
}

// GetOrder gets a specific order by ID
func (uc *OrderUseCase) GetOrder(ctx context.Context, orderID string) (*entities.Order, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// GetTodaysOrders gets all orders placed today
func (uc *OrderUseCase) GetTodaysOrders(ctx context.Context) ([]*entities.Order, error) {
	orders, err := uc.orderRepo.GetTodaysOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's orders: %w", err)
	}

	return orders, nil
}

// CooldownError represents a cooldown violation error
type CooldownError struct {
	CustomerID     string
	RemainingTime  time.Duration
	CooldownStatus map[string]any
}

func (e *CooldownError) Error() string {
	return fmt.Sprintf("customer %s is in cooldown period, remaining: %v",
		e.CustomerID, e.RemainingTime)
}

// generateOrderID generates a unique order ID in format ORD12345
func generateOrderID() (string, error) {
	max := big.NewInt(99999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	number := n.Int64() + 10000
	if number > 99999 {
		number = number%90000 + 10000
	}

	return fmt.Sprintf("ORD%05d", number), nil
}

// generateTransactionID generates a unique transaction ID in format TXN12345
func generateTransactionID() (string, error) {
	max := big.NewInt(99999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	number := n.Int64() + 10000
	if number > 99999 {
		number = number%90000 + 10000
	}

	return fmt.Sprintf("TXN%05d", number), nil
}
