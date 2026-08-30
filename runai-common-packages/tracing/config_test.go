package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig_ServiceNameOverride(t *testing.T) {
	t.Setenv("SERVICE_NAME", "test_service")
	config, _ := readConfig()
	assert.Equal(t, "test_service", config.ServiceName)
}

func TestReadConfig_CollectorEndpointOverride(t *testing.T) {
	t.Setenv("TRACING_COLLECTOR_ENDPOINT", "asdf")
	config, _ := readConfig()
	assert.Equal(t, "asdf", config.CollectorEndpoint)
}

func TestReadConfig_EnabledTrue(t *testing.T) {
	t.Setenv("TRACING_ENABLED", "true")
	config, _ := readConfig()
	assert.Equal(t, true, config.Enabled)
}
