/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orgunits

import (
	"context"
	"fmt"

	authz "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	orgunit "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

func authorizeOrgUnit(ctx context.Context, client *runaiclient.Client, orgType string, action common.WriteAction, clusterID, orgUnitID, parentDeptID string) error {
	scope, err := common.PermittedScopeForAction(ctx, client, orgUnitResourceType(orgType), action.Authz())
	if err != nil {
		return err
	}
	if common.ScopeAllowsOrgUnit(scope, orgType, clusterID, orgUnitID, parentDeptID) {
		return nil
	}
	return fmt.Errorf("not permitted to %s %s on cluster %q: %w", action, orgType, clusterID, runaimcp.ErrForbidden)
}

func orgUnitResourceType(orgType string) authz.ResourceType {
	if orgType == common.OrgTypeProject {
		return authz.RESOURCETYPE_PROJECT
	}
	return authz.RESOURCETYPE_DEPARTMENT
}

// buildResources merges the caller's quota lines over current (nil on create) by node
// pool, so a PUT that omits a node pool preserves its existing quota instead of dropping it.
func buildResources(ctx context.Context, client *runaiclient.Client, clusterID string, quota []orgUnitQuota, current []orgunit.Resources) ([]orgunit.ResourcesNullable, error) {
	byNodePool := map[string]orgunit.ResourcesNullable{}
	order := []string{}
	for _, r := range current {
		id := nodePoolID(r.NodePool)
		if id == "" {
			continue
		}
		byNodePool[id] = currentToNullable(r)
		order = append(order, id)
	}

	if len(quota) == 0 {
		resources := make([]orgunit.ResourcesNullable, 0, len(order))
		for _, id := range order {
			resources = append(resources, byNodePool[id])
		}
		return resources, nil
	}

	names := make([]string, 0, len(quota))
	for _, q := range quota {
		names = append(names, q.NodePoolName)
	}
	ids, err := common.ResolveNodePoolIDs(ctx, client, clusterID, names)
	if err != nil {
		return nil, err
	}

	for _, q := range quota {
		id := ids[q.NodePoolName]
		base, exists := byNodePool[id]
		if !exists {
			// A node pool not already in the org-unit's quota has nothing to carry forward,
			// so gpuDeserved (the one required field) must be supplied explicitly.
			if q.GPUDeserved == nil {
				return nil, fmt.Errorf("spec.quota for node pool %q: gpuDeserved is required when adding a node pool not already in the quota: %w",
					q.NodePoolName, runaimcp.ErrValidation)
			}
			base = *orgunit.NewResourcesNullable()
			base.SetNodePool(*orgunit.NewResourcesNodePool(id))
			order = append(order, id)
		}
		applyQuota(&base, q)
		byNodePool[id] = base
	}

	resources := make([]orgunit.ResourcesNullable, 0, len(order))
	for _, id := range order {
		resources = append(resources, byNodePool[id])
	}
	return resources, nil
}

// applyQuota overlays a caller quota line onto r (a carried-forward or fresh cell). It
// rebuilds only the resources the caller mentions, leaving the others as carried. Within
// a rebuilt resource, deserved/limit fall back to the current values, but overQuotaWeight
// is only set when the caller provides it — carrying a stale weight forward breaks the
// "weight must equal deserved or be empty" rule when the over-quota feature is off and
// deserved changes. CPU is cores and memory is GiB at the boundary, converted to the
// upstream millicores / MiB here.
func applyQuota(r *orgunit.ResourcesNullable, q orgUnitQuota) {
	if q.GPUDeserved != nil || q.GPULimit != nil || q.GPUOverQuotaWeight != nil {
		prev, _ := r.GetGpuOk()
		gpu := orgunit.NewResourcesNullableGpu()
		switch {
		case q.GPUDeserved != nil:
			gpu.SetDeserved(*q.GPUDeserved)
		case prev != nil && prev.Deserved != nil:
			gpu.SetDeserved(*prev.Deserved)
		}
		switch {
		case q.GPULimit != nil:
			gpu.SetLimit(*q.GPULimit)
		case prev != nil:
			copyNullableFloat(prev.Limit, gpu.SetLimit, gpu.SetLimitNil)
		}
		setFloat(q.GPUOverQuotaWeight, gpu.SetOverQuotaWeight)
		r.SetGpu(*gpu)
	}

	if q.CPUDeserved != nil || q.CPULimit != nil || q.CPUOverQuotaWeight != nil {
		prev, _ := r.GetCpuOk()
		cpu := orgunit.NewResourcesCpu()
		switch {
		case q.CPUDeserved != nil:
			cpu.SetDeserved(common.CoresToMillicores(*q.CPUDeserved))
		case prev != nil:
			copyNullableFloat(prev.Deserved, cpu.SetDeserved, cpu.SetDeservedNil)
		}
		switch {
		case q.CPULimit != nil:
			cpu.SetLimit(common.CoresToMillicores(*q.CPULimit))
		case prev != nil:
			copyNullableFloat(prev.Limit, cpu.SetLimit, cpu.SetLimitNil)
		}
		setFloat(q.CPUOverQuotaWeight, cpu.SetOverQuotaWeight)
		r.SetCpu(*cpu)
	}

	if q.MemoryDeserved != nil || q.MemoryLimit != nil || q.MemoryOverQuotaWeight != nil {
		prev, _ := r.GetMemoryOk()
		mem := orgunit.NewResourcesMemory(orgunit.UNITS_MIB)
		setMemoryField(q.MemoryDeserved, prev, memDeservedGiB, mem.SetDeserved, mem.SetDeservedNil)
		setMemoryField(q.MemoryLimit, prev, memLimitGiB, mem.SetLimit, mem.SetLimitNil)
		setFloat(q.MemoryOverQuotaWeight, mem.SetOverQuotaWeight)
		r.SetMemory(*mem)
	}

	if q.Rank != "" {
		r.SetRank(q.Rank)
	}
}

func setFloat(v *float64, set func(float64)) {
	if v != nil {
		set(*v)
	}
}

// setMemoryField writes one memory field in MiB: the caller's GiB value, or the current
// value (converted from its own units to GiB) when the caller omits it.
func setMemoryField(callerGiB *float64, prev *orgunit.ResourcesMemory, prevGiB func(*orgunit.ResourcesMemory) (float64, bool, bool), set func(float64), setNil func()) {
	if callerGiB != nil {
		set(common.GiBToMiB(*callerGiB))
		return
	}
	if prev == nil {
		return
	}
	if giB, isSet, isNull := prevGiB(prev); isNull {
		setNil()
	} else if isSet {
		set(common.GiBToMiB(giB))
	}
}

func memDeservedGiB(m *orgunit.ResourcesMemory) (float64, bool, bool) {
	return nullableGiB(m.Deserved, m.Units)
}
func memLimitGiB(m *orgunit.ResourcesMemory) (float64, bool, bool) {
	return nullableGiB(m.Limit, m.Units)
}

func nullableGiB(n orgunit.NullableFloat64, units orgunit.Units) (giB float64, isSet, isNull bool) {
	if !n.IsSet() {
		return 0, false, false
	}
	v := n.Get()
	if v == nil {
		return 0, true, true
	}
	return common.MemoryToGiB(*v, units), true, false
}

func currentToNullable(r orgunit.Resources) orgunit.ResourcesNullable {
	out := orgunit.NewResourcesNullable()
	if id := nodePoolID(r.NodePool); id != "" {
		np := orgunit.NewResourcesNodePool(id)
		if name, ok := r.NodePool.Get().GetNameOk(); ok {
			np.SetName(*name)
		}
		out.SetNodePool(*np)
	}

	gpu := orgunit.NewResourcesNullableGpu()
	gpu.SetDeserved(r.Gpu.Deserved)
	copyNullableFloat(r.Gpu.Limit, gpu.SetLimit, gpu.SetLimitNil)
	copyNullableFloat(r.Gpu.OverQuotaWeight, gpu.SetOverQuotaWeight, gpu.SetOverQuotaWeightNil)
	out.SetGpu(*gpu)

	if cur, ok := r.Cpu.Get(), r.Cpu.IsSet(); ok && cur != nil {
		cpu := orgunit.NewResourcesCpu()
		copyNullableFloat(cur.Deserved, cpu.SetDeserved, cpu.SetDeservedNil)
		copyNullableFloat(cur.Limit, cpu.SetLimit, cpu.SetLimitNil)
		copyNullableFloat(cur.OverQuotaWeight, cpu.SetOverQuotaWeight, cpu.SetOverQuotaWeightNil)
		out.SetCpu(*cpu)
	}
	if cur, ok := r.Memory.Get(), r.Memory.IsSet(); ok && cur != nil {
		mem := orgunit.NewResourcesMemory(cur.Units)
		copyNullableFloat(cur.Deserved, mem.SetDeserved, mem.SetDeservedNil)
		copyNullableFloat(cur.Limit, mem.SetLimit, mem.SetLimitNil)
		copyNullableFloat(cur.OverQuotaWeight, mem.SetOverQuotaWeight, mem.SetOverQuotaWeightNil)
		out.SetMemory(*mem)
	}
	if v := r.Rank.Get(); v != nil {
		out.SetRank(*v)
	}
	return *out
}

func copyNullableFloat(src orgunit.NullableFloat64, set func(float64), setNil func()) {
	if !src.IsSet() {
		return
	}
	if v := src.Get(); v != nil {
		set(*v)
		return
	}
	setNil()
}

func nodePoolID(np orgunit.NullableResourcesNodePool) string {
	if v := np.Get(); v != nil {
		return v.Id
	}
	return ""
}

func applyProjectCreateSpec(req *orgunit.ProjectCreationRequest, spec *orgUnitSpec) {
	if spec == nil {
		return
	}
	if spec.Description != nil {
		req.SetDescription(*spec.Description)
	}
	if spec.ParentID != nil {
		req.SetParentId(*spec.ParentID)
	}
	if spec.RequestedNamespace != nil {
		req.SetRequestedNamespace(*spec.RequestedNamespace)
	}
	if spec.EnforceRunaiScheduler != nil {
		req.SetEnforceRunaiScheduler(*spec.EnforceRunaiScheduler)
	} else {
		// The constructor injects enforceRunaiScheduler=true; clear it when the caller
		// omits the field so the cluster's default applies instead of a forced true.
		req.SetEnforceRunaiSchedulerNil()
	}
	if spec.DefaultNodePools != nil {
		req.SetDefaultNodePools(spec.DefaultNodePools)
	}
	if spec.SchedulingRules != nil {
		req.SetSchedulingRules(mergeSchedulingRules(spec.SchedulingRules, nil))
	}
}

func applyProjectUpdateSpec(req *orgunit.ProjectUpdateRequest, spec *orgUnitSpec, current *orgunit.Project) {
	if spec.Description != nil {
		req.SetDescription(*spec.Description)
	} else {
		req.SetDescription(current.Description)
	}
	if spec.EnforceRunaiScheduler != nil {
		req.SetEnforceRunaiScheduler(*spec.EnforceRunaiScheduler)
	} else {
		req.SetEnforceRunaiScheduler(current.EnforceRunaiScheduler)
	}
	if spec.DefaultNodePools != nil {
		req.SetDefaultNodePools(spec.DefaultNodePools)
	} else if current.DefaultNodePools != nil {
		req.SetDefaultNodePools(current.DefaultNodePools)
	}
	if cur, _ := current.GetSchedulingRulesOk(); spec.SchedulingRules != nil || cur != nil {
		req.SetSchedulingRules(mergeSchedulingRules(spec.SchedulingRules, cur))
	}
	// The MCP does not expose nodeTypes; carry it forward so an unrelated update (e.g. a
	// description change) doesn't wipe configured node types — the update is a full replace.
	if current.NodeTypes != nil {
		req.NodeTypes = current.NodeTypes
	}
}

func applyDepartmentCreateSpec(req *orgunit.DepartmentCreationRequest, spec *orgUnitSpec) {
	if spec == nil {
		return
	}
	if spec.Description != nil {
		req.SetDescription(*spec.Description)
	}
	if spec.DefaultNodePools != nil {
		req.SetDefaultNodePools(spec.DefaultNodePools)
	}
	if spec.SchedulingRules != nil {
		req.SetSchedulingRules(mergeSchedulingRules(spec.SchedulingRules, nil))
	}
}

func applyDepartmentUpdateSpec(req *orgunit.DepartmentUpdateRequest, spec *orgUnitSpec, current *orgunit.Department) {
	if spec.Description != nil {
		req.SetDescription(*spec.Description)
	} else {
		req.SetDescription(current.Description)
	}
	if spec.DefaultNodePools != nil {
		req.SetDefaultNodePools(spec.DefaultNodePools)
	} else if current.DefaultNodePools != nil {
		req.SetDefaultNodePools(current.DefaultNodePools)
	}
	if cur, _ := current.GetSchedulingRulesOk(); spec.SchedulingRules != nil || cur != nil {
		req.SetSchedulingRules(mergeSchedulingRules(spec.SchedulingRules, cur))
	}
	if current.NodeTypes != nil {
		req.NodeTypes = current.NodeTypes
	}
}

// mergeSchedulingRules overlays the caller's set fields onto the current rules so a
// partial update leaves the untouched limits in place (PUT replaces the whole object).
func mergeSchedulingRules(in *orgUnitSchedulingRules, current *orgunit.SchedulingRules) orgunit.SchedulingRules {
	sr := orgunit.NewSchedulingRules()
	if current != nil {
		*sr = *current
	}
	if in != nil {
		setInt32(in.InteractiveJobTimeLimitSeconds, sr.SetInteractiveJobTimeLimitSeconds)
		setInt32(in.InteractiveJobMaxIdleDurationSeconds, sr.SetInteractiveJobMaxIdleDurationSeconds)
		setInt32(in.InteractiveJobPreemptIdleDuration, sr.SetInteractiveJobPreemptIdleDurationSeconds)
		setInt32(in.TrainingJobMaxIdleDurationSeconds, sr.SetTrainingJobMaxIdleDurationSeconds)
		setInt32(in.TrainingJobTimeLimitSeconds, sr.SetTrainingJobTimeLimitSeconds)
	}
	return *sr
}

func setInt32(v *int32, set func(int32)) {
	if v != nil {
		set(*v)
	}
}
