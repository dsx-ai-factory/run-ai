package dbcleaner

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

const (
	reasonableTime = 1 * time.Hour
)

func InitCleanupJob(ctx context.Context, stopCh <-chan struct{}) {
	serverConfig := *runai_config.Get[*config.ServerConfig]()
	if serverConfig.CleanupRefreshRateHours <= 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("cleanup job is disabled (refreshRateHours=%d)", serverConfig.CleanupRefreshRateHours)
		return
	}

	logging.Infof("setting up cleanup job to run every %d hours", serverConfig.CleanupRefreshRateHours)
	ticker := time.NewTicker(time.Duration(serverConfig.CleanupRefreshRateHours) * time.Hour)

	go func() {
		runCleanup(ctx, serverConfig)

		for {
			select {
			case <-ticker.C:
				runCleanup(ctx, serverConfig)
				ticker.Reset(time.Duration(serverConfig.CleanupRefreshRateHours) * time.Hour)
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func runCleanup(ctx context.Context, serverConfig config.ServerConfig) {
	start := time.Now()

	db := runaidb.DB().WithContext(ctx)

	deleted, errors := DoCleanup(ctx, db, serverConfig)

	timeTook := time.Since(start)
	if timeTook > reasonableTime {
		logging.Warnf("Run cleanup took %s which is more than reasonable time %s. deleted %d errors %d",
			timeTook.String(), reasonableTime.String(), deleted, len(errors))
	} else {
		logging.Infof("Run cleanup took %s. deleted %d errors %d", timeTook.String(), deleted, len(errors))
	}
}

func DoCleanup(_ context.Context, db orm.DB, serverConfig config.ServerConfig) (int64, []error) {
	errors := make([]error, 0)

	obsoleteIds := make([]string, 0)

	scanRows, scanErr := db.Model(&dbmodel.AssetMetaRec{}).Unscoped().Select(dbtables.AssetMetaColumnId).Where(
		fmt.Sprintf("%s IS NOT NULL AND %s < NOW() - INTERVAL '%d' DAY",
			dbtables.AssetMetaColumnDeletedAt,
			dbtables.AssetMetaColumnDeletedAt,
			serverConfig.CleanupKeepAssetsDays)).
		Scan(&obsoleteIds)
	if scanErr != nil {
		logging.Warnf("assets cleanup failed to locate candidates for delete: %v", scanErr)
		errors = append(errors, scanErr)
		return 0, errors
	}

	if scanRows == 0 {
		return 0, errors
	}

	var deleted int64

	for tableName, assetIdKey := range dbtables.TableToClean {
		delRows, delErr := db.Table(tableName).Unscoped().Where(fmt.Sprintf("%s in ?", assetIdKey), obsoleteIds).Delete(nil)
		if delErr != nil {
			logging.Errorf("assets cleanup failed to remove %d assets from table %s: %v", len(obsoleteIds), tableName, delErr)
			errors = append(errors, delErr)
		} else {
			deleted += delRows
		}
	}

	return deleted, errors
}
