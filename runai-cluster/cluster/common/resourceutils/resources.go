/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resourceutils

import (
	"github.com/rs/zerolog/log"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	engineresourcesutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// PodRunaiAllocatedResources returns the effective allocated resources for the pod - including from runai annotations
func PodRunaiAllocatedResources(containers []corev1.Container, annotations map[string]string) corev1.ResourceList {
	resources, _ := engineresourcesutils.PodRequestsAndLimits(containers, []corev1.Container{})

	AddRunaiAnnotationsAllocatedResources(annotations, resources)
	return resources
}

// AddRunaiAnnotationsAllocatedResources
// uses correct units for gpu memory - in MB
// mainly used by cluster-sync for new workload api, and by workload-exporter.
func AddRunaiAnnotationsAllocatedResources(podAnnotations map[string]string, resources corev1.ResourceList) {
	setResourceFromAnnotation(podAnnotations, clustercommonconstants.PodAllocatedGpusAnnotation, resources,
		clustercommonconstants.GpuResourceName,
		[]corev1.ResourceName{clustercommonconstants.NvidiaGpuResourceName, clustercommonconstants.AmdGpuResourceName},
		nil)
	setResourceFromAnnotation(podAnnotations, clustercommonconstants.PodAllocatedMigGpusAnnotation, resources,
		clustercommonconstants.GpuResourceName, []corev1.ResourceName{}, nil)

	// the annotation's expected value is a number in MB, while resource.ParseQuantity parses it into Bytes,
	// so we need to convert it to Bytes.
	var gpuMemoryMbToBytesConversion float64 = 1000 * 1000
	setResourceFromAnnotation(podAnnotations, clustercommonconstants.PodAllocatedGpuMemoryAnnotation, resources,
		clustercommonconstants.GpuMemoryResourceName, []corev1.ResourceName{}, &gpuMemoryMbToBytesConversion)
}

// setResourceFromAnnotation
// the user could have asked for 0.5 gpu ->
//
//	then we will get allocated from annotations - allocated 0.5 gpu and 10GB gpu-memory.
//
// or the user could have aksed for 10GB gpu-memory ->
//
//	then we will get allocated from annotations - allocated 0.5 gpu and 10GB gpu-memory.
//
// so we take the source of truth from annotations, and discard the other matching resources
// from what is received from "PodRunaiRequestsAndLimits".
// that's why we need the param resourceNamesToDelete - we will discard the in case we find allocated in annotations.
func setResourceFromAnnotation(podAnnotations map[string]string, annotationKey string, resources corev1.ResourceList,
	resourceName corev1.ResourceName, resourceNamesToDelete []corev1.ResourceName, conversion *float64) {
	quantityStr, found := podAnnotations[annotationKey]
	if !found || quantityStr == "" || quantityStr == "0" {
		return
	}

	quantity, err := resource.ParseQuantity(quantityStr)
	if err != nil {
		log.Error().Msgf("Failed to parse %s annotation %s",
			annotationKey, quantityStr)
		return
	}

	if conversion != nil {
		quantity.Set(int64(*conversion * quantity.AsApproximateFloat64()))
	}

	setResource(resources, resourceName, quantity)

	for _, resourceNameToDelete := range resourceNamesToDelete {
		delete(resources, resourceNameToDelete)
	}
}

func setResource(resourceList corev1.ResourceList, resourceName corev1.ResourceName, quantityToSet resource.Quantity) {
	resourceList[resourceName] = quantityToSet.DeepCopy()
}
