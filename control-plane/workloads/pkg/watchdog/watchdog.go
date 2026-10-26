/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package watchdog

import (
	"fmt"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

const (
	failStuckWorkloadsQueryFormat = "UPDATE workloads SET phase = '%s',  phase_message = '%s', support_suspend = false WHERE phase = '%s' AND updated_at < NOW() - INTERVAL '%s' AND phase_message = '%s';"
)

// InitGoRoutines initializes all go routines for watchdog
func InitGoRoutines(stopCh <-chan struct{}) {
	conf := runai_config.Get[*config.WatchDog]()
	client := runaidb.DB()
	initWatchDog(stopCh, client, conf)
}

func initWatchDog(stopCh <-chan struct{}, client orm.DB, conf *config.WatchDog) {
	if conf.WakeupIntervalMin <= 0 {
		return
	}

	logging.Infof("setting up a watchdog go routine. Refresh interval %d min", conf.WakeupIntervalMin)
	ticker := time.NewTicker(time.Duration(conf.WakeupIntervalMin) * time.Minute)
	var WakeupIntervalMilliseconds = int64(conf.WakeupIntervalMin) * 60 * 1000

	go func() {
		runWatchDog(client, conf.MaxCreationDuration, WakeupIntervalMilliseconds)
		for {
			select {
			case <-ticker.C:
				runWatchDog(client, conf.MaxCreationDuration, WakeupIntervalMilliseconds)
				ticker.Reset(time.Duration(conf.WakeupIntervalMin) * time.Minute)
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func runWatchDog(client orm.DB, stuckDuration string, timeInterval int64) {
	start := time.Now()
	rowsAffected, err := client.Exec(buildFailStuckWorkloadsQuery(stuckDuration))
	if err != nil {
		logging.Errorf("failed to handle stuck workloads, err %v", err)
		return
	}
	timeTook := time.Now().UnixMilli() - start.UnixMilli()
	var logFn func(format string, args ...interface{})

	if timeTook > timeInterval {
		logFn = logging.Warnf
		logFn("Running watchdog took %d milliseconds, it is higher that the interval %d", timeTook, timeInterval)
	} else {
		logFn = logging.Debugf
		logFn("Running watchdog took took %d milliseconds", timeTook)
	}
	logFn("Rows affected %v", rowsAffected)
}

func buildFailStuckWorkloadsQuery(stuckDuration string) string {
	return fmt.Sprintf(failStuckWorkloadsQueryFormat, openapi.PHASE_FAILED, services_utils.WorkloadFailedPhaseMsg_StuckInCreation, openapi.PHASE_CREATING, stuckDuration, services.ControlPlaneCreatingMessage)
}
