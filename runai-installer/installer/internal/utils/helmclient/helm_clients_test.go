package helmclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOCIRegistryHost(t *testing.T) {
	assert.Equal(t, "nvcr.io", ociRegistryHost("oci://nvcr.io/nvidia/runai/control-plane"))
	assert.Equal(t, "nvcr.io", ociRegistryHost("nvcr.io/nvidia/runai/control-plane"))
	assert.Equal(t, "ghcr.io", ociRegistryHost("oci://ghcr.io/run-ai/chart"))
	assert.Equal(t, "nvcr.io", ociRegistryHost("oci://nvcr.io"))
}

func TestChartAuth_IsZero(t *testing.T) {
	assert.True(t, ChartAuth{}.IsZero())
	assert.False(t, ChartAuth{Username: "$oauthtoken"}.IsZero())
	assert.False(t, ChartAuth{Password: "secret"}.IsZero())
}
