package database

import (
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"day5/internal/config"
	"day5/internal/infrastructure/persistence"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database interface for dependency injection
type Database interface {
	GetDB() *gorm.DB
	Close() error
	Migrate() error
	BeginTx() *gorm.DB
}

// DatabaseImpl implements the Database interface
type DatabaseImpl struct {
	db    *gorm.DB
	mutex sync.RWMutex
}

// Singleton instance (for compatibility with existing code)
var (
	instance *DatabaseImpl
	once     sync.Once
)

// InitDatabase initializes the database connection with dependency injection support
func InitDatabase(cfg *config.DatabaseSettings) (Database, error) {
	dbImpl := &DatabaseImpl{}

	if err := dbImpl.connect(cfg); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := dbImpl.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Set singleton for legacy compatibility
	once.Do(func() {
		instance = dbImpl
	})

	return dbImpl, nil
}

// connect establishes the database connection based on dialect
func (d *DatabaseImpl) connect(cfg *config.DatabaseSettings) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	var dialector gorm.Dialector

	// Choose appropriate driver based on dialect
	switch cfg.Dialect {
	case "mysql":
		dialector = mysql.Open(cfg.GetDSN())
	case "postgres":
		dialector = postgres.Open(cfg.GetDSN())
	case "sqlite":
		dialector = sqlite.Open(cfg.GetDSN())
	default:
		return fmt.Errorf("unsupported database dialect: %s", cfg.Dialect)
	}

	// Configure GORM logger based on environment
	var gormLogger logger.Interface
	if config.Config.App.IsProduction() {
		gormLogger = logger.Default.LogMode(logger.Error)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Open database connection
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.db = db
	log.Printf("Database connected successfully using %s dialect", cfg.Dialect)
	return nil
}

// GetDB returns the database instance (thread-safe)
func (d *DatabaseImpl) GetDB() *gorm.DB {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.db
}

// BeginTx begins a new transaction (thread-safe)
func (d *DatabaseImpl) BeginTx() *gorm.DB {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.db.Begin()
}

// Close closes the database connection (thread-safe)
func (d *DatabaseImpl) Close() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.db != nil {
		sqlDB, err := d.db.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}

// Migrate runs database migrations (thread-safe)
func (d *DatabaseImpl) Migrate() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	log.Println("Running database migrations...")
	modelsToMigrate := getModelsToMigrate()
	log.Printf("Migrating %d models:", len(modelsToMigrate))

	for _, model := range modelsToMigrate {
		modelType := reflect.TypeOf(model).Elem()
		log.Printf("  - %s", modelType.Name())
	}

	err := d.db.AutoMigrate(modelsToMigrate...)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// getModelsToMigrate returns all models that need to be migrated
func getModelsToMigrate() []any {
	return persistence.GetModelsToMigrate()
}

// Legacy compatibility functions for existing code that uses global DB
// These will gradually be replaced with dependency injection

// GetDB returns the singleton database instance (legacy compatibility)
func GetDB() *gorm.DB {
	if instance == nil {
		log.Fatal("Database not initialized. Call InitDatabase first.")
	}
	return instance.GetDB()
}

// InitDB initializes the singleton database (legacy compatibility)
func InitDB() error {
	if config.Config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	db, err := InitDatabase(&config.Config.Database)
	if err != nil {
		return err
	}

	// Ensure singleton is set
	if impl, ok := db.(*DatabaseImpl); ok {
		instance = impl
	}

	return nil
}

// DatabaseManager provides a clean interface for database operations
type DatabaseManager struct {
	db Database
}

// NewDatabaseManager creates a new database manager with dependency injection
func NewDatabaseManager(db Database) *DatabaseManager {
	return &DatabaseManager{db: db}
}

// Transaction executes a function within a database transaction
func (dm *DatabaseManager) Transaction(fn func(*gorm.DB) error) error {
	tx := dm.db.BeginTx()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetConnection returns the database connection
func (dm *DatabaseManager) GetConnection() *gorm.DB {
	return dm.db.GetDB()
}

// SafeExecute executes a function with mutex protection
func (dm *DatabaseManager) SafeExecute(fn func(*gorm.DB) error) error {
	if impl, ok := dm.db.(*DatabaseImpl); ok {
		impl.mutex.Lock()
		defer impl.mutex.Unlock()
	}
	return fn(dm.db.GetDB())
}

// Repository base interface for all repositories
type Repository interface {
	SetDB(db Database)
}

// BaseRepository provides common functionality for all repositories
type BaseRepository struct {
	db Database
	mu sync.RWMutex
}

// SetDB sets the database instance for the repository
func (r *BaseRepository) SetDB(db Database) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.db = db
}

// GetDB returns the database instance (thread-safe)
func (r *BaseRepository) GetDB() *gorm.DB {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.db.GetDB()
}

// WithTx executes a function within a transaction
func (r *BaseRepository) WithTx(fn func(*gorm.DB) error) error {
	r.mu.RLock()
	db := r.db
	r.mu.RUnlock()

	tx := db.BeginTx()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
