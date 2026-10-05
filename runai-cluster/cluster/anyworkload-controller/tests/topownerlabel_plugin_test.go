/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"path/filepath"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/plugins/topownerlabel"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	dromoscontroller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	kartautils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/karta"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var _ = Describe("Top-owner label plugin upgrade safety", Serial, Ordered, func() {
	var (
		k8sManager    ctrl.Manager
		k8sClient     client.Client
		managerCtx    context.Context
		managerCancel context.CancelFunc
		namespace     = "runai-top-owner-label-test"
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

		plugin := topownerlabel.NewPlugin(k8sClient)
		fw := framework.NewFramework(framework.NewRegistry(plugin), nil)
		//nolint:staticcheck // SA1019: matches the other integration tests' recorder wiring
		controller := dromoscontroller.NewWorkloadRunnerController(
			k8sClient,
			fw,
			k8sManager.GetEventRecorderFor("top-owner-label-test-controller"),
		).WithName("top-owner-label-test-controller")
		Expect(controller.SetupWithManager(managerCtx, k8sManager)).To(Succeed())
		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(managerCtx)).To(Succeed())
		}()

		Expect(k8sClient.Create(managerCtx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		})).To(Succeed())
		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)
	})

	AfterAll(func() {
		deleteAllForType(managerCtx, k8sClient, &dromosv1alpha1.WorkloadRunnerList{})
		deleteAllForType(managerCtx, k8sClient, &batchv1.JobList{})
		deleteAllForType(managerCtx, k8sClient, &kartav1alpha1.KartaList{})
		_ = k8sClient.Delete(managerCtx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}})
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)
	})

	It("reconciles an existing unlabeled Job without changing its immutable template", func() {
		karta := &kartav1alpha1.Karta{
			ObjectMeta: metav1.ObjectMeta{
				Name: "krt-v1-batch-job-upgrade-test",
				Labels: map[string]string{
					kartautils.KartaGroupLabel:   "batch",
					kartautils.KartaVersionLabel: "v1",
					kartautils.KartaKindLabel:    "Job",
				},
			},
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
		Expect(k8sClient.Create(managerCtx, karta)).To(Succeed())
		Eventually(func(g Gomega) {
			current := &kartav1alpha1.Karta{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(karta), current)).To(Succeed())
			meta.SetStatusCondition(&current.Status.Conditions, metav1.Condition{
				Type: "Ready", Status: metav1.ConditionTrue, Reason: "Ready",
			})
			g.Expect(k8sClient.Status().Update(managerCtx, current)).To(Succeed())
		}, timeout, interval).Should(Succeed())

		job := &batchv1.Job{
			TypeMeta: metav1.TypeMeta{APIVersion: "batch/v1", Kind: "Job"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "existing-unlabeled-job",
				Namespace: namespace,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers:    []corev1.Container{{Name: "main", Image: "busybox"}},
						RestartPolicy: corev1.RestartPolicyNever,
					},
				},
			},
		}
		Expect(k8sClient.Create(managerCtx, job)).To(Succeed())
		Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(job), job)).To(Succeed())

		template := job.DeepCopy()
		template.ResourceVersion = ""
		template.UID = ""
		template.CreationTimestamp = metav1.Time{}
		template.ManagedFields = nil
		template.Status = batchv1.JobStatus{}
		workloadID := "existing-workload-id"
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      job.Name,
				Namespace: namespace,
				Annotations: map[string]string{
					commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: workloadID,
				},
			},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{Template: marshalRawExtension(template)},
		}
		Expect(k8sClient.Create(managerCtx, runner)).To(Succeed())

		Eventually(func(g Gomega) {
			current := &batchv1.Job{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(job), current)).To(Succeed())
			g.Expect(current.Labels).To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, workloadID))
			g.Expect(current.Spec.Template.Labels).NotTo(HaveKey(commonconstants.TopOwnerUidLabel))
		}, timeout, interval).Should(Succeed())
	})
})
