package container

import (
	"sync"

	"day5/internal/application/usecases"
	"day5/internal/config"
	"day5/internal/database"
	"day5/internal/domain/repositories"
	infraRepo "day5/internal/infrastructure/repositories"

	"gorm.io/gorm"
)

// Container holds all dependencies for dependency injection
// This implements the Dependency Inversion Principle from SOLID
type Container struct {
	// Database
	database database.Database

	// Repositories (infrastructure implements domain interfaces)
	productRepo     repositories.ProductRepository
	customerRepo    repositories.CustomerRepository
	cooldownRepo    repositories.CustomerCooldownRepository
	orderRepo       repositories.OrderRepository
	transactionRepo repositories.TransactionRepository

	// Use Cases (application layer)
	productUseCase     *usecases.ProductUseCase
	customerUseCase    *usecases.CustomerUseCase
	orderUseCase       *usecases.OrderUseCase
	transactionUseCase *usecases.TransactionUseCase

	// Thread safety
	mu   sync.RWMutex
	once sync.Once
}

// NewContainer creates a new dependency injection container
func NewContainer() *Container {
	return &Container{}
}

// Initialize sets up all dependencies with proper injection
func (c *Container) Initialize(cfg *config.AppConfig) error {
	var err error

	c.once.Do(func() {
		// Initialize database
		c.database, err = database.InitDatabase(&cfg.Database)
		if err != nil {
			return
		}

		// Get database connection
		db := c.database.GetDB()

		// Initialize repositories (infrastructure layer)
		c.initializeRepositories(db)

		// Initialize use cases (application layer) with repository dependencies
		c.initializeUseCases(cfg)
	})

	return err
}

// initializeRepositories sets up all repository implementations
func (c *Container) initializeRepositories(db *gorm.DB) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Infrastructure implementations of domain repository interfaces
	c.productRepo = infraRepo.NewProductRepository(db)
	c.customerRepo = infraRepo.NewCustomerRepository(db)
	c.cooldownRepo = infraRepo.NewCustomerCooldownRepository(db)
	c.orderRepo = infraRepo.NewOrderRepository(db)
	c.transactionRepo = infraRepo.NewTransactionRepository(db)
}

// initializeUseCases sets up all use cases with their dependencies
func (c *Container) initializeUseCases(cfg *config.AppConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Application layer use cases with injected dependencies
	c.productUseCase = usecases.NewProductUseCase(c.productRepo)

	c.customerUseCase = usecases.NewCustomerUseCase(
		c.customerRepo,
		c.cooldownRepo,
		cfg.Business.CooldownPeriodMinutes,
	)

	c.orderUseCase = usecases.NewOrderUseCase(
		c.orderRepo,
		c.customerUseCase,
		c.productUseCase,
		c.transactionRepo,
	)

	c.transactionUseCase = usecases.NewTransactionUseCase(
		c.transactionRepo,
		c.customerRepo,
		c.productRepo,
	)
}

// Getters for dependencies (thread-safe)

// GetDatabase returns the database instance
func (c *Container) GetDatabase() database.Database {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.database
}

// Repository getters
func (c *Container) GetProductRepository() repositories.ProductRepository {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.productRepo
}

func (c *Container) GetCustomerRepository() repositories.CustomerRepository {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.customerRepo
}

func (c *Container) GetCustomerCooldownRepository() repositories.CustomerCooldownRepository {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cooldownRepo
}

func (c *Container) GetOrderRepository() repositories.OrderRepository {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.orderRepo
}

func (c *Container) GetTransactionRepository() repositories.TransactionRepository {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.transactionRepo
}

// Use case getters
func (c *Container) GetProductUseCase() *usecases.ProductUseCase {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.productUseCase
}

func (c *Container) GetCustomerUseCase() *usecases.CustomerUseCase {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.customerUseCase
}

func (c *Container) GetOrderUseCase() *usecases.OrderUseCase {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.orderUseCase
}

func (c *Container) GetTransactionUseCase() *usecases.TransactionUseCase {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.transactionUseCase
}

// Cleanup closes all resources
func (c *Container) Cleanup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.database != nil {
		return c.database.Close()
	}
	return nil
}

// Global container instance (singleton pattern for application-wide use)
var (
	globalContainer *Container
	containerOnce   sync.Once
)

// GetGlobalContainer returns the global container instance
func GetGlobalContainer() *Container {
	containerOnce.Do(func() {
		globalContainer = NewContainer()
	})
	return globalContainer
}

// InitializeGlobalContainer initializes the global container
func InitializeGlobalContainer(cfg *config.AppConfig) error {
	container := GetGlobalContainer()
	return container.Initialize(cfg)
}
