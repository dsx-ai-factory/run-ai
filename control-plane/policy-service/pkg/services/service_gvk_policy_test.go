/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	sm "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/system_mapping"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/tree"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	urlGvkPolicies = "/api/v1/policies"
)

func gvkPolicyURL(id string) string {
	return fmt.Sprintf("%s/%s", urlGvkPolicies, id)
}

func gvkPolicyRulesURL(id string) string {
	return fmt.Sprintf("%s/%s/rules", urlGvkPolicies, id)
}

var _ = Describe("GVK Policy Service", Ordered, func() {
	router := CreateRouter()

	var backendMock *BackendMockClient

	tenantId := int32(1)
	projectId := "1111"
	gvk := api.Gvk{Group: convutils.Ref("apps"), Kind: "Deployment", Version: "v1"}

	sampleRules := func() []api.PolicyRule {
		return []api.PolicyRule{
			{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			{Field: "imagePullPolicy", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}}},
		}
	}

	newProjectPolicy := func(rules []api.PolicyRule) api.CreateGvkPolicyReq {
		return api.CreateGvkPolicyReq{
			WorkloadType: gvk,
			Scope:        api.SCOPETYPE_PROJECT,
			ProjectId:    convutils.Ref(projectId),
			Rules:        rules,
		}
	}

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()
		tree.Restart()

		backendMock = NewBackendMockClient(2, 1)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))
		authorization.SetMock(NewPoliciesAuthClient(backendMock))

		clusterMock := cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		cluster.SetMock(clusterMock)

		assets.SetMock(assets.SimpleMockClientFactory{
			SimpleMockClient: assets.NewSimpleMockClient(),
		})

		// System formatters / field groups / mappings are seeded once at suite
		// start (TestServices -> HandleSystemMapping) and preserved across specs
		// by database.DeleteAllTablesForTesting (which only removes non-system
		// rows from the mapping tables). No need to re-seed here.
	})

	Context("create", func() {
		It("creates a project-scoped GVK policy", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			Expect(created.Id).NotTo(BeEmpty())
			Expect(created.WorkloadType).To(Equal(gvk))
			Expect(created.Scope).To(Equal(api.SCOPETYPE_PROJECT))
			Expect(created.Rules).To(HaveLen(2))
			expectEffectiveRules(created, 2)
		})

		It("rejects create with empty rules", func() {
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(newProjectPolicy([]api.PolicyRule{})))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("rejects unknown field", func() {
			req := newProjectPolicy([]api.PolicyRule{
				{Field: "nonExistentField", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			})
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			// breakdown generator now reports unresolved field/specSelector
			// pairs explicitly (it runs first, with maxErrors=1, and
			// short-circuits validateRules), so the error references the
			// field-group lookup rather than validateRules' "field not
			// found under specSelector".
			Expect(response.Body.String()).To(ContainSubstring("nonExistentField"))
			Expect(response.Body.String()).To(ContainSubstring("allContainers"))
		})

		It("rejects unknown specSelector for a field", func() {
			req := newProjectPolicy([]api.PolicyRule{
				{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				{Field: "image", SpecSelector: "nonExistentSelector", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			})
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			// breakdown generator reports unresolved specSelectors before
			// validateRules runs (maxErrors=1 short-circuits), so the
			// surfaced finding mentions the missing selector by name.
			Expect(response.Body.String()).To(ContainSubstring("nonExistentSelector"))
		})

		It("rejects duplicate rules", func() {
			req := newProjectPolicy([]api.PolicyRule{
				{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}}},
			})
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("duplicate rule"))
		})

		It("rejects enforcement not applicable to field type", func() {
			req := newProjectPolicy([]api.PolicyRule{
				{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1))}}}},
			})
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("not applicable to field type"))
		})

		It("rejects KeyValue set with wrong key", func() {
			req := newProjectPolicy([]api.PolicyRule{
				{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Set: map[string]interface{}{"badKey": "hello"}}}},
			})
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("must use key"))
		})

		It("rejects invalid regex in match enforcement", func() {
			req := newProjectPolicy([]api.PolicyRule{
				{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Match: convutils.Ref("[invalid")}}},
			})
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("not a valid regular expression"))
		})

		It("rejects invalid quantity value", func() {
			req := newProjectPolicy([]api.PolicyRule{
				{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{Min: convutils.Ref("notAQuantity")}}}},
			})
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("not applicable to field type"))
		})

		It("rejects create with invalid GVK (empty kind)", func() {
			req := api.CreateGvkPolicyReq{
				WorkloadType: api.Gvk{Group: convutils.Ref("apps"), Kind: "", Version: "v1"},
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules:        sampleRules(),
			}
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("rejects create when default is longer than length.max (system mapping breakdown)", func() {
			req := newProjectPolicy([]api.PolicyRule{
				{
					Field:        "workingDir",
					SpecSelector: "allContainers",
					Enforcements: []api.PolicyRuleEnforcement{
						{Length: &api.PolicyEnforcementLengthLength{Max: convutils.Ref(int32(2))}},
						{Default: map[string]interface{}{DefaultArgName: "aaa"}},
					},
				},
			})
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			body := response.Body.String()
			Expect(body).To(ContainSubstring("workingDir"))
			Expect(body).To(ContainSubstring("no more than 2"))
		})

		It("rejects create when default is shorter than length.min (system mapping breakdown)", func() {
			req := newProjectPolicy([]api.PolicyRule{
				{
					Field:        "workingDir",
					SpecSelector: "allContainers",
					Enforcements: []api.PolicyRuleEnforcement{
						{Length: &api.PolicyEnforcementLengthLength{Min: convutils.Ref(int32(4))}},
						{Default: map[string]interface{}{DefaultArgName: "aaa"}},
					},
				},
			})
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			body := response.Body.String()
			Expect(body).To(ContainSubstring("workingDir"))
			Expect(body).To(ContainSubstring("no less than 4"))
		})

		It("rejects a second POST for the same (scope, workloadType) with 409 and leaves the original untouched", func() {
			// POST is create-only - the multi-scope refactor uses a
			// soft-deleted placeholder mechanism that, without an explicit
			// duplicate check, would silently turn a second POST into an
			// upsert (overwriting rules, resetting CreatedAt/CreatedBy on
			// the existing row). We require callers to use PUT / PATCH for
			// updates and verify the duplicate POST is rejected.
			first := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))
			Expect(first.Id).NotTo(BeEmpty())

			dupReq, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(newProjectPolicy([]api.PolicyRule{
				{Field: "workingDir", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			})))
			Expect(err).To(BeNil())
			dupReq.Header = *headerForAdmin(tenantId)

			dupResp := httptest.NewRecorder()
			router.ServeHTTP(dupResp, dupReq.WithContext(tenantContextForTest(tenantId)))
			Expect(dupResp.Code).To(Equal(http.StatusConflict), dupResp.Body.String())
			Expect(dupResp.Body.String()).To(ContainSubstring("already exists"))

			// the original policy must be untouched: same id, same rule
			// set (not the duplicate POST's payload), still valid.
			fetched := getGvkPolicyForTest(router, tenantId, first.Id)
			Expect(fetched.Id).To(Equal(first.Id))
			Expect(fetched.Rules).To(HaveLen(len(sampleRules())))
			Expect(fetched.Rules[0].Field).To(Equal(sampleRules()[0].Field))
		})

		It("allows POST again after the existing policy is deleted", func() {
			// the 409 rejection must not trap callers: once the existing
			// policy is gone, a fresh POST at the same scope is the right
			// way to create a new policy.
			first := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			delReq, err := http.NewRequest(http.MethodDelete, gvkPolicyURL(first.Id), nil)
			Expect(err).To(BeNil())
			delReq.Header = *headerForAdmin(tenantId)
			delResp := httptest.NewRecorder()
			router.ServeHTTP(delResp, delReq.WithContext(tenantContextForTest(tenantId)))
			Expect(delResp.Code).To(Equal(http.StatusNoContent))

			second := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))
			Expect(second.Id).NotTo(BeEmpty())
			// a fresh create must produce a different policy id; the
			// original was hard-deleted.
			Expect(second.Id).NotTo(Equal(first.Id))
		})
	})

	Context("get", func() {
		It("returns the created policy by ID", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			fetched := getGvkPolicyForTest(router, tenantId, created.Id)
			Expect(fetched.Id).To(Equal(created.Id))
			Expect(fetched.Rules).To(HaveLen(2))
			expectEffectiveRules(fetched, 2)
		})

		It("returns 404 for non-existent policy", func() {
			request, err := http.NewRequest(http.MethodGet, gvkPolicyURL("00000000-0000-0000-0000-000000000000"), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("list", func() {
		It("lists created GVK policies", func() {
			createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			secondPolicy := api.CreateGvkPolicyReq{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref("1112"),
				Rules:        sampleRules(),
			}
			createGvkPolicyForTest(router, tenantId, secondPolicy)

			list := listGvkPoliciesForTest(router, tenantId, "")
			Expect(list.Items).To(HaveLen(2))
		})

		It("does not return GVK policies from another tenant", func() {
			createGvkPolicyForTest(router, 1, newProjectPolicy(sampleRules()))

			list := listGvkPoliciesForTest(router, 2, "")
			Expect(list.Items).To(BeEmpty())
		})

		It("filters by scope", func() {
			createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			list := listGvkPoliciesForTest(router, tenantId, "filterBy=scope==project")
			Expect(list.Items).To(HaveLen(1))

			list = listGvkPoliciesForTest(router, tenantId, "filterBy=scope==tenant")
			Expect(list.Items).To(BeEmpty())
		})

		It("filters by type (workloadType)", func() {
			createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			batchGvk := api.Gvk{Group: convutils.Ref("batch"), Kind: "Job", Version: "v1"}
			seedMappingForGvk(batchGvk)
			createGvkPolicyForTest(router, tenantId, api.CreateGvkPolicyReq{
				WorkloadType: batchGvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules:        sampleRules(),
			})

			list := listGvkPoliciesForTest(router, tenantId, "filterBy=type==apps/Deployment/v1")
			Expect(list.Items).To(HaveLen(1))
			Expect(convutils.Deref(list.Items[0].WorkloadType.Group, "")).To(Equal("apps"))

			list = listGvkPoliciesForTest(router, tenantId, "filterBy=type==batch/Job/v1")
			Expect(list.Items).To(HaveLen(1))
			Expect(convutils.Deref(list.Items[0].WorkloadType.Group, "")).To(Equal("batch"))
		})

		It("returns 400 on invalid filter field", func() {
			request, err := http.NewRequest(http.MethodGet, urlGvkPolicies+"?filterBy=invalid==value", nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("supports sortBy and sortOrder", func() {
			createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			batchGvk := api.Gvk{Group: convutils.Ref("batch"), Kind: "Job", Version: "v1"}
			seedMappingForGvk(batchGvk)
			createGvkPolicyForTest(router, tenantId, api.CreateGvkPolicyReq{
				WorkloadType: batchGvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules:        sampleRules(),
			})

			listAsc := listGvkPoliciesForTest(router, tenantId, "sortBy=type&sortOrder=asc")
			Expect(listAsc.Items).To(HaveLen(2))
			Expect(convutils.Deref(listAsc.Items[0].WorkloadType.Group, "")).To(Equal("apps"))
			Expect(convutils.Deref(listAsc.Items[1].WorkloadType.Group, "")).To(Equal("batch"))

			listDesc := listGvkPoliciesForTest(router, tenantId, "sortBy=type&sortOrder=desc")
			Expect(listDesc.Items).To(HaveLen(2))
			Expect(convutils.Deref(listDesc.Items[0].WorkloadType.Group, "")).To(Equal("batch"))
			Expect(convutils.Deref(listDesc.Items[1].WorkloadType.Group, "")).To(Equal("apps"))
		})

		It("returns 400 on invalid sortBy field", func() {
			request, err := http.NewRequest(http.MethodGet, urlGvkPolicies+"?sortBy=invalid", nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("supports offset and limit", func() {
			for i := 0; i < 5; i++ {
				g := api.Gvk{Group: convutils.Ref(fmt.Sprintf("g%d", i)), Kind: "K", Version: "v1"}
				seedMappingForGvk(g)
				createGvkPolicyForTest(router, tenantId, api.CreateGvkPolicyReq{
					WorkloadType: g,
					Scope:        api.SCOPETYPE_PROJECT,
					ProjectId:    convutils.Ref(projectId),
					Rules:        sampleRules(),
				})
			}

			allSorted := listGvkPoliciesForTest(router, tenantId, "sortBy=type&sortOrder=asc")
			Expect(allSorted.Items).To(HaveLen(5))

			first2 := listGvkPoliciesForTest(router, tenantId, "sortBy=type&sortOrder=asc&limit=2")
			Expect(first2.Items).To(HaveLen(2))
			Expect(first2.Items[0].Id).To(Equal(allSorted.Items[0].Id))
			Expect(first2.Items[1].Id).To(Equal(allSorted.Items[1].Id))

			last2 := listGvkPoliciesForTest(router, tenantId, "sortBy=type&sortOrder=asc&offset=3&limit=10")
			Expect(last2.Items).To(HaveLen(2))
			Expect(last2.Items[0].Id).To(Equal(allSorted.Items[3].Id))
			Expect(last2.Items[1].Id).To(Equal(allSorted.Items[4].Id))

			middle := listGvkPoliciesForTest(router, tenantId, "sortBy=type&sortOrder=asc&offset=1&limit=2")
			Expect(middle.Items).To(HaveLen(2))
			Expect(middle.Items[0].Id).To(Equal(allSorted.Items[1].Id))
			Expect(middle.Items[1].Id).To(Equal(allSorted.Items[2].Id))
		})

		It("returns empty list for offset beyond total", func() {
			createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			list := listGvkPoliciesForTest(router, tenantId, "offset=100")
			Expect(list.Items).To(BeEmpty())
		})
	})

	Context("invalid policy id format", func() {
		invalidId := "not-a-uuid"

		It("returns 400 on get", func() {
			request, err := http.NewRequest(http.MethodGet, gvkPolicyURL(invalidId), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("returns 400 on put", func() {
			request, err := http.NewRequest(http.MethodPut, gvkPolicyURL(invalidId), makePayload(api.UpdateGvkPolicyReq{
				Rules: sampleRules(),
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("returns 400 on patch", func() {
			request, err := http.NewRequest(http.MethodPatch, gvkPolicyURL(invalidId), makePayload(api.UpdateGvkPolicyReq{
				Rules: sampleRules(),
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("returns 400 on delete", func() {
			request, err := http.NewRequest(http.MethodDelete, gvkPolicyURL(invalidId), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("returns 400 on delete rules", func() {
			request, err := http.NewRequest(http.MethodDelete, gvkPolicyRulesURL(invalidId), makePayload(api.DeleteGvkPolicyRulesReq{
				Rules: []api.PolicyRuleRef{{Field: "f", SpecSelector: "s"}},
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("non-existent policy id", func() {
		nonExistentId := "00000000-0000-0000-0000-000000000000"

		It("returns 404 on get", func() {
			request, err := http.NewRequest(http.MethodGet, gvkPolicyURL(nonExistentId), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("returns 404 on put", func() {
			request, err := http.NewRequest(http.MethodPut, gvkPolicyURL(nonExistentId), makePayload(api.UpdateGvkPolicyReq{
				Rules: sampleRules(),
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("returns 404 on patch", func() {
			request, err := http.NewRequest(http.MethodPatch, gvkPolicyURL(nonExistentId), makePayload(api.UpdateGvkPolicyReq{
				Rules: sampleRules(),
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("returns 404 on delete", func() {
			request, err := http.NewRequest(http.MethodDelete, gvkPolicyURL(nonExistentId), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("returns 404 on delete rules", func() {
			request, err := http.NewRequest(http.MethodDelete, gvkPolicyRulesURL(nonExistentId), makePayload(api.DeleteGvkPolicyRulesReq{
				Rules: []api.PolicyRuleRef{{Field: "f", SpecSelector: "s"}},
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("update (put)", func() {
		It("overwrites all rules", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			newRules := []api.PolicyRule{
				{Field: "workingDir", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}

			request, err := http.NewRequest(http.MethodPut, gvkPolicyURL(created.Id), makePayload(api.UpdateGvkPolicyReq{
				Rules: newRules,
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

			fetched := getGvkPolicyForTest(router, tenantId, created.Id)
			Expect(fetched.Rules).To(HaveLen(1))
			Expect(fetched.Rules[0].Field).To(Equal("workingDir"))
			expectEffectiveRules(fetched, 1)
		})

		It("rejects put with empty rules", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			request, err := http.NewRequest(http.MethodPut, gvkPolicyURL(created.Id), makePayload(api.UpdateGvkPolicyReq{
				Rules: []api.PolicyRule{},
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("updates description", func() {
			req := newProjectPolicy(sampleRules())
			req.Description = convutils.Ref("original")
			created := createGvkPolicyForTest(router, tenantId, req)

			request, err := http.NewRequest(http.MethodPut, gvkPolicyURL(created.Id), makePayload(api.UpdateGvkPolicyReq{
				Description: convutils.Ref("updated"),
				Rules:       sampleRules(),
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

			fetched := getGvkPolicyForTest(router, tenantId, created.Id)
			Expect(*fetched.Description).To(Equal("updated"))
			expectEffectiveRules(fetched, 2)
		})
	})

	Context("patch", func() {
		It("merges incoming rules with existing (replaces matching, appends new)", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			patchRules := []api.PolicyRule{
				{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}}},
				{Field: "workingDir", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}

			request, err := http.NewRequest(http.MethodPatch, gvkPolicyURL(created.Id), makePayload(api.UpdateGvkPolicyReq{
				Rules: patchRules,
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

			fetched := getGvkPolicyForTest(router, tenantId, created.Id)
			Expect(fetched.Rules).To(HaveLen(3))
			Expect(fetched.Rules[0].Field).To(Equal("image"))
			Expect(*fetched.Rules[0].Enforcements[0].Required).To(BeFalse())
			Expect(fetched.Rules[2].Field).To(Equal("workingDir"))
			expectEffectiveRules(fetched, 3)
		})

		It("patch with empty rules does not modify existing rules", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			request, err := http.NewRequest(http.MethodPatch, gvkPolicyURL(created.Id), makePayload(api.UpdateGvkPolicyReq{
				Rules: []api.PolicyRule{},
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusOK))

			fetched := getGvkPolicyForTest(router, tenantId, created.Id)
			Expect(fetched.Rules).To(HaveLen(2))
		})
	})

	Context("delete", func() {
		It("deletes a GVK policy", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			request, err := http.NewRequest(http.MethodDelete, gvkPolicyURL(created.Id), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNoContent))

			getReq, _ := http.NewRequest(http.MethodGet, gvkPolicyURL(created.Id), nil)
			getReq.Header = *headerForAdmin(tenantId)
			getResp := httptest.NewRecorder()
			router.ServeHTTP(getResp, getReq.WithContext(tenantContextForTest(tenantId)))
			Expect(getResp.Code).To(Equal(http.StatusNotFound))
		})

		It("returns 404 when deleting non-existent policy", func() {
			request, err := http.NewRequest(http.MethodDelete, gvkPolicyURL("00000000-0000-0000-0000-000000000000"), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("delete rules", func() {
		It("removes specific rules by ref", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy([]api.PolicyRule{
				{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				{Field: "imagePullPolicy", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}}},
				{Field: "workingDir", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}))

			request, err := http.NewRequest(http.MethodDelete, gvkPolicyRulesURL(created.Id), makePayload(api.DeleteGvkPolicyRulesReq{
				Rules: []api.PolicyRuleRef{
					{Field: "imagePullPolicy", SpecSelector: "allContainers"},
				},
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

			fetched := getGvkPolicyForTest(router, tenantId, created.Id)
			Expect(fetched.Rules).To(HaveLen(2))
			Expect(fetched.Rules[0].Field).To(Equal("image"))
			Expect(fetched.Rules[1].Field).To(Equal("workingDir"))
			expectEffectiveRules(fetched, 2)
		})

		It("returns error when trying to delete a non-existent rule", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			request, err := http.NewRequest(http.MethodDelete, gvkPolicyRulesURL(created.Id), makePayload(api.DeleteGvkPolicyRulesReq{
				Rules: []api.PolicyRuleRef{
					{Field: "nonexistent", SpecSelector: "allContainers"},
				},
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("returns error when trying to remove all rules", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy([]api.PolicyRule{
				{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}))

			request, err := http.NewRequest(http.MethodDelete, gvkPolicyRulesURL(created.Id), makePayload(api.DeleteGvkPolicyRulesReq{
				Rules: []api.PolicyRuleRef{
					{Field: "image", SpecSelector: "allContainers"},
				},
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("removes rules with itemSelector", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy([]api.PolicyRule{
				{Field: "ports-Name", SpecSelector: "allContainers", ItemSelector: convutils.Ref("allPorts"), Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				{Field: "ports-Name", SpecSelector: "allContainers", ItemSelector: convutils.Ref("portByContainerPort"), SelectorParams: map[string]interface{}{"containerPort": 8080}, Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}}},
			}))

			request, err := http.NewRequest(http.MethodDelete, gvkPolicyRulesURL(created.Id), makePayload(api.DeleteGvkPolicyRulesReq{
				Rules: []api.PolicyRuleRef{
					{Field: "ports-Name", SpecSelector: "allContainers", ItemSelector: convutils.Ref("allPorts")},
				},
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

			fetched := getGvkPolicyForTest(router, tenantId, created.Id)
			Expect(fetched.Rules).To(HaveLen(1))
			Expect(*fetched.Rules[0].ItemSelector).To(Equal("portByContainerPort"))
			expectEffectiveRules(fetched, 1)
		})
	})

	Context("project scope without project id", func() {
		It("rejects create with project scope but no projectId", func() {
			req := api.CreateGvkPolicyReq{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				Rules:        sampleRules(),
			}
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		Context("stored policy with scope=project but no projectId", func() {
			var badPolicyId string

			BeforeEach(func() {
				pId := uuid.New()
				badPolicyId = pId.String()
				bad := dbmodel.Policy{
					BaseModel: dbmodels.BaseModel{
						ID:        pId,
						CreatedBy: "test",
						UpdatedBy: "test",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:         "bad-no-project-id",
					WorkloadType: "apps/Deployment/v1",
					Scope:        api.SCOPETYPE_PROJECT,
					TenantId:     tenantId,
					GvkWorkload:  true,
				}
				Expect(dao.AddPolicy(context.Background(), runaidb.DB(), &bad)).To(BeNil())
			})

			It("returns error on get", func() {
				request, err := http.NewRequest(http.MethodGet, gvkPolicyURL(badPolicyId), nil)
				Expect(err).To(BeNil())
				request.Header = *headerForAdmin(tenantId)

				response := httptest.NewRecorder()
				router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns error on put", func() {
				request, err := http.NewRequest(http.MethodPut, gvkPolicyURL(badPolicyId), makePayload(api.UpdateGvkPolicyReq{
					Rules: sampleRules(),
				}))
				Expect(err).To(BeNil())
				request.Header = *headerForAdmin(tenantId)

				response := httptest.NewRecorder()
				router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns error on patch", func() {
				request, err := http.NewRequest(http.MethodPatch, gvkPolicyURL(badPolicyId), makePayload(api.UpdateGvkPolicyReq{
					Rules: sampleRules(),
				}))
				Expect(err).To(BeNil())
				request.Header = *headerForAdmin(tenantId)

				response := httptest.NewRecorder()
				router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns error on delete rules", func() {
				request, err := http.NewRequest(http.MethodDelete, gvkPolicyRulesURL(badPolicyId), makePayload(api.DeleteGvkPolicyRulesReq{
					Rules: []api.PolicyRuleRef{{Field: "image", SpecSelector: "allContainers"}},
				}))
				Expect(err).To(BeNil())
				request.Header = *headerForAdmin(tenantId)

				response := httptest.NewRecorder()
				router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Context("validation issues persistence", func() {
		// These tests verify DAO-level persistence of status fields.
		// Note: getGvkPolicy always re-validates, so we test DB persistence directly.

		It("stores validation status and issues in the database", func() {
			ctx := tenantContextForTest(tenantId)
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))
			Expect(created.Status).NotTo(BeNil())
			Expect(created.Status.Phase).To(Equal(api.POLICYPHASE_READY), "newly created policy should be Ready")
			Expect(created.Status.Issues).To(BeEmpty(), "newly created policy should have no issues")

			// Update status via DAO
			degradedPhase := api.POLICYPHASE_DEGRADED
			status := &api.PolicyStatus{
				Phase:  degradedPhase,
				Issues: []string{"specSelector portByName is not defined", "field foo has no matching formatter"},
			}
			Expect(dao.UpdatePoliciesStatus(ctx, runaidb.DB().WithContext(ctx), []string{created.Id}, status)).To(BeNil())

			// Verify by reading directly from DB
			var stored dbmodel.Policy
			result := runaidb.DB().Where("id = ?", created.Id).First(&stored)
			Expect(result).To(BeNil())
			Expect(stored.ValidationStatus).NotTo(BeNil())
			Expect(*stored.ValidationStatus).To(Equal(api.POLICYPHASE_DEGRADED))
			Expect(stored.ValidationIssues).To(HaveLen(2))
			Expect(stored.ValidationIssues).To(ContainElement("specSelector portByName is not defined"))
			Expect(stored.ValidationIssues).To(ContainElement("field foo has no matching formatter"))
		})

		It("clears validation status and issues when set to nil", func() {
			ctx := tenantContextForTest(tenantId)
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			// First set some issues
			degradedPhase := api.POLICYPHASE_DEGRADED
			status := &api.PolicyStatus{
				Phase:  degradedPhase,
				Issues: []string{"temporary issue"},
			}
			Expect(dao.UpdatePoliciesStatus(ctx, runaidb.DB().WithContext(ctx), []string{created.Id}, status)).To(BeNil())

			// Verify issues are set
			var stored dbmodel.Policy
			result := runaidb.DB().Where("id = ?", created.Id).First(&stored)
			Expect(result).To(BeNil())
			Expect(stored.ValidationStatus).NotTo(BeNil())
			Expect(stored.ValidationIssues).To(HaveLen(1))

			// Clear status by setting to nil
			Expect(dao.UpdatePoliciesStatus(ctx, runaidb.DB().WithContext(ctx), []string{created.Id}, nil)).To(BeNil())

			// Verify status is cleared
			result = runaidb.DB().Where("id = ?", created.Id).First(&stored)
			Expect(result).To(BeNil())
			Expect(stored.ValidationStatus).To(BeNil())
			Expect(stored.ValidationIssues).To(BeEmpty())
		})
	})

	Context("no authentication", func() {
		It("rejects create without auth headers", func() {
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(newProjectPolicy(sampleRules())))
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})

		It("rejects get without auth headers", func() {
			request, err := http.NewRequest(http.MethodGet, gvkPolicyURL("00000000-0000-0000-0000-000000000000"), nil)
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})

		It("rejects list without auth headers", func() {
			request, err := http.NewRequest(http.MethodGet, urlGvkPolicies, nil)
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})

		It("rejects put without auth headers", func() {
			request, err := http.NewRequest(http.MethodPut, gvkPolicyURL("00000000-0000-0000-0000-000000000000"), makePayload(api.UpdateGvkPolicyReq{
				Rules: sampleRules(),
			}))
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})

		It("rejects delete without auth headers", func() {
			request, err := http.NewRequest(http.MethodDelete, gvkPolicyURL("00000000-0000-0000-0000-000000000000"), nil)
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})

		It("rejects delete rules without auth headers", func() {
			request, err := http.NewRequest(http.MethodDelete, gvkPolicyRulesURL("00000000-0000-0000-0000-000000000000"), makePayload(api.DeleteGvkPolicyRulesReq{
				Rules: []api.PolicyRuleRef{{Field: "f", SpecSelector: "s"}},
			}))
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("org unit not found", func() {
		nonExistentProjectId := "9999"

		It("returns not found on create for non-existent project", func() {
			req := api.CreateGvkPolicyReq{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(nonExistentProjectId),
				Rules:        sampleRules(),
			}
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("returns not found on update for policy with deleted org unit", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			backendMock.DeletedScopeEntities[fmt.Sprintf("project:%s", projectId)] = true

			request, err := http.NewRequest(http.MethodPut, gvkPolicyURL(created.Id), makePayload(api.UpdateGvkPolicyReq{
				Rules: sampleRules(),
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("returns not found on delete-rules for policy with deleted org unit", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			backendMock.DeletedScopeEntities[fmt.Sprintf("project:%s", projectId)] = true

			request, err := http.NewRequest(http.MethodDelete, gvkPolicyRulesURL(created.Id), makePayload(api.DeleteGvkPolicyRulesReq{
				Rules: []api.PolicyRuleRef{{Field: "image", SpecSelector: "allContainers"}},
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("invalid workload type", func() {
		invalidGvk := api.Gvk{Group: convutils.Ref("a"), Kind: "b", Version: "c"}

		It("rejects create with invalid workload type format", func() {
			req := api.CreateGvkPolicyReq{
				WorkloadType: api.Gvk{Group: convutils.Ref(""), Kind: "", Version: ""},
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules:        sampleRules(),
			}
			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("invalid GVK: version, kind are required"))
		})

		It("reject when mapping does not exist", func() {
			req := api.CreateGvkPolicyReq{
				WorkloadType: invalidGvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules:        sampleRules(),
			}

			request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusNotFound))
			Expect(response.Body.String()).To(ContainSubstring("mapping for a/b/c not found"))
		})

		It("round-trips a policy with arbitrary GVK through get", func() {
			seedMappingForGvk(invalidGvk)

			created := createGvkPolicyForTest(router, tenantId, api.CreateGvkPolicyReq{
				WorkloadType: invalidGvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules:        sampleRules(),
			})

			fetched := getGvkPolicyForTest(router, tenantId, created.Id)
			Expect(fetched.WorkloadType).To(Equal(invalidGvk))
		})

		It("lists policies with arbitrary GVK", func() {
			seedMappingForGvk(invalidGvk)

			createGvkPolicyForTest(router, tenantId, api.CreateGvkPolicyReq{
				WorkloadType: invalidGvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules:        sampleRules(),
			})

			list := listGvkPoliciesForTest(router, tenantId, "")
			Expect(list.Items).To(HaveLen(1))
			Expect(list.Items[0].WorkloadType).To(Equal(invalidGvk))
		})

		It("updates a policy with arbitrary GVK", func() {
			seedMappingForGvk(invalidGvk)

			created := createGvkPolicyForTest(router, tenantId, api.CreateGvkPolicyReq{
				WorkloadType: invalidGvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules:        sampleRules(),
			})

			newRules := []api.PolicyRule{
				{Field: "workingDir", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}
			request, err := http.NewRequest(http.MethodPut, gvkPolicyURL(created.Id), makePayload(api.UpdateGvkPolicyReq{
				Rules: newRules,
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

			fetched := getGvkPolicyForTest(router, tenantId, created.Id)
			Expect(fetched.Rules).To(HaveLen(1))
		})

		It("deletes rules from a policy with arbitrary GVK", func() {
			seedMappingForGvk(invalidGvk)

			created := createGvkPolicyForTest(router, tenantId, api.CreateGvkPolicyReq{
				WorkloadType: invalidGvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules: []api.PolicyRule{
					{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
					{Field: "imagePullPolicy", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}}},
				},
			})

			request, err := http.NewRequest(http.MethodDelete, gvkPolicyRulesURL(created.Id), makePayload(api.DeleteGvkPolicyRulesReq{
				Rules: []api.PolicyRuleRef{{Field: "imagePullPolicy", SpecSelector: "allContainers"}},
			}))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

			fetched := getGvkPolicyForTest(router, tenantId, created.Id)
			Expect(fetched.Rules).To(HaveLen(1))
		})
	})

	// -------------------------------------------------------------------------
	// Breakdown caching
	// -------------------------------------------------------------------------

	Context("breakdown caching", func() {
		ctx := tenantContextForTest(tenantId)

		It("caches the breakdown on first get and returns it from cache on second get", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			// first get – populates the cache
			resp1 := getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(resp1.Breakdown).NotTo(BeNil())
			Expect(resp1.Breakdown.Fields).NotTo(BeEmpty())

			// verify the cache row exists in the DB
			cached, dbErr := dao.GetCachedBreakdown(ctx, created.Id)
			Expect(dbErr).To(BeNil())
			Expect(cached).NotTo(BeNil(), "cache entry should exist after first get")

			// second get – should be served from cache (breakdown still populated)
			resp2 := getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(resp2.Breakdown).NotTo(BeNil())
			Expect(resp2.Breakdown.Fields).To(Equal(resp1.Breakdown.Fields))
		})

		It("re-caches breakdown when the policy is updated", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			// prime the cache
			getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(cacheExistsForPolicy(ctx, created.Id)).To(BeTrue())

			// update the policy
			updateReq, err := http.NewRequest(http.MethodPatch, gvkPolicyURL(created.Id), makePayload(api.UpdateGvkPolicyReq{
				Rules: []api.PolicyRule{
					{Field: "image", SpecSelector: "allContainers", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}}},
				},
			}))
			Expect(err).To(BeNil())
			updateReq.Header = *headerForAdmin(tenantId)
			updateResp := httptest.NewRecorder()
			router.ServeHTTP(updateResp, updateReq.WithContext(ctx))
			Expect(updateResp.Code).To(Equal(http.StatusOK), updateResp.Body.String())

			// After a successful update, the new breakdown should be cached (same as create)
			Expect(cacheExistsForPolicy(ctx, created.Id)).To(BeTrue(), "cache should exist after policy update with new breakdown")
			Expect(dependenciesExistForPolicy(ctx, created.Id)).To(BeTrue(), "dependencies should exist after policy update")
		})

		It("clears cache when the policy is deleted", func() {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			// prime the cache
			getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(cacheExistsForPolicy(ctx, created.Id)).To(BeTrue())

			// delete the policy
			delReq, err := http.NewRequest(http.MethodDelete, gvkPolicyURL(created.Id), nil)
			Expect(err).To(BeNil())
			delReq.Header = *headerForAdmin(tenantId)
			delResp := httptest.NewRecorder()
			router.ServeHTTP(delResp, delReq.WithContext(ctx))
			Expect(delResp.Code).To(Equal(http.StatusNoContent))

			Expect(cacheExistsForPolicy(ctx, created.Id)).To(BeFalse(), "cache should be cleared after policy delete")
			Expect(dependenciesExistForPolicy(ctx, created.Id)).To(BeFalse(), "dependencies should be cleared after policy delete")
		})

		It("clears caches of subtree descendants when a cluster policy is deleted", func() {
			// project "1111" belongs to cluster ClusterIdPerTenant["1"][0]
			// per the BackendMockClient ancestry encoding.
			clusterId := backendMock.ClusterIdPerTenant["1"][0]

			clusterPolicy := createGvkPolicyForTest(router, tenantId, api.CreateGvkPolicyReq{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_CLUSTER,
				ClusterId:    convutils.Ref(clusterId),
				Rules:        sampleRules(),
			})
			projectPolicy := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))

			// prime caches for both: each get populates the breakdown cache
			// row keyed by that policy's id.
			getGvkPolicyWithBreakdownForTest(router, tenantId, clusterPolicy.Id)
			getGvkPolicyWithBreakdownForTest(router, tenantId, projectPolicy.Id)
			Expect(cacheExistsForPolicy(ctx, clusterPolicy.Id)).To(BeTrue())
			Expect(cacheExistsForPolicy(ctx, projectPolicy.Id)).To(BeTrue())

			delReq, err := http.NewRequest(http.MethodDelete, gvkPolicyURL(clusterPolicy.Id), nil)
			Expect(err).To(BeNil())
			delReq.Header = *headerForAdmin(tenantId)
			delResp := httptest.NewRecorder()
			router.ServeHTTP(delResp, delReq.WithContext(ctx))
			Expect(delResp.Code).To(Equal(http.StatusNoContent))

			Expect(cacheExistsForPolicy(ctx, clusterPolicy.Id)).To(BeFalse(),
				"cluster policy's own cache should be cleared on delete")
			Expect(cacheExistsForPolicy(ctx, projectPolicy.Id)).To(BeFalse(),
				"descendant project's cache should be cleared when its ancestor is deleted (effective rules changed)")
			Expect(dependenciesExistForPolicy(ctx, projectPolicy.Id)).To(BeFalse(),
				"descendant project's cache dependencies should be cleared too")
		})
	})

	// -------------------------------------------------------------------------
	// Component-level cache invalidation
	// -------------------------------------------------------------------------

	Context("cache invalidation on component changes", func() {
		ctx := tenantContextForTest(tenantId)

		// seed a policy, prime its breakdown cache, and return the policy id
		seedPolicyWithCache := func() string {
			created := createGvkPolicyForTest(router, tenantId, newProjectPolicy(sampleRules()))
			getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(cacheExistsForPolicy(ctx, created.Id)).To(BeTrue())
			return created.Id
		}

		type invalidationCase struct {
			description string
			invalidate  func()
		}

		DescribeTable("invalidates cache when a component is changed",
			func(tc invalidationCase) {
				policyId := seedPolicyWithCache()
				tc.invalidate()
				Expect(cacheExistsForPolicy(ctx, policyId)).To(BeFalse(),
					"cache should be gone after: %s", tc.description)
				Expect(dependenciesExistForPolicy(ctx, policyId)).To(BeFalse(),
					"dependencies should be gone after: %s", tc.description)
			},

			Entry("mapping spec selector patched", invalidationCase{
				description: "patch mapping spec selector",
				invalidate: func() {
					// create a tenant-scoped mapping overlay first, then patch it.
					// (system mappings require tenant_id=0 which breaks the locking
					// mechanism used by the patch path; tenant mappings work normally.)
					createReq, err := http.NewRequest(http.MethodPost,
						urlMappings,
						makePayload(api.MappingCreUpdRequest{
							WorkloadType: gvk,
							SpecSelectors: map[string]api.SpecSelector{
								"tenantSel": {Query: `.spec.containers[]`},
							},
						}))
					Expect(err).To(BeNil())
					createReq.Header = *headerForAdmin(tenantId)
					createResp := httptest.NewRecorder()
					router.ServeHTTP(createResp, createReq.WithContext(ctx))
					Expect(createResp.Code).To(Equal(http.StatusOK), createResp.Body.String())

					patchReq, err := http.NewRequest(http.MethodPatch,
						urlMappings,
						makePayload(api.MappingCreUpdRequest{
							WorkloadType: gvk,
							SpecSelectors: map[string]api.SpecSelector{
								"tenantSel": {Query: `.spec.initContainers[]`},
							},
						}))
					Expect(err).To(BeNil())
					patchReq.Header = *headerForAdmin(tenantId)
					resp := httptest.NewRecorder()
					router.ServeHTTP(resp, patchReq.WithContext(ctx))
					Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())
				},
			}),

			Entry("field group patched", invalidationCase{
				description: "patch field group",
				invalidate: func() {
					// find one of the field group UUIDs that the cached breakdown
					// depends on and call OnComponentDeleted directly to simulate
					// what happens when that field group is updated via the API.
					var dep dbmodel.EntitiesDependency
					res := runaidb.DB().WithContext(ctx).
						Where(fmt.Sprintf("%s = ?", "component_type"),
							string(dbmodel.ComponentTypeFieldGroup)).
						First(&dep)
					Expect(res).To(BeNil(), "expected at least one field_group dependency")

					dbErr := dao.OnComponentDeleted(ctx,
						runaidb.DB().WithContext(ctx),
						dbmodel.ComponentTypeFieldGroup,
						dep.ComponentId)
					Expect(dbErr).To(BeNil())
				},
			}),
		)
	})

	// -------------------------------------------------------------------------
	// Policy Status Lifecycle
	// -------------------------------------------------------------------------

	Context("policy status lifecycle", func() {
		var ctx context.Context
		// Two different GVKs - one for the policy under test, one unrelated
		testGvk := api.Gvk{Group: convutils.Ref("test.example.com"), Kind: "StatusTest", Version: "v1"}
		unrelatedGvk := api.Gvk{Group: convutils.Ref("test.example.com"), Kind: "UnrelatedTest", Version: "v1"}
		fgName := "statusTestFieldGroup"

		BeforeEach(func() {
			ctx = tenantContextForTest(tenantId)

			// create a field group with test fields
			fgReq := api.FieldGroupCreationRequest{
				Name: fgName,
				Fields: map[string]api.FieldGroupField{
					"testField1": {Type: api.MAPPINGFIELDTYPE_STRING, Formatter: dbmodel.BuiltinFormatterKeyValue},
					"testField2": {Type: api.MAPPINGFIELDTYPE_STRING, Formatter: dbmodel.BuiltinFormatterKeyValue},
					"testField3": {Type: api.MAPPINGFIELDTYPE_STRING, Formatter: dbmodel.BuiltinFormatterKeyValue},
				},
			}
			fgReqHttp, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(fgReq))
			Expect(err).To(BeNil())
			fgReqHttp.Header = *headerForAdmin(tenantId)
			fgResp := httptest.NewRecorder()
			router.ServeHTTP(fgResp, fgReqHttp.WithContext(ctx))
			Expect(fgResp.Code).To(Equal(http.StatusOK), fgResp.Body.String())

			// create a tenant mapping with two specSelectors for the policy under test
			mappingReq := api.MappingCreUpdRequest{
				WorkloadType: testGvk,
				SpecSelectors: map[string]api.SpecSelector{
					"specSel1": {Query: `.spec.container1`, FieldGroups: []string{fgName}},
					"specSel2": {Query: `.spec.container2`, FieldGroups: []string{fgName}},
				},
			}
			mapReqHttp, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(mappingReq))
			Expect(err).To(BeNil())
			mapReqHttp.Header = *headerForAdmin(tenantId)
			mapResp := httptest.NewRecorder()
			router.ServeHTTP(mapResp, mapReqHttp.WithContext(ctx))
			Expect(mapResp.Code).To(Equal(http.StatusOK), mapResp.Body.String())

			// create a separate unrelated mapping with specSel3
			unrelatedMappingReq := api.MappingCreUpdRequest{
				WorkloadType: unrelatedGvk,
				SpecSelectors: map[string]api.SpecSelector{
					"specSel3": {Query: `.spec.container3`, FieldGroups: []string{fgName}},
				},
			}
			unrelatedMapReqHttp, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(unrelatedMappingReq))
			Expect(err).To(BeNil())
			unrelatedMapReqHttp.Header = *headerForAdmin(tenantId)
			unrelatedMapResp := httptest.NewRecorder()
			router.ServeHTTP(unrelatedMapResp, unrelatedMapReqHttp.WithContext(ctx))
			Expect(unrelatedMapResp.Code).To(Equal(http.StatusOK), unrelatedMapResp.Body.String())
		})

		It("follows the full status lifecycle when specSelectors are removed and rules updated", func() {
			// Step 1: Create policy with 2 specSelectors - status should be Ready
			createReq := api.CreateGvkPolicyReq{
				WorkloadType: testGvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules: []api.PolicyRule{
					{Field: "testField1", SpecSelector: "specSel1", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
					{Field: "testField2", SpecSelector: "specSel2", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				},
			}
			created := createGvkPolicyForTest(router, tenantId, createReq)
			Expect(created.Id).NotTo(BeEmpty())
			Expect(created.Status).NotTo(BeNil())
			Expect(created.Status.Phase).To(Equal(api.POLICYPHASE_READY), "newly created policy should be Ready")
			Expect(created.Status.Issues).To(BeEmpty(), "newly created policy should have no issues")

			// Prime the cache by requesting the policy with breakdown
			// This creates the cache row and dependency edges that OnComponentDeleted will use
			getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(cacheExistsForPolicy(ctx, created.Id)).To(BeTrue(), "cache should exist after get with breakdown")

			// Step 2: Remove specSel2 from the mapping - policy should change to PendingUpdate
			// then Degraded when accessed (since get always validates)
			delSelReq := api.MappingDeleteSpecSelectorsReq{
				WorkloadType:  testGvk,
				SpecSelectors: []string{"specSel2"},
			}
			delSelHttp, err := http.NewRequest(http.MethodDelete, urlMappingsSpecSelectors, makePayload(delSelReq))
			Expect(err).To(BeNil())
			delSelHttp.Header = *headerForAdmin(tenantId)
			delSelResp := httptest.NewRecorder()
			router.ServeHTTP(delSelResp, delSelHttp.WithContext(ctx))
			Expect(delSelResp.Code).To(Equal(http.StatusOK), delSelResp.Body.String())

			// Verify policy was marked PendingUpdate in the DB (before we access it via API)
			// We check the DB directly because accessing via API will trigger validation
			var policyRow dbmodel.Policy
			dbRes := runaidb.DB().WithContext(ctx).
				Table("policies").
				Where("id = ?", created.Id).
				First(&policyRow)
			Expect(dbRes).To(BeNil())
			Expect(policyRow.ValidationStatus).NotTo(BeNil())
			Expect(*policyRow.ValidationStatus).To(Equal(api.POLICYPHASE_PENDING_UPDATE),
				"policy should be PendingUpdate in DB after related specSelector is removed")
			Expect(policyRow.ValidationIssues).To(BeEmpty(),
				"PendingUpdate should have no issues yet")

			// Step 3: Attempt to update policy description - should fail due to missing specSelector
			updateReq := api.UpdateGvkPolicyReq{
				Description: convutils.Ref("updated description"),
				Rules: []api.PolicyRule{
					{Field: "testField1", SpecSelector: "specSel1", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
					{Field: "testField2", SpecSelector: "specSel2", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				},
			}
			updateHttp, err := http.NewRequest(http.MethodPut, gvkPolicyURL(created.Id), makePayload(updateReq))
			Expect(err).To(BeNil())
			updateHttp.Header = *headerForAdmin(tenantId)
			updateResp := httptest.NewRecorder()
			router.ServeHTTP(updateResp, updateHttp.WithContext(ctx))
			Expect(updateResp.Code).To(Equal(http.StatusBadRequest), "update should fail with missing specSelector")
			Expect(updateResp.Body.String()).To(ContainSubstring("specSel2"), "error should mention the missing specSelector")

			// Step 4: Get policy with breakdown - should see Degraded status with issues
			// This triggers validation which updates status from PendingUpdate to Degraded
			getPolicyResp := getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(getPolicyResp.Status).NotTo(BeNil())
			Expect(getPolicyResp.Status.Phase).To(Equal(api.POLICYPHASE_DEGRADED),
				"policy should be Degraded after breakdown validation fails")
			Expect(getPolicyResp.Status.Issues).NotTo(BeEmpty(),
				"Degraded policy should have issues reported")
			Expect(getPolicyResp.Status.Issues[0]).To(ContainSubstring("specSel2"),
				"issues should mention the missing specSelector")
			Expect(getPolicyResp.Breakdown).NotTo(BeNil(), "breakdown should still be returned")
			// breakdown should contain only the valid specSelector (specSel2 couldn't be resolved)
			Expect(getPolicyResp.Breakdown.SpecSelectors).To(HaveKey("specSel1"))

			// Step 4b: Get policy again - issues should still be detected (not cached away)
			// This verifies that breakdowns with issues are not cached
			getPolicyResp2 := getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(getPolicyResp2.Status).NotTo(BeNil())
			Expect(getPolicyResp2.Status.Phase).To(Equal(api.POLICYPHASE_DEGRADED),
				"policy should still be Degraded on second get")
			Expect(getPolicyResp2.Status.Issues).NotTo(BeEmpty(),
				"issues should still be reported on second get")
			Expect(getPolicyResp2.Status.Issues[0]).To(ContainSubstring("specSel2"),
				"issues should still mention the missing specSelector on second get")

			// Step 5: Update policy removing the problematic rule - should succeed and become Ready
			fixUpdateReq := api.UpdateGvkPolicyReq{
				Rules: []api.PolicyRule{
					{Field: "testField1", SpecSelector: "specSel1", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				},
			}
			fixUpdateHttp, err := http.NewRequest(http.MethodPut, gvkPolicyURL(created.Id), makePayload(fixUpdateReq))
			Expect(err).To(BeNil())
			fixUpdateHttp.Header = *headerForAdmin(tenantId)
			fixUpdateResp := httptest.NewRecorder()
			router.ServeHTTP(fixUpdateResp, fixUpdateHttp.WithContext(ctx))
			Expect(fixUpdateResp.Code).To(Equal(http.StatusOK), fixUpdateResp.Body.String())

			var fixedPolicy api.Policy
			Expect(json.Unmarshal(fixUpdateResp.Body.Bytes(), &fixedPolicy)).To(Succeed())
			Expect(fixedPolicy.Status).NotTo(BeNil())
			Expect(fixedPolicy.Status.Phase).To(Equal(api.POLICYPHASE_READY),
				"policy should be Ready after removing problematic rule")
			Expect(fixedPolicy.Status.Issues).To(BeEmpty(),
				"Ready policy should have no issues")
			Expect(fixedPolicy.Rules).To(HaveLen(1), "policy should have 1 rule remaining")

			// Verify breakdown is cached after update (update now caches breakdown automatically)
			Expect(cacheExistsForPolicy(ctx, created.Id)).To(BeTrue(),
				"breakdown cache should exist after successful update")

			// Step 6: Remove specSel3 from unrelated mapping - policy should remain Ready
			// This tests that changes to unrelated mappings don't affect our policy
			unrelatedGvk := api.Gvk{Group: convutils.Ref("test.example.com"), Kind: "UnrelatedTest", Version: "v1"}
			delSel3Req := api.MappingDeleteSpecSelectorsReq{
				WorkloadType:  unrelatedGvk,
				SpecSelectors: []string{"specSel3"},
			}
			delSel3Http, err := http.NewRequest(http.MethodDelete, urlMappingsSpecSelectors, makePayload(delSel3Req))
			Expect(err).To(BeNil())
			delSel3Http.Header = *headerForAdmin(tenantId)
			delSel3Resp := httptest.NewRecorder()
			router.ServeHTTP(delSel3Resp, delSel3Http.WithContext(ctx))
			Expect(delSel3Resp.Code).To(Equal(http.StatusOK), delSel3Resp.Body.String())

			// verify policy status remains Ready (unrelated mapping was modified)
			policyAfterDelSel3 := getGvkPolicyForTest(router, tenantId, created.Id)
			Expect(policyAfterDelSel3.Status).NotTo(BeNil())
			Expect(policyAfterDelSel3.Status.Phase).To(Equal(api.POLICYPHASE_READY),
				"policy should remain Ready when unrelated mapping is modified")
			Expect(policyAfterDelSel3.Status.Issues).To(BeEmpty(),
				"policy should have no issues")
		})

		It("detects issues on repeated gets and updates issues when mapping changes again", func() {
			// This test verifies:
			// 1. Issues are not cached away (multiple gets still show issues)
			// 2. New mapping changes produce new/different issues

			// Create policy with two rules
			createReq := api.CreateGvkPolicyReq{
				WorkloadType: testGvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules: []api.PolicyRule{
					{Field: "testField1", SpecSelector: "specSel1", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
					{Field: "testField2", SpecSelector: "specSel2", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				},
			}
			created := createGvkPolicyForTest(router, tenantId, createReq)
			Expect(created.Status.Phase).To(Equal(api.POLICYPHASE_READY))

			// Prime the cache
			getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(cacheExistsForPolicy(ctx, created.Id)).To(BeTrue())

			// Remove specSel1 from mapping
			delSel1Req := api.MappingDeleteSpecSelectorsReq{
				WorkloadType:  testGvk,
				SpecSelectors: []string{"specSel1"},
			}
			delSel1Http, err := http.NewRequest(http.MethodDelete, urlMappingsSpecSelectors, makePayload(delSel1Req))
			Expect(err).To(BeNil())
			delSel1Http.Header = *headerForAdmin(tenantId)
			delSel1Resp := httptest.NewRecorder()
			router.ServeHTTP(delSel1Resp, delSel1Http.WithContext(ctx))
			Expect(delSel1Resp.Code).To(Equal(http.StatusOK), delSel1Resp.Body.String())

			// Helper to check issues contain expected specSelector error
			expectIssueForSpecSelector := func(issues []string, specSelector string) {
				found := false
				expectedSubstring := fmt.Sprintf("specSelector %q is not declared in the mapping", specSelector)
				for _, issue := range issues {
					if strings.Contains(issue, expectedSubstring) {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(),
					"expected to find issue containing %q in issues: %v", expectedSubstring, issues)
			}

			// First get - should show specSel1 issue
			get1 := getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(get1.Status.Phase).To(Equal(api.POLICYPHASE_DEGRADED))
			Expect(get1.Status.Issues).NotTo(BeEmpty(), "should have at least one issue for missing specSel1")
			expectIssueForSpecSelector(get1.Status.Issues, "specSel1")

			// Second get - should still show specSel1 issue (not cached away)
			get2 := getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(get2.Status.Phase).To(Equal(api.POLICYPHASE_DEGRADED))
			Expect(get2.Status.Issues).NotTo(BeEmpty(), "issues should still be present on second get")
			expectIssueForSpecSelector(get2.Status.Issues, "specSel1")

			// Third get - should still show specSel1 issue
			get3 := getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(get3.Status.Phase).To(Equal(api.POLICYPHASE_DEGRADED))
			Expect(get3.Status.Issues).NotTo(BeEmpty(), "issues should still be present on third get")
			expectIssueForSpecSelector(get3.Status.Issues, "specSel1")

			// Now also remove specSel2 from mapping
			delSel2Req := api.MappingDeleteSpecSelectorsReq{
				WorkloadType:  testGvk,
				SpecSelectors: []string{"specSel2"},
			}
			delSel2Http, err := http.NewRequest(http.MethodDelete, urlMappingsSpecSelectors, makePayload(delSel2Req))
			Expect(err).To(BeNil())
			delSel2Http.Header = *headerForAdmin(tenantId)
			delSel2Resp := httptest.NewRecorder()
			router.ServeHTTP(delSel2Resp, delSel2Http.WithContext(ctx))
			Expect(delSel2Resp.Code).To(Equal(http.StatusOK), delSel2Resp.Body.String())

			// Get after second mapping change - should now show BOTH issues
			get4 := getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)
			Expect(get4.Status.Phase).To(Equal(api.POLICYPHASE_DEGRADED))
			Expect(len(get4.Status.Issues)).To(BeNumerically(">=", 2),
				"should have at least 2 issues for both missing specSelectors, got: %v", get4.Status.Issues)

			// Verify both specSelectors are mentioned in issues
			expectIssueForSpecSelector(get4.Status.Issues, "specSel1")
			expectIssueForSpecSelector(get4.Status.Issues, "specSel2")
		})
	})
})

func createGvkPolicyForTest(router chi.Router, tenantId int32, req api.CreateGvkPolicyReq) api.Policy {
	request, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(req))
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(tenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
	Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

	var result api.Policy
	Expect(json.Unmarshal(response.Body.Bytes(), &result)).To(Succeed())
	return result
}

func getGvkPolicyForTest(router chi.Router, tenantId int32, policyId string) api.Policy {
	request, err := http.NewRequest(http.MethodGet, gvkPolicyURL(policyId), nil)
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(tenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
	Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

	var result api.Policy
	Expect(json.Unmarshal(response.Body.Bytes(), &result)).To(Succeed())
	return result
}

func expectEffectiveRules(policy api.Policy, expectedCount int) {
	Expect(policy.EffectiveRules).To(HaveLen(expectedCount), "expected %d effective rules", expectedCount)
	for _, er := range policy.EffectiveRules {
		Expect(er.SourceOfRule).NotTo(BeNil(), "effectiveRule for field=%s should have sourceOfRule", er.Field)
		Expect(er.SourceOfRule.Scope).To(Equal(api.SCOPETYPE_PROJECT), "effectiveRule for field=%s should have scope=project", er.Field)
	}
}

func listGvkPoliciesForTest(router chi.Router, tenantId int32, queryParams string) api.ListGvkPoliciesResp {
	url := urlGvkPolicies
	if queryParams != "" {
		url += "?" + queryParams
	}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(tenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
	Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

	var result api.ListGvkPoliciesResp
	Expect(json.Unmarshal(response.Body.Bytes(), &result)).To(Succeed())
	return result
}

// cacheExistsForPolicy returns true if a breakdown cache row exists for policyId.
func cacheExistsForPolicy(ctx context.Context, policyId string) bool {
	cached, err := dao.GetCachedBreakdown(ctx, policyId)
	Expect(err).To(BeNil())
	return cached != nil
}

// dependenciesExistForPolicy returns true if any dependency rows exist for policyId's cache.
func dependenciesExistForPolicy(ctx context.Context, policyId string) bool {
	// look up the cache row directly so we can use its id to find deps
	var cacheRow dbmodel.EntitiesCache
	res := runaidb.DB().WithContext(ctx).
		Where(&dbmodel.EntitiesCache{
			EntityType: dbmodel.EntityTypeBreakdown,
			EntityId:   policyId,
		}).
		First(&cacheRow)
	if res != nil {
		// no cache row → no deps either
		return false
	}
	var count int64
	depRes := runaidb.DB().WithContext(ctx).
		Model(&dbmodel.EntitiesDependency{}).
		Where("cache_id = ?", cacheRow.ID).
		Count(&count)
	Expect(depRes).To(BeNil())
	return count > 0
}

func seedMappingForGvk(gvk api.Gvk) {
	sysCtx := auth.MockAuthContext(systemMappingActor, 0)

	// Skip if the mapping already exists (e.g. seeded by handleSystemMappings).
	if _, err := dao.GetMappingByGvkAndScope(sysCtx, gvk, api.MAPPINGSCOPE_SYSTEM); err == nil {
		return
	} else if err.Code != http.StatusNotFound {
		Expect(err).To(BeNil(), "unexpected error while checking existing system mapping")
	}

	mapping := &dbmodel.MappingSelectors{
		ID:         uuid.New(),
		GvkGroup:   convutils.Deref(gvk.Group, ""),
		GvkKind:    gvk.Kind,
		GvkVersion: gvk.Version,
		ScopeType:  api.MAPPINGSCOPE_SYSTEM,
		TenantId:   0,
		SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{
			"allContainers": {
				Query:       `.spec.template.spec.containers[]`,
				FieldGroups: []string{sm.SystemPodSpecContainerFgName},
				Scope:       api.MAPPINGSCOPE_SYSTEM,
			},
		},
		CreatedBy: systemMappingActor,
		CreatedAt: time.Now(),
		UpdatedBy: systemMappingActor,
		UpdatedAt: time.Now(),
	}
	Expect(dao.SaveMapping(sysCtx, runaidb.DB().WithContext(sysCtx), mapping)).To(BeNil())
}
