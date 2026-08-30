package handlers

import (
	"context"
	corev1 "k8s.io/api/core/v1"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkspaceHandler struct {
	*WorkloadHandler
}

func (h *WorkspaceHandler) HandleUpdateItem(key client.ObjectKey) error {
	iw := &iwv2a1.InteractiveWorkload{}
	err := h.cachedClient.Get(context.Background(), key, iw)
	if err != nil {
		return err
	}

	return h.HandleInteractiveUpdate(iw)
}

func NewWorkspaceHandler(workloadHandler *WorkloadHandler) *WorkspaceHandler {
	return &WorkspaceHandler{WorkloadHandler: workloadHandler}
}

func (h *WorkloadHandler) HandleInteractiveAdd(workload *iwv2a1.InteractiveWorkload) error {
	if h.WasSynced(workload.UID, workload.Namespace) {
		zap.S().Infof("InteractiveWorkload <%s/%s> was already synced - not adding",
			workload.Namespace, workload.Name)
		return nil
	}
	zap.S().Debugf("Handling InteractiveWorkload <%s/%s> add", workload.Namespace, workload.Name)

	return h.handleInteractiveAddUpdate(workload)
}

func (h *WorkloadHandler) HandleInteractiveUpdate(workload *iwv2a1.InteractiveWorkload) error {
	zap.S().Debugf("Handling InteractiveWorkload <%s/%s> update", workload.Namespace, workload.Name)
	return h.handleInteractiveAddUpdate(workload)
}

func (h *WorkloadHandler) handleInteractiveAddUpdate(workload *iwv2a1.InteractiveWorkload) error {
	workloadSyncFields, _, err := h.parser.ParseInteractiveWorkloadToServerScheme(workload)
	if err != nil {
		return err
	}
	jobStatus := workload.Status.DeepCopyToJobStatus()
	if isLegacyWorkloadUpToDate(jobStatus.SyncStatus, workload.Generation,
		workload.Status.LastStatusChangeTime, workloadSyncFields) {
		zap.S().Debugf("InteractiveWorkload <%s/%s> is up to date - not updating",
			workload.Namespace, workload.Name)
		return nil
	}

	return h.handleUpsert(workloadSyncFields, workload.Namespace, workload.Name, constants.InteractiveWorkload)
}

func (h *WorkloadHandler) getParsedInteractiveWorkloadsInNamespace(namespaceName string,
	syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, error) {

	fieldSelector := GetRelevantRunaiWorkloadsFieldSelector()

	workloads := &iwv2a1.InteractiveWorkloadList{}
	err := h.cachedClient.List(context.Background(), workloads, &client.ListOptions{
		Namespace:     namespaceName,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		zap.S().Errorf("Failed listing InteractiveWorkloads in namespace <%s>, error: %s",
			namespaceName, err.Error())
		return []openapi.WorkloadSyncFields{}, []string{}, []client.Object{}, err
	}

	workloadSyncFields, upToDateIds, updateSyncStatus, _, err := h.parseInteractiveWorkloads(workloads.Items, syncedInNamespace)
	return workloadSyncFields, upToDateIds, updateSyncStatus, err
}

func (h *WorkloadHandler) parseInteractiveWorkloads(workloads []iwv2a1.InteractiveWorkload, syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, []corev1.Pod, error) {
	workloadsToSync := []openapi.WorkloadSyncFields{}
	upToDateIds := []string{}
	workloadPods := []corev1.Pod{}
	for i := range workloads {
		// important: to avoid bug where we insert a pointer of for-looped map value to an array
		workload := &workloads[i]
		syncedInNamespace[workload.UID] = true

		parsed, pods, innerErr := h.parser.ParseInteractiveWorkloadToServerScheme(workload)
		if innerErr != nil {
			zap.S().Warnf("Failed parsing InteractiveWorkload <%s/%s>, error: %s",
				workload.Namespace, workload.Name, innerErr.Error())
			continue
		}
		workloadPods = append(workloadPods, pods...)
		jobStatus := workload.Status.DeepCopyToJobStatus()
		if isLegacyWorkloadUpToDate(jobStatus.SyncStatus, workload.Generation,
			workload.Status.LastStatusChangeTime, parsed) {
			upToDateIds = append(upToDateIds, parsed.Id)
			continue
		}
		workloadsToSync = append(workloadsToSync, *parsed)
	}

	return workloadsToSync, upToDateIds, nil, workloadPods, nil
}
