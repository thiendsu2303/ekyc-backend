package grpcmw

import (
	"context"
	"time"

	"github.com/ekyc-backend/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryTracingInterceptor adds OpenTelemetry tracing to unary gRPC calls
func UnaryTracingInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return otelgrpc.UnaryServerInterceptor(
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
		otelgrpc.WithMeterProvider(otel.GetMeterProvider()),
	)
}

// StreamTracingInterceptor adds OpenTelemetry tracing to streaming gRPC calls
func StreamTracingInterceptor(serviceName string) grpc.StreamServerInterceptor {
	return otelgrpc.StreamServerInterceptor(
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
		otelgrpc.WithMeterProvider(otel.GetMeterProvider()),
	)
}

// UnaryLoggingInterceptor adds structured logging to unary gRPC calls
func UnaryLoggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Extract metadata (not used in this basic implementation)
		_, _ = metadata.FromIncomingContext(ctx)

		// Create span for logging
		span := trace.SpanFromContext(ctx)

		// Log request start
		log.WithContext(ctx).Info("gRPC Request Start")

		// Process request
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Log request completion
		if err != nil {
			log.WithContext(ctx).Error("gRPC Request Failed")
		} else {
			log.WithContext(ctx).Info("gRPC Request Completed")
		}

		// Add duration to span
		span.SetAttributes(attribute.String("grpc.duration", duration.String()))

		return resp, err
	}
}

// StreamLoggingInterceptor adds structured logging to streaming gRPC calls
func StreamLoggingInterceptor(log *logger.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		start := time.Now()

		// Create span for logging
		span := trace.SpanFromContext(ctx)

		// Log stream start
		log.WithContext(ctx).Info("gRPC Stream Start")

		// Process stream
		err := handler(srv, ss)

		// Calculate duration
		duration := time.Since(start)

		// Log stream completion
		if err != nil {
			log.WithContext(ctx).Error("gRPC Stream Failed")
		} else {
			log.WithContext(ctx).Info("gRPC Stream Completed")
		}

		// Add duration to span
		span.SetAttributes(attribute.String("grpc.duration", duration.String()))

		return err
	}
}

// UnaryAuthInterceptor validates JWT tokens for unary gRPC calls
func UnaryAuthInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for certain methods
		if info.FullMethod == "/auth.AuthService/SignIn" || info.FullMethod == "/auth.AuthService/SignUp" {
			return handler(ctx, req)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata not provided")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization header not provided")
		}

		// Validate token (simplified - in real implementation, use proper JWT validation)
		if authHeader[0] == "" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}

		// Add user info to context
		ctx = context.WithValue(ctx, "user_id", "extracted_user_id")

		return handler(ctx, req)
	}
}

// StreamAuthInterceptor validates JWT tokens for streaming gRPC calls
func StreamAuthInterceptor(secret string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// Skip auth for certain methods
		if info.FullMethod == "/auth.AuthService/SignIn" || info.FullMethod == "/auth.AuthService/SignUp" {
			return handler(srv, ss)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Errorf(codes.Unauthenticated, "metadata not provided")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return status.Errorf(codes.Unauthenticated, "authorization header not provided")
		}

		// Validate token (simplified)
		if authHeader[0] == "" {
			return status.Errorf(codes.Unauthenticated, "invalid token")
		}

		// Create new context with user info
		newCtx := context.WithValue(ctx, "user_id", "extracted_user_id")

		// Create wrapped stream
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

// wrappedServerStream wraps grpc.ServerStream to provide custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
