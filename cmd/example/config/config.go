package config

import (
	"github.com/razorpay/goutils/grpcserver"
	storage "github.com/razorpay/goutils/sqlstorage"
)

// Config holds the entire configuration for the payment service
type Config struct {
	// App configurations
	App App
	// Server configurations
	Server Server
	// Store configuration: sql, nosql ...
	Store storage.Config
}

// App is application configuration
type App struct {
	// Env the application runs
	Env string
	// ServiceName of the application
	ServiceName string
	// HostName is URL of the service
	HostName string
	// Logger to use, zap is prod default, klog is another option for local
	Logger string
}

// Server contains configuration for the gRPC server
type Server struct {
	// Addresses contains the addresses for the GRPC and HTTP servers
	Addresses grpcserver.ServerAddresses
	// ShutdownTimeout is the maximum duration to
	// wait for server shutdown in seconds
	ShutdownTimeout int
}
