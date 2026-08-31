/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export const MAX_REPLICAS_ALLOWED = 10000;

export interface IReplicaScalingSection {
  minReplicas: number | undefined;
  maxReplicas: number | undefined;
}

export enum EReplicaScalingSectionEvents {
  UpdateModelValue = "update:model-value",
}
