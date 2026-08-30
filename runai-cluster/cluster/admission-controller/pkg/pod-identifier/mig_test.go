package pod_identifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasMigDeviceRequest(t *testing.T) {
	annotations := map[string]string{
		"nvidia.com/mig": "3g.20gb",
	}

	assert.True(t, HasMigDeviceRequest(annotations))
}

func TestDoesntHaveMigDeviceRequest(t *testing.T) {
	annotations := map[string]string{}
	assert.False(t, HasMigDeviceRequest(annotations))
}
