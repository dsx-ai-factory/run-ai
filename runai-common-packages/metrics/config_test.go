package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig_CollectorEndpointOverride(t *testing.T) {
	t.Setenv("METRICS_EXPOSE_PORT", "123")
	config, _ := readConfig(nil)
	assert.Equal(t, 123, config.ExposePort)
}

func TestReadConfig_EnabledTrue(t *testing.T) {
	t.Setenv("TRACING_ENABLED", "true")
	config, _ := readConfig(nil)
	assert.Equal(t, true, config.Enabled)
}
