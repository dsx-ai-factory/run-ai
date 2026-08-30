import { mount, type VueWrapper } from "@vue/test-utils";
import { describe, it, expect, vi } from "vitest";
import GpuResourcesSummary from "./gpu-resources-summary.vue";
import type { IGpuRequestModel } from "@/models/compute-resource.model";

// Mock RunaiTooltip
vi.mock("@/components/common/runai-tooltip", () => ({
  RunaiTooltip: {
    name: "RunaiTooltip",
    props: ["tooltipText"],
    template: '<div class="mock-tooltip">{{ tooltipText }}</div>',
  },
}));

const defaultGpuResources: IGpuRequestModel = { gpuDevicesRequest: 2 };

const createWrapper = (props = {}) => {
  return mount(GpuResourcesSummary, {
    props: {
      gpuResources: defaultGpuResources,
      ...props,
    },
  });
};

type GpuResourcesSummaryInstance = InstanceType<typeof GpuResourcesSummary>;

describe("gpu-resources-summary.vue", () => {
  let wrapper: VueWrapper<GpuResourcesSummaryInstance>;

  describe("Html elements check", () => {
    it("renders the title and tooltip", () => {
      wrapper = createWrapper();
      expect(wrapper.text()).toContain("Total requested GPU resources");
      expect(wrapper.findComponent({ name: "RunaiTooltip" }).exists()).toBe(true);
    });
  });

  describe("Computed: totalGpuResources", () => {
    it("shows total GPUs for device request (no workers/replicas)", () => {
      wrapper = createWrapper({ gpuResources: { gpuDevicesRequest: 2 } });
      expect(wrapper.text()).toContain("2 GPUs");
      expect(wrapper.text()).toContain("(2 GPUs)");
    });

    it("shows total GPUs for device request with workers", () => {
      wrapper = createWrapper({ gpuResources: { gpuDevicesRequest: 2 }, workers: 3 });
      expect(wrapper.text()).toContain("6 GPUs");
      expect(wrapper.text()).toContain("(2 GPUs) x (3 Workers)");
    });

    it("shows total GPUs for device request with replicas", () => {
      wrapper = createWrapper({ gpuResources: { gpuDevicesRequest: 2 }, replicas: { min: 2, max: 4 } });
      expect(wrapper.text()).toContain("4-8 GPUs");
      expect(wrapper.text()).toContain("(2 GPUs) x (2-4 Replicas)");
    });

    it("shows total GPUs for portion request (no workers/replicas)", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2, gpuRequestType: "portion", gpuPortionRequest: 0.5 },
      });
      expect(wrapper.text()).toContain("1.00 GPUs");
      expect(wrapper.text()).toContain("(2 GPUs) x (50%)");
    });

    it("shows total GPUs for portion request with workers", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2, gpuRequestType: "portion", gpuPortionRequest: 0.5 },
        workers: 3,
      });
      expect(wrapper.text()).toContain("3.00 GPUs");
      expect(wrapper.text()).toContain("(2 GPUs) x (50%) x (3 Workers)");
    });

    it("shows total GPUs for portion request with replicas", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2, gpuRequestType: "portion", gpuPortionRequest: 0.5 },
        replicas: { min: 2, max: 4 },
      });
      expect(wrapper.text()).toContain("2.00-4.00 GPUs");
      expect(wrapper.text()).toContain("(2 GPUs) x (50%) x (2-4 Replicas)");
    });

    it("shows total GPU memory for memory request (no workers/replicas)", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2, gpuRequestType: "memory", gpuMemoryRequest: "8G" },
      });
      expect(wrapper.text()).toContain("16 GB");
      expect(wrapper.text()).toContain("(2 GPUs) x (8 GB)");
    });

    it("shows total GPU memory for memory request with workers", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2, gpuRequestType: "memory", gpuMemoryRequest: "8G" },
        workers: 3,
      });
      expect(wrapper.text()).toContain("48 GB");
      expect(wrapper.text()).toContain("(2 GPUs) x (8 GB) x (3 Workers)");
    });

    it("shows total GPU memory for memory request with replicas", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2, gpuRequestType: "memory", gpuMemoryRequest: "8G" },
        replicas: { min: 2, max: 4 },
      });
      expect(wrapper.text()).toContain("32-64 GB");
      expect(wrapper.text()).toContain("(2 GPUs) x (8 GB) x (2-4 Replicas)");
    });

    it("shows range when workers are defined with only max", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2 },
        workers: 1,
        replicas: { max: 5 },
      });
      expect(wrapper.text()).toContain("2-10 GPUs");
      expect(wrapper.text()).toContain("(2 GPUs) x (1-5 workers)");
    });

    it("shows range when workers are defined with only min", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2 },
        workers: 5,
        replicas: { min: 2 },
      });
      expect(wrapper.text()).toContain("4-10 GPUs");
      expect(wrapper.text()).toContain("(2 GPUs) x (2-5 workers)");
    });

    it("shows range for portion request when workers are defined with only max", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2, gpuRequestType: "portion", gpuPortionRequest: 0.5 },
        workers: 1,
        replicas: { max: 5 },
      });
      expect(wrapper.text()).toContain("1.00-5.00 GPUs");
      expect(wrapper.text()).toContain("(2 GPUs) x (50%) x (1-5 Workers)");
    });

    it("shows range for portion request when workers are defined with only min", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2, gpuRequestType: "portion", gpuPortionRequest: 0.5 },
        workers: 5,
        replicas: { min: 2 },
      });
      expect(wrapper.text()).toContain("2.00-5.00 GPUs");
      expect(wrapper.text()).toContain("(2 GPUs) x (50%) x (2-5 Workers)");
    });

    it("shows range for memory request when workers are defined with only max", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2, gpuRequestType: "memory", gpuMemoryRequest: "8G" },
        workers: 1,
        replicas: { max: 5 },
      });
      expect(wrapper.text()).toContain("16-80 GB");
      expect(wrapper.text()).toContain("(2 GPUs) x (8 GB) x (1-5 Workers)");
    });

    it("shows range for memory request when workers are defined with only min", () => {
      wrapper = createWrapper({
        gpuResources: { gpuDevicesRequest: 2, gpuRequestType: "memory", gpuMemoryRequest: "8G" },
        workers: 5,
        replicas: { min: 2 },
      });
      expect(wrapper.text()).toContain("32-80 GB");
      expect(wrapper.text()).toContain("(2 GPUs) x (8 GB) x (2-5 Workers)");
    });
  });
});
