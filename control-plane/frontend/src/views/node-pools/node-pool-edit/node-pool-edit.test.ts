/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { createRouter, createMemoryHistory } from "vue-router";
import { Quasar } from "quasar";
import { describe, expect, it, vi } from "vitest";
import NodePoolApp from "../node-pool-app.vue";
import NodePoolEdit from "./node-pool-edit.vue";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
vi.mock("@/constants/node-pool.constants", () => ({ ENABLE_NODE_POOL_CREATE_FLOW: true }));
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNodepool: vi.fn().mockResolvedValue({ id: "saved", name: "saved-pool", clusterId: "cluster" }),
    getNodepools: vi.fn().mockResolvedValue([]),
  },
}));

describe("shared node pool new/edit entry", () => {
  it("loads the saved pool when Create & continue navigates from new to edit", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        {
          path: "/node-pools",
          component: NodePoolApp,
          children: [
            { path: "new", name: NODE_POOL_ROUTE_NAMES.NODE_POOL_NEW, component: NodePoolEdit },
            { path: "edit/:id", name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT, component: NodePoolEdit },
          ],
        },
      ],
    });
    await router.push("/node-pools/new");
    await router.isReady();
    const wrapper = mount(NodePoolApp, {
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn }), router],
        stubs: {
          NodePoolEditFormV2: {
            name: "NodePoolEditFormV2",
            emits: ["created", "leave-approved"],
            template: "<p>New node pool</p>",
          },
          NodePoolEditForm: { props: ["nodePool"], template: "<p>Edit {{nodePool.name}}</p>" },
        },
      },
    });
    expect(wrapper.text()).toContain("New node pool");
    const form = wrapper.findComponent({ name: "NodePoolEditFormV2" });
    form.vm.$emit("leave-approved");
    form.vm.$emit("created", { id: "saved", name: "saved-pool", clusterId: "cluster" }, {}, "continue");
    await flushPromises();
    expect(router.currentRoute.value.query.stepName).toBe("scheduling");
    expect(clusterService.getNodepool).toHaveBeenCalledWith("saved");
    expect(wrapper.text()).toContain("Edit saved-pool");
    await router.push("/node-pools/new");
    await flushPromises();
    expect(wrapper.text()).toContain("New node pool");
    wrapper.unmount();
  });
});
