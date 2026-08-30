package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAirgapLoadExecutionTimeoutDefault(t *testing.T) {
	flag := newAirgapLoadCmd().Flags().Lookup("execution-timeout")

	require.NotNil(t, flag)
	assert.Equal(t, "3h0m0s", flag.DefValue)
}
