/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package queuesManager

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TimeoutEventMessage      = "Based on the time limit configured for your project, the job %s will be terminated in %s."
	TimeoutStoppedMessage    = "The workload has reached the running time limitation on the project (%s). Stopping the workload."
	TimeoutErrorEventMessage = "Workload %s was not terminated based on the time limit configured for your project: %s"

	IdleEventMessage      = "Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by %s remain idle, the job will be terminated in %s."
	IdleStoppingMessage   = "The workload has reached the idle time limitation on the project (%s). Stopping the workload."
	IdleErrorEventMessage = "Workload %s was not terminated based on the idle GPU timeout configured for your project: %s"
)

func createNotificationEvent(recorder record.EventRecorder, workload client.Object, jobName string, eventClock EventClock) {
	message := getMessage(jobName, eventClock)
	recorder.Event(workload, coreV1.EventTypeNormal, "WorkloadTimeoutApproaching", message)
}

func createStoppedEvent(recorder record.EventRecorder, workload client.Object, task *WorkloadTask, eventClock EventClock) {
	message := getTerminateMessage(task.StartTime, eventClock)
	recorder.Event(workload, coreV1.EventTypeNormal, "WorkloadTimeoutReached", message)
}

func createErrorNotificationEvent(recorder record.EventRecorder, workload client.Object, jobName string, eventClock EventClock, err error) {
	message := getErrorMessage(jobName, eventClock, err)
	recorder.Event(workload, coreV1.EventTypeWarning, "WorkloadTimeoutError", message)
}

func getMessage(jobName string, eventClock EventClock) string {
	timerMsg := createTimeMessage(time.Until(eventClock.DeactivateEvent))
	if eventClock.NextEventType == TaskTimerIdle {
		return fmt.Sprintf(IdleEventMessage, jobName, timerMsg)
	}
	return fmt.Sprintf(TimeoutEventMessage, jobName, timerMsg)
}

func getTerminateMessage(startTime time.Time, eventClock EventClock) string {
	timerMsg := createTimeMessage(time.Since(startTime))
	if eventClock.NextEventType == TaskTimerIdle {
		return fmt.Sprintf(IdleStoppingMessage, timerMsg)
	}
	return fmt.Sprintf(TimeoutStoppedMessage, timerMsg)
}

func getErrorMessage(jobName string, eventClock EventClock, err error) string {
	if eventClock.NextEventType == TaskTimerIdle {
		return fmt.Sprintf(IdleErrorEventMessage, jobName, err)
	}
	return fmt.Sprintf(TimeoutErrorEventMessage, jobName, err)
}

func createTimeMessage(diff time.Duration) string {
	// NOTE: IMPORTANT: do not change this syntax !!!
	// there is a dependency on this syntax in the workload-service -
	// parsing the exact timeout time.
	// runai/control-plane/workloads/pkg/notifiy_client/workload_timeout_notifier.go
	// function "parseTimeStr".
	// 34 days - 23 hours - 20 minutes

	// Extract days, hours, and minutes from the time difference
	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d days - %d hours - %d minutes", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours - %d minutes", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%d minutes", minutes)
	}

	// should not reach here, but if it will we have the event
	return fmt.Sprintf("%d days - %d hours - %d minutes", days, hours, minutes)
}

func resetIdleNotificationIfNeeded(job *WorkloadTask, termination time.Time) {
	takeAction := &job.TaskAction
	takeAction.SetEventType(TaskTimerIdle)
	if len(takeAction.GetList()) == 2 || takeAction.GetIndex() == 1 {
		return
	}

	firstTimeEvent := GetTimePercentile(job.StartTime, termination, takeAction.GetList()[1])
	if time.Now().Before(firstTimeEvent) {
		zap.S().Infof("Workload (%s) idle timer task was reset to percentile: %d", job.WorkloadName, takeAction.GetList()[1])
		takeAction.SetIndex(1)
	}
}
