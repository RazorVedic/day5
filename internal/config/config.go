package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Host string
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	DSN      string
}

var AppConfig *Config

func LoadConfig() {
	// Load .env file if it exists (for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", "pass"),
			DBName:   getEnv("DB_NAME", "day5"),
		},
	}
	// Build DSN for MySQL
	AppConfig.Database.DSN = AppConfig.Database.User + ":" +
		AppConfig.Database.Password + "@tcp(" +
		AppConfig.Database.Host + ":" + AppConfig.Database.Port + ")/" +
		AppConfig.Database.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	fmt.Println(AppConfig)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func IsDevelopment() bool {
	return AppConfig.Server.Env == "development"
}

func IsProduction() bool {
	return AppConfig.Server.Env == "production"
}
