package write

import (
	"errors"
	"testing"
	"time"

	auditconfig "github.com/run-ai/runai/control-plane/audit-service/pkg/client/config"
	openapi "github.com/run-ai/runai/control-plane/audit-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/audit-service/pkg/client/mocks"
	"github.com/run-ai/runai/runai-common-packages/config"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAPIWriter(t *testing.T) {
	mockClient := mocks.NewIAuditClient(t)
	apiWriter := APIWriter{
		client:          mockClient,
		maxWriteRetries: 10,
		writeRetryDelay: 5 * time.Second,
	}

	require.NotNil(t, apiWriter.client)
	require.Equal(t, uint(10), apiWriter.maxWriteRetries)
	require.Equal(t, 5*time.Second, apiWriter.writeRetryDelay)
}

func TestAPIWriter_Write(t *testing.T) {
	al := openapi.AuditLog{}
	mockClient := mocks.NewIAuditClient(t)
	mockClient.On("SaveAuditLog", mock.Anything, al).Return(nil)
	apiWriter := APIWriter{
		client:          mockClient,
		maxWriteRetries: 1,
	}

	err := apiWriter.Write(al)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // wait for the go routine to finish
}

func TestAPIWriter_ShouldRetryOnSaveAuditLogError(t *testing.T) {
	al := openapi.AuditLog{}
	mockClient := mocks.NewIAuditClient(t)
	mockClient.On("SaveAuditLog", mock.Anything, al).Return(errors.New("failed to save audit log")).Once()
	mockClient.On("SaveAuditLog", mock.Anything, al).Return(nil).Once()
	apiWriter := &APIWriter{
		client:          mockClient,
		maxWriteRetries: 2,
		writeRetryDelay: 10 * time.Millisecond,
	}

	err := apiWriter.Write(al)
	require.NoError(t, err) // function should still return nil, since we're using a go routine

	time.Sleep(100 * time.Millisecond) // wait for the go routine to finish
}

func TestInitAuditClient(t *testing.T) {
	defer config.FlushAll()
	err := config.Set(&auditconfig.Config{
		AuditServiceUrl: "http://test-audit-service:8080",
	})
	require.NoError(t, err)
	client := initAuditClient()
	require.NotNil(t, client)
}
