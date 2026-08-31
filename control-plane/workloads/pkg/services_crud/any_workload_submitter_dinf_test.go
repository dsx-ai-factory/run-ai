/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"testing"
	"time"

	clientsclusterservice "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/notifications"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/clusterservice"
	testing_pkg "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadMgrApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
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

func TestAnyWorkloadSubmitterDinfCreateReturnsConnectionMetadata(t *testing.T) {
	const clusterID = "dinf-connections-cluster"
	displayedCluster := testing_pkg.BuildDisplayedCluster(clusterID, "3.0.0", nil)

	originalCapabilitiesClient := checkerClusterClient.Get()
	checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().WithCluster(clusterID, displayedCluster))
	checker.InitClustersCache(testing_pkg.GetDefaultClustersCacheConfig())
	t.Cleanup(func() {
		checkerClusterClient.SetMock(originalCapabilitiesClient)
		checker.InitClustersCache(testing_pkg.GetDefaultClustersCacheConfig())
	})

	originalClusterService := clusterservice.Get()
	clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithCluster(clusterID, &displayedCluster))
	t.Cleanup(func() { clusterservice.SetMock(originalClusterService) })

	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	spec.SetWorkers(1)
	spec.SetReplicas(1)
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")
	spec.SetLeader(*leader)
	worker := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	worker.SetImage("worker:latest")
	spec.SetWorker(*worker)
	servingPort := workloadsApi.NewDistributedInferenceServingPort()
	servingPort.SetPort(8080)
	servingPort.SetAuthorizationType(workloadsApi.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_PUBLIC)
	servingPort.SetExposeExternally(false)
	spec.SetServingPort(*servingPort)

	supersetSpec := workloadsApi.NewSupersetSpec()
	require.NoError(t, convutils.CopySpecs(supersetSpec, spec))

	notificationsComposer := notifications.NewNotificationsComposer(
		string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		"dist-inf",
		"00000000-0000-0000-0000-000000000123",
		"requester@example.com",
		clusterID,
		"project-id",
		"department-id",
		time.Time{},
		nil,
		composer.FieldsComposerOptions{},
	).WithConnectionsPreProcessor("project", "")
	params := &createWorkloadParams{Response: workloadsApi.SupersetWorkload{
		Name:       "dist-inf",
		WorkloadId: "00000000-0000-0000-0000-000000000123",
		ProjectId:  "project-id",
		ClusterId:  clusterID,
		Spec:       supersetSpec,
	}}
	wc := &workloadContext{
		workloadType:             string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		workloadName:             "dist-inf",
		createdBy:                "requester@example.com",
		spec:                     supersetSpec,
		distributedInferenceSpec: spec,
		clusterCapabilities:      checker.NewClusterCapabilitiesChecker(clusterID),
		create: &workloadCreateContext{
			createParams:          params,
			project:               &orgUnitApi.Project{Name: "project"},
			namespace:             "runai-project",
			notificationsComposer: notificationsComposer,
		},
	}

	submission, submissionErr := (&AnyWorkloadSubmitterDinf{}).Create(context.Background(), wc)

	require.Nil(t, submissionErr)
	require.NotNil(t, submission)
	require.NotNil(t, submission.ConnectionMetadata)
	require.Same(t, notificationsComposer.Output, submission.ConnectionCreationFields)
	require.Len(t, submission.ConnectionMetadata.Environments, 2)

	leaderEnvironment := environmentByReplicaTypeForTest(submission.ConnectionMetadata.Environments, workloadsApi.REPLICATYPE_LEADER)
	require.NotNil(t, leaderEnvironment)
	require.Len(t, leaderEnvironment.Connections, 1)
	require.Equal(t, int32(8080), leaderEnvironment.Connections[0].GetContainerPort())

	workerEnvironment := environmentByReplicaTypeForTest(submission.ConnectionMetadata.Environments, workloadsApi.REPLICATYPE_WORKER)
	require.NotNil(t, workerEnvironment)
	require.Empty(t, workerEnvironment.Connections)
}

func TestComposeDinfConnectionMetadataCreatesLeaderAndWorkerEnvironments(t *testing.T) {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	spec.SetWorkers(2)
	spec.SetReplicas(3)
	spec.SetNodePools([]string{"pool-a"})

	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")
	spec.SetLeader(*leader)
	worker := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	worker.SetImage("worker:latest")
	spec.SetWorker(*worker)

	servingPort := workloadsApi.NewDistributedInferenceServingPort()
	servingPort.SetPort(8081)
	servingPort.SetAuthorizationType(workloadsApi.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_AUTHORIZED_USERS)
	servingPort.SetAuthorizedUsers([]string{"user@example.com"})
	spec.SetServingPort(*servingPort)

	notificationsComposer := newDinfNotificationsComposerForTest()
	wc := &workloadContext{
		create: &workloadCreateContext{notificationsComposer: notificationsComposer},
	}

	metadataErr := composeDinfConnectionMetadata(context.Background(), wc, spec)

	require.Nil(t, metadataErr)
	require.Len(t, notificationsComposer.Output.Environments, 2)
	require.Equal(t, int32(9), notificationsComposer.Output.RequestedPods.GetNumber())
	require.Equal(t, []string{"pool-a"}, notificationsComposer.Output.RequestedNodePools)

	leaderEnvironment := environmentByReplicaTypeForTest(notificationsComposer.Output.Environments, workloadsApi.REPLICATYPE_LEADER)
	require.NotNil(t, leaderEnvironment)
	require.Equal(t, "Leader", leaderEnvironment.Name)
	require.Len(t, leaderEnvironment.Connections, 1)
	leaderConnection := leaderEnvironment.Connections[0]
	require.Equal(t, "ServingPort", leaderConnection.ConnectionType)
	require.Equal(t, "ServingPort", leaderConnection.Name)
	require.Equal(t, "custom", leaderConnection.ToolType)
	require.Equal(t, int32(8081), leaderConnection.GetContainerPort())
	require.Equal(t, "authorizedUsers", leaderConnection.GetAuthorizationType())
	require.Equal(t, []string{"user@example.com"}, leaderConnection.AuthorizedUsers)
	require.Equal(t, "serving-port", leaderConnection.GetCorrelationName())

	workerEnvironment := environmentByReplicaTypeForTest(notificationsComposer.Output.Environments, workloadsApi.REPLICATYPE_WORKER)
	require.NotNil(t, workerEnvironment)
	require.Equal(t, "Worker", workerEnvironment.Name)
	require.Empty(t, workerEnvironment.Connections)

	connectionMetadata := newWorkloadsServiceConnectionMetadata(notificationsComposer.Output)
	require.NotNil(t, connectionMetadata)
	require.Len(t, connectionMetadata.Environments, 2)
}

func TestComposeDinfConnectionMetadataCreatesWorkerEnvironmentForCopiedLeaderSpec(t *testing.T) {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	spec.SetWorkers(1)

	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("shared:latest")
	spec.SetLeader(*leader)

	servingPort := workloadsApi.NewDistributedInferenceServingPort()
	servingPort.SetPort(8080)
	spec.SetServingPort(*servingPort)

	notificationsComposer := newDinfNotificationsComposerForTest()
	wc := &workloadContext{
		create: &workloadCreateContext{notificationsComposer: notificationsComposer},
	}

	metadataErr := composeDinfConnectionMetadata(context.Background(), wc, spec)

	require.Nil(t, metadataErr)
	require.Len(t, notificationsComposer.Output.Environments, 2)
	require.NotNil(t, environmentByReplicaTypeForTest(notificationsComposer.Output.Environments, workloadsApi.REPLICATYPE_LEADER))
	workerEnvironment := environmentByReplicaTypeForTest(notificationsComposer.Output.Environments, workloadsApi.REPLICATYPE_WORKER)
	require.NotNil(t, workerEnvironment)
	require.Empty(t, workerEnvironment.Connections)
}

func newDinfNotificationsComposerForTest() *notifications.NotificationsComposer {
	return notifications.NewNotificationsComposer(
		string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		"dist-inf",
		"workload-id",
		"requester@example.com",
		"cluster-id",
		"project-id",
		"department-id",
		time.Time{},
		nil,
		composer.FieldsComposerOptions{},
	)
}
