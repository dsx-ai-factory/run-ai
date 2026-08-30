package database

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/schedule"
)

const (
	cleanerJobName           = "metrics_cleaner"
	deleteReportsQueryFormat = "DELETE FROM reports WHERE created_at < NOW() - INTERVAL '%d' DAY;"
)

func InitCleanerIfEnabled(ctx context.Context) {
	conf := config.Get[*CleanerConfiguration]()
	if !conf.IsCleanupEnabled {
		return
	}

	logging.Infof("Initializing DB cleaner with interval: %d hours, TTL: %d days\n", conf.CleanerIntervalHours, conf.DeleteReportsTTLDays)
	interval := time.Duration(conf.CleanerIntervalHours) * time.Hour
	schedule.NewJob(runaidb.DB(), cleanerJobName, interval, []string{getDeleteReportsQuery(conf)}).
		WithStartTime(getCleanerStartTime()).
		Schedule(ctx)
}

func getDeleteReportsQuery(conf *CleanerConfiguration) string {
	return fmt.Sprintf(deleteReportsQueryFormat, conf.DeleteReportsTTLDays)
}

type CleanerConfiguration struct {
	IsCleanupEnabled     bool `mapstructure:"is_cleanup_enabled"`
	CleanerIntervalHours int  `mapstructure:"cleaner_interval_hours" validate:"min=0"`
	DeleteReportsTTLDays int  `mapstructure:"delete_report_ttl_days" validate:"min=0"`
}

func (c *CleanerConfiguration) SetDefaults() {
}

func getCleanerStartTime() time.Time {
	now := time.Now()
	threeAM := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	return threeAM
}
