/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	pkg "github.com/run-ai/runai/runai-cluster/cluster/runtime-rules"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/controllers"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/kubeConfig"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/prometheus"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/queuesManager"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/validation"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// BuildWorkloadRunner constructs a minimal WorkloadRunner suitable for the
// envtest fake client. Spec.Template wraps a placeholder Job manifest; the
// content is irrelevant because the runtime-rules does not yet
// inspect the wrapped Kind.
//
// UID is stamped explicitly: the fake client does not auto-assign UIDs and
// GetWorkloadValues requires both jobUuid and jobName on the synthetic
// OwnedResource the conversion stamps.
func BuildWorkloadRunner(namespace, name string, phase dromosv1alpha1.WorkloadRunnerPhase) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		TypeMeta: metav1.TypeMeta{
			Kind:       dromosv1alpha1.WorkloadRunnerKind,
			APIVersion: "run.ai/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			UID:               uuid.NewUUID(),
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels: map[string]string{
				constants.RunaiProjectLabel: testProjectName,
			},
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template: apiruntime.RawExtension{
				Raw: []byte(`{"apiVersion":"batch/v1","kind":"Job","metadata":{"name":"placeholder"}}`),
			},
		},
		Status: dromosv1alpha1.WorkloadRunnerStatus{
			Phase: phase,
		},
	}
}

// buildJobTemplateRaw builds the raw embedded Job manifest for a runner's
// Spec.Template, stamping the given metadata.labels and metadata.annotations
// (either may be nil). Mirrors where the control-plane puts the cast marker,
// workloadKind, GPU and auto-delete metadata — all on the wrapped manifest,
// which the runtime-rules reads off Spec.Template.
func buildJobTemplateRaw(name string, labels, annotations map[string]string) []byte {
	metadata := map[string]any{"name": name}
	if labels != nil {
		metadata["labels"] = labels
	}
	if annotations != nil {
		metadata["annotations"] = annotations
	}
	raw, err := json.Marshal(map[string]any{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata":   metadata,
	})
	Expect(err).To(BeNil())
	return raw
}

func initSchemaWithRunner() *apiruntime.Scheme {
	return initSchema()
}

var _ = Describe("WorkloadRunner observation wiring", func() {
	const recorderBufferSize = 100

	// convertRunner is a small helper that runs the new util.RunnerConverter
	// against a fake client (built from any additional objects the test
	// needs). Tests that don't need a wrapped child can pass no extras.
	convertRunner := func(runner *dromosv1alpha1.WorkloadRunner, extras ...client.Object) *jobs.JobWorkload {
		scheme := initSchemaWithRunner()
		objs := append([]client.Object{}, extras...)
		c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
		var jw jobs.JobWorkload
		err := util.NewRunnerConverter(c).Convert(runner, &jw)
		Expect(err).To(BeNil())
		return &jw
	}

	Describe("RunnerConverter — base projection", func() {
		It("stamps Kind, copies metadata, and synthesises an OwnedResource entry", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "convert-active", dromosv1alpha1.WorkloadRunnerPhaseRunning)

			jw := convertRunner(runner)
			Expect(jw.Kind).To(Equal(constants.KindWorkloadRunner))
			Expect(jw.Name).To(Equal("convert-active"))
			Expect(jw.Namespace).To(Equal(testNamespaceName))

			Expect(jw.Status.OwnedResources).To(HaveLen(1))
			Expect(jw.Status.OwnedResources[0].Kind.Kind).To(Equal(dromosv1alpha1.WorkloadRunnerKind))
			Expect(jw.Status.OwnedResources[0].Uid).To(Equal(runner.UID))

			Expect(jw.Status.StartTime).NotTo(BeNil())
			Expect(jw.Spec.Active).NotTo(BeNil())
			Expect(jw.Spec.Active.Value).NotTo(BeNil())
			Expect(*jw.Spec.Active.Value).To(BeTrue())

			// Phase passthrough: the auto-delete handler reads this string.
			Expect(jw.Status.WorkloadCommonStatus.Phase).To(Equal(string(dromosv1alpha1.WorkloadRunnerPhaseRunning)))
		})

		It("stamps a valid APIVersion even when the source runner's TypeMeta is empty", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "no-typemeta", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			runner.TypeMeta = metav1.TypeMeta{}

			jw := convertRunner(runner)
			Expect(jw.Kind).To(Equal(constants.KindWorkloadRunner))
			Expect(jw.APIVersion).To(Equal(dromosv1alpha1.GroupVersion.String()))
		})

		It("marks terminal phases inactive", func() {
			terminal := []dromosv1alpha1.WorkloadRunnerPhase{
				dromosv1alpha1.WorkloadRunnerPhaseCompleted,
				dromosv1alpha1.WorkloadRunnerPhaseFailed,
				dromosv1alpha1.WorkloadRunnerPhaseStopped,
				dromosv1alpha1.WorkloadRunnerPhaseSuspended,
				dromosv1alpha1.WorkloadRunnerPhaseUnknown,
			}
			for _, phase := range terminal {
				runner := BuildWorkloadRunner(testNamespaceName, "term-"+string(phase), phase)
				jw := convertRunner(runner)
				Expect(jw.Spec.Active.Value).NotTo(BeNil())
				Expect(*jw.Spec.Active.Value).To(BeFalse(), "phase %s should be inactive", phase)
				Expect(jw.Status.WorkloadCommonStatus.Phase).To(Equal(string(phase)), "phase string should passthrough")
			}
		})
	})

	Describe("RunnerConverter — timestamps", func() {
		It("prefers the run.ai/start-time annotation over CreationTimestamp", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "start-anno", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			fixed := time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC)
			runner.Annotations = map[string]string{
				constants.StartTimeAnnotation: fixed.Format(time.RFC3339),
			}
			jw := convertRunner(runner)
			Expect(jw.Status.StartTime).NotTo(BeNil())
			Expect(jw.Status.StartTime.Time.Equal(fixed)).To(BeTrue())
		})

		It("falls back to CreationTimestamp when the annotation is absent", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "start-fallback", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			created := runner.CreationTimestamp.Time
			jw := convertRunner(runner)
			Expect(jw.Status.StartTime).NotTo(BeNil())
			Expect(jw.Status.StartTime.Time.Equal(created)).To(BeTrue())
		})

		It("populates CompletionTime from run.ai/completion-time when present", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "completion-anno", dromosv1alpha1.WorkloadRunnerPhaseCompleted)
			fixed := time.Date(2024, 7, 1, 13, 0, 0, 0, time.UTC)
			runner.Annotations = map[string]string{
				constants.CompletionTimeAnnotation: fixed.Format(time.RFC3339),
			}
			jw := convertRunner(runner)
			Expect(jw.Status.CompletionTime).NotTo(BeNil())
			Expect(jw.Status.CompletionTime.Time.Equal(fixed)).To(BeTrue())
		})

		It("leaves CompletionTime nil when annotation is absent (legacy unset parity)", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "no-completion", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			jw := convertRunner(runner)
			Expect(jw.Status.CompletionTime).To(BeNil())
		})

		It("ignores invalid RFC3339 annotation values without erroring", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "bad-anno", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			runner.Annotations = map[string]string{
				constants.StartTimeAnnotation:      "not-a-timestamp",
				constants.CompletionTimeAnnotation: "also-not",
			}
			jw := convertRunner(runner)
			// invalid start-time → falls back to CreationTimestamp.
			Expect(jw.Status.StartTime).NotTo(BeNil())
			// invalid completion-time → nil.
			Expect(jw.Status.CompletionTime).To(BeNil())
		})
	})

	Describe("RunnerConverter — template annotation reads", func() {
		// withTemplateAnnotations rewrites the runner's Spec.Template to an
		// embedded Job manifest carrying the given metadata.annotations — where
		// the control-plane stamps GPU and auto-delete metadata at submission,
		// and where the converter now reads them from (no separate child fetch).
		withTemplateAnnotations := func(runner *dromosv1alpha1.WorkloadRunner, annotations map[string]string) {
			runner.Spec.Template = apiruntime.RawExtension{Raw: buildJobTemplateRaw("placeholder", nil, annotations)}
		}

		It("populates AutoDeletionTimeAfterCompletionSeconds from the template annotation", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "with-autodelete", dromosv1alpha1.WorkloadRunnerPhaseCompleted)
			withTemplateAnnotations(runner, map[string]string{
				constants.AutoDeletionTimeAfterCompletionAnnotation: "42",
			})
			jw := convertRunner(runner)
			Expect(jw.Spec.AutoDeletionTimeAfterCompletionSeconds).NotTo(BeNil())
			Expect(jw.Spec.AutoDeletionTimeAfterCompletionSeconds.Deref32()).To(Equal(int32(42)))
		})

		It("treats a zero auto-delete value as 'delete immediately' (parity with legacy)", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "autodelete-zero", dromosv1alpha1.WorkloadRunnerPhaseCompleted)
			withTemplateAnnotations(runner, map[string]string{
				constants.AutoDeletionTimeAfterCompletionAnnotation: "0",
			})
			jw := convertRunner(runner)
			Expect(jw.Spec.AutoDeletionTimeAfterCompletionSeconds).NotTo(BeNil())
			Expect(jw.Spec.AutoDeletionTimeAfterCompletionSeconds.Deref32()).To(Equal(int32(0)))
		})

		It("leaves AutoDeletion nil when the annotation is missing — not an error", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "no-autodelete", dromosv1alpha1.WorkloadRunnerPhaseCompleted)
			withTemplateAnnotations(runner, nil)
			jw := convertRunner(runner)
			Expect(jw.Spec.AutoDeletionTimeAfterCompletionSeconds).To(BeNil())
		})

		It("ignores a non-numeric auto-delete annotation without erroring", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "bad-autodelete", dromosv1alpha1.WorkloadRunnerPhaseCompleted)
			withTemplateAnnotations(runner, map[string]string{
				constants.AutoDeletionTimeAfterCompletionAnnotation: "not-an-int",
			})
			jw := convertRunner(runner)
			Expect(jw.Spec.AutoDeletionTimeAfterCompletionSeconds).To(BeNil())
		})

		It("detects GPU via gpu-fraction annotation on the template", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "gpu-fraction", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			withTemplateAnnotations(runner, map[string]string{
				constants.AnnotationGpuFraction: "0.5",
			})
			jw := convertRunner(runner)
			Expect(util.HasRequestedGpu(jw)).To(BeTrue())
		})

		It("detects GPU via gpu-fraction-num-devices annotation on the template", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "gpu-devices", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			withTemplateAnnotations(runner, map[string]string{
				constants.AnnotationGpuFractionDevices: "2",
			})
			jw := convertRunner(runner)
			Expect(util.HasRequestedGpu(jw)).To(BeTrue())
		})

		It("detects GPU via gpu-memory annotation on the template", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "gpu-memory", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			withTemplateAnnotations(runner, map[string]string{
				constants.AnnotationGpuMemory: "1024",
			})
			jw := convertRunner(runner)
			Expect(util.HasRequestedGpu(jw)).To(BeTrue())
		})

		It("reports HasRequestedGpu=false when the template carries no GPU annotation", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "no-gpu", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			withTemplateAnnotations(runner, nil)
			jw := convertRunner(runner)
			Expect(util.HasRequestedGpu(jw)).To(BeFalse())
		})

		It("detects a whole-GPU request via the resource limit (no partial annotation)", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "whole-gpu", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			runner.Spec.Template = apiruntime.RawExtension{Raw: []byte(`{"apiVersion":"batch/v1","kind":"Job","metadata":{"name":"whole-gpu"},` +
				`"spec":{"template":{"spec":{"containers":[{"name":"main","resources":{"limits":{"nvidia.com/gpu":"1"}}}]}}}}`)}
			jw := convertRunner(runner)
			Expect(util.HasRequestedGpu(jw)).To(BeTrue())
		})

		It("degrades gracefully when the template carries no annotations", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "orphan", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			// Default template (placeholder Job, no annotations).
			jw := convertRunner(runner)
			Expect(jw.Spec.AutoDeletionTimeAfterCompletionSeconds).To(BeNil())
			Expect(util.HasRequestedGpu(jw)).To(BeFalse())
			// Conversion still succeeded — base fields populated.
			Expect(jw.Name).To(Equal("orphan"))
		})
	})

	Describe("Lifecycle annotation helper", func() {
		It("stamps start-time on Running and clears completion-time if present", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "running-r", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			runner.Annotations = map[string]string{
				constants.CompletionTimeAnnotation: "2024-01-01T00:00:00Z",
			}
			scheme := initSchemaWithRunner()
			c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()

			h := handlers.NewWorkloadHandler(c, nil, nil, nil, nil)
			Expect(h.ReconcileTimestampAnnotations(context.Background(), runner)).To(Succeed())

			got, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "running-r")
			Expect(err).To(BeNil())
			Expect(got.Annotations).To(HaveKey(constants.StartTimeAnnotation))
			Expect(got.Annotations).NotTo(HaveKey(constants.CompletionTimeAnnotation))
		})

		It("stamps completion-time on Completed and clears start-time if present", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "completed-r", dromosv1alpha1.WorkloadRunnerPhaseCompleted)
			runner.Annotations = map[string]string{
				constants.StartTimeAnnotation: "2024-01-01T00:00:00Z",
			}
			scheme := initSchemaWithRunner()
			c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()

			h := handlers.NewWorkloadHandler(c, nil, nil, nil, nil)
			Expect(h.ReconcileTimestampAnnotations(context.Background(), runner)).To(Succeed())

			got, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "completed-r")
			Expect(err).To(BeNil())
			Expect(got.Annotations).To(HaveKey(constants.CompletionTimeAnnotation))
			Expect(got.Annotations).NotTo(HaveKey(constants.StartTimeAnnotation))
		})

		It("is a no-op for transient phases (Creating/Pending/Initializing/Unknown)", func() {
			for _, phase := range []dromosv1alpha1.WorkloadRunnerPhase{
				dromosv1alpha1.WorkloadRunnerPhaseCreating,
				dromosv1alpha1.WorkloadRunnerPhasePending,
				dromosv1alpha1.WorkloadRunnerPhaseInitializing,
				dromosv1alpha1.WorkloadRunnerPhaseUnknown,
			} {
				runner := BuildWorkloadRunner(testNamespaceName, "noop-"+string(phase), phase)
				scheme := initSchemaWithRunner()
				c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()
				h := handlers.NewWorkloadHandler(c, nil, nil, nil, nil)
				Expect(h.ReconcileTimestampAnnotations(context.Background(), runner)).To(Succeed())

				got, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "noop-"+string(phase))
				Expect(err).To(BeNil())
				Expect(got.Annotations).NotTo(HaveKey(constants.StartTimeAnnotation),
					"phase %s should not stamp start-time", phase)
				Expect(got.Annotations).NotTo(HaveKey(constants.CompletionTimeAnnotation),
					"phase %s should not stamp completion-time", phase)
			}
		})

		It("clears both timestamps on Suspended/Stopped", func() {
			for _, phase := range []dromosv1alpha1.WorkloadRunnerPhase{
				dromosv1alpha1.WorkloadRunnerPhaseSuspended,
				dromosv1alpha1.WorkloadRunnerPhaseStopped,
			} {
				runner := BuildWorkloadRunner(testNamespaceName, "reset-"+string(phase), phase)
				runner.Annotations = map[string]string{
					constants.StartTimeAnnotation:      "2024-01-01T00:00:00Z",
					constants.CompletionTimeAnnotation: "2024-01-02T00:00:00Z",
				}
				scheme := initSchemaWithRunner()
				c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()
				h := handlers.NewWorkloadHandler(c, nil, nil, nil, nil)
				Expect(h.ReconcileTimestampAnnotations(context.Background(), runner)).To(Succeed())

				got, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "reset-"+string(phase))
				Expect(err).To(BeNil())
				Expect(got.Annotations).NotTo(HaveKey(constants.StartTimeAnnotation),
					"phase %s should clear start-time", phase)
				Expect(got.Annotations).NotTo(HaveKey(constants.CompletionTimeAnnotation),
					"phase %s should clear completion-time", phase)
			}
		})

		It("resume after suspend yields a fresh start-time (not the pre-suspend value)", func() {
			const staleStart = "2024-01-01T00:00:00Z"
			runner := BuildWorkloadRunner(testNamespaceName, "resume-r", dromosv1alpha1.WorkloadRunnerPhaseSuspended)
			runner.Annotations = map[string]string{constants.StartTimeAnnotation: staleStart}
			scheme := initSchemaWithRunner()
			c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()
			h := handlers.NewWorkloadHandler(c, nil, nil, nil, nil)

			// Suspend observed: the stale start-time is cleared.
			Expect(h.ReconcileTimestampAnnotations(context.Background(), runner)).To(Succeed())
			suspended, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "resume-r")
			Expect(err).To(BeNil())
			Expect(suspended.Annotations).NotTo(HaveKey(constants.StartTimeAnnotation))

			// Resume observed: a fresh start-time is stamped, distinct from the
			// pre-suspend value, so the enforcement allowance resets.
			suspended.Status.Phase = dromosv1alpha1.WorkloadRunnerPhaseRunning
			Expect(h.ReconcileTimestampAnnotations(context.Background(), &suspended)).To(Succeed())
			resumed, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "resume-r")
			Expect(err).To(BeNil())
			Expect(resumed.Annotations).To(HaveKey(constants.StartTimeAnnotation))
			Expect(resumed.Annotations[constants.StartTimeAnnotation]).NotTo(Equal(staleStart))
		})

		It("is idempotent — a second Running observation does not rewrite start-time", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "idempotent", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			scheme := initSchemaWithRunner()
			c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()
			h := handlers.NewWorkloadHandler(c, nil, nil, nil, nil)

			Expect(h.ReconcileTimestampAnnotations(context.Background(), runner)).To(Succeed())
			gotFirst, _ := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "idempotent")
			firstStamp := gotFirst.Annotations[constants.StartTimeAnnotation]
			Expect(firstStamp).NotTo(BeEmpty())

			// Re-observe with the now-annotated runner; helper should no-op.
			Expect(h.ReconcileTimestampAnnotations(context.Background(), &gotFirst)).To(Succeed())
			gotSecond, _ := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "idempotent")
			Expect(gotSecond.Annotations[constants.StartTimeAnnotation]).To(Equal(firstStamp))
		})
	})

	Describe("HandleRunnerEvent annotation wiring", func() {
		It("stamps completion-time for an in-scope terminal runner (reaches the helper despite the phase gate)", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "wired-completed", dromosv1alpha1.WorkloadRunnerPhaseCompleted)
			stampTemplateLabels(runner, labelsForCastRunner("Training"))
			scheme := initSchemaWithRunner()
			c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()
			h := handlers.NewWorkloadHandler(c, nil, nil, nil, nil)

			Expect(h.HandleRunnerEvent(context.Background(), runner)).To(Succeed())

			got, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "wired-completed")
			Expect(err).To(BeNil())
			Expect(got.Annotations).To(HaveKey(constants.CompletionTimeAnnotation))
			Expect(got.Annotations).NotTo(HaveKey(constants.StartTimeAnnotation))
		})

		It("clears the stale start-time for an in-scope suspended runner", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "wired-suspended", dromosv1alpha1.WorkloadRunnerPhaseSuspended)
			stampTemplateLabels(runner, labelsForCastRunner("Workspace"))
			runner.Annotations = map[string]string{constants.StartTimeAnnotation: "2024-01-01T00:00:00Z"}
			scheme := initSchemaWithRunner()
			c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()
			h := handlers.NewWorkloadHandler(c, nil, nil, nil, nil)

			Expect(h.HandleRunnerEvent(context.Background(), runner)).To(Succeed())

			got, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "wired-suspended")
			Expect(err).To(BeNil())
			Expect(got.Annotations).NotTo(HaveKey(constants.StartTimeAnnotation))
		})

		It("leaves an out-of-scope runner's annotations untouched", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "wired-oos", dromosv1alpha1.WorkloadRunnerPhaseCompleted)
			// No cast label — not ours; the helper must not run.
			runner.Annotations = map[string]string{
				constants.StartTimeAnnotation:      "2024-01-01T00:00:00Z",
				constants.CompletionTimeAnnotation: "2024-01-02T00:00:00Z",
			}
			scheme := initSchemaWithRunner()
			c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()
			h := handlers.NewWorkloadHandler(c, nil, nil, nil, nil)

			Expect(h.HandleRunnerEvent(context.Background(), runner)).To(Succeed())

			got, err := kubeConfig.GetWLR(context.Background(), c, testNamespaceName, "wired-oos")
			Expect(err).To(BeNil())
			Expect(got.Annotations[constants.StartTimeAnnotation]).To(Equal("2024-01-01T00:00:00Z"))
			Expect(got.Annotations[constants.CompletionTimeAnnotation]).To(Equal("2024-01-02T00:00:00Z"))
		})
	})

	Describe("Read helpers", func() {
		It("GetWLR and ListWLR round-trip via the fake client", func() {
			ns := BuildNamespace(testNamespaceName, map[string]string{RunaiQueueLabel: testProjectName})
			runner := BuildWorkloadRunner(testNamespaceName, "lookup-test", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			scheme := initSchemaWithRunner()
			c := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(ns, runner).Build()
			ctx := context.Background()

			got, err := kubeConfig.GetWLR(ctx, c, testNamespaceName, "lookup-test")
			Expect(err).To(BeNil())
			Expect(got.Name).To(Equal("lookup-test"))

			list, err := kubeConfig.ListWLR(ctx, c, testNamespaceName)
			Expect(err).To(BeNil())
			Expect(list.Items).To(HaveLen(1))
			Expect(list.Items[0].Name).To(Equal("lookup-test"))
		})
	})

	Describe("WorkloadRunner enablement", func() {
		It("includes WorkloadRunner in SupportedJobKinds when the pre-3 disable flag is set", func() {
			rr := pkg.NewRuntimeRules(nil, nil, nil, nil, nil, validation.NewWorkloadValidator(), true)
			_, ok := rr.SupportedJobKinds[constants.KindWorkloadRunner]
			Expect(ok).To(BeTrue())
		})

		It("registers WorkloadRunner in SupportedJobKinds by default (rule enforcement enabled)", func() {
			rr := pkg.NewRuntimeRules(nil, nil, nil, nil, nil, validation.NewWorkloadValidator(), false)
			obj, ok := rr.SupportedJobKinds[constants.KindWorkloadRunner]
			Expect(ok).To(BeTrue())
			Expect(obj).NotTo(BeNil())
		})
	})

	Describe("Watch + periodic enqueue with the flag on", Serial, func() {
		var (
			watchClient client.WithWatch
			stopper     chan struct{}
			ctx         context.Context
			cancel      context.CancelFunc
			promClient  = prometheus.GetQuerierClient(false, "")
		)

		BeforeEach(func() {
			queuesManager.SetKillWorker(false, 0)
			ctx, cancel = context.WithCancel(context.TODO())

			ns := BuildNamespace(testNamespaceName, map[string]string{RunaiQueueLabel: testProjectName})
			projects := BuildProjectFixtures(testProjectName, testNamespaceName, 5, 5)

			scheme := initSchemaWithRunner()
			watchClient = fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(append(projects, ns)...).Build()

			qm := &queuesManager.QueuesManager{}
			rr := pkg.NewRuntimeRules(watchClient, qm, promClient, record.NewFakeRecorder(recorderBufferSize),
				[]int{queuesManager.UnknownEvent, queuesManager.DeactivateEvent}, validation.NewWorkloadValidator(), false)

			workloadCache := util.WorkloadCache{}
			rr.InitQueueManager(ctx, &workloadCache)

			// Manually start only the WorkloadRunner controller; the goal of these
			// integration tests is to verify the runner watch wiring, not the
			// legacy paths (those are exercised by add_workload_test.go).
			workloadHandler := handlers.NewWorkloadHandler(rr.K8sClient, rr.QueuesManager, rr.PrometheusQuerier, &workloadCache, rr.SupportedJobKinds)
			runnerController := controllers.NewWorkloadController(rr.K8sClient, workloadHandler, constants.KindWorkloadRunner)
			stopper = make(chan struct{})
			go runnerController.Start(ctx, stopper)

			time.Sleep(2 * time.Second)
		})

		AfterEach(func() {
			cancel()
			close(stopper)
			queuesManager.SetKillWorker(true, time.Duration(workerSleepSeconds+1))
		})

		It("observes runner events without panicking and without modifying the runner", func() {
			runner := BuildWorkloadRunner(testNamespaceName, "obs-runner", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			Expect(watchClient.Create(ctx, runner)).To(BeNil())

			// Wait past the project's 5-second TrainingJobTimeLimit. With the
			// flag on but no enforcement edge implemented, the runner must not
			// be modified, suspended, or deleted.
			time.Sleep(8 * time.Second)

			got, err := kubeConfig.GetWLR(ctx, watchClient, testNamespaceName, "obs-runner")
			Expect(err).To(BeNil())
			Expect(got.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseRunning))
			// Spec.Template should be unchanged — the runtime-rules never touched it.
			Expect(got.Spec.Template.Raw).NotTo(BeEmpty())
		})

		It("PushProjectWorkloadsToProjectQueue picks up runners pre-existing in the namespace", func() {
			// Pre-create a runner and re-run InitQueueManager → PushProjectWorkloadsToProjectQueue
			// against a fresh QueuesManager to assert no error path is hit.
			runner := BuildWorkloadRunner(testNamespaceName, "preexist-runner", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			Expect(watchClient.Create(ctx, runner)).To(BeNil())

			freshCache := util.WorkloadCache{}
			freshQM := &queuesManager.QueuesManager{}
			freshRR := pkg.NewRuntimeRules(watchClient, freshQM, promClient, record.NewFakeRecorder(recorderBufferSize),
				[]int{queuesManager.UnknownEvent, queuesManager.DeactivateEvent}, validation.NewWorkloadValidator(), false)

			freshRR.InitQueueManager(ctx, &freshCache)

			// The project queue should now exist and have processed the runner
			// through PushProjectWorkloadsToProjectQueue without erroring.
			projectQueue := freshQM.Get(testProjectName)
			Expect(projectQueue).NotTo(BeNil())
			// Runner is still in the cluster and untouched.
			got, err := kubeConfig.GetWLR(ctx, watchClient, testNamespaceName, "preexist-runner")
			Expect(err).To(BeNil())
			Expect(got.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseRunning))
		})
	})

	Describe("project enumeration with legacy workloads disabled", func() {
		It("invokes ListWLR when the pre-3 disable flag is set", func() {
			// Wrap the fake client with an interceptor that counts List calls
			// targeting WorkloadRunnerList. Runtime-rules must enumerate runners
			// even when the pre-3 disable flag is set.
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()

			ns := BuildNamespace(testNamespaceName, map[string]string{RunaiQueueLabel: testProjectName})
			projects := BuildProjectFixtures(testProjectName, testNamespaceName, 5, 5)
			runner := BuildWorkloadRunner(testNamespaceName, "ignored-runner", dromosv1alpha1.WorkloadRunnerPhaseRunning)
			scheme := initSchemaWithRunner()

			runnerListCalls := 0
			c := fakerunaiclient.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(append(projects, ns, runner)...).
				WithInterceptorFuncs(interceptor.Funcs{
					List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
						if _, ok := list.(*dromosv1alpha1.WorkloadRunnerList); ok {
							runnerListCalls++
						}
						return c.List(ctx, list, opts...)
					},
				}).
				Build()

			qm := &queuesManager.QueuesManager{}
			promClient := prometheus.GetQuerierClient(false, "")
			rr := pkg.NewRuntimeRules(c, qm, promClient, record.NewFakeRecorder(recorderBufferSize),
				[]int{queuesManager.UnknownEvent, queuesManager.DeactivateEvent}, validation.NewWorkloadValidator(), true)

			workloadCache := util.WorkloadCache{}
			rr.InitQueueManager(ctx, &workloadCache)

			Expect(runnerListCalls).To(Equal(1),
				"runtime-rules must enumerate WorkloadRunners")

			// Sanity-side: the runner remains untouched.
			got, err := kubeConfig.GetWLR(ctx, c, testNamespaceName, "ignored-runner")
			Expect(err).To(BeNil())
			Expect(got.Name).To(Equal("ignored-runner"))

			_, present := rr.SupportedJobKinds[constants.KindWorkloadRunner]
			Expect(present).To(BeTrue())
		})
	})

})
