module github.com/ekyc-backend/services/api-gateway

go 1.22

require (
	github.com/ekyc-backend/pkg v0.0.0
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/prometheus/client_golang v1.18.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.48.0
)

replace github.com/ekyc-backend/pkg => ../../pkg
