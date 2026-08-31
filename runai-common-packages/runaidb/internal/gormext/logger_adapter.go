/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gormext

import (
	"context"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"
	gormlogger "gorm.io/gorm/logger"
)

// loggerAdapter implements GORM's logger.Interface interface
type loggerAdapter struct {
	logLevel dbconfig.LogLevel
}

func NewLoggerAdapter() *loggerAdapter {
	dbConfig := config.Get[*dbconfig.Config]()
	return &loggerAdapter{
		logLevel: dbConfig.LogLevel,
	}
}

func (l *loggerAdapter) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *loggerAdapter) Info(ctx context.Context, msg string, args ...any) {
	logger.GetLogger().Infof("db: "+msg, args...)
}

func (l *loggerAdapter) Warn(ctx context.Context, msg string, args ...any) {
	logger.GetLogger().Warnf("db: "+msg, args...)
}

func (l *loggerAdapter) Error(ctx context.Context, msg string, args ...any) {
	logger.GetLogger().Errorf("db: "+msg, args...)
}

func (l *loggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel != dbconfig.LogLevelDebug {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	elapsedMs := elapsed.Milliseconds()

	logger.GetLogger().Debugf("db: Executed SQL query: %s, rows affected: %d, time: %dms", sql, rows, elapsedMs)
}
