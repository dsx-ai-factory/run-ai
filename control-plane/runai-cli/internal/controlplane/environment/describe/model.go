/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"time"

	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

// EnvironmentAsset Output types for JSON/YAML with omitempty tags
type EnvironmentAsset struct {
	Meta   AssetMeta           `json:"meta" yaml:"meta"`
	Spec   EnvironmentSpec     `json:"spec" yaml:"spec"`
	Status *AssetClusterStatus `json:"status,omitempty" yaml:"status,omitempty"`
	UsedBy *AssetUsageInfo     `json:"usedBy,omitempty" yaml:"usedBy,omitempty"`
}

type AssetMeta struct {
	Name                   string                  `json:"name" yaml:"name"`
	Description            string                  `json:"description,omitempty" yaml:"description,omitempty"`
	Scope                  string                  `json:"scope" yaml:"scope"`
	ClusterId              string                  `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	DepartmentId           string                  `json:"departmentId,omitempty" yaml:"departmentId,omitempty"`
	ProjectId              *int32                  `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	ProjectName            string                  `json:"projectName,omitempty" yaml:"projectName,omitempty"`
	TenantId               *int32                  `json:"tenantId,omitempty" yaml:"tenantId,omitempty"`
	WorkloadSupportedTypes *WorkloadSupportedTypes `json:"workloadSupportedTypes,omitempty" yaml:"workloadSupportedTypes,omitempty"`
	Id                     string                  `json:"id" yaml:"id"`
	Kind                   string                  `json:"kind" yaml:"kind"`
	CreatedBy              string                  `json:"createdBy" yaml:"createdBy"`
	CreatedAt              time.Time               `json:"createdAt" yaml:"createdAt"`
	UpdatedBy              string                  `json:"updatedBy" yaml:"updatedBy"`
	UpdatedAt              time.Time               `json:"updatedAt" yaml:"updatedAt"`
	DeletedAt              *time.Time              `json:"deletedAt,omitempty" yaml:"deletedAt,omitempty"`
	DeletedBy              string                  `json:"deletedBy,omitempty" yaml:"deletedBy,omitempty"`
	UpdateCount            *int64                  `json:"updateCount,omitempty" yaml:"updateCount,omitempty"`
}

type WorkloadSupportedTypes struct {
	Workspace     *bool  `json:"workspace,omitempty" yaml:"workspace,omitempty"`
	Training      *bool  `json:"training,omitempty" yaml:"training,omitempty"`
	Inference     *bool  `json:"inference,omitempty" yaml:"inference,omitempty"`
	Distributed   *bool  `json:"distributed,omitempty" yaml:"distributed,omitempty"`
	DistFramework string `json:"distFramework,omitempty" yaml:"distFramework,omitempty"`
}

type EnvironmentSpec struct {
	Image                string                `json:"image,omitempty" yaml:"image,omitempty"`
	ImagePullPolicy      string                `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
	WorkingDir           string                `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	Command              string                `json:"command,omitempty" yaml:"command,omitempty"`
	Args                 string                `json:"args,omitempty" yaml:"args,omitempty"`
	EnvironmentVariables []EnvironmentVariable `json:"environmentVariables,omitempty" yaml:"environmentVariables,omitempty"`
	Capabilities         []string              `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
	Connections          []Connection          `json:"connections,omitempty" yaml:"connections,omitempty"`
	RunAsUid             *int64                `json:"runAsUid,omitempty" yaml:"runAsUid,omitempty"`
	RunAsGid             *int64                `json:"runAsGid,omitempty" yaml:"runAsGid,omitempty"`
	UidGidSource         string                `json:"uidGidSource,omitempty" yaml:"uidGidSource,omitempty"`
	SupplementalGroups   string                `json:"supplementalGroups,omitempty" yaml:"supplementalGroups,omitempty"`
}

type EnvironmentVariable struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

type Connection struct {
	Name             string            `json:"name,omitempty" yaml:"name,omitempty"`
	InternalToolInfo *InternalToolInfo `json:"internalToolInfo,omitempty" yaml:"internalToolInfo,omitempty"`
	ExternalToolInfo *ExternalToolInfo `json:"externalToolInfo,omitempty" yaml:"externalToolInfo,omitempty"`
}

type InternalToolInfo struct {
	ToolType string `json:"toolType" yaml:"toolType"`
}

type ExternalToolInfo struct {
	ToolType    string `json:"toolType" yaml:"toolType"`
	ExternalUrl string `json:"externalUrl,omitempty" yaml:"externalUrl,omitempty"`
}

type AssetClusterStatus struct {
	Status  string                    `json:"status,omitempty" yaml:"status,omitempty"`
	Issues  []AssetClusterStatusIssue `json:"issues,omitempty" yaml:"issues,omitempty"`
	Message string                    `json:"message,omitempty" yaml:"message,omitempty"`
	Url     string                    `json:"url,omitempty" yaml:"url,omitempty"`
}

type AssetClusterStatusIssue struct {
	ScopeId   string `json:"scopeId" yaml:"scopeId"`
	ScopeType string `json:"scopeType" yaml:"scopeType"`
	Issue     string `json:"issue" yaml:"issue"`
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

// ToEnvironmentAsset converts OpenAPI struct to output struct with omitempty tags
func ToEnvironmentAsset(env *openapi.EnvironmentAsset) *EnvironmentAsset {
	if env == nil {
		return nil
	}

	result := &EnvironmentAsset{
		Meta: toAssetMeta(env.Meta),
		Spec: toEnvironmentSpec(env.Spec),
	}

	// Add status if available
	if env.Status.IsSet() && env.Status.Get() != nil {
		result.Status = toAssetClusterStatus(env.Status.Get())
	}

	// Add usedBy if available
	if env.UsedBy.IsSet() && env.UsedBy.Get() != nil {
		result.UsedBy = toAssetUsageInfo(env.UsedBy.Get())
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

	if meta.WorkloadSupportedTypes.IsSet() && meta.WorkloadSupportedTypes.Get() != nil {
		result.WorkloadSupportedTypes = toWorkloadSupportedTypes(meta.WorkloadSupportedTypes.Get())
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

func toWorkloadSupportedTypes(wst *openapi.WorkloadSupportedTypes) *WorkloadSupportedTypes {
	if wst == nil {
		return nil
	}

	result := &WorkloadSupportedTypes{}

	if wst.Workspace.IsSet() && wst.Workspace.Get() != nil {
		result.Workspace = wst.Workspace.Get()
	}

	if wst.Training.IsSet() && wst.Training.Get() != nil {
		result.Training = wst.Training.Get()
	}

	if wst.Inference.IsSet() && wst.Inference.Get() != nil {
		result.Inference = wst.Inference.Get()
	}

	if wst.Distributed.IsSet() && wst.Distributed.Get() != nil {
		result.Distributed = wst.Distributed.Get()
	}

	if wst.DistFramework.IsSet() && wst.DistFramework.Get() != nil {
		result.DistFramework = *wst.DistFramework.Get()
	}

	return result
}

func toEnvironmentSpec(spec openapi.EnvironmentAssetSpec) EnvironmentSpec {
	result := EnvironmentSpec{}

	if spec.HasImage() {
		result.Image = spec.GetImage()
	}

	if spec.ImagePullPolicy.IsSet() && spec.ImagePullPolicy.Get() != nil {
		result.ImagePullPolicy = string(*spec.ImagePullPolicy.Get())
	}

	if spec.WorkingDir.IsSet() && spec.WorkingDir.Get() != nil {
		result.WorkingDir = *spec.WorkingDir.Get()
	}

	if spec.Command.IsSet() && spec.Command.Get() != nil {
		result.Command = *spec.Command.Get()
	}

	if spec.Args.IsSet() && spec.Args.Get() != nil {
		result.Args = *spec.Args.Get()
	}

	if len(spec.EnvironmentVariables) > 0 {
		result.EnvironmentVariables = make([]EnvironmentVariable, 0, len(spec.EnvironmentVariables))
		for _, ev := range spec.EnvironmentVariables {
			envVar := EnvironmentVariable{}
			if ev.Name.IsSet() && ev.Name.Get() != nil {
				envVar.Name = *ev.Name.Get()
			}
			if ev.Value.IsSet() && ev.Value.Get() != nil {
				envVar.Value = *ev.Value.Get()
			}
			result.EnvironmentVariables = append(result.EnvironmentVariables, envVar)
		}
	}

	if len(spec.Capabilities) > 0 {
		result.Capabilities = make([]string, 0, len(spec.Capabilities))
		for _, cap := range spec.Capabilities {
			result.Capabilities = append(result.Capabilities, string(cap))
		}
	}

	if len(spec.Connections) > 0 {
		result.Connections = make([]Connection, 0, len(spec.Connections))
		for _, conn := range spec.Connections {
			connection := Connection{
				Name: conn.Name,
			}
			if conn.HasInternalToolInfo() {
				toolInfo := conn.GetInternalToolInfo()
				connection.InternalToolInfo = &InternalToolInfo{
					ToolType: string(toolInfo.ToolType),
				}
			}
			if conn.HasExternalToolInfo() {
				toolInfo := conn.GetExternalToolInfo()
				connection.ExternalToolInfo = &ExternalToolInfo{
					ToolType:    string(toolInfo.ToolType),
					ExternalUrl: toolInfo.ExternalUrl,
				}
			}
			result.Connections = append(result.Connections, connection)
		}
	}

	if spec.RunAsUid.IsSet() && spec.RunAsUid.Get() != nil {
		result.RunAsUid = spec.RunAsUid.Get()
	}

	if spec.RunAsGid.IsSet() && spec.RunAsGid.Get() != nil {
		result.RunAsGid = spec.RunAsGid.Get()
	}

	if spec.UidGidSource.IsSet() && spec.UidGidSource.Get() != nil {
		result.UidGidSource = string(*spec.UidGidSource.Get())
	}

	if spec.SupplementalGroups.IsSet() && spec.SupplementalGroups.Get() != nil {
		result.SupplementalGroups = *spec.SupplementalGroups.Get()
	}

	return result
}

func toAssetClusterStatus(status *openapi.AssetClusterStatusInfo) *AssetClusterStatus {
	if status == nil {
		return nil
	}

	result := &AssetClusterStatus{}

	if status.Status != nil {
		result.Status = string(*status.Status)
	}
	if status.Message != nil {
		result.Message = *status.Message
	}
	if status.Url != nil {
		result.Url = *status.Url
	}
	if len(status.Issues) > 0 {
		result.Issues = make([]AssetClusterStatusIssue, 0, len(status.Issues))
		for _, issue := range status.Issues {
			result.Issues = append(result.Issues, AssetClusterStatusIssue{
				ScopeId:   issue.ScopeId,
				ScopeType: string(issue.ScopeType),
				Issue:     string(issue.Issue),
			})
		}
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
