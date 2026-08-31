/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"log"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"
)

func GetEnvVarFromAnyContainer(containers []v1.Container, varName string) (bool, string) {
	for _, container := range containers {
		for _, envVar := range container.Env {
			if envVar.Name == varName {
				log.Printf("Environment variable '%s' found with value '%s'", envVar.Name, envVar.Value)
				return true, envVar.Value
			}
		}
	}
	return false, ""
}

func GetRunaiContainerIndex(labels map[string]string, containers []v1.Container) int {
	if len(containers) == 1 {
		return 0
	}

	for label := range labels {
		if strings.HasPrefix(label, "pipelines.kubeflow.org") {
			return 1
		}
	}

	return 0
}

func GetRequestedGpus(limits v1.ResourceList, podMeta *metav1.ObjectMeta) float64 {
	wholeGpus, found := limits[types.GpuResource]
	if found && wholeGpus.Value() != 0 {
		return float64(wholeGpus.Value())
	}

	gpuMemoryAnnotationValue, found := podMeta.Annotations[types.RunaiGpuMemory]
	if found && gpuMemoryAnnotationValue != "" {
		return 1
	}

	fractionGpu, found := podMeta.Annotations[types.RunaiGpuFraction]
	if found {
		fractionalGPU, err := strconv.ParseFloat(fractionGpu, 64)
		if err != nil {
			panic(err.Error())
		}
		return fractionalGPU
	}

	log.Printf("job requires no GPUs")
	return 0
}
