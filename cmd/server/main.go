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
	"day5/internal/router"

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
	appRouter := router.SetupRouter()

	// Create HTTP server
	server := &http.Server{
		Addr:    config.AppConfig.Server.Host + ":" + config.AppConfig.Server.Port,
		Handler: appRouter,
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
