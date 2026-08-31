/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { QToggle, Quasar } from "quasar";
import { LargeShmToggle } from ".";

describe("large-shm-toggle.vue", () => {
  function factory(props = { modelValue: true }): VueWrapper<InstanceType<typeof LargeShmToggle>> {
    return mount(LargeShmToggle, {
      props: {
        // define default props here
        ...props,
      },
      global: {
        plugins: [Quasar],
        components: {},
      },
    });
  }

  describe("Html elements check", () => {
    it("Component exists", () => {
      const wrapper = factory();
      expect(wrapper.exists()).toBe(true);
    });
    // Add more checks for HTML elements here
  });

  describe("Props", () => {
    describe("modelValue", () => {
      it("Should be a boolean", () => {
        const wrapper = factory({ modelValue: true });
        expect(wrapper.vm.$props.modelValue).toBe(true);
      });

      it("Toggle should be off when false", async () => {
        const wrapper = factory({ modelValue: false });
        const qToggleCmp = wrapper.findComponent(QToggle);
        expect(qToggleCmp.props("modelValue")).toBe(false);
      });

      it("Toggle should be on when true", async () => {
        const wrapper = factory({ modelValue: true });
        const qToggleCmp = wrapper.findComponent(QToggle);
        expect(qToggleCmp.props("modelValue")).toBe(true);
      });
    });
  });

  describe("Emits", () => {
    it("Should emit update:model-value event when toggled", async () => {
      const wrapper = factory({ modelValue: false });
      const qToggleCmp = wrapper.findComponent(QToggle);
      await qToggleCmp.trigger("click");
      expect(wrapper.emitted("update:model-value")).toBeTruthy();
    });
  });
});
