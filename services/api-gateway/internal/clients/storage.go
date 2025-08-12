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

// StorageClient handles gRPC communication with the storage service
type StorageClient struct {
	client proto.StorageServiceClient
	conn   *grpc.ClientConn
	logger *logger.Logger
	addr   string
}

// NewStorageClient creates a new storage service client
func NewStorageClient(addr string, logger *logger.Logger) (*StorageClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to storage service: %w", err)
	}

	client := proto.NewStorageServiceClient(conn)

	return &StorageClient{
		client: client,
		conn:   conn,
		logger: logger,
		addr:   addr,
	}, nil
}

// Close closes the gRPC connection
func (c *StorageClient) Close() error {
	return c.conn.Close()
}

// GetPresignedPut generates a presigned PUT URL for file upload
func (c *StorageClient) GetPresignedPut(ctx context.Context, key, contentType string, expiresIn time.Duration) (*proto.PresignedURLResponse, error) {
	req := &proto.GetPresignedPutRequest{
		Key:         key,
		ContentType: contentType,
		ExpiresIn:   int32(expiresIn.Seconds()),
	}

	resp, err := c.client.GetPresignedPut(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("storage service error: %s", st.Message())
		}
		return nil, fmt.Errorf("failed to get presigned PUT URL: %w", err)
	}

	return resp, nil
}

// GetPresignedGet generates a presigned GET URL for file download
func (c *StorageClient) GetPresignedGet(ctx context.Context, key string, expiresIn time.Duration) (*proto.PresignedURLResponse, error) {
	req := &proto.GetPresignedGetRequest{
		Key:       key,
		ExpiresIn: int32(expiresIn.Seconds()),
	}

	resp, err := c.client.GetPresignedGet(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("storage service error: %s", st.Message())
		}
		return nil, fmt.Errorf("failed to get presigned GET URL: %w", err)
	}

	return resp, nil
}

// HealthCheck checks if the storage service is healthy
func (c *StorageClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.GetPresignedGet(ctx, &proto.GetPresignedGetRequest{Key: "health-check"})
	if err != nil {
		// For health check, we don't care about the specific error
		// just that the service is reachable
		return fmt.Errorf("storage service health check failed: %w", err)
	}

	return nil
}
