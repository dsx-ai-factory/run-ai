package database

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/schedule"
)

const (
	reportsJobName = "metrics_reports"
)

var markFailedReportsAsPendingQueryFormat = fmt.Sprintf(`
	UPDATE %s 
	SET retry_count = retry_count + 1, status = '%s'
	WHERE status = '%s'
	AND retry_count < %%d
	AND created_at BETWEEN NOW() - INTERVAL '%%d minutes' AND NOW() - INTERVAL '%%d minutes';
`, models.ReportsTable, api.REPORTSTATUS_PENDING, api.REPORTSTATUS_FAILED)

var markProcessingReportsAsFailedQueryFormat = fmt.Sprintf(`
	UPDATE %s 
	SET status = '%s', status_message = 'timeout error'
	WHERE status = '%s'
	AND created_at < NOW() - INTERVAL '%%d minutes'
`, models.ReportsTable, api.REPORTSTATUS_FAILED, api.REPORTSTATUS_PROCESSING)

var markStuckDeletingReportsAsFailedQueryFormat = fmt.Sprintf(`
	UPDATE %s
	SET status = '%s', status_message = 'delete timeout'
	WHERE status in ('%s', '%s') AND status_message IS NOT NULL
	AND status_updated_at < NOW() - INTERVAL '%%d minutes'
`, models.ReportsTable, api.REPORTSTATUS_DELETION_FAILED, api.REPORTSTATUS_PENDING_DELETE, api.REPORTSTATUS_DELETING)

type CronJobsConfiguration struct {
	Reports *ReportsConfiguration `mapstructure:"reports"`
}

type ReportsConfiguration struct {
	IntervalSeconds                         int `mapstructure:"interval_seconds"`
	MaxRetryCount                           int `mapstructure:"max_retry_count"`
	RetryFailedReportsMinAgeMinutes         int `mapstructure:"retry_failed_reports_min_age_minutes"`
	RetryFailedReportsMaxAgeMinutes         int `mapstructure:"retry_failed_reports_max_age_minutes"`
	FailStuckProcessingReportsMinAgeMinutes int `mapstructure:"fail_stuck_processing_reports_min_age_minutes"`
	FailStuckDeletingReportsMinAgeMinutes   int `mapstructure:"fail_stuck_deleting_reports_min_age_minutes"`
}

func (c *CronJobsConfiguration) SetDefaults() {
	c.Reports = &ReportsConfiguration{
		IntervalSeconds:                         60,
		MaxRetryCount:                           1,
		RetryFailedReportsMinAgeMinutes:         1,
		RetryFailedReportsMaxAgeMinutes:         30,
		FailStuckProcessingReportsMinAgeMinutes: 30,
		FailStuckDeletingReportsMinAgeMinutes:   30,
	}
}

func InitReportsCronJob(ctx context.Context) {
	conf := config.Get[*CronJobsConfiguration]()

	logging.Infof("Initializing reports cron job with interval: %d seconds\n", conf.Reports.IntervalSeconds)
	interval := time.Duration(conf.Reports.IntervalSeconds) * time.Second
	schedule.NewJob(runaidb.DB(), reportsJobName, interval, []string{
		getRetryFailedReports(),
		getFailStuckProcessingReports(),
		getFailStuckDeletingReports(),
	}).
		Schedule(ctx)
}

func getRetryFailedReports() string {
	conf := config.Get[*CronJobsConfiguration]()
	return fmt.Sprintf(markFailedReportsAsPendingQueryFormat, conf.Reports.MaxRetryCount, conf.Reports.RetryFailedReportsMaxAgeMinutes, conf.Reports.RetryFailedReportsMinAgeMinutes)
}

func getFailStuckProcessingReports() string {
	conf := config.Get[*CronJobsConfiguration]()
	return fmt.Sprintf(markProcessingReportsAsFailedQueryFormat, conf.Reports.FailStuckProcessingReportsMinAgeMinutes)
}

func getFailStuckDeletingReports() string {
	conf := config.Get[*CronJobsConfiguration]()
	return fmt.Sprintf(markStuckDeletingReportsAsFailedQueryFormat, conf.Reports.FailStuckDeletingReportsMinAgeMinutes)
}
