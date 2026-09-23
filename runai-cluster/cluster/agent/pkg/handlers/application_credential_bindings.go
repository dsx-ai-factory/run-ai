/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/helmvalues"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *ApplicationsHandler) resolveCredentialBindings(ctx context.Context, applicationName, namespace string, bindings []workloadsapi.AIApplicationCredentialBinding, sources datastructs.Set[string]) ([]v1alpha1.HelmCredentialBinding, error) {
	variables := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		variables = append(variables, binding.Variable)
	}
	if err := helmvalues.ValidatePaths(variables); err != nil {
		return nil, err
	}
	var resolved []v1alpha1.HelmCredentialBinding
	for _, binding := range bindings {
		secret, err := h.findBindingSecret(ctx, namespace, binding)
		if err != nil {
			return nil, fmt.Errorf("credential for %q: %w", binding.Variable, err)
		}
		name := secret.Name
		if binding.GetUserCredential() {
			sources.Add(name)
			name = personalcreds.DestName(applicationName, name)
		}
		resolved = append(resolved, v1alpha1.HelmCredentialBinding{Variable: binding.Variable, SecretName: name})
	}
	return resolved, nil
}

func (h *ApplicationsHandler) findBindingSecret(ctx context.Context, namespace string, binding workloadsapi.AIApplicationCredentialBinding) (*corev1.Secret, error) {
	kind := genericSecretKindLabel
	switch binding.CredentialType {
	case "image-pull-secret":
		kind = dockerRegistryKindLabel
	case "ngc-api-key", "generic":
	default:
		return nil, fmt.Errorf("unsupported credential type %q", binding.CredentialType)
	}
	reader := h.cachedClient
	match := client.MatchingLabels{assetIdLabel: binding.AssetId, constants.RunaiAssetLabel: kind}
	if binding.GetUserCredential() {
		namespace = constants.NamespaceRunAi
		reader = h.runaiClient
		match[userCredentialsLabel] = boolTrueAsString
	}
	secret := &corev1.Secret{}
	if name := binding.GetExistingSecretName(); name != "" && !binding.GetUserCredential() {
		if err := reader.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, secret); err != nil {
			return nil, fmt.Errorf("selected credential Secret is unavailable in namespace %q; ensure it exists and retry: %w", namespace, err)
		}
	} else {
		if binding.AssetId == "" {
			return nil, fmt.Errorf("credential asset ID is required")
		}
		list := &corev1.SecretList{}
		if err := reader.List(ctx, list, client.InNamespace(namespace), match); err != nil {
			return nil, fmt.Errorf("failed to check credential Secret availability: %w", err)
		}
		if len(list.Items) != 1 {
			return nil, fmt.Errorf("expected one %s Secret for selected credential in namespace %q, found %d; check credential synchronization and reselect or retry", kind, namespace, len(list.Items))
		}
		secret = &list.Items[0]
	}
	if binding.CredentialType == "image-pull-secret" {
		if !binding.GetUserCredential() && secret.Type != corev1.SecretTypeDockerConfigJson && secret.Type != corev1.SecretTypeDockercfg {
			return nil, fmt.Errorf("selected credential has no compatible image-pull Secret; please reselect")
		}
		key := corev1.DockerConfigJsonKey
		if secret.Type == corev1.SecretTypeDockercfg {
			key = corev1.DockerConfigKey
		}
		if _, exists := secret.Data[key]; !exists {
			return nil, fmt.Errorf("selected image-pull Secret is missing its registry configuration; check credential synchronization and retry")
		}
	} else if secret.Type != "" && secret.Type != corev1.SecretTypeOpaque {
		return nil, fmt.Errorf("selected credential has no compatible generic Secret; please reselect")
	}
	if binding.CredentialType == "ngc-api-key" {
		if _, exists := secret.Data["NGC_API_KEY"]; !exists {
			return nil, fmt.Errorf("selected NGC Secret is missing NGC_API_KEY; check credential synchronization and retry")
		}
	}
	return secret, nil
}
