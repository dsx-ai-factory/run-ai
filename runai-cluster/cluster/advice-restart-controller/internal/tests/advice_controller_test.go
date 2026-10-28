/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	heimdallv1alpha1 "github.com/dsx-ai-factory/nvidia-recovery/advice-operator/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/advice-restart-controller/pkg/controller"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const testNS = "advice-test"

// setupWLRAndJob creates a WorkloadRunner and a PyTorchJob owned by it
func setupWLRAndJob(
	ctx context.Context,
	c client.Client,
	name, phase string,
	nodeAffinity ...*corev1.NodeAffinity,
) {
	var instruction *corev1.NodeAffinity
	if len(nodeAffinity) > 0 {
		instruction = nodeAffinity[0]
	}
	wlr := createWLR(ctx, c, name, testNS, phase, instruction)
	createPyTorchJob(ctx, c, name, testNS, name, string(wlr.GetUID()))
}

// onlyRestartRequest returns the single RestartRequest created for the advice.
func onlyRestartRequest(g Gomega, c client.Client, adviceName string) *runv1alpha1.RestartRequest {
	rrList := &runv1alpha1.RestartRequestList{}
	g.Expect(c.List(ctx, rrList,
		client.InNamespace(testNS),
		client.MatchingLabels{"run.ai/advice-name": adviceName},
	)).To(Succeed())
	g.Expect(rrList.Items).To(HaveLen(1))
	return &rrList.Items[0]
}

// excludedNodes unwraps the single node-affinity override back to the hostnames it excludes.
func excludedNodes(g Gomega, rr *runv1alpha1.RestartRequest) []string {
	g.Expect(rr.Spec.SpecOverrides).To(HaveLen(1))
	op := rr.Spec.SpecOverrides[0]
	g.Expect(op.Op).To(Equal(runv1alpha1.JSONPatchOperationAdd))
	g.Expect(op.Path).To(Equal("/spec/instructions/nodeAffinity"))
	g.Expect(op.Value).NotTo(BeNil())

	value := map[string]*corev1.NodeAffinity{}
	g.Expect(json.Unmarshal(op.Value.Raw, &value)).To(Succeed())
	affinity := value["nodeAffinity"]
	g.Expect(affinity).NotTo(BeNil())
	g.Expect(affinity.RequiredDuringSchedulingIgnoredDuringExecution).NotTo(BeNil())

	terms := affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	g.Expect(terms).To(HaveLen(1))
	g.Expect(terms[0].MatchExpressions).To(HaveLen(1))
	g.Expect(terms[0].MatchExpressions[0].Key).To(Equal(corev1.LabelHostname))
	g.Expect(terms[0].MatchExpressions[0].Operator).To(Equal(corev1.NodeSelectorOpNotIn))
	return terms[0].MatchExpressions[0].Values
}

var _ = Describe("AdviceController", Serial, Ordered, func() {
	var (
		k8sClient    client.Client
		mgrCancel    context.CancelFunc
		mgrDone      chan struct{}
		fakeRecorder *record.FakeRecorder
	)

	BeforeAll(func() {
		By("creating test namespace")
		s := runtime.NewScheme()
		Expect(clientgoscheme.AddToScheme(s)).To(Succeed())
		Expect(heimdallv1alpha1.AddToScheme(s)).To(Succeed())
		Expect(runv1alpha1.AddToScheme(s)).To(Succeed())
		Expect(dromosv1alpha1.AddToScheme(s)).To(Succeed())
		Expect(kubeflowv1.AddToScheme(s)).To(Succeed())

		var err error
		k8sClient, err = client.New(cfg, client.Options{Scheme: s})
		Expect(err).NotTo(HaveOccurred())
		createNamespace(ctx, k8sClient, testNS)

		By("setting up controller manager")
		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: s,
			Metrics: metricsserver.Options{
				BindAddress: "0",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		fakeRecorder = record.NewFakeRecorder(100)
		adviceCtrl := controller.NewAdviceController(mgr.GetClient(), fakeRecorder, "action", 3, 3)
		Expect(adviceCtrl.SetupWithManager(mgr)).To(Succeed())

		var mgrCtx context.Context
		mgrCtx, mgrCancel = context.WithCancel(ctx)
		mgrDone = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(mgrDone)
			Expect(mgr.Start(mgrCtx)).To(Succeed())
		}()
	})

	AfterAll(func() {
		mgrCancel()
		// Wait for the manager to fully shut down before finishing, so nothing leaks
		// into a subsequent run.
		Eventually(mgrDone, timeout).Should(BeClosed())
	})

	AfterEach(func() {
		drainEvents(fakeRecorder)
		deleteAllAdvices(ctx, k8sClient, testNS)
		deleteAllRestartRequests(ctx, k8sClient, testNS)
		deleteAllPyTorchJobs(ctx, k8sClient, testNS)
		deleteAllWLRs(ctx, k8sClient, testNS)
	})

	Context("RestartRequest State Check", func() {
		It("should create RestartRequest when no existing RR", func() {
			setupWLRAndJob(ctx, k8sClient, "wlr-no-rr", "Running")
			createAdvice(ctx, k8sClient, "advice-no-rr", testNS, "wlr-no-rr", heimdallv1alpha1.JobStateRestarted)

			Eventually(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": "advice-no-rr"},
				)).To(Succeed())
				return len(rrList.Items)
			}, timeout, interval).Should(Equal(1))

			Eventually(fakeRecorder.Events).Should(Receive(ContainSubstring("AdviceReceivedOnWorkload")))
			Eventually(fakeRecorder.Events).Should(Receive(ContainSubstring("RestartRequestCreated")))
		})

		It("should skip when Succeeded RR exists", func() {
			adviceName := "advice-rr-succeeded"
			createRestartRequestWithStatus(ctx, k8sClient, "rr-succeeded", testNS,
				map[string]string{
					"run.ai/advice-name": adviceName,
				}, runv1alpha1.RestartRequestPhaseSucceeded)

			createAdvice(ctx, k8sClient, adviceName, testNS, "nonexistent-job",
				heimdallv1alpha1.JobStateRestarted)

			Consistently(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": adviceName},
				)).To(Succeed())
				return len(rrList.Items)
			}, consistentlyTimeout, interval).Should(Equal(1))
		})

		It("should not create new RR when InProgress RR exists", func() {
			adviceName := "advice-rr-inprog"
			createRestartRequestWithStatus(ctx, k8sClient, "rr-inprogress", testNS,
				map[string]string{
					"run.ai/advice-name": adviceName,
				}, runv1alpha1.RestartRequestPhaseInProgress)

			createAdvice(ctx, k8sClient, adviceName, testNS, "nonexistent-job",
				heimdallv1alpha1.JobStateRestarted)

			Consistently(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": adviceName},
				)).To(Succeed())
				return len(rrList.Items)
			}, consistentlyTimeout, interval).Should(Equal(1))
		})

		It("should create new RR when Failed RR exists (retry)", func() {
			adviceName := "advice-rr-failed"
			createRestartRequestWithStatus(ctx, k8sClient, "rr-failed", testNS,
				map[string]string{
					"run.ai/advice-name": adviceName,
				}, runv1alpha1.RestartRequestPhaseFailed)

			setupWLRAndJob(ctx, k8sClient, "wlr-rr-failed", "Running")
			createAdvice(ctx, k8sClient, adviceName, testNS, "wlr-rr-failed",
				heimdallv1alpha1.JobStateRestarted)

			Eventually(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": adviceName},
				)).To(Succeed())
				return len(rrList.Items)
			}, timeout, interval).Should(Equal(2))
		})
	})

	Context("RestartRequest Watch and Retry", func() {
		It("should retry when RR transitions to Failed", func() {
			adviceName := "advice-rr-watch"
			setupWLRAndJob(ctx, k8sClient, "wlr-rr-watch", "Running")
			createAdvice(ctx, k8sClient, adviceName, testNS, "wlr-rr-watch",
				heimdallv1alpha1.JobStateRestarted)

			// Wait for first RR to be created
			var firstRR *runv1alpha1.RestartRequest
			Eventually(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": adviceName},
				)).To(Succeed())
				if len(rrList.Items) > 0 {
					firstRR = &rrList.Items[0]
				}
				return len(rrList.Items)
			}, timeout, interval).Should(Equal(1))

			// Transition the RR to Failed — this should trigger re-reconciliation via the watch
			Eventually(func() error {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(firstRR), firstRR); err != nil {
					return err
				}
				firstRR.Status.Phase = runv1alpha1.RestartRequestPhaseFailed
				return k8sClient.Status().Update(ctx, firstRR)
			}, timeout, interval).Should(Succeed())

			// A second RR should be created automatically (without modifying the Advice)
			Eventually(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": adviceName},
				)).To(Succeed())
				return len(rrList.Items)
			}, timeout, interval).Should(Equal(2))
		})

		It("should stop retrying after max retries reached", func() {
			adviceName := "advice-max-retry"
			setupWLRAndJob(ctx, k8sClient, "wlr-max-retry", "Running")

			// Pre-create 3 failed RRs (matching maxRetriesPerAdvice=3)
			for i := range 3 {
				createRestartRequestWithStatus(ctx, k8sClient,
					fmt.Sprintf("rr-max-retry-%d", i), testNS,
					map[string]string{
						"run.ai/advice-name": adviceName,
						"run.ai/created-by":  "advice-restart-controller",
					}, runv1alpha1.RestartRequestPhaseFailed)
			}

			createAdvice(ctx, k8sClient, adviceName, testNS, "wlr-max-retry",
				heimdallv1alpha1.JobStateRestarted)

			// No new RR should be created since retry limit is reached
			Consistently(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": adviceName},
				)).To(Succeed())
				return len(rrList.Items)
			}, consistentlyTimeout, interval).Should(Equal(3))
		})
	})

	Context("Advice Validation", func() {
		It("should not create RR when no job with matching name", func() {
			adviceName := "advice-no-job"
			createAdvice(ctx, k8sClient, adviceName, testNS, "nonexistent-job",
				heimdallv1alpha1.JobStateRestarted)

			Consistently(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": adviceName},
				)).To(Succeed())
				return len(rrList.Items)
			}, consistentlyTimeout, interval).Should(Equal(0))
		})

		It("should not create RR when job has no owner", func() {
			adviceName := "advice-no-owner"
			// Create PyTorchJob with no owner reference
			createPyTorchJob(ctx, k8sClient, "job-no-owner", testNS, "", "")
			createAdvice(ctx, k8sClient, adviceName, testNS, "job-no-owner",
				heimdallv1alpha1.JobStateRestarted)

			Consistently(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": adviceName},
				)).To(Succeed())
				return len(rrList.Items)
			}, consistentlyTimeout, interval).Should(Equal(0))
		})
	})

	Context("Workload Eligibility", func() {
		It("should create RR when workload is Running", func() {
			setupWLRAndJob(ctx, k8sClient, "wlr-running", "Running")
			createAdvice(ctx, k8sClient, "advice-running", testNS, "wlr-running",
				heimdallv1alpha1.JobStateRestarted)

			Eventually(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": "advice-running"},
				)).To(Succeed())
				return len(rrList.Items)
			}, timeout, interval).Should(Equal(1))
		})

		It("should create RR when workload is Degraded", func() {
			setupWLRAndJob(ctx, k8sClient, "wlr-degraded", "Degraded")
			createAdvice(ctx, k8sClient, "advice-degraded", testNS, "wlr-degraded",
				heimdallv1alpha1.JobStateRestarted)

			Eventually(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": "advice-degraded"},
				)).To(Succeed())
				return len(rrList.Items)
			}, timeout, interval).Should(Equal(1))
		})

		It("should not create RR when workload is Suspended", func() {
			setupWLRAndJob(ctx, k8sClient, "wlr-suspended", "Suspended")
			createAdvice(ctx, k8sClient, "advice-suspended", testNS, "wlr-suspended",
				heimdallv1alpha1.JobStateRestarted)

			Consistently(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": "advice-suspended"},
				)).To(Succeed())
				return len(rrList.Items)
			}, consistentlyTimeout, interval).Should(Equal(0))

			Eventually(fakeRecorder.Events).Should(Receive(ContainSubstring("AdviceIgnored")))
		})

		It("should not create RR when workload is Completed", func() {
			setupWLRAndJob(ctx, k8sClient, "wlr-completed", "Completed")
			createAdvice(ctx, k8sClient, "advice-completed", testNS, "wlr-completed",
				heimdallv1alpha1.JobStateRestarted)

			Consistently(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": "advice-completed"},
				)).To(Succeed())
				return len(rrList.Items)
			}, consistentlyTimeout, interval).Should(Equal(0))
		})

		It("should create RR when workload is Resuming", func() {
			setupWLRAndJob(ctx, k8sClient, "wlr-resuming", "Resuming")
			createAdvice(ctx, k8sClient, "advice-resuming", testNS, "wlr-resuming",
				heimdallv1alpha1.JobStateRestarted)

			Eventually(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": "advice-resuming"},
				)).To(Succeed())
				return len(rrList.Items)
			}, timeout, interval).Should(Equal(1))
		})

		DescribeTable("should not create RR in a non-restartable phase",
			func(phase string) {
				suffix := strings.ToLower(phase)
				setupWLRAndJob(ctx, k8sClient, "wlr-"+suffix, phase)
				createAdvice(ctx, k8sClient, "advice-"+suffix, testNS, "wlr-"+suffix,
					heimdallv1alpha1.JobStateRestarted)

				Consistently(func() int {
					rrList := &runv1alpha1.RestartRequestList{}
					Expect(k8sClient.List(ctx, rrList,
						client.InNamespace(testNS),
						client.MatchingLabels{"run.ai/advice-name": "advice-" + suffix},
					)).To(Succeed())
					return len(rrList.Items)
				}, consistentlyTimeout, interval).Should(Equal(0))

				// the event proves the eligibility gate was reached - without it a rotted
				// fixture (owner chain, registry key) would skip earlier and still pass
				Eventually(fakeRecorder.Events).Should(Receive(ContainSubstring("AdviceIgnored")))
			},
			Entry("Suspending", "Suspending"),
			Entry("Stopped", "Stopped"),
		)

		It("should not create RR when at restart cap", func() {
			wlrName := "wlr-at-cap"
			setupWLRAndJob(ctx, k8sClient, wlrName, "Running")

			// Create 3 succeeded RestartRequests for this workload
			for i := range 3 {
				createRestartRequestWithStatus(ctx, k8sClient,
					fmt.Sprintf("rr-cap-%d", i), testNS,
					map[string]string{
						"run.ai/workload-name": wlrName,
						"run.ai/created-by":    "advice-restart-controller",
					}, runv1alpha1.RestartRequestPhaseSucceeded)
			}

			createAdvice(ctx, k8sClient, "advice-at-cap", testNS, wlrName,
				heimdallv1alpha1.JobStateRestarted)

			Consistently(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": "advice-at-cap"},
				)).To(Succeed())
				return len(rrList.Items)
			}, consistentlyTimeout, interval).Should(Equal(0))

			Eventually(fakeRecorder.Events).Should(Receive(ContainSubstring("AdviceIgnored")))
		})
	})

	Context("RestartRequest Creation", func() {
		It("should create RR targeting the WorkloadRunner node-affinity instruction", func() {
			wlrName := "wlr-with-excl"
			setupWLRAndJob(ctx, k8sClient, wlrName, "Running")
			createAdvice(ctx, k8sClient, "advice-with-excl", testNS, wlrName,
				heimdallv1alpha1.JobStateRestarted, "node-2", "node-1")

			Eventually(func(g Gomega) {
				rr := onlyRestartRequest(g, k8sClient, "advice-with-excl")

				wlr := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: wlrName, Namespace: testNS}, wlr)).To(Succeed())

				g.Expect(rr.Spec.WorkloadRef.GVK).To(Equal(metav1.GroupVersionKind{
					Group: "run.ai", Version: "v1alpha1", Kind: "WorkloadRunner",
				}))
				g.Expect(rr.Spec.WorkloadRef.Name).To(Equal(wlrName))
				g.Expect(rr.Spec.WorkloadRef.Generation).To(Equal(wlr.GetGeneration()))
				g.Expect(excludedNodes(g, rr)).To(Equal([]string{"node-1", "node-2"}))

				g.Expect(rr.Labels["run.ai/workload-name"]).To(Equal(wlrName))
				g.Expect(rr.Labels["run.ai/created-by"]).To(Equal("advice-restart-controller"))
			}, timeout, interval).Should(Succeed())

			Eventually(fakeRecorder.Events).Should(Receive(ContainSubstring("AdviceReceivedOnWorkload")))
		})

		It("should union the exclusions with the ones already on the workload", func() {
			existingExclusions := &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      corev1.LabelHostname,
							Operator: corev1.NodeSelectorOpNotIn,
							Values:   []string{"node-b", "node-a"},
						}},
					}},
				},
			}
			setupWLRAndJob(ctx, k8sClient, "wlr-union", "Running", existingExclusions)
			createAdvice(ctx, k8sClient, "advice-union", testNS, "wlr-union",
				heimdallv1alpha1.JobStateRestarted, "node-c", "node-a")

			Eventually(func(g Gomega) {
				rr := onlyRestartRequest(g, k8sClient, "advice-union")
				g.Expect(excludedNodes(g, rr)).To(Equal([]string{"node-a", "node-b", "node-c"}))
			}, timeout, interval).Should(Succeed())
		})

		It("should create one RR per advice when two advices target the same workload", func() {
			// no-active-RR rule is what serializes them in a real cluster
			setupWLRAndJob(ctx, k8sClient, "wlr-two-advices", "Running")
			createAdvice(ctx, k8sClient, "advice-two-a", testNS, "wlr-two-advices",
				heimdallv1alpha1.JobStateRestarted, "node-1")
			createAdvice(ctx, k8sClient, "advice-two-b", testNS, "wlr-two-advices",
				heimdallv1alpha1.JobStateRestarted, "node-2")

			for adviceName, node := range map[string]string{"advice-two-a": "node-1", "advice-two-b": "node-2"} {
				Eventually(func(g Gomega) {
					rr := onlyRestartRequest(g, k8sClient, adviceName)
					g.Expect(excludedNodes(g, rr)).To(Equal([]string{node}))
				}, timeout, interval).Should(Succeed())
			}
		})

		It("should create RR without specOverrides when no exclusions", func() {
			setupWLRAndJob(ctx, k8sClient, "wlr-no-excl", "Running")
			createAdvice(ctx, k8sClient, "advice-no-excl", testNS, "wlr-no-excl",
				heimdallv1alpha1.JobStateRestarted)

			Eventually(func(g Gomega) {
				rrList := &runv1alpha1.RestartRequestList{}
				g.Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": "advice-no-excl"},
				)).To(Succeed())
				g.Expect(rrList.Items).NotTo(BeEmpty())

				rr := rrList.Items[0]
				g.Expect(rr.Spec.SpecOverrides).To(BeNil())
			}, timeout, interval).Should(Succeed())
		})

		It("should set run.ai/workload-id label on created RR matching the WorkloadRunner UID", func() {
			wlrName := "wlr-wl-id-label"
			setupWLRAndJob(ctx, k8sClient, wlrName, "Running")
			createAdvice(ctx, k8sClient, "advice-wl-id-label", testNS, wlrName,
				heimdallv1alpha1.JobStateRestarted)

			Eventually(func(g Gomega) {
				rr := onlyRestartRequest(g, k8sClient, "advice-wl-id-label")
				wlr := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: wlrName, Namespace: testNS}, wlr)).To(Succeed())
				g.Expect(rr.Labels["run.ai/workload-id"]).To(Equal(string(wlr.GetUID())))
			}, timeout, interval).Should(Succeed())
		})
	})

	Context("Non-RESTARTED state", func() {
		It("should not create RR for TERMINATED advice", func() {
			adviceName := "advice-terminated"
			terminated := heimdallv1alpha1.JobStateTerminated
			advice := &heimdallv1alpha1.Advice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      adviceName,
					Namespace: testNS,
				},
				Spec: heimdallv1alpha1.AdviceSpec{
					Job: heimdallv1alpha1.AdviceJobReference{
						HeimdallID:   "heimdall-test",
						APIVersion:   "kubeflow.org/v1",
						Kind:         "PyTorchJob",
						Name:         "some-job",
						JobID:        "some-scheduler-id",
						AttemptIndex: 0,
					},
					Advice: heimdallv1alpha1.AdviceRecommendation{
						State: &terminated,
					},
				},
			}
			Expect(k8sClient.Create(ctx, advice)).To(Succeed())

			Consistently(func() int {
				rrList := &runv1alpha1.RestartRequestList{}
				Expect(k8sClient.List(ctx, rrList,
					client.InNamespace(testNS),
					client.MatchingLabels{"run.ai/advice-name": adviceName},
				)).To(Succeed())
				return len(rrList.Items)
			}, consistentlyTimeout, interval).Should(Equal(0))
		})
	})

})
