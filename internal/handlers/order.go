package handlers

import (
	"fmt"
	"net/http"

	"day5/internal/database"
	"day5/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct{}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{}
}

// PlaceOrder handles POST /order
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req models.OrderRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	db := database.GetDB()

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if customer exists
	var customer models.Customer
	if err := db.First(&customer, "id = ?", req.CustomerID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Customer with ID %v not found", req.CustomerID),
		})
		return
	}

	// Check cooldown period
	var cooldown models.CustomerCooldown
	err := db.Where("customer_id = ?", req.CustomerID).First(&cooldown).Error
	if err == nil {
		// Cooldown record exists, check if enough time has passed
		if !cooldown.CanPlaceOrder() {
			remaining := cooldown.RemainingCooldown()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":                      "Customer is in cooldown period",
				"cooldown_remaining_seconds": int(remaining.Seconds()),
				"cooldown_remaining_minutes": fmt.Sprintf("%.1f", remaining.Minutes()),
			})
			return
		}
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check cooldown",
		})
		return
	}

	// Check if product exists and has sufficient quantity
	var product models.Product
	if err := db.First(&product, "id = ?", req.ProductID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Product with ID %v not found", req.ProductID),
		})
		return
	}

	if product.Quantity < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":              "Insufficient product quantity",
			"available_quantity": product.Quantity,
			"requested_quantity": req.Quantity,
		})
		return
	}

	// Create order
	order := models.Order{
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		UnitPrice:  product.Price,
		Status:     models.OrderStatusCompleted,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create order",
			"details": err.Error(),
		})
		return
	}

	// Update product quantity
	product.Quantity -= req.Quantity
	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update product quantity",
		})
		return
	}

	// Create transaction record
	transaction := models.Transaction{
		OrderID:     order.ID,
		CustomerID:  req.CustomerID,
		ProductID:   req.ProductID,
		Type:        models.TransactionTypeOrder,
		Amount:      order.TotalPrice,
		Quantity:    req.Quantity,
		Description: fmt.Sprintf("Order for %s (x%d)", product.ProductName, req.Quantity),
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create transaction record",
		})
		return
	}

	// Update or create cooldown record
	cooldown.CustomerID = req.CustomerID
	cooldown.UpdateLastOrderTime()
	if err := tx.Save(&cooldown).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update cooldown",
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to complete order",
		})
		return
	}

	// Load relationships for response
	db.Preload("Customer").Preload("Product").First(&order, order.ID)
	order.Customer = customer
	order.Product = product

	// Return success response
	response := order.ToResponse("Order successfully placed")
	c.JSON(http.StatusCreated, response)
}

// GetOrderHistory handles GET /orders/customer/:customer_id
func (h *OrderHandler) GetOrderHistory(c *gin.Context) {
	customerID := c.Param("customer_id")

	// Verify customer exists
	var customer models.Customer
	if err := database.GetDB().First(&customer, "id = ?", customerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Customer not found",
		})
		return
	}

	var orders []models.Order
	if err := database.GetDB().
		Preload("Customer").
		Preload("Product").
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch order history",
		})
		return
	}

	var responses []models.OrderResponse
	for _, order := range orders {
		responses = append(responses, order.ToResponse(""))
	}

	c.JSON(http.StatusOK, models.OrderHistoryResponse{
		Orders: responses,
		Count:  len(responses),
	})
}

// GetAllOrders handles GET /orders (for retailer)
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	var orders []models.Order
	if err := database.GetDB().
		Preload("Customer").
		Preload("Product").
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch orders",
		})
		return
	}

	var responses []models.OrderResponse
	for _, order := range orders {
		responses = append(responses, order.ToResponse(""))
	}

	c.JSON(http.StatusOK, models.OrderHistoryResponse{
		Orders: responses,
		Count:  len(responses),
	})
}
