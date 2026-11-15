/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { InjectionKey } from "vue";
import { orgUnitServiceApi } from "@/services/infra/client-apis/org-unit-service-api/org-unit-service-api";
import type { FairShareSimulationRequest, FairShareSimulationResponse } from "@/swagger-models/org-unit-service-client";
export type SimulationTransport = (
  poolId: string,
  request: FairShareSimulationRequest,
  signal: AbortSignal,
) => Promise<FairShareSimulationResponse>;
export const simulateNodePool: SimulationTransport = async (poolId, request, signal) => {
  const response = await orgUnitServiceApi.orgUnitApi.simulateNodePoolFairShare(poolId, request, { signal });
  return response.data;
};

// Preview harnesses may inject a fixture transport; production always uses the API.
export const simulationPreviewKey: InjectionKey<{ transport: SimulationTransport }> =
  Symbol("node-pool-simulation-preview");
