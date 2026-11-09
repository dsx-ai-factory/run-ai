/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpuinventory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Collector struct {
	client   client.Client
	nodeName string
	read     func() ([]Device, error)
	now      func() time.Time
	bootID   func() (string, error)
}

func NewCollector(kubeClient client.Client, nodeName, sysRoot, procRoot string) *Collector {
	return &Collector{client: kubeClient, nodeName: nodeName, read: func() ([]Device, error) { return ReadDevices(sysRoot, procRoot) }, now: time.Now, bootID: func() (string, error) {
		b, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
		return strings.TrimSpace(string(b)), err
	}}
}

func (c *Collector) Collect(ctx context.Context) error {
	node := &corev1.Node{}
	if err := c.client.Get(ctx, client.ObjectKey{Name: c.nodeName}, node); err != nil {
		return fmt.Errorf("get collector node %s: %w", c.nodeName, err)
	}
	before := node.DeepCopy()
	observedAt := c.now()
	devices, collectionErr := c.read()
	bootID, bootErr := c.bootID()
	if bootErr != nil {
		collectionErr = errors.Join(collectionErr, fmt.Errorf("read kernel boot ID: %w", bootErr))
	} else if bootID != node.Status.NodeInfo.BootID {
		collectionErr = errors.Join(collectionErr, fmt.Errorf("kernel boot ID does not match node status"))
	}
	observation := NewObservation(node, devices, observedAt)
	if collectionErr == nil && observation.Identity(node, c.now()) == "" {
		collectionErr = fmt.Errorf("hardware does not match a homogeneous node GPU inventory")
	}
	if collectionErr != nil {
		if _, exists := node.Annotations[Annotation]; !exists {
			return collectionErr
		}
		delete(node.Annotations, Annotation)
	} else {
		data, err := json.Marshal(observation)
		if err != nil {
			return fmt.Errorf("encode GPU observation: %w", err)
		}
		if node.Annotations == nil {
			node.Annotations = map[string]string{}
		}
		node.Annotations[Annotation] = string(data)
	}
	// A changed resource version requires recollection; otherwise a node replacement,
	// reboot or concurrent label update could receive evidence collected before it.
	patch := client.MergeFromWithOptions(before, client.MergeFromWithOptimisticLock{})
	if err := c.client.Patch(ctx, node, patch); err != nil {
		return errors.Join(collectionErr, fmt.Errorf("publish GPU observation on %s: %w", c.nodeName, err))
	}
	return collectionErr
}

func (c *Collector) Run(ctx context.Context) {
	ticker := time.NewTicker(RefreshInterval)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		requestCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		err := c.Collect(requestCtx)
		cancel()
		if err != nil {
			zap.S().Warnw("GPU identity collection unavailable", "node", c.nodeName, "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
