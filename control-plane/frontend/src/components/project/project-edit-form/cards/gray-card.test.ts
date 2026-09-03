/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, it } from "vitest";
import GrayCard from "./gray-card.vue";

describe("gray-card.vue", () => {
  it("renders a Phosphor icon when the card has a title", () => {
    const wrapper = mount(GrayCard, {
      global: { plugins: [Quasar] },
      props: {
        title: "This rule is inherited from the parent department",
        phIconName: "ph-bold ph-garage",
        hideCloseButton: true,
      },
    });

    const icon = wrapper.findComponent({ name: "QIcon" });
    expect(icon.props("name")).toBe("ph-bold ph-garage");
    expect(icon.props("size")).toBe("22px");
    expect(icon.props("color")).toBe("black-54");
  });

  it("does not render the inherited-rule header or icon without a title", () => {
    const wrapper = mount(GrayCard, {
      global: { plugins: [Quasar] },
      props: {
        phIconName: "ph-bold ph-garage",
        hideCloseButton: true,
      },
    });

    expect(wrapper.findComponent({ name: "QIcon" }).exists()).toBe(false);
  });
});
