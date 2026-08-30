package update

import (
	"context"
	"encoding/json"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/controller"
	utilsconverter "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/utils/converter"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	workloadupdater "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/status/update"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/convertors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InferenceWorkloadUpdater struct {
	updater *workloadupdater.WorkloadUpdater
}

func NewInferenceWorkloadUpdater(client client.Client, workloadType constants.WorkloadType, convertor convertors.Convertor) *InferenceWorkloadUpdater {
	return &InferenceWorkloadUpdater{
		updater: workloadupdater.NewWorkloadUpdater(client, workloadType, convertor, &WorkloadStatusPatcher{}),
	}
}

// UpdateStatus updates the status of the workload
func (wu *InferenceWorkloadUpdater) UpdateStatus(ctx context.Context, modifiedWorkload *jobs.JobWorkload) error {
	return wu.updater.UpdateStatus(ctx, modifiedWorkload)
}

type InferenceWorkloadStatusUpdaterFactory struct {
}

func (self *InferenceWorkloadStatusUpdaterFactory) CreateStatusUpdater(client client.Client) workloadupdater.StatusUpdater {
	return NewInferenceWorkloadUpdater(client, controller.TypeInferenceWorkload, utilsconverter.NewInferenceWorkloadConvertor())
}

type WorkloadStatusPatcher struct {
}

func (wu *WorkloadStatusPatcher) PatchBytes(ctx context.Context, modifiedWorkload *jobs.JobWorkload) ([]byte, error) {
	jobStatus := modifiedWorkload.Status

	return json.Marshal(map[string]interface{}{"status": map[string]interface{}{
		/// InferenceStatus
		"latestCreatedRevisionName": jobStatus.LatestCreatedRevisionName,
		"latestReadyRevisionName":   jobStatus.LatestReadyRevisionName,
		"externalUrls":              jobStatus.InferenceStatus.ExternalUrls,
		"phase":                     jobStatus.InferenceStatus.Phase,
		"phaseMessage":              jobStatus.InferenceStatus.PhaseMessage,

		/// NetworkStatus
		"services": jobStatus.ServiceStatuses,

		/// WorkloadDebugStatus
		"observedGeneration":    jobStatus.ObservedGeneration,
		"firstAcknowledgedTime": jobStatus.FirstAcknowledgedTime,
		"ownedResources":        jobStatus.OwnedResources,
		"lastStatusChangeTime":  jobStatus.LastStatusChangeTime,
	}})
}
