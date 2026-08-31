/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package queuesManager

import (
	"go.uber.org/zap"
	"time"
)

const (
	UnknownEvent    = 0
	DeactivateEvent = 100
)

// EventClock holds the event required data:
// DeactivateEvent: use for calculation next event and event message
// NextEvent: points on the next event time.
// NextEventType: which type of next event EventClock (idle/timeout)
type EventClock struct {
	DeactivateEvent time.Time
	NextEvent       time.Time
	NextEventType   EventType
}

func (e *EventClock) SetNextEventTime(workloadTask *WorkloadTask) {
	taskAction := &workloadTask.TaskAction // just for shortening the code lines.

	if taskAction.GetIndex() == UnknownEvent {
		timeNow := time.Now()
		nextNoteTime := e.DeactivateEvent
		actionList := taskAction.GetList()

		for action := 1; action < len(actionList); action++ {
			nextNoteTime = GetTimePercentile(workloadTask.StartTime, e.DeactivateEvent, actionList[action])
			if nextNoteTime.After(timeNow) {
				e.NextEvent = nextNoteTime
				taskAction.SetIndex(action)
				zap.S().Debugf("Notification timer %v was not validated, next event for workloadTask <%s> is <%d%%> notificaiton.", e.NextEventType, workloadTask.JobName, actionList[action])
				return
			}
		}
		e.NextEvent = nextNoteTime
		taskAction.SetIndex(len(actionList) - 1)
		zap.S().Debugf("Notification timer %v was not validated, next event for workloadTask <%s> is deactivation.", e.NextEventType, workloadTask.JobName)
		return
	}
	e.NextEvent = GetTimePercentile(workloadTask.StartTime, e.DeactivateEvent, taskAction.GetListIndex())
}

// GetTimePercentile calculate the time of the next event
// the percentage of the total runtime
func GetTimePercentile(startTime time.Time, earlierTime time.Time, percentage int) time.Time {
	zap.S().Debugf("earlierTime: %s, startTime: %s, percentage: %d", earlierTime.String(), startTime.String(), percentage)
	if percentage == DeactivateEvent {
		return earlierTime
	}

	delta := earlierTime.Sub(startTime)
	delta = delta * time.Duration(percentage) / 100

	earlierTime = startTime.Add(delta)
	zap.S().Debugf("newEarlierTime: %s, startTime: %s, percentage: %d", earlierTime.String(), startTime.String(), percentage)
	return earlierTime
}

func GetTheEarlier(timeout EventClock, idle EventClock) EventClock {
	if timeout.NextEvent.IsZero() && idle.NextEvent.IsZero() {
		return EventClock{}
	}

	if !timeout.NextEvent.IsZero() && !idle.NextEvent.IsZero() {
		if timeout.NextEvent.Before(idle.NextEvent) {
			return timeout
		} else {
			return idle
		}
	}

	if !timeout.NextEvent.IsZero() {
		return timeout
	} else {
		return idle
	}
}
