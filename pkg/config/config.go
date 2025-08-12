package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName string
	Environment string
	Port        int

	// Database
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// NATS
	NATSHost string
	NATSPort int

	// MinIO
	MinIOEndpoint        string
	MinIOAccessKeyID     string
	MinIOSecretAccessKey string
	MinIOBucketName      string
	MinIOUseSSL          bool

	// JWT
	JWTSecret     string
	JWTExpiration time.Duration

	// OpenTelemetry
	OTELCollectorEndpoint string
	OTELServiceName       string

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   time.Duration

	// Liveness Test
	LivenessForceFail bool
}

func Load() *Config {
	cfg := &Config{
		ServiceName: getEnv("SERVICE_NAME", "unknown"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnvAsInt("PORT", 8080),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "ekyc"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnvAsInt("REDIS_PORT", 6379),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// NATS
		NATSHost: getEnv("NATS_HOST", "localhost"),
		NATSPort: getEnvAsInt("NATS_PORT", 4222),

		// MinIO
		MinIOEndpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKeyID:     getEnv("MINIO_ACCESS_KEY_ID", "minioadmin"),
		MinIOSecretAccessKey: getEnv("MINIO_SECRET_ACCESS_KEY", "minioadmin"),
		MinIOBucketName:      getEnv("MINIO_BUCKET_NAME", "ekyc"),
		MinIOUseSSL:          getEnvAsBool("MINIO_USE_SSL", false),

		// JWT
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),

		// OpenTelemetry
		OTELCollectorEndpoint: getEnv("OTEL_COLLECTOR_ENDPOINT", "localhost:4317"),
		OTELServiceName:       getEnv("OTEL_SERVICE_NAME", "ekyc-backend"),

		// Rate Limiting
		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvAsDuration("RATE_LIMIT_WINDOW", time.Minute),

		// Liveness Test
		LivenessForceFail: getEnvAsBool("LIVENESS_FORCE_FAIL", false),
	}

	return cfg
}

func (c *Config) GetDBConnString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort)
}

func (c *Config) GetNATSAddr() string {
	return fmt.Sprintf("nats://%s:%d", c.NATSHost, c.NATSPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
