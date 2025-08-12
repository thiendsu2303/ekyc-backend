# API Gateway Service

API Gateway là entrypoint REST cho hệ thống eKYC, cung cấp authentication, rate limiting, request validation, và routing đến các service nội bộ.

## 🎯 Mục tiêu & Vai trò

- **Entrypoint REST**: Xử lý tất cả HTTP requests từ client và admin
- **Authentication & Authorization**: JWT-based auth với role-based access control
- **Rate Limiting**: Token bucket algorithm với Redis backend
- **Request Validation**: Validate input theo OpenAPI schema
- **Idempotency**: Đảm bảo operations không duplicate với Redis cache
- **Observability**: Logging, metrics, và tracing đầy đủ
- **Fan-out**: Route requests đến các service nội bộ qua gRPC

## 🏗️ Kiến trúc

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client/Admin  │───▶│  API Gateway   │───▶│  gRPC Services  │
│                 │    │                 │    │                 │
│   HTTP/JSON     │    │   Chi Router    │    │   Identity      │
│                 │    │   Middleware    │    │   Storage       │
│                 │    │   Handlers      │    │   Admin         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📁 Cấu trúc thư mục

```
services/api-gateway/
├── Dockerfile                 # Container build
├── main.go                    # Entry point
├── openapi.yaml              # OpenAPI specification
├── internal/
│   ├── config/               # Configuration management
│   │   └── config.go
│   ├── security/             # JWT authentication
│   │   └── jwt.go
│   ├── metrics/              # Prometheus metrics
│   │   └── prometheus.go
│   ├── clients/              # gRPC clients
│   │   ├── identity.go
│   │   ├── storage.go
│   │   └── admin.go
│   ├── middleware/           # HTTP middleware
│   │   ├── requestid.go
│   │   ├── correlation.go
│   │   ├── logging.go
│   │   ├── recover.go
│   │   ├── ratelimit.go
│   │   ├── idempotency.go
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── otelhttp.go
│   ├── handlers/             # HTTP handlers
│   │   └── health.go
│   └── server/               # Server setup
│       ├── server.go
│       ├── response.go
│       └── validator.go
```

## 🚀 Tính năng chính

### 1. **Authentication & Authorization**
- JWT-based authentication với HS256 signing
- Role-based access control (USER, ADMIN)
- Automatic token validation và user context injection

### 2. **Rate Limiting**
- Token bucket algorithm
- Per-IP và per-route rate limiting
- Redis backend cho distributed environments
- Configurable RPS và burst limits

### 3. **Idempotency**
- Redis-based idempotency key storage
- 15-minute TTL cho cached responses
- Automatic duplicate request handling
- Support cho tất cả POST operations

### 4. **Request Validation**
- OpenAPI schema validation
- Input sanitization và type checking
- Detailed error messages với correlation IDs

### 5. **Observability**
- Structured JSON logging với Zap
- Prometheus metrics (RED metrics)
- OpenTelemetry tracing
- Health check endpoints (/live, /ready)

### 6. **Security**
- CORS policy configuration
- Request body size limits
- Panic recovery middleware
- Secure headers

## 🔧 Cấu hình

### Environment Variables

```bash
# Service Configuration
SERVICE_NAME=api-gateway
HTTP_PORT=8080

# CORS
ALLOW_ORIGINS=*

# JWT
JWT_SECRET=your-secret-key

# Rate Limiting
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20

# Redis
REDIS_URL=redis://redis:6379

# gRPC Services
IDENTITY_GRPC_ADDR=identity:9090
STORAGE_GRPC_ADDR=storage-svc:9092
ADMIN_GRPC_ADDR=admin:9093

# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317

# Prometheus
PROMETHEUS_METRICS_PATH=/metrics

# Security
MAX_REQUEST_BODY_SIZE=2097152  # 2MB
REQUEST_TIMEOUT=5s
```

## 📡 API Endpoints

### Health Check
- `GET /live` - Process health check
- `GET /ready` - Service readiness check
- `GET /metrics` - Prometheus metrics

### Authentication
- `POST /api/v1/auth/signup` - User registration
- `POST /api/v1/auth/signin` - User authentication

### eKYC Operations
- `POST /api/v1/ekyc/session` - Create eKYC session
- `POST /api/v1/ekyc/{id}/document` - Document upload/presign
- `POST /api/v1/ekyc/{id}/selfie` - Selfie upload/presign
- `POST /api/v1/ekyc/{id}/liveness` - Liveness check upload/presign
- `GET /api/v1/ekyc/{id}/status` - Get session status

### Admin Operations
- `GET /api/v1/admin/sessions` - List sessions
- `GET /api/v1/admin/sessions/{id}` - Get session details
- `POST /api/v1/admin/sessions/{id}/decision` - Apply admin decision

## 🔄 Middleware Stack

Middleware được áp dụng theo thứ tự sau:

1. **Recover** - Panic recovery
2. **RequestID** - Unique request identification
3. **CorrelationID** - Request correlation tracking
4. **Tracing** - OpenTelemetry tracing
5. **Logging** - Structured request logging
6. **CORS** - Cross-origin resource sharing
7. **RateLimit** - Rate limiting
8. **Idempotency** - Idempotent operations
9. **Auth** - JWT authentication (cho protected routes)

## 📊 Metrics

### HTTP Metrics (RED)
- `gateway_http_requests_total` - Request count
- `gateway_http_request_duration_seconds` - Request duration
- `gateway_http_request_errors_total` - Error count

### Custom Metrics
- `gateway_idempotency_hits_total` - Idempotency cache hits
- `gateway_rate_limited_total` - Rate limited requests
- `gateway_active_connections` - Active connections

### gRPC Client Metrics
- `gateway_grpc_client_requests_total` - gRPC request count
- `gateway_grpc_client_duration_seconds` - gRPC request duration
- `gateway_grpc_client_errors_total` - gRPC error count

## 🚀 Development

### Prerequisites
- Go 1.22+
- Redis
- gRPC services (identity, storage, admin)

### Local Development
```bash
cd services/api-gateway

# Install dependencies
go mod tidy

# Run service
go run main.go

# Run tests
go test ./...

# Build binary
go build -o api-gateway main.go
```

### Docker
```bash
# Build image
docker build -t api-gateway .

# Run container
docker run -p 8080:8080 api-gateway
```

## 🧪 Testing

### Unit Tests
```bash
go test ./internal/...
```

### Integration Tests
```bash
# Requires Redis and gRPC services
go test -tags=integration ./...
```

### Load Testing
```bash
# Using k6
k6 run load-test.js
```

## 📝 Logging

### Log Format
```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "api-gateway",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "span_id": "12345678-1234-1234-1234-123456789abc",
  "correlation_id": "corr-123",
  "session_id": "session-456",
  "message": "HTTP Request",
  "method": "POST",
  "path": "/api/v1/ekyc/session",
  "status": 201,
  "duration_ms": 45,
  "remote_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "request_id": "req-789",
  "user_id": "user-123"
}
```

### Log Levels
- **DEBUG** - Detailed debugging information
- **INFO** - General operational messages
- **WARN** - Warning messages
- **ERROR** - Error conditions
- **FATAL** - Fatal errors (service shutdown)

## 🔍 Monitoring

### Health Checks
- **Liveness**: Process health (`/live`)
- **Readiness**: Service dependencies health (`/ready`)

### Metrics Dashboard
- Prometheus metrics endpoint (`/metrics`)
- Grafana dashboards cho monitoring

### Tracing
- OpenTelemetry traces
- Tempo integration cho distributed tracing

## 🚨 Troubleshooting

### Common Issues

1. **Rate Limiting Errors (429)**
   - Check Redis connectivity
   - Verify rate limit configuration
   - Monitor request patterns

2. **Authentication Errors (401)**
   - Verify JWT_SECRET configuration
   - Check token expiration
   - Validate token format

3. **gRPC Connection Errors**
   - Verify service addresses
   - Check network connectivity
   - Monitor service health

4. **Redis Connection Issues**
   - Verify REDIS_URL
   - Check Redis service health
   - Monitor Redis memory usage

### Debug Mode
```bash
# Enable debug logging
export LOG_LEVEL=debug

# Enable OpenTelemetry debug
export OTEL_LOG_LEVEL=debug
```

## 📚 References

- [Chi Router Documentation](https://github.com/go-chi/chi)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)
- [Zap Logger](https://github.com/uber-go/zap)
- [JWT Go](https://github.com/golang-jwt/jwt)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.
