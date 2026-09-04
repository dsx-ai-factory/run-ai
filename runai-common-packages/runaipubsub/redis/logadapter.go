/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package redis

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/logger"
	"github.com/ThreeDotsLabs/watermill"
)

func NewLogger(l logger.Loggerer) *Logger {
	return &Logger{
		Loggerer: l,
	}
}

type Logger struct {
	logger.Loggerer
}

func (l *Logger) Error(msg string, err error, fields watermill.LogFields) {
	if strings.Contains(msg, "read fail") && strings.Contains(err.Error(), "EOF") {
		l.Infof("%s, error: %s. %+v", msg, err.Error(), fields)
		return
	}

	l.Errorf("%s, error: %s. %+v", msg, err.Error(), fields)
}

func (l *Logger) Info(msg string, fields watermill.LogFields) {
	l.Infof("%s: %+v", msg, fields)
}

func (l *Logger) Debug(msg string, fields watermill.LogFields) {
	l.Debugf("%s: %+v", msg, fields)
}

func (l *Logger) Trace(msg string, fields watermill.LogFields) {}

func (l *Logger) Warn(msg string, fields watermill.LogFields) {
	l.Warnf(fmt.Sprintf("%s: %+v", msg, fields))
}

func (l *Logger) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return l
}
