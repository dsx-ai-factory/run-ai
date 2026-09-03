/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_controller_mutators

import (
	"fmt"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/patch"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/utils"
)

const (
	MilliResourceQuantity = 1000.0
)

func (pcm *PodControllerMutator) mutateJobResources(podController *types.PodControllerCommon,
	_ *v1beta1.AdmissionRequest) error {
	containers := podController.Spec.Template.Spec.Containers
	containerIndex := utils.GetRunaiContainerIndex(podController.Spec.Template.Labels, containers)
	container := &podController.Spec.Template.Spec.Containers[containerIndex]
	syncGpuAnnotationAndRequests(podController, container)
	return nil
}

func syncGpuAnnotationAndRequests(podController *types.PodControllerCommon, container *v1.Container) {
	gpuResourceLimit, hasGPUResourceLimit := container.Resources.Limits[types.GpuResource]
	if !hasGPUResourceLimit || gpuResourceLimit.MilliValue() <= 0 ||
		gpuResourceLimit.MilliValue() >= MilliResourceQuantity {
		return
	}

	fractionalGpuRequest := float64(gpuResourceLimit.MilliValue()) / MilliResourceQuantity
	fractionStr := fmt.Sprintf("%f", fractionalGpuRequest)
	patch.AddAnnotationToObjectMeta(&podController.Spec.Template.ObjectMeta, types.RunaiGpuFraction, fractionStr)
	patch.RemoveGpuResourceFromLimits(&container.Resources)
}
