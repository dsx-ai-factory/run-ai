/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package topownerlabel

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	dromostest "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/test"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

const (
	testWorkloadName = "workload"
	testNamespace    = "project"
)

func toUnstructured(object any) *unstructured.Unstructured {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	Expect(err).NotTo(HaveOccurred())
	return &unstructured.Unstructured{Object: content}
}

func fromUnstructured(resource *unstructured.Unstructured) *jobsetv1alpha2.JobSet {
	jobSet := &jobsetv1alpha2.JobSet{}
	Expect(runtime.DefaultUnstructuredConverter.FromUnstructured(resource.Object, jobSet)).To(Succeed())
	return jobSet
}

func newPlugin(objects ...client.Object) *Plugin {
	return NewPlugin(fake.NewClientBuilder().WithObjects(objects...).Build())
}

func newJob() *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: "batch/v1", Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testWorkloadName,
			Namespace: testNamespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{{Name: "main", Image: "test"}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
}

func newJobKarta() *kartav1alpha1.Karta {
	return &kartav1alpha1.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "job",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group: "batch", Version: "v1", Kind: "Job",
					},
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.template"),
					},
				},
			},
		},
	}
}

var _ = Describe("Plugin", func() {
	var (
		ctx   context.Context
		state *reconcile.State
	)

	BeforeEach(func() {
		ctx = context.Background()
		state = reconcile.New()
	})

	It("stamps the top-owner label when creating a workload", func() {
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				UID: types.UID("runner-uid"),
				Annotations: map[string]string{
					commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: "workload-id",
				},
			},
		}
		jobSet := dromostest.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()
		jobSet.Labels = map[string]string{
			"existing-root":                  "value",
			commonconstants.TopOwnerUidLabel: "stale-id",
		}
		jobSet.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels = map[string]string{
			"existing-pod":                   "value",
			commonconstants.TopOwnerUidLabel: "stale-id",
		}
		resource := toUnstructured(jobSet)
		plugin := newPlugin()

		Expect(plugin.MutateResource(
			ctx,
			runner,
			resource,
			dromostest.NewValidJobSetKartaWithPodTemplateSpec(),
			state,
		)).To(Succeed())
		Expect(state.Issues()).To(BeEmpty())

		mutated := fromUnstructured(resource)
		Expect(mutated.Annotations).To(HaveKeyWithValue(podTemplatesLabeledAnnotation, "true"))
		Expect(mutated.Labels).To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, "workload-id"))
		Expect(mutated.Labels).To(HaveKeyWithValue("existing-root", "value"))
		Expect(mutated.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels).
			To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, "workload-id"))
		Expect(mutated.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels).
			To(HaveKeyWithValue("existing-pod", "value"))
	})

	It("falls back to the WorkloadRunner UID", func() {
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{UID: types.UID("runner-uid")},
		}
		resource := toUnstructured(
			dromostest.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build(),
		)
		plugin := newPlugin()

		Expect(plugin.MutateResource(
			ctx,
			runner,
			resource,
			dromostest.NewValidJobSetKartaWithFragmentedPodSpec(),
			state,
		)).To(Succeed())

		mutated := fromUnstructured(resource)
		Expect(mutated.Labels).To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, "runner-uid"))
		Expect(mutated.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels).
			To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, "runner-uid"))
	})

	It("does not change pod templates on an existing unmarked workload", func() {
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				UID: types.UID("runner-uid"),
				Annotations: map[string]string{
					commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: "workload-id",
				},
			},
		}
		jobSet := dromostest.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()
		existing := toUnstructured(jobSet.DeepCopy())
		resource := toUnstructured(jobSet.DeepCopy())
		plugin := newPlugin(existing)

		Expect(plugin.MutateResource(
			ctx,
			runner,
			resource,
			dromostest.NewValidJobSetKartaWithPodTemplateSpec(),
			state,
		)).To(Succeed())

		mutated := fromUnstructured(resource)
		Expect(mutated.Labels).To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, "workload-id"))
		Expect(mutated.Annotations).NotTo(HaveKey(podTemplatesLabeledAnnotation))
		Expect(mutated.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels).
			NotTo(HaveKey(commonconstants.TopOwnerUidLabel))
	})

	It("does not change the immutable template of an existing unmarked Job", func() {
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				UID: types.UID("runner-uid"),
				Annotations: map[string]string{
					commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: "workload-id",
				},
			},
		}
		existing := toUnstructured(newJob())
		resource := toUnstructured(newJob())
		plugin := newPlugin(existing)

		Expect(plugin.MutateResource(ctx, runner, resource, newJobKarta(), state)).To(Succeed())

		mutated := &batchv1.Job{}
		Expect(runtime.DefaultUnstructuredConverter.FromUnstructured(resource.Object, mutated)).To(Succeed())
		Expect(mutated.Labels).To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, "workload-id"))
		Expect(mutated.Spec.Template.Labels).NotTo(HaveKey(commonconstants.TopOwnerUidLabel))
	})

	It("preserves pod-template labels on workloads created by the plugin", func() {
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				UID: types.UID("runner-uid"),
				Annotations: map[string]string{
					commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: "workload-id",
				},
			},
		}
		jobSet := dromostest.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()
		existing := toUnstructured(jobSet.DeepCopy())
		stampPodTemplateMarker(existing)
		resource := toUnstructured(jobSet.DeepCopy())
		plugin := newPlugin(existing)

		Expect(plugin.MutateResource(
			ctx,
			runner,
			resource,
			dromostest.NewValidJobSetKartaWithPodTemplateSpec(),
			state,
		)).To(Succeed())

		mutated := fromUnstructured(resource)
		Expect(mutated.Annotations).To(HaveKeyWithValue(podTemplatesLabeledAnnotation, "true"))
		Expect(mutated.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels).
			To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, "workload-id"))
	})

	It("still stamps the workload when Karta exposes only a bare PodSpec", func() {
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{UID: types.UID("runner-uid")},
		}
		resource := toUnstructured(
			dromostest.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build(),
		)
		plugin := newPlugin()

		Expect(plugin.MutateResource(
			ctx,
			runner,
			resource,
			dromostest.NewValidJobSetKartaWithPodSpec(),
			state,
		)).To(Succeed())

		mutated := fromUnstructured(resource)
		Expect(mutated.Labels).To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, "runner-uid"))
		Expect(mutated.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels).
			NotTo(HaveKey(commonconstants.TopOwnerUidLabel))
	})
})
