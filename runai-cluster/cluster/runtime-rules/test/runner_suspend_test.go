/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/kubeConfig"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// buildTemplateRaw marshals a minimal embedded manifest of the given
// apiVersion/kind for a runner's Spec.Template, optionally seeding nested
// fields (e.g. an already-true suspend) via extraSpec.
func buildTemplateRaw(apiVersion, kind, name string, extraSpec map[string]any) []byte {
	manifest := map[string]any{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata":   map[string]any{"name": name},
	}
	if extraSpec != nil {
		manifest["spec"] = extraSpec
	}
	raw, err := json.Marshal(manifest)
	Expect(err).To(BeNil())
	return raw
}

// templateSuspend parses the runner's Spec.Template and reads the bool at path.
func templateSuspend(runner dromosv1alpha1.WorkloadRunner, path ...string) (bool, bool) {
	obj := &unstructured.Unstructured{}
	Expect(json.Unmarshal(runner.Spec.Template.Raw, &obj.Object)).To(Succeed())
	val, found, err := unstructured.NestedBool(obj.Object, path...)
	Expect(err).To(BeNil())
	return val, found
}

var _ = Describe("WorkloadRunner native suspend", func() {
	const startTime = "2024-07-01T12:00:00Z"

	suspendWith := func(runner *dromosv1alpha1.WorkloadRunner, extras ...client.Object) (client.Client, error) {
		scheme := initSchemaWithRunner()
		objs := append([]client.Object{runner}, extras...)
		c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
		return c, util.SuspendRunner(context.Background(), c, runner.Namespace, runner.Name)
	}

	It("flips spec.suspend on a batch/v1 Job-wrapped runner and clears start-time", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "job-suspend", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		runner.Spec.Template = apiruntime.RawExtension{Raw: buildTemplateRaw("batch/v1", "Job", "job-suspend", nil)}
		runner.Annotations = map[string]string{constants.StartTimeAnnotation: startTime}

		c, err := suspendWith(runner)
		Expect(err).To(BeNil())

		got, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "job-suspend")
		Expect(err).To(BeNil())
		val, found := templateSuspend(got, "spec", "suspend")
		Expect(found).To(BeTrue())
		Expect(val).To(BeTrue())
		Expect(got.Annotations).NotTo(HaveKey(constants.StartTimeAnnotation))
	})

	It("preserves other template fields (minimal patch, no read-modify-write clobber)", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "preserve-fields", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		runner.Spec.Template = apiruntime.RawExtension{
			Raw: buildTemplateRaw("batch/v1", "Job", "preserve-fields", map[string]any{"parallelism": int64(3)}),
		}

		c, err := suspendWith(runner)
		Expect(err).To(BeNil())

		got, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "preserve-fields")
		Expect(err).To(BeNil())
		val, found := templateSuspend(got, "spec", "suspend")
		Expect(found).To(BeTrue())
		Expect(val).To(BeTrue())

		obj := &unstructured.Unstructured{}
		Expect(json.Unmarshal(got.Spec.Template.Raw, &obj.Object)).To(Succeed())
		parallelism, found, err := unstructured.NestedFloat64(obj.Object, "spec", "parallelism")
		Expect(err).To(BeNil())
		Expect(found).To(BeTrue())
		Expect(parallelism).To(Equal(float64(3)))
	})

	It("flips spec.runPolicy.suspend on a kubeflow MPIJob-wrapped runner", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "mpi-suspend", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		runner.Spec.Template = apiruntime.RawExtension{Raw: buildTemplateRaw("kubeflow.org/v1", constants.KindMpiJob, "mpi-suspend", nil)}
		runner.Annotations = map[string]string{constants.StartTimeAnnotation: startTime}

		c, err := suspendWith(runner)
		Expect(err).To(BeNil())

		got, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "mpi-suspend")
		Expect(err).To(BeNil())
		val, found := templateSuspend(got, "spec", "runPolicy", "suspend")
		Expect(found).To(BeTrue())
		Expect(val).To(BeTrue())
		Expect(got.Annotations).NotTo(HaveKey(constants.StartTimeAnnotation))
	})

	It("returns a clear error for an unsupported wrapped kind", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "bad-kind", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		runner.Spec.Template = apiruntime.RawExtension{Raw: buildTemplateRaw("apps/v1", "Deployment", "bad-kind", nil)}

		_, err := suspendWith(runner)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported wrapped kind"))
	})

	It("treats a missing runner as already-suspended (success)", func() {
		scheme := initSchemaWithRunner()
		c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).Build()
		Expect(util.SuspendRunner(context.Background(), c, testNamespaceName, "ghost")).To(Succeed())
	})

	It("is idempotent: no patch when already suspended and start-time absent", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "already-suspended", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		runner.Spec.Template = apiruntime.RawExtension{
			Raw: buildTemplateRaw("batch/v1", "Job", "already-suspended", map[string]any{"suspend": true}),
		}
		// no start-time annotation present

		scheme := initSchemaWithRunner()
		patchCount := 0
		c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).
			WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(ctx context.Context, cl client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					patchCount++
					return cl.Patch(ctx, obj, patch, opts...)
				},
			}).Build()

		Expect(util.SuspendRunner(context.Background(), c, testNamespaceName, "already-suspended")).To(Succeed())
		Expect(patchCount).To(Equal(0), "no patch should be issued when already in the desired state")
	})

	It("still clears start-time when already suspended but the annotation lingers", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "suspended-stale-anno", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		runner.Spec.Template = apiruntime.RawExtension{
			Raw: buildTemplateRaw("batch/v1", "Job", "suspended-stale-anno", map[string]any{"suspend": true}),
		}
		runner.Annotations = map[string]string{constants.StartTimeAnnotation: startTime}

		c, err := suspendWith(runner)
		Expect(err).To(BeNil())

		got, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "suspended-stale-anno")
		Expect(err).To(BeNil())
		Expect(got.Annotations).NotTo(HaveKey(constants.StartTimeAnnotation))
	})

	It("surfaces the patch error when the write fails", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "patch-fails", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		runner.Spec.Template = apiruntime.RawExtension{Raw: buildTemplateRaw("batch/v1", "Job", "patch-fails", nil)}
		runner.Annotations = map[string]string{constants.StartTimeAnnotation: startTime}

		scheme := initSchemaWithRunner()
		c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).
			WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(ctx context.Context, cl client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					return apierrors.NewForbidden(schema.GroupResource{Group: "run.ai", Resource: "workloadrunners"}, obj.GetName(), nil)
				},
			}).Build()

		err := util.SuspendRunner(context.Background(), c, testNamespaceName, "patch-fails")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("patch suspend"))
	})
})
