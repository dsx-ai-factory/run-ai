/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"

	csclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	assetsServiceApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	assetsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/datavolumes"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"

	k8sObjectsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/k8s-objects-tracker"
	k8sObjectsApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	k8sObjectsTrackerMock "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/mocks"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"

	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"

	"github.com/stretchr/testify/mock"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"
	policyService "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeAnyWorkloadSubmitter is a test double for AnyWorkloadSubmitter that lets a test
// control the reported submission methods and the result of Create, so the create-flow
// fallback logic can be exercised deterministically.
type fakeAnyWorkloadSubmitter struct {
	methods      SubmissionMethods
	submission   *Submission
	createErr    *workloadsApi.SubmissionErrorV2
	createCalled bool
}

func (f *fakeAnyWorkloadSubmitter) DetermineSubmissionMethods(context.Context, *workloadContext) SubmissionMethods {
	return f.methods
}

func (f *fakeAnyWorkloadSubmitter) CanUpdate(context.Context, SubmissionMethods, *workloadContext, string) *workloadsApi.SubmissionErrorV2 {
	return nil
}

func (f *fakeAnyWorkloadSubmitter) Create(context.Context, *workloadContext) (*Submission, *workloadsApi.SubmissionErrorV2) {
	f.createCalled = true
	return f.submission, f.createErr
}

func (f *fakeAnyWorkloadSubmitter) Update(context.Context, *workloadContext, string, *workloadMgrApi.WorkloadV2GetResponse, *models_crud.Workload) (*Update, *workloadsApi.SubmissionErrorV2) {
	return nil, nil
}

var _ = Describe("Workload composer validations", func() {
	var ctx context.Context
	var randoms *testutils.RandomValues
	var forbiddenProject *testutils.RandomValues
	var differentCluster *testutils.RandomValues

	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}

	trainingsService := NewTrainingsService(mockTenantSettingsClient)
	distributedService := NewDistributedService(mockTenantSettingsClient)
	workspaceService := NewWorkspacesService(mockTenantSettingsClient)
	inferenceService := NewInferencesService(mockTenantSettingsClient)

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		differentCluster = randoms.CloneWithDifferentCluster()

		k8sObjectsClient.SetMock(k8sObjectsTrackerMock.NewK8SObjectTrackerMock())

		// simulate another project that user does not have permission to access
		forbiddenProject = randoms.CloneWithDifferentProject().Forbidden()

		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		policyMockUpdateTime := time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms).With(forbiddenProject).With(differentCluster)
		authMock.AllowedProjectsToContext(&ctx)

		orgunit.SetMock(authMock.OrgUnitMock())
		authorization.SetMock(authMock)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, nil)).
			WithCluster(differentCluster.ClusterId, testing.BuildDisplayedCluster(differentCluster.ClusterId, "2.21.0", nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))

		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	DescribeTable("GPU memory aliases survive workspace creation and retrieval",
		func(legacy, canonical *float64, expectedCode int) {
			request := api.WorkspaceCreationRequest{
				Name: "memory-alias-test", ProjectId: randoms.ProjectIdStr, ClusterId: randoms.ClusterId,
				Spec: api.WorkspaceSpecSpec{Image: convutils.Ref("ubuntu"), Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuRequestType:    convutils.Ref(api.GPUREQUESTTYPE_PORTION),
					GpuPortionRequest: legacy, GpuMemoryPortionRequest: canonical,
				}},
			}
			created, err := workspaceService.CreateWorkspace(ctx, request)
			Expect(err).To(BeNil())
			Expect(created.Code).To(Equal(expectedCode))
			if expectedCode != http.StatusAccepted {
				return
			}
			expectedValue := convutils.Deref(legacy, convutils.Deref(canonical, 0))
			workload := created.Body.(workloadsApi.SupersetWorkload)
			Expect(workload.Spec.Compute.Get().GetGpuPortionRequest()).To(Equal(expectedValue))
			Expect(workload.Spec.Compute.Get().GpuMemoryPortionRequest.IsSet()).To(Equal(expectedValue <= 1))
			if expectedValue <= 1 {
				Expect(workload.Spec.Compute.Get().GetGpuMemoryPortionRequest()).To(Equal(expectedValue))
			}
			var stored models_crud.Workload
			Expect(runaidb.DB().Where("id = ?", workload.WorkloadId).First(&stored)).To(Succeed())
			Expect(stored.Spec.Compute.Get().GpuMemoryPortionRequest.IsSet()).To(BeFalse())
			Expect(stored.Spec.Compute.Get().GetGpuPortionRequest()).To(Equal(expectedValue))
			retrieved, err := workspaceService.GetWorkspace(ctx, workload.WorkloadId)
			Expect(err).To(BeNil())
			Expect(retrieved.Code).To(Equal(http.StatusOK))
			response := retrieved.Body.(workloadsApi.SupersetWorkload)
			Expect(response.Spec.Compute.Get().GetGpuPortionRequest()).To(Equal(expectedValue))
			Expect(response.Spec.Compute.Get().GpuMemoryPortionRequest.IsSet()).To(Equal(expectedValue <= 1))
			if expectedValue <= 1 {
				Expect(response.Spec.Compute.Get().GetGpuMemoryPortionRequest()).To(Equal(expectedValue))
			}
		},
		Entry("deprecated input", convutils.Ref(.25), nil, http.StatusAccepted),
		Entry("canonical input", nil, convutils.Ref(.25), http.StatusAccepted),
		Entry("equal aliases", convutils.Ref(.25), convutils.Ref(.25), http.StatusAccepted),
		Entry("conflicting aliases", convutils.Ref(.25), convutils.Ref(.5), http.StatusBadRequest),
		Entry("legacy whole GPU request", convutils.Ref(4.0), nil, http.StatusAccepted),
		Entry("new whole GPU request rejected", nil, convutils.Ref(4.0), http.StatusBadRequest),
		Entry("equal whole GPU aliases rejected", convutils.Ref(4.0), convutils.Ref(4.0), http.StatusBadRequest),
		Entry("legacy non-integer above one rejected", convutils.Ref(4.5), nil, http.StatusBadRequest),
	)

	Context("fixed name vs prefix name", func() {
		It("should create proper name", func() {
			requestFixed := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			requestPrefixed := api.TrainingCreationRequest{
				Name:                 "trainjob",
				UseGivenNameAsPrefix: true,
				ProjectId:            randoms.ProjectIdStr,
				ClusterId:            randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			fixedResponse, _ := trainingsService.CreateTraining(ctx, requestFixed)
			Expect(fixedResponse.Code).To(Equal(http.StatusAccepted))
			fixedWorkloadId := fixedResponse.Body.(workloadsApi.SupersetWorkload).WorkloadId

			prefixResponse, _ := trainingsService.CreateTraining(ctx, requestPrefixed)
			Expect(prefixResponse.Code).To(Equal(http.StatusAccepted))
			prefixedWorkloadId := prefixResponse.Body.(workloadsApi.SupersetWorkload).WorkloadId

			// get training should return the correct name for each of them
			getFixed, _ := trainingsService.GetTraining(ctx, fixedWorkloadId)
			Expect(getFixed.Code).To(Equal(http.StatusOK))

			getPrefixed, _ := trainingsService.GetTraining(ctx, prefixedWorkloadId)
			Expect(getFixed.Code).To(Equal(http.StatusOK))

			fixedPayload := getFixed.Body.(workloadsApi.SupersetWorkload)
			prefixedPayload := getPrefixed.Body.(workloadsApi.SupersetWorkload)

			Expect(fixedPayload.RequestedName).To(Equal(requestFixed.Name))
			Expect(prefixedPayload.RequestedName).To(Equal(requestPrefixed.Name))

			Expect(fixedPayload.Name).To(Equal(requestFixed.Name))
			Expect(prefixedPayload.Name).NotTo(Equal(requestPrefixed.Name))
			Expect(strings.HasPrefix(requestPrefixed.Name, fixedPayload.Name))

			//
			//   make sure that in workloads table the full name is written
			//
			prefixedWorkloadUuid, err := uuid.Parse(prefixedWorkloadId)
			Expect(err).To(BeNil())

			workload, dbErr := dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]()).DbGetWorkload(ctx, randoms.TenantId, prefixedWorkloadUuid)
			Expect(dbErr).To(BeNil())
			Expect(workload.Name).To(Equal(prefixedPayload.Name))
		})
	})

	Context("any-workload submitter produces no submission", func() {
		var originalWorkspaceSubmitter AnyWorkloadSubmitter

		BeforeEach(func() {
			originalWorkspaceSubmitter = anyWorkloadSubmitters[Workspace]
		})

		AfterEach(func() {
			anyWorkloadSubmitters[Workspace] = originalWorkspaceSubmitter
		})

		It("returns an error and does not persist natively when native submission is not allowed", func() {
			fake := &fakeAnyWorkloadSubmitter{methods: SubmissionMethods{AnyWorkload: true, Native: false}}
			anyWorkloadSubmitters[Workspace] = fake

			request := createWorkspaceRequest("ws-no-native", randoms.ProjectIdStr, randoms.ClusterId, "", "")

			response, err := workspaceService.CreateWorkspace(ctx, *request)
			Expect(err).To(BeNil())

			// Create was consulted, but with no submission and no native path the flow must error out
			Expect(fake.createCalled).To(BeTrue())
			Expect(response.Code).To(Equal(http.StatusInternalServerError))

			// nothing should have been persisted through the native path
			var count int64
			Expect(runaidb.DB().Table(models_crud.WorkloadsTableName).Where("tenant_id = ?", randoms.TenantId).Count(&count)).To(BeNil())
			Expect(count).To(BeZero())
		})

		It("falls back to native submission when native is allowed", func() {
			fake := &fakeAnyWorkloadSubmitter{methods: SubmissionMethods{AnyWorkload: true, Native: true}}
			anyWorkloadSubmitters[Workspace] = fake

			request := createWorkspaceRequest("ws-native-fallback", randoms.ProjectIdStr, randoms.ClusterId, "", "")

			response, err := workspaceService.CreateWorkspace(ctx, *request)
			Expect(err).To(BeNil())

			// Create was consulted, produced no submission, and the flow fell back to native
			Expect(fake.createCalled).To(BeTrue())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)
			Expect(payload.WorkloadId).NotTo(BeEmpty())

			// the workload should be retrievable, proving it was persisted natively
			getResponse, _ := workspaceService.GetWorkspace(ctx, payload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))
		})
	})

	Context("name conflict", func() {
		It("should fail", func() {
			const name = "name-conflict-test"
			requestFixed := api.TrainingCreationRequest{
				Name:      name,
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			fixedResponse, _ := trainingsService.CreateTraining(ctx, requestFixed)
			Expect(fixedResponse.Code).To(Equal(http.StatusAccepted))

			conflictResponse, _ := trainingsService.CreateTraining(ctx, requestFixed)
			Expect(conflictResponse.Code).To(Equal(http.StatusConflict))

			workspaceRequse := api.WorkspaceCreationRequest{
				Name:      name,
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.WorkspaceSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}
			conflictResponse, _ = workspaceService.CreateWorkspace(ctx, workspaceRequse)
			Expect(conflictResponse.Code).To(Equal(http.StatusConflict))
		})
	})

	Context("attempt to create workload in forbidden project", func() {
		It("should fail", func() {
			requestFixed := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: forbiddenProject.ProjectIdStr,
				ClusterId: forbiddenProject.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			org_unit.SetMock(org_unit.NewSimpleMock().WithProject(forbiddenProject.ProjectIdStr, GetProjectResponse(*forbiddenProject)))
			response, _ := trainingsService.CreateTraining(ctx, requestFixed)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("attempt to give incorrect cluster id", func() {
		It("should fail", func() {
			requestFixed := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: differentCluster.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, requestFixed)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			payload := response.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(payload.Message).To(ContainSubstring("not associated with cluster"))
		})
	})

	Context("attempt to create pvc in non-allowed storage class", func() {
		BeforeEach(func() {
			allowedConfig := k8sObjectsApi.StorageClassConfiguration{
				ScopeId: randoms.ClusterId,
			}

			forbiddenWorkloadsConfig := k8sObjectsApi.StorageClassConfiguration{
				ScopeId: randoms.ClusterId,
				Permissions: *k8sObjectsApi.NewNullableStorageClassPermissions(&k8sObjectsApi.StorageClassPermissions{
					AllowedForWorkloads: *k8sObjectsApi.NewNullableBool(convutils.Ref(false)),
				}),
			}

			forbiddenPersistentVolConfig := k8sObjectsApi.StorageClassConfiguration{
				ScopeId: randoms.ClusterId,
				Permissions: *k8sObjectsApi.NewNullableStorageClassPermissions(&k8sObjectsApi.StorageClassPermissions{
					AllowedForPersistentVolumes: *k8sObjectsApi.NewNullableBool(convutils.Ref(false)),
				}),
			}

			forbiddenEphemeralVolConfig := k8sObjectsApi.StorageClassConfiguration{
				ScopeId: randoms.ClusterId,
				Permissions: *k8sObjectsApi.NewNullableStorageClassPermissions(&k8sObjectsApi.StorageClassPermissions{
					AllowedForEphemeralVolumes: *k8sObjectsApi.NewNullableBool(convutils.Ref(false)),
				}),
			}

			k8sObjectsClient.SetMock(k8sObjectsTrackerMock.NewK8SObjectTrackerMock().
				WithStorageClassConfig("allowed", allowedConfig).
				WithStorageClassConfig("forbidden-workloads", forbiddenWorkloadsConfig).
				WithStorageClassConfig("forbidden-persistent", forbiddenPersistentVolConfig).
				WithStorageClassConfig("forbidden-ephemeral", forbiddenEphemeralVolConfig))
		})

		It("non configured storage class", func() {
			notConfigured := api.TrainingCreationRequest{
				Name:      "trainjob-2",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Pvc: &[]api.PvcInstance{
							{
								ExistingPvc: convutils.Ref(false),
								ClaimInfo: &api.ClaimInfo{
									StorageClass: convutils.Ref("not-configured"),
									Size:         convutils.Ref("1G"),
								},
							},
							{
								ClaimName:   convutils.Ref("existing"),
								ExistingPvc: convutils.Ref(true),
								ClaimInfo: &api.ClaimInfo{
									StorageClass: convutils.Ref("not-configured"),
									Size:         convutils.Ref("1G"),
								},
							},
						},
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, notConfigured)
			Expect(response.Code).To(Equal(http.StatusAccepted))
		})

		It("existing pvc", func() {
			allowed := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Pvc: &[]api.PvcInstance{
							{
								ExistingPvc: convutils.Ref(true),
								ClaimName:   convutils.Ref("existing"),
								ClaimInfo: &api.ClaimInfo{
									StorageClass: convutils.Ref("allowed"),
								},
							},
							{
								ExistingPvc: convutils.Ref(true),
								ClaimName:   convutils.Ref("existing"),
								ClaimInfo: &api.ClaimInfo{
									StorageClass: convutils.Ref("forbidden-ephemeral"),
								},
							},
							{
								ExistingPvc: convutils.Ref(true),
								ClaimName:   convutils.Ref("existing"),
								ClaimInfo: &api.ClaimInfo{
									StorageClass: convutils.Ref("forbidden-persistent"),
								},
							},
						},
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, allowed)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			forbidden := api.TrainingCreationRequest{
				Name:      "trainjob-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Pvc: &[]api.PvcInstance{
							{
								ClaimName:   convutils.Ref("existing"),
								ExistingPvc: convutils.Ref(true),
								ClaimInfo: &api.ClaimInfo{
									StorageClass: convutils.Ref("forbidden-workloads"),
								},
							},
						},
					},
				},
			}

			response, _ = trainingsService.CreateTraining(ctx, forbidden)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			payload := response.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(payload.Message).To(ContainSubstring("storage class forbidden-workloads is forbidden for running workloads"))
		})

		It("non-existing pvc", func() {
			allowed := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Pvc: &[]api.PvcInstance{
							{
								ExistingPvc: convutils.Ref(false),
								ClaimInfo: &api.ClaimInfo{
									StorageClass: convutils.Ref("allowed"),
									Size:         convutils.Ref("1G"),
								},
							},
							{
								ExistingPvc: convutils.Ref(false),
								ClaimName:   convutils.Ref("non-existing"),
								Ephemeral:   convutils.Ref(true),
								ClaimInfo: &api.ClaimInfo{
									StorageClass: convutils.Ref("forbidden-persistent"),
									Size:         convutils.Ref("1G"),
								},
							},
							{
								ExistingPvc: convutils.Ref(false),
								ClaimName:   convutils.Ref("non-existing"),
								ClaimInfo: &api.ClaimInfo{
									StorageClass: convutils.Ref("forbidden-ephemeral"),
									Size:         convutils.Ref("1G"),
								},
							},
						},
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, allowed)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			forbiddenEphemeral := api.TrainingCreationRequest{
				Name:      "trainjob-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Pvc: &[]api.PvcInstance{
							{
								ExistingPvc: convutils.Ref(false),
								ClaimName:   convutils.Ref("non-existing"),
								Ephemeral:   convutils.Ref(true),
								ClaimInfo: &api.ClaimInfo{
									StorageClass: convutils.Ref("forbidden-ephemeral"),
								},
							},
						},
					},
				},
			}

			forbiddenPersistent := api.TrainingCreationRequest{
				Name:      "trainjob-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Pvc: &[]api.PvcInstance{
							{
								ExistingPvc: convutils.Ref(false),
								ClaimName:   convutils.Ref("non-existing"),
								ClaimInfo: &api.ClaimInfo{
									StorageClass: convutils.Ref("forbidden-persistent"),
								},
							},
						},
					},
				},
			}

			response, _ = trainingsService.CreateTraining(ctx, forbiddenEphemeral)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			payload := response.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(payload.Message).To(ContainSubstring("creating ephemeral volumes on storage class forbidden-ephemeral is forbidden"))

			response, _ = trainingsService.CreateTraining(ctx, forbiddenPersistent)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			payload = response.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(payload.Message).To(ContainSubstring("creating persistent volumes on storage class forbidden-persistent is forbidden"))
		})
	})

	Context("attempt to create pvc in non-allowed storage class", func() {
		It("allows a PVC to be mounted with the same access mode across multiple paths", func() {
			request := api.TrainingCreationRequest{
				Name:      "trainjob-consistent-access",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Pvc: &[]api.PvcInstance{
							{
								ClaimName:   convutils.Ref("claim-name1"),
								ExistingPvc: convutils.Ref(true),
								ReadOnly:    convutils.Ref(false),
							},
							{
								ClaimName:   convutils.Ref("claim-name1"),
								ExistingPvc: convutils.Ref(true),
								ReadOnly:    convutils.Ref(false),
							},
						},
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, request)
			Expect(response.Code).To(Equal(http.StatusAccepted))
		})

		It("rejects a PVC mounted with both read-only and read-write access modes", func() {
			request := api.TrainingCreationRequest{
				Name:      "trainjob-mixed-access",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Pvc: &[]api.PvcInstance{
							{
								ClaimName:   convutils.Ref("claim-name1"),
								ExistingPvc: convutils.Ref(true),
								ReadOnly:    convutils.Ref(false),
							},
							{
								ClaimName:   convutils.Ref("claim-name1"),
								ExistingPvc: convutils.Ref(true),
								ReadOnly:    convutils.Ref(true),
							},
						},
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			payload := response.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(payload.Message).To(ContainSubstring("claim-name1 cannot be mounted with both readOnly: true and readOnly: false"))
		})
	})

	Context("attempt to create workload with memory limit", func() {
		It("should fail no settings", func() {
			requestFixed := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuMemoryLimit: convutils.Ref("10M"),
					},
				},
			}

			//
			//  let's disable the flag in the backend mock
			//
			mockTenantSettingsClient.On("GetGpuResourceOptimization", mock.Anything).Return(false, nil).Once()

			response, _ := trainingsService.CreateTraining(ctx, requestFixed)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			payload := response.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(payload.Message).To(ContainSubstring("gpuPortionLimit and gpuMemoryLimit require GPU Optimization settings to be enabled"))

			//
			//  let's enable the flag in the backend mock
			//
			mockTenantSettingsClient.On("GetGpuResourceOptimization", mock.Anything).Return(true, nil).Once()

			response, _ = trainingsService.CreateTraining(ctx, requestFixed)
			Expect(response.Code).To(Equal(http.StatusAccepted))
		})
	})

	Context("attempt to apply policy defaults", func() {
		It("should apply policy correctly", func() {
			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveTrainingPolicyV2: &policyService.TrainingPolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableTrainingPolicyDefaultsV2(&policyService.TrainingPolicyDefaultsV2{
							Image: *policyService.NewNullableString(convutils.Ref("ubuntu")),
							EnvironmentVariables: *policyService.NewNullableEnvironmentVariablesDefaults(&policyService.EnvironmentVariablesDefaults{
								Instances: []policyService.EnvironmentVariable{
									{
										Name:  *policyService.NewNullableString(convutils.Ref("HOME")),
										Value: *policyService.NewNullableString(convutils.Ref("home-from-policy")),
									},
									{
										Name:  *policyService.NewNullableString(convutils.Ref("MORE")),
										Value: *policyService.NewNullableString(convutils.Ref("more-from-policy")),
									},
									{
										Name:  *policyService.NewNullableString(convutils.Ref("USER")),
										Value: *policyService.NewNullableString(convutils.Ref("user-from-policy")),
									},
								},
							}),
						}),
					},
				},
			})

			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						CpuMemoryRequest: convutils.Ref("10M"),
					},
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("HOME"),
							Value: convutils.Ref("name-from-workload"),
						},
						{
							Name:    convutils.Ref("MORE"),
							Exclude: convutils.Ref(true),
						},
						{
							Name: convutils.Ref("fieldref"),
							PodFieldRef: &api.EnvironmentVariablePodFieldReference{
								Path: convutils.Ref("value"),
							},
						},
					},
				},
			}

			response, err := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)

			actualAsMap, _ := gojsonq.LoadJson(payload.Spec.EnvironmentVariables)

			expectAsMap, _ := gojsonq.LoadJson([]workloadsApi.EnvironmentVariable{
				{
					Name:  *workloadsApi.NewNullableString(convutils.Ref("HOME")),
					Value: *workloadsApi.NewNullableString(convutils.Ref("home-from-workload")),
				},
				{
					Name:    *workloadsApi.NewNullableString(convutils.Ref("MORE")),
					Exclude: *workloadsApi.NewNullableBool(convutils.Ref(true)),
				},
				{
					Name:  *workloadsApi.NewNullableString(convutils.Ref("USER")),
					Value: *workloadsApi.NewNullableString(convutils.Ref("user-from-policy")),
				},
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("fieldref")),
					PodFieldRef: *workloadsApi.NewNullableEnvironmentVariablePodFieldReference(&workloadsApi.EnvironmentVariablePodFieldReference{
						Path: *workloadsApi.NewNullableString(convutils.Ref("value"))}),
				},
			})

			Expect(actualAsMap).To(Equal(expectAsMap))
		})
	})

	Context("gpu memory defaults", func() {
		It("should apply defaults correctly", func() {
			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveTrainingPolicyV2: &policyService.TrainingPolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableTrainingPolicyDefaultsV2(&policyService.TrainingPolicyDefaultsV2{
							Compute: *policyService.NewNullableSupersetDefaultsAllOfCompute(&policyService.SupersetDefaultsAllOfCompute{
								GpuMemoryRequest:  *policyService.NewNullableString(convutils.Ref("2M")),
								GpuDevicesRequest: *policyService.NewNullableInt32(convutils.Ref(int32(1))),
								GpuRequestType:    *policyService.NewNullableGpuRequestType(convutils.Ref(policyService.GPUREQUESTTYPE_MEMORY)),
							}),
						}),
						Rules: *policyService.NewNullableTrainingPolicyRulesV2(&policyService.TrainingPolicyRulesV2{
							Compute: *policyService.NewNullableSupersetRulesAllOfCompute(&policyService.SupersetRulesAllOfCompute{
								GpuRequestType: *policyService.NewNullableGpuRequestRules(&policyService.GpuRequestRules{
									CanEdit: *policyService.NewNullableBool(convutils.Ref(false)),
								}),
								GpuDevicesRequest: *policyService.NewNullableIntegerRules(&policyService.IntegerRules{
									CanEdit: *policyService.NewNullableBool(convutils.Ref(false)),
								}),
							}),
						}),
					},
				},
			})

			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						CpuCoreRequest:   convutils.Ref(0.1),
						CpuCoreLimit:     convutils.Ref(0.5),
						CpuMemoryRequest: convutils.Ref("100M"),
						CpuMemoryLimit:   convutils.Ref("300M"),
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)

			actualAsMap, _ := gojsonq.LoadJson(payload.Spec.Compute)

			expectAsMap, _ := gojsonq.LoadJson(api.WorkspaceSpecSpecAllOfCompute{
				CpuCoreRequest:    convutils.Ref(0.1),
				CpuCoreLimit:      convutils.Ref(0.5),
				CpuMemoryRequest:  convutils.Ref("100M"),
				CpuMemoryLimit:    convutils.Ref("300M"),
				GpuMemoryRequest:  convutils.Ref("2M"),
				GpuDevicesRequest: convutils.Ref(int32(1)),
				GpuRequestType:    convutils.Ref(api.GPUREQUESTTYPE_MEMORY),
			})

			Expect(actualAsMap).To(Equal(expectAsMap))
		})
	})

	Context("gpu portion for distributed", func() {
		It("should apply defaults correctly", func() {
			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveDistributedPolicyV2: &policyService.DistributedPolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableDistributedPolicyDefaultsV2(&policyService.DistributedPolicyDefaultsV2{
							Worker: *policyService.NewNullableDistributedPolicyDefaultsV2Worker(&policyService.DistributedPolicyDefaultsV2Worker{
								Compute: *policyService.NewNullableSupersetDefaultsAllOfCompute(&policyService.SupersetDefaultsAllOfCompute{
									GpuPortionRequest: *policyService.NewNullableFloat64(convutils.Ref(0.4)),
									GpuDevicesRequest: *policyService.NewNullableInt32(convutils.Ref(int32(1))),
									GpuRequestType:    *policyService.NewNullableGpuRequestType(convutils.Ref(policyService.GPUREQUESTTYPE_PORTION)),
								}),
							}),
							Master: *policyService.NewNullableReplicaDefaultsV2(&policyService.ReplicaDefaultsV2{
								Compute: *policyService.NewNullableSupersetDefaultsAllOfCompute(&policyService.SupersetDefaultsAllOfCompute{
									GpuDevicesRequest: *policyService.NewNullableInt32(convutils.Ref(int32(1))),
								}),
							}),
						}),
					},
				},
			})

			workloadRequest := api.DistributedCreationRequest{
				Name:      "dist-job",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						CpuCoreRequest:   convutils.Ref(0.1),
						CpuCoreLimit:     convutils.Ref(0.5),
						CpuMemoryRequest: convutils.Ref("100M"),
						CpuMemoryLimit:   convutils.Ref("300M"),
					},
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						CpuCoreRequest:   convutils.Ref(0.1),
						CpuCoreLimit:     convutils.Ref(0.5),
						CpuMemoryRequest: convutils.Ref("100M"),
						CpuMemoryLimit:   convutils.Ref("300M"),
					},
				},
			}

			response, _ := distributedService.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)

			actualWorkerAsMap, _ := gojsonq.LoadJson(payload.Spec.Compute)
			actualMasterAsMap, _ := gojsonq.LoadJson(payload.MasterSpec.Get().Compute.Get())

			expectedWorkerAsMap, _ := gojsonq.LoadJson(api.WorkspaceSpecSpecAllOfCompute{
				CpuCoreRequest:          convutils.Ref(0.1),
				CpuCoreLimit:            convutils.Ref(0.5),
				CpuMemoryRequest:        convutils.Ref("100M"),
				CpuMemoryLimit:          convutils.Ref("300M"),
				GpuPortionRequest:       convutils.Ref(0.4),
				GpuMemoryPortionRequest: convutils.Ref(0.4),
				GpuDevicesRequest:       convutils.Ref(int32(1)),
				GpuRequestType:          convutils.Ref(api.GPUREQUESTTYPE_PORTION),
			})

			Expect(actualWorkerAsMap).To(Equal(expectedWorkerAsMap))

			expectedMasterAsMap, _ := gojsonq.LoadJson(api.WorkspaceSpecSpecAllOfCompute{
				CpuCoreRequest:    convutils.Ref(0.1),
				CpuCoreLimit:      convutils.Ref(0.5),
				CpuMemoryRequest:  convutils.Ref("100M"),
				CpuMemoryLimit:    convutils.Ref("300M"),
				GpuDevicesRequest: convutils.Ref(int32(1)),
			})

			Expect(actualMasterAsMap).To(Equal(expectedMasterAsMap))
		})
	})

	Context("restart policy request creation check", func() {
		It("restart policy is nil", func() {
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))
			payload := response.Body.(workloadsApi.SupersetWorkload)
			actualAsMap, _ := gojsonq.LoadJson(payload.Spec)
			expectAsMap, _ := gojsonq.LoadJson(workloadsApi.SupersetSpec{
				Image:     *workloadsApi.NewNullableString(ptr.To("ubuntu")),
				NodePools: []string{"default"},
				Category:  *workloadsApi.NewNullableString(ptr.To("Train")),
			})

			Expect(actualAsMap).To(Equal(expectAsMap))
		})

		It("restart policy is Never", func() {
			neverRestartPolicy, err := api.NewRestartPolicyFromValue("Never")
			Expect(err).ToNot(HaveOccurred())
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image:         convutils.Ref("ubuntu"),
					RestartPolicy: &neverRestartPolicy,
				},
			}

			wlNeverRestartPolicy := "Never"
			_, err = workloadsApi.NewRestartPolicyFromValue(wlNeverRestartPolicy)
			Expect(err).ToNot(HaveOccurred())
			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))
			payload := response.Body.(workloadsApi.SupersetWorkload)
			actualAsMap, _ := gojsonq.LoadJson(payload.Spec)
			expectAsMap, _ := gojsonq.LoadJson(workloadsApi.SupersetSpec{
				Image:         *workloadsApi.NewNullableString(ptr.To("ubuntu")),
				RestartPolicy: *workloadsApi.NewNullableString(ptr.To("Never")),
				NodePools:     []string{"default"},
				Category:      *workloadsApi.NewNullableString(ptr.To("Train")),
			})

			Expect(actualAsMap).To(Equal(expectAsMap))
		})

		Context("default nodepools", func() {
			var (
				workloadRequest api.TrainingCreationRequest
			)

			BeforeEach(func() {
				workloadRequest = api.TrainingCreationRequest{
					Name:      "trainjob",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.TrainingSpecSpec{
						Image: convutils.Ref("ubuntu"),
					},
				}
			})

			It("use nodepools from spec if exist", func() {
				clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
					WithNodePools(randoms.ClusterId, "my-pool"))

				workloadRequest.Spec.NodePools = ptr.To([]string{"my-pool"})

				response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
				Expect(response.Code).To(Equal(http.StatusAccepted))
				payload := response.Body.(workloadsApi.SupersetWorkload)
				actualAsMap, _ := gojsonq.LoadJson(payload.Spec)
				expectAsMap, _ := gojsonq.LoadJson(workloadsApi.SupersetSpec{
					Image:     *workloadsApi.NewNullableString(ptr.To("ubuntu")),
					NodePools: []string{"my-pool"},
					Category:  *workloadsApi.NewNullableString(ptr.To("Train")),
				})

				Expect(actualAsMap).To(Equal(expectAsMap))
			})

			It("use nodepools from project if exist", func() {
				mockProj := GetProjectResponse(*randoms)
				mockProj.DefaultNodePools = []string{"pool-a", "pool-b"}
				org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, mockProj))

				clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
					WithNodePools(randoms.ClusterId, "pool-a", "pool-b", "pool-c"))

				response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
				Expect(response.Code).To(Equal(http.StatusAccepted))
				payload := response.Body.(workloadsApi.SupersetWorkload)
				actualAsMap, _ := gojsonq.LoadJson(payload.Spec)
				expectAsMap, _ := gojsonq.LoadJson(workloadsApi.SupersetSpec{
					Image:     *workloadsApi.NewNullableString(ptr.To("ubuntu")),
					NodePools: []string{"pool-a", "pool-b"},
					Category:  *workloadsApi.NewNullableString(ptr.To("Train")),
				})

				Expect(actualAsMap).To(Equal(expectAsMap))
			})

			It("use deafult nodepool", func() {
				response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
				Expect(response.Code).To(Equal(http.StatusAccepted))
				payload := response.Body.(workloadsApi.SupersetWorkload)
				actualAsMap, _ := gojsonq.LoadJson(payload.Spec)
				expectAsMap, _ := gojsonq.LoadJson(workloadsApi.SupersetSpec{
					Image:     *workloadsApi.NewNullableString(ptr.To("ubuntu")),
					NodePools: []string{"default"},
					Category:  *workloadsApi.NewNullableString(ptr.To("Train")),
				})

				Expect(actualAsMap).To(Equal(expectAsMap))
			})

			It("distributed - master is updated too", func() {
				pytorch, err := workloadsApi.NewDistributedFrameworkFromValue(string(api.DISTRIBUTEDFRAMEWORK_PY_TORCH))
				Expect(err).NotTo(HaveOccurred())

				workloadRequest := api.DistributedCreationRequest{
					Name:      "dist-job",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.DistributedSpecSpec{
						DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
						Image:                convutils.Ref("ubuntu"),
					},
					MasterSpec: &api.MasterSpec{
						Image: convutils.Ref("ubuntu"),
					},
				}

				response, _ := distributedService.CreateDistributed(ctx, workloadRequest)
				Expect(response.Code).To(Equal(http.StatusAccepted))
				payload := response.Body.(workloadsApi.SupersetWorkload)

				// worker
				actualAsMap, _ := gojsonq.LoadJson(payload.Spec)
				expectAsMap, _ := gojsonq.LoadJson(workloadsApi.SupersetSpec{
					DistributedFramework: *workloadsApi.NewNullableDistributedFramework(pytorch),
					Image:                *workloadsApi.NewNullableString(ptr.To("ubuntu")),
					NodePools:            []string{"default"},
					Category:             *workloadsApi.NewNullableString(ptr.To("Train")),
				})
				Expect(actualAsMap).To(Equal(expectAsMap))

				// master
				actualAsMap, _ = gojsonq.LoadJson(payload.MasterSpec.Get())
				expectAsMap, _ = gojsonq.LoadJson(workloadsApi.SupersetSpec{
					Image:     *workloadsApi.NewNullableString(ptr.To("ubuntu")),
					NodePools: []string{"default"},
					Category:  *workloadsApi.NewNullableString(ptr.To("Train")),
				})
				Expect(actualAsMap).To(Equal(expectAsMap))
			})
		})

	})

	Context("gpu portion and gpu portion limit both set to 1", func() {
		It("gpu portion limit should be set to nil", func() {
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuPortionRequest: convutils.Ref(float64(1)),
						GpuPortionLimit:   convutils.Ref(float64(1)),
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))
			payload := response.Body.(workloadsApi.SupersetWorkload)
			Expect(payload.Spec.GetCompute().GpuPortionLimit.IsSet()).To(Equal(false))
		})
		It("gpu portion limit should be set to nil - distributed master spec", func() {
			workloadRequest := api.DistributedCreationRequest{
				Name:      "dist-job",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuPortionRequest: convutils.Ref(float64(1)),
						GpuPortionLimit:   convutils.Ref(float64(1)),
					},
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuPortionRequest: convutils.Ref(float64(1)),
						GpuPortionLimit:   convutils.Ref(float64(1)),
					},
				},
			}

			response, _ := distributedService.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)
			Expect(payload.Spec.GetCompute().GpuPortionLimit.IsSet()).To(Equal(false))
			Expect(payload.GetMasterSpec().Compute.Get().GpuPortionLimit.IsSet()).To(Equal(false))
		})
	})

	Context("JAX spec validation", func() {
		It("no master", func() {
			jax, err := workloadsApi.NewDistributedFrameworkFromValue("JAX")
			Expect(err).ToNot(HaveOccurred())

			workloadRequest := api.DistributedCreationRequest{
				Name:      "dist-job",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_JAX),
					Image:                convutils.Ref("ubuntu"),
				},
			}

			response, _ := distributedService.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))
			payload := response.Body.(workloadsApi.SupersetWorkload)
			actualAsMap, _ := gojsonq.LoadJson(payload.Spec)
			expectAsMap, _ := gojsonq.LoadJson(workloadsApi.SupersetSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(jax),
				Image:                *workloadsApi.NewNullableString(ptr.To("ubuntu")),
				NodePools:            []string{"default"},
				Category:             *workloadsApi.NewNullableString(convutils.Ref("Train")),
			})

			Expect(actualAsMap).To(Equal(expectAsMap))
		})

		It("with master - should fail", func() {
			workloadRequest := api.DistributedCreationRequest{
				Name:      "dist-job",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_JAX),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			response, _ := distributedService.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			submissionError, ok := response.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(ok).To(BeTrue())
			Expect(submissionError.Message).To(Equal("JAX submission cannot have master spec"))
		})
	})

	Context("Data volume validation", func() {
		It("Should resolve data volume IDs to their PVC claim names", func() {
			status := openapi.DatavolumeStatus{}
			status.SetDatavolumePvcName("data-volume-pvc")
			datavolumes.SetMock(datavolumes.NewSimpleMock().WithDataVolumes([]openapi.DatavolumeNoSharedScopes{
				{
					Id:     "data-volume-id",
					Status: status,
				},
			}))

			spec := workloadsApi.NewSupersetSpec()
			spec.SetStorage(workloadsApi.WorkspaceSpecSpecAllOfStorage{
				DataVolume: []workloadsApi.DataVolumeInstance{
					{
						Id:        *workloadsApi.NewNullableString(convutils.Ref("data-volume-id")),
						MountPath: *workloadsApi.NewNullableString(convutils.Ref("/models")),
					},
				},
			})

			claimNames, validationErr := preValidateDataVolumes(ctx, randoms.ProjectIdStr, spec, nil)

			Expect(validationErr).To(BeNil())
			Expect(claimNames).To(Equal(map[string]string{"data-volume-id": "data-volume-pvc"}))
		})

		It("Should pass data volume validation", func() {
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						DataVolume: &[]api.DataVolumeInstance{
							{
								Id:        convutils.Ref("1111"),
								MountPath: convutils.Ref("/home"),
							},
							{
								Id:        convutils.Ref("2222"),
								MountPath: convutils.Ref("/away"),
							},
						},
					},
				},
			}

			// Mocking data volumes since the policy have default data volume instances
			dataVolumesMock := datavolumes.NewSimpleMock().WithDataVolumes([]openapi.DatavolumeNoSharedScopes{
				{
					Id: "1111",
				},
				{
					Id: "2222",
				},
				{
					Id: "3333",
				},
			})
			datavolumes.SetMock(dataVolumesMock)

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))
		})

		It("Should not pass data volume validation", func() {
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						DataVolume: &[]api.DataVolumeInstance{
							{
								Id:        convutils.Ref("1111"),
								MountPath: convutils.Ref("/home"),
							},
							{
								Id:        convutils.Ref("2222"),
								MountPath: convutils.Ref("/away"),
							},
						},
					},
				},
			}

			// Mocking data volumes since the policy have default data volume instances
			dataVolumesMock := datavolumes.NewSimpleMock().WithDataVolumes([]openapi.DatavolumeNoSharedScopes{
				{
					Id: "1111",
				},
			})
			datavolumes.SetMock(dataVolumesMock)

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("Image pull secrets validation", func() {
		It("Should not pass image pull secrets validation", func() {
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					ImagePullSecrets: &[]api.ImagePullSecret{
						{
							Name: "secret-1",
						},
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			submissionError, ok := response.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(ok).To(BeTrue())
			Expect(submissionError.Message).To(Equal("secret named 'secret-1' does not exist"))
		})

		It("Should not pass image pull secrets validation", func() {
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					ImagePullSecrets: &[]api.ImagePullSecret{
						{
							Name:           "secret-1",
							UserCredential: convutils.Ref(true),
						},
					},
				},
			}

			assetsClient.SetMock(assetsClient.NewSimpleMock())

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusNotFound))
			submissionError, ok := response.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(ok).To(BeTrue())
			Expect(submissionError.Message).To(Equal("credential with name 'secret-1' not found"))
		})

		It("Should pass image pull secrets validation", func() {
			k8sObjectsClient.SetMock(k8sObjectsTrackerMock.NewK8SObjectTrackerMock().WithSecrets([]k8sObjectsApi.SecretListInfo{
				{
					Name: "secret-1",
				},
			}))

			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					ImagePullSecrets: &[]api.ImagePullSecret{
						{
							Name: "secret-1",
						},
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))
		})

		It("Should pass image pull secrets validation with user credential", func() {
			k8sObjectsClient.SetMock(k8sObjectsTrackerMock.NewK8SObjectTrackerMock().WithSecrets([]k8sObjectsApi.SecretListInfo{
				{
					Name: "secret-1",
				},
			}))

			credentialListResponseEntry := assetsServiceApi.CredentialsListResponseEntry{
				Meta: assetsServiceApi.AssetMeta{
					Name:      "secret-2",
					Scope:     assetsServiceApi.SCOPE_CLUSTER,
					ClusterId: *assetsServiceApi.NewNullableString(convutils.Ref(randoms.ClusterId)),
				},
				Spec: assetsServiceApi.CredentialsListResponseAssetSpec{
					DockerRegistry: *assetsServiceApi.NewNullableDockerRegistrySpec(
						&assetsServiceApi.DockerRegistrySpec{
							User:     *assetsServiceApi.NewNullableString(convutils.Ref("user")),
							Password: *assetsServiceApi.NewNullableString(convutils.Ref("password")),
							Url:      *assetsServiceApi.NewNullableString(convutils.Ref("url")),
						},
					),
				},
			}

			assetsClient.SetMock(assetsClient.NewSimpleMock().WithCredentials([]assetsServiceApi.CredentialsListResponseEntry{
				credentialListResponseEntry,
			}))

			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					ImagePullSecrets: &[]api.ImagePullSecret{
						{
							Name:           "secret-2",
							UserCredential: convutils.Ref(true),
						},
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))
		})
	})

	Context("UID/GID extraction from context when source is fromIdpToken", func() {
		It("should extract UID and GID from context", func() {
			// Create a context with the UID, GID and supplemental groups and required fields
			ctx := testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})
			ctx = context.WithValue(ctx, runaictx.CfUId, "1000")
			ctx = context.WithValue(ctx, runaictx.CfGId, "1000")

			// Create a test workload request with UidGidSource set to fromIdpToken
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Security: &api.WorkspaceSpecSpecAllOfSecurity{
						UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_IDP_TOKEN),
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)
			security := payload.Spec.GetSecurity()

			// Verify UID and GID were extracted correctly
			Expect(security.RunAsUid.IsSet()).To(BeTrue())
			Expect(*security.RunAsUid.Get()).To(Equal(int64(1000)))
			Expect(security.RunAsGid.IsSet()).To(BeTrue())
			Expect(*security.RunAsGid.Get()).To(Equal(int64(1000)))
		})

		It("should extract UID, GID and supplemental groups from context", func() {
			// Create a context with the UID, GID and supplemental groups
			ctx := testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})
			ctx = context.WithValue(ctx, runaictx.CfUId, "1000")
			ctx = context.WithValue(ctx, runaictx.CfGId, "1000")
			ctx = context.WithValue(ctx, runaictx.CfSupplementalGroups, "2000,3000")

			// Create a test workload request with UidGidSource set to fromIdpToken
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Security: &api.WorkspaceSpecSpecAllOfSecurity{
						UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_IDP_TOKEN),
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)
			security := payload.Spec.GetSecurity()

			// Verify UID, GID and supplemental groups were extracted correctly
			Expect(security.RunAsUid.IsSet()).To(BeTrue())
			Expect(*security.RunAsUid.Get()).To(Equal(int64(1000)))
			Expect(security.RunAsGid.IsSet()).To(BeTrue())
			Expect(*security.RunAsGid.Get()).To(Equal(int64(1000)))
			Expect(security.SupplementalGroups.IsSet()).To(BeTrue())
			Expect(*security.SupplementalGroups.Get()).To(Equal("2000,3000"))
		})

		It("should use UID, GID from the request as a fallback when context values are missing", func() {
			// Create a context without any UID/GID values
			ctx := testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

			// Verify context doesn't have UID/GID values
			Expect(runaictx.GetString(ctx, runaictx.CfUId)).To(BeNil())
			Expect(runaictx.GetString(ctx, runaictx.CfGId)).To(BeNil())

			// Create a test workload request with UidGidSource set to fromIdpToken and values in request
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Security: &api.WorkspaceSpecSpecAllOfSecurity{
						UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_IDP_TOKEN),
						RunAsUid:     convutils.Ref(int64(1234)),
						RunAsGid:     convutils.Ref(int64(4321)),
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)
			security := payload.Spec.GetSecurity()

			// Verify UID and GID were taken from request as fallback
			Expect(security.RunAsUid.IsSet()).To(BeTrue())
			Expect(*security.RunAsUid.Get()).To(Equal(int64(1234)))
			Expect(security.RunAsGid.IsSet()).To(BeTrue())
			Expect(*security.RunAsGid.Get()).To(Equal(int64(4321)))
		})

		It("should use partial fallback when only one value is missing from context", func() {
			// Create a context with only UID
			ctx := testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})
			ctx = context.WithValue(ctx, runaictx.CfUId, "1000")

			// Create a test workload request with UidGidSource set to fromIdpToken and GID in request
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Security: &api.WorkspaceSpecSpecAllOfSecurity{
						UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_IDP_TOKEN),
						RunAsGid:     convutils.Ref(int64(4321)),
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)
			security := payload.Spec.GetSecurity()

			// Verify UID from context and GID from request
			Expect(security.RunAsUid.IsSet()).To(BeTrue())
			Expect(*security.RunAsUid.Get()).To(Equal(int64(1000)))
			Expect(security.RunAsGid.IsSet()).To(BeTrue())
			Expect(*security.RunAsGid.Get()).To(Equal(int64(4321)))
		})

		It("should handle invalid UID/GID values gracefully", func() {
			// Create a context with invalid UID/GID values
			ctx := testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})
			ctx = context.WithValue(ctx, runaictx.CfUId, "invalid")
			ctx = context.WithValue(ctx, runaictx.CfGId, "invalid")

			// Create a test workload request with UidGidSource set to fromIdpToken and valid values in request
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Security: &api.WorkspaceSpecSpecAllOfSecurity{
						UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_IDP_TOKEN),
						RunAsUid:     convutils.Ref(int64(1234)),
						RunAsGid:     convutils.Ref(int64(4321)),
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)
			security := payload.Spec.GetSecurity()

			// Verify values were taken from request since context values were invalid
			Expect(security.RunAsUid.IsSet()).To(BeTrue())
			Expect(*security.RunAsUid.Get()).To(Equal(int64(1234)))
			Expect(security.RunAsGid.IsSet()).To(BeTrue())
			Expect(*security.RunAsGid.Get()).To(Equal(int64(4321)))
		})

		It("should handle invalid supplemental groups gracefully", func() {
			// Create a context with mixed valid and invalid supplemental groups
			ctx := testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})
			ctx = context.WithValue(ctx, runaictx.CfUId, "1000")
			ctx = context.WithValue(ctx, runaictx.CfGId, "1000")
			ctx = context.WithValue(ctx, runaictx.CfSupplementalGroups, "2000,invalid,3000")

			// Create a test workload request with UidGidSource set to fromIdpToken
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Security: &api.WorkspaceSpecSpecAllOfSecurity{
						UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_IDP_TOKEN),
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)
			security := payload.Spec.GetSecurity()

			// Verify only valid supplemental groups were included
			Expect(security.RunAsUid.IsSet()).To(BeTrue())
			Expect(*security.RunAsUid.Get()).To(Equal(int64(1000)))
			Expect(security.RunAsGid.IsSet()).To(BeTrue())
			Expect(*security.RunAsGid.Get()).To(Equal(int64(1000)))
			Expect(security.SupplementalGroups.IsSet()).To(BeTrue())
			Expect(*security.SupplementalGroups.Get()).To(Equal("2000,3000"))
		})

		It("should succeed even when no UID/GID values are provided", func() {
			// Create a context without any UID/GID values
			ctx := testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

			// Create a test workload request with UidGidSource set to fromIdpToken but no values
			workloadRequest := api.TrainingCreationRequest{
				Name:      "trainjob",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Security: &api.WorkspaceSpecSpecAllOfSecurity{
						UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_IDP_TOKEN),
					},
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)
			security := payload.Spec.GetSecurity()

			// Verify no UID/GID values were set
			Expect(security.RunAsUid.IsSet()).To(BeFalse())
			Expect(security.RunAsGid.IsSet()).To(BeFalse())
		})
	})

	Context("distributed inference policy application", func() {
		It("should apply comprehensive policy defaults correctly", func() {
			mockTenantSettingsClient.On("GetGpuResourceOptimization", mock.Anything).Return(true, nil).Once()
			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveDistributedInferencePolicyV2: &policyService.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableDistributedInferenceDefaultsV2(&policyService.DistributedInferenceDefaultsV2{
							NodePools:     []string{"default"},
							Category:      *policyService.NewNullableString(convutils.Ref("Build")),
							RestartPolicy: *policyService.NewNullableDistributedInferenceRestartPolicy(convutils.Ref(policyService.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART)),
							ServingPort: *policyService.NewNullableDistributedInferenceServingPort(&policyService.DistributedInferenceServingPort{
								Port:              *policyService.NewNullableInt32(convutils.Ref(int32(8080))),
								Protocol:          *policyService.NewNullableDistributedInferenceServingPortProtocol(convutils.Ref(policyService.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP)),
								AuthorizationType: *policyService.NewNullableDistributedInferenceServingPortAccessAuthorizationTypeEnum(convutils.Ref(policyService.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_PUBLIC)),
								ExposeExternally:  *policyService.NewNullableBool(convutils.Ref(true)),
								ExposedUrl:        *policyService.NewNullableString(convutils.Ref("string")),
							}),
							StartupPolicy: *policyService.NewNullableDistributedInferenceStartupPolicy(convutils.Ref(policyService.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED)),
							Workers:       *policyService.NewNullableInt32(convutils.Ref(int32(4))),
							Replicas:      *policyService.NewNullableInt32(convutils.Ref(int32(2))),
							Leader: *policyService.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyService.DistributedInferenceLeaderWorkerDefaultsV2{
								Command:         *policyService.NewNullableString(convutils.Ref("python")),
								Args:            *policyService.NewNullableString(convutils.Ref("-x my-script.py")),
								Image:           *policyService.NewNullableString(convutils.Ref("python:3.8")),
								ImagePullPolicy: *policyService.NewNullableImagePullPolicy(convutils.Ref(policyService.IMAGEPULLPOLICY_ALWAYS)),
								WorkingDir:      *policyService.NewNullableString(convutils.Ref("/home/myfolder")),
								CreateHomeDir:   *policyService.NewNullableBool(convutils.Ref(true)),
								NodeType:        *policyService.NewNullableString(convutils.Ref("my-node-type")),
								Compute: *policyService.NewNullableSupersetDefaultsAllOfCompute(&policyService.SupersetDefaultsAllOfCompute{
									GpuDevicesRequest: *policyService.NewNullableInt32(convutils.Ref(int32(1))),
									GpuRequestType:    *policyService.NewNullableGpuRequestType(convutils.Ref(policyService.GPUREQUESTTYPE_PORTION)),
									GpuPortionRequest: *policyService.NewNullableFloat64(convutils.Ref(0.5)),
									GpuPortionLimit:   *policyService.NewNullableFloat64(convutils.Ref(0.5)),
									CpuCoreRequest:    *policyService.NewNullableFloat64(convutils.Ref(0.5)),
									CpuCoreLimit:      *policyService.NewNullableFloat64(convutils.Ref(2.0)),
									CpuMemoryRequest:  *policyService.NewNullableString(convutils.Ref("20M")),
									CpuMemoryLimit:    *policyService.NewNullableString(convutils.Ref("30M")),
									LargeShmRequest:   *policyService.NewNullableBool(convutils.Ref(false)),
								}),
								Security: *policyService.NewNullableDistributedInferenceLeaderWorkerDefaultsV2Security(&policyService.DistributedInferenceLeaderWorkerDefaultsV2Security{
									UidGidSource:           *policyService.NewNullableUidGidSource(convutils.Ref(policyService.UIDGIDSOURCE_FROM_THE_IMAGE)),
									Capabilities:           []policyService.Capability{policyService.CAPABILITY_CHOWN, policyService.CAPABILITY_KILL},
									SeccompProfileType:     *policyService.NewNullableSeccompProfileType(convutils.Ref(policyService.SECCOMPPROFILETYPE_RUNTIME_DEFAULT)),
									RunAsNonRoot:           *policyService.NewNullableBool(convutils.Ref(true)),
									ReadOnlyRootFilesystem: *policyService.NewNullableBool(convutils.Ref(false)),
									RunAsUid:               *policyService.NewNullableInt64(convutils.Ref(int64(500))),
									RunAsGid:               *policyService.NewNullableInt64(convutils.Ref(int64(30))),
									SupplementalGroups:     *policyService.NewNullableString(convutils.Ref("2,3,5,8")),
								}),
								EnvironmentVariables: *policyService.NewNullableEnvironmentVariablesDefaults(&policyService.EnvironmentVariablesDefaults{
									Instances: []policyService.EnvironmentVariable{
										{
											Name:        *policyService.NewNullableString(convutils.Ref("HOME")),
											Value:       *policyService.NewNullableString(convutils.Ref("test")),
											Exclude:     *policyService.NewNullableBool(convutils.Ref(false)),
											Description: *policyService.NewNullableString(convutils.Ref("Home directory of the user.")),
										},
									},
								}),
								Annotations: *policyService.NewNullableAnnotationsDefaults(&policyService.AnnotationsDefaults{
									Instances: []policyService.Annotation{
										{
											Name:    *policyService.NewNullableString(convutils.Ref("billing")),
											Value:   *policyService.NewNullableString(convutils.Ref("my-billing-unit")),
											Exclude: *policyService.NewNullableBool(convutils.Ref(false)),
										},
									},
								}),
								Labels: *policyService.NewNullableLabelsDefaults(&policyService.LabelsDefaults{
									Instances: []policyService.Label{
										{
											Name:    *policyService.NewNullableString(convutils.Ref("stage")),
											Value:   *policyService.NewNullableString(convutils.Ref("initial-research")),
											Exclude: *policyService.NewNullableBool(convutils.Ref(false)),
										},
									},
								}),
								ImagePullSecrets: *policyService.NewNullableImagePullSecretsDefaults(&policyService.ImagePullSecretsDefaults{
									Instances: []policyService.ImagePullSecret{
										{
											Name:           convutils.Ref("Always"),
											UserCredential: *policyService.NewNullableBool(convutils.Ref(true)),
											Exclude:        *policyService.NewNullableBool(convutils.Ref(false)),
										},
									},
								}),
								Storage: *policyService.NewNullableDistributedInferenceLeaderWorkerDefaultsV2Storage(&policyService.DistributedInferenceLeaderWorkerDefaultsV2Storage{
									ConfigMapVolume: *policyService.NewNullableConfigMapsDefaults(&policyService.ConfigMapsDefaults{
										Instances: []policyService.ConfigMapInstance{
											{
												Name:        *policyService.NewNullableString(convutils.Ref("storage-instance-a")),
												ConfigMap:   *policyService.NewNullableString(convutils.Ref("string")),
												MountPath:   *policyService.NewNullableString(convutils.Ref("string")),
												SubPath:     *policyService.NewNullableString(convutils.Ref("string")),
												DefaultMode: *policyService.NewNullableString(convutils.Ref("0644")),
												Exclude:     *policyService.NewNullableBool(convutils.Ref(false)),
											},
										},
									}),
									EmptyDirVolume: *policyService.NewNullableEmptyDirsDefaults(&policyService.EmptyDirsDefaults{
										Instances: []policyService.EmptyDirInstance{
											{
												Name:      *policyService.NewNullableString(convutils.Ref("storage-instance-a")),
												Path:      *policyService.NewNullableString(convutils.Ref("/mnt/emptydir")),
												SizeLimit: *policyService.NewNullableString(convutils.Ref("1G")),
												Exclude:   *policyService.NewNullableBool(convutils.Ref(false)),
											},
										},
									}),
									Pvc: *policyService.NewNullablePvcsDefaults(&policyService.PvcsDefaults{
										Instances: []policyService.PvcInstance{
											{
												Name:        *policyService.NewNullableString(convutils.Ref("storage-instance-a")),
												Path:        *policyService.NewNullableString(convutils.Ref("/container/my-claim")),
												ExistingPvc: *policyService.NewNullableBool(convutils.Ref(false)),
												ClaimName:   *policyService.NewNullableString(convutils.Ref("my-claim")),
												ReadOnly:    *policyService.NewNullableBool(convutils.Ref(false)),
												Ephemeral:   *policyService.NewNullableBool(convutils.Ref(false)),
												DataSharing: *policyService.NewNullableBool(convutils.Ref(false)),
												Exclude:     *policyService.NewNullableBool(convutils.Ref(false)),
												ClaimInfo: *policyService.NewNullableClaimInfo(&policyService.ClaimInfo{
													Size:         *policyService.NewNullableString(convutils.Ref("1G")),
													StorageClass: *policyService.NewNullableString(convutils.Ref("my-storage-class")),
													VolumeMode:   *policyService.NewNullablePvcVolumeMode(convutils.Ref(policyService.PVCVOLUMEMODE_FILESYSTEM)),
													AccessModes: *policyService.NewNullablePvcAccessModes(&policyService.PvcAccessModes{
														ReadWriteOnce: *policyService.NewNullableBool(convutils.Ref(true)),
														ReadOnlyMany:  *policyService.NewNullableBool(convutils.Ref(false)),
														ReadWriteMany: *policyService.NewNullableBool(convutils.Ref(false)),
													}),
												}),
											},
										},
									}),
									SecretVolume: *policyService.NewNullableSecretsDefaults(&policyService.SecretsDefaults{
										Instances: []policyService.SecretInstance{
											{
												Name:        *policyService.NewNullableString(convutils.Ref("storage-instance-a")),
												MountPath:   *policyService.NewNullableString(convutils.Ref("string")),
												DefaultMode: *policyService.NewNullableString(convutils.Ref("0644")),
												Secret:      *policyService.NewNullableString(convutils.Ref("string")),
												Exclude:     *policyService.NewNullableBool(convutils.Ref(false)),
											},
										},
									}),
								}),
							}),
							Worker: *policyService.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyService.DistributedInferenceLeaderWorkerDefaultsV2{
								Command:         *policyService.NewNullableString(convutils.Ref("python")),
								Args:            *policyService.NewNullableString(convutils.Ref("-x my-script.py")),
								Image:           *policyService.NewNullableString(convutils.Ref("python:3.8")),
								ImagePullPolicy: *policyService.NewNullableImagePullPolicy(convutils.Ref(policyService.IMAGEPULLPOLICY_ALWAYS)),
								WorkingDir:      *policyService.NewNullableString(convutils.Ref("/home/myfolder")),
								CreateHomeDir:   *policyService.NewNullableBool(convutils.Ref(true)),
								NodeType:        *policyService.NewNullableString(convutils.Ref("my-node-type")),
								Compute: *policyService.NewNullableSupersetDefaultsAllOfCompute(&policyService.SupersetDefaultsAllOfCompute{
									GpuDevicesRequest: *policyService.NewNullableInt32(convutils.Ref(int32(1))),
									GpuRequestType:    *policyService.NewNullableGpuRequestType(convutils.Ref(policyService.GPUREQUESTTYPE_PORTION)),
									GpuPortionRequest: *policyService.NewNullableFloat64(convutils.Ref(0.5)),
									GpuPortionLimit:   *policyService.NewNullableFloat64(convutils.Ref(0.5)),
									CpuCoreRequest:    *policyService.NewNullableFloat64(convutils.Ref(0.5)),
									CpuCoreLimit:      *policyService.NewNullableFloat64(convutils.Ref(2.0)),
									CpuMemoryRequest:  *policyService.NewNullableString(convutils.Ref("20M")),
									CpuMemoryLimit:    *policyService.NewNullableString(convutils.Ref("30M")),
									LargeShmRequest:   *policyService.NewNullableBool(convutils.Ref(false)),
								}),
								Security: *policyService.NewNullableDistributedInferenceLeaderWorkerDefaultsV2Security(&policyService.DistributedInferenceLeaderWorkerDefaultsV2Security{
									UidGidSource:           *policyService.NewNullableUidGidSource(convutils.Ref(policyService.UIDGIDSOURCE_FROM_THE_IMAGE)),
									Capabilities:           []policyService.Capability{policyService.CAPABILITY_CHOWN, policyService.CAPABILITY_KILL},
									SeccompProfileType:     *policyService.NewNullableSeccompProfileType(convutils.Ref(policyService.SECCOMPPROFILETYPE_RUNTIME_DEFAULT)),
									RunAsNonRoot:           *policyService.NewNullableBool(convutils.Ref(true)),
									ReadOnlyRootFilesystem: *policyService.NewNullableBool(convutils.Ref(false)),
									RunAsUid:               *policyService.NewNullableInt64(convutils.Ref(int64(500))),
									RunAsGid:               *policyService.NewNullableInt64(convutils.Ref(int64(30))),
									SupplementalGroups:     *policyService.NewNullableString(convutils.Ref("2,3,5,8")),
								}),
								EnvironmentVariables: *policyService.NewNullableEnvironmentVariablesDefaults(&policyService.EnvironmentVariablesDefaults{
									Instances: []policyService.EnvironmentVariable{
										{
											Name:        *policyService.NewNullableString(convutils.Ref("HOME")),
											Value:       *policyService.NewNullableString(convutils.Ref("/home/my-folder")),
											Exclude:     *policyService.NewNullableBool(convutils.Ref(false)),
											Description: *policyService.NewNullableString(convutils.Ref("Home directory of the user.")),
										},
									},
								}),
								Annotations: *policyService.NewNullableAnnotationsDefaults(&policyService.AnnotationsDefaults{
									Instances: []policyService.Annotation{
										{
											Name:    *policyService.NewNullableString(convutils.Ref("billing")),
											Value:   *policyService.NewNullableString(convutils.Ref("my-billing-unit")),
											Exclude: *policyService.NewNullableBool(convutils.Ref(false)),
										},
									},
								}),
								Labels: *policyService.NewNullableLabelsDefaults(&policyService.LabelsDefaults{
									Instances: []policyService.Label{
										{
											Name:    *policyService.NewNullableString(convutils.Ref("stage")),
											Value:   *policyService.NewNullableString(convutils.Ref("initial-research")),
											Exclude: *policyService.NewNullableBool(convutils.Ref(false)),
										},
									},
								}),
								ImagePullSecrets: *policyService.NewNullableImagePullSecretsDefaults(&policyService.ImagePullSecretsDefaults{
									Instances: []policyService.ImagePullSecret{
										{
											Name:           convutils.Ref("Always"),
											UserCredential: *policyService.NewNullableBool(convutils.Ref(true)),
											Exclude:        *policyService.NewNullableBool(convutils.Ref(false)),
										},
									},
								}),
								Storage: *policyService.NewNullableDistributedInferenceLeaderWorkerDefaultsV2Storage(&policyService.DistributedInferenceLeaderWorkerDefaultsV2Storage{
									ConfigMapVolume: *policyService.NewNullableConfigMapsDefaults(&policyService.ConfigMapsDefaults{
										Instances: []policyService.ConfigMapInstance{
											{
												Name:        *policyService.NewNullableString(convutils.Ref("storage-instance-a")),
												ConfigMap:   *policyService.NewNullableString(convutils.Ref("string")),
												MountPath:   *policyService.NewNullableString(convutils.Ref("string")),
												SubPath:     *policyService.NewNullableString(convutils.Ref("string")),
												DefaultMode: *policyService.NewNullableString(convutils.Ref("0644")),
												Exclude:     *policyService.NewNullableBool(convutils.Ref(false)),
											},
										},
									}),
									EmptyDirVolume: *policyService.NewNullableEmptyDirsDefaults(&policyService.EmptyDirsDefaults{
										Instances: []policyService.EmptyDirInstance{
											{
												Name:      *policyService.NewNullableString(convutils.Ref("storage-instance-a")),
												Path:      *policyService.NewNullableString(convutils.Ref("/mnt/emptydir")),
												SizeLimit: *policyService.NewNullableString(convutils.Ref("1G")),
												Exclude:   *policyService.NewNullableBool(convutils.Ref(false)),
											},
										},
									}),
									Pvc: *policyService.NewNullablePvcsDefaults(&policyService.PvcsDefaults{
										Instances: []policyService.PvcInstance{
											{
												Name:        *policyService.NewNullableString(convutils.Ref("storage-instance-a")),
												Path:        *policyService.NewNullableString(convutils.Ref("/container/my-claim")),
												ExistingPvc: *policyService.NewNullableBool(convutils.Ref(false)),
												ClaimName:   *policyService.NewNullableString(convutils.Ref("my-claim")),
												ReadOnly:    *policyService.NewNullableBool(convutils.Ref(false)),
												Ephemeral:   *policyService.NewNullableBool(convutils.Ref(false)),
												DataSharing: *policyService.NewNullableBool(convutils.Ref(false)),
												Exclude:     *policyService.NewNullableBool(convutils.Ref(false)),
												ClaimInfo: *policyService.NewNullableClaimInfo(&policyService.ClaimInfo{
													Size:         *policyService.NewNullableString(convutils.Ref("1G")),
													StorageClass: *policyService.NewNullableString(convutils.Ref("my-storage-class")),
													VolumeMode:   *policyService.NewNullablePvcVolumeMode(convutils.Ref(policyService.PVCVOLUMEMODE_FILESYSTEM)),
													AccessModes: *policyService.NewNullablePvcAccessModes(&policyService.PvcAccessModes{
														ReadWriteOnce: *policyService.NewNullableBool(convutils.Ref(true)),
														ReadOnlyMany:  *policyService.NewNullableBool(convutils.Ref(false)),
														ReadWriteMany: *policyService.NewNullableBool(convutils.Ref(false)),
													}),
												}),
											},
										},
									}),
									SecretVolume: *policyService.NewNullableSecretsDefaults(&policyService.SecretsDefaults{
										Instances: []policyService.SecretInstance{
											{
												Name:        *policyService.NewNullableString(convutils.Ref("storage-instance-a")),
												MountPath:   *policyService.NewNullableString(convutils.Ref("string")),
												DefaultMode: *policyService.NewNullableString(convutils.Ref("0644")),
												Secret:      *policyService.NewNullableString(convutils.Ref("string")),
												Exclude:     *policyService.NewNullableBool(convutils.Ref(false)),
											},
										},
									}),
								}),
							}),
						}),
					},
				},
			})

			distributedInferenceService := NewDistributedInferenceService(mockTenantSettingsClient)
			workloadRequest := api.DistributedInferenceCreationRequest{
				Name:      "d2",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedInferenceSpecSpec{
					Workers:  convutils.Ref(int32(1)),
					Replicas: convutils.Ref(int32(2)),
					ServingPort: &api.DistributedInferenceServingPort{
						Port: convutils.Ref(int32(8080)),
					},
					Leader: &api.DistributedInferenceLeaderWorkerSpec{
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuMemoryLimit: convutils.Ref("1Gi"),
						},
					},
					Worker: &api.DistributedInferenceLeaderWorkerSpec{
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuMemoryLimit: convutils.Ref("1Gi"),
						},
					},
				},
			}

			response, err := distributedInferenceService.CreateDistributedInference(ctx, workloadRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)

			// Verify that policy defaults were applied while user overrides were preserved
			Expect(payload.Spec.NodePools).To(Equal([]string{"default"}))
			Expect(payload.Spec.Category.IsSet()).To(BeTrue())
			Expect(*payload.Spec.Category.Get()).To(Equal("Build"))
			Expect(payload.Spec.PriorityClass.IsSet()).To(BeFalse())

			// Verify restart and startup policies from defaults
			Expect(payload.Spec.RestartPolicy.IsSet()).To(BeTrue())
			Expect(*payload.Spec.RestartPolicy.Get()).To(Equal(string(workloadsApi.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART)))
			Expect(payload.Spec.StartupPolicy.IsSet()).To(BeTrue())
			Expect(string(*payload.Spec.StartupPolicy.Get())).To(Equal(string(workloadsApi.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED)))

			// Verify serving port - user override for port and exposeExternally, policy defaults for other fields
			Expect(payload.Spec.ServingPort.IsSet()).To(BeTrue())
			servingPort := payload.Spec.ServingPort.Get()
			Expect(servingPort.Port.IsSet()).To(BeTrue())
			Expect(*servingPort.Port.Get()).To(Equal(int32(8080)))
			Expect(servingPort.ExposeExternally.IsSet()).To(BeTrue())
			Expect(*servingPort.ExposeExternally.Get()).To(Equal(true))
			Expect(servingPort.Protocol.IsSet()).To(BeTrue())
			Expect(*servingPort.Protocol.Get()).To(Equal(string(workloadsApi.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP)))
			Expect(servingPort.AuthorizationType.IsSet()).To(BeTrue())
			Expect(*servingPort.AuthorizationType.Get()).To(Equal(string(workloadsApi.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_PUBLIC)))

			// Verify workers and replicas - user overrides
			Expect(payload.Spec.Workers.IsSet()).To(BeTrue())
			Expect(*payload.Spec.Workers.Get()).To(Equal(int32(1)))
			Expect(payload.Spec.Replicas.IsSet()).To(BeTrue())
			Expect(*payload.Spec.Replicas.Get()).To(Equal(int32(2)))

			// Verify leader spec - policy defaults combined with user overrides
			Expect(payload.Spec.Leader.IsSet()).To(BeTrue())
			leaderSpec := payload.Spec.Leader.Get()
			Expect(leaderSpec.Command.IsSet()).To(BeTrue())
			Expect(*leaderSpec.Command.Get()).To(Equal("python"))
			Expect(leaderSpec.Args.IsSet()).To(BeTrue())
			Expect(*leaderSpec.Args.Get()).To(Equal("-x my-script.py"))
			Expect(leaderSpec.Image.IsSet()).To(BeTrue())
			Expect(*leaderSpec.Image.Get()).To(Equal("python:3.8"))
			Expect(leaderSpec.WorkingDir.IsSet()).To(BeTrue())
			Expect(*leaderSpec.WorkingDir.Get()).To(Equal("/home/myfolder"))

			// Verify leader compute - user override for gpuMemoryLimit, policy defaults for others
			Expect(leaderSpec.Compute.IsSet()).To(BeTrue())
			leaderCompute := leaderSpec.Compute.Get()
			Expect(leaderCompute.GpuMemoryLimit.IsSet()).To(BeTrue())
			Expect(*leaderCompute.GpuMemoryLimit.Get()).To(Equal("1Gi"))
			Expect(leaderCompute.GpuDevicesRequest.IsSet()).To(BeTrue())
			Expect(*leaderCompute.GpuDevicesRequest.Get()).To(Equal(int32(1)))
			Expect(leaderCompute.GpuRequestType.IsSet()).To(BeTrue())
			Expect(string(*leaderCompute.GpuRequestType.Get())).To(Equal(string(workloadsApi.GPUREQUESTTYPE_PORTION)))

			Expect(payload.Spec.Worker.IsSet()).To(BeTrue())
			workerSpec := payload.Spec.Worker.Get()
			Expect(workerSpec.Command.IsSet()).To(BeTrue())
			Expect(*workerSpec.Command.Get()).To(Equal("python"))
			Expect(workerSpec.Image.IsSet()).To(BeTrue())
			Expect(*workerSpec.Image.Get()).To(Equal("python:3.8"))

			Expect(workerSpec.Compute.IsSet()).To(BeTrue())
			workerCompute := workerSpec.Compute.Get()
			Expect(workerCompute.GpuMemoryLimit.IsSet()).To(BeTrue())
			Expect(*workerCompute.GpuMemoryLimit.Get()).To(Equal("1Gi"))
			Expect(workerCompute.GpuDevicesRequest.IsSet()).To(BeTrue())
			Expect(*workerCompute.GpuDevicesRequest.Get()).To(Equal(int32(1)))

			Expect(leaderSpec.Security.IsSet()).To(BeTrue())
			leaderSecurity := leaderSpec.Security.Get()
			Expect(leaderSecurity.UidGidSource.IsSet()).To(BeTrue())
			Expect(string(*leaderSecurity.UidGidSource.Get())).To(Equal(string(workloadsApi.UIDGIDSOURCE_FROM_THE_IMAGE)))
			Expect(leaderSecurity.RunAsUid.IsSet()).To(BeTrue())
			Expect(*leaderSecurity.RunAsUid.Get()).To(Equal(int64(500)))
			Expect(leaderSecurity.RunAsGid.IsSet()).To(BeTrue())
			Expect(*leaderSecurity.RunAsGid.Get()).To(Equal(int64(30)))

			Expect(len(leaderSpec.EnvironmentVariables)).To(BeNumerically(">", 0))
			homeEnvVar := leaderSpec.EnvironmentVariables[0]
			Expect(homeEnvVar.Name.IsSet()).To(BeTrue())
			Expect(*homeEnvVar.Name.Get()).To(Equal("HOME"))
			Expect(homeEnvVar.Value.IsSet()).To(BeTrue())
			Expect(*homeEnvVar.Value.Get()).To(Equal("test"))
			Expect(homeEnvVar.Exclude.IsSet()).To(BeTrue())
			Expect(*homeEnvVar.Exclude.Get()).To(Equal(false))
			Expect(homeEnvVar.Description.IsSet()).To(BeTrue())
			Expect(*homeEnvVar.Description.Get()).To(Equal("Home directory of the user."))

			Expect(len(leaderSpec.Annotations)).To(BeNumerically(">", 0))
			billingAnnotation := leaderSpec.Annotations[0]
			Expect(billingAnnotation.Name.IsSet()).To(BeTrue())
			Expect(*billingAnnotation.Name.Get()).To(Equal("billing"))
			Expect(billingAnnotation.Value.IsSet()).To(BeTrue())
			Expect(*billingAnnotation.Value.Get()).To(Equal("my-billing-unit"))
			Expect(billingAnnotation.Exclude.IsSet()).To(BeTrue())
			Expect(*billingAnnotation.Exclude.Get()).To(Equal(false))

			Expect(len(leaderSpec.Labels)).To(BeNumerically(">", 0))
			stageLabel := leaderSpec.Labels[0]
			Expect(stageLabel.Name.IsSet()).To(BeTrue())
			Expect(*stageLabel.Name.Get()).To(Equal("stage"))
			Expect(stageLabel.Value.IsSet()).To(BeTrue())
			Expect(*stageLabel.Value.Get()).To(Equal("initial-research"))
			Expect(stageLabel.Exclude.IsSet()).To(BeTrue())
			Expect(*stageLabel.Exclude.Get()).To(Equal(false))

			Expect(len(leaderSpec.ImagePullSecrets)).To(BeNumerically(">", 0))
			imagePullSecret := leaderSpec.ImagePullSecrets[0]
			Expect(imagePullSecret.Name).ToNot(BeNil())
			Expect(*imagePullSecret.Name).To(Equal("Always"))
			Expect(imagePullSecret.UserCredential.IsSet()).To(BeTrue())
			Expect(*imagePullSecret.UserCredential.Get()).To(Equal(true))
			Expect(imagePullSecret.Exclude.IsSet()).To(BeTrue())
			Expect(*imagePullSecret.Exclude.Get()).To(Equal(false))

			Expect(leaderSpec.Storage.IsSet()).To(BeTrue())
			leaderStorage := leaderSpec.Storage.Get()

			Expect(len(leaderStorage.ConfigMapVolume)).To(BeNumerically(">", 0))
			configMapVolume := leaderStorage.ConfigMapVolume[0]
			Expect(configMapVolume.Name.IsSet()).To(BeTrue())
			Expect(*configMapVolume.Name.Get()).To(Equal("storage-instance-a"))
			Expect(configMapVolume.ConfigMap.IsSet()).To(BeTrue())
			Expect(*configMapVolume.ConfigMap.Get()).To(Equal("string"))
			Expect(configMapVolume.DefaultMode.IsSet()).To(BeTrue())
			Expect(*configMapVolume.DefaultMode.Get()).To(Equal("0644"))
			Expect(configMapVolume.Exclude.IsSet()).To(BeTrue())
			Expect(*configMapVolume.Exclude.Get()).To(Equal(false))

			Expect(len(leaderStorage.EmptyDirVolume)).To(BeNumerically(">", 0))
			emptyDirVolume := leaderStorage.EmptyDirVolume[0]
			Expect(emptyDirVolume.Name.IsSet()).To(BeTrue())
			Expect(*emptyDirVolume.Name.Get()).To(Equal("storage-instance-a"))
			Expect(emptyDirVolume.Path.IsSet()).To(BeTrue())
			Expect(*emptyDirVolume.Path.Get()).To(Equal("/mnt/emptydir"))
			Expect(emptyDirVolume.SizeLimit.IsSet()).To(BeTrue())
			Expect(*emptyDirVolume.SizeLimit.Get()).To(Equal("1G"))
			Expect(emptyDirVolume.Exclude.IsSet()).To(BeTrue())
			Expect(*emptyDirVolume.Exclude.Get()).To(Equal(false))

			Expect(len(leaderStorage.Pvc)).To(BeNumerically(">", 0))
			pvcInstance := leaderStorage.Pvc[0]
			Expect(pvcInstance.Name.IsSet()).To(BeTrue())
			Expect(*pvcInstance.Name.Get()).To(Equal("storage-instance-a"))
			Expect(pvcInstance.Path.IsSet()).To(BeTrue())
			Expect(*pvcInstance.Path.Get()).To(Equal("/container/my-claim"))
			Expect(pvcInstance.ClaimName.IsSet()).To(BeTrue())
			Expect(*pvcInstance.ClaimName.Get()).To(Equal("my-claim"))
			Expect(pvcInstance.ExistingPvc.IsSet()).To(BeTrue())
			Expect(*pvcInstance.ExistingPvc.Get()).To(Equal(false))
			Expect(pvcInstance.ReadOnly.IsSet()).To(BeTrue())
			Expect(*pvcInstance.ReadOnly.Get()).To(Equal(false))
			Expect(pvcInstance.Ephemeral.IsSet()).To(BeTrue())
			Expect(*pvcInstance.Ephemeral.Get()).To(Equal(false))
			Expect(pvcInstance.DataSharing.IsSet()).To(BeTrue())
			Expect(*pvcInstance.DataSharing.Get()).To(Equal(false))
			Expect(pvcInstance.Exclude.IsSet()).To(BeTrue())
			Expect(*pvcInstance.Exclude.Get()).To(Equal(false))

			Expect(pvcInstance.ClaimInfo.IsSet()).To(BeTrue())
			pvcClaimInfo := pvcInstance.ClaimInfo.Get()
			Expect(pvcClaimInfo.Size.IsSet()).To(BeTrue())
			Expect(*pvcClaimInfo.Size.Get()).To(Equal("1G"))
			Expect(pvcClaimInfo.StorageClass.IsSet()).To(BeTrue())
			Expect(*pvcClaimInfo.StorageClass.Get()).To(Equal("my-storage-class"))
			Expect(pvcClaimInfo.VolumeMode.IsSet()).To(BeTrue())
			Expect(string(*pvcClaimInfo.VolumeMode.Get())).To(Equal(string(workloadsApi.PVCVOLUMEMODE_FILESYSTEM)))
			Expect(pvcClaimInfo.AccessModes.IsSet()).To(BeTrue())
			accessModes := pvcClaimInfo.AccessModes.Get()
			Expect(accessModes.ReadWriteOnce.IsSet()).To(BeTrue())
			Expect(*accessModes.ReadWriteOnce.Get()).To(Equal(true))
			Expect(accessModes.ReadOnlyMany.IsSet()).To(BeTrue())
			Expect(*accessModes.ReadOnlyMany.Get()).To(Equal(false))
			Expect(accessModes.ReadWriteMany.IsSet()).To(BeTrue())
			Expect(*accessModes.ReadWriteMany.Get()).To(Equal(false))

			Expect(len(leaderStorage.SecretVolume)).To(BeNumerically(">", 0))
			secretVolume := leaderStorage.SecretVolume[0]
			Expect(secretVolume.Name.IsSet()).To(BeTrue())
			Expect(*secretVolume.Name.Get()).To(Equal("storage-instance-a"))
			Expect(secretVolume.MountPath.IsSet()).To(BeTrue())
			Expect(*secretVolume.MountPath.Get()).To(Equal("string"))
			Expect(secretVolume.DefaultMode.IsSet()).To(BeTrue())
			Expect(*secretVolume.DefaultMode.Get()).To(Equal("0644"))
			Expect(secretVolume.Secret.IsSet()).To(BeTrue())
			Expect(*secretVolume.Secret.Get()).To(Equal("string"))
			Expect(secretVolume.Exclude.IsSet()).To(BeTrue())
			Expect(*secretVolume.Exclude.Get()).To(Equal(false))

			Expect(payload.Spec.Worker.IsSet()).To(BeTrue())
			workerSpec2 := payload.Spec.Worker.Get()

			Expect(len(workerSpec2.EnvironmentVariables)).To(BeNumerically(">", 0))
			workerHomeEnvVar := workerSpec2.EnvironmentVariables[0]
			Expect(workerHomeEnvVar.Name.IsSet()).To(BeTrue())
			Expect(*workerHomeEnvVar.Name.Get()).To(Equal("HOME"))
			Expect(workerHomeEnvVar.Value.IsSet()).To(BeTrue())
			Expect(*workerHomeEnvVar.Value.Get()).To(Equal("/home/my-folder"))

			Expect(len(workerSpec2.Annotations)).To(BeNumerically(">", 0))
			workerBillingAnnotation := workerSpec2.Annotations[0]
			Expect(workerBillingAnnotation.Name.IsSet()).To(BeTrue())
			Expect(*workerBillingAnnotation.Name.Get()).To(Equal("billing"))
			Expect(workerBillingAnnotation.Value.IsSet()).To(BeTrue())
			Expect(*workerBillingAnnotation.Value.Get()).To(Equal("my-billing-unit"))

			Expect(len(workerSpec2.Labels)).To(BeNumerically(">", 0))
			workerStageLabel := workerSpec2.Labels[0]
			Expect(workerStageLabel.Name.IsSet()).To(BeTrue())
			Expect(*workerStageLabel.Name.Get()).To(Equal("stage"))
			Expect(workerStageLabel.Value.IsSet()).To(BeTrue())
			Expect(*workerStageLabel.Value.Get()).To(Equal("initial-research"))

			Expect(workerSpec2.Storage.IsSet()).To(BeTrue())
		})

		It("should apply policy defaults correctly for master-only workload", func() {
			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveDistributedInferencePolicyV2: &policyService.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableDistributedInferenceDefaultsV2(&policyService.DistributedInferenceDefaultsV2{
							NodePools:     []string{"default"},
							Category:      *policyService.NewNullableString(convutils.Ref("Train")),
							RestartPolicy: *policyService.NewNullableDistributedInferenceRestartPolicy(convutils.Ref(policyService.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART)),
							ServingPort: *policyService.NewNullableDistributedInferenceServingPort(&policyService.DistributedInferenceServingPort{
								Port:              *policyService.NewNullableInt32(convutils.Ref(int32(9090))),
								Protocol:          *policyService.NewNullableDistributedInferenceServingPortProtocol(convutils.Ref(policyService.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP)),
								AuthorizationType: *policyService.NewNullableDistributedInferenceServingPortAccessAuthorizationTypeEnum(convutils.Ref(policyService.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_AUTHORIZED_USERS)),
								ExposeExternally:  *policyService.NewNullableBool(convutils.Ref(false)),
							}),
							StartupPolicy: *policyService.NewNullableDistributedInferenceStartupPolicy(convutils.Ref(policyService.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_READY)),
							Workers:       *policyService.NewNullableInt32(convutils.Ref(int32(0))),
							Replicas:      *policyService.NewNullableInt32(convutils.Ref(int32(1))),
							Leader: *policyService.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyService.DistributedInferenceLeaderWorkerDefaultsV2{
								Command:         *policyService.NewNullableString(convutils.Ref("python")),
								Args:            *policyService.NewNullableString(convutils.Ref("-m flask run")),
								Image:           *policyService.NewNullableString(convutils.Ref("python:3.9-slim")),
								ImagePullPolicy: *policyService.NewNullableImagePullPolicy(convutils.Ref(policyService.IMAGEPULLPOLICY_IF_NOT_PRESENT)),
								WorkingDir:      *policyService.NewNullableString(convutils.Ref("/app")),
								CreateHomeDir:   *policyService.NewNullableBool(convutils.Ref(false)),
								NodeType:        *policyService.NewNullableString(convutils.Ref("inference-node")),
								Compute: *policyService.NewNullableSupersetDefaultsAllOfCompute(&policyService.SupersetDefaultsAllOfCompute{
									GpuDevicesRequest: *policyService.NewNullableInt32(convutils.Ref(int32(1))),
									CpuCoreRequest:    *policyService.NewNullableFloat64(convutils.Ref(2.0)),
									CpuCoreLimit:      *policyService.NewNullableFloat64(convutils.Ref(4.0)),
									CpuMemoryRequest:  *policyService.NewNullableString(convutils.Ref("1Gi")),
									CpuMemoryLimit:    *policyService.NewNullableString(convutils.Ref("2Gi")),
									LargeShmRequest:   *policyService.NewNullableBool(convutils.Ref(true)),
								}),
								Security: *policyService.NewNullableDistributedInferenceLeaderWorkerDefaultsV2Security(&policyService.DistributedInferenceLeaderWorkerDefaultsV2Security{
									UidGidSource:           *policyService.NewNullableUidGidSource(convutils.Ref(policyService.UIDGIDSOURCE_FROM_IDP_TOKEN)),
									RunAsNonRoot:           *policyService.NewNullableBool(convutils.Ref(true)),
									ReadOnlyRootFilesystem: *policyService.NewNullableBool(convutils.Ref(true)),
									RunAsUid:               *policyService.NewNullableInt64(convutils.Ref(int64(1000))),
									RunAsGid:               *policyService.NewNullableInt64(convutils.Ref(int64(1000))),
								}),
								EnvironmentVariables: *policyService.NewNullableEnvironmentVariablesDefaults(&policyService.EnvironmentVariablesDefaults{
									Instances: []policyService.EnvironmentVariable{
										{
											Name:        *policyService.NewNullableString(convutils.Ref("FLASK_ENV")),
											Value:       *policyService.NewNullableString(convutils.Ref("production")),
											Exclude:     *policyService.NewNullableBool(convutils.Ref(false)),
											Description: *policyService.NewNullableString(convutils.Ref("Flask environment")),
										},
									},
								}),
							}),
						}),
					},
				},
			})

			distributedInferenceService := NewDistributedInferenceService(mockTenantSettingsClient)
			workloadRequest := api.DistributedInferenceCreationRequest{
				Name:      "master-only-inference",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedInferenceSpecSpec{
					Workers:  convutils.Ref(int32(0)),
					Replicas: convutils.Ref(int32(1)),
					ServingPort: &api.DistributedInferenceServingPort{
						Port:              convutils.Ref(int32(8080)),
						AuthorizationType: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_PUBLIC),
					},
					Leader: &api.DistributedInferenceLeaderWorkerSpec{
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							CpuMemoryLimit: convutils.Ref("4Gi"),
						},
					},
				},
			}

			response, err := distributedInferenceService.CreateDistributedInference(ctx, workloadRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			payload := response.Body.(workloadsApi.SupersetWorkload)

			Expect(payload.Spec.NodePools).To(Equal([]string{"default"}))
			Expect(payload.Spec.Category.IsSet()).To(BeTrue())
			Expect(*payload.Spec.Category.Get()).To(Equal("Train"))
			Expect(payload.Spec.PriorityClass.IsSet()).To(BeFalse())

			Expect(payload.Spec.Workers.IsSet()).To(BeTrue())
			Expect(*payload.Spec.Workers.Get()).To(Equal(int32(0)))

			Expect(payload.Spec.ServingPort.IsSet()).To(BeTrue())
			servingPort := payload.Spec.ServingPort.Get()
			Expect(servingPort.Port.IsSet()).To(BeTrue())
			Expect(*servingPort.Port.Get()).To(Equal(int32(8080)))
			Expect(servingPort.AuthorizationType.IsSet()).To(BeTrue())
			Expect(*servingPort.AuthorizationType.Get()).To(Equal(string(workloadsApi.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_PUBLIC)))
			Expect(servingPort.Protocol.IsSet()).To(BeTrue())
			Expect(*servingPort.Protocol.Get()).To(Equal(string(workloadsApi.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP)))
			Expect(servingPort.ExposeExternally.IsSet()).To(BeTrue())
			Expect(*servingPort.ExposeExternally.Get()).To(Equal(false))

			Expect(payload.Spec.Leader.IsSet()).To(BeTrue())
			leaderSpec := payload.Spec.Leader.Get()
			Expect(leaderSpec.Command.IsSet()).To(BeTrue())
			Expect(*leaderSpec.Command.Get()).To(Equal("python"))
			Expect(leaderSpec.Args.IsSet()).To(BeTrue())
			Expect(*leaderSpec.Args.Get()).To(Equal("-m flask run"))
			Expect(leaderSpec.Image.IsSet()).To(BeTrue())
			Expect(*leaderSpec.Image.Get()).To(Equal("python:3.9-slim"))
			Expect(leaderSpec.WorkingDir.IsSet()).To(BeTrue())
			Expect(*leaderSpec.WorkingDir.Get()).To(Equal("/app"))
			Expect(leaderSpec.CreateHomeDir.IsSet()).To(BeTrue())
			Expect(*leaderSpec.CreateHomeDir.Get()).To(Equal(false))

			Expect(leaderSpec.Compute.IsSet()).To(BeTrue())
			leaderCompute := leaderSpec.Compute.Get()
			Expect(leaderCompute.CpuMemoryLimit.IsSet()).To(BeTrue())
			Expect(*leaderCompute.CpuMemoryLimit.Get()).To(Equal("4Gi"))
			Expect(leaderCompute.GpuDevicesRequest.IsSet()).To(BeTrue())
			Expect(*leaderCompute.GpuDevicesRequest.Get()).To(Equal(int32(1)))
			Expect(leaderCompute.CpuCoreRequest.IsSet()).To(BeTrue())
			Expect(*leaderCompute.CpuCoreRequest.Get()).To(Equal(2.0))
			Expect(leaderCompute.LargeShmRequest.IsSet()).To(BeTrue())
			Expect(*leaderCompute.LargeShmRequest.Get()).To(Equal(true))

			Expect(payload.Spec.Worker.IsSet()).To(BeFalse())

			Expect(leaderSpec.Security.IsSet()).To(BeTrue())
			leaderSecurity := leaderSpec.Security.Get()
			Expect(leaderSecurity.UidGidSource.IsSet()).To(BeTrue())
			Expect(string(*leaderSecurity.UidGidSource.Get())).To(Equal(string(workloadsApi.UIDGIDSOURCE_FROM_IDP_TOKEN)))
			Expect(leaderSecurity.RunAsNonRoot.IsSet()).To(BeTrue())
			Expect(*leaderSecurity.RunAsNonRoot.Get()).To(Equal(true))
			Expect(leaderSecurity.ReadOnlyRootFilesystem.IsSet()).To(BeTrue())
			Expect(*leaderSecurity.ReadOnlyRootFilesystem.Get()).To(Equal(true))

			Expect(len(leaderSpec.EnvironmentVariables)).To(BeNumerically(">", 0))
			flaskEnvVar := leaderSpec.EnvironmentVariables[0]
			Expect(flaskEnvVar.Name.IsSet()).To(BeTrue())
			Expect(*flaskEnvVar.Name.Get()).To(Equal("FLASK_ENV"))
			Expect(flaskEnvVar.Value.IsSet()).To(BeTrue())
			Expect(*flaskEnvVar.Value.Get()).To(Equal("production"))
			Expect(flaskEnvVar.Exclude.IsSet()).To(BeTrue())
			Expect(*flaskEnvVar.Exclude.Get()).To(Equal(false))
			Expect(flaskEnvVar.Description.IsSet()).To(BeTrue())
			Expect(*flaskEnvVar.Description.Get()).To(Equal("Flask environment"))
		})
	})

	Context("node affinity", func() {
		const (
			key      = "key"
			operator = api.MATCHEXPRESSIONOPERATOR_DOES_NOT_EXIST
		)

		var (
			nodeAffinityRequired *api.NodeAffinityRequired
		)

		BeforeEach(func() {
			nodeAffinityRequired = &api.NodeAffinityRequired{
				NodeSelectorTerms: []api.NodeSelectorTerm{
					{
						MatchExpressions: []api.MatchExpression{
							{
								Key:      key,
								Operator: operator,
							},
						},
					},
				},
			}
		})

		It("should validate cluster version", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.21.0", nil)))

			request := api.TrainingCreationRequest{
				Name:                 "job",
				UseGivenNameAsPrefix: true,
				ProjectId:            randoms.ProjectIdStr,
				ClusterId:            randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image:                convutils.Ref("ubuntu"),
					NodeAffinityRequired: nodeAffinityRequired,
				},
			}

			response, _ := trainingsService.CreateTraining(ctx, request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.(*workloadsApi.SubmissionErrorV2).Message).To(Equal("Node affinity is not supported in this cluster"))
		})

		Context("should validate match expressions", func() {
			DescribeTable("mandatory fields",
				func(expr api.MatchExpression, expectedErr string, expectedErrorCode int) {
					nodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions = append(nodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions, expr)

					request := api.DistributedCreationRequest{
						Name:                 "job",
						UseGivenNameAsPrefix: true,
						ProjectId:            randoms.ProjectIdStr,
						ClusterId:            randoms.ClusterId,
						Spec: api.DistributedSpecSpec{
							DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
							Image:                convutils.Ref("ubuntu"),
							NodeAffinityRequired: nodeAffinityRequired,
						},
					}

					response, _ := distributedService.CreateDistributed(ctx, request)
					Expect(response.Code).To(Equal(expectedErrorCode))
					Expect(response.Body.(*workloadsApi.SubmissionErrorV2).Message).To(Equal(expectedErr))
				},
				Entry("No key", api.MatchExpression{Operator: api.MATCHEXPRESSIONOPERATOR_EXISTS}, "nodeAffinityRequired: Key must be set", http.StatusBadRequest),
				Entry("No operator", api.MatchExpression{Operator: "", Key: "bla", Values: []string{"a"}}, "failed to load the workload spec:  is not a valid MatchExpressionOperator", http.StatusInternalServerError),
			)

			DescribeTable("operator : values",
				func(expr api.MatchExpression, expectedErr string) {
					nodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions = append(nodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions, expr)

					request := api.WorkspaceCreationRequest{
						Name:                 "job",
						UseGivenNameAsPrefix: true,
						ProjectId:            randoms.ProjectIdStr,
						ClusterId:            randoms.ClusterId,
						Spec: api.WorkspaceSpecSpec{
							Image:                convutils.Ref("ubuntu"),
							NodeAffinityRequired: nodeAffinityRequired,
						},
					}

					response, _ := workspaceService.CreateWorkspace(ctx, request)
					Expect(response.Code).To(Equal(http.StatusBadRequest))
					Expect(response.Body.(*workloadsApi.SubmissionErrorV2).Message).To(Equal(expectedErr))
				},
				Entry("In with empty Values", api.MatchExpression{Key: "bla", Operator: api.MATCHEXPRESSIONOPERATOR_IN}, "nodeAffinityRequired: Values must be non-empty when Operator is In"),
				Entry("NotIn with empty Values", api.MatchExpression{Key: "bla", Operator: api.MATCHEXPRESSIONOPERATOR_NOT_IN}, "nodeAffinityRequired: Values must be non-empty when Operator is NotIn"),
				Entry("Exists with non-empty Values", api.MatchExpression{Key: "bla", Operator: api.MATCHEXPRESSIONOPERATOR_EXISTS, Values: []string{"a"}}, "nodeAffinityRequired: Values must be empty when Operator is Exists"),
				Entry("DoesNotExist with non-empty Values", api.MatchExpression{Key: "bla", Operator: api.MATCHEXPRESSIONOPERATOR_DOES_NOT_EXIST, Values: []string{"a"}}, "nodeAffinityRequired: Values must be empty when Operator is DoesNotExist"),
				Entry("LT with empty Values", api.MatchExpression{Key: "bla", Operator: api.MATCHEXPRESSIONOPERATOR_LT}, "nodeAffinityRequired: Values must have a single element when Operator is Lt"),
				Entry("LT with multiple Values", api.MatchExpression{Key: "bla", Operator: api.MATCHEXPRESSIONOPERATOR_LT, Values: []string{"a", "b"}}, "nodeAffinityRequired: Values must have a single element when Operator is Lt"),
				Entry("GT with empty Values", api.MatchExpression{Key: "bla", Operator: api.MATCHEXPRESSIONOPERATOR_GT}, "nodeAffinityRequired: Values must have a single element when Operator is Gt"),
				Entry("GT with multiple Values", api.MatchExpression{Key: "bla", Operator: api.MATCHEXPRESSIONOPERATOR_GT, Values: []string{"a", "b"}}, "nodeAffinityRequired: Values must have a single element when Operator is Gt"),
			)
		})

		It("should succeed", func() {
			request := api.InferenceCreationRequest{
				Name:                 "job",
				UseGivenNameAsPrefix: true,
				ProjectId:            randoms.ProjectIdStr,
				ClusterId:            randoms.ClusterId,
				Spec: api.InferenceSpecSpec{
					Image: convutils.Ref("ubuntu"),
					ServingPort: &api.ServingPort{
						Container: convutils.Ref(int32(8080)),
						Protocol:  convutils.Ref(api.SERVINGPORTPROTOCOL_HTTP),
					},
					NodeAffinityRequired: nodeAffinityRequired,
				},
			}

			response, _ := inferenceService.CreateInference(ctx, request)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			workload := response.Body.(workloadsApi.SupersetWorkload)
			Expect(workload).NotTo(BeNil())

			affinity := workload.Spec.NodeAffinityRequired.Get()
			Expect(affinity).NotTo(BeNil())
			Expect(affinity.NodeSelectorTerms).To(HaveLen(1))
			Expect(affinity.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))
			Expect(affinity.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal(key))
			Expect(string(affinity.NodeSelectorTerms[0].MatchExpressions[0].Operator)).To(Equal(string(operator)))
		})
	})

	Context("When there are topologies", func() {
		var (
			pytorch         *workloadsApi.DistributedFramework
			workloadRequest api.DistributedCreationRequest
		)

		BeforeEach(func() {
			var err error
			pytorch, err = workloadsApi.NewDistributedFrameworkFromValue(string(api.DISTRIBUTEDFRAMEWORK_PY_TORCH))
			Expect(err).NotTo(HaveOccurred())

			workloadRequest = api.DistributedCreationRequest{
				Name:      "dist-job",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}
		})

		It("should return error when node pool is missing", func() {
			workloadRequest.Spec.NodePools = ptr.To([]string{"missing-nodepool"})

			response, _ := distributedService.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			payload := response.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(payload.Message).To(Equal("the following node pools do not exist: missing-nodepool"))
		})

		It("should return error when there are node pool with topology but topology is missing", func() {
			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithNodePoolsWithTopology(randoms.ClusterId, "no-topology", "missing-topology"))
			workloadRequest.Spec.NodePools = ptr.To([]string{"missing-topology"})

			response, _ := distributedService.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			payload := response.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(payload.Message).To(Equal("topology not found"))
		})

		It("should return workload with topology annotations", func() {
			topology := &csclient.NetworkTopology{
				Name:   "have-topology",
				Levels: []string{"level1", "level2"},
			}

			enabledTopologyAnnotation := []workloadsApi.Annotation{
				{
					Name:  *workloadsApi.NewNullableString(ptr.To("kai.scheduler/topology")),
					Value: *workloadsApi.NewNullableString(ptr.To("have-topology")),
				},
				{
					Name:  *workloadsApi.NewNullableString(ptr.To("kai.scheduler/topology-preferred-placement")),
					Value: *workloadsApi.NewNullableString(ptr.To("level2")),
				},
			}

			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
				WithNodePoolsWithTopology(randoms.ClusterId, "have-topology", "with-topology").
				WithTopology("have-topology", topology))
			workloadRequest.Spec.NodePools = ptr.To([]string{"with-topology"})
			workloadRequest.MasterSpec.NodePools = ptr.To([]string{"with-topology"})

			response, _ := distributedService.CreateDistributed(ctx, workloadRequest)
			Expect(response.Code).To(Equal(http.StatusAccepted))
			payload := response.Body.(workloadsApi.SupersetWorkload)

			// worker
			actualAsMap, _ := gojsonq.LoadJson(payload.Spec)
			expectAsMap, _ := gojsonq.LoadJson(workloadsApi.SupersetSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(pytorch),
				Image:                *workloadsApi.NewNullableString(ptr.To("ubuntu")),
				NodePools:            []string{"with-topology"},
				Category:             *workloadsApi.NewNullableString(ptr.To("Train")),
				Annotations:          enabledTopologyAnnotation,
			})
			Expect(actualAsMap).To(Equal(expectAsMap))

			// master
			actualAsMap, _ = gojsonq.LoadJson(payload.MasterSpec.Get())
			expectAsMap, _ = gojsonq.LoadJson(workloadsApi.SupersetSpec{
				Image:     *workloadsApi.NewNullableString(ptr.To("ubuntu")),
				Category:  *workloadsApi.NewNullableString(ptr.To("Train")),
				NodePools: []string{"with-topology"},
			})
			Expect(actualAsMap).To(Equal(expectAsMap))
		})
	})
})

var _ = Describe("setPreemptibilityFromPriority", func() {
	priorities := map[string]int32{"train": 50, "build": 100, "inference": 125}

	It("sets preemptible when priority value < 100", func() {
		spec := &workloadsApi.SupersetSpec{}
		spec.PriorityClass.Set(convutils.Ref("train"))
		setPreemptibilityFromPriority(spec, priorities)
		Expect(spec.Preemptibility.IsSet()).To(BeTrue())
		Expect(*spec.Preemptibility.Get()).To(Equal(workloadsApi.PREEMPTIBILITY_PREEMPTIBLE))
	})

	It("sets non-preemptible when priority value >= 100", func() {
		spec := &workloadsApi.SupersetSpec{}
		spec.PriorityClass.Set(convutils.Ref("build"))
		setPreemptibilityFromPriority(spec, priorities)
		Expect(spec.Preemptibility.IsSet()).To(BeTrue())
		Expect(*spec.Preemptibility.Get()).To(Equal(workloadsApi.PREEMPTIBILITY_NON_PREEMPTIBLE))
	})

	It("does not override when preemptibility already set", func() {
		spec := &workloadsApi.SupersetSpec{}
		spec.PriorityClass.Set(convutils.Ref("train"))
		spec.SetPreemptibility(workloadsApi.PREEMPTIBILITY_NON_PREEMPTIBLE)
		setPreemptibilityFromPriority(spec, priorities)
		Expect(*spec.Preemptibility.Get()).To(Equal(workloadsApi.PREEMPTIBILITY_NON_PREEMPTIBLE))
	})

	It("does nothing when priority name not found", func() {
		spec := &workloadsApi.SupersetSpec{}
		spec.PriorityClass.Set(convutils.Ref("unknown"))
		setPreemptibilityFromPriority(spec, priorities)
		Expect(spec.Preemptibility.IsSet()).To(BeFalse())
	})

	It("does nothing when priority class is not set", func() {
		spec := &workloadsApi.SupersetSpec{}
		setPreemptibilityFromPriority(spec, priorities)
		Expect(spec.Preemptibility.IsSet()).To(BeFalse())
	})
})

var _ = Describe("setMasterPreemptibilityFromPriority", func() {
	priorities := map[string]int32{"train": 50, "build": 100}

	It("sets preemptible when priority value < 100", func() {
		spec := &workloadsApi.MasterSpec{}
		spec.PriorityClass.Set(convutils.Ref("train"))
		setMasterPreemptibilityFromPriority(spec, priorities)
		Expect(spec.Preemptibility.IsSet()).To(BeTrue())
		Expect(*spec.Preemptibility.Get()).To(Equal(workloadsApi.PREEMPTIBILITY_PREEMPTIBLE))
	})

	It("sets non-preemptible when priority value >= 100", func() {
		spec := &workloadsApi.MasterSpec{}
		spec.PriorityClass.Set(convutils.Ref("build"))
		setMasterPreemptibilityFromPriority(spec, priorities)
		Expect(spec.Preemptibility.IsSet()).To(BeTrue())
		Expect(*spec.Preemptibility.Get()).To(Equal(workloadsApi.PREEMPTIBILITY_NON_PREEMPTIBLE))
	})
})
