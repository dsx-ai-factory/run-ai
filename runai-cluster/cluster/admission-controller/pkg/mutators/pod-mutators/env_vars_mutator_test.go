/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
)

func TestInjectJobEnvVarsFromPod(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-pod",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "my-container",
				},
			},
		},
	}
	container := &pod.Spec.Containers[0]
	injectJobEnvVarsFromPod(pod, container)

	assert.Equal(t, len(pod.Spec.Containers[0].Env), 4)
	for _, envVar := range pod.Spec.Containers[0].Env {
		switch envVar.Name {
		case "JOB_NAME", "jobName":
			assert.Equal(t, envVar.Value, "my-pod")
		case "JOB_UUID", "jobUUID":
			assert.Equal(t, envVar.ValueFrom.FieldRef.FieldPath, "metadata.uid")
		default:
			assert.Fail(t, "unexpected env var")
		}
	}
}

func TestInjectJobEnvVarsFromPodOwner(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-pod",
			OwnerReferences: []metav1.OwnerReference{
				{
					Name: "job-1",
					UID:  "12345",
				},
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "my-container",
				},
			},
		},
	}
	container := &pod.Spec.Containers[0]
	injectJobEnvVarsFromPodOwner(pod, container)

	assert.Equal(t, len(pod.Spec.Containers[0].Env), 4)
	for _, envVar := range pod.Spec.Containers[0].Env {
		switch envVar.Name {
		case "JOB_NAME", "jobName":
			assert.Equal(t, envVar.Value, "job-1")
		case "JOB_UUID", "jobUUID":
			assert.Equal(t, envVar.Value, "12345")
		default:
			assert.Fail(t, "unexpected env var")
		}
	}
}

func TestInjectGpuNumEnvironmentVariable(t *testing.T) {
	container := &v1.Container{
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{types.GpuResource: resource.MustParse("1")},
		},
	}

	injectGpuNumEnvironmentVariable(container)
	assert.Equal(t, len(container.Env), 1)
	assert.Equal(t, container.Env[0].Name, "RUNAI_NUM_OF_GPUS")
	assert.Equal(t, container.Env[0].Value, "1")
}

func TestInjectGpuNumEnvironmentVariableNoLimits(t *testing.T) {
	container := &v1.Container{
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1")},
		},
	}

	injectGpuNumEnvironmentVariable(container)
	assert.Equal(t, len(container.Env), 0)
}

func TestInjectReporterLibraryEnvVar(t *testing.T) {
	container := &v1.Container{}
	injectEnvVars(container)

	assert.Equal(t, len(container.Env), 3)
	for _, envVar := range container.Env {
		switch envVar.Name {
		case "podUUID", "POD_UUID":
			assert.Equal(t, envVar.ValueFrom.FieldRef.FieldPath, "metadata.uid")
		case "NODE_NAME":
			assert.Equal(t, envVar.ValueFrom.FieldRef.FieldPath, "spec.nodeName")
		default:
			assert.Fail(t, "unexpected env var")
		}
	}
}

func TestSetEnvVariablesForPod(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-pod",
			OwnerReferences: []metav1.OwnerReference{
				{
					Name: "job-1",
					UID:  "12345",
				},
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{types.GpuResource: resource.MustParse("1")},
					},
					Name: "my-container",
				},
			},
		},
	}
	mutator := NewPodMutator(nil, nil, nil)
	err := mutator.setEnvVariablesForPod(pod, nil)
	assert.Nil(t, err)
	assert.Equal(t, len(pod.Spec.Containers[0].Env), 8)
}

func TestSetEnvVariablesForPodJob(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-pod",
		},
		Spec: v1.PodSpec{
			SchedulerName: types.RunaiScheduler,
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{types.GpuResource: resource.MustParse("1")},
					},
					Name: "my-container",
				},
			},
		},
	}
	mutator := NewPodMutator(nil, nil, nil)
	err := mutator.setEnvVariablesForPod(pod, nil)
	assert.Nil(t, err)
	assert.Equal(t, len(pod.Spec.Containers[0].Env), 8)
}
