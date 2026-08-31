/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kafka

import "github.com/run-ai/runai/runai-common-packages/runaipubsub/logger"

type Logger struct {
	logger logger.Loggerer
}

func NewLogger(logger logger.Loggerer) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

type ErrorLogger struct {
	logger logger.Loggerer
}

func NewErrorLogger(logger logger.Loggerer) *ErrorLogger {
	return &ErrorLogger{
		logger: logger,
	}
}

func (el *ErrorLogger) Printf(format string, args ...interface{}) {
	el.logger.Errorf(format, args...)
}
