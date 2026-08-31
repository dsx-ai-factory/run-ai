/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { shallowMount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, it, expect } from "vitest";
import PolicyRuleSummary from "./policy-rule-summary.vue";
import { RunaiBtn } from "@/components/common/runai-btn";
import { gvkPolicyUtil } from "@/utils/gvk-policy.util";

function mountSummary() {
  const rule = gvkPolicyUtil.createEmptyRuleDraft();
  rule.field = "image";
  rule.specSelector = "allContainers";
  return shallowMount(PolicyRuleSummary, {
    props: { rule, index: 0 },
    global: { plugins: [Quasar] },
  });
}

describe("policy-rule-summary.vue", () => {
  it("renders the edit/delete icons as Phosphor glyphs", () => {
    const html = mountSummary().html();
    expect(html).toContain("ph-bold ph-pencil-simple");
    expect(html).toContain("ph-bold ph-x");
    expect(html).not.toContain("fa-");
  });

  it("emits edit / delete when the icon buttons are clicked", async () => {
    const wrapper = mountSummary();
    const [editBtn, deleteBtn] = wrapper.findAllComponents(RunaiBtn); // template order: edit, delete
    await editBtn.vm.$emit("click");
    await deleteBtn.vm.$emit("click");
    expect(wrapper.emitted("edit")).toBeTruthy();
    expect(wrapper.emitted("delete")).toBeTruthy();
  });
});
