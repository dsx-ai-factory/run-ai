/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/dao"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
)

const (
	reportsCountPerStatus = "reports_count_per_status"
)

type Processor struct {
	FetchLimit         int
	RunPeriod          int
	WaitGroup          WaitGrouper
	ConcurrencyLimitCh chan struct{}

	Collector     Collectorer
	ReportsGetter PendingReportsGetter
}

func NewProcessor(collector Collectorer) *Processor {
	conf := config.Get[*ReportsConfiguration]()
	return &Processor{
		FetchLimit: conf.FetchLimit,
		RunPeriod:  conf.RunPeriodSec,

		WaitGroup:          &sync.WaitGroup{},
		ConcurrencyLimitCh: make(chan struct{}, conf.ConcurrencyLimit),
		ReportsGetter:      dao.NewPostgresReportsDao(),
		Collector:          collector,
	}
}

func (p *Processor) Start(ctx context.Context) {
	metrics.GaugeVec(reportsCountPerStatus).Register("Number of reports grouped by their current status", []string{"status"})
	ticker := time.NewTicker(time.Duration(p.RunPeriod) * time.Second)
	for {
		select {
		case <-ticker.C:
			p.processReports(ctx)
			go p.writeMetricsData(ctx)
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (p *Processor) WaitStop() {
	p.WaitGroup.Wait()
}

func (p *Processor) fetchPendingReports(ctx context.Context) []models.Report {
	reports, err := p.ReportsGetter.GetPendingReports(ctx, p.FetchLimit)
	if err != nil {
		logging.Errorf("failed to get reports: %v", err)
		return []models.Report{}
	}

	return reports
}

func (p *Processor) processReports(ctx context.Context) {
	reports := p.fetchPendingReports(ctx)
	logging.Debugf("fetched %d reports", len(reports))
	for _, report := range reports {
		select {
		case p.ConcurrencyLimitCh <- struct{}{}:
			go p.processReport(&report)
		case <-ctx.Done():
			return
		default:
			logging.Warnf("concurrency limit reached, skipping report %s with name %s", report.ID, report.Name)
		}
	}
}

func (p *Processor) processReport(report *models.Report) {
	p.WaitGroup.Add(1)
	defer p.threadDone(report)
	logging.Debugf("processing report %s with name %s", report.ID, report.Name)
	p.Collector.Collect(*report)
}

func (p *Processor) threadDone(report *models.Report) {
	p.WaitGroup.Done()
	<-p.ConcurrencyLimitCh
	logging.Debugf("report %s with name %s processed", report.ID, report.Name)
}

func (p *Processor) writeMetricsData(ctx context.Context) {
	statusMap, err := p.fetchTotalReportsByStatus(ctx)
	if err != nil {
		logging.Errorf("Failed to fetch total reports by status. Error: %v", err)
		return
	}

	for status, totalReportsByStatus := range statusMap {
		metrics.GaugeVec(reportsCountPerStatus).Get().WithLabels(map[string]string{
			"status": status,
		}).Set(float64(totalReportsByStatus))
	}
}

func (p *Processor) fetchTotalReportsByStatus(ctx context.Context) (map[string]int, error) {
	statusMap, apiErr := p.ReportsGetter.GetTotalReportsByStatus(ctx)
	if apiErr != nil {
		return nil, fmt.Errorf("failed to get total reports by status: %v", apiErr)
	}

	return statusMap, nil
}
