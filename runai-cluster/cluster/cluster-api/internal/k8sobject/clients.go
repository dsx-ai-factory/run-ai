/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package k8sobject

import (
	"context"
	"net/http"

	workloadsapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

// ElementGetter narrows the workloads-client surface so tests don't embed
// the 40-method generated interface.
type ElementGetter interface {
	GetWorkloadElement(ctx context.Context, workloadID, elementID string) (*workloadsapi.WorkloadElementResponse, *http.Response, error)
}

// PodGetter narrows the workloads-client surface to the single-pod fetch.
// The (workloadID, podID) shape is what cluster-api needs to resolve a pod
type PodGetter interface {
	GetWorkloadPod(ctx context.Context, workloadID, podID string) (*workloadsapi.Pod, *http.Response, error)
}

type elementGetter struct{ api workloadsapi.WorkloadsAPI }

func (g *elementGetter) GetWorkloadElement(ctx context.Context, workloadID, elementID string) (*workloadsapi.WorkloadElementResponse, *http.Response, error) {
	return g.api.GetWorkloadElement(ctx, workloadID, elementID).Execute()
}

type podGetter struct{ api workloadsapi.PodsAPI }

func (g *podGetter) GetWorkloadPod(ctx context.Context, workloadID, podID string) (*workloadsapi.Pod, *http.Response, error) {
	return g.api.GetWorkloadPod(ctx, workloadID, podID).Execute()
}

func NewClients(api *workloadsapi.APIClient) (ElementGetter, PodGetter) {
	return &elementGetter{api: api.WorkloadsAPI}, &podGetter{api: api.PodsAPI}
}
