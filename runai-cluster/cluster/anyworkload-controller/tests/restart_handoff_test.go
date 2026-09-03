/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"path/filepath"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/restart"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/test"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var workloadRunnerGVK = metav1.GroupVersionKind{
	Group:   dromosv1alpha1.GroupVersion.Group,
	Version: dromosv1alpha1.GroupVersion.Version,
	Kind:    dromosv1alpha1.WorkloadRunnerKind,
}

// newRestartRequest builds a RestartRequest
func newRestartRequest(name, namespace, workloadName string, gvk metav1.GroupVersionKind) *runv1alpha1.RestartRequest {
	return &runv1alpha1.RestartRequest{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: runv1alpha1.RestartRequestSpec{
			WorkloadRef: runv1alpha1.WorkloadReference{GVK: gvk, Name: workloadName},
		},
	}
}

// setPhase drives a RestartRequest to a phase
func setPhase(ctx context.Context, k8sClient client.Client, rr *runv1alpha1.RestartRequest, phase runv1alpha1.RestartRequestPhase) {
	GinkgoHelper()
	Eventually(func(g Gomega) {
		current := &runv1alpha1.RestartRequest{}
		g.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(rr), current)).To(Succeed())
		current.Status.Phase = phase
		g.Expect(k8sClient.Status().Update(ctx, current)).To(Succeed())
	}, timeout, interval).Should(Succeed())
}

// handoffOccurred reports whether the runner carries the restart lock for rr.
func handoffOccurred(ctx context.Context, k8sClient client.Client, runner *dromosv1alpha1.WorkloadRunner, rr *runv1alpha1.RestartRequest) bool {
	current := &dromosv1alpha1.WorkloadRunner{}
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), current); err != nil {
		return false
	}
	return current.Labels[restart.RequestRefLabel] == string(rr.UID) &&
		current.Spec.ManagedBy == restart.RestartControllerManagedBy
}

var _ = Describe("Restart Handoff Reconciler Integration Tests", Serial, Ordered, func() {
	var (
		k8sManager     ctrl.Manager
		k8sClient      client.Client
		managerCtx     context.Context
		managerCancel  context.CancelFunc
		workloadNS     = "runai-restart-handoff-test"
		controllerName = "restart-handoff-test-controller"
	)

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)

		absCertDir, err := filepath.Abs(webhookCertDir)
		Expect(err).NotTo(HaveOccurred())

		k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme:  scheme.Scheme,
			Metrics: metricsserver.Options{BindAddress: "0"},
			WebhookServer: webhook.NewServer(webhook.Options{
				CertDir: absCertDir,
				Port:    webhookPort,
			}),
		})
		Expect(err).NotTo(HaveOccurred())

		k8sClient = k8sManager.GetClient()

		//nolint:staticcheck // SA1019: matches the other integration tests' recorder wiring
		reconciler := restart.NewHandoffReconciler(k8sClient, k8sManager.GetEventRecorderFor(controllerName)).
			WithName(controllerName)

		Expect(reconciler.SetupWithManager(k8sManager, 10)).To(Succeed())

		fw := framework.NewFramework(framework.NewRegistry(), nil)
		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(managerCtx)).To(Succeed())
		}()

		Expect(k8sClient.Create(managerCtx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: workloadNS},
		})).To(Succeed())

		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)
	})

	AfterAll(func() {
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)
	})

	AfterEach(func() {
		deleteAllForType(managerCtx, k8sClient, &dromosv1alpha1.WorkloadRunnerList{})
		Eventually(func(g Gomega) {
			runners := &dromosv1alpha1.WorkloadRunnerList{}
			g.Expect(k8sClient.List(managerCtx, runners)).To(Succeed())
			g.Expect(runners.Items).To(BeEmpty())
		}, timeout, interval).Should(Succeed())

		deleteAllForType(managerCtx, k8sClient, &runv1alpha1.RestartRequestList{})
		Eventually(func(g Gomega) {
			requests := &runv1alpha1.RestartRequestList{}
			g.Expect(k8sClient.List(managerCtx, requests)).To(Succeed())
			g.Expect(requests.Items).To(BeEmpty())
		}, timeout, interval).Should(Succeed())
	})

	It("takes ownership of the runner when a restart request arrives", func() {
		runnerName := "handoff-runner"
		runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, workloadNS,
			test.NewJobSetBuilder(runnerName, workloadNS).WithDefault().Build(), nil)
		generationBeforeHandoff := runner.Generation

		rr := createResource(managerCtx, k8sClient,
			newRestartRequest("handoff-rr", workloadNS, runner.Name, workloadRunnerGVK))
		Eventually(func() bool {
			return handoffOccurred(managerCtx, k8sClient, runner, rr)
		}, timeout, interval).Should(BeTrue())

		By("setting the label and managedBy")
		currentRunner := &dromosv1alpha1.WorkloadRunner{}
		Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), currentRunner)).To(Succeed())

		Expect(currentRunner.Generation).To(Equal(generationBeforeHandoff + 1))

		By("marking the request InProgress and taking ownership of it")
		Eventually(func(g Gomega) {
			currentRR := &runv1alpha1.RestartRequest{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
			g.Expect(currentRR.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseInProgress))
			g.Expect(currentRR.Finalizers).To(ContainElement(restart.RestartRequestFinalizer))
			g.Expect(currentRR.OwnerReferences).To(HaveLen(1))
			g.Expect(currentRR.OwnerReferences[0].Name).To(Equal(runner.Name))
			g.Expect(currentRR.OwnerReferences[0].Kind).To(Equal(dromosv1alpha1.WorkloadRunnerKind))
			g.Expect(currentRR.OwnerReferences[0].Controller).To(BeNil())
			g.Expect(currentRR.OwnerReferences[0].BlockOwnerDeletion).To(BeNil())
		}, timeout, interval).Should(Succeed())
	})

	DescribeTable("hands the runner back on every terminal phase",
		func(phase runv1alpha1.RestartRequestPhase) {
			runnerName := "handback-runner"
			runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, workloadNS,
				test.NewJobSetBuilder(runnerName, workloadNS).WithDefault().Build(), nil)

			generationAfterHandback := runner.Generation + 2

			rr := createResource(managerCtx, k8sClient,
				newRestartRequest("handback-rr", workloadNS, runner.Name, workloadRunnerGVK))
			Eventually(func() bool {
				return handoffOccurred(managerCtx, k8sClient, runner, rr)
			}, timeout, interval).Should(BeTrue())

			setPhase(managerCtx, k8sClient, rr, phase)

			Eventually(func(g Gomega) {
				currentRunner := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), currentRunner)).To(Succeed())
				g.Expect(currentRunner.Labels).NotTo(HaveKey(restart.RequestRefLabel))
				g.Expect(currentRunner.Spec.ManagedBy).To(BeEmpty())
				g.Expect(currentRunner.Generation).To(Equal(generationAfterHandback))
			}, timeout, interval).Should(Succeed())

			By("not flapping the runner back under restart control")
			Consistently(func(g Gomega) {
				currentRunner := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), currentRunner)).To(Succeed())
				g.Expect(currentRunner.Labels).NotTo(HaveKey(restart.RequestRefLabel))
				g.Expect(currentRunner.Spec.ManagedBy).To(BeEmpty())
				g.Expect(currentRunner.Generation).To(Equal(generationAfterHandback))

				currentRR := &runv1alpha1.RestartRequest{}
				g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
				g.Expect(currentRR.Status.Phase).To(Equal(phase))
			}, time.Second*2, interval).Should(Succeed())
		},
		Entry("Succeeded", runv1alpha1.RestartRequestPhaseSucceeded),
		Entry("Failed", runv1alpha1.RestartRequestPhaseFailed),
	)

	It("refuses a second request while the runner is already locked", func() {
		runnerName := "locked-runner"
		runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, workloadNS,
			test.NewJobSetBuilder(runnerName, workloadNS).WithDefault().Build(), nil)
		winner := createResource(managerCtx, k8sClient,
			newRestartRequest("winner-rr", workloadNS, runner.Name, workloadRunnerGVK))
		Eventually(func() bool {
			return handoffOccurred(managerCtx, k8sClient, runner, winner)
		}, timeout, interval).Should(BeTrue())

		loser := createResource(managerCtx, k8sClient,
			newRestartRequest("loser-rr", workloadNS, runner.Name, workloadRunnerGVK))

		Eventually(func(g Gomega) {
			currentRR := &runv1alpha1.RestartRequest{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(loser), currentRR)).To(Succeed())
			g.Expect(currentRR.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseFailed))
			g.Expect(currentRR.Status.Conditions).NotTo(BeEmpty())
			g.Expect(currentRR.Status.Conditions[0].Message).To(ContainSubstring("already being restarted"))
		}, timeout, interval).Should(Succeed())

		By("leaving the winner's lock on the runner intact after the loser goes terminal")
		Consistently(func(g Gomega) {
			g.Expect(handoffOccurred(managerCtx, k8sClient, runner, winner)).To(BeTrue())
		}, time.Second*2, interval).Should(Succeed())
	})

	It("fails a request whose workload does not exist", func() {
		rr := createResource(managerCtx, k8sClient,
			newRestartRequest("ghost-rr", workloadNS, "no-such-runner", workloadRunnerGVK))

		Eventually(func(g Gomega) {
			currentRR := &runv1alpha1.RestartRequest{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
			g.Expect(currentRR.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseFailed))
			g.Expect(currentRR.Status.Conditions).NotTo(BeEmpty())
			g.Expect(currentRR.Status.Conditions[0].Message).To(ContainSubstring("not found"))
		}, timeout, interval).Should(Succeed())

		By("emitting a WorkloadNotFound event on the request")
		Eventually(func(g Gomega) {
			events := &corev1.EventList{}
			g.Expect(k8sClient.List(managerCtx, events, client.InNamespace(workloadNS))).To(Succeed())
			reasons := make([]string, 0, len(events.Items))
			for _, e := range events.Items {
				if e.InvolvedObject.Name == rr.Name {
					reasons = append(reasons, e.Reason)
				}
			}
			g.Expect(reasons).To(ContainElement("WorkloadNotFound"))
		}, timeout, interval).Should(Succeed())
	})

	It("never rewrites the phase of a terminal request whose workload is gone", func() {
		runnerName := "terminal-rr-runner"
		runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, workloadNS,
			test.NewJobSetBuilder(runnerName, workloadNS).WithDefault().Build(), nil)
		rr := createResource(managerCtx, k8sClient,
			newRestartRequest("terminal-rr", workloadNS, runner.Name, workloadRunnerGVK))
		Eventually(func() bool {
			return handoffOccurred(managerCtx, k8sClient, runner, rr)
		}, timeout, interval).Should(BeTrue())

		setPhase(managerCtx, k8sClient, rr, runv1alpha1.RestartRequestPhaseSucceeded)
		deleteAndPollUntilDeleted(managerCtx, k8sClient, runner)

		Consistently(func(g Gomega) {
			currentRR := &runv1alpha1.RestartRequest{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
			g.Expect(currentRR.Status.Phase).To(Equal(runv1alpha1.RestartRequestPhaseSucceeded))
		}, time.Second*2, interval).Should(Succeed())
	})

	It("blocks request deletion while the workload lives and releases it once gone", func() {
		runnerName := "finalizer-runner"
		runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, workloadNS,
			test.NewJobSetBuilder(runnerName, workloadNS).WithDefault().Build(), nil)
		request := newRestartRequest("finalizer-rr", workloadNS, runner.Name, workloadRunnerGVK)
		request.Labels = map[string]string{commonconstants.RunaiWorkloadNameLabel: runner.Name}
		rr := createResource(managerCtx, k8sClient, request)
		Eventually(func() bool {
			return handoffOccurred(managerCtx, k8sClient, runner, rr)
		}, timeout, interval).Should(BeTrue())

		Expect(k8sClient.Delete(managerCtx, rr)).To(Succeed())

		By("keeping the request alive while the workload exists")

		Eventually(func(g Gomega) {
			currentRR := &runv1alpha1.RestartRequest{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
			g.Expect(currentRR.DeletionTimestamp).NotTo(BeNil())
		}, timeout, interval).Should(Succeed())
		Consistently(func(g Gomega) {
			currentRR := &runv1alpha1.RestartRequest{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), currentRR)).To(Succeed())
			g.Expect(currentRR.Finalizers).To(ContainElement(restart.RestartRequestFinalizer))
		}, time.Second*2, interval).Should(Succeed())

		By("releasing it once the workload is deleted")
		Expect(k8sClient.Delete(managerCtx, runner)).To(Succeed())
		Eventually(func() bool {
			err := k8sClient.Get(managerCtx, client.ObjectKeyFromObject(rr), &runv1alpha1.RestartRequest{})
			return errors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue())
	})

})
