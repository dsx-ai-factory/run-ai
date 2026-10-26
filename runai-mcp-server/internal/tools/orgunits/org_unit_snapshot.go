/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orgunits

import (
	orgunit "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

func projectSnapshot(p *orgunit.Project) map[string]any {
	if p == nil {
		return nil
	}
	snap := map[string]any{
		"id":                    p.Id,
		"name":                  p.Name,
		"clusterId":             p.ClusterId,
		"description":           p.Description,
		"parentId":              p.ParentId,
		"enforceRunaiScheduler": p.EnforceRunaiScheduler,
		"quota":                 quotaFromResources(p.Resources),
	}
	if ns := p.RequestedNamespace.Get(); ns != nil {
		snap["requestedNamespace"] = *ns
	}
	if p.DefaultNodePools != nil {
		snap["defaultNodePools"] = p.DefaultNodePools
	}
	if v := p.Status.Phase; v != nil {
		snap["phase"] = string(*v)
	}
	addSchedulingRulesSnapshot(snap, p.SchedulingRules.Get())
	return snap
}

func departmentSnapshot(d *orgunit.Department) map[string]any {
	if d == nil {
		return nil
	}
	snap := map[string]any{
		"id":          d.Id,
		"name":        d.Name,
		"clusterId":   d.ClusterId,
		"description": d.Description,
		"childCount":  len(d.Children),
		"quota":       quotaFromResources(d.Resources),
	}
	if d.DefaultNodePools != nil {
		snap["defaultNodePools"] = d.DefaultNodePools
	}
	addSchedulingRulesSnapshot(snap, d.SchedulingRules.Get())
	return snap
}

func quotaFromResources(resources []orgunit.Resources) []map[string]any {
	out := make([]map[string]any, 0, len(resources))
	for _, r := range resources {
		q := map[string]any{}
		if np := r.NodePool.Get(); np != nil {
			q["nodePoolId"] = np.Id
			if np.Name != nil {
				q["nodePoolName"] = *np.Name
			}
		}
		q["gpuDeserved"] = r.Gpu.Deserved
		putFloat(q, "gpuLimit", r.Gpu.Limit.Get())
		putFloat(q, "gpuOverQuotaWeight", r.Gpu.OverQuotaWeight.Get())
		putCPUMemory(q, r.Cpu.Get(), r.Memory.Get())
		if v := r.Rank.Get(); v != nil {
			q["rank"] = *v
		}
		out = append(out, q)
	}
	return out
}

func quotaFromNullable(resources []orgunit.ResourcesNullable) []map[string]any {
	out := make([]map[string]any, 0, len(resources))
	for _, r := range resources {
		q := map[string]any{}
		if np := r.NodePool.Get(); np != nil {
			q["nodePoolId"] = np.Id
			if np.Name != nil {
				q["nodePoolName"] = *np.Name
			}
		}
		if g := r.Gpu; g != nil {
			putFloat(q, "gpuDeserved", g.Deserved)
			putFloat(q, "gpuLimit", g.Limit.Get())
			putFloat(q, "gpuOverQuotaWeight", g.OverQuotaWeight.Get())
		}
		putCPUMemory(q, r.Cpu.Get(), r.Memory.Get())
		if v := r.Rank.Get(); v != nil {
			q["rank"] = *v
		}
		out = append(out, q)
	}
	return out
}

// putCPUMemory writes CPU (cores) and memory (GiB) into a quota cell, converting from the
// upstream millicores / mixed memory units so the snapshot matches the input boundary.
func putCPUMemory(q map[string]any, c *orgunit.ResourcesCpu, m *orgunit.ResourcesMemory) {
	if c != nil {
		putFloatScaled(q, "cpuDeserved", c.Deserved.Get(), common.MillicoresToCores)
		putFloatScaled(q, "cpuLimit", c.Limit.Get(), common.MillicoresToCores)
		putFloat(q, "cpuOverQuotaWeight", c.OverQuotaWeight.Get())
	}
	if m != nil {
		putMemory(q, "memoryDeserved", m.Deserved.Get(), m.Units)
		putMemory(q, "memoryLimit", m.Limit.Get(), m.Units)
		putFloat(q, "memoryOverQuotaWeight", m.OverQuotaWeight.Get())
	}
}

func putFloat(m map[string]any, key string, v *float64) {
	if v != nil {
		m[key] = *v
	}
}

func putFloatScaled(m map[string]any, key string, v *float64, conv func(float64) float64) {
	if v != nil {
		m[key] = conv(*v)
	}
}

func putMemory(m map[string]any, key string, v *float64, units orgunit.Units) {
	if v != nil {
		m[key] = common.MemoryToGiB(*v, units)
	}
}

func addSchedulingRulesSnapshot(snap map[string]any, sr *orgunit.SchedulingRules) {
	if sr == nil {
		return
	}
	if m := schedulingRulesMap(sr); len(m) > 0 {
		snap["schedulingRules"] = m
	}
}

func schedulingRulesMap(sr *orgunit.SchedulingRules) map[string]any {
	m := map[string]any{}
	putInt(m, "interactiveJobTimeLimitSeconds", sr.InteractiveJobTimeLimitSeconds.Get())
	putInt(m, "interactiveJobMaxIdleDurationSeconds", sr.InteractiveJobMaxIdleDurationSeconds.Get())
	putInt(m, "interactiveJobPreemptIdleDurationSeconds", sr.InteractiveJobPreemptIdleDurationSeconds.Get())
	putInt(m, "trainingJobMaxIdleDurationSeconds", sr.TrainingJobMaxIdleDurationSeconds.Get())
	putInt(m, "trainingJobTimeLimitSeconds", sr.TrainingJobTimeLimitSeconds.Get())
	return m
}

func putInt(m map[string]any, key string, v *int32) {
	if v != nil {
		m[key] = *v
	}
}

func createPreview(in manageOrgUnitInput, resources []orgunit.ResourcesNullable) map[string]any {
	snap := map[string]any{
		"name":      in.Name,
		"clusterId": in.ClusterID,
		"quota":     quotaFromNullable(resources),
	}
	addSpecPreviewFields(snap, in.Spec)
	return snap
}

func updatePreview(before map[string]any, spec *orgUnitSpec, resources []orgunit.ResourcesNullable) map[string]any {
	after := copyMap(before)
	addSpecPreviewFields(after, spec)
	after["quota"] = quotaFromNullable(resources)
	return after
}

func addSpecPreviewFields(snap map[string]any, spec *orgUnitSpec) {
	if spec == nil {
		return
	}
	if spec.Description != nil {
		snap["description"] = *spec.Description
	}
	if spec.ParentID != nil {
		snap["parentId"] = *spec.ParentID
	}
	if spec.RequestedNamespace != nil {
		snap["requestedNamespace"] = *spec.RequestedNamespace
	}
	if spec.EnforceRunaiScheduler != nil {
		snap["enforceRunaiScheduler"] = *spec.EnforceRunaiScheduler
	}
	if spec.DefaultNodePools != nil {
		snap["defaultNodePools"] = spec.DefaultNodePools
	}
	if spec.SchedulingRules != nil {
		base, _ := snap["schedulingRules"].(map[string]any)
		snap["schedulingRules"] = overlaySchedulingRules(base, spec.SchedulingRules)
	}
}

// overlaySchedulingRules previews the merged rules: the current map with the caller's set
// fields applied on top (mirroring mergeSchedulingRules on the write path).
func overlaySchedulingRules(base map[string]any, in *orgUnitSchedulingRules) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	putInt(out, "interactiveJobTimeLimitSeconds", in.InteractiveJobTimeLimitSeconds)
	putInt(out, "interactiveJobMaxIdleDurationSeconds", in.InteractiveJobMaxIdleDurationSeconds)
	putInt(out, "interactiveJobPreemptIdleDurationSeconds", in.InteractiveJobPreemptIdleDuration)
	putInt(out, "trainingJobMaxIdleDurationSeconds", in.TrainingJobMaxIdleDurationSeconds)
	putInt(out, "trainingJobTimeLimitSeconds", in.TrainingJobTimeLimitSeconds)
	return out
}

func copyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
