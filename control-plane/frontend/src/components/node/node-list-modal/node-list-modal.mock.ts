/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Node } from "@/swagger-models/cluster-service-client";
import type { INodeTable } from "@/table-models/node.table-model";

export const mockNode1: Node = {
  name: "node-gpu-1",
  nodePool: "gpu-pool",
  status: "Ready",
  clusterUuid: "test-cluster-123",
  gpuInfo: {
    gpuType: "NVIDIA-A100",
    gpuCount: 8,
  },
  createdAt: "2021-01-01T00:00:00Z",
  updatedAt: "2021-01-01T00:00:00Z",
};

export const mockNode2: Node = {
  name: "node-gpu-2",
  nodePool: "gpu-pool",
  status: "Ready",
  clusterUuid: "test-cluster-123",
  gpuInfo: {
    gpuType: "NVIDIA-A100",
    gpuCount: 8,
  },
  createdAt: "2021-01-01T00:00:00Z",
  updatedAt: "2021-01-01T00:00:00Z",
};

export const mockNode3: Node = {
  name: "node-gpu-3",
  nodePool: "gpu-pool",
  status: "NotReady",
  clusterUuid: "test-cluster-123",
  gpuInfo: {
    gpuType: "NVIDIA-A100",
    gpuCount: 8,
  },
  createdAt: "2021-01-01T00:00:00Z",
  updatedAt: "2021-01-01T00:00:00Z",
};

export const mockNodes: INodeTable[] = [mockNode1 as INodeTable, mockNode2 as INodeTable, mockNode3 as INodeTable];
