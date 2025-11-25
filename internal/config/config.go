package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	ServerPort     string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	LogLevel       string
	OTelEnabled    bool
	EventQueueSize int
	Hostname       string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "ordersdb"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		OTelEnabled:    getEnvBool("OTEL_ENABLED", true),
		EventQueueSize: getEnvInt("EVENT_QUEUE_SIZE", 100),
		Hostname:       getEnv("HOSTNAME", "localhost"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}
