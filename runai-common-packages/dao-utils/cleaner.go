package dao_utils

import (
	"time"

	"gorm.io/gorm"
)

func InitCleaner(stopCh <-chan struct{}, client *gorm.DB, cronJobName string, cleanerIntervalHours int, queries []string) {
	cleanerIntervalSeconds := cleanerIntervalHours * 60 * 60
	job := NewCronJobWithStartTime(cronJobName, cleanerIntervalSeconds, queries, getCleanerStartTime())
	job.Schedule(stopCh, client)
}

func getCleanerStartTime() time.Time {
	now := time.Now()
	threeAM := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	return threeAM
}
