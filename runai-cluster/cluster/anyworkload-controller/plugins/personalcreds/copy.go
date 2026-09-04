/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package personalcreds

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
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

const dockerRegistryKindLabel = "docker-registry"

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

	dest := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pcreds.DestName(runner.Name, sourceName),
			Namespace: runner.Namespace,
		},
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

	if err := buildDestSecret(dest, source, key); err != nil {
		return state.AddPluginIssue(ctx, p, source, fmt.Errorf("failed to build destination secret for %s: %w", sourceName, err), true, "", issue.Internal)
	}

	if err := controllerutil.SetControllerReference(runner, dest, k8sClient.Scheme()); err != nil {
		return state.AddPluginIssue(ctx, p, dest, fmt.Errorf("failed to set owner reference on %s/%s: %w", dest.Namespace, dest.Name, err), true, "", issue.Internal)
	}

	// Server-side apply rejects managedFields and resourceVersion in the
	// request body — strip both before Patch.
	dest.SetManagedFields(nil)
	dest.SetResourceVersion("")
	if err := k8sClient.Patch(ctx, dest, client.Apply, client.ForceOwnership, client.FieldOwner(FieldManager)); err != nil {
		return state.AddPluginIssue(ctx, p, dest, fmt.Errorf("failed to apply destination secret %s/%s: %w", dest.Namespace, dest.Name, err), true, "", issue.Storage)
	}
	return nil
}

// buildDestSecret populates dest in-place with the decrypted form of source:
// it copies the Secret Type, all labels, all annotations, and the Data map
// with each value decrypted using the provided encryption key. Returns an
// error if source has no Data or any value fails to decrypt.
func buildDestSecret(dest, source *corev1.Secret, key []byte) error {
	dest.Type = source.Type
	// The agent stores private docker-registry credentials as Opaque (see
	// agent/pkg/handlers/assets_builders.go) and identifies them via the
	// run.ai/asset label. The legacy workload-controller restores the proper
	// type during copy (workload-common/builders/credentials.go:173); mirror
	// that here so the kubelet honors the imagePullSecret.
	if source.Labels[constants.RunaiAssetLabel] == dockerRegistryKindLabel {
		dest.Type = corev1.SecretTypeDockerConfigJson
	}

	if len(source.Labels) > 0 {
		dest.Labels = make(map[string]string, len(source.Labels))
		for k, v := range source.Labels {
			dest.Labels[k] = v
		}
	}
	if len(source.Annotations) > 0 {
		dest.Annotations = make(map[string]string, len(source.Annotations))
		for k, v := range source.Annotations {
			dest.Annotations[k] = v
		}
	}

	if len(source.Data) == 0 {
		return fmt.Errorf("source secret %s/%s has no data", source.Namespace, source.Name)
	}
	dest.Data = make(map[string][]byte, len(source.Data))
	for k, encrypted := range source.Data {
		decrypted, err := cryptoutils.DecryptData(encrypted, key)
		if err != nil {
			return fmt.Errorf("failed to decrypt key %s: %w", k, err)
		}
		dest.Data[k] = decrypted
	}
	return nil
}
