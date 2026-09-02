/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ISelectOption } from "@/models/global.model";

export const NEVER_SCALE_DOWN_ID = "never";
export const DEFAULT_SCALE_DOWN_TO_ZERO_ID = "60";
export const CUSTOM_SCALE_DOWN_ID = "custom";
export const SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE = 1;
export const MIN_REPLICAS_TO_ALLOW_SCALE_TO_ZERO_OPTIONS = 0;

export const SCALE_DOWN_OPTIONS: Array<ISelectOption> = [
  {
    value: NEVER_SCALE_DOWN_ID,
    label: "Never",
  },
  {
    value: DEFAULT_SCALE_DOWN_TO_ZERO_ID,
    label: "After 1 minute of inactivity",
  },
  {
    value: "300",
    label: "After 5 minutes of inactivity",
  },
  {
    value: "900",
    label: "After 15 minutes of inactivity",
  },
  {
    value: "1800",
    label: "After 30 minutes of inactivity",
  },
  {
    value: CUSTOM_SCALE_DOWN_ID,
    label: "Custom",
  },
];

export interface IScaleDownData {
  scaleToZeroRetentionSeconds: number | undefined;
}

export enum EScaleDownSectionEvents {
  UpdateModelValue = "update:modelValue",
}
