package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/deployers/helm_deployer"
	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/providers"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// HelmApplicationFinalizer is the finalizer for Applications with spec.helm
	HelmApplicationFinalizer = "run.ai/helm-application"

	AnnotationRunaiApp = "run.ai/application"

	// AnnotationHelmLastFailedGeneration stores Application.Generation when we last skipped retry, due to failed phase.
	// Used to avoid a loop of nonstop retries. should retry when the user changes the spec (following generation to increase)
	// of after sensible time of RequeueAfterFailed time
	AnnotationHelmLastFailedGeneration = "run.ai/helm-last-failed-generation"

	// RequeueAfterFailed is used when the last install/upgrade failed and spec has not changed in a few
	// minutes. we requeue the request hoping this time it will succeed.
	RequeueAfterFailed = 5 * time.Minute
)

// reconcileHelmDeployment handles reconciliation for Applications with spec.helm (controller install mode)
func (c *ApplicationController) reconcileHelmDeployment(ctx context.Context, application *v1alpha1.Application) (ctrl.Result, error) {
	appKey := types.NamespacedName{
		Namespace: application.Namespace,
		Name:      application.Name,
	}

	log.Info().Msgf("--- reconciling helm for application %s ---", appKey.String())

	helmReleaseClient, err := c.releaseClientFactory(appKey, c.downloaderFactory)
	if err != nil {
		err := fmt.Errorf("failed to create helm client for application %s: %w", appKey.String(), err)
		log.Error().Err(err).Send()
		return ctrl.Result{}, err
	}

	releaseInfo, err := helmReleaseClient.GetRelease()
	if err != nil {
		err := fmt.Errorf("failed to get release info for application %s: %w", appKey.String(), err)
		log.Error().Err(err).Send()
		return ctrl.Result{}, err
	}

	deployed := releaseInfo != nil && releaseInfo.Info != nil && releaseInfo.Info.Deleted.IsZero()

	// initialize the status with the information we know at this point
	status := helm_deployer.NewAppStatus().
		WithReleaseInfo(releaseInfo).
		WithDeployedChart(helm_deployer.ChartUrl(application.Spec.Helm))

	// If a Helm operation is already in progress, wait for it to settle before taking any action.
	// this will normally happen if the controller is restarting, and the task queue is empty so we
	// based our knowledge on the information in the release itself.
	if releaseInfo != nil && helm_deployer.IsHelmWorkingOnThisRelease(releaseInfo.Info) {
		log.Warn().Msgf("release %s has in-progress status %s, waiting", appKey.String(), releaseInfo.Info.Status)
		c.updateHelmStatus(ctx, appKey, status)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Handle deletion - remove finalizer only after uninstall succeeds or after MaxUninstallAttempts failures
	if !application.DeletionTimestamp.IsZero() {
		log.Info().Msgf("application %s is marked for deletion", appKey.String())

		if !controllerutil.ContainsFinalizer(application, HelmApplicationFinalizer) {
			log.Debug().Msgf("application %s has no finalizer, can be removed", application.GetName())
			return ctrl.Result{}, nil
		}

		if !deployed {
			log.Info().Msgf("application %s is not deployed, cleaning up leftover resources", appKey.String())

			remaining, err := c.deleteApplicationPVCs(ctx, appKey)
			if err != nil {
				log.Error().Err(err).Msgf("failed to delete PVCs for application %s", appKey.String())
				return ctrl.Result{}, err
			}
			if remaining > 0 {
				log.Info().Msgf("application %s still has %d PVC(s) pending deletion, requeueing", appKey.String(), remaining)
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}

			// fetch so we have the latest ResourceVersion and avoid conflict when removing finalizer.
			if err := c.client.Get(ctx, appKey, application); err != nil {
				err := fmt.Errorf("failed to get application for finalizer removal: %w", err)
				log.Error().Err(err).Send()
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(application, HelmApplicationFinalizer)

			if err := c.client.Update(ctx, application); err != nil {
				err := fmt.Errorf("failed to remove finalizer: %w", err)
				log.Error().Err(err).Send()
				return ctrl.Result{}, err
			}

			log.Debug().Msgf("removed finalizer of application %s", appKey.String())

			return ctrl.Result{}, nil
		}

		if isTaskAlive, taskAction := c.helmWorkers.IsHelmTaskAlive(appKey); isTaskAlive {
			log.Info().Msgf("helm task %s for %s is queued or running, waiting before removing the application", taskAction, appKey.String())
			c.updateHelmStatus(ctx, appKey, status)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		task, err := helm_deployer.NewTask(ctx, helm_deployer.TaskActionUninstall, helmReleaseClient, c.client, application)
		if err != nil {
			c.updateHelmStatus(ctx, appKey, status.WithError(err))
			err := fmt.Errorf("failed to create uninstall task for %s: %w", appKey.String(), err)
			log.Error().Err(err).Send()
			return ctrl.Result{}, err
		}

		c.submitTask(ctx, appKey, status, task)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(application, HelmApplicationFinalizer) {
		log.Debug().Msgf("application %s has no finalizer yet, adding finalizer", application.GetName())

		// fetch so we have the latest ResourceVersion and avoid conflict.
		if err := c.client.Get(ctx, appKey, application); err != nil {
			err := fmt.Errorf("failed to get application for finalizer: %w", err)
			log.Error().Err(err).Send()
			return ctrl.Result{}, err
		}

		if !controllerutil.ContainsFinalizer(application, HelmApplicationFinalizer) {
			patch := client.MergeFrom(application.DeepCopy())

			controllerutil.AddFinalizer(application, HelmApplicationFinalizer)

			if err := c.client.Patch(ctx, application, patch); err != nil {
				err := fmt.Errorf("failed to add finalizer: %w", err)
				log.Error().Err(err).Send()
				return ctrl.Result{}, err
			}

			// now that we have the finalizer, follow up to process the application
		}
	}

	if isTaskAlive, action := c.helmWorkers.IsHelmTaskAlive(appKey); isTaskAlive {
		log.Warn().Msgf("helm task %s for %s is queued or running, skipping install/upgrade", action, appKey.String())
		c.updateHelmStatus(ctx, appKey, status)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// If the last install/upgrade already failed, don't send the same task again (avoids frequent reconcile loop).
	// Retry after RequeueAfterFailed or when the user changes the spec (generation increases).
	if application.Status.ActionStatus != nil && application.Status.ActionStatus.Phase == helm_deployer.ActionFailed {
		var lastFailedGen int64
		if a := application.Annotations[AnnotationHelmLastFailedGeneration]; a != "" {
			_, err = fmt.Sscanf(a, "%d", &lastFailedGen)
			if err != nil {
				err := fmt.Errorf("failed to parse last failed generation for %s: %w", appKey.String(), err)
				log.Error().Err(err).Send()
				return ctrl.Result{}, err
			}
		}

		gen := application.Generation

		// spec has not changed since the last failure — skip retry
		if lastFailedGen == 0 || lastFailedGen >= gen {
			log.Warn().Msgf("last helm operation for %s failed (%s), spec was not changed since, retry after %v", appKey.String(), application.Status.HelmStatus.Phase, RequeueAfterFailed)

			patch := client.MergeFrom(application.DeepCopy())

			if application.Annotations == nil {
				application.Annotations = make(map[string]string)
			}

			application.Annotations[AnnotationHelmLastFailedGeneration] = fmt.Sprintf("%d", gen)

			if err := c.client.Patch(ctx, application, patch); err != nil {
				err := fmt.Errorf("failed to update application annotations: %w", err)
				log.Error().Err(err).Send()
				return ctrl.Result{}, err
			}

			return ctrl.Result{RequeueAfter: RequeueAfterFailed}, nil
		}
		// Spec has changed since last failure; allow retry (fall through).
	}

	var task helm_deployer.Task

	if !deployed {
		log.Info().Msgf("app %s is not deployed, created install task", appKey.String())
		task, err = helm_deployer.NewTask(ctx, helm_deployer.TaskActionInstall, helmReleaseClient, c.client, application)
	} else {
		if helm_deployer.ReleaseMatchesDesired(releaseInfo, application.Spec.Helm, application.Status.HelmStatus) {
			log.Info().Msgf("release already matches spec for %s, skipping upgrade", appKey.String())

			// Recovery: if ActionStatus is stuck in a non-terminal phase (e.g. "running") it means
			// the worker's final status write was lost to concurrent K8s status-update conflicts.
			// Since the release is already deployed and matches the desired spec, it is safe to
			// mark the action as completed here so the application does not remain stuck.
			if isNonTerminalActionPhase(application) {
				log.Warn().Msgf("recovering stuck action status for %s (was %q), marking as completed",
					appKey.String(), application.Status.ActionStatus.Phase)
				c.updateHelmStatus(ctx, appKey, status.RecoveryCompleted())
			} else {
				c.updateHelmStatus(ctx, appKey, status)
			}

			return ctrl.Result{}, nil
		}

		task, err = helm_deployer.NewTask(ctx, helm_deployer.TaskActionUpgrade, helmReleaseClient, c.client, application)
	}

	if err != nil {
		c.updateHelmStatus(ctx, appKey, status.WithError(err))

		err := fmt.Errorf("failed to create helm task for %s: %w", appKey.String(), err)
		log.Error().Err(err).Send()
		return ctrl.Result{}, err
	}

	c.submitTask(ctx, appKey, status, task)
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// submitTask reserves a worker slot, writes "queued" status, then dispatches the task to the
// worker channel. Writing status before dispatch ensures the worker's status updates (started,
// completed) always come after — a worker can only start once the job is in the channel.
// If reserve fails (e.g. channel inactive) ActionStatus is left unchanged and only the
// HelmStatus error field is updated.
func (c *ApplicationController) submitTask(ctx context.Context, appKey types.NamespacedName, status *helm_deployer.AppStatus, task helm_deployer.Task) {
	if err := c.helmWorkers.reserve(appKey, task); err != nil {
		c.updateHelmStatus(ctx, appKey, status.WithError(err))
		return
	}
	c.updateHelmStatus(ctx, appKey, status.TaskQueued(task))
	if err := c.helmWorkers.dispatch(appKey, task); err != nil {
		c.updateHelmStatus(ctx, appKey, status.WithError(err))
	}
}

// updateHelmStatus updates the application status based on the Helm deployment result
func (c *ApplicationController) updateHelmStatus(ctx context.Context, appKey types.NamespacedName, appStatusChange *helm_deployer.AppStatus) {
	clientKey := types.NamespacedName{Namespace: appKey.Namespace, Name: appKey.Name}

	// Use a higher retry count: each Reconcile call issues two Status().Update calls
	// (one in reconcileHelmDeployment, one in UpdateApplicationStatus), so under high
	// concurrency the worker's write can lose several consecutive races.
	maxRetries := 10

	for attempt := 1; attempt <= maxRetries; attempt++ {
		var application v1alpha1.Application
		err := c.client.Get(ctx, clientKey, &application)
		if err != nil {
			log.Error().Err(err).Msgf("failed to load application %s/%s for status update", application.Namespace, application.Name)
			return
		}

		changed := appStatusChange.ApplyOnApplication(&application)
		if !changed {
			log.Debug().Msgf("application %s/%s status %s unchanged", application.Namespace, application.Name, application.Status.ActionStatus.Phase)
			return
		}

		err = c.client.Status().Update(ctx, &application)
		if err == nil {
			log.Info().Msgf("updated application %s/%s status %s", application.Namespace, application.Name, application.Status.ActionStatus.Phase)
			return
		}

		if !errors.IsConflict(err) {
			log.Error().Err(err).Msgf("failed to update application %s/%s status", application.Namespace, application.Name)
			return
		}

		if attempt < maxRetries {
			log.Warn().Msgf("status update conflict for application %s/%s on attempt #%d, retrying", application.Namespace, application.Name, attempt)
		} else {
			log.Error().Msgf("give up status update for application %s/%s, after %d attempts", application.Namespace, application.Name, attempt)
		}
	}
}

// isNonTerminalActionPhase returns true when the application's ActionStatus phase is one of
// the in-progress values set exclusively by a worker goroutine during task execution
// (started, downloading, running). It intentionally excludes "queued", which is written by
// the reconciler at submission time and is not a worker-execution state, so that a normally
// submitted task that has not yet been picked up is never mistakenly advanced to "completed".
func isNonTerminalActionPhase(application *v1alpha1.Application) bool {
	if application.Status.ActionStatus == nil {
		return false
	}
	switch application.Status.ActionStatus.Phase {
	case helm_deployer.ActionStarted,
		helm_deployer.ActionDownloading,
		helm_deployer.ActionRunning:
		return true
	}
	return false
}

// deleteApplicationPVCs deletes PVCs that belong to the given application and
// returns the number of PVCs still present (including those now pending deletion).
// Helm uninstall does not remove PVCs that carry a "helm.sh/resource-policy: keep"
// annotation, or PVCs created by StatefulSet volumeClaimTemplates that happen to
// carry standard Helm labels. This function ensures they are cleaned up before the
// Application finalizer is removed.
func (c *ApplicationController) deleteApplicationPVCs(ctx context.Context, appKey types.NamespacedName) (int, error) {
	pvcList := &corev1.PersistentVolumeClaimList{}
	listOpts := &client.ListOptions{
		Namespace: appKey.Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{
			providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue,
		}),
	}

	if err := c.client.List(ctx, pvcList, listOpts); err != nil {
		return 0, fmt.Errorf("failed to list PVCs in namespace %s: %w", appKey.Namespace, err)
	}

	remaining := 0
	for i := range pvcList.Items {
		pvc := &pvcList.Items[i]

		annotations := pvc.GetAnnotations()
		releaseName, exists := annotations[constants.ResourceReleaseAnnotationName]
		if !exists || releaseName != appKey.Name {
			continue
		}

		if pvc.DeletionTimestamp != nil {
			remaining++
			continue
		}

		log.Info().Msgf("deleting leftover PVC %s/%s for application %s", pvc.Namespace, pvc.Name, appKey.String())
		if err := c.client.Delete(ctx, pvc); err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return 0, fmt.Errorf("failed to delete PVC %s/%s: %w", pvc.Namespace, pvc.Name, err)
		}
		remaining++
	}

	return remaining, nil
}
