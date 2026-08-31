/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { GroupedMetricsResponse } from "@/swagger-models/cluster-service-client";

export const mockWorkloadByCategoryGroupedMetrics: GroupedMetricsResponse = {
  measurements: [
    {
      type: "ACTIVE_JOBS",
      groups: [
        {
          key: "category",
          value: "Build",
        },
      ],
      values: [
        { value: "5", timestamp: "2024-03-01T00:00:00.000Z" },
        { value: "7", timestamp: "2024-03-01T00:00:15.000Z" },
        { value: "6", timestamp: "2024-03-01T00:00:30.000Z" },
        { value: "8", timestamp: "2024-03-01T00:00:45.000Z" },
        { value: "9", timestamp: "2024-03-01T00:01:00.000Z" },
        { value: "7", timestamp: "2024-03-01T00:01:15.000Z" },
        { value: "6", timestamp: "2024-03-01T00:01:30.000Z" },
        { value: "8", timestamp: "2024-03-01T00:01:45.000Z" },
        { value: "10", timestamp: "2024-03-01T00:02:00.000Z" },
        { value: "9", timestamp: "2024-03-01T00:02:15.000Z" },
        { value: "7", timestamp: "2024-03-01T00:02:30.000Z" },
        { value: "8", timestamp: "2024-03-01T00:02:45.000Z" },
        { value: "6", timestamp: "2024-03-01T00:03:00.000Z" },
        { value: "5", timestamp: "2024-03-01T00:03:15.000Z" },
        { value: "7", timestamp: "2024-03-01T00:03:30.000Z" },
        { value: "8", timestamp: "2024-03-01T00:03:45.000Z" },
        { value: "9", timestamp: "2024-03-01T00:04:00.000Z" },
        { value: "11", timestamp: "2024-03-01T00:04:15.000Z" },
        { value: "10", timestamp: "2024-03-01T00:04:30.000Z" },
        { value: "8", timestamp: "2024-03-01T00:04:45.000Z" },
      ],
    },
    {
      type: "ACTIVE_JOBS",
      groups: [
        {
          key: "category",
          value: "Train",
        },
      ],
      values: [
        { value: "12", timestamp: "2024-03-01T00:00:00.000Z" },
        { value: "15", timestamp: "2024-03-01T00:00:15.000Z" },
        { value: "18", timestamp: "2024-03-01T00:00:30.000Z" },
        { value: "16", timestamp: "2024-03-01T00:00:45.000Z" },
        { value: "14", timestamp: "2024-03-01T00:01:00.000Z" },
        { value: "17", timestamp: "2024-03-01T00:01:15.000Z" },
        { value: "19", timestamp: "2024-03-01T00:01:30.000Z" },
        { value: "20", timestamp: "2024-03-01T00:01:45.000Z" },
        { value: "18", timestamp: "2024-03-01T00:02:00.000Z" },
        { value: "16", timestamp: "2024-03-01T00:02:15.000Z" },
        { value: "15", timestamp: "2024-03-01T00:02:30.000Z" },
        { value: "17", timestamp: "2024-03-01T00:02:45.000Z" },
        { value: "19", timestamp: "2024-03-01T00:03:00.000Z" },
        { value: "21", timestamp: "2024-03-01T00:03:15.000Z" },
        { value: "18", timestamp: "2024-03-01T00:03:30.000Z" },
        { value: "16", timestamp: "2024-03-01T00:03:45.000Z" },
        { value: "14", timestamp: "2024-03-01T00:04:00.000Z" },
        { value: "13", timestamp: "2024-03-01T00:04:15.000Z" },
        { value: "15", timestamp: "2024-03-01T00:04:30.000Z" },
        { value: "17", timestamp: "2024-03-01T00:04:45.000Z" },
      ],
    },
    {
      type: "ACTIVE_JOBS",
      groups: [
        {
          key: "category",
          value: "Deploy",
        },
      ],
      values: [
        { value: "3", timestamp: "2024-03-01T00:00:00.000Z" },
        { value: "4", timestamp: "2024-03-01T00:00:15.000Z" },
        { value: "2", timestamp: "2024-03-01T00:00:30.000Z" },
        { value: "5", timestamp: "2024-03-01T00:00:45.000Z" },
        { value: "6", timestamp: "2024-03-01T00:01:00.000Z" },
        { value: "4", timestamp: "2024-03-01T00:01:15.000Z" },
        { value: "3", timestamp: "2024-03-01T00:01:30.000Z" },
        { value: "5", timestamp: "2024-03-01T00:01:45.000Z" },
        { value: "7", timestamp: "2024-03-01T00:02:00.000Z" },
        { value: "6", timestamp: "2024-03-01T00:02:15.000Z" },
        { value: "4", timestamp: "2024-03-01T00:02:30.000Z" },
        { value: "5", timestamp: "2024-03-01T00:02:45.000Z" },
        { value: "3", timestamp: "2024-03-01T00:03:00.000Z" },
        { value: "2", timestamp: "2024-03-01T00:03:15.000Z" },
        { value: "4", timestamp: "2024-03-01T00:03:30.000Z" },
        { value: "5", timestamp: "2024-03-01T00:03:45.000Z" },
        { value: "6", timestamp: "2024-03-01T00:04:00.000Z" },
        { value: "8", timestamp: "2024-03-01T00:04:15.000Z" },
        { value: "7", timestamp: "2024-03-01T00:04:30.000Z" },
        { value: "5", timestamp: "2024-03-01T00:04:45.000Z" },
      ],
    },
  ],
};
