package handlers

import (
	"context"
	"fmt"
	"sync"

	swagger "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/utils/ptr"

	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	_ "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ApplicationHandler struct {
	resourceName  constants.ClusterResource
	runaiClient   client.Client
	backendClient *runaiBackendClient.BackendClient
}

func NewApplicationHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client) *ApplicationHandler {
	return &ApplicationHandler{
		resourceName:  constants.ApplicationResourceName,
		runaiClient:   runaiClient,
		backendClient: runaiBackendClient,
	}
}

func (h *ApplicationHandler) HandleUpdateItem(key client.ObjectKey) error {
	application := &v1alpha1.Application{}
	err := h.runaiClient.Get(context.Background(), key, application)
	if err != nil {
		return err
	}

	return h.HandleAddUpdate(application)
}

func (h *ApplicationHandler) HandleAddUpdate(application *v1alpha1.Application) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.UpsertOperation)()
	backendApplication, err := parsers.ParseBackendApplication(application, false)
	if err != nil {
		return err
	}
	request := h.backendClient.WorkloadsClient.AIApplicationsAPI.
		SyncApplications(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		AIApplicationsForSync(swagger.AIApplicationsForSync{
			FullSync:     ptr.To(false),
			Applications: []swagger.AIApplicationForSync{backendApplication},
		})
	res, err := h.backendClient.WorkloadsClient.AIApplicationsAPI.SyncApplicationsExecute(request)
	return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Upsert application %s", application.Name))
}

func (h *ApplicationHandler) HandleDelete(application *v1alpha1.Application) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.DeleteOperation)()

	backendApplication, err := parsers.ParseBackendApplication(application, true)
	if err != nil {
		return err
	}
	request := h.backendClient.WorkloadsClient.AIApplicationsAPI.
		SyncApplications(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		AIApplicationsForSync(swagger.AIApplicationsForSync{
			FullSync:     ptr.To(false),
			Applications: []swagger.AIApplicationForSync{backendApplication},
		})
	res, err := h.backendClient.WorkloadsClient.AIApplicationsAPI.SyncApplicationsExecute(request)
	return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Deleted application %s", application.Name))
}

func (h *ApplicationHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.SyncOperation)()
	defer waitGroup.Done()
	zap.S().Infof("Start applications Sync")

	applications := &v1alpha1.ApplicationList{}
	kubeErr := h.runaiClient.List(context.Background(), applications)
	if kubeErr != nil {
		zap.S().Errorf("Failed listing applications, error: %s", kubeErr.Error())
		return
	}

	backendApplications := []swagger.AIApplicationForSync{}
	for _, application := range applications.Items {
		backendApplication, err := parsers.ParseBackendApplication(&application, false)
		if err != nil {
			return
		}
		backendApplications = append(backendApplications, backendApplication)
	}
	request := h.backendClient.WorkloadsClient.AIApplicationsAPI.
		SyncApplications(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		AIApplicationsForSync(swagger.AIApplicationsForSync{
			FullSync:     ptr.To(true),
			Applications: backendApplications,
		})
	_, err := h.backendClient.WorkloadsClient.AIApplicationsAPI.SyncApplicationsExecute(request)
	if err != nil {
		zap.S().Errorf("Failed syncing applications: %+v", err)
	}

}
