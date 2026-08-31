/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres

import "github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"

// migrationLogger implements the golang-migrate.Logger interface
type migrationLogger struct{}

func (l *migrationLogger) Printf(format string, v ...interface{}) {
	logger.GetLogger().Infof(format, v...)
}

func (l *migrationLogger) Verbose() bool {
	return true
}
