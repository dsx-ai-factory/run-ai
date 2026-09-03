/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/config"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPodMutator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pod Mutator (default node pools) Unit Tests")
}

const (
	namespaceProjectLabelKey = "test.example.com/project-on-namespace"
	nodePoolLabelKey         = "test.example.com/node-pool"
	defaultNodepoolName      = "default"

	projectNamespace = "team-a-ns"
	projectName      = "team-a-project"

	customNodePool      = "np-x"
	customNodePoolKey   = "node.example.com/np-x"
	customNodePoolValue = "true"
)

var _ = Describe("PodMutator default node pools", func() {
	BeforeEach(func() {
		DeferCleanup(config.SetForTest(config.PodGroupAssignerConfig{
			NodePoolLabelKey:         nodePoolLabelKey,
			NamespaceProjectLabelKey: namespaceProjectLabelKey,
			DefaultNodepoolName:      defaultNodepoolName,
		}))
	})

	It("adds a DoesNotExist expression for the implicit default node pool", func() {
		mutator := newPodMutatorForTest(
			namespaceWithProject(projectNamespace, projectName),
			projectWithDefaultNodePools(projectName, defaultNodepoolName),
		)
		pod := newPod(nil, nil)

		mutator.mutateNodeAffinityForDefaultNodePools(context.Background(), pod, projectNamespace)

		expr := singleMatchExpression(pod)
		Expect(expr.Key).To(Equal(nodePoolLabelKey))
		Expect(expr.Operator).To(Equal(corev1.NodeSelectorOpDoesNotExist))
	})

	It("adds an In expression resolved from a custom node pool's labelKey/labelValue", func() {
		mutator := newPodMutatorForTest(
			namespaceWithProject(projectNamespace, projectName),
			projectWithDefaultNodePools(projectName, customNodePool),
			kaiNodePool(customNodePool, customNodePoolKey, customNodePoolValue, v1alpha1.NodePoolReady),
		)
		pod := newPod(nil, nil)

		mutator.mutateNodeAffinityForDefaultNodePools(context.Background(), pod, projectNamespace)

		expr := singleMatchExpression(pod)
		Expect(expr.Key).To(Equal(customNodePoolKey))
		Expect(expr.Operator).To(Equal(corev1.NodeSelectorOpIn))
		Expect(expr.Values).To(Equal([]string{customNodePoolValue}))
	})

	It("skips a node pool that is being deleted, leaving the pod untouched", func() {
		mutator := newPodMutatorForTest(
			namespaceWithProject(projectNamespace, projectName),
			projectWithDefaultNodePools(projectName, customNodePool),
			kaiNodePool(customNodePool, customNodePoolKey, customNodePoolValue, v1alpha1.NodePoolDeleting),
		)
		pod := newPod(nil, nil)

		mutator.mutateNodeAffinityForDefaultNodePools(context.Background(), pod, projectNamespace)

		Expect(pod.Spec.Affinity).To(BeNil())
	})

	It("honors an explicit node-pool label on the pod over the project defaults", func() {
		mutator := newPodMutatorForTest(
			namespaceWithProject(projectNamespace, projectName),
			projectWithDefaultNodePools(projectName, defaultNodepoolName),
			kaiNodePool(customNodePool, customNodePoolKey, customNodePoolValue, v1alpha1.NodePoolReady),
		)
		pod := newPod(map[string]string{nodePoolLabelKey: customNodePool}, nil)

		mutator.mutateNodeAffinityForDefaultNodePools(context.Background(), pod, projectNamespace)

		expr := singleMatchExpression(pod)
		Expect(expr.Key).To(Equal(customNodePoolKey))
		Expect(expr.Operator).To(Equal(corev1.NodeSelectorOpIn))
		Expect(expr.Values).To(Equal([]string{customNodePoolValue}))
	})

	It("does not mutate when the pod already references a node pool in its node affinity", func() {
		existing := nodeAffinity(corev1.NodeSelectorRequirement{
			Key:      customNodePoolKey,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{customNodePoolValue},
		})
		mutator := newPodMutatorForTest(
			namespaceWithProject(projectNamespace, projectName),
			projectWithDefaultNodePools(projectName, defaultNodepoolName),
			kaiNodePool(customNodePool, customNodePoolKey, customNodePoolValue, v1alpha1.NodePoolReady),
		)
		pod := newPod(nil, existing)

		mutator.mutateNodeAffinityForDefaultNodePools(context.Background(), pod, projectNamespace)

		// Unchanged: still exactly the single pre-existing expression, no default appended.
		expr := singleMatchExpression(pod)
		Expect(expr.Key).To(Equal(customNodePoolKey))
		Expect(expr.Operator).To(Equal(corev1.NodeSelectorOpIn))
	})

	It("does nothing when the namespace has no project label", func() {
		mutator := newPodMutatorForTest(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: projectNamespace}},
		)
		pod := newPod(nil, nil)

		mutator.mutateNodeAffinityForDefaultNodePools(context.Background(), pod, projectNamespace)

		Expect(pod.Spec.Affinity).To(BeNil())
	})

	It("does nothing when the project has no default node pools", func() {
		mutator := newPodMutatorForTest(
			namespaceWithProject(projectNamespace, projectName),
			projectWithDefaultNodePools(projectName),
		)
		pod := newPod(nil, nil)

		mutator.mutateNodeAffinityForDefaultNodePools(context.Background(), pod, projectNamespace)

		Expect(pod.Spec.Affinity).To(BeNil())
	})
})

func newPodMutatorForTest(initObjs ...client.Object) *PodMutator {
	scheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).
		Build()

	return NewPodMutator(fakeClient)
}

func namespaceWithProject(name, project string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{namespaceProjectLabelKey: project},
		},
	}
}

func projectWithDefaultNodePools(name string, defaultNodePools ...string) *v1alpha1.Project {
	return &v1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       v1alpha1.ProjectSpec{DefaultNodePools: defaultNodePools},
	}
}

func kaiNodePool(name, labelKey, labelValue string, phase v1alpha1.NodePoolPhase) *v1alpha1.NodePool {
	return &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       v1alpha1.NodePoolSpec{LabelKey: labelKey, LabelValue: labelValue},
		Status:     v1alpha1.NodePoolStatus{Phase: phase},
	}
}

func newPod(labels map[string]string, affinity *corev1.Affinity) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-1", Namespace: projectNamespace, Labels: labels},
		Spec:       corev1.PodSpec{Affinity: affinity},
	}
}

func nodeAffinity(expr corev1.NodeSelectorRequirement) *corev1.Affinity {
	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{MatchExpressions: []corev1.NodeSelectorRequirement{expr}},
				},
			},
		},
	}
}

func singleMatchExpression(pod *corev1.Pod) corev1.NodeSelectorRequirement {
	Expect(pod.Spec.Affinity).NotTo(BeNil())
	Expect(pod.Spec.Affinity.NodeAffinity).NotTo(BeNil())
	required := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	Expect(required).NotTo(BeNil())
	Expect(required.NodeSelectorTerms).To(HaveLen(1))
	Expect(required.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))
	return required.NodeSelectorTerms[0].MatchExpressions[0]
}
