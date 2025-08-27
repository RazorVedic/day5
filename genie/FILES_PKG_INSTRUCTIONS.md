**YOU MUST NOT implement any dependency or package. Always use the existing package, e.g., use the server pkg for creating server, pubsub pkg for creating pubsub service.**

## Packages and Dependencies

### Server
- Use `pkg/server/*` to create a new HTTP and gRPC server.
- This is mainly used in `main.go` files to create an HTTP and gRPC server, and also gRPC connections and clients.
- This package exposes functionalities to create a new gRPC connection, create a new HTTP and gRPC server, and run those servers.
- Every server must register the health check service.
- During initialization of the server, figure out the necessary dependencies and pass them to the `New` function.
- An example of server initialization is given below:

```go
customerConn, err := server.NewConnection(config.GrpcServiceAddress.Customer)
if err != nil {
  log.Fatalf(ctx, "error creating customer connection: %v", err)
}
customersClient := customerrpc.NewCustomersClient(customerConn)

// create server with the above-created services
// server creates grpc server along with a grpc gateway
srv, err := server.New(ctx,
  server.WithConfig(config.Server),
  server.WithGRPCHandler(grpcHandler(ctx, paymentService, healthService, observeServices)),
  server.WithHTTPHandler(httpHandler(ctx, observeServices)),
  server.WithPassportService(passportService),
  server.WithObserverService(observeServices),
  server.WithCustomerClient(customersClient),
)
if err != nil {
  log.Fatalf(ctx, "could not start server: %+v", err)
}
```

```go
func grpcHandler(_ context.Context,
  svc payment.Service,
  healthcheckService healthcheck.Service,
  obs *observability.Observability,
) func(*grpc.Server) error {

  return func(server *grpc.Server) error {
    // register payment server
    paymentrpc.RegisterPaymentsServer(
      server,
      payment.NewServer(svc, obs),
    )

    // register healthcheck server
    healthcheckrpc.RegisterHealthCheckServiceServer(
      server,
      healthcheck.NewServer(healthcheckService),
    )

    return nil
  }
}
```

```go
func httpHandler(
  ctx context.Context,
  obs *observability.Observability,
) func(*runtime.ServeMux, string) error {

  return func(mux *runtime.ServeMux, address string) error {
    grpcOptions := []grpc.DialOption{
      grpc.WithTransportCredentials(insecure.NewCredentials()),
      grpc.WithUnaryInterceptor(interceptor.UnaryClientTracingInterceptor(obs.Tracer)),
    }

    // register HTTP endpoint for payment server
    err := paymentrpc.RegisterPaymentsHandlerFromEndpoint(
      ctx, mux, address,
      grpcOptions,
    )
    if err != nil {
      return err
    }

    // register healthcheck service
    err = healthcheckrpc.RegisterHealthCheckServiceHandlerFromEndpoint(
      ctx, mux, address, grpcOptions,
    )
    if err != nil {
      return err
    }

    return nil
  }
}
```

- An example to create a gRPC client:

```go
vpaConn, err := server.NewConnection(config.GrpcServiceAddress.Vpa)
if err != nil {
  log.Fatalf(ctx, "error creating vpa connection: %v", err)
}
vpaClient := vparpc.NewVpaClient(vpaConn)
```

Here we are creating a gRPC client for the vpa service.

### PubSub
- Use `pkg/pubsub/*` to create a new pub or sub instance.
- This package exposes functionality to publish an event via pub and to run a consumer with a handler function via sub.
- We won't use this directly. Every service will have a `Publisher` interface, and we will inject this from `main.go`.
- Example to initialize pub service:

```go
pubService, err := pub.New(
  ctx,
  config.Pub,
  paymenttopic.PublishTopics,
)
if err != nil {
  log.Fatalf(ctx, "could not start publisher service, err: %+v", err)
}
```

### Observability
- Use `pkg/observability/*` to create a new observability instance.
- This package exposes functionality to initialize the observability tools such as logger, tracer, and metrics. These tools can then be globally used, e.g., `log.Infof()`.
- Always make sure to close the observability.
- Example to initialize observability:

```go
// observability services
// create logger, tracer, and metrics
observeServices, ctx, err := observability.New(
  ctx,
  config.App.Logger,
  config.Metric,
  config.Sentry,
  &config.Tracing,
)
defer observability.CloseTrace(ctx, observeServices.Tracer)
if err != nil {
  log.Fatalf(ctx, "could not create observer services, err: %+v", err)
}
```

### Storage
- Use `pkg/storage/*` to create a new storage instance.
- This exposes functionality to create a new storage instance and it also abstracts the actual DB that is being used using the `Store` interface.
- Always use this interface in the repo layer of any service to perform DB operations such as CRUD operations.
- Example to initialize storage:

```go
store, err := storage.New(ctx, config.Store)
if err != nil {
  log.Fatalf(ctx, "could not create storage object, err: %+v", err)
}
```

### ConfigLoader
- Use `pkg/configloader/*` to load configurations from a given TOML file.
- Example of its usage:

```go
opts := configloader.NewOptions("toml", "./config/payment", "default")
loader := configloader.NewLoader(opts)
err := loader.Load(env, &config)
if err != nil {
  fmt.Println(err.Error())
  os.Exit(1)
}
```