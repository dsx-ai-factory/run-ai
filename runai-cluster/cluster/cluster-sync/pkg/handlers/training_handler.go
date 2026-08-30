package handlers

import (
	"context"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TrainingHandler struct {
	*WorkloadHandler
}

func (h *TrainingHandler) HandleUpdateItem(key client.ObjectKey) error {
	tw := &v2alpha1.TrainingWorkload{}
	err := h.cachedClient.Get(context.Background(), key, tw)
	if err != nil {
		return err
	}

	return h.HandleTrainUpdate(tw)
}

func NewTrainingHandler(workloadHandler *WorkloadHandler) *TrainingHandler {
	return &TrainingHandler{WorkloadHandler: workloadHandler}
}

func (h *WorkloadHandler) HandleTrainAdd(workload *v2alpha1.TrainingWorkload) error {
	if h.WasSynced(workload.UID, workload.Namespace) {
		zap.S().Infof("TrainingWorkload <%s/%s> was already synced - not adding",
			workload.Namespace, workload.Name)
		return nil
	}
	zap.S().Debugf("Handling TrainingWorkload <%s/%s> add", workload.Namespace, workload.Name)

	return h.handleTrainAddUpdate(workload)
}

func (h *WorkloadHandler) HandleTrainUpdate(workload *v2alpha1.TrainingWorkload) error {
	zap.S().Debugf("Handling TrainingWorkload <%s/%s> update", workload.Namespace, workload.Name)
	return h.handleTrainAddUpdate(workload)
}

func (h *WorkloadHandler) handleTrainAddUpdate(workload *v2alpha1.TrainingWorkload) error {
	workloadSyncFields, _, err := h.parser.ParseTrainingWorkloadToServerScheme(workload)
	if err != nil {
		return err
	}
	jobStatus := workload.Status.DeepCopyToJobStatus()
	if isLegacyWorkloadUpToDate(jobStatus.SyncStatus, workload.Generation,
		workload.Status.LastStatusChangeTime, workloadSyncFields) {
		zap.S().Debugf("TrainingWorkload <%s/%s> is up to date - not updating",
			workload.Namespace, workload.Name)
		return nil
	}

	return h.handleUpsert(workloadSyncFields, workload.Namespace, workload.Name, constants.TrainingWorkload)
}

func (h *WorkloadHandler) getParsedTrainingWorkloadsInNamespace(namespaceName string,
	syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, error) {

	fieldSelector := GetRelevantRunaiWorkloadsFieldSelector()

	workloads := &v2alpha1.TrainingWorkloadList{}
	err := h.cachedClient.List(context.Background(), workloads, &client.ListOptions{
		Namespace:     namespaceName,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		zap.S().Errorf("Failed listing TrainingWorkloads in namespace <%s>, error: %s",
			namespaceName, err.Error())
		return []openapi.WorkloadSyncFields{}, []string{}, []client.Object{}, err
	}

	workloadSyncFields, upToDateIds, updateSyncStatus, _, err := h.parseTrainingWorkloads(workloads.Items, syncedInNamespace)
	return workloadSyncFields, upToDateIds, updateSyncStatus, err
}

func (h *WorkloadHandler) parseTrainingWorkloads(workloads []v2alpha1.TrainingWorkload, syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, []corev1.Pod, error) {
	workloadsToSync := []openapi.WorkloadSyncFields{}
	upToDateIds := []string{}
	workloadPods := []corev1.Pod{}
	for i := range workloads {
		// important: to avoid bug where we insert a pointer of for-looped map value to an array
		workload := &workloads[i]
		syncedInNamespace[workload.UID] = true

		parsed, pods, innerErr := h.parser.ParseTrainingWorkloadToServerScheme(workload)
		if innerErr != nil {
			zap.S().Warnf("Failed parsing TrainingWorkload <%s/%s>, error: %s",
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

func FilterWorkloads(workloadFields field_groups.WorkloadFields) bool {
	return workloadFields.Usage == constants.RunaiWorkloadRelevantUsage
}

func GetRelevantRunaiWorkloadsFieldSelector() fields.Selector {
	return fields.OneTermEqualSelector(
		constants.RunaiWorkloadByUsageIndexer, constants.RunaiWorkloadRelevantUsage)
}
