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
	"day5/internal/infrastructure/container"
	httpInterface "day5/internal/interfaces/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from TOML file
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting %s v%s in %s mode",
		config.Config.App.Name,
		config.Config.App.Version,
		config.Config.App.Environment)

	// Set Gin mode based on environment
	if config.Config.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize dependency injection container
	appContainer := container.NewContainer()
	if err := appContainer.Initialize(config.Config); err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Ensure cleanup on exit
	defer func() {
		if err := appContainer.Cleanup(); err != nil {
			log.Printf("Error during cleanup: %v", err)
		}
	}()

	// Initialize HTTP router with dependency injection
	httpRouter := httpInterface.NewRouter(appContainer)
	router := httpRouter.SetupRoutes()

	// Create HTTP server with configuration from TOML
	server := &http.Server{
		Addr:         config.Config.Server.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  time.Duration(config.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(config.Config.Server.Timeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", server.Addr)
		log.Printf("Environment: %s", config.Config.App.Environment)
		log.Printf("Database: %s", config.Config.Database.Dialect)
		log.Printf("Cooldown period: %d minutes", config.Config.Business.CooldownPeriodMinutes)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// Health check for the application
func init() {
	// Set log format
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
