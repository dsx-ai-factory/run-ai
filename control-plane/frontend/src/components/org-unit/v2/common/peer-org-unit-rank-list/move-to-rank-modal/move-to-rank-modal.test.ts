/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import MoveModal from "./move-to-rank-modal.vue";
import type { ISelectOption } from "@/models/global.model";
import { RunaiSelect } from "@/components/common/runai-select";
describe("move-to-rank unsupported option", () => {
  it("leaves Destination rank empty when an unsupported option is displayed", async () => {
    const wrapper = mount(MoveModal, {
      props: { selectedCount: 2, showUnsupportedNoOverQuota: true },
      global: { plugins: [Quasar] },
    });
    await wrapper.vm.$nextTick();
    const select = wrapper.findComponent(RunaiSelect);
    const unsupported = select
      .props("options")
      .find(
        (option): option is ISelectOption =>
          typeof option === "object" && option !== null && option.label === "No over quota",
      );
    if (!unsupported) throw new Error("Unsupported rank explanation is missing");
    expect(unsupported.disable).toBe(true);
    expect(unsupported.value).not.toBe(select.props("modelValue"));
    expect(document.querySelector(".q-field__native")?.textContent).not.toContain("No over quota");
    wrapper.unmount();
  });
  it("never emits an unsupported rank, even if supplied programmatically", async () => {
    const wrapper = mount(MoveModal, {
      props: { selectedCount: 2, showUnsupportedNoOverQuota: true },
      global: { plugins: [Quasar] },
    });
    await wrapper.vm.$nextTick();
    const select = wrapper.findComponent(RunaiSelect);
    select.vm.$emit("update:modelValue", "__unsupported_no_over_quota__");
    await wrapper.vm.$nextTick();
    const move = Array.from(document.querySelectorAll("button"))
      .filter((button) => button.textContent?.includes("Move 2"))
      .at(-1);
    if (!move) throw new Error("Move action is missing");
    move.click();
    expect(wrapper.emitted("move")).toBeUndefined();
    select.vm.$emit("update:modelValue", "Highest");
    await wrapper.vm.$nextTick();
    move.click();
    expect(wrapper.emitted("move")).toEqual([["Highest"]]);
    wrapper.unmount();
  });
});
