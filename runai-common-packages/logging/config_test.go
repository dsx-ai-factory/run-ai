package logging

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfiguration_FromEnv(t *testing.T) {
	os.Setenv("LOG_LEVEL", "error")
	config := readConfig()

	assert.Equal(t, config.LogLevel, "error")

	os.Clearenv()
}
