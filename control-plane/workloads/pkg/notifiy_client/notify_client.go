package notifiy_client

import (
	"context"
	"fmt"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"strconv"
	"time"

	"github.com/avast/retry-go/v4"
	notifyclient "github.com/run-ai/runai/control-plane/notifications-service/pkg/client"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

const (
	workloadStatusChangeCategory = "Workload status change"
	workloadTimeoutCategory      = "Workload timeout"

	templateName = "workloads"

	nameProperty    = "name"
	typeProperty    = "workload_type"
	projectProperty = "project"
	clusterProperty = "cluster"
	timeProperty    = "time"
	urlProperty     = "workload_url"

	// the line separator for the message body of the email notification
	lineSeparator = "<br>"

	workloadInfoUrl = "%s/workloads?sortBy=name&advancedFilters=[" +
		"{\"term\":\"%s\",\"value\":\"==\",\"label\":\"Project\",\"name\":\"projectName\",\"type\":\"FreeText\"}," +
		"{\"term\":\"%s\",\"value\":\"==\",\"label\":\"Workload\",\"name\":\"name\",\"type\":\"FreeText\"}," +
		"{\"term\":\"%s\",\"value\":\"=@\",\"label\":\"Type\",\"name\":\"type\",\"type\":\"FreeText\"}]"
)

type (
	WorkloadStatusChangeType string
	WorkloadTimeoutType      string
)

const (
	WorkloadStatusChangeCreating     WorkloadStatusChangeType = "Creating"
	WorkloadStatusChangePending      WorkloadStatusChangeType = "Pending"
	WorkloadStatusChangeInitializing WorkloadStatusChangeType = "Initializing"
	WorkloadStatusChangeRunning      WorkloadStatusChangeType = "Running"
	WorkloadStatusChangeDegraded     WorkloadStatusChangeType = "Degraded"
	WorkloadStatusChangeCompleted    WorkloadStatusChangeType = "Completed"
	WorkloadStatusChangeFailed       WorkloadStatusChangeType = "Failed"
	WorkloadStatusChangeStopped      WorkloadStatusChangeType = "Stopped"
	WorkloadStatusChangeDeleting     WorkloadStatusChangeType = "Deleting"

	WorkloadTimeoutTypeWorkspace WorkloadTimeoutType = "Workspace workload"
	WorkloadTimeoutTypeTraining  WorkloadTimeoutType = "Training workload"
)

type NotificationAdditionalData struct {
	WorkloadName string
	WorkloadType string
	ProjectName  string
	ClusterId    string
	TenantId     uint
	User         string
	Time         time.Time
	URL          string
}

func InitGoRoutines(ctx context.Context, stopCh <-chan struct{}) {
	notifyClient := NewNotifyClienter(runai_config.Get[*config.ServicesConfig]())
	go initGoRoutines(ctx, stopCh, notifyClient, runai_config.Get[*config.Notify]())
}

func initGoRoutines(ctx context.Context, stopCh <-chan struct{}, notifyClient NotifyClient,
	conf *config.Notify,
) {
	retryAttempts := 5
	retryDelaySeconds := 5
	err := retry.Do(func() error {
		return initNotifyClienter(ctx, conf, notifyClient)
	}, retry.Delay(time.Duration(retryDelaySeconds)*time.Second), retry.Attempts(uint(retryAttempts)), retry.OnRetry(func(attempt uint, err error) {
		logging.Debugf("Failed to init notify client, retrying, attempt: %d, error: %s", attempt, err.Error())
	}))
	if err != nil {
		logging.Errorf("Failed to init notify client after %d retries, error: %s", retryAttempts, err.Error())
		return
	}
	InitWorkloadPhaseNotifier(ctx, stopCh, conf, notifyClient)
	InitWorkloadTimeoutNotifier(ctx, stopCh, conf, notifyClient)
}

func initNotifyClienter(ctx context.Context, conf *config.Notify, notifyClient NotifyClient) error {
	notifyConfig := &notifyclient.NotifyClientConfig{
		BatchSize:        conf.BatchSize,
		BatchIntervalSec: conf.BatchIntervalSec,
		NotifyUrl:        conf.NotifyUrl,
		ConfigUrl:        conf.ConfigUrl,
	}

	err := notifyClient.InitNotifyClient(ctx, notifyConfig)
	if err != nil {
		return fmt.Errorf("failed to init notify client, error: %s", err.Error())
	}

	err = registerNotifications(notifyClient)
	if err != nil {
		return fmt.Errorf("failed to register notifications, error: %s", err.Error())
	}

	logging.Infof("Notify client initialized successfully")
	return nil
}

// TODOs:
// 1. tenant url - need to open a follow up task to use the same function of "createTenantUrl" that is used upon tenant creation,
// move the func to shared-packages and use it from workload-service.
// or if if we don't even know it then it needs to be an internal API to tenant-service...

func notifyWorkloadEvent(ctx context.Context, builder notifyclient.NotifyBuilderer,
	title, messageBody string,
	additionalData NotificationAdditionalData,
) {
	tenantIdStr := strconv.FormatUint(uint64(additionalData.TenantId), 10)

	builder = builder.WithTitle(title).WithDescription(messageBody).WithTenant(tenantIdStr).WithUser(additionalData.User)

	// if "submittedBy" is empty - will not send an email, but will still send to distribution list

	builder = builder.WithProperty(nameProperty, additionalData.WorkloadName).
		WithProperty(typeProperty, additionalData.WorkloadType).
		WithProperty(projectProperty, additionalData.ProjectName).
		// WithProperty(clusterProperty, additionalData.ClusterId).//no need at the moment
		WithProperty(timeProperty, getTimeStr(additionalData.Time)).
		WithProperty(urlProperty, additionalData.URL)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Sending notification: %s // %+v", title, additionalData)
	builder.Notify()
}

func registerNotifications(notifyClient NotifyClient) error {
	logging.Infof("Registering notifications")
	return notifyClient.RegisterNotifications("workload-service", []*notifyclient.NotificationKind{
		{
			Category:    workloadStatusChangeCategory,
			Type:        string(WorkloadStatusChangeCreating),
			Description: "Workload and associated resources are being created in the cluster",
			Template:    templateName,
		},
		{
			Category:    workloadStatusChangeCategory,
			Type:        string(WorkloadStatusChangePending),
			Description: "Workload is now awaiting the scheduling of all pods",
			Template:    templateName,
		},
		{
			Category:    workloadStatusChangeCategory,
			Type:        string(WorkloadStatusChangeInitializing),
			Description: "Workload is now pulling images and starting initialize containers",
			Template:    templateName,
		},
		{
			Category:    workloadStatusChangeCategory,
			Type:        string(WorkloadStatusChangeRunning),
			Description: "Workload is actively processing with all pods initialized",
			Template:    templateName,
		},
		{
			Category:    workloadStatusChangeCategory,
			Type:        string(WorkloadStatusChangeDegraded),
			Description: "Workload has issues or reduced resources",
			Template:    templateName,
		},
		{
			Category:    workloadStatusChangeCategory,
			Type:        string(WorkloadStatusChangeCompleted),
			Description: "Workload has successfully completed its process",
			Template:    templateName,
		},
		{
			Category:    workloadStatusChangeCategory,
			Type:        string(WorkloadStatusChangeFailed),
			Description: "Workload has failed to complete successfully",
			Template:    templateName,
		},
		{
			Category:    workloadStatusChangeCategory,
			Type:        string(WorkloadStatusChangeStopped),
			Description: "Workload stopped, all resources are up but pods are down",
			Template:    templateName,
		},
		{
			Category:    workloadStatusChangeCategory,
			Type:        string(WorkloadStatusChangeDeleting),
			Description: "Workload and associated resources are being removed from the cluster",
			Template:    templateName,
		},

		{
			Category:    workloadTimeoutCategory,
			Type:        string(WorkloadTimeoutTypeWorkspace),
			Description: "Workspace workload is about to be stopped due to timeout",
			Template:    templateName,
		},
		{
			Category:    workloadTimeoutCategory,
			Type:        string(WorkloadTimeoutTypeTraining),
			Description: "Training workload is about to be stopped due to timeout",
			Template:    templateName,
		},
	}, ptr.To(true))
}

// general helper functions for the notification messages
func getWorkloadInfoUrl(tenantBaseUrl, projectName, workloadName, workloadType string) string {
	return fmt.Sprintf(workloadInfoUrl, tenantBaseUrl, projectName, workloadName, workloadType)
}

func getTimeStr(tim time.Time) string {
	return tim.UTC().Format(time.DateTime) + " UTC"
}
