package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the API Gateway
type Metrics struct {
	// HTTP request metrics (RED metrics)
	RequestTotal    *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	RequestErrors   *prometheus.CounterVec

	// Custom metrics
	IdempotencyHits   *prometheus.CounterVec
	RateLimitedTotal  *prometheus.CounterVec
	ActiveConnections *prometheus.GaugeVec

	// gRPC client metrics
	GRPCClientRequests *prometheus.CounterVec
	GRPCClientDuration *prometheus.HistogramVec
	GRPCClientErrors   *prometheus.CounterVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	metrics := &Metrics{
		// HTTP request metrics
		RequestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status", "user_role"},
		),

		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),

		RequestErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_http_request_errors_total",
				Help: "Total number of HTTP request errors",
			},
			[]string{"method", "path", "error_type"},
		),

		// Custom metrics
		IdempotencyHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_idempotency_hits_total",
				Help: "Total number of idempotency cache hits",
			},
			[]string{"route"},
		),

		RateLimitedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_rate_limited_total",
				Help: "Total number of rate limited requests",
			},
			[]string{"ip", "route"},
		),

		ActiveConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gateway_active_connections",
				Help: "Number of active connections",
			},
			[]string{"type"},
		),

		// gRPC client metrics
		GRPCClientRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_grpc_client_requests_total",
				Help: "Total number of gRPC client requests",
			},
			[]string{"service", "method", "status"},
		),

		GRPCClientDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_grpc_client_duration_seconds",
				Help:    "gRPC client request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method"},
		),

		GRPCClientErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_grpc_client_errors_total",
				Help: "Total number of gRPC client errors",
			},
			[]string{"service", "method", "error_type"},
		),
	}

	return metrics
}

// RecordHTTPRequest records an HTTP request metric
func (m *Metrics) RecordHTTPRequest(method, path, status, userRole string) {
	m.RequestTotal.WithLabelValues(method, path, status, userRole).Inc()
}

// RecordHTTPRequestDuration records an HTTP request duration metric
func (m *Metrics) RecordHTTPRequestDuration(method, path, status string, duration float64) {
	m.RequestDuration.WithLabelValues(method, path, status).Observe(duration)
}

// RecordHTTPRequestError records an HTTP request error metric
func (m *Metrics) RecordHTTPRequestError(method, path, errorType string) {
	m.RequestErrors.WithLabelValues(method, path, errorType).Inc()
}

// RecordIdempotencyHit records an idempotency cache hit
func (m *Metrics) RecordIdempotencyHit(route string) {
	m.IdempotencyHits.WithLabelValues(route).Inc()
}

// RecordRateLimited records a rate limited request
func (m *Metrics) RecordRateLimited(ip, route string) {
	m.RateLimitedTotal.WithLabelValues(ip, route).Inc()
}

// SetActiveConnections sets the number of active connections
func (m *Metrics) SetActiveConnections(connType string, count float64) {
	m.ActiveConnections.WithLabelValues(connType).Set(count)
}

// RecordGRPCClientRequest records a gRPC client request metric
func (m *Metrics) RecordGRPCClientRequest(service, method, status string) {
	m.GRPCClientRequests.WithLabelValues(service, method, status).Inc()
}

// RecordGRPCClientDuration records a gRPC client request duration metric
func (m *Metrics) RecordGRPCClientDuration(service, method string, duration float64) {
	m.GRPCClientDuration.WithLabelValues(service, method).Observe(duration)
}

// RecordGRPCClientError records a gRPC client error metric
func (m *Metrics) RecordGRPCClientError(service, method, errorType string) {
	m.GRPCClientErrors.WithLabelValues(service, method, errorType).Inc()
}
