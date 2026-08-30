package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectHandler struct {
	resourceName  constants.ClusterResource
	backendClient *runaiBackendClient.BackendClient
	cachedClient  client.Client
}

func NewProjectHandler(runaiBackendClient *runaiBackendClient.BackendClient, cachedClient client.Client) *ProjectHandler {
	return &ProjectHandler{
		resourceName:  constants.ProjectResourceName,
		backendClient: runaiBackendClient,
		cachedClient:  cachedClient,
	}
}

func (projectHandler *ProjectHandler) HandleUpdateItem(key client.ObjectKey) error {
	project := &kaiv1alpha1.Project{}
	err := projectHandler.cachedClient.Get(context.Background(), key, project)
	if err != nil {
		return err
	}
	return projectHandler.HandleAddUpdate(project)
}

func (projectHandler *ProjectHandler) HandleAddUpdate(project *kaiv1alpha1.Project) error {
	syncTime := metav1.Now()
	updateStatusErr := projectHandler.updateProjectStatus(project)
	if updateStatusErr == nil {
		err := projectHandler.updateProjectSyncStatus(project, syncTime)
		if err != nil {
			zap.S().Warnf("Failed to update last sync time in project %s, err: %v", project.Name, err)
		}
	}
	return updateStatusErr
}
func (projectHandler *ProjectHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start projects Sync")

	projectsInCluster := []openapi.ProjectInSyncRequest{}
	projects := &kaiv1alpha1.ProjectList{}
	kubeErr := projectHandler.cachedClient.List(context.Background(), projects)
	if kubeErr != nil {
		zap.S().Errorf("%+v", kubeErr)
		return
	}
	for _, project := range projects.Items {
		projectsInCluster = append(projectsInCluster, openapi.ProjectInSyncRequest{Id: project.Labels[parsers.ProjectIdLabel]})
	}
	{
		defer metrics.GetMetricsExporter().MeasureExecutionTime(projectHandler.resourceName, constants.SyncOperation)()
		response, err := projectHandler.backendClient.OrgUnitClient.SyncClusterProjects(
			projectHandler.backendClient.Context(),
			config.Config().ClusterUuid,
			projectsInCluster,
		)
		err = projectHandler.backendClient.HandleHttpResponse(response, err, "Sync projects")
		if err != nil {
			zap.S().Errorf("Had an error while syncing projects: %+v", err)
		}
	}
}

// updateProjectSyncStatus records the last sync time as an annotation on the project.
// On the new kai.resources Project CRD the sync timestamp is no longer a status field;
// it is kept as an annotation managed exclusively by cluster-sync.
func (projectHandler *ProjectHandler) updateProjectSyncStatus(project *kaiv1alpha1.Project, syncTime metav1.Time) error {
	patchBytes, err := json.Marshal(map[string]any{"metadata": map[string]any{
		"annotations": map[string]string{
			parsers.LastSyncTimestampAnnotation: syncTime.UTC().Format(time.RFC3339),
		},
	}})
	if err != nil {
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = projectHandler.cachedClient.Patch(context.Background(), project, patch)
	if err != nil {
		return err
	}
	return nil
}

func (projectHandler *ProjectHandler) updateProjectStatus(project *kaiv1alpha1.Project) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(projectHandler.resourceName, constants.UpdateOperation)()
	toUpdate := parsers.ParseKaiProjectToUpdateStatusRequest(project)
	res, err := projectHandler.backendClient.OrgUnitClient.UpdateProjectsStatus(
		projectHandler.backendClient.Context(),
		toUpdate,
	)
	return projectHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Updated status of project %s", project.Name))
}

func (projectHandler *ProjectHandler) HandleDelete(project *kaiv1alpha1.Project) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(projectHandler.resourceName, constants.DeleteOperation)()
	toDelete := parsers.ParseKaiProjectToDeleteRequest(project)

	res, err := projectHandler.backendClient.OrgUnitClient.UpdateProjectsStatus(
		projectHandler.backendClient.Context(),
		toDelete,
	)
	return projectHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Deleted project %s", project.Name))
}
