/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { onScopeDispose, ref, watch, type Ref } from "vue";
import {
  buildSimulationRequest,
  mapSimulationResponse,
  type ISimulationScenario,
} from "@/utils/node-pool-scheduling.util/node-pool-simulation.util";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
export enum ESimulationStatus {
  Idle = "idle",
  Loading = "loading",
  Ready = "ready",
  Error = "error",
  Unavailable = "unavailable",
}
export function useNodePoolSimulation(source: Ref<ISimulationScenario | undefined>) {
  const results = ref<ReturnType<typeof mapSimulationResponse>>();
  const status = ref<ESimulationStatus>(ESimulationStatus.Idle);
  const message = ref("");
  const canRetry = ref(false);
  let requestKey: string | undefined;
  type Mode = ISimulationScenario["draft"]["mode"];
  const lastResultByMode = new Map<Mode, { key: string; rows: ReturnType<typeof mapSimulationResponse> }>();
  let activeMode: Mode;
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
      resultContext = undefined;
      activeMode = undefined;
      lastResultByMode.clear();
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
      resultContext = undefined;
      activeMode = undefined;
      lastResultByMode.clear();
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
    const mode = scenario.draft.mode;
    const delay = mode === activeMode ? 300 : 0;
    activeMode = mode;
    if (context !== resultContext) {
      const cached = lastResultByMode.get(mode);
      results.value = cached?.key === key ? cached.rows : undefined;
      resultContext = context;
    }
    const current = revision;
    status.value = ESimulationStatus.Loading;
    timer = setTimeout(async () => {
      controller = new AbortController();
      try {
        const response = await orgUnitService.simulateNodePoolFairShare(scenario.poolId, request, controller.signal);
        if (current !== revision) return;
        const rows = mapSimulationResponse(response);
        lastResultByMode.set(mode, { key, rows });
        results.value = rows;
        status.value = ESimulationStatus.Ready;
      } catch (error) {
        if (current !== revision) return;
        results.value = undefined;
        lastResultByMode.delete(mode);
        const failure = (error ?? {}) as {
          statusCode?: number;
          message?: string;
          response?: { status?: number; data?: { message?: unknown } };
        };
        const response = failure.response;
        const code = failure.statusCode ?? response?.status;
        const serverMessage = failure.statusCode ? failure.message : response?.data?.message;
        const unavailable = code === 404;
        status.value = unavailable ? ESimulationStatus.Unavailable : ESimulationStatus.Error;
        canRetry.value = code !== 400 && code !== 403;
        message.value =
          code === 400
            ? typeof serverMessage === "string" && serverMessage.trim()
              ? serverMessage
              : "Simulation inputs are invalid. Your changes are preserved."
            : code === 403
              ? "You’re missing permissions to simulate GPU allocation. Your changes are preserved."
              : unavailable
                ? "Simulation is unavailable. Your changes are preserved."
                : "Simulation failed. Your changes are preserved. Try again.";
      }
    }, delay);
  }
  const retry = () => update(true);
  watch(source, () => update(), { deep: true, immediate: true, flush: "sync" });
  onScopeDispose(cancel);
  return { results, status, message, canRetry, retry };
}
