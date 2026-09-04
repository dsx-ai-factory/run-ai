/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	commonprometheus "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/common/prometheus"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PrometheusRuleResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runaiv1.RunaiConfig
}

func NewPrometheusRuleResource(operandName string, k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) *PrometheusRuleResource {
	return &PrometheusRuleResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (r *PrometheusRuleResource) OperandName() string {
	return r.operandName
}

func (r *PrometheusRuleResource) Object(ctx context.Context) (client.Object, error) {
	var (
		name      = "nmc-mgmt-prometheus-rules"
		namespace = r.runaiConfig.Namespace
	)

	rule := &monitoringv1.PrometheusRule{}

	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, rule)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	rule.Name = name
	rule.Namespace = namespace

	return commonprometheus.SetCommonPrometheusRuleFields(
		rule,
		r.runaiConfig,
	), nil
}

func (r *PrometheusRuleResource) Status(ctx context.Context) error {
	resource, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if resource.GetUID() == "" {
		return fmt.Errorf("rule [%s] is missing", resource.GetName())
	}

	return nil
}

func (r *PrometheusRuleResource) Enabled() bool {
	return true
}
