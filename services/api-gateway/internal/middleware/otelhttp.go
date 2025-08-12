package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tracing middleware adds OpenTelemetry tracing
func Tracing(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Add custom attributes to span
				if span := trace.SpanFromContext(r.Context()); span != nil {
					span.SetAttributes(
						attribute.String("http.method", r.Method),
						attribute.String("http.url", r.URL.String()),
						attribute.String("http.user_agent", r.UserAgent()),
						attribute.String("http.remote_addr", r.RemoteAddr),
					)
				}
				next.ServeHTTP(w, r)
			}),
			serviceName,
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return r.Method + " " + r.URL.Path
			}),
		)
	}
}
