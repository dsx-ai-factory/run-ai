package utils

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetK8sClientQPS(t *testing.T) {
	assert.Equal(t, defaultK8sClientQPS, GetK8sClientQPS(), "not set -> default")

	err := os.Setenv(EnvVarK8sClientQPS, "550")
	assert.NoError(t, err)
	assert.Equal(t, float32(550), GetK8sClientQPS(), "from env var")

	err = os.Setenv(EnvVarK8sClientQPS, "")
	assert.NoError(t, err)
	assert.Equal(t, defaultK8sClientQPS, GetK8sClientQPS(), "empty env var")

	err = os.Setenv(EnvVarK8sClientQPS, "xyz")
	assert.NoError(t, err)
	assert.Equal(t, defaultK8sClientQPS, GetK8sClientQPS(), "invalid vlaue env var")
}

func TestGetK8sClientBurst(t *testing.T) {
	assert.Equal(t, defaultK8sClientBurst, GetK8sClientBurst(), "not set -> default")

	err := os.Setenv(EnvVarK8sClientBurst, "550")
	assert.NoError(t, err)
	assert.Equal(t, 550, GetK8sClientBurst(), "from env var")

	err = os.Setenv(EnvVarK8sClientBurst, "")
	assert.NoError(t, err)
	assert.Equal(t, defaultK8sClientBurst, GetK8sClientBurst(), "empty env var")

	err = os.Setenv(EnvVarK8sClientBurst, "xyz")
	assert.NoError(t, err)
	assert.Equal(t, defaultK8sClientBurst, GetK8sClientBurst(), "invalid vlaue env var")
}
