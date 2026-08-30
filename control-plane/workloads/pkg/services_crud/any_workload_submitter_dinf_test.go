package services_crud

import (
	"testing"

	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadMgrApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/stretchr/testify/require"
)

func TestProjectDistributedInferenceSpec(t *testing.T) {
	superset := &workloadsApi.SupersetSpec{}
	superset.SetWorkers(2)
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")
	superset.SetLeader(*leader)
	worker := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	worker.SetImage("worker:latest")
	superset.SetWorker(*worker)

	projected, err := projectDistributedInferenceSpec(superset)

	require.NoError(t, err)
	require.Equal(t, int32(2), projected.GetWorkers())
	projectedLeader := projected.GetLeader()
	projectedWorker := projected.GetWorker()
	require.Equal(t, "leader:latest", projectedLeader.GetImage())
	require.Equal(t, "worker:latest", projectedWorker.GetImage())
}

func TestBuildDinfWorkloadV2Metadata(t *testing.T) {
	name := "dist-inf"
	projectID := "project-a"
	category := "Inference"
	priority := "high"
	preemptibility := workloadsApi.PREEMPTIBILITY_NON_PREEMPTIBLE
	spec := &workloadsApi.DistributedInferenceSpecSpec{}
	spec.SetCategory(category)
	spec.SetPriorityClass(priority)
	spec.SetPreemptibility(preemptibility)
	configuration := &workloadMgrApi.WorkloadV2Configuration{}

	metadata := buildDinfWorkloadV2Metadata(&name, &projectID, configuration, spec)

	require.Equal(t, name, *metadata.Name)
	require.Equal(t, projectID, *metadata.ProjectId)
	require.Same(t, configuration, metadata.Configuration)
	require.Equal(t, category, *metadata.Category.Get())
	require.Equal(t, priority, *metadata.Priority.Get())
	require.Equal(t, workloadMgrApi.PREEMPTIBILITY_NON_PREEMPTIBLE, *metadata.Preemptibility.Get())
}

func TestManifestName(t *testing.T) {
	name, ok := manifestName(map[string]interface{}{
		"metadata": map[string]interface{}{"name": "dist-inf-abcd"},
	})
	require.True(t, ok)
	require.Equal(t, "dist-inf-abcd", name)

	_, ok = manifestName(map[string]interface{}{"metadata": "invalid"})
	require.False(t, ok)
}
