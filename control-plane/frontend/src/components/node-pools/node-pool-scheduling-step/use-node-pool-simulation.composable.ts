/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { onScopeDispose, ref, watch, type Ref } from "vue";
import { buildSimulationRequest, mapSimulationResponse, type ISimulationScenario } from "./node-pool-simulation.util";
import type { SimulationTransport } from "./node-pool-simulation.service";
export enum ESimulationStatus {
  Idle = "idle",
  Loading = "loading",
  Ready = "ready",
  Error = "error",
  Unavailable = "unavailable",
}
export function useNodePoolSimulation(source: Ref<ISimulationScenario | undefined>, transport: SimulationTransport) {
  const results = ref<ReturnType<typeof mapSimulationResponse>>();
  const status = ref<ESimulationStatus>(ESimulationStatus.Idle);
  const message = ref("");
  const canRetry = ref(false);
  let requestKey: string | undefined;
  let resultContext: string | undefined;
  let revision = 0;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let controller: AbortController | undefined;
  function cancel() {
    revision++;
    clearTimeout(timer);
    controller?.abort();
  }
  function update(force = false) {
    const scenario = source.value;
    if (!scenario) {
      cancel();
      requestKey = undefined;
      results.value = undefined;
      message.value = "";
      canRetry.value = false;
      status.value = ESimulationStatus.Idle;
      return;
    }
    let request: ReturnType<typeof buildSimulationRequest>;
    try {
      request = buildSimulationRequest(scenario);
    } catch (error) {
      cancel();
      requestKey = undefined;
      results.value = undefined;
      status.value = ESimulationStatus.Unavailable;
      canRetry.value = false;
      message.value = (error as Error).message;
      return;
    }
    const key = JSON.stringify([scenario.poolId, request]);
    // Draft settings outside the API payload must not cancel or restart a simulation.
    if (!force && key === requestKey) return;
    cancel();
    requestKey = key;
    message.value = "";
    canRetry.value = false;
    const context = JSON.stringify([scenario.poolId, scenario.capacity, scenario.draft.mode]);
    if (context !== resultContext) results.value = undefined;
    const current = revision;
    status.value = ESimulationStatus.Loading;
    timer = setTimeout(async () => {
      controller = new AbortController();
      try {
        const response = await transport(scenario.poolId, request, controller.signal);
        if (current !== revision) return;
        results.value = mapSimulationResponse(response);
        resultContext = context;
        status.value = ESimulationStatus.Ready;
      } catch (error) {
        if (current !== revision) return;
        results.value = undefined;
        const response = (error as { response?: { status?: number; data?: { message?: unknown } } }).response;
        const code = response?.status;
        const unavailable = code === 404;
        status.value = unavailable ? ESimulationStatus.Unavailable : ESimulationStatus.Error;
        canRetry.value = code !== 400 && code !== 403;
        message.value =
          code === 400
            ? typeof response?.data?.message === "string" && response.data.message.trim()
              ? response.data.message
              : "Simulation inputs are invalid. Your changes are preserved."
            : code === 403
              ? "You’re missing permissions to simulate GPU allocation. Your changes are preserved."
              : unavailable
                ? "Simulation is unavailable. Your changes are preserved."
                : "Simulation failed. Your changes are preserved. Try again.";
      }
    }, 300);
  }
  const retry = () => update(true);
  watch(source, () => update(), { deep: true, immediate: true, flush: "sync" });
  onScopeDispose(cancel);
  return { results, status, message, canRetry, retry };
}
