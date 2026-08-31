/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export const mockNodePoolFreeGpusResponse = {
  type: "FREE_GPUS",
  timestamp: "2024-03-10T09:57:42.397731288Z",
  values: [
    {
      value: "10",
      groups: [
        {
          key: "Nodepool",
          value: "noode-pool-1",
        },
        {
          key: "Node",
          value: "runai-cluster-worker",
        },
      ],
    },
    {
      value: "13",
      groups: [
        {
          key: "Nodepool",
          value: "default",
        },
        {
          key: "Node",
          value: "runai-cluster-worker2",
        },
      ],
    },
  ],
};
export const mockNodeTotalGpusResponse = {
  type: "TOTAL_GPUS",
  timestamp: "2024-03-10T09:56:52.32760532Z",
  values: [
    {
      value: "16",
      groups: [
        {
          key: "Node",
          value: "runai-cluster-worker",
        },
      ],
    },
    {
      value: "16",
      groups: [
        {
          key: "Node",
          value: "runai-cluster-worker2",
        },
      ],
    },
  ],
};
