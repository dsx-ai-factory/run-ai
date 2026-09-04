/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"context"
	"fmt"
	"time"

	audit_config "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/schedule"
)

const (
	cronJobName                = "audit-log-cleaner"
	deleteAuditLogsQueryFormat = "DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL '%d' DAY"
)

func InitCleaner(ctx context.Context) {
	conf := config.Get[*audit_config.ServiceConfig]()
	if !conf.CleanDbEnabled {
		return
	}
	logging.Infof("Initializing DB cleaner with interval: %d hours, TTL: %d days\n", conf.CleanIntervalHours, conf.CleanAuditTtlDays)
	client := runaidb.DB()
	schedule.NewJob(client, cronJobName, time.Duration(conf.CleanIntervalHours)*time.Hour, []string{getDeleteAuditQuery()}).
		WithStartTime(getCleanerStartTime()).
		Schedule(ctx)
}

func getDeleteAuditQuery() string {
	conf := config.Get[*audit_config.ServiceConfig]()
	return fmt.Sprintf(deleteAuditLogsQueryFormat, conf.CleanAuditTtlDays)
}

func getCleanerStartTime() time.Time {
	now := time.Now()
	threeAM := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	return threeAM
}
