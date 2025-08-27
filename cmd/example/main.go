package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/razorpay/goutils/configloader"
	server "github.com/razorpay/goutils/grpcserver"
	"github.com/razorpay/goutils/healthcheck"
	logger "github.com/razorpay/goutils/logger/v3"
	storage "github.com/razorpay/goutils/sqlstorage"
	healthrpc "github.com/razorpay/rpc/healthcheck/v1"

	"github.com/razorpay/go-foundation-v2/cmd/example/config"
	"github.com/razorpay/go-foundation-v2/internal/example"
	"github.com/razorpay/go-foundation-v2/internal/example/service"
	paymentrpc "github.com/razorpay/go-foundation-v2/rpc/gofoundationv2/v1/example/payment"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get app env
	env := configloader.GetAppEnv()

	// setup logger
	slogger := slog.New(
		logger.NewHandler(nil),
	).With(
		slog.String("service", "example"),
		slog.String("app_env", env),
	)

	// parse config values
	config := config.Config{}
	loader := configloader.New(
		configloader.WithConfigDir("./config/example"),
	)
	if err := loader.Load(env, &config); err != nil {
		logger.WithError(slogger, err).Error("could not load config")
		os.Exit(1)
	}

	// create store which would be used by services created below
	store, err := storage.New(ctx, config.Store)
	if err != nil {
		logger.WithError(slogger, err).Error("could not create storage object")
		os.Exit(1)
	}

	healthService := healthcheck.NewService()
	// TODO: add support for store db check
	healthService.AddReadinessCheck(
		"db-connection", healthcheck.DatabasePingCheck(nil),
	)
	healthService.AddLivenessCheck(
		"max-gc-pause", healthcheck.GCMaxPauseCheck(100*time.Millisecond),
	)

	// create payment service
	paymentService, err := service.New(
		service.WithStorage(store),
	)
	if err != nil {
		logger.WithError(slogger, err).Error("error creating payment service")
	}

	// create server with the above created services
	srv, err := server.NewServer(
		config.Server.Addresses,
		grpcHandler(paymentService, healthService),
		httpHandler(ctx),
	)
	if err != nil {
		logger.WithError(slogger, err).Error("could not start server")
	}

	// graceful shutdown
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigterm:
			slogger.Info("sigterm received")
		case <-ctx.Done():
			slogger.Info("context done, bye")
			return
		}

		err := srv.Stop(ctx, config.Server.ShutdownTimeout)
		if err != nil {
			logger.WithError(slogger, err).
				Error("error shutting down server(s)")
		}

		cancel()
	}()

	// starts http and grpc server
	wg.Add(1)
	go func() {
		defer wg.Done()
		slogger.Info(
			"starting server(s)",
			slog.String("http_server_addr", config.Server.Addresses.Http),
			slog.String("grpc_server_addr", config.Server.Addresses.Grpc),
		)
		err = srv.Start(ctx)
		if err != nil {
			logger.WithError(slogger, err).Error("error running server(s)")
			cancel()
		}

		slogger.Info("cancel() context")
		cancel()
	}()

	// wait for all go routines to shut down
	// then exit main routine
	wg.Wait()
}

// grpcHandler register all the servers
// all Register code are generated under rpc/
func grpcHandler(
	svc example.Service,
	healthcheckSvc healthcheck.Service,
) func(*grpc.Server) error {
	return func(server *grpc.Server) error {
		// register payment server
		paymentrpc.RegisterPaymentServiceServer(
			server,
			example.NewServer(svc))

		healthrpc.RegisterHealthCheckServiceServer(
			server,
			healthcheck.NewGRPCServer(healthcheckSvc),
		)

		return nil
	}
}

// httpHandler registers the proxy for all the grpc handlers
func httpHandler(ctx context.Context) func(*runtime.ServeMux, string) error {
	return func(mux *runtime.ServeMux, address string) error {
		grpcOptions := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		// register http endpoint for payment server
		err := paymentrpc.RegisterPaymentServiceHandlerFromEndpoint(
			ctx, mux, address,
			grpcOptions,
		)
		if err != nil {
			return err
		}

		// register healthcheck service
		err = healthrpc.RegisterHealthCheckServiceHandlerFromEndpoint(
			ctx, mux, address, grpcOptions,
		)
		if err != nil {
			return err
		}

		return nil
	}
}
