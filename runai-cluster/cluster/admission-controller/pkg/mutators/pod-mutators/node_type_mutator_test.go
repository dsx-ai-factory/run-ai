/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"encoding/json"
	"testing"

	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/kube-client/mock"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"
)

func TestMutateNodeAffinitySanity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}

	project := unstructured.Unstructured{}

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)
	m.EXPECT().GetCachedReader().Return(fakeCachedReader()).AnyTimes()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		Spec: v1.PodSpec{
			SchedulerName: types.RunaiScheduler,
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateNodeAffinityForNodeType(&pod, nil)
	assert.Nil(t, err)
}

func TestMutateNodeAffinityInteractive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}

	affinitySlice := []string{"aaa", "bbb", "ccc"}
	project := projectWithNodeAffinity(true, affinitySlice)

	// A custom (non-builtin) priority class with a value at/above the preemption threshold must be
	// treated as interactive. This is the regression: the old name-based check returned
	// train for any priority class name that wasn't "build"/"interactive-preemptible".
	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)
	m.EXPECT().GetCachedReader().Return(fakeCachedReader(priorityClass("research-high", 120))).AnyTimes()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
			Labels: map[string]string{
				"priorityClassName": "research-high",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Spec: v1.PodSpec{
			SchedulerName: types.RunaiScheduler,
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateNodeAffinityForNodeType(&pod, nil)
	assert.Nil(t, err)
	expression := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].
		MatchExpressions[0]
	assert.Equal(t, "run.ai/type", expression.Key)
	assert.Equal(t, v1.NodeSelectorOpIn, expression.Operator)
	assert.Equal(t, affinitySlice, expression.Values)
}

func TestMutateNodeAffinityTrain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}

	affinitySlice := []string{"aaa", "bbb", "ccc"}
	project := projectWithNodeAffinity(false, affinitySlice)

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)
	m.EXPECT().GetCachedReader().Return(fakeCachedReader(priorityClass("train", 50))).AnyTimes()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
			Labels: map[string]string{
				"priorityClassName": "train",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Spec: v1.PodSpec{
			SchedulerName: types.RunaiScheduler,
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateNodeAffinityForNodeType(&pod, nil)
	assert.Nil(t, err)
	expression := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].
		MatchExpressions[0]
	assert.Equal(t, "run.ai/type", expression.Key)
	assert.Equal(t, v1.NodeSelectorOpIn, expression.Operator)
	assert.Equal(t, affinitySlice, expression.Values)
}

// The preemptibility label (kai.scheduler/preemptibility) takes precedence over the priority class:
// a non-preemptible workload gets the interactive affinity even with no priority class set.
func TestMutateNodeAffinityFromPreemptibilityLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}

	affinitySlice := []string{"aaa", "bbb", "ccc"}
	project := projectWithNodeAffinity(true, affinitySlice)

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)
	m.EXPECT().GetCachedReader().Return(fakeCachedReader()).AnyTimes()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
			Labels: map[string]string{
				"kai.scheduler/preemptibility": "non-preemptible",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Spec: v1.PodSpec{
			SchedulerName: types.RunaiScheduler,
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateNodeAffinityForNodeType(&pod, nil)
	assert.Nil(t, err)
	expression := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].
		MatchExpressions[0]
	assert.Equal(t, "run.ai/type", expression.Key)
	assert.Equal(t, v1.NodeSelectorOpIn, expression.Operator)
	assert.Equal(t, affinitySlice, expression.Values)
}

func TestMutateKubevirtNodeAffinity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}
	project := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": testProjectName,
			},
		},
	}

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)
	m.EXPECT().GetCachedReader().Return(fakeCachedReader(priorityClass("train", 50))).AnyTimes()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
			Labels: map[string]string{
				"priorityClassName": "train",
			},
		},
		Spec: v1.PodSpec{
			SchedulerName: types.RunaiScheduler,
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	mutator.params.IsKubevirtEnabledCluster = true
	err := mutator.mutateNodeAffinityForNodeType(&pod, nil)
	assert.Nil(t, err)
	expression := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].
		MatchExpressions[0]
	assert.Equal(t, kubevirtAntiNodeAffinityKey, expression.Key)
	assert.Equal(t, v1.NodeSelectorOpNotIn, expression.Operator)
	assert.Equal(t, []string{"true"}, expression.Values)
}

func TestMutateKubevirtV1NodeAffinity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
			Labels: map[string]string{
				"priorityClassName": "train",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: kubevirtObjAPIVersion,
				},
			},
		},
		Spec: v1.PodSpec{
			SchedulerName: types.RunaiScheduler,
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	mutator.params.IsKubevirtEnabledCluster = true
	err := mutator.mutateNodeAffinityForNodeType(&pod, nil)
	assert.Nil(t, err)
	assert.Nil(t, pod.Spec.Affinity)
}

func getNodeAffinitySlice(affinity []string) []interface{} {
	interfaceSlice := make([]interface{}, len(affinity))
	for i, v := range affinity {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}

func fakeCachedReader(objs ...client.Object) client.Reader {
	scheme := runtime.NewScheme()
	utilruntime.Must(schedulingv1.AddToScheme(scheme))
	utilruntime.Must(v1.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func priorityClass(name string, value int32) *schedulingv1.PriorityClass {
	return &schedulingv1.PriorityClass{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Value:      value,
	}
}

// projectWithNodeAffinity builds a Project fixture carrying the node types in BOTH
// locations: on spec (legacy run.ai CRD) and as the JSON-encoded annotation the agent
// writes (kai.resources CRD). Keeping both means these specs hold under either value of
// the kai.resources annotation location that is read.
func projectWithNodeAffinity(interactive bool, affinity []string) unstructured.Unstructured {
	specField, annotationKey := "nodeAffinityTrain", commonconstants.NodeAffinityTrainAnnotation
	if interactive {
		specField, annotationKey = "nodeAffinityInteractive", commonconstants.NodeAffinityInteractiveAnnotation
	}

	encoded, err := json.Marshal(affinity)
	if err != nil {
		panic(err)
	}

	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":        testProjectName,
				"annotations": map[string]interface{}{annotationKey: string(encoded)},
			},
			"spec": map[string]interface{}{
				specField: getNodeAffinitySlice(affinity),
			},
		},
	}
}
