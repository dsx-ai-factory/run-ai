package resources

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func SumResources(left, right v1.ResourceList) v1.ResourceList {
	total := left.DeepCopy()
	if total == nil {
		total = make(v1.ResourceList)
	}

	for resourceName, resourceQuantity := range right {
		sum, seenResource := total[resourceName]
		if seenResource {
			sum.Add(resourceQuantity)
			total[resourceName] = sum
		} else {
			total[resourceName] = resourceQuantity.DeepCopy()
		}
	}
	return total
}

// PodRequestsAndLimits returns the effective resources for the pod - max of (sum containers, each init container)
func PodRequestsAndLimits(containers, initContainers []v1.Container) (reqs, limits v1.ResourceList) {
	reqs, limits = v1.ResourceList{}, v1.ResourceList{}
	for _, container := range containers {
		addResourceList(reqs, container.Resources.Requests)
		addResourceList(limits, container.Resources.Limits)
	}
	// init containers define the minimum of any resource
	for _, container := range initContainers {
		maxResourceList(reqs, container.Resources.Requests)
		maxResourceList(limits, container.Resources.Limits)
	}
	return
}

// GetExceedingResources returns the resources from request that exceed the capacity. If a resource is missing from capacity, it is considered to be 0.
func GetExceedingResources(request, capacity v1.ResourceList) v1.ResourceList {
	exceeding := make(v1.ResourceList)

	for resourceName, requestedQuantity := range request {
		capacityQuantity := capacity[resourceName]

		// we consider a negative value as unlimited
		if capacityQuantity.AsApproximateFloat64() < 0 {
			continue
		}

		// Compare requested and capacity quantities
		if requestedQuantity.Cmp(capacityQuantity) == 1 { // 1 means requestedQuantity > capacityQuantity
			requested := requestedQuantity
			requested.Sub(capacityQuantity)
			exceeding[resourceName] = requested
		}
	}

	return exceeding
}

// PodRunaiRequestsAndLimits returns the effective resources for the pod - including requests from runai annotations
func PodRunaiRequestsAndLimits(containers, initContainers []v1.Container, annotations map[string]string) (
	requests, limits v1.ResourceList) {
	requests, limits = PodRequestsAndLimits(containers, initContainers)
	AddRunaiAnnotationsRequestedResources(annotations, requests, limits)
	AddGpuLimitsFromContainers(containers, limits)
	return
}

// AddRunaiAnnotationsRequestedResources
// adds to both request and limit
// uses correct units for gpu memory - in MB
// mainly used by external-workload-integrator and cluster-sync for new workload api.
func AddRunaiAnnotationsRequestedResources(podAnnotations map[string]string,
	requests, limits v1.ResourceList) {
	// currently, gpu fraction, memory and mig from annotation are only reflected in the first container,
	// even if there are several containers. so we calculate it once (for each pod).

	// Add gpu memory request
	gpuMemoryMb, found := podAnnotations[constants.AnnotationGpuMemory]
	if found && gpuMemoryMb != "" {
		quantity, err := resource.ParseQuantity(gpuMemoryMb)
		if err != nil {
			zap.S().Warnf("Failed to parse %s annotation %s", constants.AnnotationGpuMemory, gpuMemoryMb)
		} else {
			// the annotation's expected value is a number in MB, while resource.ParseQuantity parses it into Bytes,
			// so we need to re-set it as MB.
			quantity.Set(int64(quantity.AsApproximateFloat64() * 1000 * 1000))
			addResource(requests, constants.GpuMemoryResourceName, quantity)
			addResource(limits, constants.GpuMemoryResourceName, quantity)
		}
	}

	// Add gpu fraction request. The gpu-fraction annotation carries the per-device
	// fraction (e.g. "0.5"); when the workload requests multiple devices at that
	// fraction (via gpu-fraction-num-devices), scale so the reported request/limit
	// reflects the total fractional GPUs the workload needs (fraction * device count).
	gpuFraction, found := podAnnotations[constants.AnnotationGpuFraction]
	if found && gpuFraction != "" {
		quantity, err := resource.ParseQuantity(gpuFraction)
		if err != nil {
			zap.S().Warnf("Failed to parse %s annotation %s", constants.AnnotationGpuFraction, gpuFraction)
		} else {
			if numDevicesStr, ok := podAnnotations[constants.AnnotationGpuFractionDevices]; ok && numDevicesStr != "" {
				if numDevices, parseErr := strconv.ParseInt(numDevicesStr, 10, 64); parseErr != nil {
					zap.S().Warnf("Failed to parse %s annotation %s", constants.AnnotationGpuFractionDevices, numDevicesStr)
				} else if numDevices > 1 {
					quantity = *resource.NewMilliQuantity(quantity.MilliValue()*numDevices, resource.DecimalSI)
				}
			}
			addResource(requests, constants.NvidiaGpuResourceName, quantity)
			addResource(limits, constants.NvidiaGpuResourceName, quantity)
		}
	}

	// Add mig requests
	for key, val := range podAnnotations {
		if !strings.HasPrefix(key, constants.AnnotationMigProfilePrefix) {
			continue
		}
		quantity, err := resource.ParseQuantity(val)
		if err != nil {
			zap.S().Warnf("Failed to parse mig %s annotation %s", key, val)
		} else {
			requests[v1.ResourceName(key)] = quantity
			limits[v1.ResourceName(key)] = quantity
		}
	}
}

func SumResourceRequirements(left, right v1.ResourceRequirements) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: SumResources(left.Requests, right.Requests),
		Limits:   SumResources(left.Limits, right.Limits),
	}
}

func AddGpuLimitsFromContainers(containers []v1.Container, limits v1.ResourceList) {
	if len(containers) == 0 {
		return
	}

	var gpuLimit *resource.Quantity = nil
	var err error

	for i, container := range containers {
		gpuLimit, err = GetGpuLimitsFromEnvVars(container.Env)
		if err != nil {
			zap.S().Warnf("Failed to get gpu limits from env vars of container %d, error: %s", i, err.Error())
			continue
		}
		if gpuLimit != nil {
			break
		}
	}
	if gpuLimit == nil {
		return
	}

	q := gpuLimit.AsApproximateFloat64()
	if q <= 1 {
		// if the value is less than or equal to 1,
		// it means it's the limit for GPU resources (a fraction or a whole GPU).
		limits[constants.NvidiaGpuResourceName] = *gpuLimit
	} else {
		// else - it means it's gpu-memory limit.
		limits[constants.GpuMemoryResourceName] = *gpuLimit
	}
}

func GetGpuLimitsFromEnvVars(envVars []v1.EnvVar) (*resource.Quantity, error) {
	for _, envVar := range envVars {
		if envVar.Name == constants.GpuMemoryLimitEnvVar {
			return parseGpuMemoryLimitValue(envVar.Value)
		}
	}
	return nil, nil
}

func parseGpuMemoryLimitValue(value string) (*resource.Quantity, error) {
	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		return nil, err
	}

	return &quantity, nil
}

// addResourceList adds the resources in newList to list
func addResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
		} else {
			value.Add(quantity)
			list[name] = value
		}
	}
}

// maxResourceList sets list to the greater of list/newList for every resource
// either list
func maxResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
			continue
		} else {
			if quantity.Cmp(value) > 0 {
				list[name] = quantity.DeepCopy()
			}
		}
	}
}

func addResource(resourceList v1.ResourceList, resourceName v1.ResourceName, quantityToAdd resource.Quantity) {
	existing, found := resourceList[resourceName]
	if !found {
		resourceList[resourceName] = quantityToAdd
		return
	}
	existing.Add(quantityToAdd)
	resourceList[resourceName] = existing
}

func EqualResourceLists(got, want v1.ResourceList) error {
	gotKeys := maps.Keys(got)
	sortResourceNames(gotKeys)
	wantKeys := maps.Keys(want)
	sortResourceNames(wantKeys)

	if !reflect.DeepEqual(gotKeys, wantKeys) {
		return fmt.Errorf("found different sets of keys that expected. got = %v, want %v", got, want)
	}

	for resourceName, quantity := range got {
		if quantity.Cmp(want[resourceName]) != 0 {
			return fmt.Errorf("found a different value than expected for resource %s . got = %v, want %v",
				resourceName, got, want)
		}
	}

	return nil
}

func sortResourceNames(list []v1.ResourceName) {
	sort.Slice(list, func(i, j int) bool {
		return list[i] < list[j]
	})
}
