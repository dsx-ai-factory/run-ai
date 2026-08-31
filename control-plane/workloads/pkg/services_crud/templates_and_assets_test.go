/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	clientsclusterservice "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/run-ai/runai/control-plane/workloads/mocks/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/assets"
	assetsClient "github.com/run-ai/runai/control-plane/workloads/pkg/clients/assets"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/clusterservice"
	k8sObjectsClient "github.com/run-ai/runai/control-plane/workloads/pkg/clients/k8s-objects-tracker"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/org_unit"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	"github.com/run-ai/runai/control-plane/workloads/pkg/testing_crud"
	assetsApi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	k8sObjectsTrackerMock "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/mocks"
	policyService "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

var _ = Describe("Workload composer validations", func() {
	var ctx context.Context
	var randoms *testutils.RandomValues

	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}

	trainingsService := NewTrainingsService(mockTenantSettingsClient)
	distributedService := NewDistributedService(mockTenantSettingsClient)

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()

		k8sObjectsClient.SetMock(k8sObjectsTrackerMock.NewK8SObjectTrackerMock())

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
		authorization.SetMock(authMock)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))

		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	Context("submit using template", func() {
		It("training", func() {
			//
			//   create the policy
			//
			trainingPolicy := policyService.TrainingPolicyDefaultsAndRulesV2{
				Defaults: *policyService.NewNullableTrainingPolicyDefaultsV2(&policyService.TrainingPolicyDefaultsV2{
					Image: *policyService.NewNullableString(convutils.Ref("ubuntu")),
					EnvironmentVariables: *policyService.NewNullableEnvironmentVariablesDefaults(&policyService.EnvironmentVariablesDefaults{
						Instances: []policyService.EnvironmentVariable{
							{
								Name:  *policyService.NewNullableString(convutils.Ref("env-1")),
								Value: *policyService.NewNullableString(convutils.Ref("env-1-from-policy")),
							},
							{
								Name:  *policyService.NewNullableString(convutils.Ref("env-2")),
								Value: *policyService.NewNullableString(convutils.Ref("env-2-from-policy")),
							},
							{
								Name:  *policyService.NewNullableString(convutils.Ref("env-3")),
								Value: *policyService.NewNullableString(convutils.Ref("env-3-from-policy")),
							},
						},
					}),
				}),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt:        time.Now(),
					EffectiveTrainingPolicyV2: &trainingPolicy,
				},
			})

			//
			//    create the template
			//
			ts := TemplatesService{}

			trainingTemplate := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "ts-1",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.TrainingTemplateSpec{
					Image: convutils.Ref("image-from-template"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:    convutils.Ref("env-1"),
							Exclude: convutils.Ref(true),
						},
						{
							Name:  convutils.Ref("env-2"),
							Value: convutils.Ref("env-2-from-template"),
						},
						{
							Name:  convutils.Ref("env-4"),
							Value: convutils.Ref("env-4-from-template"),
						},
						{
							Name:  convutils.Ref("env-5"),
							Value: convutils.Ref("env-5-from-template"),
						},
					},
				},
			}

			unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_TRAINING, &trainingTemplate)
			Expect(err).To(BeNil())

			err = ts.CreateTemplate(ctx, unifiedRequest)
			Expect(err).To(BeNil())

			//
			//   create the workload
			//
			workloadRequest := api.TrainingCreationRequest{
				Name:       "train-1",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(unifiedRequest.template.Meta.Id),

				Spec: api.TrainingSpecSpec{
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:    convutils.Ref("env-4"),
							Exclude: convutils.Ref(true),
						},
						{
							Name:  convutils.Ref("env-6"),
							Value: convutils.Ref("env-6-from-spec"),
						},
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			workload := response.Body.(workloadsApi.SupersetWorkload)

			expectedWorkloadEnvVars := []api.EnvironmentVariable{
				{
					Name:    convutils.Ref("env-1"),
					Exclude: convutils.Ref(true),
				},
				{
					Name:  convutils.Ref("env-2"),
					Value: convutils.Ref("env-2-from-template"),
				},
				{
					Name:  convutils.Ref("env-3"),
					Value: convutils.Ref("env-3-from-policy"),
				},
				{
					Name:    convutils.Ref("env-4"),
					Exclude: convutils.Ref(true),
				},
				{
					Name:  convutils.Ref("env-5"),
					Value: convutils.Ref("env-5-from-template"),
				},
				{
					Name:  convutils.Ref("env-6"),
					Value: convutils.Ref("env-6-from-spec"),
				},
			}

			testutils.Expect(workload.Spec.EnvironmentVariables).ToEqual(expectedWorkloadEnvVars)
		})

		It("distributed - template containing master + worker", func() {
			//
			//    create the template
			//
			ts := TemplatesService{}

			distributedTemplate := api.DistributedTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "ts-1",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.DistributedTemplateSpec{
					Image:      convutils.Ref("image-template-worker"),
					WorkingDir: convutils.Ref("dir-template-worker"),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Image:      convutils.Ref("image-template-master"),
						WorkingDir: convutils.Ref("dir-template-master"),
					},
				},
			}

			unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, &distributedTemplate)
			Expect(err).To(BeNil())

			err = ts.CreateTemplate(ctx, unifiedRequest)
			Expect(err).To(BeNil())

			//
			//   create the workload - master + spec
			//
			workloadRequest := api.DistributedCreationRequest{
				Name:       "dist-1",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(unifiedRequest.template.Meta.Id),

				Spec: api.DistributedSpecSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("image-workload-worker"),
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("image-workload-master"),
				},
			}

			response, _ := distributedService.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			workload := response.Body.(workloadsApi.SupersetWorkload)

			expectedWorkload := api.SupersetWorkload{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),

					Image:      convutils.Ref("image-workload-worker"),
					WorkingDir: convutils.Ref("dir-template-worker"),
					Category:   convutils.Ref("Train"),
					NodePools: &[]string{
						"default",
					},
				},
				MasterSpec: &api.MasterSpec{
					Image:      convutils.Ref("image-workload-master"),
					WorkingDir: convutils.Ref("dir-template-master"),
					Category:   convutils.Ref("Train"),
					NodePools: &[]string{
						"default",
					},
				},
			}
			testutils.Expect(workload.Spec).ToEqual(expectedWorkload.Spec)
			testutils.Expect(workload.MasterSpec).ToEqual(expectedWorkload.MasterSpec)

			//
			//   create the master same as
			//
			workloadRequest = api.DistributedCreationRequest{
				Name:       "dist-2",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(unifiedRequest.template.Meta.Id),

				Spec: api.DistributedSpecSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("image-workload-worker"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			response, _ = distributedService.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			workload = response.Body.(workloadsApi.SupersetWorkload)

			expectedWorkload = api.SupersetWorkload{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),

					Image:      convutils.Ref("image-workload-worker"),
					WorkingDir: convutils.Ref("dir-template-worker"),
					Category:   convutils.Ref("Train"),
					NodePools: &[]string{
						"default",
					},
				},
				MasterSpec: &api.MasterSpec{
					Image:      convutils.Ref("image-workload-worker"),
					WorkingDir: convutils.Ref("dir-template-worker"),
					Category:   convutils.Ref("Train"),
					NodePools: &[]string{
						"default",
					},
				},
			}
			testutils.Expect(workload.Spec).ToEqual(expectedWorkload.Spec)
			testutils.Expect(workload.MasterSpec).ToEqual(expectedWorkload.MasterSpec)
		})

		It("distributed - template with master same as worker", func() {
			//
			//    create the template
			//
			ts := TemplatesService{}

			distributedTemplate := api.DistributedTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "ts-1",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.DistributedTemplateSpec{
					Image:      convutils.Ref("image-template-worker"),
					WorkingDir: convutils.Ref("dir-template-worker"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, &distributedTemplate)
			Expect(err).To(BeNil())

			err = ts.CreateTemplate(ctx, unifiedRequest)
			Expect(err).To(BeNil())

			//
			//   create the workload - master + spec
			//
			workloadRequest := api.DistributedCreationRequest{
				Name:       "dist-1",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(unifiedRequest.template.Meta.Id),

				Spec: api.DistributedSpecSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("image-workload-worker"),
					Category:             convutils.Ref("Train"),
					NodePools: &[]string{
						"default",
					},
				},
				MasterSpec: &api.MasterSpec{
					Image:    convutils.Ref("image-workload-master"),
					Category: convutils.Ref("Train"),
					NodePools: &[]string{
						"default",
					},
				},
			}

			response, _ := distributedService.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			workload := response.Body.(workloadsApi.SupersetWorkload)

			expectedWorkload := api.SupersetWorkload{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),

					Image:      convutils.Ref("image-workload-worker"),
					WorkingDir: convutils.Ref("dir-template-worker"),
					Category:   convutils.Ref("Train"),
					NodePools: &[]string{
						"default",
					},
				},
				MasterSpec: &api.MasterSpec{
					Image:      convutils.Ref("image-workload-master"),
					WorkingDir: convutils.Ref("dir-template-worker"),
					Category:   convutils.Ref("Train"),
					NodePools: &[]string{
						"default",
					},
				},
			}
			testutils.Expect(workload.Spec).ToEqual(expectedWorkload.Spec)
			testutils.Expect(workload.MasterSpec).ToEqual(expectedWorkload.MasterSpec)

			//
			//   create the master same as
			//
			workloadRequest = api.DistributedCreationRequest{
				Name:       "dist-2",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(unifiedRequest.template.Meta.Id),

				Spec: api.DistributedSpecSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("image-workload-worker"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			response, _ = distributedService.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			workload = response.Body.(workloadsApi.SupersetWorkload)

			expectedWorkload = api.SupersetWorkload{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),

					Image:      convutils.Ref("image-workload-worker"),
					WorkingDir: convutils.Ref("dir-template-worker"),
					Category:   convutils.Ref("Train"),
					NodePools: &[]string{
						"default",
					},
				},
				MasterSpec: &api.MasterSpec{
					Image:      convutils.Ref("image-workload-worker"),
					WorkingDir: convutils.Ref("dir-template-worker"),
					Category:   convutils.Ref("Train"),
					NodePools: &[]string{
						"default",
					},
				},
			}
			testutils.Expect(workload.Spec).ToEqual(expectedWorkload.Spec)
			testutils.Expect(workload.MasterSpec).ToEqual(expectedWorkload.MasterSpec)
		})
	})

	Context("assets validity in submitting workloads", func() {
		var ubuntuForWorkspace assetsApi.AssetMeta
		var ubuntuForDistributed assetsApi.AssetMeta
		var redhatForDistributed assetsApi.AssetMeta
		var ubuntu assetsApi.AssetMeta
		var redhat assetsApi.AssetMeta
		var compute assetsApi.AssetMeta
		var configMap assetsApi.AssetMeta

		BeforeEach(func() {
			//
			//     CREATE ASSETS
			//
			ubuntu = assetsApi.AssetMeta{
				Kind:      assetsApi.ASSETKIND_ENVIRONMENT,
				Id:        uuid.NewString(),
				Scope:     assetsApi.SCOPE_CLUSTER,
				ClusterId: *assetsApi.NewNullableString(convutils.Ref(randoms.ClusterId)),
				Name:      "env-cross-types",
			}

			ubuntuForWorkspace = assetsApi.AssetMeta{
				Kind:      assetsApi.ASSETKIND_ENVIRONMENT,
				Id:        uuid.NewString(),
				Scope:     assetsApi.SCOPE_CLUSTER,
				ClusterId: *assetsApi.NewNullableString(convutils.Ref(randoms.ClusterId)),
				Name:      "env-for-workspace",
				WorkloadSupportedTypes: *assetsApi.NewNullableWorkloadSupportedTypes(&assetsApi.WorkloadSupportedTypes{
					Workspace: *assetsApi.NewNullableBool(convutils.Ref(true)),
				}),
			}

			ubuntuForDistributed = assetsApi.AssetMeta{
				Kind:      assetsApi.ASSETKIND_ENVIRONMENT,
				Id:        uuid.NewString(),
				Scope:     assetsApi.SCOPE_CLUSTER,
				ClusterId: *assetsApi.NewNullableString(convutils.Ref(randoms.ClusterId)),
				Name:      "env-1-for-distributed",
				WorkloadSupportedTypes: *assetsApi.NewNullableWorkloadSupportedTypes(&assetsApi.WorkloadSupportedTypes{
					Distributed: *assetsApi.NewNullableBool(convutils.Ref(true)),
				}),
			}

			redhat = assetsApi.AssetMeta{
				Kind:      assetsApi.ASSETKIND_ENVIRONMENT,
				Id:        uuid.NewString(),
				Scope:     assetsApi.SCOPE_CLUSTER,
				ClusterId: *assetsApi.NewNullableString(convutils.Ref(randoms.ClusterId)),
				Name:      "env-2-cross-types",
			}

			redhatForDistributed = assetsApi.AssetMeta{
				Kind:      assetsApi.ASSETKIND_ENVIRONMENT,
				Id:        uuid.NewString(),
				Scope:     assetsApi.SCOPE_CLUSTER,
				ClusterId: *assetsApi.NewNullableString(convutils.Ref(randoms.ClusterId)),
				Name:      "env-2-for-distributed",
				WorkloadSupportedTypes: *assetsApi.NewNullableWorkloadSupportedTypes(&assetsApi.WorkloadSupportedTypes{
					Distributed: *assetsApi.NewNullableBool(convutils.Ref(true)),
				}),
			}

			compute = assetsApi.AssetMeta{
				Kind:      assetsApi.ASSETKIND_COMPUTE,
				Id:        uuid.NewString(),
				Scope:     assetsApi.SCOPE_CLUSTER,
				ClusterId: *assetsApi.NewNullableString(convutils.Ref(randoms.ClusterId)),
				Name:      "compute",
			}

			configMap = assetsApi.AssetMeta{
				Kind:      assetsApi.ASSETKIND_CONFIG_MAP,
				Id:        uuid.NewString(),
				Scope:     assetsApi.SCOPE_CLUSTER,
				ClusterId: *assetsApi.NewNullableString(convutils.Ref(randoms.ClusterId)),
				Name:      "config-map",
			}

			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(&ubuntu).WithMapping(api.WORKLOADTYPE_DISTRIBUTED_INFERENCE, ubuntu.Id, &assetsApi.SupersetSpec{
				Image: *assetsApi.NewNullableString(convutils.Ref("ubuntu")),
			}).
				WithAssetMeta(&ubuntuForWorkspace).WithMapping(api.WORKLOADTYPE_WORKSPACE, ubuntuForWorkspace.Id, &assetsApi.SupersetSpec{
				Image: *assetsApi.NewNullableString(convutils.Ref("ubuntu")),
			}).
				WithAssetMeta(&ubuntuForDistributed).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, ubuntuForDistributed.Id, &assetsApi.SupersetSpec{
				Image: *assetsApi.NewNullableString(convutils.Ref("ubuntu")),
			}).
				WithAssetMeta(&redhatForDistributed).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, redhatForDistributed.Id, &assetsApi.SupersetSpec{
				Image: *assetsApi.NewNullableString(convutils.Ref("redhat")),
			}).
				WithAssetMeta(&redhat).WithMapping(api.WORKLOADTYPE_DISTRIBUTED_INFERENCE, redhat.Id, &assetsApi.SupersetSpec{
				Image: *assetsApi.NewNullableString(convutils.Ref("redhat")),
			}).
				WithAssetMeta(&compute).WithMapping(api.WORKLOADTYPE_WORKSPACE, compute.Id, &assetsApi.SupersetSpec{
				Compute: *assetsApi.NewNullableDistributedInferenceLeaderWorkerSpecCompute(&assetsApi.DistributedInferenceLeaderWorkerSpecCompute{
					GpuDevicesRequest: *assetsApi.NewNullableInt32(convutils.Ref(int32(12))),
				}),
			}).
				WithAssetMeta(&configMap).WithMapping(api.WORKLOADTYPE_WORKSPACE, configMap.Id, &assetsApi.SupersetSpec{
				Storage: *assetsApi.NewNullableSupersetSpecAllOfStorage(&assetsApi.SupersetSpecAllOfStorage{
					ConfigMapVolume: []assetsApi.ConfigMapInstance{
						{
							Name: *assetsApi.NewNullableString(convutils.Ref("my-cm")),
						},
					},
				}),
			}))

			//
			//   we want simple policies that provides working directory
			//
			workspacePolicy := policyService.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: *policyService.NewNullableWorkspacePolicyDefaultsV2(&policyService.WorkspacePolicyDefaultsV2{
					WorkingDir: *policyService.NewNullableString(convutils.Ref("/policy/workspace/dir")),
				}),
			}

			distributedPolicy := policyService.DistributedPolicyDefaultsAndRulesV2{
				Defaults: *policyService.NewNullableDistributedPolicyDefaultsV2(&policyService.DistributedPolicyDefaultsV2{
					Worker: *policyService.NewNullableDistributedPolicyDefaultsV2Worker(&policyService.DistributedPolicyDefaultsV2Worker{
						WorkingDir: *policyService.NewNullableString(convutils.Ref("/policy/distributed/worker/dir")),
					}),
					Master: *policyService.NewNullableReplicaDefaultsV2(&policyService.ReplicaDefaultsV2{
						WorkingDir: *policyService.NewNullableString(convutils.Ref("/policy/distributed/master/dir")),
					}),
				}),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt:           time.Now(),
					EffectiveWorkspacePolicyV2:   &workspacePolicy,
					EffectiveDistributedPolicyV2: &distributedPolicy,
				},
			})
		})

		It("workspace, all assets ok", func() {
			// Positive -> workspace
			payload, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_WORKSPACE),
				AssetIds: &api.AssetSubmissionIds{
					Environment: &ubuntuForWorkspace.Id,
					Compute:     &compute.Id,
					Datasources: []string{
						configMap.Id,
					},
				},
			}, false)
			Expect(err).To(BeNil())

			expectedPayload := api.SupersetSpec{
				Image: convutils.Ref("ubuntu"),
				Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuDevicesRequest: convutils.Ref(int32(12)),
				},
				Storage: &api.WorkspaceSpecSpecAllOfStorage{
					ConfigMapVolume: &[]api.ConfigMapInstance{
						{
							Name: convutils.Ref("my-cm"),
						},
					},
				},
			}
			testutils.Expect(payload).ToEqual(expectedPayload)
		})

		It("workspace, datasource only", func() {
			// Positive -> workspace
			payload, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_WORKSPACE),
				AssetIds: &api.AssetSubmissionIds{
					Datasources: []string{
						configMap.Id,
					},
				},
			}, false)
			Expect(err).To(BeNil())

			expectedPayload := api.SupersetSpec{
				Storage: &api.WorkspaceSpecSpecAllOfStorage{
					ConfigMapVolume: &[]api.ConfigMapInstance{
						{
							Name: convutils.Ref("my-cm"),
						},
					},
				},
			}
			testutils.Expect(payload).ToEqual(expectedPayload)
		})

		It("workspace, environment only", func() {
			// Positive -> workspace
			payload, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_WORKSPACE),
				AssetIds: &api.AssetSubmissionIds{
					Environment: &ubuntuForWorkspace.Id,
				},
			}, false)
			Expect(err).To(BeNil())

			expectedPayload := api.SupersetSpec{
				Image: convutils.Ref("ubuntu"),
			}
			testutils.Expect(payload).ToEqual(expectedPayload)
		})

		It("workspace, compute only", func() {
			// Positive -> workspace
			payload, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_WORKSPACE),
				AssetIds: &api.AssetSubmissionIds{
					Compute: &compute.Id,
				},
			}, false)
			Expect(err).To(BeNil())

			expectedPayload := api.SupersetSpec{
				Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuDevicesRequest: convutils.Ref(int32(12)),
				},
			}
			testutils.Expect(payload).ToEqual(expectedPayload)
		})

		It("workspace, no assets", func() {
			// Positive -> workspace
			_, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_WORKSPACE),
				AssetIds:     &api.AssetSubmissionIds{},
			}, false)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("at least one asset ID is required when submitting with assets"))
		})

		It("workspace, compute sent instead of env", func() {
			// Positive -> workspace
			_, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_WORKSPACE),
				AssetIds: &api.AssetSubmissionIds{
					Environment: &compute.Id,
				},
			}, false)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("is not an environment asset"))
		})

		It("workspace, env sent instead of compute", func() {
			// Positive -> workspace
			_, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_WORKSPACE),
				AssetIds: &api.AssetSubmissionIds{
					Compute: &ubuntuForWorkspace.Id,
				},
			}, false)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("is not a compute asset"))
		})

		It("workspace, env sent instead of datasource", func() {
			// Positive -> workspace
			_, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_WORKSPACE),
				AssetIds: &api.AssetSubmissionIds{
					Datasources: []string{
						compute.Id,
						ubuntuForWorkspace.Id,
					},
				},
			}, false)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("is not a datasource asset"))
		})

		It("workspace, distributed env asset sent instead of workspace asset", func() {
			// Positive -> workspace
			_, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_WORKSPACE),
				AssetIds: &api.AssetSubmissionIds{
					Environment: &ubuntuForDistributed.Id,
				},
			}, false)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("is not supported for workload of type Workspace"))
		})

		It("distributed, workspace env asset sent instead of distributed asset", func() {
			// Positive -> workspace
			_, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_DISTRIBUTED),
				AssetIds: &api.AssetSubmissionIds{
					Environment: &ubuntuForWorkspace.Id,
				},
			}, false)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("is not supported for workload of type Distributed"))
		})

		It("distributed, worker + master", func() {
			// Positive -> workspace
			payload, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_DISTRIBUTED),
				AssetIds: &api.AssetSubmissionIds{
					Environment: &ubuntuForDistributed.Id,
				},
				MasterAssetIds: &api.AssetSubmissionIds{
					Environment: &redhatForDistributed.Id,
				},
			}, false)
			Expect(err).To(BeNil())

			expectedPayload := map[string]interface{}{
				workerSection: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
				masterSection: api.SupersetSpec{
					Image: convutils.Ref("redhat"),
				},
			}
			testutils.Expect(payload).ToEqual(expectedPayload)
		})

		It("distributed inference, worker + leader", func() {
			// Positive -> workspace
			payload, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
				AssetIds: &api.AssetSubmissionIds{
					Environment: &ubuntu.Id,
				},
				LeaderAssetIds: &api.AssetSubmissionIds{
					Environment: &redhat.Id,
				},
			}, false)
			Expect(err).To(BeNil())

			expectedPayload := map[string]interface{}{
				workerSection: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
				leaderSection: api.SupersetSpec{
					Image: convutils.Ref("redhat"),
				},
			}
			testutils.Expect(payload).ToEqual(expectedPayload)
		})

		It("distributed, worker only", func() {
			// Positive -> workspace
			payload, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_DISTRIBUTED),
				AssetIds: &api.AssetSubmissionIds{
					Environment: &ubuntuForDistributed.Id,
				},
			}, false)
			Expect(err).To(BeNil())

			expectedPayload := map[string]interface{}{
				workerSection: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}
			testutils.Expect(payload).ToEqual(expectedPayload)
		})

		It("distributed, master only", func() {
			// Positive -> workspace
			payload, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_DISTRIBUTED),
				MasterAssetIds: &api.AssetSubmissionIds{
					Environment: &ubuntuForDistributed.Id,
				},
			}, false)
			Expect(err).To(BeNil())

			expectedPayload := map[string]interface{}{
				masterSection: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}
			testutils.Expect(payload).ToEqual(expectedPayload)
		})

		It("distributed, worker same as master", func() {
			// Positive -> workspace
			payload, err := loadAssets(context.Background(), &createWorkloadParams{
				WorkloadType: string(api.WORKLOADTYPE_DISTRIBUTED),
				AssetIds: &api.AssetSubmissionIds{
					Environment: &ubuntuForDistributed.Id,
				},
			}, true)
			Expect(err).To(BeNil())

			// sanitize the payload for comparing with the expected
			delete(payload, "assetsMeta")

			expectedPayload := map[string]interface{}{
				workerSection: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
				masterSection: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}
			testutils.Expect(payload).ToEqual(expectedPayload)
		})

		It("workspace submission with asset template and spec", func() {
			ts := TemplatesService{}
			ws := WorkspacesService{}

			//
			//    CREATE WORKSPACE TEMPLATE
			//
			workspaceTemplateReq := api.WorkspaceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "ws-1",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.WorkspaceTemplateSpec{
					Command: convutils.Ref("command-from-template"),
				},
			}

			unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, &workspaceTemplateReq)
			Expect(err).To(BeNil())

			err = ts.CreateTemplate(ctx, unifiedRequest)
			Expect(err).To(BeNil())

			//
			//   create the workload
			//
			workloadRequest := api.WorkspaceCreationRequest{
				Name:       "ws-1",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(unifiedRequest.template.Meta.Id),
				Assets: &api.AssetSubmissionIds{
					Environment: &ubuntuForWorkspace.Id,
				},
				Spec: api.WorkspaceSpecSpec{
					Args: convutils.Ref("args-from-spec"),
				},
			}

			response, _ := ws.CreateWorkspace(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			workload := response.Body.(workloadsApi.SupersetWorkload)
			Expect(workload).NotTo(BeNil())
			Expect(workload.Spec).NotTo(BeNil())

			Expect(*workload.Spec.Image.Get()).To(Equal("ubuntu")) // from the asset
			Expect(*workload.Spec.Command.Get()).To(Equal("command-from-template"))
			Expect(*workload.Spec.Args.Get()).To(Equal("args-from-spec"))
		})

		It("distributed submission with asset template and spec", func() {
			ts := TemplatesService{}
			ds := DistributedService{}

			//
			//    CREATE DISTRIBUTED TEMPLATE
			//
			distributedTemplateReq := api.DistributedTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "ds-1",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.DistributedTemplateSpec{
					Command: convutils.Ref("command-from-template-worker"),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Command: convutils.Ref("command-from-template-master"),
					},
				},
			}

			unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, &distributedTemplateReq)
			Expect(err).To(BeNil())

			err = ts.CreateTemplate(ctx, unifiedRequest)
			Expect(err).To(BeNil())

			//
			//   create the workload - master AND worker
			//
			workloadRequest := api.DistributedCreationRequest{
				Name:       "ds-1",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(unifiedRequest.template.Meta.Id),
				Assets: &api.AssetSubmissionIds{
					Environment: &ubuntuForDistributed.Id,
				},
				MasterAssets: &api.AssetSubmissionIds{
					Environment: &redhatForDistributed.Id,
				},
				Spec: api.DistributedSpecSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Args:                 convutils.Ref("args-from-spec-worker"),
				},
				MasterSpec: &api.MasterSpec{
					Args: convutils.Ref("args-from-spec-master"),
				},
			}

			response, _ := ds.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			workload := response.Body.(workloadsApi.SupersetWorkload)
			Expect(workload).NotTo(BeNil())
			Expect(workload.Spec).NotTo(BeNil())
			Expect(workload.MasterSpec).NotTo(BeNil())

			Expect(*workload.Spec.Image.Get()).To(Equal("ubuntu")) // from the asset
			Expect(*workload.Spec.Command.Get()).To(Equal("command-from-template-worker"))
			Expect(*workload.Spec.Args.Get()).To(Equal("args-from-spec-worker"))

			Expect(*workload.MasterSpec.Get().Image.Get()).To(Equal("redhat")) // from the asset
			Expect(*workload.MasterSpec.Get().Command.Get()).To(Equal("command-from-template-master"))
			Expect(*workload.MasterSpec.Get().Args.Get()).To(Equal("args-from-spec-master"))

			//
			//   create the workload - master same as worker
			//
			workloadRequest = api.DistributedCreationRequest{
				Name:       "ds-2",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(unifiedRequest.template.Meta.Id),
				Assets: &api.AssetSubmissionIds{
					Environment: &ubuntuForDistributed.Id,
				},
				Spec: api.DistributedSpecSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Args:                 convutils.Ref("args-from-spec-worker"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			response, _ = ds.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			workload = response.Body.(workloadsApi.SupersetWorkload)
			Expect(workload).NotTo(BeNil())
			Expect(workload.Spec).NotTo(BeNil())
			Expect(workload.MasterSpec).NotTo(BeNil())

			Expect(*workload.Spec.Image.Get()).To(Equal("ubuntu")) // from the asset
			Expect(*workload.Spec.Command.Get()).To(Equal("command-from-template-worker"))
			Expect(*workload.Spec.Args.Get()).To(Equal("args-from-spec-worker"))

			Expect(*workload.MasterSpec.Get().Image.Get()).To(Equal("ubuntu")) // from the asset
			Expect(*workload.MasterSpec.Get().Command.Get()).To(Equal("command-from-template-worker"))
			Expect(*workload.MasterSpec.Get().Args.Get()).To(Equal("args-from-spec-worker"))

		})
	})

	//
	//    field        policy    template    asset    spec      result
	//    ------------ --------  ----------  -------  --------- ---------
	//    image        img-pol   img-templ   img-ass  img-spec  img-spec
	//    workingDir   wd-pol    wd-templ    wd-ass             wd-ass
	//    command      cmd-pol   cmd-templ                      cmd-templ
	//    args         args-pol                                 cmd-args
	//    envvar       end-pol   env-temp    env-ass  env-spec  env-pol + env-temp + env-ass + env-spec
	//
	Context("policy + assets + templates + spec combo", func() {
		It("should take the correct order", func() {
			//
			//   set the policy
			//
			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveTrainingPolicyV2: &policyService.TrainingPolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableTrainingPolicyDefaultsV2(&policyService.TrainingPolicyDefaultsV2{
							Image:      *policyService.NewNullableString(convutils.Ref("img-pol")),
							WorkingDir: *policyService.NewNullableString(convutils.Ref("wd-pol")),
							Command:    *policyService.NewNullableString(convutils.Ref("cmd-pol")),
							Args:       *policyService.NewNullableString(convutils.Ref("args-pol")),
							EnvironmentVariables: *policyService.NewNullableEnvironmentVariablesDefaults(&policyService.EnvironmentVariablesDefaults{
								Instances: []policyService.EnvironmentVariable{
									{
										Name:  *policyService.NewNullableString(convutils.Ref("env-pol")),
										Value: *policyService.NewNullableString(convutils.Ref("env-pol")),
									},
								},
							}),
						}),
					},
				},
			})

			//
			//   create the asset
			//
			assetId := uuid.NewString()

			assetMeta := assetsApi.AssetMeta{
				Kind:      assetsApi.ASSETKIND_ENVIRONMENT,
				Id:        assetId,
				Name:      "env",
				Scope:     assetsApi.SCOPE_CLUSTER,
				TenantId:  *assetsApi.NewNullableInt32(convutils.Ref(int32(randoms.TenantId))),
				ClusterId: *assetsApi.NewNullableString(convutils.Ref(randoms.ClusterId)),
			}

			assetFlatSpec := assetsApi.SupersetSpec{
				Image:      *assetsApi.NewNullableString(convutils.Ref("img-ass")),
				WorkingDir: *assetsApi.NewNullableString(convutils.Ref("wd-ass")),
				EnvironmentVariables: []assetsApi.EnvironmentVariable{
					{
						Name:  *assetsApi.NewNullableString(convutils.Ref("env-ass")),
						Value: *assetsApi.NewNullableString(convutils.Ref("env-ass")),
					},
				},
			}

			assetsClient.SetMock(assetsClient.NewSimpleMock().
				WithAssetMeta(&assetMeta).WithMapping(api.WORKLOADTYPE_TRAINING, assetId, &assetFlatSpec))

			//
			//  create the template
			//
			ts := TemplatesService{}

			trainingTemplate := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "ts-1",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.TrainingTemplateSpec{
					Image:      convutils.Ref("img-templ"),
					WorkingDir: convutils.Ref("wd-templ"),
					Command:    convutils.Ref("cmd-templ"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("env-templ"),
							Value: convutils.Ref("env-templ"),
						},
					},
				},
			}

			unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_TRAINING, &trainingTemplate)
			Expect(err).To(BeNil())

			err = ts.CreateTemplate(ctx, unifiedRequest)
			Expect(err).To(BeNil())

			//
			//   create the workload
			//
			workloadRequest := api.TrainingCreationRequest{
				Name:       "train-1",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(unifiedRequest.template.Meta.Id),
				Assets: &api.AssetSubmissionIds{
					Environment: convutils.Ref(assetId),
				},

				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("img-spec"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("env-spec"),
							Value: convutils.Ref("env-spec"),
						},
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			//
			//    field        policy    template    asset    spec      result
			//    ------------ --------  ----------  -------  --------- ---------
			//    image        img-pol   img-templ   img-ass  img-spec  img-spec
			//    workingDir   wd-pol    wd-templ    wd-ass             wd-ass
			//    command      cmd-pol   cmd-templ                      cmd-templ
			//    args         args-pol                                 cmd-pol
			//    envvar       end-pol   env-temp    env-ass  env-spec  env-pol + env-temp + env-ass + env-spec
			//
			workload := response.Body.(workloadsApi.SupersetWorkload)
			Expect(*workload.Spec.Image.Get()).To(Equal("img-spec"))
			Expect(*workload.Spec.WorkingDir.Get()).To(Equal("wd-ass"))
			Expect(*workload.Spec.Command.Get()).To(Equal("cmd-templ"))
			Expect(*workload.Spec.Args.Get()).To(Equal("args-pol"))
			Expect(len(workload.Spec.EnvironmentVariables)).To(Equal(4))
		})
	})

	//
	//  EXCLUSION TEST (x = variable excluded)
	//    env          policy    template    asset    spec      result
	//    ------------ --------  ----------  -------  --------- ---------
	//    a            pol       X                              X
	//    b            pol       X           as                 as
	//    c            pol       X           as       X         X
	//    d                      temp        X        spec      spec
	//
	Context("policy + assets + templates + spec combo", func() {
		It("should exclude/unexclude variables in the correct order", func() {
			//
			//   set the policy
			//
			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveTrainingPolicyV2: &policyService.TrainingPolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableTrainingPolicyDefaultsV2(&policyService.TrainingPolicyDefaultsV2{
							EnvironmentVariables: *policyService.NewNullableEnvironmentVariablesDefaults(&policyService.EnvironmentVariablesDefaults{
								Instances: []policyService.EnvironmentVariable{
									{
										Name:  *policyService.NewNullableString(convutils.Ref("a")),
										Value: *policyService.NewNullableString(convutils.Ref("pol")),
									},
									{
										Name:  *policyService.NewNullableString(convutils.Ref("b")),
										Value: *policyService.NewNullableString(convutils.Ref("pol")),
									},
									{
										Name:  *policyService.NewNullableString(convutils.Ref("c")),
										Value: *policyService.NewNullableString(convutils.Ref("pol")),
									},
								},
							}),
						}),
					},
				},
			})

			//
			//  create the template
			//
			ts := TemplatesService{}

			trainingTemplate := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "ts-1",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.TrainingTemplateSpec{
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:    convutils.Ref("a"),
							Exclude: convutils.Ref(true),
						},
						{
							Name:    convutils.Ref("b"),
							Exclude: convutils.Ref(true),
						},
						{
							Name:    convutils.Ref("c"),
							Exclude: convutils.Ref(true),
						},
						{
							Name:  convutils.Ref("d"),
							Value: convutils.Ref("temp"),
						},
					},
				},
			}

			unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_TRAINING, &trainingTemplate)
			Expect(err).To(BeNil())

			err = ts.CreateTemplate(ctx, unifiedRequest)
			Expect(err).To(BeNil())

			//
			//   create the asset
			//
			assetId := uuid.NewString()

			assetMeta := assetsApi.AssetMeta{
				Kind:      assetsApi.ASSETKIND_ENVIRONMENT,
				Id:        assetId,
				Name:      "env",
				Scope:     assetsApi.SCOPE_CLUSTER,
				TenantId:  *assetsApi.NewNullableInt32(convutils.Ref(int32(randoms.TenantId))),
				ClusterId: *assetsApi.NewNullableString(convutils.Ref(randoms.ClusterId)),
			}

			assetFlatSpec := assetsApi.SupersetSpec{
				EnvironmentVariables: []assetsApi.EnvironmentVariable{
					{
						Name:  *assetsApi.NewNullableString(convutils.Ref("b")),
						Value: *assetsApi.NewNullableString(convutils.Ref("as")),
					},
					{
						Name:    *assetsApi.NewNullableString(convutils.Ref("c")),
						Exclude: *assetsApi.NewNullableBool(convutils.Ref(true)),
					},
					{
						Name:  *assetsApi.NewNullableString(convutils.Ref("d")),
						Value: *assetsApi.NewNullableString(convutils.Ref("as")),
					},
				},
			}

			assetsClient.SetMock(assetsClient.NewSimpleMock().
				WithAssetMeta(&assetMeta).WithMapping(api.WORKLOADTYPE_TRAINING, assetId, &assetFlatSpec))

			//
			//   create the workload
			//
			workloadRequest := api.TrainingCreationRequest{
				Name:       "train-1",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(unifiedRequest.template.Meta.Id),
				Assets: &api.AssetSubmissionIds{
					Environment: convutils.Ref(assetId),
				},

				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("img-spec"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:    convutils.Ref("c"),
							Exclude: convutils.Ref(true),
						},
						{
							Name:  convutils.Ref("d"),
							Value: convutils.Ref("spec"),
						},
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			//
			//    env          policy    template    asset    spec      result
			//    ------------ --------  ----------  -------  --------- ---------
			//    a            pol       X                              X
			//    b            pol       X           as                 as
			//    c            pol       X           as       X         X
			//    d                      temp        X        spec      spec
			//
			workload := response.Body.(workloadsApi.SupersetWorkload)
			Expect(len(workload.Spec.EnvironmentVariables)).To(Equal(4))
			for _, env := range workload.Spec.EnvironmentVariables {
				switch *env.Name.Get() {
				case "a":
					Expect(convutils.BoolIsTrue(env.Exclude.Get())).To(Equal(true))
				case "b":
					Expect(convutils.BoolIsTrue(env.Exclude.Get())).To(Equal(false))
					Expect(*env.Value.Get()).To(Equal("as"))
				case "c":
					Expect(convutils.BoolIsTrue(env.Exclude.Get())).To(Equal(true))
				case "d":
					Expect(convutils.BoolIsTrue(env.Exclude.Get())).To(Equal(false))
					Expect(*env.Value.Get()).To(Equal("spec"))
				}
			}
		})
	})
})
