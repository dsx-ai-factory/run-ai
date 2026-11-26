/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises } from "@vue/test-utils";
import { createMemoryHistory, createRouter } from "vue-router";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { describe, expect, it, vi } from "vitest";
import Step from "./node-pool-topology-step.vue";
import Selector from "@/components/node-pools/node-pool-edit-form/node-pool-network-section/network-topology-selection/network-topology-selection.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import type { NetworkTopology, Nodepool } from "@/swagger-models/cluster-service-client";

describe("permissions changing during topology creation", () => {
  it("keeps the pending request guarded until completion when topology Read is revoked", async () => {
    let finish!: (value: NetworkTopology) => void;
    vi.spyOn(clusterService, "getNetworkTopologies").mockResolvedValue([]);
    vi.spyOn(clusterService, "createNetworkTopology").mockImplementation(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        }),
    );
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: "/", component: { template: "<div />" } }],
    });
    await router.push("/");
    const wrapper = mount(Step, {
      props: {
        pool: { ...nodePoolUtil.createNodePoolDraft("cluster"), id: "pool", name: "Training" } as Nodepool,
        supports: { mnnvl: true, numa: true, gpuDevice: true },
        canSave: true,
        canReadTopologies: true,
        canCreateTopology: true,
      },
      global: {
        plugins: [Quasar, router, createTestingPinia({ createSpy: vi.fn })],
        stubs: { NetworkTopologyForm: true },
      },
    });
    try {
      await flushPromises();
      const selector = wrapper.findComponent(Selector);
      selector.vm.setModalOpen(true);
      const pending = selector.vm.createNetworkTopology({
        name: "Pending fabric",
        clusterId: "cluster",
        levels: ["rack"],
      });
      await flushPromises();
      expect(wrapper.emitted("saving")?.at(-1)).toEqual([true]);
      await wrapper.setProps({ canReadTopologies: false });
      expect(wrapper.findComponent(Selector).exists()).toBe(true);
      expect(wrapper.get('[aid="topology-save-close"]').attributes("disabled")).toBeDefined();
      expect(await requestToLeave()).toBe(false);
      finish({
        id: "fabric",
        name: "Pending fabric",
        clusterId: "cluster",
        levels: ["rack"],
        tenantId: 1,
        createdAt: "2026-11-17T00:00:00Z",
      });
      await pending;
      await flushPromises();
      expect(wrapper.findComponent(Selector).exists()).toBe(false);
      expect(wrapper.text()).toContain("permission to view network topologies");
      expect(wrapper.emitted("saving")?.at(-1)).toEqual([false]);
    } finally {
      wrapper.unmount();
      vi.restoreAllMocks();
    }
  });
});
