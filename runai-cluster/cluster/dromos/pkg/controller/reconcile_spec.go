/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	controllerutils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/controller"
	kartautils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/karta"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
)

func (c *WorkloadRunnerController) reconcileSpec(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	state *reconcile.State,
) reconcile.StepResult {
	logger := log.FromContext(ctx)

	template, err := c.templateParser.Parse(runner.Spec.Template)
	if err != nil {
		logger.Error(err, "Failed to parse template for WorkloadRunner", "namespace", runner.Namespace, "name", runner.Name)
		state.AddControllerIssue(runner, err, true, "failed to parse template for WorkloadRunner", issue.Internal)
		return reconcile.WithError(fmt.Errorf("failed to parse template: %w", err))
	}

	gvk := template.GroupVersionKind()
	logger.Info(
		"Parsed template with GVK for WorkloadRunner",
		"gvk", gvk.String(),
		"namespace", runner.Namespace,
		"name", runner.Name,
	)

	karta, err := c.kartaUtils.GetKartaForGVK(ctx, gvk)
	if err != nil {
		if apierrors.IsNotFound(err) {
			kartaNotFoundMsg := fmt.Sprintf("Karta for GVK %s does not exist", gvk.String())
			state.AddControllerIssue(runner, err, true, kartaNotFoundMsg, issue.Internal)
			c.setPrerequisitesPassedCondition(
				runner,
				dromosv1alpha1.ReasonKartaNotFound,
				kartaNotFoundMsg,
			)
			runner.Status.Phase = dromosv1alpha1.WorkloadRunnerPhaseUnknown
			runner.Status.ResourcePhase = dromosv1alpha1.WorkloadRunnerPhaseUnknown
			c.recorder.Event(runner, corev1.EventTypeWarning, "KartaCRNotFound",
				kartaNotFoundMsg)
			reconcile.SetControllerIssues(ctx, runner, state)
			if err := c.applyStatus(ctx, runner); err != nil {
				return reconcile.WithError(err)
			}
			// In case of missing Karta, there is no need to requeue every x minutes.
			// we are watching Karta cr, the runner will reconcile when we detect a change.
			return reconcile.DoNotRequeue()
		}
		state.AddControllerIssue(runner, err, true, "failed to get Karta for GVK", issue.Internal)
		return reconcile.WithError(fmt.Errorf("failed to get Karta for GVK: %s: %w", gvk.String(), err))
	}

	if isValid, err := c.validateAttachedGVK(ctx, runner, gvk, karta); err != nil {
		state.AddControllerIssue(runner, err, true, "error validating attached GVK", issue.Internal)
		return reconcile.WithError(err)
	} else if !isValid {
		logger.V(2).Info( //nolint:mnd // log verbosity level
			"Warning: Attached GVK / its Karta is not valid, skipping reconciliation",
			"gvk", gvk.String(),
			"namespace", runner.Namespace,
			"name", runner.Name,
		)

		state.AddControllerIssue(runner, err, true, "Karta is not valid", issue.Internal)
		reconcile.SetControllerIssues(ctx, runner, state)
		runner.Status.Phase = dromosv1alpha1.WorkloadRunnerPhaseUnknown
		runner.Status.ResourcePhase = dromosv1alpha1.WorkloadRunnerPhaseUnknown
		if err = c.applyStatus(ctx, runner); err != nil {
			logger.Error(err, "Failed to apply status")
			return reconcile.WithError(err)
		}
		return reconcile.DoNotRequeue()
	}

	if err = c.watchForGVK(ctx, gvk); err != nil {
		logger.Error(err, "Failed to setup watch for WorkloadRunner", "namespace", runner.Namespace, "name", runner.Name)
		state.AddControllerIssue(runner, err, true, "GVK watch failed", issue.Internal)
		return reconcile.WithError(fmt.Errorf("failed to setup watch: %w", err))
	}

	err = c.buildResources(ctx, runner, template, karta, state)
	if err != nil {
		logger.Error(
			err,
			"Failed to get/create attached resource for WorkloadRunner",
			"namespace", runner.Namespace,
			"name", runner.Name,
		)
		state.AddControllerIssue(runner, err, true, "failed to build resources for WorkloadRunner", issue.Internal)

		return reconcile.WithError(fmt.Errorf("failed to get/create attached resource: %w", err))
	}

	logger.Info("Successfully reconciled spec for WorkloadRunner", "namespace", runner.Namespace, "name", runner.Name)

	return reconcile.Continue()
}

func (c *WorkloadRunnerController) watchForGVK(ctx context.Context, gvk schema.GroupVersionKind) error {
	logger := log.FromContext(ctx)

	c.watchMutex.Lock()
	defer c.watchMutex.Unlock()

	if c.activeWatches[gvk] {
		return nil
	}

	logger.Info("Setting up dynamic watch for GVK", "gvk", gvk.String())

	err := controllerutils.AddDynamicWatchForGVK(&c.manager, &c.wrController, gvk)
	if err != nil {
		return err
	}

	c.activeWatches[gvk] = true

	logger.Info("Successfully set up dynamic watch for GVK", "gvk", gvk.String())

	return nil
}
func (c *WorkloadRunnerController) buildResources(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	template *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	state *reconcile.State,
) error {
	if err := c.pluginFramework.RunHandleAdditionalResources(ctx, runner, template, karta, c.client, state); err != nil {
		return fmt.Errorf("HandleAdditionalResources failed: %w", err)
	}

	if _, err := c.handleAttachedResource(ctx, runner, template, karta, state); err != nil {
		return fmt.Errorf("failed to create/update attached resource: %w", err)
	}

	return nil
}

func (c *WorkloadRunnerController) handleAttachedResource(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	template *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	state *reconcile.State,
) (*unstructured.Unstructured, error) {
	if err := c.applyTemplateModification(ctx, runner, template, karta, state); err != nil {
		return nil, err
	}

	return c.applyAttachedResource(ctx, template, state)
}

func (c *WorkloadRunnerController) applyAttachedResource(
	ctx context.Context,
	template *unstructured.Unstructured,
	state *reconcile.State,
) (*unstructured.Unstructured, error) {
	applyOptions := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner(controllerName),
	}

	if err := c.client.Patch(ctx, template, client.Apply, applyOptions...); err != nil {
		state.AddControllerIssue(template, err, true, "failed to apply attached resource", issue.Internal)
		return nil, fmt.Errorf("failed to apply attached resource: %w", err)
	}

	log.FromContext(ctx).Info(
		"Successfully applied attached resource",
		"namespace", template.GetNamespace(),
		"name", template.GetName(),
		"gvk", template.GroupVersionKind().String(),
	)

	return template, nil
}

func (c *WorkloadRunnerController) applyTemplateModification(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	template *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	state *reconcile.State,
) error {
	template.SetNamespace(runner.Namespace)

	if err := ctrl.SetControllerReference(runner, template, c.manager.GetScheme()); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	if err := c.pluginFramework.RunMutateResource(ctx, runner, template, karta, state); err != nil {
		return fmt.Errorf("failed to run mutation plugins: %w", err)
	}

	return nil
}

func (c *WorkloadRunnerController) validateAttachedGVK(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	gvk schema.GroupVersionKind,
	karta *kartav1alpha1.Karta,
) (bool, error) {
	logger := log.FromContext(ctx)

	if !kartautils.IsReady(karta) {
		logger.V(2).Info("Karta for GVK is not ready", "gvk", gvk.String()) //nolint:mnd // log verbosity level
		c.setPrerequisitesPassedCondition(
			runner,
			dromosv1alpha1.ReasonKartaNotReady,
			fmt.Sprintf("Karta for GVK %s is not ready", gvk.String()),
		)

		return false, nil
	}

	c.setKartaReadyCondition(runner)

	return true, nil
}

func (c *WorkloadRunnerController) setPrerequisitesPassedCondition(
	runner *dromosv1alpha1.WorkloadRunner,
	reason, message string,
) {
	meta.SetStatusCondition(&runner.Status.Conditions, metav1.Condition{
		Type:    dromosv1alpha1.ConditionTypePrerequisitesPassed,
		Status:  metav1.ConditionFalse,
		Reason:  reason,
		Message: message,
	})
}

func (c *WorkloadRunnerController) setKartaReadyCondition(runner *dromosv1alpha1.WorkloadRunner) {
	meta.SetStatusCondition(&runner.Status.Conditions, metav1.Condition{
		Type:    dromosv1alpha1.ConditionTypePrerequisitesPassed,
		Status:  metav1.ConditionTrue,
		Reason:  "KartaReady",
		Message: "Karta is ready",
	})
}
