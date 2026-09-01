/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/runaidb"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/clients/assets"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/run-ai/runai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/tree"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
)

var _ = Describe("Mapping Breakdown APIs", Ordered, func() {
	router := CreateRouter()

	var backendMock *BackendMockClient

	tenantId := int32(1)
	gvk := api.Gvk{Group: convutils.Ref("apps"), Kind: "Deployment", Version: "v1"}

	const (
		fmtBuiltin = "KeyValue"
		fmtCustom  = "test-custom-formatter"
		fgaName    = "test-fg-alpha"
		fgbName    = "test-fg-beta"
		selAlpha   = "selectorAlpha"
		selBeta    = "selectorBeta"
		selGamma   = "selectorGamma"
	)

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
	})

	seedMapping := func() {
		tenantCtx := auth.MockAuthContext("test-user", tenantId)
		db := runaidb.DB()

		setA := `{"a": $ARGS.named.name}`
		getA := `.a`
		setB := `{"b": $ARGS.named.port}`
		getB := `.b`
		setC := `{"c": $ARGS.named.flag}`
		customFmt := &dbmodel.MappingFormatters{
			ID:        uuid.New(),
			Name:      fmtCustom,
			ScopeType: api.MAPPINGSCOPE_TENANT,
			TenantId:  tenantId,
			Fields: map[string]api.FormatterField{
				"fieldA": {SetExpr: &setA, GetExpr: &getA, ArgsInfo: map[string]api.MappingFieldType{"name": api.MAPPINGFIELDTYPE_STRING}},
				"fieldB": {SetExpr: &setB, GetExpr: &getB, ArgsInfo: map[string]api.MappingFieldType{"port": api.MAPPINGFIELDTYPE_INTEGER}},
				"fieldC": {SetExpr: &setC, ArgsInfo: map[string]api.MappingFieldType{"flag": api.MAPPINGFIELDTYPE_BOOLEAN}},
			},
			CreatedBy: "test-user",
			CreatedAt: time.Now(),
			UpdatedBy: "test-user",
			UpdatedAt: time.Now(),
		}
		Expect(dao.CreateFormatter(tenantCtx, customFmt)).To(BeNil())

		fga := newFieldGroupForTest(fgaName, api.MAPPINGSCOPE_TENANT, tenantId, "test-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"fieldA": {Formatter: fmtBuiltin, SpecRef: "specA", Type: api.MAPPINGFIELDTYPE_STRING, Scope: api.MAPPINGSCOPE_TENANT},
				"fieldB": {Formatter: fmtCustom, SpecRef: "specB", Type: api.MAPPINGFIELDTYPE_STRING, Scope: api.MAPPINGSCOPE_TENANT, ItemSelectors: []string{"itemSelAlpha"}},
				"fieldC": {Formatter: fmtBuiltin, SpecRef: "specC", Type: api.MAPPINGFIELDTYPE_BOOLEAN, Scope: api.MAPPINGSCOPE_TENANT},
				"fieldD": {Formatter: fmtBuiltin, SpecRef: "specD", Type: api.MAPPINGFIELDTYPE_INTEGER, Scope: api.MAPPINGSCOPE_TENANT, ItemSelectors: []string{"itemSelAlpha", "itemSelShared"}},
			},
			map[string]api.ItemSelectorsWithScopeValue{
				"itemSelAlpha":  {Query: `(.items // [])[]`, Scope: api.MAPPINGSCOPE_TENANT},
				"itemSelShared": {Query: `(.shared // [])[]`, Scope: api.MAPPINGSCOPE_TENANT},
			},
		)
		Expect(dao.SaveFieldGroup(tenantCtx, db, fga)).To(BeNil())

		fgb := newFieldGroupForTest(fgbName, api.MAPPINGSCOPE_TENANT, tenantId, "test-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"fieldA": {Formatter: fmtCustom, SpecRef: "specA-alt", Type: api.MAPPINGFIELDTYPE_STRING, Scope: api.MAPPINGSCOPE_TENANT, ItemSelectors: []string{"itemSelBeta"}},
				"fieldB": {Formatter: fmtBuiltin, SpecRef: "specB-alt", Type: api.MAPPINGFIELDTYPE_INTEGER, Scope: api.MAPPINGSCOPE_TENANT},
				"fieldC": {Formatter: fmtCustom, SpecRef: "specC-alt", Type: api.MAPPINGFIELDTYPE_STRING, Scope: api.MAPPINGSCOPE_TENANT},
				"fieldE": {Formatter: fmtBuiltin, SpecRef: "specE", Type: api.MAPPINGFIELDTYPE_QUANTITY, Scope: api.MAPPINGSCOPE_TENANT},
			},
			map[string]api.ItemSelectorsWithScopeValue{
				"itemSelBeta": {Query: `(.beta // [])[]`, Scope: api.MAPPINGSCOPE_TENANT},
			},
		)
		Expect(dao.SaveFieldGroup(tenantCtx, db, fgb)).To(BeNil())

		mapping := &dbmodel.MappingSelectors{
			ID:         uuid.New(),
			GvkGroup:   convutils.Deref(gvk.Group, ""),
			GvkKind:    gvk.Kind,
			GvkVersion: gvk.Version,
			ScopeType:  api.MAPPINGSCOPE_TENANT,
			TenantId:   tenantId,
			SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{
				selAlpha: {Query: `.spec.alpha`, FieldGroups: []string{fgaName}, Scope: api.MAPPINGSCOPE_TENANT},
				selBeta:  {Query: `.spec.beta`, FieldGroups: []string{fgbName}, Scope: api.MAPPINGSCOPE_TENANT},
				selGamma: {Query: `.spec.gamma`, FieldGroups: []string{fgaName}, Scope: api.MAPPINGSCOPE_TENANT},
			},
			CreatedBy: "test-user",
			CreatedAt: time.Now(),
			UpdatedBy: "test-user",
			UpdatedAt: time.Now(),
		}
		Expect(dao.SaveMapping(tenantCtx, runaidb.DB().WithContext(tenantCtx), mapping)).To(BeNil())
	}

	Context("mapping minimal", func() {
		It("returns all fields with their spec selectors and item selectors", func() {
			seedMapping()

			resp := getMappingMinimalForTest(router, tenantId, api.GetMappingMinimalInfoReq{
				WorkloadType: gvk,
			})

			Expect(resp.SpecSelectors).To(HaveKey(selAlpha))
			Expect(resp.SpecSelectors).To(HaveKey(selBeta))
			Expect(resp.SpecSelectors).To(HaveKey(selGamma))

			Expect(resp.Fields).To(HaveKey("fieldA"))
			Expect(resp.Fields["fieldA"]).To(HaveKey(selAlpha))
			Expect(resp.Fields["fieldA"]).To(HaveKey(selBeta))
			Expect(resp.Fields["fieldA"]).To(HaveKey(selGamma))

			Expect(resp.Fields).To(HaveKey("fieldB"))
			Expect(resp.Fields["fieldB"]).To(HaveKey(selAlpha))
			Expect(resp.Fields["fieldB"][selAlpha]).To(HaveKey("itemSelAlpha"))

			Expect(resp.Fields).To(HaveKey("fieldD"))
			Expect(resp.Fields["fieldD"]).To(HaveKey(selAlpha))
			Expect(resp.Fields["fieldD"][selAlpha]).To(HaveLen(2))
			Expect(resp.Fields["fieldD"][selAlpha]).To(HaveKey("itemSelAlpha"))
			Expect(resp.Fields["fieldD"][selAlpha]).To(HaveKey("itemSelShared"))

			Expect(resp.Fields).To(HaveKey("fieldE"))
			Expect(resp.Fields["fieldE"]).To(HaveKey(selBeta))
			Expect(resp.Fields["fieldE"]).NotTo(HaveKey(selAlpha))
		})
	})

	Context("mapping field info", func() {
		It("returns breakdown for specific field refs", func() {
			seedMapping()

			resp := getMappingFieldInfoForTest(router, tenantId, api.GetMappingFieldInfoReq{
				WorkloadType: gvk,
				FieldsRef: []api.BreakdownFieldRef{
					{FieldName: "fieldA", SpecSelector: convutils.Ref(selAlpha)},
					{FieldName: "fieldD", SpecSelector: convutils.Ref(selAlpha)},
				},
			})

			Expect(resp.Breakdown).NotTo(BeNil())
			Expect(resp.Breakdown.Fields).To(HaveKey("fieldA"))
			Expect(resp.Breakdown.Fields).To(HaveKey("fieldD"))
			Expect(resp.Breakdown.Fields).NotTo(HaveKey("fieldB"))
			Expect(resp.Breakdown.Fields).NotTo(HaveKey("fieldE"))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey(selAlpha))
			Expect(resp.Breakdown.SpecSelectors).NotTo(HaveKey(selBeta))

			Expect(resp.Breakdown.Fields["fieldD"]).To(HaveLen(1))
			Expect(resp.Breakdown.Fields["fieldD"][0].ItemSelectors).To(HaveKey("itemSelAlpha"))
			Expect(resp.Breakdown.Fields["fieldD"][0].ItemSelectors["itemSelAlpha"].Query).To(Equal(`(.items // [])[]`))
			Expect(resp.Breakdown.Fields["fieldD"][0].ItemSelectors).To(HaveKey("itemSelShared"))
			Expect(resp.Breakdown.Fields["fieldD"][0].ItemSelectors["itemSelShared"].Query).To(Equal(`(.shared // [])[]`))
		})

		It("returns empty breakdown for invalid field ref", func() {
			seedMapping()

			resp := getMappingFieldInfoForTest(router, tenantId, api.GetMappingFieldInfoReq{
				WorkloadType: gvk,
				FieldsRef: []api.BreakdownFieldRef{
					{FieldName: "nonExistentField", SpecSelector: convutils.Ref(selAlpha)},
				},
			})

			Expect(resp.Breakdown).NotTo(BeNil())
			Expect(resp.Breakdown.Fields).To(BeEmpty())
			Expect(resp.Breakdown.SpecSelectors).To(BeEmpty())
		})

		It("returns breakdown with fields from multiple selectors", func() {
			seedMapping()

			resp := getMappingFieldInfoForTest(router, tenantId, api.GetMappingFieldInfoReq{
				WorkloadType: gvk,
				FieldsRef: []api.BreakdownFieldRef{
					{FieldName: "fieldA", SpecSelector: convutils.Ref(selAlpha)},
					{FieldName: "fieldA", SpecSelector: convutils.Ref(selBeta)},
					{FieldName: "fieldE", SpecSelector: convutils.Ref(selBeta)},
				},
			})

			Expect(resp.Breakdown).NotTo(BeNil())
			Expect(resp.Breakdown.Fields).To(HaveKey("fieldA"))
			Expect(resp.Breakdown.Fields["fieldA"]).To(HaveLen(2))
			Expect(resp.Breakdown.Fields).To(HaveKey("fieldE"))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey(selAlpha))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey(selBeta))
		})

		It("returns all spec selectors when specSelector is omitted", func() {
			seedMapping()

			resp := getMappingFieldInfoForTest(router, tenantId, api.GetMappingFieldInfoReq{
				WorkloadType: gvk,
				FieldsRef: []api.BreakdownFieldRef{
					{FieldName: "fieldA"},
				},
			})

			Expect(resp.Breakdown).NotTo(BeNil())
			Expect(resp.Breakdown.Fields).To(HaveKey("fieldA"))
			Expect(resp.Breakdown.Fields["fieldA"]).To(HaveLen(2))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey(selAlpha))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey(selBeta))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey(selGamma))
		})

		It("returns only matching spec selector when field exists in one selector", func() {
			seedMapping()

			resp := getMappingFieldInfoForTest(router, tenantId, api.GetMappingFieldInfoReq{
				WorkloadType: gvk,
				FieldsRef: []api.BreakdownFieldRef{
					{FieldName: "fieldE"},
				},
			})

			Expect(resp.Breakdown).NotTo(BeNil())
			Expect(resp.Breakdown.Fields).To(HaveKey("fieldE"))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey(selBeta))
			Expect(resp.Breakdown.SpecSelectors).NotTo(HaveKey(selAlpha))
			Expect(resp.Breakdown.SpecSelectors).NotTo(HaveKey(selGamma))
		})

		It("returns breakdown combining explicit and omitted specSelectors", func() {
			seedMapping()

			resp := getMappingFieldInfoForTest(router, tenantId, api.GetMappingFieldInfoReq{
				WorkloadType: gvk,
				FieldsRef: []api.BreakdownFieldRef{
					{FieldName: "fieldA", SpecSelector: convutils.Ref(selAlpha)},
					{FieldName: "fieldE"},
				},
			})

			Expect(resp.Breakdown).NotTo(BeNil())
			Expect(resp.Breakdown.Fields).To(HaveKey("fieldA"))
			Expect(resp.Breakdown.Fields).To(HaveKey("fieldE"))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey(selAlpha))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey(selBeta))
			Expect(resp.Breakdown.SpecSelectors).NotTo(HaveKey(selGamma))
		})

		It("accepts empty group for core Kubernetes resources like Pod", func() {
			tenantCtx := auth.MockAuthContext("test-user", tenantId)
			db := runaidb.DB()

			podGvk := api.Gvk{Group: convutils.Ref(""), Kind: "Pod", Version: "v1"}
			fgName := "fgForPodTest"
			fg := newFieldGroupForTest(fgName, api.MAPPINGSCOPE_TENANT, tenantId, "test-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"annotations": {Formatter: fmtBuiltin, SpecRef: "annotations", Type: api.MAPPINGFIELDTYPE_STRING, Scope: api.MAPPINGSCOPE_TENANT},
				},
				nil,
			)
			Expect(dao.SaveFieldGroup(tenantCtx, db, fg)).To(BeNil())

			mapping := &dbmodel.MappingSelectors{
				ID:         uuid.New(),
				GvkGroup:   convutils.Deref(podGvk.Group, ""),
				GvkKind:    podGvk.Kind,
				GvkVersion: podGvk.Version,
				ScopeType:  api.MAPPINGSCOPE_TENANT,
				TenantId:   tenantId,
				SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{
					"podSpec": {Query: `.spec`, FieldGroups: []string{fgName}, Scope: api.MAPPINGSCOPE_TENANT},
				},
				CreatedBy: "test-user",
				CreatedAt: time.Now(),
				UpdatedBy: "test-user",
				UpdatedAt: time.Now(),
			}
			Expect(dao.SaveMapping(tenantCtx, db, mapping)).To(BeNil())

			resp := getMappingFieldInfoForTest(router, tenantId, api.GetMappingFieldInfoReq{
				WorkloadType: podGvk,
				FieldsRef: []api.BreakdownFieldRef{
					{FieldName: "annotations", SpecSelector: convutils.Ref("podSpec")},
				},
			})

			Expect(resp.Breakdown).NotTo(BeNil())
			Expect(resp.Breakdown.Fields).To(HaveKey("annotations"))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey("podSpec"))
		})

		It("returns spec selector dependency in field-info breakdown response", func() {
			tenantCtx := auth.MockAuthContext("test-user", tenantId)
			db := runaidb.DB()

			fgName := "fgForDepTest"
			fg := newFieldGroupForTest(fgName, api.MAPPINGSCOPE_TENANT, tenantId, "test-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"annotations": {Formatter: fmtBuiltin, SpecRef: "annotations", Type: api.MAPPINGFIELDTYPE_STRING, Scope: api.MAPPINGSCOPE_TENANT},
				},
				nil,
			)
			Expect(dao.SaveFieldGroup(tenantCtx, db, fg)).To(BeNil())

			parentSelector := "containerByName"
			dependencyElement := "metadata"
			depGvk := api.Gvk{Group: convutils.Ref("dep.test"), Kind: "TestWorkload", Version: "v1"}

			mapping := &dbmodel.MappingSelectors{
				ID:         uuid.New(),
				GvkGroup:   convutils.Deref(depGvk.Group, ""),
				GvkKind:    depGvk.Kind,
				GvkVersion: depGvk.Version,
				ScopeType:  api.MAPPINGSCOPE_TENANT,
				TenantId:   tenantId,
				SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{
					"containerByName": {
						Query: `.spec.containers[] | select(.name == $ARGS.named.containerName)`,
						ArgsInfo: map[string]api.MappingFieldType{
							"containerName": api.MAPPINGFIELDTYPE_STRING,
						},
						FieldGroups: []string{fgName},
						Scope:       api.MAPPINGSCOPE_TENANT,
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
						Scope: api.MAPPINGSCOPE_TENANT,
					},
				},
				CreatedBy: "test-user",
				CreatedAt: time.Now(),
				UpdatedBy: "test-user",
				UpdatedAt: time.Now(),
			}
			Expect(dao.CreateMapping(tenantCtx, mapping)).To(BeNil())

			resp := getMappingFieldInfoForTest(router, tenantId, api.GetMappingFieldInfoReq{
				WorkloadType: depGvk,
				FieldsRef: []api.BreakdownFieldRef{
					{FieldName: "annotations", SpecSelector: convutils.Ref("containerMeta")},
				},
			})

			Expect(resp.Breakdown).NotTo(BeNil())
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey("containerMeta"))
			Expect(resp.Breakdown.SpecSelectors["containerMeta"].Dependency).ToNot(BeNil())
			Expect(resp.Breakdown.SpecSelectors["containerMeta"].Dependency.SpecSelector).To(Equal(&parentSelector))
			Expect(resp.Breakdown.SpecSelectors["containerMeta"].Dependency.Path).To(Equal(dependencyElement))

			// The parent selector (containerByName) must also be included in
			// the breakdown even though no rule references it directly. The
			// policy engine needs its query to resolve the dependency path
			// when ensuring intermediate spec elements exist before set
			// enforcements run.
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey("containerByName"),
				"parent selector must be in breakdown for the engine to resolve the dependency path")
			parentBD := resp.Breakdown.SpecSelectors["containerByName"]
			Expect(parentBD.Query).To(Equal(`.spec.containers[] | select(.name == $ARGS.named.containerName)`))
			Expect(parentBD.ArgsInfo).To(HaveKey("containerName"))
		})

		It("includes the full transitive dependency chain in field-info breakdown", func() {
			tenantCtx := auth.MockAuthContext("test-user", tenantId)
			db := runaidb.DB()

			fgName := "fgForChainTest"
			fg := newFieldGroupForTest(fgName, api.MAPPINGSCOPE_TENANT, tenantId, "test-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"annotations": {Formatter: fmtBuiltin, SpecRef: "annotations", Type: api.MAPPINGFIELDTYPE_STRING, Scope: api.MAPPINGSCOPE_TENANT},
				},
				nil,
			)
			Expect(dao.SaveFieldGroup(tenantCtx, db, fg)).To(BeNil())

			// Three-level chain: leafSel -> midSel -> rootSel
			leafParent := "midSel"
			midParent := "rootSel"
			chainGvk := api.Gvk{Group: convutils.Ref("chain.test"), Kind: "ChainWorkload", Version: "v1"}

			mapping := &dbmodel.MappingSelectors{
				ID:         uuid.New(),
				GvkGroup:   convutils.Deref(chainGvk.Group, ""),
				GvkKind:    chainGvk.Kind,
				GvkVersion: chainGvk.Version,
				ScopeType:  api.MAPPINGSCOPE_TENANT,
				TenantId:   tenantId,
				SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{
					"rootSel": {
						Query:       `.spec`,
						FieldGroups: []string{fgName},
						Scope:       api.MAPPINGSCOPE_TENANT,
					},
					"midSel": {
						Query:       `.spec.template`,
						FieldGroups: []string{fgName},
						Dependency: &api.SpecSelectorDependency{
							SpecSelector: &midParent,
							Path:         "template",
						},
						Scope: api.MAPPINGSCOPE_TENANT,
					},
					"leafSel": {
						Query:       `.spec.template.metadata`,
						FieldGroups: []string{fgName},
						Dependency: &api.SpecSelectorDependency{
							SpecSelector: &leafParent,
							Path:         "metadata",
						},
						Scope: api.MAPPINGSCOPE_TENANT,
					},
				},
				CreatedBy: "test-user",
				CreatedAt: time.Now(),
				UpdatedBy: "test-user",
				UpdatedAt: time.Now(),
			}
			Expect(dao.CreateMapping(tenantCtx, mapping)).To(BeNil())

			// Only reference the leaf; the breakdown must walk the chain.
			resp := getMappingFieldInfoForTest(router, tenantId, api.GetMappingFieldInfoReq{
				WorkloadType: chainGvk,
				FieldsRef: []api.BreakdownFieldRef{
					{FieldName: "annotations", SpecSelector: convutils.Ref("leafSel")},
				},
			})

			Expect(resp.Breakdown).NotTo(BeNil())
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey("leafSel"))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey("midSel"),
				"intermediate parent must be included")
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey("rootSel"),
				"root parent must be included")
		})
	})
})

func getMappingMinimalForTest(router chi.Router, tenantId int32, req api.GetMappingMinimalInfoReq) api.GetMappingMinimalResp {
	request, err := http.NewRequest(http.MethodPost, "/api/v1/policies/mappings/minimal", makePayload(req))
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(tenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
	Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

	var result api.GetMappingMinimalResp
	Expect(json.Unmarshal(response.Body.Bytes(), &result)).To(Succeed())
	return result
}

func getMappingFieldInfoForTest(router chi.Router, tenantId int32, req api.GetMappingFieldInfoReq) api.Breakdown {
	request, err := http.NewRequest(http.MethodPost, "/api/v1/policies/mappings/field-info", makePayload(req))
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(tenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
	Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

	var result api.Breakdown
	Expect(json.Unmarshal(response.Body.Bytes(), &result)).To(Succeed())
	return result
}
