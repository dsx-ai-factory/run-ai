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
	"github.com/run-ai/runai/runai-application-packages/convutils"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
)

var _ = Describe("BreakdownGenerator", Ordered, func() {
	router := CreateRouter()

	var backendMock *BackendMockClient

	tenantId := int32(1)
	projectId := "1111"
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

	seedBreakdownMapping := func() {
		tenantCtx := auth.MockAuthContext("test-user", tenantId)
		db := database.DB()

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
				"fieldA": {
					SetExpr:  &setA,
					GetExpr:  &getA,
					ArgsInfo: map[string]api.MappingFieldType{"name": api.MAPPINGFIELDTYPE_STRING},
				},
				"fieldB": {
					SetExpr:  &setB,
					GetExpr:  &getB,
					ArgsInfo: map[string]api.MappingFieldType{"port": api.MAPPINGFIELDTYPE_INTEGER},
				},
				"fieldC": {
					SetExpr:  &setC,
					ArgsInfo: map[string]api.MappingFieldType{"flag": api.MAPPINGFIELDTYPE_BOOLEAN},
				},
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

		// selAlpha, selGama ->
		//	  fldA builtin str
		//    fldB custom  str isAlpha
		//    fldC builtin bool
		//    fldD builtin int isAlpha + isShared
		// selBeta ->
		//	  fldA custom	  str isBeta
		//    fldC custom   str
		//    fldE builtin  quantity
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
		Expect(dao.SaveMapping(tenantCtx, dao_utils.DB().WithContext(tenantCtx), mapping)).To(BeNil())
	}

	//
	// fldA + selAlpha -> fldA builtin str
	// fldA + selBeta  -> fldB custom  str  isBeta
	// fldB + selAlpha -> fldB custom  str  isAlpha
	// fldB + selBeta  -> fldB builtin int
	// fldC + selAlpha -> fldC builtin bool
	// fldC + selGama  -> fldC builtin bool
	// fldD + selAlpha -> fldD builtin int  isAlpha + isShated
	// fldE + selBeta  -> fldE builtin quantity
	//
	policyRules := func() []api.PolicyRule {
		required := []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}
		return []api.PolicyRule{
			{Field: "fieldA", SpecSelector: selAlpha, Enforcements: required},
			{Field: "fieldA", SpecSelector: selBeta, ItemSelector: convutils.Ref("itemSelBeta"), Enforcements: required},
			{Field: "fieldB", SpecSelector: selAlpha, ItemSelector: convutils.Ref("itemSelAlpha"), Enforcements: required},
			{Field: "fieldB", SpecSelector: selBeta, Enforcements: required},
			{Field: "fieldC", SpecSelector: selAlpha, Enforcements: required},
			{Field: "fieldC", SpecSelector: selGamma, Enforcements: required},
			{Field: "fieldD", SpecSelector: selAlpha, ItemSelector: convutils.Ref("itemSelAlpha"), Enforcements: required},
			{Field: "fieldE", SpecSelector: selBeta, Enforcements: required},
		}
	}

	It("returns correct variations when field appears in two field groups", func() {
		seedBreakdownMapping()

		req := api.CreateGvkPolicyReq{
			WorkloadType: gvk,
			Scope:        api.SCOPETYPE_PROJECT,
			ProjectId:    convutils.Ref(projectId),
			Rules:        policyRules(),
		}
		created := createGvkPolicyForTest(router, tenantId, req)
		resp := getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)

		Expect(resp.Breakdown).NotTo(BeNil())
		bd := resp.Breakdown

		// fieldA: two variations (one from fga via selAlpha, one from fgb via selBeta)
		Expect(bd.Fields).To(HaveKey("fieldA"))
		Expect(bd.Fields["fieldA"]).To(HaveLen(2))
		var fromAlpha, fromBeta *api.BreakdownFieldVariation
		for i := range bd.Fields["fieldA"] {
			v := &bd.Fields["fieldA"][i]
			if v.SpecRef == "specA" {
				fromAlpha = v
			} else if v.SpecRef == "specA-alt" {
				fromBeta = v
			}
		}
		Expect(fromAlpha).NotTo(BeNil(), "expected variation from fga")
		Expect(fromAlpha.SpecSelectors).To(ConsistOf(selAlpha))
		Expect(fromAlpha.Formatter.Builtin).NotTo(BeNil())
		Expect(*fromAlpha.Formatter.Builtin).To(Equal(fmtBuiltin))
		Expect(fromAlpha.Type).To(Equal(api.MAPPINGFIELDTYPE_STRING))

		Expect(fromBeta).NotTo(BeNil(), "expected variation from fgb")
		Expect(fromBeta.SpecSelectors).To(ConsistOf(selBeta))
		Expect(fromBeta.Formatter.SetExpr).NotTo(BeNil())
		Expect(*fromBeta.Formatter.SetExpr).To(Equal(`{"a": $ARGS.named.name}`))
		Expect(fromBeta.Formatter.GetExpr).NotTo(BeNil())
		Expect(*fromBeta.Formatter.GetExpr).To(Equal(`.a`))
		Expect(fromBeta.Formatter.ArgsInfo).To(HaveKeyWithValue("name", api.MAPPINGFIELDTYPE_STRING))
		Expect(fromBeta.ItemSelectors).To(HaveKey("itemSelBeta"))
		Expect(fromBeta.ItemSelectors["itemSelBeta"].Query).To(Equal(`(.beta // [])[]`))

		// fieldB: two variations (fga via selAlpha with custom, fgb via selBeta with builtin)
		Expect(bd.Fields).To(HaveKey("fieldB"))
		Expect(bd.Fields["fieldB"]).To(HaveLen(2))
		var fieldBFromAlpha, fieldBFromBeta *api.BreakdownFieldVariation
		for i := range bd.Fields["fieldB"] {
			v := &bd.Fields["fieldB"][i]
			if v.SpecRef == "specB" {
				fieldBFromAlpha = v
			} else if v.SpecRef == "specB-alt" {
				fieldBFromBeta = v
			}
		}
		Expect(fieldBFromAlpha).NotTo(BeNil(), "expected fieldB variation from fga")
		Expect(fieldBFromAlpha.SpecSelectors).To(ConsistOf(selAlpha))
		Expect(fieldBFromAlpha.Formatter.SetExpr).NotTo(BeNil())
		Expect(*fieldBFromAlpha.Formatter.SetExpr).To(Equal(`{"b": $ARGS.named.port}`))
		Expect(fieldBFromAlpha.Formatter.GetExpr).NotTo(BeNil())
		Expect(*fieldBFromAlpha.Formatter.GetExpr).To(Equal(`.b`))
		Expect(fieldBFromAlpha.Formatter.ArgsInfo).To(HaveKeyWithValue("port", api.MAPPINGFIELDTYPE_INTEGER))
		Expect(fieldBFromAlpha.ItemSelectors).To(HaveKey("itemSelAlpha"))
		Expect(fieldBFromAlpha.ItemSelectors["itemSelAlpha"].Query).To(Equal(`(.items // [])[]`))

		Expect(fieldBFromBeta).NotTo(BeNil(), "expected fieldB variation from fgb")
		Expect(fieldBFromBeta.SpecSelectors).To(ConsistOf(selBeta))
		Expect(fieldBFromBeta.Formatter.Builtin).NotTo(BeNil())
		Expect(*fieldBFromBeta.Formatter.Builtin).To(Equal(fmtBuiltin))
		Expect(fieldBFromBeta.Type).To(Equal(api.MAPPINGFIELDTYPE_INTEGER))

		// fieldC: from fga via both selAlpha and selGamma (same field group -> single variation, two selectors)
		Expect(bd.Fields).To(HaveKey("fieldC"))
		Expect(bd.Fields["fieldC"]).To(HaveLen(1))
		Expect(bd.Fields["fieldC"][0].SpecSelectors).To(ConsistOf(selAlpha, selGamma))
		Expect(bd.Fields["fieldC"][0].SpecRef).To(Equal("specC"))
		Expect(bd.Fields["fieldC"][0].Type).To(Equal(api.MAPPINGFIELDTYPE_BOOLEAN))
		Expect(bd.Fields["fieldC"][0].Formatter.Builtin).NotTo(BeNil())
		Expect(*bd.Fields["fieldC"][0].Formatter.Builtin).To(Equal(fmtBuiltin))
		Expect(bd.Fields["fieldC"][0].Formatter.SetExpr).To(BeNil())

		// fieldD: from fga (selAlpha), has two item selectors
		Expect(bd.Fields).To(HaveKey("fieldD"))
		Expect(bd.Fields["fieldD"]).To(HaveLen(1))
		Expect(bd.Fields["fieldD"][0].SpecSelectors).To(ConsistOf(selAlpha))
		Expect(bd.Fields["fieldD"][0].ItemSelectors).To(HaveLen(2))
		Expect(bd.Fields["fieldD"][0].ItemSelectors).To(HaveKey("itemSelAlpha"))
		Expect(bd.Fields["fieldD"][0].ItemSelectors["itemSelAlpha"].Query).To(Equal(`(.items // [])[]`))
		Expect(bd.Fields["fieldD"][0].ItemSelectors).To(HaveKey("itemSelShared"))
		Expect(bd.Fields["fieldD"][0].ItemSelectors["itemSelShared"].Query).To(Equal(`(.shared // [])[]`))
		Expect(bd.Fields["fieldD"][0].Type).To(Equal(api.MAPPINGFIELDTYPE_INTEGER))

		// fieldE: from fgb (selBeta)
		Expect(bd.Fields).To(HaveKey("fieldE"))
		Expect(bd.Fields["fieldE"]).To(HaveLen(1))
		Expect(bd.Fields["fieldE"][0].SpecSelectors).To(ConsistOf(selBeta))
		Expect(bd.Fields["fieldE"][0].SpecRef).To(Equal("specE"))
		Expect(bd.Fields["fieldE"][0].Type).To(Equal(api.MAPPINGFIELDTYPE_QUANTITY))

		// spec selectors: all three referenced
		Expect(bd.SpecSelectors).To(HaveKey(selAlpha))
		Expect(bd.SpecSelectors[selAlpha].Query).To(Equal(`.spec.alpha`))
		Expect(bd.SpecSelectors).To(HaveKey(selBeta))
		Expect(bd.SpecSelectors[selBeta].Query).To(Equal(`.spec.beta`))
		Expect(bd.SpecSelectors).To(HaveKey(selGamma))
		Expect(bd.SpecSelectors[selGamma].Query).To(Equal(`.spec.gamma`))

		// possibleEnforcements: derived from each variation's type +
		// itemSelectors presence. Spot-check three shapes:
		//   - string + no itemSelectors  (fieldA/fromAlpha)
		//   - string + itemSelectors     (fieldA/fromBeta — adds items)
		//   - quantity + no itemSelectors (fieldE — uses quantityRange)
		Expect(fromAlpha.PossibleEnforcements).To(ConsistOf(
			api.MAPPINGPOSSIBLEENFORCEMENT_SET,
			api.MAPPINGPOSSIBLEENFORCEMENT_DEFAULT,
			api.MAPPINGPOSSIBLEENFORCEMENT_REQUIRED,
			api.MAPPINGPOSSIBLEENFORCEMENT_BLOCKED,
			api.MAPPINGPOSSIBLEENFORCEMENT_MATCH,
			api.MAPPINGPOSSIBLEENFORCEMENT_OPTIONS,
			api.MAPPINGPOSSIBLEENFORCEMENT_LENGTH,
		))
		Expect(fromBeta.PossibleEnforcements).To(ConsistOf(
			api.MAPPINGPOSSIBLEENFORCEMENT_SET,
			api.MAPPINGPOSSIBLEENFORCEMENT_DEFAULT,
			api.MAPPINGPOSSIBLEENFORCEMENT_REQUIRED,
			api.MAPPINGPOSSIBLEENFORCEMENT_BLOCKED,
			api.MAPPINGPOSSIBLEENFORCEMENT_MATCH,
			api.MAPPINGPOSSIBLEENFORCEMENT_OPTIONS,
			api.MAPPINGPOSSIBLEENFORCEMENT_LENGTH,
			api.MAPPINGPOSSIBLEENFORCEMENT_ITEMS,
		))
		Expect(bd.Fields["fieldE"][0].PossibleEnforcements).To(ConsistOf(
			api.MAPPINGPOSSIBLEENFORCEMENT_SET,
			api.MAPPINGPOSSIBLEENFORCEMENT_DEFAULT,
			api.MAPPINGPOSSIBLEENFORCEMENT_REQUIRED,
			api.MAPPINGPOSSIBLEENFORCEMENT_BLOCKED,
			api.MAPPINGPOSSIBLEENFORCEMENT_QUANTITY_RANGE,
		))
	})

	It("returns nil breakdown when not requested", func() {
		seedBreakdownMapping()

		req := api.CreateGvkPolicyReq{
			WorkloadType: gvk,
			Scope:        api.SCOPETYPE_PROJECT,
			ProjectId:    convutils.Ref(projectId),
			Rules:        policyRules(),
		}
		created := createGvkPolicyForTest(router, tenantId, req)

		request, err := http.NewRequest(http.MethodGet, "/api/v1/policies/"+created.Id, nil)
		Expect(err).To(BeNil())
		request.Header = *headerForAdmin(tenantId)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

		var resp api.GetGvkPolicyResp
		Expect(json.Unmarshal(response.Body.Bytes(), &resp)).To(Succeed())
		Expect(resp.Id).To(Equal(created.Id))
		Expect(resp.Breakdown).To(BeNil())
	})

	Context("internal endpoint: get by type and scope", func() {
		projectId2 := "1112"

		It("returns the correct policy with breakdown for each project", func() {
			seedBreakdownMapping()

			rulesP1 := []api.PolicyRule{
				{Field: "fieldA", SpecSelector: selAlpha, Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				{Field: "fieldD", SpecSelector: selAlpha, ItemSelector: convutils.Ref("itemSelAlpha"), Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}
			createdP1 := createGvkPolicyForTest(router, tenantId, api.CreateGvkPolicyReq{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules:        rulesP1,
			})

			rulesP2 := []api.PolicyRule{
				{Field: "fieldB", SpecSelector: selBeta, Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				{Field: "fieldE", SpecSelector: selBeta, Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}
			createdP2 := createGvkPolicyForTest(router, tenantId, api.CreateGvkPolicyReq{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId2),
				Rules:        rulesP2,
			})

			Expect(createdP1.Id).NotTo(Equal(createdP2.Id))

			respP1 := getGvkPolicyByTypeAndScopeForTest(router, tenantId, api.WorkloadTypeAndScope{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
			})
			Expect(respP1.Id).To(Equal(createdP1.Id))
			Expect(respP1.Breakdown).NotTo(BeNil())
			Expect(respP1.Breakdown.Fields).To(HaveKey("fieldA"))
			Expect(respP1.Breakdown.Fields).To(HaveKey("fieldD"))
			Expect(respP1.Breakdown.Fields).NotTo(HaveKey("fieldB"))
			Expect(respP1.Breakdown.Fields).NotTo(HaveKey("fieldE"))
			Expect(respP1.Breakdown.SpecSelectors).To(HaveKey(selAlpha))
			Expect(respP1.Breakdown.SpecSelectors).NotTo(HaveKey(selBeta))

			respP2 := getGvkPolicyByTypeAndScopeForTest(router, tenantId, api.WorkloadTypeAndScope{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId2),
			})
			Expect(respP2.Id).To(Equal(createdP2.Id))
			Expect(respP2.Breakdown).NotTo(BeNil())
			Expect(respP2.Breakdown.Fields).To(HaveKey("fieldB"))
			Expect(respP2.Breakdown.Fields).To(HaveKey("fieldE"))
			Expect(respP2.Breakdown.Fields).NotTo(HaveKey("fieldA"))
			Expect(respP2.Breakdown.Fields).NotTo(HaveKey("fieldD"))
			Expect(respP2.Breakdown.SpecSelectors).To(HaveKey(selBeta))
			Expect(respP2.Breakdown.SpecSelectors).NotTo(HaveKey(selAlpha))
		})
	})

	Context("dependency chain inclusion", func() {
		// These tests verify that when a policy references a spec selector
		// that has a dependency declaration, the breakdown returned for the
		// policy includes the full parent chain. The policy engine needs
		// the parent selectors' queries to ensure dependency paths exist
		// before set enforcements run; if the parent is missing the engine
		// silently no-ops on path creation.

		fgDepName := "fg-dep-chain"

		seedChainMapping := func() {
			tenantCtx := auth.MockAuthContext("test-user", tenantId)
			db := database.DB()

			fg := newFieldGroupForTest(fgDepName, api.MAPPINGSCOPE_TENANT, tenantId, "test-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"fieldA": {Formatter: fmtBuiltin, SpecRef: "specA", Type: api.MAPPINGFIELDTYPE_STRING, Scope: api.MAPPINGSCOPE_TENANT},
				},
				nil,
			)
			Expect(dao.SaveFieldGroup(tenantCtx, db, fg)).To(BeNil())

			midParent := "rootSel"
			leafParent := "midSel"

			mapping := &dbmodel.MappingSelectors{
				ID:         uuid.New(),
				GvkGroup:   convutils.Deref(gvk.Group, ""),
				GvkKind:    gvk.Kind,
				GvkVersion: gvk.Version,
				ScopeType:  api.MAPPINGSCOPE_TENANT,
				TenantId:   tenantId,
				SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{
					"rootSel": {
						Query:       `.spec`,
						FieldGroups: []string{fgDepName},
						Scope:       api.MAPPINGSCOPE_TENANT,
					},
					"midSel": {
						Query:       `.spec.template`,
						FieldGroups: []string{fgDepName},
						Dependency: &api.SpecSelectorDependency{
							SpecSelector: &midParent,
							Path:         "template",
						},
						Scope: api.MAPPINGSCOPE_TENANT,
					},
					"leafSel": {
						Query:       `.spec.template.metadata`,
						FieldGroups: []string{fgDepName},
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
			Expect(dao.SaveMapping(tenantCtx, dao_utils.DB().WithContext(tenantCtx), mapping)).To(BeNil())
		}

		It("includes the immediate parent selector when only the child is referenced", func() {
			seedChainMapping()

			rules := []api.PolicyRule{
				{Field: "fieldA", SpecSelector: "midSel", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}
			req := api.CreateGvkPolicyReq{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules:        rules,
			}
			created := createGvkPolicyForTest(router, tenantId, req)
			resp := getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)

			Expect(resp.Breakdown).NotTo(BeNil())
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey("midSel"))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey("rootSel"),
				"parent selector must be included for the engine to resolve the dependency path")
			Expect(resp.Breakdown.SpecSelectors["rootSel"].Query).To(Equal(`.spec`))
		})

		It("includes the full transitive parent chain (leaf -> mid -> root)", func() {
			seedChainMapping()

			rules := []api.PolicyRule{
				{Field: "fieldA", SpecSelector: "leafSel", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}
			req := api.CreateGvkPolicyReq{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
				Rules:        rules,
			}
			created := createGvkPolicyForTest(router, tenantId, req)
			resp := getGvkPolicyWithBreakdownForTest(router, tenantId, created.Id)

			Expect(resp.Breakdown).NotTo(BeNil())
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey("leafSel"))
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey("midSel"),
				"intermediate parent must be included")
			Expect(resp.Breakdown.SpecSelectors).To(HaveKey("rootSel"),
				"root parent must be included")
		})
	})
})

func getGvkPolicyByTypeAndScopeForTest(router chi.Router, tenantId int32, req api.WorkloadTypeAndScope) api.GetGvkPolicyResp {
	request, err := http.NewRequest(http.MethodPost, "/internal/policies", makePayload(req))
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(tenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
	Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

	var result api.GetGvkPolicyResp
	Expect(json.Unmarshal(response.Body.Bytes(), &result)).To(Succeed())
	return result
}

func getGvkPolicyWithBreakdownForTest(router chi.Router, tenantId int32, policyId string) api.GetGvkPolicyResp {
	request, err := http.NewRequest(http.MethodGet, "/api/v1/policies/"+policyId+"?breakdown=true", nil)
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(tenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
	Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

	var result api.GetGvkPolicyResp
	Expect(json.Unmarshal(response.Body.Bytes(), &result)).To(Succeed())
	return result
}
