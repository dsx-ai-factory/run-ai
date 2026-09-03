/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"context"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterWideReplicatedReconciler adapts the (project-controller-style) replication handlers,
// to the shared-objects reconciler pattern: it runs the handler
type ClusterWideReplicatedReconciler struct {
	client          client.Client
	gvkCommon       *ClusterWideGVKCommon
	resourceHandler *ClusterWideReplicatedResourceHandler
}

func NewClusterWideReplicatedReconciler(
	ctx context.Context, c client.Client, clusterWideSecrets, clusterWidePvcs, clusterWideConfigMaps bool,
) (*ClusterWideReplicatedReconciler, error) {
	gvkCommon, err := NewClusterWideGVKCommon(ctx, c, &ClusterWideReplicatorGVKConfig{
		ClusterWideSecrets:    clusterWideSecrets,
		ClusterWidePvcs:       clusterWidePvcs,
		ClusterWideConfigMaps: clusterWideConfigMaps,
	})
	if err != nil {
		return nil, err
	}
	return &ClusterWideReplicatedReconciler{
		client:          c,
		gvkCommon:       gvkCommon,
		resourceHandler: NewClusterWideReplicatedResourceHandler(c, gvkCommon),
	}, nil
}

func (r *ClusterWideReplicatedReconciler) Reconcile(ctx context.Context, project *kaiorg.Project) error {
	handleErr := r.resourceHandler.HandleResource(ctx, *project)
	return handleErr
}
