/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package restart

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerUtil "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

// Protocol constants for the workload restart flow.
const (
	// RestartControllerManagedBy is the spec.managedBy marker set on a workload
	RestartControllerManagedBy = "restart-controller-workloadrunner"

	// RequestRefLabel holds the UID of the active RestartRequest on a workload
	RequestRefLabel = "run.ai/request-ref"

	// RestartRequestFinalizer blocks RestartRequest deletion
	RestartRequestFinalizer = "run.ai/restart-request-finalizer"
)

// Event reasons for the restart handoff reconciler.
const (
	restartInitiatedEventReason = "RestartInitiated"
	restartFailedEventReason    = "RestartFailed"
	restartSucceededEventReason = "RestartSucceeded"
	workloadNotFoundEventReason = "WorkloadNotFound"

	restartFailedConditionType = "RestartFailed"
)

var (
	handoffLogger                = logger.Log.WithName("reconcile.handoff")
	errWorkloadAlreadyRestarting = errors.New("workload is already being restarted by another request")
	errWorkloadAlreadyManaged    = errors.New("workload is already managed by another controller")
)

// HandoffReconciler owns the RestartRequest lifecycle
type HandoffReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder
	name          string
}

// NewHandoffReconciler creates a new HandoffReconciler.
func NewHandoffReconciler(client client.Client, eventRecorder record.EventRecorder) *HandoffReconciler {
	return &HandoffReconciler{
		Client:        client,
		Scheme:        client.Scheme(),
		EventRecorder: eventRecorder,
		name:          "restart-handoff",
	}
}

// WithName overrides the controller name (used by tests running multiple managers).
func (r *HandoffReconciler) WithName(name string) *HandoffReconciler {
	r.name = name
	return r
}

// SetupWithManager sets up the controller with the Manager.
func (r *HandoffReconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconciles int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&runv1alpha1.RestartRequest{}, builder.WithPredicates(knownDeletionPredicate())).
		Watches(
			&dromosv1alpha1.WorkloadRunner{},
			handler.EnqueueRequestsFromMapFunc(r.mapWorkloadToRestartRequests),
			builder.WithPredicates(workloadDeletionPredicate()),
		).
		Named(r.name).
		WithOptions(controllerUtil.Options{MaxConcurrentReconciles: maxConcurrentReconciles,
			RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
				constants.RateLimiterBaseDelay, constants.RateLimiterMaxDelay),
		}).
		Complete(r)
}

// knownDeletionPredicate drops delete events whose final state is unknown.
func knownDeletionPredicate() predicate.Predicate {
	return predicate.Funcs{
		DeleteFunc: func(e event.DeleteEvent) bool {
			return !e.DeleteStateUnknown
		},
	}
}

// workloadDeletionPredicate exist so that reconciler can re-enqueue RestartRequests whose owning workload is gone
func workloadDeletionPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		UpdateFunc:  func(e event.UpdateEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return !e.DeleteStateUnknown },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
}

func (r *HandoffReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	trxLogger := handoffLogger.WithValues("id", fmt.Sprintf("%s:%s", req.Namespace, req.Name))
	ctx = logger.IntoContext(ctx, trxLogger)

	trxLogger.Info("reconcile started")
	defer trxLogger.Info("reconcile ended")

	restartRequest := &runv1alpha1.RestartRequest{}
	err := r.Get(ctx, req.NamespacedName, restartRequest)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			trxLogger.Info("restart request has been deleted")
			return ctrl.Result{}, nil
		}
		trxLogger.Error(err, "failed to load the restart request")
		return ctrl.Result{}, err
	}

	if !restartRequest.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, restartRequest)
	}

	if !controllerutil.ContainsFinalizer(restartRequest, RestartRequestFinalizer) {
		base := restartRequest.DeepCopy()
		controllerutil.AddFinalizer(restartRequest, RestartRequestFinalizer)
		if err = r.Patch(ctx, restartRequest, client.MergeFrom(base)); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	workloadRef := &restartRequest.Spec.WorkloadRef

	if r.checkIfCompleted(ctx, restartRequest) {
		workload, err := r.getWorkload(ctx, workloadRef, restartRequest.Namespace)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				trxLogger.Info("terminal restart request, workload already gone - nothing to hand back")
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
		// verify ownership of the workload before cleaning up the label and managedBy
		if ref, ok := workload.Labels[RequestRefLabel]; ok && ref != string(restartRequest.UID) {
			trxLogger.Info("workload is locked by another restart request - nothing to hand back",
				"workload", workload.Name, "existingRef", ref)
			return ctrl.Result{}, nil
		}
		r.emitRestartOutcomeEvent(workload, restartRequest)
		if err = r.finalizeHandoff(ctx, workload); err != nil {
			trxLogger.Error(err, "failed to cleanup workload after restart")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	workload, err := r.getWorkload(ctx, workloadRef, restartRequest.Namespace)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			trxLogger.Error(err, "referenced workload not found", "workload", workloadRef.Name)
			notFoundMsg := fmt.Sprintf("Referenced workload %s/%s not found", workloadRef.GVK.Kind, workloadRef.Name)
			r.EventRecorder.Event(restartRequest, corev1.EventTypeWarning, workloadNotFoundEventReason, notFoundMsg)
			if updateErr := updateRestartRequestPhase(ctx, r.Client, restartRequest, runv1alpha1.RestartRequestPhaseFailed, notFoundMsg); updateErr != nil {
				return ctrl.Result{}, updateErr
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if r.checkIfPending(ctx, restartRequest) {
		if err = updateRestartRequestPhase(ctx, r.Client, restartRequest, runv1alpha1.RestartRequestPhaseInProgress, "Restart request is being processed"); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err = r.updateWorkloadForRestart(ctx, workload, restartRequest); err != nil {
		trxLogger.Error(err, "failed to update workload for restart")
		if errors.Is(err, errWorkloadAlreadyRestarting) || errors.Is(err, errWorkloadAlreadyManaged) ||
			k8sErrors.IsInvalid(err) || k8sErrors.IsForbidden(err) {
			if updateErr := updateRestartRequestPhase(ctx, r.Client, restartRequest, runv1alpha1.RestartRequestPhaseFailed, err.Error()); updateErr != nil {
				return ctrl.Result{}, updateErr
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	trxLogger.Info("reconcile completed successfully")

	return ctrl.Result{}, nil
}

func (r *HandoffReconciler) checkIfCompleted(ctx context.Context, restartRequest *runv1alpha1.RestartRequest) bool {
	if restartRequest.Status.Phase == runv1alpha1.RestartRequestPhaseSucceeded ||
		restartRequest.Status.Phase == runv1alpha1.RestartRequestPhaseFailed {
		logger.FromContext(ctx).Info("restart request already completed", "phase", restartRequest.Status.Phase)
		return true
	}
	return false
}

func (r *HandoffReconciler) checkIfPending(ctx context.Context, restartRequest *runv1alpha1.RestartRequest) bool {
	if restartRequest.Status.Phase == "" || restartRequest.Status.Phase == runv1alpha1.RestartRequestPhasePending {
		logger.FromContext(ctx).Info("restart request is pending", "phase", restartRequest.Status.Phase)
		return true
	}
	return false
}

func (r *HandoffReconciler) getWorkload(ctx context.Context, workloadRef *runv1alpha1.WorkloadReference, restartRequestNamespace string) (*dromosv1alpha1.WorkloadRunner, error) {
	workload := &dromosv1alpha1.WorkloadRunner{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      workloadRef.Name,
		Namespace: restartRequestNamespace,
	}, workload)

	if err != nil {
		return nil, err
	}
	return workload, nil
}

// updateWorkloadForRestart updates the workload with specific restart request
func (r *HandoffReconciler) updateWorkloadForRestart(ctx context.Context, workload *dromosv1alpha1.WorkloadRunner, restartRequest *runv1alpha1.RestartRequest) error {
	trxLogger := logger.FromContext(ctx)

	if workload.Labels == nil {
		workload.Labels = make(map[string]string)
	}

	// Refuse if workload is already being restarted by another request
	existingRef, exists := workload.Labels[RequestRefLabel]
	if exists {
		if existingRef != string(restartRequest.UID) {
			return fmt.Errorf("%w (existing ref: %s)", errWorkloadAlreadyRestarting, existingRef)
		}
		trxLogger.Info("workload already set with restarted UID", "uid", restartRequest.UID, "workload", workload.Name)
		return nil
	}

	// no label means nobody restart-owns it, so a foreign managedBy is a delegation we must not overwrite.
	// our own value without a label is an orphan awaiting the restart reconciler's self-heal - let it be acquired
	if workload.Spec.ManagedBy != "" && workload.Spec.ManagedBy != RestartControllerManagedBy {
		return fmt.Errorf("managedBy %q: %w", workload.Spec.ManagedBy, errWorkloadAlreadyManaged)
	}

	workload.Labels[RequestRefLabel] = string(restartRequest.UID)
	workload.Spec.ManagedBy = RestartControllerManagedBy

	trxLogger.Info("updating workload with restart label and managedBy",
		"restartUID", restartRequest.UID,
		"workload", fmt.Sprintf("%s/%s", workload.Namespace, workload.Name))

	if err := r.Update(ctx, workload); err != nil {
		return fmt.Errorf("failed to update workload: %w", err)
	}

	base := restartRequest.DeepCopy()
	if err := controllerutil.SetOwnerReference(workload, restartRequest, r.Scheme); err != nil {
		return fmt.Errorf("failed to set workload as owner of restart request: %w", err)
	}

	if err := r.Patch(ctx, restartRequest, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("failed to patch restart request with owner reference: %w", err)
	}

	r.EventRecorder.Event(workload, corev1.EventTypeNormal, restartInitiatedEventReason, "Workload restart initiated")
	return nil
}

// emitRestartOutcomeEvent emits restart outcome
func (r *HandoffReconciler) emitRestartOutcomeEvent(workload *dromosv1alpha1.WorkloadRunner, restartRequest *runv1alpha1.RestartRequest) {
	if restartRequest.Status.Phase == runv1alpha1.RestartRequestPhaseSucceeded {
		r.EventRecorder.Event(workload, corev1.EventTypeNormal, restartSucceededEventReason, "Workload restarted successfully")
		return
	}
	failMsg := "Workload restart failed"
	if cond := meta.FindStatusCondition(restartRequest.Status.Conditions, restartFailedConditionType); cond != nil && cond.Message != "" {
		failMsg += ": " + cond.Message
	}
	r.EventRecorder.Event(workload, corev1.EventTypeWarning, restartFailedEventReason, failMsg)
}

// handleDeletion handles the case where a RestartRequest has been marked for deletion
func (r *HandoffReconciler) handleDeletion(ctx context.Context, restartRequest *runv1alpha1.RestartRequest) (ctrl.Result, error) {
	trxLogger := logger.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(restartRequest, RestartRequestFinalizer) {
		return ctrl.Result{}, nil
	}

	workload, err := r.getWorkload(ctx, &restartRequest.Spec.WorkloadRef, restartRequest.Namespace)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			trxLogger.Info("workload not found, removing restart request finalizer")
			return ctrl.Result{}, r.removeFinalizer(ctx, restartRequest)
		}
		return ctrl.Result{}, err
	}

	if !workload.DeletionTimestamp.IsZero() {
		trxLogger.Info("workload is being deleted, removing restart request finalizer")
		return ctrl.Result{}, r.removeFinalizer(ctx, restartRequest)
	}

	trxLogger.Info("workload still exists, keeping restart request finalizer")
	return ctrl.Result{}, nil
}

func (r *HandoffReconciler) removeFinalizer(ctx context.Context, restartRequest *runv1alpha1.RestartRequest) error {
	base := restartRequest.DeepCopy()
	controllerutil.RemoveFinalizer(restartRequest, RestartRequestFinalizer)
	return r.Patch(ctx, restartRequest, client.MergeFrom(base))
}

// mapWorkloadToRestartRequests maps a deleted WorkloadRunner to the RestartRequests
func (r *HandoffReconciler) mapWorkloadToRestartRequests(ctx context.Context, obj client.Object) []reconcile.Request {
	rrList := &runv1alpha1.RestartRequestList{}
	if err := r.List(ctx, rrList,
		client.InNamespace(obj.GetNamespace()),
		client.MatchingLabels{constants.RunaiWorkloadNameLabel: obj.GetName()}); err != nil {
		handoffLogger.Error(err, "failed to list restart requests for deleted workload", "workload", obj.GetName())
		return nil
	}
	var requests []reconcile.Request
	for i := range rrList.Items {
		rr := &rrList.Items[i]
		// If not yet terminating, explicitly delete the RR so handleDeletion can remove the finalizer
		if rr.DeletionTimestamp.IsZero() {
			if err := r.Delete(ctx, rr); err != nil && !k8sErrors.IsNotFound(err) {
				handoffLogger.Error(err, "failed to delete restart request for deleted workload",
					"restartRequest", rr.Name, "workload", obj.GetName())
			}
		}
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: rr.Name, Namespace: rr.Namespace},
		})
	}
	return requests
}

// finalizeHandoff clears the restart metadata and give-up control workload
func (r *HandoffReconciler) finalizeHandoff(ctx context.Context, workload *dromosv1alpha1.WorkloadRunner) error {
	trxLogger := logger.FromContext(ctx)

	needsUpdate := false

	if workload.Labels != nil {
		if _, exists := workload.Labels[RequestRefLabel]; exists {
			delete(workload.Labels, RequestRefLabel)
			needsUpdate = true
		}
	}

	if workload.Spec.ManagedBy == RestartControllerManagedBy {
		workload.Spec.ManagedBy = ""
		needsUpdate = true
	}

	if !needsUpdate {
		return nil
	}

	trxLogger.Info("cleaning up workload after restart completion",
		"workload", fmt.Sprintf("%s/%s", workload.Namespace, workload.Name))

	if err := r.Update(ctx, workload); err != nil {
		return fmt.Errorf("failed to cleanup workload: %w", err)
	}

	return nil
}
