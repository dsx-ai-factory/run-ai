package logging_test

import (
	"encoding/json"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/stretchr/testify/assert"
)

func TestInfo_Log3Entries(t *testing.T) {
	config := &logging.Config{LogLevel: "info"}
	logger, logHook := logging.GetDebugHookLogger(config)

	logger.Debugf("debug debug")
	logger.Infof("info info")
	logger.Warnf("warning warning")
	logger.Errorf("error error")

	assert.Equal(t, 3, logHook.ActualEventsCount)
}

func TestWarn_Log2Entries(t *testing.T) {
	config := &logging.Config{LogLevel: "warn"}
	logger, logHook := logging.GetDebugHookLogger(config)

	logger.Debugf("debug debug")
	logger.Infof("info info")
	logger.Warnf("warning warning")
	logger.Errorf("error error")

	assert.Equal(t, 2, logHook.ActualEventsCount)
}

func TestError_Log1Entries(t *testing.T) {
	config := &logging.Config{LogLevel: "error"}
	logger, logHook := logging.GetDebugHookLogger(config)

	logger.Debugf("debug debug")
	logger.Infof("info info")
	logger.Warnf("warning warning")
	logger.Errorf("error error")

	assert.Equal(t, 1, logHook.ActualEventsCount)
}

func TestDebug_Log4Entries(t *testing.T) {
	config := &logging.Config{LogLevel: "debug"}
	logger, logHook := logging.GetDebugHookLogger(config)

	logger.Debugf("debug debug")
	logger.Infof("info info")
	logger.Warnf("warning warning")
	logger.Errorf("error error")

	assert.Equal(t, 4, logHook.ActualEventsCount)
}

func TestPattern_Aggregation(t *testing.T) {
	logger, writer := logging.GetDebugWriter(&logging.Config{LogLevel: "info"})

	logger.Infof("User %s logged in", "alice")
	logger.Infof("User %s logged in", "bob")
	logger.Infof("User %s failed to login", "adel")

	assert.Len(t, writer.LogEvents, 3)

	var patterns []string
	for _, raw := range writer.LogEvents {
		var event map[string]interface{}
		assert.NoError(t, json.Unmarshal([]byte(raw), &event))
		p, _ := event["log_pattern"].(string)
		patterns = append(patterns, p)
	}

	assert.Equal(t, "User %s logged in", patterns[0])
	assert.Equal(t, patterns[0], patterns[1])
	assert.Equal(t, "User %s failed to login", patterns[2])
}
