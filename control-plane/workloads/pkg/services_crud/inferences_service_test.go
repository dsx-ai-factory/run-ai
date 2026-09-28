/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	workloadsMgrClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/workloads_mgr"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	workloadsMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	image1 = "image:v1"
	image2 = "image:v2"
)

var _ = Describe("inference service validations", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	var ctxAnotherTenant context.Context
	var ctxClusterWithOwnershipProtection context.Context

	const protectedClusterId = "protected-cluster-id"

	var randoms *testutils.RandomValues
	var randomsAnotherTenant *testutils.RandomValues
	var randomsClusterWithOwnershipProtection *testutils.RandomValues

	workloadsService := services.NewWorkloadsService()
	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}

	podsService := services.NewPodsService(mockTenantSettingsClient)
	inferencesService := NewInferencesService(mockTenantSettingsClient)
	workspacesService := NewWorkspacesService(mockTenantSettingsClient)

	creationRequest := api.InferenceCreationRequest{}
	updateRequest := api.InferenceUpdateRequest{}

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		randomsAnotherTenant = randoms.CloneWithDifferentTenant().Forbidden()
		randomsClusterWithOwnershipProtection = randomsAnotherTenant.CloneWithDifferentCluster().Allowed()

		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})
		ctxAnotherTenant = testing.SetContext(randomsAnotherTenant.TenantIdStr, testing.TestUser, []string{})
		ctxClusterWithOwnershipProtection = testing.SetContext(randomsClusterWithOwnershipProtection.TenantIdStr, testing.TestUser, []string{})

		policyMockUpdateTime := time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms).With(randomsAnotherTenant).With(randomsClusterWithOwnershipProtection)
		authMock.AllowedProjectsToContext(&ctx)
		authMock.AllowedProjectsToContext(&ctxAnotherTenant)
		authMock.AllowedProjectsToContext(&ctxClusterWithOwnershipProtection)

		orgunit.SetMock(authMock.OrgUnitMock())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, nil)).
			WithCluster(randomsAnotherTenant.ClusterId, testing.BuildDisplayedCluster(randomsAnotherTenant.ClusterId, testing.TestClusterVersionNativeSubmission, nil)).
			WithCluster(randomsClusterWithOwnershipProtection.ClusterId, testing.BuildDisplayedCluster(randomsClusterWithOwnershipProtection.ClusterId, testing.TestClusterVersionForUpdateInference, convutils.Ref(true))))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))

		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))

		// Default the workloads-manager to "not found" so that update requests for workloads that
		// do not exist in the native DB resolve to a proper 404 (the update path now consults
		// workloads-manager before concluding a workload is missing). Tests that need a specific
		// workloads-manager behavior override this mock locally.
		workloadsMgrClient.SetMock(&workloadsMgrMock{
			getErr: &workloadsMgrApi.Error{Code: http.StatusNotFound, Message: "workload not found"},
		})

		creationRequest = api.InferenceCreationRequest{
			Name:      "inf-1",
			ProjectId: randoms.ProjectIdStr,
			ClusterId: randoms.ClusterId,
			Spec: api.InferenceSpecSpec{
				Image: convutils.Ref(image1),
				ServingPort: &api.ServingPort{
					Container: convutils.Ref(int32(8080)),
					Protocol:  convutils.Ref(api.SERVINGPORTPROTOCOL_HTTP),
				},
			},
		}

		updateRequest = api.InferenceUpdateRequest{
			Spec: api.InferenceUpdateSpecSpec{
				Image: convutils.Ref(image2),
			},
		}
	})

	AfterAll(func() {
		testing.CleanTables()
	})

	Context("create", func() {
		It("should create inference with security fields (hostIpc, hostNetwork, capabilities)", func() {
			request := api.InferenceCreationRequest{
				Name:      "inf-security",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.InferenceSpecSpec{
					Image: convutils.Ref("ubuntu"),
					ServingPort: &api.ServingPort{
						Container: convutils.Ref(int32(8080)),
					},
					Security: &api.InferenceSpecSpecAllOfSecurity{
						HostIpc:     convutils.Ref(true),
						HostNetwork: convutils.Ref(true),
						Capabilities: &[]api.Capability{
							api.CAPABILITY_IPC_LOCK,
							api.CAPABILITY_NET_BIND_SERVICE,
						},
					},
				},
			}

			createResponse, _ := inferencesService.CreateInference(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())

			getResponse, _ := inferencesService.GetInference(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.Spec.Security.IsSet()).To(BeTrue())
			security := getPayload.Spec.Security.Get()
			Expect(security.HostIpc.IsSet()).To(BeTrue())
			Expect(*security.HostIpc.Get()).To(BeTrue())
			Expect(security.HostNetwork.IsSet()).To(BeTrue())
			Expect(*security.HostNetwork.Get()).To(BeTrue())
		})

		It("get should be sensitive to workload type", func() {
			request := api.InferenceCreationRequest{
				Name:      "inf-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.InferenceSpecSpec{
					Image: convutils.Ref("ubuntu"),
					ServingPort: &api.ServingPort{
						Container: convutils.Ref(int32(8080)),
					},
				},
			}

			createResponse, _ := inferencesService.CreateInference(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())

			//  expecting the actual phase to be Creating...
			createPayload.ActualPhase = convutils.Ref(workloadsApi.PHASE_CREATING)

			// get trainings should return the correct workload
			getResponse, _ := inferencesService.GetInference(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload).To(BeEquivalentTo(createPayload))

			// get workspace should return 404 not found
			getResponse, _ = workspacesService.GetWorkspace(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("delete", func() {
		It("should be reflected in get", func() {
			request := api.InferenceCreationRequest{
				Name:      "inf-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.InferenceSpecSpec{
					Image: convutils.Ref("ubuntu"),
					ServingPort: &api.ServingPort{
						Container: convutils.Ref(int32(8080)),
						Protocol:  convutils.Ref(api.SERVINGPORTPROTOCOL_GRPC),
					},
				},
			}

			createResponse, _ := inferencesService.CreateInference(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			// desired state -> running, actual phase -> creating
			getResponse, _ := inferencesService.GetInference(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_CREATING))

			// simulate real-life time b/w creating and deleting
			time.Sleep(time.Millisecond * 200)

			phaseResponse, _ := inferencesService.DeleteInference(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusAccepted))

			getResponse, _ = inferencesService.GetInference(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload = getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DeletedAt).To(Not(BeNil()))
		})
	})

	Context("update spec", func() {
		It("returns a populated response for an any-workload inference update", func() {
			displayedCluster := testing.BuildDisplayedCluster(randoms.ClusterId, "3.0.0", convutils.Ref(true))
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, displayedCluster))
			checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
				WithCluster(randoms.ClusterId, &displayedCluster).
				WithDefaultNodePool(randoms.ClusterId))

			workloadID := uuid.New()
			spec := workloadsApi.NewSupersetSpec()
			spec.SetImage(image1)
			servingPort := workloadsApi.NewSupersetServingPort()
			servingPort.SetContainer(8080)
			spec.SetServingPort(*servingPort)
			autoscaling := workloadsApi.NewAutoScaling()
			autoscaling.SetMinReplicas(0)
			autoscaling.SetMaxReplicas(1)
			spec.SetAutoscaling(*autoscaling)

			snapshot := workloadV2SnapshotForTest(workloadID, api.WORKLOADTYPE_INFERENCE, randoms)
			snapshot.Name = "inference-v2"
			snapshot.RequestedName = "inference-v2"
			snapshot.Namespace = "runai-project"
			snapshot.Spec = spec
			Expect(dao_crud.DbUpsertWorkloadV2Snapshot(ctx, snapshot)).To(BeNil())

			updatedWorkload, copyErr := workloadManagerUpdateResultForSnapshot(
				snapshot,
				knativeGroup,
				knativeVersion,
				knativeKind,
				"knative-existing",
			)
			Expect(copyErr).NotTo(HaveOccurred())
			Expect(updatedWorkload.Metadata.Id).To(Equal(snapshot.ID.String()))
			updatedWorkload.Metadata.Name = "workloads-manager-name"
			updatedWorkload.Metadata.CreatedBy = "workloads-manager-user"
			mgr := &workloadsMgrMock{
				getResponse:    workloadsMgrApi.NewWorkloadV2GetResponse(*updatedWorkload),
				updateResponse: updatedWorkload,
			}
			originalWorkloadsManager := workloadsMgrClient.Get()
			workloadsMgrClient.SetMock(mgr)
			DeferCleanup(func() { workloadsMgrClient.SetMock(originalWorkloadsManager) })

			updateResponse, _ := inferencesService.UpdateInferenceSpec(ctx, workloadID.String(), api.InferenceUpdateRequest{
				Spec: api.InferenceUpdateSpecSpec{
					Autoscaling: &api.AutoScaling{MinReplicas: convutils.Ref(int32(1))},
				},
			})

			Expect(updateResponse.Code).To(Equal(http.StatusAccepted), "response body: %#v", updateResponse.Body)
			response := updateResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(response.WorkloadId).To(Equal(workloadID.String()))
			Expect(response.Name).To(Equal(snapshot.Name))
			Expect(response.ProjectId).To(Equal(snapshot.ProjectId))
			Expect(response.ClusterId).To(Equal(snapshot.ClusterId))
			Expect(response.CreatedBy).To(Equal(snapshot.CreatedBy))
			Expect(response.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(response.Spec).NotTo(BeNil())
			Expect(response.Spec.Autoscaling.Get().MinReplicas.Get()).To(Equal(convutils.Ref(int32(1))))

			body, marshalErr := json.Marshal(response)
			Expect(marshalErr).NotTo(HaveOccurred())
			var generatedClientResponse workloadsApi.Inference
			Expect(json.Unmarshal(body, &generatedClientResponse)).To(Succeed())
			Expect(generatedClientResponse.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
		})

		It("should be reflected in get", func() {
			createResponse, _ := inferencesService.CreateInference(ctx, creationRequest)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())
			Expect(createPayload.Spec).NotTo(BeNil())
			Expect(createPayload.Spec.Image).NotTo(BeNil())
			Expect(createPayload.Spec.Image.IsSet()).To(BeTrue())
			Expect(*createPayload.Spec.Image.Get()).To(Equal(image1))

			workloadId := createPayload.WorkloadId

			getResponse, _ := inferencesService.GetInference(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.Spec.Image.IsSet()).To(BeTrue())
			Expect(*getPayload.Spec.Image.Get()).To(Equal(image1))

			updateResponse, _ := inferencesService.UpdateInferenceSpec(ctx, workloadId, updateRequest)
			Expect(updateResponse.Code).To(Equal(http.StatusAccepted))
			updatePayload := updateResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(updatePayload.Spec).NotTo(BeNil())
			Expect(updatePayload.Spec.Image).NotTo(BeNil())
			Expect(updatePayload.Spec.Image.IsSet()).To(BeTrue())
			Expect(*updatePayload.Spec.Image.Get()).To(Equal(image2))

			getResponse, _ = inferencesService.GetInference(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getPayload = getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.Spec.Image.IsSet()).To(BeTrue())
			Expect(*getPayload.Spec.Image.Get()).To(Equal(image2))
		})
	})

	Context("update spec - tests for add/remove/replace spec updates", func() {
		It("result spec should be as expected", func() {
			// replacing image value
			// leaving command as-is
			// adding args
			// replacing environment variables
			// removing probes
			// leaving autoscaling as-is
			fullSpecJson := `{
				"servingPort": {
					"container": 8080,
					"protocol": "http"
				},
				"image": "image:v1",
				"command": "cmd1",
				"environmentVariables": [
					{"name": "key1", "value": "val1"},
					{"name": "key2", "value": "val2"}
				],
				"probes": {
					"readiness": {
						"initialDelaySeconds": 10,
						"timeoutSeconds": 5
					}
				},
				"autoscaling": {
					"minReplicas": 1,
					"maxReplicas": 3
				}
			}`

			updateSpecJson := `{
				"image": "image:v2",
				"args": "--arg1 --arg2",
				"environmentVariables": [
					{"name": "key1", "value": "val3"}
				],
				"probes": {}
			}`

			fullSpec := workloadsApi.SupersetSpec{}
			updateSpec := api.InferenceUpdateSpecSpec{}

			err := json.Unmarshal([]byte(fullSpecJson), &fullSpec)
			Expect(err).To(BeNil())
			err = json.Unmarshal([]byte(updateSpecJson), &updateSpec)
			Expect(err).To(BeNil())

			copyUpdateSpecIntoFullSpec(ctx, &fullSpec, &updateSpec)

			Expect(fullSpec.Image.IsSet()).To(BeTrue())
			Expect(*fullSpec.Image.Get()).To(Equal("image:v2"))
			Expect(fullSpec.Command.IsSet()).To(BeTrue())
			Expect(*fullSpec.Command.Get()).To(Equal("cmd1"))
			Expect(fullSpec.Args.IsSet()).To(BeTrue())
			Expect(*fullSpec.Args.Get()).To(Equal("--arg1 --arg2"))
			Expect(fullSpec.EnvironmentVariables).ToNot(BeNil())
			Expect(len(fullSpec.EnvironmentVariables)).To(Equal(1))
			Expect(*fullSpec.EnvironmentVariables[0].Value.Get()).To(Equal("val3"))
			Expect(fullSpec.Probes.IsSet()).To(BeTrue())
			Expect(fullSpec.Probes.Get().Readiness.IsSet()).To(BeFalse())
			Expect(fullSpec.Autoscaling.IsSet()).To(BeTrue())
			Expect(fullSpec.Autoscaling.Get().MinReplicas.IsSet()).To(BeTrue())
			Expect(*fullSpec.Autoscaling.Get().MinReplicas.Get()).To(Equal(int32(1)))
			Expect(fullSpec.Autoscaling.Get().MaxReplicas.IsSet()).To(BeTrue())
			Expect(*fullSpec.Autoscaling.Get().MaxReplicas.Get()).To(Equal(int32(3)))
		})

		Context("update spec - more tests for add/remove/replace spec updates", func() {
			It("result spec should be as expected", func() {
				// leaving image as-is
				// replacing command to ""
				// removing environment variables
				// replacing autoscaling values
				fullSpecJson := `{
				"servingPort": {
					"container": 8080,
					"protocol": "http"
				},
				"image": "image:v1",
				"command": "cmd1",
				"environmentVariables": [
					{"name": "key1", "value": "val1"}
				],
				"autoscaling": {
					"minReplicas": 1,
					"maxReplicas": 3
				}
			}`

				updateSpecJson := `{
				"command": "",
				"environmentVariables": [],
				"autoscaling": {
					"maxReplicas": 5
				}
			}`

				fullSpec := workloadsApi.SupersetSpec{}
				updateSpec := api.InferenceUpdateSpecSpec{}

				err := json.Unmarshal([]byte(fullSpecJson), &fullSpec)
				Expect(err).To(BeNil())
				err = json.Unmarshal([]byte(updateSpecJson), &updateSpec)
				Expect(err).To(BeNil())

				copyUpdateSpecIntoFullSpec(ctx, &fullSpec, &updateSpec)

				Expect(fullSpec.Image.IsSet()).To(BeTrue())
				Expect(*fullSpec.Image.Get()).To(Equal("image:v1"))
				Expect(fullSpec.Command.IsSet()).To(BeTrue())
				Expect(*fullSpec.Command.Get()).To(Equal(""))
				Expect(fullSpec.EnvironmentVariables).ToNot(BeNil())
				Expect(len(fullSpec.EnvironmentVariables)).To(Equal(0))
				Expect(fullSpec.Autoscaling.IsSet()).To(BeTrue())
				Expect(fullSpec.Autoscaling.Get().MinReplicas.IsSet()).To(BeFalse())
				Expect(fullSpec.Autoscaling.Get().MaxReplicas.IsSet()).To(BeTrue())
				Expect(*fullSpec.Autoscaling.Get().MaxReplicas.Get()).To(Equal(int32(5)))
			})
		})

		Context("attempt to update spec with workload id that doesn't exists", func() {
			It("should fail", func() {
				randomWorkloadId := uuid.New().String()

				updateResponse, _ := inferencesService.UpdateInferenceSpec(ctx, randomWorkloadId, updateRequest)
				Expect(updateResponse.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("update spec by another user, who is authorized for the scope", func() {
			It("should succeed", func() {
				createResponse, _ := inferencesService.CreateInference(ctx, creationRequest)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))

				createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
				Expect(createPayload.WorkloadId).NotTo(BeNil())

				anotherCtx := testing.SetContext(randoms.TenantIdStr, testing.AnotherUser, []string{})
				updateResponse, _ := inferencesService.UpdateInferenceSpec(anotherCtx, createPayload.WorkloadId, updateRequest)
				Expect(updateResponse.Code).To(Equal(http.StatusAccepted))
			})
		})

		Context("attempt to update spec by another user, who is not authorized for the scope", func() {
			It("should fail", func() {
				createResponse, _ := inferencesService.CreateInference(ctx, creationRequest)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))

				createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
				Expect(createPayload.WorkloadId).NotTo(BeNil())

				anotherCtx := testing.SetContext(randomsAnotherTenant.TenantIdStr, testing.AnotherUser, []string{})
				updateResponse, _ := inferencesService.UpdateInferenceSpec(anotherCtx, createPayload.WorkloadId, updateRequest)
				Expect(updateResponse.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("attempt to update spec while ownership protection is set", func() {
			It("owner should succeed, another user should fail", func() {
				creationRequest.ProjectId = randomsClusterWithOwnershipProtection.ProjectIdStr
				creationRequest.ClusterId = randomsClusterWithOwnershipProtection.ClusterId

				org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randomsClusterWithOwnershipProtection.ProjectIdStr, GetProjectResponse(*randomsClusterWithOwnershipProtection)))
				clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randomsClusterWithOwnershipProtection.ClusterId))

				createResponse, _ := inferencesService.CreateInference(ctxClusterWithOwnershipProtection, creationRequest)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))

				createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
				Expect(createPayload.WorkloadId).NotTo(BeNil())

				//owner should succeed
				updateResponse, _ := inferencesService.UpdateInferenceSpec(ctxClusterWithOwnershipProtection, createPayload.WorkloadId, updateRequest)
				Expect(updateResponse.Code).To(Equal(http.StatusAccepted))

				// but another user should fail
				ctxAnotherUser := testing.SetProjectContext(randomsClusterWithOwnershipProtection.TenantIdStr, testing.AnotherUser, []string{}, randomsClusterWithOwnershipProtection.ProjectIdStr)
				updateResponse, _ = inferencesService.UpdateInferenceSpec(ctxAnotherUser, createPayload.WorkloadId, updateRequest)
				Expect(updateResponse.Code).To(Equal(http.StatusForbidden))
			})
		})

		Context("attempt to update spec with a policy compliance error", func() {
			It("should fail with the compliance issues in the response", func() {
				createResponse, _ := inferencesService.CreateInference(ctx, creationRequest)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))

				createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
				workloadId := createPayload.WorkloadId

				// trying to update with an empty environment variable,
				// while Name is a mandatory field
				updateRequest = api.InferenceUpdateRequest{
					Spec: api.InferenceUpdateSpecSpec{
						EnvironmentVariables: &[]api.EnvironmentVariable{
							{},
						},
					},
				}

				updateResponse, _ := inferencesService.UpdateInferenceSpec(ctx, workloadId, updateRequest)
				Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))
				updatePayload := updateResponse.Body.(*workloadsApi.SubmissionErrorV2)
				Expect(updatePayload.Message).NotTo(BeNil())
				Expect(updatePayload.Message).To(ContainSubstring("does not comply with the policy"))
				Expect(updatePayload.ComplianceIssues).NotTo(BeNil())
			})
		})

		Context("create inference - with serving configuration", func() {
			It("min version valid with deprecated InitializationTimeoutSeconds and new InitializationTimeoutSeconds (both) ", func() {
				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.22", convutils.Ref(true))).
					WithCluster(protectedClusterId, testing.BuildDisplayedCluster(protectedClusterId, "2.22", nil)))

				request := api.InferenceCreationRequest{
					Name:      "inf-1",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.InferenceSpecSpec{
						Image: convutils.Ref("ubuntu"),
						ServingPort: &api.ServingPort{
							Container: convutils.Ref(int32(8080)),
						},
						Autoscaling: &api.AutoScaling{
							InitializationTimeoutSeconds: convutils.Ref(int32(60)),
							MinReplicas:                  ptr.To(int32(1)),
						},
						ServingConfiguration: &api.ServingConfiguration{
							InitializationTimeoutSeconds: convutils.Ref(int32(61)),
						},
					},
				}

				createResponse, _ := inferencesService.CreateInference(ctx, request)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))

				createResponseBody := createResponse.Body.(workloadsApi.SupersetWorkload)
				Expect(createResponseBody.WorkloadId).NotTo(BeNil())

				Expect(*createResponseBody.Spec.ServingConfiguration.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(61)))
				Expect(*createResponseBody.Spec.Autoscaling.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(61))) // both should be set to the ServingConfiguration InitializationTimeoutSeconds value
				Expect(*createResponseBody.Spec.Autoscaling.Get().MinReplicas.Get()).To(Equal(int32(1)))                   // make sure the MinReplicas is set as well and not overridden by the ServingConfiguration
			})

			It("min version valid with new InitializationTimeoutSeconds and RequestTimeoutSeconds ", func() {
				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.22", convutils.Ref(true))).
					WithCluster(protectedClusterId, testing.BuildDisplayedCluster(protectedClusterId, "2.22", nil)))

				request := api.InferenceCreationRequest{
					Name:      "inf-1",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.InferenceSpecSpec{
						Image: convutils.Ref("ubuntu"),
						ServingPort: &api.ServingPort{
							Container: convutils.Ref(int32(8080)),
						},
						Autoscaling: &api.AutoScaling{
							MinReplicas: ptr.To(int32(1)),
						},
						ServingConfiguration: &api.ServingConfiguration{
							InitializationTimeoutSeconds: convutils.Ref(int32(62)),
							RequestTimeoutSeconds:        convutils.Ref(int32(63)),
						},
					},
				}

				createResponse, _ := inferencesService.CreateInference(ctx, request)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))

				createResponseBody := createResponse.Body.(workloadsApi.SupersetWorkload)
				Expect(createResponseBody.WorkloadId).NotTo(BeNil())

				Expect(*createResponseBody.Spec.ServingConfiguration.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(62)))
				Expect(*createResponseBody.Spec.ServingConfiguration.Get().RequestTimeoutSeconds.Get()).To(Equal(int32(63)))
				Expect(*createResponseBody.Spec.Autoscaling.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(62))) // both should be set to the ServingConfiguration InitializationTimeoutSeconds value
				Expect(*createResponseBody.Spec.Autoscaling.Get().MinReplicas.Get()).To(Equal(int32(1)))                   // make sure the MinReplicas is set as well and not overridden by the ServingConfiguration
			})

			It("min version valid with AutoScaling deprecated InitializationTimeoutSeconds ", func() {
				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.22", convutils.Ref(true))).
					WithCluster(protectedClusterId, testing.BuildDisplayedCluster(protectedClusterId, "2.22", nil)))

				request := api.InferenceCreationRequest{
					Name:      "inf-1",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.InferenceSpecSpec{
						Image: convutils.Ref("ubuntu"),
						ServingPort: &api.ServingPort{
							Container: convutils.Ref(int32(8080)),
						},
						Autoscaling: &api.AutoScaling{
							InitializationTimeoutSeconds: convutils.Ref(int32(60)),
						},
					},
				}

				createResponse, _ := inferencesService.CreateInference(ctx, request)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))

				createResponseBody := createResponse.Body.(workloadsApi.SupersetWorkload)
				Expect(createResponseBody.WorkloadId).NotTo(BeNil())

				Expect(*createResponseBody.Spec.ServingConfiguration.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(60)))
				Expect(*createResponseBody.Spec.Autoscaling.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(60)))
			})

			It("min version not valid with ServingConfiguration InitializationTimeoutSeconds ", func() {
				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.21", convutils.Ref(true))).
					WithCluster(protectedClusterId, testing.BuildDisplayedCluster(protectedClusterId, "2.21", nil)))

				request := api.InferenceCreationRequest{
					Name:      "inf-1",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.InferenceSpecSpec{
						Image: convutils.Ref("ubuntu"),
						ServingPort: &api.ServingPort{
							Container: convutils.Ref(int32(8080)),
						},
						ServingConfiguration: &api.ServingConfiguration{
							InitializationTimeoutSeconds: convutils.Ref(int32(62)),
							RequestTimeoutSeconds:        convutils.Ref(int32(63)),
						},
					},
				}

				createResponse, _ := inferencesService.CreateInference(ctx, request)
				Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
				createPayload := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
				Expect(createPayload.Message).NotTo(BeNil())
				Expect(createPayload.Message).To(ContainSubstring("RequestTimeoutSeconds is not supported in this cluster (minimal supported cluster version: 2.22)"))
			})

		})

		//Context("update inference - with serving configuration", func() {
		//	It("min version valid with AutoScaling deprecated InitializationTimeoutSeconds and update", func() {
		//		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
		//			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.22", convutils.Ref(true))).
		//			WithCluster(protectedClusterId, testing.BuildDisplayedCluster(protectedClusterId, "2.22", nil)))
		//
		//		request := api.InferenceCreationRequest{
		//			Name:      "inf-1",
		//			ProjectId: randoms.ProjectIdStr,
		//			ClusterId: randoms.ClusterId,
		//			Spec: api.InferenceSpecSpec{
		//				Image: convutils.Ref("ubuntu"),
		//				ServingPort: &api.ServingPort{
		//					Container: convutils.Ref(int32(8080)),
		//				},
		//				Autoscaling: &api.AutoScaling{
		//					InitializationTimeoutSeconds: convutils.Ref(int32(60)),
		//				},
		//			},
		//		}
		//
		//		createResponse, _ := inferencesService.CreateInference(ctx, request)
		//		Expect(createResponse.Code).To(Equal(http.StatusAccepted))
		//
		//		createResponseBody := createResponse.Body.(workloadsApi.SupersetWorkload)
		//		Expect(createResponseBody.WorkloadId).NotTo(BeNil())
		//
		//		Expect(*createResponseBody.Spec.ServingConfiguration.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(60)))
		//		Expect(*createResponseBody.Spec.Autoscaling.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(60)))
		//
		//		updateRequest := api.InferenceUpdateRequest{
		//			Spec: api.InferenceUpdateSpecSpec{
		//				Autoscaling: &api.AutoScaling{
		//					InitializationTimeoutSeconds: convutils.Ref(int32(61)),
		//				},
		//				ServingConfiguration: &api.ServingConfiguration{
		//					InitializationTimeoutSeconds: convutils.Ref(int32(62)),
		//				},
		//			},
		//		}
		//		updateResponse, _ := inferencesService.UpdateInferenceSpec(ctx, createResponseBody.WorkloadId, updateRequest)
		//		Expect(updateResponse.Code).To(Equal(http.StatusAccepted))
		//
		//		updateResponseBody := updateResponse.Body.(workloadsApi.SupersetWorkload)
		//		Expect(createResponseBody.WorkloadId).NotTo(BeNil())
		//
		//		Expect(*updateResponseBody.Spec.ServingConfiguration.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(62)))
		//		Expect(*updateResponseBody.Spec.Autoscaling.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(62)))
		//	})
		//	// TODO add more test for update inference with serving configuration
		//})

	})

	Context("Get workload and pod Metrics", func() {
		It("Request inference workload and pod metrics", func() {
			requestedWorkloadMetrics := []api.InferenceWorkloadMetricType{
				api.INFERENCEWORKLOADMETRICTYPE_THROUGHPUT,
				api.INFERENCEWORKLOADMETRICTYPE_LATENCY,
			}

			requestedPodMetrics := []api.InferencePodMetricType{
				api.INFERENCEPODMETRICTYPE_THROUGHPUT,
				api.INFERENCEPODMETRICTYPE_LATENCY,
			}

			baseWorkload := services.SetupBaseWorkload(testing.TestUuid1, services.DefaultProjectId, randoms.ClusterId)
			baseWorkload.Type = auth.WorkloadTypeInference
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: services.DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{baseWorkload},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// workload metrics
			resp, err := inferencesService.GetInferenceWorkloadMetrics(ctx, baseWorkload.Id, requestedWorkloadMetrics,
				time.Now(), time.Now().Add(-4*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(207))

			// pod metrics
			apiResponse, _ = podsService.InternalSyncPods(ctx, api.InternalPodsSyncRequest{
				ClusterId: testing.ClusterUuid3,
				Pods:      []api.PodSyncFields{services.SetupBasePod(testing.TestUuid4, baseWorkload.Id, services.DefaultProjectId, baseWorkload.ClusterId)},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			resp, err = inferencesService.GetInferenceWorkloadPodMetrics(ctx, baseWorkload.Id, testing.TestUuid4, requestedPodMetrics,
				time.Now(), time.Now().Add(-4*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(207))
		})
	})

	Context("resource ID validation", func() {
		It("should validate and fix resource IDs correctly", func() {
			// Test valid 2-part resource ID
			result := validateAndFixResourceId("nvidia/llama-3.1-8b-instruct", "nvidia")
			Expect(result).To(Equal("nvidia/llama-3.1-8b-instruct"))

			// Test 3-part resource ID with matching orgName
			result = validateAndFixResourceId("nvidia/nvidia/llama-3.1-8b-instruct", "nvidia")
			Expect(result).To(Equal("nvidia/llama-3.1-8b-instruct"))

			// Test 3-part resource ID with non-matching orgName
			result = validateAndFixResourceId("nvidia/meta/llama-3.1-8b-instruct", "meta")
			Expect(result).To(Equal("nvidia/meta/llama-3.1-8b-instruct"))

			// Test empty resource ID
			result = validateAndFixResourceId("", "nvidia")
			Expect(result).To(Equal(""))

			// Test single part resource ID
			result = validateAndFixResourceId("llama-3.1-8b-instruct", "nvidia")
			Expect(result).To(Equal("llama-3.1-8b-instruct"))

			// Test 4-part resource ID
			result = validateAndFixResourceId("nvidia/nvidia/meta/llama-3.1-8b-instruct", "nvidia")
			Expect(result).To(Equal("nvidia/nvidia/meta/llama-3.1-8b-instruct"))
		})
	})
})
