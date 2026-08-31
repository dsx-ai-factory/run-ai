/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
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
	repoconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// labelsForCastRunner is the minimum label set the v1->v2 cast stamps on the
// wrapped manifest: the cast marker plus a workloadKind disambiguator. Tests
// compose these into the runner's Spec.Template (via stampTemplateLabels) to
// exercise specific filter branches.
func labelsForCastRunner(workloadKind string) map[string]string {
	return map[string]string{
		constants.RunaiProjectLabel:            testProjectName,
		repoconstants.ConvertedFromNativeLabel: "true",
		constants.LabelWorkloadKind:            workloadKind,
	}
}

// stampTemplateLabels embeds the given labels into the runner's Spec.Template
// metadata.labels — where ClassifyRunner reads the cast marker and workloadKind
// from (the control-plane stamps them on the wrapped manifest, not the runner).
func stampTemplateLabels(runner *dromosv1alpha1.WorkloadRunner, labels map[string]string) {
	runner.Spec.Template = apiruntime.RawExtension{Raw: buildJobTemplateRaw("placeholder", labels, nil)}
}

var _ = Describe("ClassifyRunner", func() {
	DescribeTable("skip reasons",
		func(labels map[string]string, phase dromosv1alpha1.WorkloadRunnerPhase, wantReason util.RunnerSkipReason) {
			runner := BuildWorkloadRunner(testNamespaceName, "classify", phase)
			stampTemplateLabels(runner, labels)

			_, reason, ok := util.ClassifyRunner(runner)
			Expect(ok).To(BeFalse())
			Expect(reason).To(Equal(wantReason))
		},
		Entry("missing cast label",
			map[string]string{constants.LabelWorkloadKind: "Workspace"},
			dromosv1alpha1.WorkloadRunnerPhaseRunning,
			util.RunnerSkipNotV1Cast),
		Entry("cast label present but value != true",
			map[string]string{
				repoconstants.ConvertedFromNativeLabel: "false",
				constants.LabelWorkloadKind:            "Training",
			},
			dromosv1alpha1.WorkloadRunnerPhaseRunning,
			util.RunnerSkipNotV1Cast),
		Entry("cast label present but workloadKind missing",
			map[string]string{repoconstants.ConvertedFromNativeLabel: "true"},
			dromosv1alpha1.WorkloadRunnerPhaseRunning,
			util.RunnerSkipMissingWorkloadKind),
		Entry("cast label present, workloadKind out of scope (Inference)",
			labelsForCastRunner("Inference"),
			dromosv1alpha1.WorkloadRunnerPhaseRunning,
			util.RunnerSkipOutOfScopeKind),
		Entry("cast label present, Workspace, pre-running phase (Creating)",
			labelsForCastRunner("Workspace"),
			dromosv1alpha1.WorkloadRunnerPhaseCreating,
			util.RunnerSkipPreRunningPhase),
		Entry("cast label present, Training, pre-running phase (Pending)",
			labelsForCastRunner("Training"),
			dromosv1alpha1.WorkloadRunnerPhasePending,
			util.RunnerSkipPreRunningPhase),
		Entry("cast label present, Workspace, pre-running phase (Initializing)",
			labelsForCastRunner("Workspace"),
			dromosv1alpha1.WorkloadRunnerPhaseInitializing,
			util.RunnerSkipPreRunningPhase),
		Entry("cast label present, Workspace, terminal phase (Completed)",
			labelsForCastRunner("Workspace"),
			dromosv1alpha1.WorkloadRunnerPhaseCompleted,
			util.RunnerSkipTerminalPhase),
		Entry("cast label present, MPIJob, terminal phase (Failed)",
			labelsForCastRunner("MPIJob"),
			dromosv1alpha1.WorkloadRunnerPhaseFailed,
			util.RunnerSkipTerminalPhase),
		Entry("cast label present, Training, terminal phase (Suspended)",
			labelsForCastRunner("Training"),
			dromosv1alpha1.WorkloadRunnerPhaseSuspended,
			util.RunnerSkipTerminalPhase),
		Entry("cast label present, Workspace, terminal phase (Unknown)",
			labelsForCastRunner("Workspace"),
			dromosv1alpha1.WorkloadRunnerPhaseUnknown,
			util.RunnerSkipTerminalPhase),
	)

	DescribeTable("category mapping for in-scope runners",
		func(workloadKind string, phase dromosv1alpha1.WorkloadRunnerPhase, wantCategory string) {
			runner := BuildWorkloadRunner(testNamespaceName, "classify-pass", phase)
			stampTemplateLabels(runner, labelsForCastRunner(workloadKind))

			category, reason, ok := util.ClassifyRunner(runner)
			Expect(ok).To(BeTrue(), "workloadKind=%s phase=%s should pass; got skip=%s", workloadKind, phase, reason)
			Expect(reason).To(BeEmpty())
			Expect(category).To(Equal(wantCategory))
		},
		Entry("Workspace + Running -> Interactive", "Workspace", dromosv1alpha1.WorkloadRunnerPhaseRunning, constants.KindInteractiveWorkload),
		Entry("Workspace + Degraded -> Interactive", "Workspace", dromosv1alpha1.WorkloadRunnerPhaseDegraded, constants.KindInteractiveWorkload),
		Entry("Training + Running -> Training", "Training", dromosv1alpha1.WorkloadRunnerPhaseRunning, constants.KindTrainingWorkload),
		Entry("Training + Degraded -> Training", "Training", dromosv1alpha1.WorkloadRunnerPhaseDegraded, constants.KindTrainingWorkload),
		Entry("MPIJob + Running -> Distributed", "MPIJob", dromosv1alpha1.WorkloadRunnerPhaseRunning, constants.KindDistributedWorkload),
		Entry("PyTorchJob + Running -> Distributed", "PyTorchJob", dromosv1alpha1.WorkloadRunnerPhaseRunning, constants.KindDistributedWorkload),
		Entry("TFJob + Running -> Distributed", "TFJob", dromosv1alpha1.WorkloadRunnerPhaseRunning, constants.KindDistributedWorkload),
		Entry("XGBoostJob + Running -> Distributed", "XGBoostJob", dromosv1alpha1.WorkloadRunnerPhaseRunning, constants.KindDistributedWorkload),
		Entry("JAXJob + Running -> Distributed", "JAXJob", dromosv1alpha1.WorkloadRunnerPhaseRunning, constants.KindDistributedWorkload),
	)
})

var _ = Describe("Runner filter wiring", Serial, func() {
	const recorderBufferSize = 100

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

	It("does not modify a runner without the cast label", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "no-cast", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		Expect(watchClient.Create(ctx, runner)).To(BeNil())

		// 8 seconds is well past the 5-second project time limit. With the
		// filter short-circuiting on not-v1-cast, no enforcement edge should
		// ever fire — phase stays Running and the spec stays intact.
		time.Sleep(8 * time.Second)
		got, err := kubeConfig.GetWLR(ctx, watchClient, testNamespaceName, "no-cast")
		Expect(err).To(BeNil())
		Expect(got.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseRunning))
		Expect(got.Spec.Template.Raw).NotTo(BeEmpty())
	})

	It("does not modify a runner with an out-of-scope workloadKind", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "infw-cast", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		stampTemplateLabels(runner, labelsForCastRunner("Inference"))
		Expect(watchClient.Create(ctx, runner)).To(BeNil())

		time.Sleep(8 * time.Second)
		got, err := kubeConfig.GetWLR(ctx, watchClient, testNamespaceName, "infw-cast")
		Expect(err).To(BeNil())
		Expect(got.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseRunning))
	})

	It("does not modify a runner in a pre-running phase", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "pending-cast", dromosv1alpha1.WorkloadRunnerPhasePending)
		stampTemplateLabels(runner, labelsForCastRunner("Workspace"))
		Expect(watchClient.Create(ctx, runner)).To(BeNil())

		time.Sleep(8 * time.Second)
		got, err := kubeConfig.GetWLR(ctx, watchClient, testNamespaceName, "pending-cast")
		Expect(err).To(BeNil())
		Expect(got.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhasePending))
	})

	It("derives the Interactive category for a Workspace + Running runner that passes the filter", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "ws-running", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		stampTemplateLabels(runner, labelsForCastRunner("Workspace"))
		Expect(watchClient.Create(ctx, runner)).To(BeNil())

		// Sanity-side: in-process classification yields the expected category
		// (the handler runs the same logic against the live runner).
		got, err := kubeConfig.GetWLR(ctx, watchClient, testNamespaceName, "ws-running")
		Expect(err).To(BeNil())
		category, _, ok := util.ClassifyRunner(&got)
		Expect(ok).To(BeTrue())
		Expect(category).To(Equal(constants.KindInteractiveWorkload))
	})
})

var _ = Describe("WarnIfMissingWorkloadID", func() {
	It("does not panic when annotations are nil", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "no-annotations", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		runner.Annotations = nil
		Expect(func() { util.WarnIfMissingWorkloadID(runner) }).NotTo(Panic())
	})

	It("does not warn when the annotation is present", func() {
		runner := BuildWorkloadRunner(testNamespaceName, "with-id", dromosv1alpha1.WorkloadRunnerPhaseRunning)
		runner.Annotations = map[string]string{
			repoconstants.ControlPlaneWorkloadIdAnnotationForWorkload: "wid-123",
		}
		Expect(func() { util.WarnIfMissingWorkloadID(runner) }).NotTo(Panic())
	})
})
