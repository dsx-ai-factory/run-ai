/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_service

import (
	"context"

	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

// FactoryInterface defines and interface for client factory
type FactoryInterface interface {
	// Client returns a workload service client
	Client() ClientInterface
}

// ClientInterface defines all the API the client provides
type ClientInterface interface {
	// InternalGetResourceClaimTemplateWorkloads lists the workloads in the request's scope whose pods
	// reference the resource claim template with the request's name
	InternalGetResourceClaimTemplateWorkloads(ctx context.Context,
		request *workloadsServiceClient.InternalResourceClaimTemplateWorkloadsRequest, offset int32, limit int32,
	) (*workloadsServiceClient.ResourceClaimTemplateWorkloadsResponse, *workloadsServiceClient.Error)
}
