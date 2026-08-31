/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/secrets"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewSecretResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *SecretResource {
	return &SecretResource{operandName: operandName, k8sReader: k8sReader, runaiConfig: runaiConfig}
}

func (s *SecretResource) OperandName() string { return s.operandName }

func (s *SecretResource) Object(ctx context.Context) (client.Object, error) {
	ic, err := secrets.GetConfigurationSecret(ctx, s.k8sReader)
	if err != nil {
		return nil, err
	}
	secret, err := secrets.GetDeploymentSecret(ctx, s.k8sReader, DeploymentName)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration secret for agent: %w", err)
	}
	vault, err := secrets.NewVault(DeploymentName, ic, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault for agent: %w", err)
	}
	vault.AddFromSecret(secrets.ClientName, "CLIENT_NAME")
	vault.AddFromSecret(secrets.ClientSecret, "CLIENT_SECRET")
	vault.AddFromSecret(secrets.ClusterID, "CLUSTER_ID")
	return vault.Generate()
}

func (s *SecretResource) Status(ctx context.Context) error {
	secret, err := s.Object(ctx)
	if err != nil {
		return err
	}
	if secret.GetUID() == "" {
		return fmt.Errorf("secret [%s] is missing", secret.GetName())
	}
	return nil
}

func (s *SecretResource) Enabled() bool { return true }
