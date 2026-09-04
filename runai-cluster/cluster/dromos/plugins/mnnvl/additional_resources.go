/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"context"
	"errors"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// HandleAdditionalResources creates ComputeDomain for MNNVL workloads
func (p *Plugin) HandleAdditionalResources(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	logger := log.FromContext(ctx)

	config := extractMnnvlConfig(runner)
	if config == nil || config.Mode != dromosv1alpha1.MnnvlModeRequired {
		return nil
	}

	numNodes, err := p.calculateNumNodes(ctx, resource, karta)
	if err != nil {
		return state.AddPluginIssue(ctx, p, resource, err, true, "node calculation error", issue.Compute)
	}

	if numNodes == 0 {
		logger.Info("no GPU resources found for the workload, skipping ComputeDomain creation", "workloadName", runner.Name)
		return nil
	}

	params := computeDomainWorkloadParams{
		WorkloadKind:      runner.GetObjectKind().GroupVersionKind().Kind,
		JobType:           resource.GetKind(),
		WorkloadInitials:  workloadRunnerInitials,
		WorkloadName:      runner.Name,
		WorkloadNamespace: runner.Namespace,
		NumNodes:          numNodes,
		ElasticEnabled:    p.clusterElasticCDEnabled,
	}

	computeDomain, err := buildComputeDomain(
		ctx,
		params,
		k8sClient.Scheme(),
		runner,
	)
	if err != nil {
		return state.AddPluginIssue(ctx, p, resource, err, true, "compute domain build failed", issue.Compute)
	}

	if computeDomain == nil {
		err = errors.New("failed to build compute domain")
		return state.AddPluginIssue(ctx, p, resource, err, true, "", issue.Compute)
	}

	err = p.applyComputeDomain(ctx, k8sClient, computeDomain)
	if err != nil {
		return state.AddPluginIssue(ctx, p, resource, err, true, "", issue.Compute)
	}
	return nil
}
