/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { TelemetryResponse } from "@/swagger-models/cluster-service-client";

export const idleAllocatedGpuWidgetMock: TelemetryResponse = {
  type: "IDLE_ALLOCATED_GPUS",
  timestamp: "2024-09-03T06:36:40.106935459Z",
  values: [
    {
      value: "2",
    },
  ],
};
