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

// ResourceClaim coverage focuses on the single project-scope sync surface,
// since claims are not created from assets - only via the Kubernetes resource
// claim controller - and therefore cannot exist in cluster or department scope.

var _ = Describe("resource claims service", func() {
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

	It("round-trips sync + list with template name preserved", func() {
		claims := api.ResourceClaims{Items: []api.ResourceClaim{
			{
				Name:                      "pod-a-claim-0-abc",
				Namespace:                 fmt.Sprintf("runai-%d", randomScopes.ProjectId),
				ResourceClaimTemplateName: ptr.To("rct-gpu-1"),
				Spec:                      map[string]interface{}{"devices": "req-1"},
			},
		}}
		applyRcs(randomScopes, randomScopes.RcsProjectSyncUrl(), claims)

		list := listRcs(randomScopes, randomScopes.RcsProjectListUrl())
		Expect(list.Items).To(HaveLen(1))
		Expect(list.Items[0].Name).To(Equal("pod-a-claim-0-abc"))
		Expect(list.Items[0].ResourceClaimTemplateName).To(Equal(ptr.To("rct-gpu-1")))
		Expect(list.Items[0].Spec).To(HaveKey("devices"))
	})

	It("prunes stale claims within the namespace on re-sync", func() {
		claims := api.ResourceClaims{Items: []api.ResourceClaim{
			{Name: "claim-a", Namespace: fmt.Sprintf("runai-%d", randomScopes.ProjectId)},
			{Name: "claim-b", Namespace: fmt.Sprintf("runai-%d", randomScopes.ProjectId)},
		}}
		applyRcs(randomScopes, randomScopes.RcsProjectSyncUrl(), claims)

		trimmed := api.ResourceClaims{Items: claims.Items[:1]}
		applyRcs(randomScopes, randomScopes.RcsProjectSyncUrl(), trimmed)

		list := listRcs(randomScopes, randomScopes.RcsProjectListUrl())
		Expect(namesOfRcs(list)).To(ConsistOf("claim-a"))
	})
})

func (rs *RandomScopes) RcsProjectSyncUrl() string {
	return fmt.Sprintf("/api/v1/resource-claims/%s/namespaces/runai-%d", rs.ClusterId, rs.ProjectId)
}

func (rs *RandomScopes) RcsProjectListUrl() string {
	return fmt.Sprintf("/api/v1/resource-claims?clusterId=%s&scopeType=project&scopeId=%s", rs.ClusterId, rs.ProjectIdStr)
}

func applyRcs(rs *RandomScopes, url string, payload api.ResourceClaims) {
	req, err := http.NewRequest("POST", url, makePayload(payload))
	Expect(err).To(BeNil())
	req.Header = *headerForAdmin(rs)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("POST %s body=%s", url, resp.Body.String()))
}

func listRcs(rs *RandomScopes, url string) api.ResourceClaims {
	req, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	req.Header = *headerForAdmin(rs)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("GET %s body=%s", url, resp.Body.String()))

	return unmarshalRcs(resp.Body)
}

func unmarshalRcs(buf *bytes.Buffer) api.ResourceClaims {
	result := api.ResourceClaims{Items: make([]api.ResourceClaim, 0)}
	Expect(json.Unmarshal(buf.Bytes(), &result)).To(Succeed())
	return result
}

func namesOfRcs(rcs api.ResourceClaims) []string {
	names := make([]string, 0, len(rcs.Items))
	for _, rc := range rcs.Items {
		names = append(names, rc.Name)
	}
	return names
}
