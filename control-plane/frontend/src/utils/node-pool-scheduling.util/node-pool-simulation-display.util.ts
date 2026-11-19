/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
export interface ISimulationDisplaySource {
  demand: number;
  allocated: number;
  guaranteed: number;
}

// Round totals first, then derive their parts so every displayed sum stays consistent.
// API values remain untouched for allocation calculations and bar geometry.
export function integerSimulationValues(result: ISimulationDisplaySource) {
  const demand = Math.max(0, Math.round(result.demand));
  const allocated = Math.min(demand, Math.max(0, Math.round(result.allocated)));
  const guaranteed = Math.min(allocated, Math.max(0, Math.round(result.guaranteed)));
  return { demand, allocated, guaranteed, overQuota: allocated - guaranteed, unmet: demand - allocated };
}
