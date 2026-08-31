/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package templates

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/templates/resources/configmaps"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/templates/resources/service_monitors"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// name specifies the name of this operand for logging purposes.
	name = "templates"
)

// RBAC permissions the operand requires for resource management
//+kubebuilder:rbac:namespace=runai,groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=configmaps,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=secrets,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=monitoring,groups="",resources=configmaps,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=cert-manager.io,resources=issuers,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create;update;watch;patch;delete;list

type operand struct {
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
	resources   []operands.Resource
}

var _ operands.Operand = &operand{}

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
	return true
}

func Operand(kbClient kc.HybridClient, runaiConfig runv1.RunaiConfig) operands.Operand {
	return &operand{
		k8sReader:   kbClient,
		runaiConfig: &runaiConfig,
		resources: []operands.Resource{
			configmaps.NewProxyCert(name, kbClient, runaiConfig),
			configmaps.NewLimitRange(name, kbClient, runaiConfig),
			configmaps.NewNvidiaDevicePlugin(name, kbClient, runaiConfig),
			service_monitors.NewServiceMonitor(name, kbClient, runaiConfig),
		},
	}
}
