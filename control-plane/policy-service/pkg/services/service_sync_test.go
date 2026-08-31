/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"fmt"

	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	orgUnitService "github.com/run-ai/runai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	scope "github.com/run-ai/runai/runai-application-packages/scopeinfo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/clients/assets"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/clients/cluster"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/tree"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
)

var (
	urlSyncWorkspace = "/api/v2/policy/workspaces/sync"
	urlSyncConfig    = "/api/v2/policy/sync-config"
)

var samplePolicy = api.WorkspacePolicyDefaultsAndRulesV2{
	Defaults: &api.WorkspacePolicyDefaultsV2{
		Image: convutils.Ref("ubuntu"),
	},
}

var _ = Describe("Apply delete impact of no sync", func() {
	router := CreateRouter()

	var backendMock *BackendMockClient
	var clusterMock cluster.SimpleMockClient
	var oldClusterId string

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()

		tree.Restart()

		// lets have 2 tenants x 2 clusters
		backendMock = NewBackendMockClient(2, 2)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))
		authorization.SetMock(NewPoliciesAuthClient(backendMock))

		// arrange the clusters so that
		// tenant 1 both clusters are ok
		// tenant 2 has one old cluster, one new cluster
		clusterMock = cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		oldClusterId = backendMock.ClusterIdPerTenant["2"][0]
		clusterMock.Clusters[oldClusterId].Version = *clusterApi.NewNullableString(convutils.Ref("2.20.0"))
		cluster.SetMock(clusterMock)

		assets.SetMock(assets.SimpleMockClientFactory{
			SimpleMockClient: assets.NewSimpleMockClient(),
		})
	})

	Context("try to update tenant policy with no approval", func() {
		It("should be ok for new clusters, not for old", func() {
			tenant1policy := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &samplePolicy,
			}

			tenant2policy := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t2",
				},
				Policy: &samplePolicy,
			}

			// policy for tenant 1 we can apply, because all clusters there are new
			applyPolicyForTest(router, urlPolicyWorkspace, makePayload(tenant1policy), 1, http.MethodPatch)

			// no updates are sent to the new cluster, and old cluster did not approve so nothing send
			// to it either
			for _, cluster := range clusterMock.Clusters {
				updates := getUpdatesForSync(router, urlSyncWorkspace, cluster, 1)
				Expect(len(updates.Policies)).To(Equal(0))
			}

			// we are also able to delete it
			expectError(router, fmt.Sprintf("%s?scope=%s", urlPolicyWorkspace, scope.Tenant), makePayload(tenant1policy), 1, http.MethodDelete, http.StatusNoContent)

			// now that we deleted it, all clusters of tenant 1 should get deletion event because deletions
			// are populated to all the clusters no matter if approved or not
			for _, cluster := range clusterMock.Clusters {
				updates := getUpdatesForSync(router, urlSyncWorkspace, cluster, 1)
				if cluster.TenantId == 1 {
					Expect(len(updates.Policies)).To(Equal(1))
					Expect(convutils.BoolIsTrue(updates.Policies[0].IsDeleted)).To(BeTrue())
				} else {
					Expect(len(updates.Policies)).To(Equal(0))
				}
			}

			msg := expectError(router, urlPolicyWorkspace, makePayload(tenant2policy), 2, http.MethodPatch, http.StatusPreconditionFailed)
			Expect(msg).To(ContainSubstring("running version 2.20 or earlier. In such case, the platform needs your approval"))

			// delete is always allowed, but since the policy already deleted it should return error 404
			expectError(router, fmt.Sprintf("%s?scope=%s", urlPolicyWorkspace, scope.Tenant), makePayload(tenant2policy), 2, http.MethodDelete, http.StatusNotFound)

			// no change in the updates from last time
			for _, cluster := range clusterMock.Clusters {
				updates := getUpdatesForSync(router, urlSyncWorkspace, cluster, 1)
				if cluster.TenantId == 1 {
					Expect(len(updates.Policies)).To(Equal(1))
					Expect(convutils.BoolIsTrue(updates.Policies[0].IsDeleted)).To(BeTrue())
				} else {
					Expect(len(updates.Policies)).To(Equal(0))
				}
			}

			// update the policy of tenant 2 + force delete the cluster removal approve-deletion flag
			applyPolicyForTest(router, fmt.Sprintf("%s?approveClusterDeletion=true", urlPolicyWorkspace), makePayload(tenant2policy), 2, http.MethodPatch)

			// now the situation is this:
			// tenant 1 -> both cluster gets deletion from the recent DELETION
			// tenant 2 ->
			//    new cluster no updates
			//    old cluster -> receives deletion for this update (because now it is approved)
			for _, cluster := range clusterMock.Clusters {
				updates := getUpdatesForSync(router, urlSyncWorkspace, cluster, 1)
				if cluster.TenantId == 2 && cluster.Uuid != oldClusterId {
					Expect(len(updates.Policies)).To(Equal(0))
				} else {
					Expect(len(updates.Policies)).To(Equal(1))
					Expect(convutils.BoolIsTrue(updates.Policies[0].IsDeleted)).To(BeTrue())
				}
			}
		})
	})

	Context("try to update cluster policy with no approval", func() {
		It("should be ok for new clusters, not for old", func() {
			newClusterId := backendMock.ClusterIdPerTenant["2"][1]

			newClusterPol := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					ClusterId: convutils.Ref(newClusterId),
					Name:      "pol-cl-new",
				},
				Policy: &samplePolicy,
			}

			applyPolicyForTest(router, urlPolicyWorkspace, makePayload(newClusterPol), 2, http.MethodPatch)

			oldClusterPol := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					ClusterId: convutils.Ref(oldClusterId),
					Name:      "pol-cl-old",
				},
				Policy: &samplePolicy,
			}

			msg := expectError(router, urlPolicyWorkspace, makePayload(oldClusterPol), 2, http.MethodPatch, http.StatusPreconditionFailed)
			Expect(msg).To(ContainSubstring("running version 2.20 or earlier. In such case, the platform needs your approval"))

			// delete both of them, in the old cluster policy was rejected so it does not exist
			expectError(router, fmt.Sprintf("%s?scope=%s&clusterId=%s", urlPolicyWorkspace, scope.Cluster, newClusterId), makePayload(newClusterPol), 2, http.MethodDelete, http.StatusNoContent)
			expectError(router, fmt.Sprintf("%s?scope=%s&clusterId=%s", urlPolicyWorkspace, scope.Cluster, oldClusterId), makePayload(oldClusterPol), 2, http.MethodDelete, http.StatusNotFound)

			// tenant 2 - we get updates for deletion the new cluster, tenant 1 and old cluster - nothing to update
			for _, cluster := range clusterMock.Clusters {
				updates := getUpdatesForSync(router, urlSyncWorkspace, cluster, 1)
				if cluster.TenantId == 2 && cluster.Uuid != oldClusterId {
					Expect(len(updates.Policies)).To(Equal(1))
					Expect(convutils.BoolIsTrue(updates.Policies[0].IsDeleted)).To(BeTrue())
				} else {
					Expect(len(updates.Policies)).To(Equal(0))
				}
			}

			// now there is approval for deletion, the backend returns deletion message
			applyPolicyForTest(router, fmt.Sprintf("%s?approveClusterDeletion=true", urlPolicyWorkspace), makePayload(oldClusterPol), 2, http.MethodPatch)

			updates := getUpdatesForSync(router, urlSyncWorkspace, clusterMock.Clusters[oldClusterId], 1)
			Expect(len(updates.Policies)).To(Equal(1))
			Expect(convutils.BoolIsTrue(updates.Policies[0].IsDeleted)).To(Equal(true))

			// old cluster upgraded to 2.22
			clusterMock.Clusters[oldClusterId].Version = *clusterApi.NewNullableString(convutils.Ref("2.22.0"))
			cluster.SetMock(clusterMock)

			// no need for approval anymore
			applyPolicyForTest(router, urlPolicyWorkspace, makePayload(oldClusterPol), 2, http.MethodPatch)

			// now we expect to receive only deletion requests, just like in new cluster
			// and since we UPDATED the policy, we do not get anything
			updates = getUpdatesForSync(router, urlSyncWorkspace, clusterMock.Clusters[oldClusterId], 1)
			Expect(len(updates.Policies)).To(Equal(0))

			// but if we delete it, we get the deletion event in the cluster
			deletePolicyForTest(router, 2, fmt.Sprintf("%s?scope=cluster&clusterId=%s", urlPolicyWorkspace, oldClusterId))

			updates = getUpdatesForSync(router, urlSyncWorkspace, clusterMock.Clusters[oldClusterId], 1)
			Expect(len(updates.Policies)).To(Equal(1))
			Expect(convutils.BoolIsTrue(updates.Policies[0].IsDeleted)).To(Equal(true))
		})
	})

	Context("simulate scenario that user is not willing to delete", func() {
		It("should send back no-changes to the agent", func() {
			oldClusterPol := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					ClusterId: convutils.Ref(oldClusterId),
					Name:      "pol-cl-old",
				},
				Policy: &samplePolicy,
			}

			msg := expectError(router, urlPolicyWorkspace, makePayload(oldClusterPol), 2, http.MethodPatch, http.StatusPreconditionFailed)
			Expect(msg).To(ContainSubstring("running version 2.20 or earlier. In such case, the platform needs your approval"))

			// for all the clusters trying to sync should return "no changes" because
			// there is no approval to delete anything on the cluster
			for _, cluster := range clusterMock.Clusters {
				updates := getUpdatesForSync(router, urlSyncWorkspace, cluster, 1)
				Expect(len(updates.Policies)).To(Equal(0))
			}

			// update twice to verify both create and update paths work (upsert behaviour)
			applySyncConfig(router, urlSyncConfig, makePayload(api.SyncConfig{
				ClusterId:                        oldClusterId,
				AvoidDeletionsOfObsoletePolicies: false,
			}), 2)

			applySyncConfig(router, urlSyncConfig, makePayload(api.SyncConfig{
				ClusterId:                        oldClusterId,
				AvoidDeletionsOfObsoletePolicies: true,
			}), 2)

			// now it is ok to apply the policy
			applyPolicyForTest(router, fmt.Sprintf("%s?approveClusterDeletion=true", urlPolicyWorkspace), makePayload(oldClusterPol), 2, http.MethodPatch)

			updates := getUpdatesForSync(router, urlSyncWorkspace, clusterMock.Clusters[oldClusterId], 1)
			Expect(len(updates.Policies)).To(Equal(0))

			// let's also try upgrading the old cluster to 2.22
			clusterMock.Clusters[oldClusterId].Version = *clusterApi.NewNullableString(convutils.Ref("2.22.0"))
			cluster.SetMock(clusterMock)

			// no need for approval anymore
			applyPolicyForTest(router, urlPolicyWorkspace, makePayload(oldClusterPol), 2, http.MethodPatch)

			// now we expect to receive no-change updates and not "deletion" updates
			updates = getUpdatesForSync(router, urlSyncWorkspace, clusterMock.Clusters[oldClusterId], 1)
			Expect(len(updates.Policies)).To(Equal(0))
		})
	})
})

func getUpdatesForSync(router chi.Router, url string, cluster *clusterApi.DisplayedCluster, fromUpdate int64) api.PolicySyncResponse {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?clusterId=%s&after=%d", url, cluster.Uuid, fromUpdate), nil)
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(cluster.TenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(tenantContextForTest(cluster.TenantId)))
	Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

	var result api.PolicySyncResponse
	jsonErr := json.Unmarshal(response.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	return result
}
