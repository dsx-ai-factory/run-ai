/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prometheus

import (
	"context"
	"fmt"

	gosecrets "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/secrets"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_prometheus "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/prometheus"
	runaiv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func SetRWAuthSecretFields(ctx context.Context, secret *corev1.Secret, k8sReader client.Reader) (*corev1.Secret, error) {
	icSecret := &corev1.Secret{}
	if err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      gosecrets.InstallationConfigurationSecretName,
		Namespace: runaiv1.DefaultNamespace,
	}, icSecret); err != nil {
		return nil, fmt.Errorf("failed to read installation-configuration secret: %w", err)
	}

	configBytes, ok := icSecret.Data[gosecrets.InstallationConfigurationSecretName]
	if !ok || len(configBytes) == 0 {
		return nil, fmt.Errorf("installation-configuration secret missing %q data key", gosecrets.InstallationConfigurationSecretName)
	}
	internalData := &runaiv1_internal.Internal{}
	if err := yaml.Unmarshal(configBytes, internalData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal installation-configuration: %w", err)
	}

	secret.Type = corev1.SecretTypeOpaque
	secret.Data = map[string][]byte{}

	if internalData.Global != nil && internalData.Global.ControlPlane != nil && internalData.Global.ControlPlane.Config != nil {
		for _, remoteWrite := range internalData.Global.ControlPlane.Config.GetMetricsRemoteWrites() {
			userKey, passwordKey := runv1_prometheus.GetSecretIdentifiers(remoteWrite.GetLabel())
			secret.Data[userKey] = []byte(remoteWrite.Credentials.GetUsername())
			secret.Data[passwordKey] = []byte(remoteWrite.Credentials.GetPassword())
		}
	}

	return secret, nil
}
