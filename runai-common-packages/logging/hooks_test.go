package logging_test

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/stretchr/testify/assert"
)

func TestWithContext_AddField(t *testing.T) {
	ctx := context.WithValue(context.Background(), "mykey", "myval")
	testLogger, logHook := logging.GetDebugWriter(&logging.Config{LogLevel: "debug"})

	testLogger.WithCtx(func() (map[interface{}]string, context.Context) {
		return map[interface{}]string{
			"mykey": "mylogkey",
		}, ctx
	}).Infof("info info")

	assert.Len(t, logHook.LogEvents, 1)
	assert.Contains(t, logHook.LogEvents[0], "mylogkey")
}

func TestWithContext_AddMultipleField(t *testing.T) {
	ctx := context.WithValue(context.Background(), "mykey", "myval")
	ctx = context.WithValue(ctx, "mykey2", "myval2")
	testLogger, logHook := logging.GetDebugWriter(&logging.Config{LogLevel: "debug"})

	testLogger.WithCtx(func() (map[interface{}]string, context.Context) {
		return map[interface{}]string{
			"mykey":  "mylogkey",
			"mykey2": "mylogkey2",
		}, ctx
	}).Infof("info info")

	assert.Len(t, logHook.LogEvents, 1)
	assert.Contains(t, logHook.LogEvents[0], "mylogkey")
	assert.Contains(t, logHook.LogEvents[0], "mylogkey2")
}

func TestWithContext_NilFieldIgnored(t *testing.T) {
	ctx := context.WithValue(context.Background(), "mykey", nil)
	testLogger, logHook := logging.GetDebugWriter(&logging.Config{LogLevel: "debug"})

	testLogger.WithCtx(func() (map[interface{}]string, context.Context) {
		return map[interface{}]string{
			"mykey": "mylogkey",
		}, ctx
	}).Infof("info info")

	assert.Len(t, logHook.LogEvents, 1)
	assert.NotContains(t, logHook.LogEvents[0], "mylogkey")
}

func TestWithContext_AddArrField(t *testing.T) {
	ctx := context.WithValue(context.Background(), "mykey", []string{"myval", "asdf"})

	testLogger, logHook := logging.GetDebugWriter(&logging.Config{LogLevel: "debug"})

	testLogger.WithCtx(func() (map[interface{}]string, context.Context) {
		return map[interface{}]string{
			"mykey": "mylogkey",
		}, ctx
	}).Infof("info info")

	assert.Len(t, logHook.LogEvents, 1)
	assert.Contains(t, logHook.LogEvents[0], "[myval,asdf]")
}
