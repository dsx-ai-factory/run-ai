/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"fmt"
	"net/http"
	"time"

	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"

	"github.com/stretchr/testify/mock"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"

	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	workload_settings "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/settings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	policyService "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	setting_mocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("distributed service validations", func() {
	var ctx context.Context
	var randoms *testutils.RandomValues

	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}

	distributedService := NewDistributedService(mockTenantSettingsClient)
	trainingsService := NewTrainingsService(mockTenantSettingsClient)

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		policyMockUpdateTime := time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
		authMock.AllowedProjectsToContext(&ctx)

		orgunit.SetMock(authMock.OrgUnitMock())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	Context("create", func() {
		It("get should be sensitive to workload type", func() {
			request := api.DistributedCreationRequest{
				Name:      "tf-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_TF),
					Image:                convutils.Ref("ubuntu"),
					Category:             convutils.Ref("Train"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())
			expectWorkloadActions(ctx, createPayload.WorkloadId, "create")

			//  expecting the actual phase to be Creating...
			createPayload.ActualPhase = convutils.Ref(workloadsApi.PHASE_CREATING)

			// get distributed should return the correct workload
			getResponse, _ := distributedService.GetDistributed(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload).To(BeEquivalentTo(createPayload))

			// get workspace should return 404 not found
			getResponse, _ = trainingsService.GetTraining(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("should fail if framework is not available", func() {
			// deps where tf is not available
			frameworkDeps := testing.GetCustomFrameworkDependencies(true, true, true, true, false, false, false, false, true, true)
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedClusterWithCustomFrameworks(randoms.ClusterId, testing.TestDefaultClusterVersion, nil, frameworkDeps)))

			request := api.DistributedCreationRequest{
				Name:      "tf-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_TF),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			createResponse, err := distributedService.CreateDistributed(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			apiErr := createResponse.Body.(*api.Error)
			Expect(apiErr.Message).To(Equal(fmt.Sprintf(services_utils.FrameworkNotAvailableMsg, api.DISTRIBUTEDFRAMEWORK_TF)))
		})
	})

	Context("no master + mpi", func() {
		It("should be rejected, not supported", func() {
			request := api.DistributedCreationRequest{
				Name:      "mpi-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Image:                convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			err := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(err.Message).To(ContainSubstring("either provide a masterSpec or enable masterSpecSameAsMaster"))
		})
	})

	Context("no master + pytorch", func() {
		It("should be ok", func() {
			request := api.DistributedCreationRequest{
				Name:      "pytorch-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))
		})

		It("submits segment size natively when the tenant setting is disabled", func() {
			const topologyName = "segment-topology"
			const nodePoolName = "topology-node-pool"

			// Segmented distributed workloads require cluster >= 2.27 (MinVersionOfSegmentTopology).
			// Stay below the no-native-submission version so submission still routes natively.
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.27.0", nil)))
			checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
				WithNodePoolsWithTopology(randoms.ClusterId, topologyName, nodePoolName).
				WithTopology(topologyName, &clusterApi.NetworkTopology{
					Name:   topologyName,
					Levels: []string{"rack", "block"},
				}))

			cfg := runai_config.Get[*config.ServicesConfig]()
			originalConfig := *cfg
			cfg.AnyWorkloadSubmissionEnabled = false
			cfg.AnyWorkloadDistributedSubmission = true
			DeferCleanup(func() { *cfg = originalConfig })

			settingsManager := setting_mocks.NewSettingsManager(GinkgoT())
			settingsManager.On(
				"Get",
				mock.Anything,
				randoms.TenantIdStr,
				(&workload_settings.AnyWorkloadSubmissionEnabled{}).Key(),
			).Return(&workload_settings.AnyWorkloadSubmissionEnabled{Enabled: false}, nil).Once()
			setting_pkg.SetMock(settingsManager)
			DeferCleanup(func() { setting_pkg.SetMock(nil) })

			request := api.DistributedCreationRequest{
				Name:      "segmented-pytorch",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(4)),
					SegmentSize:          convutils.Ref(int32(2)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
					NodePools:            convutils.Ref([]string{nodePoolName}),
				},
			}

			createResponse, err := distributedService.CreateDistributed(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))
		})

		It("rejects non-PyTorch segment configuration before checking segment capability", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.26.9", nil)))
			checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

			request := api.DistributedCreationRequest{
				Name:      "segmented-tf",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(4)),
					SegmentSize:          convutils.Ref(int32(2)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_TF),
					Image:                convutils.Ref("ubuntu"),
				},
			}

			createResponse, err := distributedService.CreateDistributed(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			apiErr := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(apiErr.Message).To(ContainSubstring("supported only for PyTorch distributed workloads"))
			Expect(apiErr.Message).NotTo(ContainSubstring("cluster version"))
		})
	})

	Context("master + worker", func() {
		It("should be ok", func() {
			request := api.DistributedCreationRequest{
				Name:      "pytorch-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("redhat"),
				},
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			getResponse, _ := distributedService.GetDistributed(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload).To(Equal(payloadCreate))
		})
	})

	Context("master + same as", func() {
		It("should be rejected", func() {
			request := api.DistributedCreationRequest{
				Name:      "pytorch-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			err := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(err.Message).To(ContainSubstring("masterSpec should not be provided when masterSpecSameAsMaster is requested"))
		})
	})

	Context("master and worker / master same as / no master + distributed policy", func() {
		It("should take defaults from the propert replica", func() {
			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveDistributedPolicyV2: &policyService.DistributedPolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableDistributedPolicyDefaultsV2(
							&policyService.DistributedPolicyDefaultsV2{
								Master: *policyService.NewNullableReplicaDefaultsV2(&policyService.ReplicaDefaultsV2{
									Command: *policyService.NewNullableString(convutils.Ref("command-master")),
									Args:    *policyService.NewNullableString(convutils.Ref("args-master")),
								}),
								Worker: *policyService.NewNullableDistributedPolicyDefaultsV2Worker(&policyService.DistributedPolicyDefaultsV2Worker{
									Command:    *policyService.NewNullableString(convutils.Ref("command-worker")),
									WorkingDir: *policyService.NewNullableString(convutils.Ref("directory-worker")),
								}),
							}),
					},
				},
			})

			//
			//   when specifying worker and master policy defaults are taken from each replica respecively
			//
			requestNotSame := api.DistributedCreationRequest{
				Name:      "pytorch-2",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, requestNotSame)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			specWorkload := createResponse.Body.(workloadsApi.SupersetWorkload)

			Expect(convutils.Deref(specWorkload.Spec.Command.Get(), "")).To(Equal("command-worker"))
			Expect(specWorkload.Spec.Args.Get()).To(BeNil())
			Expect(convutils.Deref(specWorkload.Spec.WorkingDir.Get(), "")).To(Equal("directory-worker"))

			Expect(specWorkload.MasterSpec).NotTo(BeNil())
			Expect(convutils.Deref(specWorkload.MasterSpec.Get().Command.Get(), "")).To(Equal("command-master"))
			Expect(convutils.Deref(specWorkload.MasterSpec.Get().Args.Get(), "")).To(Equal("args-master"))
			Expect(specWorkload.MasterSpec.Get().WorkingDir.Get()).To(BeNil())

			//
			//   when doing worker-same-as-master the defaults are taken from the worker for both master AND worker
			//
			requestSameAs := api.DistributedCreationRequest{
				Name:      "pytorch-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}
			createResponse, _ = distributedService.CreateDistributed(ctx, requestSameAs)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			specWorkload = createResponse.Body.(workloadsApi.SupersetWorkload)

			Expect(convutils.Deref(specWorkload.Spec.Command.Get(), "")).To(Equal("command-worker"))
			Expect(specWorkload.Spec.Args.Get()).To(BeNil())
			Expect(convutils.Deref(specWorkload.Spec.WorkingDir.Get(), "")).To(Equal("directory-worker"))

			Expect(specWorkload.MasterSpec).NotTo(BeNil())
			Expect(convutils.Deref(specWorkload.MasterSpec.Get().Command.Get(), "")).To(Equal("command-worker"))
			Expect(specWorkload.MasterSpec.Get().Args.Get()).To(BeNil())
			Expect(convutils.Deref(specWorkload.MasterSpec.Get().WorkingDir.Get(), "")).To(Equal("directory-worker"))

			//
			//   when doing no-master we have only worker
			//
			requestNoMaster := api.DistributedCreationRequest{
				Name:      "pytorch-3",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
			}
			createResponse, _ = distributedService.CreateDistributed(ctx, requestNoMaster)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			specWorkload = createResponse.Body.(workloadsApi.SupersetWorkload)

			Expect(convutils.Deref(specWorkload.Spec.Command.Get(), "")).To(Equal("command-worker"))
			Expect(specWorkload.Spec.Args.Get()).To(BeNil())
			Expect(convutils.Deref(specWorkload.Spec.WorkingDir.Get(), "")).To(Equal("directory-worker"))

			Expect(specWorkload.MasterSpec.Get()).To(BeNil())
		})
	})

	Context("master same as", func() {
		It("defaults taken from the worker policy", func() {
			request := api.DistributedCreationRequest{
				Name:      "pytorch-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveDistributedPolicyV2: &policyService.DistributedPolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableDistributedPolicyDefaultsV2(
							&policyService.DistributedPolicyDefaultsV2{
								Worker: *policyService.NewNullableDistributedPolicyDefaultsV2Worker(&policyService.DistributedPolicyDefaultsV2Worker{
									Command: *policyService.NewNullableString(convutils.Ref("command-from-worker")),
								}),
							}),
					},
				},
			})

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(*payload.MasterSpec.Get().Command.Get()).To(Equal("command-from-worker"))
		})
	})

	Context("nodepools -> worker != master", func() {
		It("two different list", func() {
			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
				WithNodePools(randoms.ClusterId, "np-1", "np-2", "np-3"))

			request := api.DistributedCreationRequest{
				Name:      "pytorch-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
					NodePools: &[]string{
						"np-1",
						"np-3",
					},
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
					NodePools: &[]string{
						"np-2",
						"np-3",
					},
				},
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			Expect(createResponse.Body.(*workloadsApi.SubmissionErrorV2).Message).To(ContainSubstring(ErrNodePoolsDiff))
		})

		It("same list, order difference is ok", func() {
			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
				WithNodePools(randoms.ClusterId, "np-1", "np-2", "np-3"))

			request := api.DistributedCreationRequest{
				Name:      "pytorch-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
					NodePools: &[]string{
						"np-1",
						"np-3",
					},
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
					NodePools: &[]string{
						"np-3",
						"np-1",
					},
				},
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))
		})

		It("master node pools not provided this is ok", func() {
			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
				WithNodePools(randoms.ClusterId, "np-1", "np-2", "np-3"))

			request := api.DistributedCreationRequest{
				Name:      "pytorch-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
					NodePools: &[]string{
						"np-1",
						"np-3",
					},
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))
		})

		It("both empty", func() {
			request := api.DistributedCreationRequest{
				Name:      "pytorch-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))
		})
	})

	Context("masterGpu != 0 in PyTorch", func() {
		It("allowed", func() {
			request := api.DistributedCreationRequest{
				Name:      "pytorch-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuPortionRequest: convutils.Ref(0.5),
						GpuPortionLimit:   convutils.Ref(0.8),
					},
				},
			}

			mockTenantSettingsClient.On("GetGpuResourceOptimization", mock.Anything).Return(true, nil).Once()

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(*payload.MasterSpec.Get().Compute.Get().GpuPortionRequest.Get()).To(Equal(0.5))
			Expect(*payload.MasterSpec.Get().Compute.Get().GpuPortionLimit.Get()).To(Equal(0.8))
		})
	})

	Context("masterGpu != 0 in MPI", func() {
		It("forbidden", func() {
			request := api.DistributedCreationRequest{
				Name:      "mpi-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(5)),
					},
				},
			}

			mockTenantSettingsClient.On("GetGpuResourceOptimization", mock.Anything).Return(true, nil).Once()

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			submissionErr := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(submissionErr.Message).To(ContainSubstring("gpuDevicesRequest should not be provided for master replica of MPI workload"))

			request.MasterSpec = &api.MasterSpec{
				Image: convutils.Ref("ubuntu"),
				Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuRequestType: convutils.Ref(api.GPUREQUESTTYPE_MEMORY),
				},
			}

			createResponse, _ = distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			submissionErr = createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(submissionErr.Message).To(ContainSubstring("gpuRequestType should not be provided for master replica of MPI workload"))

			request.MasterSpec = &api.MasterSpec{
				Image: convutils.Ref("ubuntu"),
				Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuPortionRequest: convutils.Ref(0.3),
				},
			}

			createResponse, _ = distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			submissionErr = createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(submissionErr.Message).To(ContainSubstring("gpuMemoryPortionRequest should not be provided for master replica of MPI workload"))

			request.MasterSpec = &api.MasterSpec{
				Image: convutils.Ref("ubuntu"),
				Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuPortionLimit: convutils.Ref(0.3),
				},
			}

			createResponse, _ = distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			submissionErr = createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(submissionErr.Message).To(ContainSubstring("gpuMemoryPortionLimit should not be provided for master replica of MPI workload"))

			request.MasterSpec = &api.MasterSpec{
				Image: convutils.Ref("ubuntu"),
				Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuMemoryRequest: convutils.Ref("1G"),
				},
			}

			createResponse, _ = distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			submissionErr = createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(submissionErr.Message).To(ContainSubstring("gpuMemoryRequest should not be provided for master replica of MPI workload"))

			request.MasterSpec = &api.MasterSpec{
				Image: convutils.Ref("ubuntu"),
				Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuMemoryLimit: convutils.Ref("1G"),
				},
			}

			createResponse, _ = distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			submissionErr = createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(submissionErr.Message).To(ContainSubstring("gpuMemoryLimit should not be provided for master replica of MPI workload"))
		})
	})

	Context("masterGpu != 0 in MPI workloads -> master same as worker", func() {
		It("non relevant fields cleared - portion", func() {
			request := api.DistributedCreationRequest{
				Name:      "mpi-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Image:                convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(5)),
						GpuRequestType:    convutils.Ref(api.GPUREQUESTTYPE_PORTION),
						GpuPortionRequest: convutils.Ref(0.5),
						GpuPortionLimit:   convutils.Ref(0.6),
					},
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			mockTenantSettingsClient.On("GetGpuResourceOptimization", mock.Anything).Return(true, nil).Once()

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			workload := createResponse.Body.(workloadsApi.SupersetWorkload)

			Expect(workload.Spec).NotTo(BeNil())
			Expect(workload.Spec.Compute.Get()).NotTo(BeNil())
			Expect(*workload.Spec.Compute.Get().GpuDevicesRequest.Get()).To(Equal(int32(5)))
			Expect(*workload.Spec.Compute.Get().GpuRequestType.Get()).To(Equal(workloadsApi.GPUREQUESTTYPE_PORTION))
			Expect(*workload.Spec.Compute.Get().GpuPortionRequest.Get()).To(Equal(0.5))
			Expect(*workload.Spec.Compute.Get().GpuPortionLimit.Get()).To(Equal(0.6))

			Expect(workload.MasterSpec.Get()).NotTo(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get()).NotTo(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuDevicesRequest.Get()).To(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuRequestType.Get()).To(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuPortionRequest.Get()).To(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuPortionLimit.Get()).To(BeNil())
		})

		It("non relevant fields cleared - memory", func() {
			request := api.DistributedCreationRequest{
				Name:      "mpi-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Image:                convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(5)),
						GpuRequestType:    convutils.Ref(api.GPUREQUESTTYPE_MEMORY),
						GpuMemoryRequest:  convutils.Ref("1G"),
						GpuMemoryLimit:    convutils.Ref("2G"),
					},
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			mockTenantSettingsClient.On("GetGpuResourceOptimization", mock.Anything).Return(true, nil).Once()

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			workload := createResponse.Body.(workloadsApi.SupersetWorkload)

			Expect(workload.Spec).NotTo(BeNil())
			Expect(workload.Spec.Compute.Get()).NotTo(BeNil())
			Expect(*workload.Spec.Compute.Get().GpuDevicesRequest.Get()).To(Equal(int32(5)))
			Expect(*workload.Spec.Compute.Get().GpuRequestType.Get()).To(Equal(workloadsApi.GPUREQUESTTYPE_MEMORY))
			Expect(*workload.Spec.Compute.Get().GpuMemoryRequest.Get()).To(Equal("1G"))
			Expect(*workload.Spec.Compute.Get().GpuMemoryLimit.Get()).To(Equal("2G"))

			Expect(workload.MasterSpec.Get()).NotTo(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get()).NotTo(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuDevicesRequest.Get()).To(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuRequestType.Get()).To(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuMemoryRequest.Get()).To(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuMemoryLimit.Get()).To(BeNil())
		})
	})

	Context("masterGpu != 0 in MPI workloads -> master same as worker + POLICY defaults", func() {
		It("non relevant fields cleared - portion", func() {
			request := api.DistributedCreationRequest{
				Name:      "mpi-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveDistributedPolicyV2: &policyService.DistributedPolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableDistributedPolicyDefaultsV2(
							&policyService.DistributedPolicyDefaultsV2{
								Master: *policyService.NewNullableReplicaDefaultsV2(&policyService.ReplicaDefaultsV2{
									Command: *policyService.NewNullableString(convutils.Ref("command-master")),
									Args:    *policyService.NewNullableString(convutils.Ref("args-master")),
									Compute: *policyService.NewNullableSupersetDefaultsAllOfCompute(&policyService.SupersetDefaultsAllOfCompute{
										GpuDevicesRequest: *policyService.NewNullableInt32(convutils.Ref(int32(5))),
										GpuRequestType:    *policyService.NewNullableGpuRequestType(convutils.Ref(policyService.GPUREQUESTTYPE_MEMORY)),
										GpuMemoryRequest:  *policyService.NewNullableString(convutils.Ref("1G")),
										GpuMemoryLimit:    *policyService.NewNullableString(convutils.Ref("2G")),
									}),
								}),
								Worker: *policyService.NewNullableDistributedPolicyDefaultsV2Worker(&policyService.DistributedPolicyDefaultsV2Worker{
									Command:    *policyService.NewNullableString(convutils.Ref("command-worker")),
									WorkingDir: *policyService.NewNullableString(convutils.Ref("directory-worker")),
									Compute: *policyService.NewNullableSupersetDefaultsAllOfCompute(&policyService.SupersetDefaultsAllOfCompute{
										GpuDevicesRequest: *policyService.NewNullableInt32(convutils.Ref(int32(6))),
										GpuRequestType:    *policyService.NewNullableGpuRequestType(convutils.Ref(policyService.GPUREQUESTTYPE_MEMORY)),
										GpuMemoryRequest:  *policyService.NewNullableString(convutils.Ref("3G")),
										GpuMemoryLimit:    *policyService.NewNullableString(convutils.Ref("4G")),
									}),
								}),
							}),
					},
				},
			})

			mockTenantSettingsClient.On("GetGpuResourceOptimization", mock.Anything).Return(true, nil).Once()

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			workload := createResponse.Body.(workloadsApi.SupersetWorkload)

			Expect(workload.Spec).NotTo(BeNil())
			Expect(workload.Spec.Compute.Get()).NotTo(BeNil())
			Expect(*workload.Spec.Compute.Get().GpuDevicesRequest.Get()).To(Equal(int32(6)))
			Expect(*workload.Spec.Compute.Get().GpuRequestType.Get()).To(Equal(workloadsApi.GPUREQUESTTYPE_MEMORY))
			Expect(*workload.Spec.Compute.Get().GpuMemoryRequest.Get()).To(Equal("3G"))
			Expect(*workload.Spec.Compute.Get().GpuMemoryLimit.Get()).To(Equal("4G"))

			Expect(workload.MasterSpec.Get()).NotTo(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get()).NotTo(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuDevicesRequest.Get()).To(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuRequestType.Get()).To(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuPortionRequest.Get()).To(BeNil())
			Expect(workload.MasterSpec.Get().Compute.Get().GpuPortionLimit.Get()).To(BeNil())
		})
	})

	Context("delete", func() {
		var workloadId string

		BeforeEach(func() {
			request := api.DistributedCreationRequest{
				Name:      "xg-boost-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_XG_BOOST),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId = payloadCreate.WorkloadId
		})

		It("should be reflected in get", func() {
			// desired state -> running, actual phase -> creating
			getResponse, _ := distributedService.GetDistributed(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_CREATING))

			// simulate real-life time b/w creating and deleting
			time.Sleep(time.Millisecond * 200)

			phaseResponse, _ := distributedService.DeleteDistributed(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusAccepted))

			getResponse, _ = distributedService.GetDistributed(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload = getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DeletedAt).To(Not(BeNil()))
		})

		It("should fail if framework is not available", func() {
			// clear config caches before applying new config
			checker.InitClustersCache(testing.GetDefaultClustersCacheConfig())
			// deps where pytorch is not available and not suspendable
			frameworkDeps := testing.GetCustomFrameworkDependencies(true, true, true, true, true, true, false, false, false, false)
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedClusterWithCustomFrameworks(randoms.ClusterId, testing.TestDefaultClusterVersion, nil, frameworkDeps)))

			deleteResponse, err := distributedService.DeleteDistributed(ctx, workloadId)
			Expect(err).NotTo(HaveOccurred())
			Expect(deleteResponse.Code).To(Equal(http.StatusBadRequest))
			apiErr := deleteResponse.Body.(*api.Error)
			Expect(apiErr.Message).To(Equal(fmt.Sprintf(services_utils.FrameworkNotAvailableMsg, api.DISTRIBUTEDFRAMEWORK_XG_BOOST)))
		})
	})

	Context("delete or get from forbidden project", func() {
		It("should fail", func() {
			request := api.DistributedCreationRequest{
				Name:      "tf-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_TF),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			// make the project forbidden (update all mocks)
			randoms.Forbidden()
			authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
			authMock.AllowedProjectsToContext(&ctx)
			orgunit.SetMock(authMock.OrgUnitMock())

			// get should fail with 404
			getResponse, _ := distributedService.GetDistributed(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))

			// delete should fail with 403
			deleteResponse, _ := distributedService.DeleteDistributed(ctx, workloadId)
			Expect(deleteResponse.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("suspend/resume", func() {
		var workloadId string

		BeforeEach(func() {
			request := api.DistributedCreationRequest{
				Name:      "pytorch-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					NumWorkers:           convutils.Ref(int32(1)),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("redhat"),
				},
			}

			createResponse, _ := distributedService.CreateDistributed(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId = payloadCreate.WorkloadId
		})

		Context("success", func() {
			It("suspend successful", func() {
				nowPlus1Min := time.Now().Add(1 * time.Minute)
				dbErr := dao.DbUpdateWorkloadPhase(ctx, randoms.TenantId, workloadId, api.WorkloadPatchFields{Phase: api.PHASE_RUNNING}, nowPlus1Min)
				Expect(dbErr).To(BeNil())

				// simulate real-life time b/w creating and suspending
				time.Sleep(time.Millisecond * 200)

				suspendResponse, err := distributedService.SuspendDistributed(ctx, workloadId)
				Expect(err).NotTo(HaveOccurred())
				Expect(suspendResponse.Code).To(Equal(http.StatusAccepted))

				getResponse, _ := distributedService.GetDistributed(ctx, workloadId)
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
				Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_STOPPED))
				Expect(getPayload.ActualPhase).NotTo(BeNil())
				Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_STOPPING))
			})

			It("resume successful", func() {
				nowPlus1Min := time.Now().Add(1 * time.Minute)
				dbErr := dao.DbUpdateWorkloadPhase(ctx, randoms.TenantId, workloadId, api.WorkloadPatchFields{Phase: api.PHASE_STOPPED}, nowPlus1Min)
				Expect(dbErr).To(BeNil())

				// simulate real-life time b/w creating and suspending
				time.Sleep(time.Millisecond * 200)

				suspendResponse, err := distributedService.ResumeDistributed(ctx, workloadId)
				Expect(err).NotTo(HaveOccurred())
				Expect(suspendResponse.Code).To(Equal(http.StatusAccepted))

				getResponse, _ := distributedService.GetDistributed(ctx, workloadId)
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
				Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
				Expect(getPayload.ActualPhase).NotTo(BeNil())
				Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_RESUMING))
			})
		})

		Context("failure", func() {
			BeforeEach(func() {
				// clear config caches before applying new config
				checker.InitClustersCache(testing.GetDefaultClustersCacheConfig())
			})

			DescribeTable("suspend fails if",
				func(available bool, expectedMsg string) {
					// deps where pytorch is not available and not suspendable
					frameworkDeps := testing.GetCustomFrameworkDependencies(true, true, available, false, true, false, false, true, true, true)
					checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
						WithCluster(randoms.ClusterId, testing.BuildDisplayedClusterWithCustomFrameworks(randoms.ClusterId, testing.TestDefaultClusterVersion, nil, frameworkDeps)))

					nowPlus1Min := time.Now().Add(1 * time.Minute)
					dbErr := dao.DbUpdateWorkloadPhase(ctx, randoms.TenantId, workloadId, api.WorkloadPatchFields{Phase: api.PHASE_RUNNING}, nowPlus1Min)
					Expect(dbErr).To(BeNil())

					suspendResponse, err := distributedService.SuspendDistributed(ctx, workloadId)
					Expect(err).NotTo(HaveOccurred())
					Expect(suspendResponse.Code).To(Equal(http.StatusBadRequest))
					apiErr := suspendResponse.Body.(*api.Error)
					Expect(apiErr.Message).To(Equal(fmt.Sprintf(expectedMsg, api.DISTRIBUTEDFRAMEWORK_PY_TORCH)))
				},
				Entry("framework not available", false, services_utils.FrameworkNotAvailableMsg),
				Entry("framework not suspendable", true, services_utils.FrameworkNotSuspendableMsg),
			)

			DescribeTable("resume fails if",
				func(available bool, expectedMsg string) {
					// deps where pytorch is not available and not suspendable
					frameworkDeps := testing.GetCustomFrameworkDependencies(true, true, available, false, false, false, true, true, true, true)
					checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
						WithCluster(randoms.ClusterId, testing.BuildDisplayedClusterWithCustomFrameworks(randoms.ClusterId, testing.TestDefaultClusterVersion, nil, frameworkDeps)))

					nowPlus1Min := time.Now().Add(1 * time.Minute)
					dbErr := dao.DbUpdateWorkloadPhase(ctx, randoms.TenantId, workloadId, api.WorkloadPatchFields{Phase: api.PHASE_STOPPED}, nowPlus1Min)
					Expect(dbErr).To(BeNil())

					// simulate real-life time b/w creating and suspending
					time.Sleep(time.Millisecond * 200)

					resumeResponse, err := distributedService.ResumeDistributed(ctx, workloadId)
					Expect(err).NotTo(HaveOccurred())
					Expect(resumeResponse.Code).To(Equal(http.StatusBadRequest))
					apiErr := resumeResponse.Body.(*api.Error)
					Expect(apiErr.Message).To(Equal(fmt.Sprintf(expectedMsg, api.DISTRIBUTEDFRAMEWORK_PY_TORCH)))
				},
				Entry("framework not available", false, services_utils.FrameworkNotAvailableMsg),
				Entry("framework not suspendable", true, services_utils.FrameworkNotSuspendableMsg),
			)
		})
	})
})
