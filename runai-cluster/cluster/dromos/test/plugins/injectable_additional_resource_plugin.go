/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plugins

import (
	"context"
	"sync"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const InjectableAdditionalResourcePluginName = "injectable-additional-resource-plugin"

type InjectableAdditionalResourcePlugin struct {
	mu  sync.Mutex
	err error
}

func NewInjectableAdditionalResourcePlugin() *InjectableAdditionalResourcePlugin {
	return &InjectableAdditionalResourcePlugin{}
}

func (p *InjectableAdditionalResourcePlugin) SetError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.err = err
}

func (p *InjectableAdditionalResourcePlugin) Name() string {
	return InjectableAdditionalResourcePluginName
}

func (p *InjectableAdditionalResourcePlugin) HandleAdditionalResources(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

var _ framework.AdditionalResourcePlugin = &InjectableAdditionalResourcePlugin{}
