/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gvk

import (
	"context"
	"encoding/json"
	"testing"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAttachedResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GVK Mutation Suite")
}

var _ = Describe("Runner GVK Labels Mutator", func() {
	var (
		ctx     context.Context
		mutator *LabelMutator
	)

	BeforeEach(func() {
		ctx = context.Background()
		mutator = NewLabelMutator()
	})

	Describe("Mutation Name", func() {
		It("returns the correct mutate name", func() {
			Expect(mutator.Name()).To(Equal("LabelMutator"))
		})
	})

	Describe("MutateResource", func() {
		Context("with Pod template", func() {
			It("adds GVK labels to the resource", func() {
				runner := &dromosv1alpha1.WorkloadRunner{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "default",
					},
					Spec: dromosv1alpha1.WorkloadRunnerSpec{
						Template: newPodTemplate("test-pod", "default"),
					},
				}

				warnings, err := mutator.Mutate(ctx, runner)
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(BeEmpty())

				labels := runner.GetLabels()
				Expect(labels).To(HaveKey(LabelKeyGroup))
				Expect(labels[LabelKeyGroup]).To(Equal(""))
				Expect(labels).To(HaveKey(LabelKeyVersion))
				Expect(labels[LabelKeyVersion]).To(Equal("v1"))
				Expect(labels).To(HaveKey(LabelKeyKind))
				Expect(labels[LabelKeyKind]).To(Equal("Pod"))
			})
		})

		Context("with Job template", func() {
			It("adds GVK labels with batch group to the resource", func() {
				runner := &dromosv1alpha1.WorkloadRunner{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "default",
					},
					Spec: dromosv1alpha1.WorkloadRunnerSpec{
						Template: newJobTemplate("test-job", "default"),
					},
				}

				warnings, err := mutator.Mutate(ctx, runner)
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(BeEmpty())

				labels := runner.GetLabels()
				Expect(labels).To(HaveKey(LabelKeyGroup))
				Expect(labels[LabelKeyGroup]).To(Equal("batch"))
				Expect(labels).To(HaveKey(LabelKeyVersion))
				Expect(labels[LabelKeyVersion]).To(Equal("v1"))
				Expect(labels).To(HaveKey(LabelKeyKind))
				Expect(labels[LabelKeyKind]).To(Equal("Job"))
			})
		})

		Context("with existing labels", func() {
			It("preserves existing labels and adds GVK labels", func() {
				runner := &dromosv1alpha1.WorkloadRunner{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "default",
						Labels: map[string]string{
							"existing-label": "existing-value",
							"app":            "my-app",
						},
					},
					Spec: dromosv1alpha1.WorkloadRunnerSpec{
						Template: newPodTemplate("test-pod", "default"),
					},
				}

				warnings, err := mutator.Mutate(ctx, runner)
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(BeEmpty())

				labels := runner.GetLabels()
				Expect(labels).To(HaveLen(5))
				Expect(labels["existing-label"]).To(Equal("existing-value"))
				Expect(labels["app"]).To(Equal("my-app"))
				Expect(labels[LabelKeyGroup]).To(Equal(""))
				Expect(labels[LabelKeyVersion]).To(Equal("v1"))
				Expect(labels[LabelKeyKind]).To(Equal("Pod"))
			})
		})

		Context("with invalid template", func() {
			It("skips mutation for empty template without error", func() {
				runner := &dromosv1alpha1.WorkloadRunner{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "default",
					},
					Spec: dromosv1alpha1.WorkloadRunnerSpec{
						Template: runtime.RawExtension{Raw: []byte("{}")},
					},
				}

				runner.SetLabels(map[string]string{
					"existing-label": "existing-value",
				})

				warnings, err := mutator.Mutate(ctx, runner)
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(HaveLen(1))
				Expect(warnings[0]).To(ContainSubstring("template missing required apiVersion or kind fields"))

				labels := runner.GetLabels()
				Expect(labels).To(HaveLen(1))
				Expect(labels["existing-label"]).To(Equal("existing-value"))
			})
		})
	})
})

func newPodTemplate(name, namespace string) runtime.RawExtension {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: "nginx",
				},
			},
		},
	}

	data, _ := json.Marshal(pod)
	return runtime.RawExtension{Raw: data}
}

func newJobTemplate(name, namespace string) runtime.RawExtension {
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "busybox",
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	data, _ := json.Marshal(job)
	return runtime.RawExtension{Raw: data}
}
