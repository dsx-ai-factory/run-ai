/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/kube-client/mock"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/patch"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/component-helpers/scheduling/corev1/nodeaffinity"
)

func TestGuardedGPUPlacementUpgrade(t *testing.T) {
	const gpuKey = "nvidia.com/gpu.product"
	resourceLabel := constants.AcceleratorResourceNodeLabel("nvidia.com/gpu")
	in := func(key, value string) v1.NodeSelectorRequirement {
		return v1.NodeSelectorRequirement{Key: key, Operator: v1.NodeSelectorOpIn, Values: []string{value}}
	}
	guard := []v1.NodeSelectorRequirement{in("metadata.name", constants.GPUPlacementAdmissionNode), {
		Key: "metadata.name", Operator: v1.NodeSelectorOpNotIn, Values: []string{constants.GPUPlacementAdmissionNode},
	}}
	caller := v1.NodeSelectorRequirement{Key: "metadata.name", Operator: v1.NodeSelectorOpNotIn, Values: []string{"excluded-node"}}
	makePod := func() *v1.Pod {
		return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testNamespaceName, Annotations: map[string]string{
			constants.AnnotationGPUPlacementResource: "nvidia.com/gpu", constants.AnnotationGPUTypeLabel: gpuKey + "=A100", constants.AnnotationNodepools: "chosen",
		}}, Spec: v1.PodSpec{Affinity: &v1.Affinity{NodeAffinity: &v1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
			{MatchExpressions: []v1.NodeSelectorRequirement{in(resourceLabel, "true"), in(gpuKey, "A100")}, MatchFields: append(append([]v1.NodeSelectorRequirement{}, guard...), caller)},
		}}}}}}
	}
	match := func(pod *v1.Pod, name string, labels map[string]string, expected bool) {
		t.Helper()
		matches, err := nodeaffinity.GetRequiredNodeAffinity(pod).Match(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels}})
		require.NoError(t, err)
		require.Equal(t, expected, matches, "node %s, labels %v", name, labels)
	}
	a100 := map[string]string{gpuKey: "A100", resourceLabel: "true", "pool": "chosen"}
	h100 := map[string]string{gpuKey: "H100", resourceLabel: "true", "pool": "chosen"}
	pod := makePod()
	match(pod, "a100-node", a100, false)
	// This helper is the same-key replacement used by the older webhook.
	patch.AddNodeAffinityToPodSpec(&pod.Spec, []v1.NodeSelectorRequirement{in(gpuKey, "H100")})
	match(pod, "h100-node", h100, false)
	match(pod, constants.GPUPlacementAdmissionNode, h100, false)

	pod = makePod()
	kube := mock.NewMockClient(gomock.NewController(t))
	ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"runai/queue": testProjectName}}}
	kube.EXPECT().GetNamespace(testNamespaceName).Return(ns, nil).AnyTimes()
	kube.EXPECT().GetProject(testProjectName).Return(&unstructured.Unstructured{}, nil).AnyTimes()
	pool := unstructured.Unstructured{Object: map[string]any{"spec": map[string]any{"labelKey": "pool", "labelValue": "chosen"}, "status": map[string]any{"phase": "Ready"}}}
	kube.EXPECT().GetNodePool("chosen").Return(&pool, nil)
	pm := createMutator(testNamespaceName, false)
	pm.client = kube
	require.NoError(t, pm.mutateNodeAffinityForDefaultNodePools(pod, nil))
	require.NoError(t, pm.enforceSelectedGPUPlacement(pod))
	match(pod, "a100-node", a100, true)
	match(pod, "h100-node", h100, false)
	match(pod, "excluded-node", a100, false)
	match(pod, "other-resource", map[string]string{gpuKey: "A100", "pool": "chosen"}, false)
	match(pod, "other-pool", map[string]string{gpuKey: "A100", resourceLabel: "true", "pool": "other"}, false)
	before := pod.DeepCopy()
	require.NoError(t, pm.enforceSelectedGPUPlacement(pod))
	require.Equal(t, before, pod)

	// An older controller may already have overwritten the model expression.
	// Pinning the model independently must leave that contradiction unschedulable.
	pod = makePod()
	patch.AddNodeAffinityToPodSpec(&pod.Spec, []v1.NodeSelectorRequirement{in(gpuKey, "H100")})
	kube.EXPECT().GetNodePool("chosen").Return(&pool, nil)
	require.NoError(t, pm.enforceSelectedGPUPlacement(pod))
	match(pod, "h100-node", h100, false)
	match(pod, "a100-node", a100, false)

	plain := &v1.Pod{}
	require.NoError(t, pm.enforceSelectedGPUPlacement(plain))
	require.Equal(t, &v1.Pod{}, plain)

	// If a pool disappeared after submission, admission must not release the barrier.
	pod = makePod()
	kube.EXPECT().GetNodePool("chosen").Return(nil, errors.NewNotFound(schema.GroupResource{Resource: "nodepools"}, "chosen"))
	require.Error(t, pm.enforceSelectedGPUPlacement(pod))
	match(pod, "a100-node", a100, false)

	pod = makePod()
	delete(pod.Annotations, constants.AnnotationNodepools)
	pod.Annotations[constants.AnnotationGPUTypeLabel] = "broken"
	require.Error(t, pm.enforceSelectedGPUPlacement(pod))
	match(pod, "a100-node", a100, false)
	pod = makePod()
	delete(pod.Annotations, constants.AnnotationNodepools)
	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	terms[0].MatchFields = terms[0].MatchFields[:1]
	require.Error(t, pm.enforceSelectedGPUPlacement(pod))

	pod = makePod()
	delete(pod.Annotations, constants.AnnotationNodepools)
	terms = pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	for i := range 2 {
		terms[0].MatchFields[i].Values = []string{constants.GPUPlacementAdmissionNode + "-empty"}
	}
	require.NoError(t, pm.enforceSelectedGPUPlacement(pod))
	match(pod, "a100-node", a100, false)

	// Even a caller's own identical intentional contradiction survives reinvocation.
	pod = makePod()
	delete(pod.Annotations, constants.AnnotationNodepools)
	term := &pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0]
	term.MatchFields = append(append([]v1.NodeSelectorRequirement{}, guard...), guard...)
	require.NoError(t, pm.enforceSelectedGPUPlacement(pod))
	require.NoError(t, pm.enforceSelectedGPUPlacement(pod))
	require.Equal(t, guard, term.MatchFields)
	match(pod, "a100-node", a100, false)
}
