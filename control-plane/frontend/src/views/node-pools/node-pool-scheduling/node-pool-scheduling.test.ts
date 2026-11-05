/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises } from "@vue/test-utils";
import { createMemoryHistory, createRouter, RouterView } from "vue-router";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, it, vi } from "vitest";
import View from "./node-pool-scheduling.vue";
import Step from "@/components/node-pools/node-pool-scheduling-step/node-pool-scheduling-step.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { createNodePoolDraft } from "@/components/node-pools/node-pool-create-flow/node-pool-create-flow.model";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { useAppStore } from "@/stores/app.store";

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: { getNodepool: vi.fn(), createNodepool: vi.fn(), patchNodepool: vi.fn(), updateNodepool: vi.fn() },
}));
const pool = { ...createNodePoolDraft("cluster"), id: "one", name: "training", clusterName: "Tokyo" } as Nodepool;
async function render(realStep = false) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/:id", component: View },
      { path: "/", component: { template: "<div>Pool list</div>" } },
    ],
  });
  await router.push("/one");
  await router.isReady();
  const wrapper = mount(RouterView, {
    global: {
      plugins: [router, Quasar, createTestingPinia({ createSpy: vi.fn })],
      stubs: realStep ? {} : { NodePoolSchedulingStep: true },
    },
  });
  return { wrapper, router };
}
beforeEach(() => vi.clearAllMocks());
describe("scheduling route", () => {
  it.each(["loading", "error"])("allows header Close in the %s state", async (state) => {
    vi.mocked(clusterService.getNodepool).mockImplementation(() =>
      state === "error" ? Promise.reject(new Error("Unavailable")) : new Promise(() => {}),
    );
    const { wrapper } = await render(true);
    await flushPromises();
    expect(useAppStore().setPageLoading).toHaveBeenCalledWith(false);
    let allowed: boolean | undefined;
    void requestToLeave().then((value) => {
      allowed = value;
    });
    await flushPromises();
    expect(allowed).toBe(true);
    wrapper.unmount();
  });

  it("uses header leave approval once without asking again during navigation", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue(pool);
    const { wrapper, router } = await render(true);
    await flushPromises();
    await wrapper.get('[aid="mode-ranked"]').trigger("click");
    const approved = requestToLeave();
    await flushPromises();
    (document.querySelector('[aid="leave-page-modal-cancel-btn"]') as HTMLButtonElement).click();
    expect(await approved).toBe(true);
    let navigated = false;
    void router.push("/").then(() => {
      navigated = true;
    });
    await flushPromises();
    expect(navigated).toBe(true);
    expect(wrapper.text()).toContain("Pool list");
    wrapper.unmount();
  });
  it("loads by route ID without creating or updating a pool", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue(pool);
    const { wrapper } = await render();
    await flushPromises();
    expect(clusterService.getNodepool).toHaveBeenCalledWith("one");
    expect(wrapper.findComponent(Step).props("pool")).toEqual(pool);
    expect(clusterService.createNodepool).not.toHaveBeenCalled();
    expect(clusterService.patchNodepool).not.toHaveBeenCalled();
    expect(clusterService.updateNodepool).not.toHaveBeenCalled();
    wrapper.unmount();
  });

  it("shows a load failure and retries without a stale screen", async () => {
    vi.mocked(clusterService.getNodepool)
      .mockRejectedValueOnce(new Error("Service unavailable"))
      .mockResolvedValueOnce(pool);
    const { wrapper } = await render();
    await flushPromises();
    expect(wrapper.find('[role="alert"]').text()).toContain("Service unavailable");
    expect(wrapper.findComponent(Step).exists()).toBe(false);
    await wrapper.get('[aid="retry-scheduling-pool"]').trigger("click");
    await flushPromises();
    expect(wrapper.findComponent(Step).props("pool").id).toBe("one");
    wrapper.unmount();
  });

  it("ignores a response for an older route", async () => {
    let resolveFirst!: (value: Nodepool) => void;
    vi.mocked(clusterService.getNodepool)
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveFirst = resolve;
          }),
      )
      .mockResolvedValueOnce({ ...pool, id: "two", name: "second" });
    const { wrapper, router } = await render();
    await router.push("/two");
    await flushPromises();
    resolveFirst(pool);
    await flushPromises();
    expect(wrapper.findComponent(Step).props("pool").id).toBe("two");
    wrapper.unmount();
  });
});
