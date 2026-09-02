/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export interface IAutoScaleOptions {
  latencyEnabled: boolean;
  isAutoScaleEnabled: boolean;
}

export interface IAutoScaleData {
  minReplicas: number | undefined;
  maxReplicas: number | undefined;
  thresholdMetric: string | undefined;
  thresholdValue: number | undefined;
  scaleToZeroRetentionSeconds: number | undefined;
}

export const MIN_REPLICAS_FOR_LATENCY_METRIC = 1;

export enum EAutoScaleSectionEvents {
  OnAutoScaleChanged = "on-auto-scale-changed",
}
