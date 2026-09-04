/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, shallowMount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import { ref } from "vue";

import WorkloadStructure from "./workload-structure.vue";

// services
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// utils
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";

import type { Workload, WorkloadElement, WorkloadElementResponse, Pod } from "@/swagger-models/workloads-client";

vi.mock("@/services/cluster/workload.service/workload.service");

vi.mock("@/utils/interval.util", () => ({
  intervalUtil: {
    startInterval: vi.fn(),
    stopInterval: vi.fn(),
  },
}));

const routeQuery = ref<Record<string, string | undefined>>({});
const routeParams = ref<Record<string, string>>({ workloadId: "wid-1" });
const routerReplace = vi.fn();
const routerPush = vi.fn();

vi.mock("vue-router", async (importOriginal) => {
  const actual = (await importOriginal()) as object;
  return {
    ...actual,
    useRoute: () => ({
      get query() {
        return routeQuery.value;
      },
      get params() {
        return routeParams.value;
      },
      name: "workload-structure",
    }),
    useRouter: () => ({
      replace: routerReplace,
      push: routerPush,
    }),
  };
});

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

const buildWorkload = (overrides: Partial<Workload> = {}): Workload =>
  ({
    id: "wid-1",
    name: "test-workload",
    type: "Workspace",
    phase: "Running",
    clusterId: "cluster-1",
    projectName: "test-project",
    namespace: "ns-1",
    runningPods: 2,
    requestedPods: { number: 3 },
    phaseUpdatedAt: "2026-04-30T00:00:00Z",
    k8sPhaseUpdatedAt: "2026-04-30T00:00:00Z",
    updatedAt: "2026-04-30T00:00:00Z",
    createdAt: "2026-04-30T00:00:00Z",
    source: "CLI",
    deletedAt: null,
    tenantId: 1,
    ...overrides,
  }) as Workload;

const buildElement = (id: string): WorkloadElement =>
  ({
    id,
    name: `name-${id}`,
    hierarchyLevel: 1,
    createdAt: "2026-04-30T00:00:00Z",
  }) as WorkloadElement;

const buildElementResponse = (id: string): WorkloadElementResponse =>
  ({
    id,
    name: `name-${id}`,
    hierarchyLevel: 1,
    createdAt: "2026-04-30T00:00:00Z",
    allocatedResources: { gpu: 1, gpuMemory: null, cpu: null, cpuMemory: null },
  }) as unknown as WorkloadElementResponse;

const mountPage = (): VueWrapper<InstanceType<typeof WorkloadStructure>> => {
  // Provide a teleport target for completeness — shallowMount stubs Teleport,
  // but having the target avoids any stray DOM warnings.
  const teleportTarget = document.createElement("div");
  teleportTarget.id = "header-left-teleport";
  document.body.appendChild(teleportTarget);

  return shallowMount(WorkloadStructure, {
    attachTo: document.body,
    global: {
      plugins: [Quasar, pinia],
    },
  });
};

describe("workload-structure.vue", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    routeQuery.value = {};
    routeParams.value = { workloadId: "wid-1" };
    vi.mocked(workloadService.getWorkloadById).mockResolvedValue(buildWorkload());
    vi.mocked(workloadService.getWorkloadElements).mockResolvedValue([] as WorkloadElement[]);
    vi.mocked(workloadService.getWorkloadPods).mockResolvedValue([] as Pod[]);
    vi.mocked(workloadService.getWorkloadElement).mockResolvedValue(buildElementResponse("el-1"));
  });

  afterEach(() => {
    document.body.innerHTML = "";
  });

  describe("initial load", () => {
    it("calls getWorkloadById, getWorkloadElements and getWorkloadPods on mount", async () => {
      mountPage();
      await flushPromises();

      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("wid-1");
      expect(workloadService.getWorkloadElements).toHaveBeenCalledWith("wid-1");
      expect(workloadService.getWorkloadPods).toHaveBeenCalled();
    });

    it("fetches per-element details for every returned element (no batch endpoint)", async () => {
      vi.mocked(workloadService.getWorkloadElements).mockResolvedValue([
        buildElement("el-1"),
        buildElement("el-2"),
        buildElement("el-3"),
      ]);
      mountPage();
      await flushPromises();

      expect(workloadService.getWorkloadElement).toHaveBeenCalledTimes(3);
      expect(workloadService.getWorkloadElement).toHaveBeenCalledWith("wid-1", "el-1");
      expect(workloadService.getWorkloadElement).toHaveBeenCalledWith("wid-1", "el-2");
      expect(workloadService.getWorkloadElement).toHaveBeenCalledWith("wid-1", "el-3");
    });

    it("does not render the diagram when there is no workload", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      vi.mocked(workloadService.getWorkloadById).mockRejectedValue(new Error("boom"));
      const wrapper = mountPage();
      await flushPromises();

      expect(wrapper.findComponent({ name: "WorkloadArchitectureDiagram" }).exists()).toBe(false);
      errorSpy.mockRestore();
    });

    it("forwards loaded elements + per-element details to the diagram", async () => {
      vi.mocked(workloadService.getWorkloadElements).mockResolvedValue([buildElement("el-1"), buildElement("el-2")]);
      vi.mocked(workloadService.getWorkloadElement).mockImplementation(async (_, elementId) =>
        buildElementResponse(elementId),
      );
      const wrapper = mountPage();
      await flushPromises();

      const diagram = wrapper.findComponent({ name: "WorkloadArchitectureDiagram" });
      expect(diagram.exists()).toBe(true);
      const elementsProp = diagram.props("elements") as WorkloadElement[];
      expect(elementsProp.map((e) => e.id)).toEqual(["el-1", "el-2"]);

      const detailsProp = diagram.props("elementDetailsById") as Record<string, WorkloadElementResponse>;
      expect(Object.keys(detailsProp).sort()).toEqual(["el-1", "el-2"]);
    });
  });

  describe("auto-refresh interval", () => {
    it("starts the WorkloadStructure interval on mount", async () => {
      mountPage();
      await flushPromises();

      expect(intervalUtil.startInterval).toHaveBeenCalledWith(EIntervalLabels.WorkloadStructure, expect.any(Function));
    });

    it("stops the WorkloadStructure interval on unmount", async () => {
      const wrapper = mountPage();
      await flushPromises();
      vi.mocked(intervalUtil.stopInterval).mockClear();

      wrapper.unmount();

      expect(intervalUtil.stopInterval).toHaveBeenCalledWith(EIntervalLabels.WorkloadStructure);
    });

    it("the registered tick callback re-invokes the workload service", async () => {
      mountPage();
      await flushPromises();

      // Pull the registered callback out of the most recent startInterval call.
      const startCalls = vi.mocked(intervalUtil.startInterval).mock.calls;
      const tick = startCalls[startCalls.length - 1][1] as () => void;

      vi.mocked(workloadService.getWorkloadById).mockClear();
      tick();
      await flushPromises();

      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("wid-1");
    });
  });

  describe("element selection (URL → diagram → context pods)", () => {
    it("passes selectedElementId from the URL down to the diagram", async () => {
      routeQuery.value = { elementId: "el-42" };
      const wrapper = mountPage();
      await flushPromises();

      const diagram = wrapper.findComponent({ name: "WorkloadArchitectureDiagram" });
      expect(diagram.exists()).toBe(true);
      expect(diagram.props("selectedElementId")).toBe("el-42");
    });

    it("filters context pods by elementId when one is selected", async () => {
      routeQuery.value = { elementId: "el-42" };
      mountPage();
      await flushPromises();

      // Two getWorkloadPods calls happen on mount: one for diagramPods (no filter),
      // one for contextPods (with elementId filter). Find the one with a filter.
      const podCalls = vi.mocked(workloadService.getWorkloadPods).mock.calls;
      const contextCall = podCalls.find((args) => Array.isArray(args[6]) && (args[6] as string[]).length > 0);
      expect(contextCall, "getWorkloadPods called with an elementId filter").toBeTruthy();
      expect(contextCall?.[6] as string[]).toEqual([expect.stringContaining("elementId")]);
    });

    it("does NOT filter context pods when no element is selected", async () => {
      routeQuery.value = {};
      mountPage();
      await flushPromises();

      const callsWithFilter = vi
        .mocked(workloadService.getWorkloadPods)
        .mock.calls.filter((args) => Array.isArray(args[6]) && (args[6] as string[]).length > 0);
      expect(callsWithFilter).toHaveLength(0);
    });

    it("forwards node-selected from the diagram to router.replace with an elementId query", async () => {
      const wrapper = mountPage();
      await flushPromises();

      const diagram = wrapper.findComponent({ name: "WorkloadArchitectureDiagram" });
      diagram.vm.$emit("node-selected", "el-99");
      await flushPromises();

      expect(routerReplace).toHaveBeenCalled();
      const lastCall = routerReplace.mock.calls[routerReplace.mock.calls.length - 1][0];
      expect(lastCall.query.elementId).toBe("el-99");
    });

    it("forwards node-selected with null (root) by clearing the elementId query", async () => {
      routeQuery.value = { elementId: "el-99" };
      const wrapper = mountPage();
      await flushPromises();

      const diagram = wrapper.findComponent({ name: "WorkloadArchitectureDiagram" });
      diagram.vm.$emit("node-selected", null);
      await flushPromises();

      expect(routerReplace).toHaveBeenCalled();
      const lastCall = routerReplace.mock.calls[routerReplace.mock.calls.length - 1][0];
      expect(lastCall.query.elementId).toBeUndefined();
    });
  });
});
