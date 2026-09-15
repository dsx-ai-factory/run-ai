/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodHandler struct {
	resourceName           constants.ClusterResource
	backendClient          *runaiBackendClient.BackendClient
	cachedClient           cachedclient.CachedClient
	maxNumOfPodsPerRequest int

	podParser *parsers.PodParser

	// namespace -> UIDs of already synced pods
	syncedObjects      map[string]map[types.UID]bool
	syncedObjectsMutex sync.RWMutex

	// handle pods after their workload is handled.
	// map of top owner workload uuid -> pods by key
	enqueuePodsAddition      map[string]map[client.ObjectKey]*coreV1.Pod
	enqueuePodsAdditionMutex sync.RWMutex
	syncFailedProjects       map[string]bool
	syncFailedProjectsMutex  sync.RWMutex

	syncPodsWithWorkloads bool
}

func NewPodHandler(runaiBackendClient *runaiBackendClient.BackendClient, cachedClient cachedclient.CachedClient) *PodHandler {
	return &PodHandler{
		resourceName:             constants.PodResourceName,
		cachedClient:             cachedClient,
		backendClient:            runaiBackendClient,
		syncedObjects:            map[string]map[types.UID]bool{},
		syncedObjectsMutex:       sync.RWMutex{},
		podParser:                parsers.NewPodParser(cachedClient),
		maxNumOfPodsPerRequest:   config.Config().MaxNumOfPodsPerRequest,
		enqueuePodsAddition:      map[string]map[client.ObjectKey]*coreV1.Pod{},
		enqueuePodsAdditionMutex: sync.RWMutex{},
		syncFailedProjects:       map[string]bool{},
		syncFailedProjectsMutex:  sync.RWMutex{},
		syncPodsWithWorkloads:    config.Config().SyncPodsWithWorkloads,
	}
}

func (h *PodHandler) HandleUpdateItem(key client.ObjectKey) error {
	pod := &coreV1.Pod{}
	err := h.cachedClient.Get(context.Background(), key, pod)
	if err != nil {
		return err
	}

	return h.HandlePodUpdate(pod)
}

func (h *PodHandler) HandlePodAdd(pod *coreV1.Pod) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.AddOperation)()
	if WasSynced(pod.UID, pod.Namespace, h.syncedObjects, &h.syncedObjectsMutex) {
		zap.S().Infof("Pod <%s/%s> was already synced - not adding", pod.Namespace, pod.Name)
		return nil
	}

	return h.handleAddUpdate(pod)
}

func (h *PodHandler) HandlePodDelete(pod *coreV1.Pod) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.DeleteOperation)()
	request := h.backendClient.WorkloadsClient.PodsAPI.DeletePod(h.backendClient.Context(), string(pod.UID))

	res, err := request.Execute()
	if res != nil {
		_ = res.Body.Close()
	}
	return h.backendClient.HandleHttpResponse(res, err,
		fmt.Sprintf("Deleted pod <%s/%s>", pod.Namespace, pod.Name))
}

func (h *PodHandler) HandlePodUpdate(pod *coreV1.Pod) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.UpdateOperation)()
	return h.handleAddUpdate(pod)
}

func (h *PodHandler) handleAddUpdate(pod *coreV1.Pod) error {
	podSyncFields, err := h.podParser.ParsePodToServerScheme(pod)
	if err != nil {
		//TODO: remove err if not used...
		return err
	}

	request := h.backendClient.WorkloadsClient.PodsAPI.UpsertPod(h.backendClient.Context(), podSyncFields.Id)
	request = request.PodSyncFields(*podSyncFields)

	res, err := request.Execute()
	if res != nil {
		defer func() { _ = res.Body.Close() }()
	}
	logDetailedApiErrorMessageIfNeeded(res, err, "upsert pod")
	if err != nil && res != nil && res.StatusCode == http.StatusPreconditionFailed {
		h.enqueuePodsHandlingForLaterIfNeeded(res, err, pod, podSyncFields.WorkloadId)
		return nil
	}
	return h.backendClient.HandleHttpResponse(res, err,
		fmt.Sprintf("Upserted pod <%s/%s>", pod.Namespace, pod.Name))
}

func (h *PodHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	if h.syncPodsWithWorkloads {
		zap.S().Infof("Pods sync skipped - will sync with workloads")
		return
	}
	defer waitGroup.Done()
	zap.S().Infof("Start sync pods")

	UpdateProjectHandlerWrapper(h.handleUpdateInProject, h.cachedClient, "Pods")

	// try again for the projects that failed sync
	currentSyncFailedProjects := []string{}
	h.syncFailedProjectsMutex.RLock()
	for projectName := range h.syncFailedProjects {
		currentSyncFailedProjects = append(currentSyncFailedProjects, projectName)
	}
	// unlock read while handling
	h.syncFailedProjectsMutex.RUnlock()

	for _, projectName := range currentSyncFailedProjects {
		h.reSyncPodsInProject(projectName)
	}
}

func (h *PodHandler) handleUpdateInProject(project *util.ProjectData) {
	projectId, err := util.GetProjectId(project)
	if err != nil {
		return
	}

	namespaceName := project.Namespace
	if namespaceName == "" {
		zap.S().Warnf("Failed getting namespace name for project <%s> while syncing Pods",
			project.Name)
		return
	}

	syncedInNamespace := map[types.UID]bool{}
	podsToSync, err := h.getPodsToSync(namespaceName, syncedInNamespace)
	if err != nil {
		return
	}

	err = h.handleSyncInProjectWithPagination(podsToSync, projectId, project.Name)
	if err != nil {
		zap.S().Errorf("Had an error while syncing Pods in project <%s[%s]>: %s",
			project.Name, projectId, err.Error())
		syncedInNamespace = map[types.UID]bool{}
	}

	h.syncedObjectsMutex.Lock()
	h.syncedObjects[namespaceName] = syncedInNamespace
	h.syncedObjectsMutex.Unlock()

	h.updateSyncFailedProjects(project.Name, err)
}

func (h *PodHandler) syncPodsFromWorkload(projectName string, podsPerWorkload map[string][]coreV1.Pod, projectId string, syncedInNamespace map[types.UID]bool, namespaceName string) {
	zap.S().Infof("Running sync pods of workloads in project <%s[%s]>", projectName, projectId)
	allPodsToSync := map[string][]openapi.PodSyncFields{}
	podsPerWorkloadMaxNumber := float64(0)
	for workload := range podsPerWorkload {
		podsToSync, err := h.convertPodsToSync(podsPerWorkload[workload], syncedInNamespace)
		if err != nil {
			zap.S().Errorf("Had an error while syncing Pods in project <%s[%s]>: %s",
				projectName, projectId, err.Error())
			syncedInNamespace = map[types.UID]bool{}
		}
		allPodsToSync[workload] = podsToSync
		podsPerWorkloadMaxNumber = math.Max(podsPerWorkloadMaxNumber, float64(len(podsToSync)))
	}
	err := h.handleSyncInWorkload(allPodsToSync, projectId, projectName, int(podsPerWorkloadMaxNumber+1))
	if err != nil {
		zap.S().Errorf("Had an error while syncing Pods per workload in project <%s[%s]>: %s",
			projectName, projectId, err.Error())
		syncedInNamespace = map[types.UID]bool{}
	}

	h.syncedObjectsMutex.Lock()
	h.syncedObjects[namespaceName] = syncedInNamespace
	h.syncedObjectsMutex.Unlock()

	h.updateSyncFailedProjects(projectName, err)
}

func (h *PodHandler) getPodsToSync(namespaceName string,
	syncedInNamespace map[types.UID]bool) ([]openapi.PodSyncFields, error) {

	runaiSchedulerFieldSelector :=
		fields.OneTermEqualSelector(constants.PodBySchedulerNameIndexer, constants.RunaiSchedulerName)

	pods := &coreV1.PodList{}
	err := h.cachedClient.List(context.TODO(), pods,
		&client.ListOptions{Namespace: namespaceName, FieldSelector: runaiSchedulerFieldSelector})
	if err != nil {
		zap.S().Errorf("Failed listing pods in namespace <%s>, error: %s",
			namespaceName, err.Error())
		return []openapi.PodSyncFields{}, err
	}

	podsToSync, err := h.convertPodsToSync(pods.Items, syncedInNamespace)
	if err != nil {
		return nil, err
	}

	return podsToSync, nil
}

func (h *PodHandler) convertPodsToSync(pods []coreV1.Pod, syncedInNamespace map[types.UID]bool) ([]openapi.PodSyncFields, error) {
	podsToSync := []openapi.PodSyncFields{}
	for _, pod := range pods {
		parsedPod, err := h.podParser.ParsePodToServerScheme(&pod)
		if err != nil {
			zap.S().Errorf("Failed parsing pod <%s/%s>, error: %s",
				pod.Namespace, pod.Name, err.Error())
			return nil, err
		}

		syncedInNamespace[pod.UID] = true
		podsToSync = append(podsToSync, *parsedPod)
	}
	return podsToSync, nil
}

func (h *PodHandler) handleSyncInProjectWithPagination(podsToSync []openapi.PodSyncFields,
	projectId string, projectName string) error {
	if len(podsToSync) <= h.maxNumOfPodsPerRequest {
		return h.sendSyncPodsRequests(podsToSync, projectId, projectName, nil, nil)
	}

	// sort all pods first
	sortNewApiPodsById(podsToSync)

	for i := 0; i < len(podsToSync); i += h.maxNumOfPodsPerRequest - 1 {
		chunkEndIndex := i + h.maxNumOfPodsPerRequest
		if chunkEndIndex > len(podsToSync) {
			chunkEndIndex = len(podsToSync)
		}

		chunk := podsToSync[i:chunkEndIndex]

		var from *string = nil
		var to *string = nil
		if i > 0 {
			from = ptr.To(chunk[0].Id)
		}
		if chunkEndIndex < len(podsToSync) {
			to = ptr.To(chunk[len(chunk)-1].Id)
		}

		if err := h.sendSyncPodsRequests(chunk, projectId, projectName, from, to); err != nil {
			return err
		}
	}
	return nil
}

func (h *PodHandler) handleSyncInWorkload(podsToSync map[string][]openapi.PodSyncFields, projectId string, projectName string, podsPerWorkloadMaxNumber int) error {
	chunkSize := int(math.Max(2, float64(h.maxNumOfPodsPerRequest/podsPerWorkloadMaxNumber)))

	podsWorkloads := []string{}
	for i := range podsToSync {
		podsWorkloads = append(podsWorkloads, i)
	}

	for i := 0; i < len(podsWorkloads); i += chunkSize - 1 {
		chunkEndIndex := i + chunkSize
		if chunkEndIndex > len(podsWorkloads) {
			chunkEndIndex = len(podsWorkloads)
		}

		workloadsInChunk := podsWorkloads[i:chunkEndIndex]
		chunk := map[string][]openapi.PodSyncFields{}
		for _, workload := range workloadsInChunk {
			chunk[workload] = podsToSync[workload]
		}

		if err := h.sendSyncPodsPerWorkloadRequests(chunk, projectId, projectName); err != nil {
			return err
		}
	}
	return nil
}

func (h *PodHandler) sendSyncPodsRequests(podsToSync []openapi.PodSyncFields,
	projectId string, projectName string, from, to *string) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.SyncOperation)()
	msg := fmt.Sprintf("Sync Pods in project <%s[%s]>", projectName, projectId)
	if from != nil && *from != "" {
		msg = fmt.Sprintf("%s; from <%s>", msg, *from)
	}
	if to != nil && *to != "" {
		msg = fmt.Sprintf("%s; to <%s>", msg, *to)
	} else {
		msg = fmt.Sprintf("Finished %s", msg)
	}

	request := h.backendClient.WorkloadsClient.PodsAPI.SyncWorkloadsPods(h.backendClient.Context())
	request = request.PodsSyncRequest(openapi.PodsSyncRequest{
		ProjectId: projectId,
		ClusterId: ptr.To(config.Config().ClusterUuid),
		Pods:      podsToSync,
		From:      from,
		To:        to,
	})

	res, err := request.Execute()
	if res != nil {
		defer func() { _ = res.Body.Close() }()
	}
	logDetailedApiErrorMessageIfNeeded(res, err, msg)
	statusCode, err := h.backendClient.HandleHttpResponseGetStatus(res, err, msg)
	if statusCode == http.StatusRequestEntityTooLarge && h.maxNumOfPodsPerRequest > 2 {
		// reduce by 20%
		h.maxNumOfPodsPerRequest = int(h.maxNumOfPodsPerRequest * 4 / 5)
		zap.S().Warnf("Changing maxNumOfPodsPerRequest to: %d", h.maxNumOfPodsPerRequest)
	}
	return err
}

func (h *PodHandler) sendSyncPodsPerWorkloadRequests(podsToSync map[string][]openapi.PodSyncFields, projectId string, projectName string) error {
	msg := fmt.Sprintf("Sync Pods in project <%s[%s]>", projectName, projectId)

	request := h.backendClient.WorkloadsClient.PodsAPI.SyncPods(h.backendClient.Context())
	podsRequest := []openapi.WorkloadPodsSyncRequest{}
	for workload := range podsToSync {
		podsRequest = append(podsRequest, openapi.WorkloadPodsSyncRequest{
			ProjectId:  projectId,
			ClusterId:  ptr.To(config.Config().ClusterUuid),
			WorkloadId: workload,
			Pods:       podsToSync[workload],
		})
	}
	request = request.WorkloadPodsSyncRequest(podsRequest)

	res, err := request.Execute()
	if res != nil {
		defer func() { _ = res.Body.Close() }()
	}
	logDetailedApiErrorMessageIfNeeded(res, err, msg)
	statusCode, err := h.backendClient.HandleHttpResponseGetStatus(res, err, msg)
	if statusCode == http.StatusRequestEntityTooLarge && h.maxNumOfPodsPerRequest > 2 {
		// reduce by 20%
		h.maxNumOfPodsPerRequest = int(h.maxNumOfPodsPerRequest * 4 / 5)
		zap.S().Warnf("Changing maxNumOfPodsPerRequest to: %d", h.maxNumOfPodsPerRequest)
	}
	return err
}

// TODO: rename after removing deprecated
func sortNewApiPodsById(pods []openapi.PodSyncFields) {
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].Id < pods[j].Id
	})
}

func (h *PodHandler) enqueuePodsHandlingForLaterIfNeeded(res *http.Response, err error,
	pod *coreV1.Pod, workloadId string) {

	zap.S().Infof("Workload <%s> of pod <%s/%s> was not added yet - enqueueing pod to be added later",
		workloadId, pod.Namespace, pod.Name)

	h.enqueuePodsAdditionMutex.Lock()
	defer h.enqueuePodsAdditionMutex.Unlock()

	_, found := h.enqueuePodsAddition[workloadId]
	if !found {
		h.enqueuePodsAddition[workloadId] = map[client.ObjectKey]*coreV1.Pod{}
	}
	h.enqueuePodsAddition[workloadId][client.ObjectKeyFromObject(pod)] = pod.DeepCopy()
}

// HandleAdditionOfWorkloadPodsAfterItWasAdded handles pods after their workload is handled.
// If pods are synced before their workload is synced,
// we will enqueue them and handle them after their workload is synced.
func (h *PodHandler) HandleAdditionOfWorkloadPodsAfterItWasAdded(workloadId string,
	podRateLimiter *cachedclient.RateLimitedController) error {
	// take the batch out of the map before handling it, so no other goroutine can write to
	// the map we are ranging over, and pods enqueued meanwhile survive for the next drain
	h.enqueuePodsAdditionMutex.Lock()
	podsToHandle, found := h.enqueuePodsAddition[workloadId]
	delete(h.enqueuePodsAddition, workloadId)
	h.enqueuePodsAdditionMutex.Unlock()
	if !found || len(podsToHandle) == 0 {
		return nil
	}

	zap.S().Infof("Handling addition of %d enqueued pods of workload %s",
		len(podsToHandle), workloadId)

	var err error = nil
	for key, pod := range podsToHandle {
		zap.S().Infof("Detected <%s/%s> as child of workload <%s>",
			key.Namespace, key.Name, workloadId)
		if podRateLimiter != nil {
			podRateLimiter.AddRateLimited(key)
		} else {
			innerErr := h.handleAddUpdate(pod)
			if innerErr != nil {
				err = multierror.Append(err, innerErr)
			}
		}
	}

	return err
}

func (h *PodHandler) ResyncPodsAfterWorkloadAdded(projectName string) {
	h.syncFailedProjectsMutex.RLock()
	_, found := h.syncFailedProjects[projectName]
	// unlock read while handling
	h.syncFailedProjectsMutex.RUnlock()
	if !found {
		return
	}

	h.syncFailedProjectsMutex.Lock()
	delete(h.syncFailedProjects, projectName)
	h.syncFailedProjectsMutex.Unlock()

	go h.reSyncPodsInProject(projectName)
}

func (h *PodHandler) reSyncPodsInProject(projectName string) {
	zap.S().Infof("Handling re-sync of pods in project %s", projectName)

	project, err := util.GetProject(projectName, h.cachedClient)
	if err != nil {
		return
	}

	h.handleUpdateInProject(project)
}

func (h *PodHandler) updateSyncFailedProjects(projectName string, syncErr error) {
	h.syncFailedProjectsMutex.RLock()
	_, found := h.syncFailedProjects[projectName]
	h.syncFailedProjectsMutex.RUnlock()

	if found && syncErr != nil {
		return
	}
	if !found && syncErr == nil {
		return
	}

	h.syncFailedProjectsMutex.Lock()
	defer h.syncFailedProjectsMutex.Unlock()

	if syncErr != nil {
		h.syncFailedProjects[projectName] = true
	} else {
		delete(h.syncFailedProjects, projectName)
	}
}
