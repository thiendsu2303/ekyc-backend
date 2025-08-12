package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ekyc-backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type EventBus interface {
	Publish(ctx context.Context, subject string, event interface{}) error
	Subscribe(ctx context.Context, subject string, queueGroup string, handler EventHandler) error
	Close() error
}

type EventHandler func(ctx context.Context, event []byte) error

type NATSEventBus struct {
	conn   *nats.Conn
	logger *logger.Logger
}

type EventMetadata struct {
	EventID       string    `json:"event_id"`
	CorrelationID string    `json:"correlation_id"`
	SessionID     string    `json:"session_id"`
	OccurredAt    time.Time `json:"occurred_at"`
	EventType     string    `json:"event_type"`
	SourceService string    `json:"source_service"`
}

func NewNATSEventBus(natsURL string, logger *logger.Logger) (*NATSEventBus, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NATSEventBus{
		conn:   conn,
		logger: logger,
	}, nil
}

func (n *NATSEventBus) Publish(ctx context.Context, subject string, event interface{}) error {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	// Create event metadata
	metadata := EventMetadata{
		EventID:       uuid.New().String(),
		OccurredAt:    time.Now().UTC(),
		EventType:     subject,
		SourceService: "unknown", // This will be set by the service
	}

	// Add correlation ID and session ID from context if available
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			metadata.CorrelationID = id
		}
	}

	if sessionID := ctx.Value("session_id"); sessionID != nil {
		if id, ok := sessionID.(string); ok {
			metadata.SessionID = id
		}
	}

	// Create envelope with metadata and event
	envelope := map[string]interface{}{
		"metadata": metadata,
		"data":     event,
	}

	// Serialize envelope
	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Add tracing attributes
	span.SetAttributes(
		attribute.String("nats.subject", subject),
		attribute.String("nats.event_id", metadata.EventID),
		attribute.String("nats.correlation_id", metadata.CorrelationID),
		attribute.String("nats.session_id", metadata.SessionID),
	)

	// Publish to NATS
	if err := n.conn.Publish(subject, payload); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	// Publish to DLQ if there's an error (simulated)
	if metadata.EventType == "error" {
		dlqSubject := subject + ".dlq"
		n.conn.Publish(dlqSubject, payload)
	}

	n.logger.WithContext(ctx).WithTraceID(span.SpanContext().TraceID().String()).
		WithSpanID(span.SpanContext().SpanID().String()).
		WithCorrelationID(metadata.CorrelationID).
		WithSessionID(metadata.SessionID).
		Info("Event published")

	return nil
}

func (n *NATSEventBus) Subscribe(ctx context.Context, subject string, queueGroup string, handler EventHandler) error {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	// Subscribe with queue group for load balancing
	subscription, err := n.conn.QueueSubscribe(subject, queueGroup, func(msg *nats.Msg) {
		// Create new context for each message
		msgCtx := context.Background()

		// Extract correlation ID and session ID from message headers if available
		if msg.Header != nil {
			if correlationID := msg.Header.Get("X-Correlation-ID"); correlationID != "" {
				msgCtx = context.WithValue(msgCtx, "correlation_id", correlationID)
			}
			if sessionID := msg.Header.Get("X-Session-ID"); sessionID != "" {
				msgCtx = context.WithValue(msgCtx, "session_id", sessionID)
			}
		}

		// Create span for message processing
		msgSpan := trace.SpanFromContext(msgCtx)
		defer msgSpan.End()

		// Process message
		if err := handler(msgCtx, msg.Data); err != nil {
			msgSpan.RecordError(err)
			n.logger.WithContext(msgCtx).Error("Failed to process message")
		}

		// Acknowledge message
		msg.Ack()
	})

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to subscribe to NATS: %w", err)
	}

	// Add subscription to context for cleanup
	ctx = context.WithValue(ctx, "subscription", subscription)

	n.logger.WithContext(ctx).Info("Subscribed to NATS subject")

	return nil
}

func (n *NATSEventBus) Close() error {
	if n.conn != nil {
		n.conn.Close()
	}
	return nil
}

// Helper function to create correlation ID
func NewCorrelationID() string {
	return uuid.New().String()
}

// Helper function to create session ID
func NewSessionID() string {
	return uuid.New().String()
}
