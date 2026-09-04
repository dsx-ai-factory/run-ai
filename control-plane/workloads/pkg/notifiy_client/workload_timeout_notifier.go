/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notifiy_client

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/utils/errors"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var (
	lastFetchedTimeoutTimePerTenant map[string]time.Time

	workloadTypeToTimeoutType = map[string]WorkloadTimeoutType{
		services_crud.Training:  WorkloadTimeoutTypeTraining,
		services_crud.Workspace: WorkloadTimeoutTypeWorkspace,
	}
)

const (
	workloadTimeoutTitle = "Workload %s is about to stop due to timeout"
	timeoutMessage       = "Based on the time limit configured for your project, the workload %s will be stopped at %s."
	idleTimeoutMessage   = "Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by %s remain idle, the workload will be terminated at %s."

	workloadTimeoutApproachingEventReason = "WorkloadTimeoutApproaching"
)

func InitWorkloadTimeoutNotifier(ctx context.Context, stopCh <-chan struct{}, conf *config.Notify, notifyClient NotifyClient) {
	if conf.TimeoutIntervalMillis <= 0 {
		logging.Infof("NotifyTimeoutIntervalMillis is set to 0, skipping workload timeout notifier initialization")
		return
	}

	logging.Infof("Setting up a go routine for workload timeout notifier. Notify interval: %d milliseconds", conf.TimeoutIntervalMillis)
	interval := time.Duration(conf.TimeoutIntervalMillis) * time.Millisecond
	ticker := time.NewTicker(interval)

	lastFetchedTimeoutTimePerTenant = map[string]time.Time{}

	// sleep for the interval time, to let the correct data to be filled after startup
	time.Sleep(interval)

	notifierLock := &sync.Mutex{}
	go func() {
		go runWorkloadTimeoutNotifier(ctx, conf, notifierLock, notifyClient)
		for {
			select {
			case <-ticker.C:
				go runWorkloadTimeoutNotifier(ctx, conf, notifierLock, notifyClient)
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func runWorkloadTimeoutNotifier(ctx context.Context, conf *config.Notify, notifierLock *sync.Mutex,
	notifyClient NotifyClient,
) {
	if !notifierLock.TryLock() {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Workload timeout notifier is still in progress - not starting new sync")
		return
	}
	defer notifierLock.Unlock()

	tenantsToNotify := notifyClient.GetEnabledTenants()
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Starting workload timeout notifier for %d tenants - %+v", len(tenantsToNotify), tenantsToNotify)

	for _, tenantId := range tenantsToNotify {
		lastFetchTime, ok := lastFetchedTimeoutTimePerTenant[tenantId]
		if !ok {
			lastFetchTime = time.Now().UTC().Add(-time.Duration(conf.TimeoutIntervalMillis) * time.Millisecond)
			lastFetchedTimeoutTimePerTenant[tenantId] = lastFetchTime
		}

		timeoutApproachingEvents, dbErr := dao.DbGetTimeoutEventsWithReason(ctx, tenantId, workloadTimeoutApproachingEventReason, lastFetchTime)
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get events with reason %s for tenant %s, error: %v",
				workloadTimeoutApproachingEventReason, tenantId, dbErr)
			continue
		}

		tenantBaseUrl, err := notifyClient.GetTenantBaseUrl(tenantId)
		if err != nil {
			if !errors.IsNotFound(err) {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant base url for tenant %s, error: %v", tenantId, err)
			}
			continue
		}

		for _, event := range *timeoutApproachingEvents {
			notifyWorkloadTimeout(ctx, notifyClient, event, tenantBaseUrl)
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Notified workload timeout for %d workloads in tenant %s", len(*timeoutApproachingEvents), tenantId)

		// update last fetch time per tenant with db timestamp
		numOfEventsPerTenant := len(*timeoutApproachingEvents)
		if numOfEventsPerTenant > 0 {
			lastFetchedTimeoutTimePerTenant[tenantId] = (*timeoutApproachingEvents)[numOfEventsPerTenant-1].CreatedAt
		}
	}
}

func notifyWorkloadTimeout(ctx context.Context, notifyClient NotifyClient, workloadEvent models.TimeoutEventsOfWorkloadView, tenantBaseUrl string) {
	additionalData := NotificationAdditionalData{
		WorkloadName: workloadEvent.InvolvedObjectName,
		WorkloadType: workloadEvent.WorkloadType,
		ProjectName:  workloadEvent.ProjectName,
		ClusterId:    workloadEvent.ClusterID.String(),
		TenantId:     workloadEvent.TenantID,
		User:         workloadEvent.SubmittedBy,
		Time:         workloadEvent.CreatedAt,
		URL:          getWorkloadInfoUrl(tenantBaseUrl, workloadEvent.ProjectName, workloadEvent.InvolvedObjectName, workloadEvent.WorkloadType),
	}

	workloadTimeoutType, ok := workloadTypeToTimeoutType[workloadEvent.WorkloadType]
	if !ok {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to find timeout type for workload type %s for workload <%s> (name: <%s>) in project <%s>",
			workloadEvent.WorkloadType, workloadEvent.WorkloadID, workloadEvent.InvolvedObjectName, workloadEvent.ProjectName)
		return
	}

	title, messageBody := getTimeoutNotificationTitleAndMessageBody(ctx, additionalData, workloadEvent)

	builder := notifyClient.BuildNotification(workloadTimeoutCategory, string(workloadTimeoutType))
	notifyWorkloadEvent(ctx, builder, title, messageBody, additionalData)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Notified workload timeout type <%v> for workload name <%s> in project <%s> in tenant %d",
		workloadTimeoutType, workloadEvent.InvolvedObjectName, workloadEvent.ProjectName, workloadEvent.TenantID)
}

func getTimeoutNotificationTitleAndMessageBody(ctx context.Context, additionalData NotificationAdditionalData,
	workloadEvent models.TimeoutEventsOfWorkloadView,
) (string, string) {
	title := fmt.Sprintf(workloadTimeoutTitle, additionalData.WorkloadName)

	timeoutBaseMessage := timeoutMessage
	if strings.Contains(strings.ToLower(workloadEvent.Message), "idle") {
		timeoutBaseMessage = idleTimeoutMessage
	}
	messageBody := ""
	timeoutTime, err := calculateTimeoutTime(ctx, workloadEvent.CreatedAt, workloadEvent.Message)
	if err == nil {
		messageBody = fmt.Sprintf(timeoutBaseMessage, additionalData.WorkloadName, getTimeStr(*timeoutTime))
	} else {
		messageBody = workloadEvent.Message
	}

	return title, messageBody
}

func calculateTimeoutTime(ctx context.Context, eventCreatedAt time.Time, eventMessage string) (*time.Time, error) {
	splitter := "will be terminated in "
	messageParts := strings.Split(eventMessage, splitter)
	if len(messageParts) < 2 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to calculate timeout time for event message <%s>", eventMessage)
		return nil, fmt.Errorf("failed to calculate timeout time for event message <%s>", eventMessage)
	}

	timeStr := messageParts[1]
	days, hours, minutes := parseTimeStr(ctx, timeStr)

	timeoutTime := eventCreatedAt.Add(time.Duration(days) * 24 * time.Hour).
		Add(time.Duration(hours) * time.Hour).
		Add(time.Duration(minutes) * time.Minute)

	return &timeoutTime, nil
}

func parseTimeStr(ctx context.Context, timeStr string) (int, int, int) {
	// acceptable formats of timeStr:
	// %d days - %d hours - %d minutes
	// %d hours - %d minutes
	// %d minutes

	daysRegexp := regexp.MustCompile(`(\d+) days`)
	hoursRegexp := regexp.MustCompile(`(\d+) hours`)
	minutesRegexp := regexp.MustCompile(`(\d+) minutes`)

	days := parseRegexpInt(ctx, daysRegexp, timeStr, "days")
	hours := parseRegexpInt(ctx, hoursRegexp, timeStr, "hours")
	minutes := parseRegexpInt(ctx, minutesRegexp, timeStr, "minutes")

	return days, hours, minutes
}

func parseRegexpInt(ctx context.Context, regexp *regexp.Regexp, timeStr string, element string) int {
	match := regexp.FindStringSubmatch(timeStr)
	if len(match) < 2 {
		return 0
	}

	value, err := strconv.Atoi(match[1])
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to parse int %s from time string: <%s>",
			element, timeStr)
		return 0
	}
	return value
}
