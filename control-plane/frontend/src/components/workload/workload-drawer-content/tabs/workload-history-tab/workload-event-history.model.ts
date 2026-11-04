/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export type WorkloadHistoryType = "logs" | "user-actions";

export const workloadHistoryTypes: { label: string; value: WorkloadHistoryType }[] = [
  { label: "Logs", value: "logs" },
  { label: "User actions", value: "user-actions" },
];
