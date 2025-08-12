package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/ekyc-backend/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/ekyc-backend/pkg/contracts/proto"
)

// AdminClient handles gRPC communication with the admin service
type AdminClient struct {
	client proto.AdminServiceClient
	conn   *grpc.ClientConn
	logger *logger.Logger
	addr   string
}

// NewAdminClient creates a new admin service client
func NewAdminClient(addr string, logger *logger.Logger) (*AdminClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to admin service: %w", err)
	}

	client := proto.NewAdminServiceClient(conn)

	return &AdminClient{
		client: client,
		conn:   conn,
		logger: logger,
		addr:   addr,
	}, nil
}

// Close closes the gRPC connection
func (c *AdminClient) Close() error {
	return c.conn.Close()
}

// ListSessions retrieves a list of eKYC sessions with filtering and pagination
func (c *AdminClient) ListSessions(ctx context.Context, filter *proto.SessionFilter, page, size int32) (*proto.SessionListResponse, error) {
	req := &proto.ListSessionsRequest{
		Filter: filter,
		Page:   page,
		Size:   size,
	}

	resp, err := c.client.ListSessions(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("admin service error: %s", st.Message())
		}
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return resp, nil
}

// GetSessionDetail retrieves detailed information about a specific session
func (c *AdminClient) GetSessionDetail(ctx context.Context, sessionID string) (*proto.SessionDetailResponse, error) {
	req := &proto.GetSessionDetailRequest{
		SessionId: sessionID,
	}

	resp, err := c.client.GetSessionDetail(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("admin service error: %s", st.Message())
		}
		return nil, fmt.Errorf("failed to get session detail: %w", err)
	}

	return resp, nil
}

// ApplyDecision applies an admin decision to a session
func (c *AdminClient) ApplyDecision(ctx context.Context, sessionID, status, note, adminID string) error {
	req := &proto.ApplyDecisionRequest{
		SessionId: sessionID,
		Status:    status,
		Note:      note,
		AdminId:   adminID,
	}

	_, err := c.client.ApplyDecision(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return fmt.Errorf("admin service error: %s", st.Message())
		}
		return fmt.Errorf("failed to apply decision: %w", err)
	}

	return nil
}

// HealthCheck checks if the admin service is healthy
func (c *AdminClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.GetSessionDetail(ctx, &proto.GetSessionDetailRequest{SessionId: "health-check"})
	if err != nil {
		// For health check, we don't care about the specific error
		// just that the service is reachable
		return fmt.Errorf("admin service health check failed: %w", err)
	}

	return nil
}
