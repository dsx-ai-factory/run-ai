/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { TrainingSpecificRunParams, WorkspaceSpecificRunParams } from "@/swagger-models/assets-service-client";
import type { IUIWorkloadSpecificEnv } from "@/models/workload.model";

export interface IWorkloadSpecificRunParams extends WorkspaceSpecificRunParams, TrainingSpecificRunParams {}
export type SpecificRunParamsKeys = (keyof IUIWorkloadSpecificEnv)[];
