import { mount } from "@vue/test-utils";
import { describe, it, expect, vi } from "vitest";
import HybridComputeResourceSection from "./hybrid-compute-resource-section.vue";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { DistributedFramework, type Toleration } from "@/swagger-models/assets-service-client";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("hybrid-compute-resource-section.vue", () => {
  const factory = (props = {}) =>
    mount(HybridComputeResourceSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        selectedName: "",
        workloadSpec: {},
        sectionDisabled: false,
        policy: {},
        nodePoolsResources: [],
        supportMultiFractionGpu: false,
        ...props,
      },
    });

  it("renders correctly", () => {
    const wrapper = factory();
    expect(wrapper.exists()).toBe(true);
  });

  it("disables node affinity section when sectionDisabled is true", async () => {
    const wrapper = factory({ sectionDisabled: true });
    const nfSection = wrapper.findComponent({ name: "node-affinity" });

    expect(nfSection.vm.disabled).toBe(true);
  });

  it("emits load-section-templates event", async () => {
    const wrapper = factory();
    await wrapper.findComponent({ name: "hybrid-section-wrapper" }).vm.$emit("load-section-templates");
    expect(wrapper.emitted("load-section-templates")).toBeTruthy();
  });

  it("emits clear event", async () => {
    const wrapper = factory();
    await wrapper.findComponent({ name: "hybrid-section-wrapper" }).vm.$emit("clear");
    expect(wrapper.emitted("clear")).toBeTruthy();
  });

  it("emits reset event", async () => {
    const wrapper = factory();
    await wrapper.findComponent({ name: "hybrid-section-wrapper" }).vm.$emit("reset");
    expect(wrapper.emitted("reset")).toBeTruthy();
  });

  it("updates tolerations", async () => {
    const wrapper = factory();
    const tolerations: Toleration[] = [{ key: "key1", operator: "Equal", value: "value1", effect: "NoSchedule" }];
    await wrapper.vm.updateTolerations(tolerations);
    expect(wrapper.emitted("update")?.[0]?.[0]).toEqual({ tolerations: tolerations });
  });

  it("updates largeShmRequest", async () => {
    const wrapper = factory();
    await wrapper.vm.updateLargeShmRequest(true);
    expect(wrapper.emitted("update")?.[0]?.[0]).toEqual({ compute: { largeShmRequest: true } });
  });

  it("shows GPU resources section and summary by default", () => {
    const wrapper = factory();
    expect(wrapper.findComponent({ name: "gpu-resources-section" }).exists()).toBe(true);
    expect(wrapper.findComponent({ name: "gpu-resources-summary" }).exists()).toBe(true);
  });

  it("hides GPU resources section and summary when hideGpuResources is true", () => {
    const wrapper = factory({ hideGpuResources: true });
    expect(wrapper.findComponent({ name: "gpu-resources-section" }).exists()).toBe(false);
    expect(wrapper.findComponent({ name: "gpu-resources-summary" }).exists()).toBe(false);
  });

  describe("segment size", () => {
    it("shows the input for PyTorch worker specifications", () => {
      const wrapper = factory({
        workloadSpec: {
          distributedFramework: DistributedFramework.PyTorch,
          numWorkers: 4,
          segmentSize: 2,
        },
      });
      const segmentSizeInput = wrapper.findComponent({ name: "distributed-segment-size-input" });
      const workersInput = wrapper.findComponent({ name: "workers-input" });
      const gpuResources = wrapper.findComponent({ name: "gpu-resources-section" });
      const workerElasticity = wrapper.findComponent({ name: "worker-elasticity-section" });

      expect(segmentSizeInput.exists()).toBe(true);
      expect(segmentSizeInput.props("segmentSize")).toBe(2);
      expect(
        gpuResources.element.compareDocumentPosition(workersInput.element) & Node.DOCUMENT_POSITION_FOLLOWING,
      ).not.toBe(0);
      expect(
        workersInput.element.compareDocumentPosition(segmentSizeInput.element) & Node.DOCUMENT_POSITION_FOLLOWING,
      ).not.toBe(0);
      expect(
        segmentSizeInput.element.compareDocumentPosition(workerElasticity.element) & Node.DOCUMENT_POSITION_FOLLOWING,
      ).not.toBe(0);
    });

    it("does not show the input for other distributed frameworks", () => {
      const wrapper = factory({
        workloadSpec: {
          distributedFramework: DistributedFramework.Tf,
          numWorkers: 4,
        },
      });

      expect(wrapper.findComponent({ name: "distributed-segment-size-input" }).exists()).toBe(false);
    });

    it("does not show the input when the specification has no workers", () => {
      const wrapper = factory({
        workloadSpec: {
          distributedFramework: DistributedFramework.PyTorch,
        },
      });

      expect(wrapper.findComponent({ name: "distributed-segment-size-input" }).exists()).toBe(false);
    });

    it("updates segmentSize without changing the rest of the specification", async () => {
      const workloadSpec = {
        distributedFramework: DistributedFramework.PyTorch,
        numWorkers: 4,
        segmentSize: null,
      };
      const wrapper = factory({ workloadSpec });

      await wrapper.findComponent({ name: "distributed-segment-size-input" }).vm.$emit("update", 2);

      expect(wrapper.emitted("update")?.[0]?.[0]).toEqual({ ...workloadSpec, segmentSize: 2 });
    });
  });
});
