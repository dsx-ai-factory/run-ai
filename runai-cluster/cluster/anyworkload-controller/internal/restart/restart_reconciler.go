/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package restart

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	jsonpatch "github.com/evanphx/json-patch/v5"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerUtil "sigs.k8s.io/controller-runtime/pkg/controller"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	dromospredicates "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/predicates"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/template"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

const (
	// frameworkClearConditionType reports whether the framework (the child workload object) was cleared
	frameworkClearConditionType = "FrameworkClear"

	// childDeletionRequeueInterval is the poll interval while waiting for the child object to disappear
	childDeletionRequeueInterval = 2 * time.Second

	// defaultChildDeletionTimeout applies when the configured timeout is not a positive duration
	defaultChildDeletionTimeout = 10 * time.Minute
)

var (
	restartLogger                     = logger.Log.WithName("reconcile.restart")
	errRequestRefLabelNotFound        = errors.New("request-ref label not found")
	errRestartRequestNotFound         = errors.New("restart request not found")
	errRestartRequestWorkloadMismatch = errors.New("restart request targets another workload")
)

// RestartReconciler performs the restart itself on a WorkloadRunner the HandoffReconciler
type RestartReconciler struct {
	client.Client
	name string
	// childDeletionTimeout bounds the wait for the child object to disappear
	childDeletionTimeout time.Duration
}

// NewRestartReconciler creates a new RestartReconciler.
func NewRestartReconciler(client client.Client, childDeletionTimeout time.Duration) *RestartReconciler {
	if childDeletionTimeout <= 0 {
		childDeletionTimeout = defaultChildDeletionTimeout
	}
	return &RestartReconciler{
		Client:               client,
		name:                 "restart-reconciler",
		childDeletionTimeout: childDeletionTimeout,
	}
}

// WithName overrides the controller name (used by tests running multiple managers).
func (r *RestartReconciler) WithName(name string) *RestartReconciler {
	r.name = name
	return r
}

// SetupWithManager sets up the controller with the Manager.
func (r *RestartReconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconciles int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dromosv1alpha1.WorkloadRunner{}, builder.WithPredicates(
			predicate.GenerationChangedPredicate{},
			dromospredicates.ManagedByPredicate(RestartControllerManagedBy),
		)).
		Named(r.name).
		WithOptions(controllerUtil.Options{MaxConcurrentReconciles: maxConcurrentReconciles,
			RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
				constants.RateLimiterBaseDelay, constants.RateLimiterMaxDelay),
		}).
		Complete(r)
}

func (r *RestartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	trxLogger := restartLogger.WithValues("id", fmt.Sprintf("%s:%s", req.Namespace, req.Name))
	ctx = logger.IntoContext(ctx, trxLogger)

	trxLogger.Info("reconcile started")
	defer trxLogger.Info("reconcile ended")

	workload := &dromosv1alpha1.WorkloadRunner{}
	if err := r.Get(ctx, req.NamespacedName, workload); err != nil {
		if k8sErrors.IsNotFound(err) {
			trxLogger.Info("workload has been deleted")
			return ctrl.Result{}, nil
		}
		trxLogger.Error(err, "failed to load the workload")
		return ctrl.Result{}, err
	}

	// Complements ManagedByPredicate: events queued before managedBy was set still reach here.
	if workload.Spec.ManagedBy != RestartControllerManagedBy {
		trxLogger.Info("workload is managed by another controller, skipping", "managedBy", workload.Spec.ManagedBy)
		return ctrl.Result{}, nil
	}

	restartRequest, err := r.getRestartRequest(ctx, workload)
	if err != nil {
		return ctrl.Result{}, err
	}

	if r.isRestartCompleted(ctx, restartRequest) {
		// handing the workload back is the HandoffReconciler's job
		return ctrl.Result{}, nil
	}

	// The HandoffReconciler bumps the workload generation by 1 (label + managedBy update).
	// A second bump (delta=2) means spec overrides were already persisted in a previous
	// reconcile round. Since the workload patch happens after the child delete in
	// applyRestart, delta=2 implies the full restart already completed.
	// Any other delta (!= 1) means the RestartRequest is stale or invalid.
	generationDelta := workload.GetGeneration() - restartRequest.Spec.WorkloadRef.Generation
	if generationDelta <= 0 || generationDelta > 2 {
		return ctrl.Result{}, updateRestartRequestPhase(ctx, r.Client, restartRequest, runv1alpha1.RestartRequestPhaseFailed,
			fmt.Sprintf("generation mismatch: workload=%d, restartRequest=%d",
				workload.GetGeneration(), restartRequest.Spec.WorkloadRef.Generation))
	}
	if generationDelta == 2 {
		// Crash recovery: the controller died between applying the restart and finalizing it, so the
		// child was already deleted (deleted=true). setWorkloadPhaseToRestarting is idempotent.
		if err = r.setWorkloadPhaseToRestarting(ctx, workload); err != nil {
			return ctrl.Result{}, err
		}
		return r.finalizeRestart(ctx, workload, restartRequest, true, nil)
	}

	if err = r.setRestartRequestInProgress(ctx, restartRequest); err != nil {
		return ctrl.Result{}, err
	}

	if err = r.setWorkloadPhaseToRestarting(ctx, workload); err != nil {
		return ctrl.Result{}, err
	}

	deleted, deletedErr := r.applyRestart(ctx, workload, restartRequest)
	if k8sErrors.IsConflict(deletedErr) {
		// retry from a fresh
		return ctrl.Result{}, deletedErr
	}
	result, err := r.finalizeRestart(ctx, workload, restartRequest, deleted, deletedErr)
	if err != nil {
		trxLogger.Error(err, "failed to finalize restart")
	}
	return result, err
}

// getRestartRequest resolves the RestartRequest the run.ai/request-ref label points at
func (r *RestartReconciler) getRestartRequest(ctx context.Context, workload *dromosv1alpha1.WorkloadRunner) (*runv1alpha1.RestartRequest, error) {
	trxLogger := logger.FromContext(ctx)

	// the HandoffReconciler writes label and managedBy together, so a missing label is an anomaly
	requestRef := workload.Labels[RequestRefLabel]
	if requestRef == "" {
		trxLogger.Error(errRequestRefLabelNotFound, "clearing restart ownership")
		return nil, r.clearRestartOwnershipWithErr(ctx, workload, errRequestRefLabelNotFound)
	}

	list := &runv1alpha1.RestartRequestList{}
	err := r.List(ctx, list, client.InNamespace(workload.Namespace), client.MatchingFields{MetadataUIDField: requestRef})
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		trxLogger.Error(errRestartRequestNotFound, "clearing restart ownership", "requestRef", requestRef)
		return nil, r.clearRestartOwnershipWithErr(ctx, workload, errRestartRequestNotFound)
	}

	if list.Items[0].Spec.WorkloadRef.Name != workload.Name {
		return nil, fmt.Errorf("%w (workloadRef: %s)", errRestartRequestWorkloadMismatch, list.Items[0].Spec.WorkloadRef.Name)
	}
	return &list.Items[0], nil
}

// clearRestartOwnershipWithErr hands the workload back to dromos and returns cause with the clear error.
func (r *RestartReconciler) clearRestartOwnershipWithErr(ctx context.Context, workload *dromosv1alpha1.WorkloadRunner, cause error) error {
	delete(workload.Labels, RequestRefLabel)
	workload.Spec.ManagedBy = ""
	if err := r.Update(ctx, workload); err != nil {
		return fmt.Errorf("failed to clear restart ownership (cause :%v): %w", cause, err)
	}
	return cause
}

// isRestartCompleted returns true if the restart request phase is Succeeded or Failed (reconcile should end).
func (r *RestartReconciler) isRestartCompleted(ctx context.Context, restartRequest *runv1alpha1.RestartRequest) bool {
	if restartRequest.Status.Phase == runv1alpha1.RestartRequestPhaseSucceeded ||
		restartRequest.Status.Phase == runv1alpha1.RestartRequestPhaseFailed {
		logger.FromContext(ctx).Info("restart request already completed", "phase", restartRequest.Status.Phase)
		return true
	}
	return false
}

// setRestartRequestInProgress sets the restart request phase to InProgress
func (r *RestartReconciler) setRestartRequestInProgress(ctx context.Context, restartRequest *runv1alpha1.RestartRequest) error {
	if restartRequest.Status.Phase == runv1alpha1.RestartRequestPhaseInProgress {
		return nil
	}
	return updateRestartRequestPhase(ctx, r.Client, restartRequest, runv1alpha1.RestartRequestPhaseInProgress, "Restart in progress")
}

// setWorkloadPhaseToRestarting sets the workload phase to Restarting
func (r *RestartReconciler) setWorkloadPhaseToRestarting(ctx context.Context, workload *dromosv1alpha1.WorkloadRunner) error {
	if workload.Status.Phase == dromosv1alpha1.WorkloadRunnerPhaseRestarting {
		return nil
	}
	workload.Status.Phase = dromosv1alpha1.WorkloadRunnerPhaseRestarting
	return r.Status().Update(ctx, workload)
}

// applyRestart apply the overrides and deletes the child
func (r *RestartReconciler) applyRestart(ctx context.Context, workload *dromosv1alpha1.WorkloadRunner, restartRequest *runv1alpha1.RestartRequest) (bool, error) {
	patchedWorkload, err := r.applySpecOverrides(ctx, restartRequest, workload)
	if err != nil {
		return false, err
	}

	childDeleted, deleteChildErr := r.deleteChild(ctx, workload)
	if deleteChildErr != nil {
		logger.FromContext(ctx).Info("failed to delete child workload object",
			"workload", workload.Name, "namespace", workload.Namespace, "error", deleteChildErr)
		return childDeleted, deleteChildErr
	}

	if patchedWorkload != nil {
		err := r.Patch(ctx, patchedWorkload, client.MergeFromWithOptions(workload, client.MergeFromWithOptimisticLock{}))
		if err != nil {
			return childDeleted, fmt.Errorf("apply spec overrides after child delete: %w", err)
		}
	}

	return childDeleted, nil
}

// applySpecOverrides computes the patched workload and dry-runs it
func (r *RestartReconciler) applySpecOverrides(ctx context.Context, restartRequest *runv1alpha1.RestartRequest, workload *dromosv1alpha1.WorkloadRunner) (*dromosv1alpha1.WorkloadRunner, error) {
	overrides := restartRequest.Spec.SpecOverrides
	if len(overrides) == 0 {
		logger.FromContext(ctx).Info("no spec overrides to apply")
		return nil, nil
	}

	patchedWorkload, err := applyOverrides(workload, overrides)
	if err != nil {
		return nil, fmt.Errorf("apply spec overrides: %w", err)
	}

	if err = r.Patch(ctx, patchedWorkload, client.MergeFrom(workload),
		&client.PatchOptions{DryRun: []string{metav1.DryRunAll}}); err != nil {
		return nil, fmt.Errorf("spec overrides dry-run update failed: %w", err)
	}
	return patchedWorkload, nil
}

// applyOverrides applies the RFC 6902 overrides to the whole workload object and returns the result.
func applyOverrides(workload *dromosv1alpha1.WorkloadRunner, overrides runv1alpha1.SpecOverrides) (*dromosv1alpha1.WorkloadRunner, error) {
	// if the WLR spec.instructions is nil initialize empty instructions
	// otherwise the patch will fail overrides for path (add /spec/instructions/nodeAffinity)
	if workload.Spec.Instructions == nil && overridesTargetInstructions(overrides) {
		workload = workload.DeepCopy()
		workload.Spec.Instructions = &dromosv1alpha1.Instructions{}
	}
	workloadJSON, err := json.Marshal(workload)
	if err != nil {
		return nil, fmt.Errorf("marshal workload: %w", err)
	}
	patchJSON, err := json.Marshal(overrides)
	if err != nil {
		return nil, fmt.Errorf("marshal spec overrides: %w", err)
	}
	patchJSON, err = unwrapSpecOverrideValues(patchJSON)
	if err != nil {
		return nil, fmt.Errorf("unwrap spec override values: %w", err)
	}
	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		return nil, fmt.Errorf("decode patch: %w", err)
	}
	patched, err := patch.Apply(workloadJSON)
	if err != nil {
		return nil, fmt.Errorf("apply patch: %w", err)
	}
	patchedWorkload := workload.DeepCopy()
	if err = json.Unmarshal(patched, patchedWorkload); err != nil {
		return nil, fmt.Errorf("unmarshal patched workload: %w", err)
	}
	return patchedWorkload, nil
}

// overridesTargetInstructions check if path override under spec.instructions
func overridesTargetInstructions(overrides runv1alpha1.SpecOverrides) bool {
	for _, o := range overrides {
		if strings.HasPrefix(o.Path, "/spec/instructions/") {
			return true
		}
	}
	return false
}

// unwrapSpecOverrideValues rewrites patch JSON so any operation value that is an object
// with exactly one key "value" is replaced by that inner value. This allows CRD-compliant
// object values to still apply as scalars/arrays per RFC 6902.
func unwrapSpecOverrideValues(patchJSON []byte) ([]byte, error) {
	var ops []map[string]any
	if err := json.Unmarshal(patchJSON, &ops); err != nil {
		return nil, err
	}
	for i := range ops {
		v, ok := ops[i]["value"]
		if !ok {
			continue
		}
		m, ok := v.(map[string]any)
		if !ok || len(m) != 1 {
			continue
		}
		inner, ok := m["value"]
		if !ok {
			continue
		}
		ops[i]["value"] = inner
	}
	return json.Marshal(ops)
}

// childObject returns a handle on the object renders from spec.template
func childObject(workload *dromosv1alpha1.WorkloadRunner) (*unstructured.Unstructured, error) {
	child, err := template.NewParser().Parse(workload.Spec.Template)
	if err != nil {
		return nil, fmt.Errorf("parse workload template: %w", err)
	}
	child.SetNamespace(workload.Namespace)
	return child, nil
}

// deleteChild deletes the child workload object
func (r *RestartReconciler) deleteChild(ctx context.Context, workload *dromosv1alpha1.WorkloadRunner) (bool, error) {
	child, err := childObject(workload)
	if err != nil {
		return false, err
	}
	if err = r.Delete(ctx, child); err != nil {
		if k8sErrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}
	return true, nil
}

// isChildGone reports whether the child object is fully deleted from the API server
func (r *RestartReconciler) isChildGone(ctx context.Context, workload *dromosv1alpha1.WorkloadRunner) (bool, *metav1.Time, error) {
	child, err := childObject(workload)
	if err != nil {
		return false, nil, err
	}
	found := &unstructured.Unstructured{}
	found.SetGroupVersionKind(child.GroupVersionKind())
	if err = r.Get(ctx, client.ObjectKeyFromObject(child), found); err != nil {
		if k8sErrors.IsNotFound(err) {
			return true, nil, nil
		}
		return false, nil, err
	}
	return false, found.GetDeletionTimestamp(), nil
}

// finalizeRestart waits for the child object to be fully deleted
func (r *RestartReconciler) finalizeRestart(ctx context.Context, workload *dromosv1alpha1.WorkloadRunner,
	restartRequest *runv1alpha1.RestartRequest, frameworkDeleted bool, deletedErr error) (ctrl.Result, error) {
	if deletedErr != nil {
		return ctrl.Result{}, r.updateStatus(ctx, restartRequest, frameworkDeleted, deletedErr)
	}

	gone, deletionTimestamp, err := r.isChildGone(ctx, workload)
	if err != nil {
		return ctrl.Result{}, err
	}
	if gone {
		return ctrl.Result{}, r.updateStatus(ctx, restartRequest, frameworkDeleted, nil)
	}

	if deletionTimestamp != nil && time.Since(deletionTimestamp.Time) > r.childDeletionTimeout {
		logger.FromContext(ctx).Info("child workload object deletion timed out", "timeout", r.childDeletionTimeout)
		return ctrl.Result{}, r.updateStatus(ctx, restartRequest, frameworkDeleted,
			fmt.Errorf("child deletion timed out after %s", r.childDeletionTimeout))
	}

	logger.FromContext(ctx).Info("child workload object not yet fully deleted, requeuing")
	return ctrl.Result{RequeueAfter: childDeletionRequeueInterval}, nil
}

// updateStatus writes the framework-clear condition.
func (r *RestartReconciler) updateStatus(ctx context.Context, restartRequest *runv1alpha1.RestartRequest, frameworkDeleted bool, deletedErr error) error {
	setFrameworkClearCondition(restartRequest, frameworkDeleted, deletedErr)
	if deletedErr != nil {
		return updateRestartRequestPhase(ctx, r.Client, restartRequest, runv1alpha1.RestartRequestPhaseFailed, deletedErr.Error())
	}
	return updateRestartRequestPhase(ctx, r.Client, restartRequest, runv1alpha1.RestartRequestPhaseSucceeded, "Restart completed successfully")
}

// setFrameworkClearCondition records the outcome of the deletion.
func setFrameworkClearCondition(restartRequest *runv1alpha1.RestartRequest, frameworkDeleted bool, deleteErr error) {
	var condition metav1.Condition
	switch {
	case deleteErr != nil:
		condition = metav1.Condition{
			Type:               frameworkClearConditionType,
			Status:             metav1.ConditionFalse,
			Reason:             "Failed",
			Message:            deleteErr.Error(),
			LastTransitionTime: metav1.Now(),
		}
	case frameworkDeleted:
		condition = metav1.Condition{
			Type:               frameworkClearConditionType,
			Status:             metav1.ConditionTrue,
			Reason:             "Deleted",
			Message:            "Framework is cleared",
			LastTransitionTime: metav1.Now(),
		}
	default:
		condition = metav1.Condition{
			Type:               frameworkClearConditionType,
			Status:             metav1.ConditionTrue,
			Reason:             "NotFound",
			Message:            "No framework found to delete",
			LastTransitionTime: metav1.Now(),
		}
	}
	meta.SetStatusCondition(&restartRequest.Status.Conditions, condition)
}

// updateRestartRequestPhase sets the RestartRequest
func updateRestartRequestPhase(ctx context.Context, c client.Client, restartRequest *runv1alpha1.RestartRequest,
	phase runv1alpha1.RestartRequestPhase, message string) error {
	restartRequest.Status.Phase = phase

	condition := metav1.Condition{
		Type:               restartFailedConditionType,
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Message:            message,
	}

	switch phase {
	case runv1alpha1.RestartRequestPhaseSucceeded:
		condition.Reason = "Succeeded"
	case runv1alpha1.RestartRequestPhaseFailed:
		condition.Status = metav1.ConditionTrue
		condition.Reason = "Failed"
	case runv1alpha1.RestartRequestPhaseInProgress:
		condition.Reason = "InProgress"
	default:
		condition.Reason = "Pending"
	}

	meta.SetStatusCondition(&restartRequest.Status.Conditions, condition)

	return c.Status().Update(ctx, restartRequest)
}
