/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_controller_mutators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/utils"
)

func TestSyncGpuAnnotationAndRequests(t *testing.T) {
	podController := utils.CreatePodController("my-namespace")
	container := &podController.Spec.Template.Spec.Containers[0]
	container.Resources.Limits = v1.ResourceList{
		types.GpuResource: resource.MustParse("500m"),
	}

	syncGpuAnnotationAndRequests(podController, container)

	specTemplateAnnotations := podController.Spec.Template.Annotations
	assert.Equal(t, len(specTemplateAnnotations), 1)
	assert.Equal(t, specTemplateAnnotations[types.RunaiGpuFraction], "0.500000")
	_, found := container.Resources.Limits[types.GpuResource]
	assert.False(t, found)
}

func TestSyncGpuAnnotationAndRequestsNoLimit(t *testing.T) {
	podController := utils.CreatePodController("my-namespace")
	container := &podController.Spec.Template.Spec.Containers[0]

	syncGpuAnnotationAndRequests(podController, container)

	specTemplateAnnotations := podController.Spec.Template.Annotations
	assert.Equal(t, len(specTemplateAnnotations), 0)
}

func TestSyncGpuAnnotationAndRequestsWholeGpu(t *testing.T) {
	podController := utils.CreatePodController("my-namespace")
	container := &podController.Spec.Template.Spec.Containers[0]
	container.Resources.Limits = v1.ResourceList{
		types.GpuResource: resource.MustParse("1"),
	}

	syncGpuAnnotationAndRequests(podController, container)

	specTemplateAnnotations := podController.Spec.Template.Annotations
	assert.Equal(t, len(specTemplateAnnotations), 0)
	_, found := container.Resources.Limits[types.GpuResource]
	assert.True(t, found)
}

func TestMutateJobResources(t *testing.T) {
	podController := utils.CreatePodController("my-namespace")
	container := &podController.Spec.Template.Spec.Containers[0]
	container.Resources.Limits = v1.ResourceList{
		types.GpuResource: resource.MustParse("0.5"),
	}

	mutator := createMutator("my-namespace")
	err := mutator.mutateJobResources(podController, nil)
	assert.Nil(t, err)

	specTemplateAnnotations := podController.Spec.Template.Annotations
	assert.Equal(t, len(specTemplateAnnotations), 1)
	assert.Equal(t, specTemplateAnnotations[types.RunaiGpuFraction], "0.500000")
	_, found := container.Resources.Limits[types.GpuResource]
	assert.False(t, found)
}
