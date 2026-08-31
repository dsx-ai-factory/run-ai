/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reporter

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"go.uber.org/zap"
)

type LoggerReporter struct {
	logger *zap.Logger
}

func NewLoggerReporter(logger *zap.Logger) *LoggerReporter {
	return &LoggerReporter{logger: logger}
}

func (cr *LoggerReporter) Report(t *transaction.Txn) error {
	cr.logger.Info("", zap.Any("info", t.GetInfo()), zap.Any("metrics", t.GetMetrics()))
	return nil
}
