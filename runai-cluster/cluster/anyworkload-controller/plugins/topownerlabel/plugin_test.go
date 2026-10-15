/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package topownerlabel

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/mnnvlconv"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/plugins/mnnvl"

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

	It("uses the API reader when an existing Job is absent from the cache", func() {
		runner := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{UID: types.UID("runner-uid")}}
		existing := toUnstructured(newJob())
		resource := toUnstructured(newJob())
		originalTemplate, _, err := unstructured.NestedMap(resource.Object, "spec", "template")
		Expect(err).NotTo(HaveOccurred())
		plugin, err := New(ctx, framework.PluginDeps{
			Client:    fake.NewClientBuilder().Build(),
			APIReader: fake.NewClientBuilder().WithObjects(existing).Build(),
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(plugin.(framework.ResourceMutatePlugin).MutateResource(ctx, runner, resource, newJobKarta(), state)).To(Succeed())

		Expect(resource.GetAnnotations()).NotTo(HaveKey(podTemplatesLabeledAnnotation))
		template, _, err := unstructured.NestedMap(resource.Object, "spec", "template")
		Expect(err).NotTo(HaveOccurred())
		Expect(template).To(Equal(originalTemplate))
	})

	It("requires an API reader rather than falling back to the cache", func() {
		_, err := New(ctx, framework.PluginDeps{Client: fake.NewClientBuilder().Build()})
		Expect(err).To(MatchError("top-owner-label plugin requires a non-nil APIReader in PluginDeps"))
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
			"existing-root":        "value",
			legacyTopOwnerUIDLabel: "stale-id",
		}
		jobSet.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels = map[string]string{
			"existing-pod":         "value",
			legacyTopOwnerUIDLabel: "stale-id",
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
		Expect(mutated.Labels).To(HaveKeyWithValue(legacyTopOwnerUIDLabel, "workload-id"))
		Expect(mutated.Labels).To(HaveKeyWithValue(commonconstants.TopOwnerWorkloadIdLabel, "workload-id"))
		Expect(mutated.Labels).To(HaveKeyWithValue("existing-root", "value"))
		Expect(mutated.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels).
			To(HaveKeyWithValue(legacyTopOwnerUIDLabel, "workload-id"))
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
		Expect(mutated.Labels).To(HaveKeyWithValue(legacyTopOwnerUIDLabel, "runner-uid"))
		Expect(mutated.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels).
			To(HaveKeyWithValue(legacyTopOwnerUIDLabel, "runner-uid"))
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
		Expect(mutated.Labels).To(HaveKeyWithValue(legacyTopOwnerUIDLabel, "workload-id"))
		Expect(mutated.Annotations).NotTo(HaveKey(podTemplatesLabeledAnnotation))
		Expect(mutated.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels).
			NotTo(HaveKey(legacyTopOwnerUIDLabel))
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
		Expect(mutated.Labels).To(HaveKeyWithValue(legacyTopOwnerUIDLabel, "workload-id"))
		Expect(mutated.Spec.Template.Labels).NotTo(HaveKey(legacyTopOwnerUIDLabel))
	})

	DescribeTable("preserves the stored identity of an existing unmarked Job",
		func(labelKey string) {
			runner := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
				UID: types.UID("replacement-runner-uid"),
				Annotations: map[string]string{
					commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: "changed-id",
				},
			}}
			existing := toUnstructured(newJob())
			existing.SetLabels(map[string]string{labelKey: "original-id"})
			resource := toUnstructured(newJob())
			originalTemplate, _, err := unstructured.NestedMap(resource.Object, "spec", "template")
			Expect(err).NotTo(HaveOccurred())

			Expect(newPlugin(existing).MutateResource(ctx, runner, resource, newJobKarta(), state)).To(Succeed())

			Expect(resource.GetLabels()).To(HaveKeyWithValue(commonconstants.TopOwnerWorkloadIdLabel, "original-id"))
			Expect(resource.GetLabels()).To(HaveKeyWithValue(legacyTopOwnerUIDLabel, "original-id"))
			Expect(resource.GetAnnotations()).NotTo(HaveKey(podTemplatesLabeledAnnotation))
			template, _, err := unstructured.NestedMap(resource.Object, "spec", "template")
			Expect(err).NotTo(HaveOccurred())
			Expect(template).To(Equal(originalTemplate))
		},
		Entry("with the workload ID label", commonconstants.TopOwnerWorkloadIdLabel),
		Entry("with the legacy UID label", legacyTopOwnerUIDLabel),
	)

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
			To(HaveKeyWithValue(legacyTopOwnerUIDLabel, "workload-id"))
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
		Expect(mutated.Labels).To(HaveKeyWithValue(legacyTopOwnerUIDLabel, "runner-uid"))
		Expect(mutated.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels).
			NotTo(HaveKey(legacyTopOwnerUIDLabel))
	})

	It("preserves the persisted Job labels when the runner workload ID changes", func() {
		runner := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
			UID:         types.UID("runner-uid"),
			Annotations: map[string]string{commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: "original-id"},
		}}
		reader := fake.NewClientBuilder().Build()
		plugin := NewPlugin(reader)
		existing := toUnstructured(newJob())
		Expect(plugin.MutateResource(ctx, runner, existing, newJobKarta(), state)).To(Succeed())
		Expect(reader.Create(ctx, existing)).To(Succeed())
		originalTemplate, _, err := unstructured.NestedMap(existing.Object, "spec", "template")
		Expect(err).NotTo(HaveOccurred())

		runner.Annotations[commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload] = "changed-id"
		resource := toUnstructured(newJob())
		Expect(plugin.MutateResource(ctx, runner, resource, newJobKarta(), state)).To(Succeed())
		template, _, err := unstructured.NestedMap(resource.Object, "spec", "template")
		Expect(err).NotTo(HaveOccurred())
		Expect(template).To(Equal(originalTemplate))
		Expect(resource.GetLabels()).To(HaveKeyWithValue(commonconstants.TopOwnerWorkloadIdLabel, "original-id"))
	})

	It("keeps PyTorch master and worker labels aligned with MNNVL affinity across reconciliations", func() {
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{Name: testWorkloadName, Namespace: testNamespace, UID: types.UID("runner-uid")},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{Instructions: &dromosv1alpha1.Instructions{
				GpuOptimizations: &dromosv1alpha1.GpuOptimizations{Mnnvl: &dromosv1alpha1.MnnvlConfig{
					Mode:                            dromosv1alpha1.MnnvlModeRequired,
					GpuNetworkAccelerationLabelKeys: []string{"nvidia.com/gpu.clique"},
				}},
			}},
		}
		podTemplate := toUnstructured(&dromostest.NewJobSetBuilder(testWorkloadName, testNamespace).
			WithDefault().Build().Spec.ReplicatedJobs[0].Template.Spec.Template)
		original := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "kubeflow.org/v1", "kind": "PyTorchJob",
			"metadata": map[string]any{"name": testWorkloadName, "namespace": testNamespace},
			"spec": map[string]any{"pytorchReplicaSpecs": map[string]any{
				"Master": map[string]any{"replicas": int64(1), "template": podTemplate.DeepCopy().Object},
				"Worker": map[string]any{"replicas": int64(2), "template": podTemplate.DeepCopy().Object},
			}},
		}}
		karta := newJobKarta()
		karta.Spec.StructureDefinition.RootComponent.Kind = &kartav1alpha1.GroupVersionKind{
			Group: "kubeflow.org", Version: "v1", Kind: "PyTorchJob",
		}
		karta.Spec.StructureDefinition.RootComponent.SpecDefinition.PodTemplateSpecPath = ptr.To(".spec.pytorchReplicaSpecs[].template")
		karta.Spec.StructureDefinition.RootComponent.InstanceIdPath = ptr.To(".spec.pytorchReplicaSpecs | keys[]")
		reader := fake.NewClientBuilder().Build()
		fw := framework.NewFramework(framework.NewRegistry(
			NewPlugin(reader), mnnvl.NewPlugin(false,
				mnnvl.WithPodAffinityProvider(mnnvlconv.ProvidePodAffinity), mnnvl.WithWorkloadLabelStampDisabled()),
		), nil)
		resource := original.DeepCopy()
		Expect(fw.RunMutateResource(ctx, runner, resource, karta, state)).To(Succeed())
		Expect(reader.Create(ctx, resource)).To(Succeed())
		// The UID fallback becomes persisted identity even if CP later supplies an ID.
		runner.SetAnnotations(map[string]string{commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: "late-cp-id"})
		for range 2 {
			resource = original.DeepCopy()
			Expect(fw.RunMutateResource(ctx, runner, resource, karta, reconcile.New())).To(Succeed())
			Expect(resource.GetLabels()).To(HaveKeyWithValue(commonconstants.TopOwnerWorkloadIdLabel, "runner-uid"))
			for _, replica := range []string{"Master", "Worker"} {
				templateMap, _, err := unstructured.NestedMap(resource.Object, "spec", "pytorchReplicaSpecs", replica, "template")
				Expect(err).NotTo(HaveOccurred())
				template := &corev1.PodTemplateSpec{}
				Expect(runtime.DefaultUnstructuredConverter.FromUnstructured(templateMap, template)).To(Succeed())
				Expect(template.Labels).To(HaveKeyWithValue(commonconstants.TopOwnerWorkloadIdLabel, "runner-uid"))
				Expect(template.Labels).To(HaveKeyWithValue("run.ai/top-owner-uid", "runner-uid"))
				terms := template.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution
				Expect(terms).To(HaveLen(1))
				Expect(terms[0].PodAffinityTerm.LabelSelector.MatchLabels).
					To(HaveKeyWithValue("run.ai/top-owner-uid", "runner-uid"))
			}
		}
	})

})
