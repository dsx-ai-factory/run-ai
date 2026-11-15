/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { FairShareSimulationRequest, FairShareSimulationResponse } from "@/swagger-models/org-unit-service-client";
import type { SimulationTransport } from "./node-pool-simulation.service";
// Authored contract fixture, never a scheduling implementation. A preview harness
// explicitly injects this transport. Production does not import this module.
export function createSimulationPreview(
  request: FairShareSimulationRequest,
  response: FairShareSimulationResponse,
): SimulationTransport {
  const expected = JSON.stringify(request);
  const fixture = JSON.stringify(response);
  return async (_poolId, actual, signal) => {
    if (signal.aborted) throw new DOMException("Aborted", "AbortError");
    if (JSON.stringify(actual) !== expected) throw new Error("No authored fixture for this scenario.");
    return JSON.parse(fixture);
  };
}
