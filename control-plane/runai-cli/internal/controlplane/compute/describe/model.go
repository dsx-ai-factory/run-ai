package describe

import (
	"time"

	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

// ComputeAsset Output types for JSON/YAML with omitempty tags
type ComputeAsset struct {
	Meta   AssetMeta       `json:"meta" yaml:"meta"`
	Spec   ComputeSpec     `json:"spec" yaml:"spec"`
	UsedBy *AssetUsageInfo `json:"usedBy,omitempty" yaml:"usedBy,omitempty"`
}

type AssetMeta struct {
	Name         string     `json:"name" yaml:"name"`
	Description  string     `json:"description,omitempty" yaml:"description,omitempty"`
	Scope        string     `json:"scope" yaml:"scope"`
	ClusterId    string     `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	DepartmentId string     `json:"departmentId,omitempty" yaml:"departmentId,omitempty"`
	ProjectId    *int32     `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	ProjectName  string     `json:"projectName,omitempty" yaml:"projectName,omitempty"`
	TenantId     *int32     `json:"tenantId,omitempty" yaml:"tenantId,omitempty"`
	Id           string     `json:"id" yaml:"id"`
	Kind         string     `json:"kind" yaml:"kind"`
	CreatedBy    string     `json:"createdBy" yaml:"createdBy"`
	CreatedAt    time.Time  `json:"createdAt" yaml:"createdAt"`
	UpdatedBy    string     `json:"updatedBy" yaml:"updatedBy"`
	UpdatedAt    time.Time  `json:"updatedAt" yaml:"updatedAt"`
	DeletedAt    *time.Time `json:"deletedAt,omitempty" yaml:"deletedAt,omitempty"`
	DeletedBy    string     `json:"deletedBy,omitempty" yaml:"deletedBy,omitempty"`
	UpdateCount  *int64     `json:"updateCount,omitempty" yaml:"updateCount,omitempty"`
}

type ComputeSpec struct {
	CpuCoreLimit      *float64           `json:"cpuCoreLimit,omitempty" yaml:"cpuCoreLimit,omitempty"`
	CpuCoreRequest    *float64           `json:"cpuCoreRequest,omitempty" yaml:"cpuCoreRequest,omitempty"`
	CpuMemoryLimit    string             `json:"cpuMemoryLimit,omitempty" yaml:"cpuMemoryLimit,omitempty"`
	CpuMemoryRequest  string             `json:"cpuMemoryRequest,omitempty" yaml:"cpuMemoryRequest,omitempty"`
	ExtendedResources []ExtendedResource `json:"extendedResources,omitempty" yaml:"extendedResources,omitempty"`
	GpuDevicesRequest *int32             `json:"gpuDevicesRequest,omitempty" yaml:"gpuDevicesRequest,omitempty"`
	GpuMemoryLimit    string             `json:"gpuMemoryLimit,omitempty" yaml:"gpuMemoryLimit,omitempty"`
	GpuMemoryRequest  string             `json:"gpuMemoryRequest,omitempty" yaml:"gpuMemoryRequest,omitempty"`
	GpuPortionLimit   *float64           `json:"gpuPortionLimit,omitempty" yaml:"gpuPortionLimit,omitempty"`
	GpuPortionRequest *float64           `json:"gpuPortionRequest,omitempty" yaml:"gpuPortionRequest,omitempty"`
	GpuRequestType    string             `json:"gpuRequestType,omitempty" yaml:"gpuRequestType,omitempty"`
	LargeShmRequest   *bool              `json:"largeShmRequest,omitempty" yaml:"largeShmRequest,omitempty"`
}

type ExtendedResource struct {
	Name     string `json:"name" yaml:"name"`
	Quantity string `json:"quantity" yaml:"quantity"`
}

type AssetUsageInfo struct {
	Workspaces  []WorkloadRefAndStatus `json:"workspaces,omitempty" yaml:"workspaces,omitempty"`
	Trainings   []WorkloadRefAndStatus `json:"trainings,omitempty" yaml:"trainings,omitempty"`
	Distributed []WorkloadRefAndStatus `json:"distributed,omitempty" yaml:"distributed,omitempty"`
	Inferences  []WorkloadRefAndStatus `json:"inferences,omitempty" yaml:"inferences,omitempty"`
	Templates   []AssetRef             `json:"templates,omitempty" yaml:"templates,omitempty"`
}

type WorkloadRefAndStatus struct {
	Name   string `json:"name" yaml:"name"`
	Id     string `json:"id" yaml:"id"`
	Status string `json:"status,omitempty" yaml:"status,omitempty"`
}

type AssetRef struct {
	Name string `json:"name" yaml:"name"`
	Id   string `json:"id" yaml:"id"`
}

// ToComputeAsset converts OpenAPI struct to output struct with omitempty tags
func ToComputeAsset(compute *openapi.ComputeAsset) *ComputeAsset {
	if compute == nil {
		return nil
	}

	result := &ComputeAsset{
		Meta: toAssetMeta(compute.Meta),
		Spec: toComputeSpec(compute.Spec),
	}

	// Add usedBy if available
	if compute.UsedBy.IsSet() && compute.UsedBy.Get() != nil {
		result.UsedBy = toAssetUsageInfo(compute.UsedBy.Get())
	}

	return result
}

func toAssetMeta(meta openapi.AssetMeta) AssetMeta {
	result := AssetMeta{
		Name:      meta.Name,
		Scope:     string(meta.Scope),
		Id:        meta.Id,
		Kind:      string(meta.Kind),
		CreatedBy: meta.CreatedBy,
		CreatedAt: meta.CreatedAt,
		UpdatedBy: meta.UpdatedBy,
		UpdatedAt: meta.UpdatedAt,
	}

	// Add optional fields only if they have values
	if meta.Description.IsSet() && meta.Description.Get() != nil {
		result.Description = *meta.Description.Get()
	}

	if meta.ClusterId.IsSet() && meta.ClusterId.Get() != nil {
		result.ClusterId = *meta.ClusterId.Get()
	}

	if meta.DepartmentId.IsSet() && meta.DepartmentId.Get() != nil {
		result.DepartmentId = *meta.DepartmentId.Get()
	}

	if meta.ProjectId.IsSet() && meta.ProjectId.Get() != nil {
		result.ProjectId = meta.ProjectId.Get()
	}

	if meta.HasProjectName() && meta.ProjectName.Get() != nil {
		result.ProjectName = *meta.ProjectName.Get()
	}

	if meta.TenantId.IsSet() && meta.TenantId.Get() != nil {
		result.TenantId = meta.TenantId.Get()
	}

	// Only include deletedAt if it's not the zero time
	if meta.DeletedAt != nil && !meta.DeletedAt.IsZero() {
		result.DeletedAt = meta.DeletedAt
	}

	if meta.DeletedBy != nil && *meta.DeletedBy != "" {
		result.DeletedBy = *meta.DeletedBy
	}

	if meta.UpdateCount != nil {
		result.UpdateCount = meta.UpdateCount
	}

	return result
}

func toComputeSpec(spec openapi.ComputeAssetSpec) ComputeSpec {
	result := ComputeSpec{}

	if spec.CpuCoreLimit.IsSet() && spec.CpuCoreLimit.Get() != nil {
		result.CpuCoreLimit = spec.CpuCoreLimit.Get()
	}

	if spec.CpuCoreRequest.IsSet() && spec.CpuCoreRequest.Get() != nil {
		result.CpuCoreRequest = spec.CpuCoreRequest.Get()
	}

	if spec.CpuMemoryLimit.IsSet() && spec.CpuMemoryLimit.Get() != nil {
		result.CpuMemoryLimit = *spec.CpuMemoryLimit.Get()
	}

	if spec.CpuMemoryRequest.IsSet() && spec.CpuMemoryRequest.Get() != nil {
		result.CpuMemoryRequest = *spec.CpuMemoryRequest.Get()
	}

	if len(spec.ExtendedResources) > 0 {
		result.ExtendedResources = make([]ExtendedResource, 0, len(spec.ExtendedResources))
		for _, er := range spec.ExtendedResources {
			extRes := ExtendedResource{}
			if er.Resource.IsSet() && er.Resource.Get() != nil {
				extRes.Name = *er.Resource.Get()
			}
			if er.Quantity.IsSet() && er.Quantity.Get() != nil {
				extRes.Quantity = *er.Quantity.Get()
			}
			result.ExtendedResources = append(result.ExtendedResources, extRes)
		}
	}

	if spec.GpuDevicesRequest.IsSet() && spec.GpuDevicesRequest.Get() != nil {
		result.GpuDevicesRequest = spec.GpuDevicesRequest.Get()
	}

	if spec.GpuMemoryLimit.IsSet() && spec.GpuMemoryLimit.Get() != nil {
		result.GpuMemoryLimit = *spec.GpuMemoryLimit.Get()
	}

	if spec.GpuMemoryRequest.IsSet() && spec.GpuMemoryRequest.Get() != nil {
		result.GpuMemoryRequest = *spec.GpuMemoryRequest.Get()
	}

	if spec.GpuPortionLimit.IsSet() && spec.GpuPortionLimit.Get() != nil {
		result.GpuPortionLimit = spec.GpuPortionLimit.Get()
	}

	if spec.GpuPortionRequest.IsSet() && spec.GpuPortionRequest.Get() != nil {
		result.GpuPortionRequest = spec.GpuPortionRequest.Get()
	}

	if spec.GpuRequestType.IsSet() && spec.GpuRequestType.Get() != nil {
		result.GpuRequestType = string(*spec.GpuRequestType.Get())
	}

	if spec.LargeShmRequest.IsSet() && spec.LargeShmRequest.Get() != nil {
		result.LargeShmRequest = spec.LargeShmRequest.Get()
	}

	return result
}

func toAssetUsageInfo(usedBy *openapi.AssetUsageInfo) *AssetUsageInfo {
	if usedBy == nil {
		return nil
	}

	result := &AssetUsageInfo{}

	if len(usedBy.Workspaces) > 0 {
		result.Workspaces = make([]WorkloadRefAndStatus, 0, len(usedBy.Workspaces))
		for _, w := range usedBy.Workspaces {
			result.Workspaces = append(result.Workspaces, toWorkloadRefAndStatus(w))
		}
	}

	if len(usedBy.Trainings) > 0 {
		result.Trainings = make([]WorkloadRefAndStatus, 0, len(usedBy.Trainings))
		for _, t := range usedBy.Trainings {
			result.Trainings = append(result.Trainings, toWorkloadRefAndStatus(t))
		}
	}

	if len(usedBy.Distributed) > 0 {
		result.Distributed = make([]WorkloadRefAndStatus, 0, len(usedBy.Distributed))
		for _, d := range usedBy.Distributed {
			result.Distributed = append(result.Distributed, toWorkloadRefAndStatus(d))
		}
	}

	if len(usedBy.Inferences) > 0 {
		result.Inferences = make([]WorkloadRefAndStatus, 0, len(usedBy.Inferences))
		for _, i := range usedBy.Inferences {
			result.Inferences = append(result.Inferences, toWorkloadRefAndStatus(i))
		}
	}

	if len(usedBy.Templates) > 0 {
		result.Templates = make([]AssetRef, 0, len(usedBy.Templates))
		for _, tmpl := range usedBy.Templates {
			result.Templates = append(result.Templates, AssetRef{
				Name: tmpl.Name,
				Id:   tmpl.Id,
			})
		}
	}

	return result
}

func toWorkloadRefAndStatus(w openapi.WorkloadRefAndStatus) WorkloadRefAndStatus {
	result := WorkloadRefAndStatus{
		Name: w.Name,
		Id:   w.Id,
	}

	if w.Status != nil {
		result.Status = string(*w.Status)
	}

	return result
}
