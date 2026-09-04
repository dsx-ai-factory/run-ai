/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	osecrets "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/common/secrets"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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
	namespace := s.runaiConfig.Namespace

	secret := &v1.Secret{}
	err := s.k8sReader.Get(ctx, types.NamespacedName{Name: DeploymentName, Namespace: namespace}, secret)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	password := string(secret.Data[constants.WorkloadAuthRedisPasswordKey])
	if password == "" {
		workloadsAuthSecret := &v1.Secret{}
		err = s.k8sReader.Get(ctx, types.NamespacedName{Name: constants.WorkloadAuthSecretName, Namespace: namespace}, workloadsAuthSecret)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		password = string(workloadsAuthSecret.Data[constants.WorkloadAuthRedisPasswordKey])
	}
	if password == "" {
		password, err = osecrets.GeneratePassword()
		if err != nil {
			return nil, fmt.Errorf("failed to generate Redis password: %w", err)
		}
	}

	secret.ObjectMeta.Name = DeploymentName
	secret.ObjectMeta.Namespace = namespace

	if secret.ObjectMeta.Labels == nil {
		secret.ObjectMeta.Labels = map[string]string{}
	}
	secret.ObjectMeta.Labels["app"] = DeploymentName

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}
	secret.Data[constants.WorkloadAuthRedisPasswordKey] = []byte(password)
	secret.Data[valkeyConfFileName] = []byte(fmt.Sprintf(`requirepass "%s"`, password))

	return secret, nil
}

func (s *SecretResource) Status(ctx context.Context) error {
	resource, err := s.Object(ctx)
	if err != nil {
		return err
	}
	if resource.GetUID() == "" {
		return fmt.Errorf("Secret [%s] is missing", resource.GetName())
	}
	return nil
}

func (s *SecretResource) Enabled() bool { return true }
