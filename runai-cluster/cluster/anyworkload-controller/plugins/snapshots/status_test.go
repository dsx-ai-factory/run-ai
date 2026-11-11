/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
)

var optIn = map[string]string{optInAnnotation: "true"}

var _ = Describe("Status", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	gpu := gpuContainer("main")
	ownProfile := corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault}

	// wantIssue is part of the warning; empty means no warning at all.
	DescribeTable("reports the phase, with a warning when an opt-in can't be honoured",
		func(karta *kartav1alpha1.Karta, workload func() *unstructured.Unstructured,
			wantPhase dromosv1alpha1.PluginPhase, wantIssue string) {
			plugin := newPlugin(kartaStub{karta: karta}, true)
			status := &dromosv1alpha1.PluginStatus{Phase: dromosv1alpha1.PluginPhaseUnknown}
			state := reconcile.New()

			Expect(plugin.Status(ctx, runnerFor(workload()), nil, status, state)).To(Succeed())

			Expect(status.Phase).To(Equal(wantPhase))
			issues := state.PluginIssues(plugin)
			if wantIssue == "" {
				Expect(issues).To(BeEmpty())
				return
			}
			Expect(issues).To(HaveLen(1))
			Expect(issues[0].IsCritical).To(BeFalse(), "an ineligible workload's phase must not change")
			Expect(issues[0].Error()).To(ContainSubstring(wantIssue))
		},
		Entry("not opted in", deploymentKarta(),
			func() *unstructured.Unstructured { return deployment(nil, gpu) },
			dromosv1alpha1.PluginPhaseDisabled, ""),
		// The signal is on the workload (REQ-02-1); a pod label alone is not it.
		Entry("only a pod label", deploymentKarta(),
			func() *unstructured.Unstructured {
				return withPodLabel(deployment(nil, gpu), constants.SnapshotOptInLabel, "true")
			},
			dromosv1alpha1.PluginPhaseDisabled, ""),
		Entry("eligible", deploymentKarta(),
			func() *unstructured.Unstructured { return deployment(optIn, gpu) },
			dromosv1alpha1.PluginPhaseReady, ""),
		Entry("two pod templates", pairKarta(),
			func() *unstructured.Unstructured { return pair(optIn, gpu) },
			dromosv1alpha1.PluginPhaseDisabled, "2 pod templates"),
		Entry("pods cannot carry labels", bareKarta(),
			func() *unstructured.Unstructured { return bare(optIn, gpu) },
			dromosv1alpha1.PluginPhaseDisabled, "cannot carry labels"),
		Entry("no GPU container", deploymentKarta(),
			func() *unstructured.Unstructured {
				return deployment(optIn, corev1.Container{Name: "main", Image: "busybox"})
			},
			dromosv1alpha1.PluginPhaseDisabled, "no container requests"),
		Entry("two GPU containers", deploymentKarta(),
			func() *unstructured.Unstructured { return deployment(optIn, gpu, gpuContainer("sidecar")) },
			dromosv1alpha1.PluginPhaseDisabled, "both request"),
		Entry("own seccomp profile", deploymentKarta(),
			func() *unstructured.Unstructured { return deployment(optIn, withSeccomp(gpu, ownProfile)) },
			dromosv1alpha1.PluginPhaseDisabled, "seccomp"),
	)

	// The framework starts a plugin at Unknown, which holds the workload in
	// Initializing; a failed lookup must not leave it there.
	It("is Disabled when the lookup fails", func() {
		plugin := newPlugin(kartaStub{err: errors.New("cache not synced")}, true)
		status := &dromosv1alpha1.PluginStatus{Phase: dromosv1alpha1.PluginPhaseUnknown}

		err := plugin.Status(ctx, runnerFor(deployment(optIn, gpu)), nil, status, reconcile.New())

		Expect(err).To(HaveOccurred())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
	})

	// The plugin runs for every workload, so one that never opted in must not get
	// an issue from a lookup it doesn't need.
	It("skips the lookup when not opted in", func() {
		plugin := newPlugin(kartaStub{err: errors.New("cache not synced")}, true)
		status := &dromosv1alpha1.PluginStatus{Phase: dromosv1alpha1.PluginPhaseUnknown}

		Expect(plugin.Status(ctx, runnerFor(deployment(nil, gpu)), nil, status, reconcile.New())).To(Succeed())

		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
	})

	It("is Disabled without Snapshot, with no warning", func() {
		plugin := newPlugin(kartaStub{karta: deploymentKarta()}, false)
		status := &dromosv1alpha1.PluginStatus{Phase: dromosv1alpha1.PluginPhaseUnknown}
		state := reconcile.New()

		Expect(plugin.Status(ctx, runnerFor(deployment(optIn, gpu)), nil, status, state)).To(Succeed())

		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
		Expect(state.PluginIssues(plugin)).To(BeEmpty())
	})
})

type kartaStub struct {
	karta *kartav1alpha1.Karta
	err   error
}

func (s kartaStub) GetKartaForGVK(context.Context, schema.GroupVersionKind) (*kartav1alpha1.Karta, error) {
	return s.karta, s.err
}

func runnerFor(workload *unstructured.Unstructured) *dromosv1alpha1.WorkloadRunner {
	raw, err := json.Marshal(workload.Object)
	Expect(err).NotTo(HaveOccurred())
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{Name: "llama", Namespace: "runai-proj"},
		Spec:       dromosv1alpha1.WorkloadRunnerSpec{Template: runtime.RawExtension{Raw: raw}},
	}
}

func deploymentKarta() *kartav1alpha1.Karta {
	return kartaWith(kartav1alpha1.ComponentDefinition{
		Name:           "deployment",
		Kind:           &kartav1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: ptr.To(".spec.template")},
	})
}

// pairKarta is a kind with a leader and a worker template, like LeaderWorkerSet.
func pairKarta() *kartav1alpha1.Karta {
	return kartaWith(
		kartav1alpha1.ComponentDefinition{
			Name: "pair",
			Kind: &kartav1alpha1.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Pair"},
		},
		kartav1alpha1.ComponentDefinition{
			Name:           "leader",
			OwnerRef:       ptr.To("pair"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: ptr.To(".spec.leader")},
		},
		kartav1alpha1.ComponentDefinition{
			Name:           "worker",
			OwnerRef:       ptr.To("pair"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: ptr.To(".spec.worker")},
		},
	)
}

// bareKarta is a kind that embeds a pod spec with no metadata.
func bareKarta() *kartav1alpha1.Karta {
	return kartaWith(kartav1alpha1.ComponentDefinition{
		Name:           "bare",
		Kind:           &kartav1alpha1.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Bare"},
		SpecDefinition: &kartav1alpha1.SpecDefinition{PodSpecPath: ptr.To(".spec.podSpec")},
	})
}

func kartaWith(root kartav1alpha1.ComponentDefinition, children ...kartav1alpha1.ComponentDefinition) *kartav1alpha1.Karta {
	return &kartav1alpha1.Karta{Spec: kartav1alpha1.KartaSpec{
		StructureDefinition: kartav1alpha1.StructureDefinition{RootComponent: root, ChildComponents: children},
	}}
}

// deployment is a Deployment whose own annotations carry the opt-in, and whose
// pod template has a label of its own, so stamping can be seen to keep it.
func deployment(annotations map[string]string, containers ...corev1.Container) *unstructured.Unstructured {
	return toUnstructured(&appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{Name: "llama", Namespace: "runai-proj", Annotations: annotations},
		Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "llama"}},
			Spec:       corev1.PodSpec{Containers: containers},
		}},
	})
}

func withPodLabel(workload *unstructured.Unstructured, key, value string) *unstructured.Unstructured {
	Expect(unstructured.SetNestedField(workload.Object, value, "spec", "template", "metadata", "labels", key)).To(Succeed())
	return workload
}

func pair(annotations map[string]string, containers ...corev1.Container) *unstructured.Unstructured {
	template := toMap(&corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: containers}})
	workload := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "example.com/v1",
		"kind":       "Pair",
		"metadata":   map[string]any{"name": "llama", "namespace": "runai-proj"},
		"spec":       map[string]any{"leader": template, "worker": runtime.DeepCopyJSON(template)},
	}}
	workload.SetAnnotations(annotations)
	return workload
}

func bare(annotations map[string]string, containers ...corev1.Container) *unstructured.Unstructured {
	workload := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "example.com/v1",
		"kind":       "Bare",
		"metadata":   map[string]any{"name": "llama", "namespace": "runai-proj"},
		"spec":       map[string]any{"podSpec": toMap(&corev1.PodSpec{Containers: containers})},
	}}
	workload.SetAnnotations(annotations)
	return workload
}

func toUnstructured(object runtime.Object) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: toMap(object)}
}

func toMap(object any) map[string]any {
	converted, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	Expect(err).NotTo(HaveOccurred())
	return converted
}

func gpuContainer(name string) corev1.Container {
	return corev1.Container{
		Name:  name,
		Image: "vllm/vllm-openai:v0.31.0",
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")},
		},
	}
}

func withSeccomp(container corev1.Container, profile corev1.SeccompProfile) corev1.Container {
	container.SecurityContext = &corev1.SecurityContext{SeccompProfile: &profile}
	return container
}
