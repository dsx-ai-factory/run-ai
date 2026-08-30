package database

import (
	"fmt"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"time"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"gorm.io/gorm"
)

const (
	cronJobName               = "workloads_cleaner"
	workloadDeleteQueryFormat = "DELETE FROM workloads WHERE " +
		"(deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '%d' DAY) OR" +
		"(phase IN ('Failed', 'Completed') AND phase_updated_at < NOW() - INTERVAL '%d' DAY) OR" +
		"(project_id = '') OR" +
		"(phase = 'Creating' AND phase_message = 'The workload is being created in the control plane' AND phase_updated_at < NOW() - INTERVAL '%d' MINUTE);"
	eventsDeleteQuery = "DELETE FROM events WHERE " +
		"NOT EXISTS (SELECT 1 FROM workloads WHERE workloads.id = events.involved_object_uid ) AND " +
		"NOT EXISTS (SELECT 1 FROM workloads_children WHERE workloads_children.child_id = events.involved_object_uid);"
	workloadsViewRefreshQuery        = "REFRESH MATERIALIZED VIEW CONCURRENTLY workloads_materialized_view WITH DATA"
	workloadsPartialViewRefreshQuery = "REFRESH MATERIALIZED VIEW CONCURRENTLY workloads_partial_materialized_view WITH DATA"

	// chartUploadsDeleteQueryFormat removes staged chart archives that are no longer needed: either
	// they have outlived the retention window (covers uploads that were never deployed, e.g. an
	// abandoned form) or the application that references them is already deployed (Helm has the
	// release in-cluster and will not re-fetch the chart).
	chartUploadsDeleteQueryFormat = "DELETE FROM application_chart_uploads acu WHERE " +
		"(acu.created_at < NOW() - INTERVAL '%d' MINUTE) OR " +
		"EXISTS (SELECT 1 FROM application_crud_sync cs JOIN applications a ON a.id = cs.application_id " +
		"WHERE a.actual_phase = 'deployed' AND cs.spec->'chartRef'->>'chartId' = acu.id::text);"
)

func InitGoRoutines(stopCh <-chan struct{}) {
	client := dao_utils.DB()

	dbCleanerConf := runai_config.Get[*config.DBCleanerConfig]()
	serverConf := runai_config.Get[*config.ServerConfig]()
	chartUploadConf := runai_config.Get[*config.ChartUploadConfig]()

	cleanupQueries := []string{getWorkloadsDeleteQuery(dbCleanerConf), eventsDeleteQuery}
	// A non-positive retention would reap every staged chart on the next run (0 => older than now,
	// negative => a future cutoff), so treat <=0 as "cleanup disabled", matching InitRefresher below.
	if chartUploadConf.RetentionMinutes > 0 {
		cleanupQueries = append(cleanupQueries, getChartUploadsDeleteQuery(chartUploadConf))
	} else {
		logging.Warnf("chart_retention_minutes is %d; disabling transient chart cleanup", chartUploadConf.RetentionMinutes)
	}

	dao_utils.InitCleaner(stopCh, client, cronJobName, dbCleanerConf.CleanerIntervalHours, cleanupQueries)
	InitRefresher(stopCh, client, serverConf)
}

func InitRefresher(stopCh <-chan struct{}, client *gorm.DB, conf *config.ServerConfig) {
	if conf.ViewRefreshRateMilliseconds <= 0 {
		return
	}

	logging.Infof("setting up a go routine view refresher. Refresh interval %d", conf.ViewRefreshRateMilliseconds)
	ticker := time.NewTicker(time.Duration(conf.ViewRefreshRateMilliseconds) * time.Millisecond)
	refreshQuery := workloadsPartialViewRefreshQuery
	if !conf.UsePartialMaterializedView {
		refreshQuery = workloadsViewRefreshQuery
	}

	go func() {
		runRefresher(client, int64(conf.ViewRefreshRateMilliseconds), refreshQuery)
		for {
			select {
			case <-ticker.C:
				runRefresher(client, int64(conf.ViewRefreshRateMilliseconds), refreshQuery)
				ticker.Reset(time.Duration(conf.ViewRefreshRateMilliseconds) * time.Millisecond)
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func runRefresher(client *gorm.DB, timeInterval int64, refreshQuery string) {
	start := time.Now()
	result := client.Exec(refreshQuery)
	if result.Error != nil {
		logging.Errorf("failed to refresh workloads view, err %v", result.Error)
	} else {
		timeTook := time.Now().UnixMilli() - start.UnixMilli()
		if timeTook > timeInterval {
			logging.Warnf("Run view refresh took %d milliseconds, it is higher than the interval %d", timeTook, timeInterval)
		} else {
			logging.Debugf("Run view refresh took %d milliseconds", timeTook)
		}
	}
}

func getWorkloadsDeleteQuery(conf *config.DBCleanerConfig) string {
	return fmt.Sprintf(workloadDeleteQueryFormat, conf.DeletedWorkloadsDaysToLive, conf.FinishedWorkloadsDaysToLive, conf.CreatingWorkloadsMinutesToLive)
}

func getChartUploadsDeleteQuery(conf *config.ChartUploadConfig) string {
	return fmt.Sprintf(chartUploadsDeleteQueryFormat, conf.RetentionMinutes)
}
