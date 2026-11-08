/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpu_pci_inventory

import (
	"context"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const name = "gpu-pci-inventory"

//+kubebuilder:rbac:namespace=runai,groups=apps,resources=daemonsets,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;create;update;watch;patch;delete;list

type operand struct {
	config    *runv1.RunaiConfig
	resources []operands.Resource
}

func (o *operand) Name() string                   { return name }
func (o *operand) Resources() []operands.Resource { return o.resources }
func (o *operand) Enabled() bool                  { return *o.config.Spec.ClusterSync.Enabled }

func (o *operand) Status(context.Context) (operands.Status, error) {
	// Hardware discovery is optional: absent driver paths must not block the installation.
	return operands.ReadyStatus(), nil
}

func Operand(reader kc.HybridClient, config *runv1.RunaiConfig) operands.Operand {
	return &operand{config: config, resources: []operands.Resource{
		&daemonSetResource{reader: reader, config: config},
		&serviceAccountResource{reader: reader, config: config},
	}}
}
