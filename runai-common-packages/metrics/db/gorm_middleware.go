package db

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

const metricsRefreshIntervalSeconds = 30

type gormPrometheusMetricsMiddleware struct {
	errorCallback func(e error)
	db            *gorm.DB
	dbStats       *stats
	refreshOnce   sync.Once
}

func newGormPrometheusMetricsMiddleware(errorCallback func(e error)) *gormPrometheusMetricsMiddleware {
	return &gormPrometheusMetricsMiddleware{
		errorCallback: errorCallback,
	}
}
func (p *gormPrometheusMetricsMiddleware) Name() string {
	return "RunaiGormPrometheus"
}

func (p *gormPrometheusMetricsMiddleware) Initialize(db *gorm.DB) error { // can be called repeatedly
	if db == nil {
		p.errorCallback(fmt.Errorf("RunaiGormPrometheus failed to initialize, got nil db"))
	}

	p.db = db
	labels := map[string]string{}
	p.dbStats = newStats(labels)

	p.refreshOnce.Do(func() {
		go func() {
			for range time.Tick(time.Duration(metricsRefreshIntervalSeconds) * time.Second) {
				p.refresh()
			}
		}()
	})

	return nil
}

func (p *gormPrometheusMetricsMiddleware) refresh() {
	if db, err := p.db.DB(); err == nil {
		p.dbStats.set(db.Stats())
	} else {
		p.errorCallback(fmt.Errorf("RunaiGormPrometheus failed to collect db status, got error: %v", err))
	}
}
