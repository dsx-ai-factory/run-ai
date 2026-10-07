/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/clusters"
	orgUnit "github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dbtesting"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test coverage for the three scope-split RCT sync endpoints + the shared list
// endpoint. Each sync endpoint is treated as a full-set replace within its scope
// so the prune-on-sync behaviour is exercised explicitly; separate scopes must
// not clobber each other even when template names collide.

var _ = Describe("resource claim templates service", func() {
	var randomScopes *RandomScopes

	BeforeEach(func() {
		randomScopes = NewRandomsSet()

		backendMock := NewRandomsBasedBackendMockClient(randomScopes)
		orgUnit.SetMock(backendMock)
		authorization.SetMock(backendMock)

		router = CreateRouter()
		dbtesting.CleanTables()

		clusters.SetMock(clusters.NewSimpleMock().WithCluster(randomScopes.ClusterId, &clustersApi.DisplayedCluster{}))
	})

	It("round-trips cluster-scope sync, list, and spec payload", func() {
		rcts := api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			{
				Name:      "rct-cluster-1",
				Namespace: "runai",
				Origin:    originAssetsPtr(),
				AssetId:   ptr.To("asset-1"),
				Spec:      map[string]interface{}{"spec": map[string]interface{}{"devices": "cluster-1"}},
			},
			{
				Name:      "rct-cluster-2",
				Namespace: "runai",
				Origin:    originExistingPtr(),
			},
		}}

		applyRcts(randomScopes, randomScopes.RctsClusterSyncUrl(), rcts)

		list := listRcts(randomScopes, randomScopes.RctsClusterListUrl())
		Expect(list.Items).To(HaveLen(2))

		byName := map[string]api.ResourceClaimTemplate{}
		for _, item := range list.Items {
			byName[item.Name] = item
		}
		Expect(byName["rct-cluster-1"].Namespace).To(Equal("runai"))
		Expect(byName["rct-cluster-1"].Origin).ToNot(BeNil())
		Expect(*byName["rct-cluster-1"].Origin).To(Equal(api.ORIGIN_ASSETS))
		Expect(byName["rct-cluster-1"].AssetId).To(Equal(ptr.To("asset-1")))
		Expect(byName["rct-cluster-1"].Spec).To(HaveKey("spec"))
	})

	It("prunes stale rows within a scope but leaves other scopes alone", func() {
		clusterRcts := api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			{Name: "rct-foo", Namespace: "runai"},
			{Name: "rct-bar", Namespace: "runai"},
		}}
		applyRcts(randomScopes, randomScopes.RctsClusterSyncUrl(), clusterRcts)

		deptRcts := api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			{Name: "rct-foo", Namespace: "runai"}, // same name as cluster scope - must not collide
		}}
		applyRcts(randomScopes, randomScopes.RctsDepartmentSyncUrl(), deptRcts)

		projectRcts := api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			{Name: "rct-foo", Namespace: fmt.Sprintf("runai-%d", randomScopes.ProjectId)},
		}}
		applyRcts(randomScopes, randomScopes.RctsProjectSyncUrl(), projectRcts)

		// Drop rct-foo from the cluster scope sync: only rct-bar should remain
		// there, while department and project rct-foo rows must survive.
		trimmed := api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			{Name: "rct-bar", Namespace: "runai"},
		}}
		applyRcts(randomScopes, randomScopes.RctsClusterSyncUrl(), trimmed)

		clusterList := listRcts(randomScopes, randomScopes.RctsClusterListUrl())
		Expect(namesOfRcts(clusterList)).To(ConsistOf("rct-bar"))

		deptList := listRcts(randomScopes, randomScopes.RctsDepartmentListUrl())
		Expect(namesOfRcts(deptList)).To(ConsistOf("rct-foo"))

		projList := listRcts(randomScopes, randomScopes.RctsProjectListUrl())
		Expect(namesOfRcts(projList)).To(ConsistOf("rct-foo"))
	})

	It("filters list by name and origin", func() {
		rcts := api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			{Name: "rct-assets", Namespace: "runai", Origin: originAssetsPtr(), AssetId: ptr.To("a-1")},
			{Name: "rct-existing", Namespace: "runai", Origin: originExistingPtr()},
		}}
		applyRcts(randomScopes, randomScopes.RctsClusterSyncUrl(), rcts)

		byName := listRcts(randomScopes, randomScopes.RctsClusterListUrl()+"&name=rct-assets")
		Expect(namesOfRcts(byName)).To(ConsistOf("rct-assets"))

		byOrigin := listRcts(randomScopes, randomScopes.RctsClusterListUrl()+"&origin=existing")
		Expect(namesOfRcts(byOrigin)).To(ConsistOf("rct-existing"))

		all := listRcts(randomScopes, randomScopes.RctsClusterListUrl()+"&origin=all")
		Expect(namesOfRcts(all)).To(ConsistOf("rct-assets", "rct-existing"))
	})
})

// --- URL helpers ------------------------------------------------------------

func (rs *RandomScopes) RctsClusterSyncUrl() string {
	return fmt.Sprintf("/api/v1/resource-claim-templates/%s", rs.ClusterId)
}

func (rs *RandomScopes) RctsDepartmentSyncUrl() string {
	return fmt.Sprintf("/api/v1/resource-claim-templates/%s/departments/%s", rs.ClusterId, rs.DepartmentId)
}

func (rs *RandomScopes) RctsProjectSyncUrl() string {
	return fmt.Sprintf("/api/v1/resource-claim-templates/%s/namespaces/runai-%d", rs.ClusterId, rs.ProjectId)
}

func (rs *RandomScopes) RctsClusterListUrl() string {
	return fmt.Sprintf("/api/v1/resource-claim-templates?clusterId=%s&scopeType=cluster&scopeId=%s", rs.ClusterId, rs.ClusterId)
}

func (rs *RandomScopes) RctsDepartmentListUrl() string {
	return fmt.Sprintf("/api/v1/resource-claim-templates?clusterId=%s&scopeType=department&scopeId=%s", rs.ClusterId, rs.DepartmentId)
}

func (rs *RandomScopes) RctsProjectListUrl() string {
	return fmt.Sprintf("/api/v1/resource-claim-templates?clusterId=%s&scopeType=project&scopeId=%s", rs.ClusterId, rs.ProjectIdStr)
}

// --- request helpers --------------------------------------------------------

func applyRcts(rs *RandomScopes, url string, payload api.ResourceClaimTemplates) {
	req, err := http.NewRequest("POST", url, makePayload(payload))
	Expect(err).To(BeNil())
	req.Header = *headerForAdmin(rs)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("POST %s body=%s", url, resp.Body.String()))
}

func listRcts(rs *RandomScopes, url string) api.ResourceClaimTemplates {
	req, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	req.Header = *headerForAdmin(rs)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("GET %s body=%s", url, resp.Body.String()))

	return unmarshalRcts(resp.Body)
}

func unmarshalRcts(buf *bytes.Buffer) api.ResourceClaimTemplates {
	result := api.ResourceClaimTemplates{Items: make([]api.ResourceClaimTemplate, 0)}
	Expect(json.Unmarshal(buf.Bytes(), &result)).To(Succeed())
	return result
}

func namesOfRcts(rcts api.ResourceClaimTemplates) []string {
	names := make([]string, 0, len(rcts.Items))
	for _, rct := range rcts.Items {
		names = append(names, rct.Name)
	}
	return names
}

func originAssetsPtr() *api.Origin {
	v := api.ORIGIN_ASSETS
	return &v
}

func originExistingPtr() *api.Origin {
	v := api.ORIGIN_EXISTING
	return &v
}
