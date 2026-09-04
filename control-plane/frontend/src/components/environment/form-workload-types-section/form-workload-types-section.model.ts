/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { EWorkloadName } from "@/models/workload.model";
import type { DistributedFramework } from "@/swagger-models/assets-service-client";

export interface IWorkloadTypesSectionModel {
  workloadTypes?: Record<EWorkloadName, boolean>;
  distributed: {
    isDistributed: boolean;
    distFramework: DistributedFramework | null;
  };
}
