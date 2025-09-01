package http

import (
	"net/http"

	"day5/internal/infrastructure/container"
	"day5/internal/interfaces/middleware"

	"github.com/gin-gonic/gin"
)

// Router sets up HTTP routes with Clean Architecture and dependency injection
type Router struct {
	container *container.Container
}

// NewRouter creates a new HTTP router with dependency injection
func NewRouter(container *container.Container) *Router {
	return &Router{
		container: container,
	}
}

// SetupRoutes configures all HTTP routes with proper dependency injection
func (r *Router) SetupRoutes() *gin.Engine {
	router := gin.New()

	// Apply middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", r.healthCheck)

	// API routes with dependency injection
	r.setupAPIRoutes(router)

	return router
}

// setupAPIRoutes configures API routes with handlers that use dependency injection
func (r *Router) setupAPIRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")

	// Initialize handlers with use cases from container
	productHandler := NewProductHandler(r.container.GetProductUseCase())
	customerHandler := NewCustomerHandler(r.container.GetCustomerUseCase())
	orderHandler := NewOrderHandler(r.container.GetOrderUseCase())
	transactionHandler := NewTransactionHandler(r.container.GetTransactionUseCase())

	// === PRODUCT ROUTES (For Retailer) ===
	productRoutes := api.Group("/product")
	{
		productRoutes.POST("", productHandler.CreateProduct)    // Create product
		productRoutes.GET("/:id", productHandler.GetProduct)    // Get single product
		productRoutes.PUT("/:id", productHandler.UpdateProduct) // Update product
	}

	// Products collection routes
	api.GET("/products", productHandler.GetProducts)                    // List all products
	api.GET("/products/search", productHandler.SearchProducts)          // Search products
	api.GET("/products/available", productHandler.GetAvailableProducts) // Available products

	// === CUSTOMER ROUTES ===
	customerRoutes := api.Group("/customer")
	{
		customerRoutes.POST("", customerHandler.CreateCustomer)                // Register customer
		customerRoutes.GET("/:id", customerHandler.GetCustomer)                // Get single customer
		customerRoutes.GET("/:id/cooldown", customerHandler.GetCooldownStatus) // Cooldown status
	}

	// Customers collection routes
	api.GET("/customers", customerHandler.GetCustomers)           // List all customers
	api.GET("/customers/search", customerHandler.SearchCustomers) // Search customers

	// === ORDER ROUTES ===
	orderRoutes := api.Group("/order")
	{
		orderRoutes.POST("", orderHandler.PlaceOrder)  // Place order
		orderRoutes.GET("/:id", orderHandler.GetOrder) // Get single order
	}

	// Orders collection routes
	ordersRoutes := api.Group("/orders")
	{
		ordersRoutes.GET("", orderHandler.GetAllOrders)                          // All orders (retailer)
		ordersRoutes.GET("/today", orderHandler.GetTodaysOrders)                 // Today's orders
		ordersRoutes.GET("/customer/:customer_id", orderHandler.GetOrderHistory) // Customer order history
	}

	// === TRANSACTION ROUTES (For Retailer Business Analytics) ===
	transactionRoutes := api.Group("/transactions")
	{
		transactionRoutes.GET("", transactionHandler.GetTransactionHistory)                                       // Transaction history
		transactionRoutes.GET("/stats", transactionHandler.GetTransactionStats)                                   // Business stats
		transactionRoutes.GET("/stats/comprehensive", transactionHandler.GetComprehensiveStats)                   // All periods
		transactionRoutes.GET("/customer/:customer_id/summary", transactionHandler.GetCustomerTransactionSummary) // Customer summary
		transactionRoutes.GET("/revenue/analytics", transactionHandler.GetRevenueAnalytics)                       // Revenue analytics
	}
}

// healthCheck provides a health check endpoint
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "day5-retailer-api",
		"version":   "2.0.0",
		"timestamp": "2025-08-28T12:00:00Z",
	})
}
