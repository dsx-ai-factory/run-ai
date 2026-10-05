/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	goErrors "errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/metrics"
	clusterCommonConstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	pcreds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/hashicorp/go-multierror"

	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
	configmaps "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config-maps"
	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"github.com/rs/zerolog/log"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	maxNumWorkloadsV2Workers = 4

	KindWorkloadRunner       = "WorkloadRunner"
	ApiVersionWorkloadRunner = "run.ai/v1alpha1"

	requestRefLabel = "run.ai/request-ref"
)

func NewWorkloadsV2Handler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {

	object := &unstructured.Unstructured{}
	objectInspector := UnstructuredObjectInspector{}

	handlingOptions := HandlingOptions{deletePropagationOrphan: false}

	commonHandler, err := NewUnstructuredHandler(runaiBackendClient, runaiClient, cachedClient, object, objectInspector, handlingOptions)
	if err != nil {
		return nil, err
	}

	agentConfigData, _ := configmaps.GetAgentConfig(cachedClient)
	updateCount, err := initialWorkloadsV2SyncCounter(
		cachedClient, agentConfigData.WorkloadsV2SyncCounter,
	)
	if err != nil {
		return nil, err
	}

	return &WorkloadsV2Handler{
		CommonHandler: commonHandler,
		updateCount:   updateCount,
		metricLabel:   "workloads_v2",
	}, nil
}

func initialWorkloadsV2SyncCounter(
	cachedClient client.Reader,
	configuredCounter int64,
) (int64, error) {
	runners := &dromosv1alpha1.WorkloadRunnerList{}
	if err := cachedClient.List(context.Background(), runners); err != nil {
		return 0, fmt.Errorf("failed to inspect WorkloadRunners for ownership backfill: %w", err)
	}
	for index := range runners.Items {
		runner := &runners.Items[index]
		if runner.Labels[constants.ControlPlaneManagedLabel] == "true" &&
			runner.Annotations[clusterCommonConstants.WorkloadOwnerUIDAnnotation] == "" {
			log.Info().Msg("Resetting workloads v2 sync counter to backfill WorkloadRunner ownership")
			return 0, nil
		}
	}
	return configuredCounter, nil
}

// WorkloadsV2SyncReport is the report to be sent to the backend on each sync request
// reporting success/failure of the previous sync cycle.
type WorkloadsV2SyncReport struct {
	numWorkloads int
	theReport    workloadsManagerApi.ClusterSyncReport
}

var syncReport WorkloadsV2SyncReport

type WorkloadsV2Handler struct {
	*CommonHandler

	updateCount int64
	metricLabel string
}

func (w *WorkloadsV2Handler) MetricLabel() string {
	return w.metricLabel
}

func (w *WorkloadsV2Handler) Handle() error {
	var err error
	var prevHighestUpdate int64
	numWorkers := maxNumWorkloadsV2Workers

	for {
		backendWorkloads, innerErr := w.getBackendWorkloads(w.updateCount)
		if innerErr != nil {
			log.Error().Msgf("Had an error while requesting workloads from backend, update count > %d, error: %s", w.updateCount, innerErr.Error())
			return innerErr
		}

		if backendWorkloads == nil {
			log.Error().Msgf("workloads v2 were not received from the backend for update count > %d", w.updateCount)
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
		log.Debug().Msgf("workloads v2 response -> prev highest update: %d, highest update: %d, highest system update: %d, size: %d", prevHighestUpdate, highestUpdate, highestSysUpdate, numWorkloads)

		syncReport.numWorkloads = numWorkloads

		if numWorkloads == 0 {
			log.Debug().Msgf("No v2 workloads to handle for update count > %d", w.updateCount)
			break
		}

		syncReport.theReport.Updates = make([]workloadsManagerApi.ClusterSyncReportUpdate, numWorkloads)

		innerErr = w.ProcessWorkloadConcurrently(numWorkloads, numWorkers, backendWorkloads)
		if innerErr != nil {
			log.Error().Msgf("Had an error while processing workloads, error: %s", innerErr.Error())
			err = multierror.Append(err, innerErr)
		}

		// update the highest update counter
		// NOTE! we're updating the sync counter even if there was an error in processing thr workloads
		// because there is no point to attempting again.
		innerErr = w.setSyncCounter(highestUpdate)
		if innerErr != nil {
			log.Error().Msgf("Had an error while updating workloads sync counter, error: %s", innerErr.Error())
			err = multierror.Append(err, innerErr)
		}

		// if the highest update counter is the same as the highest system update counter, it means that there are no more cluster workloads to handle
		if highestUpdate == highestSysUpdate {
			log.Debug().Msgf("No more workloads to handle, take a rest")
			break
		}
	}

	if err != nil {
		log.Error().Msgf("Failed to handle some of cluster workloads v2 (create or update), error: %s", err.Error())
		return err
	}

	log.Debug().Msgf("Finished handling v2 workloads")
	return nil
}

func (w *WorkloadsV2Handler) getBackendWorkloads(prevHighestUpdate int64) (*workloadsManagerApi.WorkloadsV2Updates, error) {
	var syncResponse *workloadsManagerApi.WorkloadsV2Updates
	var res *http.Response
	var err error

	request := w.backendClient.WorkloadsMgrClient.WorkloadsV2API.GetWorkloadV2Updates(w.backendClient.Context())
	request = request.ClusterId(config.Config().ClusterUuid)
	request = request.UpdatesAfter(prevHighestUpdate)

	if syncReport.numWorkloads > 0 {
		request = request.ClusterSyncReport(syncReport.theReport)
	}

	syncResponse, res, err = request.Execute()
	if err != nil {
		log.Error().Msgf("Had an error while preparing v2 get request: %s", err.Error())
		return nil, err
	}

	err = w.backendClient.HandleHttpResponse(res, err, "get workloads v2 updates")
	if err != nil {
		log.Error().Msgf("Had an error while getting v2 workloads: %s", err.Error())
		return nil, err
	}

	return syncResponse, nil
}

func (w *WorkloadsV2Handler) ProcessWorkloadConcurrently(numWorkloads int, numWorkers int, workloads *workloadsManagerApi.WorkloadsV2Updates) error {
	var wg sync.WaitGroup
	var err error

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
			perWorkloadError, innerErr := w.ProcessWorkloads(workloads.Workloads[start:end], &wg)
			if innerErr != nil {
				err = multierror.Append(err, innerErr)
			}

			for idx := start; idx < end; idx++ {
				syncReport.theReport.Updates[idx] = workloadsManagerApi.ClusterSyncReportUpdate{
					Id:                convutils.Ref(workloads.Workloads[idx].Id),
					UpdateCount:       convutils.Ref(workloads.Workloads[idx].UpdateCount),
					UpdateAttemptTime: convutils.Ref(time.Now()),
				}
				if perWorkloadError[idx-start] != nil {
					syncReport.theReport.Updates[idx].Message = *workloadsManagerApi.NewNullableString(convutils.Ref(perWorkloadError[idx-start].Error()))
				}
			}
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
	return err
}

func (w *WorkloadsV2Handler) ProcessWorkloads(workloads []workloadsManagerApi.WorkloadV2Update, wg *sync.WaitGroup) ([]error, error) {
	defer wg.Done()
	var err error

	clusterUpdates, allObjectsErrors, innerErr := w.createClusterUpdates(workloads)
	if innerErr != nil {
		log.Error().Msgf("Had an error while convering workloads: %s", innerErr.Error())
		err = multierror.Append(err, innerErr)
	}

	metrics.GetMetricsExporter().AddResourceAction(w.MetricLabel(), metrics.ActionCreateOrUpdate, len(clusterUpdates.ToUpdate))
	metrics.GetMetricsExporter().AddResourceAction(w.MetricLabel(), metrics.ActionDelete, len(clusterUpdates.ToDelete))

	if len(clusterUpdates.ToUpdate) > 0 {
		_, perObjectError, innerErr := w.CreateOrUpdateObjects(clusterUpdates.ToUpdate)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}

		for idx, objectErr := range perObjectError {
			toIndex := clusterUpdates.UpdateRef[idx]
			allObjectsErrors[toIndex] = objectErr
		}
	}

	if len(clusterUpdates.ToDelete) > 0 {
		perObjectError, innerErr := w.DeleteAnyWorkloadObjects(clusterUpdates.ToDelete, clusterUpdates.ToDeleteGvk)
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

func (w *WorkloadsV2Handler) DeleteAnyWorkloadObjects(objectNames []types.NamespacedName, objectsGVKs []schema.GroupVersionKind) ([]error, error) {
	var err error
	perObjectError := make([]error, len(objectNames))

	for idx, objectName := range objectNames {
		object, innerErr := w.getWorkloadObject(objectName, objectsGVKs[idx])

		if innerErr != nil {
			if !isWorkloadDeletionComplete(innerErr) {
				log.Error().Msgf("Failed to get %s <%s/%s>, error: %s", w.objectTypeStr(), objectName.Namespace, objectName.Name, innerErr.Error())
				err = multierror.Append(err, innerErr)
				perObjectError[idx] = multierror.Append(perObjectError[idx], innerErr)
			} else {
				log.Debug().Msgf("%s <%s/%s> is already deleted from the cluster - doing nothing",
					w.objectTypeStr(), objectName.Namespace, objectName.Name)
				perObjectError[idx] = errors.NewNotFound(
					schema.GroupResource{Group: objectsGVKs[idx].Group, Resource: objectsGVKs[idx].Kind},
					objectName.Name,
				)
			}
			continue
		}

		innerErr = w.deleteObject(object)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			perObjectError[idx] = multierror.Append(perObjectError[idx], innerErr)
			w.createDeletionEvent(object, innerErr)
		}
	}

	return perObjectError, err
}

func (w *WorkloadsV2Handler) getWorkloadObject(objectName types.NamespacedName, gvk schema.GroupVersionKind) (client.Object, error) {
	// first we check for runner workload
	runner := &dromosv1alpha1.WorkloadRunner{}
	err := w.cachedClient.Get(context.Background(), objectName, runner)
	if err != nil && errors.IsNotFound(err) {
		// else check if external workload
		unstructuredObj := &unstructured.Unstructured{}
		unstructuredObj.SetGroupVersionKind(gvk)
		err = w.cachedClient.Get(context.Background(), objectName, unstructuredObj)
		return unstructuredObj, err
	}
	return runner, err
}

func (w *WorkloadsV2Handler) setSyncCounter(highestUpdate int64) error {
	w.updateCount = highestUpdate

	err := configmaps.UpdateWorkloadsV2SyncCounter(w.runaiClient, w.cachedClient, highestUpdate)
	if err != nil {
		return err
	}
	log.Info().Msgf("Updated workloads v2 sync counter to: %d", highestUpdate)
	return nil
}

func (w *WorkloadsV2Handler) createClusterUpdates(backendUpdates []workloadsManagerApi.WorkloadV2Update) (clusterWorkloadsAndPoliciesUpdates, []error, error) {
	clusterUpdates := clusterWorkloadsAndPoliciesUpdates{
		UpdateRef: make([]int, 0, len(backendUpdates)),
		ToUpdate:  make([]client.Object, 0, len(backendUpdates)),

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

	errorList := make([]error, 0, len(backendUpdates))
	for idx, backendUpdate := range backendUpdates {
		namespace := projectToNamespaceMap[convutils.Deref(backendUpdate.ProjectId, "")]
		if namespace == "" {
			log.Error().Msgf("Unknown project %s, skipped", convutils.Deref(backendUpdate.ProjectId, ""))
		}

		if backendUpdate.DesiredPhase == workloadsManagerApi.DESIREDPHASE_DELETED {
			clusterUpdates.ToDelete = append(clusterUpdates.ToDelete, types.NamespacedName{
				Namespace: namespace,
				Name:      backendUpdate.Name,
			})
			clusterUpdates.ToDeleteGvk = append(clusterUpdates.ToDeleteGvk, schema.GroupVersionKind{
				Group:   convutils.Deref(backendUpdate.Gvk.Group.Get(), ""),
				Version: backendUpdate.Gvk.Version,
				Kind:    backendUpdate.Gvk.Kind,
			})
			clusterUpdates.DeleteRef = append(clusterUpdates.DeleteRef, idx)
			continue
		}

		clusterWorkload, err := w.convertToClusterResource(backendUpdate, namespace)
		if err != nil {
			errorList = append(errorList, err)
			perWorkloadErrors[idx] = err
			continue
		}

		clusterUpdates.ToUpdate = append(clusterUpdates.ToUpdate, clusterWorkload)
		clusterUpdates.UpdateRef = append(clusterUpdates.UpdateRef, idx)
	}

	return clusterUpdates, perWorkloadErrors, goErrors.Join(errorList...)
}

func (w *WorkloadsV2Handler) convertToClusterResource(update workloadsManagerApi.WorkloadV2Update, namespace string) (client.Object, error) {
	obj := client.Object(&unstructured.Unstructured{
		Object: update.WorkloadManifest,
	})
	workloadId := update.Id
	runner, err := wrapWithWorkloadRunner(obj, workloadId, update.Configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to build workload runner: %w", err)
	}
	runnerUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(runner)
	if err != nil {
		return nil, fmt.Errorf("failed to convert workload runner to unstructured: %w", err)
	}
	obj = client.Object(&unstructured.Unstructured{
		Object: runnerUnstructured,
	})
	obj.SetName(update.Name)
	obj.SetNamespace(namespace)

	annotations := obj.GetAnnotations()
	labels := obj.GetLabels()

	if obj.GetAnnotations() == nil {
		annotations = make(map[string]string)
	}
	if obj.GetLabels() == nil {
		labels = make(map[string]string)
	}

	// Set default annotations
	annotations[clusterCommonConstants.AnnotationsBackendWorkloadCounter] = fmt.Sprintf("%d", update.UpdateCount)
	annotations[constants.ControlPlaneWorkloadIdAnnotationForWorkload] = update.Id
	annotations[annotationClusterId] = config.Config().ClusterUuid
	if update.CreatedBy != "" {
		annotations[clusterCommonConstants.WorkloadOwnerUIDAnnotation] = update.CreatedBy
	}

	// Set default labels
	labels[composer.ControlPlaneManagedLabel] = "true"
	labels[clusterCommonConstants.WorkloadLabel] = "true"

	obj.SetAnnotations(annotations)
	obj.SetLabels(labels)

	return obj, nil
}

type UnstructuredObjectInspector struct {
}

// DeepCopySpec
//
//nolint:staticcheck
func (w UnstructuredObjectInspector) DeepCopySpec(from, to client.Object) {
	fromUnstructured, ok := from.(*unstructured.Unstructured)
	if !ok {
		return
	}
	toUnstructured, ok := to.(*unstructured.Unstructured)
	if !ok {
		return
	}

	// Snapshot the in-cluster restart state before the overwrite below wipes it.
	// 'from' is the CP payload and is reused across update retries, so it must never
	// be mutated here - state preserved on one attempt would leak into the next.
	restartState := snapshotRestartState(toUnstructured)

	// Only copy the "spec" field, preserving metadata on the target object
	fromSpec, found, err := unstructured.NestedFieldCopy(fromUnstructured.Object, "spec")
	if err != nil || !found {
		return
	}
	if err := unstructured.SetNestedField(toUnstructured.Object, fromSpec, "spec"); err != nil {
		return
	}

	// annotations
	if to.GetAnnotations() == nil {
		to.SetAnnotations(map[string]string{})
	}
	to.SetAnnotations(from.GetAnnotations())

	// labels - copy labels from source to target (important for category and priority updates)
	if to.GetLabels() == nil {
		to.SetLabels(map[string]string{})
	}
	to.SetLabels(from.GetLabels())

	// Re-apply the restart state on top of the freshly copied payload.
	restartState.applyTo(toUnstructured)
}

// wlrRestartState is the restart bookkeeping the restart controller writes on a live
// WorkloadRunner. It must survive the CP sync overwrite until handback clears it.
type wlrRestartState struct {
	requestRef   string
	managedBy    string
	nodeAffinity interface{}
}

// snapshotRestartState captures the restart state of a live WorkloadRunner.
// Returns the zero value (which applyTo treats as a no-op) for any other kind.
func snapshotRestartState(live *unstructured.Unstructured) wlrRestartState {
	if live.GetKind() != KindWorkloadRunner {
		return wlrRestartState{}
	}

	state := wlrRestartState{requestRef: live.GetLabels()[requestRefLabel]}

	if managedBy, found, err := unstructured.NestedString(live.Object, "spec", "managedBy"); err == nil && found {
		state.managedBy = managedBy
	}
	if nodeAffinity, found, err := unstructured.NestedFieldCopy(live.Object, "spec", "instructions", "nodeAffinity"); err == nil && found {
		state.nodeAffinity = nodeAffinity
	}
	return state
}

// applyTo writes the snapshotted restart state back onto the live object. Empty
// fields are skipped so a CP-provided value is not clobbered by an unset one.
func (s wlrRestartState) applyTo(live *unstructured.Unstructured) {
	if s.requestRef != "" {
		labels := live.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[requestRefLabel] = s.requestRef
		live.SetLabels(labels)
	}

	if s.managedBy != "" {
		if err := unstructured.SetNestedField(live.Object, s.managedBy, "spec", "managedBy"); err != nil {
			log.Warn().Msgf("failed to preserve spec.managedBy on WorkloadRunner <%s/%s>: %s",
				live.GetNamespace(), live.GetName(), err.Error())
		}
	}

	if s.nodeAffinity != nil {
		if err := unstructured.SetNestedField(live.Object, s.nodeAffinity, "spec", "instructions", "nodeAffinity"); err != nil {
			log.Warn().Msgf("failed to preserve spec.instructions.nodeAffinity on WorkloadRunner <%s/%s>: %s",
				live.GetNamespace(), live.GetName(), err.Error())
		}
	}
}

func (w UnstructuredObjectInspector) IsUpToDate(left, right client.Object) bool {
	if left == nil || right == nil {
		return false
	}

	rightAnnotations := right.GetAnnotations()

	counterStr, ok := rightAnnotations[clusterCommonConstants.AnnotationsBackendWorkloadCounter]
	if !ok || counterStr == "" {
		return false
	}

	rightAnnotationsCounter, err := strconv.Atoi(counterStr)
	if err != nil {
		log.Warn().Msgf("Had an error while converting backend workload counter to int object annotation: %s , err: %s", rightAnnotations[clusterCommonConstants.AnnotationsBackendWorkloadCounter], err.Error())
		return false
	}

	leftAnnotations := left.GetAnnotations()

	counterStr, ok = leftAnnotations[clusterCommonConstants.AnnotationsBackendWorkloadCounter]
	if !ok || counterStr == "" {
		return false
	}

	leftAnnotationsCounter, err := strconv.Atoi(counterStr)
	if err != nil {
		log.Warn().Msgf("Had an error while converting backend workload counter to int object annotation: %s , err: %s", leftAnnotations[clusterCommonConstants.AnnotationsBackendWorkloadCounter], err.Error())
		return false
	}

	if leftAnnotationsCounter <= rightAnnotationsCounter {
		// if the backend workload counter is the same or higher than the cluster workload counter,
		// it means that the backend workload is newer than the cluster workload
		return true
	}

	return false
}

func wrapWithWorkloadRunner(manifest client.Object, workloadId string, configuration *workloadsManagerApi.WorkloadV2UpdateConfiguration) (*dromosv1alpha1.WorkloadRunner, error) {
	manifestAnnotations := manifest.GetAnnotations()
	if manifestAnnotations == nil {
		manifestAnnotations = make(map[string]string)
	}
	manifestAnnotations[constants.ControlPlaneWorkloadIdAnnotationForWorkload] = workloadId
	manifest.SetAnnotations(manifestAnnotations)

	runner := &dromosv1alpha1.WorkloadRunner{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindWorkloadRunner,
			APIVersion: ApiVersionWorkloadRunner,
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template: runtime.RawExtension{Object: manifest},
		},
	}
	if err := applyNetworkingSubmissionProfile(manifest, runner, configuration); err != nil {
		return nil, fmt.Errorf("invalid networking submission profile: %w", err)
	}

	if configuration == nil {
		return runner, nil
	}
	if err := applyInferenceEndpoint(runner, configuration.InferenceEndpoint.Get()); err != nil {
		return nil, fmt.Errorf("invalid inference endpoint configuration: %w", err)
	}

	hasMnnvl := configuration.Mnnvl != nil && configuration.Mnnvl.Mode != nil
	hasNetworkingEndpoints := hasNetworkingEndpointsConfiguration(configuration)
	hasLegacyNetworking := !hasNetworkingEndpoints && hasLegacyNetworkingConfiguration(configuration)
	hasStorage := len(configuration.Pvcs) > 0

	if hasMnnvl || hasLegacyNetworking || hasStorage {
		runner.Spec.Instructions = &dromosv1alpha1.Instructions{}
	}

	if hasMnnvl {
		mode := dromosv1alpha1.MnnvlModeNone
		if *configuration.Mnnvl.Mode == workloadsManagerApi.MNNVLMODE_REQUIRED {
			mode = dromosv1alpha1.MnnvlModeRequired
		}

		runner.Spec.Instructions.GpuOptimizations = &dromosv1alpha1.GpuOptimizations{
			Mnnvl: &dromosv1alpha1.MnnvlConfig{
				Mode:                            mode,
				GpuNetworkAccelerationLabelKeys: configuration.Mnnvl.GpuNetworkAccelerationLabelKeys,
			},
		}
	}

	if hasNetworkingEndpoints {
		netConfig, err := convertNetworkingEndpoints(configuration.NetworkingEndpoints, workloadconfig.IsOpenshift())
		if err != nil {
			return nil, fmt.Errorf("failed to convert networking endpoints: %w", err)
		}
		if err := endpoints.Set(runner, netConfig); err != nil {
			return nil, fmt.Errorf("failed to set networking endpoints annotation: %w", err)
		}
	} else if hasLegacyNetworking {
		netConfig, err := convertWorkloadV2NetworkingConfig(configuration, workloadconfig.IsOpenshift())
		if err != nil {
			return nil, fmt.Errorf("failed to convert networking configuration: %w", err)
		}
		runner.Spec.Instructions.Networking = netConfig
	}

	if hasStorage {
		storageConfig, err := convertStorageConfig(configuration.Pvcs, configuration.PvcsConfig.Get())
		if err != nil {
			return nil, fmt.Errorf("failed to convert storage configuration: %w", err)
		}
		runner.Spec.Instructions.Storage = storageConfig
	}

	if convutils.HasValue(configuration.UserCredentials) {
		pc := configuration.UserCredentials.Get()
		if len(pc.Sources) > 0 {
			creds := make([]pcreds.Cred, 0, len(pc.Sources))
			for _, s := range pc.Sources {
				creds = append(creds, pcreds.Cred{SourceSecret: s.Ref})
			}
			if err := pcreds.Set(runner, creds); err != nil {
				return nil, fmt.Errorf("failed to set personal-credentials annotation: %w", err)
			}
		}
	}

	return runner, nil
}

func hasNetworkingEndpointsConfiguration(configuration *workloadsManagerApi.WorkloadV2UpdateConfiguration) bool {
	if configuration == nil {
		return false
	}
	return len(configuration.NetworkingEndpoints) > 0
}

func hasLegacyNetworkingConfiguration(configuration *workloadsManagerApi.WorkloadV2UpdateConfiguration) bool {
	if configuration == nil {
		return false
	}
	legacyNetworking := configuration.Networking.Get()
	return legacyNetworking != nil && (len(legacyNetworking.Services) > 0 || legacyNetworking.ExternalAccess.IsSet())
}
