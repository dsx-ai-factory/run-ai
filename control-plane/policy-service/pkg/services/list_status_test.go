/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"fmt"

	orgUnitService "github.com/run-ai/runai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/clients/cluster"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtesting"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/tree"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
)

const (
	urlTrainingSyncUpdate    = "/api/v2/policy/trainings/sync/status"
	urlWorkspaceSyncUpdate   = "/api/v2/policy/workspaces/sync/status"
	urlDistributedSyncUpdate = "/api/v2/policy/distributed/sync/status"
	urlInferenceSyncUpdate   = "/api/v2/policy/inferences/sync/status"
)

var (
	tenantWorkspacePolicy   = api.PolicyCreationFields{Scope: api.SCOPETYPE_TENANT, Name: "pol-w"}
	tenantTrainingPolicy    = api.PolicyCreationFields{Scope: api.SCOPETYPE_TENANT, Name: "pol-t"}
	tenantDistributedPolicy = api.PolicyCreationFields{Scope: api.SCOPETYPE_TENANT, Name: "pol-d"}
	tenantInferencePolicy   = api.PolicyCreationFields{Scope: api.SCOPETYPE_TENANT, Name: "pol-i"}
)

var (
	departWorkspacePolicy  = api.PolicyCreationFields{Scope: api.SCOPETYPE_DEPARTMENT, DepartmentId: convutils.Ref("111"), Name: "pol-w-d11"}
	projectWorkspacePolicy = api.PolicyCreationFields{Scope: api.SCOPETYPE_PROJECT, ProjectId: convutils.Ref(int32(1111)), Name: "pol-w-p1"}
)

var _ = Describe("ListStatusTest", func() {
	router := CreateRouter()
	tenantId := int32(1)

	var backendMock *BackendMockClient
	var clusters []string

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()

		dbtesting.RestartUpdateCounter()

		tree.Restart()

		backendMock = NewBackendMockClient(1, 2)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))
		authorization.SetMock(NewPoliciesAuthClient(backendMock))

		clusterMock := cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		cluster.SetMock(clusterMock)

		clusters = backendMock.ClusterIdPerTenant["1"]
	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Behavior
	/////////////////////////////////////////////////////////////////////////////////////////////////////////

	Context("status in list policies", func() {
		It("correct statuses are reported", func() {
			workspace := createWorkspacePolicyForStatusTest(router, tenantWorkspacePolicy)
			training := createTrainingPolicyForStatusTest(router, tenantTrainingPolicy)
			distributed := createDistributedPolicyForStatusTest(router, tenantDistributedPolicy)
			inference := createInferencePolicyForStatusTest(router, tenantInferencePolicy)
			distributedInference := createDistributedInferencePolicyForStatusTest(router, tenantInferencePolicy)

			idsToNick := map[string]string{
				clusters[0]:                  "cluster-0",
				clusters[1]:                  "cluster-1",
				workspace.Meta.Id:            "workspace-policy",
				training.Meta.Id:             "training-policy",
				distributed.Meta.Id:          "distributed-policy",
				inference.Meta.Id:            "inference-policy",
				distributedInference.Meta.Id: "distributed-inference-policy",
			}

			//
			//   expecting the status of policies to be ready
			//
			validateListSyncStatus(router, tenantId, "policy created, we expect Not Ready for both clusters", idsToNick, map[PolicyAndClusterId]api.PolicySyncStatusOfCluster{
				NewPolicyAndClusterId(workspace.Meta.Id, clusters[0], idsToNick):            NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(workspace.Meta.Id, clusters[1], idsToNick):            NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
				NewPolicyAndClusterId(training.Meta.Id, clusters[0], idsToNick):             NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(training.Meta.Id, clusters[1], idsToNick):             NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
				NewPolicyAndClusterId(distributed.Meta.Id, clusters[0], idsToNick):          NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(distributed.Meta.Id, clusters[1], idsToNick):          NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
				NewPolicyAndClusterId(inference.Meta.Id, clusters[0], idsToNick):            NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(inference.Meta.Id, clusters[1], idsToNick):            NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
				NewPolicyAndClusterId(distributedInference.Meta.Id, clusters[0], idsToNick): NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(distributedInference.Meta.Id, clusters[1], idsToNick): NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
			})
		})
	})

	Context("status while doing delete", func() {
		It("correct statuses are reported", func() {
			workspace := createWorkspacePolicyForStatusTest(router, tenantWorkspacePolicy)
			training := createTrainingPolicyForStatusTest(router, tenantTrainingPolicy)

			idsToNick := map[string]string{
				clusters[0]:       "cluster-0",
				clusters[1]:       "cluster-1",
				workspace.Meta.Id: "workspace-policy",
				training.Meta.Id:  "training-policy",
			}

			//
			//  and reports back about the status update
			//
			validateListSyncStatus(router, tenantId, "cluster 0 received the workspace update", idsToNick, map[PolicyAndClusterId]api.PolicySyncStatusOfCluster{
				NewPolicyAndClusterId(workspace.Meta.Id, clusters[0], idsToNick): NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(workspace.Meta.Id, clusters[1], idsToNick): NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
				NewPolicyAndClusterId(training.Meta.Id, clusters[0], idsToNick):  NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(training.Meta.Id, clusters[1], idsToNick):  NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
			})

			//
			//   now the policy is being deleted
			//
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace))

			validateListSyncStatus(router, tenantId, "only training remains", idsToNick, map[PolicyAndClusterId]api.PolicySyncStatusOfCluster{
				NewPolicyAndClusterId(training.Meta.Id, clusters[0], idsToNick): NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(training.Meta.Id, clusters[1], idsToNick): NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
			})
		})
	})

	Context("deletion followed by re-creation", func() {
		It("cluster sync + list status is correct", func() {
			workspace := createWorkspacePolicyForStatusTest(router, tenantWorkspacePolicy)
			training := createTrainingPolicyForStatusTest(router, tenantTrainingPolicy)

			idsToNick := map[string]string{
				clusters[0]:       "cluster-0",
				clusters[1]:       "cluster-1",
				workspace.Meta.Id: "workspace-policy",
				training.Meta.Id:  "training-policy",
			}

			validateListSyncStatus(router, tenantId, "cluster 0 received the workspace update", idsToNick, map[PolicyAndClusterId]api.PolicySyncStatusOfCluster{
				NewPolicyAndClusterId(workspace.Meta.Id, clusters[0], idsToNick): NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(workspace.Meta.Id, clusters[1], idsToNick): NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
				NewPolicyAndClusterId(training.Meta.Id, clusters[0], idsToNick):  NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(training.Meta.Id, clusters[1], idsToNick):  NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
			})

			//
			//    the workspace policy is being Deleted
			//
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace))

			validateListSyncStatus(router, tenantId, "clusters received the deletion request", idsToNick, map[PolicyAndClusterId]api.PolicySyncStatusOfCluster{
				NewPolicyAndClusterId(training.Meta.Id, clusters[0], idsToNick): NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(training.Meta.Id, clusters[1], idsToNick): NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
			})

			//
			//   the workspace policy is re-created
			//
			workspace = createWorkspacePolicyForStatusTest(router, tenantWorkspacePolicy)

			idsToNick[workspace.Meta.Id] = "new-workspace-policy"

			//
			//   the status goes back to not ready for both clusters
			//
			validateListSyncStatus(router, tenantId, "clusters received the deletion request", idsToNick, map[PolicyAndClusterId]api.PolicySyncStatusOfCluster{
				NewPolicyAndClusterId(workspace.Meta.Id, clusters[0], idsToNick): NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(workspace.Meta.Id, clusters[1], idsToNick): NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
				NewPolicyAndClusterId(training.Meta.Id, clusters[0], idsToNick):  NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(training.Meta.Id, clusters[1], idsToNick):  NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
			})
		})
	})

	Context("deletion of parent policy", func() {
		It("list status is correct", func() {
			//
			//   create tenant + department policies
			//
			tenantPolicy := createWorkspacePolicyForStatusTest(router, tenantWorkspacePolicy)
			departPolicy := createWorkspacePolicyForStatusTest(router, departWorkspacePolicy)
			projectPolicy := createWorkspacePolicyForStatusTest(router, projectWorkspacePolicy)

			idsToNick := map[string]string{
				clusters[0]:           "cluster-0",
				clusters[1]:           "cluster-1",
				tenantPolicy.Meta.Id:  "tenant-policy",
				departPolicy.Meta.Id:  "depart-policy",
				projectPolicy.Meta.Id: "project-policy",
			}

			//
			//   make them synced with the cluster. we expect updates 1 and 2, one per policy
			//
			validateListSyncStatus(router, tenantId, "cluster 0 received the updates", idsToNick, map[PolicyAndClusterId]api.PolicySyncStatusOfCluster{
				NewPolicyAndClusterId(tenantPolicy.Meta.Id, clusters[0], idsToNick):  NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(departPolicy.Meta.Id, clusters[0], idsToNick):  NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(projectPolicy.Meta.Id, clusters[0], idsToNick): NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(tenantPolicy.Meta.Id, clusters[1], idsToNick):  NewPolicySyncStatusOfCluster(clusters[1], PolicyStatusReady, nil),
			})

			//
			//    delete the tenant policy.
			//
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace))

			//
			//    tenant policy gone.
			//
			validateListSyncStatus(router, tenantId, "clusters received the deletion request", idsToNick, map[PolicyAndClusterId]api.PolicySyncStatusOfCluster{
				NewPolicyAndClusterId(departPolicy.Meta.Id, clusters[0], idsToNick):  NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(projectPolicy.Meta.Id, clusters[0], idsToNick): NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
			})
		})
	})

	Context("project/dept, deletion + re-creation, single cluster", func() {
		It("list status is correct", func() {
			//
			//   set single-cluster mode
			//
			backendMock = NewBackendMockClient(1, 1)
			orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
			orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))

			clusterMock := cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
			cluster.SetMock(clusterMock)

			tenantIdStr := fmt.Sprintf("%d", tenantId)

			clusters := backendMock.ClusterIdPerTenant[tenantIdStr]

			//
			//   create department + tenant policy
			//
			departPolicy := createWorkspacePolicyForStatusTest(router, departWorkspacePolicy)
			projectPolicy := createWorkspacePolicyForStatusTest(router, projectWorkspacePolicy)

			idsToNick := map[string]string{
				clusters[0]:           "cluster",
				departPolicy.Meta.Id:  "dept-policy",
				projectPolicy.Meta.Id: "proj-policy",
			}

			//
			//   sync the cluster with the workspaces
			//
			validateListSyncStatus(router, tenantId, "cluster 0 received the updates", idsToNick, map[PolicyAndClusterId]api.PolicySyncStatusOfCluster{
				NewPolicyAndClusterId(departPolicy.Meta.Id, clusters[0], idsToNick):  NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
				NewPolicyAndClusterId(projectPolicy.Meta.Id, clusters[0], idsToNick): NewPolicySyncStatusOfCluster(clusters[0], PolicyStatusReady, nil),
			})

			//
			//   now the policy is being deleted
			//
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=department&departmentId=111", urlPolicyWorkspace))
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			//
			//   deletion 2 processed
			//
			validateListSyncStatus(router, tenantId, "cluster 0 received the updates", idsToNick, map[PolicyAndClusterId]api.PolicySyncStatusOfCluster{})
		})
	})
})

type PolicyAndClusterId struct {
	PolicyId  string
	ClusterId string
}

func NewPolicyAndClusterId(policyId string, clusterId string, idsToNick map[string]string) PolicyAndClusterId {
	return PolicyAndClusterId{
		PolicyId:  fmt.Sprintf("%s (%s)", idsToNick[policyId], policyId),
		ClusterId: fmt.Sprintf("%s (%s)", idsToNick[clusterId], clusterId),
	}
}

func validateListSyncStatus(router chi.Router, tenantId int32, description string, idsToNick map[string]string, expected map[PolicyAndClusterId]api.PolicySyncStatusOfCluster) {
	actualResponse := listPoliciesForTest(router, tenantId, urlPolicy)

	actualStatus := make(map[PolicyAndClusterId]api.PolicySyncStatusOfCluster)
	for _, policy := range actualResponse.Policies {
		for _, clusterStatus := range policy.Status.Clusters {
			actualStatus[NewPolicyAndClusterId(policy.Meta.Id, clusterStatus.ClusterId, idsToNick)] = clusterStatus
		}
	}

	Expect(actualStatus).To(Equal(expected), description)
}

func createWorkspacePolicyForStatusTest(router chi.Router, meta api.PolicyCreationFields) *api.WorkspacePolicyV2 {
	workspaceBody := api.WorkspacePolicyOverwriteRequestV2{
		Meta: meta,
		Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
			Defaults: &api.WorkspacePolicyDefaultsV2{
				Security: &api.SupersetDefaultsAllOfSecurity{
					RunAsUid: convutils.Ref(int64(1)),
				},
			},
		},
	}

	workspaceRawResp := overwriteWorkspacePolicyForTest(router, &workspaceBody, 1)

	var workspaceResp api.WorkspacePolicyV2
	err := json.Unmarshal(workspaceRawResp.Body.Bytes(), &workspaceResp)
	Expect(err).To(BeNil())

	return &workspaceResp
}

func createTrainingPolicyForStatusTest(router chi.Router, meta api.PolicyCreationFields) *api.TrainingPolicyV2 {
	trainingBody := api.TrainingPolicyOverwriteRequestV2{
		Meta: meta,
		Policy: &api.TrainingPolicyDefaultsAndRulesV2{
			Defaults: &api.TrainingPolicyDefaultsV2{
				Security: &api.SupersetDefaultsAllOfSecurity{
					RunAsUid: convutils.Ref(int64(1)),
				},
			},
		},
	}

	trainingRawResp := overwriteTrainingPolicyForTest(router, &trainingBody, 1)

	var trainingResp api.TrainingPolicyV2
	err := json.Unmarshal(trainingRawResp.Body.Bytes(), &trainingResp)
	Expect(err).To(BeNil())

	return &trainingResp
}

func createDistributedPolicyForStatusTest(router chi.Router, meta api.PolicyCreationFields) *api.DistributedPolicyV2 {
	distributedBody := api.DistributedPolicyOverwriteRequestV2{
		Meta: meta,
		Policy: &api.DistributedPolicyDefaultsAndRulesV2{
			Defaults: &api.DistributedPolicyDefaultsV2{
				Worker: &api.DistributedPolicyDefaultsV2Worker{
					Security: &api.SupersetDefaultsAllOfSecurity{
						RunAsUid: convutils.Ref(int64(1)),
					},
				},
				Master: &api.ReplicaDefaultsV2{
					Security: &api.SupersetDefaultsAllOfSecurity{
						RunAsUid: convutils.Ref(int64(2)),
					},
				},
			},
		},
	}

	distributedRawResp := overwriteDistributedPolicyForTest(router, &distributedBody, 1)

	var distributedResp api.DistributedPolicyV2
	err := json.Unmarshal(distributedRawResp.Body.Bytes(), &distributedResp)
	Expect(err).To(BeNil())

	return &distributedResp
}

func createDistributedInferencePolicyForStatusTest(router chi.Router, meta api.PolicyCreationFields) *api.DistributedInferencePolicyV2 {
	distributedInferenceBody := api.DistributedInferencePolicyOverwriteRequestV2{
		Meta: meta,
		Policy: &api.DistributedInferencePolicyDefaultsAndRulesV2{
			Defaults: &api.DistributedInferenceDefaultsV2{
				Leader: &api.DistributedInferenceLeaderWorkerDefaultsV2{
					Security: &api.DistributedInferenceLeaderWorkerDefaultsV2Security{
						RunAsUid: convutils.Ref(int64(1)),
					},
				},
			},
		},
	}

	distributedInferenceRawResp := overwriteDistributedInferencePolicyForTest(router, &distributedInferenceBody, 1)

	var distributedInferenceResp api.DistributedInferencePolicyV2
	err := json.Unmarshal(distributedInferenceRawResp.Body.Bytes(), &distributedInferenceResp)
	Expect(err).To(BeNil())

	return &distributedInferenceResp
}

func createInferencePolicyForStatusTest(router chi.Router, meta api.PolicyCreationFields) *api.InferencePolicyV2 {
	inferenceBody := api.InferencePolicyOverwriteRequestV2{
		Meta: meta,
		Policy: &api.InferencePolicyDefaultsAndRulesV2{
			Defaults: &api.InferencePolicyDefaultsV2{
				Security: &api.InferencePolicyDefaultsV2AllOfSecurity{
					RunAsUid: convutils.Ref(int64(1)),
				},
			},
		},
	}

	inferenceRawResp := overwriteInferencePolicyForTest(router, &inferenceBody, 1)

	var inferenceResp api.InferencePolicyV2
	err := json.Unmarshal(inferenceRawResp.Body.Bytes(), &inferenceResp)
	Expect(err).To(BeNil())

	return &inferenceResp
}

func NewPolicySyncStatusOfCluster(clusterId string, status string, errorMessage *string) api.PolicySyncStatusOfCluster {
	return api.PolicySyncStatusOfCluster{
		ClusterId:    clusterId,
		Status:       status,
		ErrorMessage: errorMessage,
	}
}
