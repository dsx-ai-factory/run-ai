/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaictx_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/assert"
)

type logHook struct {
	logEvents []zerolog.Event
}

func (logHook *logHook) Run(logEvent *zerolog.Event, level zerolog.Level, message string) {
	logHook.logEvents = append(logHook.logEvents, *logEvent)
}

func TestWithContext_AddField(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "test")
	testLogger := logging.Init(nil).WithCtx(runaictx.LoggerHook(ctx)).(*logging.ConsoleLogger)
	logHook := &logHook{}
	testLogger.Logger = testLogger.Logger.Hook(logHook)

	testLogger.Infof("info info")

	buf := getEventBuf(logHook.logEvents[0])
	assert.Len(t, logHook.logEvents, 1)
	assert.Contains(t, buf, "tenant_id")
}

func getEventBuf(e zerolog.Event) string {
	reflectedE := reflect.ValueOf(e)
	buf := reflectedE.FieldByName("buf")
	return string(buf.Bytes())
}
