/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export const uniqueEmailChannelRunaiId = "runai-email";

export const workloadNotificationCategories = ["Workload status change", "Workload timeout"];

export enum NotificationChannelLabels {
  EMAIL = "Email",
  SLACK = "Slack",
}

export interface CategorizedNotificationSettings {
  [category: string]: NotificationSettings[];
}

export interface NotificationSettings {
  type: string;
  enabled: boolean;
  category: string;
  description?: string;
}
