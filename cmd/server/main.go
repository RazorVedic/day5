package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"day5/internal/config"
	"day5/internal/database"
	"day5/internal/handlers"
	"day5/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDatabase()

	// Set Gin mode based on environment
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := setupRouter()

	// Create HTTP server
	server := &http.Server{
		Addr:    config.AppConfig.Server.Host + ":" + config.AppConfig.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter() *gin.Engine {
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

	// Product routes
	api.POST("/product", productHandler.CreateProduct)

	// Additional routes for testing (optional)
	api.GET("/products", productHandler.GetProducts)
	api.GET("/product/:id", productHandler.GetProduct)
}
