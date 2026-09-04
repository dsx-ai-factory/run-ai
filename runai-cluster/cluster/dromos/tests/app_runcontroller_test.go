/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/app"
	dromosconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Exercises app.RunController's consumer extension points for registering additional reconcilers and webhook paths
var _ = Describe("RunController manager setup", Serial, Ordered, func() {
	var (
		managerCtx        context.Context
		managerCancel     context.CancelFunc
		runControllerDone chan error
		k8sClient         client.Client
		reconciledNames   sync.Map
		setupWebhookURL   string
		setupNamespace    = "runai-app-setup-test"
		setupWebhookPath  = "/setup-registered-path"
	)

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)
		setupWebhookURL = fmt.Sprintf("https://127.0.0.1:%d%s", webhookPort, setupWebhookPath)

		absCertDir, err := filepath.Abs(webhookCertDir)
		Expect(err).NotTo(HaveOccurred())

		testConfig := &dromosconfig.DromosConfig{
			WebhookPort:       webhookPort,
			CertDir:           absCertDir,
			MetricsPort:       0,
			MaxMutationRounds: 5,
		}

		setup := func(_ context.Context, mgr ctrl.Manager) error {
			// add extra webhook path
			mgr.GetWebhookServer().Register(setupWebhookPath,
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

			// add extra reconciler
			return ctrl.NewControllerManagedBy(mgr).
				Named("setup-noop-reconciler").
				For(&dromosv1alpha1.WorkloadRunner{}).
				Complete(reconcile.Func(func(_ context.Context, req reconcile.Request) (reconcile.Result, error) {
					reconciledNames.Store(req.Name, true)
					return reconcile.Result{}, nil
				}))
		}

		runControllerDone = make(chan error, 1)
		go func() {
			defer GinkgoRecover()
			runControllerDone <- app.RunController(managerCtx,
				app.WithRestConfig(cfg),
				app.WithConfig(testConfig),
				app.WithManagerSetup(setup),
			)
		}()

		k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).NotTo(HaveOccurred())

		ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: setupNamespace}}
		_ = k8sClient.Create(managerCtx, ns)

		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)
	})

	AfterAll(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		managerCancel()
		Eventually(runControllerDone, timeout, interval).Should(Receive(BeNil()))
		waitForWebhookToBeClosed(webhookUrlValidate)
	})

	It("serves a webhook path registered by a manager setup function", func() {
		httpClient := &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}
		Eventually(func() (int, error) {
			req, err := http.NewRequestWithContext(managerCtx, http.MethodGet, setupWebhookURL, nil)
			if err != nil {
				return 0, err
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				return 0, err
			}
			defer func() { _ = resp.Body.Close() }()
			return resp.StatusCode, nil
		}, timeout, interval).Should(Equal(http.StatusOK))
	})

	It("delivers events to a setup-registered reconciler via the injected rest config", func() {
		runnerName := "setup-test-runner-" + rand.String(4)
		pod := newTestPod(runnerName+"-pod", setupNamespace)
		createWorkloadRunner(managerCtx, k8sClient, runnerName, setupNamespace, pod, nil)

		Eventually(func() bool {
			_, ok := reconciledNames.Load(runnerName)
			return ok
		}, timeout, interval).Should(BeTrue())
	})

	It("keeps the built-in controller and webhook wiring intact", func() {
		runnerName := "setup-test-builtin-" + rand.String(4)
		pod := newTestPod(runnerName+"-pod", setupNamespace)
		// Simple pass, admission still works with setup functions present
		createWorkloadRunner(managerCtx, k8sClient, runnerName, setupNamespace, pod, nil)

		// No Karta exists, the built-in WorkloadRunner reports Unknown
		updated := &dromosv1alpha1.WorkloadRunner{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{
				Name:      runnerName,
				Namespace: setupNamespace,
			}, updated)).To(Succeed())
			g.Expect(updated.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseUnknown))
		}, timeout, interval).Should(Succeed())
	})
})
