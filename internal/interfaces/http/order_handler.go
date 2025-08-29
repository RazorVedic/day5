package http

import (
	"net/http"
	"strconv"

	"day5/internal/application/usecases"
	"day5/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

// OrderHandler handles HTTP requests for order operations
type OrderHandler struct {
	orderUseCase *usecases.OrderUseCase
}

// NewOrderHandler creates a new order handler with dependency injection
func NewOrderHandler(orderUseCase *usecases.OrderUseCase) *OrderHandler {
	return &OrderHandler{
		orderUseCase: orderUseCase,
	}
}

// OrderResponse represents the HTTP response for order operations
type OrderResponse struct {
	ID           string  `json:"id"`
	CustomerID   string  `json:"customer_id"`
	CustomerName string  `json:"customer_name,omitempty"`
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name,omitempty"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	TotalAmount  float64 `json:"total_amount"`
	OrderDate    string  `json:"order_date"`
	CreatedAt    string  `json:"created_at"`
	Message      string  `json:"message,omitempty"`
}

// OrderHistoryResponse represents the response for order history
type OrderHistoryResponse struct {
	Orders  []*OrderResponse `json:"orders"`
	Count   int              `json:"count"`
	Message string           `json:"message,omitempty"`
}

// PlaceOrder handles POST /api/v1/order
// @Summary Place a new order
// @Description Places an order with cooldown validation and inventory management
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body usecases.PlaceOrderRequest true "Order details"
// @Success 201 {object} OrderResponse
// @Failure 400 {object} map[string]any
// @Failure 429 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/order [post]
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req usecases.PlaceOrderRequest

	// Struct tags used:
	// - json:"customer_id" - JSON field mapping for request/response
	// - binding:"required" - ensures field is present in request
	// - binding:"gt=0" - validates that numeric values are greater than 0
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Call use case (business logic layer)
	orderResponse, err := h.orderUseCase.PlaceOrder(c.Request.Context(), &req)
	if err != nil {
		// Handle cooldown errors specifically
		if cooldownErr, ok := err.(*usecases.CooldownError); ok {
			response := gin.H{
				"error":                      "Customer is in cooldown period",
				"cooldown_remaining_seconds": int(cooldownErr.RemainingTime.Seconds()),
				"cooldown_remaining_minutes": cooldownErr.RemainingTime.Minutes(),
			}
			// Add full cooldown status if available
			if cooldownErr.CooldownStatus != nil {
				for k, v := range cooldownErr.CooldownStatus {
					response[k] = v
				}
			}
			c.JSON(http.StatusTooManyRequests, response)
			return
		}

		// Handle other business logic errors
		if err.Error() == "insufficient quantity" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Insufficient product quantity",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to place order",
			"details": err.Error(),
		})
		return
	}

	// Convert use case response to HTTP response
	response := &OrderResponse{
		ID:           orderResponse.ID,
		CustomerID:   orderResponse.CustomerID,
		CustomerName: orderResponse.CustomerName,
		ProductID:    orderResponse.ProductID,
		ProductName:  orderResponse.ProductName,
		Quantity:     orderResponse.Quantity,
		UnitPrice:    orderResponse.UnitPrice,
		TotalAmount:  orderResponse.TotalAmount,
		OrderDate:    orderResponse.OrderDate.Format("2006-01-02T15:04:05Z"),
		Message:      orderResponse.Message,
	}

	c.JSON(http.StatusCreated, response)
}

// GetOrderHistory handles GET /api/v1/orders/customer/:customer_id
// @Summary Get customer order history
// @Description Retrieves order history for a specific customer
// @Tags Orders
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} OrderHistoryResponse
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/orders/customer/{customer_id} [get]
func (h *OrderHandler) GetOrderHistory(c *gin.Context) {
	customerID := c.Param("customer_id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Customer ID is required",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	orders, err := h.orderUseCase.GetOrderHistory(c.Request.Context(), customerID, limit, offset)
	if err != nil {
		if err.Error() == "customer not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Customer not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve order history",
			"details": err.Error(),
		})
		return
	}

	orderResponses := make([]*OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = h.entityToResponse(order, "")
	}

	response := &OrderHistoryResponse{
		Orders:  orderResponses,
		Count:   len(orderResponses),
		Message: "Order history retrieved successfully",
	}

	c.JSON(http.StatusOK, response)
}

// GetAllOrders handles GET /api/v1/orders
// @Summary Get all orders (retailer view)
// @Description Retrieves all orders with pagination for retailer analytics
// @Tags Orders
// @Produce json
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} OrderHistoryResponse
// @Failure 500 {object} map[string]any
// @Router /api/v1/orders [get]
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	orders, err := h.orderUseCase.GetAllOrders(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve orders",
			"details": err.Error(),
		})
		return
	}

	orderResponses := make([]*OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = h.entityToResponse(order, "")
	}

	response := &OrderHistoryResponse{
		Orders:  orderResponses,
		Count:   len(orderResponses),
		Message: "Orders retrieved successfully",
	}

	c.JSON(http.StatusOK, response)
}

// GetOrder handles GET /api/v1/order/:id
// @Summary Get an order by ID
// @Description Retrieves a specific order by its ID
// @Tags Orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} OrderResponse
// @Failure 404 {object} map[string]any
// @Router /api/v1/order/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Order ID is required",
		})
		return
	}

	order, err := h.orderUseCase.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Order not found",
		})
		return
	}

	response := h.entityToResponse(order, "")
	c.JSON(http.StatusOK, response)
}

// GetTodaysOrders handles GET /api/v1/orders/today
// @Summary Get today's orders
// @Description Retrieves all orders placed today
// @Tags Orders
// @Produce json
// @Success 200 {object} OrderHistoryResponse
// @Failure 500 {object} map[string]any
// @Router /api/v1/orders/today [get]
func (h *OrderHandler) GetTodaysOrders(c *gin.Context) {
	orders, err := h.orderUseCase.GetTodaysOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve today's orders",
			"details": err.Error(),
		})
		return
	}

	orderResponses := make([]*OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = h.entityToResponse(order, "")
	}

	response := &OrderHistoryResponse{
		Orders:  orderResponses,
		Count:   len(orderResponses),
		Message: "Today's orders retrieved successfully",
	}

	c.JSON(http.StatusOK, response)
}

// Helper method to convert domain entity to HTTP response
func (h *OrderHandler) entityToResponse(order *entities.Order, message string) *OrderResponse {
	response := &OrderResponse{
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		ProductID:   order.ProductID,
		Quantity:    order.Quantity,
		UnitPrice:   order.UnitPrice,
		TotalAmount: order.TotalAmount,
		OrderDate:   order.OrderDate.Format("2006-01-02T15:04:05Z"),
		CreatedAt:   order.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Message:     message,
	}

	// Add related entity information if available
	if order.Customer != nil {
		response.CustomerName = order.Customer.Name
	}
	if order.Product != nil {
		response.ProductName = order.Product.ProductName
	}

	return response
}
