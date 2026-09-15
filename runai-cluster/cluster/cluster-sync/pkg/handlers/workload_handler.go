/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	corev1 "k8s.io/api/core/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dinfwv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	"go.uber.org/zap"
	"k8s.io/utils/ptr"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkloadHandler struct {
	resourceName                constants.ClusterResource
	cachedClient                client.Client
	backendClient               *runaiBackendClient.BackendClient
	parser                      *parsers.WorkloadParser
	maxNumOfWorkloadsPerRequest int
	syncPodsWithWorkloads       bool

	// namespace -> UIDs of already synced workloads
	syncedObjects      map[string]map[types.UID]bool
	syncedObjectsMutex sync.RWMutex

	podHandler     *PodHandler
	podRateLimiter *cachedclient.RateLimitedController
}

type WorkloadsByTypes struct {
	Training             map[string]*twv2a1.TrainingWorkload
	Interactive          map[string]*iwv2a1.InteractiveWorkload
	Inference            map[string]*infwv2a1.InferenceWorkload
	Distributed          map[string]*diwv2.DistributedWorkload
	DistributedInference map[string]*dinfwv1.DistributedInferenceWorkload
	External             map[string]*v1alpha1.ExternalWorkload
	ids                  []string
}

// isLegacyWorkloadUpToDate compares against the last status written before the
// legacy APIs became read-only. The temporary conversion keeps that status
// immutable while reusing the existing workload checksum calculation.
func isLegacyWorkloadUpToDate(syncStatus *common.WorkloadSyncStatus, generation int64,
	lastStatusChangeTime *metav1.Time, parsedWorkload *openapi.WorkloadSyncFields) bool {
	if syncStatus == nil {
		return false
	}

	statusCopy := &v1alpha1.WorkloadSyncStatus{
		LastSyncGeneration:       syncStatus.LastSyncGeneration,
		LastSyncStatusChangeTime: syncStatus.LastSyncStatusChangeTime,
		LastSyncStatusChecksum:   syncStatus.LastSyncStatusChecksum,
	}
	return isExternalWorkloadUpToDateAndSetSyncStatus(
		statusCopy, generation, lastStatusChangeTime, parsedWorkload,
	)
}

// GetParsedWorkloadsFn lists all workloads in the namespace; adds their UIDs to syncedInNamespace.
// Returns: the parsed objects that need to be synced;
// a list of the IDs of the objects that are up-to-date (so no need to send full parsed object to server);
// a list of workloads that need their sync status to be updated;
// error - if occurred.
type GetParsedWorkloadsFn func(namespaceName string, syncedInNamespace map[types.UID]bool) (
	[]openapi.WorkloadSyncFields, []string, []client.Object, error)

func NewWorkloadHandler(runaiBackendClient *runaiBackendClient.BackendClient,
	cachedClient client.Client, podHandler *PodHandler, podRateLimiter *cachedclient.RateLimitedController) *WorkloadHandler {
	return &WorkloadHandler{
		resourceName:                constants.DistributedWorkloadResourceName,
		cachedClient:                cachedClient,
		backendClient:               runaiBackendClient,
		parser:                      parsers.NewWorkloadParser(cachedClient),
		maxNumOfWorkloadsPerRequest: config.Config().MaxNumOfWorkloadsPerRequest,
		syncPodsWithWorkloads:       config.Config().SyncPodsWithWorkloads,
		syncedObjects:               map[string]map[types.UID]bool{},
		syncedObjectsMutex:          sync.RWMutex{},
		podHandler:                  podHandler,
		podRateLimiter:              podRateLimiter,
	}
}

func (h *WorkloadHandler) HandleRunaiWorkloadDelete(object client.Object) error {
	kind := ""
	switch object.(type) {
	case *twv2a1.TrainingWorkload:
		kind = constants.TrainingWorkload
	case *iwv2a1.InteractiveWorkload:
		kind = constants.InteractiveWorkload
	case *infwv2a1.InferenceWorkload:
		kind = constants.InferenceWorkload
	case *diwv2.DistributedWorkload:
		kind = constants.DistributedWorkload
	case *dinfwv1.DistributedInferenceWorkload:
		kind = constants.DistributedInferenceWorkload
	}

	zap.S().Debugf("Handling %s <%s/%s> delete", kind, object.GetNamespace(), object.GetName())

	workloadId := topowner.GetRunaiWorkloadId(object.GetUID(), object.GetAnnotations())
	return h.handleDelete(workloadId, object.GetNamespace(), object.GetName(), kind)
}

func (h *WorkloadHandler) handleDelete(uid string, namespace, name, kind string) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.DeleteOperation)()
	request := h.backendClient.WorkloadsClient.WorkloadsAPI.DeleteWorkload(h.backendClient.Context(), uid)

	res, err := request.Execute()
	if res != nil {
		_ = res.Body.Close()
	}
	return h.backendClient.HandleHttpResponse(res, err,
		fmt.Sprintf("Deleted %s <%s/%s>", kind, namespace, name))
}

func (h *WorkloadHandler) WasSynced(objectUID types.UID, objectNamespace string) bool {
	return WasSynced(objectUID, objectNamespace, h.syncedObjects, &h.syncedObjectsMutex)
}

func (h *WorkloadHandler) handleUpsert(workloadSyncFields *openapi.WorkloadSyncFields,
	namespace, name, kind string) error {
	{
		defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.UpsertOperation)()
		workloadId := workloadSyncFields.Id
		request := h.backendClient.WorkloadsClient.WorkloadsAPI.UpsertWorkload(h.backendClient.Context(), workloadId)
		request = request.WorkloadSyncFields(*workloadSyncFields)

		res, err := request.Execute()
		if res != nil {
			defer func() { _ = res.Body.Close() }()
		}
		logDetailedApiErrorMessageIfNeeded(res, err, "upsert workload")
		finalError := h.backendClient.HandleHttpResponse(res, err,
			fmt.Sprintf("Upserted %s <%s/%s>", kind, namespace, name))
		if finalError != nil {
			return finalError
		}
	}

	_ = h.podHandler.HandleAdditionOfWorkloadPodsAfterItWasAdded(workloadSyncFields.Id, h.podRateLimiter)
	h.podHandler.ResyncPodsAfterWorkloadAdded(workloadSyncFields.ProjectName)
	return nil
}

func (h *WorkloadHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start ExternalWorkload Sync")

	if h.syncPodsWithWorkloads {
		UpdateProjectHandlerWrapper(h.handleUpdateInProjectWorkloadPagination, h.cachedClient, "ExternalWorkloads")
	} else {
		UpdateProjectHandlerWrapper(h.handleUpdateInProject, h.cachedClient, "ExternalWorkloads")
	}
}

func (h *WorkloadHandler) getClusterWorkloadsByProject(project *util.ProjectData) (*WorkloadsByTypes, error) {
	namespaceName := project.Namespace
	if namespaceName == "" {
		zap.S().Warnf("Failed getting namespace name for project <%s> while syncing ExternalWorkloads",
			project.Name)
		return nil, nil
	}

	fieldSelector := GetRelevantRunaiWorkloadsFieldSelector()

	workloadIds := []string{}
	trainingWorkloads := &twv2a1.TrainingWorkloadList{}
	trainingWorkloadsMap := map[string]*twv2a1.TrainingWorkload{}
	err := h.cachedClient.List(context.Background(), trainingWorkloads, &client.ListOptions{
		Namespace:     namespaceName,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	for i := range trainingWorkloads.Items {
		workload := trainingWorkloads.Items[i]
		workloadId := topowner.GetRunaiWorkloadId(workload.UID, workload.Annotations)
		trainingWorkloadsMap[workloadId] = &(workload)
		workloadIds = append(workloadIds, workloadId)
	}

	interactiveWorkloads := &iwv2a1.InteractiveWorkloadList{}
	interactiveWorkloadsMap := map[string]*iwv2a1.InteractiveWorkload{}
	err = h.cachedClient.List(context.Background(), interactiveWorkloads, &client.ListOptions{
		Namespace:     namespaceName,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	for i := range interactiveWorkloads.Items {
		workload := interactiveWorkloads.Items[i]
		workloadId := topowner.GetRunaiWorkloadId(workload.UID, workload.Annotations)
		interactiveWorkloadsMap[workloadId] = &(workload)
		workloadIds = append(workloadIds, workloadId)
	}

	inferenceWorkloads := &infwv2a1.InferenceWorkloadList{}
	inferenceWorkloadsMap := map[string]*infwv2a1.InferenceWorkload{}
	err = h.cachedClient.List(context.Background(), inferenceWorkloads, &client.ListOptions{
		Namespace:     namespaceName,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	for i := range inferenceWorkloads.Items {
		workload := inferenceWorkloads.Items[i]
		workloadId := topowner.GetRunaiWorkloadId(workload.UID, workload.Annotations)
		inferenceWorkloadsMap[workloadId] = &(workload)
		workloadIds = append(workloadIds, workloadId)
	}

	distributedWorkloads := &diwv2.DistributedWorkloadList{}
	distributedWorkloadsMap := map[string]*diwv2.DistributedWorkload{}
	err = h.cachedClient.List(context.Background(), distributedWorkloads, &client.ListOptions{
		Namespace:     namespaceName,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	for i := range distributedWorkloads.Items {
		workload := distributedWorkloads.Items[i]
		workloadId := topowner.GetRunaiWorkloadId(workload.UID, workload.Annotations)
		distributedWorkloadsMap[workloadId] = &(workload)
		workloadIds = append(workloadIds, workloadId)
	}

	distributedInferenceWorkloads := &dinfwv1.DistributedInferenceWorkloadList{}
	distributedInferenceWorkloadsMap := map[string]*dinfwv1.DistributedInferenceWorkload{}
	err = h.cachedClient.List(context.Background(), distributedInferenceWorkloads, &client.ListOptions{
		Namespace:     namespaceName,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	for i := range distributedInferenceWorkloads.Items {
		workload := distributedInferenceWorkloads.Items[i]
		workloadId := topowner.GetRunaiWorkloadId(workload.UID, workload.Annotations)
		distributedInferenceWorkloadsMap[workloadId] = &(workload)
		workloadIds = append(workloadIds, workloadId)
	}

	externalWorkloads := &v1alpha1.ExternalWorkloadList{}
	externalWorkloadsMap := map[string]*v1alpha1.ExternalWorkload{}
	err = h.cachedClient.List(context.Background(), externalWorkloads, &client.ListOptions{Namespace: namespaceName})
	if err != nil {
		return nil, err
	}
	for i := range externalWorkloads.Items {
		workload := externalWorkloads.Items[i]
		workloadId := topowner.GetRunaiWorkloadId(workload.Spec.OriginalWorkloadUID, workload.Annotations)
		externalWorkloadsMap[workloadId] = &(workload)
		workloadIds = append(workloadIds, workloadId)
	}

	sort.Strings(workloadIds)

	return &WorkloadsByTypes{
		Training:             trainingWorkloadsMap,
		Interactive:          interactiveWorkloadsMap,
		Inference:            inferenceWorkloadsMap,
		Distributed:          distributedWorkloadsMap,
		DistributedInference: distributedInferenceWorkloadsMap,
		External:             externalWorkloadsMap,
		ids:                  workloadIds,
	}, nil
}

func (h *WorkloadHandler) handleUpdateInProject(project *util.ProjectData) {
	projectId, err := util.GetProjectId(project)
	if err != nil {
		return
	}

	namespaceName := project.Namespace
	if namespaceName == "" {
		zap.S().Warnf("Failed getting namespace name for project <%s> while syncing ExternalWorkloads",
			project.Name)
		return
	}

	syncedInNamespace := map[types.UID]bool{}
	allWorkloadsToSync, allUpToDateIds, allWorkloadsToUpdateSyncStatus, err := h.getAllWorkloadsToSync(namespaceName, syncedInNamespace)
	if err != nil {
		return
	}

	var errMsg *string = nil
	err = h.handleSyncInProject(allWorkloadsToSync, allUpToDateIds, projectId, project.Name)
	if err != nil {
		zap.S().Errorf("Had an error while syncing Workloads (v2) in project <%s[%s]>, error: %s",
			project.Name, projectId, err.Error())
		syncedInNamespace = map[types.UID]bool{}
		errMsg = ptr.To(err.Error())
	}

	h.syncedObjectsMutex.Lock()
	h.syncedObjects[namespaceName] = syncedInNamespace
	h.syncedObjectsMutex.Unlock()

	go h.updateWorkloadsSyncStatus(allWorkloadsToUpdateSyncStatus, errMsg)
}

func (h *WorkloadHandler) handleUpdateInProjectWorkloadPagination(project *util.ProjectData) {
	projectId, err := util.GetProjectId(project)
	if err != nil {
		return
	}

	workloadTypes, err := h.getClusterWorkloadsByProject(project)

	if err != nil {
		return
	}
	namespaceName := project.Namespace
	if namespaceName == "" {
		zap.S().Warnf("Failed getting namespace name for project <%s> while syncing ExternalWorkloads",
			project.Name)
		return
	}

	syncedInNamespace := map[types.UID]bool{}

	err = h.handleSyncInProjectPaginated(workloadTypes, projectId, project.Name, syncedInNamespace, namespaceName)
	if err != nil {
		zap.S().Errorf("Had an error while syncing Workloads (v2) in project <%s[%s]>, error: %s",
			project.Name, projectId, err.Error())
		syncedInNamespace = map[types.UID]bool{}
	}

	h.syncedObjectsMutex.Lock()
	h.syncedObjects[namespaceName] = syncedInNamespace
	h.syncedObjectsMutex.Unlock()
}

func (h *WorkloadHandler) getAllWorkloadsToSync(namespaceName string,
	syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, error) {
	getParsedWorkloadsFns := []GetParsedWorkloadsFn{
		h.getParsedExternalWorkloadsInNamespace,
		h.getParsedTrainingWorkloadsInNamespace,
		h.getParsedInteractiveWorkloadsInNamespace,
		h.getParsedInferenceWorkloadsInNamespace,
		h.getParsedDistributedWorkloadsInNamespace,
		h.getParsedDistributedInferenceWorkloadsInNamespace}

	allWorkloadsToSync := []openapi.WorkloadSyncFields{}
	allUpToDateIds := []string{}
	allWorkloadsToUpdateSyncStatus := []client.Object{}

	for _, getParsedWorkloadsFn := range getParsedWorkloadsFns {
		workloadsToSync, upToDateIds, workloadsToUpdateSyncStatus, err := getParsedWorkloadsFn(namespaceName, syncedInNamespace)
		if err != nil {
			return []openapi.WorkloadSyncFields{}, []string{}, []client.Object{}, err
		}

		allWorkloadsToSync = append(allWorkloadsToSync, workloadsToSync...)
		allUpToDateIds = append(allUpToDateIds, upToDateIds...)
		allWorkloadsToUpdateSyncStatus = append(allWorkloadsToUpdateSyncStatus, workloadsToUpdateSyncStatus...)
	}

	return allWorkloadsToSync, allUpToDateIds, allWorkloadsToUpdateSyncStatus, nil
}

func (h *WorkloadHandler) handleSyncInProject(
	workloadsToSync []openapi.WorkloadSyncFields, upToDateIds []string,
	projectId string, projectName string) error {
	if len(workloadsToSync) <= h.maxNumOfWorkloadsPerRequest {
		return h.sendSyncWorkloadsRequests(workloadsToSync, upToDateIds, projectId, projectName, nil, nil)
	}

	// sort all workloads first
	sortWorkloadsById(workloadsToSync)

	for i := 0; i < len(workloadsToSync); i += h.maxNumOfWorkloadsPerRequest - 1 {
		chunkEndIndex := i + h.maxNumOfWorkloadsPerRequest
		if chunkEndIndex > len(workloadsToSync) {
			chunkEndIndex = len(workloadsToSync)
		}

		chunk := workloadsToSync[i:chunkEndIndex]

		var from *string = nil
		var to *string = nil
		if i > 0 {
			from = ptr.To(chunk[0].Id)
		}
		if chunkEndIndex < len(workloadsToSync) {
			to = ptr.To(chunk[len(chunk)-1].Id)
		}

		if err := h.sendSyncWorkloadsRequests(chunk, upToDateIds, projectId, projectName, from, to); err != nil {
			return err
		}
	}
	return nil
}

type IterableWorkload struct {
	ObjectMeta metav1.ObjectMeta
}

func (h *WorkloadHandler) handleSyncInProjectPaginated(workloads *WorkloadsByTypes, projectId string, projectName string, syncedInNamespace map[types.UID]bool, namespaceName string) error {

	moreThanOneChunk := len(workloads.ids) >= h.maxNumOfWorkloadsPerRequest

	for chunkIndex := 0; chunkIndex < len(workloads.ids); chunkIndex += h.maxNumOfWorkloadsPerRequest - 1 {

		allUpToDateIds := []string{}
		allWorkloadsToUpdateSyncStatus := []client.Object{}

		chunkEndIndex := chunkIndex + h.maxNumOfWorkloadsPerRequest
		if chunkEndIndex > len(workloads.ids) {
			chunkEndIndex = len(workloads.ids)
		}

		chunk := []openapi.WorkloadSyncFields{}
		podsSyncedInNamespace := map[types.UID]bool{}
		podsChunk := map[string][]corev1.Pod{}
		for idIndex := chunkIndex; idIndex < chunkEndIndex; idIndex++ {
			workloadId := workloads.ids[idIndex]
			parsedWorkload, upToDateIds, workloadsToUpdateSyncStatus, pods, err := h.getWorkload(workloadId, workloads, syncedInNamespace)
			if err != nil {
				return err
			}
			if len(parsedWorkload) > 0 {
				chunk = append(chunk, parsedWorkload[0])
			}
			allUpToDateIds = append(allUpToDateIds, upToDateIds...)
			allWorkloadsToUpdateSyncStatus = append(allWorkloadsToUpdateSyncStatus, workloadsToUpdateSyncStatus...)
			podsChunk[workloadId] = pods
		}

		var from *string = nil
		var to *string = nil
		if moreThanOneChunk {
			if chunkIndex > 0 {
				from = ptr.To(workloads.ids[chunkIndex])
			}
			if chunkEndIndex < len(workloads.ids) {
				to = ptr.To(workloads.ids[chunkEndIndex-1])
			}
		}

		var errMsg *string = nil
		if err := h.sendSyncWorkloadsRequests(chunk, allUpToDateIds, projectId, projectName, from, to); err != nil {
			zap.S().Errorf("Had an error while syncing Workloads (v2) in project <%s[%s]>, error: %s",
				projectName, projectId, err.Error())
			syncedInNamespace = map[types.UID]bool{}
			errMsg = ptr.To(err.Error())
		}
		// this is done since h.syncPodsWithWorkloads is true
		h.podHandler.syncPodsFromWorkload(projectName, podsChunk, projectId, podsSyncedInNamespace, namespaceName)

		go h.updateWorkloadsSyncStatus(allWorkloadsToUpdateSyncStatus, errMsg)
	}

	return nil
}

func (h *WorkloadHandler) getWorkload(id string, workloads *WorkloadsByTypes, syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, []corev1.Pod, error) {
	training, found := workloads.Training[id]
	if found && training != nil {
		return h.parseTrainingWorkloads([]twv2a1.TrainingWorkload{*training}, syncedInNamespace)
	}
	interactive, found := workloads.Interactive[id]
	if found && interactive != nil {
		return h.parseInteractiveWorkloads([]iwv2a1.InteractiveWorkload{*interactive}, syncedInNamespace)
	}
	distributed, found := workloads.Distributed[id]
	if found && distributed != nil {
		return h.parseDistributedWorkloads([]diwv2.DistributedWorkload{*distributed}, syncedInNamespace)
	}
	inference, found := workloads.Inference[id]
	if found && inference != nil {
		return h.ParseInferenceWorkloads([]infwv2a1.InferenceWorkload{*inference}, syncedInNamespace)
	}
	distributedInference, found := workloads.DistributedInference[id]
	if found && distributedInference != nil {
		return h.parseDistributedInferenceWorkloads([]dinfwv1.DistributedInferenceWorkload{*distributedInference}, syncedInNamespace)
	}
	external, found := workloads.External[id]
	if found && external != nil {
		return h.parseExternalWorkloads([]v1alpha1.ExternalWorkload{*external}, syncedInNamespace)
	}
	return []openapi.WorkloadSyncFields{}, []string{}, []client.Object{}, []corev1.Pod{}, fmt.Errorf("failed to find workload with id %s", id)
}

// upToDateIds - should be all up-to-date ids for each request, even with pagination.
func (h *WorkloadHandler) sendSyncWorkloadsRequests(
	workloadsToSync []openapi.WorkloadSyncFields, upToDateIds []string,
	projectId string, projectName string, from, to *string) error {
	msg := fmt.Sprintf("Sync Workloads (v2) in project <%s[%s]>", projectName, projectId)

	if from != nil && *from != "" {
		msg = fmt.Sprintf("%s; from <%s>", msg, *from)
	}
	if to != nil && *to != "" {
		msg = fmt.Sprintf("%s; to <%s>", msg, *to)
	} else {
		msg = fmt.Sprintf("Finished %s", msg)
	}

	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.SyncOperation)()
	request := h.backendClient.WorkloadsClient.WorkloadsAPI.SyncWorkloads(h.backendClient.Context())
	request = request.WorkloadsSyncRequest(openapi.WorkloadsSyncRequest{
		ProjectId:   projectId,
		ClusterId:   ptr.To(config.Config().ClusterUuid),
		Workloads:   workloadsToSync,
		From:        from,
		To:          to,
		IdsToIgnore: upToDateIds,
	})

	res, err := request.Execute()
	if res != nil {
		defer func() { _ = res.Body.Close() }()
	}
	logDetailedApiErrorMessageIfNeeded(res, err, msg)
	statusCode, err := h.backendClient.HandleHttpResponseGetStatus(res, err, msg)
	if statusCode == http.StatusRequestEntityTooLarge && h.maxNumOfWorkloadsPerRequest > 2 {
		// reduce by 20%
		h.maxNumOfWorkloadsPerRequest = int(h.maxNumOfWorkloadsPerRequest * 4 / 5)
		zap.S().Warnf("Changing maxNumOfWorkloadsPerRequest to: %d", h.maxNumOfWorkloadsPerRequest)
	}
	if err == nil {
		syncedWorkloadIds := []string{}
		for _, w := range workloadsToSync {
			syncedWorkloadIds = append(syncedWorkloadIds, w.Id)
		}
		zap.S().Debugf("%s with ids: %+v // upToDateIds: %+v", msg, syncedWorkloadIds, upToDateIds)
	}

	return err
}

func (h *WorkloadHandler) updateWorkloadsSyncStatus(workloadsToUpdateSyncStatus []client.Object, errMsg *string) {
	for _, obj := range workloadsToUpdateSyncStatus {
		switch workload := obj.(type) {
		case *v1alpha1.ExternalWorkload:
			_ = h.updateExternalWorkloadSyncStatus(workload, errMsg)
		}
	}
}

func sortWorkloadsById(workloads []openapi.WorkloadSyncFields) {
	sort.Slice(workloads, func(i, j int) bool {
		return workloads[i].Id < workloads[j].Id
	})
}

func timeDeref(t *metav1.Time) metav1.Time {
	if t == nil {
		return metav1.Time{}
	}
	return *t
}

func logDetailedApiErrorMessageIfNeeded(res *http.Response, err error, action string) {
	if err != nil && res != nil && res.StatusCode >= 400 {
		// Read the full response body
		bodyBytes, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			zap.S().Errorf("Failed to read response body: %v", readErr)
		} else {
			// Restore the body for potential future reads
			res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Try to parse as API error first
			apiError := &openapi.Error{}
			if jsonErr := json.Unmarshal(bodyBytes, apiError); jsonErr == nil {
				zap.S().Errorf("Failed to %s, status: %s, error message: %s, full response body: %s",
					action, res.Status, apiError.Message, string(bodyBytes))
			} else {
				// If not a valid API error, just log the raw body
				zap.S().Errorf("Failed to %s, status: %s, full response body: %s",
					action, res.Status, string(bodyBytes))
			}
		}
	}
}
