/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package service

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/workload"
	v1 "k8s.io/api/core/v1"
)

type PodsGetter interface {
	GetPod(ctx context.Context, workload *workload.Workload, preferredPod string) (*v1.Pod, error)
}

type SelectorFactoryFN func(workload *workload.Workload) workload.Selector
