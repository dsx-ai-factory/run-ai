/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"errors"
	"fmt"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (c *WorkloadRunnerController) pluginStatusCheck(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, state *reconcile.State) error {
	logger := log.FromContext(ctx)

	if err := c.pluginFramework.RunResourcesStatus(ctx, runner, c.client, state); err != nil {
		return err
	}

	logger.Info("Successfully updated plugin status of WorkloadRunner", "namespace", runner.Namespace, "name", runner.Name)

	return nil
}

// reconcileStatusAndRunPostPhase computes and persists the WorkloadRunner status, then runs the post-phase plugins
func (c *WorkloadRunnerController) reconcileStatusAndRunPostPhase(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, specResult reconcile.StepResult, state *reconcile.State) reconcile.StepResult {
	logger := log.FromContext(ctx)

	// Check if prerequisites failed (Karta missing, Karta not valid, etc.)
	// If so, phase was already set by reconcileSpec - don't override it
	prereqCond := meta.FindStatusCondition(runner.Status.Conditions, dromosv1alpha1.ConditionTypePrerequisitesPassed)
	if prereqCond != nil && prereqCond.Status == metav1.ConditionFalse {
		// Prerequisites not met - return the same result from reconcileSpec
		return specResult
	}

	pscErr := c.pluginStatusCheck(ctx, runner, state)

	reconcile.SetControllerIssues(ctx, runner, state)
	attachedResource, err := c.getAttachedResource(ctx, runner)
	if err != nil {
		_ = c.setFailedStatusForUnknownResourceState(ctx, runner, fmt.Errorf("failed to get attached resource: %w", err))
		return reconcile.WithError(err)
	}

	gvk := attachedResource.GroupVersionKind()
	karta, err := c.kartaUtils.GetKartaForGVK(ctx, gvk)
	if err != nil {
		_ = c.setFailedStatusForUnknownResourceState(ctx, runner, fmt.Errorf("failed to get karta for GVK %s: %w", gvk.String(), err))
		return reconcile.WithError(err)
	}

	setLastReconcileCondition(runner, specResult.Err)

	resourcePhase, err := c.calcResourcePhase(ctx, karta, attachedResource)
	if err != nil {
		_ = c.setFailedStatusForUnknownResourceState(ctx, runner, err)
		return reconcile.WithError(err)
	}
	runner.Status.ResourcePhase = resourcePhase

	phase := c.calcPhase(karta, attachedResource, resourcePhase, runner.Status.Plugins, state.HasCriticalControllerIssues())
	runner.Status.Phase = phase

	if err := c.applyStatus(ctx, runner); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(
				err,
				"WorkloadRunner not found during status update",
				"namespace", runner.Namespace,
				"name", runner.Name,
			)

			return reconcile.DoNotRequeue()
		}

		logger.Error(err, "Failed to update status of WorkloadRunner", "namespace", runner.Namespace, "name", runner.Name)

		return reconcile.WithError(err)
	}

	if pscErr != nil {
		logger.Error(pscErr, "Failed to check plugin status", "namespace", runner.Namespace, "name", runner.Name)
		return reconcile.WithError(pscErr)
	}

	logger.Info("Successfully updated status of WorkloadRunner", "namespace", runner.Namespace, "name", runner.Name)

	if err := c.pluginFramework.RunPostPhase(ctx, runner, attachedResource, karta, c.client); err != nil {
		logger.Error(err, "Failed to run post-phase plugins", "namespace", runner.Namespace, "name", runner.Name)
		return reconcile.WithError(fmt.Errorf("run post-phase: %w", err))
	}

	return reconcile.Continue()
}

func (c *WorkloadRunnerController) setFailedStatusForUnknownResourceState(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, err error) error {
	runner.Status.ResourcePhase = dromosv1alpha1.WorkloadRunnerPhaseUnknown
	runner.Status.Phase = dromosv1alpha1.WorkloadRunnerPhaseFailed

	setLastReconcileCondition(runner, err)

	return c.applyStatus(ctx, runner)
}

const statusFieldManager = "anyworkload"

func (c *WorkloadRunnerController) applyStatus(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner) error {
	statusObj := &dromosv1alpha1.WorkloadRunner{
		TypeMeta: metav1.TypeMeta{
			APIVersion: dromosv1alpha1.GroupVersion.String(),
			Kind:       dromosv1alpha1.WorkloadRunnerKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      runner.Name,
			Namespace: runner.Namespace,
		},
		Status: runner.Status,
	}
	return c.client.Status().Patch(ctx, statusObj, client.Apply, client.ForceOwnership, client.FieldOwner(statusFieldManager))
}

func setLastReconcileCondition(runner *dromosv1alpha1.WorkloadRunner, err error) {
	if err != nil {
		meta.SetStatusCondition(&runner.Status.Conditions, metav1.Condition{
			Type:    dromosv1alpha1.LastReconcileSucceeded,
			Status:  metav1.ConditionFalse,
			Reason:  dromosv1alpha1.ReasonReconcileFailed,
			Message: err.Error(),
		})
	} else {
		meta.SetStatusCondition(&runner.Status.Conditions, metav1.Condition{
			Type:    dromosv1alpha1.LastReconcileSucceeded,
			Status:  metav1.ConditionTrue,
			Reason:  dromosv1alpha1.ReasonSucceeded,
			Message: "",
		})
	}
}

func (c *WorkloadRunnerController) calcResourcePhase(ctx context.Context, karta *kartav1alpha1.Karta, resource *unstructured.Unstructured) (dromosv1alpha1.WorkloadRunnerPhase, error) {
	if karta == nil {
		return dromosv1alpha1.WorkloadRunnerPhaseUnknown, errors.New("karta is nil")
	}
	if resource == nil {
		return dromosv1alpha1.WorkloadRunnerPhaseUnknown, errors.New("attached resource is nil")
	}

	phase, err := c.kartaUtils.GetResourcePhase(ctx, karta, resource)
	if err != nil {
		return dromosv1alpha1.WorkloadRunnerPhaseUnknown, fmt.Errorf("failed to get resource phase: %w", err)
	}

	return phase, nil
}

// pluginAggregate represents the combined health of all active plugins.
type pluginAggregate int

const (
	pluginAggregateReady pluginAggregate = iota
	pluginAggregateNotReady
	pluginAggregateFailed
)

// aggregatePluginPhase computes the worst-case aggregate of all active (non-Disabled) plugin phases.
func aggregatePluginPhase(plugins map[string]*dromosv1alpha1.PluginStatus) pluginAggregate {
	result := pluginAggregateReady
	for _, ps := range plugins {
		if ps == nil {
			continue
		}
		switch ps.Phase {
		case dromosv1alpha1.PluginPhaseFailed:
			return pluginAggregateFailed
		case dromosv1alpha1.PluginPhaseCreating,
			dromosv1alpha1.PluginPhasePending,
			dromosv1alpha1.PluginPhaseInitializing,
			dromosv1alpha1.PluginPhaseUnknown:
			result = pluginAggregateNotReady
		}
	}
	return result
}

func (c *WorkloadRunnerController) calcPhase(karta *kartav1alpha1.Karta, resource *unstructured.Unstructured,
	resourcePhase dromosv1alpha1.WorkloadRunnerPhase, plugins map[string]*dromosv1alpha1.PluginStatus, controllerHasCriticalIssues bool) dromosv1alpha1.WorkloadRunnerPhase {
	if karta == nil {
		return dromosv1alpha1.WorkloadRunnerPhaseUnknown
	}
	if resource == nil {
		return dromosv1alpha1.WorkloadRunnerPhaseUnknown
	}

	// Terminal resource phases take precedence
	switch resourcePhase {
	case dromosv1alpha1.WorkloadRunnerPhaseFailed,
		dromosv1alpha1.WorkloadRunnerPhaseCompleted,
		dromosv1alpha1.WorkloadRunnerPhaseSuspended,
		dromosv1alpha1.WorkloadRunnerPhaseStopped:
		return resourcePhase
	}

	// Critical controller issues → Degraded
	if controllerHasCriticalIssues {
		return dromosv1alpha1.WorkloadRunnerPhaseDegraded
	}

	agg := aggregatePluginPhase(plugins)
	switch agg {
	case pluginAggregateFailed:
		return dromosv1alpha1.WorkloadRunnerPhaseDegraded
	case pluginAggregateNotReady:
		if resourcePhase == dromosv1alpha1.WorkloadRunnerPhaseRunning {
			return dromosv1alpha1.WorkloadRunnerPhaseInitializing
		}
		return resourcePhase
	default: // pluginAggregateReady
		return resourcePhase
	}
}
