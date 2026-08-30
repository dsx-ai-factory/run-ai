package logging

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

type CtxLoggerHook func() (map[interface{}]string, context.Context)

type CtxHookAdapter struct {
	enrichFields map[interface{}]string
}

func (logHook *CtxHookAdapter) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}
	for ctxField, logField := range logHook.enrichFields {
		addStr(e, ctx, ctxField, logField)
	}
}

func addStr(e *zerolog.Event, ctx context.Context, contextField interface{}, fieldName string) {
	value := ctx.Value(contextField)
	switch ctx.Value(contextField).(type) {
	case string:
		e.Str(fieldName, value.(string))
	case []string:
		arr := strings.Join(value.([]string), ",")
		e.Str(fieldName, fmt.Sprintf("[%s]", arr))
	}
}

func NewLogDebugHook() *LogDebugHook {
	return &LogDebugHook{
		mu:                sync.Mutex{},
		ActualEventsCount: 0,
	}
}

type LogDebugHook struct {
	mu                sync.Mutex
	ActualEventsCount int // counts the actual events of the logger considering the log level (because this is a hook)
}

func (hook *LogDebugHook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	hook.mu.Lock()
	defer hook.mu.Unlock()

	hook.ActualEventsCount = hook.ActualEventsCount + 1
}

func GetDebugHookLogger(config *Config) (Logger, *LogDebugHook) {
	testLogger := Init(config).(*ConsoleLogger)
	debugAdapter := NewLogDebugHook()
	testLogger.Logger = testLogger.Logger.Hook(debugAdapter)
	return testLogger, debugAdapter
}

type LogDebugWriter struct {
	mu        sync.Mutex
	LogEvents []string
}

// Write implements the io.Writer interface.
func (adapter *LogDebugWriter) Write(p []byte) (n int, err error) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	// Capture the log line
	logEvent := string(p)

	// Append the log event to the array
	adapter.LogEvents = append(adapter.LogEvents, logEvent)
	return len(p), nil
}

func GetDebugWriter(config *Config) (Logger, *LogDebugWriter) {
	testLogger := Init(config).(*ConsoleLogger)
	debugAdapter := &LogDebugWriter{}
	debug := zerolog.MultiLevelWriter(debugAdapter)

	// Create a zerolog logger with the MultiLevelWriter
	logger := zerolog.New(debug).With().Timestamp().Logger()

	testLogger.Logger = logger
	return testLogger, debugAdapter
}
