package server

import (
	"context"
	"net/http"
	"time"

	"github.com/ekyc-backend/pkg/logger"
	"github.com/ekyc-backend/pkg/storage"
	"github.com/ekyc-backend/services/api-gateway/internal/config"
	"github.com/ekyc-backend/services/api-gateway/internal/handlers"
	"github.com/ekyc-backend/services/api-gateway/internal/metrics"
	"github.com/ekyc-backend/services/api-gateway/internal/middleware"
	"github.com/ekyc-backend/services/api-gateway/internal/security"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the HTTP server
type Server struct {
	config        *config.Config
	logger        *logger.Logger
	redisClient   *storage.Redis
	jwtManager    *security.JWTManager
	metrics       *metrics.Metrics
	healthHandler *handlers.HealthHandler
	server        *http.Server
	router        *chi.Mux
}

// NewServer creates a new HTTP server
func NewServer(
	cfg *config.Config,
	logger *logger.Logger,
	redisClient *storage.Redis,
	jwtManager *security.JWTManager,
	metrics *metrics.Metrics,
	healthHandler *handlers.HealthHandler,
) *Server {
	s := &Server{
		config:        cfg,
		logger:        logger,
		redisClient:   redisClient,
		jwtManager:    jwtManager,
		metrics:       metrics,
		healthHandler: healthHandler,
	}

	s.setupRouter()
	s.setupServer()

	return s
}

// setupRouter sets up the Chi router with all middleware and routes
func (s *Server) setupRouter() {
	s.router = chi.NewRouter()

	// Middleware stack (order matters)
	s.router.Use(
		// Recovery and panic handling
		middleware.Recover(s.logger),

		// Request ID and correlation ID
		middleware.RequestID(),
		middleware.CorrelationID(),

		// OpenTelemetry tracing
		middleware.Tracing(s.config.ServiceName),

		// Logging
		middleware.Logging(s.logger),

		// CORS
		middleware.CORS(s.config.AllowOrigins),

		// Rate limiting
		middleware.RateLimit(s.redisClient, s.config.RateLimitRPS, s.config.RateLimitBurst, s.logger),

		// Idempotency (for POST requests)
		middleware.Idempotency(s.redisClient, s.logger),
	)

	// Health check endpoints (no auth required)
	s.router.Get("/live", s.healthHandler.Live)
	s.router.Get("/ready", s.healthHandler.Ready)
	s.router.Handle(s.config.PrometheusMetricsPath, promhttp.Handler())

	// API routes
	s.router.Route("/api/v1", func(r chi.Router) {
		// Auth routes (no auth required)
		r.Route("/auth", func(r chi.Router) {
			// TODO: Add auth handlers
		})

		// eKYC routes (require USER or ADMIN role)
		r.Route("/ekyc", func(r chi.Router) {
			r.Use(middleware.Auth(s.jwtManager, s.logger))
			r.Use(middleware.RequireAnyRole("USER", "ADMIN"))

			// TODO: Add eKYC handlers
		})

		// Admin routes (require ADMIN role)
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(s.jwtManager, s.logger))
			r.Use(middleware.RequireRole("ADMIN", s.logger))

			// TODO: Add admin handlers
		})
	})
}

// setupServer sets up the HTTP server
func (s *Server) setupServer() {
	s.server = &http.Server{
		Addr:         s.config.GetHTTPAddr(),
		Handler:      s.router,
		ReadTimeout:  s.config.RequestTimeout,
		WriteTimeout: s.config.RequestTimeout,
		IdleTimeout:  60 * time.Second,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// GetRouter returns the Chi router (useful for testing)
func (s *Server) GetRouter() *chi.Mux {
	return s.router
}
