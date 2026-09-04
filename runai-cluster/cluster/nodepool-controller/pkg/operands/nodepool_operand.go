/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package operands

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceFunction func(context.Context, client.Reader, *v1alpha1.NodePool, *common.NodePoolControllerParams, string) (client.Object, error)
type ResourceStatusFunction func(context.Context, client.Reader, *v1alpha1.NodePool, *common.NodePoolControllerParams, string) (Status, error)

// NodePoolOperand is an interface each operand has to implement in order to
// smoothly integrate with the controller.
// An Operand is defined as a group of Kubernetes resources that
// together form a logical unit, such as a micro-service, database, etc
type NodePoolOperand interface {
	// Name returns a readable operand name for logging
	Name() string

	// ResourcesForNodePool returns a slice of Kubernetes resources
	// that should be Created/Updated/Deleted by the controller to satisfy
	// the needs of this operand.
	ResourcesForNodePool(ctx context.Context, k8sReader client.Reader,
		nodePool *v1alpha1.NodePool, params *common.NodePoolControllerParams) ([]ResourceOld, error)

	// Status returns the status of the operand as calculated by the operand.
	// If some resources are not yet available for usage, a status with 'Ready'
	// set to false and 'Reasons' stated should be returned.
	Status(ctx context.Context, k8sReader client.Reader,
		nodePool *v1alpha1.NodePool, params *common.NodePoolControllerParams) (Status, error)
}
