module github.com/ekyc-backend/services/storage-svc

go 1.22

require (
	github.com/ekyc-backend/pkg v0.0.0
	github.com/gin-gonic/gin v1.9.1
	github.com/minio/minio-go/v7 v7.0.69
)

replace github.com/ekyc-backend/pkg => ../../pkg
