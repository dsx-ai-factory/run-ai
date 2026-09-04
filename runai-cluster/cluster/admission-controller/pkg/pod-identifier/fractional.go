/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_identifier

import (
	v1 "k8s.io/api/core/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/utils"
)

func IsFractionalPod(podAnnotations map[string]string, containers []v1.Container) bool {
	if HasGpuFractionalAnnotation(podAnnotations) {
		return true
	}
	hasFractionalGpuRequest, _ := HasFractionalEnvVar(containers)
	return hasFractionalGpuRequest
}

func HasGpuFractionalAnnotation(podAnnotations map[string]string) bool {
	_, hasGpuFractionAnnotation := podAnnotations[types.RunaiGpuFraction]
	_, hasGpuMemoryAnnotation := podAnnotations[types.RunaiGpuMemory]
	return hasGpuFractionAnnotation || hasGpuMemoryAnnotation
}

func HasFractionalEnvVar(containers []v1.Container) (bool, string) {
	exists, value := utils.GetEnvVarFromAnyContainer(containers, types.RunaiGpuFracEnvVar)
	return exists, value
}
