/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { QBtn, Quasar } from "quasar";
import MoveToRankModal from "./move-to-rank-modal.vue";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";

describe("move to rank modal", () => {
  it("preserves the existing rank order, project wording and dialog defaults", async () => {
    const wrapper = mount(MoveToRankModal, {
      props: { selectedCount: 1, currentRank: "Medium" },
      global: { plugins: [Quasar] },
    });
    await wrapper.vm.$nextTick();
    const options = wrapper.findComponent(RunaiSelect).props("options");
    expect(options.map((option) => typeof option === "object" && option?.value)).toEqual(
      ENTITY_PRIORITY_SELECT_OPTIONS.map((option) => option.value),
    );
    expect(options.find((option) => typeof option === "object" && option?.value === "Medium")).toMatchObject({
      disable: true,
    });
    expect(wrapper.findComponent(RunaiBaseDialog).props()).toMatchObject({ size: "width-sm", noPadding: false });
    expect(document.body.textContent).toContain("1 selected project");
    wrapper.unmount();
  });

  it("uses supplied choices without changing their order or enabling disabled options", async () => {
    const wrapper = mount(MoveToRankModal, {
      props: {
        selectedCount: 2,
        currentRank: "current",
        rankOptions: [
          { label: "Last", value: "last" },
          { label: "Unavailable", value: "unavailable", disable: true },
          { label: "Current", value: "current" },
        ],
      },
      global: { plugins: [Quasar] },
    });
    await wrapper.vm.$nextTick();
    const select = wrapper.findComponent(RunaiSelect);
    expect(select.props("options")).toEqual([
      { label: "Last", value: "last", disable: false },
      { label: "Unavailable", value: "unavailable", disable: true },
      { label: "Current", value: "current", disable: true },
    ]);
    const move = wrapper.findAllComponents(QBtn).find((button) => button.props("label") === "Move 2 projects");
    if (!move) throw new Error("Missing move action");
    for (const rank of ["unavailable", "current", "unknown"]) {
      select.vm.$emit("update:modelValue", rank);
      await wrapper.vm.$nextTick();
      move.vm.$emit("click");
    }
    expect(wrapper.emitted("move")).toBeUndefined();
    select.vm.$emit("update:modelValue", "last");
    await wrapper.vm.$nextTick();
    move.vm.$emit("click");
    expect(wrapper.emitted("move")).toEqual([["last"]]);
    wrapper.unmount();
  });
});
