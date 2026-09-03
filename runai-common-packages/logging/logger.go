/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logging

import (
	"github.com/rs/zerolog"
)

const patternFieldName = "log_pattern"

var globalLogger Logger

type Logger interface {
	Actual() interface{}
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	WithCtx(CtxLoggerHook) Logger
}

func NewConsoleLogger(zlogger zerolog.Logger) *ConsoleLogger {
	return &ConsoleLogger{
		Logger: zlogger,
	}
}

type ConsoleLogger struct {
	Logger zerolog.Logger
}

func (cl *ConsoleLogger) Actual() interface{} {
	return cl.Logger
}

func (cl *ConsoleLogger) Debugf(format string, args ...interface{}) {
	cl.Logger.Debug().Str(patternFieldName, format).Msgf(format, args...)
}

func (cl *ConsoleLogger) Infof(format string, args ...interface{}) {
	cl.Logger.Info().Str(patternFieldName, format).Msgf(format, args...)
}

func (cl *ConsoleLogger) Warnf(format string, args ...interface{}) {
	cl.Logger.Warn().Str(patternFieldName, format).Msgf(format, args...)
}

func (cl *ConsoleLogger) Errorf(format string, args ...interface{}) {
	cl.Logger.Error().Str(patternFieldName, format).Msgf(format, args...)
}

func (cl *ConsoleLogger) Fatalf(format string, args ...interface{}) {
	cl.Logger.Fatal().Str(patternFieldName, format).Msgf(format, args...)
}

// gets:
//
//	mapping - ctxkey: logkey mapping, convert field from ctx to logger
//	ctx - the context to get the values from
//
// return: a new contextual logger
func (cl *ConsoleLogger) WithCtx(hook CtxLoggerHook) Logger {
	mapping, ctx := hook()
	l := cl.Logger.Hook(&CtxHookAdapter{enrichFields: mapping}).With().Ctx(ctx).Logger()
	return NewConsoleLogger(l)
}
