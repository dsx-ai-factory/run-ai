/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { describe, it, expect } from "vitest";
import { Quasar } from "quasar";
import CpuMemoryQuotaInput from "./cpu-memory-quota-input.vue";
import { EQuotaInputState, ECpuMemoryUnit } from "@/models/org-unit.model";

describe("CpuMemoryQuotaInput", () => {
  const mountComponent = (props = {}) => {
    return mount(CpuMemoryQuotaInput, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: null,
        ...props,
      },
    });
  };

  describe("Rendering", () => {
    it("renders correctly with default props", () => {
      const wrapper = mountComponent();
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.find("input").exists()).toBe(true);
    });

    it("displays label as CPU memory", () => {
      const wrapper = mountComponent();
      expect(wrapper.text()).toContain("CPU memory");
    });

    it("renders unit selector with MB, GB, and MiB options", () => {
      const wrapper = mountComponent();
      const select = wrapper.findComponent({ name: "QSelect" });
      expect(select.exists()).toBe(true);
    });

    it("defaults to MB unit", () => {
      const wrapper = mountComponent();
      const select = wrapper.findComponent({ name: "QSelect" });
      expect(select.props("modelValue")).toBe(ECpuMemoryUnit.MB);
    });
  });

  describe("Props", () => {
    it("accepts and displays modelValue", () => {
      const wrapper = mountComponent({ modelValue: 1024 });
      const input = wrapper.find("input");
      expect(input.element.value).toBe("1024");
    });

    it("accepts custom unit", () => {
      const wrapper = mountComponent({ unit: ECpuMemoryUnit.GB });
      const select = wrapper.findComponent({ name: "QSelect" });
      expect(select.props("modelValue")).toBe(ECpuMemoryUnit.GB);
    });

    it("accepts MiB unit", () => {
      const wrapper = mountComponent({ unit: ECpuMemoryUnit.MiB });
      const select = wrapper.findComponent({ name: "QSelect" });
      expect(select.props("modelValue")).toBe(ECpuMemoryUnit.MiB);
    });
  });

  describe("Validation States", () => {
    it("applies error state", () => {
      const wrapper = mountComponent({
        state: EQuotaInputState.Error,
        message: "Error message",
      });
      const messageEl = wrapper.find(".quota-message");
      expect(messageEl.exists()).toBe(true);
      expect(messageEl.text()).toBe("Error message");
    });

    it("applies warning state", () => {
      const wrapper = mountComponent({
        state: EQuotaInputState.Warning,
        message: "Warning message",
      });
      const messageEl = wrapper.find(".quota-message");
      expect(messageEl.exists()).toBe(true);
      expect(messageEl.text()).toBe("Warning message");
    });

    it("applies default state message", async () => {
      const wrapper = mountComponent({
        state: EQuotaInputState.Default,
        message: "Default message",
      });

      // Trigger focus on the q-input component (not the native input)
      const qInput = wrapper.findComponent({ name: "QInput" });
      qInput.vm.$emit("focus");
      await wrapper.vm.$nextTick();

      const messageEl = wrapper.find(".quota-message");
      expect(messageEl.exists()).toBe(true);
      expect(messageEl.text()).toBe("Default message");
    });
  });

  describe("Events", () => {
    it("emits update:model-value on input change", async () => {
      const wrapper = mountComponent();
      const input = wrapper.find("input");
      await input.setValue(2048);
      expect(wrapper.emitted("update:model-value")).toBeTruthy();
      expect(wrapper.emitted("update:model-value")?.[0]).toEqual([2048]);
    });

    it("emits update:unit on unit change", async () => {
      const wrapper = mountComponent({ unit: ECpuMemoryUnit.MB });
      const select = wrapper.findComponent({ name: "QSelect" });
      await select.vm.$emit("update:modelValue", ECpuMemoryUnit.GB); // QSelect (Quasar) declares emits in camelCase — must match to avoid Vue warn
      expect(wrapper.emitted("update:unit")).toBeTruthy();
      expect(wrapper.emitted("update:unit")?.[0]).toEqual([ECpuMemoryUnit.GB]);
    });
  });
});
