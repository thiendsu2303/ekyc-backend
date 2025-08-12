module github.com/ekyc-backend/services/liveness

go 1.22

require (
	github.com/ekyc-backend/pkg v0.0.0
	github.com/nats-io/nats.go v1.33.1
)

replace github.com/ekyc-backend/pkg => ../../pkg
