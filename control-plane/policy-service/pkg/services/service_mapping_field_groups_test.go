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
	"net/url"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const urlFieldGroups = "/api/v1/policies/field-groups"

func fieldGroupByNameURL(name string) string {
	return fmt.Sprintf("%s/%s", urlFieldGroups, name)
}

func fieldGroupFieldsURL(name string) string {
	return fmt.Sprintf("%s/%s/fields", urlFieldGroups, name)
}

func fieldGroupItemSelectorsURL(name string) string {
	return fmt.Sprintf("%s/%s/item-selectors", urlFieldGroups, name)
}

// mappingTestTenantCtx matches auth.TenantIdOf: CfTenantId must be a string (MockAuthContext uses int32 and yields tenant 0).
func mappingTestTenantCtx(tenantID int32) context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfSubjectID, "tenant-user")
	return context.WithValue(ctx, runaictx.CfTenantId, fmt.Sprintf("%d", tenantID))
}

// Field group service behavior: merged shape validation with and without a system row for the same name.
var _ = Describe("Service mapping field groups", Ordered, func() {
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
	const tenantB int32 = 2
	tenantCtxA := mappingTestTenantCtx(tenantA)
	tenantCtxB := mappingTestTenantCtx(tenantB)
	sysAdminCtx := auth.MockAuthContext(systemMappingActor, 0)

	sharedSelector := "sharedItemSelector"

	Context("authentication", func() {
		It("rejects create without auth headers", func() {
			body := api.FieldGroupCreationRequest{
				Name: "fg1",
			}

			request, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(body))
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})
	})

	// Covers SpecMappingAPIServicer field-group handlers: list, create, get, patch, delete, delete fields, delete item selectors.
	Context("field group APIs", func() {
		It("lists, creates, gets, patches, deletes fields and item selectors, then deletes the group", func() {
			name := "tenantFgHttpCrud"
			desc0 := "initial"
			desc1 := "updated"

			createBody := api.FieldGroupCreationRequest{
				Name:        name,
				Description: &desc0,
				Fields: map[string]api.FieldGroupField{
					"fldWithSelector": {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "specA",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: []string{"s1"},
					},
					"fldWithNoSelector": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "specB",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
				ItemSelectors: map[string]api.ItemSelector{
					"s1": {Query: "."},
					"s2": {Query: ".unused"},
				},
			}

			reqList0, err := http.NewRequest(http.MethodGet, urlFieldGroups, nil)
			Expect(err).To(BeNil())
			reqList0.Header = *headerForAdmin(tenantA)
			respList0 := httptest.NewRecorder()
			router.ServeHTTP(respList0, reqList0.WithContext(tenantContextForTest(tenantA)))
			Expect(respList0.Code).To(Equal(http.StatusOK), respList0.Body.String())

			reqPost, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantA)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantA)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())

			var created api.FieldGroup
			Expect(json.Unmarshal(respPost.Body.Bytes(), &created)).To(Succeed())
			Expect(created.Name).To(Equal(name))
			Expect(created.Description).ToNot(BeNil())
			Expect(*created.Description).To(Equal(desc0))
			Expect(created.Fields).To(HaveKey("fldWithSelector"))
			Expect(created.Fields["fldWithSelector"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(created.Fields).To(HaveKey("fldWithNoSelector"))
			Expect(created.Fields["fldWithNoSelector"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(created.ItemSelectors).To(HaveKey("s1"))
			Expect(created.ItemSelectors["s1"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(created.ItemSelectors).To(HaveKey("s2"))
			Expect(created.ItemSelectors["s2"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))

			reqList1, err := http.NewRequest(http.MethodGet, urlFieldGroups, nil)
			Expect(err).To(BeNil())
			reqList1.Header = *headerForAdmin(tenantA)
			respList1 := httptest.NewRecorder()
			router.ServeHTTP(respList1, reqList1.WithContext(tenantContextForTest(tenantA)))
			Expect(respList1.Code).To(Equal(http.StatusOK), respList1.Body.String())
			var listResp api.FieldGroupListResponse
			Expect(json.Unmarshal(respList1.Body.Bytes(), &listResp)).To(Succeed())
			found := false
			var listItem api.FieldGroup
			for _, it := range listResp.Items {
				if it.Name == name {
					listItem = it
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
			Expect(listItem.Fields["fldWithSelector"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(listItem.ItemSelectors["s2"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))

			reqGet, err := http.NewRequest(http.MethodGet, fieldGroupByNameURL(name), nil)
			Expect(err).To(BeNil())
			reqGet.Header = *headerForAdmin(tenantA)
			respGet := httptest.NewRecorder()
			router.ServeHTTP(respGet, reqGet.WithContext(tenantContextForTest(tenantA)))
			Expect(respGet.Code).To(Equal(http.StatusOK), respGet.Body.String())
			var gotByName api.FieldGroup
			Expect(json.Unmarshal(respGet.Body.Bytes(), &gotByName)).To(Succeed())
			Expect(gotByName.Fields["fldWithSelector"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(gotByName.ItemSelectors["s1"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))

			patchBody := api.FieldGroupUpdateRequest{Description: convutils.Ref(desc1)}
			reqPatch, err := http.NewRequest(http.MethodPatch, fieldGroupByNameURL(name), makePayload(patchBody))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantA)
			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantA)))
			Expect(respPatch.Code).To(Equal(http.StatusOK), respPatch.Body.String())
			var patched api.FieldGroup
			Expect(json.Unmarshal(respPatch.Body.Bytes(), &patched)).To(Succeed())
			Expect(patched.Description).ToNot(BeNil())
			Expect(*patched.Description).To(Equal(desc1))
			Expect(patched.Fields["fldWithSelector"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(patched.Fields["fldWithNoSelector"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(patched.ItemSelectors["s1"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(patched.ItemSelectors["s2"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))

			reqDelFields, err := http.NewRequest(http.MethodDelete, fieldGroupFieldsURL(name), makePayload(api.FieldGroupDeleteFieldsRequest{
				Fields: []string{"fldWithNoSelector"},
			}))
			Expect(err).To(BeNil())
			reqDelFields.Header = *headerForAdmin(tenantA)
			respDelFields := httptest.NewRecorder()
			router.ServeHTTP(respDelFields, reqDelFields.WithContext(tenantContextForTest(tenantA)))
			Expect(respDelFields.Code).To(Equal(http.StatusOK), respDelFields.Body.String())
			var afterDelFields api.FieldGroup
			Expect(json.Unmarshal(respDelFields.Body.Bytes(), &afterDelFields)).To(Succeed())
			Expect(afterDelFields.Fields).To(HaveKey("fldWithSelector"))
			Expect(afterDelFields.Fields).ToNot(HaveKey("fldWithNoSelector"))
			Expect(afterDelFields.ItemSelectors).To(HaveKey("s1"))
			Expect(afterDelFields.ItemSelectors).To(HaveKey("s2"))

			reqDelSels, err := http.NewRequest(http.MethodDelete, fieldGroupItemSelectorsURL(name), makePayload(api.FieldGroupDeleteItemSelectorsRequest{
				ItemSelectors: []string{"s2"},
			}))
			Expect(err).To(BeNil())
			reqDelSels.Header = *headerForAdmin(tenantA)
			respDelSels := httptest.NewRecorder()
			router.ServeHTTP(respDelSels, reqDelSels.WithContext(tenantContextForTest(tenantA)))
			Expect(respDelSels.Code).To(Equal(http.StatusOK), respDelSels.Body.String())
			var afterDelSels api.FieldGroup
			Expect(json.Unmarshal(respDelSels.Body.Bytes(), &afterDelSels)).To(Succeed())
			Expect(afterDelSels.Fields).To(HaveKey("fldWithSelector"))
			Expect(afterDelSels.ItemSelectors).To(HaveKey("s1"))
			Expect(afterDelSels.ItemSelectors).ToNot(HaveKey("s2"))

			reqGet2, err := http.NewRequest(http.MethodGet, fieldGroupByNameURL(name), nil)
			Expect(err).To(BeNil())
			reqGet2.Header = *headerForAdmin(tenantA)
			respGet2 := httptest.NewRecorder()
			router.ServeHTTP(respGet2, reqGet2.WithContext(tenantContextForTest(tenantA)))
			Expect(respGet2.Code).To(Equal(http.StatusOK), respGet2.Body.String())
			var afterSub api.FieldGroup
			Expect(json.Unmarshal(respGet2.Body.Bytes(), &afterSub)).To(Succeed())
			Expect(afterSub.Fields).To(HaveKey("fldWithSelector"))
			Expect(afterSub.Fields["fldWithSelector"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(afterSub.Fields).ToNot(HaveKey("fldWithNoSelector"))
			Expect(afterSub.ItemSelectors).To(HaveKey("s1"))
			Expect(afterSub.ItemSelectors["s1"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(afterSub.ItemSelectors).ToNot(HaveKey("s2"))

			reqDelGroup, err := http.NewRequest(http.MethodDelete, fieldGroupByNameURL(name), nil)
			Expect(err).To(BeNil())
			reqDelGroup.Header = *headerForAdmin(tenantA)
			respDelGroup := httptest.NewRecorder()
			router.ServeHTTP(respDelGroup, reqDelGroup.WithContext(tenantContextForTest(tenantA)))
			Expect(respDelGroup.Code).To(Equal(http.StatusNoContent), respDelGroup.Body.String())

			reqGet404, err := http.NewRequest(http.MethodGet, fieldGroupByNameURL(name), nil)
			Expect(err).To(BeNil())
			reqGet404.Header = *headerForAdmin(tenantA)
			respGet404 := httptest.NewRecorder()
			router.ServeHTTP(respGet404, reqGet404.WithContext(tenantContextForTest(tenantA)))
			Expect(respGet404.Code).To(Equal(http.StatusNotFound), respGet404.Body.String())
		})

		It("returns 400 when PATCH omits description, fields, and itemSelectors", func() {
			name := "tenantFgEmptyPatch"
			desc := "x"
			createBody := api.FieldGroupCreationRequest{
				Name:        name,
				Description: &desc,
				Fields: map[string]api.FieldGroupField{
					"only": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "p",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqPost, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantA)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantA)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())

			reqPatch, err := http.NewRequest(http.MethodPatch, fieldGroupByNameURL(name), makePayload(api.FieldGroupUpdateRequest{}))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantA)
			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantA)))
			Expect(respPatch.Code).To(Equal(http.StatusBadRequest), respPatch.Body.String())
			Expect(respPatch.Body.String()).To(ContainSubstring("patch must include description, fields, or itemSelectors"))
		})
	})

	Context("createFieldGroup merged validation", func() {
		It("allows a tenant-only group when each field's item selectors are defined on the tenant row", func() {
			ms := MappingApiService{}
			name := "svcFgTenantSelfContained"

			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f1": newFieldForTest("x", []string{"s1"}, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{
					"s1": newItemSelForTest(".", api.MAPPINGSCOPE_TENANT),
				},
			)

			merged, apiErr := ms.createFieldGroup(tenantCtxA, row)
			Expect(apiErr).To(BeNil())
			Expect(merged.Fields).To(HaveKey("f1"))
			Expect(merged.Fields["f1"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(merged.ItemSelectors).To(HaveKey("s1"))
			Expect(merged.ItemSelectors["s1"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))

			again, err := ms.getFieldGroup(tenantCtxA, name)
			Expect(err).To(BeNil())
			Expect(again.Fields["f1"].ItemSelectors).To(ConsistOf("s1"))
			Expect(again.Fields["f1"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(again.ItemSelectors["s1"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
		})

		It("rejects a tenant-only group when a field references an item selector not present in the merged map", func() {
			ms := MappingApiService{}
			name := "svcFgTenantMissingSelector"

			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f1": newFieldForTest("x", []string{sharedSelector}, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)

			_, apiErr := ms.createFieldGroup(tenantCtxA, row)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(ContainSubstring("field f1 references item selector \"sharedItemSelector\", which is not defined in field group svcFgTenantMissingSelector"))
		})

		It("allows a tenant field to reference an item selector defined only on the system row after system create", func() {
			ms := MappingApiService{}
			name := "svcFgSystemThenTenantMerge"

			sysRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{
					"sysField": newFieldForTest("a", []string{sharedSelector}, api.MAPPINGSCOPE_SYSTEM),
				},
				map[string]api.ItemSelectorsWithScopeValue{
					sharedSelector: newItemSelForTest(".", api.MAPPINGSCOPE_SYSTEM),
				},
			)
			_, apiErr := ms.createFieldGroup(sysAdminCtx, sysRow)
			Expect(apiErr).To(BeNil())

			tenantRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"tenantField": newFieldForTest("b", []string{sharedSelector}, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			merged, apiErr := ms.createFieldGroup(tenantCtxA, tenantRow)
			Expect(apiErr).To(BeNil())

			Expect(merged.Fields).To(HaveKey("sysField"))
			Expect(merged.Fields["sysField"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(merged.Fields).To(HaveKey("tenantField"))
			Expect(merged.Fields["tenantField"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(merged.ItemSelectors).To(HaveKey(sharedSelector))
			Expect(merged.ItemSelectors[sharedSelector].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))

			loaded, err := ms.getFieldGroup(tenantCtxA, name)
			Expect(err).To(BeNil())
			Expect(loaded.Fields["sysField"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(loaded.Fields["tenantField"].ItemSelectors).To(ConsistOf(sharedSelector))
			Expect(loaded.Fields["tenantField"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(loaded.ItemSelectors[sharedSelector].Query).To(Equal("."))
			Expect(loaded.ItemSelectors[sharedSelector].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
		})

		It("uses system item selector until tenant patch overlays it, then delete restores system scope", func() {
			ms := MappingApiService{}
			name := "fgPatchTenantOverlaySelThenUndo"
			sysQuery := `.systemSelQuery`
			tenantQuery := `.tenantOverlayQuery`

			sysRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f1": newFieldForTest("specA", []string{sharedSelector}, api.MAPPINGSCOPE_SYSTEM),
				},
				map[string]api.ItemSelectorsWithScopeValue{
					sharedSelector: newItemSelForTest(sysQuery, api.MAPPINGSCOPE_SYSTEM),
				},
			)
			_, err := ms.createFieldGroup(sysAdminCtx, sysRow)
			Expect(err).To(BeNil())

			tenantRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"tenantF": newFieldForTest("specB", []string{sharedSelector}, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err = ms.createFieldGroup(tenantCtxA, tenantRow)
			Expect(err).To(BeNil())

			afterCreate, err := ms.getFieldGroup(tenantCtxA, name)
			Expect(err).To(BeNil())
			Expect(afterCreate.Fields["tenantF"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(afterCreate.ItemSelectors[sharedSelector].Query).To(Equal(sysQuery))
			Expect(afterCreate.ItemSelectors[sharedSelector].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)

			_, perr := ms.patchFieldGroup(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupUpdateRequest{
				ItemSelectors: map[string]api.ItemSelector{
					sharedSelector: {Query: tenantQuery},
				},
			}, cv)
			Expect(perr).To(BeNil())

			afterPatch, err := ms.getFieldGroup(tenantCtxA, name)
			Expect(err).To(BeNil())
			Expect(afterPatch.ItemSelectors[sharedSelector].Query).To(Equal(tenantQuery))
			Expect(afterPatch.ItemSelectors[sharedSelector].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(afterPatch.Fields["tenantF"].ItemSelectors).To(ConsistOf(sharedSelector))

			_, delErr := ms.deleteFieldGroupItemSelectors(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupDeleteItemSelectorsRequest{
				ItemSelectors: []string{sharedSelector},
			})
			Expect(delErr).To(BeNil())

			afterDelete, err := ms.getFieldGroup(tenantCtxA, name)
			Expect(err).To(BeNil())
			Expect(afterDelete.ItemSelectors[sharedSelector].Query).To(Equal(sysQuery))
			Expect(afterDelete.ItemSelectors[sharedSelector].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(afterDelete.Fields["tenantF"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
		})

		It("patching the tenant row still validates merged shape against system item selectors", func() {
			ms := MappingApiService{}
			name := "svcFgPatchWithSystemSelectors"

			sysRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{
					"sysField": newFieldForTest("a", []string{sharedSelector}, api.MAPPINGSCOPE_SYSTEM),
				},
				map[string]api.ItemSelectorsWithScopeValue{
					sharedSelector: newItemSelForTest(".", api.MAPPINGSCOPE_SYSTEM),
				},
			)
			_, errSys := ms.createFieldGroup(sysAdminCtx, sysRow)
			Expect(errSys).To(BeNil())

			tenantRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"existing": newFieldForTest("b", []string{sharedSelector}, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, errTenant := ms.createFieldGroup(tenantCtxA, tenantRow)
			Expect(errTenant).To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)

			patched, apiErr := ms.patchFieldGroup(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupUpdateRequest{
				Fields: map[string]api.FieldGroupField{
					"patched": {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "c",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: []string{sharedSelector},
					},
				},
			}, cv)
			Expect(apiErr).To(BeNil())
			Expect(patched.Fields).To(HaveKey("patched"))
			Expect(patched.Fields["patched"].ItemSelectors).To(ConsistOf(sharedSelector))
			Expect(patched.Fields["patched"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(patched.Fields["sysField"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(patched.Fields["existing"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(patched.ItemSelectors[sharedSelector].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
		})
	})

	Context("create returns conflict when scope row already exists", func() {
		It("field group — tenant scope", func() {
			ms := MappingApiService{}
			name := "fgDupCreateTenant"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f1": {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "s",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: nil,
						Scope:         api.MAPPINGSCOPE_TENANT,
					},
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(tenantCtxA, row)
			Expect(err).To(BeNil())

			row2 := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f2": {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "t",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: nil,
						Scope:         api.MAPPINGSCOPE_TENANT,
					},
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err2 := ms.createFieldGroup(tenantCtxA, row2)
			Expect(err2).ToNot(BeNil())
			Expect(err2.Code).To(Equal(int32(http.StatusConflict)))
			Expect(err2.Message).To(ContainSubstring("already exists"))
		})

		It("field group — system scope", func() {
			ms := MappingApiService{}
			name := "fgDupCreateSystem"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f1": {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "s",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: nil,
						Scope:         api.MAPPINGSCOPE_SYSTEM,
					},
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(sysAdminCtx, row)
			Expect(err).To(BeNil())

			row2 := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f2": {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "t",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: nil,
						Scope:         api.MAPPINGSCOPE_SYSTEM,
					},
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err2 := ms.createFieldGroup(sysAdminCtx, row2)
			Expect(err2).ToNot(BeNil())
			Expect(err2.Code).To(Equal(int32(http.StatusConflict)))
		})

		It("workload mapping — tenant scope", func() {
			ms := MappingApiService{}
			gvk := api.Gvk{Group: convutils.Ref("dupmap.example"), Kind: "DupKind", Version: "v1"}
			r1 := newTenantMappingRowForConflictTest(tenantA, "tenant-user", gvk)
			_, err := ms.createMapping(tenantCtxA, r1, nil)
			Expect(err).To(BeNil())

			r2 := newTenantMappingRowForConflictTest(tenantA, "tenant-user", gvk)
			_, err2 := ms.createMapping(tenantCtxA, r2, nil)
			Expect(err2).ToNot(BeNil())
			Expect(err2.Code).To(Equal(int32(http.StatusConflict)))
			Expect(err2.Message).To(ContainSubstring("already exists"))
		})
	})

	Context("field group service edge cases", func() {
		It("rejects create when validateFieldGroupFormatters fails (unknown formatter)", func() {
			ms := MappingApiService{}
			name := "fgFmtMissing"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f1": {
						Formatter:     "formatter-does-not-exist-xyz",
						SpecRef:       "s",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: nil,
						Scope:         api.MAPPINGSCOPE_TENANT,
					},
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, apiErr := ms.createFieldGroup(tenantCtxA, row)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(ContainSubstring(`formatter "formatter-does-not-exist-xyz" not found`))
		})

		It("rejects create when the formatter has no matching field definition key", func() {
			ms := MappingApiService{}
			now := time.Now().UTC()
			fmtName := "fmtSingleFieldDef"
			Expect(dao.CreateFormatter(tenantCtxA, &dbmodel.MappingFormatters{
				ID:        uuid.New(),
				Name:      fmtName,
				ScopeType: api.MAPPINGSCOPE_TENANT,
				TenantId:  tenantA,
				Fields:    map[string]api.FormatterField{"onlyKey": minimalFormatterField("a")},
				CreatedBy: "tenant-user",
				CreatedAt: now,
				UpdatedBy: "tenant-user",
				UpdatedAt: now,
			})).To(BeNil())

			name := "fgFmtKeyMismatch"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"wrongKey": {
						Formatter:     fmtName,
						SpecRef:       "s",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: nil,
						Scope:         api.MAPPINGSCOPE_TENANT,
					},
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, apiErr := ms.createFieldGroup(tenantCtxA, row)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(ContainSubstring("has no field definition named"))
			Expect(apiErr.Message).To(ContainSubstring("wrongKey"))
		})

		It("returns 400 from getFieldGroup when the name is only whitespace", func() {
			ms := MappingApiService{}
			_, err := ms.getFieldGroup(tenantCtxA, "   ")
			Expect(err).ToNot(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(Equal("field group name is required"))
		})

		It("returns 400 over HTTP when getting a field group whose name is only whitespace", func() {
			reqGet, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", urlFieldGroups, url.PathEscape("  ")), nil)
			Expect(err).To(BeNil())
			reqGet.Header = *headerForAdmin(tenantA)
			respGet := httptest.NewRecorder()
			router.ServeHTTP(respGet, reqGet.WithContext(tenantContextForTest(tenantA)))
			Expect(respGet.Code).To(Equal(http.StatusBadRequest), respGet.Body.String())
		})

		It("returns 404 when patching the tenant row but only a system row exists for that name", func() {
			ms := MappingApiService{}
			name := "fgSystemOnlyPatchTenant"
			sysRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{
					"s1": newFieldForTest("a", nil, api.MAPPINGSCOPE_SYSTEM),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(sysAdminCtx, sysRow)
			Expect(err).To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)

			_, apiErr := ms.patchFieldGroup(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupUpdateRequest{
				Description: convutils.Ref("nope"),
			}, cv)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusNotFound)))
			Expect(apiErr.Message).To(ContainSubstring("tenant scoped field group fgSystemOnlyPatchTenant not found"))
		})

		It("initializes nil fields and itemSelectors maps on patch via jsonb null then applies description", func() {
			db := runaidb.DB()

			ms := MappingApiService{}
			name := "fgNilMapsPatch"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"keep": newFieldForTest("k", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(tenantCtxA, row)
			Expect(err).To(BeNil())

			dbRow, derr := dao.GetFieldGroupByNameAndScope(tenantCtxA, db, name, api.MAPPINGSCOPE_TENANT)
			Expect(derr).To(BeNil())
			Expect(runaidb.DB().Exec(
				"UPDATE mapping_field_groups SET fields = 'null'::jsonb, item_selectors = 'null'::jsonb WHERE id = ?",
				dbRow.ID,
			)).Error().To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)

			newDesc := "after-null-json"
			_, perr := ms.patchFieldGroup(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupUpdateRequest{
				Description: convutils.Ref(newDesc),
				Fields: map[string]api.FieldGroupField{
					"added": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "p",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}, cv)
			Expect(perr).To(BeNil())

			loaded, gerr := ms.getFieldGroup(tenantCtxA, name)
			Expect(gerr).To(BeNil())
			Expect(loaded.Description).ToNot(BeNil())
			Expect(*loaded.Description).To(Equal(newDesc))
			Expect(loaded.Fields).To(HaveKey("added"))
			Expect(loaded.Fields["added"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
		})

		It("rejects patch when merged fields fail due to formatter not containing the specified field", func() {
			ms := MappingApiService{}
			now := time.Now().UTC()
			fmtName := "fmtForPatchMerged"
			Expect(dao.CreateFormatter(tenantCtxA, &dbmodel.MappingFormatters{
				ID:        uuid.New(),
				Name:      fmtName,
				ScopeType: api.MAPPINGSCOPE_TENANT,
				TenantId:  tenantA,
				Fields:    map[string]api.FormatterField{"okKey": minimalFormatterField("a")},
				CreatedBy: "tenant-user",
				CreatedAt: now,
				UpdatedBy: "tenant-user",
				UpdatedAt: now,
			})).To(BeNil())

			name := "fgPatchFmtMergedFail"
			sysRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{
					"sysF": newFieldForTest("a", nil, api.MAPPINGSCOPE_SYSTEM),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(sysAdminCtx, sysRow)
			Expect(err).To(BeNil())

			tenantRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err = ms.createFieldGroup(tenantCtxA, tenantRow)
			Expect(err).To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)

			_, apiErr := ms.patchFieldGroup(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupUpdateRequest{
				Fields: map[string]api.FieldGroupField{
					"badMerged": {
						Formatter: fmtName,
						SpecRef:   "x",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}, cv)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(ContainSubstring("field \"badMerged\" uses formatter \"fmtForPatchMerged\", which has no field definition named \"badMerged\""))
		})

		It("patches only the system row when both scopes exist (tenant row unchanged)", func() {
			db := runaidb.DB()

			ms := MappingApiService{}
			name := "fgPatchSystemIgnoresTenant"
			sysDesc := "sys-before"
			tenDesc := "ten-fixed"
			sysRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{
					"sysF": newFieldForTest("a", nil, api.MAPPINGSCOPE_SYSTEM),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			sysRow.Description = &sysDesc
			_, err := ms.createFieldGroup(sysAdminCtx, sysRow)
			Expect(err).To(BeNil())

			tenantRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"tenF": newFieldForTest("b", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			tenantRow.Description = &tenDesc
			_, err = ms.createFieldGroup(tenantCtxA, tenantRow)
			Expect(err).To(BeNil())

			cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)

			sysAfter := "sys-after"
			_, perr := ms.patchFieldGroup(sysAdminCtx, name, api.MAPPINGSCOPE_SYSTEM, &api.FieldGroupUpdateRequest{
				Description: convutils.Ref(sysAfter),
			}, cv)
			Expect(perr).To(BeNil())

			sysStored, derr := dao.GetFieldGroupByNameAndScope(sysAdminCtx, db, name, api.MAPPINGSCOPE_SYSTEM)
			Expect(derr).To(BeNil())
			Expect(sysStored.Description).ToNot(BeNil())
			Expect(*sysStored.Description).To(Equal(sysAfter))

			tenStored, derr := dao.GetFieldGroupByNameAndScope(tenantCtxA, db, name, api.MAPPINGSCOPE_TENANT)
			Expect(derr).To(BeNil())
			Expect(tenStored.Description).ToNot(BeNil())
			Expect(*tenStored.Description).To(Equal(tenDesc))
		})

		It("returns 404 when deleting tenant fields but only a system row exists", func() {
			ms := MappingApiService{}
			name := "fgSystemOnlyDelTenantFields"
			sysRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{
					"s1": newFieldForTest("a", nil, api.MAPPINGSCOPE_SYSTEM),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(sysAdminCtx, sysRow)
			Expect(err).To(BeNil())

			_, apiErr := ms.deleteFieldGroupFields(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupDeleteFieldsRequest{
				Fields: []string{"s1"},
			})
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusNotFound)))
			Expect(apiErr.Message).To(ContainSubstring("tenant scoped field group fgSystemOnlyDelTenantFields not found"))
		})

		It("returns validation error when deleting a field that does not exist on the tenant row", func() {
			ms := MappingApiService{}
			name := "fgEmptyFieldsDelete"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(tenantCtxA, row)
			Expect(err).To(BeNil())

			_, apiErr := ms.deleteFieldGroupFields(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupDeleteFieldsRequest{
				Fields: []string{"nope"},
			})
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(ContainSubstring("field nope does not exist"))
		})

		It("returns validation error when delete-fields key is not a valid mapping field name", func() {
			ms := MappingApiService{}
			name := "fgInvalidDelFieldKey"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"ok": newFieldForTest("a", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(tenantCtxA, row)
			Expect(err).To(BeNil())

			_, apiErr := ms.deleteFieldGroupFields(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupDeleteFieldsRequest{
				Fields: []string{"-bad"},
			})
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(ContainSubstring("field name \"-bad\" is invalid"))
		})

		It("returns the cannot-delete message when removing an item selector still referenced by a field", func() {
			ms := MappingApiService{}
			name := "fgSelInUse"
			sel := "inUseSel"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f1": newFieldForTest("a", []string{sel}, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{
					sel: newItemSelForTest(".x", api.MAPPINGSCOPE_TENANT),
				},
			)
			_, err := ms.createFieldGroup(tenantCtxA, row)
			Expect(err).To(BeNil())

			_, apiErr := ms.deleteFieldGroupItemSelectors(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupDeleteItemSelectorsRequest{
				ItemSelectors: []string{sel},
			})
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(Equal(fmt.Sprintf("cannot delete item selector %q, it is being used by field f1", sel)))
		})

		It("returns validation error when deleting an item selector key that does not exist on the tenant row", func() {
			ms := MappingApiService{}
			name := "fgMissingSelDel"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{},
				map[string]api.ItemSelectorsWithScopeValue{
					"only": newItemSelForTest(".", api.MAPPINGSCOPE_TENANT),
				},
			)
			_, err := ms.createFieldGroup(tenantCtxA, row)
			Expect(err).To(BeNil())

			_, apiErr := ms.deleteFieldGroupItemSelectors(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupDeleteItemSelectorsRequest{
				ItemSelectors: []string{"ghost"},
			})
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(ContainSubstring("item selector ghost does not exist"))
		})

		It("uses overrideSliceWhileMerging so tenant item selector list on a field replaces the system list", func() {
			ms := MappingApiService{}
			name := "fgMergeSliceField"
			fname := "sliceField"
			sysRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{
					fname: {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "a",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: []string{"sa", "sb"},
						Scope:         api.MAPPINGSCOPE_SYSTEM,
					},
				},
				map[string]api.ItemSelectorsWithScopeValue{
					"sa": newItemSelForTest(".sa", api.MAPPINGSCOPE_SYSTEM),
					"sb": newItemSelForTest(".sb", api.MAPPINGSCOPE_SYSTEM),
				},
			)
			_, err := ms.createFieldGroup(sysAdminCtx, sysRow)
			Expect(err).To(BeNil())

			tenantRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					fname: {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "b",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: []string{"tx"},
						Scope:         api.MAPPINGSCOPE_TENANT,
					},
				},
				map[string]api.ItemSelectorsWithScopeValue{
					"tx": newItemSelForTest(".tx", api.MAPPINGSCOPE_TENANT),
				},
			)
			_, err = ms.createFieldGroup(tenantCtxA, tenantRow)
			Expect(err).To(BeNil())

			merged, gerr := ms.getFieldGroup(tenantCtxA, name)
			Expect(gerr).To(BeNil())
			Expect(merged.Fields[fname].ItemSelectors).To(ConsistOf("tx"))
			Expect(merged.Fields[fname].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(merged.ItemSelectors["sa"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(merged.ItemSelectors["sb"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(merged.ItemSelectors["tx"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
		})

		It("rejects patch when a new field references an item selector not present in the merged map (validateFieldGroupShape)", func() {
			ms := MappingApiService{}
			name := "fgPatchMissingSelector"
			sel := "definedSel"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"existing": newFieldForTest("a", []string{sel}, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{
					sel: newItemSelForTest(".", api.MAPPINGSCOPE_TENANT),
				},
			)
			_, err := ms.createFieldGroup(tenantCtxA, row)
			Expect(err).To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)

			orphan := "orphanSel"
			_, apiErr := ms.patchFieldGroup(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupUpdateRequest{
				Fields: map[string]api.FieldGroupField{
					"newField": {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "b",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: []string{orphan},
					},
				},
			}, cv)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(ContainSubstring("field newField references item selector \"orphanSel\", which is not defined in field group fgPatchMissingSelector"))
		})

		It("rejects system-scoped patch when a field uses an item selector defined only on the tenant row", func() {
			ms := MappingApiService{}
			name := "fgSysPatchNoTenantSel"
			selOnlyOnTenant := "tenantOnlySel"

			sysRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(sysAdminCtx, sysRow)
			Expect(err).To(BeNil())

			tenantRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{},
				map[string]api.ItemSelectorsWithScopeValue{
					selOnlyOnTenant: newItemSelForTest(`.tenant`, api.MAPPINGSCOPE_TENANT),
				},
			)
			_, err = ms.createFieldGroup(tenantCtxA, tenantRow)
			Expect(err).To(BeNil())

			cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)

			_, apiErr := ms.patchFieldGroup(sysAdminCtx, name, api.MAPPINGSCOPE_SYSTEM, &api.FieldGroupUpdateRequest{
				Fields: map[string]api.FieldGroupField{
					"needsTenantSel": {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "s",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: []string{selOnlyOnTenant},
					},
				},
			}, cv)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(ContainSubstring("field needsTenantSel references item selector \"tenantOnlySel\", which is not defined in field group fgSysPatchNoTenantSel"))
		})

		It("initializes nil fields map on deleteFieldGroupFields after jsonb null", func() {
			db := runaidb.DB()

			ms := MappingApiService{}
			name := "fgDelFieldsNilMap"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"keep": newFieldForTest("k", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(tenantCtxA, row)
			Expect(err).To(BeNil())

			dbRow, derr := dao.GetFieldGroupByNameAndScope(tenantCtxA, db, name, api.MAPPINGSCOPE_TENANT)
			Expect(derr).To(BeNil())
			Expect(runaidb.DB().Exec("UPDATE mapping_field_groups SET fields = 'null'::jsonb WHERE id = ?", dbRow.ID)).Error().To(BeNil())

			_, apiErr := ms.deleteFieldGroupFields(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupDeleteFieldsRequest{
				Fields: []string{"missingField"},
			})
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(ContainSubstring("field missingField does not exist"))
		})

		It("initializes nil itemSelectors map on deleteFieldGroupItemSelectors after jsonb null", func() {
			db := runaidb.DB()

			ms := MappingApiService{}
			name := "fgDelSelsNilMap"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{},
				map[string]api.ItemSelectorsWithScopeValue{
					"only": newItemSelForTest(".", api.MAPPINGSCOPE_TENANT),
				},
			)
			_, err := ms.createFieldGroup(tenantCtxA, row)
			Expect(err).To(BeNil())

			dbRow, derr := dao.GetFieldGroupByNameAndScope(tenantCtxA, db, name, api.MAPPINGSCOPE_TENANT)
			Expect(derr).To(BeNil())
			Expect(runaidb.DB().Exec("UPDATE mapping_field_groups SET item_selectors = 'null'::jsonb WHERE id = ?", dbRow.ID)).Error().To(BeNil())

			_, apiErr := ms.deleteFieldGroupItemSelectors(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupDeleteItemSelectorsRequest{
				ItemSelectors: []string{"missingSel"},
			})
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(ContainSubstring("item selector missingSel does not exist"))
		})

		It("rejects create when formatter name is empty or whitespace (validateFieldGroupFormatters)", func() {
			ms := MappingApiService{}
			cases := []struct {
				groupSuffix string
				fmt         string
			}{
				{"empty", ""},
				{"spaces", "   "},
				{"tabs", "\t  "},
			}
			for _, c := range cases {
				row := newFieldGroupForTest("fgEmptyFmt"+c.groupSuffix, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
					map[string]api.FieldGroupFieldsWithScopeValue{
						"f1": {
							Formatter:     c.fmt,
							SpecRef:       "x",
							Type:          api.MAPPINGFIELDTYPE_STRING,
							ItemSelectors: nil,
							Scope:         api.MAPPINGSCOPE_TENANT,
						},
					},
					map[string]api.ItemSelectorsWithScopeValue{},
				)
				_, apiErr := ms.createFieldGroup(tenantCtxA, row)
				Expect(apiErr).ToNot(BeNil())
				Expect(apiErr.Message).To(Equal(`field f1: formatter name is required`))
			}
		})

		It("rejects patch when a merged field has whitespace-only formatter name", func() {
			ms := MappingApiService{}
			name := "fgPatchWhitespaceFormatter"
			_, err := ms.createFieldGroup(tenantCtxA, newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"base": newFieldForTest("a", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			))
			Expect(err).To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)

			_, apiErr := ms.patchFieldGroup(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupUpdateRequest{
				Fields: map[string]api.FieldGroupField{
					"badFmt": {
						Formatter: " \t ",
						SpecRef:   "z",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}, cv)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Message).To(Equal(`field badFmt: formatter name is required`))
		})
	})

	Context("tenant isolation for field groups", func() {
		It("does not expose another tenant's field group on get or list", func() {
			ms := MappingApiService{}
			name := "fgOwnedByTenantBOnly"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantB, "tenant-b-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f1": newFieldForTest("x", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(tenantCtxB, row)
			Expect(err).To(BeNil())

			visibleB, verr := ms.getFieldGroup(tenantCtxB, name)
			Expect(verr).To(BeNil())
			Expect(visibleB.Fields["f1"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))

			_, gerr := ms.getFieldGroup(tenantCtxA, name)
			Expect(gerr).ToNot(BeNil())
			Expect(gerr.Code).To(Equal(int32(http.StatusNotFound)))
			Expect(gerr.Message).To(ContainSubstring("field group named fgOwnedByTenantBOnly not found"))

			items, lerr := ms.listFieldGroups(tenantCtxA, 0, 500)
			Expect(lerr).To(BeNil())
			for _, it := range items {
				Expect(it.Name).ToNot(Equal(name))
			}
		})

		It("returns 404 when another tenant patches or deletes a tenant-scoped group of another tenant", func() {
			ms := MappingApiService{}
			name := "fgTenantBMutateFromA"
			row := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantB, "tenant-b-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f1": newFieldForTest("x", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			_, err := ms.createFieldGroup(tenantCtxB, row)
			Expect(err).To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)

			_, perr := ms.patchFieldGroup(tenantCtxA, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupUpdateRequest{
				Description: convutils.Ref("from-a"),
			}, cv)
			Expect(perr).ToNot(BeNil())
			Expect(perr.Code).To(Equal(int32(http.StatusNotFound)))
			Expect(perr.Message).To(ContainSubstring("field group named fgTenantBMutateFromA not found"))

			derr := ms.deleteFieldGroup(tenantCtxA, name, api.MAPPINGSCOPE_TENANT)
			Expect(derr).ToNot(BeNil())
			Expect(derr.Code).To(Equal(int32(http.StatusNotFound)))
			Expect(derr.Message).To(ContainSubstring("tenant scoped field group fgTenantBMutateFromA not found"))
		})

		It("does not expose another tenant's field group over HTTP GET", func() {
			name := "fgHttpTenantBOnly"
			desc := "secret"
			createBody := api.FieldGroupCreationRequest{
				Name:        name,
				Description: &desc,
				Fields: map[string]api.FieldGroupField{
					"f1": {
						Formatter: dbmodel.BuiltinFormatterKeyValue,
						SpecRef:   "p",
						Type:      api.MAPPINGFIELDTYPE_STRING,
					},
				},
			}
			reqPost, err := http.NewRequest(http.MethodPost, urlFieldGroups, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantB)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantB)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())
			var createdIso api.FieldGroup
			Expect(json.Unmarshal(respPost.Body.Bytes(), &createdIso)).To(Succeed())
			Expect(createdIso.Fields["f1"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))

			reqGet, err := http.NewRequest(http.MethodGet, fieldGroupByNameURL(name), nil)
			Expect(err).To(BeNil())
			reqGet.Header = *headerForAdmin(tenantA)
			respGet := httptest.NewRecorder()
			router.ServeHTTP(respGet, reqGet.WithContext(tenantContextForTest(tenantA)))
			Expect(respGet.Code).To(Equal(http.StatusNotFound), respGet.Body.String())
			Expect(respGet.Body.String()).To(ContainSubstring("field group named fgHttpTenantBOnly not found"))
		})
	})

})

func newItemSelForTest(query string, scope api.MappingScope) api.ItemSelectorsWithScopeValue {
	return api.ItemSelectorsWithScopeValue{
		Query: query,
		Scope: scope,
	}
}

func newFieldForTest(specRef string, itemSelectors []string, scope api.MappingScope) api.FieldGroupFieldsWithScopeValue {
	return api.FieldGroupFieldsWithScopeValue{
		Formatter:     dbmodel.BuiltinFormatterKeyValue,
		SpecRef:       specRef,
		Type:          api.MAPPINGFIELDTYPE_STRING,
		ItemSelectors: itemSelectors,
		Scope:         scope,
	}
}

func newTenantMappingRowForConflictTest(tenantID int32, subject string, gvk api.Gvk) *dbmodel.MappingSelectors {
	now := time.Now().UTC()
	return &dbmodel.MappingSelectors{
		ID:            uuid.New(),
		GvkGroup:      convutils.Deref(gvk.Group, ""),
		GvkKind:       gvk.Kind,
		GvkVersion:    gvk.Version,
		ScopeType:     api.MAPPINGSCOPE_TENANT,
		TenantId:      tenantID,
		SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{},
		CreatedAt:     now,
		CreatedBy:     subject,
		UpdatedAt:     now,
		UpdatedBy:     subject,
	}
}

func newFieldGroupForTest(name string, scope api.MappingScope, tenantId int32, subject string,
	fields map[string]api.FieldGroupFieldsWithScopeValue,
	itemSelectors map[string]api.ItemSelectorsWithScopeValue,
) *dbmodel.MappingFieldGroups {
	now := time.Now().UTC()
	return &dbmodel.MappingFieldGroups{
		ID:            uuid.New(),
		Name:          name,
		ScopeType:     scope,
		TenantId:      tenantId,
		Fields:        fields,
		ItemSelectors: itemSelectors,
		CreatedAt:     now,
		CreatedBy:     subject,
		UpdatedAt:     now,
		UpdatedBy:     subject,
	}
}
