/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, afterEach } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { defineComponent, onUnmounted } from "vue";
import WorkloadEventHistory from "./workload-event-history.vue";
import type { Workload } from "@/swagger-models/workloads-client";

const stoppedLogs = vi.fn();
const stoppedActions = vi.fn();
let wrapper: VueWrapper | undefined;
const render = (extra = {}) => {
  wrapper = mount(WorkloadEventHistory, {
    props: { workload: { id: "one", clusterId: "cluster" } as Workload, drawerWidth: "654px", ...extra },
    global: {
      stubs: {
        RunaiSelect: {
          name: "RunaiSelect",
          props: ["modelValue", "options"],
          template: '<div aid="type">{{ modelValue }}</div>',
        },
        WorkloadHistoryTab: defineComponent({
          name: "WorkloadHistoryTab",
          setup() {
            onUnmounted(stoppedLogs);
          },
          template: '<div aid="logs">Logs</div>',
        }),
        WorkloadActivityTab: defineComponent({
          name: "WorkloadActivityTab",
          setup() {
            onUnmounted(stoppedActions);
          },
          template: '<div aid="actions">User actions</div>',
        }),
      },
    },
  });
  return wrapper;
};
afterEach(() => {
  wrapper?.unmount();
  vi.clearAllMocks();
});
const choose = async (value: string) => {
  wrapper?.findComponent({ name: "RunaiSelect" }).vm.$emit("update:model-value", value);
  await wrapper?.vm.$nextTick();
};

describe("Event history type selector", () => {
  it("defaults to Logs and offers exactly Logs and User actions", () => {
    const view = render();
    expect(view.find('[aid="logs"]').exists()).toBe(true);
    expect(view.find('[aid="actions"]').exists()).toBe(false);
    expect(view.findComponent({ name: "RunaiSelect" }).props("options")).toEqual([
      { label: "Logs", value: "logs" },
      { label: "User actions", value: "user-actions" },
    ]);
  });
  it("unmounts the inactive view so its requests and polling are cleaned up", async () => {
    const view = render();
    await choose("user-actions");
    expect(view.emitted("update:initialType")?.at(-1)).toEqual(["user-actions"]);
    expect(view.find('[aid="actions"]').exists()).toBe(true);
    expect(view.find('[aid="logs"]').exists()).toBe(false);
    expect(stoppedLogs).toHaveBeenCalledTimes(1);
    await choose("logs");
    expect(stoppedActions).toHaveBeenCalledTimes(1);
    expect(view.find('[aid="logs"]').exists()).toBe(true);
  });
  it("resets to Logs when selecting another workload", async () => {
    const view = render();
    await choose("user-actions");
    await view.setProps({ workload: { id: "two", clusterId: "cluster" } as Workload });
    expect(view.find('[aid="logs"]').exists()).toBe(true);
  });
  it("supports existing Activity links as User actions inside Event history", () => {
    expect(render({ initialType: "user-actions" }).find('[aid="actions"]').exists()).toBe(true);
  });
  it("hides the selector and preserves the element empty state", () => {
    const view = render({ showElementEmptyState: true, initialType: "user-actions" });
    expect(view.findComponent({ name: "RunaiSelect" }).exists()).toBe(false);
    expect(view.find('[aid="actions"]').exists()).toBe(false);
    expect(view.findComponent({ name: "WorkloadHistoryTab" }).attributes("show-element-empty-state")).toBe("true");
  });
});
