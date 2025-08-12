.PHONY: help dev build up down clean lint test migrate proto openapi seed

# Default target
help:
	@echo "Available targets:"
	@echo "  dev      - Start development environment (docker compose up --build)"
	@echo "  build    - Build all services"
	@echo "  up       - Start services (docker compose up)"
	@echo "  down     - Stop services (docker compose down)"
	@echo "  clean    - Clean up containers, volumes, and images"
	@echo "  lint     - Run golangci-lint (if available)"
	@echo "  test     - Run unit and integration tests"
	@echo "  migrate  - Run database migrations"
	@echo "  proto    - Generate gRPC code from protobuf"
	@echo "  openapi  - Generate/validate REST stubs from OpenAPI"
	@echo "  seed     - Create demo user and data"

# Development environment
dev:
	@echo "Starting development environment..."
	docker compose up --build

# Build all services
build:
	@echo "Building all services..."
	docker compose build

# Start services
up:
	@echo "Starting services..."
	docker compose up -d

# Stop services
down:
	@echo "Stopping services..."
	docker compose down

# Clean up everything
clean:
	@echo "Cleaning up..."
	docker compose down -v --remove-orphans
	docker system prune -f
	docker volume prune -f

# Lint code (if golangci-lint is available)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install it first:"; \
		echo "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run tests
test:
	@echo "Running tests..."
	@cd pkg && go test -v ./...
	@cd services/api-gateway && go test -v ./...
	@cd services/identity && go test -v ./...
	@cd services/doc-ocr && go test -v ./...
	@cd services/face-match && go test -v ./...
	@cd services/liveness && go test -v ./...
	@cd services/scoring && go test -v ./...
	@cd services/storage-svc && go test -v ./...
	@cd services/admin && go test -v ./...

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@if command -v goose >/dev/null 2>&1; then \
		cd migrations && goose postgres "postgres://postgres:postgres@localhost:5432/ekyc?sslmode=disable" up; \
	elif command -v atlas >/dev/null 2>&1; then \
		atlas migrate apply --url "postgres://postgres:postgres@localhost:5432/ekyc?sslmode=disable"; \
	else \
		echo "Neither goose nor atlas found. Install one of them:"; \
		echo "go install github.com/pressly/goose/v3/cmd/goose@latest"; \
		echo "or"; \
		echo "go install ariga.io/atlas/cmd/atlas@latest"; \
	fi

# Generate gRPC code from protobuf
proto:
	@echo "Generating gRPC code..."
	@if command -v protoc >/dev/null 2>&1; then \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			pkg/contracts/proto/*.proto; \
	else \
		echo "protoc not found. Install Protocol Buffers compiler first."; \
	fi

# Generate/validate REST stubs from OpenAPI
openapi:
	@echo "Generating REST stubs from OpenAPI..."
	@if command -v openapi-generator-cli >/dev/null 2>&1; then \
		openapi-generator-cli generate -i pkg/contracts/openapi.yaml \
			-g go-server -o services/api-gateway/generated; \
	else \
		echo "openapi-generator-cli not found. Install it first:"; \
		echo "npm install -g @openapitools/openapi-generator-cli"; \
	fi

# Create demo user and data
seed:
	@echo "Creating demo data..."
	@echo "This will create a demo user and sample eKYC session"
	@echo "Make sure the database is running and accessible"
	@echo "You can run this after starting the services with 'make dev'"

# Wait for database to be ready
wait-db:
	@echo "Waiting for database to be ready..."
	@until docker compose exec -T postgres pg_isready -U postgres; do \
		echo "Database not ready, waiting..."; \
		sleep 2; \
	done
	@echo "Database is ready!"

# Show logs
logs:
	docker compose logs -f

# Show logs for specific service
logs-%:
	docker compose logs -f $(subst logs-,,$@)

# Scale services
scale-%:
	docker compose up --scale $(subst scale-,,$@)=2 -d

# Health check
health:
	@echo "Checking service health..."
	@curl -f http://localhost:8080/health || echo "API Gateway: ❌"
	@curl -f http://localhost:8081/health || echo "Identity: ❌"
	@curl -f http://localhost:8082/health || echo "Doc OCR: ❌"
	@curl -f http://localhost:8083/health || echo "Face Match: ❌"
	@curl -f http://localhost:8084/health || echo "Liveness: ❌"
	@curl -f http://localhost:8085/health || echo "Scoring: ❌"
	@curl -f http://localhost:8086/health || echo "Storage: ❌"
	@curl -f http://localhost:8087/health || echo "Admin: ❌"
	@curl -f http://localhost:9090/-/healthy || echo "Prometheus: ❌"
	@curl -f http://localhost:3000/api/health || echo "Grafana: ❌"
	@curl -f http://localhost:3200/ready || echo "Tempo: ❌"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install ariga.io/atlas/cmd/atlas@latest
	@echo "Tools installed successfully!"

# Show service URLs
urls:
	@echo "Service URLs:"
	@echo "  API Gateway:    http://localhost:8080"
	@echo "  Identity:       http://localhost:8081"
	@echo "  Doc OCR:        http://localhost:8082"
	@echo "  Face Match:     http://localhost:8083"
	@echo "  Liveness:       http://localhost:8084"
	@echo "  Scoring:        http://localhost:8085"
	@echo "  Storage:        http://localhost:8086"
	@echo "  Admin:          http://localhost:8087"
	@echo "  Prometheus:     http://localhost:9090"
	@echo "  Grafana:        http://localhost:3000 (admin/admin)"
	@echo "  Tempo:          http://localhost:3200"
	@echo "  MinIO Console:  http://localhost:9001 (minioadmin/minioadmin)"
	@echo "  NATS:           nats://localhost:4222"
