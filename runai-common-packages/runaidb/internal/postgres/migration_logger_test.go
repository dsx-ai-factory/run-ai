package postgres

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger/mocks"
	"github.com/stretchr/testify/require"
)

func TestVerbose(t *testing.T) {
	migLogger := &migrationLogger{}
	require.True(t, migLogger.Verbose())
}

func TestPrintf(t *testing.T) {
	migLogger := &migrationLogger{}

	loggerMock := mocks.NewLogger(t)
	loggerMock.On("Infof", "test %s", "message").Return()
	logger.SetLogger(loggerMock)
	defer logger.SetLogger(logger.NewDummyLogger())

	migLogger.Printf("test %s", "message")
}
