/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package inferenceendpoint

import (
	"context"
	"fmt"
	ep "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromos "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	karta "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (p *Plugin) HandleAdditionalResources(ctx context.Context, runner *dromos.WorkloadRunner, _ *unstructured.Unstructured, _ *karta.Karta, c client.Client, state *reconcile.State) error {
	if !runner.DeletionTimestamp.IsZero() {
		return nil
	}
	if err := p.reconcile(ctx, runner, c); err != nil {
		return state.AddPluginIssue(ctx, p, runner, err, true, "", issue.Internal)
	}
	return nil
}

func (p *Plugin) reconcile(ctx context.Context, runner *dromos.WorkloadRunner, c client.Client) error {
	cfg, err := ep.GetInferenceEndpoint(runner)
	if err != nil {
		return err
	}
	if err := p.available(); err != nil {
		if cfg == nil && meta.IsNoMatchError(err) {
			return nil
		}
		return err
	}
	name, err := ep.InferenceEndpointName(runner)
	if err != nil {
		return err
	}
	live := &unstructured.Unstructured{}
	live.SetGroupVersionKind(ep.InferenceEndpointGVK)
	err = c.Get(ctx, client.ObjectKey{Namespace: runner.Namespace, Name: name}, live)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("read InferenceEndpoint: %w", err)
	}
	if cfg == nil {
		if apierrors.IsNotFound(err) || !metav1.IsControlledBy(live, runner) {
			return nil
		}
		// Preconditions protect against a delete/recreate or ownership change between
		// the read and delete. Never delete a newly substituted object on a stale read.
		uid, rv := live.GetUID(), live.GetResourceVersion()
		if err := c.Delete(ctx, live, client.Preconditions{UID: &uid, ResourceVersion: &rv}); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("delete InferenceEndpoint: %w", err)
		}
		return nil
	}
	desired, buildErr := ep.BuildInferenceEndpoint(runner, cfg)
	if buildErr != nil {
		return buildErr
	}
	if err == nil {
		if !metav1.IsControlledBy(live, runner) {
			return fmt.Errorf("InferenceEndpoint %s/%s is not controlled by runner %s", runner.Namespace, name, runner.Name)
		}
		if !live.GetDeletionTimestamp().IsZero() {
			return fmt.Errorf("InferenceEndpoint %s/%s is terminating", runner.Namespace, name)
		}
		if apiequality.Semantic.DeepEqual(live.Object["spec"], desired.Object["spec"]) {
			return nil
		}
		live.Object["spec"] = desired.Object["spec"]
		// Update retains resourceVersion, so a concurrent change cannot be overwritten.
		// Strict validation rejects fields the CRD would otherwise prune and cause us to reapply forever.
		if err := c.Update(ctx, live, client.FieldValidation(metav1.FieldValidationStrict)); err != nil {
			return fmt.Errorf("update InferenceEndpoint: %w", err)
		}
		return nil
	}
	if err := controllerutil.SetControllerReference(runner, desired, c.Scheme()); err != nil {
		return fmt.Errorf("own InferenceEndpoint: %w", err)
	}
	if err := c.Create(ctx, desired, client.FieldValidation(metav1.FieldValidationStrict)); err != nil {
		// The Get above reads the informer cache, which can lag a Create that already
		// succeeded. The Owns watch delivers the object and the next reconcile converges.
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return fmt.Errorf("create InferenceEndpoint: %w", err)
	}
	return nil
}
