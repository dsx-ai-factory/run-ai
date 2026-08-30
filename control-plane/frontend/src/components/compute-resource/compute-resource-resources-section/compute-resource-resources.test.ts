import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import ComputeResourceResources from "./compute-resource-resources-section.vue";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { Quasar } from "quasar";
import type { DefineComponent } from "vue";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { GpuRequestType, type ExtendedResource } from "@/swagger-models/assets-service-client";
import type { IResourcesSectionModel } from "@/models/compute-resource.model";

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

describe("compute-resource-resources-section.vue", () => {
  let computeResourceCmp: VueWrapper<InstanceType<typeof ComputeResourceResources>>;

  beforeEach(() => {
    const resourcesModel: IResourcesSectionModel = {
      gpuDevicesRequest: 1,
      gpuRequestType: GpuRequestType.Portion,
      gpuPortionRequest: 0,
      cpuCore: { request: 0, limit: 0 },
      cpuMemory: { request: "0M", limit: null },
    };
    const largeShmRequest = false;
    const extendedResources: ExtendedResource[] = [];

    computeResourceCmp = mount(ComputeResourceResources, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        resourcesModel,
        largeShmRequest,
        extendedResources,
      },
    });
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(computeResourceCmp.html()).toContain("Resources");
    expect(computeResourceCmp.classes()).toContain("compute-resource-resources-section");
  });

  test("component will emit the correct data when receiving new data from the shm-toggle component", async () => {
    const shmToggleCmp = computeResourceCmp.findComponent<DefineComponent>("[aid=shm-toggle]");
    shmToggleCmp.trigger("click");
    await computeResourceCmp.vm.$nextTick();
    expect(computeResourceCmp.emitted("update-shm-request")).toBeTruthy();
    const expectedEmit: boolean[][] = [[true]];
    expect(computeResourceCmp.emitted("update-shm-request")).toEqual(expectedEmit);
  });
});
