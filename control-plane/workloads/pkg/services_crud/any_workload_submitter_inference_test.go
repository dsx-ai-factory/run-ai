/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/notifications"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"

	workloadsMgrClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/workloads_mgr"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	testing_pkg "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

type inferenceUpdateWorkloadsManagerMock struct {
	workload            *workloadMgrApi.WorkloadV2GetResponse
	updatedWorkload     *workloadMgrApi.WorkloadV2
	updateErr           *workloadMgrApi.Error
	getCalls            int
	internalUpdateCalls int
	publicUpdateCalls   int
}

func (m *inferenceUpdateWorkloadsManagerMock) BatchWorkloads(_ context.Context, _ workloadMgrApi.WorkloadsV2Batch) (*workloadMgrApi.BatchResponse, *workloadMgrApi.Error) {
	return nil, nil
}

func (m *inferenceUpdateWorkloadsManagerMock) CreateWorkloadV2(_ context.Context, _ workloadMgrApi.WorkloadV2CreateRequest) (*workloadMgrApi.WorkloadV2, *workloadMgrApi.Error) {
	return nil, nil
}

func (m *inferenceUpdateWorkloadsManagerMock) GetWorkloadV2ById(_ context.Context, _ string) (*workloadMgrApi.WorkloadV2GetResponse, *workloadMgrApi.Error) {
	m.getCalls++
	return m.workload, nil
}

func (m *inferenceUpdateWorkloadsManagerMock) UpdateWorkloadV2(_ context.Context, _ string, _ workloadMgrApi.WorkloadV2UpdateRequest) (*workloadMgrApi.WorkloadV2, *workloadMgrApi.Error) {
	m.publicUpdateCalls++
	return nil, nil
}

func (m *inferenceUpdateWorkloadsManagerMock) UpdateWorkloadInternal(_ context.Context, _ string, _ workloadMgrApi.WorkloadV2UpdateRequest) (*workloadMgrApi.WorkloadV2, *workloadMgrApi.Error) {
	m.internalUpdateCalls++
	return m.updatedWorkload, m.updateErr
}

var _ = Describe("sendUpdateToWorkloadMgr", func() {
	It("returns the populated response from the internal workload update endpoint", func() {
		updatedWorkload := &workloadMgrApi.WorkloadV2{
			Metadata: workloadMgrApi.WorkloadV2Metadata{Id: "workload-id"},
		}
		mgr := &inferenceUpdateWorkloadsManagerMock{updatedWorkload: updatedWorkload}
		originalWorkloadsManager := workloadsMgrClient.Get()
		workloadsMgrClient.SetMock(mgr)
		DeferCleanup(func() { workloadsMgrClient.SetMock(originalWorkloadsManager) })

		actual, updateErr := sendUpdateToWorkloadMgr(context.Background(), "workload-id", &Update{})

		Expect(updateErr).To(BeNil())
		Expect(actual).To(BeIdenticalTo(updatedWorkload))
		Expect(mgr.internalUpdateCalls).To(Equal(1))
		Expect(mgr.publicUpdateCalls).To(BeZero())
	})

	It("rejects a successful update without a workload response", func() {
		mgr := &inferenceUpdateWorkloadsManagerMock{}
		originalWorkloadsManager := workloadsMgrClient.Get()
		workloadsMgrClient.SetMock(mgr)
		DeferCleanup(func() { workloadsMgrClient.SetMock(originalWorkloadsManager) })

		actual, updateErr := sendUpdateToWorkloadMgr(context.Background(), "workload-id", &Update{})

		Expect(actual).To(BeNil())
		Expect(updateErr).NotTo(BeNil())
		Expect(int(updateErr.Code)).To(Equal(http.StatusServiceUnavailable))
	})
})

var _ = Describe("buildInferenceWorkloadV2Metadata", func() {
	It("passes scheduling metadata to workloads-manager", func() {
		name := "inference"
		projectID := "project-a"
		category := "Deploy"
		priority := "very-high"
		preemptibility := workloadsApi.PREEMPTIBILITY_PREEMPTIBLE
		spec := &workloadsApi.InferenceSpecSpec{}
		spec.SetCategory(category)
		spec.SetPriorityClass(priority)
		spec.SetPreemptibility(preemptibility)
		configuration := &workloadMgrApi.WorkloadV2Configuration{}

		metadata := buildInferenceWorkloadV2Metadata(&name, &projectID, configuration, spec)

		Expect(*metadata.Name).To(Equal(name))
		Expect(*metadata.ProjectId).To(Equal(projectID))
		Expect(metadata.Configuration).To(BeIdenticalTo(configuration))
		Expect(*metadata.Category.Get()).To(Equal(category))
		Expect(*metadata.Priority.Get()).To(Equal(priority))
		Expect(*metadata.Preemptibility.Get()).To(Equal(workloadMgrApi.PREEMPTIBILITY_PREEMPTIBLE))
	})
})

var _ = Describe("AnyWorkloadSubmitterInference.Create", Serial, func() {
	It("returns serving-port and exposed-URL connection metadata", func() {
		const clusterID = "inference-connections-cluster"
		displayedCluster := testing_pkg.BuildDisplayedCluster(clusterID, "3.0.0", nil)
		displayedCluster.SetDomain("https://mock-q7z3k.example.com")
		clusterConfig := displayedCluster.Status.Get().GetConfig()
		clusterConfig.SetRunaiOauth2ProxyImage("test.registry/oauth2-proxy:test")
		displayedCluster.Status.Get().SetConfig(clusterConfig)
		originalCapabilitiesClient := checkerClusterClient.Get()
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().WithCluster(clusterID, displayedCluster))
		checker.InitClustersCache(testing_pkg.GetDefaultClustersCacheConfig())
		DeferCleanup(func() {
			checkerClusterClient.SetMock(originalCapabilitiesClient)
			checker.InitClustersCache(testing_pkg.GetDefaultClustersCacheConfig())
		})

		originalClusterService := clusterservice.Get()
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithCluster(clusterID, &displayedCluster))
		DeferCleanup(func() { clusterservice.SetMock(originalClusterService) })

		servingPort := workloadsApi.NewSupersetServingPort()
		servingPort.SetContainer(8888)
		exposedURL := workloadsApi.NewExposedUrl()
		exposedURL.SetContainer(7860)
		exposedURL.SetToolType("gradio")
		exposedURL.SetToolName("chat")
		spec := workloadsApi.NewSupersetSpec()
		spec.SetImage("jupyter/base-notebook:latest")
		spec.SetServingPort(*servingPort)
		spec.SetExposedUrls([]workloadsApi.ExposedUrl{*exposedURL})
		inferenceSpec := &workloadsApi.InferenceSpecSpec{}
		Expect(convutils.CopySpecs(inferenceSpec, spec)).To(Succeed())

		notificationsComposer := notifications.NewNotificationsComposer(
			string(workloadsApi.WORKLOADTYPE_INFERENCE),
			"inference",
			"workload-id",
			"test@example.com",
			clusterID,
			"project-id",
			"department-id",
			time.Now(),
			nil,
			composer.FieldsComposerOptions{},
		)
		params := &createWorkloadParams{Response: workloadsApi.SupersetWorkload{
			Name:      "inference",
			ProjectId: "project-id",
			ClusterId: clusterID,
			Spec:      spec,
		}}
		wc := &workloadContext{
			workloadType:        string(workloadsApi.WORKLOADTYPE_INFERENCE),
			workloadName:        "inference",
			createdBy:           "test@example.com",
			spec:                spec,
			inferenceSpec:       inferenceSpec,
			clusterCapabilities: checker.NewClusterCapabilitiesChecker(clusterID),
			create: &workloadCreateContext{
				createParams:          params,
				project:               &orgUnitApi.Project{Name: "project"},
				namespace:             "runai-project",
				notificationsComposer: notificationsComposer,
			},
		}

		submission, submissionErr := (&AnyWorkloadSubmitterInference{}).Create(context.Background(), wc)

		Expect(submissionErr).To(BeNil())
		Expect(submission).NotTo(BeNil())
		Expect(submission.ConnectionMetadata).NotTo(BeNil())
		Expect(submission.ConnectionCreationFields).To(BeIdenticalTo(notificationsComposer.Output))
		Expect(submission.ConnectionMetadata.Environments).To(HaveLen(1))
		connections := submission.ConnectionMetadata.Environments[0].Connections
		Expect(connections).To(HaveLen(2))

		var servingPortConnection, exposedURLConnection *workloadsApi.Connection
		for idx := range connections {
			switch connections[idx].ConnectionType {
			case "ServingPort":
				servingPortConnection = &connections[idx]
			case "ExternalUrl":
				exposedURLConnection = &connections[idx]
			}
		}

		Expect(servingPortConnection).NotTo(BeNil())
		Expect(servingPortConnection.GetContainerPort()).To(Equal(int32(8888)))
		Expect(servingPortConnection.GetCorrelationName()).To(Equal("serving-port"))
		Expect(exposedURLConnection).NotTo(BeNil())
		Expect(exposedURLConnection.GetContainerPort()).To(Equal(int32(7860)))
		Expect(exposedURLConnection.GetCorrelationName()).To(Equal("url"))
		Expect(exposedURLConnection.GetServiceResourceName()).To(Equal("infw-inference-url"))
	})
})

var _ = Describe("AnyWorkloadSubmitterInference.CanUpdate", func() {
	const tenantID = int32(1)

	// CanUpdate no longer queries workloads-manager. The caller (UpdateInferenceSpec) already
	// resolved the workload, its cluster and the mode it was created in; CanUpdate only verifies
	// that the workload's original submission mode is still supported by the cluster today.
	newCapabilityChecker := func(clusterID, version string) *checker.ClusterCapabilitiesChecker {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(clusterID, testing_pkg.BuildDisplayedCluster(clusterID, version, nil)))
		checker.InitClusterCapabilitiesChecker("", testing_pkg.GetDefaultClustersCacheConfig(), "")
		return checker.NewClusterCapabilitiesChecker(clusterID)
	}

	BeforeEach(func() {
		// Keep the deployment-level any-workload gate disabled so the 2.6-style "both supported"
		// branch does not reach the per-tenant setting lookup (which has no mock in these tests).
		cfg := runai_config.Get[*config.ServicesConfig]()
		originalConfig := *cfg
		cfg.AnyWorkloadSubmissionEnabled = false
		DeferCleanup(func() { *cfg = originalConfig })
	})

	It("returns an internal error when the cluster capabilities are unavailable", func() {
		wc := &workloadContext{workloadType: Inference, tenantId: tenantID}

		err := (&AnyWorkloadSubmitterInference{}).
			CanUpdate(context.Background(), SubmissionMethodNative, wc, "workload-id")

		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusInternalServerError))
	})

	It("allows updating a native workload while the cluster still supports native submission", func() {
		wc := &workloadContext{
			workloadType:        Inference,
			tenantId:            tenantID,
			clusterCapabilities: newCapabilityChecker("inference-native-supported", "2.30.0"),
		}
		err := (&AnyWorkloadSubmitterInference{}).
			CanUpdate(context.Background(), SubmissionMethodNative, wc, "workload-id")

		Expect(err).To(BeNil())
	})

	It("blocks updating a native workload once the cluster no longer supports native submission", func() {
		wc := &workloadContext{
			workloadType:        Inference,
			tenantId:            tenantID,
			clusterCapabilities: newCapabilityChecker("inference-native-dropped", "3.0.0"),
		}

		err := (&AnyWorkloadSubmitterInference{}).
			CanUpdate(context.Background(), SubmissionMethodNative, wc, "workload-id")

		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusServiceUnavailable))
	})

	It("allows updating an any-workload workload while the cluster supports any-workload submission", func() {
		wc := &workloadContext{
			workloadType:        Inference,
			tenantId:            tenantID,
			clusterCapabilities: newCapabilityChecker("inference-anyworkload-supported", "3.0.0"),
		}

		err := (&AnyWorkloadSubmitterInference{}).
			CanUpdate(context.Background(), SubmissionMethodAnyWorkload, wc, "workload-id")

		Expect(err).To(BeNil())
	})

	It("blocks updating an any-workload workload when the cluster does not support any-workload submission", func() {
		wc := &workloadContext{
			workloadType:        Inference,
			tenantId:            tenantID,
			clusterCapabilities: newCapabilityChecker("inference-anyworkload-unsupported", "2.22.0"),
		}

		err := (&AnyWorkloadSubmitterInference{}).
			CanUpdate(context.Background(), SubmissionMethodAnyWorkload, wc, "workload-id")

		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusServiceUnavailable))
	})
})
