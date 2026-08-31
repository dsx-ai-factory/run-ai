/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
	commonprometheus "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/common/prometheus"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_prometheus "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RWAuthSecretResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewRWAuthSecretResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *RWAuthSecretResource {
	return &RWAuthSecretResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (r *RWAuthSecretResource) OperandName() string {
	return r.operandName
}

func (r *RWAuthSecretResource) Object(ctx context.Context) (client.Object, error) {
	var (
		namespace = r.runaiConfig.Namespace
	)

	if shouldUseOpenshiftBuiltInPrometheus(r.runaiConfig) {
		namespace = consts.OpenshiftMonitoringNamespace
	}

	secret := &corev1.Secret{}
	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Name:      runv1_prometheus.SecretName,
		Namespace: namespace,
	}, secret)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	secret.Name = runv1_prometheus.SecretName
	secret.Namespace = namespace

	return commonprometheus.SetRWAuthSecretFields(ctx, secret, r.k8sReader)
}

func (r *RWAuthSecretResource) Status(ctx context.Context) error {
	secret, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if secret.GetUID() == "" {
		return fmt.Errorf("secret [%s] is missing", secret.GetName())
	}

	return nil
}

func (r *RWAuthSecretResource) Enabled() bool {
	return !r.runaiConfig.Spec.Internal.Global.NMC.IsEnabled()
}
