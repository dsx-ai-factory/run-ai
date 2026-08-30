package resources

// Holds the resourceQuota type and everything that builds one from the org-unit
// schema — unit conversion, sentinel stripping, and nullable accessors — so the
// tool file stays focused on request handling rather than the upstream quirks
// these untangle.

import (
	orgunit "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type resourceQuota struct {
	Deserved        *float64 `json:"deserved"        jsonschema:"Configured quota assigned to this project on this node pool. Null when the resource is not metered."`
	Limit           *float64 `json:"limit"           jsonschema:"Maximum amount the project can consume on this node pool, including over-quota. Null when unlimited or not set."`
	OverQuotaWeight *float64 `json:"overQuotaWeight" jsonschema:"Fair-share weight (1-20) when projects compete for over-quota resources within the same rank. Null when not set."`
	Allocated       *float64 `json:"allocated"       jsonschema:"Currently allocated amount on this cell from the upstream quotaStatus. Null when the quota status is not populated upstream."`
}

func normalizeQuota(q resourceQuota) resourceQuota {
	return resourceQuota{
		Deserved:        round3(nilIfNegative(q.Deserved)),
		Limit:           round3(nilIfNegative(q.Limit)),
		OverQuotaWeight: round3(nilIfNonPositive(q.OverQuotaWeight)),
		Allocated:       round3(q.Allocated),
	}
}

func nilIfNegative(v *float64) *float64 {
	if v == nil || *v < 0 {
		return nil
	}
	return v
}

func nilIfNonPositive(v *float64) *float64 {
	if v == nil || *v <= 0 {
		return nil
	}
	return v
}

func round3(v *float64) *float64 {
	if v == nil {
		return nil
	}
	r := common.RoundTo(*v, 3)
	return &r
}

func scaleByDivisor(value *float64, divisor float64) *float64 {
	if value == nil {
		return nil
	}
	scaled := *value / divisor
	return &scaled
}

func convertToGiB(value float64, units orgunit.Units) float64 {
	return common.MemoryToGiB(value, units)
}

func nullableFloatValue(n orgunit.NullableFloat64) *float64 {
	if !n.IsSet() {
		return nil
	}
	return n.Get()
}

func nullableCPU(n orgunit.NullableResourcesCpu) (*orgunit.ResourcesCpu, bool) {
	if !n.IsSet() {
		return nil, false
	}
	v := n.Get()
	if v == nil {
		return nil, false
	}
	return v, true
}

func nullableMemory(n orgunit.NullableResourcesMemory) (*orgunit.ResourcesMemory, bool) {
	if !n.IsSet() {
		return nil, false
	}
	v := n.Get()
	if v == nil {
		return nil, false
	}
	return v, true
}

func gpuDeserved(r orgunit.ResourcesPerNodePool) *float64 {
	if r.Gpu == nil {
		return nil
	}
	return nullableFloatValue(r.Gpu.Deserved)
}

func gpuLimit(r orgunit.ResourcesPerNodePool) *float64 {
	if r.Gpu == nil {
		return nil
	}
	return nullableFloatValue(r.Gpu.Limit)
}

func gpuOverQuotaWeight(r orgunit.ResourcesPerNodePool) *float64 {
	if r.Gpu == nil {
		return nil
	}
	return nullableFloatValue(r.Gpu.OverQuotaWeight)
}

func cpuCoresDeserved(r orgunit.ResourcesPerNodePool) *float64 {
	cpu, ok := nullableCPU(r.Cpu)
	if !ok {
		return nil
	}
	return scaleByDivisor(nullableFloatValue(cpu.Deserved), common.MillicoresPerCore)
}

func cpuCoresLimit(r orgunit.ResourcesPerNodePool) *float64 {
	cpu, ok := nullableCPU(r.Cpu)
	if !ok {
		return nil
	}
	return scaleByDivisor(nullableFloatValue(cpu.Limit), common.MillicoresPerCore)
}

func cpuOverQuotaWeight(r orgunit.ResourcesPerNodePool) *float64 {
	cpu, ok := nullableCPU(r.Cpu)
	if !ok {
		return nil
	}
	return nullableFloatValue(cpu.OverQuotaWeight)
}

func memoryGiBDeserved(r orgunit.ResourcesPerNodePool) *float64 {
	mem, ok := nullableMemory(r.Memory)
	if !ok {
		return nil
	}
	v := nullableFloatValue(mem.Deserved)
	if v == nil {
		return nil
	}
	gib := convertToGiB(*v, mem.Units)
	return &gib
}

func memoryGiBLimit(r orgunit.ResourcesPerNodePool) *float64 {
	mem, ok := nullableMemory(r.Memory)
	if !ok {
		return nil
	}
	v := nullableFloatValue(mem.Limit)
	if v == nil {
		return nil
	}
	gib := convertToGiB(*v, mem.Units)
	return &gib
}

func memoryOverQuotaWeight(r orgunit.ResourcesPerNodePool) *float64 {
	mem, ok := nullableMemory(r.Memory)
	if !ok {
		return nil
	}
	return nullableFloatValue(mem.OverQuotaWeight)
}
