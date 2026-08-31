/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gomodules.xyz/jsonpatch/v2"
	"k8s.io/apimachinery/pkg/api/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/controller"
	internaltest "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/internaltest"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/test"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/test/plugins"
	sdkv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

const (
	testNamespace = "test-namespace"

	jobsetGroup   = "jobset.x-k8s.io"
	jobsetVersion = "v1alpha2"
	jobsetKind    = "JobSet"
)

var jobsetGVK = schema.GroupVersionKind{
	Group:   jobsetGroup,
	Version: jobsetVersion,
	Kind:    jobsetKind,
}

var _ = Describe("WorkloadRunnerController", Serial, Ordered, func() {
	var (
		ns                       *corev1.Namespace
		k8sManager               ctrl.Manager
		k8sClient                client.Client
		managerCtx               context.Context
		managerCancel            context.CancelFunc
		phasePlugin              *plugins.InjectableStatusPlugin
		additionalResourcePlugin *plugins.InjectableAdditionalResourcePlugin
	)

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)

		phasePlugin = plugins.NewInjectableStatusPlugin()
		additionalResourcePlugin = plugins.NewInjectableAdditionalResourcePlugin()
		var err error
		k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
			Metrics: metricsserver.Options{
				BindAddress: "0",
			},
			WebhookServer: webhook.NewServer(webhook.Options{
				CertDir: webhookCertDir,
				Port:    webhookPort,
			}),
		})
		Expect(err).NotTo(HaveOccurred())

		registry := framework.NewRegistry(
			plugins.NewResourceMutateAddEnvVarPlugin(),
			plugins.NewResourceCreateConfigMapPlugin(),
			phasePlugin,
			additionalResourcePlugin,
		)
		fw := framework.NewFramework(registry, nil)
		workloadCtrl := controller.NewWorkloadRunnerController(k8sManager.GetClient(), fw, k8sManager.GetEventRecorderFor("workloadrunner-controller-test")).WithName("workloadrunner-controller-test")
		err = workloadCtrl.SetupWithManager(managerCtx, k8sManager)
		Expect(err).NotTo(HaveOccurred())
		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		go func() {
			defer GinkgoRecover()
			err = k8sManager.Start(managerCtx)
			Expect(err).NotTo(HaveOccurred())
		}()

		k8sClient = k8sManager.GetClient()

		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		_ = k8sClient.Create(managerCtx, ns)
	})

	AfterAll(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)
	})

	Context("submitting a workload", Ordered, func() {
		var (
			runner       *dromosv1alpha1.WorkloadRunner
			karta        *kartav1alpha1.Karta
			runnerName   = "test-runner1"
			workloadName = "test-jobset1"
		)

		BeforeAll(func() {
			karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			runner = createWorkloadRunner(managerCtx, k8sClient, runnerName, testNamespace, newJobSet(workloadName), nil)
		})

		AfterAll(func() {
			deleteAllForType(managerCtx, k8sClient, &dromosv1alpha1.WorkloadRunnerList{})
			deleteAllForType(managerCtx, k8sClient, &kartav1alpha1.KartaList{})
			deleteAllForType(managerCtx, k8sClient, &jobsetv1alpha2.JobSetList{})

			deleteAndPollUntilDeleted(managerCtx, k8sClient, runner)
			deleteAndPollUntilDeleted(managerCtx, k8sClient, karta)
			deleteAllJobSets(managerCtx, k8sClient)
		})

		It("creates workload resource", func() {
			jobset := &unstructured.Unstructured{}
			jobset.SetGroupVersionKind(jobsetGVK)

			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      workloadName,
					Namespace: testNamespace,
				}, jobset)
			}, timeout, interval).Should(Succeed())
		})

		It("sets owner reference", func() {
			jobset := &unstructured.Unstructured{}
			jobset.SetGroupVersionKind(jobsetGVK)

			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      workloadName,
					Namespace: testNamespace,
				}, jobset)
			}, timeout, interval).Should(Succeed())

			ownerRefs := jobset.GetOwnerReferences()
			Expect(ownerRefs).To(HaveLen(1))
			Expect(ownerRefs[0].Kind).To(Equal(dromosv1alpha1.WorkloadRunnerKind))
			Expect(ownerRefs[0].Name).To(Equal(runnerName))
			Expect(ownerRefs[0].Controller).NotTo(BeNil())
			Expect(*ownerRefs[0].Controller).To(BeTrue())
		})

		It("updates WorkloadRunner status to Initializing", func() {
			updated := &dromosv1alpha1.WorkloadRunner{}

			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      runnerName,
					Namespace: testNamespace,
				}, updated)
				// Status should be Unknown as the jobset conditions are not set in test.
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(updated.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseUnknown))
			}, timeout, interval).Should(Succeed())

		})
	})

	Context("missing prerequisites", Ordered, func() {
		Context("when Karta does not exist", func() {
			var runner *dromosv1alpha1.WorkloadRunner
			runnerName := "no-karta-runner"
			workloadName := "no-karta-jobset"

			BeforeAll(func() {
				runner = createWorkloadRunner(managerCtx, k8sClient, runnerName, testNamespace, newJobSet(workloadName), nil)
			})

			AfterAll(func() {
				deleteAndPollUntilDeleted(managerCtx, k8sClient, runner)
				deleteAllJobSets(managerCtx, k8sClient)
			})

			It("updates phase to Unknown", func() {
				updated := &dromosv1alpha1.WorkloadRunner{}

				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
					return err == nil && updated.Status.Phase != ""
				}, timeout, interval).Should(BeTrue())

				Expect(updated.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseUnknown))
				Expect(updated.Status.Issues).To(HaveLen(1))
				Expect(updated.Status.Issues[0].Kind).To(Equal("WorkloadRunner"))
				Expect(updated.Status.Issues[0].Name).To(Equal(runnerName))
				Expect(updated.Status.Issues[0].Group).To(Equal("Internal"))
				Expect(updated.Status.Issues[0].Description).To(ContainSubstring("Karta for GVK"))
			})

			It("sets Failed condition True with KartaNotFound reason", func() {
				updated := &dromosv1alpha1.WorkloadRunner{}

				Eventually(func() bool {
					_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
					cond := meta.FindStatusCondition(updated.Status.Conditions, dromosv1alpha1.ConditionTypePrerequisitesPassed)
					return cond != nil && cond.Reason != ""
				}, timeout, interval).Should(BeTrue())

				cond := meta.FindStatusCondition(updated.Status.Conditions, dromosv1alpha1.ConditionTypePrerequisitesPassed)
				Expect(cond).NotTo(BeNil())
				Expect(cond.Reason).To(Equal(dromosv1alpha1.ReasonKartaNotFound))
				Expect(cond.Status).To(Equal(metav1.ConditionFalse))

				Expect(updated.Status.Issues).To(HaveLen(1))
				Expect(updated.Status.Issues[0].Kind).To(Equal("WorkloadRunner"))
				Expect(updated.Status.Issues[0].Name).To(Equal(runnerName))
				Expect(updated.Status.Issues[0].Group).To(Equal("Internal"))
				Expect(updated.Status.Issues[0].Description).To(ContainSubstring("Karta for GVK"))
			})
		})

		Context("when underlying resource fails", func() {
			var runner *dromosv1alpha1.WorkloadRunner
			var karta *kartav1alpha1.Karta
			runnerName := "resource-failed-runner"
			workloadName := "resource-failed-jobset"

			BeforeAll(func() {
				karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
				runner = createWorkloadRunner(managerCtx, k8sClient, runnerName, testNamespace, newJobSet(workloadName), nil)

				// Wait for the JobSet to be created
				jobset := &unstructured.Unstructured{}
				jobset.SetGroupVersionKind(jobsetGVK)
				Eventually(func() error {
					return k8sClient.Get(ctx, types.NamespacedName{
						Name:      workloadName,
						Namespace: testNamespace,
					}, jobset)
				}, timeout, interval).Should(Succeed())

				// Set the JobSet status to Failed by adding a Failed condition
				err := unstructured.SetNestedSlice(jobset.Object, []interface{}{
					map[string]interface{}{
						"type":               "Failed",
						"status":             "True",
						"reason":             "JobSetFailed",
						"message":            "JobSet failed",
						"lastTransitionTime": metav1.Now().Format(time.RFC3339),
					},
				}, "status", "conditions")
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Status().Update(ctx, jobset)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterAll(func() {
				deleteAndPollUntilDeleted(managerCtx, k8sClient, runner)
				deleteAndPollUntilDeleted(managerCtx, k8sClient, karta)
				deleteAllJobSets(managerCtx, k8sClient)
			})

			It("updates phase to Failed when underlying resource fails", func() {
				updated := &dromosv1alpha1.WorkloadRunner{}

				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
					return err == nil && updated.Status.Phase == dromosv1alpha1.WorkloadRunnerPhaseFailed
				}, timeout, interval).Should(BeTrue())

				Expect(updated.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseFailed))
				Expect(updated.Status.ResourcePhase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseFailed))
			})
		})

		Context("when CRD does not exist", Ordered, func() {
			var (
				karta        *kartav1alpha1.Karta
				runner       *dromosv1alpha1.WorkloadRunner
				runnerName   = "no-crd-runner"
				workloadName = "no-crd-jobset"
			)

			BeforeAll(func() {
				karta = createKartaWithStatus(managerCtx, k8sClient, newKarta(test.NewValidJobSetKarta(), true, false))
				runner = createWorkloadRunner(managerCtx, k8sClient, runnerName, testNamespace, newJobSet(workloadName), nil)
			})

			AfterAll(func() {
				deleteAndPollUntilDeleted(managerCtx, k8sClient, runner)
				deleteAndPollUntilDeleted(managerCtx, k8sClient, karta)
				deleteAllJobSets(managerCtx, k8sClient)
			})

			It("updates phase to Unknown", func() {
				updated := &dromosv1alpha1.WorkloadRunner{}
				Eventually(func() bool {
					_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
					return updated.Status.Phase != ""
				}, timeout, interval).Should(BeTrue())

				Expect(updated.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseUnknown))
			})

			It("sets Failed condition True with KartaNotReady reason", func() {
				updated := &dromosv1alpha1.WorkloadRunner{}

				Eventually(func() bool {
					_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
					cond := meta.FindStatusCondition(updated.Status.Conditions, dromosv1alpha1.ConditionTypePrerequisitesPassed)
					return cond != nil && cond.Reason != ""
				}, timeout, interval).Should(BeTrue())

				cond := meta.FindStatusCondition(updated.Status.Conditions, dromosv1alpha1.ConditionTypePrerequisitesPassed)
				Expect(cond).NotTo(BeNil())
				Expect(cond.Reason).To(Equal(dromosv1alpha1.ReasonKartaNotReady))
				Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			})
		})

		Context("when Karta is not ready", Ordered, func() {
			var (
				karta        *kartav1alpha1.Karta
				runner       *dromosv1alpha1.WorkloadRunner
				runnerName   = "not-ready-runner"
				workloadName = "not-ready-jobset"
			)

			BeforeAll(func() {
				karta = createKartaWithStatus(managerCtx, k8sClient, newKarta(test.NewValidJobSetKarta(), false, true))
				runner = createWorkloadRunner(managerCtx, k8sClient, runnerName, testNamespace, newJobSet(workloadName), nil)
			})

			AfterAll(func() {
				deleteAndPollUntilDeleted(managerCtx, k8sClient, runner)
				deleteAndPollUntilDeleted(managerCtx, k8sClient, karta)
				deleteAllJobSets(managerCtx, k8sClient)
			})

			It("sets Failed condition to True with KartaNotReady reason", func() {
				updated := &dromosv1alpha1.WorkloadRunner{}

				Eventually(func() bool {
					_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
					cond := meta.FindStatusCondition(updated.Status.Conditions, dromosv1alpha1.ConditionTypePrerequisitesPassed)
					return cond != nil && cond.Reason != ""
				}, timeout, interval).Should(BeTrue())

				cond := meta.FindStatusCondition(updated.Status.Conditions, dromosv1alpha1.ConditionTypePrerequisitesPassed)
				Expect(cond).NotTo(BeNil())
				Expect(cond.Reason).To(Equal(dromosv1alpha1.ReasonKartaNotReady))
				Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			})
		})
	})

	Context("phase calculation", Ordered, func() {
		var (
			karta  *kartav1alpha1.Karta
			runner *dromosv1alpha1.WorkloadRunner
		)

		AfterEach(func() {
			deleteAndPollUntilDeleted(managerCtx, k8sClient, runner)
			deleteAndPollUntilDeleted(managerCtx, k8sClient, karta)
			deleteAllJobSets(managerCtx, k8sClient)
		})

		It("calculates Initializing phase from resource conditions", func() {
			karta = internaltest.NewJobSetKartaBuilder("karta-initializing").
				WithDefaults().
				WithStatusMatcher(kartav1alpha1.InitializingStatus, []kartav1alpha1.ExpectedCondition{
					{Type: "StartupPolicyInProgress", Status: ptr.To("True")},
				}).
				Build()
			karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(karta))

			runner = createWorkloadRunner(managerCtx, k8sClient, "runner-init", testNamespace, newJobSet("jobset-init"), nil)

			updateJobSetStatus(managerCtx, k8sClient, "jobset-init", testNamespace, []interface{}{
				makeCondition("StartupPolicyInProgress", "True"),
			})

			Eventually(func() bool {
				updated := &dromosv1alpha1.WorkloadRunner{}
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
				return updated.Status.ResourcePhase == dromosv1alpha1.WorkloadRunnerPhaseInitializing &&
					updated.Status.Phase == dromosv1alpha1.WorkloadRunnerPhaseInitializing
			}, timeout, interval).Should(BeTrue())
		})

		It("calculates Running phase from resource conditions", func() {
			karta = internaltest.NewJobSetKartaBuilder("karta-running").
				WithDefaults().
				WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
					{Type: "StartupPolicyCompleted", Status: ptr.To("True")},
					{Type: "Completed", Status: ptr.To("False")},
					{Type: "Failed", Status: ptr.To("False")},
				}).
				Build()
			karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(karta))

			runner = createWorkloadRunner(managerCtx, k8sClient, "runner-running", testNamespace, newJobSet("jobset-running"), nil)

			updateJobSetStatus(managerCtx, k8sClient, "jobset-running", testNamespace, []interface{}{
				makeCondition("StartupPolicyCompleted", "True"),
				makeCondition("Completed", "False"),
				makeCondition("Failed", "False"),
			})

			Eventually(func() bool {
				updated := &dromosv1alpha1.WorkloadRunner{}
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
				return updated.Status.ResourcePhase == dromosv1alpha1.WorkloadRunnerPhaseRunning &&
					updated.Status.Phase == dromosv1alpha1.WorkloadRunnerPhaseRunning
			}, timeout, interval).Should(BeTrue())
		})

		It("calculates Completed phase from resource conditions", func() {
			karta = internaltest.NewJobSetKartaBuilder("karta-completed").
				WithDefaults().
				WithStatusMatcher(kartav1alpha1.CompletedStatus, []kartav1alpha1.ExpectedCondition{
					{Type: "Completed", Status: ptr.To("True")},
				}).
				Build()
			karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(karta))

			runner = createWorkloadRunner(managerCtx, k8sClient, "runner-completed", testNamespace, newJobSet("jobset-completed"), nil)

			updateJobSetStatus(managerCtx, k8sClient, "jobset-completed", testNamespace, []interface{}{
				makeCondition("Completed", "True"),
			})

			Eventually(func() bool {
				updated := &dromosv1alpha1.WorkloadRunner{}
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
				return updated.Status.ResourcePhase == dromosv1alpha1.WorkloadRunnerPhaseCompleted &&
					updated.Status.Phase == dromosv1alpha1.WorkloadRunnerPhaseCompleted
			}, timeout, interval).Should(BeTrue())
		})

		It("calculates Failed phase from resource conditions", func() {
			karta = internaltest.NewJobSetKartaBuilder("karta-failed").
				WithDefaults().
				WithStatusMatcher(kartav1alpha1.FailedStatus, []kartav1alpha1.ExpectedCondition{
					{Type: "Failed", Status: ptr.To("True")},
				}).
				Build()
			karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(karta))

			runner = createWorkloadRunner(managerCtx, k8sClient, "runner-failed", testNamespace, newJobSet("jobset-failed"), nil)

			updateJobSetStatus(managerCtx, k8sClient, "jobset-failed", testNamespace, []interface{}{
				makeCondition("Failed", "True"),
			})

			Eventually(func() bool {
				updated := &dromosv1alpha1.WorkloadRunner{}
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
				return updated.Status.ResourcePhase == dromosv1alpha1.WorkloadRunnerPhaseUnknown &&
					updated.Status.Phase == dromosv1alpha1.WorkloadRunnerPhaseUnknown
			}, timeout, interval).Should(BeTrue())
		})

		It("sets Phase to Degraded when resource is Running and plugin reports Failed", func() {
			phasePlugin.SetPhase(dromosv1alpha1.PluginPhaseFailed)

			karta = internaltest.NewJobSetKartaBuilder("karta-degraded").
				WithDefaults().
				WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
					{Type: "StartupPolicyCompleted", Status: ptr.To("True")},
					{Type: "Completed", Status: ptr.To("False")},
					{Type: "Failed", Status: ptr.To("False")},
				}).
				Build()
			karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(karta))

			jobset := test.NewJobSetBuilder("jobset-degraded", testNamespace).WithReplicas(3).Build()
			runner = createWorkloadRunner(managerCtx, k8sClient, "runner-degraded", testNamespace, jobset, nil)

			updateJobSetStatus(managerCtx, k8sClient, "jobset-degraded", testNamespace, []interface{}{
				makeCondition("StartupPolicyCompleted", "True"),
				makeCondition("Completed", "False"),
				makeCondition("Failed", "False"),
			})

			Eventually(func() bool {
				updated := &dromosv1alpha1.WorkloadRunner{}
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
				if err != nil {
					return false
				}
				return updated.Status.ResourcePhase == dromosv1alpha1.WorkloadRunnerPhaseRunning &&
					updated.Status.Phase == dromosv1alpha1.WorkloadRunnerPhaseDegraded
			}, timeout, interval).Should(BeTrue())

			updated := &dromosv1alpha1.WorkloadRunner{}
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)).To(Succeed())
			Expect(updated.Status.Plugins).To(HaveKey(plugins.InjectableStatusPluginName))
			Expect(updated.Status.Plugins[plugins.InjectableStatusPluginName].Phase).To(Equal(dromosv1alpha1.PluginPhaseFailed))
		})

		It("critical controller issue yields Phase to Unknown when resource is Running but Karta becomes invalid", func() {
			phasePlugin.SetPhase(dromosv1alpha1.PluginPhaseDisabled)

			karta = internaltest.NewJobSetKartaBuilder("karta-critical-degraded").
				WithDefaults().
				WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
					{Type: "StartupPolicyCompleted", Status: ptr.To("True")},
					{Type: "Completed", Status: ptr.To("False")},
					{Type: "Failed", Status: ptr.To("False")},
				}).
				Build()
			karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(karta))

			runner = createWorkloadRunner(managerCtx, k8sClient, "runner-critical-degraded", testNamespace, newJobSet("jobset-critical-degraded"), nil)

			updateJobSetStatus(managerCtx, k8sClient, "jobset-critical-degraded", testNamespace, []interface{}{
				makeCondition("StartupPolicyCompleted", "True"),
				makeCondition("Completed", "False"),
				makeCondition("Failed", "False"),
			})

			Eventually(func() bool {
				updated := &dromosv1alpha1.WorkloadRunner{}
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
				return updated.Status.Phase == dromosv1alpha1.WorkloadRunnerPhaseRunning
			}, timeout, interval).Should(BeTrue())

			// make Karta invalid - yield critical controller issue + trigger reconcile
			Eventually(func() error {
				if err := k8sClient.Get(managerCtx, client.ObjectKeyFromObject(karta), karta); err != nil {
					return err
				}
				now := metav1.Now()
				karta.Status.Conditions = []metav1.Condition{
					{
						Type:               string(sdkv1alpha1.CRDExists),
						Status:             metav1.ConditionTrue,
						LastTransitionTime: now,
						Reason:             "CRDExists",
						Message:            "CRD exists",
					},
				}
				return k8sClient.Status().Update(managerCtx, karta)
			}, timeout, interval).Should(Succeed())

			Eventually(func() bool {
				updated := &dromosv1alpha1.WorkloadRunner{}
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated); err != nil {
					return false
				}
				return updated.Status.Phase == dromosv1alpha1.WorkloadRunnerPhaseUnknown
			}, timeout, interval).Should(BeTrue())

			updated := &dromosv1alpha1.WorkloadRunner{}
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)).To(Succeed())
			Expect(updated.Status.Issues).To(HaveLen(1))
			Expect(updated.Status.Issues[0].Description).To(ContainSubstring("Karta is not valid"))
		})

		It("buildResources failure yields Degraded phase with controller issue", func() {
			additionalResourcePlugin.SetError(nil)
			phasePlugin.SetPhase(dromosv1alpha1.PluginPhaseDisabled)

			karta = internaltest.NewJobSetKartaBuilder("karta-build-fail").
				WithDefaults().
				WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
					{Type: "StartupPolicyCompleted", Status: ptr.To("True")},
					{Type: "Completed", Status: ptr.To("False")},
					{Type: "Failed", Status: ptr.To("False")},
				}).
				Build()
			karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(karta))

			runner = createWorkloadRunner(managerCtx, k8sClient, "runner-build-fail", testNamespace, newJobSet("jobset-build-fail"), nil)

			updateJobSetStatus(managerCtx, k8sClient, "jobset-build-fail", testNamespace, []interface{}{
				makeCondition("StartupPolicyCompleted", "True"),
				makeCondition("Completed", "False"),
				makeCondition("Failed", "False"),
			})

			Eventually(func() bool {
				updated := &dromosv1alpha1.WorkloadRunner{}
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)
				return updated.Status.Phase == dromosv1alpha1.WorkloadRunnerPhaseRunning
			}, timeout, interval).Should(BeTrue())

			additionalResourcePlugin.SetError(fmt.Errorf("simulated ingress conflict"))

			Eventually(func() error {
				current := &dromosv1alpha1.WorkloadRunner{}
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), current); err != nil {
					return err
				}
				jobset := newJobSet("jobset-build-fail")
				jobset.Spec.ReplicatedJobs[0].Replicas = int32(2)
				current.Spec.Template = marshalRawExtension(jobset)
				return k8sClient.Update(ctx, current)
			}, timeout, interval).Should(Succeed())

			Eventually(func() bool {
				updated := &dromosv1alpha1.WorkloadRunner{}
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated); err != nil {
					return false
				}
				return updated.Status.Phase == dromosv1alpha1.WorkloadRunnerPhaseDegraded
			}, timeout, interval).Should(BeTrue())

			updated := &dromosv1alpha1.WorkloadRunner{}
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), updated)).To(Succeed())
			Expect(updated.Status.Issues).To(HaveLen(1))
			Expect(updated.Status.Issues[0].Group).To(Equal("Internal"))
			Expect(updated.Status.Issues[0].Description).To(ContainSubstring("failed to build resources"))

			additionalResourcePlugin.SetError(nil)
		})
	})

	Context("updating workload replicas", func() {
		var (
			karta        *kartav1alpha1.Karta
			runner       *dromosv1alpha1.WorkloadRunner
			runnerName   = "jobset-update-runner"
			workloadName = "test-job-set"
		)

		BeforeEach(func() {
			karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			runner = createWorkloadRunner(managerCtx, k8sClient, runnerName, testNamespace, newJobSet(workloadName), nil)
		})

		AfterEach(func() {
			deleteAndPollUntilDeleted(managerCtx, k8sClient, runner)
			deleteAndPollUntilDeleted(managerCtx, k8sClient, karta)
			deleteAllJobSets(managerCtx, k8sClient)
		})

		Context("plugin with mutate pod spec functionality - add env var", Ordered, func() {
			BeforeAll(func() {
				runnerName = "env-var-runner"
				workloadName = "env-var-jobset"
			})

			AfterAll(func() {
				deleteAndPollUntilDeleted(managerCtx, k8sClient, runner)
				deleteAndPollUntilDeleted(managerCtx, k8sClient, karta)
				deleteAllJobSets(managerCtx, k8sClient)
			})

			It("creates JobSet with env var in containers", func() {
				karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
				runner = createWorkloadRunner(managerCtx, k8sClient, runnerName, testNamespace, newJobSet(workloadName), nil)

				jobset := &jobsetv1alpha2.JobSet{}
				Eventually(func() error {
					return k8sClient.Get(managerCtx, types.NamespacedName{
						Name:      workloadName,
						Namespace: testNamespace,
					}, jobset)
				}, timeout, interval).Should(Succeed())

				Expect(jobset.Spec.ReplicatedJobs).NotTo(BeEmpty())
				for _, replicatedJob := range jobset.Spec.ReplicatedJobs {
					for _, container := range replicatedJob.Template.Spec.Template.Spec.Containers {
						Expect(container.Env).To(ContainElement(corev1.EnvVar{
							Name:  plugins.EnvVarName,
							Value: runnerName,
						}))
					}
				}
			})
		})

		Context("plugin with create additional resources - configmap", Ordered, func() {
			BeforeAll(func() {
				runnerName = "configmap-runner"
				workloadName = "configmap-jobset"
			})

			AfterAll(func() {
				deleteAndPollUntilDeleted(managerCtx, k8sClient, runner)
				deleteAndPollUntilDeleted(managerCtx, k8sClient, karta)
				deleteAllJobSets(managerCtx, k8sClient)
			})

			It("creates JobSet and ConfigMap", func() {
				karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
				runner = createWorkloadRunner(managerCtx, k8sClient, runnerName, testNamespace, newJobSet(workloadName), nil)

				jobset := &jobsetv1alpha2.JobSet{}
				Eventually(func() error {
					return k8sClient.Get(managerCtx, types.NamespacedName{
						Name:      workloadName,
						Namespace: testNamespace,
					}, jobset)
				}, timeout, interval).Should(Succeed())

				// Plugin CreateConfigMap should create a ConfigMap with owner reference to the runner.
				configMap := &corev1.ConfigMap{}
				Eventually(func() error {
					return k8sClient.Get(managerCtx, types.NamespacedName{
						Name:      runnerName + plugins.ConfigMapNameSuffix,
						Namespace: testNamespace,
					}, configMap)
				}, timeout, interval).Should(Succeed())

				Expect(configMap.Data).To(HaveKeyWithValue("runner-name", runnerName))
				Expect(configMap.Data).To(HaveKeyWithValue("attached-resource-name", workloadName))
				Expect(configMap.Data).To(HaveKeyWithValue("attached-resource-kind", "JobSet"))

				Expect(configMap.OwnerReferences).To(HaveLen(1))
				Expect(configMap.OwnerReferences[0].Name).To(Equal(runnerName))
				Expect(configMap.OwnerReferences[0].Kind).To(Equal("WorkloadRunner"))
			})

			It("reconcile after ConfigMap changed - and patch it", func() {
				karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
				runner = createWorkloadRunner(managerCtx, k8sClient, runnerName, testNamespace, newJobSet(workloadName), nil)

				jobset := &jobsetv1alpha2.JobSet{}
				Eventually(func() error {
					return k8sClient.Get(managerCtx, types.NamespacedName{
						Name:      workloadName,
						Namespace: testNamespace,
					}, jobset)
				}, timeout, interval).Should(Succeed())

				configMap := &corev1.ConfigMap{}
				Eventually(func() error {
					return k8sClient.Get(managerCtx, types.NamespacedName{
						Name:      runnerName + plugins.ConfigMapNameSuffix,
						Namespace: testNamespace,
					}, configMap)
				}, timeout, interval).Should(Succeed())

				Expect(configMap.Data).To(HaveKeyWithValue("runner-name", runnerName))

				Eventually(func() error {
					updated := configMap.DeepCopy()
					updated.Data["runner-name"] = "updated-runner-name"
					return k8sClient.Update(managerCtx, updated)
				}, timeout, interval).Should(Succeed())

				Eventually(func() bool {
					err := k8sClient.Get(managerCtx, types.NamespacedName{
						Name:      runnerName + plugins.ConfigMapNameSuffix,
						Namespace: testNamespace,
					}, configMap)
					if errors.IsNotFound(err) {
						return false
					}
					Expect(err).ToNot(HaveOccurred())
					return configMap.Data["runner-name"] == runnerName
				}, timeout, interval).Should(BeTrue())
			})
		})

		Context("workload runner update - replicas, annotations and labels", Ordered, func() {
			var (
				originJobSet *jobsetv1alpha2.JobSet
				updateJobSet *jobsetv1alpha2.JobSet
			)

			BeforeAll(func() {
				originJobSet = &jobsetv1alpha2.JobSet{}
				updateJobSet = &jobsetv1alpha2.JobSet{}

				runnerName = "lifecycle-runner"
				workloadName = "lifecycle-jobset"
				karta = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			})

			AfterEach(func() {
				updateJobSet.DeepCopyInto(originJobSet)
				updateJobSet = &jobsetv1alpha2.JobSet{}
			})

			AfterAll(func() {
				deleteAndPollUntilDeleted(managerCtx, k8sClient, runner)
				deleteAndPollUntilDeleted(managerCtx, k8sClient, karta)
				deleteAllJobSets(managerCtx, k8sClient)
			})

			It("creates workload runner with 1 replica", func() {
				originJobSet = newJobSet(workloadName)
				originJobSet.Spec.ReplicatedJobs[0].Replicas = int32(1)
				runner = createWorkloadRunner(managerCtx, k8sClient, runnerName, testNamespace, originJobSet, nil)

				Eventually(func() error {
					return k8sClient.Get(managerCtx, types.NamespacedName{
						Name:      workloadName,
						Namespace: testNamespace,
					}, updateJobSet)
				}, timeout, interval).Should(Succeed())

				Expect(updateJobSet.Spec.ReplicatedJobs).To(HaveLen(1))
				Expect(updateJobSet.Spec.ReplicatedJobs[0].Replicas).To(Equal(int32(1)))
			})

			It("updates workload runner spec to 2 replicas", func() {
				jobset := newJobSet(workloadName)
				jobset.Spec.ReplicatedJobs[0].Replicas = int32(2)
				Eventually(func() error { return updateAttachedResource(k8sClient, runnerName, jobset) }, timeout, interval).Should(Succeed())
				Eventually(func() bool { return getJobSetWithReplicaCheck(k8sClient, workloadName, updateJobSet, 2) }, timeout, interval).Should(BeTrue())

				Expect(updateJobSet.Spec.ReplicatedJobs[0].Replicas).To(Equal(int32(2)))

				patches, err := UpdatePatches(originJobSet, updateJobSet)
				Expect(err).ToNot(HaveOccurred())
				Expect(patches).To(ContainElements(jsonpatch.Operation{
					Operation: "replace",
					Path:      "/spec/replicatedJobs/0/replicas",
					Value:     json.Number("2"),
				}))
			})

			It("adds annotation and label to workload attached resource", func() {
				jobset := newJobSet(workloadName)
				jobset.Spec.ReplicatedJobs[0].Replicas = int32(2)
				jobset.Annotations = map[string]string{"test-annotation": "test-value1"}
				jobset.Labels = map[string]string{"test-label": "test-value2"}
				Eventually(func() error { return updateAttachedResource(k8sClient, runnerName, jobset) }, timeout, interval).Should(Succeed())

				Eventually(func() bool {
					err := k8sClient.Get(ctx, types.NamespacedName{
						Name:      workloadName,
						Namespace: testNamespace,
					}, updateJobSet)
					if err != nil {
						return false
					}
					return updateJobSet.Annotations["test-annotation"] == "test-value1" && updateJobSet.Labels["test-label"] == "test-value2"
				}, timeout, interval).Should(BeTrue())

				patches, err := UpdatePatches(originJobSet, updateJobSet)
				Expect(err).ToNot(HaveOccurred())
				Expect(patches).To(ContainElements(
					jsonpatch.Operation{
						Operation: "add",
						Path:      "/metadata/annotations",
						Value:     map[string]interface{}{"test-annotation": "test-value1"},
					},
					jsonpatch.Operation{
						Operation: "add",
						Path:      "/metadata/labels",
						Value:     map[string]interface{}{"test-label": "test-value2"},
					}))
			})

			It("updates workload runner annotation and label to different values", func() {
				jobset := newJobSet(workloadName)
				jobset.Spec.ReplicatedJobs[0].Replicas = int32(2)
				jobset.Annotations = map[string]string{"test-annotation": "test-value3"}
				jobset.Labels = map[string]string{"test-label": "test-value4"}
				Eventually(func() error { return updateAttachedResource(k8sClient, runnerName, jobset) }, timeout, interval).Should(Succeed())

				Eventually(func() bool {
					err := k8sClient.Get(ctx, types.NamespacedName{
						Name:      workloadName,
						Namespace: testNamespace,
					}, updateJobSet)
					if err != nil {
						return false
					}
					return updateJobSet.Annotations["test-annotation"] == "test-value3" && updateJobSet.Labels["test-label"] == "test-value4"
				}, timeout, interval).Should(BeTrue())

				patches, err := UpdatePatches(originJobSet, updateJobSet)
				Expect(err).ToNot(HaveOccurred())
				Expect(patches).To(ContainElements(
					jsonpatch.Operation{
						Operation: "replace",
						Path:      "/metadata/annotations/test-annotation",
						Value:     "test-value3",
					},
					jsonpatch.Operation{
						Operation: "replace",
						Path:      "/metadata/labels/test-label",
						Value:     "test-value4",
					}))
			})

			It("deletes workload runner annotation and label", func() {
				jobset := newJobSet(workloadName)
				jobset.Annotations = make(map[string]string)
				jobset.Labels = make(map[string]string)
				jobset.Spec.ReplicatedJobs[0].Replicas = int32(2)
				Eventually(func() error { return updateAttachedResource(k8sClient, runnerName, jobset) }, timeout, interval).Should(Succeed())

				Eventually(func() bool {
					err := k8sClient.Get(ctx, types.NamespacedName{
						Name:      workloadName,
						Namespace: testNamespace,
					}, updateJobSet)
					if err != nil {
						return false
					}
					_, hasAnnotation := updateJobSet.Annotations["test-annotation"]
					_, hasLabel := updateJobSet.Labels["test-label"]
					return !hasAnnotation && !hasLabel
				}, timeout, interval).Should(BeTrue())
				patches, err := UpdatePatches(originJobSet, updateJobSet)
				Expect(err).ToNot(HaveOccurred())
				Expect(patches).To(ContainElements(
					jsonpatch.Operation{
						Operation: "remove",
						Path:      "/metadata/annotations",
						Value:     nil,
					},
					jsonpatch.Operation{
						Operation: "remove",
						Path:      "/metadata/labels",
						Value:     nil,
					}))
			})
		})
	})
})

func UpdatePatches(origin, updated client.Object) ([]jsonpatch.Operation, error) {
	originRaw, _ := json.Marshal(origin)
	updatedRaw, _ := json.Marshal(updated)
	return jsonpatch.CreatePatch(originRaw, updatedRaw)
}

func updateResourceStatus[T client.Object](ctx context.Context, k8sClient client.Client, obj T) {
	Eventually(func() error {
		err := k8sClient.Status().Update(ctx, obj)
		return err
	}, timeout, interval).Should(Succeed())
}

func createKartaWithStatus(ctx context.Context, k8sClient client.Client, karta *kartav1alpha1.Karta) *kartav1alpha1.Karta {
	kartaCopy := karta.DeepCopy()
	karta = createResource(ctx, k8sClient, karta)
	karta.Status = kartaCopy.Status
	updateResourceStatus(ctx, k8sClient, karta)
	Eventually(func() bool {
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      karta.Name,
			Namespace: karta.Namespace,
		}, karta)
		return err == nil && karta.Status.Conditions != nil
	}, timeout, interval).Should(BeTrue())
	return karta
}

func newKarta(karta *kartav1alpha1.Karta, ready bool, crdExists bool) *kartav1alpha1.Karta {
	now := metav1.Now()

	readyStatus := metav1.ConditionFalse
	if ready && crdExists {
		readyStatus = metav1.ConditionTrue
	}

	conditions := []metav1.Condition{
		{
			Type:               string(sdkv1alpha1.Ready),
			Status:             readyStatus,
			LastTransitionTime: now,
			Reason:             "Ready",
			Message:            "Karta is ready",
		},
	}

	karta = karta.DeepCopy()
	karta.Status.Conditions = conditions
	return karta
}

func newKartaWithValidStatus(karta *kartav1alpha1.Karta) *kartav1alpha1.Karta {
	return newKarta(karta, true, true)
}

func newJobSet(name string) *jobsetv1alpha2.JobSet {
	parallelism := int32(1)
	completions := int32(1)
	backoffLimit := int32(0)
	replicas := int32(3)

	js := &jobsetv1alpha2.JobSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: jobsetGroup + "/" + jobsetVersion,
			Kind:       jobsetKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
		Spec: jobsetv1alpha2.JobSetSpec{
			ReplicatedJobs: []jobsetv1alpha2.ReplicatedJob{
				{
					Name:     "workers",
					Replicas: replicas,
					Template: batchv1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Parallelism:  &parallelism,
							Completions:  &completions,
							BackoffLimit: &backoffLimit,
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									RestartPolicy: corev1.RestartPolicyNever,
									Containers: []corev1.Container{
										{
											Name:  "sleep",
											Image: "busybox",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return js
}

func getJobSetWithReplicaCheck(k8sClient client.Reader, workloadName string, jobsetObject *jobsetv1alpha2.JobSet, replicas int32) bool {
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      workloadName,
		Namespace: testNamespace,
	}, jobsetObject)
	if err != nil {
		if errors.IsNotFound(err) {
			return false
		}
		Expect(err).ToNot(HaveOccurred())
	}

	Expect(err).ToNot(HaveOccurred())
	if err != nil {
		return false
	}

	return len(jobsetObject.Spec.ReplicatedJobs) > 0 && jobsetObject.Spec.ReplicatedJobs[0].Replicas == replicas
}

func updateAttachedResource(k8sClient client.Client, runnerName string, jobObject *jobsetv1alpha2.JobSet) error {
	current := &dromosv1alpha1.WorkloadRunner{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      runnerName,
		Namespace: testNamespace,
	}, current)
	if err != nil {
		return err
	}

	current.Spec.Template = marshalRawExtension(jobObject)
	return k8sClient.Update(ctx, current)
}
