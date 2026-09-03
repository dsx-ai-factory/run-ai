/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/logger"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/logger/mocks"
	"github.com/stretchr/testify/require"
)

func TestVerbose(t *testing.T) {
	migLogger := &migrationLogger{}
	require.True(t, migLogger.Verbose())
}

func TestPrintf(t *testing.T) {
	migLogger := &migrationLogger{}

	loggerMock := mocks.NewLogger(t)
	loggerMock.On("Infof", "test %s", "message").Return()
	logger.SetLogger(loggerMock)
	defer logger.SetLogger(logger.NewDummyLogger())

	migLogger.Printf("test %s", "message")
}
