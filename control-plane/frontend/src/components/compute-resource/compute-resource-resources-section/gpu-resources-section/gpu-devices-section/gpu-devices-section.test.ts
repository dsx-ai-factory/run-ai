/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, type VueWrapper } from "@vue/test-utils";
import GpuDevicesSection from "./gpu-devices-section.vue";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { IGpuRequestModel } from "@/models/compute-resource.model";
import type { SupersetDefaultsAllOfCompute, SupersetRulesAllOfCompute } from "@/swagger-models/workloads-client";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

const defaultModelValue: IGpuRequestModel = {
  gpuDevicesRequest: 1,
  gpuRequestType: undefined,
};
const defaultPolicyRules: SupersetRulesAllOfCompute = {
  gpuDevicesRequest: { min: 0 },
} as SupersetRulesAllOfCompute;
const defaultPolicyDefaults: SupersetDefaultsAllOfCompute = {
  gpuPortionRequest: 0.5,
  gpuPortionLimit: 0.5,
} as SupersetDefaultsAllOfCompute;

const createWrapper = (props = {}) => {
  return mount(GpuDevicesSection, {
    global: { plugins: [Quasar, pinia] },
    props: {
      modelValue: defaultModelValue,
      policyRules: defaultPolicyRules,
      policyDefaults: defaultPolicyDefaults,
      supportGpuLimit: true,
      ...props,
    },
  });
};

describe("gpu-devices-section.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof GpuDevicesSection>>;

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  describe("Html elements check", () => {
    it("renders GPU devices input and toggle", () => {
      wrapper = createWrapper();
      expect(wrapper.html()).toContain("GPU devices");
      expect(wrapper.findComponent({ name: "policy-number-input" }).exists()).toBe(true);
      expect(wrapper.findComponent({ name: "runai-toggle" }).exists()).toBe(true);
    });
  });

  describe("Props", () => {
    it("disables input and toggle when sectionDisabled is true", () => {
      wrapper = createWrapper({ sectionDisabled: true });
      expect(wrapper.findComponent({ name: "policy-number-input" }).props("disable")).toBe(true);
      expect(wrapper.findComponent({ name: "runai-toggle" }).props("disable")).toBe(true);
    });

    it("sets is-integer prop to true for GPU devices input", () => {
      wrapper = createWrapper();
      expect(wrapper.findComponent({ name: "policy-number-input" }).props("isInteger")).toBe(true);
    });

    it("enables toggle for >1 device only if supportMultiFractionGpu is true", async () => {
      wrapper = createWrapper({ modelValue: { gpuDevicesRequest: 2 }, supportMultiFractionGpu: false });
      expect(wrapper.findComponent({ name: "runai-toggle" }).props("disable")).toBe(true);
      await wrapper.setProps({ supportMultiFractionGpu: true });
      expect(wrapper.findComponent({ name: "runai-toggle" }).props("disable")).toBe(false);
    });

    it("uses policyDefaults when enabling fractioning", async () => {
      const customDefaults = { gpuPortionRequest: 0.7, gpuPortionLimit: 0.7 };
      wrapper = createWrapper({ policyDefaults: customDefaults });
      await wrapper.findComponent({ name: "runai-toggle" }).vm.$emit("update:modelValue", true);
      const emits = wrapper.emitted("update:model-value");
      expect((emits as IGpuRequestModel[][])[0][0].gpuPortionRequest).toBe(0.7);
      expect((emits as IGpuRequestModel[][])[0][0].gpuPortionLimit).toBe(0.7);
    });

    it("disables toggle when gpu fractioning is not supported for the cluster", () => {
      wrapper = createWrapper({ supportGpuFractioning: false });
      expect(wrapper.findComponent({ name: "runai-toggle" }).props("disable")).toBe(true);
      expect(wrapper.vm.fractioningDisabledTooltip).toBe("GPU fractioning is not supported for the selected cluster");
    });
  });

  describe("Emits", () => {
    it("emits update:model-value when GPU devices input changes", async () => {
      wrapper = createWrapper();
      await wrapper.findComponent({ name: "policy-number-input" }).vm.$emit("update:model-value", 2);
      const emits = wrapper.emitted("update:model-value");
      expect(Array.isArray(emits)).toBe(true);
      expect((emits as IGpuRequestModel[][])[0][0].gpuDevicesRequest).toBe(2);
    });

    it("emits reset when GPU devices set to 0", async () => {
      wrapper = createWrapper();
      await wrapper.findComponent({ name: "policy-number-input" }).vm.$emit("update:model-value", 0);
      const emits = wrapper.emitted("update:model-value");
      expect(Array.isArray(emits)).toBe(true);
      expect((emits as IGpuRequestModel[][])[0][0]).toMatchObject({
        gpuDevicesRequest: 0,
        gpuMemoryRequest: undefined,
        gpuMemoryLimit: undefined,
        gpuPortionRequest: undefined,
        gpuPortionLimit: undefined,
        gpuRequestType: undefined,
      });
    });

    it("emits update:model-value with portion fields when fractioning enabled", async () => {
      wrapper = createWrapper();
      await wrapper.findComponent({ name: "runai-toggle" }).vm.$emit("update:modelValue", true);
      const emits = wrapper.emitted("update:model-value");
      expect(Array.isArray(emits)).toBe(true);
      expect((emits as IGpuRequestModel[][])[0][0].gpuRequestType).toBeDefined();
      expect((emits as IGpuRequestModel[][])[0][0].gpuPortionRequest).toBe(defaultPolicyDefaults.gpuPortionRequest);
      expect((emits as IGpuRequestModel[][])[0][0].gpuPortionLimit).toBe(defaultPolicyDefaults.gpuPortionLimit);
    });

    it("emits update:model-value with portion fields undefined when fractioning disabled", async () => {
      wrapper = createWrapper();
      await wrapper.findComponent({ name: "runai-toggle" }).vm.$emit("update:modelValue", true);
      await wrapper.findComponent({ name: "runai-toggle" }).vm.$emit("update:modelValue", false);
      const emits = wrapper.emitted("update:model-value");
      expect(Array.isArray(emits)).toBe(true);
      const lastEmit = (emits as IGpuRequestModel[][])[(emits as IGpuRequestModel[][]).length - 1][0];
      expect(lastEmit.gpuRequestType).toBeUndefined();
      expect(lastEmit.gpuPortionRequest).toBeUndefined();
      expect(lastEmit.gpuPortionLimit).toBeUndefined();
    });
  });

  describe("Computed", () => {
    it("fractioning toggle is enabled/disabled based on devices and prop", async () => {
      wrapper = createWrapper({ modelValue: { gpuDevicesRequest: 1 } });
      expect(wrapper.findComponent({ name: "runai-toggle" }).props("disable")).toBe(false);
      await wrapper.setProps({ modelValue: { gpuDevicesRequest: 0 } });
      expect(wrapper.findComponent({ name: "runai-toggle" }).props("disable")).toBe(true);
      await wrapper.setProps({ modelValue: { gpuDevicesRequest: 2 }, supportMultiFractionGpu: false });
      expect(wrapper.findComponent({ name: "runai-toggle" }).props("disable")).toBe(true);
      await wrapper.setProps({ supportMultiFractionGpu: true });
      expect(wrapper.findComponent({ name: "runai-toggle" }).props("disable")).toBe(false);
    });

    it("shows multi-device cluster version tooltip when fractioning is disabled", () => {
      wrapper = createWrapper({ modelValue: { gpuDevicesRequest: 2 }, supportMultiFractionGpu: false });
      expect(wrapper.vm.fractioningDisabledTooltip).toBe(
        "Cluster version does not support GPU fractioning for multiple devices, to enable this set 1 GPU device",
      );
    });

    it("shows minimum device tooltip when no gpu devices are set", () => {
      wrapper = createWrapper({ modelValue: { gpuDevicesRequest: 0 } });
      expect(wrapper.vm.fractioningDisabledTooltip).toBe("To use GPU fractioning, set at least 1 GPU device");
    });
  });
});
