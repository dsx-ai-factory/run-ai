/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, it, expect } from "vitest";
import NotificationChannelsSelector from "./notification-channels-selector.vue";
import type { NotificationChannelOption } from "./notification-channels-selector.types";

interface NotificationChannelsSelectorProps {
  options: NotificationChannelOption[];
}

describe("notification-channels-selector.vue", () => {
  const mockOptions: NotificationChannelOption[] = [
    {
      id: "email",
      label: "Email",
      isSelected: true,
      isDisabled: false,
      disabledTooltipText: "",
    },
    {
      id: "slack",
      label: "Slack",
      isSelected: false,
      isDisabled: true,
      disabledTooltipText: "Slack integration not configured",
    },
  ];

  function createWrapper(
    props: NotificationChannelsSelectorProps,
  ): VueWrapper<InstanceType<typeof NotificationChannelsSelector>> {
    return mount(NotificationChannelsSelector, {
      global: {
        plugins: [Quasar],
      },
      props: props,
    });
  }

  describe("Rendering", () => {
    it("renders the component with title", () => {
      const wrapper = createWrapper({
        options: mockOptions,
      });

      expect(wrapper.text()).toContain("Send me system notifications via:");
    });

    it("renders all notification channel options", () => {
      const wrapper = createWrapper({
        options: mockOptions,
      });

      const checkboxes = wrapper.findAllComponents({ name: "q-checkbox" });
      expect(checkboxes).toHaveLength(2);
    });

    it("displays correct labels for each option", () => {
      const wrapper = createWrapper({
        options: mockOptions,
      });

      const checkboxLabels = wrapper.findAllComponents({ name: "q-checkbox" });
      expect(checkboxLabels[0].props("label")).toBe("Email");
      expect(checkboxLabels[1].props("label")).toBe("Slack");
    });
  });

  describe("Props", () => {
    it("passes correct model-value to checkboxes based on isSelected", () => {
      const wrapper = createWrapper({
        options: mockOptions,
      });

      const checkboxes = wrapper.findAllComponents({ name: "q-checkbox" });
      expect(checkboxes[0].props("modelValue")).toBe(true); // Email is selected
      expect(checkboxes[1].props("modelValue")).toBe(false); // Slack is not selected
    });

    it("disables checkboxes when isDisabled is true", () => {
      const wrapper = createWrapper({
        options: mockOptions,
      });

      const checkboxes = wrapper.findAllComponents({ name: "q-checkbox" });
      expect(checkboxes[0].props("disable")).toBe(false); // Email is enabled
      expect(checkboxes[1].props("disable")).toBe(true); // Slack is disabled
    });

    it("renders tooltip for disabled options with disabledTooltipText", () => {
      const wrapper = createWrapper({
        options: mockOptions,
      });

      const tooltips = wrapper.findAllComponents({ name: "q-tooltip" });
      expect(tooltips).toHaveLength(1); // Only Slack has tooltip
      // Check that the tooltip component exists for the disabled option
      expect(tooltips[0].exists()).toBe(true);
    });

    it("does not render tooltip for enabled options", () => {
      const wrapper = createWrapper({
        options: mockOptions,
      });

      const tooltips = wrapper.findAllComponents({ name: "q-tooltip" });
      expect(tooltips).toHaveLength(1); // Only disabled option has tooltip
    });
  });

  describe("Events", () => {
    it('emits "channel-toggled" event with option id when checkbox is clicked', async () => {
      const wrapper = createWrapper({
        options: mockOptions,
      });

      const emailCheckbox = wrapper.findAllComponents({ name: "q-checkbox" })[0];
      await emailCheckbox.vm.$emit("update:modelValue");

      expect(wrapper.emitted()).toHaveProperty("channel-toggled");
      expect(wrapper.emitted("channel-toggled")?.[0]).toEqual(["email"]);
    });

    it('emits "channel-toggled" event with correct id for different options', async () => {
      const wrapper = createWrapper({
        options: mockOptions,
      });

      const slackCheckbox = wrapper.findAllComponents({ name: "q-checkbox" })[1];
      await slackCheckbox.vm.$emit("update:modelValue");

      expect(wrapper.emitted("channel-toggled")?.[0]).toEqual(["slack"]);
    });

    it('emits "channel-toggled" event even for disabled options', async () => {
      const wrapper = createWrapper({
        options: mockOptions,
      });

      const slackCheckbox = wrapper.findAllComponents({ name: "q-checkbox" })[1];
      await slackCheckbox.vm.$emit("update:modelValue");

      expect(wrapper.emitted("channel-toggled")?.[0]).toEqual(["slack"]);
    });
  });

  describe("Edge Cases", () => {
    it("handles empty options array", () => {
      const wrapper = createWrapper({
        options: [],
      });

      const checkboxes = wrapper.findAllComponents({ name: "q-checkbox" });
      expect(checkboxes).toHaveLength(0);
    });

    it("handles options without disabledTooltipText", () => {
      const optionsWithoutTooltip: NotificationChannelOption[] = [
        {
          label: "Email",
          id: "email",
          isSelected: false,
          isDisabled: true,
          disabledTooltipText: "",
        },
      ];

      const wrapper = createWrapper({
        options: optionsWithoutTooltip,
      });

      const tooltips = wrapper.findAllComponents({ name: "q-tooltip" });
      expect(tooltips).toHaveLength(0); // No tooltip when disabledTooltipText is empty
    });

    it("handles all options selected", () => {
      const allSelectedOptions: NotificationChannelOption[] = [
        {
          id: "email",
          label: "Email",
          isSelected: true,
          isDisabled: false,
          disabledTooltipText: "",
        },
        {
          id: "slack",
          label: "Slack",
          isSelected: true,
          isDisabled: false,
          disabledTooltipText: "",
        },
      ];

      const wrapper = createWrapper({
        options: allSelectedOptions,
      });

      const checkboxes = wrapper.findAllComponents({ name: "q-checkbox" });
      expect(checkboxes[0].props("modelValue")).toBe(true);
      expect(checkboxes[1].props("modelValue")).toBe(true);
    });
  });
});
