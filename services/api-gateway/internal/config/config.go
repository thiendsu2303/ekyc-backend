package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the API Gateway
type Config struct {
	// Service configuration
	ServiceName string
	HTTPPort    int

	// CORS configuration
	AllowOrigins []string

	// JWT configuration
	JWTSecret string

	// Rate limiting
	RateLimitRPS   int
	RateLimitBurst int

	// Redis configuration
	RedisURL string

	// gRPC service addresses
	IdentityGRPCAddr string
	StorageGRPCAddr  string
	AdminGRPCAddr    string

	// OpenTelemetry configuration
	OTELExporterOTLPEndpoint string

	// Prometheus configuration
	PrometheusMetricsPath string

	// Security
	MaxRequestBodySize int64
	RequestTimeout     time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	viper.SetDefault("SERVICE_NAME", "api-gateway")
	viper.SetDefault("HTTP_PORT", 8080)
	viper.SetDefault("ALLOW_ORIGINS", "*")
	viper.SetDefault("JWT_SECRET", "changeme")
	viper.SetDefault("RATE_LIMIT_RPS", 10)
	viper.SetDefault("RATE_LIMIT_BURST", 20)
	viper.SetDefault("REDIS_URL", "redis://redis:6379")
	viper.SetDefault("IDENTITY_GRPC_ADDR", "identity:9090")
	viper.SetDefault("STORAGE_GRPC_ADDR", "storage-svc:9092")
	viper.SetDefault("ADMIN_GRPC_ADDR", "admin:9093")
	viper.SetDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4317")
	viper.SetDefault("PROMETHEUS_METRICS_PATH", "/metrics")
	viper.SetDefault("MAX_REQUEST_BODY_SIZE", 2*1024*1024) // 2MB
	viper.SetDefault("REQUEST_TIMEOUT", "5s")

	// Read from environment variables
	viper.AutomaticEnv()

	// Parse origins
	originsStr := viper.GetString("ALLOW_ORIGINS")
	var origins []string
	if originsStr == "*" {
		origins = []string{"*"}
	} else {
		origins = strings.Split(originsStr, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
	}

	// Parse request timeout
	timeoutStr := viper.GetString("REQUEST_TIMEOUT")
	requestTimeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REQUEST_TIMEOUT: %w", err)
	}

	config := &Config{
		ServiceName:              viper.GetString("SERVICE_NAME"),
		HTTPPort:                 viper.GetInt("HTTP_PORT"),
		AllowOrigins:             origins,
		JWTSecret:                viper.GetString("JWT_SECRET"),
		RateLimitRPS:             viper.GetInt("RATE_LIMIT_RPS"),
		RateLimitBurst:           viper.GetInt("RATE_LIMIT_BURST"),
		RedisURL:                 viper.GetString("REDIS_URL"),
		IdentityGRPCAddr:         viper.GetString("IDENTITY_GRPC_ADDR"),
		StorageGRPCAddr:          viper.GetString("STORAGE_GRPC_ADDR"),
		AdminGRPCAddr:            viper.GetString("ADMIN_GRPC_ADDR"),
		OTELExporterOTLPEndpoint: viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"),
		PrometheusMetricsPath:    viper.GetString("PROMETHEUS_METRICS_PATH"),
		MaxRequestBodySize:       viper.GetInt64("MAX_REQUEST_BODY_SIZE"),
		RequestTimeout:           requestTimeout,
	}

	// Validate required fields
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validate ensures all required configuration is present
func (c *Config) validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("SERVICE_NAME is required")
	}
	if c.HTTPPort <= 0 || c.HTTPPort > 65535 {
		return fmt.Errorf("HTTP_PORT must be between 1 and 65535")
	}
	if c.JWTSecret == "" || c.JWTSecret == "changeme" {
		return fmt.Errorf("JWT_SECRET must be set to a secure value")
	}
	if c.RedisURL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	if c.IdentityGRPCAddr == "" {
		return fmt.Errorf("IDENTITY_GRPC_ADDR is required")
	}
	if c.StorageGRPCAddr == "" {
		return fmt.Errorf("STORAGE_GRPC_ADDR is required")
	}
	if c.AdminGRPCAddr == "" {
		return fmt.Errorf("ADMIN_GRPC_ADDR is required")
	}
	return nil
}

// GetHTTPAddr returns the HTTP address string
func (c *Config) GetHTTPAddr() string {
	return fmt.Sprintf(":%d", c.HTTPPort)
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return viper.GetString("ENV") == "development" || viper.GetString("ENV") == "dev"
}
