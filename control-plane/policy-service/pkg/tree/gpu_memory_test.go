/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/scopes"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGPUMemoryPolicyPatchAndReset(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		old, current := "gpuPortionRequest", "gpuMemoryPortionRequest"
		if reverse {
			old, current = current, old
		}
		t.Run(old, func(t *testing.T) {
			existing := &dbmodel.PolicyWithOptionalEffective{Policy: dbmodel.Policy{
				DefaultsV2: dbmodels.JsonBMap{"compute": map[string]any{old: json.Number("0.25")}},
				RulesV2:    dbmodels.JsonBMap{"compute": map[string]any{old: map[string]any{"canEdit": false}}},
			}}
			request := &ApplyPolicyRequest{Defaults: map[string]any{"compute": map[string]any{current: json.Number("0.5")}}}
			if err := applyRequestedChangesOnGivenPolicy(existing, request); err != nil {
				t.Fatal(err)
			}
			if got := existing.Policy.DefaultsV2["compute"].(map[string]any); len(got) != 1 || got["gpuMemoryPortionRequest"] != json.Number("0.5") {
				t.Fatalf("stale alias after patch: %#v", got)
			}
			stored, err := existing.Policy.ForDB()
			if err != nil {
				t.Fatal(err)
			}
			if stored.DefaultsV2["compute"].(map[string]any)["gpuPortionRequest"] != json.Number("0.5") {
				t.Fatal("DB key changed")
			}
			if err := applyRequestedChangesOnGivenPolicy(existing, &ApplyPolicyRequest{FieldsToReset: []string{"compute." + old}}); err != nil {
				t.Fatal(err)
			}
			if len(existing.Policy.DefaultsV2) != 0 || len(existing.Policy.RulesV2) != 0 {
				t.Fatalf("reset left aliases: %#v", existing.Policy)
			}
		})
	}
}

func TestGPUMemoryPolicyInheritance(t *testing.T) {
	for _, parentKey := range []string{"gpuPortionRequest", "gpuMemoryPortionRequest"} {
		t.Run(parentKey, func(t *testing.T) {
			childKey := "gpuMemoryPortionRequest"
			if parentKey == childKey {
				childKey = "gpuPortionRequest"
			}
			parent := &dbmodel.PolicyWithOptionalEffective{Policy: dbmodel.Policy{RulesV2: dbmodels.JsonBMap{"compute": map[string]any{parentKey: map[string]any{"canEdit": false}}}}}
			child := &dbmodel.PolicyWithOptionalEffective{Policy: dbmodel.Policy{RulesV2: dbmodels.JsonBMap{"compute": map[string]any{childKey: map[string]any{"canEdit": true}}}}}
			for _, p := range []*dbmodel.PolicyWithOptionalEffective{parent, child} {
				if err := p.NormalizeGPUMemory(); err != nil {
					t.Fatal(err)
				}
			}
			ht := &HierarchyTree{}
			parentSource, err := (scopes.OrgUnit{Scope: api.SCOPETYPE_TENANT, Id: "1"}).AsSourceOfRule("Workspace")
			if err != nil {
				t.Fatal(err)
			}
			childSource, err := (scopes.OrgUnit{Scope: api.SCOPETYPE_PROJECT, Id: "2"}).AsSourceOfRule("Workspace")
			if err != nil {
				t.Fatal(err)
			}
			ht.InitSourceOfRules(parent.Policy.RulesV2, parentSource, false)
			inherited := ht.DetermineSourceOfRules(child.Policy.RulesV2, parent.Policy.RulesV2)
			rules, mergeErr := ht.MergeRules(parent.Policy.RulesV2, child.Policy.RulesV2, inherited, childSource)
			if mergeErr != nil {
				t.Fatal(mergeErr)
			}
			rule := rules["compute"].(map[string]any)["gpuMemoryPortionRequest"].(map[string]any)
			if rule["canEdit"] != false {
				t.Fatal("child alias bypassed parent lock")
			}
			if !reflect.DeepEqual(rule["sourceOfRule"], parentSource) {
				t.Fatal("lost source of inherited rule")
			}
		})
	}
}

func TestLegacyWholePolicyCanBeNormalizedAndPatched(t *testing.T) {
	request := &ApplyPolicyRequest{
		Defaults: map[string]any{"compute": map[string]any{"gpuPortionRequest": json.Number("4")}},
		Rules:    map[string]any{"compute": map[string]any{"gpuPortionRequest": map[string]any{"canEdit": false, "required": true}}},
	}
	if err := gpucompat.ValidateMemoryRequestInput(request.Defaults); err != nil {
		t.Fatal(err)
	}
	if err := normalizeGPUPolicyRequest(request); err != nil {
		t.Fatal(err)
	}
	existing := &dbmodel.PolicyWithOptionalEffective{}
	if err := applyRequestedChangesOnGivenPolicy(existing, request); err != nil {
		t.Fatal("legacy request rejected after normalization", err)
	}
	stored, err := existing.Policy.ForDB()
	if err != nil {
		t.Fatal(err)
	}
	if stored.DefaultsV2["compute"].(map[string]any)["gpuPortionRequest"] != json.Number("4") {
		t.Fatal("lost whole GPU default")
	}
	if stored.RulesV2["compute"].(map[string]any)["gpuPortionRequest"].(map[string]any)["canEdit"] != false {
		t.Fatal("lost lock")
	}
}

func TestGPUMemoryAliasesDuringRecursivePolicyCalculation(t *testing.T) {
	for _, parentKey := range []string{"gpuPortionRequest", "gpuMemoryPortionRequest"} {
		t.Run(parentKey, func(t *testing.T) {
			childKey := "gpuMemoryPortionRequest"
			if parentKey == childKey {
				childKey = "gpuPortionRequest"
			}
			parent := &dbmodel.PolicyWithOptionalEffective{Policy: dbmodel.Policy{
				BaseModel: dbmodels.BaseModel{ID: uuid.New()}, TenantId: 1,
				Scope: api.SCOPETYPE_TENANT, WorkloadType: string(WorkloadTypeWorkspace),
				DefaultsV2: dbmodels.JsonBMap{"compute": map[string]any{parentKey: .25}},
				RulesV2:    dbmodels.JsonBMap{"compute": map[string]any{parentKey: map[string]any{"canEdit": true}}},
			}}
			child := &dbmodel.PolicyWithOptionalEffective{Policy: dbmodel.Policy{
				BaseModel: dbmodels.BaseModel{ID: uuid.New()}, TenantId: 1,
				Scope: api.SCOPETYPE_PROJECT, ProjectId: convutils.Ref(int32(2)), WorkloadType: string(WorkloadTypeWorkspace),
				DefaultsV2: dbmodels.JsonBMap{"compute": map[string]any{childKey: .5}},
			}}
			db, mock := dbtesting.NewMocker(t)
			for _, policy := range []*dbmodel.PolicyWithOptionalEffective{parent, child} {
				mock.ExpectQuery("SELECT * FROM effective_policies WHERE effective_policies.policy_id = ? ORDER BY effective_policies.policy_id LIMIT ?", policy.ID, 1).
					Returns([]dbmodel.EffectivePolicy{})
				mock.ExpectTransaction().Begin()
				mock.ExpectExec("INSERT INTO effective_policies (policy_id,tenant_id,defaults_v2,rules_v2,imposed_assets,updated_at,update_count,default_from_order,validation_error,deleted_at) VALUES (?,?,?,?,?,?,?,?,?,?)").Returns(1)
				mock.ExpectTransaction().Commit()
			}
			ctx := context.Background()
			policies := map[string]*dbmodel.PolicyWithOptionalEffective{parent.ID.String(): parent, child.ID.String(): child}
			effectiveContext := NewEffectiveContext(ctx, WorkloadTypeWorkspace, []scopes.OrgUnit{child.OrgUnit(), parent.OrgUnit()}, policies)
			ctx = context.WithValue(ctx, EffectiveContextKey, effectiveContext)
			ht := NewHierarchyTree(1, WorkloadTypeWorkspace)
			result, err := ht.calculateEffectivePolicy(ctx, db, child.ID.String(), 0, nil, nil, nil, nil)
			require.Nil(t, err)
			require.Equal(t, map[string]any{"gpuMemoryPortionRequest": .5}, result.DefaultsV2["compute"])
			require.Equal(t, map[string]any{"gpuMemoryPortionRequest": .25}, parent.EffectivePolicy.DefaultsV2["compute"])
			require.NotContains(t, result.RulesV2["compute"], "gpuPortionRequest")
			require.Contains(t, result.RulesV2["compute"], "gpuMemoryPortionRequest")
		})
	}
}
