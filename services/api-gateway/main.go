package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ekyc-backend/pkg/logger"
	"github.com/ekyc-backend/pkg/otel"
	"github.com/ekyc-backend/pkg/storage"
	"github.com/ekyc-backend/services/api-gateway/internal/clients"
	"github.com/ekyc-backend/services/api-gateway/internal/config"
	"github.com/ekyc-backend/services/api-gateway/internal/handlers"
	"github.com/ekyc-backend/services/api-gateway/internal/metrics"
	"github.com/ekyc-backend/services/api-gateway/internal/security"
	"github.com/ekyc-backend/services/api-gateway/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.ServiceName)
	defer logger.Sync()

	logger.Info("Starting API Gateway")

	// Initialize OpenTelemetry
	if err := otel.InitTracer(cfg, cfg.ServiceName); err != nil {
		logger.Error("Failed to initialize OpenTelemetry tracer")
	}

	if err := otel.InitMeter(cfg, cfg.ServiceName); err != nil {
		logger.Error("Failed to initialize OpenTelemetry meter")
	}

	// Initialize Redis client
	redisClient, err := storage.NewRedis(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Initialize gRPC clients
	identityClient, err := clients.NewIdentityClient(cfg.IdentityGRPCAddr, logger)
	if err != nil {
		logger.Fatal("Failed to create identity client")
	}
	defer identityClient.Close()

	storageClient, err := clients.NewStorageClient(cfg.StorageGRPCAddr, logger)
	if err != nil {
		logger.Fatal("Failed to create storage client")
	}
	defer storageClient.Close()

	adminClient, err := clients.NewAdminClient(cfg.AdminGRPCAddr, logger)
	if err != nil {
		logger.Fatal("Failed to create admin client")
	}
	defer adminClient.Close()

	// Initialize JWT manager
	jwtManager := security.NewJWTManager(cfg.JWTSecret)

	// Initialize metrics
	metrics := metrics.NewMetrics()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(logger, identityClient, storageClient, adminClient)

	// Initialize server
	srv := server.NewServer(cfg, logger, redisClient, jwtManager, metrics, healthHandler)

	// Start server
	go func() {
		logger.Info("Starting HTTP server", "addr", cfg.GetHTTPAddr())
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
