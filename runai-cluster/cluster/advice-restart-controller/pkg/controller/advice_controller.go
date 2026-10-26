/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaicommon "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"

	"github.com/dsx-ai-factory/nvidia-recovery/advice-operator/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/advice-restart-controller/internal/workloadtype"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

const requeueDelay = 30 * time.Second

// nonRestartablePhases are phases where restart should not be triggered
var nonRestartablePhases = map[string]struct{}{
	string(dromosv1alpha1.WorkloadRunnerPhaseSuspended):  {},
	string(dromosv1alpha1.WorkloadRunnerPhaseSuspending): {},
	string(dromosv1alpha1.WorkloadRunnerPhaseCompleted):  {},
	string(dromosv1alpha1.WorkloadRunnerPhaseStopped):    {},
}

const phaseRestarting = string(dromosv1alpha1.WorkloadRunnerPhaseRestarting)

type AdviceController struct {
	client              client.Client
	recorder            record.EventRecorder
	mode                string
	restartLimit        int
	maxRetriesPerAdvice int
	handlerFactories    map[schema.GroupVersionKind]workloadtype.HandlerFactory
}

func NewAdviceController(c client.Client, recorder record.EventRecorder, mode string, restartLimit, maxRetriesPerAdvice int) *AdviceController {
	if mode != ModeLog && mode != ModeAction {
		panic(fmt.Sprintf("invalid mode %q: must be %q or %q", mode, ModeLog, ModeAction))
	}
	return &AdviceController{
		client:              c,
		recorder:            recorder,
		mode:                mode,
		restartLimit:        restartLimit,
		maxRetriesPerAdvice: maxRetriesPerAdvice,
		handlerFactories: map[schema.GroupVersionKind]workloadtype.HandlerFactory{
			workloadtype.GVKWorkloadRunner: workloadtype.NewWorkloadRunnerHandler,
		},
	}
}

func (c *AdviceController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	advice := &v1alpha1.Advice{}
	if err := c.client.Get(ctx, req.NamespacedName, advice); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !isRestartAdvice(advice) {
		logger.V(1).Info("ignoring advice, state is not RESTARTED",
			"state", advice.Spec.Advice.State)
		return ctrl.Result{}, nil
	}

	jobRef := advice.Spec.Job
	if jobRef.Name == "" || jobRef.APIVersion == "" || jobRef.Kind == "" {
		logger.V(0).Info("WARNING: advice missing job name/apiVersion/kind, skipping",
			"apiVersion", jobRef.APIVersion, "kind", jobRef.Kind, "jobName", jobRef.Name)
		return ctrl.Result{}, nil
	}

	result, err := c.checkExistingRestartRequests(ctx, advice)
	switch {
	case err != nil:
		return ctrl.Result{}, err
	case result != nil:
		return *result, nil
	}

	topOwner, err := c.findTopOwner(ctx, advice)
	if err != nil {
		return ctrl.Result{}, err
	}
	if topOwner == nil {
		logger.V(0).Info("WARNING: job not found for advice, skipping", "jobName", jobRef.Name)
		return ctrl.Result{}, nil
	}

	topOwnerGVK := topOwner.GroupVersionKind()
	factory, ok := c.handlerFactories[topOwnerGVK]
	if !ok {
		logger.Info("unsupported workload type, skipping", "gvk", topOwnerGVK.String())
		return ctrl.Result{}, nil
	}

	handler, err := factory(topOwner)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("create handler for %s/%s: %w", topOwner.GetNamespace(), topOwner.GetName(), err)
	}

	result, err = c.checkWorkloadEligibility(ctx, advice, topOwner, handler)
	switch {
	case err != nil:
		logger.Error(err, "checking workload eligibility", "workload", topOwner.GetName())
		return ctrl.Result{}, err
	case result != nil:
		return *result, nil
	}

	var nodeExclusions []string
	if advice.Spec.Advice.NodeConstraints != nil {
		nodeExclusions = advice.Spec.Advice.NodeConstraints.Exclusions
	}

	c.emitAdviceReceivedOnWorkloadEvent(advice, topOwner, nodeExclusions)

	if err := c.createRestartRequest(ctx, advice, topOwner, nodeExclusions, handler); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// checkExistingRestartRequests checks for existing RestartRequests linked to this advice.
// Returns a non-nil result to skip/requeue, or nil to proceed with creating a new RR.
func (c *AdviceController) checkExistingRestartRequests(ctx context.Context, advice *v1alpha1.Advice) (*ctrl.Result, error) {
	logger := log.FromContext(ctx)

	rrList := &runv1alpha1.RestartRequestList{}
	if err := c.client.List(ctx, rrList,
		client.InNamespace(advice.Namespace),
		client.MatchingLabels{LabelAdviceName: advice.Name}); err != nil {
		return nil, fmt.Errorf("failed to list RestartRequests for advice %s: %w", advice.Name, err)
	}

	failedCount := 0
	for i := range rrList.Items {
		rr := &rrList.Items[i]
		switch rr.Status.Phase {
		case runv1alpha1.RestartRequestPhaseSucceeded:
			logger.Info("advice already processed, RestartRequest succeeded",
				"restartRequest", rr.Name)
			return &ctrl.Result{}, nil

		case runv1alpha1.RestartRequestPhaseInProgress, runv1alpha1.RestartRequestPhasePending, "":
			logger.Info("RestartRequest still in progress, requeuing",
				"restartRequest", rr.Name, "phase", rr.Status.Phase)
			return &ctrl.Result{RequeueAfter: requeueDelay}, nil

		case runv1alpha1.RestartRequestPhaseFailed:
			failedCount++
		}
	}

	if failedCount > 0 && failedCount >= c.maxRetriesPerAdvice {
		logger.Info("retry limit reached for advice, giving up",
			"failedCount", failedCount, "maxRetries", c.maxRetriesPerAdvice)
		return &ctrl.Result{}, nil
	}
	if failedCount > 0 {
		logger.Info("previous RestartRequest(s) failed, retrying",
			"failedCount", failedCount, "maxRetries", c.maxRetriesPerAdvice)
	}

	return nil, nil
}

// findTopOwner finds the job referenced by the advice, walks the owner chain to the top-level
// workload, and returns it as unstructured. Returns nil if the job is not found.
func (c *AdviceController) findTopOwner(ctx context.Context, advice *v1alpha1.Advice) (*unstructured.Unstructured, error) {
	jobRef := advice.Spec.Job

	job := &unstructured.Unstructured{}
	job.SetGroupVersionKind(schema.FromAPIVersionAndKind(jobRef.APIVersion, jobRef.Kind))
	if err := c.client.Get(ctx, types.NamespacedName{Namespace: advice.Namespace, Name: jobRef.Name}, job); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get job %s for advice %s: %w", jobRef.Name, advice.Name, err)
	}

	topOwnerMeta, err := topowner.GetTopResourceOwner(ctx, job, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to find top owner for job %s: %w", job.GetName(), err)
	}

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(topOwnerMeta.GroupVersionKind())
	if err := c.client.Get(ctx, types.NamespacedName{
		Namespace: topOwnerMeta.Namespace,
		Name:      topOwnerMeta.Name,
	}, u); err != nil {
		return nil, fmt.Errorf("failed to fetch top owner %s/%s: %w", topOwnerMeta.Namespace, topOwnerMeta.Name, err)
	}

	return u, nil
}

// checkWorkloadEligibility determines whether to proceed with the advice.
func (c *AdviceController) checkWorkloadEligibility(ctx context.Context, advice *v1alpha1.Advice,
	topOwner *unstructured.Unstructured, handler workloadtype.WorkloadHandler) (*ctrl.Result, error) {
	logger := log.FromContext(ctx)

	phase := handler.GetPhase()

	if phase == phaseRestarting {
		logger.Info("workload is mid-restart, requeuing advice",
			"workload", topOwner.GetName(), "phase", phase)
		return &ctrl.Result{RequeueAfter: requeueDelay}, nil
	}

	if _, ok := nonRestartablePhases[phase]; ok {
		logger.Info("ignoring advice, workload in non-restartable state",
			"workload", topOwner.GetName(), "phase", phase)
		c.emitAdviceIgnoredNonRestartablePhaseEvent(advice, topOwner, phase)
		return &ctrl.Result{}, nil
	}

	// Count succeeded RestartRequests for this workload created by this controller
	rrList := &runv1alpha1.RestartRequestList{}
	if err := c.client.List(ctx, rrList,
		client.InNamespace(topOwner.GetNamespace()),
		client.MatchingLabels{
			constants.RunaiWorkloadNameLabel: topOwner.GetName(),
			LabelCreatedBy:                   CreatedByValue,
		}); err != nil {
		return nil, fmt.Errorf("failed to list RestartRequests for workload %s: %w", topOwner.GetName(), err)
	}

	succeededCount := 0
	for i := range rrList.Items {
		if rrList.Items[i].Status.Phase == runv1alpha1.RestartRequestPhaseSucceeded {
			succeededCount++
		}
	}

	if succeededCount >= c.restartLimit {
		logger.Info("restart threshold exceeded, skipping",
			"workload", topOwner.GetName(),
			"succeededRestarts", succeededCount, "restartLimit", c.restartLimit)
		c.emitAdviceIgnoredRestartLimitEvent(advice, topOwner, c.restartLimit)
		return &ctrl.Result{}, nil
	}

	return nil, nil
}

// createRestartRequest creates a RestartRequest for the advice, or logs what would happen in log mode.
func (c *AdviceController) createRestartRequest(ctx context.Context, advice *v1alpha1.Advice,
	topOwner *unstructured.Unstructured, nodeExclusions []string, handler workloadtype.WorkloadHandler) error {
	logger := log.FromContext(ctx)

	specOverrides, err := handler.BuildSpecOverrides(nodeExclusions)
	if err != nil {
		return fmt.Errorf("failed to build spec overrides: %w", err)
	}

	topOwnerGVK := topOwner.GetObjectKind().GroupVersionKind()
	rr := &runv1alpha1.RestartRequest{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("advice-%s-", advice.Name),
			Namespace:    advice.Namespace,
			Labels: map[string]string{
				LabelAdviceName:                     advice.Name,
				constants.RunaiWorkloadNameLabel:    topOwner.GetName(),
				LabelCreatedBy:                      CreatedByValue,
				runaicommon.TopOwnerWorkloadIdLabel: topowner.GetRunaiWorkloadId(topOwner.GetUID(), topOwner.GetAnnotations()),
			},
		},
		Spec: runv1alpha1.RestartRequestSpec{
			WorkloadRef: runv1alpha1.WorkloadReference{
				GVK: metav1.GroupVersionKind{
					Group:   topOwnerGVK.Group,
					Version: topOwnerGVK.Version,
					Kind:    topOwnerGVK.Kind,
				},
				Name:       topOwner.GetName(),
				Generation: topOwner.GetGeneration(),
			},
			SpecOverrides: specOverrides,
		},
	}

	if c.mode == ModeLog {
		logger.Info("LOG MODE: would create RestartRequest",
			"workload", topOwner.GetName(),
			"workloadRef", rr.Spec.WorkloadRef,
			"specOverrides", rr.Spec.SpecOverrides)
		c.emitAdviceProcessedLogOnlyEvent(advice, topOwner)
		return nil
	}

	if err = c.client.Create(ctx, rr); err != nil {
		return fmt.Errorf("failed to create RestartRequest for advice %s: %w", advice.Name, err)
	}

	logger.Info("RestartRequest created",
		"restartRequest", rr.Name,
		"workload", topOwner.GetName(), "nodeExclusions", nodeExclusions)

	c.emitRestartRequestCreatedEvent(advice, topOwner)

	return nil
}

// isRestartAdvice returns true when the advice recommends a restart (state == RESTARTED).
// Other states (e.g. RUNNING, COMPLETED) are not relevant for this controller and do not trigger restart logic.
func isRestartAdvice(advice *v1alpha1.Advice) bool {
	return advice.Spec.Advice.State != nil && *advice.Spec.Advice.State == v1alpha1.JobStateRestarted
}

func (c *AdviceController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Advice{}).
		Watches(
			&runv1alpha1.RestartRequest{},
			handler.EnqueueRequestsFromMapFunc(c.mapRestartRequestToAdvice),
			builder.WithPredicates(c.restartRequestPhasePredicate()),
		).
		Complete(c)
}

// mapRestartRequestToAdvice maps a RestartRequest back to its parent Advice for re-reconciliation.
func (c *AdviceController) mapRestartRequestToAdvice(_ context.Context, obj client.Object) []reconcile.Request {
	adviceName := obj.GetLabels()[LabelAdviceName]
	if adviceName == "" {
		return nil
	}
	return []reconcile.Request{{
		NamespacedName: types.NamespacedName{
			Name:      adviceName,
			Namespace: obj.GetNamespace(),
		},
	}}
}

// restartRequestPhasePredicate filters RestartRequest events to only phase transitions
// on RRs created by this controller.
func (c *AdviceController) restartRequestPhasePredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool { return false },
		DeleteFunc: func(e event.DeleteEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectNew.GetLabels()[LabelCreatedBy] != CreatedByValue {
				return false
			}
			oldRR, ok1 := e.ObjectOld.(*runv1alpha1.RestartRequest)
			newRR, ok2 := e.ObjectNew.(*runv1alpha1.RestartRequest)
			if !ok1 || !ok2 {
				return false
			}
			return oldRR.Status.Phase != newRR.Status.Phase
		},
	}
}
