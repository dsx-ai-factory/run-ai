/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/deployers/helm_deployer"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/test"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// these tests are very similar to the controller tests, but they use the real workers pool
// to allow running complete operation end-to-end unlike the controller test that runs up
// to the point that a task is queued for execution.
var _ = Describe("helm controller e2e test", Ordered, func() {
	var (
		ctx            context.Context
		cancel         context.CancelFunc
		k8sCli         client.Client
		store          *storage.Storage
		testController *ApplicationController
	)

	appKey := types.NamespacedName{Name: "e2e-app", Namespace: "ns1"}

	BeforeEach(func() {
		// we need to create contexts that will allow us to control the workers pool
		ctx, cancel = context.WithCancel(context.Background())

		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(runv1alpha1.AddToScheme(scheme)).To(Succeed())

		k8sCli = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&runv1alpha1.Application{}).
			Build()

		Expect(k8sCli.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: appKey.Namespace}})).To(Succeed())

		store = test.NewTestHelmStore()

		testController = NewApplicationControllerForTest(k8sCli,
			func(id types.NamespacedName, downloaderFactory helm_deployer.DownloaderFactory) (*helm_deployer.ReleaseClient, error) {
				return test.NewTestReleaseClient(id, store, downloaderFactory), nil
			},
			helm_deployer.NewDownloaderForTest,
		)

		// Start workers so jobs will run and not only queued
		for i := 0; i < helmWorkerCount; i++ {
			testController.helmWorkers.waitGroup.Add(1)
			go testController.helmWorker(ctx, testController.helmWorkers.jobsChannel)
		}

		// Close the channel when the context is done. Capture channel and workers so we close
		// the correct instance after each BeforeEach overwrites testController (avoids "close of closed channel").
		jobsCh := testController.helmWorkers.jobsChannel
		hw := &testController.helmWorkers
		go func() {
			<-ctx.Done()
			close(jobsCh)
			hw.waitGroup.Wait()
		}()

		DeferCleanup(cancel)
	})

	Context("run install", func() {
		It("runs install task to completion and updates status", func() {
			abs, err := filepath.Abs("../../test/data/chart-a")
			Expect(err).NotTo(HaveOccurred())
			chartURL := "file://" + abs

			app := &runv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:        appKey.Name,
					Namespace:   appKey.Namespace,
					Annotations: map[string]string{AnnotationRunaiApp: "true"},
				},
				Spec: runv1alpha1.ApplicationSpec{
					Helm: &runv1alpha1.HelmSpec{ChartUrl: chartURL},
				},
			}
			Expect(k8sCli.Create(ctx, app)).To(Succeed())

			// First reconcile: install task queued and picked by worker.
			result, err := testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Duration(0)))

			// Wait until a worker finishes (task no longer in runningTasks).
			Eventually(func() bool {
				alive, _ := testController.helmWorkers.IsHelmTaskAlive(appKey)
				return !alive
			}, 30*time.Second, 200*time.Millisecond).Should(BeTrue())

			// Reconcile until status reflects completion (avoids race where a single Reconcile
			// runs before the worker's status update is visible or triggers a second install).
			Eventually(func() bool {
				_, _ = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey})
				app := &runv1alpha1.Application{}
				if err := k8sCli.Get(ctx, appKey, app); err != nil {
					return false
				}
				return app.Status.ActionStatus != nil &&
					app.Status.ActionStatus.Phase == helm_deployer.ActionCompleted &&
					strings.Contains(app.Status.ActionStatus.Message, "install completed")
			}, 15*time.Second, 200*time.Millisecond).Should(BeTrue())

			app = &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, appKey, app)).To(Succeed())
			Expect(app.Status.HelmStatus).NotTo(BeNil())
			Expect(app.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionInstall)))
			Expect(app.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionCompleted))
			Expect(app.Status.ActionStatus.StartedAt.IsZero()).To(BeFalse())
			Expect(app.Status.ActionStatus.CompletedAt.IsZero()).To(BeFalse())
			Expect(app.Status.ActionStatus.Message).To(ContainSubstring("install completed"))
			Expect(app.Status.ActionStatus.Error).To(Equal(""))
			Expect(app.Status.HelmStatus.Phase).To(Equal(string(release.StatusDeployed)))
			Expect(app.Status.HelmStatus.ChartUrl).To(Equal(chartURL))
		})
	})

	Context("install task fails", func() {
		It("reports TaskFailed status when chart cannot be located", func() {
			// Use a file:// URL that does not exist so the test downloader fails in prepareToInstall.
			invalidChartURL := "file:///nonexistent-chart-path-for-e2e-test"

			app := &runv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:        appKey.Name,
					Namespace:   appKey.Namespace,
					Finalizers:  []string{HelmApplicationFinalizer},
					Annotations: map[string]string{AnnotationRunaiApp: "true"},
				},
				Spec: runv1alpha1.ApplicationSpec{
					Helm: &runv1alpha1.HelmSpec{ChartUrl: invalidChartURL},
				},
			}
			Expect(k8sCli.Create(ctx, app)).To(Succeed())

			_, err := testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				alive, _ := testController.helmWorkers.IsHelmTaskAlive(appKey)
				return !alive
			}, 30*time.Second, 200*time.Millisecond).Should(BeTrue())

			// Reconcile until status reflects failure (avoids race with status update visibility).
			Eventually(func() bool {
				_, _ = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey})
				app := &runv1alpha1.Application{}
				if err := k8sCli.Get(ctx, appKey, app); err != nil {
					return false
				}
				return app.Status.ActionStatus != nil &&
					app.Status.ActionStatus.Phase == helm_deployer.ActionFailed
			}, 15*time.Second, 200*time.Millisecond).Should(BeTrue())

			app = &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, appKey, app)).To(Succeed())
			Expect(app.Status.ActionStatus).NotTo(BeNil())
			Expect(app.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionInstall)))
			Expect(app.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionFailed))
			Expect(app.Status.ActionStatus.Message).To(Equal("install failed"))
			Expect(app.Status.ActionStatus.Error).NotTo(BeEmpty())
			Expect(app.Status.ActionStatus.Error).To(ContainSubstring("nonexistent"))
		})
	})

	Context("full app cycle: install, upgrade, uninstall", func() {
		It("runs install, then upgrade, then uninstall to completion", func() {
			chart1Abs, err := filepath.Abs("../../test/data/chart-a")
			Expect(err).NotTo(HaveOccurred())
			chart1URL := "file://" + chart1Abs

			chart2Abs, err := filepath.Abs("../../test/data/chart-b")
			Expect(err).NotTo(HaveOccurred())
			chart2URL := "file://" + chart2Abs

			app := &runv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:        appKey.Name,
					Namespace:   appKey.Namespace,
					Annotations: map[string]string{AnnotationRunaiApp: "true"},
				},
				Spec: runv1alpha1.ApplicationSpec{
					Helm: &runv1alpha1.HelmSpec{ChartUrl: chart1URL},
				},
			}
			Expect(k8sCli.Create(ctx, app)).To(Succeed())

			// --- Install ---
			_, err = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				alive, _ := testController.helmWorkers.IsHelmTaskAlive(appKey)
				return !alive
			}, 30*time.Second, 200*time.Millisecond).Should(BeTrue())

			Eventually(func() bool {
				_, _ = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey})
				app := &runv1alpha1.Application{}
				if err := k8sCli.Get(ctx, appKey, app); err != nil {
					return false
				}
				return app.Status.ActionStatus != nil &&
					app.Status.ActionStatus.Phase == helm_deployer.ActionCompleted &&
					strings.Contains(app.Status.ActionStatus.Message, "install completed")
			}, 15*time.Second, 200*time.Millisecond).Should(BeTrue())

			app = &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, appKey, app)).To(Succeed())
			Expect(app.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionInstall)))
			Expect(app.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionCompleted))
			Expect(app.Status.ActionStatus.StartedAt.IsZero()).To(BeFalse())
			Expect(app.Status.ActionStatus.CompletedAt.IsZero()).To(BeFalse())
			Expect(app.Status.ActionStatus.Message).To(ContainSubstring("install completed"))
			Expect(app.Status.ActionStatus.Error).To(Equal(""))
			Expect(app.Status.HelmStatus.Phase).To(Equal(string(release.StatusDeployed)))
			Expect(app.Status.HelmStatus.ChartUrl).To(Equal(chart1URL))

			// --- Upgrade ---
			app.Spec.Helm.ChartUrl = chart2URL
			Expect(k8sCli.Update(ctx, app)).To(Succeed())

			_, err = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				alive, _ := testController.helmWorkers.IsHelmTaskAlive(appKey)
				return !alive
			}, 30*time.Second, 200*time.Millisecond).Should(BeTrue())

			Eventually(func() bool {
				_, _ = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey})
				app := &runv1alpha1.Application{}
				if err := k8sCli.Get(ctx, appKey, app); err != nil {
					return false
				}
				return app.Status.ActionStatus != nil &&
					app.Status.ActionStatus.Phase == helm_deployer.ActionCompleted &&
					strings.Contains(app.Status.ActionStatus.Message, "upgrade completed")
			}, 15*time.Second, 200*time.Millisecond).Should(BeTrue())

			app = &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, appKey, app)).To(Succeed())
			Expect(app.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionUpgrade)))
			Expect(app.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionCompleted))
			Expect(app.Status.ActionStatus.StartedAt.IsZero()).To(BeFalse())
			Expect(app.Status.ActionStatus.CompletedAt.IsZero()).To(BeFalse())
			Expect(app.Status.ActionStatus.Message).To(ContainSubstring("upgrade completed"))
			Expect(app.Status.ActionStatus.Error).To(Equal(""))
			Expect(app.Status.HelmStatus.Phase).To(Equal(string(release.StatusDeployed)))
			Expect(app.Status.HelmStatus.ChartUrl).To(Equal(chart2URL))

			// --- Uninstall ---
			Expect(k8sCli.Delete(ctx, app)).To(Succeed())

			_, err = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				alive, _ := testController.helmWorkers.IsHelmTaskAlive(appKey)
				return !alive
			}, 30*time.Second, 200*time.Millisecond).Should(BeTrue())

			// Reconcile until finalizer is removed and app is gone.
			Eventually(func() bool {
				_, _ = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey})
				gone := &runv1alpha1.Application{}
				return apierrors.IsNotFound(k8sCli.Get(ctx, appKey, gone))
			}, 15*time.Second, 200*time.Millisecond).Should(BeTrue())
		})
	})

	Context("two apps in parallel: install then uninstall", func() {
		It("installs two apps with different charts in parallel, then uninstalls both", func() {
			chartAAbs, err := filepath.Abs("../../test/data/chart-a")
			Expect(err).NotTo(HaveOccurred())
			chartAUrl := "file://" + chartAAbs

			chartBAbs, err := filepath.Abs("../../test/data/chart-b")
			Expect(err).NotTo(HaveOccurred())
			chartBUrl := "file://" + chartBAbs

			appKey1 := types.NamespacedName{Name: "e2e-app-1", Namespace: "ns1"}
			appKey2 := types.NamespacedName{Name: "e2e-app-2", Namespace: "ns1"}

			app1 := &runv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:        appKey1.Name,
					Namespace:   appKey1.Namespace,
					Annotations: map[string]string{AnnotationRunaiApp: "true"},
				},
				Spec: runv1alpha1.ApplicationSpec{
					Helm: &runv1alpha1.HelmSpec{ChartUrl: chartAUrl},
				},
			}
			app2 := &runv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:        appKey2.Name,
					Namespace:   appKey2.Namespace,
					Annotations: map[string]string{AnnotationRunaiApp: "true"},
				},
				Spec: runv1alpha1.ApplicationSpec{
					Helm: &runv1alpha1.HelmSpec{ChartUrl: chartBUrl},
				},
			}
			Expect(k8sCli.Create(ctx, app1)).To(Succeed())
			Expect(k8sCli.Create(ctx, app2)).To(Succeed())

			// --- Install both (reconcile both so two install tasks are queued; workers run them in parallel) ---
			_, err = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey1})
			Expect(err).NotTo(HaveOccurred())
			_, err = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey2})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				alive1, _ := testController.helmWorkers.IsHelmTaskAlive(appKey1)
				alive2, _ := testController.helmWorkers.IsHelmTaskAlive(appKey2)
				return !alive1 && !alive2
			}, 30*time.Second, 200*time.Millisecond).Should(BeTrue())

			Eventually(func() bool {
				_, _ = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey1})
				_, _ = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey2})
				a1, a2 := &runv1alpha1.Application{}, &runv1alpha1.Application{}
				if k8sCli.Get(ctx, appKey1, a1) != nil || k8sCli.Get(ctx, appKey2, a2) != nil {
					return false
				}
				return a1.Status.ActionStatus != nil && a1.Status.ActionStatus.Phase == helm_deployer.ActionCompleted &&
					a2.Status.ActionStatus != nil && a2.Status.ActionStatus.Phase == helm_deployer.ActionCompleted
			}, 15*time.Second, 200*time.Millisecond).Should(BeTrue())

			app1 = &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, appKey1, app1)).To(Succeed())
			Expect(app1.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionInstall)))
			Expect(app1.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionCompleted))
			Expect(app1.Status.ActionStatus.StartedAt.IsZero()).To(BeFalse())
			Expect(app1.Status.ActionStatus.CompletedAt.IsZero()).To(BeFalse())
			Expect(app1.Status.ActionStatus.Message).To(ContainSubstring("install completed"))
			Expect(app1.Status.ActionStatus.Error).To(Equal(""))
			Expect(app1.Status.HelmStatus.Phase).To(Equal(string(release.StatusDeployed)))
			Expect(app1.Status.HelmStatus.ChartUrl).To(Equal(chartAUrl))

			app2 = &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, appKey2, app2)).To(Succeed())
			Expect(app2.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionInstall)))
			Expect(app2.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionCompleted))
			Expect(app2.Status.ActionStatus.StartedAt.IsZero()).To(BeFalse())
			Expect(app2.Status.ActionStatus.CompletedAt.IsZero()).To(BeFalse())
			Expect(app2.Status.ActionStatus.Message).To(ContainSubstring("install completed"))
			Expect(app2.Status.ActionStatus.Error).To(Equal(""))
			Expect(app2.Status.HelmStatus.Phase).To(Equal(string(release.StatusDeployed)))
			Expect(app2.Status.HelmStatus.ChartUrl).To(Equal(chartBUrl))

			// --- Uninstall both ---
			Expect(k8sCli.Delete(ctx, app1)).To(Succeed())
			Expect(k8sCli.Delete(ctx, app2)).To(Succeed())

			_, err = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey1})
			Expect(err).NotTo(HaveOccurred())
			_, err = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey2})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				alive1, _ := testController.helmWorkers.IsHelmTaskAlive(appKey1)
				alive2, _ := testController.helmWorkers.IsHelmTaskAlive(appKey2)
				return !alive1 && !alive2
			}, 30*time.Second, 200*time.Millisecond).Should(BeTrue())

			Eventually(func() bool {
				_, _ = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey1})
				_, _ = testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey2})
				gone1, gone2 := &runv1alpha1.Application{}, &runv1alpha1.Application{}
				return apierrors.IsNotFound(k8sCli.Get(ctx, appKey1, gone1)) &&
					apierrors.IsNotFound(k8sCli.Get(ctx, appKey2, gone2))
			}, 15*time.Second, 200*time.Millisecond).Should(BeTrue())
		})
	})
})
