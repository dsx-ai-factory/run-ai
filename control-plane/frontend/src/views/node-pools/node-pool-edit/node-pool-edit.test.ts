/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { createRouter, createMemoryHistory } from "vue-router";
import { Notify, Quasar } from "quasar";
import { ForbiddenError, NotFoundError } from "@/models/http-response.model";
import { usePermissionStore } from "@/stores/permissions.store";
import { useAppStore } from "@/stores/app.store";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { describe, expect, it, vi } from "vitest";
import NodePoolApp from "../node-pool-app.vue";
import NodePoolEdit from "./node-pool-edit.vue";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
vi.mock("@/constants/node-pool.constants", () => ({ ENABLE_NODE_POOL_CREATE_FLOW: true }));
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNodepool: vi.fn().mockResolvedValue({ id: "saved", name: "saved-pool", clusterId: "cluster" }),
    patchNodepool: vi.fn(),
    getNodepools: vi.fn().mockResolvedValue([]),
  },
}));

describe("shared node pool new/edit entry", () => {
  it.each(["identity", "scheduling", "topology"])(
    "shows a retry screen at the direct %s URL for a deleted pool",
    async (step) => {
      const poolId = "deleted-pool";
      const forbidden = new ForbiddenError();
      const notFound = new NotFoundError("Node pool not found");
      vi.mocked(clusterService.getNodepool).mockClear().mockRejectedValueOnce(forbidden).mockRejectedValueOnce(notFound);
      const router = createRouter({
        history: createMemoryHistory(),
        routes: [{ path: "/node-pools/edit/:id", name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT, component: NodePoolEdit }],
      });
      await router.push({
        name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT,
        params: { id: poolId },
        query: { stepName: step },
      });
      await router.isReady();
      const wrapper = mount(NodePoolApp, {
        global: { plugins: [Quasar, createTestingPinia({ createSpy: vi.fn }), router] },
      });
      try {
        const retry = `[aid="retry-${step}-pool"]`;
        const save = `[aid="${step}-save-close"]`;
        await vi.waitFor(() => expect(wrapper.get('[role="alert"]').text()).toContain(forbidden.message), {
          timeout: 5000,
        });
        expect(wrapper.find(save).exists()).toBe(false);
        expect(useAppStore().setPageTitleOverride).toHaveBeenLastCalledWith(null);
        expect(await requestToLeave()).toBe(true);

        await wrapper.get(retry).trigger("click");
        await flushPromises();
        expect(wrapper.get('[role="alert"]').text()).toContain(notFound.message);
        expect(wrapper.get(retry).isVisible()).toBe(true);
        expect(wrapper.find(save).exists()).toBe(false);
        expect(vi.mocked(clusterService.getNodepool).mock.calls).toEqual([[poolId], [poolId]]);
        expect(router.currentRoute.value.query.stepName).toBe(step);
      } finally {
        wrapper.unmount();
      }
    },
  );

  it("opens saved identity, locks its four fixed fields, and saves only the description before continuing", async () => {
    const snapshot = {
      id: "saved",
      clusterId: "cluster",
      clusterName: "Cluster A",
      name: "my-pool",
      labelKey: "rack",
      labelValue: "west",
      description: "original",
    };
    vi.mocked(clusterService.getNodepool).mockResolvedValueOnce(
      snapshot as Awaited<ReturnType<typeof clusterService.getNodepool>>,
    );
    vi.mocked(clusterService.patchNodepool).mockResolvedValueOnce({ ...snapshot, description: "edited" } as Awaited<
      ReturnType<typeof clusterService.patchNodepool>
    >);
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: "/node-pools/edit/:id", name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT, component: NodePoolEdit }],
    });
    await router.push("/node-pools/edit/saved");
    const pinia = createTestingPinia({ createSpy: vi.fn });
    vi.mocked(usePermissionStore(pinia).hasPermission).mockReturnValue(true);
    const wrapper = mount(NodePoolApp, {
      global: { plugins: [[Quasar, { plugins: { Notify } }], pinia, router], stubs: { NodePoolScheduling: true } },
    });
    try {
      await vi.waitFor(() => expect(wrapper.find('[aid="identity-save-continue"]').exists()).toBe(true), {
        timeout: 5000,
      });
      const locked = wrapper.findAll(".locked-field input");
      expect(locked.map((field) => (field.element as HTMLInputElement).value)).toEqual([
        "Cluster A",
        "my-pool",
        "rack",
        "west",
      ]);
      expect(locked.every((field) => field.attributes("disabled") !== undefined)).toBe(true);
      const description = wrapper.get("textarea");
      expect((description.element as HTMLTextAreaElement).value).toBe("original");
      await description.setValue("edited");
      await wrapper.get('[aid="identity-save-continue"]').trigger("click");
      await vi.waitFor(() => expect(router.currentRoute.value.query.stepName).toBe("scheduling"));
      expect(vi.mocked(clusterService.patchNodepool).mock.calls).toEqual([["saved", { description: "edited" }]]);
    } finally {
      wrapper.unmount();
    }
  });

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
          NodePoolScheduling: { template: "<p>Scheduling saved-pool</p>" },
          NodePoolTopology: { template: "<p>Topology saved-pool</p>" },
          NodePoolIdentity: { template: "<p>Cluster &amp; name</p>" },
        },
      },
    });
    expect(wrapper.text()).toContain("New node pool");
    const form = wrapper.findComponent({ name: "NodePoolEditFormV2" });
    form.vm.$emit("leave-approved");
    form.vm.$emit("created", { id: "saved", name: "saved-pool", clusterId: "cluster" }, {}, "continue");
    await flushPromises();
    expect(router.currentRoute.value.query.stepName).toBe("scheduling");
    expect(wrapper.text()).toContain("Scheduling saved-pool");
    vi.mocked(clusterService.getNodepool).mockClear();
    await router.push("/node-pools/edit/saved?stepName=topology");
    await flushPromises();
    expect(wrapper.text()).toContain("Topology saved-pool");
    expect(clusterService.getNodepool).not.toHaveBeenCalled();
    await router.push("/node-pools/edit/saved");
    await flushPromises();
    expect(wrapper.text()).toContain("Cluster & name");
    await router.push("/node-pools/new");
    await flushPromises();
    expect(wrapper.text()).toContain("New node pool");
    wrapper.unmount();
  });
});
