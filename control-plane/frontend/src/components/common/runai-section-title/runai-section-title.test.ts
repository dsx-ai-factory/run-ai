/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import SectionTitle from "./runai-section-title.vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
describe("section title help icon", () => {
  it("keeps question help by default and supports an explicit information icon", async () => {
    const wrapper = mount(SectionTitle, {
      props: { title: "Resources", tooltipText: "Details" },
      global: { plugins: [Quasar] },
    });
    expect(wrapper.findComponent(RunaiTooltip).props("icon")).toBe("ph-bold ph-question");
    await wrapper.setProps({ tooltipIcon: "ph ph-info" });
    expect(wrapper.findComponent(RunaiTooltip).props("icon")).toBe("ph ph-info");
    wrapper.unmount();
  });
});
