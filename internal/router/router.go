package router

import (
	"net/http"

	"day5/internal/handlers"
	"day5/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter creates and configures the main application router
func SetupRouter() *gin.Engine {
	router := gin.New()

	// Apply middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "day5",
			"version": "1.0.0",
		})
	})

	// API routes
	setupAPIRoutes(router)

	return router
}

func setupAPIRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")

	// Initialize handlers
	productHandler := handlers.NewProductHandler()
	customerHandler := handlers.NewCustomerHandler()
	orderHandler := handlers.NewOrderHandler()
	transactionHandler := handlers.NewTransactionHandler()

	// === PRODUCT ROUTES (For Retailer) ===
	api.POST("/product", productHandler.CreateProduct)    // Add a product
	api.PUT("/product/:id", productHandler.UpdateProduct) // Update product price/quantity
	api.GET("/products", productHandler.GetProducts)      // View all products (also for customers)
	api.GET("/product/:id", productHandler.GetProduct)    // Get single product

	// === CUSTOMER ROUTES ===
	api.POST("/customer", customerHandler.CreateCustomer) // Register a customer
	api.GET("/customers", customerHandler.GetCustomers)   // List all customers (for retailer)
	api.GET("/customer/:id", customerHandler.GetCustomer) // Get single customer

	// === ORDER ROUTES ===
	api.POST("/order", orderHandler.PlaceOrder)                            // Customer places an order (with cooldown)
	api.GET("/orders/customer/:customer_id", orderHandler.GetOrderHistory) // Customer views order history
	api.GET("/orders", orderHandler.GetAllOrders)                          // Retailer views all orders

	// === TRANSACTION ROUTES (For Retailer Business Analytics) ===
	api.GET("/transactions", transactionHandler.GetTransactionHistory)     // Detailed transaction history
	api.GET("/transactions/stats", transactionHandler.GetTransactionStats) // Business statistics dashboard
}
