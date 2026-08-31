/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package nodepools reads node pool names from the cluster for the policy field rules.
//
// It exists so that sdk/apis/workload_fields/field_groups can offer node pool options
// without importing the NodePool CRD types. Importing the kai.resources API package pulls
// KAI-scheduler's pkg/apis/kai/v1 and from there its scheduler internals - code written
// against controller-runtime >= 0.23 - and the control plane services compile
// field_groups while pinning controller-runtime to 0.22.4 via a replace directive.
//
// field_groups declares the NodePoolNameLister interface; this package implements it.
// The control plane never imports this package.
package nodepools

import (
	"context"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NameLister lists node pool names from the cluster.
type NameLister struct {
	reader client.Reader
}

// NewNameLister returns a NameLister backed by the given reader.
func NewNameLister(reader client.Reader) *NameLister {
	return &NameLister{reader: reader}
}

// ListSchedulableNodePoolNames returns the names of the node pools a workload may be
// submitted to, skipping any that are being deleted.
func (l *NameLister) ListSchedulableNodePoolNames(ctx context.Context) ([]string, error) {
	return l.listKaiNames(ctx)
}

func (l *NameLister) listKaiNames(ctx context.Context) ([]string, error) {
	nodePools := &kaiv1alpha1.NodePoolList{}
	if err := l.reader.List(ctx, nodePools); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(nodePools.Items))
	for _, nodePool := range nodePools.Items {
		if nodePool.Status.Phase == kaiv1alpha1.NodePoolDeleting {
			continue
		}
		names = append(names, nodePool.Name)
	}
	return names, nil
}
