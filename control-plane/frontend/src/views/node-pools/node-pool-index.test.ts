/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { createTestingPinia } from "@pinia/testing";
import { mount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createMemoryHistory, createRouter } from "vue-router";
import { afterEach, describe, expect, it, vi } from "vitest";
import { usePermissionStore } from "@/stores/permissions.store";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import NodePoolIndex from "./node-pool-index.vue";

const flag = vi.hoisted(() => ({ enabled: true }));
vi.mock("@/constants/node-pool.constants", () => ({
  get ENABLE_NODE_POOL_CREATE_FLOW() {
    return flag.enabled;
  },
}));

async function renderGrid() {
  vi.spyOn(clusterService, "countNodepools").mockResolvedValue(1);
  vi.spyOn(clusterService, "getNodepools").mockResolvedValue([]);
  vi.spyOn(clusterService, "getNodeTelemetries").mockResolvedValue([]);
  const pinia = createTestingPinia({ createSpy: vi.fn });
  vi.mocked(usePermissionStore(pinia).hasPermission).mockReturnValue(true);
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/node-pools", component: NodePoolIndex },
      { path: "/node-pools/edit/:id", name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT, component: { template: "<div />" } },
    ],
  });
  await router.push("/node-pools");
  const wrapper = mount(NodePoolIndex, {
    attachTo: document.body,
    global: {
      plugins: [Quasar, pinia, router],
      directives: { permission: vi.fn() },
      stubs: {
        RunaiTableWrapper: { template: "<div><slot name='actions' /></div>" },
        RunaiPageActions: { template: "<div><slot name='selected-rows-actions' /></div>" },
      },
    },
  });
  await wrapper.setData({ selectedRows: [{ id: "100000", name: "default", phase: "Ready", isDefault: true }] });
  return { wrapper, router };
}

afterEach(() => {
  flag.enabled = true;
  vi.restoreAllMocks();
});

describe("node pool grid Edit action", () => {
  it.each([
    ["Description", "identity"],
    ["Scheduling mode", "scheduling"],
    ["Topology & placement", "topology"],
  ])("opens the %s step on the selected pool", async (label, stepName) => {
    const { wrapper, router } = await renderGrid();
    try {
      await wrapper.get('[aid="node-pool-edit-btn"]').trigger("click");
      await flushPromises();
      expect(router.currentRoute.value.path).toBe("/node-pools");
      const items = Array.from(document.querySelectorAll<HTMLElement>('[role="menuitem"]'));
      expect(items.map((item) => item.textContent?.trim())).toEqual([
        "Description",
        "Scheduling mode",
        "Topology & placement",
      ]);
      items.find((item) => item.textContent?.trim() === label)?.click();
      await flushPromises();
      expect(router.currentRoute.value.fullPath).toBe(`/node-pools/edit/100000?stepName=${stepName}`);
    } finally {
      wrapper.unmount();
    }
  });

  it("opens the original editor directly when the new wizard is off", async () => {
    flag.enabled = false;
    const { wrapper, router } = await renderGrid();
    try {
      await wrapper.get('[aid="node-pool-edit-btn"]').trigger("click");
      await flushPromises();
      expect(document.querySelector('[role="menuitem"]')).toBeNull();
      expect(router.currentRoute.value.fullPath).toBe("/node-pools/edit/100000");
    } finally {
      wrapper.unmount();
    }
  });
});
