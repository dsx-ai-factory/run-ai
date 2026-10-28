// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"math"
	"net/http"
	"strconv"

	"github.com/distribution/reference"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	catalog "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client/generated"
	"k8s.io/apimachinery/pkg/util/validation"
)

const nimModelProfileEnv = "NIM_MODEL_PROFILE"

// resolveNimGPUAllocation operates on the policy-mutated spec, before compliance
// validation and persistence. Manual and policy-default GPU allocations win.
func (nsi *NimServicesImpl) resolveNimGPUAllocation(ctx context.Context, projectID string, spec *api.NimServiceSpec) (bool, *api.Error) {
	if spec.GpuType == nil {
		return false, nil
	}
	if *spec.GpuType == "" || len(validation.IsValidLabelValue(*spec.GpuType)) != 0 {
		return false, api.NewError(http.StatusBadRequest, "gpuType must be an exact, nonempty nvidia.com/gpu.product label value", nil)
	}
	if hasNimGPUAllocation(spec.Compute) {
		return false, nil
	}
	if spec.MultiNode != nil {
		return false, api.NewError(http.StatusBadRequest, "automatic GPU sizing is supported for single-node NIM Services; specify resources explicitly for multi-node services", nil)
	}
	if spec.Image == nil {
		return false, api.NewError(http.StatusBadRequest, "automatic GPU sizing requires an image with an explicit tag", nil)
	}
	named, err := reference.ParseNormalizedNamed(*spec.Image)
	if err != nil {
		return false, api.NewError(http.StatusBadRequest, "automatic GPU sizing requires a valid image with an explicit tag", err)
	}
	tagged, ok := named.(reference.NamedTagged)
	_, digest := named.(reference.Digested)
	if !ok || digest {
		return false, api.NewError(http.StatusBadRequest, "automatic GPU sizing requires an explicit image tag without a digest", nil)
	}
	project, projectErr := org_unit.Get().GetProjectById(ctx, projectID)
	if projectErr != nil {
		return false, api.NewError(int(projectErr.Code), projectErr.Message, nil)
	}
	request := catalog.NewRightSizingRequest(project.ClusterId, named.Name(), tagged.Tag())
	request.SetGpuType(*spec.GpuType)
	response, apiErr := nsi.recommendGPU(ctx, *request)
	if apiErr != nil {
		return false, apiErr
	}
	result, apiErr := selectNimSizingResult(response, *spec.GpuType)
	if apiErr != nil {
		return false, apiErr
	}
	if apiErr = pinNimSizingProfile(spec, result.GetProfileId()); apiErr != nil {
		return false, apiErr
	}
	allocation := result.Recommendation.Get()
	if spec.Compute == nil {
		spec.Compute = &api.NimServiceSpecAllOfCompute{}
	}
	spec.Compute.GpuDevicesRequest = convutils.Ref(int32(allocation.GpuCount))
	if allocation.AllocationType == "gpu_memory" {
		spec.Compute.GpuMemoryRequest = convutils.Ref(strconv.FormatInt(allocation.GetGpuMemoryPerDeviceBytes(), 10))
	}
	return true, nil
}

func hasNimGPUAllocation(c *api.NimServiceSpecAllOfCompute) bool {
	return c != nil && (c.GpuDevicesRequest != nil || c.GpuMemoryRequest != nil || c.GpuMemoryLimit != nil ||
		c.GpuMemoryPortionRequest != nil || c.GpuMemoryPortionLimit != nil || c.GpuPortionRequest != nil || c.GpuPortionLimit != nil)
}

func selectNimSizingResult(response *catalog.RightSizingResponse, gpuType string) (*catalog.RightSizingResult, *api.Error) {
	var selected *catalog.RightSizingResult
	if response != nil {
		for i := range response.Results {
			result := &response.Results[i]
			if result.GpuType != gpuType {
				continue
			}
			a := result.Recommendation.Get()
			if result.Status != "recommended" || a == nil {
				reason := result.GetReason()
				if reason == "" {
					reason = "no eligible recommendation"
				}
				return nil, api.NewError(http.StatusBadRequest, "cannot automatically size NIM Service: "+reason, nil)
			}
			if a.GpuCount < 1 || a.GpuCount > math.MaxInt32 || result.GetProfileId() == "" ||
				(a.AllocationType != "gpu_memory" && a.AllocationType != "gpu_count") ||
				(a.AllocationType == "gpu_memory" && (a.GetGpuMemoryPerDeviceBytes() < 1 || a.GpuCount != 1)) {
				return nil, api.NewError(http.StatusBadGateway, "recommendation service returned an invalid NIM allocation or profile", nil)
			}
			if a.Estimated && a.GpuCount > 1 {
				return nil, api.NewError(http.StatusBadRequest, "automatic NIM sizing requires a validated multi-GPU profile; the available count is only estimated; specify resources explicitly", nil)
			}
			if selected != nil {
				b := selected.Recommendation.Get()
				if selected.GetProfileId() != result.GetProfileId() || b.AllocationType != a.AllocationType || b.GpuCount != a.GpuCount || b.GetGpuMemoryPerDeviceBytes() != a.GetGpuMemoryPerDeviceBytes() {
					return nil, api.NewError(http.StatusBadRequest, "GPU capacities sharing this product label require different allocations or profiles; specify resources explicitly", nil)
				}
			}
			selected = result
		}
	}
	if selected == nil {
		return nil, api.NewError(http.StatusBadRequest, "no GPU sizing recommendation is available for the requested gpuType", nil)
	}
	return selected, nil
}

func pinNimSizingProfile(spec *api.NimServiceSpec, profile string) *api.Error {
	var cache *api.NimCache
	if spec.ModelStore != nil {
		cache = spec.ModelStore.NimCache
	}
	if cache != nil && cache.Profile != nil && *cache.Profile != profile {
		return api.NewError(http.StatusBadRequest, "modelStore.nimCache.profile conflicts with the automatically selected profile; specify GPU resources explicitly to use another profile", nil)
	}
	found := false
	var environmentVariables []api.EnvironmentVariable
	if spec.EnvironmentVariables != nil {
		environmentVariables = *spec.EnvironmentVariables
	}
	for _, env := range environmentVariables {
		if env.Name == nil || *env.Name != nimModelProfileEnv {
			continue
		}
		if found || env.Value == nil || *env.Value != profile || env.Secret != nil || env.ConfigMap != nil || env.PodFieldRef != nil || env.UserCredential != nil || (env.Exclude != nil && *env.Exclude) {
			return api.NewError(http.StatusBadRequest, "NIM_MODEL_PROFILE conflicts with the automatically selected profile; specify GPU resources explicitly to use another profile", nil)
		}
		found = true
	}
	if cache != nil {
		cache.Profile = convutils.Ref(profile)
	}
	if !found {
		environmentVariables = append(environmentVariables, api.EnvironmentVariable{Name: convutils.Ref(nimModelProfileEnv), Value: convutils.Ref(profile)})
		spec.EnvironmentVariables = &environmentVariables
	}
	return nil
}
