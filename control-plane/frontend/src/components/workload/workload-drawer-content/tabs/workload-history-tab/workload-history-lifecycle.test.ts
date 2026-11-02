/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import WorkloadEventHistory from "./workload-event-history.vue";
import { EIntervalLabels } from "@/models/interval.model";
import type { Workload } from "@/swagger-models/workloads-client";

const { loadHistory, startInterval, stopInterval, notify } = vi.hoisted(() => ({
  loadHistory: vi.fn(),
  startInterval: vi.fn(),
  stopInterval: vi.fn(),
  notify: vi.fn(),
}));
vi.mock("@/stores/workload.store", () => ({
  useWorkloadStore: () => ({
    loadWorkloadHistory: loadHistory,
    getWorkloadHistoryList: () => [],
    workloadHistoryFilterBy: {},
  }),
}));
vi.mock("@/utils/interval.util", () => ({ intervalUtil: { startInterval, stopInterval } }));
let wrapper: VueWrapper | undefined;
beforeEach(() => vi.clearAllMocks());
afterEach(() => wrapper?.unmount());

function render(): VueWrapper {
  wrapper = mount(WorkloadEventHistory, {
    props: { workload: { id: "workload", clusterId: "cluster" } as Workload, drawerWidth: "654px" },
    global: {
      mocks: { $q: { notify } },
      stubs: {
        RunaiSelect: { name: "RunaiSelect", props: ["modelValue", "options"], template: "<div />" },
        WorkloadActivityTab: true,
        RunaiDefaultLoader: true,
        WorkloadPhaseHistoryChart: true,
        WorkloadEventHistoryTable: true,
      },
    },
  });
  return wrapper;
}

async function selectActions(): Promise<void> {
  wrapper?.findComponent({ name: "RunaiSelect" }).vm.$emit("update:model-value", "user-actions");
  await flushPromises();
}

describe("Event history loading lifecycle", () => {
  it("starts polling after the active Logs view finishes loading", async () => {
    loadHistory.mockResolvedValueOnce(undefined);
    render();
    await flushPromises();
    expect(startInterval).toHaveBeenCalledWith(EIntervalLabels.WorkloadHistoryTab, expect.any(Function));
  });

  it("does not restart Logs polling after switching to User actions during loading", async () => {
    let resolve!: () => void;
    loadHistory.mockReturnValueOnce(
      new Promise<void>((done) => {
        resolve = done;
      }),
    );
    render();
    expect(loadHistory).toHaveBeenCalledOnce();
    await selectActions();
    expect(stopInterval).toHaveBeenCalledWith(EIntervalLabels.WorkloadHistoryTab);
    resolve();
    await flushPromises();
    expect(startInterval).not.toHaveBeenCalled();
  });

  it("does not notify about a Logs request that fails after switching modes", async () => {
    let reject!: (error: Error) => void;
    loadHistory.mockReturnValueOnce(
      new Promise<void>((_, fail) => {
        reject = fail;
      }),
    );
    render();
    await selectActions();
    reject(new Error("late Logs failure"));
    await flushPromises();
    expect(notify).not.toHaveBeenCalled();
    expect(startInterval).not.toHaveBeenCalled();
  });
});
