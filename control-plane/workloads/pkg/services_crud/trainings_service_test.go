/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	assestMock "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"

	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("training service validations", func() {
	var ctx context.Context
	var differentCtx context.Context
	var randoms *testutils.RandomValues
	var differentRandoms *testutils.RandomValues

	const protectedClusterId = "protected-cluster-id"

	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}

	workspacesService := NewWorkspacesService(mockTenantSettingsClient)
	trainingsService := NewTrainingsService(mockTenantSettingsClient)
	workloadPropertiesService := services.NewWorkloadPropertiesService()

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})
		differentRandoms = testutils.NewRandomValues()
		differentCtx = testing.SetContext(differentRandoms.TenantIdStr, testing.AnotherUser, []string{})

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
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionForOwnershipProtection, convutils.Ref(true))).
			WithCluster(protectedClusterId, testing.BuildDisplayedCluster(protectedClusterId, testing.TestClusterVersionForOwnershipProtection, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))

		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	Context("create", func() {
		It("get should be sensitive to workload type", func() {
			request := api.TrainingCreationRequest{
				Name:      "ws-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())

			//  expecting the actual phase to be Creating...
			createPayload.ActualPhase = convutils.Ref(workloadsApi.PHASE_CREATING)

			// get trainings should return the correct workload
			getResponse, _ := trainingsService.GetTraining(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload).To(BeEquivalentTo(createPayload))

			// get workspace should return 404 not found
			getResponse, _ = workspacesService.GetWorkspace(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("priority class for training > 2.18 (applyPriorityClassLabelWorkaround)", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.21.9999", convutils.Ref(true))).
				WithCluster(protectedClusterId, testing.BuildDisplayedCluster(protectedClusterId, "2.21.9999", nil)))

			priorityClass := constants.PriorityClassBuild

			request := api.TrainingCreationRequest{
				Name:      "ws-11112",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image:         convutils.Ref("ubuntu"),
					PriorityClass: &priorityClass,
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())

			//  expecting the actual phase to be Creating...
			createPayload.ActualPhase = convutils.Ref(workloadsApi.PHASE_CREATING)

			// get trainings should return the correct workload
			getResponse, _ := trainingsService.GetTraining(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload).To(BeEquivalentTo(createPayload))

			Expect(getPayload.Spec.PriorityClass.Get()).To(BeNil())
			Expect(len(getPayload.Spec.Labels)).To(BeEquivalentTo(1))
			Expect(labelsToMap(getPayload.Spec.Labels)).To(BeEquivalentTo(map[string]string{"priorityClassName": "build"}))
		})

		It("Invalid priorityClass", func() {
			priorityClass := "not-existing-priority-class"

			request := api.TrainingCreationRequest{
				Name:      "ws-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image:         convutils.Ref("ubuntu"),
					PriorityClass: &priorityClass,
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			responseErr := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			expectedErrMessage := "priority class for Training is limited to: build, interactive-preemptible, train"
			Expect(responseErr.Message).To(Equal(expectedErrMessage))
		})

		It("non existent node pools", func() {
			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithNodePools(randoms.ClusterId, "np2", "np3"))

			request := api.TrainingCreationRequest{
				Name:      "ws-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					NodePools: &[]string{
						"np1",
						"np2",
						"np3",
						"np4",
					},
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			responseErr := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			expectedErrMessage := "the following node pools do not exist: np1, np4"
			Expect(responseErr.Message).To(Equal(expectedErrMessage))
		})

		It("priority class for training > 2.22 (applyPriorityClassLabelWorkaround not enabled)", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.22.0", convutils.Ref(true))))

			priorityClass := "medium-low"

			// the priority must exist from the cluster scope
			priority, daoErr := dao.GetPriorityByName(ctx, randoms.TenantIdStr, priorityClass)
			Expect(daoErr).To(BeNil())
			statusResponse, err := workloadPropertiesService.WorkloadPrioritiesStatus(ctx, randoms.ClusterId, api.PrioritiesForStatus{
				FullSync: false,
				Priorities: []api.PrioritiesForStatusPrioritiesInner{{
					Id:     priority.ID.String(),
					Name:   priorityClass,
					Value:  50,
					Source: priority.Source,
					Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
				}}})
			Expect(err).To(BeNil())
			Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

			request := api.TrainingCreationRequest{
				Name:      "ws-11113",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image:         convutils.Ref("ubuntu"),
					PriorityClass: &priorityClass,
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())

			//  expecting the actual phase to be Creating...
			createPayload.ActualPhase = convutils.Ref(workloadsApi.PHASE_CREATING)

			// get trainings should return the correct workload
			getResponse, _ := trainingsService.GetTraining(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload).To(BeEquivalentTo(createPayload))

			Expect(getPayload.Spec.PriorityClass.Get()).ToNot(BeNil())
			Expect(len(getPayload.Spec.Labels)).To(BeEquivalentTo(0))
		})

	})

	Context("delete", func() {
		It("should be reflected in get", func() {
			request := api.TrainingCreationRequest{
				Name:      "tr-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			// desired state -> running, actual phase -> creating
			getResponse, _ := trainingsService.GetTraining(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_CREATING))

			// simulate real-life time b/w creating and deleting
			time.Sleep(time.Millisecond * 200)

			phaseResponse, _ := trainingsService.DeleteTraining(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusAccepted))

			getResponse, _ = trainingsService.GetTraining(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload = getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DeletedAt).To(Not(BeNil()))
		})
		It("test ownership protection", func() {
			request := api.TrainingCreationRequest{
				Name:      "tr-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			// desired state -> running, actual phase -> creating
			getResponse, _ := trainingsService.GetTraining(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_CREATING))

			differentCtx = testing.SetProjectContext(differentRandoms.TenantIdStr, testing.AnotherUser, []string{}, differentRandoms.ProjectIdStr)
			phaseResponse, _ := trainingsService.DeleteTraining(differentCtx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("suspend", func() {
		It("should be reflected in get", func() {
			request := api.TrainingCreationRequest{
				Name:      "tr-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			// desired state -> running, actual phase -> creating
			getResponse, _ := trainingsService.GetTraining(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_CREATING))

			// attempting to suspend will fail, because we need the status to be running
			phaseResponse, _ := trainingsService.SuspendTraining(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusBadRequest))
			errPayload := phaseResponse.Body.(*api.Error)
			Expect(errPayload.Message).To(ContainSubstring("suspend request is not permitted for Creating workloads"))

			// forcefully change the phase to: running
			nowPlus1Min := time.Now().Add(1 * time.Minute)
			dbErr := dao.DbUpdateWorkloadPhase(ctx, randoms.TenantId, workloadId, api.WorkloadPatchFields{Phase: api.PHASE_RUNNING}, nowPlus1Min)
			Expect(dbErr).To(BeNil())

			// simulate real-life time b/w creating and suspending
			time.Sleep(time.Millisecond * 200)

			phaseResponse, _ = trainingsService.SuspendTraining(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusAccepted))

			getResponse, _ = trainingsService.GetTraining(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload = getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(BeEquivalentTo(workloadsApi.WORKLOADDESIREDPHASE_STOPPED))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(BeEquivalentTo(workloadsApi.PHASE_STOPPING))
		})
		It("test ownership protection", func() {
			request := api.TrainingCreationRequest{
				Name:      "tr-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			// desired state -> running, actual phase -> creating
			getResponse, _ := trainingsService.GetTraining(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_CREATING))

			anotherCtx := testing.SetProjectContext(randoms.TenantIdStr, testing.AnotherUser, []string{}, randoms.ProjectIdStr)
			phaseResponse, _ := trainingsService.SuspendTraining(anotherCtx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("test environment variables", func() {
		It("test user credentials protection - should success", func() {
			// to ensure the cluster supports user credentials
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, nil)))
			// training creation request with 3 environment variables using 2 user credentials
			request := api.TrainingCreationRequest{
				Name:      "tr-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					EnvironmentVariables: convutils.Ref([]api.EnvironmentVariable{
						{
							Name: convutils.Ref("my-env-var"),
							UserCredential: convutils.Ref(api.EnvironmentVariableUserCredential{
								Name: convutils.Ref("my-credential"),
								Key:  convutils.Ref("my-key"),
							}),
						},
						{
							Name: convutils.Ref("my-other-env-var"),
							UserCredential: convutils.Ref(api.EnvironmentVariableUserCredential{
								Name: convutils.Ref("my-other-credential"),
								Key:  convutils.Ref("my-other-key"),
							}),
						},
						{
							Name: convutils.Ref("my-other-other-env-var"),
							UserCredential: convutils.Ref(api.EnvironmentVariableUserCredential{
								Name: convutils.Ref("my-credential"),
								Key:  convutils.Ref("my-other-other-key"),
							}),
						},
					}),
				},
			}

			// mock the assets client to return 3 user credentials (2 are needed for the env vars)
			assetsClientMock := &assestMock.IAssetsClient{}
			assets.SetMock(assetsClientMock)
			assetsClientMock.On("ListCredentialsAssets", ctx, true).
				Return(&assetsApi.CredentialsListResponse{
					Entries: []assetsApi.CredentialsListResponseEntry{
						{
							Meta: assetsApi.AssetMeta{
								Name:    "my-credential",
								Id:      "my-credential-id",
								Kind:    assetsApi.ASSETKIND_GENERIC_SECRET,
								Private: *assetsApi.NewNullableBool(convutils.Ref(true)),
							},
						},
						{
							Meta: assetsApi.AssetMeta{
								Name:    "my-other-credential",
								Id:      "my-other-credential-id",
								Kind:    assetsApi.ASSETKIND_GENERIC_SECRET,
								Private: *assetsApi.NewNullableBool(convutils.Ref(true)),
							},
						},
						{
							Meta: assetsApi.AssetMeta{
								Name:    "my-3th-credential",
								Id:      "my-3th-credential-id",
								Kind:    assetsApi.ASSETKIND_GENERIC_SECRET,
								Private: *assetsApi.NewNullableBool(convutils.Ref(true)),
							},
						},
					},
				}, nil)

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			// As part of the test, we expect the assets mapping to be created when syncing the workload
			listHttpRes, listErr := ListUpdates(ctx, Training, 0, randoms.ClusterId, &config.WorkloadsSyncConfig{BulkSize: 40}, nil)
			Expect(listErr).To(BeNil())
			resp := httpResponseBodyToListUpdatesResponse(listHttpRes)

			Expect(len(resp.Workloads)).To(Equal(1))
			Expect(resp.Workloads[0].Workload.Name).To(Equal("tr-1"))
			Expect(len(*resp.Workloads[0].Workload.AssetsMapping)).To(Equal(2))

			type assetsInfo struct {
				Kind api.AssetKind
				Id   string
			}

			actualAssetsMapping := make(map[string]assetsInfo)

			for _, mapping := range *resp.Workloads[0].Workload.AssetsMapping {
				actualAssetsMapping[mapping.Name] = assetsInfo{Kind: mapping.Kind, Id: mapping.Id}
			}
			Expect(actualAssetsMapping["my-credential"]).To(Equal(assetsInfo{Kind: api.ASSETKIND_GENERIC_SECRET, Id: "my-credential-id"}))
			Expect(actualAssetsMapping["my-other-credential"]).To(Equal(assetsInfo{Kind: api.ASSETKIND_GENERIC_SECRET, Id: "my-other-credential-id"}))
		})

		It("test user credentials protection with a made up credential- should return error", func() {
			// to ensure the cluster supports user credentials
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "9.9.9", nil)))

			// training creation request with 3 environment variables using 2 user credentials
			request := api.TrainingCreationRequest{
				Name:      "tr-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					EnvironmentVariables: convutils.Ref([]api.EnvironmentVariable{
						{
							Name: convutils.Ref("my-env-var"),
							UserCredential: convutils.Ref(api.EnvironmentVariableUserCredential{
								Name: convutils.Ref("my-made-up-credential"),
								Key:  convutils.Ref("my-key"),
							}),
						},
						{
							Name: convutils.Ref("my-other-env-var"),
							UserCredential: convutils.Ref(api.EnvironmentVariableUserCredential{
								Name: convutils.Ref("my-other-credential"),
								Key:  convutils.Ref("my-other-key"),
							}),
						},
						{
							Name: convutils.Ref("my-other-other-env-var"),
							UserCredential: convutils.Ref(api.EnvironmentVariableUserCredential{
								Name: convutils.Ref("my-credential"),
								Key:  convutils.Ref("my-other-other-key"),
							}),
						},
					}),
				},
			}

			// mock the assets client to return 3 user credentials but without the made up credential
			assetsClientMock := &assestMock.IAssetsClient{}
			assets.SetMock(assetsClientMock)
			assetsClientMock.On("ListCredentialsAssets", ctx, true).
				Return(&assetsApi.CredentialsListResponse{
					Entries: []assetsApi.CredentialsListResponseEntry{
						{
							Meta: assetsApi.AssetMeta{
								Name:    "my-credential",
								Id:      "my-credential-id",
								Kind:    assetsApi.ASSETKIND_GENERIC_SECRET,
								Private: *assetsApi.NewNullableBool(convutils.Ref(true)),
							},
						},
						{
							Meta: assetsApi.AssetMeta{
								Name:    "my-other-credential",
								Id:      "my-other-credential-id",
								Kind:    assetsApi.ASSETKIND_GENERIC_SECRET,
								Private: *assetsApi.NewNullableBool(convutils.Ref(true)),
							},
						},
						{
							Meta: assetsApi.AssetMeta{
								Name:    "my-3th-credential",
								Id:      "my-3th-credential-id",
								Kind:    assetsApi.ASSETKIND_GENERIC_SECRET,
								Private: *assetsApi.NewNullableBool(convutils.Ref(true)),
							},
						},
					},
				}, nil)

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusNotFound))
		})
		It("test user credentials protection with unsupported cluster - should fail", func() {
			// user credentials are supported only for clusters with version >= 2.23
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.20.0", nil)))
			// training creation request with 3 environment variables using 2 user credentials
			request := api.TrainingCreationRequest{
				Name:      "tr-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					EnvironmentVariables: convutils.Ref([]api.EnvironmentVariable{
						{
							Name: convutils.Ref("my-env-var"),
							UserCredential: convutils.Ref(api.EnvironmentVariableUserCredential{
								Name: convutils.Ref("my-credential"),
								Key:  convutils.Ref("my-key"),
							}),
						},
						{
							Name: convutils.Ref("my-other-env-var"),
							UserCredential: convutils.Ref(api.EnvironmentVariableUserCredential{
								Name: convutils.Ref("my-other-credential"),
								Key:  convutils.Ref("my-other-key"),
							}),
						},
						{
							Name: convutils.Ref("my-other-other-env-var"),
							UserCredential: convutils.Ref(api.EnvironmentVariableUserCredential{
								Name: convutils.Ref("my-credential"),
								Key:  convutils.Ref("my-other-other-key"),
							}),
						},
					}),
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("resume", func() {
		It("should be reflected in get", func() {
			request := api.TrainingCreationRequest{
				Name:      "tr-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			// desired state -> running, actual phase -> creating
			getResponse, _ := trainingsService.GetTraining(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_CREATING))

			// attempting to suspend will fail, because we need the status to be sopped
			phaseResponse, _ := trainingsService.ResumeTraining(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusBadRequest))
			errPayload := phaseResponse.Body.(*api.Error)
			Expect(errPayload.Message).To(ContainSubstring("resume request is permitted only for stopped workloads (currently the workload is in Creating phase)"))

			// forcefully change the phase to: running
			nowPlus1Min := time.Now().Add(1 * time.Minute)
			dbErr := dao.DbUpdateWorkloadPhase(ctx, randoms.TenantId, workloadId, api.WorkloadPatchFields{Phase: api.PHASE_STOPPED}, nowPlus1Min)
			Expect(dbErr).To(BeNil())

			// simulate real-life time b/w creating and suspending
			time.Sleep(time.Millisecond * 200)

			phaseResponse, _ = trainingsService.ResumeTraining(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusAccepted))

			getResponse, _ = trainingsService.GetTraining(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload = getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(BeEquivalentTo(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(BeEquivalentTo(workloadsApi.PHASE_RESUMING))
		})
	})

	Context("delete or get from forbidden project", func() {
		It("should fail", func() {
			request := api.TrainingCreationRequest{
				Name:      "tr-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			createResponse, _ := trainingsService.CreateTraining(ctx, request)
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
			getResponse, _ := trainingsService.GetTraining(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))

			// delete should fail with 403
			deleteResponse, _ := trainingsService.DeleteTraining(ctx, workloadId)
			Expect(deleteResponse.Code).To(Equal(http.StatusForbidden))
		})
	})
})

func labelsToMap(labels []workloadsApi.Label) map[string]string {
	result := make(map[string]string)
	for _, label := range labels {
		result[*label.Name.Get()] = *label.Value.Get()
	}
	return result
}
