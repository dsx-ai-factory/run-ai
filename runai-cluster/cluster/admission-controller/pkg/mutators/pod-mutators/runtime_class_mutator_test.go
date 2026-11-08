/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	nodev1 "k8s.io/api/node/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/tests"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
)

const confidentialRuntimeClassName = "kata-qemu-nvidia-gpu-snp-runtime-rs"

func runtimeClassMutator(confidentialComputing bool, builder *fake.ClientBuilder, objs ...client.Object) *PodMutator {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(nodev1.AddToScheme(scheme))
	reader := builder.WithScheme(scheme).WithObjects(objs...).Build()
	return NewPodMutator(tests.NewClient(reader), &types.MutatingParams{ConfidentialComputing: confidentialComputing}, nil)
}

func runtimeClass(name string, nodeSelector map[string]string) *nodev1.RuntimeClass {
	return &nodev1.RuntimeClass{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Handler:    name,
		Scheduling: &nodev1.Scheduling{NodeSelector: nodeSelector},
	}
}

func confidentialGpuRuntimeClass() *nodev1.RuntimeClass {
	return runtimeClass(confidentialRuntimeClassName, map[string]string{
		ccReadyStateNodeLabel:                "true",
		"katacontainers.io/kata-runtime":     "true",
		"amd.feature.node.kubernetes.io/snp": "true",
	})
}

func cpuRuntimeClass() *nodev1.RuntimeClass {
	return runtimeClass("kata-qemu-nvidia-cpu-runtime-rs", map[string]string{"katacontainers.io/kata-runtime": "true"})
}

func gpuPod() *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: testNamespaceName},
		Spec: v1.PodSpec{Containers: []v1.Container{{
			Name: "main",
			Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{nvidiaGpuResourceName: resource.MustParse("1")},
			},
		}}},
	}
}

func TestConfidentialRuntimeClassMutator(t *testing.T) {
	t.Run("assigns the confidential runtime class and its node selector", func(t *testing.T) {
		pod := gpuPod()
		mutator := runtimeClassMutator(true, fake.NewClientBuilder(), cpuRuntimeClass(), confidentialGpuRuntimeClass())

		assert.NoError(t, mutator.mutateConfidentialRuntimeClass(pod, nil))
		assert.Equal(t, confidentialRuntimeClassName, *pod.Spec.RuntimeClassName)
		assert.Equal(t, "true", pod.Spec.NodeSelector[ccReadyStateNodeLabel])
	})

	t.Run("rejects a GPU pod when no confidential runtime class exists", func(t *testing.T) {
		pod := gpuPod()
		mutator := runtimeClassMutator(true, fake.NewClientBuilder(), cpuRuntimeClass())

		err := mutator.mutateConfidentialRuntimeClass(pod, nil)
		assert.ErrorContains(t, err, "no runtime class schedules onto nvidia.com/cc.ready.state=true nodes")
		assert.Nil(t, pod.Spec.RuntimeClassName)
	})

	t.Run("rejects a GPU pod when runtime classes cannot be listed", func(t *testing.T) {
		pod := gpuPod()
		builder := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return errors.New("runtimeclasses is forbidden")
			},
		})
		mutator := runtimeClassMutator(true, builder)

		err := mutator.mutateConfidentialRuntimeClass(pod, nil)
		assert.ErrorContains(t, err, "runtimeclasses is forbidden")
		assert.Nil(t, pod.Spec.RuntimeClassName)
	})

	t.Run("leaves a pod that names its own runtime class", func(t *testing.T) {
		pod := gpuPod()
		pod.Spec.RuntimeClassName = ptr.To("custom")
		mutator := runtimeClassMutator(true, fake.NewClientBuilder())

		assert.NoError(t, mutator.mutateConfidentialRuntimeClass(pod, nil))
		assert.Equal(t, "custom", *pod.Spec.RuntimeClassName)
	})

	t.Run("leaves a pod without GPUs", func(t *testing.T) {
		pod := gpuPod()
		pod.Spec.Containers[0].Resources = v1.ResourceRequirements{}
		mutator := runtimeClassMutator(true, fake.NewClientBuilder())

		assert.NoError(t, mutator.mutateConfidentialRuntimeClass(pod, nil))
		assert.Nil(t, pod.Spec.RuntimeClassName)
	})

	t.Run("does nothing when confidential computing is off", func(t *testing.T) {
		pod := gpuPod()
		mutator := runtimeClassMutator(false, fake.NewClientBuilder())

		assert.NoError(t, mutator.mutateConfidentialRuntimeClass(pod, nil))
		assert.Nil(t, pod.Spec.RuntimeClassName)
	})
}
