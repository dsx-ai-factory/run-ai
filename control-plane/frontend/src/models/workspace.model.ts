/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Workspace } from "@/swagger-models/assets-service-client";
import type { IToolItem } from "./workload.model";

export interface IExtendJobValues {
  usedMemory?: string;
  usedCPUs?: string;
  gpusUtilization?: string;
  usedGpuMemory?: string;
  currentAllocatedGPUsMemoryBytes?: number;
}
export interface IWorkspaceList extends Workspace {
  connections?: Array<IToolItem>;
  extendJobValues?: IExtendJobValues;
}
