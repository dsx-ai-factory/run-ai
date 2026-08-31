/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"fmt"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/patch"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/kube-client/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func getFieldFromName(name string, field string) string {
	return fmt.Sprintf("%s%s", name, field)
}

func createMockedNodePool(m *mock.MockClient, name string, phase string, mockRegistry bool) unstructured.Unstructured {
	nodePool := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"labelKey":   getFieldFromName(name, "Key"),
				"labelValue": getFieldFromName(name, "Value"),
			},
			"status": map[string]interface{}{
				"phase": phase,
			},
		},
	}
	if mockRegistry {
		m.EXPECT().GetNodePool(name).Return(&nodePool, nil)
	}
	return nodePool
}

func TestProjectDefaultNodePoolsSanity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mock.NewMockClient(ctrl)
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}
	project := unstructured.Unstructured{}

	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateNodeAffinityForDefaultNodePools(&pod, nil)
	assert.Nil(t, err)
}

func TestNodepoolLabelBackwardsCompatibility(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mock.NewMockClient(ctrl)
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	npNameA := "np-a"
	createMockedNodePool(m, npNameA, "Ready", false)
	npNameB := "np-b"
	createMockedNodePool(m, npNameB, "Empty", true)

	nodePoolsNames := []interface{}{npNameA, npNameB}
	project := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": testProjectName,
			},
			"spec": map[string]interface{}{
				"defaultNodePools": nodePoolsNames,
			},
		},
	}
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
			Labels: map[string]string{
				patch.NodePoolNameLabel: "np-b",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateNodeAffinityForDefaultNodePools(&pod, nil)
	assert.Nil(t, err)

	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	assert.Equal(t, 1, len(terms))
	validateExpressionInOp(t, npNameB, terms[0].MatchExpressions[0])
}

func TestDefaultListWithNoAffinity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mock.NewMockClient(ctrl)
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	npNameA := "np-a"
	createMockedNodePool(m, npNameA, "Ready", true)
	npNameB := "np-b"
	createMockedNodePool(m, npNameB, "Empty", true)

	nodePoolsNames := []interface{}{npNameA, npNameB}
	project := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": testProjectName,
			},
			"spec": map[string]interface{}{
				"defaultNodePools": nodePoolsNames,
			},
		},
	}
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateNodeAffinityForDefaultNodePools(&pod, nil)
	assert.Nil(t, err)

	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	validateExpressionInOp(t, npNameA, terms[0].MatchExpressions[0])
	validateExpressionInOp(t, npNameB, terms[1].MatchExpressions[0])
}

func TestDefaultListWithAffinity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mock.NewMockClient(ctrl)
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	npNameA := "np-a"
	npA := createMockedNodePool(m, npNameA, "Ready", true)
	npNameB := "np-b"
	npB := createMockedNodePool(m, npNameB, "Empty", true)

	nodePoolsNames := []interface{}{npNameA, npNameB}
	project := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": testProjectName,
			},
			"spec": map[string]interface{}{
				"defaultNodePools": nodePoolsNames,
			},
		},
	}
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)
	m.EXPECT().GetNodePoolsList().Return(&unstructured.UnstructuredList{Items: []unstructured.Unstructured{npA, npB}}, nil)

	prevNodePoolName1 := "prevnp1"
	prevNodePoolName2 := "prevnp2"
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Spec: v1.PodSpec{
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      getFieldFromName(prevNodePoolName1, "Key"),
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{getFieldFromName(prevNodePoolName1, "Value")},
									},
								},
							},
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      getFieldFromName(prevNodePoolName2, "Key"),
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{getFieldFromName(prevNodePoolName2, "Value")},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateNodeAffinityForDefaultNodePools(&pod, nil)
	assert.Nil(t, err)

	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	assert.Equal(t, 4, len(terms))
	assert.Equal(t, 2, len(terms[0].MatchExpressions))
	assert.Equal(t, 2, len(terms[1].MatchExpressions))
	assert.Equal(t, 2, len(terms[2].MatchExpressions))
	assert.Equal(t, 2, len(terms[3].MatchExpressions))

	validateExpressionInOp(t, prevNodePoolName1, terms[0].MatchExpressions[0])
	validateExpressionInOp(t, npNameA, terms[0].MatchExpressions[1])
	validateExpressionInOp(t, prevNodePoolName2, terms[1].MatchExpressions[0])
	validateExpressionInOp(t, npNameA, terms[1].MatchExpressions[1])
	validateExpressionInOp(t, prevNodePoolName1, terms[2].MatchExpressions[0])
	validateExpressionInOp(t, npNameB, terms[2].MatchExpressions[1])
	validateExpressionInOp(t, prevNodePoolName2, terms[3].MatchExpressions[0])
	validateExpressionInOp(t, npNameB, terms[3].MatchExpressions[1])
}

func TestHasNodePoolInAffinity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mock.NewMockClient(ctrl)
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	npNameA := "np-a"
	npA := createMockedNodePool(m, npNameA, "Ready", false)
	npNameB := "np-b"
	npB := createMockedNodePool(m, npNameB, "Empty", false)

	nodePoolsNames := []interface{}{npNameA}
	project := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": testProjectName,
			},
			"spec": map[string]interface{}{
				"defaultNodePools": nodePoolsNames,
			},
		},
	}
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)
	m.EXPECT().GetNodePoolsList().Return(&unstructured.UnstructuredList{Items: []unstructured.Unstructured{npA, npB}}, nil)

	differentValue := "SomeDifferentValue"
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Spec: v1.PodSpec{
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      getFieldFromName(npNameB, "Key"),
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{differentValue, getFieldFromName(npNameB, "Value")},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateNodeAffinityForDefaultNodePools(&pod, nil)
	assert.Nil(t, err)

	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	assert.Equal(t, 1, len(terms))
	assert.Equal(t, 1, len(terms[0].MatchExpressions))

	expression := terms[0].MatchExpressions[0]
	assert.Equal(t, 2, len(expression.Values))
	assert.Equal(t, getFieldFromName(npNameB, "Key"), expression.Key)
	assert.Equal(t, v1.NodeSelectorOpIn, expression.Operator)
	assert.Equal(t, differentValue, expression.Values[0])
	assert.Equal(t, getFieldFromName(npNameB, "Value"), expression.Values[1])
}

func TestHasAffinityWithNotInOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mock.NewMockClient(ctrl)
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	npNameA := "np-a"
	npA := createMockedNodePool(m, npNameA, "Ready", true)

	nodePoolsNames := []interface{}{npNameA}
	project := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": testProjectName,
			},
			"spec": map[string]interface{}{
				"defaultNodePools": nodePoolsNames,
			},
		},
	}
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)
	m.EXPECT().GetNodePoolsList().Return(&unstructured.UnstructuredList{Items: []unstructured.Unstructured{npA}}, nil)

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Spec: v1.PodSpec{
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      getFieldFromName(npNameA, "Key"),
										Operator: v1.NodeSelectorOpNotIn,
										Values:   []string{getFieldFromName(npNameA, "Value")},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateNodeAffinityForDefaultNodePools(&pod, nil)
	assert.Nil(t, err)

	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	assert.Equal(t, 1, len(terms))
	assert.Equal(t, 1, len(terms[0].MatchExpressions))

	validateExpressionInOp(t, npNameA, terms[0].MatchExpressions[0])
}

func TestDeletingNodePool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mock.NewMockClient(ctrl)
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	npNameA := "np-a"
	createMockedNodePool(m, npNameA, patch.NodePoolDeleting, true)
	npNameB := "np-b"
	createMockedNodePool(m, npNameB, "Ready", true)

	nodePoolsNames := []interface{}{npNameA, npNameB}
	project := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": testProjectName,
			},
			"spec": map[string]interface{}{
				"defaultNodePools": nodePoolsNames,
			},
		},
	}
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateNodeAffinityForDefaultNodePools(&pod, nil)
	assert.Nil(t, err)

	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	assert.Equal(t, 1, len(terms))
	assert.Equal(t, 1, len(terms[0].MatchExpressions))

	validateExpressionInOp(t, npNameB, terms[0].MatchExpressions[0])
}

func validateExpressionInOp(t *testing.T, name string, expression v1.NodeSelectorRequirement) {
	assert.Equal(t, getFieldFromName(name, "Key"), expression.Key)
	assert.Equal(t, v1.NodeSelectorOpIn, expression.Operator)
	assert.Equal(t, getFieldFromName(name, "Value"), expression.Values[0])
}
