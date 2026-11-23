/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises } from "@vue/test-utils";
import { createMemoryHistory, createRouter, RouterView } from "vue-router";
import { createTestingPinia } from "@pinia/testing";
import { Notify, Quasar } from "quasar";
import { beforeEach, describe, expect, it, vi } from "vitest";
import View from "./node-pool-scheduling.vue";
import { NodePoolSchedulingStep as Step } from "@/components/node-pools/node-pool-scheduling-step";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
const { createNodePoolDraft } = nodePoolUtil;
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import { ForbiddenError, NotFoundError } from "@/models/http-response.model";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { useAppStore } from "@/stores/app.store";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNodepoolMetrics: vi.fn().mockResolvedValue({ measurements: [] }),
    getNodepool: vi.fn(),
    createNodepool: vi.fn(),
    updateNodepool: vi.fn(),
  },
}));
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    getDepartmentNodePoolResources: vi.fn().mockResolvedValue({ nodePoolResources: [] }),
    countDepartmentNodePoolResources: vi.fn().mockResolvedValue(0),
    sumDepartmentNodePoolResources: vi.fn().mockResolvedValue({ weightsSum: 0, deservedSum: 0 }),
  },
}));
vi.mock("@/services/control-plane/org-unit-settings.service/org-unit-settings.service", () => ({
  orgUnitSettingsService: { getLimitQuotaOverSubscription: vi.fn().mockResolvedValue({ value: false }) },
}));
vi.mock("@/services/authorization/permission.service/permission.service", () => ({
  permissionService: {
    getPermittedScopesForAction: vi.fn().mockResolvedValue({ read: { system: true }, update: { system: true } }),
  },
}));
const pool = { ...createNodePoolDraft("cluster"), id: "one", name: "training", clusterName: "Tokyo" } as Nodepool;
async function render(realStep = false) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/:id", name: NODE_POOL_ROUTE_NAMES.NODE_POOL_SCHEDULING, component: View },
      {
        path: "/edit/:id",
        name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT,
        component: { template: "<div>Topology step</div>" },
      },
      { path: "/", name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX, component: { template: "<div>Pool list</div>" } },
    ],
  });
  await router.push("/one");
  await router.isReady();
  const wrapper = mount(RouterView, {
    global: {
      plugins: [router, [Quasar, { plugins: { Notify } }], createTestingPinia({ createSpy: vi.fn })],
      directives: { "click-outside": {}, permission: {} },
      stubs: realStep ? {} : { NodePoolSchedulingStep: true },
    },
  });
  return { wrapper, router };
}
beforeEach(() => vi.clearAllMocks());
describe("scheduling route", () => {
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
  it("opens topology with the successfully saved pool ID after Save & continue", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue(pool);
    const { wrapper, router } = await render();
    await flushPromises();
    wrapper.findComponent(Step).vm.$emit("saved", { ...pool, id: "saved-pool" }, "continue");
    await flushPromises();
    expect(router.currentRoute.value.name).toBe(NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT);
    expect(router.currentRoute.value.params.id).toBe("saved-pool");
    expect(router.currentRoute.value.query.stepName).toBe("topology");
    expect(wrapper.text()).toContain("Topology step");
    wrapper.unmount();
  });

  it("keeps Save & close returning to the node pool list", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue(pool);
    const { wrapper, router } = await render();
    await flushPromises();
    wrapper.findComponent(Step).vm.$emit("saved", pool, "close");
    await flushPromises();
    expect(router.currentRoute.value.name).toBe(NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX);
    expect(wrapper.text()).toContain("Pool list");
    wrapper.unmount();
  });

  it("does not navigate when partial persistence only updates the saved pool snapshot", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue(pool);
    const { wrapper, router } = await render();
    await flushPromises();
    wrapper.findComponent(Step).vm.$emit("pool-saved", { ...pool, name: "partially saved" });
    await flushPromises();
    expect(router.currentRoute.value.name).toBe(NODE_POOL_ROUTE_NAMES.NODE_POOL_SCHEDULING);
    expect(wrapper.findComponent(Step).props("pool").name).toBe("partially saved");
    wrapper.unmount();
  });

  it("loads by route ID without creating or updating a pool", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue(pool);
    const { wrapper } = await render();
    await flushPromises();
    expect(clusterService.getNodepool).toHaveBeenCalledWith("one");
    expect(wrapper.findComponent(Step).props("pool")).toEqual(pool);
    expect(clusterService.createNodepool).not.toHaveBeenCalled();
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

  it.each([new ForbiddenError(), new NotFoundError()])(
    "removes the saved editor when refreshing the same pool returns $statusCode",
    async (error) => {
      vi.mocked(clusterService.getNodepool).mockResolvedValueOnce(pool).mockRejectedValue(error);
      const { wrapper } = await render(true);
      await flushPromises();
      expect(wrapper.get('[aid="scheduling-save-close"]').isVisible()).toBe(true);

      wrapper.findComponent(Step).vm.$emit("refresh");
      await flushPromises();
      expect(wrapper.find('[aid="scheduling-save-close"]').exists()).toBe(false);
      expect(wrapper.get('[role="alert"]').text()).toContain(error.message);
      expect(useAppStore().setPageTitleOverride).toHaveBeenLastCalledWith(null);

      await wrapper.get('[aid="retry-scheduling-pool"]').trigger("click");
      await flushPromises();
      expect(clusterService.getNodepool).toHaveBeenCalledTimes(3);
      expect(clusterService.getNodepool).toHaveBeenLastCalledWith(pool.id);
      expect(wrapper.get('[role="alert"]').text()).toContain(error.message);
      expect(wrapper.find('[aid="scheduling-save-close"]').exists()).toBe(false);
      expect(clusterService.updateNodepool).not.toHaveBeenCalled();
      wrapper.unmount();
    },
  );

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
