/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { onScopeDispose, ref, watch, type Ref } from "vue";
import { buildSimulationRequest, mapSimulationResponse, type SimulationScenario } from "./node-pool-simulation.util";
import type { SimulationTransport } from "./node-pool-simulation.service";
export function useNodePoolSimulation(source: Ref<SimulationScenario | undefined>, transport: SimulationTransport) {
  const results = ref<ReturnType<typeof mapSimulationResponse>>();
  const status = ref<"idle" | "loading" | "ready" | "error" | "unavailable">("idle");
  const message = ref("");
  let revision = 0;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let controller: AbortController | undefined;
  function cancel() {
    revision++;
    clearTimeout(timer);
    controller?.abort();
  }
  function retry() {
    cancel();
    results.value = undefined;
    message.value = "";
    const scenario = source.value;
    if (!scenario) {
      status.value = "idle";
      return;
    }
    const current = revision;
    let request: ReturnType<typeof buildSimulationRequest>;
    try {
      request = buildSimulationRequest(scenario);
    } catch (error) {
      status.value = "unavailable";
      message.value = (error as Error).message;
      return;
    }
    status.value = "loading";
    timer = setTimeout(async () => {
      controller = new AbortController();
      try {
        const response = await transport(scenario.poolId, request, controller.signal);
        if (current !== revision) return;
        results.value = mapSimulationResponse(response);
        status.value = "ready";
      } catch (error) {
        if (current !== revision) return;
        const code = (error as { response?: { status?: number } }).response?.status;
        const unavailable = code === 404 || code === 501;
        status.value = unavailable ? "unavailable" : "error";
        message.value = unavailable
          ? "Simulation is not available on this backend yet. Your changes are preserved."
          : "Simulation failed. Your changes are preserved. Try again.";
      }
    }, 300);
  }
  watch(source, retry, { deep: true, immediate: true, flush: "sync" });
  onScopeDispose(cancel);
  return { results, status, message, retry };
}
