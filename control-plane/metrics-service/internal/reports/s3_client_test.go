package reports

import (
	"fmt"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

func TestIsObjectNotFound(t *testing.T) {
	client := S3Client{}
	assert.True(t, client.isObjectNotFound(minio.ErrorResponse{Code: "NoSuchKey"}))
	assert.False(t, client.isObjectNotFound(fmt.Errorf("other error")))
}
