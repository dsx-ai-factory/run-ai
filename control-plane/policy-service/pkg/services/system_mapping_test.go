/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
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
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const formattersApiPath = "/api/v1/policies/formatter"

func minimalFormatterField(arg string) api.FormatterField {
	setExpr := fmt.Sprintf(`{"x": $ARGS.named.%s}`, arg)
	return api.FormatterField{
		SetExpr: &setExpr,
		ArgsInfo: map[string]api.MappingFieldType{
			arg: api.MAPPINGFIELDTYPE_STRING,
		},
	}
}

func systemFormatterByNameURL(name string) string {
	return fmt.Sprintf("%s/%s", formattersApiPath, name)
}

var _ = Describe("SystemMappingTest", func() {
	router := CreateRouter()

	var backendMock *BackendMockClient
	var clusterMock cluster.SimpleMockClient
	var sysAdminCtx = auth.MockAuthContext(systemMappingActor, 0)

	const tenantID int32 = 2

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()

		backendMock = NewBackendMockClient(2, 2)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))
		authorization.SetMock(NewPoliciesAuthClient(backendMock))

		clusterMock = cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		cluster.SetMock(clusterMock)
	})

	Context("system formatter description versioning", func() {
		It("creates runai-podSpec and seeds fields when missing", func() {
			ms := MappingApiService{}

			handleSystemFormatters(sysAdminCtx, ms, nil)

			row, err := ms.getFormatter(sysAdminCtx, sm.SystemPodSpecFormatterName)
			Expect(err).To(BeNil())
			Expect(row).ToNot(BeNil())
			Expect(row.Description).ToNot(BeNil())
			Expect(*row.Description).To(Equal(systemPodSpecFormatterDesc))
			Expect(len(row.Fields)).To(BeNumerically(">", 0))
		})

		It("does not update formatter when description is unchanged", func() {
			name := "runai-system-version-noop"
			wantDescription := "v-fixed"
			now := time.Now().UTC()

			initial := &dbmodel.MappingFormatters{
				ID:          uuid.New(),
				Name:        name,
				Description: &wantDescription,
				ScopeType:   api.MAPPINGSCOPE_SYSTEM,
				TenantId:    0,
				Fields: map[string]api.FormatterField{
					"keep": minimalFormatterField("k"),
				},
				CreatedBy: systemMappingActor,
				CreatedAt: now,
				UpdatedBy: systemMappingActor,
				UpdatedAt: now,
			}
			Expect(dao.CreateFormatter(auth.MockAuthContext(systemMappingActor, 0), initial)).To(BeNil())

			// If the description matches, this builder should not be applied.
			handleSystemFormatter(
				auth.MockAuthContext(systemMappingActor, 0),
				MappingApiService{},
				name,
				wantDescription,
				func() map[string]api.FormatterField {
					return map[string]api.FormatterField{
						"newField": minimalFormatterField("newArg"),
					}
				},
			)

			after, err := MappingApiService{}.getFormatter(auth.MockAuthContext(systemMappingActor, 0), name)
			Expect(err).To(BeNil())
			Expect(after.Description).ToNot(BeNil())
			Expect(*after.Description).To(Equal(wantDescription))
			Expect(after.Fields).To(HaveKey("keep"))
			Expect(after.Fields).ToNot(HaveKey("newField"))
		})

		It("updates formatter when description changes", func() {
			name := "runai-system-version-upgrade"
			oldDescription := "v-old"
			newDescription := "v-new"
			now := time.Now().UTC()

			initial := &dbmodel.MappingFormatters{
				ID:          uuid.New(),
				Name:        name,
				Description: &oldDescription,
				ScopeType:   api.MAPPINGSCOPE_SYSTEM,
				TenantId:    0,
				Fields: map[string]api.FormatterField{
					"keep": minimalFormatterField("k"),
				},
				CreatedBy: systemMappingActor,
				CreatedAt: now,
				UpdatedBy: systemMappingActor,
				UpdatedAt: now,
			}
			Expect(dao.CreateFormatter(auth.MockAuthContext(systemMappingActor, 0), initial)).To(BeNil())

			handleSystemFormatter(
				auth.MockAuthContext(systemMappingActor, 0),
				MappingApiService{},
				name,
				newDescription,
				func() map[string]api.FormatterField {
					return map[string]api.FormatterField{
						"newField": minimalFormatterField("newArg"),
					}
				},
			)

			after, err := MappingApiService{}.getFormatter(auth.MockAuthContext(systemMappingActor, 0), name)
			Expect(err).To(BeNil())
			Expect(after.Description).ToNot(BeNil())
			Expect(*after.Description).To(Equal(newDescription))
			Expect(after.Fields).ToNot(HaveKey("keep"))
			Expect(after.Fields).To(HaveKey("newField"))
		})
	})

	Context("list/get API with system and tenant formatters", func() {
		It("returns both system and tenant formatters for the tenant", func() {
			handleSystemFormatters(sysAdminCtx, MappingApiService{}, nil)

			tenantFormatterName := "tenant-visible-formatter"
			createBody := api.FormatterCreationRequest{
				Name: tenantFormatterName,
				Fields: map[string]api.FormatterField{
					"ports": minimalFormatterField("containerPort"),
				},
			}

			reqPost, err := http.NewRequest(http.MethodPost, formattersApiPath, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantID)

			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantID)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())

			reqList, err := http.NewRequest(http.MethodGet, formattersApiPath, nil)
			Expect(err).To(BeNil())
			reqList.Header = *headerForAdmin(tenantID)

			respList := httptest.NewRecorder()
			router.ServeHTTP(respList, reqList.WithContext(tenantContextForTest(tenantID)))
			Expect(respList.Code).To(Equal(http.StatusOK), respList.Body.String())

			var listResp api.FormatterListResponse
			Expect(json.Unmarshal(respList.Body.Bytes(), &listResp)).To(Succeed())

			foundSystem := false
			foundTenant := false
			for _, item := range listResp.Items {
				if item.Name == sm.SystemPodSpecFormatterName {
					foundSystem = true
				}
				if item.Name == tenantFormatterName {
					foundTenant = true
				}
			}
			Expect(foundSystem).To(BeTrue())
			Expect(foundTenant).To(BeTrue())

			reqGetSystem, err := http.NewRequest(http.MethodGet, systemFormatterByNameURL(sm.SystemPodSpecFormatterName), nil)
			Expect(err).To(BeNil())
			reqGetSystem.Header = *headerForAdmin(tenantID)

			respGetSystem := httptest.NewRecorder()
			router.ServeHTTP(respGetSystem, reqGetSystem.WithContext(tenantContextForTest(tenantID)))
			Expect(respGetSystem.Code).To(Equal(http.StatusOK), respGetSystem.Body.String())

			var systemFormatter api.Formatter
			Expect(json.Unmarshal(respGetSystem.Body.Bytes(), &systemFormatter)).To(Succeed())
			Expect(systemFormatter.Name).To(Equal(sm.SystemPodSpecFormatterName))
			Expect(systemFormatter.Description).ToNot(BeNil())
			Expect(*systemFormatter.Description).To(Equal(systemPodSpecFormatterDesc))
			Expect(len(systemFormatter.Fields)).To(BeNumerically(">", 0))
		})
	})

	Context("system formatter creation", func() {
		It("seeds apps/Deployment/v1 system mapping after field groups exist", func() {
			db := runaidb.DB()
			ms := MappingApiService{}
			cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)

			handleSystemFormatters(sysAdminCtx, ms, cv)
			handleSystemFieldGroups(sysAdminCtx, db, ms, cv)
			handleSystemMappings(sysAdminCtx, db, ms, cv)

			gvk, _ := sm.BuildSystemAppsDeploymentV1Mapping()
			mapping, err := ms.getMapping(sysAdminCtx, gvk)
			Expect(err).To(BeNil())
			Expect(mapping.Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(mapping.Description).ToNot(BeNil())
			Expect(*mapping.Description).To(Equal(systemDeploymentMappingDesc))

			Expect(mapping.SpecSelectors).To(HaveKey("workloadSpec"))
			Expect(mapping.SpecSelectors).To(HaveKey("workloadMetadata"))
			Expect(mapping.SpecSelectors).To(HaveKey("podSpec"))
			Expect(mapping.SpecSelectors).To(HaveKey("podMetadata"))
			Expect(mapping.SpecSelectors).To(HaveKey("allContainers"))
			Expect(mapping.SpecSelectors).To(HaveKey("allInitContainers"))
			Expect(mapping.SpecSelectors).To(HaveKey("allAppContainers"))
			Expect(mapping.SpecSelectors).To(HaveKey("appContainerByName"))
			Expect(mapping.SpecSelectors).To(HaveKey("initContainerByName"))

			workloadSpec := mapping.SpecSelectors["workloadSpec"]
			Expect(workloadSpec.Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(workloadSpec.FieldGroups).To(Equal([]string{sm.SystemDeploymentSpecFgName, sm.SystemSelectorFgName}))
			Expect(workloadSpec.Query).To(ContainSubstring(".spec"))

			workloadMeta := mapping.SpecSelectors["workloadMetadata"]
			Expect(workloadMeta.Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(workloadMeta.FieldGroups).To(Equal([]string{sm.SystemMetadataFgName}))
			Expect(workloadMeta.Query).To(ContainSubstring(".metadata"))

			podSpec := mapping.SpecSelectors["podSpec"]
			Expect(podSpec.Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(podSpec.FieldGroups).To(Equal([]string{sm.SystemPodSpecFgName, sm.SystemPodSpecAndContainerFgName}))
			Expect(podSpec.Query).To(ContainSubstring(".spec.template.spec"))

			podMeta := mapping.SpecSelectors["podMetadata"]
			Expect(podMeta.Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(podMeta.FieldGroups).To(Equal([]string{sm.SystemMetadataFgName}))
			Expect(podMeta.Query).To(ContainSubstring(".spec.template.metadata"))

			byName := mapping.SpecSelectors["appContainerByName"]
			Expect(byName.ArgsInfo).To(HaveKeyWithValue("containerName", api.MAPPINGFIELDTYPE_STRING))
			Expect(byName.Query).To(ContainSubstring("$ARGS.named.containerName"))
		})
	})

	Context("system workload mapping description versioning", func() {
		It("does not update mapping when description is unchanged (skips seed)", func() {
			db := runaidb.DB()
			ms := MappingApiService{}
			cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)

			handleSystemFormatters(sysAdminCtx, ms, cv)
			handleSystemFieldGroups(sysAdminCtx, db, ms, cv)
			handleSystemMappings(sysAdminCtx, db, ms, cv)

			gvk, _ := sm.BuildSystemAppsDeploymentV1Mapping()
			row, getErr := dao.GetMappingByGvkAndScope(sysAdminCtx, gvk, api.MAPPINGSCOPE_SYSTEM)
			Expect(getErr).To(BeNil())
			Expect(row.SpecSelectors).To(HaveKey("workloadSpec"))

			// The system mapping is a suite-wide singleton (seeded once in TestServices and
			// preserved by DeleteAllTablesForTesting); restore it before this spec ends so
			// later specs - in any order under --randomize-all - still see the canonical
			// `workloadSpec` selector.
			originalSelectors := make(map[string]api.SpecSelectorsWithScopeValue, len(row.SpecSelectors))
			for k, v := range row.SpecSelectors {
				originalSelectors[k] = v
			}
			DeferCleanup(func() {
				restored, rerr := dao.GetMappingByGvkAndScope(sysAdminCtx, gvk, api.MAPPINGSCOPE_SYSTEM)
				Expect(rerr).To(BeNil())
				restored.SpecSelectors = originalSelectors
				Expect(dao.SaveMapping(sysAdminCtx, runaidb.DB().WithContext(sysAdminCtx), restored)).To(BeNil())
			})

			delete(row.SpecSelectors, "workloadSpec")
			Expect(dao.SaveMapping(sysAdminCtx, runaidb.DB().WithContext(sysAdminCtx), row)).To(BeNil())

			handleSystemMappings(sysAdminCtx, db, ms, cv)

			after, err := dao.GetMappingByGvkAndScope(sysAdminCtx, gvk, api.MAPPINGSCOPE_SYSTEM)
			Expect(err).To(BeNil())
			Expect(after.Description).ToNot(BeNil())
			Expect(*after.Description).To(Equal(systemDeploymentMappingDesc))
			Expect(after.SpecSelectors).ToNot(HaveKey("workloadSpec"))
		})

		It("updates description and spec selectors when description differs", func() {
			db := runaidb.DB()
			ms := MappingApiService{}
			cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)

			handleSystemFormatters(sysAdminCtx, ms, cv)
			handleSystemFieldGroups(sysAdminCtx, db, ms, cv)
			handleSystemMappings(sysAdminCtx, db, ms, cv)

			gvk, _ := sm.BuildSystemAppsDeploymentV1Mapping()
			row, getErr := dao.GetMappingByGvkAndScope(sysAdminCtx, gvk, api.MAPPINGSCOPE_SYSTEM)
			Expect(getErr).To(BeNil())

			staleDesc := "apps/Deployment/v1 mapping (stale)"
			row.Description = &staleDesc
			delete(row.SpecSelectors, "workloadSpec")
			delete(row.SpecSelectors, "allContainers")
			Expect(dao.SaveMapping(sysAdminCtx, runaidb.DB().WithContext(sysAdminCtx), row)).To(BeNil())

			handleSystemMappings(sysAdminCtx, db, ms, cv)

			mapping, err := ms.getMapping(sysAdminCtx, gvk)
			Expect(err).To(BeNil())
			Expect(mapping.Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(mapping.Description).ToNot(BeNil())
			Expect(*mapping.Description).To(Equal(systemDeploymentMappingDesc))
			Expect(mapping.SpecSelectors).To(HaveKey("workloadSpec"))
			Expect(mapping.SpecSelectors).To(HaveKey("allContainers"))
			Expect(mapping.SpecSelectors["workloadSpec"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))

			updated, uerr := dao.GetMappingByGvkAndScope(sysAdminCtx, gvk, api.MAPPINGSCOPE_SYSTEM)
			Expect(uerr).To(BeNil())
			Expect(updated.UpdatedBy).To(Equal(systemMappingActor))
		})
	})

	Context("system field group seeding", func() {
		It("seeds podSpec field groups with inverted systemItemSel refs after runai-podSpec formatter exists", func() {
			db := runaidb.DB()
			ms := MappingApiService{}
			cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)

			handleSystemFormatters(sysAdminCtx, ms, cv)
			handleSystemFieldGroups(sysAdminCtx, db, ms, cv)

			podSpec, err := ms.getFieldGroup(sysAdminCtx, sm.SystemPodSpecFgName)
			Expect(err).To(BeNil())
			Expect(podSpec.Fields).To(HaveKey("dnsOptions"))
			Expect(podSpec.Fields["dnsOptions"].Formatter).To(Equal(sm.SystemPodSpecFormatterName))
			Expect(podSpec.Fields["dnsOptions"].ItemSelectors).To(BeEmpty())
			Expect(podSpec.Fields["dnsOptions"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(podSpec.ItemSelectors).To(HaveKey("allDnsOptions"))
			Expect(podSpec.ItemSelectors).To(HaveKey("dnsOptionByName"))
			allDnsOpts := podSpec.ItemSelectors["allDnsOptions"]
			Expect(allDnsOpts.Description).ToNot(BeNil())
			Expect(*allDnsOpts.Description).To(Equal("all DNS options"))
			Expect(allDnsOpts.Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			dnsByName := podSpec.ItemSelectors["dnsOptionByName"]
			Expect(dnsByName.Description).ToNot(BeNil())
			Expect(*dnsByName.Description).To(Equal("DNS option with a given name"))
			Expect(dnsByName.Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))

			containerFg, err := ms.getFieldGroup(sysAdminCtx, sm.SystemPodSpecContainerFgName)
			Expect(err).To(BeNil())
			Expect(containerFg.Fields).To(HaveKey("environmentVariables-FromConfigMap"))
			Expect(containerFg.Fields["environmentVariables-FromConfigMap"].ItemSelectors).To(BeEmpty())
			Expect(containerFg.Fields["environmentVariables-FromConfigMap"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(containerFg.ItemSelectors).To(HaveKey("envVarFromConfigMap"))
			envFromCM := containerFg.ItemSelectors["envVarFromConfigMap"]
			Expect(envFromCM.Description).ToNot(BeNil())
			Expect(*envFromCM.Description).To(Equal("all environment variables sourced from a ConfigMap"))
			Expect(envFromCM.Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))

			andCtr, err := ms.getFieldGroup(sysAdminCtx, sm.SystemPodSpecAndContainerFgName)
			Expect(err).To(BeNil())
			Expect(andCtr.Fields).To(HaveKey("resourceRequests"))
			Expect(andCtr.Fields["resourceRequests"].ItemSelectors).To(BeEmpty())
			Expect(andCtr.Fields["resourceRequests"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(andCtr.ItemSelectors).To(HaveKey("allResourceRequests"))
			allReq := andCtr.ItemSelectors["allResourceRequests"]
			Expect(allReq.Description).ToNot(BeNil())
			Expect(*allReq.Description).To(Equal("all resource requests"))
			Expect(allReq.Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
		})
	})

	Context("system field group description versioning", func() {
		It("does not update field group when description is unchanged", func() {
			db := runaidb.DB()

			ms := MappingApiService{}
			name := "runai-system-fg-version-noop"
			wantDescription := "v-fixed"
			now := time.Now().UTC()

			initial := &dbmodel.MappingFieldGroups{
				ID:          uuid.New(),
				Name:        name,
				Description: &wantDescription,
				ScopeType:   api.MAPPINGSCOPE_SYSTEM,
				TenantId:    0,
				Fields: map[string]api.FieldGroupFieldsWithScopeValue{
					"keep": {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "keepSpec",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: nil,
						Scope:         api.MAPPINGSCOPE_SYSTEM,
					},
				},
				ItemSelectors: map[string]api.ItemSelectorsWithScopeValue{},
				CreatedBy:     systemMappingActor,
				CreatedAt:     now,
				UpdatedBy:     systemMappingActor,
				UpdatedAt:     now,
			}
			_, apiErr := ms.createFieldGroup(sysAdminCtx, initial)
			Expect(apiErr).To(BeNil())

			cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)

			handleSystemFieldGroup(
				sysAdminCtx,
				db,
				ms,
				name,
				wantDescription,
				func() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
					return map[string]api.FieldGroupField{
						"newField": {
							Formatter:     dbmodel.BuiltinFormatterKeyValue,
							SpecRef:       "newSpec",
							Type:          api.MAPPINGFIELDTYPE_STRING,
							ItemSelectors: nil,
						},
					}, map[string]api.ItemSelector{}
				},
				cv)

			after, err := ms.getFieldGroup(sysAdminCtx, name)
			Expect(err).To(BeNil())
			Expect(after.Description).ToNot(BeNil())
			Expect(*after.Description).To(Equal(wantDescription))
			Expect(after.Fields).To(HaveKey("keep"))
			Expect(after.Fields["keep"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(after.Fields).ToNot(HaveKey("newField"))
		})

		It("updates field group when description changes", func() {
			db := runaidb.DB()
			ms := MappingApiService{}
			name := "runai-system-fg-version-upgrade"
			oldDescription := "v-old"
			newDescription := "v-new"
			now := time.Now().UTC()

			initial := &dbmodel.MappingFieldGroups{
				ID:          uuid.New(),
				Name:        name,
				Description: &oldDescription,
				ScopeType:   api.MAPPINGSCOPE_SYSTEM,
				TenantId:    0,
				Fields: map[string]api.FieldGroupFieldsWithScopeValue{
					"keep": {
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       "keepSpec",
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: nil,
						Scope:         api.MAPPINGSCOPE_SYSTEM,
					},
				},
				ItemSelectors: map[string]api.ItemSelectorsWithScopeValue{},
				CreatedBy:     systemMappingActor,
				CreatedAt:     now,
				UpdatedBy:     systemMappingActor,
				UpdatedAt:     now,
			}
			_, apiErr := ms.createFieldGroup(sysAdminCtx, initial)
			Expect(apiErr).To(BeNil())

			cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)

			handleSystemFieldGroup(
				sysAdminCtx,
				db,
				ms,
				name,
				newDescription,
				func() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
					return map[string]api.FieldGroupField{
						"newField": {
							Formatter:     dbmodel.BuiltinFormatterKeyValue,
							SpecRef:       "newSpec",
							Type:          api.MAPPINGFIELDTYPE_STRING,
							ItemSelectors: nil,
						},
					}, map[string]api.ItemSelector{}
				},
				cv)

			after, err := ms.getFieldGroup(sysAdminCtx, name)
			Expect(err).To(BeNil())
			Expect(after.Description).ToNot(BeNil())
			Expect(*after.Description).To(Equal(newDescription))
			Expect(after.Fields).To(HaveKey("keep"))
			Expect(after.Fields["keep"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
			Expect(after.Fields).To(HaveKey("newField"))
			Expect(after.Fields["newField"].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
		})

		It("prunes fields and item selectors the builder no longer defines when description changes", func() {
			db := runaidb.DB()
			ms := MappingApiService{}
			name := "runai-system-fg-version-prune"
			oldDescription := "v-old"
			newDescription := "v-new"

			initial := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{
					"keep":    newFieldForTest("keepSpec", []string{"keptSel"}, api.MAPPINGSCOPE_SYSTEM),
					"stale":   newFieldForTest("staleSpec", []string{"staleSel"}, api.MAPPINGSCOPE_SYSTEM),
					"movesTo": newFieldForTest("movesSpec", []string{"staleSel"}, api.MAPPINGSCOPE_SYSTEM),
				},
				map[string]api.ItemSelectorsWithScopeValue{
					"keptSel":  newItemSelForTest(`.kept[]`, api.MAPPINGSCOPE_SYSTEM),
					"staleSel": newItemSelForTest(`.stale[]`, api.MAPPINGSCOPE_SYSTEM),
				},
			)
			initial.Description = &oldDescription
			_, apiErr := ms.createFieldGroup(sysAdminCtx, initial)
			Expect(apiErr).To(BeNil())

			builder := func() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
				field := func(specRef string, itemSelectors ...string) api.FieldGroupField {
					return api.FieldGroupField{
						Formatter:     dbmodel.BuiltinFormatterKeyValue,
						SpecRef:       specRef,
						Type:          api.MAPPINGFIELDTYPE_STRING,
						ItemSelectors: itemSelectors,
					}
				}
				return map[string]api.FieldGroupField{
					"keep":     field("keepSpec", "keptSel"),
					"movesTo":  field("movesSpec", "keptSel"),
					"newField": field("newSpec"),
				}, map[string]api.ItemSelector{
					"keptSel": {Query: `.kept[]`},
				}
			}

			cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)
			pruneSystemFieldGroup(sysAdminCtx, db, ms, name, newDescription, builder)
			handleSystemFieldGroup(sysAdminCtx, db, ms, name, newDescription, builder, cv)

			after, err := ms.getFieldGroup(sysAdminCtx, name)
			Expect(err).To(BeNil())
			Expect(*after.Description).To(Equal(newDescription))
			Expect(after.Fields).To(HaveLen(3))
			Expect(after.Fields).To(HaveKey("keep"))
			Expect(after.Fields).To(HaveKey("newField"))
			Expect(after.Fields).ToNot(HaveKey("stale"))
			Expect(after.Fields["movesTo"].ItemSelectors).To(Equal([]string{"keptSel"}))
			Expect(after.ItemSelectors).To(HaveLen(1))
			Expect(after.ItemSelectors).To(HaveKey("keptSel"))
		})
	})

	Context("re-seeding a database seeded before the resource claim rename", func() {
		It("removes the reservedResourceClaims and container resourceClaims keys", func() {
			db := runaidb.DB().WithContext(sysAdminCtx)
			ms := MappingApiService{}

			podSpecRow, err := dao.GetFieldGroupByNameAndScope(sysAdminCtx, db, sm.SystemPodSpecFgName, api.MAPPINGSCOPE_SYSTEM)
			Expect(err).To(BeNil())
			andCtrRow, err := dao.GetFieldGroupByNameAndScope(sysAdminCtx, db, sm.SystemPodSpecAndContainerFgName, api.MAPPINGSCOPE_SYSTEM)
			Expect(err).To(BeNil())
			formatterRow, err := dao.GetFormatterByName(sysAdminCtx, db, sm.SystemPodSpecFormatterName)
			Expect(err).To(BeNil())

			// System rows are suite-wide singletons (see DeleteAllTablesForTesting); restore them for later specs.
			podSpecOrig, andCtrOrig, formatterOrig := *podSpecRow, *andCtrRow, *formatterRow
			podSpecOrig.Fields, podSpecOrig.ItemSelectors = maps.Clone(podSpecRow.Fields), maps.Clone(podSpecRow.ItemSelectors)
			andCtrOrig.Fields, andCtrOrig.ItemSelectors = maps.Clone(andCtrRow.Fields), maps.Clone(andCtrRow.ItemSelectors)
			formatterOrig.Fields = maps.Clone(formatterRow.Fields)
			DeferCleanup(func() {
				Expect(dao.SaveFieldGroup(sysAdminCtx, db, &podSpecOrig)).To(BeNil())
				Expect(dao.SaveFieldGroup(sysAdminCtx, db, &andCtrOrig)).To(BeNil())
				Expect(dao.SaveFormatter(sysAdminCtx, db, &formatterOrig)).To(BeNil())
				Expect(dao.WipeAllEntityCaches(sysAdminCtx)).To(BeNil())
			})

			oldItemSelectorNames := map[string]string{
				"allResourceClaims":          "allReservedResourceClaims",
				"resourceClaimByName":        "reservedResourceClaimByName",
				"allDirectResourceClaims":    "allClaimBasedReservedResourceClaims",
				"allTemplatedResourceClaims": "allTemplateBasedReservedResourceClaims",
			}
			oldFieldNames := map[string]string{
				"resourceClaims":                           "reservedResourceClaims",
				"resourceClaims-name":                      "reservedResourceClaims-Name",
				"resourceClaims-resourceClaimName":         "reservedResourceClaims-ClaimName",
				"resourceClaims-resourceClaimTemplateName": "reservedResourceClaims-ClaimTemplate",
			}

			for newName, oldName := range oldItemSelectorNames {
				Expect(podSpecRow.ItemSelectors).To(HaveKey(newName))
				podSpecRow.ItemSelectors[oldName] = podSpecRow.ItemSelectors[newName]
				delete(podSpecRow.ItemSelectors, newName)
			}
			for newName, oldName := range oldFieldNames {
				Expect(podSpecRow.Fields).To(HaveKey(newName))
				field := podSpecRow.Fields[newName]
				var oldSelectors []string
				for _, sel := range field.ItemSelectors {
					oldSelectors = append(oldSelectors, oldItemSelectorNames[sel])
				}
				field.ItemSelectors = oldSelectors
				podSpecRow.Fields[oldName] = field
				delete(podSpecRow.Fields, newName)
			}
			oldPodSpecDesc := "podSpec fields (v3)"
			podSpecRow.Description = &oldPodSpecDesc
			Expect(dao.SaveFieldGroup(sysAdminCtx, db, podSpecRow)).To(BeNil())

			andCtrRow.Fields["resourceClaims"] = api.FieldGroupFieldsWithScopeValue{
				Formatter: sm.SystemPodSpecFormatterName,
				Type:      api.MAPPINGFIELDTYPE_OBJECT,
				Scope:     api.MAPPINGSCOPE_SYSTEM,
			}
			andCtrRow.Fields["resourceClaims-Name"] = newFieldForTest("name", []string{"allResourceClaims"}, api.MAPPINGSCOPE_SYSTEM)
			andCtrRow.Fields["resourceClaims-Request"] = newFieldForTest("request", []string{"allResourceClaims"}, api.MAPPINGSCOPE_SYSTEM)
			andCtrRow.ItemSelectors["allResourceClaims"] = newItemSelForTest(`(.resources.claims // [])[]`, api.MAPPINGSCOPE_SYSTEM)
			oldAndCtrDesc := "podSpecAndContainer fields (v3)"
			andCtrRow.Description = &oldAndCtrDesc
			Expect(dao.SaveFieldGroup(sysAdminCtx, db, andCtrRow)).To(BeNil())

			formatterRow.Fields["reservedResourceClaims"] = formatterRow.Fields["resourceClaims"]
			oldFormatterDesc := "formatter for podSpec fields (v5)"
			formatterRow.Description = &oldFormatterDesc
			Expect(dao.SaveFormatter(sysAdminCtx, db, formatterRow)).To(BeNil())

			cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)
			handleSystemFormatters(sysAdminCtx, ms, cv)
			handleSystemFieldGroups(sysAdminCtx, db, ms, cv)
			handleSystemMappings(sysAdminCtx, db, ms, cv)

			podSpec, gerr := ms.getFieldGroup(sysAdminCtx, sm.SystemPodSpecFgName)
			Expect(gerr).To(BeNil())
			Expect(*podSpec.Description).To(Equal(systemPodSpecFgDesc))
			for newName, oldName := range oldFieldNames {
				Expect(podSpec.Fields).To(HaveKey(newName))
				Expect(podSpec.Fields).ToNot(HaveKey(oldName))
			}
			for newName, oldName := range oldItemSelectorNames {
				Expect(podSpec.ItemSelectors).To(HaveKey(newName))
				Expect(podSpec.ItemSelectors).ToNot(HaveKey(oldName))
			}
			Expect(podSpec.ItemSelectors["allResourceClaims"].Query).To(Equal(`(.resourceClaims // [])[]`))

			andCtr, gerr := ms.getFieldGroup(sysAdminCtx, sm.SystemPodSpecAndContainerFgName)
			Expect(gerr).To(BeNil())
			Expect(*andCtr.Description).To(Equal(systemPodSpecAndContainerFgDesc))
			Expect(andCtr.Fields).ToNot(HaveKey("resourceClaims"))
			Expect(andCtr.Fields).ToNot(HaveKey("resourceClaims-Name"))
			Expect(andCtr.Fields).ToNot(HaveKey("resourceClaims-Request"))
			Expect(andCtr.ItemSelectors).ToNot(HaveKey("allResourceClaims"))
			Expect(andCtr.Fields).To(HaveKey("resourceRequests"))

			formatter, ferr := ms.getFormatter(sysAdminCtx, sm.SystemPodSpecFormatterName)
			Expect(ferr).To(BeNil())
			Expect(*formatter.Description).To(Equal(systemPodSpecFormatterDesc))
			Expect(formatter.Fields).ToNot(HaveKey("reservedResourceClaims"))
			Expect(formatter.Fields).To(HaveKey("resourceClaims"))
		})
	})

	Context("tenant overlay of system item selectors", func() {
		const tenantID int32 = 2
		tenantCtx := mappingTestTenantCtx(tenantID)

		It("exposes the system selector definition after the tenant removes an overlapping key", func() {
			ms := MappingApiService{}
			name := "fgTenantOverlaySel"
			selName := "sharedOverlaySel"

			sysRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
				map[string]api.FieldGroupFieldsWithScopeValue{},
				map[string]api.ItemSelectorsWithScopeValue{
					selName: newItemSelForTest(`sysQuery`, api.MAPPINGSCOPE_SYSTEM),
				},
			)
			_, err := ms.createFieldGroup(sysAdminCtx, sysRow)
			Expect(err).To(BeNil())

			tenantRow := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantID, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"usesSel": newFieldForTest("spec", []string{selName}, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{
					selName: newItemSelForTest(`tenantQuery`, api.MAPPINGSCOPE_TENANT),
				},
			)
			_, err = ms.createFieldGroup(tenantCtx, tenantRow)
			Expect(err).To(BeNil())

			before, gerr := ms.getFieldGroup(tenantCtx, name)
			Expect(gerr).To(BeNil())
			Expect(before.Fields["usesSel"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(before.ItemSelectors[selName].Query).To(Equal(`tenantQuery`))
			Expect(before.ItemSelectors[selName].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))

			_, delErr := ms.deleteFieldGroupItemSelectors(tenantCtx, name, api.MAPPINGSCOPE_TENANT, &api.FieldGroupDeleteItemSelectorsRequest{
				ItemSelectors: []string{selName},
			})
			Expect(delErr).To(BeNil())

			after, gerr := ms.getFieldGroup(tenantCtx, name)
			Expect(gerr).To(BeNil())
			Expect(after.Fields).To(HaveKey("usesSel"))
			Expect(after.Fields["usesSel"].Scope).To(Equal(api.MAPPINGSCOPE_TENANT))
			Expect(after.ItemSelectors[selName].Query).To(Equal(`sysQuery`))
			Expect(after.ItemSelectors[selName].Scope).To(Equal(api.MAPPINGSCOPE_SYSTEM))
		})
	})

})
