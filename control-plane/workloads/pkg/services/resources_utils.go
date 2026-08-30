package services

import (
	"github.com/lib/pq"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	log "github.com/run-ai/runai/runai-common-packages/logging"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

func getWorkloadRequestedResources(
	requestedGPULimit *float64,
	requestedGPURequest *float64,
	requestedGPUMemoryLimit *string,
	requestedGPUMemoryRequest *string,
	requestedCPULimit *float64,
	requestedCPURequest *float64,
	requestedCPUMemoryLimit *string,
	requestedCPUMemoryRequest *string,
) *api.WorkloadRequestResources {
	resources := &api.WorkloadRequestResources{}

	if requestedGPULimit != nil || requestedGPURequest != nil {
		resources.GpuRequestType = (*api.GpuRequestType)(ptr.To(string(api.GPUREQUESTTYPE_PORTION)))
		resources.Gpu = &api.RequestResourceCores{
			Limit:   requestedGPULimit,
			Request: requestedGPURequest,
		}
	} else if requestedGPUMemoryLimit != nil || requestedGPUMemoryRequest != nil {
		resources.GpuRequestType = (*api.GpuRequestType)(ptr.To(string(api.GPUREQUESTTYPE_MEMORY)))
		resources.GpuMemory = &api.RequestResourceQuantity{
			Limit:   requestedGPUMemoryLimit,
			Request: requestedGPUMemoryRequest,
		}
	}

	if requestedCPULimit != nil || requestedCPURequest != nil {
		resources.Cpu = &api.RequestResourceCores{
			Limit:   requestedCPULimit,
			Request: requestedCPURequest,
		}
	}

	if requestedCPUMemoryLimit != nil || requestedCPUMemoryRequest != nil {
		resources.CpuMemory = &api.RequestResourceQuantity{
			Limit:   requestedCPUMemoryLimit,
			Request: requestedCPUMemoryRequest,
		}
	}

	return resources
}

func getPodsRequestedResources(
	requestedGPULimit *float64,
	requestedGPURequest *float64,
	requestedGPUMemoryLimit *pq.StringArray,
	requestedGPUMemoryRequest *pq.StringArray,
	requestedCPULimit *float64,
	requestedCPURequest *float64,
	requestedCPUMemoryLimit *pq.StringArray,
	requestedCPUMemoryRequest *pq.StringArray,
) *api.WorkloadRequestResources {
	resources := &api.WorkloadRequestResources{}

	if requestedGPULimit != nil || requestedGPURequest != nil {
		resources.GpuRequestType = (*api.GpuRequestType)(ptr.To(string(api.GPUREQUESTTYPE_PORTION)))
		resources.Gpu = &api.RequestResourceCores{
			Limit:   requestedGPULimit,
			Request: requestedGPURequest,
		}
	} else if requestedGPUMemoryLimit != nil || requestedGPUMemoryRequest != nil {
		resources.GpuRequestType = (*api.GpuRequestType)(ptr.To(string(api.GPUREQUESTTYPE_MEMORY)))
		resources.GpuMemory = &api.RequestResourceQuantity{
			Limit:   sumQuantities(requestedGPUMemoryLimit),
			Request: sumQuantities(requestedGPUMemoryRequest),
		}
	}

	if requestedCPULimit != nil || requestedCPURequest != nil {
		resources.Cpu = &api.RequestResourceCores{
			Limit:   requestedCPULimit,
			Request: requestedCPURequest,
		}
	}

	if requestedCPUMemoryLimit != nil || requestedCPUMemoryRequest != nil {
		resources.CpuMemory = &api.RequestResourceQuantity{
			Limit:   sumQuantities(requestedCPUMemoryLimit),
			Request: sumQuantities(requestedCPUMemoryRequest),
		}
	}

	return resources
}

type WorkloadOrRevisionResources struct {
	RequestedGPULimit         *float64
	RequestedGPURequest       *float64
	RequestedGPUMemoryLimit   *string
	RequestedGPUMemoryRequest *string
	RequestedCPULimit         *float64
	RequestedCPURequest       *float64
	RequestedCPUMemoryLimit   *string
	RequestedCPUMemoryRequest *string
	RequestedMigDevice        pq.StringArray
}

func setResourcesGeneric(requestedResources *api.WorkloadRequestResources) (*WorkloadOrRevisionResources, error) {
	var resources WorkloadOrRevisionResources

	if api.IsZeroValue(requestedResources) {
		return &resources, nil
	}

	if requestedResources.GpuRequestType != nil && *requestedResources.GpuRequestType == api.GPUREQUESTTYPE_PORTION && !api.IsZeroValue(requestedResources.Gpu) {
		resources.RequestedGPULimit = requestedResources.Gpu.Limit
		resources.RequestedGPURequest = requestedResources.Gpu.Request
	} else {
		if requestedResources.GpuRequestType != nil && *requestedResources.GpuRequestType == api.GPUREQUESTTYPE_MEMORY && !api.IsZeroValue(requestedResources.GpuMemory) {
			limit, request, err := validateRequestResourceQuantity(requestedResources.GpuMemory)
			if err != nil {
				return nil, err
			}
			resources.RequestedGPUMemoryLimit = limit
			resources.RequestedGPUMemoryRequest = request
		}
	}

	if !api.IsZeroValue(requestedResources.Cpu) {
		resources.RequestedCPULimit = requestedResources.Cpu.Limit
		resources.RequestedCPURequest = requestedResources.Cpu.Request
	}

	if !api.IsZeroValue(requestedResources.CpuMemory) {

		if requestedResources.CpuMemory.Limit != nil {
			limit, err := validateQuantity(*requestedResources.CpuMemory.Limit)
			if err != nil {
				return nil, err
			}
			resources.RequestedCPUMemoryLimit = limit
		}

		if requestedResources.CpuMemory.Request != nil {
			request, err := validateQuantity(*requestedResources.CpuMemory.Request)
			if err != nil {
				return nil, err
			}
			resources.RequestedCPUMemoryRequest = request
		}
	}

	return &resources, nil
}

func setWorkloadResources(apiWorkload api.WorkloadSyncFields, modelWorkload *models.Workload) error {
	resources, err := setResourcesGeneric(apiWorkload.RequestedResources)
	if err != nil {
		return err
	}

	modelWorkload.RequestedGPULimit = resources.RequestedGPULimit
	modelWorkload.RequestedGPURequest = resources.RequestedGPURequest
	modelWorkload.RequestedGPUMemoryLimit = resources.RequestedGPUMemoryLimit
	modelWorkload.RequestedGPUMemoryRequest = resources.RequestedGPUMemoryRequest
	modelWorkload.RequestedCPULimit = resources.RequestedCPULimit
	modelWorkload.RequestedCPURequest = resources.RequestedCPURequest
	modelWorkload.RequestedCPUMemoryLimit = resources.RequestedCPUMemoryLimit
	modelWorkload.RequestedCPUMemoryRequest = resources.RequestedCPUMemoryRequest
	modelWorkload.RequestedMigDevice = resources.RequestedMigDevice

	return nil
}

func setRevisionResources(apiRevision api.RevisionSyncFields, modelRevision *models.Revision) error {
	resources, err := setResourcesGeneric(apiRevision.RequestedResources)
	if err != nil {
		return err
	}

	modelRevision.RequestedGPULimit = resources.RequestedGPULimit
	modelRevision.RequestedGPURequest = resources.RequestedGPURequest
	modelRevision.RequestedGPUMemoryLimit = resources.RequestedGPUMemoryLimit
	modelRevision.RequestedGPUMemoryRequest = resources.RequestedGPUMemoryRequest
	modelRevision.RequestedCPULimit = resources.RequestedCPULimit
	modelRevision.RequestedCPURequest = resources.RequestedCPURequest
	modelRevision.RequestedCPUMemoryLimit = resources.RequestedCPUMemoryLimit
	modelRevision.RequestedCPUMemoryRequest = resources.RequestedCPUMemoryRequest
	modelRevision.RequestedMigDevice = resources.RequestedMigDevice

	return nil
}

func getPodRequestedResources(
	requestedGPULimit *float64,
	requestedGPURequest *float64,
	requestedGPUMemoryLimit *string,
	requestedGPUMemoryRequest *string,
	requestedCPULimit *float64,
	requestedCPURequest *float64,
	requestedCPUMemoryLimit *string,
	requestedCPUMemoryRequest *string,
	requestedMigDevice *string,
) *api.PodRequestResources {
	resources := &api.PodRequestResources{}

	if requestedGPULimit != nil || requestedGPURequest != nil {
		resources.GpuRequestType = (*api.GpuRequestType)(ptr.To(string(api.GPUREQUESTTYPE_PORTION)))
		resources.Gpu = &api.RequestResourceCores{
			Limit:   requestedGPULimit,
			Request: requestedGPURequest,
		}
	} else if requestedGPUMemoryLimit != nil || requestedGPUMemoryRequest != nil {
		resources.GpuRequestType = (*api.GpuRequestType)(ptr.To(string(api.GPUREQUESTTYPE_MEMORY)))
		resources.GpuMemory = &api.RequestResourceQuantity{
			Limit:   requestedGPUMemoryLimit,
			Request: requestedGPUMemoryRequest,
		}
	}

	if requestedCPULimit != nil || requestedCPURequest != nil {
		resources.Cpu = &api.RequestResourceCores{
			Limit:   requestedCPULimit,
			Request: requestedCPURequest,
		}
	}

	if requestedCPUMemoryLimit != nil || requestedCPUMemoryRequest != nil {
		resources.CpuMemory = &api.RequestResourceQuantity{
			Limit:   requestedCPUMemoryLimit,
			Request: requestedCPUMemoryRequest,
		}
	}

	return resources
}

func setPodResources(apiPod *api.PodRequestResources, modelPod models.Pod) (models.Pod, error) {
	if api.IsZeroValue(apiPod) {
		return modelPod, nil
	}

	if apiPod.GpuRequestType != nil && *apiPod.GpuRequestType == api.GPUREQUESTTYPE_PORTION && !api.IsZeroValue(apiPod.Gpu) {
		modelPod.RequestedGPULimit = apiPod.Gpu.Limit
		modelPod.RequestedGPURequest = apiPod.Gpu.Request
	} else {
		if apiPod.GpuRequestType != nil && *apiPod.GpuRequestType == api.GPUREQUESTTYPE_MEMORY && !api.IsZeroValue(apiPod.GpuMemory) {
			limit, request, err := validateRequestResourceQuantity(apiPod.GpuMemory)
			if err != nil {
				return models.Pod{}, err
			}
			modelPod.RequestedGPUMemoryLimit = limit
			modelPod.RequestedGPUMemoryRequest = request
		}
	}

	if !api.IsZeroValue(apiPod.Cpu) {
		modelPod.RequestedCPULimit = apiPod.Cpu.Limit
		modelPod.RequestedCPURequest = apiPod.Cpu.Request
	}

	if !api.IsZeroValue(apiPod.CpuMemory) {
		limit, request, err := validateRequestResourceQuantity(apiPod.CpuMemory)
		if err != nil {
			return models.Pod{}, err
		}
		modelPod.RequestedCPUMemoryLimit = limit
		modelPod.RequestedCPUMemoryRequest = request
	}

	return modelPod, nil
}

func validateRequestResourceQuantity(requestResourceQuantity *api.RequestResourceQuantity) (limit *string, request *string, err error) {
	if requestResourceQuantity.Limit != nil {
		limit, err = validateQuantity(*requestResourceQuantity.Limit)
		if err != nil {
			return nil, nil, err
		}
	}

	if requestResourceQuantity.Request != nil {
		request, err = validateQuantity(*requestResourceQuantity.Request)
		if err != nil {
			return nil, nil, err
		}
	}

	return limit, request, err
}

func validateQuantity(field string) (*string, error) {
	if api.IsZeroValue(field) {
		return nil, nil
	}

	_, err := resource.ParseQuantity(field)
	if err != nil {
		log.Errorf("Quantity is invalid <%s>.", field)
		return nil, err
	}

	return &field, nil
}

func sumQuantities(quantities *pq.StringArray) *string {
	if quantities == nil {
		return nil
	}

	sum := resource.Quantity{}
	for _, quantity := range *quantities {
		parsedResource, err := resource.ParseQuantity(quantity)
		if err != nil {
			log.Warnf("Malformed quantity stored in DB, <%s>", quantity)
			continue
		}

		sum.Add(parsedResource)
	}

	return ptr.To(sum.String())
}

// aggregatePodsRequestedResources aggregates requested resource fields from individual pods
// into the same structure used at the workload level. Completed pods are excluded.
func aggregatePodsRequestedResources(pods []models.Pod) *api.WorkloadRequestResources {
	var gpuLimitSum, gpuRequestSum, cpuLimitSum, cpuRequestSum float64
	var hasGPULimit, hasGPURequest, hasCPULimit, hasCPURequest bool
	var gpuMemLimits, gpuMemRequests, cpuMemLimits, cpuMemRequests pq.StringArray

	for _, p := range pods {
		if p.CompletedAt != nil {
			continue
		}
		if p.RequestedGPULimit != nil {
			gpuLimitSum += *p.RequestedGPULimit
			hasGPULimit = true
		}
		if p.RequestedGPURequest != nil {
			gpuRequestSum += *p.RequestedGPURequest
			hasGPURequest = true
		}
		if p.RequestedGPUMemoryLimit != nil {
			gpuMemLimits = append(gpuMemLimits, *p.RequestedGPUMemoryLimit)
		}
		if p.RequestedGPUMemoryRequest != nil {
			gpuMemRequests = append(gpuMemRequests, *p.RequestedGPUMemoryRequest)
		}
		if p.RequestedCPULimit != nil {
			cpuLimitSum += *p.RequestedCPULimit
			hasCPULimit = true
		}
		if p.RequestedCPURequest != nil {
			cpuRequestSum += *p.RequestedCPURequest
			hasCPURequest = true
		}
		if p.RequestedCPUMemoryLimit != nil {
			cpuMemLimits = append(cpuMemLimits, *p.RequestedCPUMemoryLimit)
		}
		if p.RequestedCPUMemoryRequest != nil {
			cpuMemRequests = append(cpuMemRequests, *p.RequestedCPUMemoryRequest)
		}
	}

	var gpuLimit, gpuRequest, cpuLimit, cpuRequest *float64
	if hasGPULimit {
		gpuLimit = &gpuLimitSum
	}
	if hasGPURequest {
		gpuRequest = &gpuRequestSum
	}
	if hasCPULimit {
		cpuLimit = &cpuLimitSum
	}
	if hasCPURequest {
		cpuRequest = &cpuRequestSum
	}

	return getPodsRequestedResources(
		gpuLimit, gpuRequest,
		nilIfEmpty(gpuMemLimits), nilIfEmpty(gpuMemRequests),
		cpuLimit, cpuRequest,
		nilIfEmpty(cpuMemLimits), nilIfEmpty(cpuMemRequests),
	)
}

// aggregatePodsAllocatedResources aggregates allocated resource fields from individual pods.
// Completed pods are excluded from the sum (matching the workloads_view pattern).
func aggregatePodsAllocatedResources(pods []models.Pod) *api.WorkloadAllocatedResources {
	var gpuSum, cpuSum float64
	var hasGPU, hasCPU bool
	var gpuMemories, cpuMemories pq.StringArray

	for _, p := range pods {
		if p.CompletedAt != nil {
			continue
		}
		if p.AllocatedGPU != nil {
			gpuSum += *p.AllocatedGPU
			hasGPU = true
		}
		if p.AllocatedGPUMemory != nil {
			gpuMemories = append(gpuMemories, *p.AllocatedGPUMemory)
		}
		if p.AllocatedCPU != nil {
			cpuSum += *p.AllocatedCPU
			hasCPU = true
		}
		if p.AllocatedCPUMemory != nil {
			cpuMemories = append(cpuMemories, *p.AllocatedCPUMemory)
		}
	}

	result := &api.WorkloadAllocatedResources{}
	if hasGPU {
		result.Gpu = &gpuSum
	}
	result.GpuMemory = sumQuantities(nilIfEmpty(gpuMemories))
	if hasCPU {
		result.Cpu = &cpuSum
	}
	result.CpuMemory = sumQuantities(nilIfEmpty(cpuMemories))
	return result
}

func nilIfEmpty(arr pq.StringArray) *pq.StringArray {
	if len(arr) == 0 {
		return nil
	}
	return &arr
}
