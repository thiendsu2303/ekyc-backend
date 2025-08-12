package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func New(serviceName string) *Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"

	// Add service name to all logs
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &Logger{logger}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := []zap.Field{}

	// Add trace context if available
	if span := trace.SpanFromContext(ctx); span != nil {
		spanCtx := span.SpanContext()
		if spanCtx.TraceID().IsValid() {
			fields = append(fields, zap.String("trace_id", spanCtx.TraceID().String()))
		}
		if spanCtx.SpanID().IsValid() {
			fields = append(fields, zap.String("span_id", spanCtx.SpanID().String()))
		}
	}

	// Add correlation ID from context
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			fields = append(fields, zap.String("correlation_id", id))
		}
	}

	// Add session ID from context
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		if id, ok := sessionID.(string); ok {
			fields = append(fields, zap.String("session_id", id))
		}
	}

	if len(fields) > 0 {
		return &Logger{l.Logger.With(fields...)}
	}

	return l
}

func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

func (l *Logger) WithSessionID(sessionID string) *Logger {
	return &Logger{l.Logger.With(zap.String("session_id", sessionID))}
}

func (l *Logger) WithCorrelationID(correlationID string) *Logger {
	return &Logger{l.Logger.With(zap.String("correlation_id", correlationID))}
}

func (l *Logger) WithTraceID(traceID string) *Logger {
	return &Logger{l.Logger.With(zap.String("trace_id", traceID))}
}

func (l *Logger) WithSpanID(spanID string) *Logger {
	return &Logger{l.Logger.With(zap.String("span_id", spanID))}
}

// Mask sensitive data for logging
func (l *Logger) WithMaskedIDNumber(idNumber string) *Logger {
	if len(idNumber) > 4 {
		masked := idNumber[:2] + "****" + idNumber[len(idNumber)-2:]
		return &Logger{l.Logger.With(zap.String("id_number_masked", masked))}
	}
	return &Logger{l.Logger.With(zap.String("id_number_masked", "****"))}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
