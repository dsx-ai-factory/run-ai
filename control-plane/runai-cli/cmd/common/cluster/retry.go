package cluster

import (
	"time"
)

const (
	SleepTimeDivider = 10
)

func GetIterationMinimalTime(defaultSleep, waitTimeout time.Duration) time.Duration {
	newSleepTime := waitTimeout / SleepTimeDivider
	if newSleepTime < defaultSleep {
		return newSleepTime
	}
	return defaultSleep
}
