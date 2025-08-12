# eKYC Backend - Monorepo

Backend monorepo cho hệ thống eKYC (electronic Know Your Customer) sử dụng Go 1.22 và microservices architecture.

## 🏗️ Kiến trúc

Hệ thống được thiết kế theo kiến trúc microservices với các thành phần chính:

### Services
- **API Gateway** (Port 8080): REST API endpoint, authentication, rate limiting
- **Identity** (Port 8081): Quản lý eKYC sessions, state machine
- **Document OCR** (Port 8082): Xử lý OCR documents
- **Face Match** (Port 8083): So sánh khuôn mặt
- **Liveness** (Port 8084): Kiểm tra liveness
- **Scoring** (Port 8085): Engine đánh giá và quyết định
- **Storage Service** (Port 8086): Quản lý file storage (MinIO)
- **Admin** (Port 8087): Giao diện quản trị

### Infrastructure
- **PostgreSQL**: Database chính
- **Redis**: Cache, session, idempotency
- **NATS**: Message bus, pub/sub
- **MinIO**: Object storage (S3-compatible)
- **OpenTelemetry**: Distributed tracing
- **Prometheus**: Metrics collection
- **Grafana**: Monitoring dashboards
- **Tempo**: Trace storage và query

## 🚀 Khởi chạy

### Yêu cầu hệ thống
- Docker & Docker Compose
- Go 1.22+
- Make (optional)

### Khởi chạy nhanh
```bash
# Clone repository
git clone <repository-url>
cd ekyc-backend

# Khởi chạy toàn bộ hệ thống
make dev

# Hoặc sử dụng docker compose trực tiếp
docker compose up --build
```

### Các lệnh Makefile
```bash
make help          # Hiển thị tất cả commands
make dev           # Khởi chạy development environment
make build         # Build tất cả services
make up            # Start services
make down          # Stop services
make clean         # Clean up containers và volumes
make test          # Chạy tests
make migrate       # Chạy database migrations
make health        # Kiểm tra health của services
make urls          # Hiển thị URLs của services
```

## 📊 Monitoring & Observability

### Dashboards
- **Backend Overview**: RPS, error rate, latency, NATS metrics
- **Traces**: Distributed tracing với Tempo
- **NATS Overview**: Message throughput, connections

### URLs
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Tempo**: http://localhost:3200
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)
- **NATS**: nats://localhost:4222

## 🔧 Development

### Cấu trúc thư mục
```
ekyc-backend/
├── pkg/                    # Shared packages
│   ├── config/            # Configuration management
│   ├── logger/            # Structured logging (zap)
│   ├── otel/              # OpenTelemetry setup
│   ├── httpmw/            # HTTP middlewares
│   ├── grpcmw/            # gRPC interceptors
│   ├── events/            # Event bus (NATS)
│   ├── storage/           # Redis & MinIO clients
│   ├── db/                # Database connection
│   └── errors/            # Error handling
├── services/               # Microservices
│   ├── api-gateway/       # REST API gateway
│   ├── identity/          # Session management
│   ├── doc-ocr/           # Document processing
│   ├── face-match/        # Face comparison
│   ├── liveness/          # Liveness detection
│   ├── scoring/           # Decision engine
│   ├── storage-svc/       # File storage
│   └── admin/             # Admin interface
├── migrations/             # Database migrations
├── deploy/                 # Infrastructure configs
│   ├── grafana/           # Dashboards & datasources
│   ├── prometheus/        # Prometheus config
│   ├── otel-collector/    # OpenTelemetry config
│   └── tempo/             # Tempo config
├── docker-compose.yml      # Infrastructure setup
├── Makefile               # Development commands
└── go.work               # Go workspace
```

### API Endpoints

#### Authentication
- `POST /api/v1/auth/signup` - Đăng ký user
- `POST /api/v1/auth/signin` - Đăng nhập

#### eKYC
- `POST /api/v1/ekyc/session` - Tạo session mới
- `POST /api/v1/ekyc/{id}/document` - Upload document
- `POST /api/v1/ekyc/{id}/selfie` - Upload selfie
- `POST /api/v1/ekyc/{id}/liveness` - Thực hiện liveness check
- `GET /api/v1/ekyc/{id}/status` - Lấy trạng thái session

#### Admin
- `GET /api/v1/admin/sessions` - Danh sách sessions
- `POST /api/v1/admin/sessions/{id}/decision` - Quyết định admin

### Middlewares
- **Request ID**: Tự động tạo unique request ID
- **Correlation ID**: Tracking request flow
- **Session ID**: User session management
- **Idempotency**: Tránh duplicate requests
- **Rate Limiting**: Token bucket algorithm
- **JWT Auth**: Bearer token authentication
- **Tracing**: OpenTelemetry integration
- **Logging**: Structured logging với correlation

### Events (NATS)
- `ocr.request/result` - OCR processing
- `face.request/result` - Face matching
- `liveness.request/result` - Liveness detection
- `kyc.decision` - KYC decision events
- `admin.decision` - Admin decision events
- `audit.log` - Audit trail

## 🧪 Testing

```bash
# Chạy tất cả tests
make test

# Chạy tests cho package cụ thể
cd pkg && go test -v ./...
cd services/api-gateway && go test -v ./...
```

## 📝 Database

### Migrations
```bash
# Sử dụng goose
go install github.com/pressly/goose/v3/cmd/goose@latest
cd migrations && goose postgres "postgres://postgres:postgres@localhost:5432/ekyc?sslmode=disable" up

# Hoặc sử dụng atlas
go install ariga.io/atlas/cmd/atlas@latest
atlas migrate apply --url "postgres://postgres:postgres@localhost:5432/ekyc?sslmode=disable"
```

### Schema
- `users` - User accounts
- `ekyc_sessions` - eKYC sessions
- `person_pii` - Personal information
- `ekyc_artifacts` - Uploaded files
- `ekyc_results` - Processing results
- `ekyc_decisions` - Admin decisions
- `audit_logs` - Audit trail

## 🔒 Security

- JWT authentication với HS256
- Rate limiting per IP
- Idempotency keys
- PII masking trong logs
- Secure headers

## 📈 Metrics

### Prometheus Metrics
- `http_requests_total` - Request count
- `http_request_duration_seconds` - Response time
- `nats_messages_published_total` - NATS messages
- `worker_job_duration_seconds` - Worker performance

### RED Metrics
- **Rate**: Requests per second
- **Errors**: Error rate percentage
- **Duration**: Response time percentiles (P50, P95, P99)

## 🚨 Troubleshooting

### Common Issues
1. **Port conflicts**: Kiểm tra ports đã được sử dụng
2. **Database connection**: Đảm bảo PostgreSQL đã sẵn sàng
3. **Memory issues**: Tăng Docker memory limit nếu cần

### Logs
```bash
# Xem logs của tất cả services
make logs

# Xem logs của service cụ thể
make logs-api-gateway
make logs-identity
```

### Health Check
```bash
# Kiểm tra health của tất cả services
make health
```

## 🤝 Contributing

1. Fork repository
2. Tạo feature branch
3. Commit changes
4. Push to branch
5. Tạo Pull Request

## 📄 License

MIT License - xem file LICENSE để biết thêm chi tiết.

## 📞 Support

- Issues: GitHub Issues
- Documentation: README này
- Team: eKYC Development Team
