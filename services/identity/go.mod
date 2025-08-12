module github.com/ekyc-backend/services/identity

go 1.22

require (
	github.com/ekyc-backend/pkg v0.0.0
	github.com/gin-gonic/gin v1.9.1
	github.com/nats-io/nats.go v1.33.1
	google.golang.org/grpc v1.61.0
)

replace github.com/ekyc-backend/pkg => ../../pkg
