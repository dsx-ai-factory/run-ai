/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	sm "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/system_mapping"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	urlMappings              = "/api/v1/policies/mappings"
	urlMappingsSpecSelectors = "/api/v1/policies/mappings/specSelectors"
)

func mappingListFindGvk(list api.MappingListResponse, gvk api.Gvk) (api.Mapping, bool) {
	for i := range list.Items {
		m := list.Items[i]
		if convutils.Deref(m.WorkloadType.Group, "") == convutils.Deref(gvk.Group, "") && m.WorkloadType.Kind == gvk.Kind && m.WorkloadType.Version == gvk.Version {
			return m, true
		}
	}
	return api.Mapping{}, false
}

// Covers SpecMappingAPIServicer workload mapping handlers: list, create, patch, delete spec selectors, delete mapping.
var _ = Describe("Service mapping spec selectors (mappings API)", Ordered, func() {
	router := CreateRouter()

	var backendMock *BackendMockClient
	var clusterMock cluster.SimpleMockClient

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()

		backendMock = NewBackendMockClient(2, 2)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))
		authorization.SetMock(NewPoliciesAuthClient(backendMock))

		clusterMock = cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		cluster.SetMock(clusterMock)
	})

	const tenantA int32 = 1

	Context("authentication", func() {
		It("rejects create mapping without auth headers", func() {
			body := api.MappingCreUpdRequest{
				WorkloadType: api.Gvk{Group: convutils.Ref("x"), Kind: "Y", Version: "v1"},
			}
			request, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(body))
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("mapping APIs", func() {
		It("lists, creates, patches, deletes spec selectors, then deletes the mapping", func() {
			fgName := "tenantFgForMappingHttp"
			fgDesc := "fg for mapping http test"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"mapFld": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "specPath",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.http.test"), Kind: "TestWorkload", Version: "v1"}
			mapDesc := "tenant mapping http crud"

			reqList0, err := http.NewRequest(http.MethodGet, urlMappings, nil)
			Expect(err).To(BeNil())
			reqList0.Header = *headerForAdmin(tenantA)
			respList0 := httptest.NewRecorder()
			router.ServeHTTP(respList0, reqList0.WithContext(tenantContextForTest(tenantA)))
			Expect(respList0.Code).To(Equal(http.StatusOK), respList0.Body.String())
			var list0 api.MappingListResponse
			Expect(json.Unmarshal(respList0.Body.Bytes(), &list0)).To(Succeed())
			_, found0 := mappingListFindGvk(list0, gvk)
			Expect(found0).To(BeFalse())

			createBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &mapDesc,
				SpecSelectors: map[string]api.SpecSelector{
					"selMain": {
						Query:       ".spec.template",
						FieldGroups: []string{fgName},
					},
				},
			}
			reqPost, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantA)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantA)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())

			var created api.Mapping
			Expect(json.Unmarshal(respPost.Body.Bytes(), &created)).To(Succeed())
			Expect(created.WorkloadType).To(Equal(gvk))
			Expect(created.Description).ToNot(BeNil())
			Expect(*created.Description).To(Equal(mapDesc))
			Expect(created.Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(created.SpecSelectors).To(HaveKey("selMain"))
			Expect(created.SpecSelectors["selMain"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(created.SpecSelectors["selMain"].FieldGroups).To(Equal([]string{fgName}))

			reqList1, err := http.NewRequest(http.MethodGet, urlMappings, nil)
			Expect(err).To(BeNil())
			reqList1.Header = *headerForAdmin(tenantA)
			respList1 := httptest.NewRecorder()
			router.ServeHTTP(respList1, reqList1.WithContext(tenantContextForTest(tenantA)))
			Expect(respList1.Code).To(Equal(http.StatusOK), respList1.Body.String())
			var list1 api.MappingListResponse
			Expect(json.Unmarshal(respList1.Body.Bytes(), &list1)).To(Succeed())
			listed, found1 := mappingListFindGvk(list1, gvk)
			Expect(found1).To(BeTrue())
			Expect(listed.SpecSelectors).To(HaveKey("selMain"))
			Expect(listed.SpecSelectors["selMain"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))

			descPatched := "patched description"
			patchBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &descPatched,
				SpecSelectors: map[string]api.SpecSelector{
					"selExtra": {
						Query:       ".metadata.labels",
						FieldGroups: []string{fgName},
					},
				},
			}
			reqPatch, err := http.NewRequest(http.MethodPatch, urlMappings, makePayload(patchBody))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantA)
			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantA)))
			Expect(respPatch.Code).To(Equal(http.StatusOK), respPatch.Body.String())
			var patched api.Mapping
			Expect(json.Unmarshal(respPatch.Body.Bytes(), &patched)).To(Succeed())
			Expect(patched.Description).ToNot(BeNil())
			Expect(*patched.Description).To(Equal(descPatched))
			Expect(patched.SpecSelectors).To(HaveKey("selMain"))
			Expect(patched.SpecSelectors).To(HaveKey("selExtra"))
			Expect(patched.SpecSelectors["selExtra"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))

			reqDelSels, err := http.NewRequest(http.MethodDelete, urlMappingsSpecSelectors, makePayload(api.MappingDeleteSpecSelectorsReq{
				WorkloadType:  gvk,
				SpecSelectors: []string{"selExtra"},
			}))
			Expect(err).To(BeNil())
			reqDelSels.Header = *headerForAdmin(tenantA)
			respDelSels := httptest.NewRecorder()
			router.ServeHTTP(respDelSels, reqDelSels.WithContext(tenantContextForTest(tenantA)))
			Expect(respDelSels.Code).To(Equal(http.StatusOK), respDelSels.Body.String())
			var afterDelSels api.Mapping
			Expect(json.Unmarshal(respDelSels.Body.Bytes(), &afterDelSels)).To(Succeed())
			Expect(afterDelSels.SpecSelectors).To(HaveKey("selMain"))
			Expect(afterDelSels.SpecSelectors).ToNot(HaveKey("selExtra"))

			reqDelMap, err := http.NewRequest(http.MethodDelete, urlMappings, makePayload(api.MappingDeletionRequest{
				WorkloadType: gvk,
			}))
			Expect(err).To(BeNil())
			reqDelMap.Header = *headerForAdmin(tenantA)
			respDelMap := httptest.NewRecorder()
			router.ServeHTTP(respDelMap, reqDelMap.WithContext(tenantContextForTest(tenantA)))
			Expect(respDelMap.Code).To(Equal(http.StatusNoContent), respDelMap.Body.String())

			reqList2, err := http.NewRequest(http.MethodGet, urlMappings, nil)
			Expect(err).To(BeNil())
			reqList2.Header = *headerForAdmin(tenantA)
			respList2 := httptest.NewRecorder()
			router.ServeHTTP(respList2, reqList2.WithContext(tenantContextForTest(tenantA)))
			Expect(respList2.Code).To(Equal(http.StatusOK), respList2.Body.String())
			var list2 api.MappingListResponse
			Expect(json.Unmarshal(respList2.Body.Bytes(), &list2)).To(Succeed())
			_, found2 := mappingListFindGvk(list2, gvk)
			Expect(found2).To(BeFalse())
		})

		It("list mappings respects filterBy on group, kind, and version", func() {
			fgName := "tenantFgForMappingListFilter"
			fgDesc := "fg for mapping list filter test"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"fld": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "specPath",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			postMapping := func(gvk api.Gvk, desc string) {
				body := api.MappingCreUpdRequest{
					WorkloadType: gvk,
					Description:  &desc,
					SpecSelectors: map[string]api.SpecSelector{
						"s": {Query: ".", FieldGroups: []string{fgName}},
					},
				}
				req, e := http.NewRequest(http.MethodPost, urlMappings, makePayload(body))
				Expect(e).To(BeNil())
				req.Header = *headerForAdmin(tenantA)
				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantA)))
				Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())
			}

			gvkA := api.Gvk{Group: convutils.Ref("mapping.list.filter"), Kind: "WorkloadA", Version: "v1"}
			gvkB := api.Gvk{Group: convutils.Ref("mapping.list.filter"), Kind: "WorkloadB", Version: "v1"}
			gvkC := api.Gvk{Group: convutils.Ref("other.group"), Kind: "WorkloadA", Version: "v2"}

			postMapping(gvkA, "ma")
			postMapping(gvkB, "mb")
			postMapping(gvkC, "mc")

			getList := func(filterBy string) api.MappingListResponse {
				u, e := url.Parse(urlMappings)
				Expect(e).To(BeNil())
				if filterBy != "" {
					q := url.Values{}
					q.Set("filterBy", filterBy)
					u.RawQuery = q.Encode()
				}
				req, e := http.NewRequest(http.MethodGet, u.String(), nil)
				Expect(e).To(BeNil())
				req.Header = *headerForAdmin(tenantA)
				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantA)))
				Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())
				var list api.MappingListResponse
				Expect(json.Unmarshal(resp.Body.Bytes(), &list)).To(Succeed())
				return list
			}

			full := getList("")
			// The list returns tenant + system mappings; the suite-wide seed creates
			// real system mappings so the total count is >3. Instead of checking exact
			// count, verify the 3 tenant mappings we created are present.
			_, okA := mappingListFindGvk(full, gvkA)
			_, okB := mappingListFindGvk(full, gvkB)
			_, okC := mappingListFindGvk(full, gvkC)
			Expect(okA && okB && okC).To(BeTrue())

			onlyB := getList("kind==WorkloadB")
			Expect(onlyB.Items).To(HaveLen(1))
			m, found := mappingListFindGvk(onlyB, gvkB)
			Expect(found).To(BeTrue())
			Expect(m.WorkloadType).To(Equal(gvkB))

			onlyCByGroup := getList("group==other.group")
			Expect(onlyCByGroup.Items).To(HaveLen(1))
			m, found = mappingListFindGvk(onlyCByGroup, gvkC)
			Expect(found).To(BeTrue())

			onlyCByVersion := getList("version==v2")
			Expect(onlyCByVersion.Items).To(HaveLen(1))
			m, found = mappingListFindGvk(onlyCByVersion, gvkC)
			Expect(found).To(BeTrue())

			twoByKind := getList("kind==WorkloadA")
			Expect(twoByKind.Items).To(HaveLen(2))
			_, okA = mappingListFindGvk(twoByKind, gvkA)
			_, okC = mappingListFindGvk(twoByKind, gvkC)
			Expect(okA && okC).To(BeTrue())

			combined := getList("group==mapping.list.filter,kind==WorkloadA")
			Expect(combined.Items).To(HaveLen(1))
			m, found = mappingListFindGvk(combined, gvkA)
			Expect(found).To(BeTrue())

			badU, err := url.Parse(urlMappings)
			Expect(err).To(BeNil())
			badU.RawQuery = url.Values{"filterBy": []string{"notAField==x"}}.Encode()
			reqBad, err := http.NewRequest(http.MethodGet, badU.String(), nil)
			Expect(err).To(BeNil())
			reqBad.Header = *headerForAdmin(tenantA)
			respBad := httptest.NewRecorder()
			router.ServeHTTP(respBad, reqBad.WithContext(tenantContextForTest(tenantA)))
			Expect(respBad.Code).To(Equal(http.StatusBadRequest), respBad.Body.String())
		})

		It("returns 409 when creating the same tenant mapping twice", func() {
			fgName := "tenantFgForMappingDup"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"a": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "s",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.dup.test"), Kind: "Dup", Version: "v1"}
			desc := "d"
			body := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &desc,
				SpecSelectors: map[string]api.SpecSelector{
					"s": {Query: ".", FieldGroups: []string{fgName}},
				},
			}
			req1, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(body))
			Expect(err).To(BeNil())
			req1.Header = *headerForAdmin(tenantA)
			resp1 := httptest.NewRecorder()
			router.ServeHTTP(resp1, req1.WithContext(tenantContextForTest(tenantA)))
			Expect(resp1.Code).To(Equal(http.StatusOK), resp1.Body.String())

			req2, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(body))
			Expect(err).To(BeNil())
			req2.Header = *headerForAdmin(tenantA)
			resp2 := httptest.NewRecorder()
			router.ServeHTTP(resp2, req2.WithContext(tenantContextForTest(tenantA)))
			Expect(resp2.Code).To(Equal(http.StatusConflict), resp2.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(resp2.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring("already exists"))
		})

		It("returns 400 when spec selector references an unknown field group", func() {
			gvk := api.Gvk{Group: convutils.Ref("mapping.badfg.test"), Kind: "Bad", Version: "v1"}
			desc := "bad"
			body := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &desc,
				SpecSelectors: map[string]api.SpecSelector{
					"sel": {Query: ".", FieldGroups: []string{"fieldGroupDoesNotExist789"}},
				},
			}
			req, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(body))
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantA)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantA)))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(resp.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring("field group"))
			Expect(apiErr.Message).To(ContainSubstring("does not exist"))
		})

		It("returns 400 when POST create uses an invalid specSelector map key (validateSpecSelectorMapKeys)", func() {
			fgName := "tenantFgBadKeyCreate"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"a": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "s",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.badkey.create"), Kind: "K", Version: "v1"}
			desc := "d"
			body := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &desc,
				SpecSelectors: map[string]api.SpecSelector{
					"bad!name": {Query: ".", FieldGroups: []string{fgName}},
				},
			}
			req, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(body))
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantA)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantA)))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(resp.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring("specSelector"))
			Expect(apiErr.Message).To(ContainSubstring("bad!name"))
		})

		It("returns 400 when PATCH adds a spec selector with an invalid map key (validateSpecSelectorMapKeys)", func() {
			fgName := "tenantFgForBadPatchKey"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"a": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "s",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.badkey.patch"), Kind: "K", Version: "v1"}
			mapDesc := "m"
			createBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &mapDesc,
				SpecSelectors: map[string]api.SpecSelector{
					"okSel": {Query: ".", FieldGroups: []string{fgName}},
				},
			}
			reqPost, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantA)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantA)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())

			patchBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				SpecSelectors: map[string]api.SpecSelector{
					"-bad": {Query: ".x", FieldGroups: []string{fgName}},
				},
			}
			reqPatch, err := http.NewRequest(http.MethodPatch, urlMappings, makePayload(patchBody))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantA)
			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantA)))
			Expect(respPatch.Code).To(Equal(http.StatusBadRequest), respPatch.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(respPatch.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring("specSelector"))
			Expect(apiErr.Message).To(ContainSubstring("-bad"))
		})

		It("returns 400 when POST create has invalid JQ in specSelector query (validateSpecSelectorJQExprs)", func() {
			fgName := "tenantFgBadJQQueryCreate"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"a": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "s",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.badjq.create"), Kind: "K", Version: "v1"}
			desc := "d"
			body := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &desc,
				SpecSelectors: map[string]api.SpecSelector{
					"badJQ": {Query: "abc(", FieldGroups: []string{fgName}},
				},
			}
			req, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(body))
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantA)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantA)))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(resp.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring("specSelector"))
			Expect(apiErr.Message).To(ContainSubstring("badJQ"))
			Expect(apiErr.Message).To(ContainSubstring("query"))
			Expect(apiErr.Message).To(ContainSubstring("invalid"))
		})

		It("returns 400 when PATCH adds a spec selector with invalid JQ in query (validateSpecSelectorJQExprs)", func() {
			fgName := "tenantFgBadJQPatchQuery"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"a": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "s",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.badjq.patch"), Kind: "K", Version: "v1"}
			mapDesc := "m"
			createBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &mapDesc,
				SpecSelectors: map[string]api.SpecSelector{
					"okSel": {Query: ".", FieldGroups: []string{fgName}},
				},
			}
			reqPost, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantA)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantA)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())

			patchBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				SpecSelectors: map[string]api.SpecSelector{
					"badSel": {Query: "abc(", FieldGroups: []string{fgName}},
				},
			}
			reqPatch, err := http.NewRequest(http.MethodPatch, urlMappings, makePayload(patchBody))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantA)
			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantA)))
			Expect(respPatch.Code).To(Equal(http.StatusBadRequest), respPatch.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(respPatch.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring("specSelector"))
			Expect(apiErr.Message).To(ContainSubstring("badSel"))
			Expect(apiErr.Message).To(ContainSubstring("query"))
			Expect(apiErr.Message).To(ContainSubstring("invalid"))
		})

		It("returns 400 when POST create has an argsInfo arg not referenced in the query (validateSpecSelectorJQExprs)", func() {
			fgName := "tenantFgArgsInfoUnreferenced"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"a": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "s",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.argsinfo.unref.create"), Kind: "K", Version: "v1"}
			desc := "d"
			body := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &desc,
				SpecSelectors: map[string]api.SpecSelector{
					"argSel": {
						Query: ".spec",
						ArgsInfo: map[string]api.MappingFieldType{
							"missingArg": api.MAPPINGFIELDTYPE_STRING,
						},
						FieldGroups: []string{fgName},
					},
				},
			}
			req, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(body))
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantA)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantA)))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(resp.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring("specSelector"))
			Expect(apiErr.Message).To(ContainSubstring("argSel"))
			Expect(apiErr.Message).To(ContainSubstring("missingArg"))
			Expect(apiErr.Message).To(ContainSubstring("not referenced in the query"))
		})

		It("returns 400 when POST create has an argsEnum arg that is neither in argsInfo nor referenced in the query (validateSpecSelectorJQExprs)", func() {
			fgName := "tenantFgArgsEnumUnref"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"a": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "s",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.argsenum.unref.create"), Kind: "K", Version: "v1"}
			desc := "d"
			body := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &desc,
				SpecSelectors: map[string]api.SpecSelector{
					"enumSel": {
						Query: ".spec",
						ArgsEnum: map[string][]string{
							"orphanEnum": {"A", "B"},
						},
						FieldGroups: []string{fgName},
					},
				},
			}
			req, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(body))
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantA)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantA)))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(resp.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring("specSelector"))
			Expect(apiErr.Message).To(ContainSubstring("enumSel"))
			Expect(apiErr.Message).To(ContainSubstring("enumeration is not allowed for arg"))
			Expect(apiErr.Message).To(ContainSubstring("orphanEnum"))
		})

		It("accepts POST create when argsInfo is referenced in the query and argsEnum is on a string arg", func() {
			fgName := "tenantFgArgsHappy"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"a": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "s",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.argshappy.create"), Kind: "K", Version: "v1"}
			desc := "d"
			body := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &desc,
				SpecSelectors: map[string]api.SpecSelector{
					"happySel": {
						Query: `.spec | select(.name == $ARGS.named.containerName)`,
						ArgsInfo: map[string]api.MappingFieldType{
							"containerName": api.MAPPINGFIELDTYPE_STRING,
						},
						ArgsEnum: map[string][]string{
							"containerName": {"main", "sidecar"},
						},
						FieldGroups: []string{fgName},
					},
				},
			}
			req, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(body))
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantA)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantA)))
			Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())
			var created api.Mapping
			Expect(json.Unmarshal(resp.Body.Bytes(), &created)).To(Succeed())
			Expect(created.SpecSelectors).To(HaveKey("happySel"))
			Expect(created.SpecSelectors["happySel"].ArgsInfo).To(HaveKeyWithValue("containerName", api.MAPPINGFIELDTYPE_STRING))
			Expect(created.SpecSelectors["happySel"].ArgsEnum).To(HaveKey("containerName"))
		})

		It("PATCH initializes nil spec_selectors JSON (jsonb null) before merging new selectors", func() {
			fgName := "tenantFgForNilJsonb"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"a": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "s",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.nil.jsonb.patch"), Kind: "W", Version: "v1"}
			mapDesc := "has row"
			createBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &mapDesc,
				SpecSelectors: map[string]api.SpecSelector{
					"willClear": {Query: ".old", FieldGroups: []string{fgName}},
				},
			}
			reqPost, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantA)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantA)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())

			Expect(runaidb.DB().Exec(
				`UPDATE mapping_selectors SET spec_selectors = 'null'::jsonb WHERE gvk_group = ? AND gvk_kind = ? AND gvk_version = ? AND scope_type = ? AND tenant_id = ?`,
				convutils.Deref(gvk.Group, ""), gvk.Kind, gvk.Version, string(api.MAPPINGSCOPE_TENANT), tenantA,
			)).Error().To(BeNil())

			tenantCtx := mappingTestTenantCtx(tenantA)
			reloaded, derr := dao.GetMappingByGvkAndScope(tenantCtx, gvk, api.MAPPINGSCOPE_TENANT)
			Expect(derr).To(BeNil())
			Expect(reloaded.SpecSelectors).To(BeNil())

			patchBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				SpecSelectors: map[string]api.SpecSelector{
					"afterNil": {Query: ".new", FieldGroups: []string{fgName}},
				},
			}
			reqPatch, err := http.NewRequest(http.MethodPatch, urlMappings, makePayload(patchBody))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantA)
			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantA)))
			Expect(respPatch.Code).To(Equal(http.StatusOK), respPatch.Body.String())
			var patched api.Mapping
			Expect(json.Unmarshal(respPatch.Body.Bytes(), &patched)).To(Succeed())
			Expect(patched.SpecSelectors).To(HaveKey("afterNil"))
			Expect(patched.SpecSelectors["afterNil"].Query).To(Equal(".new"))
		})

		It("returns 404 when DELETE specSelectors targets tenant scope but only a system mapping exists for the GVK", func() {
			db := runaidb.DB()
			ms := MappingApiService{}
			sysCtx := auth.MockAuthContext(systemMappingActor, 0)
			cvSys := NewMappingConsistencyValidator(sysCtx, api.MAPPINGSCOPE_SYSTEM)

			handleSystemFormatters(sysCtx, ms, cvSys)
			handleSystemFieldGroups(sysCtx, db, ms, cvSys)

			gvk := api.Gvk{Group: convutils.Ref("mapping.sys.only.delsel"), Kind: "X", Version: "v1"}
			sysDesc := "system only"
			sysRow := &dbmodel.MappingSelectors{
				ID:          uuid.New(),
				GvkGroup:    convutils.Deref(gvk.Group, ""),
				GvkKind:     gvk.Kind,
				GvkVersion:  gvk.Version,
				Description: &sysDesc,
				ScopeType:   api.MAPPINGSCOPE_SYSTEM,
				TenantId:    0,
				SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{
					"sysSel": {
						Query:       ".s",
						FieldGroups: []string{sm.SystemPodSpecFgName},
						Scope:       api.MAPPINGSCOPE_SYSTEM,
					},
				},
				CreatedAt: time.Now().UTC(),
				CreatedBy: systemMappingActor,
				UpdatedAt: time.Now().UTC(),
				UpdatedBy: systemMappingActor,
			}
			Expect(dao.CreateMapping(sysCtx, sysRow)).To(BeNil())

			reqDel, err := http.NewRequest(http.MethodDelete, urlMappingsSpecSelectors, makePayload(api.MappingDeleteSpecSelectorsReq{
				WorkloadType:  gvk,
				SpecSelectors: []string{"sysSel"},
			}))
			Expect(err).To(BeNil())
			reqDel.Header = *headerForAdmin(tenantA)
			respDel := httptest.NewRecorder()
			router.ServeHTTP(respDel, reqDel.WithContext(tenantContextForTest(tenantA)))
			Expect(respDel.Code).To(Equal(http.StatusNotFound), respDel.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(respDel.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring("tenant scoped mapping"))
			Expect(apiErr.Message).To(ContainSubstring("not found"))
		})

		It("returns 404 when PATCH update is requested for a GVK with no tenant mapping row", func() {
			gvk := api.Gvk{Group: convutils.Ref("mapping.patch.missing"), Kind: "None", Version: "v1"}
			desc := "nope"
			patchBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &desc,
			}
			reqPatch, err := http.NewRequest(http.MethodPatch, urlMappings, makePayload(patchBody))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantA)
			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantA)))
			Expect(respPatch.Code).To(Equal(http.StatusNotFound), respPatch.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(respPatch.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring("tenant scoped mapping"))
			Expect(apiErr.Message).To(ContainSubstring("not found"))
		})

		It("creates a mapping with empty specSelectors then PATCH adds selectors", func() {
			fgName := "tenantFgForEmptyMapCreate"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"a": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "s",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.empty.then.patch"), Kind: "W", Version: "v1"}
			mapDesc := "starts empty"
			createBody := api.MappingCreUpdRequest{
				WorkloadType:  gvk,
				Description:   &mapDesc,
				SpecSelectors: map[string]api.SpecSelector{},
			}
			reqPost, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantA)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantA)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())
			var created api.Mapping
			Expect(json.Unmarshal(respPost.Body.Bytes(), &created)).To(Succeed())
			Expect(created.SpecSelectors).To(BeEmpty())

			patchBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				SpecSelectors: map[string]api.SpecSelector{
					"lateSel": {Query: ".spec", FieldGroups: []string{fgName}},
				},
			}
			reqPatch, err := http.NewRequest(http.MethodPatch, urlMappings, makePayload(patchBody))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantA)
			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantA)))
			Expect(respPatch.Code).To(Equal(http.StatusOK), respPatch.Body.String())
			var patched api.Mapping
			Expect(json.Unmarshal(respPatch.Body.Bytes(), &patched)).To(Succeed())
			Expect(patched.SpecSelectors).To(HaveKey("lateSel"))
			Expect(patched.SpecSelectors["lateSel"].Query).To(Equal(".spec"))
		})

		It("returns 400 when PATCH violates mapping consistency (duplicate field across field groups)", func() {
			sharedField := "sharedFieldName"
			fgA := "tenantFgConsistA"
			fgB := "tenantFgConsistB"
			for _, pair := range []struct {
				name string
				fg   string
			}{{fgA, "a"}, {fgB, "b"}} {
				desc := "d"
				postFg := api.FieldGroupCreationRequest{
					Name:        pair.name,
					Description: &desc,
					Fields: map[string]api.FieldGroupField{
						sharedField: {
							Formatter: dbmodel.BuiltinFormatterKeyValue,
							SpecRef:   pair.fg,
							Type:      api.MAPPINGFIELDTYPE_STRING,
						},
					},
				}
				reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
				Expect(err).To(BeNil())
				reqFg.Header = *headerForAdmin(tenantA)
				respFg := httptest.NewRecorder()
				router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
				Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())
			}

			gvk := api.Gvk{Group: convutils.Ref("mapping.consist.patch"), Kind: "C", Version: "v1"}
			mapDesc := "m"
			createBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &mapDesc,
				SpecSelectors: map[string]api.SpecSelector{
					"okSel": {Query: ".", FieldGroups: []string{fgA}},
				},
			}
			reqPost, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantA)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantA)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())

			patchBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				SpecSelectors: map[string]api.SpecSelector{
					"badSel": {Query: ".spec", FieldGroups: []string{fgA, fgB}},
				},
			}
			reqPatch, err := http.NewRequest(http.MethodPatch, urlMappings, makePayload(patchBody))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantA)
			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantA)))
			Expect(respPatch.Code).To(Equal(http.StatusBadRequest), respPatch.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(respPatch.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring(sharedField))
			Expect(apiErr.Message).To(ContainSubstring("badSel"))
		})

		It("returns 400 when DELETE specSelectors lists an invalid selector key", func() {
			fgName := "tenantFgForDelSelNames"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"x": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "p",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.delsel.names"), Kind: "D", Version: "v1"}
			desc := "d"
			createBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &desc,
				SpecSelectors: map[string]api.SpecSelector{
					"good": {Query: ".", FieldGroups: []string{fgName}},
				},
			}
			reqPost, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantA)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantA)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())

			reqDel, err := http.NewRequest(http.MethodDelete, urlMappingsSpecSelectors, makePayload(api.MappingDeleteSpecSelectorsReq{
				WorkloadType:  gvk,
				SpecSelectors: []string{"bad!key"},
			}))
			Expect(err).To(BeNil())
			reqDel.Header = *headerForAdmin(tenantA)
			respDel := httptest.NewRecorder()
			router.ServeHTTP(respDel, reqDel.WithContext(tenantContextForTest(tenantA)))
			Expect(respDel.Code).To(Equal(http.StatusBadRequest), respDel.Body.String())
			var apiErr api.Error
			Expect(json.Unmarshal(respDel.Body.Bytes(), &apiErr)).To(Succeed())
			Expect(apiErr.Message).To(ContainSubstring("specSelector"))
			Expect(apiErr.Message).To(ContainSubstring("bad!key"))
		})
	})

	Context("mapping list merge (system + tenant GVK)", func() {
		It("returns merged spec selectors with tenant keys and values overriding system for the same key", func() {
			db := runaidb.DB()
			ms := MappingApiService{}
			sysCtx := auth.MockAuthContext(systemMappingActor, 0)
			cvSys := NewMappingConsistencyValidator(sysCtx, api.MAPPINGSCOPE_SYSTEM)

			handleSystemFormatters(sysCtx, ms, cvSys)
			handleSystemFieldGroups(sysCtx, db, ms, cvSys)

			gvk := api.Gvk{Group: convutils.Ref("mapping.merge.list"), Kind: "Merged", Version: "v1"}
			sysDesc := "system row"
			sysRow := &dbmodel.MappingSelectors{
				ID:            uuid.New(),
				GvkGroup:      convutils.Deref(gvk.Group, ""),
				GvkKind:       gvk.Kind,
				GvkVersion:    gvk.Version,
				Description:   &sysDesc,
				ScopeType:     api.MAPPINGSCOPE_SYSTEM,
				TenantId:      0,
				SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{},
				CreatedAt:     time.Now().UTC(),
				CreatedBy:     systemMappingActor,
				UpdatedAt:     time.Now().UTC(),
				UpdatedBy:     systemMappingActor,
			}
			sysRow.SpecSelectors["onlySystem"] = api.SpecSelectorsWithScopeValue{
				Query:       "fromSystemOnly",
				FieldGroups: []string{sm.SystemPodSpecFgName},
				Scope:       api.MAPPINGSCOPE_SYSTEM,
			}
			sysRow.SpecSelectors["bothKeys"] = api.SpecSelectorsWithScopeValue{
				Query:       "queryFromSystem",
				FieldGroups: []string{sm.SystemPodSpecFgName},
				Scope:       api.MAPPINGSCOPE_SYSTEM,
			}
			Expect(dao.CreateMapping(sysCtx, sysRow)).To(BeNil())

			fgName := "tenantFgMergeList"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"t": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "p",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			tenDesc := "tenant row"
			createBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &tenDesc,
				SpecSelectors: map[string]api.SpecSelector{
					"onlyTenant": {Query: "fromTenantOnly", FieldGroups: []string{fgName}},
					"bothKeys":   {Query: "queryFromTenant", FieldGroups: []string{fgName}},
				},
			}
			reqPost, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantA)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantA)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())

			reqList, err := http.NewRequest(http.MethodGet, urlMappings, nil)
			Expect(err).To(BeNil())
			reqList.Header = *headerForAdmin(tenantA)
			respList := httptest.NewRecorder()
			router.ServeHTTP(respList, reqList.WithContext(tenantContextForTest(tenantA)))
			Expect(respList.Code).To(Equal(http.StatusOK), respList.Body.String())
			var list api.MappingListResponse
			Expect(json.Unmarshal(respList.Body.Bytes(), &list)).To(Succeed())
			merged, found := mappingListFindGvk(list, gvk)
			Expect(found).To(BeTrue())
			Expect(merged.SpecSelectors).To(HaveKey("onlySystem"))
			Expect(merged.SpecSelectors["onlySystem"].Query).To(Equal("fromSystemOnly"))
			Expect(merged.SpecSelectors["onlySystem"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(merged.SpecSelectors).To(HaveKey("onlyTenant"))
			Expect(merged.SpecSelectors["onlyTenant"].Query).To(Equal("fromTenantOnly"))
			Expect(merged.SpecSelectors["onlyTenant"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(merged.SpecSelectors["bothKeys"].Query).To(Equal("queryFromTenant"))
			Expect(merged.SpecSelectors["bothKeys"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
		})
	})

	Context("system-scoped updateMapping (service)", func() {
		It("updates only the system row when both system and tenant mappings exist for the same GVK", func() {
			db := runaidb.DB()
			ms := MappingApiService{}
			sysCtx := auth.MockAuthContext(systemMappingActor, 0)
			cvSys := NewMappingConsistencyValidator(sysCtx, api.MAPPINGSCOPE_SYSTEM)
			tenantCtx := mappingTestTenantCtx(tenantA)

			handleSystemFormatters(sysCtx, ms, cvSys)
			handleSystemFieldGroups(sysCtx, db, ms, cvSys)

			gvk := api.Gvk{Group: convutils.Ref("mapping.sys.tenant.both"), Kind: "Both", Version: "v1"}

			sysDesc := "sys"
			sysRow := &dbmodel.MappingSelectors{
				ID:          uuid.New(),
				GvkGroup:    convutils.Deref(gvk.Group, ""),
				GvkKind:     gvk.Kind,
				GvkVersion:  gvk.Version,
				Description: convutils.Ref(sysDesc),
				ScopeType:   api.MAPPINGSCOPE_SYSTEM,
				TenantId:    0,
				SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{
					"sysSel": {
						Query:       ".sys",
						FieldGroups: []string{sm.SystemPodSpecFgName},
						Scope:       api.MAPPINGSCOPE_SYSTEM,
					},
				},
				CreatedAt: time.Now().UTC(),
				CreatedBy: systemMappingActor,
				UpdatedAt: time.Now().UTC(),
				UpdatedBy: systemMappingActor,
			}
			Expect(dao.CreateMapping(sysCtx, sysRow)).To(BeNil())

			fgName := "tenantFgSysIso"
			fgDesc := "fg"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"z": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "p",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			tenDesc := "ten"
			tenRow := &dbmodel.MappingSelectors{
				ID:          uuid.New(),
				GvkGroup:    convutils.Deref(gvk.Group, ""),
				GvkKind:     gvk.Kind,
				GvkVersion:  gvk.Version,
				Description: convutils.Ref(tenDesc),
				ScopeType:   api.MAPPINGSCOPE_TENANT,
				TenantId:    tenantA,
				SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{
					"tenSel": {
						Query:       ".tenant",
						FieldGroups: []string{fgName},
						Scope:       api.MAPPINGSCOPE_TENANT,
					},
				},
				CreatedAt: time.Now().UTC(),
				CreatedBy: "tenant-user",
				UpdatedAt: time.Now().UTC(),
				UpdatedBy: "tenant-user",
			}
			Expect(dao.CreateMapping(tenantCtx, tenRow)).To(BeNil())

			patchDesc := "sys patched"
			_, uerr := ms.updateMapping(sysCtx, &api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &patchDesc,
				SpecSelectors: map[string]api.SpecSelector{
					"sysAdded": {Query: ".added", FieldGroups: []string{sm.SystemPodSpecFgName}},
				},
			}, api.MAPPINGSCOPE_SYSTEM, cvSys)
			Expect(uerr).To(BeNil())

			reloadedTen, err := dao.GetMappingByGvkAndScope(tenantCtx, gvk, api.MAPPINGSCOPE_TENANT)
			Expect(err).To(BeNil())
			Expect(reloadedTen.SpecSelectors).To(HaveLen(1))
			Expect(reloadedTen.SpecSelectors).To(HaveKey("tenSel"))
			Expect(reloadedTen.SpecSelectors["tenSel"].Query).To(Equal(".tenant"))

			reloadedSys, err := dao.GetMappingByGvkAndScope(sysCtx, gvk, api.MAPPINGSCOPE_SYSTEM)
			Expect(err).To(BeNil())
			Expect(reloadedSys.SpecSelectors).To(HaveKey("sysSel"))
			Expect(reloadedSys.SpecSelectors).To(HaveKey("sysAdded"))
			Expect(*reloadedSys.Description).To(Equal(patchDesc))
		})

		It("persists and returns spec selector dependency in create, list, and get responses", func() {
			fgName := "tenantFgForDependency"
			fgDesc := "fg for dependency test"
			postFg := api.FieldGroupCreationRequest{
				Name:        fgName,
				Description: &fgDesc,
				Fields: map[string]api.FieldGroupField{
					"annotations": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "annotations",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqFg, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(postFg))
			Expect(err).To(BeNil())
			reqFg.Header = *headerForAdmin(tenantA)
			respFg := httptest.NewRecorder()
			router.ServeHTTP(respFg, reqFg.WithContext(tenantContextForTest(tenantA)))
			Expect(respFg.Code).To(Equal(http.StatusOK), respFg.Body.String())

			gvk := api.Gvk{Group: convutils.Ref("mapping.dependency.test"), Kind: "TestWorkload", Version: "v1"}
			mapDesc := "mapping with dependency"
			parentSelector := "containerByName"
			dependencyElement := "metadata"

			createBody := api.MappingCreUpdRequest{
				WorkloadType: gvk,
				Description:  &mapDesc,
				SpecSelectors: map[string]api.SpecSelector{
					"containerByName": {
						Query: `.spec.containers[] | select(.name == $ARGS.named.containerName)`,
						ArgsInfo: map[string]api.MappingFieldType{
							"containerName": api.MAPPINGFIELDTYPE_STRING,
						},
						FieldGroups: []string{fgName},
					},
					"containerMeta": {
						Query: `.spec.containers[] | select(.name == $ARGS.named.containerName) | .metadata`,
						ArgsInfo: map[string]api.MappingFieldType{
							"containerName": api.MAPPINGFIELDTYPE_STRING,
						},
						FieldGroups: []string{fgName},
						Dependency: &api.SpecSelectorDependency{
							SpecSelector: &parentSelector,
							Path:         dependencyElement,
						},
					},
				},
			}

			// Create mapping
			reqCreate, err := http.NewRequest(http.MethodPost, urlMappings, makePayload(createBody))
			Expect(err).To(BeNil())
			reqCreate.Header = *headerForAdmin(tenantA)
			respCreate := httptest.NewRecorder()
			router.ServeHTTP(respCreate, reqCreate.WithContext(tenantContextForTest(tenantA)))
			Expect(respCreate.Code).To(Equal(http.StatusOK), respCreate.Body.String())

			var created api.Mapping
			Expect(json.Unmarshal(respCreate.Body.Bytes(), &created)).To(Succeed())

			// Verify dependency in create response
			Expect(created.SpecSelectors).To(HaveKey("containerMeta"))
			Expect(created.SpecSelectors["containerMeta"].Dependency).ToNot(BeNil())
			Expect(created.SpecSelectors["containerMeta"].Dependency.SpecSelector).To(Equal(&parentSelector))
			Expect(created.SpecSelectors["containerMeta"].Dependency.Path).To(Equal(dependencyElement))

			// Verify containerByName has no dependency
			Expect(created.SpecSelectors).To(HaveKey("containerByName"))
			Expect(created.SpecSelectors["containerByName"].Dependency).To(BeNil())

			// List mappings and verify dependency is returned
			reqList, err := http.NewRequest(http.MethodGet, urlMappings, nil)
			Expect(err).To(BeNil())
			reqList.Header = *headerForAdmin(tenantA)
			respList := httptest.NewRecorder()
			router.ServeHTTP(respList, reqList.WithContext(tenantContextForTest(tenantA)))
			Expect(respList.Code).To(Equal(http.StatusOK), respList.Body.String())

			var list api.MappingListResponse
			Expect(json.Unmarshal(respList.Body.Bytes(), &list)).To(Succeed())
			listed, found := mappingListFindGvk(list, gvk)
			Expect(found).To(BeTrue())
			Expect(listed.SpecSelectors).To(HaveKey("containerMeta"))
			Expect(listed.SpecSelectors["containerMeta"].Dependency).ToNot(BeNil())
			Expect(listed.SpecSelectors["containerMeta"].Dependency.SpecSelector).To(Equal(&parentSelector))
			Expect(listed.SpecSelectors["containerMeta"].Dependency.Path).To(Equal(dependencyElement))
		})
	})
})
