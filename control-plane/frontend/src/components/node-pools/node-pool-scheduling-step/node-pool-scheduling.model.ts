/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { SchedulingCapabilitiesNullable } from "@/swagger-models/cluster-service-client";

export type SchedulingMode = "proportional" | "ranked" | "guaranteed";
export interface DepartmentSchedulingEdit {
  id: string;
  weight?: number;
  quota?: number;
  limit?: number | null;
  rank?: string;
}
export interface SchedulingSnapshot {
  mode?: SchedulingMode;
  automaticWeights: boolean;
  cpuQuota: boolean;
  departments: Record<string, DepartmentSchedulingEdit>;
  capabilities?: SchedulingCapabilitiesNullable;
}
export type SchedulingDraft = SchedulingSnapshot;
export function createSchedulingDraft(snapshot: SchedulingSnapshot): SchedulingDraft {
  return JSON.parse(JSON.stringify(snapshot));
}
export function toSchedulingCapabilities(draft: SchedulingDraft): SchedulingCapabilitiesNullable {
  if (!draft.mode) return { ...draft.capabilities };
  return {
    adjustableWeights: draft.mode !== "guaranteed" || !draft.automaticWeights,
    quota: draft.mode === "guaranteed",
    ranks: draft.mode !== "proportional",
    cpuQuota: draft.mode === "guaranteed" && draft.cpuQuota,
  };
}
