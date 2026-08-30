package pod_mutators

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/patch"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/utils"
)

const (
	CpuDefaultRequestGpuFactor                              = "cpuDefaultRequestGpuFactor"
	CpuDefaultLimitGpuFactor                                = "cpuDefaultLimitGpuFactor"
	MemoryDefaultRequestGpuFactor                           = "memoryDefaultRequestGpuFactor"
	MemoryDefaultLimitGpuFactor                             = "memoryDefaultLimitGpuFactor"
	CpuDefaultRequestCpuLimitFactorNoGpu                    = "cpuDefaultRequestCpuLimitFactorNoGpu"
	MemoryDefaultRequestMemoryLimitFactorNoGpu              = "memoryDefaultRequestMemoryLimitFactorNoGpu"
	Requests                                   resourceType = "requests"
	Limits                                     resourceType = "limits"
)

type resourceType = string

func (pm *PodMutator) mutatePodResources(pod *v1.Pod, req *v1beta1.AdmissionRequest) error {
	containers := pod.Spec.Containers
	containerIndex := utils.GetRunaiContainerIndex(pod.Labels, containers)
	requirements := &containers[containerIndex].Resources

	configMap, err := pm.client.GetRunaiConfigMap()
	if err != nil || len(configMap.Data) == 0 {
		log.Printf("Failed to get default-limit-range config map")
		//lint:ignore nilerr this error should not fail the mutator
		return nil
	}

	gpuToCpuFactor := utils.GetRequestedGpus(requirements.Limits, &pod.ObjectMeta)

	// If no GPUs via traditional resources, check DRA claims
	if gpuToCpuFactor == 0 && pm.draParser != nil && len(pod.Spec.ResourceClaims) > 0 {
		namespace := getDestinationNamespace(pod, req)
		requestedGpus := pm.draParser.ResolveDRAGpuCount(
			context.TODO(), namespace, pod.Spec.ResourceClaims)
		gpuToCpuFactor = float64(requestedGpus)
	}

	if gpuToCpuFactor == 0 {
		setRequestResourceLimitFactor(requirements, configMap, v1.ResourceMemory)
		setRequestResourceLimitFactor(requirements, configMap, v1.ResourceCPU)
		return nil
	}

	setCpuResource(requirements, configMap, gpuToCpuFactor, Requests)
	setCpuResource(requirements, configMap, gpuToCpuFactor, Limits)
	setMemoryResource(requirements, configMap, gpuToCpuFactor, Requests)
	setMemoryResource(requirements, configMap, gpuToCpuFactor, Limits)
	return nil
}

func setRequestResourceLimitFactor(requirements *v1.ResourceRequirements, configMap *v1.ConfigMap,
	resourceName v1.ResourceName) {
	var request *resource.Quantity
	var limit *resource.Quantity
	var limitFactorStr string
	var found bool

	switch resourceName {
	case v1.ResourceCPU:
		request = requirements.Requests.Cpu()
		limit = requirements.Limits.Cpu()
		limitFactorStr, found = configMap.Data[CpuDefaultRequestCpuLimitFactorNoGpu]
	case v1.ResourceMemory:
		request = requirements.Requests.Memory()
		limit = requirements.Limits.Memory()
		limitFactorStr, found = configMap.Data[MemoryDefaultRequestMemoryLimitFactorNoGpu]
	default:
		return
	}

	if !request.IsZero() || limit.IsZero() || !found {
		log.Printf("No need to set %v request, request: %v, limit: %v", resourceName, request.Value(),
			limit.Value())
		return
	}

	limitFactor, err := strconv.ParseFloat(limitFactorStr, 64)
	if err != nil {
		log.Printf("error while getting %v request factor %v", resourceName, err)
		return
	}

	var valueToSet int
	var valueAsString string

	if resourceName == v1.ResourceCPU {
		valueToSet = int(float64(limit.MilliValue()) * limitFactor)
		valueAsString = fmt.Sprintf("%dm", valueToSet)
	} else if resourceName == v1.ResourceMemory {
		valueToSet = int(float64(limit.Value()) * limitFactor)
		valueAsString = fmt.Sprintf("%d", valueToSet)
	}

	log.Printf("Setting request %v of: %v", resourceName, valueAsString)
	addResourceTypeToRequirements(requirements, resourceName, Requests, valueAsString)
}

func setCpuResource(requirements *v1.ResourceRequirements, configMap *v1.ConfigMap, requestedGpus float64,
	resourceToSet resourceType) {
	setResource(requirements, configMap, requestedGpus, resourceToSet, v1.ResourceCPU,
		CpuDefaultRequestGpuFactor, CpuDefaultLimitGpuFactor)
}

func setMemoryResource(requirements *v1.ResourceRequirements, configMap *v1.ConfigMap, requestedGpus float64,
	resourceToSet resourceType) {
	setResource(requirements, configMap, requestedGpus, resourceToSet, v1.ResourceMemory,
		MemoryDefaultRequestGpuFactor, MemoryDefaultLimitGpuFactor)
}

func setResource(requirements *v1.ResourceRequirements, configMap *v1.ConfigMap, requestedGpus float64,
	resourceToSet resourceType, resourceName v1.ResourceName, defaultRequestsGpuFactor string,
	defaultLimitGpuFactor string) {
	var resourceQuantity resource.Quantity
	var resourceFactor string
	var found bool

	if resourceToSet == Requests {
		resourceQuantity = requirements.Requests[resourceName]
		resourceFactor, found = configMap.Data[defaultRequestsGpuFactor]
	} else {
		resourceQuantity = requirements.Limits[resourceName]
		resourceFactor, found = configMap.Data[defaultLimitGpuFactor]
	}

	if !resourceQuantity.IsZero() || !found {
		log.Printf("No need to set %s, resource type: %v", resourceName, resourceToSet)
		return
	}

	newResourceQuantity, err := resource.ParseQuantity(resourceFactor)
	if err != nil {
		log.Printf("error while getting %s quantity %v", resourceName, err)
		return
	}

	var valueToSet int64
	var valueAsString string
	switch resourceName {
	case v1.ResourceCPU:
		valueToSet = int64(float64(newResourceQuantity.MilliValue()) * requestedGpus)
		valueAsString = fmt.Sprintf("%dm", valueToSet)
	case v1.ResourceMemory:
		valueToSet = int64(float64(newResourceQuantity.Value()) * requestedGpus)
		valueAsString = fmt.Sprintf("%d", valueToSet)
	default:
		log.Printf("unknown resource type: %s", resourceName)
	}

	log.Printf("Setting %s of: %v, resource type: %v, ", resourceName, valueAsString, resourceToSet)
	addResourceTypeToRequirements(requirements, resourceName, resourceToSet, valueAsString)
}

func addResourceTypeToRequirements(requirements *v1.ResourceRequirements, resourceName v1.ResourceName,
	resourceToSet resourceType, value string) {
	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		log.Printf("Failed to parse quantity: %v", value)
		return
	}

	if resourceToSet == Requests {
		patch.AddResourcesToRequests(requirements, resourceName, quantity)
	} else if resourceToSet == Limits {
		patch.AddResourcesToLimits(requirements, resourceName, quantity)
	}
}
