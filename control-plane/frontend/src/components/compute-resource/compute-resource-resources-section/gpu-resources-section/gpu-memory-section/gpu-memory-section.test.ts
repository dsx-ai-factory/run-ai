import { mount, type VueWrapper } from "@vue/test-utils";
import GpuMemorySection from "./gpu-memory-section.vue";
import { Quasar } from "quasar";
import { describe, it, expect, afterEach, vi } from "vitest";
import type { IGpuRequestModel } from "@/models/compute-resource.model";
import type { SupersetDefaultsAllOfCompute, SupersetRulesAllOfCompute } from "@/swagger-models/workloads-client";

const defaultModelValue: IGpuRequestModel = {
  gpuDevicesRequest: 1,
  gpuRequestType: "portion",
  gpuPortionRequest: 0.5,
  gpuPortionLimit: 0.5,
};
const defaultPolicyRules: SupersetRulesAllOfCompute = {};
const defaultPolicyDefaults: SupersetDefaultsAllOfCompute = {};

const createWrapper = (props = {}) =>
  mount(GpuMemorySection, {
    global: { plugins: [Quasar] },
    props: {
      modelValue: defaultModelValue,
      policyRules: defaultPolicyRules,
      policyDefaults: defaultPolicyDefaults,
      ...props,
    },
  });

describe("gpu-memory-section.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof GpuMemorySection>>;

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  describe("Html elements check", () => {
    it("renders the section and select", () => {
      wrapper = createWrapper();
      expect(wrapper.html()).toContain("Set the fraction for each device");
      expect(wrapper.findComponent({ name: "runai-select" }).exists()).toBe(true);
    });

    it("renders portion inputs when gpuRequestType is 'portion'", () => {
      wrapper = createWrapper({
        modelValue: { ...defaultModelValue, gpuRequestType: "portion" },
        supportGpuLimit: true,
      });
      expect(wrapper.findAllComponents({ name: "policy-number-input" }).length).toBe(2);
    });

    it("renders memory inputs when gpuRequestType is 'memory'", () => {
      wrapper = createWrapper({
        modelValue: { ...defaultModelValue, gpuRequestType: "memory", gpuMemoryRequest: "500M" },
        supportGpuLimit: true,
      });
      expect(wrapper.findAllComponents({ name: "policy-quantity-input" }).length).toBe(2);
    });
  });

  describe("Props", () => {
    it("disables all inputs when sectionDisabled is true", () => {
      wrapper = createWrapper({ sectionDisabled: true });
      wrapper.findAllComponents({ name: "policy-number-input" }).forEach((input) => {
        expect(input.props("disable")).toBe(true);
      });
      wrapper.findAllComponents({ name: "policy-quantity-input" }).forEach((input) => {
        expect(input.props("disable")).toBe(true);
      });
      expect(wrapper.findComponent({ name: "runai-select" }).props("disable")).toBe(true);
    });
  });

  describe("Emits", () => {
    it("emits update:model-value when request type is changed", async () => {
      wrapper = createWrapper();
      const select = wrapper.findComponent({ name: "runai-select" });
      await select.vm.$emit("update:model-value", { value: "memory" });
      expect(wrapper.emitted("update:model-value")).toBeTruthy();
    });
    it("emits update:model-value when portion request is changed", async () => {
      wrapper = createWrapper({ modelValue: { ...defaultModelValue, gpuRequestType: "portion" } });
      const requestInput = wrapper.findAllComponents({ name: "policy-number-input" })[0];
      await requestInput.vm.$emit("update:model-value", "60");
      expect(wrapper.emitted("update:model-value")).toBeTruthy();
    });
    it("emits update:model-value when memory request is changed", async () => {
      wrapper = createWrapper({
        modelValue: { ...defaultModelValue, gpuRequestType: "memory", gpuMemoryRequest: "500M" },
      });
      const requestInput = wrapper.findAllComponents({ name: "policy-quantity-input" })[0];
      await requestInput.vm.$emit("update:model-value", "600");
      expect(wrapper.emitted("update:model-value")).toBeTruthy();
    });
  });

  describe("Computed", () => {
    it("computes gpuPortionRequest and gpuPortionLimit correctly", () => {
      wrapper = createWrapper({ modelValue: { ...defaultModelValue, gpuPortionRequest: 0.25, gpuPortionLimit: 0.75 } });
      const vm = wrapper.vm as InstanceType<typeof GpuMemorySection>;
      expect(vm.gpuPortionRequest).toBe(25);
      expect(vm.gpuPortionLimit).toBe(75);
    });
    it("computes gpuMemoryRequest and gpuMemoryLimit correctly", () => {
      wrapper = createWrapper({
        modelValue: {
          ...defaultModelValue,
          gpuRequestType: "memory",
          gpuMemoryRequest: "500M",
          gpuMemoryLimit: "1000M",
        },
      });
      const vm = wrapper.vm as InstanceType<typeof GpuMemorySection>;
      expect(vm.gpuMemoryRequest).toBe(500);
      expect(vm.gpuMemoryLimit).toBe(1000);
    });
  });
});
