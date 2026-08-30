package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/metrics"

	"github.com/hashicorp/go-multierror"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-application-packages/concurrency"
	workloads_openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config"
	configmaps "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config-maps"
	dinfwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"

	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	maxNumWorkloadWorkers = 4

	annotationClusterId = "clusterId"
)

type clusterWorkloadsAndPoliciesUpdates struct {
	ToUpdate  []client.Object
	UpdateRef []int // the reference of this update to the array of workloads/policies that were received from the CP

	ToDelete    []types.NamespacedName
	ToDeleteGvk []schema.GroupVersionKind // we need the GVK when deleting
	DeleteRef   []int                     // the reference of this deletion to the array of workloads/policies that were received from the CP
}

// WorkloadsSyncReportInfo is the report to be sent to the agent on  each sync request
// reporting success/failure of the previous sync cycle. it is not a must-sent
// kind of information, but it helps the control plane in displaying failure/success
type WorkloadsSyncReportInfo struct {
	numWorkloads int
	theReport    workloads_openapi.ClusterSyncReport
}

var syncReportPerWorkloadType = concurrency.NewConcurrentMap[*WorkloadsSyncReportInfo](nil)

type WorkloadsHandler struct {
	*CommonHandler
	syncCounter  int64
	workloadType string
	metricLabel  string
}

func isWorkloadDeletionComplete(err error) bool {
	return apierrors.IsNotFound(err) || meta.IsNoMatchError(err)
}

func (w *WorkloadsHandler) deleteLegacyWorkloadObjects(objectNames []types.NamespacedName, objectGVKs []schema.GroupVersionKind) ([]error, error) {
	var err error
	perObjectErrors := make([]error, len(objectNames))

	for idx, objectName := range objectNames {
		object, innerErr := w.getObjectByGVK(objectName.Name, objectName.Namespace, objectGVKs[idx])
		if innerErr != nil {
			if isWorkloadDeletionComplete(innerErr) {
				log.Debug().Msgf("%s <%s/%s> is already deleted from the cluster - doing nothing",
					objectGVKs[idx].Kind, objectName.Namespace, objectName.Name)
				continue
			}

			log.Error().Msgf("Failed to get %s <%s/%s>, error: %s", objectGVKs[idx].Kind, objectName.Namespace, objectName.Name, innerErr.Error())
			err = multierror.Append(err, innerErr)
			perObjectErrors[idx] = innerErr
			continue
		}

		innerErr = w.deleteObject(object)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			perObjectErrors[idx] = innerErr
			w.createDeletionEvent(object, innerErr)
		}
	}

	return perObjectErrors, err
}

func (w *WorkloadsHandler) MetricLabel() string {
	return w.metricLabel
}

func (w *WorkloadsHandler) Handle() error {
	var err error
	var prevHighestUpdate int64
	numWorkers := maxNumWorkloadWorkers

	//
	//  make sure each workload type has an entry in the sync report
	//
	syncReport, found := syncReportPerWorkloadType.Load(w.workloadType)
	if !found {
		syncReport = &WorkloadsSyncReportInfo{}
		syncReportPerWorkloadType.Store(w.workloadType, syncReport)
	}

	for {
		backendWorkloads, innerErr := w.getBackendWorkloads(w.syncCounter)
		if innerErr != nil {
			log.Error().Msgf("Had an error while getting backend %s workloads, sync counter: %d, error: %s", w.workloadType, w.syncCounter, innerErr.Error())
			return innerErr
		}

		if backendWorkloads == nil {
			log.Error().Msgf("%s workloads were not received from the backend for sync counter %d", w.workloadType, w.syncCounter)
			return innerErr
		}

		var highestUpdate, highestSysUpdate int64
		if backendWorkloads.ResponseInfo != nil && backendWorkloads.ResponseInfo.HighestUpdate != nil {
			highestUpdate = *backendWorkloads.ResponseInfo.HighestUpdate
		}

		if backendWorkloads.SystemInfo != nil && backendWorkloads.SystemInfo.HighestUpdate != nil {
			highestSysUpdate = *backendWorkloads.SystemInfo.HighestUpdate
		}

		numWorkloads := len(backendWorkloads.Workloads)
		log.Debug().Msgf("Backend %s workloads response - prev highest update: %d, highest update: %d, highest system update: %d, size: %d", w.workloadType, prevHighestUpdate, highestUpdate, highestSysUpdate, numWorkloads)

		syncReport.numWorkloads = numWorkloads

		if numWorkloads == 0 {
			log.Debug().Msgf("No %s workloads to handle while counter is %d", w.workloadType, w.syncCounter)
			break
		}

		syncReport.theReport.Updates = make([]workloads_openapi.ClusterSyncReportUpdate, numWorkloads)

		innerErr = w.ProcessWorkloadConcurrently(context.Background(), numWorkloads, numWorkers, backendWorkloads)
		if innerErr != nil {
			log.Error().Msgf("Had an error while processing %s workloads, error: %s", w.workloadType, innerErr.Error())
			err = multierror.Append(err, innerErr)
		}

		// update the highest update counter
		// NOTE! we're updating the sync counter even if there was an error in processing thr workloads
		// because there is no point to attempting again.
		innerErr = w.setSyncCounter(highestUpdate)
		if innerErr != nil {
			log.Error().Msgf("Had an error while updating %s workloads sync counter, error: %s", w.workloadType, innerErr.Error())
			err = multierror.Append(err, innerErr)
		}

		// if the highest update counter is the same as the highest system update counter, it means that there are no more cluster workloads to handle
		if highestUpdate == highestSysUpdate {
			log.Debug().Msgf("No more %s workloads to handle", w.workloadType)
			break
		}
	}

	if err != nil {
		log.Error().Msgf("Failed to handle some of cluster workloads (create or update), error: %s", err.Error())
		return err
	}

	log.Debug().Msgf("Finished handling %s workloads", w.workloadType)
	return nil
}

func (w *WorkloadsHandler) ProcessWorkloadConcurrently(ctx context.Context, numWorkloads int, numWorkers int, workloads *workloads_openapi.WorkloadsSyncResponse) error {
	var wg sync.WaitGroup
	var err error

	syncReport, _ := syncReportPerWorkloadType.Load(w.workloadType)

	// to be on the safe side make sure at least one worker is running
	if numWorkers <= 0 {
		log.Warn().Msgf("number of workers changed from %d to 1", numWorkers)
		numWorkers = 1
	}

	chunkSize := numWorkloads / numWorkers

	if chunkSize == 0 {
		// if there are fewer workloads than workers, we will have only one worker who will handle all workloads
		numWorkers = 1
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		// Calculate the start and end indices for each worker's chunk
		start := i * chunkSize
		end := (i + 1) * chunkSize

		// The last worker may get extra items if numWorkloads is not divisible by numWorkers
		if i == numWorkers-1 {
			end = numWorkloads
		}

		log.Debug().Msgf("Worker %d: start processing, start: %v, end: %v", i, start, end)
		go func() {
			perWorkloadError, innerErr := w.ProcessWorkloads(ctx, workloads.Workloads[start:end], &wg)
			if innerErr != nil {
				err = multierror.Append(err, innerErr)
			}

			for idx := start; idx < end; idx++ {
				syncReport.theReport.Updates[idx] = workloads_openapi.ClusterSyncReportUpdate{
					Id:                convutils.Ref(workloads.Workloads[idx].Workload.WorkloadId),
					UpdateCount:       workloads.Workloads[idx].UpdateCount,
					UpdateAttemptTime: convutils.Ref(time.Now()),
				}
				if perWorkloadError[idx-start] != nil {
					syncReport.theReport.Updates[idx].Message = *workloads_openapi.NewNullableString(convutils.Ref(perWorkloadError[idx-start].Error()))
				}
			}
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
	return err
}

func (w *WorkloadsHandler) ProcessWorkloads(ctx context.Context, workloads []workloads_openapi.WorkloadsSyncInfo, wg *sync.WaitGroup) ([]error, error) {
	defer wg.Done()
	var err error

	clusterUpdates, allObjectsErrors, innerErr := w.getClusterUpdates(ctx, workloads)
	if innerErr != nil {
		log.Error().Msgf("Had an error while convering workloads: %s", innerErr.Error())
		err = multierror.Append(err, innerErr)
	}

	metrics.GetMetricsExporter().AddResourceAction(w.MetricLabel(), metrics.ActionDelete, len(clusterUpdates.ToDelete))

	if len(clusterUpdates.ToDelete) > 0 {
		perObjectError, innerErr := w.deleteLegacyWorkloadObjects(clusterUpdates.ToDelete, clusterUpdates.ToDeleteGvk)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}

		for idx, objectErr := range perObjectError {
			toIndex := clusterUpdates.DeleteRef[idx]
			allObjectsErrors[toIndex] = objectErr
		}
	}

	return allObjectsErrors, err
}

func (w *WorkloadsHandler) setSyncCounter(highestUpdate int64) error {
	w.syncCounter = highestUpdate
	switch w.workloadType {
	case common.WorkloadTypeInteractive:
		err := configmaps.UpdateInteractiveWorkloadsSyncCounters(w.runaiClient, w.cachedClient, highestUpdate)
		if err != nil {
			return err
		}
		log.Info().Msgf("Updated interactive workloads sync counter to: %d", highestUpdate)
	case common.WorkloadTypeTraining:
		err := configmaps.UpdateTrainingWorkloadsSyncCounters(w.runaiClient, w.cachedClient, highestUpdate)
		if err != nil {
			return err
		}
		log.Info().Msgf("Updated training workloads sync counter to: %d", highestUpdate)
	case common.WorkloadTypeDistributed:
		err := configmaps.UpdateDistributedWorkloadsSyncCounters(w.runaiClient, w.cachedClient, highestUpdate)
		if err != nil {
			return err
		}
		log.Info().Msgf("Updated distributed workloads sync counter to: %d", highestUpdate)
	case common.WorkloadTypeInference:
		err := configmaps.UpdateInferenceWorkloadsSyncCounters(w.runaiClient, w.cachedClient, highestUpdate)
		if err != nil {
			return err
		}
		log.Info().Msgf("Updated inference workloads sync counter to: %d", highestUpdate)
	case common.WorkloadTypeDistributedInference:
		err := configmaps.UpdateDistributedInferenceWorkloadsSyncCounters(w.runaiClient, w.cachedClient, highestUpdate)
		if err != nil {
			return err
		}
		log.Info().Msgf("Updated distributed inference workloads sync counter to: %d", highestUpdate)
	}
	return nil
}

func (w *WorkloadsHandler) getClusterUpdates(ctx context.Context, backendUpdates []workloads_openapi.WorkloadsSyncInfo) (clusterWorkloadsAndPoliciesUpdates, []error, error) {
	clusterUpdates := clusterWorkloadsAndPoliciesUpdates{
		DeleteRef: make([]int, 0, len(backendUpdates)),
		ToDelete:  make([]types.NamespacedName, 0),
	}
	perWorkloadErrors := make([]error, len(backendUpdates))

	projectToNamespaceMap, err := w.ProjectToNamespaceList()
	if err != nil {
		for idx := range perWorkloadErrors {
			perWorkloadErrors[idx] = err
		}
		return clusterUpdates, perWorkloadErrors, err
	}
	for idx, backendUpdate := range backendUpdates {
		if backendUpdate.Workload.DesiredPhase != workloads_openapi.WORKLOADDESIREDPHASE_DELETED {
			perWorkloadErrors[idx] = fmt.Errorf("legacy workload %s only supports deletion", backendUpdate.Workload.Name)
			continue
		}

		namespace := projectToNamespaceMap[backendUpdate.Workload.ProjectId]
		if namespace == "" {
			log.Error().Msgf("Unknown project %s, skipped", backendUpdate.Workload.ProjectId)
			perWorkloadErrors[idx] = fmt.Errorf("unknown project %s", backendUpdate.Workload.ProjectId)
			continue
		}

		clusterUpdates.ToDelete = append(clusterUpdates.ToDelete, types.NamespacedName{
			Namespace: namespace,
			Name:      backendUpdate.Workload.Name,
		})
		clusterUpdates.ToDeleteGvk = append(clusterUpdates.ToDeleteGvk, w.gvk)
		clusterUpdates.DeleteRef = append(clusterUpdates.DeleteRef, idx)
	}

	return clusterUpdates, perWorkloadErrors, nil
}

func (w *WorkloadsHandler) getBackendWorkloads(prevHighestUpdate int64) (*workloads_openapi.WorkloadsSyncResponse, error) {
	var syncResponse *workloads_openapi.WorkloadsSyncResponse
	var res *http.Response
	var err error

	syncReport, _ := syncReportPerWorkloadType.Load(w.workloadType)

	switch w.workloadType {
	case common.WorkloadTypeInteractive:
		request := w.backendClient.WorkloadsClient.WorkspacesAPI.RequestWorkspacesUpdates(w.backendClient.Context())
		request = request.ClusterId(config.Config().ClusterUuid)
		request = request.UpdatesAfter(prevHighestUpdate)
		if syncReport.numWorkloads > 0 {
			request = request.ClusterSyncReport(syncReport.theReport)
		}
		syncResponse, res, err = request.Execute()
		err = w.backendClient.HandleHttpResponse(res, err, "get workspace updates")
	case common.WorkloadTypeTraining:
		request := w.backendClient.WorkloadsClient.TrainingsAPI.RequestTrainingsUpdates(w.backendClient.Context())
		request = request.ClusterId(config.Config().ClusterUuid)
		request = request.UpdatesAfter(prevHighestUpdate)
		if syncReport.numWorkloads > 0 {
			request = request.ClusterSyncReport(syncReport.theReport)
		}
		syncResponse, res, err = request.Execute()
		err = w.backendClient.HandleHttpResponse(res, err, "get trainings updates")
	case common.WorkloadTypeDistributed:
		request := w.backendClient.WorkloadsClient.DistributedAPI.RequestDistributedUpdates(w.backendClient.Context())
		request = request.ClusterId(config.Config().ClusterUuid)
		request = request.UpdatesAfter(prevHighestUpdate)
		if syncReport.numWorkloads > 0 {
			request = request.ClusterSyncReport(syncReport.theReport)
		}
		syncResponse, res, err = request.Execute()
		err = w.backendClient.HandleHttpResponse(res, err, "get distributed updates")
	case common.WorkloadTypeInference:
		request := w.backendClient.WorkloadsClient.InferencesAPI.RequestInferencesUpdates(w.backendClient.Context())
		request = request.ClusterId(config.Config().ClusterUuid)
		request = request.UpdatesAfter(prevHighestUpdate)
		if syncReport.numWorkloads > 0 {
			request = request.ClusterSyncReport(syncReport.theReport)
		}
		syncResponse, res, err = request.Execute()
		err = w.backendClient.HandleHttpResponse(res, err, "get inference updates")
	case common.WorkloadTypeDistributedInference:
		request := w.backendClient.WorkloadsClient.DistributedInferencesAPI.RequestDistributedInferencesUpdates(w.backendClient.Context())
		request = request.ClusterId(config.Config().ClusterUuid)
		request = request.UpdatesAfter(prevHighestUpdate)
		if syncReport.numWorkloads > 0 {
			request = request.ClusterSyncReport(syncReport.theReport)
		}
		syncResponse, res, err = request.Execute()
		err = w.backendClient.HandleHttpResponse(res, err, "get distributed inference updates")
	}

	if err != nil {
		log.Error().Msgf("Had an error while getting backend workload: %s", err.Error())
		return nil, err
	}

	return syncResponse, nil
}

func NewWorkloadsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader, workloadType string) (ResourceHandler, error) {
	var object client.Object

	switch workloadType {
	case common.WorkloadTypeInteractive:
		object = &iwv2.InteractiveWorkload{}

	case common.WorkloadTypeTraining:
		object = &twv2.TrainingWorkload{}

	case common.WorkloadTypeDistributed:
		object = &dwv2.DistributedWorkload{}

	case common.WorkloadTypeInference:
		object = &infwv2.InferenceWorkload{}

	case common.WorkloadTypeDistributedInference:
		object = &dinfwv1.DistributedInferenceWorkload{}
	default:
		return nil, fmt.Errorf("unsupported workload type: %s", workloadType)
	}

	handlingOptions := HandlingOptions{deletePropagationOrphan: false}
	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient, object, nil, handlingOptions)
	if err != nil {
		return nil, err
	}

	agentConfigData, _ := configmaps.GetAgentConfig(cachedClient)
	switch workloadType {
	case common.WorkloadTypeDistributedInference:
		return &WorkloadsHandler{
			CommonHandler: commonHandler,
			workloadType:  workloadType,
			syncCounter:   agentConfigData.DistributedInferenceWorkloadsSyncCounter,
			metricLabel:   "distributed_inference_workload",
		}, nil
	case common.WorkloadTypeInteractive:
		return &WorkloadsHandler{
			CommonHandler: commonHandler,
			workloadType:  workloadType,
			syncCounter:   agentConfigData.InteractiveWorkloadsSyncCounter,
			metricLabel:   "interactive_workload",
		}, nil
	case common.WorkloadTypeTraining:
		return &WorkloadsHandler{
			CommonHandler: commonHandler,
			workloadType:  workloadType,
			syncCounter:   agentConfigData.TrainingWorkloadsSyncCounter,
			metricLabel:   "training_workload",
		}, nil
	case common.WorkloadTypeDistributed:
		return &WorkloadsHandler{
			CommonHandler: commonHandler,
			workloadType:  workloadType,
			syncCounter:   agentConfigData.DistributedWorkloadsSyncCounter,
			metricLabel:   "distributed_workload",
		}, nil
	case common.WorkloadTypeInference:
		return &WorkloadsHandler{
			CommonHandler: commonHandler,
			workloadType:  workloadType,
			syncCounter:   agentConfigData.InferenceWorkloadsSyncCounter,
			metricLabel:   "inference_workload",
		}, nil
	}

	return nil, fmt.Errorf("unsupported workload type: %s", workloadType)
}

func NewInteractiveWorkloadsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	return NewWorkloadsHandler(runaiBackendClient, runaiClient, cachedClient, common.WorkloadTypeInteractive)
}

func NewTrainingWorkloadsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	return NewWorkloadsHandler(runaiBackendClient, runaiClient, cachedClient, common.WorkloadTypeTraining)
}

func NewDistributedWorkloadsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	return NewWorkloadsHandler(runaiBackendClient, runaiClient, cachedClient, common.WorkloadTypeDistributed)
}

func NewInferenceWorkloadsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	return NewWorkloadsHandler(runaiBackendClient, runaiClient, cachedClient, common.WorkloadTypeInference)
}

func NewDistributedInferenceWorkloadsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	return NewWorkloadsHandler(runaiBackendClient, runaiClient, cachedClient, common.WorkloadTypeDistributedInference)
}
