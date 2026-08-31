/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ingest

import (
	"testing"
	"time"

	"github.com/google/uuid"
	audit_config "github.com/run-ai/runai/control-plane/audit-service/internal/config"
	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/audit-service/internal/ingest/mocks"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestDbSinkHandler test that db will store to DB in 4 batches, while the last one is due to a shutdown
func TestDbSinkHandler(t *testing.T) {
	batcherMock := mocks.NewBatcherer[models.AuditLog](t)
	config.Set(&audit_config.BatcherConfig{
		MaxSize:     1,
		IntervalSec: 1,
	})
	defer config.FlushAll()
	dbSink := NewDbSinkHandler(batcherMock)

	batcherMock.On("Add", mock.Anything).Return(nil).Times(4)
	batcherMock.On("WaitDone").Return(true)

	addLogs(dbSink, 4)

	// wait for 1 interval, where all full batches are stored
	time.Sleep(2 * time.Second)
	// once stopped - only the partial batch (1 log) sent stored due to a max interval
	require.NoError(t, dbSink.Stop())
}

func addLogs(dbSink *DbSinkHandler, logCount int) {
	for i := 0; i < logCount; i++ {
		_ = dbSink.Handle(newAuditLog())
	}
}

func newAuditLog() *models.AuditLog {
	return &models.AuditLog{LogId: uuid.New()}
}
