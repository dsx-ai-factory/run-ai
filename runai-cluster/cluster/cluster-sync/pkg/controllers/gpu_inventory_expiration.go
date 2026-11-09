/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/gpuinventory"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type gpuInventoryExpiration struct {
	reader  client.Reader
	enqueue func(client.ObjectKey)
	expired map[types.UID]string
}

func newGPUInventoryExpiration(reader client.Reader, enqueue func(client.ObjectKey)) *gpuInventoryExpiration {
	return &gpuInventoryExpiration{reader: reader, enqueue: enqueue, expired: map[types.UID]string{}}
}
func (e *gpuInventoryExpiration) scan(ctx context.Context, now time.Time) error {
	nodes := &corev1.NodeList{}
	if err := e.reader.List(ctx, nodes); err != nil {
		return fmt.Errorf("list nodes for GPU identity expiry: %w", err)
	}
	expired := map[types.UID]string{}
	for i := range nodes.Items {
		node := &nodes.Items[i]
		if node.Labels[v1alpha1.NodePoolNameLabel] == clusterconstants.ExcludedNodepoolName {
			continue
		}
		raw := node.Annotations[gpuinventory.Annotation]
		if raw == "" || gpuinventory.ReportedIdentity(node, now) != "" {
			continue
		}
		expired[node.UID] = raw
		if e.expired[node.UID] != raw {
			e.enqueue(client.ObjectKeyFromObject(node))
		}
	}
	e.expired = expired
	return nil
}

// StartGPUInventoryExpiration handles time passing without a Node informer event.
// The normal retry queue reads the latest node before sending its cleared ID.
func StartGPUInventoryExpiration(ctx context.Context, reader client.Reader, controller *cachedclient.RateLimitedController) {
	tracker := newGPUInventoryExpiration(reader, controller.AddRateLimited)
	ticker := time.NewTicker(gpuinventory.RefreshInterval)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		if err := tracker.scan(ctx, time.Now()); err != nil {
			zap.S().Warnw("GPU identity expiry scan failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
