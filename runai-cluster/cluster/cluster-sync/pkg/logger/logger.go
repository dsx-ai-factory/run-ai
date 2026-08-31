/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logger

import (
	"flag"
	"github.com/go-logr/zapr"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
)

func InitLogger() {
	var logLevel int
	flag.IntVar(&logLevel, "log-level", int(zapcore.InfoLevel), "Log level to use, -1 is Debug, 0 is Info and so forth")
	flag.Parse()
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zapcore.Level(logLevel))
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.Encoding = "console"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := zapConfig.Build()
	zap.ReplaceGlobals(logger)
	ctrl.SetLogger(zapr.NewLogger(logger))

	if logLevel == int(zapcore.InfoLevel) {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
