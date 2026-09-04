/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export enum ENotificationBarVariant {
  Info = "info",
  Warning = "warning",
  Error = "error",
  Default = "default",
  Success = "success",
}

export interface INotificationBarCta {
  label: string;
  onClick: () => void;
}

export const variantConfig = {
  [ENotificationBarVariant.Info]: {
    background: "variant-info",
    icon: "ph-bold ph-info",
    iconColor: "variant-icon-info",
  },
  [ENotificationBarVariant.Warning]: {
    background: "variant-warning",
    icon: "ph-bold ph-warning-circle",
    iconColor: "variant-icon-warning",
  },
  [ENotificationBarVariant.Error]: {
    background: "variant-error",
    icon: "ph-fill ph-x-circle",
    iconColor: "variant-icon-error",
  },
  [ENotificationBarVariant.Default]: {
    background: "variant-default",
    icon: "ph-bold ph-question",
    iconColor: "variant-icon-default",
  },
  [ENotificationBarVariant.Success]: {
    background: "variant-success",
    icon: "ph-bold ph-check-circle",
    iconColor: "variant-icon-success",
  },
};
