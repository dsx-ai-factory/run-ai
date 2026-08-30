package notifiy_client

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/control-plane/workloads/pkg/utils/errors"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var (
	lastFetchedPhaseTimePerTenant map[string]time.Time

	phaseToStatusChangeNotificationType = map[api.Phase]WorkloadStatusChangeType{
		api.PHASE_CREATING:     WorkloadStatusChangeCreating,
		api.PHASE_PENDING:      WorkloadStatusChangePending,
		api.PHASE_INITIALIZING: WorkloadStatusChangeInitializing,
		api.PHASE_RUNNING:      WorkloadStatusChangeRunning,
		api.PHASE_DEGRADED:     WorkloadStatusChangeDegraded,
		api.PHASE_COMPLETED:    WorkloadStatusChangeCompleted,
		api.PHASE_FAILED:       WorkloadStatusChangeFailed,
		api.PHASE_STOPPED:      WorkloadStatusChangeStopped,
		api.PHASE_DELETING:     WorkloadStatusChangeDeleting,
	}
)

const (
	workloadStatusChangeTitle        = "Workload %s status changed from %s to %s"
	workloadStatusNewTitle           = "Workload %s status changed to %s"
	statusChangeTitle                = "Status changed from %s to %s"
	statusNewTitle                   = "Status changed to %s"
	atTimeMessage                    = "at %s"
	detailsMessage                   = "Details:"
	workloadAdditionalDetailsMessage = "Additional details: %s"
)

func InitWorkloadPhaseNotifier(ctx context.Context, stopCh <-chan struct{}, conf *config.Notify, notifyClient NotifyClient) {
	if conf.PhaseChangeIntervalMillis <= 0 {
		logging.Infof("NotifyPhaseChangeIntervalMillis is set to 0, skipping workload phase notifier initialization")
		return
	}

	logging.Infof("Setting up a go routine for workload phase notifier. Notify interval: %d milliseconds", conf.PhaseChangeIntervalMillis)
	interval := time.Duration(conf.PhaseChangeIntervalMillis) * time.Millisecond
	ticker := time.NewTicker(interval)

	lastFetchedPhaseTimePerTenant = map[string]time.Time{}

	// sleep for the interval time, to let the correct data to be filled after startup
	time.Sleep(interval)

	notifierLock := &sync.Mutex{}
	go func() {
		go runWorkloadPhaseNotifier(ctx, conf, notifierLock, notifyClient)
		for {
			select {
			case <-ticker.C:
				go runWorkloadPhaseNotifier(ctx, conf, notifierLock, notifyClient)
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func runWorkloadPhaseNotifier(ctx context.Context, conf *config.Notify, notifierLock *sync.Mutex,
	notifyClient NotifyClient,
) {
	if !notifierLock.TryLock() {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Workload phase notifier is still in progress - not starting new sync")
		return
	}
	defer notifierLock.Unlock()

	tenantsToNotify := notifyClient.GetEnabledTenants()
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Starting workload phase notifier for %d tenants - %+v", len(tenantsToNotify), tenantsToNotify)

	for _, tenantId := range tenantsToNotify {
		lastFetchTime, ok := lastFetchedPhaseTimePerTenant[tenantId]
		if !ok {
			lastFetchTime = time.Now().UTC().Add(-time.Duration(conf.PhaseChangeIntervalMillis) * time.Millisecond)
			lastFetchedPhaseTimePerTenant[tenantId] = lastFetchTime
		}

		tenantRecords, dbErr := dao.DbGetTenantWorkloadPhaseHistoryRecords(ctx, tenantId, lastFetchTime)
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload history records for tenant %s, error: %v", tenantId, dbErr)
			continue
		}

		tenantBaseUrl, err := notifyClient.GetTenantBaseUrl(tenantId)
		if err != nil {
			if !errors.IsNotFound(err) {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant base url for tenant %s, error: %v", tenantId, err)
			}
			continue
		}

		recordsByWorkloadId := groupWorkloadsPhaseHistoryByWorkloadID(tenantRecords)

		for _, workloadRecords := range recordsByWorkloadId {
			notifyWorkloadPhaseChanges(ctx, notifyClient, workloadRecords, tenantBaseUrl)
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Notified workload phase change for %d workloads in tenant %s", len(recordsByWorkloadId), tenantId)

		// update last fetch time per tenant with db timestamp
		numOfRecordsPerTenant := len(*tenantRecords)
		if numOfRecordsPerTenant > 0 {
			lastFetchedPhaseTimePerTenant[tenantId] = (*tenantRecords)[numOfRecordsPerTenant-1].CreatedAt
		}
	}
}

func notifyWorkloadPhaseChanges(ctx context.Context, notifyClient NotifyClient, workloadRecords []models.WorkloadPhasesView,
	tenantBaseUrl string,
) {
	numOfRecords := len(workloadRecords)
	if numOfRecords == 0 {
		return
	}

	firstRecord := workloadRecords[0]
	lastRecord := workloadRecords[numOfRecords-1]
	additionalData := NotificationAdditionalData{
		WorkloadName: firstRecord.Name,
		WorkloadType: firstRecord.Type,
		ProjectName:  firstRecord.ProjectName,
		ClusterId:    firstRecord.ClusterID.String(),
		TenantId:     firstRecord.TenantID,
		User:         firstRecord.SubmittedBy,
		Time:         lastRecord.CreatedAt,
		URL:          getWorkloadInfoUrl(tenantBaseUrl, firstRecord.ProjectName, firstRecord.Name, firstRecord.Type),
	}

	workloadStatusChangeType, ok := phaseToStatusChangeNotificationType[api.Phase(lastRecord.Phase)]
	if !ok {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to find status change type for phase <%s> for workload <%s> (name: <%s>) in project <%s>",
			lastRecord.Phase, firstRecord.WorkloadID, firstRecord.Name, firstRecord.ProjectName)
		return
	}

	notifyStatusChange(ctx, notifyClient, workloadStatusChangeType,
		workloadRecords, additionalData)
}

func notifyStatusChange(ctx context.Context, notifyClient NotifyClient, workloadStatusChangeType WorkloadStatusChangeType,
	workloadRecords []models.WorkloadPhasesView,
	additionalData NotificationAdditionalData,
) {
	title, messageBody := getPhaseChangeNotificationTitleAndMessageBody(additionalData, workloadRecords)

	builder := notifyClient.BuildNotification(workloadStatusChangeCategory, string(workloadStatusChangeType))
	notifyWorkloadEvent(ctx, builder, title, messageBody, additionalData)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Notified workload phase changes <%v> for workload name <%s> in project <%s> in tenant %d",
		workloadStatusChangeType, additionalData.WorkloadName, additionalData.ProjectName, additionalData.TenantId)
}

func getPhaseChangeNotificationTitleAndMessageBody(additionalData NotificationAdditionalData, workloadRecords []models.WorkloadPhasesView) (string, string) {
	oldStatus := workloadRecords[0].OldPhase
	newStatus := workloadRecords[len(workloadRecords)-1].Phase

	var title string
	if oldStatus != "" {
		title = fmt.Sprintf(workloadStatusChangeTitle, additionalData.WorkloadName, oldStatus, newStatus)
	} else {
		title = fmt.Sprintf(workloadStatusNewTitle, additionalData.WorkloadName, newStatus)
	}

	messageParts := []string{title}

	additionalInfoMsg := getAdditionalInfoMsg(workloadRecords)
	if additionalInfoMsg != "" {
		messageParts = append(messageParts, additionalInfoMsg)
	}

	messageBody := strings.Join(messageParts, lineSeparator)

	return title, messageBody
}

func getAdditionalInfoMsg(records []models.WorkloadPhasesView) string {
	if len(records) == 1 {
		return getPhaseMessageInnerMsg(records[0].PhaseMessage)
	}

	recordsMessages := []string{detailsMessage}
	for _, record := range records {
		recordsMessages = append(recordsMessages, getSingleRecordAdditionalInfoMsg(record))
	}

	return strings.Join(recordsMessages, lineSeparator)
}

func getSingleRecordAdditionalInfoMsg(record models.WorkloadPhasesView) string {
	innerMsg := ""
	if record.OldPhase != "" {
		innerMsg = fmt.Sprintf(statusChangeTitle, record.OldPhase, record.Phase)
	} else {
		innerMsg = fmt.Sprintf(statusNewTitle, record.Phase)
	}

	if record.PhaseMessage != "" {
		innerMsg = fmt.Sprintf("%s. %s", innerMsg, getPhaseMessageInnerMsg(record.PhaseMessage))
	}

	atTime := fmt.Sprintf(atTimeMessage, getTimeStr(record.CreatedAt))
	innerMsg = fmt.Sprintf("%s, %s", innerMsg, atTime)

	return innerMsg
}

func getPhaseMessageInnerMsg(phaseMessage string) string {
	if phaseMessage == "" {
		return ""
	}

	return fmt.Sprintf(workloadAdditionalDetailsMessage, phaseMessage)
}

func groupWorkloadsPhaseHistoryByWorkloadID(tenantWorkloadHistoryRecords *[]models.WorkloadPhasesView) map[uuid.UUID][]models.WorkloadPhasesView {
	workloadsPhaseHistory := make(map[uuid.UUID][]models.WorkloadPhasesView)

	for _, record := range *tenantWorkloadHistoryRecords {
		_, ok := workloadsPhaseHistory[record.WorkloadID]
		if !ok {
			workloadsPhaseHistory[record.WorkloadID] = []models.WorkloadPhasesView{}
		}

		workloadsPhaseHistory[record.WorkloadID] = append(workloadsPhaseHistory[record.WorkloadID], record)
	}

	return workloadsPhaseHistory
}
