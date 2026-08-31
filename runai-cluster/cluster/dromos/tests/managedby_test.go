/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/plugins/storage"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/test"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// These specs verify the managedBy contract
var _ = Describe("ManagedBy yield", Serial, Ordered, func() {
	const otherController = "other-controller"
	const consistentlyDuration = 2 * time.Second

	var (
		k8sManager              ctrl.Manager
		k8sClient               client.Client
		managerCtx              context.Context
		managerCancel           context.CancelFunc
		managedByNamespace      = "managedby-test"
		managedByControllerName = "managedby-test-controller"
	)

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)

		absCertDir, err := filepath.Abs(webhookCertDir)
		Expect(err).NotTo(HaveOccurred())

		k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
			Metrics: metricsserver.Options{
				BindAddress: "0",
			},
			WebhookServer: webhook.NewServer(webhook.Options{
				CertDir: absCertDir,
				Port:    webhookPort,
			}),
		})
		Expect(err).NotTo(HaveOccurred())

		k8sClient = k8sManager.GetClient()

		registry := framework.NewRegistry(storage.NewPlugin(false))
		fw := framework.NewFramework(registry, nil)
		managedByController := controller.NewWorkloadRunnerController(k8sClient, fw,
			k8sManager.GetEventRecorderFor(managedByControllerName)).WithName(managedByControllerName)
		Expect(managedByController.SetupWithManager(managerCtx, k8sManager)).To(Succeed())

		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		startManager(k8sManager, managerCtx, managerCancel)

		ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: managedByNamespace}}
		_ = k8sClient.Create(managerCtx, ns)

		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)
	})

	AfterAll(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		deleteAllForType(managerCtx, k8sClient, &v1.PersistentVolumeClaimList{})
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)
	})

	AfterEach(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		deleteAllForType(managerCtx, k8sClient, &v1.PersistentVolumeClaimList{})
	})

	setManagedBy := func(runner *dromosv1alpha1.WorkloadRunner, value string) {
		Eventually(func() error {
			fresh := &dromosv1alpha1.WorkloadRunner{}
			if err := k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), fresh); err != nil {
				return err
			}
			fresh.Spec.ManagedBy = value
			return k8sClient.Update(managerCtx, fresh)
		}, timeout, interval).Should(Succeed())
	}

	It("produces no side effects and no status for a runner born managed", func() {
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))

		pvcName := "pvc-" + rand.String(4)
		runner := newStorageRunner(managedByNamespace, nil, newDefaultPvc(pvcName))
		runner.Spec.ManagedBy = otherController
		runner = createResource(managerCtx, k8sClient, runner)

		Consistently(func(g Gomega) {
			pvc := &v1.PersistentVolumeClaim{}
			err := k8sClient.Get(managerCtx, types.NamespacedName{Name: pvcName, Namespace: managedByNamespace}, pvc)
			g.Expect(apierrors.IsNotFound(err)).To(BeTrue(), "PVC must not be created for a managed runner")

			fresh := &dromosv1alpha1.WorkloadRunner{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), fresh)).To(Succeed())
			g.Expect(fresh.Status.Phase).To(BeEmpty())
			g.Expect(fresh.Status.Plugins).To(BeEmpty())
			g.Expect(fresh.Status.Conditions).To(BeEmpty())

			// Read through the API reader, not the cached client
			events := &v1.EventList{}
			g.Expect(k8sManager.GetAPIReader().List(managerCtx, events,
				client.InNamespace(managedByNamespace))).To(Succeed())
			var runnerEvents []string
			for _, event := range events.Items {
				if event.InvolvedObject.UID == runner.UID {
					runnerEvents = append(runnerEvents, event.Reason+": "+event.Message)
				}
			}
			g.Expect(runnerEvents).To(BeEmpty(), "no events must be emitted for a managed runner")
		}, consistentlyDuration, interval).Should(Succeed())
	})

	It("freezes side effects on takeover and resumes on handback", func() {
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))

		pvcName := "pvc-" + rand.String(4)
		runner := createResource(managerCtx, k8sClient, newStorageRunner(managedByNamespace, nil, newDefaultPvc(pvcName)))

		pvc := &v1.PersistentVolumeClaim{}
		Eventually(func() error {
			return k8sClient.Get(managerCtx, types.NamespacedName{Name: pvcName, Namespace: managedByNamespace}, pvc)
		}, timeout, interval).Should(Succeed())

		// Change owner
		setManagedBy(runner, otherController)

		// Create side affect
		Expect(k8sClient.Delete(managerCtx, pvc)).To(Succeed())

		// strip finalizers to let the delete complete
		Eventually(func(g Gomega) {
			fresh := &v1.PersistentVolumeClaim{}
			err := k8sClient.Get(managerCtx, types.NamespacedName{Name: pvcName, Namespace: managedByNamespace}, fresh)
			if apierrors.IsNotFound(err) {
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
			fresh.Finalizers = nil
			g.Expect(k8sClient.Update(managerCtx, fresh)).To(Succeed())
			g.Expect(apierrors.IsNotFound(
				k8sClient.Get(managerCtx, types.NamespacedName{Name: pvcName, Namespace: managedByNamespace}, fresh),
			)).To(BeTrue())
		}, timeout, interval).Should(Succeed())

		pvcKey := types.NamespacedName{Name: pvcName, Namespace: managedByNamespace}
		Consistently(func(g Gomega) {
			err := k8sClient.Get(managerCtx, pvcKey, &v1.PersistentVolumeClaim{})
			g.Expect(apierrors.IsNotFound(err)).To(BeTrue(), "PVC must not be recreated while managed")
		}, consistentlyDuration, interval).Should(Succeed())

		// Handback: clearing managedBy
		setManagedBy(runner, "")

		// PVC should exist after reconcile
		Eventually(func() error {
			return k8sClient.Get(managerCtx, pvcKey, &v1.PersistentVolumeClaim{})
		}, timeout, interval).Should(Succeed())
	})
})
