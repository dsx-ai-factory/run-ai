/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gormext

import (
	"context"
	"testing"
	"time"

	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"
	mocks "github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger/mocks"
	"github.com/stretchr/testify/mock"
)

func setupMockLogger(t *testing.T) *mocks.Logger {
	loggerMock := mocks.NewLogger(t)
	logger.SetLogger(loggerMock)
	t.Cleanup(func() {
		logger.SetLogger(logger.NewDummyLogger())
	})
	return loggerMock
}

func TestInfo(t *testing.T) {
	loggerMock := setupMockLogger(t)
	adapter := &loggerAdapter{}
	loggerMock.On("Infof", "db: info log: %s", "arg").Return()
	adapter.Info(context.Background(), "info log: %s", "arg")
}

func TestWarn(t *testing.T) {
	loggerMock := setupMockLogger(t)
	adapter := &loggerAdapter{}
	loggerMock.On("Warnf", "db: warn log: %s", "arg").Return()
	adapter.Warn(context.Background(), "warn log: %s", "arg")
}

func TestError(t *testing.T) {
	loggerMock := setupMockLogger(t)
	adapter := &loggerAdapter{}
	loggerMock.On("Errorf", "db: error log: %s", "arg").Return()
	adapter.Error(context.Background(), "error log: %s", "arg")
}

func TestTrace_DebugLevel_LogsSQL(t *testing.T) {
	loggerMock := setupMockLogger(t)
	adapter := &loggerAdapter{logLevel: dbconfig.LogLevelDebug}
	loggerMock.On("Debugf", "db: Executed SQL query: %s, rows affected: %d, time: %dms", "SELECT * FROM users", int64(5), mock.Anything).Return()

	adapter.Trace(context.Background(), time.Now(), func() (string, int64) { return "SELECT * FROM users", 5 }, nil)
}

func TestTrace_NotDebugLevel_DoesNotLog(t *testing.T) {
	setupMockLogger(t)
	adapter := &loggerAdapter{logLevel: dbconfig.LogLevelInfo}

	adapter.Trace(context.Background(), time.Now(), func() (string, int64) { return "SELECT * FROM users", 5 }, nil)
}
