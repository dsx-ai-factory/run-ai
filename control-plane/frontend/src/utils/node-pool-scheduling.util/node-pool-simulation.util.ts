/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { isSupportedSchedulingMode } from "@/models/node-pool.model";
import {
  FairShareSimulationRank,
  type FairShareSimulationRequest,
  type FairShareSimulationResponse,
} from "@/swagger-models/org-unit-service-client/index";
import {
  toSchedulingCapabilities,
  validateDepartmentValues,
  type SchedulingDraft,
} from "@/models/node-pool-scheduling.model";
export interface ISimulationScenario {
  poolId: string;
  capacity: number;
  draft: SchedulingDraft;
  ids: string[];
  demand: Record<string, number>;
}
// Initial slider position only. The backend owns the default demand and limit cap.
export const defaultSimulationDemand = (capacity: number) => capacity / 2;
export function buildSimulationRequest(scenario: ISimulationScenario): FairShareSimulationRequest {
  const { capacity, draft, demand } = scenario;
  if (!Number.isFinite(capacity) || capacity < 0) throw new Error("Node pool capacity is unavailable.");
  if (!isSupportedSchedulingMode(draft.mode)) throw new Error("Select a supported scheduling mode to simulate.");
  const capabilities = toSchedulingCapabilities(draft);
  // orgUnits serves both as requested rows and scenario overrides. Include all
  // off-screen edits; splitting requests would silently change the competition.
  const ids = [...new Set([...scenario.ids, ...Object.keys(draft.departments), ...Object.keys(demand)])];
  if (!ids.length || ids.length > 500)
    throw new Error("Simulation supports 1–500 departments including off-screen edits. Your draft is preserved.");
  return {
    nodePool: {
      capacity: { gpu: capacity },
      schedulingCapabilities: {
        quota: !!capabilities.quota,
        ranks: !!capabilities.ranks,
        adjustableWeights: !!capabilities.adjustableWeights,
        limits: !!capabilities.limits,
      },
    },
    orgUnitType: "department",
    orgUnits: ids.map((id) => {
      const edit = draft.departments[id];
      const rank = capabilities.ranks ? edit?.rank : undefined;
      if (rank !== undefined && !Object.values(FairShareSimulationRank).some((value) => value === rank))
        throw new Error("Select a valid rank to simulate.");
      const gpu = {
        ...(capabilities.quota && edit?.quota !== undefined ? { deserved: edit.quota } : {}),
        ...(capabilities.limits && edit?.limit !== undefined ? { limit: edit.limit ?? -1 } : {}),
        ...(capabilities.adjustableWeights && edit?.weight !== undefined ? { overQuotaWeight: edit.weight } : {}),
      };
      const errors = validateDepartmentValues({
        id,
        weight: gpu.overQuotaWeight,
        quota: gpu.deserved,
        limit: gpu.limit,
      });
      if (errors.length) throw new Error(errors[0]);
      return {
        id,
        ...(rank !== undefined ? { rank: rank as FairShareSimulationRank } : {}),
        ...(Object.keys(gpu).length ? { gpu } : {}),
        ...(demand[id] !== undefined ? { demand: { gpu: demand[id] } } : {}),
      };
    }),
  };
}
export function mapSimulationResponse(response: FairShareSimulationResponse) {
  return Object.fromEntries(
    response.orgUnits.map((row) => [
      row.id,
      {
        allocated: row.fairShare.gpu,
        guaranteed: row.guaranteed.gpu,
        overQuota: row.overQuota.gpu,
        unmet: row.unmet.gpu,
        demand: row.demand.gpu,
        weightPercentage: row.weightShareInRank * 100,
      },
    ]),
  );
}
