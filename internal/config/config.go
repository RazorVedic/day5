package config

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/viper"
)

// AppConfig holds the entire application configuration
type AppConfig struct {
	App      AppSettings      `mapstructure:"app"`
	Server   ServerSettings   `mapstructure:"server"`
	Database DatabaseSettings `mapstructure:"database"`
	Logging  LoggingSettings  `mapstructure:"logging"`
	Business BusinessSettings `mapstructure:"business"`
	Security SecuritySettings `mapstructure:"security"`
	Cache    CacheSettings    `mapstructure:"cache"`
}

// AppSettings contains general application settings
type AppSettings struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerSettings contains HTTP server configuration
type ServerSettings struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Timeout      int    `mapstructure:"timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseSettings contains database configuration with multi-dialect support
type DatabaseSettings struct {
	Dialect         string `mapstructure:"dialect"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	Loc             string `mapstructure:"loc"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`

	// SQLite specific
	SQLitePath string `mapstructure:"sqlite_path"`

	// PostgreSQL specific
	PostgresSSLMode  string `mapstructure:"postgres_sslmode"`
	PostgresTimezone string `mapstructure:"postgres_timezone"`
}

// LoggingSettings contains logging configuration
type LoggingSettings struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// BusinessSettings contains business logic configuration
type BusinessSettings struct {
	CooldownPeriodMinutes int    `mapstructure:"cooldown_period_minutes"`
	DefaultCurrency       string `mapstructure:"default_currency"`
	CurrencyPrecision     int    `mapstructure:"currency_precision"`
}

// SecuritySettings contains security-related configuration
type SecuritySettings struct {
	JWTSecret              string `mapstructure:"jwt_secret"`
	JWTExpiryHours         int    `mapstructure:"jwt_expiry_hours"`
	RateLimitRequests      int    `mapstructure:"rate_limit_requests"`
	RateLimitWindowMinutes int    `mapstructure:"rate_limit_window_minutes"`
}

// CacheSettings contains cache configuration
type CacheSettings struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Global configuration instance
var Config *AppConfig

// LoadConfig loads configuration from TOML file using viper
func LoadConfig() error {
	viper.SetConfigName(getConfigName())
	viper.SetConfigType("toml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")    // For tests
	viper.AddConfigPath("../../config") // For deeper test directories

	// Enable automatic environment variable substitution
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal into our config struct
	Config = &AppConfig{}
	if err := viper.Unmarshal(Config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required configuration
	if err := validateConfig(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	return nil
}

// getConfigName determines which config file to load based on environment
func getConfigName() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("GO_ENV")
	}
	if env == "" {
		env = "dev" // default to development
	}

	switch env {
	case "production", "prod":
		return "prod"
	case "test", "testing":
		return "test"
	default:
		return "dev"
	}
}

// validateConfig validates the loaded configuration
func validateConfig() error {
	if Config.Database.Dialect == "" {
		return fmt.Errorf("database dialect is required")
	}

	validDialects := []string{"mysql", "postgres", "sqlite"}
	if !slices.Contains(validDialects, Config.Database.Dialect) {
		return fmt.Errorf("unsupported database dialect: %s. Supported: %v",
			Config.Database.Dialect, validDialects)
	}

	// Dialect-specific validation
	switch Config.Database.Dialect {
	case "mysql", "postgres":
		if Config.Database.Host == "" {
			return fmt.Errorf("database host is required for %s", Config.Database.Dialect)
		}
		if Config.Database.Name == "" {
			return fmt.Errorf("database name is required for %s", Config.Database.Dialect)
		}
	case "sqlite":
		if Config.Database.SQLitePath == "" {
			return fmt.Errorf("sqlite_path is required for SQLite dialect")
		}
	}

	if Config.Server.Port <= 0 || Config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", Config.Server.Port)
	}

	return nil
}

// GetDSN returns the appropriate database connection string based on dialect
func (db *DatabaseSettings) GetDSN() string {
	switch db.Dialect {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
			db.Username, db.Password, db.Host, db.Port, db.Name,
			db.Charset, db.ParseTime, db.Loc)

	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			db.Host, db.Port, db.Username, db.Password, db.Name,
			db.PostgresSSLMode, db.PostgresTimezone)

	case "sqlite":
		return db.SQLitePath

	default:
		panic(fmt.Sprintf("unsupported database dialect: %s", db.Dialect))
	}
}

// Convenience methods for common checks
func (a *AppSettings) IsProduction() bool {
	return a.Environment == "production" || a.Environment == "prod"
}

func (a *AppSettings) IsDevelopment() bool {
	return a.Environment == "development" || a.Environment == "dev"
}

func (a *AppSettings) IsTest() bool {
	return a.Environment == "test" || a.Environment == "testing"
}

// GetServerAddress returns the complete server address
func (s *ServerSettings) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Legacy compatibility functions for existing code
func IsProduction() bool {
	if Config == nil {
		return false
	}
	return Config.App.IsProduction()
}

func GetServerHost() string {
	if Config == nil {
		return "0.0.0.0"
	}
	return Config.Server.Host
}

func GetServerPort() string {
	if Config == nil {
		return "8080"
	}
	return fmt.Sprintf("%d", Config.Server.Port)
}
