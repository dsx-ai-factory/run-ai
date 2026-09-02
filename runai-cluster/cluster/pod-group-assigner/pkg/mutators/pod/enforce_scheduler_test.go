/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"context"
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaipgconstants "github.com/kai-scheduler/KAI-scheduler/pkg/podgrouper/podgrouper/plugins/constants"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/config"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	enforceAnnotationKey = "test.example.com/enforce-scheduler-name"

	schedulerName      = "test-scheduler"
	otherSchedulerName = "default-scheduler"
)

// enforcementConfig is the config the enforcement-related specs run under. It carries the
// node-pool vocabulary too, since Handle always runs the default-node-pools mutation.
func enforcementConfig() config.PodGroupAssignerConfig {
	return config.PodGroupAssignerConfig{
		NodePoolLabelKey:              nodePoolLabelKey,
		NamespaceProjectLabelKey:      namespaceProjectLabelKey,
		DefaultNodepoolName:           defaultNodepoolName,
		EnforceSchedulerAnnotationKey: enforceAnnotationKey,
		SchedulerName:                 schedulerName,
	}
}

var _ = Describe("PodMutator scheduler enforcement", func() {
	BeforeEach(func() {
		DeferCleanup(config.SetForTest(enforcementConfig()))
	})

	Describe("isSchedulerEnforcedForNamespace", func() {
		DescribeTable("resolves the enforcement annotation",
			func(annotations map[string]string, expected bool) {
				mutator := newPodMutatorForTest(&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: projectNamespace, Annotations: annotations},
				})

				enforced, err := mutator.isSchedulerEnforcedForNamespace(context.Background(), projectNamespace)

				Expect(err).NotTo(HaveOccurred())
				Expect(enforced).To(Equal(expected))
			},
			Entry("true", map[string]string{enforceAnnotationKey: "true"}, true),
			Entry("false", map[string]string{enforceAnnotationKey: "false"}, false),
			Entry("absent", map[string]string{}, false),
			Entry("no annotations at all", nil, false),
			// A malformed value must never force a scheduler onto a pod by accident.
			Entry("unparseable", map[string]string{enforceAnnotationKey: "yes-please"}, false),
			Entry("a different annotation", map[string]string{"other/annotation": "true"}, false),
		)

		It("returns an error when the namespace cannot be read", func() {
			mutator := newPodMutatorForTest()

			_, err := mutator.isSchedulerEnforcedForNamespace(context.Background(), projectNamespace)

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("mutateSchedulerName", func() {
		It("sets the configured scheduler on a pod that has none", func() {
			mutator := newPodMutatorForTest()
			pod := newPod(nil, nil)

			mutator.mutateSchedulerName(context.Background(), pod, projectNamespace)

			Expect(pod.Spec.SchedulerName).To(Equal(schedulerName))
		})

		It("overrides a scheduler the submitter asked for", func() {
			mutator := newPodMutatorForTest()
			pod := newPod(nil, nil)
			pod.Spec.SchedulerName = otherSchedulerName

			mutator.mutateSchedulerName(context.Background(), pod, projectNamespace)

			Expect(pod.Spec.SchedulerName).To(Equal(schedulerName))
		})

		It("leaves a pod that already runs on the configured scheduler untouched", func() {
			mutator := newPodMutatorForTest()
			pod := newPod(nil, nil)
			pod.Spec.SchedulerName = schedulerName

			mutator.mutateSchedulerName(context.Background(), pod, projectNamespace)

			Expect(pod.Spec.SchedulerName).To(Equal(schedulerName))
		})
	})

	Describe("Handle", func() {
		It("patches the scheduler name when the namespace enforces it", func() {
			mutator := newPodMutatorForTest(
				enforcing(namespaceWithProject(projectNamespace, projectName), "true"),
				projectWithDefaultNodePools(projectName),
			)
			pod := newPod(nil, nil)
			pod.Spec.SchedulerName = otherSchedulerName

			response := mutator.Handle(context.Background(), admissionRequestForPod(pod))

			Expect(response.Allowed).To(BeTrue())
			Expect(patchedPod(pod, response).Spec.SchedulerName).To(Equal(schedulerName))
		})

		DescribeTable("leaves the pod alone when the namespace does not enforce",
			func(annotationValue string) {
				mutator := newPodMutatorForTest(
					enforcing(namespaceWithProject(projectNamespace, projectName), annotationValue),
					projectWithDefaultNodePools(projectName, defaultNodepoolName),
				)
				pod := newPod(nil, nil)
				pod.Spec.SchedulerName = otherSchedulerName

				response := mutator.Handle(context.Background(), admissionRequestForPod(pod))

				Expect(response.Allowed).To(BeTrue())
				// No patch at all - not the scheduler, and not the project's default node pools.
				Expect(response.Patches).To(BeEmpty())
			},
			Entry("explicitly false", "false"),
			Entry("annotation missing", ""),
		)

		It("rejects the request when the namespace cannot be read", func() {
			// Fail closed: SchedulerName is immutable after creation, so a pod admitted
			// without enforcement could never be corrected.
			mutator := newPodMutatorForTest()
			pod := newPod(nil, nil)

			response := mutator.Handle(context.Background(), admissionRequestForPod(pod))

			Expect(response.Allowed).To(BeFalse())
		})

		It("applies the project's default node pools alongside the scheduler name", func() {
			mutator := newPodMutatorForTest(
				enforcing(namespaceWithProject(projectNamespace, projectName), "true"),
				projectWithDefaultNodePools(projectName, customNodePool),
				kaiNodePool(customNodePool, customNodePoolKey, customNodePoolValue, v1alpha1.NodePoolReady),
			)
			pod := newPod(nil, nil)

			response := mutator.Handle(context.Background(), admissionRequestForPod(pod))

			Expect(response.Allowed).To(BeTrue())
			mutated := patchedPod(pod, response)
			Expect(mutated.Spec.SchedulerName).To(Equal(schedulerName))
			Expect(mutated.Labels).To(HaveKeyWithValue(kaipgconstants.ProjectLabelKey, projectName))
			Expect(singleMatchExpression(mutated).Key).To(Equal(customNodePoolKey))
		})

		It("rejects a request whose object is not a pod", func() {
			mutator := newPodMutatorForTest(
				enforcing(namespaceWithProject(projectNamespace, projectName), "true"),
			)

			response := mutator.Handle(context.Background(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Namespace: projectNamespace,
					Object:    runtime.RawExtension{Raw: []byte("not json")},
				},
			})

			Expect(response.Allowed).To(BeFalse())
		})
	})
})

// enforcing stamps the enforcement annotation onto a namespace fixture. An empty value
// leaves the annotation off entirely, modelling a namespace project-controller never touched.
func enforcing(namespace *corev1.Namespace, value string) *corev1.Namespace {
	if value == "" {
		return namespace
	}

	if namespace.Annotations == nil {
		namespace.Annotations = map[string]string{}
	}
	namespace.Annotations[enforceAnnotationKey] = value

	return namespace
}

func admissionRequestForPod(pod *corev1.Pod) admission.Request {
	raw, err := json.Marshal(pod)
	Expect(err).NotTo(HaveOccurred())

	return admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Namespace: pod.Namespace,
			Object:    runtime.RawExtension{Raw: raw},
		},
	}
}

// patchedPod applies the response's JSON patch to the original pod, so specs assert on the
// resulting pod rather than on patch internals.
func patchedPod(original *corev1.Pod, response admission.Response) *corev1.Pod {
	originalRaw, err := json.Marshal(original)
	Expect(err).NotTo(HaveOccurred())

	patchRaw, err := json.Marshal(response.Patches)
	Expect(err).NotTo(HaveOccurred())

	patch, err := jsonpatch.DecodePatch(patchRaw)
	Expect(err).NotTo(HaveOccurred())

	patchedRaw, err := patch.Apply(originalRaw)
	Expect(err).NotTo(HaveOccurred())

	mutated := &corev1.Pod{}
	Expect(json.Unmarshal(patchedRaw, mutated)).To(Succeed())

	return mutated
}
