package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ekyc-backend/pkg/logger"
	"github.com/ekyc-backend/services/api-gateway/internal/clients"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	logger         *logger.Logger
	identityClient *clients.IdentityClient
	storageClient  *clients.StorageClient
	adminClient    *clients.AdminClient
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	logger *logger.Logger,
	identityClient *clients.IdentityClient,
	storageClient *clients.StorageClient,
	adminClient *clients.AdminClient,
) *HealthHandler {
	return &HealthHandler{
		logger:         logger,
		identityClient: identityClient,
		storageClient:  storageClient,
		adminClient:    adminClient,
	}
}

// Live handles the /live endpoint (process health)
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	// Simple process health check - always return 200 if the process is running
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "api-gateway",
	}

	json.NewEncoder(w).Encode(response)
}

// Ready handles the /ready endpoint (service health)
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check all downstream services
	health := h.checkServiceHealth(ctx)

	if health.Healthy {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(health)
}

// checkServiceHealth checks the health of all downstream services
func (h *HealthHandler) checkServiceHealth(ctx context.Context) *HealthStatus {
	health := &HealthStatus{
		Status:    "unhealthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   "api-gateway",
		Checks:    make(map[string]CheckResult),
	}

	// Check Identity Service
	identityHealthy := h.checkIdentityHealth(ctx)
	health.Checks["identity"] = identityHealthy

	// Check Storage Service
	storageHealthy := h.checkStorageHealth(ctx)
	health.Checks["storage"] = storageHealthy

	// Check Admin Service
	adminHealthy := h.checkAdminHealth(ctx)
	health.Checks["admin"] = adminHealthy

	// Overall health
	overallHealthy := identityHealthy.Healthy && storageHealthy.Healthy && adminHealthy.Healthy
	if overallHealthy {
		health.Status = "healthy"
	}

	return health
}

// checkIdentityHealth checks the health of the identity service
func (h *HealthHandler) checkIdentityHealth(ctx context.Context) CheckResult {
	if h.identityClient == nil {
		return CheckResult{
			Name:      "identity",
			Status:    "unhealthy",
			Message:   "Client not initialized",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}

	err := h.identityClient.HealthCheck(ctx)
	if err != nil {
		return CheckResult{
			Name:      "identity",
			Status:    "unhealthy",
			Message:   err.Error(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}

	return CheckResult{
		Name:      "identity",
		Status:    "healthy",
		Message:   "Service responding",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// checkStorageHealth checks the health of the storage service
func (h *HealthHandler) checkStorageHealth(ctx context.Context) CheckResult {
	if h.storageClient == nil {
		return CheckResult{
			Name:      "storage",
			Status:    "unhealthy",
			Message:   "Client not initialized",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}

	err := h.storageClient.HealthCheck(ctx)
	if err != nil {
		return CheckResult{
			Name:      "storage",
			Status:    "unhealthy",
			Message:   err.Error(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}

	return CheckResult{
		Name:      "storage",
		Status:    "healthy",
		Message:   "Service responding",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// checkAdminHealth checks the health of the admin service
func (h *HealthHandler) checkAdminHealth(ctx context.Context) CheckResult {
	if h.adminClient == nil {
		return CheckResult{
			Name:      "admin",
			Status:    "unhealthy",
			Message:   "Client not initialized",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}

	err := h.adminClient.HealthCheck(ctx)
	if err != nil {
		return CheckResult{
			Name:      "admin",
			Status:    "unhealthy",
			Message:   err.Error(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}

	return CheckResult{
		Name:      "admin",
		Status:    "healthy",
		Message:   "Service responding",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Service   string                 `json:"service"`
	Checks    map[string]CheckResult `json:"checks"`
	Healthy   bool                   `json:"healthy"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Healthy   bool   `json:"healthy"`
}
