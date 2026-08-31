/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, it } from "vitest";

import DistributedSegmentSizeInput from "./distributed-segment-size-input.vue";

describe("distributed-segment-size-input.vue", () => {
  const factory = (props = {}) =>
    mount(DistributedSegmentSizeInput, {
      global: {
        plugins: [Quasar],
      },
      props: {
        segmentSize: null,
        sectionDisabled: false,
        ...props,
      },
    });

  it("renders the Figma title, tooltip, and positive integer input", () => {
    const wrapper = factory({ segmentSize: 2 });
    const section = wrapper.findComponent({ name: "runai-section" });
    const input = wrapper.findComponent({ name: "policy-number-input" });

    expect(section.props("title")).toBe("Set the number of workers per segment");
    expect(section.props("tooltipText")).toContain("Leave it blank to disable segmentation");
    expect(input.props("modelValue")).toBe(2);
    expect(input.props("minValue")).toBe(1);
    expect(input.props("isInteger")).toBe(true);
    expect(input.props("aid")).toBe("segment-size-input");
  });

  it("renders with an unset segment size", () => {
    const wrapper = factory({ segmentSize: undefined });

    expect(wrapper.findComponent({ name: "policy-number-input" }).props("modelValue")).toBeNull();
  });

  it("emits a numeric segment size", async () => {
    const wrapper = factory();

    await wrapper.findComponent({ name: "policy-number-input" }).vm.$emit("update:model-value", "4");

    expect(wrapper.emitted("update")?.[0]).toEqual([4]);
  });

  it("emits null when the input is cleared", async () => {
    const wrapper = factory({ segmentSize: 2 });

    await wrapper.findComponent({ name: "policy-number-input" }).vm.$emit("update:model-value", "");

    expect(wrapper.emitted("update")?.[0]).toEqual([null]);
  });

  it("disables the input with the section", () => {
    const wrapper = factory({ sectionDisabled: true });

    expect(wrapper.findComponent({ name: "policy-number-input" }).props("disable")).toBe(true);
  });
});
