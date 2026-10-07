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

// Test coverage for the collapsed RCT API: a single per-item CRUD surface
// (POST/PUT/DELETE on the base path) plus a single Sync endpoint whose body
// may carry items at any scope. Each sync call still performs a per-scope
// prune so templates in windows the caller did not mention are left alone.

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

	It("round-trips cluster-scope sync and list payload", func() {
		rcts := api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			clusterRct(randomScopes, "rct-cluster-1", withOrigin(api.ORIGIN_ASSETS), withAssetId("asset-1")),
			clusterRct(randomScopes, "rct-cluster-2", withOrigin(api.ORIGIN_EXISTING)),
		}}

		syncRcts(randomScopes, rcts)

		list := listRcts(randomScopes, randomScopes.RctsClusterListUrl())
		Expect(list.Items).To(HaveLen(2))

		byName := map[string]api.ResourceClaimTemplate{}
		for _, item := range list.Items {
			byName[item.Name] = item
		}
		Expect(byName["rct-cluster-1"].Namespace).To(Equal("runai"))
		Expect(byName["rct-cluster-1"].ScopeType).To(Equal("cluster"))
		Expect(byName["rct-cluster-1"].ScopeId).To(Equal(randomScopes.ClusterId))
		Expect(byName["rct-cluster-1"].Origin).ToNot(BeNil())
		Expect(*byName["rct-cluster-1"].Origin).To(Equal(api.ORIGIN_ASSETS))
		Expect(byName["rct-cluster-1"].AssetId).To(Equal(ptr.To("asset-1")))
	})

	It("sync accepts a mixed-scope payload and prunes cluster-wide", func() {
		mixed := api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			clusterRct(randomScopes, "rct-foo"),
			clusterRct(randomScopes, "rct-bar"),
			departmentRct(randomScopes, "rct-foo"), // same name as cluster scope - must not collide
			projectRct(randomScopes, "rct-foo"),
		}}
		syncRcts(randomScopes, mixed)

		// Second sync only mentions one cluster-scope row. Everything the
		// payload did not surface - including the department and project rows
		// - must be pruned, because the agent is reporting the full cluster
		// inventory.
		trimmed := api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			clusterRct(randomScopes, "rct-bar"),
		}}
		syncRcts(randomScopes, trimmed)

		Expect(namesOfRcts(listRcts(randomScopes, randomScopes.RctsClusterListUrl()))).To(ConsistOf("rct-bar"))
		Expect(listRcts(randomScopes, randomScopes.RctsDepartmentListUrl()).Items).To(BeEmpty())
		Expect(listRcts(randomScopes, randomScopes.RctsProjectListUrl()).Items).To(BeEmpty())
	})

	It("sync preserves cluster-scope rows with duplicate names across scopes", func() {
		mixed := api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			clusterRct(randomScopes, "rct-foo"),
			departmentRct(randomScopes, "rct-foo"),
			projectRct(randomScopes, "rct-foo"),
		}}
		syncRcts(randomScopes, mixed)

		// Re-send the same payload: every row must survive even though the
		// name repeats across scopes. (A name-based prune would wipe two of
		// the three rows here; the id-keyed prune does not.)
		syncRcts(randomScopes, mixed)

		Expect(namesOfRcts(listRcts(randomScopes, randomScopes.RctsClusterListUrl()))).To(ConsistOf("rct-foo"))
		Expect(namesOfRcts(listRcts(randomScopes, randomScopes.RctsDepartmentListUrl()))).To(ConsistOf("rct-foo"))
		Expect(namesOfRcts(listRcts(randomScopes, randomScopes.RctsProjectListUrl()))).To(ConsistOf("rct-foo"))
	})

	It("filters list by name and origin", func() {
		rcts := api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			clusterRct(randomScopes, "rct-assets", withOrigin(api.ORIGIN_ASSETS), withAssetId("a-1")),
			clusterRct(randomScopes, "rct-existing", withOrigin(api.ORIGIN_EXISTING)),
		}}
		syncRcts(randomScopes, rcts)

		byName := listRcts(randomScopes, randomScopes.RctsClusterListUrl()+"&name=rct-assets")
		Expect(namesOfRcts(byName)).To(ConsistOf("rct-assets"))

		byOrigin := listRcts(randomScopes, randomScopes.RctsClusterListUrl()+"&origin=existing")
		Expect(namesOfRcts(byOrigin)).To(ConsistOf("rct-existing"))

		all := listRcts(randomScopes, randomScopes.RctsClusterListUrl()+"&origin=all")
		Expect(namesOfRcts(all)).To(ConsistOf("rct-assets", "rct-existing"))
	})

	It("per-item create and update at the same name are idempotent", func() {
		resp := createRct(randomScopes, clusterRct(randomScopes, "rct-one"))
		Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

		resp = updateRct(randomScopes, clusterRct(randomScopes, "rct-one", withNamespace("runai-updated")))
		Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

		list := listRcts(randomScopes, randomScopes.RctsClusterListUrl())
		Expect(list.Items).To(HaveLen(1))
		Expect(list.Items[0].Name).To(Equal("rct-one"))
		Expect(list.Items[0].Namespace).To(Equal("runai-updated"))
	})

	It("per-item create does not prune siblings in the same window", func() {
		syncRcts(randomScopes, api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			clusterRct(randomScopes, "rct-a"),
			clusterRct(randomScopes, "rct-b"),
		}})

		resp := createRct(randomScopes, clusterRct(randomScopes, "rct-c"))
		Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

		Expect(namesOfRcts(listRcts(randomScopes, randomScopes.RctsClusterListUrl()))).To(ConsistOf("rct-a", "rct-b", "rct-c"))
	})

	It("per-item delete removes exactly one template", func() {
		syncRcts(randomScopes, api.ResourceClaimTemplates{Items: []api.ResourceClaimTemplate{
			clusterRct(randomScopes, "rct-a"),
			clusterRct(randomScopes, "rct-b"),
			departmentRct(randomScopes, "rct-a"),
		}})

		resp := deleteRct(randomScopes, "cluster", randomScopes.ClusterId, "rct-a")
		Expect(resp.Code).To(Equal(http.StatusNoContent), resp.Body.String())

		Expect(namesOfRcts(listRcts(randomScopes, randomScopes.RctsClusterListUrl()))).To(ConsistOf("rct-b"))
		// department-scope rct-a must survive the cluster-scope delete.
		Expect(namesOfRcts(listRcts(randomScopes, randomScopes.RctsDepartmentListUrl()))).To(ConsistOf("rct-a"))
	})

	It("rejects cluster-scope create whose scopeId does not equal clusterId", func() {
		rct := clusterRct(randomScopes, "rct-mismatch")
		rct.ScopeId = "not-the-cluster-id"

		resp := createRct(randomScopes, rct)
		Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
	})

	It("rejects create whose scopeType is missing a scopeId", func() {
		rct := departmentRct(randomScopes, "rct-no-scope")
		rct.ScopeId = ""

		resp := createRct(randomScopes, rct)
		Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
	})
})

// --- URL helpers ------------------------------------------------------------

func (rs *RandomScopes) rctsBaseUrl() string {
	return fmt.Sprintf("/api/v1/resource-claim-templates?clusterId=%s", rs.ClusterId)
}

func (rs *RandomScopes) rctsSyncUrl() string {
	return fmt.Sprintf("/api/v1/resource-claim-templates/sync?clusterId=%s", rs.ClusterId)
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

// --- RCT builders -----------------------------------------------------------

type rctOption func(*api.ResourceClaimTemplate)

func withOrigin(o api.Origin) rctOption {
	return func(r *api.ResourceClaimTemplate) {
		v := o
		r.Origin = &v
	}
}

func withAssetId(id string) rctOption {
	return func(r *api.ResourceClaimTemplate) {
		s := id
		r.AssetId = &s
	}
}

func withNamespace(ns string) rctOption {
	return func(r *api.ResourceClaimTemplate) {
		r.Namespace = ns
	}
}

func clusterRct(rs *RandomScopes, name string, opts ...rctOption) api.ResourceClaimTemplate {
	rct := api.ResourceClaimTemplate{
		Name:      name,
		Namespace: "runai",
		ScopeType: "cluster",
		ScopeId:   rs.ClusterId,
	}
	for _, o := range opts {
		o(&rct)
	}
	return rct
}

func departmentRct(rs *RandomScopes, name string, opts ...rctOption) api.ResourceClaimTemplate {
	rct := api.ResourceClaimTemplate{
		Name:      name,
		Namespace: "runai",
		ScopeType: "department",
		ScopeId:   rs.DepartmentId,
	}
	for _, o := range opts {
		o(&rct)
	}
	return rct
}

func projectRct(rs *RandomScopes, name string, opts ...rctOption) api.ResourceClaimTemplate {
	rct := api.ResourceClaimTemplate{
		Name:      name,
		Namespace: fmt.Sprintf("runai-%d", rs.ProjectId),
		ScopeType: "project",
		ScopeId:   rs.ProjectIdStr,
	}
	for _, o := range opts {
		o(&rct)
	}
	return rct
}

// --- request helpers --------------------------------------------------------

func createRct(rs *RandomScopes, rct api.ResourceClaimTemplate) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", rs.rctsBaseUrl(), makePayload(rct))
	Expect(err).To(BeNil())
	req.Header = *headerForAdmin(rs)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func updateRct(rs *RandomScopes, rct api.ResourceClaimTemplate) *httptest.ResponseRecorder {
	req, err := http.NewRequest("PUT", rs.rctsBaseUrl(), makePayload(rct))
	Expect(err).To(BeNil())
	req.Header = *headerForAdmin(rs)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func deleteRct(rs *RandomScopes, scopeType, scopeId, name string) *httptest.ResponseRecorder {
	url := fmt.Sprintf("%s&scopeType=%s&scopeId=%s&name=%s", rs.rctsBaseUrl(), scopeType, scopeId, name)
	req, err := http.NewRequest("DELETE", url, nil)
	Expect(err).To(BeNil())
	req.Header = *headerForAdmin(rs)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func syncRcts(rs *RandomScopes, payload api.ResourceClaimTemplates) {
	req, err := http.NewRequest("POST", rs.rctsSyncUrl(), makePayload(payload))
	Expect(err).To(BeNil())
	req.Header = *headerForAdmin(rs)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("POST %s body=%s", rs.rctsSyncUrl(), resp.Body.String()))
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
