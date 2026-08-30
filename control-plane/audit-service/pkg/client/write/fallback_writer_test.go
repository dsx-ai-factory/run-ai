package write

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	openapi "github.com/run-ai/runai/control-plane/audit-service/pkg/client/generated"
)

func TestFallbackWriter(t *testing.T) {
	mockPrimaryWriter := newMockWriter(t)
	mockFallbackWriter := newMockWriter(t)

	fallbackWriter := NewFallbackWriter(mockPrimaryWriter, mockFallbackWriter)

	require.NotNil(t, fallbackWriter.primary)
	require.NotNil(t, fallbackWriter.fallback)
}

func TestFallbackWriter_Write_PrimaryWrites(t *testing.T) {
	al := openapi.AuditLog{}

	mockPrimaryWriter := newMockWriter(t)
	mockPrimaryWriter.On("Write", al).Return(nil)
	mockFallbackWriter := newMockWriter(t)

	fallbackWriter := NewFallbackWriter(mockPrimaryWriter, mockFallbackWriter)

	err := fallbackWriter.Write(al)
	require.NoError(t, err)
}

func TestFallbackWriter_Write_PrimaryFailsButFallbackWrites(t *testing.T) {
	al := openapi.AuditLog{}

	mockPrimaryWriter := newMockWriter(t)
	mockPrimaryWriter.On("Write", al).Return(errors.New("failed to write"))
	mockFallbackWriter := newMockWriter(t)
	mockFallbackWriter.On("Write", al).Return(nil)

	fallbackWriter := NewFallbackWriter(mockPrimaryWriter, mockFallbackWriter)

	err := fallbackWriter.Write(al)
	require.NoError(t, err)
}

func TestFallbackWriter_Write_BothPrimaryAndFallbackFail(t *testing.T) {
	al := openapi.AuditLog{}

	mockPrimaryWriter := newMockWriter(t)
	mockPrimaryWriter.On("Write", al).Return(errors.New("failed to write"))
	mockFallbackWriter := newMockWriter(t)
	mockFallbackWriter.On("Write", al).Return(errors.New("failed to write"))

	fallbackWriter := NewFallbackWriter(mockPrimaryWriter, mockFallbackWriter)

	err := fallbackWriter.Write(al)
	require.EqualError(t, err, "failed to write")
}
