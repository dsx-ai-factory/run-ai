/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func Init(config *Config) Logger {
	if config == nil {
		config = readConfig()
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false, TimeFormat: time.RFC3339})
	if !config.Pretty {
		logger = zerolog.New(os.Stdout)
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		logger.Warn().Str("input_level", config.LogLevel).Msg("Failed to parse log using default info level")
		level = zerolog.InfoLevel
	}
	logger = logger.Level(level).With().Timestamp().Logger()

	globalLogger = NewConsoleLogger(logger)
	return globalLogger
}

func Debugf(format string, args ...interface{}) {
	if globalLogger == nil {
		return
	}
	globalLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	if globalLogger == nil {
		return
	}
	globalLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	if globalLogger == nil {
		return
	}
	globalLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	if globalLogger == nil {
		return
	}
	globalLogger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	if globalLogger == nil {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
		os.Exit(1)
	}
	globalLogger.Fatalf(format, args...)
}

func Actual() Logger {
	if globalLogger == nil {
		return nil
	}
	return globalLogger
}

// WithCtx - returns a new contextual logger
func WithCtx(hook CtxLoggerHook) Logger {
	if globalLogger == nil {
		globalLogger = Init(nil)
	}
	return globalLogger.WithCtx(hook)
}
