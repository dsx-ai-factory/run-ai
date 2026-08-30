package dao_utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"gorm.io/gorm"
)

type CronJob struct {
	name            string
	intervalSeconds int
	queries         []string
	startTime       time.Time
}

func NewCronJob(name string, intervalSeconds int, queries []string) *CronJob {
	return &CronJob{
		name:            name,
		intervalSeconds: intervalSeconds,
		queries:         queries,
	}
}

func NewCronJobWithStartTime(name string, intervalSeconds int, queries []string, startTime time.Time) *CronJob {
	return &CronJob{
		name:            name,
		intervalSeconds: intervalSeconds,
		queries:         queries,
		startTime:       startTime,
	}
}

func (j *CronJob) Schedule(stopCh <-chan struct{}, client *gorm.DB) {
	if !isPGCronInstalled(client) {
		logging.Infof("PG cron is not installed, using Go routine for job %s", j.name)
		j.runAsGoRoutine(stopCh, client)
		return
	}
	if err := j.runAsPGCron(client); err != nil {
		logging.Errorf("Failed to set up PG cron job for job %s, falling back to Go routine: %v", j.name, err)
		j.runAsGoRoutine(stopCh, client)
	}
}

func (j *CronJob) runAsPGCron(client *gorm.DB) error {
	result := client.Exec("DO $$ " +
		"BEGIN " +
		//Attempt to drop the job if it already exists
		"BEGIN " +
		fmt.Sprintf("PERFORM cron.schedule('%s'); ", j.name) +
		fmt.Sprintf("IF FOUND THEN PERFORM cron.unschedule('%s'); ", j.name) +
		"END IF; " +
		"EXCEPTION " +
		// Do nothing if the job doesn't exist
		"WHEN UNDEFINED_OBJECT THEN NULL;" +
		"END; " +
		//-- Create the job with the new settings
		fmt.Sprintf("PERFORM cron.schedule('%s', '%s', ", j.name, getCronJobSchedule(j.intervalSeconds)) +
		strings.Join(j.queries, " ") +
		"); END $$;")

	return result.Error
}

func (j *CronJob) runAsGoRoutine(stopCh <-chan struct{}, client *gorm.DB) {
	needToReset := true
	initialDelay := getInitialDelay(j.startTime)
	if initialDelay == 0 {
		initialDelay = time.Duration(j.intervalSeconds) * time.Second
		needToReset = false
	}

	ticker := time.NewTicker(initialDelay)

	go func() {
		runQueries(client, j.queries)
		for {
			select {
			case <-ticker.C:
				runQueries(client, j.queries)
				if needToReset {
					ticker.Reset(time.Duration(j.intervalSeconds) * time.Second)
					needToReset = false
				}
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func getCronJobSchedule(intervalSeconds int) string {
	if intervalSeconds < 60 {
		return fmt.Sprintf("%d seconds", intervalSeconds)
	}

	minutes := intervalSeconds / 60
	if minutes < 60 {
		return fmt.Sprintf("*/%d * * * *", minutes)
	}

	hours := minutes / 60
	return fmt.Sprintf("0 */%d * * *", hours)
}

func isPGCronInstalled(client *gorm.DB) bool {
	var count int64 = 0

	result := client.Table("pg_extension").Where("extname = 'pg_cron'").Count(&count)
	if result.Error != nil {
		logging.Errorf("Failed to query pg_cron extension availability. %+v", result.Error)
		return false
	}

	return count > 0
}

func getInitialDelay(startTime time.Time) time.Duration {
	if startTime.IsZero() {
		return 0
	}

	now := time.Now()

	// Ensure startTime is set for today
	startTime = time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), startTime.Nanosecond(), now.Location())
	if now.After(startTime) {
		startTime = startTime.AddDate(0, 0, 1) // If already past startTime, set it for tomorrow
	}

	return startTime.Sub(now)
}

func runQueries(client *gorm.DB, queries []string) {
	var result *gorm.DB
	for _, query := range queries {
		if result = client.Exec(query); result.Error != nil {
			logging.Errorf("Failed to exec cron job query %s. %+v", query, result.Error)
		}
	}
}
