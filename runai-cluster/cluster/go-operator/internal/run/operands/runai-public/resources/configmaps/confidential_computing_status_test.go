/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configmaps

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const testInitdataConfigMap = "runai-cc-initdata"

func confidentialGpuNode(name string, mutate func(labels map[string]string)) *corev1.Node {
	labels := map[string]string{
		clusterconstants.KataNvidiaGpuNodeLabel:  "true",
		clusterconstants.AmdSevSnpNodeLabel:      "true",
		clusterconstants.NvidiaCCReadyStateLabel: "true",
		clusterconstants.KataRuntimeNodeLabel:    "true",
	}
	if mutate != nil {
		mutate(labels)
	}
	return &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels}}
}

func cpuNode(name string) *corev1.Node {
	return &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{
		clusterconstants.AmdSevSnpNodeLabel:   "true",
		clusterconstants.KataRuntimeNodeLabel: "true",
	}}}
}

func initdataConfigMap(content string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: testInitdataConfigMap, Namespace: namespace},
		Data:       map[string]string{initdataConfigMapKey: content},
	}
}

func confidentialStatus(initdataName string, objects ...client.Object) runv1.PublicStatus {
	kubeClient := fake.NewClientBuilder().WithObjects(objects...).Build()
	return getConfidentialComputingStatus(context.Background(), kubeClient, kubeClient, initdataName)
}

func TestConfidentialComputingStatus(t *testing.T) {
	tests := []struct {
		name              string
		initdata          string
		objects           []client.Object
		expectedAvailable bool
		expectedReason    string
	}{
		{
			name: "every GPU node ready",
			objects: []client.Object{
				confidentialGpuNode("gpu-1", nil),
				confidentialGpuNode("gpu-2", func(l map[string]string) {
					delete(l, clusterconstants.AmdSevSnpNodeLabel)
					l[clusterconstants.IntelTdxNodeLabel] = "true"
				}),
				cpuNode("cpu-1"),
			},
			expectedAvailable: true,
			expectedReason:    "",
		},
		{
			name: "some GPU nodes ready",
			objects: []client.Object{
				confidentialGpuNode("gpu-1", nil),
				confidentialGpuNode("gpu-2", func(l map[string]string) { l[clusterconstants.NvidiaCCReadyStateLabel] = "false" }),
			},
			expectedAvailable: true,
			expectedReason:    "1 of 2 GPU nodes are ready for confidential workloads",
		},
		{
			name: "no GPU node ready",
			objects: []client.Object{
				confidentialGpuNode("gpu-1", func(l map[string]string) { delete(l, clusterconstants.AmdSevSnpNodeLabel) }),
				confidentialGpuNode("gpu-2", func(l map[string]string) { delete(l, clusterconstants.KataRuntimeNodeLabel) }),
			},
			expectedAvailable: false,
			expectedReason:    "0 of 2 GPU nodes are ready for confidential workloads",
		},
		{
			name:              "KRAB not detected",
			objects:           []client.Object{cpuNode("cpu-1")},
			expectedAvailable: false,
			expectedReason: "no node is labeled kata.feature.node.kubernetes.io/nvidia-gpu=true; " +
				"KRAB is not installed or found no NVIDIA GPU",
		},
		{
			name:              "initdata ConfigMap present",
			initdata:          testInitdataConfigMap,
			objects:           []client.Object{confidentialGpuNode("gpu-1", nil), initdataConfigMap("version = \"0.1.0\"")},
			expectedAvailable: true,
			expectedReason:    "",
		},
		{
			name:              "initdata ConfigMap missing",
			initdata:          testInitdataConfigMap,
			objects:           []client.Object{confidentialGpuNode("gpu-1", nil)},
			expectedAvailable: false,
			expectedReason:    "initdata ConfigMap runai/runai-cc-initdata not found",
		},
		{
			name:              "initdata ConfigMap empty",
			initdata:          testInitdataConfigMap,
			objects:           []client.Object{confidentialGpuNode("gpu-1", nil), initdataConfigMap("  \n")},
			expectedAvailable: false,
			expectedReason:    "initdata ConfigMap runai/runai-cc-initdata has no initdata.toml key or it is empty",
		},
		{
			name:              "nodes and initdata both failing",
			initdata:          testInitdataConfigMap,
			objects:           []client.Object{confidentialGpuNode("gpu-1", func(l map[string]string) { delete(l, clusterconstants.KataRuntimeNodeLabel) })},
			expectedAvailable: false,
			expectedReason: "0 of 1 GPU nodes are ready for confidential workloads; " +
				"initdata ConfigMap runai/runai-cc-initdata not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := confidentialStatus(test.initdata, test.objects...)
			assert.Equal(t, test.expectedAvailable, *status.Available)
			assert.Equal(t, test.expectedReason, status.Reason)
		})
	}
}

func TestConfidentialComputingStatus_NodeListError(t *testing.T) {
	kubeClient := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
			return errors.New("nodes is forbidden")
		},
	}).Build()

	status := getConfidentialComputingStatus(context.Background(), kubeClient, kubeClient, "")
	assert.False(t, *status.Available)
	assert.Equal(t, "failed to list GPU nodes: nodes is forbidden", status.Reason)
}

func TestWithConfidentialComputingStatus(t *testing.T) {
	healthyStack := func() runv1.PublicStatus {
		return runv1.PublicStatus{
			Available:  runv1.AvailableStatus(true),
			Components: map[string]runv1.PublicStatus{"nvidia": {Available: runv1.AvailableStatus(true)}},
		}
	}

	t.Run("unavailable confidential computing fails the GPU stack", func(t *testing.T) {
		confidential := runv1.PublicStatus{Available: runv1.AvailableStatus(false), Reason: "0 of 1 GPU nodes"}

		status := withConfidentialComputingStatus(healthyStack(), confidential)
		assert.False(t, *status.Available)
		assert.Equal(t, "0 of 1 GPU nodes", status.Reason)
		assert.Equal(t, confidential, status.Components[confidentialComputingComponentName])
		assert.Contains(t, status.Components, "nvidia")
	})

	t.Run("available confidential computing keeps the GPU stack available", func(t *testing.T) {
		confidential := runv1.PublicStatus{Available: runv1.AvailableStatus(true), Reason: "1 of 2 GPU nodes"}

		status := withConfidentialComputingStatus(healthyStack(), confidential)
		assert.True(t, *status.Available)
		assert.Empty(t, status.Reason)
		assert.Equal(t, confidential, status.Components[confidentialComputingComponentName])
	})

	t.Run("an earlier GPU stack failure keeps its reason", func(t *testing.T) {
		stack := healthyStack()
		stack.Available = runv1.AvailableStatus(false)
		stack.Reason = "gpu fractioning unavailable"
		confidential := runv1.PublicStatus{Available: runv1.AvailableStatus(false), Reason: "0 of 1 GPU nodes"}

		status := withConfidentialComputingStatus(stack, confidential)
		assert.False(t, *status.Available)
		assert.Equal(t, "gpu fractioning unavailable", status.Reason)
	})
}
