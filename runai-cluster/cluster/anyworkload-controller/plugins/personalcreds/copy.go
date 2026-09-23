/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package personalcreds

import (
	"context"
	"fmt"

	pcreds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	cryptoutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders/utils"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (p *Plugin) HandleAdditionalResources(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	_ *unstructured.Unstructured,
	_ *kartav1alpha1.Karta,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	if !pcreds.Has(runner) {
		return nil
	}

	entries, err := pcreds.Get(runner)
	if err != nil {
		return state.AddPluginIssue(ctx, p, runner, fmt.Errorf("invalid %s annotation: %w", pcreds.Annotation, err), true, "", issue.Internal)
	}

	key, err := cryptoutils.GetEncryptionKey()
	if err != nil {
		return state.AddPluginIssue(ctx, p, runner, fmt.Errorf("encryption key unavailable: %w", err), true, "", issue.Internal)
	}

	for _, e := range entries {
		if err := p.materializeDestSecret(ctx, runner, e.SourceSecret, key, k8sClient, state); err != nil {
			return err
		}
	}
	return nil
}

// materializeDestSecret fetches the encrypted source Secret from the runai
// namespace, decrypts its Data with the cluster encryption key, and
// server-side-applies a destination Secret in the workload namespace owned by
// the runner. The destination name is deterministic via DestName(runner, source);
// if a Secret with that name already exists and is not controlled by this
// runner, materialization fails with a Storage issue.
func (p *Plugin) materializeDestSecret(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	sourceName string,
	key []byte,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	source := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sourceName,
			Namespace: p.runaiNS,
		},
	}
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: sourceName, Namespace: p.runaiNS}, source); err != nil {
		return state.AddPluginIssue(ctx, p, source, fmt.Errorf("failed to get source secret %s/%s: %w", p.runaiNS, sourceName, err), true, "", issue.Storage)
	}

	dest, err := pcreds.BuildDestination(
		pcreds.DestName(runner.Name, sourceName),
		runner.Namespace,
		source,
		key,
	)
	if err != nil {
		return state.AddPluginIssue(ctx, p, source, fmt.Errorf("failed to build destination secret for %s: %w", sourceName, err), true, "", issue.Internal)
	}

	existing := &corev1.Secret{}
	getErr := k8sClient.Get(ctx, client.ObjectKeyFromObject(dest), existing)
	switch {
	case getErr == nil:
		if !metav1.IsControlledBy(existing, runner) {
			return state.AddPluginIssue(ctx, p, dest, fmt.Errorf("destination secret %s/%s exists but is not owned by runner %s", dest.Namespace, dest.Name, runner.Name), true, "", issue.Storage)
		}
	case !apierrors.IsNotFound(getErr):
		return state.AddPluginIssue(ctx, p, dest, fmt.Errorf("failed to check destination secret %s/%s: %w", dest.Namespace, dest.Name, getErr), true, "", issue.Storage)
	}

	if err := controllerutil.SetControllerReference(runner, dest, k8sClient.Scheme()); err != nil {
		return state.AddPluginIssue(ctx, p, dest, fmt.Errorf("failed to set owner reference on %s/%s: %w", dest.Namespace, dest.Name, err), true, "", issue.Internal)
	}

	// Server-side apply rejects managedFields and resourceVersion in the
	// request body — strip both before Patch.
	dest.SetManagedFields(nil)
	dest.SetResourceVersion("")
	//nolint:staticcheck // Typed Secrets require the Patch API for server-side apply.
	if err := k8sClient.Patch(ctx, dest, client.Apply, client.ForceOwnership, client.FieldOwner(FieldManager)); err != nil {
		return state.AddPluginIssue(ctx, p, dest, fmt.Errorf("failed to apply destination secret %s/%s: %w", dest.Namespace, dest.Name, err), true, "", issue.Storage)
	}
	return nil
}
