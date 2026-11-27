/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import MoveModal from "@/components/node-pools/scheduling-move-to-rank-dialog/scheduling-move-to-rank-dialog.vue";
import type { ISelectOption } from "@/models/global.model";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog/index";
import { RunaiSelect } from "@/components/common/runai-select/index";
describe("move-to-rank destinations", () => {
  it("omits No over quota from the available destinations", async () => {
    const wrapper = mount(MoveModal, {
      props: { selectedCount: 2 },
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
    expect(unsupported).toBeUndefined();
    wrapper.unmount();
  });
  it("orders ranks highest first, disables the current rank and retains local dialog layout", async () => {
    const wrapper = mount(MoveModal, {
      props: { selectedCount: 1, currentRank: "Medium" },
      global: { plugins: [Quasar] },
    });
    await wrapper.vm.$nextTick();
    const options = wrapper.findComponent(RunaiSelect).props("options");
    expect(options.map((option) => typeof option === "object" && option?.value)).toEqual([
      "Highest",
      "VeryHigh",
      "High",
      "MediumHigh",
      "Medium",
      "MediumLow",
      "Low",
      "VeryLow",
      "Lowest",
    ]);
    expect(options.find((option) => typeof option === "object" && option?.value === "Medium")).toMatchObject({
      disable: true,
    });
    expect(wrapper.findComponent(RunaiBaseDialog).props()).toMatchObject({
      customClass: "scheduling-move-dialog",
      noPadding: true,
    });
    expect(document.body.textContent).toContain("1 selected department");
    wrapper.unmount();
  });
  it("never emits an unsupported rank, even if supplied programmatically", async () => {
    const wrapper = mount(MoveModal, {
      props: { selectedCount: 2 },
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
