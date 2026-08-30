package schedule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

type Job struct {
	db        orm.DB
	name      string
	interval  time.Duration
	queries   []string
	startTime time.Time
}

func NewJob(db orm.DB, name string, interval time.Duration, queries []string) *Job {
	return &Job{
		db:       db,
		name:     name,
		interval: interval,
		queries:  queries,
	}
}

func (j *Job) WithStartTime(t time.Time) *Job {
	j.startTime = t
	return j
}

func (j *Job) Schedule(ctx context.Context) {
	if j.interval <= 0 {
		logger.GetLogger().Errorf("schedule: job %q has non-positive interval %s, not scheduling", j.name, j.interval)
		return
	}
	if !isPGCronInstalled(j.db) {
		logger.GetLogger().Infof("pg_cron not installed, scheduling job %q as goroutine", j.name)
		j.scheduleGoroutine(ctx)
		return
	}
	if err := j.schedulePGCron(); err != nil {
		logger.GetLogger().Errorf("Failed to schedule pg_cron job %q, falling back to goroutine: %v", j.name, err)
		j.scheduleGoroutine(ctx)
	}
}

func (j *Job) schedulePGCron() error {
	_, _ = j.db.Exec("SELECT cron.unschedule(?)", j.name)
	_, err := j.db.Exec(
		"SELECT cron.schedule(?, ?, ?)",
		j.name, toCronExpression(j.interval), strings.Join(j.queries, "; "),
	)
	return err
}

func (j *Job) scheduleGoroutine(ctx context.Context) {
	initialDelay := getInitialDelay(j.startTime)
	needToReset := initialDelay > 0
	if !needToReset {
		initialDelay = j.interval
	}

	ticker := time.NewTicker(initialDelay)

	go func() {
		runQueries(j.db, j.queries)
		for {
			select {
			case <-ticker.C:
				runQueries(j.db, j.queries)
				if needToReset {
					ticker.Reset(j.interval)
					needToReset = false
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func isPGCronInstalled(db orm.DB) bool {
	var count int64
	if err := db.Table("pg_extension").Where("extname = ?", "pg_cron").Count(&count); err != nil {
		logger.GetLogger().Errorf("Failed to query pg_cron availability: %v", err)
		return false
	}
	return count > 0
}

func toCronExpression(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("*/%d * * * *", minutes)
	}
	hours := minutes / 60
	return fmt.Sprintf("0 */%d * * *", hours)
}

func getInitialDelay(startTime time.Time) time.Duration {
	if startTime.IsZero() {
		return 0
	}
	now := time.Now()
	st := time.Date(now.Year(), now.Month(), now.Day(),
		startTime.Hour(), startTime.Minute(), startTime.Second(), startTime.Nanosecond(), now.Location())
	if now.After(st) {
		st = st.AddDate(0, 0, 1)
	}
	return st.Sub(now)
}

func runQueries(db orm.DB, queries []string) {
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			logger.GetLogger().Errorf("Failed to exec job query %q: %v", q, err)
		}
	}
}
