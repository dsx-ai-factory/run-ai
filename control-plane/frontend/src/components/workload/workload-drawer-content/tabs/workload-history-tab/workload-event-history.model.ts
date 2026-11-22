/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export enum EWorkloadHistoryType {
  Logs = "logs",
  UserActions = "user-actions",
}

export type WorkloadHistoryType = `${EWorkloadHistoryType}`;

export const workloadHistoryTypes: { label: string; value: WorkloadHistoryType }[] = [
  { label: "Logs", value: EWorkloadHistoryType.Logs },
  { label: "User actions", value: EWorkloadHistoryType.UserActions },
];
