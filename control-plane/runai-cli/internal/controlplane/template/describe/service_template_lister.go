/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"context"
	"fmt"
	"sort"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	instancecfg "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	workloads_openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"go.uber.org/zap"
)

type Lister interface {
	List(ctx context.Context, filterBy []string, includeAncestors bool, limit, offset *int32) (*workloads_openapi.TemplatesListResponse, error)
}

type Service struct {
	lister    Lister
	clusterId string
}

func NewService(lister Lister, clusterId string) *Service {
	return &Service{
		lister:    lister,
		clusterId: clusterId,
	}
}

func (s *Service) FindByName(ctx context.Context, templateName string, projectId string, workloadType string) (*workloads_openapi.SupersetTemplate, error) {
	templates, err := s.FetchTemplates(ctx, templateName, projectId, workloadType)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template: %w", err)
	}
	if len(templates) == 0 {
		if workloadType != "" {
			return nil, fmt.Errorf("the template '%s' does not support %s workloads", templateName, spectype.APIStringToDisplayName(workloadType))
		}
		return nil, fmt.Errorf("template '%s' not found", templateName)
	}

	// Sort templates by scope priority: project > department > cluster > tenant > system
	sortedTemplates := SortByScope(templates)

	// Use the first (highest priority) template
	return &sortedTemplates[0], nil
}

func (s *Service) FetchTemplates(ctx context.Context, templateName string, projectId string, workloadType string) ([]workloads_openapi.SupersetTemplate, error) {
	filterBy := []string{fmt.Sprintf("name==%s", templateName)}

	if workloadType != "" {
		filterBy = append(filterBy, fmt.Sprintf("workloadType==%s", workloadType))
	}

	if projectId != "" {
		filterBy = append(filterBy, fmt.Sprintf("projectId==%s", projectId))
	}

	useIncludeAncestors := projectId != ""

	response, err := s.lister.List(ctx, filterBy, useIncludeAncestors, nil, nil)
	if err != nil {
		return nil, err
	}

	entries := response.GetEntries()
	return entries, nil
}

func (s *Service) GetByName(ctx context.Context, templateName, projectId string) (*Template, error) {
	rawTemplate, err := s.FindByName(ctx, templateName, projectId, "")
	if err != nil {
		return nil, err
	}

	displayTemplate, err := s.buildTemplateFromApiResponse(rawTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to build template: %w", err)
	}

	return displayTemplate, nil
}

// SortByScope sorts templates by scope priority
// Priority order: project > department > cluster > tenant > system
//
// Note: This sorting is for backward compatibility with versions prior to 2.24.
// From version 2.24 onwards, templates are guaranteed to be unique within their scope tree,
// and the backend will return only the highest priority template. This client-side sorting
// ensures correct behavior when working with older backend versions that may return multiple
// templates with the same name across different scopes.
func SortByScope(templates []workloads_openapi.SupersetTemplate) []workloads_openapi.SupersetTemplate {
	if len(templates) <= 1 {
		return templates
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([]workloads_openapi.SupersetTemplate, len(templates))
	copy(sorted, templates)

	sort.Slice(sorted, func(i, j int) bool {
		return GetScopePriority(sorted[i].Meta.ScopeType) < GetScopePriority(sorted[j].Meta.ScopeType)
	})

	return sorted
}

// GetScopePriority returns the priority value for a scope type
// Lower value = higher priority
func GetScopePriority(scopeType workloads_openapi.ScopeType) int {
	switch scopeType {
	case workloads_openapi.SCOPETYPE_PROJECT:
		return 1
	case workloads_openapi.SCOPETYPE_DEPARTMENT:
		return 2
	case workloads_openapi.SCOPETYPE_CLUSTER:
		return 3
	case workloads_openapi.SCOPETYPE_TENANT:
		return 4
	case workloads_openapi.SCOPETYPE_SYSTEM:
		return 5
	default:
		return 6
	}
}

func (s *Service) buildTemplateFromApiResponse(rawTemplate *workloads_openapi.SupersetTemplate) (*Template, error) {
	builder := NewTemplateDescriptionBuilder()

	setMetaInformation(builder, rawTemplate)

	if rawTemplate.Spec != nil {
		setSpecInformation(builder, rawTemplate.Spec)
	}

	// Handle master spec for distributed workloads (when master != worker)
	if rawTemplate.HasCombinedMasterSpec() {
		masterSpec := rawTemplate.GetCombinedMasterSpec()
		setMasterSpecInformation(builder, &masterSpec)
	}

	template := builder.Build()
	template.Raw = rawTemplate
	return template, nil
}

func setMetaInformation(builder *TemplateDescriptionBuilder, rawTemplate *workloads_openapi.SupersetTemplate) {
	meta := &rawTemplate.Meta

	builder.
		SetID(meta.Id).
		SetName(meta.Name).
		SetType(string(meta.WorkloadType)).
		SetScope(string(meta.ScopeType)).
		SetScopeID(meta.ScopeId).
		SetCreatedAt(meta.CreatedAt).
		SetCreatedBy(meta.CreatedBy)

	if meta.HasDescription() {
		if desc := meta.GetDescription(); desc != "" {
			builder.SetDescription(&desc)
		}
	}

	if meta.HasClusterId() {
		if clusterId := meta.GetClusterId(); clusterId != "" {
			builder.SetClusterID(&clusterId)
		}
	}

	if projectId, ok := meta.GetProjectIdOk(); ok && projectId != nil {
		builder.SetProjectID(projectId)
	}

	if meta.HasDepartmentId() {
		if deptId := meta.GetDepartmentId(); deptId != "" {
			builder.SetDepartmentID(&deptId)
		}
	}

	if updatedAt, ok := meta.GetUpdatedAtOk(); ok && updatedAt != nil {
		builder.SetUpdatedAt(updatedAt)
	}

	if meta.HasUpdatedBy() {
		if updatedBy := meta.GetUpdatedBy(); updatedBy != "" {
			builder.SetUpdatedBy(&updatedBy)
		}
	}
}

func setSpecInformation(builder *TemplateDescriptionBuilder, rawSpec *workloads_openapi.SupersetSpec) {
	if rawSpec.HasPreemptibility() {
		preempt := rawSpec.GetPreemptibility()
		isPreemptible := preempt == workloads_openapi.PREEMPTIBILITY_PREEMPTIBLE
		builder.SetPreemptible(&isPreemptible)
	}

	// Set string fields
	if rawSpec.HasPriorityClass() {
		if priorityClass := rawSpec.GetPriorityClass(); priorityClass != "" {
			builder.SetPriorityClass(&priorityClass)
		}
	}

	if rawSpec.HasImage() {
		if image := rawSpec.GetImage(); image != "" {
			builder.SetImage(&image)
		}
	}

	if rawSpec.HasCommand() {
		if command := rawSpec.GetCommand(); command != "" {
			builder.SetCommand(&command)
		}
	}

	if rawSpec.HasArgs() {
		if args := rawSpec.GetArgs(); args != "" {
			builder.SetArgs(&args)
		}
	}

	if rawSpec.HasWorkingDir() {
		if workingDir := rawSpec.GetWorkingDir(); workingDir != "" {
			builder.SetWorkingDir(&workingDir)
		}
	}

	if len(rawSpec.NodePools) > 0 {
		builder.SetNodePools(rawSpec.NodePools)
	}

	if len(rawSpec.EnvironmentVariables) > 0 {
		envVars := buildEnvironmentVariables(rawSpec.EnvironmentVariables)
		if len(envVars) > 0 {
			builder.SetEnvironmentVariables(envVars)
		}
	}

	if rawSpec.HasCompute() {
		buildComputeResources(builder, rawSpec.GetCompute())
	}

	setNetworks(builder, rawSpec)
}

func buildEnvironmentVariables(apiEnvVars []workloads_openapi.EnvironmentVariable) map[string]string {
	logger := instancecfg.GetInstance().GetLogger()
	envVars := make(map[string]string)
	for i, envVar := range apiEnvVars {
		name := envVar.GetName()
		value := envVar.GetValue()
		if name == "" {
			// Log debug message about invalid env var (empty name)
			logger.Debug("environment variable has empty name, skipping", zap.Int("index", i))
			continue
		}
		if _, exists := envVars[name]; exists {
			// Log debug message about duplicate env var name
			logger.Debug("duplicate environment variable, using last value", zap.String("name", name))
		}
		envVars[name] = value
	}
	return envVars
}

func buildComputeResources(builder *TemplateDescriptionBuilder, rawCompute workloads_openapi.WorkspaceSpecSpecAllOfCompute) {
	compute := &Compute{}
	hasAnyComputeField := false

	if gpuDevices, ok := rawCompute.GetGpuDevicesRequestOk(); ok && gpuDevices != nil {
		compute.GpuDevicesRequest = gpuDevices
		hasAnyComputeField = true
	}

	if rawCompute.HasGpuMemoryRequest() {
		if gpuMemReq := rawCompute.GetGpuMemoryRequest(); gpuMemReq != "" {
			compute.GpuMemoryRequest = &gpuMemReq
			hasAnyComputeField = true
		}
	}

	if rawCompute.HasGpuMemoryLimit() {
		if gpuMemLim := rawCompute.GetGpuMemoryLimit(); gpuMemLim != "" {
			compute.GpuMemoryLimit = &gpuMemLim
			hasAnyComputeField = true
		}
	}

	if cpuCore, ok := rawCompute.GetCpuCoreRequestOk(); ok && cpuCore != nil {
		compute.CpuCoreRequest = cpuCore
		hasAnyComputeField = true
	}

	if cpuCore, ok := rawCompute.GetCpuCoreLimitOk(); ok && cpuCore != nil {
		compute.CpuCoreLimit = cpuCore
		hasAnyComputeField = true
	}

	if rawCompute.HasCpuMemoryRequest() {
		if cpuMemReq := rawCompute.GetCpuMemoryRequest(); cpuMemReq != "" {
			compute.CpuMemoryRequest = &cpuMemReq
			hasAnyComputeField = true
		}
	}

	if rawCompute.HasCpuMemoryLimit() {
		if cpuMemLim := rawCompute.GetCpuMemoryLimit(); cpuMemLim != "" {
			compute.CpuMemoryLimit = &cpuMemLim
			hasAnyComputeField = true
		}
	}

	if hasAnyComputeField {
		builder.SetCompute(compute)
	}
}

func setMasterSpecInformation(builder *TemplateDescriptionBuilder, rawMasterSpec *workloads_openapi.MasterTemplateSpec) {
	if rawMasterSpec == nil {
		return
	}

	masterSpec := &MasterSpec{}
	hasAnyField := false

	if rawMasterSpec.HasImage() {
		if image := rawMasterSpec.GetImage(); image != "" {
			masterSpec.Image = &image
			hasAnyField = true
		}
	}

	if rawMasterSpec.HasCommand() {
		if command := rawMasterSpec.GetCommand(); command != "" {
			masterSpec.Command = &command
			hasAnyField = true
		}
	}

	if rawMasterSpec.HasArgs() {
		if args := rawMasterSpec.GetArgs(); args != "" {
			masterSpec.Args = &args
			hasAnyField = true
		}
	}

	if rawMasterSpec.HasWorkingDir() {
		if workingDir := rawMasterSpec.GetWorkingDir(); workingDir != "" {
			masterSpec.WorkingDir = &workingDir
			hasAnyField = true
		}
	}

	if len(rawMasterSpec.NodePools) > 0 {
		masterSpec.NodePools = rawMasterSpec.NodePools
		hasAnyField = true
	}

	if len(rawMasterSpec.EnvironmentVariables) > 0 {
		envVars := buildMasterEnvironmentVariables(rawMasterSpec.EnvironmentVariables)
		if len(envVars) > 0 {
			masterSpec.EnvironmentVariables = envVars
			hasAnyField = true
		}
	}

	if rawMasterSpec.HasCompute() {
		compute := buildMasterComputeResources(rawMasterSpec.GetCompute())
		if compute != nil {
			masterSpec.Compute = compute
			hasAnyField = true
		}
	}

	// Handle ports
	for _, apiPort := range rawMasterSpec.Ports {
		if apiPort.HasContainer() {
			port := Port{
				Container: apiPort.GetContainer(),
			}
			if external, ok := apiPort.GetExternalOk(); ok && external != nil {
				port.External = external
			}
			masterSpec.Ports = append(masterSpec.Ports, port)
			hasAnyField = true
		}
	}

	// Handle exposed URLs
	for _, apiUrl := range rawMasterSpec.ExposedUrls {
		if apiUrl.HasUrl() {
			if url := apiUrl.GetUrl(); url != "" {
				exposedUrl := ExposedUrl{
					Url:  url,
					Name: apiUrl.GetName(),
				}
				masterSpec.ExposedUrls = append(masterSpec.ExposedUrls, exposedUrl)
				hasAnyField = true
			}
		}
	}

	if hasAnyField {
		builder.SetMasterSpec(masterSpec)
	}
}

// buildMasterEnvironmentVariables transforms environment variables from master spec API response
func buildMasterEnvironmentVariables(apiEnvVars []workloads_openapi.EnvironmentVariable) map[string]string {
	logger := instancecfg.GetInstance().GetLogger()
	envVars := make(map[string]string)
	for i, envVar := range apiEnvVars {
		name := envVar.GetName()
		value := envVar.GetValue()
		if name == "" {
			logger.Debug("master environment variable has empty name, skipping", zap.Int("index", i))
			continue
		}
		if _, exists := envVars[name]; exists {
			logger.Debug("duplicate master environment variable, using last value", zap.String("name", name))
		}
		envVars[name] = value
	}
	return envVars
}

func buildMasterComputeResources(rawCompute workloads_openapi.WorkspaceSpecSpecAllOfCompute) *Compute {
	compute := &Compute{}
	hasAnyComputeField := false

	if gpuDevices, ok := rawCompute.GetGpuDevicesRequestOk(); ok && gpuDevices != nil {
		compute.GpuDevicesRequest = gpuDevices
		hasAnyComputeField = true
	}

	if rawCompute.HasGpuMemoryRequest() {
		if gpuMemReq := rawCompute.GetGpuMemoryRequest(); gpuMemReq != "" {
			compute.GpuMemoryRequest = &gpuMemReq
			hasAnyComputeField = true
		}
	}

	if rawCompute.HasGpuMemoryLimit() {
		if gpuMemLim := rawCompute.GetGpuMemoryLimit(); gpuMemLim != "" {
			compute.GpuMemoryLimit = &gpuMemLim
			hasAnyComputeField = true
		}
	}

	if cpuCore, ok := rawCompute.GetCpuCoreRequestOk(); ok && cpuCore != nil {
		compute.CpuCoreRequest = cpuCore
		hasAnyComputeField = true
	}

	if cpuCore, ok := rawCompute.GetCpuCoreLimitOk(); ok && cpuCore != nil {
		compute.CpuCoreLimit = cpuCore
		hasAnyComputeField = true
	}

	if rawCompute.HasCpuMemoryRequest() {
		if cpuMemReq := rawCompute.GetCpuMemoryRequest(); cpuMemReq != "" {
			compute.CpuMemoryRequest = &cpuMemReq
			hasAnyComputeField = true
		}
	}

	if rawCompute.HasCpuMemoryLimit() {
		if cpuMemLim := rawCompute.GetCpuMemoryLimit(); cpuMemLim != "" {
			compute.CpuMemoryLimit = &cpuMemLim
			hasAnyComputeField = true
		}
	}

	if hasAnyComputeField {
		return compute
	}
	return nil
}

func setNetworks(builder *TemplateDescriptionBuilder, rawSpec *workloads_openapi.SupersetSpec) {
	for _, apiPort := range rawSpec.Ports {
		if apiPort.HasContainer() {
			port := Port{
				Container: apiPort.GetContainer(),
			}
			if external, ok := apiPort.GetExternalOk(); ok && external != nil {
				port.External = external
			}
			builder.AddPort(port)
		}
	}

	for _, apiUrl := range rawSpec.ExposedUrls {
		if apiUrl.HasUrl() {
			if url := apiUrl.GetUrl(); url != "" {
				exposedUrl := ExposedUrl{
					Url:  url,
					Name: apiUrl.GetName(),
				}
				builder.AddExposedUrl(exposedUrl)
			}
		}
	}
}
