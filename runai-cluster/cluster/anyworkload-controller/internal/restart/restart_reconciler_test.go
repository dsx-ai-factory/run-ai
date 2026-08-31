/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package restart

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

var _ = Describe("unwrapSpecOverrideValues", func() {
	unwrap := func(in string) []map[string]any {
		out, err := unwrapSpecOverrideValues([]byte(in))
		Expect(err).NotTo(HaveOccurred())
		var ops []map[string]any
		Expect(json.Unmarshal(out, &ops)).To(Succeed())
		return ops
	}

	It("unwraps a scalar wrapped in a single-key value object", func() {
		ops := unwrap(`[{"op":"replace","path":"/spec/x","value":{"value":3}}]`)
		Expect(ops[0]["value"]).To(Equal(float64(3)))
	})

	It("unwraps an array wrapped in a single-key value object", func() {
		ops := unwrap(`[{"op":"add","path":"/spec/x","value":{"value":["a","b"]}}]`)
		Expect(ops[0]["value"]).To(Equal([]any{"a", "b"}))
	})

	It("leaves a plain object value with multiple keys untouched", func() {
		ops := unwrap(`[{"op":"replace","path":"/spec/x","value":{"mode":"Required","count":2}}]`)
		Expect(ops[0]["value"]).To(Equal(map[string]any{"mode": "Required", "count": float64(2)}))
	})

	It("leaves a single-key object not named value untouched", func() {
		ops := unwrap(`[{"op":"replace","path":"/spec/x","value":{"mode":"Required"}}]`)
		Expect(ops[0]["value"]).To(Equal(map[string]any{"mode": "Required"}))
	})

	It("leaves operations without a value untouched", func() {
		ops := unwrap(`[{"op":"remove","path":"/spec/x"}]`)
		Expect(ops[0]).NotTo(HaveKey("value"))
	})

	It("unwraps exactly one level of nesting", func() {
		ops := unwrap(`[{"op":"replace","path":"/spec/x","value":{"value":{"value":1}}}]`)
		Expect(ops[0]["value"]).To(Equal(map[string]any{"value": float64(1)}))
	})

	It("fails on non-array patch JSON", func() {
		_, err := unwrapSpecOverrideValues([]byte(`{"op":"remove"}`))
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("applyOverrides", func() {
	op := func(opType runv1alpha1.JSONPatchOperationType, path, rawValue string) runv1alpha1.JSONPatchOperation {
		o := runv1alpha1.JSONPatchOperation{Op: opType, Path: path}
		if rawValue != "" {
			o.Value = &runtime.RawExtension{Raw: []byte(rawValue)}
		}
		return o
	}

	It("applies an override to an existing instructions field", func() {
		workload := &dromosv1alpha1.WorkloadRunner{
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Instructions: &dromosv1alpha1.Instructions{
					GpuOptimizations: &dromosv1alpha1.GpuOptimizations{
						Mnnvl: &dromosv1alpha1.MnnvlConfig{Mode: dromosv1alpha1.MnnvlModeRequired},
					},
				},
			},
		}
		patched, err := applyOverrides(workload, runv1alpha1.SpecOverrides{
			op(runv1alpha1.JSONPatchOperationReplace, "/spec/instructions/gpuOptimizations/mnnvl/mode", `"Preferred"`),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(patched.Spec.Instructions.GpuOptimizations.Mnnvl.Mode).To(Equal(dromosv1alpha1.MnnvlMode("Preferred")))
		// the input workload is not mutated
		Expect(workload.Spec.Instructions.GpuOptimizations.Mnnvl.Mode).To(Equal(dromosv1alpha1.MnnvlModeRequired))
	})

	It("materializes nil instructions so an add under it applies (plain-workload restart)", func() {
		workload := &dromosv1alpha1.WorkloadRunner{}
		patched, err := applyOverrides(workload, runv1alpha1.SpecOverrides{
			op(runv1alpha1.JSONPatchOperationAdd, "/spec/instructions/gpuOptimizations", `{"mnnvl":{"mode":"Required"}}`),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(patched.Spec.Instructions.GpuOptimizations.Mnnvl.Mode).To(Equal(dromosv1alpha1.MnnvlModeRequired))
		// the input workload is not mutated
		Expect(workload.Spec.Instructions).To(BeNil())
	})

	It("leaves nil instructions alone when no override targets it (no spurious spec diff)", func() {
		workload := &dromosv1alpha1.WorkloadRunner{}
		patched, err := applyOverrides(workload, runv1alpha1.SpecOverrides{
			op(runv1alpha1.JSONPatchOperationAdd, "/spec/managedBy", `"someone"`),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(patched.Spec.Instructions).To(BeNil())
	})

	It("fails when adding under a missing parent other than instructions (typos stay loud)", func() {
		workload := &dromosv1alpha1.WorkloadRunner{}
		_, err := applyOverrides(workload, runv1alpha1.SpecOverrides{
			op(runv1alpha1.JSONPatchOperationAdd, "/spec/instrctions/nodeAffinity", `{"required":"gpu"}`),
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("apply patch"))
	})
})

// Reconcile tests that should be guarded even if webhook validation is disabled for some reason
var _ = Describe("RestartReconciler.Reconcile defenses", func() {
	const (
		testNS     = "test-ns"
		runnerName = "runner"
		requestUID = "request-uid-1"
	)

	// newManagedRunner is a workload mid-handoff: locked to the restart controller and
	// pointing at the request.
	newManagedRunner := func(generation int64) *dromosv1alpha1.WorkloadRunner {
		return &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:       runnerName,
				Namespace:  testNS,
				Generation: generation,
				Labels:     map[string]string{RequestRefLabel: requestUID},
			},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				ManagedBy: RestartControllerManagedBy,
			},
		}
	}

	newRequest := func(workloadGeneration int64, overrides runv1alpha1.SpecOverrides) *runv1alpha1.RestartRequest {
		return &runv1alpha1.RestartRequest{
			ObjectMeta: metav1.ObjectMeta{Name: "request", Namespace: testNS, UID: requestUID},
			Spec: runv1alpha1.RestartRequestSpec{
				WorkloadRef: runv1alpha1.WorkloadReference{
					Name:       runnerName,
					Generation: workloadGeneration,
				},
				SpecOverrides: overrides,
			},
		}
	}
	newFakeClient := func(initObjs ...client.Object) client.Client {
		s := runtime.NewScheme()
		Expect(dromosv1alpha1.AddToScheme(s)).To(Succeed())
		Expect(runv1alpha1.AddToScheme(s)).To(Succeed())
		return fake.NewClientBuilder().
			WithScheme(s).
			WithObjects(initObjs...).
			WithStatusSubresource(&dromosv1alpha1.WorkloadRunner{}, &runv1alpha1.RestartRequest{}).
			WithIndex(&runv1alpha1.RestartRequest{}, MetadataUIDField, func(o client.Object) []string {
				return []string{string(o.GetUID())}
			}).
			Build()
	}

	It("fails a stale-generation request before touching the workload", func() {
		validOverride := runv1alpha1.SpecOverrides{{
			Op:    runv1alpha1.JSONPatchOperationAdd,
			Path:  "/spec/instructions/gpuOptimizations",
			Value: &runtime.RawExtension{Raw: []byte(`{"mnnvl":{"mode":"Required"}}`)},
		}}
		restartRequest := newRequest(1, validOverride)
		runner := newManagedRunner(5)
		c := newFakeClient(runner, restartRequest)
		restartReconciler := NewRestartReconciler(c, 0)
		reconcileRequest := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(runner)}
		_, err := restartReconciler.Reconcile(context.Background(), reconcileRequest)
		Expect(err).NotTo(HaveOccurred())

		currentWorkload := &dromosv1alpha1.WorkloadRunner{}
		Expect(c.Get(context.Background(), client.ObjectKeyFromObject(runner), currentWorkload)).To(Succeed())
		currentRequest := &runv1alpha1.RestartRequest{}
		Expect(c.Get(context.Background(), client.ObjectKeyFromObject(restartRequest), currentRequest)).To(Succeed())

		Expect(currentRequest.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseFailed))
		failed := apimeta.FindStatusCondition(currentRequest.Status.Conditions, restartFailedConditionType)
		Expect(failed).NotTo(BeNil())
		Expect(failed.Message).To(ContainSubstring("generation mismatch"))

		// no override applied
		Expect(currentWorkload.Spec.Instructions).To(BeNil())
		Expect(currentWorkload.Status.Phase).NotTo(Equal(dromosv1alpha1.WorkloadRunnerPhaseRestarting))
		Expect(currentWorkload.Spec.ManagedBy).To(Equal(RestartControllerManagedBy))
	})

	It("fails a request whose overrides cannot apply, leaving the workload spec untouched", func() {
		badOverride := runv1alpha1.SpecOverrides{{
			Op:    runv1alpha1.JSONPatchOperationAdd,
			Path:  "/spec/no-such-parent/child",
			Value: &runtime.RawExtension{Raw: []byte(`{}`)},
		}}

		// the request pins the pre-lock generation; the handoff lock bump puts the workload one ahead
		restartRequest := newRequest(1, badOverride)
		runner := newManagedRunner(2)
		c := newFakeClient(runner, restartRequest)
		restartReconciler := NewRestartReconciler(c, 0)
		reconcileRequest := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(runner)}
		_, err := restartReconciler.Reconcile(context.Background(), reconcileRequest)
		Expect(err).NotTo(HaveOccurred())

		// get fresh objects
		currentWorkload := &dromosv1alpha1.WorkloadRunner{}
		Expect(c.Get(context.Background(), client.ObjectKeyFromObject(runner), currentWorkload)).To(Succeed())
		currentRequest := &runv1alpha1.RestartRequest{}
		Expect(c.Get(context.Background(), client.ObjectKeyFromObject(restartRequest), currentRequest)).To(Succeed())

		// restart failed
		Expect(currentRequest.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseFailed))
		failed := apimeta.FindStatusCondition(currentRequest.Status.Conditions, restartFailedConditionType)
		Expect(failed).NotTo(BeNil())
		Expect(failed.Message).To(ContainSubstring("apply spec overrides"))

		// handback on terminal requests stays the HandoffReconciler's job
		Expect(currentWorkload.Spec.ManagedBy).To(Equal(RestartControllerManagedBy))
	})
})
