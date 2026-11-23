/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises } from "@vue/test-utils";
import { createMemoryHistory, createRouter, RouterView } from "vue-router";
import { createTestingPinia } from "@pinia/testing";
import { Quasar, Notify } from "quasar";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { usePermissionStore } from "@/stores/permissions.store";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import View from "./node-pool-topology.vue";
import Step from "@/components/node-pools/node-pool-topology-step/node-pool-topology-step.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: { getNodepool: vi.fn() },
}));
vi.mock("@/composables/node-pool/node-pool-scheduling-data.util", () => ({
  loadSchedulingResources: vi.fn().mockResolvedValue({ nodes: 0, gpus: 0 }),
}));
const pool = { ...nodePoolUtil.createNodePoolDraft("cluster"), id: "one", name: "Training" } as Nodepool;
async function render(realStep = false, path = "/one?stepName=topology") {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/:id", name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT, component: View },
      { path: "/", name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX, component: { template: "<div>Pool list</div>" } },
    ],
  });
  await router.push(path);
  await router.isReady();
  const pinia = createTestingPinia({ createSpy: vi.fn });
  vi.mocked(usePermissionStore(pinia).hasPermission).mockReturnValue(true);
  const wrapper = mount(RouterView, {
    global: {
      plugins: [[Quasar, { plugins: { Notify } }], pinia, router],
      stubs: { NetworkTopologySelection: true, ...(realStep ? {} : { NodePoolTopologyStep: true }) },
    },
  });
  await flushPromises();
  return { wrapper, router };
}
beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(clusterService.getNodepool).mockResolvedValue(pool);
});
describe("topology route", () => {
  it("uses the shared page loader while fetching the pool and keeps Close available", async () => {
    let resolvePool!: (value: Nodepool) => void;
    vi.mocked(clusterService.getNodepool).mockImplementationOnce(
      () =>
        new Promise<Nodepool>((resolve) => {
          resolvePool = resolve;
        }),
    );
    const { wrapper } = await render();
    await flushPromises();
    expect(wrapper.find(".runai-default-loader").exists()).toBe(true);
    expect(wrapper.text()).not.toContain("Loading node pool");
    expect(wrapper.findComponent(Step).exists()).toBe(false);
    expect(await requestToLeave()).toBe(true);
    resolvePool(pool);
    await flushPromises();
    expect(wrapper.find(".runai-default-loader").exists()).toBe(false);
    expect(wrapper.findComponent(Step).props("pool").id).toBe("one");
    wrapper.unmount();
  });

  it("loads a direct URL by saved pool ID and retries failed reads", async () => {
    vi.mocked(clusterService.getNodepool).mockRejectedValueOnce(new Error("Unavailable"));
    const { wrapper } = await render();
    expect(wrapper.get('[role="alert"]').text()).toContain("Unavailable");
    await wrapper.get('[aid="retry-topology-pool"]').trigger("click");
    await flushPromises();
    expect(wrapper.findComponent(Step).props("pool").id).toBe("one");
    wrapper.unmount();
  });
  it("blocks route changes during a save and leaves after confirmed persistence", async () => {
    const { wrapper, router } = await render();
    wrapper.findComponent(Step).vm.$emit("saving", true);
    await flushPromises();
    await router.push("/");
    expect(router.currentRoute.value.params.id).toBe("one");
    wrapper.findComponent(Step).vm.$emit("saving", false);
    wrapper.findComponent(Step).vm.$emit("saved", pool);
    await flushPromises();
    expect(router.currentRoute.value.name).toBe(NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX);
    wrapper.unmount();
  });
  it("lets the user stay or discard when leaving a dirty topology step", async () => {
    const { wrapper, router } = await render(true);
    const step = wrapper.findComponent(Step);
    const radios = step.findAllComponents({ name: "runai-radio-options" });
    radios[0].vm.$emit("update:model-value", "dont-use");
    // Editing the supported placement field works independently of the cluster metadata fixture.
    radios[1].vm.$emit("update:model-value", "spread");
    await flushPromises();
    let done = false;
    void router.push("/").then(() => {
      done = true;
    });
    await flushPromises();
    expect(done).toBe(false);
    (document.querySelector('[aid="leave-page-modal-stay-btn"]') as HTMLButtonElement).click();
    await flushPromises();
    expect(router.currentRoute.value.params.id).toBe("one");
    void router.push("/");
    await flushPromises();
    (document.querySelector('[aid="leave-page-modal-cancel-btn"]') as HTMLButtonElement).click();
    await flushPromises();
    expect(router.currentRoute.value.name).toBe(NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX);
    wrapper.unmount();
  });
});
