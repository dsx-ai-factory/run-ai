/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GVK Workload Validation and Mutations", func() {
	var (
		impl          *WorkloadsV2Impl
		ctx           context.Context
		randoms       *testutils.RandomValues
		workloadsMock *workloads.SimpleMock
		jobManifest   map[string]interface{}
		csclient      *clusterservice.SimpleMock
	)

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()

		csclient = clusterservice.NewSimpleMock()
		csclient.WithDefaultNodePool(randoms.ClusterId)
		clusters.SetMock(csclient)

		k8sUtil := NewK8sResourceUtil()
		impl = &WorkloadsV2Impl{
			k8sUtil: k8sUtil,
		}

		ctx = setContext("1", "user-a", []string{})

		authMock.SetNoPermissions().AddPermittedTenants(randoms)
		org_unit.SetMock(org_unit.NewSimpleMock().WithRandoms(randoms))
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)))
		checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

		// Setup workloads client mock
		workloadsMock = workloads.NewSimpleMock().
			WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories)
		workloads.SetMock(workloadsMock)

		// Setup test manifest
		jobManifest = map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "Job",
			"metadata": map[string]interface{}{
				"name": "test-job",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "test-container",
								"image": "nginx",
							},
						},
						"restartPolicy": "Never",
					},
				},
			},
		}

		policies.SetMock(policies.SimpleMockClientFactory{})

		dao.RestartUpdateCountForTest()
	})

	AfterEach(func() {
		cleanTables()
	})

	// Helper function to create a workload type with proper structure
	createWorkloadType := func(name, group string, clusterId, phase, version string, priority, category *string) workloadsApi.WorkloadTypeConfig {
		workloadType := workloadsApi.WorkloadTypeConfig{
			Name:       name,
			Group:      group,
			CategoryId: "default-category",
			PriorityId: "default-priority",
		}

		if priority != nil {
			workloadType.PriorityName = priority
		}

		if category != nil {
			workloadType.CategoryName = category
		}

		if clusterId != "" && phase != "" && version != "" {
			workloadType.ClusterStatuses = []workloadsApi.WorkloadTypeStatus{
				{
					ClusterId: &clusterId,
					Phase:     &phase,
					Version:   &version,
				},
			}
		}

		return workloadType
	}

	Context("validation failures", func() {
		It("should fail when workload type is not found", func() {
			// Don't add any workload types to mock

			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: jobManifest,
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(result).To(BeNil())
			Expect(apiErrs).ToNot(BeEmpty())
			Expect(apiErrs[0].Code).To(Equal(int32(400)))
			Expect(apiErrs[0].Message).To(ContainSubstring("workload type not found"))
			Expect(apiErrs[0].Message).To(ContainSubstring("batch"))
			Expect(apiErrs[0].Message).To(ContainSubstring("Job"))
		})

		It("should fail when workload type has no cluster status", func() {
			// Setup mock with workload type but no cluster statuses
			workloadType := createWorkloadType("Job", "batch", "", "", "", nil, nil)
			workloadsMock.WithWorkloadType(workloadType)

			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: jobManifest,
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(result).To(BeNil())
			Expect(apiErrs).ToNot(BeEmpty())
			Expect(apiErrs[0].Code).To(Equal(int32(400)))
			Expect(apiErrs[0].Message).To(ContainSubstring("no cluster status information"))
		})

		It("should fail when workload type is not ready in target cluster", func() {
			// Setup mock with workload type in failed state
			workloadType := createWorkloadType("Job", "batch", randoms.ClusterId, "Failed", "v1", nil, nil)
			workloadsMock.WithWorkloadType(workloadType)

			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: jobManifest,
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(result).To(BeNil())
			Expect(apiErrs).ToNot(BeEmpty())
			Expect(apiErrs[0].Code).To(Equal(int32(400)))
			Expect(apiErrs[0].Message).To(ContainSubstring("workload type is not ready"))
			Expect(apiErrs[0].Message).To(ContainSubstring(randoms.ClusterId))
			Expect(apiErrs[0].Message).To(ContainSubstring("Failed"))
		})

		It("should fail when workload type is not ready with condition messages", func() {
			// Setup mock with workload type in failed state with conditions
			conditionType := "Ready"
			conditionStatus := "False"
			conditionMessage := "Installation failed due to missing dependencies"

			workloadType := createWorkloadType("Job", "batch", randoms.ClusterId, "Failed", "v1", nil, nil)
			// Add conditions to the cluster status
			workloadType.ClusterStatuses[0].Conditions = []workloadsApi.Condition{
				{
					Type:    conditionType,
					Status:  conditionStatus,
					Message: &conditionMessage,
				},
			}
			workloadsMock.WithWorkloadType(workloadType)

			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: jobManifest,
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(result).To(BeNil())
			Expect(apiErrs).ToNot(BeEmpty())
			Expect(apiErrs[0].Code).To(Equal(int32(400)))
			Expect(apiErrs[0].Message).To(ContainSubstring("workload type is not ready"))
			Expect(apiErrs[0].Message).To(ContainSubstring("Failed"))
			Expect(apiErrs[0].Message).To(ContainSubstring(conditionMessage))
		})

		It("should fail when workload type exists in different cluster but not target cluster", func() {
			// Setup mock with workload type ready in different cluster
			workloadType := createWorkloadType("Job", "batch", uuid.NewString(), "Ready", "v1", nil, nil)
			workloadsMock.WithWorkloadType(workloadType)

			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: jobManifest,
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(result).To(BeNil())
			Expect(apiErrs).ToNot(BeEmpty())
			Expect(apiErrs[0].Code).To(Equal(int32(400)))
			Expect(apiErrs[0].Message).To(ContainSubstring("workload type has no status information for cluster"))
			Expect(apiErrs[0].Message).To(ContainSubstring(randoms.ClusterId))
			Expect(apiErrs[0].Message).To(ContainSubstring("version v1"))
		})
	})

	Context("default mutations", func() {
		BeforeEach(func() {
			// Setup mock with ready workload type that has defaults
			defaultPriority := "high"
			defaultCategory := "training"
			workloadType := createWorkloadType("Job", "batch", randoms.ClusterId, "Ready", "v1", &defaultPriority, &defaultCategory)
			workloadsMock.WithWorkloadType(workloadType)
		})

		It("should apply defaults when not provided", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: jobManifest,
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(apiErrs).To(BeEmpty())
			Expect(result).ToNot(BeNil())
			Expect(result.Metadata.Priority).ToNot(BeNil())
			Expect(*result.Metadata.Priority).To(Equal("high"))
			Expect(result.Metadata.Category).ToNot(BeNil())
			Expect(*result.Metadata.Category).To(Equal("training"))

			// Verify it's applied to the manifest labels
			workloadMetadata := result.WorkloadManifest["metadata"].(map[string]interface{})
			labels := workloadMetadata["labels"].(map[string]interface{})
			Expect(labels[PriorityLabel]).To(Equal("high"))
			Expect(labels[CategoryLabel]).To(Equal("training"))
		})

		It("should not override explicitly provided priority", func() {
			explicitPriority := "low"
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					Priority:  &explicitPriority,
				},
				Manifest: jobManifest,
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(apiErrs).To(BeEmpty())
			Expect(result).ToNot(BeNil())
			Expect(result.Metadata.Priority).ToNot(BeNil())
			Expect(*result.Metadata.Priority).To(Equal("low"))

			// Verify it's applied to the manifest labels
			workloadMetadata := result.WorkloadManifest["metadata"].(map[string]interface{})
			labels := workloadMetadata["labels"].(map[string]interface{})
			Expect(labels[PriorityLabel]).To(Equal("low"))
		})

		It("should handle empty string priority as not provided", func() {
			emptyPriority := ""
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					Priority:  &emptyPriority,
				},
				Manifest: jobManifest,
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(apiErrs).To(BeEmpty())
			Expect(result).ToNot(BeNil())
			Expect(result.Metadata.Priority).ToNot(BeNil())
			Expect(*result.Metadata.Priority).To(Equal("high"))
		})

		It("should not apply defaults when workload type has no defaults", func() {
			// Setup mock with workload type without defaults
			workloadTypeNoDefaults := createWorkloadType("Job", "batch", randoms.ClusterId, "Ready", "v1", nil, nil)
			workloads.SetMock(workloads.NewSimpleMock().WithWorkloadType(workloadTypeNoDefaults))

			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: jobManifest,
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(apiErrs).To(BeEmpty())
			Expect(result).ToNot(BeNil())
			Expect(result.Metadata.Priority).To(BeNil())
			Expect(result.Metadata.Category).ToNot(BeNil())
			Expect(*result.Metadata.Category).To(Equal(DefaultCategory))
		})
	})
})
