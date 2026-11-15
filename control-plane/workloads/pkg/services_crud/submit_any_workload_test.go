/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/notifications"
	workloadsServerApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/google/uuid"
)

func workloadManagerUpdateResultForSnapshot(
	snapshot *models_crud.WorkloadV2Snapshot,
	group string,
	version string,
	kind string,
	manifestName string,
) (*workloadMgrApi.WorkloadV2, error) {
	metadata := workloadMgrApi.WorkloadV2Metadata{}
	if err := convutils.CopySpecs(&metadata, snapshot); err != nil {
		return nil, err
	}
	metadata.ProjectName = "project"
	metadata.DepartmentName = "department"
	metadata.Gvk = workloadMgrApi.GVK{
		Group:   *workloadMgrApi.NewNullableString(convutils.Ref(group)),
		Version: version,
		Kind:    kind,
	}

	return &workloadMgrApi.WorkloadV2{
		Metadata:     metadata,
		DesiredPhase: workloadMgrApi.DESIREDPHASE_RUNNING,
		WorkloadManifest: map[string]interface{}{
			"apiVersion": group + "/" + version,
			"kind":       kind,
			"metadata":   map[string]interface{}{"name": manifestName},
		},
	}, nil
}

func TestWorkloadFromV2SnapshotUsesSnapshotFieldsAndDesiredPhase(t *testing.T) {
	workloadID := uuid.New()
	createdAt := time.Now().Add(-time.Hour).Truncate(time.Microsecond)
	spec := workloadsApi.NewSupersetSpec()
	spec.SetImage("image:v2")
	masterSpec := workloadsApi.NewMasterSpec()
	masterSpec.SetImage("leader:v2")
	masterSameAsWorker := true
	snapshot := &models_crud.WorkloadV2Snapshot{
		ID:                 workloadID,
		CreatedAt:          createdAt,
		CreatedBy:          "snapshot-user",
		ClusterId:          "snapshot-cluster",
		DepartmentId:       "snapshot-department",
		ProjectId:          "snapshot-project",
		RequestedName:      "requested-name",
		Name:               "snapshot-name",
		Spec:               spec,
		MasterSpec:         masterSpec,
		MasterSameAsWorker: &masterSameAsWorker,
	}

	response := workloadFromV2Snapshot(snapshot, workloadsApi.WORKLOADDESIREDPHASE_RUNNING)

	if response.WorkloadId != workloadID.String() || response.Name != "snapshot-name" {
		t.Fatalf("expected snapshot identity, got id=%q name=%q", response.WorkloadId, response.Name)
	}
	if response.ProjectId != "snapshot-project" || response.ClusterId != "snapshot-cluster" {
		t.Fatalf("expected snapshot project and cluster, got project=%q cluster=%q", response.ProjectId, response.ClusterId)
	}
	if response.DepartmentId == nil || *response.DepartmentId != "snapshot-department" {
		t.Fatalf("expected snapshot department, got %#v", response.DepartmentId)
	}
	if response.CreatedBy != "snapshot-user" || !response.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected snapshot creation metadata, got user=%q createdAt=%s", response.CreatedBy, response.CreatedAt)
	}
	if response.DesiredPhase != workloadsApi.WORKLOADDESIREDPHASE_RUNNING {
		t.Fatalf("expected Running desired phase, got %q", response.DesiredPhase)
	}
	if response.RequestedName != "requested-name" || !reflect.DeepEqual(response.Spec, spec) {
		t.Fatalf("expected requested name and updated typed spec from the snapshot")
	}
	if !reflect.DeepEqual(response.MasterSpec.Get(), masterSpec) || response.MasterSpecSameAsWorker.Get() == nil || !*response.MasterSpecSameAsWorker.Get() {
		t.Fatalf("expected distributed inference master fields from the snapshot")
	}
	if response.ActualPhase != nil {
		t.Fatalf("expected actual phase to remain omitted, got %q", *response.ActualPhase)
	}

	body, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal update response: %v", err)
	}
	var decoded workloadsApi.SupersetWorkload
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("generated client rejected update response: %v", err)
	}
}

func TestNewWorkloadsServiceConnectionMetadataExtractsConnectionMetadata(t *testing.T) {
	environments := []workloadsApi.Environment{
		{
			Name: "env",
			Connections: []workloadsApi.Connection{
				{
					Name:                "Jupyter",
					ConnectionType:      "NodePort",
					ContainerPort:       convutils.Ref(int32(8888)),
					ServiceResourceName: *workloadsApi.NewNullableString(convutils.Ref("job-test-np")),
					PortResourceName:    *workloadsApi.NewNullableString(convutils.Ref("port-8888")),
				},
			},
		},
	}
	externalConnections := []workloadsApi.Connection{
		{
			Name:                "TensorBoard",
			ConnectionType:      "ExternalUrl",
			ServiceResourceName: *workloadsApi.NewNullableString(convutils.Ref("job-test-url")),
			ExposedUrlPath:      *workloadsApi.NewNullableString(convutils.Ref("/p1/test")),
		},
	}
	creation := &workloadsApi.WorkloadCreationFields{
		Type:                string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		Name:                "test",
		Id:                  "workloads-service-id",
		ClusterId:           "cluster-id",
		ProjectId:           "project-id",
		CreatedAt:           time.Now(),
		Environments:        environments,
		ExternalConnections: externalConnections,
	}

	metadata := newWorkloadsServiceConnectionMetadata(creation)

	if metadata == nil {
		t.Fatalf("expected metadata")
	}
	if !reflect.DeepEqual(metadata.Environments, environments) {
		t.Fatalf("expected environments to be extracted, got %#v", metadata.Environments)
	}
	if !reflect.DeepEqual(metadata.ExternalConnections, externalConnections) {
		t.Fatalf("expected external connections to be extracted, got %#v", metadata.ExternalConnections)
	}
}

func TestWorkloadsClientConnectionMetadataToServerMetadata(t *testing.T) {
	metadata := &WorkloadsServiceConnectionMetadata{
		Environments: []workloadsApi.Environment{
			{
				Name: "env",
				Connections: []workloadsApi.Connection{
					{
						Name:                "Jupyter",
						ConnectionType:      "NodePort",
						ContainerPort:       convutils.Ref(int32(8888)),
						ServiceResourceName: *workloadsApi.NewNullableString(convutils.Ref("job-test-np")),
						PortResourceName:    *workloadsApi.NewNullableString(convutils.Ref("port-8888")),
					},
				},
			},
		},
	}

	serverMetadata, err := workloadsClientConnectionMetadataToServerMetadata(metadata)

	if err != nil {
		t.Fatalf("expected conversion to succeed, got %v", err)
	}
	if len(serverMetadata.Environments) != 1 || len(serverMetadata.Environments[0].Connections) != 1 {
		t.Fatalf("expected server environments to be converted, got %#v", serverMetadata.Environments)
	}
	connection := serverMetadata.Environments[0].Connections[0]
	if connection.ServiceResourceName == nil || *connection.ServiceResourceName != "job-test-np" {
		t.Fatalf("expected serviceResourceName job-test-np, got %#v", connection.ServiceResourceName)
	}
	if connection.PortResourceName == nil || *connection.PortResourceName != "port-8888" {
		t.Fatalf("expected portResourceName port-8888, got %#v", connection.PortResourceName)
	}
}

func TestWorkloadsServiceCreationFieldsForConnectionMetadataUsesWorkloadsManagerMetadata(t *testing.T) {
	createdAt := time.Now()
	preemptibility := workloadMgrApi.PREEMPTIBILITY_NON_PREEMPTIBLE
	sourceCategory := "Build"
	creation := &workloadsApi.WorkloadCreationFields{
		Type:               string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		Group:              convutils.Ref("run.ai"),
		Version:            convutils.Ref("v2alpha1"),
		Name:               "old-name",
		Id:                 "old-id",
		SubmittedBy:        convutils.Ref("old-user"),
		ClusterId:          "old-cluster",
		ProjectId:          "old-project",
		CreatedAt:          createdAt.Add(-time.Hour),
		Images:             []string{"jupyter/scipy-notebook"},
		RequestedNodePools: []string{"default"},
	}
	metadata := workloadsServiceConnectionMetadata{
		Environments: []workloadsServerApi.Environment{
			workloadsServerEnvironmentForTest("env", "job-wm-np", "port-8888"),
		},
	}
	workload := &workloadMgrApi.WorkloadV2{
		Metadata: workloadMgrApi.WorkloadV2Metadata{
			Name:           "wm-name",
			ProjectId:      "wm-project",
			Category:       *workloadMgrApi.NewNullableString(&sourceCategory),
			Preemptibility: *workloadMgrApi.NewNullablePreemptibility(&preemptibility),
			Id:             "wm-id",
			Gvk:            workloadMgrApi.GVK{Group: *workloadMgrApi.NewNullableString(convutils.Ref("batch")), Version: "v1", Kind: "Job"},
			ProjectName:    "project-name",
			ClusterId:      "wm-cluster",
			TenantId:       1001,
			DepartmentId:   "department-id",
			DepartmentName: "department-name",
			CreatedAt:      createdAt,
			CreatedBy:      "wm-user",
			UpdatedAt:      createdAt,
			UpdatedBy:      "wm-user",
		},
		DesiredPhase: workloadMgrApi.DESIREDPHASE_RUNNING,
	}

	serverCreation, err := workloadsServiceCreationFieldsForConnectionMetadata(workload, creation, metadata)

	if err != nil {
		t.Fatalf("expected conversion to succeed, got %v", err)
	}
	if serverCreation == nil {
		t.Fatalf("expected fallback creation fields")
	}
	if serverCreation.Id != "wm-id" || serverCreation.Name != "wm-name" {
		t.Fatalf("expected workloads-manager id/name, got id=%q name=%q", serverCreation.Id, serverCreation.Name)
	}
	if serverCreation.Group != "batch" || serverCreation.Version != "v1" {
		t.Fatalf("expected workloads-manager GVK, got %s/%s", serverCreation.Group, serverCreation.Version)
	}
	if serverCreation.ProjectId != "wm-project" || serverCreation.ClusterId != "wm-cluster" {
		t.Fatalf("expected workloads-manager project/cluster, got project=%q cluster=%q", serverCreation.ProjectId, serverCreation.ClusterId)
	}
	if serverCreation.SourceApi == nil || *serverCreation.SourceApi != workloadsServerApi.SOURCEAPI_WORKLOADS_V2 {
		t.Fatalf("expected WorkloadsV2 source api, got %#v", serverCreation.SourceApi)
	}
	if serverCreation.Preemptibility == nil || *serverCreation.Preemptibility != string(workloadMgrApi.PREEMPTIBILITY_NON_PREEMPTIBLE) {
		t.Fatalf("expected preemptibility from workloads-manager, got %#v", serverCreation.Preemptibility)
	}
	if len(serverCreation.Environments) != 1 || len(serverCreation.Environments[0].Connections) != 1 {
		t.Fatalf("expected metadata environments to be copied, got %#v", serverCreation.Environments)
	}
}

func TestNewWorkloadsServiceConnectionMetadataReturnsNilWhenNoMetadataExists(t *testing.T) {
	creation := &workloadsApi.WorkloadCreationFields{
		Type:      string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		Name:      "test",
		Id:        "workloads-service-id",
		ClusterId: "cluster-id",
		ProjectId: "project-id",
		CreatedAt: time.Now(),
	}

	if metadata := newWorkloadsServiceConnectionMetadata(creation); metadata != nil {
		t.Fatalf("expected nil metadata, got %#v", metadata)
	}
}

func TestNewWorkloadsServiceConnectionMetadataReturnsNilWhenMetadataIsEmpty(t *testing.T) {
	creation := &workloadsApi.WorkloadCreationFields{
		Type:                string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		Name:                "test",
		Id:                  "workloads-service-id",
		ClusterId:           "cluster-id",
		ProjectId:           "project-id",
		CreatedAt:           time.Now(),
		Environments:        []workloadsApi.Environment{},
		ExternalConnections: []workloadsApi.Connection{},
	}

	if metadata := newWorkloadsServiceConnectionMetadata(creation); metadata != nil {
		t.Fatalf("expected nil metadata, got %#v", metadata)
	}
}

func TestNewWorkloadsServiceConnectionMetadataReturnsNilForNilCreation(t *testing.T) {
	if metadata := newWorkloadsServiceConnectionMetadata(nil); metadata != nil {
		t.Fatalf("expected nil metadata, got %#v", metadata)
	}
}

func TestComposeDistributedConnectionMetadataCorrelatesMpiWorkerAndLauncher(t *testing.T) {
	framework := workloadsApi.DISTRIBUTEDFRAMEWORK_MPI
	spec := &workloadsApi.SupersetSpec{
		Ports: []workloadsApi.Port{
			connectionMetadataPort(8081, workloadsApi.PORTSERVICETYPE_NODE_PORT),
		},
	}
	spec.DistributedFramework.Set(&framework)
	spec.Image.Set(convutils.Ref("worker-image"))

	masterSpec := &workloadsApi.MasterSpec{
		Ports: []workloadsApi.Port{
			connectionMetadataPort(8082, workloadsApi.PORTSERVICETYPE_NODE_PORT),
		},
	}
	masterSpec.Image.Set(convutils.Ref("launcher-image"))

	notificationsComposer := notifications.NewNotificationsComposer(
		string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		"dist",
		"workload-id",
		"test@example.com",
		"cluster-id",
		"project-id",
		"department-id",
		time.Now(),
		nil,
		composer.FieldsComposerOptions{WorkloadV2NetworkingSupported: true},
	)

	wc := &workloadContext{
		spec:       spec,
		masterSpec: masterSpec,
		create: &workloadCreateContext{
			notificationsComposer: notificationsComposer,
		},
	}

	if metadataErr := composeDistributedConnectionMetadata(context.Background(), wc, "project"); metadataErr != nil {
		t.Fatalf("expected metadata composition to succeed, got %v", metadataErr.Message)
	}
	if notificationsComposer.Output.DistributedFramework == nil ||
		*notificationsComposer.Output.DistributedFramework != string(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI) {
		t.Fatalf("expected normalized MPI framework, got %#v", notificationsComposer.Output.DistributedFramework)
	}

	metadata := newWorkloadsServiceConnectionMetadata(notificationsComposer.Output)
	if metadata == nil {
		t.Fatalf("expected connection metadata")
	}

	workerEnv := environmentByReplicaTypeForTest(metadata.Environments, workloadsApi.REPLICATYPE_WORKER)
	if workerEnv == nil || len(workerEnv.Connections) != 1 {
		t.Fatalf("expected one worker connection, got %#v", workerEnv)
	}
	if workerEnv.Connections[0].GetServiceResourceName() != "kf-dist-np-worker" {
		t.Fatalf("expected worker service kf-dist-np-worker, got %q", workerEnv.Connections[0].GetServiceResourceName())
	}
	if workerEnv.Connections[0].GetPortResourceName() != "port-8081" {
		t.Fatalf("expected worker port port-8081, got %q", workerEnv.Connections[0].GetPortResourceName())
	}

	masterEnv := environmentByReplicaTypeForTest(metadata.Environments, workloadsApi.REPLICATYPE_MASTER)
	if masterEnv == nil || len(masterEnv.Connections) != 1 {
		t.Fatalf("expected one master connection, got %#v", masterEnv)
	}
	if masterEnv.Connections[0].GetServiceResourceName() != "kf-dist-np-launcher" {
		t.Fatalf("expected MPI launcher service kf-dist-np-launcher, got %q", masterEnv.Connections[0].GetServiceResourceName())
	}
	if masterEnv.Connections[0].GetPortResourceName() != "port-8082" {
		t.Fatalf("expected launcher port port-8082, got %q", masterEnv.Connections[0].GetPortResourceName())
	}
}

func TestComposeDistributedConnectionMetadataNormalizesFrameworkForFallback(t *testing.T) {
	frameworks := []workloadsApi.DistributedFramework{
		workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
	}

	for _, framework := range frameworks {
		framework := framework
		t.Run(string(framework), func(t *testing.T) {
			spec := &workloadsApi.SupersetSpec{}
			spec.DistributedFramework.Set(&framework)
			spec.Image.Set(convutils.Ref("worker-image"))

			notificationsComposer := notifications.NewNotificationsComposer(
				string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
				"dist",
				"workload-id",
				"test@example.com",
				"cluster-id",
				"project-id",
				"department-id",
				time.Now(),
				nil,
				composer.FieldsComposerOptions{},
			)

			wc := &workloadContext{
				spec: spec,
				create: &workloadCreateContext{
					notificationsComposer: notificationsComposer,
				},
			}

			if metadataErr := composeDistributedConnectionMetadata(context.Background(), wc, "project"); metadataErr != nil {
				t.Fatalf("expected metadata composition to succeed, got %v", metadataErr.Message)
			}
			if notificationsComposer.Output.DistributedFramework == nil {
				t.Fatal("expected normalized distributed framework")
			}
			if got := *notificationsComposer.Output.DistributedFramework; got != string(framework) {
				t.Fatalf("expected framework %q, got %q", framework, got)
			}
		})
	}
}

func connectionMetadataPort(containerPort int32, serviceType workloadsApi.PortServiceType) workloadsApi.Port {
	port := workloadsApi.Port{}
	port.Container.Set(&containerPort)
	port.ServiceType.Set(&serviceType)
	return port
}

func environmentByReplicaTypeForTest(environments []workloadsApi.Environment, replicaType workloadsApi.ReplicaType) *workloadsApi.Environment {
	for idx := range environments {
		if environments[idx].ReplicaType.Get() != nil && *environments[idx].ReplicaType.Get() == replicaType {
			return &environments[idx]
		}
	}
	return nil
}

func workloadsServerEnvironmentForTest(envName string, serviceName string, portName string) workloadsServerApi.Environment {
	return workloadsServerApi.Environment{
		Name: envName,
		Connections: []workloadsServerApi.Connection{
			{
				Name:                "Jupyter",
				ConnectionType:      "NodePort",
				ServiceResourceName: convutils.Ref(serviceName),
				PortResourceName:    convutils.Ref(portName),
			},
		},
	}
}
