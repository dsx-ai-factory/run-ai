/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package db

import (
	"database/sql"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/runaimetrics"
)

type stats struct {
	maxOpenConnections runaimetrics.MetricGauge // Maximum number of open connections to the database.

	// Pool status
	openConnections runaimetrics.MetricGauge // The number of established connections both in use and idle.
	inUse           runaimetrics.MetricGauge // The number of connections currently in use.
	idle            runaimetrics.MetricGauge // The number of idle connections.

	// Counters
	waitCount         runaimetrics.MetricGauge // The total number of connections waited for.
	waitDuration      runaimetrics.MetricGauge // The total time blocked waiting for a new connection.
	maxIdleClosed     runaimetrics.MetricGauge // The total number of connections closed due to SetMaxIdleConns.
	maxLifetimeClosed runaimetrics.MetricGauge // The total number of connections closed due to SetConnMaxLifetime.
	maxIdleTimeClosed runaimetrics.MetricGauge // The total number of connections closed due to SetConnMaxIdleTime.
}

func newStats(labels map[string]string) *stats {
	return &stats{
		maxOpenConnections: metrics.Gauge("db_max_open_connections").Register("Maximum number of open connections to the database", labels),
		openConnections:    metrics.Gauge("db_open_connections").Register("The number of established connections both in use and idle", labels),
		inUse:              metrics.Gauge("db_in_use_connections").Register("The number of connections currently in use", labels),
		idle:               metrics.Gauge("db_idle_connections").Register("The number of idle connections", labels),
		waitCount:          metrics.Gauge("db_waited_connections").Register("The total number of connections waited for", labels),
		waitDuration:       metrics.Gauge("db_waited_duration").Register("The total time blocked waiting for a new connection", labels),
		maxIdleClosed:      metrics.Gauge("db_max_idle_closed_connections").Register("The total number of connections closed due to SetMaxIdleConns", labels),
		maxLifetimeClosed:  metrics.Gauge("db_max_lifetime_closed_connections").Register("The total number of connections closed due to SetConnMaxLifetime", labels),
		maxIdleTimeClosed:  metrics.Gauge("db_max_idle_time_closed_connections").Register("The total number of connections closed due to SetConnMaxIdleTime", labels),
	}
}

func (stats *stats) set(dbStats sql.DBStats) {
	stats.maxOpenConnections.Set(float64(dbStats.MaxOpenConnections))
	stats.openConnections.Set(float64(dbStats.OpenConnections))
	stats.inUse.Set(float64(dbStats.InUse))
	stats.idle.Set(float64(dbStats.Idle))
	stats.waitCount.Set(float64(dbStats.WaitCount))
	stats.waitDuration.Set(float64(dbStats.WaitDuration))
	stats.maxIdleClosed.Set(float64(dbStats.MaxIdleClosed))
	stats.maxLifetimeClosed.Set(float64(dbStats.MaxLifetimeClosed))
	stats.maxIdleTimeClosed.Set(float64(dbStats.MaxIdleTimeClosed))
}
