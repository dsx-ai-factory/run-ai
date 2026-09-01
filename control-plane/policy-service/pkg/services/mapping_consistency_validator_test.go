/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

// MappingConsistencyValidator: single-selector scenarios (multi-selector support comes later).
var _ = Describe("MappingConsistencyValidator", Ordered, func() {
	const tenantA int32 = 1
	tenantCtxA := mappingTestTenantCtx(tenantA)
	sysAdminCtx := auth.MockAuthContext(systemMappingActor, 0)

	db := runaidb.DB().WithContext(sysAdminCtx)

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()
	})

	It("allows a single field group with one field (one selector)", func() {
		fg := newFieldGroupForTest("fgConsistOne", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"cpu": newFieldForTest(".spec.cpu", nil, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		Expect(dao.CreateFieldGroup(tenantCtxA, db, fg)).To(BeNil())

		Expect(validateMappingChangeForTest(tenantCtxA, api.MAPPINGSCOPE_TENANT, "selA", specSelectorForTest("fgConsistOne"))).To(BeNil())
	})

	It("rejects when two tenant field groups both define the same field key", func() {
		g1 := newFieldGroupForTest("fgConsistT1", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"shared": newFieldForTest(".spec.shared1", nil, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		g2 := newFieldGroupForTest("fgConsistT2", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"shared": newFieldForTest(".spec.shared2", nil, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		Expect(dao.CreateFieldGroup(tenantCtxA, db, g1)).To(BeNil())
		Expect(dao.CreateFieldGroup(tenantCtxA, db, g2)).To(BeNil())

		err := validateMappingChangeForTest(tenantCtxA, api.MAPPINGSCOPE_TENANT, "selDup", specSelectorForTest("fgConsistT1", "fgConsistT2"))
		Expect(err).ToNot(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
		Expect(err.Message).To(ContainSubstring("field shared is sourced from both fgConsistT1 and fgConsistT2 in specSelector selDup"))
	})

	It("allows tenant-scoped validation when a selector lists a system-only field group name (ignored for overlap checks)", func() {
		sysG := newFieldGroupForTest("fgConsistSys", api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
			map[string]api.FieldGroupFieldsWithScopeValue{
				"mem": newFieldForTest(".spec.memSys", nil, api.MAPPINGSCOPE_SYSTEM),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		tenG := newFieldGroupForTest("fgConsistTen", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"mem": newFieldForTest(".spec.memTen", nil, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		Expect(dao.CreateFieldGroup(sysAdminCtx, db, sysG)).To(BeNil())
		Expect(dao.CreateFieldGroup(tenantCtxA, db, tenG)).To(BeNil())

		Expect(validateMappingChangeForTest(tenantCtxA, api.MAPPINGSCOPE_TENANT, "selCrossScope",
			specSelectorForTest("fgConsistSys", "fgConsistTen"))).To(BeNil())
	})

	It("rejects unknown field group names on a selector", func() {
		err := validateMappingChangeForTest(tenantCtxA, api.MAPPINGSCOPE_TENANT, "selMissing", specSelectorForTest("doesNotExist"))
		Expect(err).ToNot(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
		Expect(err.Message).To(ContainSubstring("field group(s)"))
		Expect(err.Message).To(ContainSubstring("doesNotExist"))
		Expect(err.Message).To(ContainSubstring("does not exist"))
	})

	It("rejects when any referenced field group name is unknown (multiple names)", func() {
		fg := newFieldGroupForTest("fgKnownMulti", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"x": newFieldForTest(".spec.x", nil, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		Expect(dao.CreateFieldGroup(tenantCtxA, db, fg)).To(BeNil())

		err := validateMappingChangeForTest(tenantCtxA, api.MAPPINGSCOPE_TENANT, "selMulti",
			specSelectorForTest("fgKnownMulti", "typoOne", "typoTwo"))
		Expect(err).ToNot(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
		Expect(err.Message).To(ContainSubstring("typoOne"))
		Expect(err.Message).To(ContainSubstring("typoTwo"))
		Expect(err.Message).To(ContainSubstring("does not exist"))
	})

	It("rejects unknown field group names when validating a system-scoped mapping", func() {
		cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)
		sel := api.SpecSelectorsWithScopeValue{
			Query:       ".spec",
			FieldGroups: []string{"noSuchSystemFg"},
			Scope:       api.MAPPINGSCOPE_SYSTEM,
			ArgsInfo:    map[string]api.MappingFieldType{},
			ArgsEnum:    map[string][]string{},
		}
		err := cv.ValidateMappingChange(sysAdminCtx, db, map[string]api.SpecSelectorsWithScopeValue{"main": sel})
		Expect(err).ToNot(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
		Expect(err.Message).To(ContainSubstring("noSuchSystemFg"))
		Expect(err.Message).To(ContainSubstring("does not exist"))
	})

	It("rejects when the same field group is listed more than once in a selector", func() {
		dupName := "fgListedTwice"
		sel := specSelectorForTest(dupName, dupName)
		err := validateMappingChangeForTest(tenantCtxA, api.MAPPINGSCOPE_TENANT, "selDupFgNames", sel)
		Expect(err).ToNot(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
		Expect(err.Message).To(Equal(fmt.Sprintf("field group %s is multiply defined in selector selDupFgNames", dupName)))
	})

	It("rejects when two tenant field groups both define the same item selector key", func() {
		shared := "sharedItemSel"
		g1 := newFieldGroupForTest("fgIselA", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"f1": newFieldForTest(".spec.f1", []string{shared}, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{
				shared: newItemSelForTest(".spec.containers", api.MAPPINGSCOPE_TENANT),
			},
		)
		g2 := newFieldGroupForTest("fgIselB", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"f2": newFieldForTest(".spec.f2", []string{shared}, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{
				shared: newItemSelForTest(".spec.init", api.MAPPINGSCOPE_TENANT),
			},
		)
		Expect(dao.CreateFieldGroup(tenantCtxA, db, g1)).To(BeNil())
		Expect(dao.CreateFieldGroup(tenantCtxA, db, g2)).To(BeNil())

		err := validateMappingChangeForTest(tenantCtxA, api.MAPPINGSCOPE_TENANT, "selItemDup", specSelectorForTest("fgIselA", "fgIselB"))
		Expect(err).ToNot(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
		Expect(err.Message).To(ContainSubstring("itemSelector " + shared + " is sourced from both fgIselA and fgIselB in specSelector selItemDup"))
	})

	It("allows the same string as a field key in one group and an item selector key in another", func() {
		// Field and item selector namespaces are independent.
		sameKey := "ports"
		g1 := newFieldGroupForTest("fgFieldNS", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				sameKey: newFieldForTest(".spec.portsField", nil, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		g2 := newFieldGroupForTest("fgItemNS", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"other": newFieldForTest(".spec.other", []string{sameKey}, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{
				sameKey: newItemSelForTest(".spec.portsPath", api.MAPPINGSCOPE_TENANT),
			},
		)
		Expect(dao.CreateFieldGroup(tenantCtxA, db, g1)).To(BeNil())
		Expect(dao.CreateFieldGroup(tenantCtxA, db, g2)).To(BeNil())

		Expect(validateMappingChangeForTest(tenantCtxA, api.MAPPINGSCOPE_TENANT, "selNS", specSelectorForTest("fgFieldNS", "fgItemNS"))).To(BeNil())
	})

	It("ValidateMappingChange runs parallel validation for multiple distinct field-group combinations", func() {
		base := "fgMapParBase"
		Expect(dao.CreateFieldGroup(tenantCtxA, db, newFieldGroupForTest(base, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"b": newFieldForTest(".spec.b", nil, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		))).To(BeNil())

		specs := map[string]api.SpecSelectorsWithScopeValue{}
		for i := range 12 {
			ex := fmt.Sprintf("fgMapParX%d", i)
			Expect(dao.CreateFieldGroup(tenantCtxA, db, newFieldGroupForTest(ex, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"x": newFieldForTest(".spec.x", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			))).To(BeNil())
			specs[fmt.Sprintf("sel%d", i)] = specSelectorForTest(base, ex)
		}

		cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)
		Expect(cv.ValidateMappingChange(tenantCtxA, db, specs)).To(BeNil())
	})

	It("ValidateMappingChange returns validation error when exactly one parallel selector job is inconsistent", func() {
		okOnly := newFieldGroupForTest("fgParFailOk", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"aOnly": newFieldForTest(".spec.a", nil, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		dup1 := newFieldGroupForTest("fgParFailDup1", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"sharedPar": newFieldForTest(".spec.s1", nil, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		dup2 := newFieldGroupForTest("fgParFailDup2", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
			map[string]api.FieldGroupFieldsWithScopeValue{
				"sharedPar": newFieldForTest(".spec.s2", nil, api.MAPPINGSCOPE_TENANT),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		for _, fg := range []*dbmodel.MappingFieldGroups{okOnly, dup1, dup2} {
			Expect(dao.CreateFieldGroup(tenantCtxA, db, fg)).To(BeNil())
		}

		specs := map[string]api.SpecSelectorsWithScopeValue{
			"selParOk":  specSelectorForTest("fgParFailOk"),
			"selParBad": specSelectorForTest("fgParFailDup1", "fgParFailDup2"),
		}
		cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)
		err := cv.ValidateMappingChange(tenantCtxA, db, specs)
		Expect(err).ToNot(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
		Expect(err.Message).To(ContainSubstring("field sharedPar is sourced from both fgParFailDup1 and fgParFailDup2 in specSelector selParBad"))
	})

	It("ValidateMappingChange with system scope loads only system field group rows", func() {
		sysA := newFieldGroupForTest("fgSysValA", api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
			map[string]api.FieldGroupFieldsWithScopeValue{
				"sa": newFieldForTest(".spec.sa", nil, api.MAPPINGSCOPE_SYSTEM),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		sysB := newFieldGroupForTest("fgSysValB", api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
			map[string]api.FieldGroupFieldsWithScopeValue{
				"sb": newFieldForTest(".spec.sb", nil, api.MAPPINGSCOPE_SYSTEM),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		Expect(dao.CreateFieldGroup(sysAdminCtx, db, sysA)).To(BeNil())
		Expect(dao.CreateFieldGroup(sysAdminCtx, db, sysB)).To(BeNil())

		sel := api.SpecSelectorsWithScopeValue{
			Query:       ".spec",
			FieldGroups: []string{"fgSysValA", "fgSysValB"},
			Scope:       api.MAPPINGSCOPE_SYSTEM,
			ArgsInfo:    map[string]api.MappingFieldType{},
			ArgsEnum:    map[string][]string{},
		}
		cv := NewMappingConsistencyValidator(sysAdminCtx, api.MAPPINGSCOPE_SYSTEM)
		Expect(cv.ValidateMappingChange(sysAdminCtx, db, map[string]api.SpecSelectorsWithScopeValue{"main": sel})).To(BeNil())
	})

	It("ValidateMappingChange with tenant scope ignores system-only field group names on selectors", func() {
		sysOnly := newFieldGroupForTest("fgSysOnlyName", api.MAPPINGSCOPE_SYSTEM, 0, systemMappingActor,
			map[string]api.FieldGroupFieldsWithScopeValue{
				"z": newFieldForTest(".spec.z", nil, api.MAPPINGSCOPE_SYSTEM),
			},
			map[string]api.ItemSelectorsWithScopeValue{},
		)
		Expect(dao.CreateFieldGroup(sysAdminCtx, db, sysOnly)).To(BeNil())

		Expect(validateMappingChangeForTest(tenantCtxA, api.MAPPINGSCOPE_TENANT, "ok", specSelectorForTest("fgSysOnlyName"))).To(BeNil())
	})

	It("ValidateMappingChange allows empty fieldGroups on a selector", func() {
		cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)
		empty := api.SpecSelectorsWithScopeValue{
			Query:       ".",
			FieldGroups: []string{},
			Scope:       api.MAPPINGSCOPE_TENANT,
			ArgsInfo:    map[string]api.MappingFieldType{},
			ArgsEnum:    map[string][]string{},
		}
		Expect(cv.ValidateMappingChange(tenantCtxA, db, map[string]api.SpecSelectorsWithScopeValue{"empty": empty})).To(BeNil())
	})

	It("ValidateMappingChange skips duplicate selectors that list the same field groups in different order", func() {
		for _, name := range []string{"fgDedA", "fgDedB", "fgDedC"} {
			fg := newFieldGroupForTest(name, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					name: newFieldForTest(".spec."+name, nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			Expect(dao.CreateFieldGroup(tenantCtxA, db, fg)).To(BeNil())
		}

		s1 := specSelectorForTest("fgDedA", "fgDedB", "fgDedC")
		s2 := api.SpecSelectorsWithScopeValue{
			Query:       s1.Query,
			FieldGroups: []string{"fgDedC", "fgDedA", "fgDedB"},
			Scope:       s1.Scope,
			ArgsInfo:    map[string]api.MappingFieldType{},
			ArgsEnum:    map[string][]string{},
		}

		cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)
		Expect(cv.ValidateMappingChange(tenantCtxA, db, map[string]api.SpecSelectorsWithScopeValue{
			"selector1": s1,
			"selector2": s2,
		})).To(BeNil())
	})

	Context("ValidateFieldGroupChange", func() {
		specSelForStoredMapping := func(fieldGroups ...string) api.SpecSelectorsWithScopeValue {
			return api.SpecSelectorsWithScopeValue{
				Query:       ".spec",
				FieldGroups: fieldGroups,
				Scope:       api.MAPPINGSCOPE_TENANT,
				ArgsInfo:    map[string]api.MappingFieldType{},
				ArgsEnum:    map[string][]string{},
			}
		}

		newTenantMappingRow := func(specSelectors map[string]api.SpecSelectorsWithScopeValue) *dbmodel.MappingSelectors {
			now := time.Now().UTC()
			return &dbmodel.MappingSelectors{
				ID:            uuid.New(),
				GvkGroup:      "apps",
				GvkKind:       "Deployment",
				GvkVersion:    "v1",
				ScopeType:     api.MAPPINGSCOPE_TENANT,
				TenantId:      tenantA,
				SpecSelectors: specSelectors,
				CreatedAt:     now,
				CreatedBy:     "tenant-user",
				UpdatedAt:     now,
				UpdatedBy:     "tenant-user",
			}
		}

		It("returns nil when no mapping references the field group", func() {
			fg := newFieldGroupForTest("fgNoMappingRef", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"x": newFieldForTest(".spec.x", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			Expect(dao.CreateFieldGroup(tenantCtxA, db, fg)).To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)
			Expect(cv.ValidateFieldGroupChange(tenantCtxA, db, "fgNoMappingRef")).To(BeNil())
		})

		It("lists mappings that reference a field group via jsonb fieldGroups", func() {
			fg := newFieldGroupForTest("fgListedByDao", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"a": newFieldForTest(".spec.a", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			Expect(dao.CreateFieldGroup(tenantCtxA, db, fg)).To(BeNil())

			row := newTenantMappingRow(map[string]api.SpecSelectorsWithScopeValue{
				"main": specSelForStoredMapping("fgListedByDao"),
			})
			Expect(dao.CreateMapping(tenantCtxA, row)).To(BeNil())

			rows, err := dao.ListMappingsReferencingFieldGroup(tenantCtxA, db, "fgListedByDao")
			Expect(err).To(BeNil())
			Expect(rows).To(HaveLen(1))
			Expect(rows[0].GvkKind).To(Equal("Deployment"))
		})

		It("fails when a stored mapping selector lists two tenant groups with the same field key", func() {
			g1 := newFieldGroupForTest("fgChgM1", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"dup": newFieldForTest(".spec.d1", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			g2 := newFieldGroupForTest("fgChgM2", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"dup": newFieldForTest(".spec.d2", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			Expect(dao.CreateFieldGroup(tenantCtxA, db, g1)).To(BeNil())
			Expect(dao.CreateFieldGroup(tenantCtxA, db, g2)).To(BeNil())

			row := newTenantMappingRow(map[string]api.SpecSelectorsWithScopeValue{
				"main": specSelForStoredMapping("fgChgM1", "fgChgM2"),
			})
			Expect(dao.CreateMapping(tenantCtxA, row)).To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)
			verr := cv.ValidateFieldGroupChange(tenantCtxA, db, "fgChgM1")
			Expect(verr).ToNot(BeNil())
			Expect(verr.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(verr.Message).To(ContainSubstring("field dup is sourced from both fgChgM1 and fgChgM2 in specSelector selector main in apps/Deployment/v1 mapping"))
		})

		It("fails when a mapping lists two tenant groups that share an item selector key", func() {
			shared := "sharedIselChg"
			g1 := newFieldGroupForTest("fgChgIs1", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f1": newFieldForTest(".a", []string{shared}, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{shared: newItemSelForTest(".x", api.MAPPINGSCOPE_TENANT)},
			)
			g2 := newFieldGroupForTest("fgChgIs2", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"f2": newFieldForTest(".b", []string{shared}, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{shared: newItemSelForTest(".y", api.MAPPINGSCOPE_TENANT)},
			)
			Expect(dao.CreateFieldGroup(tenantCtxA, db, g1)).To(BeNil())
			Expect(dao.CreateFieldGroup(tenantCtxA, db, g2)).To(BeNil())

			row := newTenantMappingRow(map[string]api.SpecSelectorsWithScopeValue{
				"pod": specSelForStoredMapping("fgChgIs1", "fgChgIs2"),
			})
			Expect(dao.CreateMapping(tenantCtxA, row)).To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)
			verr := cv.ValidateFieldGroupChange(tenantCtxA, db, "fgChgIs1")
			Expect(verr).ToNot(BeNil())
			Expect(verr.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(verr.Message).To(ContainSubstring("itemSelector " + shared + " is sourced from both fgChgIs1 and fgChgIs2 in specSelector selector pod in apps/Deployment/v1 mapping"))
		})

		It("deduplicates selectors that list the same field groups", func() {
			fg := newFieldGroupForTest("fgDupSel", api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"only": newFieldForTest(".spec.only", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			Expect(dao.CreateFieldGroup(tenantCtxA, db, fg)).To(BeNil())

			sels := map[string]api.SpecSelectorsWithScopeValue{
				"sel0": specSelForStoredMapping("fgDupSel"),
				"sel1": specSelForStoredMapping("fgDupSel"),
			}
			Expect(dao.CreateMapping(tenantCtxA, newTenantMappingRow(sels))).To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)
			Expect(cv.ValidateFieldGroupChange(tenantCtxA, db, "fgDupSel")).To(BeNil())
		})

		It("runs parallel validation when selectors use different field-group sets", func() {
			base := "fgParBase"
			baseFg := newFieldGroupForTest(base, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
				map[string]api.FieldGroupFieldsWithScopeValue{
					"b": newFieldForTest(".spec.b", nil, api.MAPPINGSCOPE_TENANT),
				},
				map[string]api.ItemSelectorsWithScopeValue{},
			)
			Expect(dao.CreateFieldGroup(tenantCtxA, db, baseFg)).To(BeNil())

			sels := make(map[string]api.SpecSelectorsWithScopeValue)
			for i := range 12 {
				exName := fmt.Sprintf("fgParX%d", i)
				exFg := newFieldGroupForTest(exName, api.MAPPINGSCOPE_TENANT, tenantA, "tenant-user",
					map[string]api.FieldGroupFieldsWithScopeValue{
						"x": newFieldForTest(".spec.x", nil, api.MAPPINGSCOPE_TENANT),
					},
					map[string]api.ItemSelectorsWithScopeValue{},
				)
				Expect(dao.CreateFieldGroup(tenantCtxA, db, exFg)).To(BeNil())
				sels[fmt.Sprintf("sel%d", i)] = specSelForStoredMapping(base, exName)
			}
			Expect(dao.CreateMapping(tenantCtxA, newTenantMappingRow(sels))).To(BeNil())

			cv := NewMappingConsistencyValidator(tenantCtxA, api.MAPPINGSCOPE_TENANT)
			Expect(cv.ValidateFieldGroupChange(tenantCtxA, db, base)).To(BeNil())
		})
	})
})

func specSelectorForTest(fieldGroups ...string) api.SpecSelectorsWithScopeValue {
	return api.SpecSelectorsWithScopeValue{
		Query:       ".metadata.name",
		FieldGroups: fieldGroups,
		Scope:       api.MAPPINGSCOPE_TENANT,
		ArgsInfo:    map[string]api.MappingFieldType{},
		ArgsEnum:    map[string][]string{},
	}
}

func validateMappingChangeForTest(ctx context.Context, mappingScope api.MappingScope, specSelName string, sel api.SpecSelectorsWithScopeValue) *api.Error {
	v := NewMappingConsistencyValidator(ctx, mappingScope)
	return v.ValidateMappingChange(ctx, runaidb.DB().WithContext(ctx), map[string]api.SpecSelectorsWithScopeValue{
		specSelName: sel,
	})
}
