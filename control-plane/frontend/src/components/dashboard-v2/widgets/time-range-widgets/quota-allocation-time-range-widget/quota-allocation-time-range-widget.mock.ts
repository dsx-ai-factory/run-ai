/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { MetricsType, type MetricsResponse } from "@/swagger-models/cluster-service-client";

export const mockQuotaAllocationMetrics: MetricsResponse = {
  measurements: [
    {
      type: MetricsType.AllocatedGpu,
      values: [
        { value: "12", timestamp: "2024-03-01T00:00:00.000Z" },
        { value: "14", timestamp: "2024-03-01T04:00:00.000Z" },
        { value: "10", timestamp: "2024-03-01T08:00:00.000Z" },
        { value: "16", timestamp: "2024-03-01T12:00:00.000Z" },
        { value: "18", timestamp: "2024-03-01T16:00:00.000Z" },
        { value: "15", timestamp: "2024-03-01T20:00:00.000Z" },
        { value: "13", timestamp: "2024-03-02T00:00:00.000Z" },
      ],
    },
    {
      type: MetricsType.GpuQuota,
      values: [
        { value: "20", timestamp: "2024-03-01T00:00:00.000Z" },
        { value: "20", timestamp: "2024-03-01T04:00:00.000Z" },
        { value: "20", timestamp: "2024-03-01T08:00:00.000Z" },
        { value: "20", timestamp: "2024-03-01T12:00:00.000Z" },
        { value: "20", timestamp: "2024-03-01T16:00:00.000Z" },
        { value: "20", timestamp: "2024-03-01T20:00:00.000Z" },
        { value: "20", timestamp: "2024-03-02T00:00:00.000Z" },
      ],
    },
  ],
};
