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

// IdentityClient handles gRPC communication with the identity service
type IdentityClient struct {
	client proto.IdentityServiceClient
	conn   *grpc.ClientConn
	logger *logger.Logger
	addr   string
}

// NewIdentityClient creates a new identity service client
func NewIdentityClient(addr string, logger *logger.Logger) (*IdentityClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to identity service: %w", err)
	}

	client := proto.NewIdentityServiceClient(conn)

	return &IdentityClient{
		client: client,
		conn:   conn,
		logger: logger,
		addr:   addr,
	}, nil
}

// Close closes the gRPC connection
func (c *IdentityClient) Close() error {
	return c.conn.Close()
}

// SignUp registers a new user
func (c *IdentityClient) SignUp(ctx context.Context, email, password string) error {
	req := &proto.SignUpRequest{
		Email:    email,
		Password: password,
	}

	_, err := c.client.SignUp(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return fmt.Errorf("identity service error: %s", st.Message())
		}
		return fmt.Errorf("failed to sign up: %w", err)
	}

	return nil
}

// SignIn authenticates a user and returns user info
func (c *IdentityClient) SignIn(ctx context.Context, email, password string) (*proto.User, error) {
	req := &proto.SignInRequest{
		Email:    email,
		Password: password,
	}

	resp, err := c.client.SignIn(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("identity service error: %s", st.Message())
		}
		return nil, fmt.Errorf("failed to sign in: %w", err)
	}

	return resp.User, nil
}

// CreateSession creates a new eKYC session for a user
func (c *IdentityClient) CreateSession(ctx context.Context, userID string) (string, error) {
	req := &proto.CreateSessionRequest{
		UserId: userID,
	}

	resp, err := c.client.CreateSession(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return "", fmt.Errorf("identity service error: %s", st.Message())
		}
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return resp.SessionId, nil
}

// DocumentUploaded notifies that a document has been uploaded
func (c *IdentityClient) DocumentUploaded(ctx context.Context, sessionID, key, docType string) error {
	req := &proto.DocumentUploadedRequest{
		SessionId: sessionID,
		Key:       key,
		Type:      docType,
	}

	_, err := c.client.DocumentUploaded(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return fmt.Errorf("identity service error: %s", st.Message())
		}
		return fmt.Errorf("failed to notify document upload: %w", err)
	}

	return nil
}

// SelfieUploaded notifies that a selfie has been uploaded
func (c *IdentityClient) SelfieUploaded(ctx context.Context, sessionID, key string) error {
	req := &proto.SelfieUploadedRequest{
		SessionId: sessionID,
		Key:       key,
	}

	_, err := c.client.SelfieUploaded(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return fmt.Errorf("identity service error: %s", st.Message())
		}
		return fmt.Errorf("failed to notify selfie upload: %w", err)
	}

	return nil
}

// LivenessUploaded notifies that a liveness check has been uploaded
func (c *IdentityClient) LivenessUploaded(ctx context.Context, sessionID, key string) error {
	req := &proto.LivenessUploadedRequest{
		SessionId: sessionID,
		Key:       key,
	}

	_, err := c.client.LivenessUploaded(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return fmt.Errorf("identity service error: %s", st.Message())
		}
		return fmt.Errorf("failed to notify liveness upload: %w", err)
	}

	return nil
}

// GetStatus retrieves the current status of an eKYC session
func (c *IdentityClient) GetStatus(ctx context.Context, sessionID string) (*proto.SessionStatus, error) {
	req := &proto.GetStatusRequest{
		SessionId: sessionID,
	}

	resp, err := c.client.GetStatus(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("identity service error: %s", st.Message())
		}
		return nil, fmt.Errorf("failed to get session status: %w", err)
	}

	return resp.Status, nil
}

// HealthCheck checks if the identity service is healthy
func (c *IdentityClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.GetStatus(ctx, &proto.GetStatusRequest{SessionId: "health-check"})
	if err != nil {
		// For health check, we don't care about the specific error
		// just that the service is reachable
		return fmt.Errorf("identity service health check failed: %w", err)
	}

	return nil
}
