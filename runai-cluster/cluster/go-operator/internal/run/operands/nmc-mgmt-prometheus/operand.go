/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nmc_mgmt_prometheus

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/common/shared_data"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/nmc-mgmt-prometheus/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	// name specifies the name of this operand for logging purposes.
	name = "nmc-mgmt-prometheus"
)

// RBAC permissions the operand requires for resource management
//+kubebuilder:rbac:namespace=runai,groups=monitoring.coreos.com,resources=prometheuses,verbs=create;get;update;watch;patch;list;delete
//+kubebuilder:rbac:namespace=runai,groups=monitoring.coreos.com,resources=prometheusrules,verbs=create;get;update;watch;patch;list;delete
//+kubebuilder:rbac:namespace=openshift-monitoring,groups="",resources=configmaps,verbs=get;update;watch
//+kubebuilder:rbac:namespace=openshift-monitoring,groups="",resources=secrets,verbs=get;create;list
//+kubebuilder:rbac:namespace=runai,groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=gateway.networking.k8s.io,resources=httproutes/status,verbs=get

type operand struct {
	k8sReader   kc.HybridClient
	runaiConfig *runv1.RunaiConfig
	resources   []operands.Resource
	sharedData  *shared_data.SharedData
}

func (o *operand) Name() string {
	return name
}

func (o *operand) Resources() []operands.Resource {
	return o.resources
}

func (o *operand) Status(ctx context.Context) (operands.Status, error) {
	return utils.Status(ctx, o.resources)
}

func (o *operand) Enabled() bool {
	return *o.runaiConfig.Spec.Prometheus.Enabled &&
		!o.runaiConfig.Spec.Internal.Development.Headless &&
		o.runaiConfig.Spec.Internal.Global.NMC.IsEnabled()
}

func Operand(kbClient kc.HybridClient, runaiConfig *runv1.RunaiConfig, sharedData *shared_data.SharedData) operands.Operand {
	return &operand{
		k8sReader:   kbClient,
		runaiConfig: runaiConfig,
		resources: []operands.Resource{
			resources.NewRWAuthSecretResource(name, kbClient, runaiConfig),
			resources.NewPrometheusResource(name, kbClient, runaiConfig),
			resources.NewPrometheusRuleResource(name, kbClient, runaiConfig),
			resources.NewServiceResource(name, kbClient, runaiConfig),
			resources.NewIngressResource(name, kbClient, runaiConfig),
			resources.NewHTTPRouteResource(name, kbClient, runaiConfig),
			resources.NewTLSSecretResource(name, kbClient, runaiConfig, sharedData),
		},
		sharedData: sharedData,
	}
}
