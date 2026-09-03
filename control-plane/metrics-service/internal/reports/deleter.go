/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports

import (
	"context"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"k8s.io/utils/ptr"
)

type Deleter struct {
	waitGroup          WaitGrouper
	concurrencyLimitCh chan struct{}

	reportsUpdater ReportsUpdater
	reportsDeleter ReportsDeleter
	s3Client       S3Deleter
}

func NewDeleter(s3Client S3Deleter) *Deleter {
	conf := config.Get[*ReportsConfiguration]()
	reportsDao := dao.NewPostgresReportsDao()
	return NewTestDeleter(
		&sync.WaitGroup{},
		conf.DeleteConcurrencyLimit,
		reportsDao,
		reportsDao,
		s3Client,
	)
}

func NewTestDeleter(
	waitGroup WaitGrouper,
	concurrencyLimit int,
	reportsUpdater ReportsUpdater,
	reportsDeleter ReportsDeleter,
	s3Client S3Deleter,
) *Deleter {
	return &Deleter{
		waitGroup:          waitGroup,
		concurrencyLimitCh: make(chan struct{}, concurrencyLimit),
		reportsUpdater:     reportsUpdater,
		reportsDeleter:     reportsDeleter,
		s3Client:           s3Client,
	}
}

func (d *Deleter) ConcurrencyInUse() int {
	return len(d.concurrencyLimitCh)
}

func (d *Deleter) Start(ctx context.Context) {
	conf := config.Get[*ReportsConfiguration]()
	ticker := time.NewTicker(time.Duration(conf.RunPeriodSec) * time.Second)
	for {
		select {
		case <-ticker.C:
			d.processPendingDeleteReports(ctx)
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (d *Deleter) WaitStop() {
	d.waitGroup.Wait()
}

func (d *Deleter) fetchPendingDeleteReports(ctx context.Context) []models.Report {
	conf := config.Get[*ReportsConfiguration]()
	reports, err := d.reportsDeleter.GetPendingDeleteReports(ctx, conf.DeleteFetchLimit)
	if err != nil {
		logging.Errorf("failed to get pending-delete reports: %v", err)
		return []models.Report{}
	}

	return reports
}

func (d *Deleter) processPendingDeleteReports(ctx context.Context) {
	reports := d.fetchPendingDeleteReports(ctx)
	logging.Debugf("fetched %d reports for deletion", len(reports))
	for _, report := range reports {
		select {
		case d.concurrencyLimitCh <- struct{}{}:
			d.waitGroup.Add(1)
			go d.processDeletingReport(ctx, &report)
		case <-ctx.Done():
			return
		default:
			logging.Warnf("delete concurrency limit reached, skipping report %s", report.ID)
		}
	}
}

func (d *Deleter) processDeletingReport(ctx context.Context, report *models.Report) {
	logging.Debugf("deleting report %s", report.ID)
	d.deleteReport(ctx, report.TenantId, report.ID.String())
	logging.Debugf("report %s deletion attempt finished", report.ID)
	d.waitGroup.Done()
	<-d.concurrencyLimitCh
}

func (d *Deleter) deleteReport(ctx context.Context, tenantId int32, reportId string) {
	if dbErr := d.reportsUpdater.UpdateReportStatus(tenantId, reportId, api.REPORTSTATUS_DELETING, ptr.To(api.REPORTSTATUS_PENDING_DELETE)); dbErr != nil {
		logging.Errorf("failed to update report status to deleting, tenant_id [%d] report_id: [%s]. err: %v", tenantId, reportId, dbErr)
		return
	}

	if err := d.s3Client.Delete(reportId); err != nil {
		logging.Errorf("failed to delete report object from S3, tenant_id [%d] report_id: [%s]. err: %v", tenantId, reportId, err)
		if dbErr := d.reportsUpdater.UpdateReportStatusWithMessage(tenantId, reportId, api.REPORTSTATUS_PENDING_DELETE, err.Error(), ptr.To(api.REPORTSTATUS_DELETING)); dbErr != nil {
			logging.Warnf("failed to update report status message while deleting, tenant_id [%d] report_id: [%s]. err: %v", tenantId, reportId, dbErr)
		}
		return
	}

	if dbErr := d.reportsDeleter.DeleteReport(context.Background(), tenantId, reportId); dbErr != nil {
		logging.Errorf("failed to delete report from database, tenant_id [%d] report_id: [%s]. err: %v", tenantId, reportId, dbErr)
		return
	}

	logging.Debugf("deleted report, tenant_id [%d], report_id: [%s]", tenantId, reportId)
}
