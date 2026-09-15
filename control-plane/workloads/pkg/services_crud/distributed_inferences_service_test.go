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

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	workloadsMgrClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/workloads_mgr"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	workloadsMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("distributed inference service validations", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	var ctxAnotherTenant context.Context
	var ctxClusterWithOwnershipProtection context.Context

	const protectedClusterId = "protected-cluster-id"

	var randoms *testutils.RandomValues
	var randomsAnotherTenant *testutils.RandomValues
	var randomsClusterWithOwnershipProtection *testutils.RandomValues

	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}

	distributedInferencesService := NewDistributedInferenceService(mockTenantSettingsClient)
	workspacesService := NewWorkspacesService(mockTenantSettingsClient)

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

		backendMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms).With(randomsAnotherTenant).With(randomsClusterWithOwnershipProtection)
		backendMock.AllowedProjectsToContext(&ctx)
		backendMock.AllowedProjectsToContext(&ctxAnotherTenant)
		backendMock.AllowedProjectsToContext(&ctxClusterWithOwnershipProtection)

		orgunit.SetMock(backendMock.OrgUnitMock())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedClusterWithLWS(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, true)).
			WithCluster(randomsAnotherTenant.ClusterId, testing.BuildDisplayedClusterWithLWS(randomsAnotherTenant.ClusterId, testing.TestClusterVersionNativeSubmission, true)).
			WithCluster(randomsClusterWithOwnershipProtection.ClusterId, testing.BuildDisplayedClusterWithLWS(randomsClusterWithOwnershipProtection.ClusterId, testing.TestClusterVersionForUpdateInference, true)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))
	})

	AfterAll(func() {
		testing.CleanTables()
	})

	Context("create", func() {
		DescribeTable("resolves IdP identity independently for each distributed inference replica",
			func(leaderSource, workerSource api.UidGidSource) {
				ctx = context.WithValue(ctx, runaictx.CfUId, "2001")
				ctx = context.WithValue(ctx, runaictx.CfGId, "2002")
				ctx = context.WithValue(ctx, runaictx.CfSupplementalGroups, "2003,2004")
				security := func(source api.UidGidSource) *api.DistributedInferenceLeaderWorkerSpecSecurity {
					result := &api.DistributedInferenceLeaderWorkerSpecSecurity{
						UidGidSource: convutils.Ref(source),
						HostIpc:      convutils.Ref(true),
					}
					if source != api.UIDGIDSOURCE_FROM_IDP_TOKEN {
						result.RunAsUid = convutils.Ref(int64(1001))
						result.RunAsGid = convutils.Ref(int64(1002))
						result.SupplementalGroups = convutils.Ref("1003")
					}
					return result
				}
				request := api.DistributedInferenceCreationRequest{
					Name:      "inference-idp",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.DistributedInferenceSpecSpec{
						ServingPort: &api.DistributedInferenceServingPort{
							Port:     convutils.Ref(int32(8080)),
							Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
						},
						Workers:  convutils.Ref(int32(1)),
						Replicas: convutils.Ref(int32(1)),
						Leader:   &api.DistributedInferenceLeaderWorkerSpec{Image: convutils.Ref("leader-image"), Security: security(leaderSource)},
						Worker:   &api.DistributedInferenceLeaderWorkerSpec{Image: convutils.Ref("worker-image"), Security: security(workerSource)},
					},
				}
				response, err := distributedInferencesService.CreateDistributedInference(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusAccepted))
				payload := response.Body.(workloadsApi.SupersetWorkload)
				getResponse, err := distributedInferencesService.GetDistributedInference(ctx, payload.WorkloadId)
				Expect(err).NotTo(HaveOccurred())
				Expect(getResponse.Code).To(Equal(http.StatusOK))
				stored := getResponse.Body.(workloadsApi.SupersetWorkload)
				for i, replica := range []*workloadsApi.DistributedInferenceLeaderWorkerSpec{stored.Spec.Leader.Get(), stored.Spec.Worker.Get()} {
					source := []api.UidGidSource{leaderSource, workerSource}[i]
					uid, gid, groups := int64(1001), int64(1002), "1003"
					if source == api.UIDGIDSOURCE_FROM_IDP_TOKEN {
						uid, gid, groups = 2001, 2002, "2003,2004"
					}
					Expect(replica).NotTo(BeNil())
					result := replica.GetSecurity()
					Expect(result.GetRunAsUid()).To(Equal(uid))
					Expect(result.GetRunAsGid()).To(Equal(gid))
					Expect(result.GetSupplementalGroups()).To(Equal(groups))
					Expect(result.GetHostIpc()).To(BeTrue())
				}
			},
			Entry("both use IdP", api.UIDGIDSOURCE_FROM_IDP_TOKEN, api.UIDGIDSOURCE_FROM_IDP_TOKEN),
			Entry("only leader uses IdP", api.UIDGIDSOURCE_FROM_IDP_TOKEN, api.UIDGIDSOURCE_CUSTOM),
			Entry("only worker uses IdP", api.UIDGIDSOURCE_CUSTOM, api.UIDGIDSOURCE_FROM_IDP_TOKEN),
		)

		It("should create distributed inference with security fields (hostIpc, hostNetwork) on leader and worker", func() {
			request := api.DistributedInferenceCreationRequest{
				Name:      "dist-inf-security",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedInferenceSpecSpec{
					ServingPort: &api.DistributedInferenceServingPort{
						Port:     convutils.Ref(int32(8080)),
						Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
					},
					Workers:       convutils.Ref(int32(2)),
					Replicas:      convutils.Ref(int32(1)),
					StartupPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED),
					RestartPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART),
					Leader: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("leader-image:v1"),
						Security: &api.DistributedInferenceLeaderWorkerSpecSecurity{
							HostIpc:     convutils.Ref(true),
							HostNetwork: convutils.Ref(true),
							Capabilities: &[]api.Capability{
								api.CAPABILITY_IPC_LOCK,
							},
						},
					},
					Worker: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("worker-image:v1"),
						Security: &api.DistributedInferenceLeaderWorkerSpecSecurity{
							HostIpc:     convutils.Ref(true),
							HostNetwork: convutils.Ref(true),
						},
					},
				},
			}

			createResponse, _ := distributedInferencesService.CreateDistributedInference(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())

			getResponse, _ := distributedInferencesService.GetDistributedInference(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.WorkloadId).To(Equal(createPayload.WorkloadId))
		})

		It("get should be sensitive to workload type", func() {
			request := api.DistributedInferenceCreationRequest{
				Name:      "dist-inf-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedInferenceSpecSpec{
					ServingPort: &api.DistributedInferenceServingPort{
						Port:     convutils.Ref(int32(8080)),
						Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
					},
					Workers:       convutils.Ref(int32(2)),
					Replicas:      convutils.Ref(int32(1)),
					StartupPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED),
					RestartPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART),
					Leader: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("leader-image:v1"),
					},
					Worker: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("worker-image:v1"),
					},
				},
			}

			createResponse, _ := distributedInferencesService.CreateDistributedInference(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())

			//  expecting the actual phase to be Creating...
			createPayload.ActualPhase = convutils.Ref(workloadsApi.PHASE_CREATING)

			// get distributed inferences should return the correct workload
			getResponse, _ := distributedInferencesService.GetDistributedInference(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload).To(BeEquivalentTo(createPayload))

			// get workspace should return 404 not found
			getResponse, _ = workspacesService.GetWorkspace(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("should succeed with comprehensive distributed inference configuration", func() {
			request := api.DistributedInferenceCreationRequest{
				Name:      "comprehensive-dist-inf",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedInferenceSpecSpec{
					NodePools: &[]string{
						"gpu-pool-a",
						"gpu-pool-b",
						"cpu-pool",
					},
					ServingPort: &api.DistributedInferenceServingPort{
						Port:              convutils.Ref(int32(8081)),
						Protocol:          convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
						AuthorizationType: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_AUTHORIZED_USERS),
						AuthorizedUsers: &[]string{
							"user-a@example.com",
							"user-b@example.com",
						},
						ExposeExternally: convutils.Ref(true),
						ExposedUrl:       convutils.Ref("https://my-distributed-inference.example.com"),
					},
					RestartPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART),
					StartupPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_READY),
					Workers:       convutils.Ref(int32(3)),
					Replicas:      convutils.Ref(int32(2)),
					Category:      convutils.Ref("Train"),
					//PriorityClass: convutils.Ref("low"),
					Leader: &api.DistributedInferenceLeaderWorkerSpec{
						Command:         convutils.Ref("python"),
						Args:            convutils.Ref("-x my-script.py"),
						Image:           convutils.Ref("python:3.8"),
						NodeType:        convutils.Ref("my-node-type"),
						ImagePullPolicy: convutils.Ref(api.IMAGEPULLPOLICY_IF_NOT_PRESENT),
						CreateHomeDir:   convutils.Ref(true),
						WorkingDir:      convutils.Ref("/home/leader-folder"),
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuDevicesRequest: convutils.Ref(int32(2)),
							GpuRequestType:    convutils.Ref(api.GPUREQUESTTYPE_PORTION),
							GpuPortionRequest: convutils.Ref(float64(0.5)),
							CpuCoreRequest:    convutils.Ref(float64(1.0)),
							CpuCoreLimit:      convutils.Ref(float64(4)),
							CpuMemoryRequest:  convutils.Ref("2Gi"),
							CpuMemoryLimit:    convutils.Ref("8Gi"),
						},
						Storage: &api.DistributedInferenceLeaderWorkerSpecStorage{
							Pvc: &[]api.PvcInstance{
								{
									ExistingPvc: convutils.Ref(false),
									Path:        convutils.Ref("/data/leader-pvc"),
									ReadOnly:    convutils.Ref(false),
									ClaimInfo: &api.ClaimInfo{
										Size:         convutils.Ref("5Gi"),
										StorageClass: convutils.Ref("fast-ssd"),
									},
								},
							},
							ConfigMapVolume: &[]api.ConfigMapInstance{
								{
									ConfigMap: convutils.Ref("leader-config"),
									MountPath: convutils.Ref("/config/leader"),
									Name:      convutils.Ref("leader-app-config"),
								},
							},
							SecretVolume: &[]api.SecretInstance{
								{
									Secret:    convutils.Ref("leader-secret"),
									MountPath: convutils.Ref("/secrets/leader"),
									Name:      convutils.Ref("leader-credentials"),
								},
							},
							EmptyDirVolume: &[]api.EmptyDirInstance{
								{
									Path:      convutils.Ref("/tmp/leader-scratch"),
									Name:      convutils.Ref("leader-temp-storage"),
									SizeLimit: convutils.Ref("1Gi"),
								},
							},
						},
						Probes: &api.Probes{
							Readiness: &api.Probe{
								Handler: &api.ProbeHandler{
									HttpGet: api.ProbeHandlerHttpGet{
										Port: convutils.Ref(int32(8080)),
									},
								},
							},
						},
						EnvironmentVariables: &[]api.EnvironmentVariable{
							{
								Name:  convutils.Ref("HOME"),
								Value: convutils.Ref("/home/leader-folder"),
							},
							{
								Name:  convutils.Ref("LEADER_ID"),
								Value: convutils.Ref("leader-001"),
							},
							{
								Name:  convutils.Ref("DISTRIBUTED_ROLE"),
								Value: convutils.Ref("leader"),
							},
						},
						Tolerations: &[]api.Toleration{
							{
								Name:     convutils.Ref("gpu-taint"),
								Operator: convutils.Ref(api.TOLERATIONOPERATOR_EQUAL),
								Key:      convutils.Ref("nvidia.com/gpu"),
								Value:    convutils.Ref("true"),
								Effect:   convutils.Ref(api.TOLERATIONEFFECT_NO_SCHEDULE),
							},
							{
								Name:     convutils.Ref("leader-taint"),
								Operator: convutils.Ref(api.TOLERATIONOPERATOR_EXISTS),
								Key:      convutils.Ref("distributed-inference/leader"),
								Effect:   convutils.Ref(api.TOLERATIONEFFECT_NO_EXECUTE),
							},
						},
						Labels: &[]api.Label{
							{
								Name:  convutils.Ref("component"),
								Value: convutils.Ref("distributed-inference-leader"),
							},
							{
								Name:  convutils.Ref("version"),
								Value: convutils.Ref("v1.0.0"),
							},
						},
						Annotations: &[]api.Annotation{
							{
								Name:  convutils.Ref("deployment.kubernetes.io/revision"),
								Value: convutils.Ref("1"),
							},
							{
								Name:  convutils.Ref("distributed-inference/role"),
								Value: convutils.Ref("leader"),
							},
						},
						ImagePullSecrets: &[]api.ImagePullSecret{
							{
								Name: "registry-secret",
							},
							{
								Name:           "leader-user-credential",
								UserCredential: convutils.Ref(true),
							},
							{
								Name:           "shared-registry",
								UserCredential: convutils.Ref(false),
							},
						},
						PodAffinity: &api.PodAffinity{
							Type: convutils.Ref(api.PODAFFINITYTYPE_REQUIRED),
							Key:  convutils.Ref("distributed-inference-group"),
						},
					},
					Worker: &api.DistributedInferenceLeaderWorkerSpec{
						Command:         convutils.Ref("python"),
						Args:            convutils.Ref("-x worker-script.py"),
						Image:           convutils.Ref("python:3.8"),
						NodeType:        convutils.Ref("worker-node-type"),
						ImagePullPolicy: convutils.Ref(api.IMAGEPULLPOLICY_ALWAYS),
						CreateHomeDir:   convutils.Ref(true),
						WorkingDir:      convutils.Ref("/home/worker-folder"),
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuDevicesRequest: convutils.Ref(int32(1)),
							GpuRequestType:    convutils.Ref(api.GPUREQUESTTYPE_MEMORY),
							GpuMemoryRequest:  convutils.Ref("8Gi"),
							CpuCoreRequest:    convutils.Ref(float64(0.5)),
							CpuCoreLimit:      convutils.Ref(float64(2)),
							CpuMemoryRequest:  convutils.Ref("1Gi"),
							CpuMemoryLimit:    convutils.Ref("4Gi"),
						},
						Storage: &api.DistributedInferenceLeaderWorkerSpecStorage{
							Pvc: &[]api.PvcInstance{
								{
									ExistingPvc: convutils.Ref(false),
									Path:        convutils.Ref("/data/worker-pvc"),
									ReadOnly:    convutils.Ref(false),
									ClaimInfo: &api.ClaimInfo{
										Size:         convutils.Ref("2Gi"),
										StorageClass: convutils.Ref("standard-ssd"),
									},
								},
							},
							ConfigMapVolume: &[]api.ConfigMapInstance{
								{
									ConfigMap: convutils.Ref("worker-config"),
									MountPath: convutils.Ref("/config/worker"),
									Name:      convutils.Ref("worker-app-config"),
								},
							},
							SecretVolume: &[]api.SecretInstance{
								{
									Secret:    convutils.Ref("worker-secret"),
									MountPath: convutils.Ref("/secrets/worker"),
									Name:      convutils.Ref("worker-credentials"),
								},
							},
							EmptyDirVolume: &[]api.EmptyDirInstance{
								{
									Path:      convutils.Ref("/tmp/worker-scratch"),
									Name:      convutils.Ref("worker-temp-storage"),
									SizeLimit: convutils.Ref("512Mi"),
								},
							},
						},
						Security: &api.DistributedInferenceLeaderWorkerSpecSecurity{
							RunAsUid: convutils.Ref(int64(1001)),
							RunAsGid: convutils.Ref(int64(1001)),
							Capabilities: &[]api.Capability{
								api.CAPABILITY_NET_BIND_SERVICE,
								api.CAPABILITY_SYS_TIME,
							},
						},
						NodeAffinityRequired: &api.NodeAffinityRequired{
							NodeSelectorTerms: []api.NodeSelectorTerm{
								{
									MatchExpressions: []api.MatchExpression{
										{
											Key:      "worker-type",
											Operator: api.MATCHEXPRESSIONOPERATOR_IN,
											Values: []string{
												"gpu-worker",
												"compute-worker",
											},
										},
										{
											Key:      "zone",
											Operator: api.MATCHEXPRESSIONOPERATOR_IN,
											Values: []string{
												"zone-a",
												"zone-b",
											},
										},
									},
								},
							},
						},
						Probes: &api.Probes{
							Readiness: &api.Probe{
								Handler: &api.ProbeHandler{
									HttpGet: api.ProbeHandlerHttpGet{
										Port: convutils.Ref(int32(9090)),
									},
								},
							},
						},
						EnvironmentVariables: &[]api.EnvironmentVariable{
							{
								Name:  convutils.Ref("HOME"),
								Value: convutils.Ref("/home/worker-folder"),
							},
							{
								Name:  convutils.Ref("WORKER_ID"),
								Value: convutils.Ref("worker-${POD_NAME}"),
							},
							{
								Name:  convutils.Ref("DISTRIBUTED_ROLE"),
								Value: convutils.Ref("worker"),
							},
							{
								Name:  convutils.Ref("LEADER_ADDRESS"),
								Value: convutils.Ref("${LEADER_SERVICE_HOST}"),
							},
						},
						Tolerations: &[]api.Toleration{
							{
								Name:     convutils.Ref("gpu-worker-taint"),
								Operator: convutils.Ref(api.TOLERATIONOPERATOR_EQUAL),
								Key:      convutils.Ref("nvidia.com/gpu"),
								Value:    convutils.Ref("true"),
								Effect:   convutils.Ref(api.TOLERATIONEFFECT_NO_SCHEDULE),
							},
							{
								Name:     convutils.Ref("worker-taint"),
								Operator: convutils.Ref(api.TOLERATIONOPERATOR_EXISTS),
								Key:      convutils.Ref("distributed-inference/worker"),
								Effect:   convutils.Ref(api.TOLERATIONEFFECT_PREFER_NO_SCHEDULE),
							},
						},
						Labels: &[]api.Label{
							{
								Name:  convutils.Ref("component"),
								Value: convutils.Ref("distributed-inference-worker"),
							},
							{
								Name:  convutils.Ref("version"),
								Value: convutils.Ref("v1.0.0"),
							},
						},
						Annotations: &[]api.Annotation{
							{
								Name:  convutils.Ref("deployment.kubernetes.io/revision"),
								Value: convutils.Ref("1"),
							},
							{
								Name:  convutils.Ref("distributed-inference/role"),
								Value: convutils.Ref("worker"),
							},
						},
						ImagePullSecrets: &[]api.ImagePullSecret{
							{
								Name: "registry-secret",
							},
							{
								Name:           "worker-user-credential",
								UserCredential: convutils.Ref(true),
							},
							{
								Name:           "shared-registry",
								UserCredential: convutils.Ref(false),
							},
						},
						PodAffinity: &api.PodAffinity{
							Type: convutils.Ref(api.PODAFFINITYTYPE_PREFERRED),
							Key:  convutils.Ref("distributed-inference-worker-group"),
						},
					},
				},
			}

			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithNodePools(randoms.ClusterId, "gpu-pool-a", "gpu-pool-b", "cpu-pool"))

			createResponse, _ := distributedInferencesService.CreateDistributedInference(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())

			// get distributed inference should return the correct workload
			getResponse, _ := distributedInferencesService.GetDistributedInference(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.WorkloadId).To(Equal(createPayload.WorkloadId))
		})
	})

	Context("delete", func() {
		It("should be reflected in get", func() {
			request := api.DistributedInferenceCreationRequest{
				Name:      "dist-inf-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedInferenceSpecSpec{
					ServingPort: &api.DistributedInferenceServingPort{
						Port:     convutils.Ref(int32(8080)),
						Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
					},
					Workers:       convutils.Ref(int32(2)),
					Replicas:      convutils.Ref(int32(1)),
					StartupPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED),
					RestartPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART),
					Leader: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("leader-image:v1"),
					},
					Worker: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("worker-image:v1"),
					},
				},
			}

			createResponse, _ := distributedInferencesService.CreateDistributedInference(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			// desired state -> running, actual phase -> creating
			getResponse, _ := distributedInferencesService.GetDistributedInference(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_CREATING))

			// simulate real-life time b/w creating and deleting
			time.Sleep(time.Millisecond * 200)

			phaseResponse, _ := distributedInferencesService.DeleteDistributedInference(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusAccepted))

			getResponse, _ = distributedInferencesService.GetDistributedInference(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload = getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DeletedAt).To(Not(BeNil()))
		})
	})

	Context("update spec", func() {
		It("returns a populated response for an any-workload distributed inference update", func() {
			displayedCluster := testing.BuildDisplayedClusterWithLWS(randoms.ClusterId, "3.0.0", true)
			displayedCluster.SetDomain("https://mock-q7z3k.example.com")
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, displayedCluster))
			checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
				WithCluster(randoms.ClusterId, &displayedCluster).
				WithDefaultNodePool(randoms.ClusterId))

			workloadID := uuid.New()
			spec := workloadsApi.NewSupersetSpec()
			spec.SetWorkers(2)
			spec.SetReplicas(1)
			spec.SetStartupPolicy(workloadsApi.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED)
			spec.SetRestartPolicy(string(workloadsApi.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART))
			leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
			leader.SetImage("leader-image:v1")
			spec.SetLeader(*leader)
			worker := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
			worker.SetImage("worker-image:v1")
			spec.SetWorker(*worker)
			servingPort := workloadsApi.NewSupersetServingPort()
			servingPort.SetPort(8080)
			servingPort.SetContainer(8080)
			spec.SetServingPort(*servingPort)

			snapshot := workloadV2SnapshotForTest(workloadID, api.WORKLOADTYPE_DISTRIBUTED_INFERENCE, randoms)
			snapshot.Name = "distributed-inference-v2"
			snapshot.RequestedName = "distributed-inference-v2"
			snapshot.Namespace = "runai-project"
			snapshot.Spec = spec
			Expect(dao_crud.DbUpsertWorkloadV2Snapshot(ctx, snapshot)).To(BeNil())

			updatedWorkload, copyErr := workloadManagerUpdateResultForSnapshot(
				snapshot,
				constants.LwsGroup,
				constants.LwsVersion,
				constants.LwsKind,
				"lws-existing",
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

			updateResponse, _ := distributedInferencesService.UpdateDistributedInferenceSpec(ctx, workloadID.String(), api.UpdateRequest{
				Spec: api.UpdateSpecSpec{Replicas: convutils.Ref(int32(3))},
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
			Expect(response.Spec.Replicas.Get()).To(Equal(convutils.Ref(int32(3))))

			body, marshalErr := json.Marshal(response)
			Expect(marshalErr).NotTo(HaveOccurred())
			var generatedClientResponse workloadsApi.DistributedInference
			Expect(json.Unmarshal(body, &generatedClientResponse)).To(Succeed())
			Expect(generatedClientResponse.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
		})

		It("should update replicas field successfully", func() {
			// Create a distributed inference workload first
			creationRequest := api.DistributedInferenceCreationRequest{
				Name:      "dist-inf-update-test",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedInferenceSpecSpec{
					ServingPort: &api.DistributedInferenceServingPort{
						Port:     convutils.Ref(int32(8080)),
						Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
					},
					Workers:       convutils.Ref(int32(2)),
					Replicas:      convutils.Ref(int32(1)),
					StartupPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED),
					RestartPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART),
					Leader: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("leader-image:v1"),
					},
					Worker: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("worker-image:v1"),
					},
				},
			}

			createResponse, _ := distributedInferencesService.CreateDistributedInference(ctx, creationRequest)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())

			workloadId := createPayload.WorkloadId

			// Update the replicas field
			updateRequest := api.UpdateRequest{
				Spec: api.UpdateSpecSpec{
					Replicas: convutils.Ref(int32(3)),
				},
			}

			updateResponse, _ := distributedInferencesService.UpdateDistributedInferenceSpec(ctx, workloadId, updateRequest)
			Expect(updateResponse.Code).To(Equal(http.StatusAccepted))

			// Verify the update was applied by getting the workload
			getResponse, _ := distributedInferencesService.GetDistributedInference(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.Spec.Replicas.IsSet()).To(BeTrue())
			Expect(*getPayload.Spec.Replicas.Get()).To(Equal(int32(3)))
		})

		It("should handle empty update request gracefully", func() {
			// Create a distributed inference workload first
			creationRequest := api.DistributedInferenceCreationRequest{
				Name:      "dist-inf-empty-update-test",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedInferenceSpecSpec{
					ServingPort: &api.DistributedInferenceServingPort{
						Port:     convutils.Ref(int32(8080)),
						Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
					},
					Workers:       convutils.Ref(int32(2)),
					Replicas:      convutils.Ref(int32(1)),
					StartupPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED),
					RestartPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART),
					Leader: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("leader-image:v1"),
					},
					Worker: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("worker-image:v1"),
					},
				},
			}

			createResponse, _ := distributedInferencesService.CreateDistributedInference(ctx, creationRequest)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())

			workloadId := createPayload.WorkloadId

			// Update with empty spec (should be handled gracefully)
			updateRequest := api.UpdateRequest{
				Spec: api.UpdateSpecSpec{},
			}

			updateResponse, _ := distributedInferencesService.UpdateDistributedInferenceSpec(ctx, workloadId, updateRequest)
			Expect(updateResponse.Code).To(Equal(http.StatusAccepted))

			// Verify the workload still exists and hasn't changed
			getResponse, _ := distributedInferencesService.GetDistributedInference(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.Spec.Replicas.IsSet()).To(BeTrue())
			Expect(*getPayload.Spec.Replicas.Get()).To(Equal(int32(1))) // Should remain unchanged
		})

		It("should handle cluster capability validation", func() {
			// Set up cluster with insufficient version for distributed inference updates
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedClusterWithLWS(randoms.ClusterId, "2.22", true)).
				WithCluster(randomsAnotherTenant.ClusterId, testing.BuildDisplayedClusterWithLWS(randomsAnotherTenant.ClusterId, "2.22", true)))

			// Create a distributed inference workload
			creationRequest := api.DistributedInferenceCreationRequest{
				Name:      "dist-inf-capability-test",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedInferenceSpecSpec{
					ServingPort: &api.DistributedInferenceServingPort{
						Port:     convutils.Ref(int32(8080)),
						Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
					},
					Workers:       convutils.Ref(int32(2)),
					Replicas:      convutils.Ref(int32(1)),
					StartupPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED),
					RestartPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART),
					Leader: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("leader-image:v1"),
					},
					Worker: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("worker-image:v1"),
					},
				},
			}

			createResponse, _ := distributedInferencesService.CreateDistributedInference(ctx, creationRequest)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())

			workloadId := createPayload.WorkloadId

			// Try to update - should fail due to cluster capability
			updateRequest := api.UpdateRequest{
				Spec: api.UpdateSpecSpec{
					Replicas: convutils.Ref(int32(3)),
				},
			}

			updateResponse, _ := distributedInferencesService.UpdateDistributedInferenceSpec(ctx, workloadId, updateRequest)
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))

			// Reset cluster mock to supported version
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedClusterWithLWS(randoms.ClusterId, testing.TestDefaultClusterVersion, true)).
				WithCluster(randomsAnotherTenant.ClusterId, testing.BuildDisplayedClusterWithLWS(randomsAnotherTenant.ClusterId, testing.TestDefaultClusterVersion, true)))
		})

		It("should handle non-existent workload", func() {
			nonExistentWorkloadId := uuid.NewString()
			workloadsMgrClient.SetMock(&workloadsMgrMock{
				getErr: &workloadsMgrApi.Error{Code: http.StatusNotFound, Message: "workload not found"},
			})

			updateRequest := api.UpdateRequest{
				Spec: api.UpdateSpecSpec{
					Replicas: convutils.Ref(int32(3)),
				},
			}

			updateResponse, _ := distributedInferencesService.UpdateDistributedInferenceSpec(ctx, nonExistentWorkloadId, updateRequest)
			Expect(updateResponse.Code).To(Equal(http.StatusNotFound))
		})
	})
})
