/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/patch"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/kube-client/mock"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/component-helpers/scheduling/corev1/nodeaffinity"
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

func TestExplicitGPUSelectionRequiresPoolsOnAllBranches(t *testing.T) {
	const key = "nvidia.com/gpu.product"
	model := v1.NodeSelectorRequirement{Key: key, Operator: v1.NodeSelectorOpIn, Values: []string{"H100"}}
	pools := map[KeyValuePair]bool{{Key: key, Value: "H100"}: true}
	terms := []v1.NodeSelectorTerm{{MatchExpressions: []v1.NodeSelectorRequirement{model}}}
	assert.True(t, hasNodePoolInNodeAffinity(terms, pools, true))
	assert.True(t, hasNodePoolInNodeAffinity(terms, pools, false))
	terms = append(terms, v1.NodeSelectorTerm{MatchFields: []v1.NodeSelectorRequirement{{Key: "metadata.name", Operator: v1.NodeSelectorOpIn, Values: []string{"other-worker"}}}})
	assert.False(t, hasNodePoolInNodeAffinity(terms, pools, true))
	assert.True(t, hasNodePoolInNodeAffinity(terms, pools, false))
}

func TestGeneratedGPURequirementsKeepProjectPoolDefaults(t *testing.T) {
	const modelKey = "nvidia.com/gpu.product"
	resourceKey := constants.AcceleratorResourceNodeLabel("nvidia.com/gpu")
	in := func(key, value string) v1.NodeSelectorRequirement {
		return v1.NodeSelectorRequirement{Key: key, Operator: v1.NodeSelectorOpIn, Values: []string{value}}
	}

	for _, tc := range []struct {
		name, model, explicitPool string
		callerModel               bool
	}{
		{name: "resource-only inference"},
		{name: "model and resource both coincide with pools", model: "H100"},
		{name: "resource only with a converted caller nodeSelector", callerModel: true},
		{name: "explicit resource pool", explicitPool: "nvidia"},
		{name: "explicit model pool", model: "H100", explicitPool: "h100s"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := mock.NewMockClient(gomock.NewController(t))
			project := defaultTestProject()
			project.Object["spec"] = map[string]any{defaultNodePoolsFieldName: []any{"blue"}}
			client.EXPECT().GetNamespace(testNamespaceName).Return(defaultTestNamespace(), nil).AnyTimes()
			client.EXPECT().GetProject(testProjectName).Return(project, nil).AnyTimes()
			list := &unstructured.UnstructuredList{}
			for _, pool := range []struct{ name, key, value string }{{"blue", "fabric", "blue"}, {"nvidia", resourceKey, "true"}, {"h100s", modelKey, "H100"}} {
				obj := unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": pool.name},
					"spec": map[string]any{"labelKey": pool.key, "labelValue": pool.value}, "status": map[string]any{"phase": "Ready"}}}
				list.Items = append(list.Items, obj)
				client.EXPECT().GetNodePool(pool.name).Return(&obj, nil).AnyTimes()
			}
			client.EXPECT().GetNodePoolsList().Return(list, nil).AnyTimes()
			pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testNamespaceName, Annotations: map[string]string{}}}
			patch.AddPlacementRequirements(&pod.Spec, pod.Annotations, []v1.NodeSelectorRequirement{in(resourceKey, "true")})
			if tc.model != "" {
				pod.Annotations[constants.AnnotationGPUTypeLabel] = modelKey + "=" + tc.model
			}
			if tc.model != "" || tc.callerModel {
				patch.AddPlacementRequirements(&pod.Spec, pod.Annotations, []v1.NodeSelectorRequirement{in(modelKey, "H100")})
			}
			if tc.explicitPool != "" {
				pod.Annotations[constants.AnnotationNodepools] = tc.explicitPool
			}
			mutator := &PodMutator{client: client}
			for range 2 {
				require.NoError(t, mutator.mutateNodeAffinityForDefaultNodePools(pod, nil))
				for _, fabric := range []string{"red", "blue"} {
					node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: fabric, Labels: map[string]string{
						resourceKey: "true", modelKey: "H100", "fabric": fabric,
					}}}
					eligible, err := nodeaffinity.GetRequiredNodeAffinity(pod).Match(node)
					require.NoError(t, err)
					assert.Equal(t, fabric == "blue" || tc.explicitPool != "", eligible, fabric)
				}
			}
		})
	}
}

func TestAdmissionKeepsGpuModelAndConflictingDefaultPool(t *testing.T) {
	const gpuKey = "nvidia.com/gpu.product"
	m := mock.NewMockClient(gomock.NewController(t))
	project := &unstructured.Unstructured{Object: map[string]interface{}{
		"spec": map[string]interface{}{defaultNodePoolsFieldName: []interface{}{"default-h100"}},
	}}
	namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"runai/queue": testProjectName}}}
	m.EXPECT().GetNamespace(testNamespaceName).Return(namespace, nil).Times(2)
	m.EXPECT().GetProject(testProjectName).Return(project, nil).Times(2)

	makePool := func(name, model string) unstructured.Unstructured {
		return unstructured.Unstructured{Object: map[string]interface{}{
			"metadata": map[string]interface{}{"name": name},
			"spec":     map[string]interface{}{"labelKey": gpuKey, "labelValue": model},
			"status":   map[string]interface{}{"phase": "Ready"},
		}}
	}
	defaultPool := makePool("default-h100", "H100")
	m.EXPECT().GetNodePool("default-h100").Return(&defaultPool, nil).Times(2)
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testNamespaceName, Annotations: map[string]string{constants.AnnotationGPUTypeLabel: gpuKey + "=A100"}}}
	model := v1.NodeSelectorRequirement{Key: gpuKey, Operator: v1.NodeSelectorOpIn, Values: []string{"A100"}}
	pool := v1.NodeSelectorRequirement{Key: gpuKey, Operator: v1.NodeSelectorOpIn, Values: []string{"H100"}}
	patch.AddPlacementRequirements(&pod.Spec, pod.Annotations, []v1.NodeSelectorRequirement{model})

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	for range 2 {
		require.NoError(t, mutator.mutateNodeAffinityForDefaultNodePools(pod, nil))
		require.Equal(t, []v1.NodeSelectorTerm{{MatchExpressions: []v1.NodeSelectorRequirement{model, pool}}}, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)
	}
}
