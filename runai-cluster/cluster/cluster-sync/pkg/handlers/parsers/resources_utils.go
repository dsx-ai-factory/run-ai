/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"context"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/common/node-pool-utils/converter"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func parseRequestedResources(requests, limits corev1.ResourceList) openapi.NullablePodRequestResources {
	result := openapi.PodRequestResources{}
	extendedResources := map[string]openapi.WorkloadsExtendedResource{}

	for resourceName, resourceQuantity := range requests {
		switch resourceName {
		case corev1.ResourceCPU:
			result.Cpu.Set(&openapi.RequestResourceCores{
				Request: getApiFloat64Quantity(&resourceQuantity),
			})
			continue
		case corev1.ResourceMemory:
			result.CpuMemory.Set(&openapi.RequestResourceQuantity{
				Request: getApiStringQuantity(&resourceQuantity),
			})
			continue
		case constants.GpuResourceName, constants.AmdGpuResourceName:
			result.GpuRequestType = getGpuRequestType(openapi.GPUREQUESTTYPE_PORTION)
			result.Gpu.Set(&openapi.RequestResourceCores{
				Request: getApiFloat64Quantity(&resourceQuantity),
			})
			continue
		case constants.GpuMemoryResourceName:
			result.GpuRequestType = getGpuRequestType(openapi.GPUREQUESTTYPE_MEMORY)
			result.GpuMemory.Set(&openapi.RequestResourceQuantity{
				Request: getApiStringQuantity(&resourceQuantity),
			})
			continue
		}

		extendedResources[string(resourceName)] = openapi.WorkloadsExtendedResource{
			Resource: *openapi.NewNullableString(ptr.To(string(resourceName))),
			Quantity: getApiStringQuantity(&resourceQuantity),
		}
	}

	for resourceName, resourceQuantity := range limits {
		switch resourceName {
		case corev1.ResourceCPU:
			cpu := result.Cpu.Get()
			if cpu == nil {
				cpu = &openapi.RequestResourceCores{}
			}
			cpu.Limit = getApiFloat64Quantity(&resourceQuantity)
			result.Cpu.Set(cpu)
			continue
		case corev1.ResourceMemory:
			cpuMemory := result.CpuMemory.Get()
			if cpuMemory == nil {
				cpuMemory = &openapi.RequestResourceQuantity{}
			}
			cpuMemory.Limit = getApiStringQuantity(&resourceQuantity)
			result.CpuMemory.Set(cpuMemory)
			continue
		case constants.GpuResourceName, constants.AmdGpuResourceName:
			result.GpuRequestType = getGpuRequestType(openapi.GPUREQUESTTYPE_PORTION)
			gpu := result.Gpu.Get()
			if gpu == nil {
				gpu = &openapi.RequestResourceCores{}
			}
			gpu.Limit = getApiFloat64Quantity(&resourceQuantity)
			result.Gpu.Set(gpu)
			continue
		case constants.GpuMemoryResourceName:
			result.GpuRequestType = getGpuRequestType(openapi.GPUREQUESTTYPE_MEMORY)
			gpuMemory := result.GpuMemory.Get()
			if gpuMemory == nil {
				gpuMemory = &openapi.RequestResourceQuantity{}
			}
			gpuMemory.Limit = getApiStringQuantity(&resourceQuantity)
			result.GpuMemory.Set(gpuMemory)
			continue
		}

		extendedResource, found := extendedResources[string(resourceName)]
		if found {
			extendedResource.Quantity = getApiStringQuantity(&resourceQuantity)
			extendedResources[string(resourceName)] = extendedResource
		} else {
			extendedResources[string(resourceName)] = openapi.WorkloadsExtendedResource{
				Resource: *openapi.NewNullableString(ptr.To(string(resourceName))),
				Quantity: getApiStringQuantity(&resourceQuantity),
			}
		}
	}

	if len(extendedResources) > 0 {
		if result.ExtendedResources == nil {
			result.ExtendedResources = []openapi.WorkloadsExtendedResource{}
		}
		for _, extendedResource := range extendedResources {
			result.ExtendedResources = append(result.ExtendedResources, extendedResource)
		}
	}

	nullableResult := openapi.NullablePodRequestResources{}
	nullableResult.Set(&result)
	return nullableResult
}

func parseAllocatedResources(allocatedResources corev1.ResourceList,
	requestedResources openapi.NullablePodRequestResources) openapi.NullableAllocatedResources {
	result := openapi.AllocatedResources{}

	for resourceName, resourceQuantity := range allocatedResources {
		switch resourceName {
		case corev1.ResourceCPU:
			result.Cpu = getApiFloat64Quantity(&resourceQuantity)
			continue
		case corev1.ResourceMemory:
			result.CpuMemory = getApiStringQuantity(&resourceQuantity)
			continue
		case clustercommonconstants.GpuResourceName:
			result.Gpu = getApiFloat64Quantity(&resourceQuantity)
			continue
		case clustercommonconstants.GpuMemoryResourceName:
			result.GpuMemory = getApiStringQuantity(&resourceQuantity)
			continue
		}
	}

	result.ExtendedResources = requestedResources.Get().ExtendedResources

	nullableResult := openapi.NullableAllocatedResources{}
	nullableResult.Set(&result)
	return nullableResult
}

func getPodRequestedNodePools(podMeta metav1.ObjectMeta,
	affinity *corev1.Affinity, cachedReader client.Reader) []string {

	nodePoolsSources := converter.NewNodePoolsSources(affinity, cachedReader, converter.IgnoreAffinityConversionErrors, podMeta.Annotations, nil)
	identifiers := converter.NodePoolIdentifiers{
		NodePoolAssignmentLabelKey: v1alpha1.NodePoolNameLabel,
		DefaultNodepoolName:        v1alpha1.DefaultNodePoolName,
		UnexistingNodepoolSentinel: v1alpha1.UnexistingRunaiNodePool,
		AnnotationNodepoolsKey:     commonconstants.AnnotationNodepools,
	}
	nodePools, err := converter.GetRequestedNodePools(context.Background(), podMeta, "Pod", identifiers, nodePoolsSources)
	if err != nil {
		// even if error occurred, we still want to continue
		zap.S().Warnf("For pod <%s/%s>, had an error when getting node pool names, err: <%s>",
			podMeta.Namespace, podMeta.Name, err.Error())
	}

	return nodePools
}

func parseResourcesForSchedulingCondition(resources corev1.ResourceList) *openapi.QueueResources {
	result := openapi.QueueResources{}

	for resourceName, resourceQuantity := range resources {
		switch resourceName {
		case clustercommonconstants.NvidiaGpuResourceName:
			result.Gpu = resourceQuantity.AsApproximateFloat64()
			continue
		case corev1.ResourceCPU:
			result.Cpu = resourceQuantity.AsApproximateFloat64()
			continue
		case corev1.ResourceMemory:
			result.Memory = resourceQuantity.AsApproximateFloat64()
			continue
		}
	}

	return &result
}

func multiplyNumberField(numberField *flds.NumberField, multiplyFactor int) *flds.NumberField {
	if numberField == nil {
		return numberField
	}

	floatQuantity := numberField.Deref()
	floatQuantity *= float64(multiplyFactor)
	return flds.NewOptionalNumberField(&floatQuantity)
}

func multiplyQuantityField(quantityField *flds.QuantityField, multiplyFactor int) *flds.QuantityField {
	if quantityField == nil {
		return quantityField
	}

	quantity := quantityField.Deref()
	parsedQuantity, err := resource.ParseQuantity(quantity)
	if err != nil {
		return quantityField
	}

	result := ptr.To(parsedQuantity.DeepCopy())
	for i := 0; i < multiplyFactor-1; i++ {
		result.Add(parsedQuantity)
	}
	return flds.NewQuantityField(result.String())
}

func multiplyExtendedResources(extendedResources *flds.ExtendedResourcesField, multiplyFactor int) *flds.ExtendedResourcesField {
	if extendedResources == nil {
		return extendedResources
	}

	result := map[string]flds.ExtendedResourcesItemAndRules{}

	for key, extendedResource := range extendedResources.Deref() {
		parsedQuantity := getInnerQuantity(extendedResource)
		if parsedQuantity == nil {
			continue
		}

		resultQuantity := ptr.To(parsedQuantity.DeepCopy())
		for i := 0; i < multiplyFactor-1; i++ {
			resultQuantity.Add(*parsedQuantity)
		}

		resultExtendedResource := (ptr.To((extendedResources.Deref())[key])).DeepCopy()
		resultExtendedResource.Value.Quantity = ptr.To(resultQuantity.String())
		result[key] = *resultExtendedResource
	}

	if len(result) == 0 {
		return nil
	}
	return flds.NewExtendedResourcesField(result)
}

func MultiplyResourceAllocationFields(resourceAllocationFields *field_groups.ResourceAllocationFields, multiplyFactor int) *field_groups.ResourceAllocationFields {
	if resourceAllocationFields == nil {
		return resourceAllocationFields
	}

	unifyGpuFields(resourceAllocationFields)

	result := &field_groups.ResourceAllocationFields{
		Gpu:               multiplyNumberField(resourceAllocationFields.Gpu, multiplyFactor),
		GpuLimit:          multiplyNumberField(resourceAllocationFields.GpuLimit, multiplyFactor),
		GpuMemory:         multiplyQuantityField(resourceAllocationFields.GpuMemory, multiplyFactor),
		GpuMemoryLimit:    multiplyQuantityField(resourceAllocationFields.GpuMemoryLimit, multiplyFactor),
		Cpu:               multiplyNumberField(resourceAllocationFields.Cpu, multiplyFactor),
		CpuLimit:          multiplyNumberField(resourceAllocationFields.CpuLimit, multiplyFactor),
		Memory:            multiplyQuantityField(resourceAllocationFields.Memory, multiplyFactor),
		MemoryLimit:       multiplyQuantityField(resourceAllocationFields.MemoryLimit, multiplyFactor),
		ExtendedResources: multiplyExtendedResources(resourceAllocationFields.ExtendedResources, multiplyFactor),
	}

	return result
}

func addNumberField(left, right *flds.NumberField) *flds.NumberField {
	if left == nil {
		return right
	}

	if right == nil {
		return left
	}

	result := left.Deref() + right.Deref()
	return flds.NewOptionalNumberField(&result)
}

func addQuantityField(left, right *flds.QuantityField) *flds.QuantityField {
	if left == nil {
		return right
	}

	if right == nil {
		return left
	}

	leftQuantity := left.Deref()
	rightQuantity := right.Deref()

	parsedLeftQuantity, err := resource.ParseQuantity(leftQuantity)
	if err != nil {
		return right
	}

	parsedRightQuantity, err := resource.ParseQuantity(rightQuantity)
	if err != nil {
		return left
	}

	result := ptr.To(parsedLeftQuantity.DeepCopy())
	result.Add(parsedRightQuantity)
	return flds.NewQuantityField(result.String())
}

func getInnerQuantity(er flds.ExtendedResourcesItemAndRules) *resource.Quantity {
	if er.Value == nil {
		return nil
	}

	quantity := er.Value.Quantity
	if quantity == nil {
		return nil
	}

	parsedQuantity, err := resource.ParseQuantity(*quantity)
	if err != nil {
		return nil
	}

	return &parsedQuantity
}

func addExtendedResources(left, right *flds.ExtendedResourcesField) *flds.ExtendedResourcesField {
	if left == nil {
		return right
	}

	if right == nil {
		return left
	}

	leftExtendedResources := left.Deref()
	rightExtendedResources := right.Deref()

	result := map[string]flds.ExtendedResourcesItemAndRules{}

	for key, leftExtendedResource := range leftExtendedResources {
		result[key] = leftExtendedResource
	}

	for key := range rightExtendedResources {
		if foundResult, ok := result[key]; ok {
			leftQuantity := getInnerQuantity(foundResult)
			if leftQuantity == nil {
				continue
			}

			rightExtendedResource := rightExtendedResources[key]
			rightQuantity := getInnerQuantity(rightExtendedResource)
			if rightQuantity == nil {
				continue
			}

			sumQuantities := ptr.To(leftQuantity.DeepCopy())
			sumQuantities.Add(*rightQuantity)
			result[key].Value.Quantity = ptr.To(sumQuantities.String())
		} else {
			rightExtendedResource := (ptr.To(rightExtendedResources[key])).DeepCopy()
			result[key] = *rightExtendedResource
		}
	}

	if len(result) == 0 {
		return nil
	}

	return flds.NewExtendedResourcesField(result)
}

func convertIntToNumberField(intField *flds.IntField) *flds.NumberField {
	if intField == nil {
		return nil
	}

	result := float64(intField.Deref())
	return flds.NewOptionalNumberField(&result)
}

func unifyGpuFields(resourceAllocationFields *field_groups.ResourceAllocationFields) {
	if resourceAllocationFields == nil {
		return
	}

	if resourceAllocationFields.Gpu == nil && resourceAllocationFields.GpuDevices != nil {
		resourceAllocationFields.Gpu = convertIntToNumberField(resourceAllocationFields.GpuDevices)
	}
	resourceAllocationFields.GpuDevices = nil
}

func AddResourceAllocationFields(left, right *field_groups.ResourceAllocationFields) *field_groups.ResourceAllocationFields {
	if left == nil {
		return right
	}

	if right == nil {
		return left
	}

	unifyGpuFields(left)
	unifyGpuFields(right)

	result := &field_groups.ResourceAllocationFields{
		Gpu:               addNumberField(left.Gpu, right.Gpu),
		GpuLimit:          addNumberField(left.GpuLimit, right.GpuLimit),
		GpuMemory:         addQuantityField(left.GpuMemory, right.GpuMemory),
		GpuMemoryLimit:    addQuantityField(left.GpuMemoryLimit, right.GpuMemoryLimit),
		Cpu:               addNumberField(left.Cpu, right.Cpu),
		CpuLimit:          addNumberField(left.CpuLimit, right.CpuLimit),
		Memory:            addQuantityField(left.Memory, right.Memory),
		MemoryLimit:       addQuantityField(left.MemoryLimit, right.MemoryLimit),
		ExtendedResources: addExtendedResources(left.ExtendedResources, right.ExtendedResources),
	}

	return result
}
