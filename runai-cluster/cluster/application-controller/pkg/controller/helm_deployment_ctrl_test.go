/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/test"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/deployers/helm_deployer"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// The tests below validate the behavior of the controller itself.
// The building blocks are:
//   - Controller that works with:
//   - in-memory HelmReleaseClient, so the release info is similar to production but in memory
//   - file://local-directory downloader
//   - We do not have a worker pool, we only validate that the correct task was submitted, and then we
//     drain the channel. the functionality of the HelmTasks will be tested separately.
//
// ┌──────────────────────────────────────────────────────────────────────────────────────┐
// │  ┌───────────────────┐     ┌──────────────────────────────────────────────────────┐  │
// │  │  fake.Client (K8s)│     │  test.NewTestHelmStore()                             │  │
// │  │  + scheme         │     │      storage.Init(driver.NewMemory())                │  │
// │  │  + namespace      │     │      in-memory Helm release state                    │  │
// │  └────────┬──────────┘     └───────────────────────┬──────────────────────────────┘  │
// │           ▼                                        ▼                                 │
// │  ┌────────────────────────────────────────────────────────────────────────────────┐  │
// │  │  NewApplicationControllerForTest(k8sCli, releaseClientFactory,                 │  │
// │  │                                  downloaderFactory                             │  │
// │  │  ┌─────────────────────────────────────────────────────────────────────────┐   │  │
// │  │  │  ApplicationController                                                  │   │  │
// │  │  │    • Reconcile(ctx, Request)                                            │   │  │
// │  │  │    • releaseClientFactory(id) ──► NewTestReleaseClient (local store)    │   │  │
// │  │  │    • downloaderFactory() ─► NewDownloaderForTest (local file)           │   │  │
// │  │  │    • helmWorkers.jobsChannel (buffered, no consumers)                   │   │  │
// │  │  │    • helmWorkers.runningTasks (map)                                     │   │  │
// │  │  └─────────────────────────────────────────────────────────────────────────┘   │  │
// │  └────────────────────────────────────────────────────────────────────────────────┘  │
// │           │  HelmJobsChannel()                                                       │
// │           ▼                                                                          │
// │  ┌───────────────────┐     drainJob(appKey, tag)                                     │
// │  │  HelmJob channel  │  ◄── Receive job, then ClearHelmTask(appKey)                  │
// │  │  (test drains)    │                                                               │
// │  └───────────────────┘                                                               │
// └──────────────────────────────────────────────────────────────────────────────────────┘
//
// Simulating release lifecycle (tests write to store directly):
// ┌────────────────────────────────────────────────────────────────────────────────┐
// │  store.Create(Release{Status: PendingInstall})  →  "install in progress"       │
// │  store.Update(Release{Status: Deployed})        →  "install done"              │
// │  store.Update(Release{Status: Uninstalling})    →  "uninstall in progress"     │
// │  store.Delete(name, version)                    →  "uninstall done"            │
// └────────────────────────────────────────────────────────────────────────────────┘
var _ = Describe("helm deployment controller test", func() {
	var (
		ctx            context.Context
		k8sCli         client.Client
		store          *storage.Storage
		testController *ApplicationController
	)

	chartUrl1 := "https://charts.example.com/myapp-1.0.0.tgz"
	chartUrl2 := "https://charts.example.com/myapp-2.0.0.tgz"

	appKey1 := types.NamespacedName{Name: "app-1", Namespace: "ns1"}

	makeApp := func(appKey types.NamespacedName) *runv1alpha1.Application {
		return &runv1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:       appKey.Name,
				Namespace:  appKey.Namespace,
				Finalizers: []string{HelmApplicationFinalizer},
				Annotations: map[string]string{
					AnnotationRunaiApp: "true",
				},
			},
			Spec: runv1alpha1.ApplicationSpec{
				Helm: &runv1alpha1.HelmSpec{
					ChartUrl: "https://charts.example.com/myapp-1.0.0.tgz",
				},
			},
		}
	}

	drainJob := func(appKey types.NamespacedName, tag string) HelmJob {
		var job HelmJob
		Expect(testController.HelmJobsChannel()).To(Receive(&job), tag)
		testController.ClearHelmTask(appKey)
		return job
	}

	reconcile := func(appKey types.NamespacedName) ctrl.Result {
		result, err := testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey})
		Expect(err).NotTo(HaveOccurred())
		return result
	}

	getApplication := func(appKey types.NamespacedName) *runv1alpha1.Application {
		app := &runv1alpha1.Application{}
		Expect(k8sCli.Get(ctx, appKey, app)).To(Succeed())
		return app
	}

	BeforeEach(func() {
		ctx = context.Background()

		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(runv1alpha1.AddToScheme(scheme)).To(Succeed())

		k8sCli = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&runv1alpha1.Application{}).
			Build()

		Expect(k8sCli.Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: appKey1.Namespace},
		})).To(Succeed())

		store = test.NewTestHelmStore()

		testController = NewApplicationControllerForTest(k8sCli,
			func(id types.NamespacedName, downloaderFactory helm_deployer.DownloaderFactory) (*helm_deployer.ReleaseClient, error) {
				return test.NewTestReleaseClient(id, store, downloaderFactory), nil
			},
			helm_deployer.NewDownloader,
		)
	})

	Context("spec is missing", func() {
		It("should fail with error", func() {
			// create the app with no spec
			app := makeApp(appKey1)
			app.Spec.Helm = nil
			Expect(k8sCli.Create(ctx, app)).To(Succeed())

			_, err := testController.Reconcile(ctx, ctrl.Request{NamespacedName: appKey1})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("spec.helm is required"))
		})
	})

	Context("chartRef", func() {
		It("uses chartRef and reports correct ChartUrl in status", func() {
			app := makeApp(appKey1)
			app.Spec.Helm = &runv1alpha1.HelmSpec{
				ChartRef: &runv1alpha1.ChartRef{
					Repository: "https://charts.example.com/bitnami",
					Name:       "myapp",
					Version:    "1.0.0",
				},
			}
			Expect(k8sCli.Create(ctx, app)).To(Succeed())

			// Reconcile: install task queued; status should have derived URL (repository/name:version)
			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Second))

			job := drainJob(appKey1, "after install")
			Expect(job.Task.Action()).To(Equal(helm_deployer.TaskActionInstall))

			current := getApplication(appKey1)
			Expect(current.Status.HelmStatus.ChartUrl).To(Equal("https://charts.example.com/bitnami/myapp:1.0.0"))

			// Simulate install completing; status ChartUrl should remain correct
			Expect(store.Create(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info:      &release.Info{Status: release.StatusDeployed},
				Version:   1,
			})).To(Succeed())

			result = reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeZero())

			current = getApplication(appKey1)
			Expect(current.Status.HelmStatus.Phase).To(Equal(string(release.StatusDeployed)))
			Expect(current.Status.HelmStatus.ChartUrl).To(Equal("https://charts.example.com/bitnami/myapp:1.0.0"))
		})
	})

	Context("happy path: install then uninstall", func() {
		It("should install the app and uninstall it successfully", func() {
			// create the app
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())

			// reconcile: install task queued
			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Second))

			job := drainJob(appKey1, "after install")
			Expect(job.Task.Action()).To(Equal(helm_deployer.TaskActionInstall))

			current := getApplication(appKey1)
			Expect(current.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionInstall)))
			Expect(current.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionQueued))
			Expect(current.Status.HelmStatus.ChartUrl).To(Equal("https://charts.example.com/myapp-1.0.0.tgz"))

			// simulate install in-progress
			Expect(store.Create(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info: &release.Info{
					Status: release.StatusPendingInstall,
				},
				Version: 1,
			})).To(Succeed())

			// reconcile: updates the status
			result = reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Second))
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			current = getApplication(appKey1)
			Expect(current.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionInstall))) // unchanged
			Expect(current.Status.HelmStatus.Phase).To(Equal(string(release.StatusPendingInstall)))
			Expect(current.Status.HelmStatus.ChartUrl).To(Equal("https://charts.example.com/myapp-1.0.0.tgz"))

			// simulate install completing
			Expect(store.Update(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info: &release.Info{
					Status: release.StatusDeployed,
				},
				Version: 1,
			})).To(Succeed())

			// Reconcile: deployed, no new job, only status updated
			result = reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeZero())
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			current = getApplication(appKey1)
			Expect(current.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionInstall)))
			Expect(current.Status.HelmStatus.Phase).To(Equal(string(release.StatusDeployed)))
			Expect(current.Status.HelmStatus.ChartUrl).To(Equal(chartUrl1))

			// delete the app
			Expect(k8sCli.Delete(ctx, getApplication(appKey1))).To(Succeed())

			// reconcile: uninstall submitted
			result = reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Second))

			job = drainJob(appKey1, "after uninstall")
			Expect(job.Task.Action()).To(Equal(helm_deployer.TaskActionUninstall))

			// simulate uninstall in progress
			Expect(store.Update(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info: &release.Info{
					Status: release.StatusUninstalling,
				},
				Version: 1,
			})).To(Succeed())

			result = reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Second))
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			current = getApplication(appKey1)
			Expect(current.DeletionTimestamp).NotTo(BeNil())
			Expect(current.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionUninstall)))
			Expect(current.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionQueued))
			Expect(current.Status.HelmStatus.ChartUrl).To(Equal(chartUrl1))

			// Simulate uninstall completed
			_, err := store.Delete(appKey1.Name, 1)
			Expect(err).NotTo(HaveOccurred())

			// Reconcile: not deployed → finalizer removed, app deleted
			result = reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeZero())

			gone := &runv1alpha1.Application{}
			Expect(apierrors.IsNotFound(k8sCli.Get(ctx, appKey1, gone))).To(BeTrue())
		})
	})

	Context("application is not deployed", func() {
		It("can be removed without uninstall", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())
			Expect(k8sCli.Delete(ctx, getApplication(appKey1))).To(Succeed())

			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeZero())
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			gone := &runv1alpha1.Application{}
			Expect(apierrors.IsNotFound(k8sCli.Get(ctx, appKey1, gone))).To(BeTrue())
		})
	})

	Context("app is currently uninstalling", func() {
		It("waits while release is uninstalling, then installs when release is gone", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())

			// simulate release is currently uninstalling
			Expect(store.Create(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info: &release.Info{
					Status: release.StatusUninstalling,
				},
				Version: 1,
			})).To(Succeed())

			// First reconcile: in-progress detected, requeue — no task submitted
			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Second))
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			// Simulate uninstall completing: remove release from store
			_, err := store.Delete(appKey1.Name, 1)
			Expect(err).NotTo(HaveOccurred())

			// Second reconcile: no release → install task submitted
			result = reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Second))
			job := drainJob(appKey1, "after install")
			Expect(job.Task.Action()).To(Equal(helm_deployer.TaskActionInstall))

			current := getApplication(appKey1)
			Expect(current.DeletionTimestamp).To(BeNil())
			Expect(current.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionInstall)))
			Expect(current.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionQueued))
		})
	})

	Context("uninstall submit fails", func() {
		It("updates status with error and requeues when job channel is inactive", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())

			reconcile(appKey1) // sync status
			_ = drainJob(appKey1, "drain install if any")

			// simulate the release being deployed
			Expect(store.Create(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info:      &release.Info{Status: release.StatusDeployed},
				Version:   1,
			})).To(Succeed())

			// now delete the application
			Expect(k8sCli.Delete(ctx, getApplication(appKey1))).To(Succeed())

			// Make submit return "channel is inactive"
			testController.helmWorkers.jobsChannel = nil

			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Second))

			// action is still "install", because the new action cannot be fullfilled.
			app := getApplication(appKey1)
			Expect(app.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionInstall)))
			Expect(app.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionQueued))
			Expect(app.Status.HelmStatus.Error).To(ContainSubstring("cannot submit uninstall"))
		})
	})

	Context("install/upgrade skipped when task already queued or running", func() {
		It("requeues and does not submit a second task", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())

			// First reconcile: install task submitted, job in channel, runningTasks has appKey1
			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Second))
			Expect(testController.HelmJobsChannel()).NotTo(BeEmpty())

			// Do not drain — task remains "queued" from controller's perspective
			// Second reconcile: IsHelmTaskAlive true -> skip install/upgrade
			result = reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Second))

			// Only one job was ever submitted
			var job HelmJob
			Expect(testController.HelmJobsChannel()).To(Receive(&job))
			Expect(job.Task.Action()).To(Equal(helm_deployer.TaskActionInstall))
			Expect(testController.HelmJobsChannel()).To(BeEmpty())
		})
	})

	Context("last failed generation annotation invalid", func() {
		It("error for annotation does not parse as integer", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())

			//  simulate an install action that has failed so the next reconcile will attempt to continue from there
			app := getApplication(appKey1)
			app.Status.ActionStatus = &runv1alpha1.ActionStatus{
				Phase: helm_deployer.ActionFailed,
				Type:  string(helm_deployer.TaskActionInstall),
			}

			app.Annotations[AnnotationHelmLastFailedGeneration] = "notanumber"

			_, err := testController.reconcileHelmDeployment(ctx, app)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse last failed generation"))
		})
	})

	Context("last helm operation failed and spec unchanged", func() {
		It("requeues after RequeueAfterFailed and sets last-failed-generation annotation", func() {
			app := makeApp(appKey1)
			app.Annotations[AnnotationHelmLastFailedGeneration] = "1" // matches Generation after Create
			Expect(k8sCli.Create(ctx, app)).To(Succeed())

			// Simulate: last install failed, spec unchanged (lastFailedGen >= gen)
			app.Status.ActionStatus = &runv1alpha1.ActionStatus{
				Phase: helm_deployer.ActionFailed,
				Type:  string(helm_deployer.TaskActionInstall),
			}
			app.Status.HelmStatus = &runv1alpha1.HelmStatus{
				Phase: "failed",
			}
			app.Generation = 1

			// simulate the release being deployed
			Expect(store.Create(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info:      &release.Info{Status: release.StatusDeployed},
				Version:   1,
			})).To(Succeed())

			result, err := testController.reconcileHelmDeployment(ctx, app)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(RequeueAfterFailed))
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			updated := &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, appKey1, updated)).To(Succeed())
			Expect(updated.Annotations).To(HaveKeyWithValue(AnnotationHelmLastFailedGeneration, "1"))
		})
	})

	Context("upgrade following spec change", func() {
		It("submits upgrade task when spec changes after deploy", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())

			// Install: queue install, drain, then simulate deploy
			reconcile(appKey1)
			drainJob(appKey1, "install")

			Expect(store.Create(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info:      &release.Info{Status: release.StatusDeployed},
				Version:   1,
			})).To(Succeed())

			reconcile(appKey1) // status updated: ChartUrl = chartUrl1

			// Change spec to new chart, so ReleaseMatchesDesired is false
			app := getApplication(appKey1)
			app.Spec.Helm.ChartUrl = chartUrl2
			app.Generation = 2
			Expect(k8sCli.Update(ctx, app)).To(Succeed())

			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeNumerically(">", time.Second))

			job := drainJob(appKey1, "after upgrade")
			Expect(job.Task.Action()).To(Equal(helm_deployer.TaskActionUpgrade))

			current := getApplication(appKey1)
			Expect(current.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionUpgrade)))
			Expect(current.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionQueued))
		})
	})

	Context("deletion with no finalizer", func() {
		It("returns no requeue when application has no finalizer, can be removed", func() {
			app := makeApp(appKey1)
			app.Finalizers = nil
			app.DeletionTimestamp = ptr.To(metav1.Now())

			result, err := testController.reconcileHelmDeployment(ctx, app)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())
		})
	})

	Context("deletion while uninstall task is queued", func() {
		It("requeues after 5s and does not remove finalizer while task is alive", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())
			Expect(store.Create(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info:      &release.Info{Status: release.StatusDeployed},
				Version:   1,
			})).To(Succeed())
			reconcile(appKey1)
			_ = drainJob(appKey1, "drain install if any")
			reconcile(appKey1)

			Expect(k8sCli.Delete(ctx, getApplication(appKey1))).To(Succeed())
			reconcile(appKey1) // uninstall submitted
			// Do NOT drain — uninstall task stays "queued or running"

			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(Equal(5 * time.Second))

			current := getApplication(appKey1)
			Expect(controllerutil.ContainsFinalizer(current, HelmApplicationFinalizer)).To(BeTrue())
			Expect(current.DeletionTimestamp).NotTo(BeNil())
			Expect(current.Status.ActionStatus.Type).To(Equal(string(helm_deployer.TaskActionUninstall)))
			Expect(current.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionQueued))
		})
	})

	Context("add finalizer", func() {
		It("adds finalizer and follow up to install", func() {
			app := makeApp(appKey1)
			app.Finalizers = nil
			Expect(k8sCli.Create(ctx, app)).To(Succeed())

			result, err := testController.reconcileHelmDeployment(ctx, app)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeNumerically(">", 0))

			job := drainJob(appKey1, "after install")
			Expect(job.Task.Action()).To(Equal(helm_deployer.TaskActionInstall))

			current := getApplication(appKey1)
			Expect(controllerutil.ContainsFinalizer(current, HelmApplicationFinalizer)).To(BeTrue())
		})
	})

	Context("failing to get application when adding finalizer", func() {
		It("should return error when Get fails while adding finalizer", func() {
			// Controller uses client that fails Get for Application
			failingClient := test.NewFailingGetClient(k8sCli, fmt.Errorf("get failed"))

			ctrl := NewApplicationControllerForTest(failingClient,
				func(id types.NamespacedName, downloaderFactory helm_deployer.DownloaderFactory) (*helm_deployer.ReleaseClient, error) {
					return test.NewTestReleaseClient(id, store, downloaderFactory), nil
				},
				helm_deployer.NewDownloader,
			)

			// App exists, is deleted, has finalizer, no release (not deployed)
			app := makeApp(appKey1)
			Expect(k8sCli.Create(ctx, app)).To(Succeed())
			Expect(k8sCli.Delete(ctx, app)).To(Succeed())
			// No store.Create(release) → not deployed

			// Build app for reconcile (can't Get; wrapper would fail)
			appToReconcile := &runv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appKey1.Name,
					Namespace: appKey1.Namespace,
				},
				Spec: runv1alpha1.ApplicationSpec{
					Helm: &runv1alpha1.HelmSpec{ChartUrl: chartUrl1},
				},
				// set DeletionTimestamp, Finalizers, Annotations as in makeApp
			}

			_, err := ctrl.reconcileHelmDeployment(ctx, appToReconcile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get application"))
		})
	})

	Context("failing to get application when removing finalizer", func() {
		It("returns error when Get fails before removing finalizer (app not deployed)", func() {
			failingClient := test.NewFailingGetClient(k8sCli, fmt.Errorf("get failed"))
			ctrl := NewApplicationControllerForTest(failingClient,
				func(id types.NamespacedName, downloaderFactory helm_deployer.DownloaderFactory) (*helm_deployer.ReleaseClient, error) {
					return test.NewTestReleaseClient(id, store, downloaderFactory), nil
				},
				helm_deployer.NewDownloader,
			)

			// App is deleted, has finalizer, no release → not deployed → controller tries to remove finalizer, then Get fails
			app := makeApp(appKey1)
			Expect(k8sCli.Create(ctx, app)).To(Succeed())
			Expect(k8sCli.Delete(ctx, app)).To(Succeed())
			// No store.Create(release) → !deployed

			now := metav1.Now()
			appToReconcile := &runv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:              appKey1.Name,
					Namespace:         appKey1.Namespace,
					DeletionTimestamp: &now,
					Finalizers:        []string{HelmApplicationFinalizer},
					Annotations:       map[string]string{AnnotationRunaiApp: "true"},
				},
				Spec: runv1alpha1.ApplicationSpec{
					Helm: &runv1alpha1.HelmSpec{ChartUrl: chartUrl1},
				},
			}

			_, err := ctrl.reconcileHelmDeployment(ctx, appToReconcile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get application for finalizer removal"))
		})
	})

	Context("stuck action status recovery", func() {
		It("reconciler advances ActionStatus to 'completed' when the release is deployed, matches spec, and no task is alive", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())

			// Reconcile: install task queued; HelmStatus.ChartUrl recorded so
			// ReleaseMatchesDesired can compare against the spec later.
			reconcile(appKey1)
			drainJob(appKey1, "install task")

			current := getApplication(appKey1)
			Expect(current.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionQueued))
			Expect(current.Status.HelmStatus.ChartUrl).To(Equal(chartUrl1))

			// Simulate the worker having set the phase to "running" without
			// subsequently writing "completed" (e.g. all retries were exhausted
			// due to concurrent status-update conflicts).
			stuck := current.DeepCopy()
			stuck.Status.ActionStatus.Phase = helm_deployer.ActionRunning
			stuck.Status.ActionStatus.Message = "install task is running"
			Expect(k8sCli.Status().Update(ctx, stuck)).To(Succeed())

			// Simulate the Helm install completing in the background.
			Expect(store.Create(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info:      &release.Info{Status: release.StatusDeployed},
				Version:   1,
			})).To(Succeed())

			// Precondition: ActionStatus is non-terminal, no task is alive.
			current = getApplication(appKey1)
			Expect(current.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionRunning))
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			// Reconcile: release is deployed and matches spec, no task alive,
			// ActionStatus is non-terminal → recovery writes "completed".
			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeZero())
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			recovered := getApplication(appKey1)
			Expect(recovered.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionCompleted))
			Expect(recovered.Status.ActionStatus.Message).To(Equal("completed"))
			Expect(recovered.Status.HelmStatus.Phase).To(Equal(string(release.StatusDeployed)))
		})

		It("does not overwrite ActionStatus when it is already 'completed'", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())

			reconcile(appKey1)
			drainJob(appKey1, "install task")

			// ActionStatus already in its terminal completed state.
			completed := getApplication(appKey1)
			completed.Status.ActionStatus.Phase = helm_deployer.ActionCompleted
			completed.Status.ActionStatus.Message = "install completed"
			Expect(k8sCli.Status().Update(ctx, completed)).To(Succeed())

			Expect(store.Create(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info:      &release.Info{Status: release.StatusDeployed},
				Version:   1,
			})).To(Succeed())

			// Reconcile: no recovery triggered, no new task, message left untouched.
			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeZero())
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			current := getApplication(appKey1)
			Expect(current.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionCompleted))
			Expect(current.Status.ActionStatus.Message).To(Equal("install completed"))
		})

		It("does not overwrite ActionStatus when it is 'queued' (reconciler-set, not a worker-execution phase)", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())

			// Reconcile: install task submitted; ActionStatus.Phase = "queued".
			reconcile(appKey1)
			drainJob(appKey1, "install task")

			current := getApplication(appKey1)
			Expect(current.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionQueued))

			// Release is deployed (e.g. install ran in a previous controller instance).
			Expect(store.Create(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info:      &release.Info{Status: release.StatusDeployed},
				Version:   1,
			})).To(Succeed())

			// Reconcile: "queued" is not a worker-execution phase — no recovery triggered.
			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeZero())
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			current = getApplication(appKey1)
			Expect(current.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionQueued))
		})

		It("does not overwrite ActionStatus when it is already 'failed'", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())

			reconcile(appKey1)
			drainJob(appKey1, "install task")

			// ActionStatus already in its terminal failed state.
			failed := getApplication(appKey1)
			failed.Status.ActionStatus.Phase = helm_deployer.ActionFailed
			failed.Status.ActionStatus.Message = "install failed"
			Expect(k8sCli.Status().Update(ctx, failed)).To(Succeed())

			Expect(store.Create(&release.Release{
				Name:      appKey1.Name,
				Namespace: appKey1.Namespace,
				Info:      &release.Info{Status: release.StatusDeployed},
				Version:   1,
			})).To(Succeed())

			// Reconcile: "failed" is terminal — no recovery applied; reconciler
			// requeues after RequeueAfterFailed as the spec has not changed.
			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(Equal(RequeueAfterFailed))
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			current := getApplication(appKey1)
			Expect(current.Status.ActionStatus.Phase).To(Equal(helm_deployer.ActionFailed))
			Expect(current.Status.ActionStatus.Message).To(Equal("install failed"))
		})
	})

	Context("leftover PVCs are deleted before finalizer removal", func() {
		It("deletes PVCs belonging to the application and waits for them to disappear", func() {
			Expect(k8sCli.Create(ctx, makeApp(appKey1))).To(Succeed())

			// Create a PVC that belongs to this application (Helm-managed labels + release annotation)
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "data-app-1-0",
					Namespace: appKey1.Namespace,
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "Helm",
					},
					Annotations: map[string]string{
						"meta.helm.sh/release-name": appKey1.Name,
					},
				},
			}
			Expect(k8sCli.Create(ctx, pvc)).To(Succeed())

			// Create a PVC that belongs to a DIFFERENT application — should not be touched
			otherPvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "data-other-app-0",
					Namespace: appKey1.Namespace,
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "Helm",
					},
					Annotations: map[string]string{
						"meta.helm.sh/release-name": "other-app",
					},
				},
			}
			Expect(k8sCli.Create(ctx, otherPvc)).To(Succeed())

			// Delete the application (no release in store → !deployed)
			Expect(k8sCli.Delete(ctx, getApplication(appKey1))).To(Succeed())

			// First reconcile: deletes the PVC, requeues to wait for it to disappear
			result := reconcile(appKey1)
			Expect(result.RequeueAfter).To(Equal(5 * time.Second))
			Expect(testController.HelmJobsChannel()).To(BeEmpty())

			// The application's PVC should already be gone (fake client deletes immediately)
			deletedPvc := &corev1.PersistentVolumeClaim{}
			Expect(apierrors.IsNotFound(k8sCli.Get(ctx, types.NamespacedName{
				Name: "data-app-1-0", Namespace: appKey1.Namespace,
			}, deletedPvc))).To(BeTrue())

			// Second reconcile: no PVCs remaining → finalizer removed, app deleted
			result = reconcile(appKey1)
			Expect(result.RequeueAfter).To(BeZero())

			// Application should be gone
			gone := &runv1alpha1.Application{}
			Expect(apierrors.IsNotFound(k8sCli.Get(ctx, appKey1, gone))).To(BeTrue())

			// The other application's PVC should still exist
			survivingPvc := &corev1.PersistentVolumeClaim{}
			Expect(k8sCli.Get(ctx, types.NamespacedName{
				Name: "data-other-app-0", Namespace: appKey1.Namespace,
			}, survivingPvc)).To(Succeed())
		})
	})
})
