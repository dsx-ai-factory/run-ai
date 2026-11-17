/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { QCheckbox, QInput, QSelect, Quasar } from "quasar";
import SchedulingFilters from "./scheduling-filters.vue";
import { RunaiPageFilterWrapper } from "@/components/common/runai-advanced-filters/runai-page-filter-wrapper";
import { RunaiAdvancedFilterSummary } from "@/components/common/runai-advanced-filters/runai-advanced-filter-summary";
import { EColumnFilterType, EFilterOperator, type IFilterBy, type IEnumBasedFilterModel } from "@/models/filter.model";

function mountFilters(filterBy: IFilterBy = { advancedFilters: [] }, ranked = true) {
  return mount(SchedulingFilters, {
    props: { filterBy, ranked },
    global: {
      plugins: [Quasar],
      stubs: { QMenu: { template: "<div><slot /></div>", methods: { updatePosition() {} } } },
    },
  });
}
describe("scheduling filters", () => {
  it("applies the Department operator and term while retaining search and unrelated filters", async () => {
    const rank: IEnumBasedFilterModel = {
      name: "rank",
      label: "Rank",
      type: EColumnFilterType.EnumBased,
      selectedValues: ["High"],
      selectOptions: [{ value: "High", label: "High" }],
    };
    const original: IFilterBy = { searchTerm: "alpha", advancedFilters: [rank] };
    const wrapper = mountFilters(original);
    await wrapper.get(".q-item--clickable").trigger("click");
    const select = wrapper.getComponent(QSelect);
    expect((select.props("options") ?? []).map((option: { label: string }) => option.label)).toEqual([
      "Contains",
      "Does not contain",
      "Equals",
      "Does not equal",
    ]);
    select.vm.$emit("update:modelValue", EFilterOperator.NotEquals);
    wrapper.getComponent(QInput).vm.$emit("update:modelValue", "Beta");
    await wrapper.vm.$nextTick();
    wrapper.getComponent(RunaiPageFilterWrapper).vm.$emit("save");
    expect(wrapper.emitted("update-filter-by")?.[0][0]).toEqual({
      searchTerm: "alpha",
      advancedFilters: [
        rank,
        {
          name: "department",
          label: "Department",
          type: EColumnFilterType.FreeText,
          term: "Beta",
          value: EFilterOperator.NotEquals,
        },
      ],
    });
    expect(original.advancedFilters).toEqual([rank]);
    wrapper.unmount();
  });
  it("discards Department edits when its editor closes", async () => {
    const department = {
      name: "department",
      label: "Department",
      type: EColumnFilterType.FreeText,
      term: "Alpha",
      value: EFilterOperator.Contains,
    };
    const wrapper = mountFilters({ advancedFilters: [department] });
    wrapper.getComponent(RunaiAdvancedFilterSummary).vm.$emit("select", department);
    await wrapper.vm.$nextTick();
    wrapper.getComponent(QInput).vm.$emit("update:modelValue", "unsaved");
    wrapper.getComponent(RunaiPageFilterWrapper).vm.$emit("close");
    await wrapper.vm.$nextTick();
    expect(wrapper.findComponent(RunaiPageFilterWrapper).exists()).toBe(false);
    expect(wrapper.emitted("update-filter-by")).toBeUndefined();
    expect(department.term).toBe("Alpha");
    wrapper.getComponent(RunaiAdvancedFilterSummary).vm.$emit("select", department);
    await wrapper.vm.$nextTick();
    expect(wrapper.getComponent(QInput).props("modelValue")).toBe("Alpha");
    wrapper.unmount();
  });
  it("isolates Rank checkbox changes until Apply and removes only the chosen chip", async () => {
    const rank: IEnumBasedFilterModel = {
      name: "rank",
      label: "Rank",
      type: EColumnFilterType.EnumBased,
      selectedValues: ["High"],
      selectOptions: [
        { value: "High", label: "High" },
        { value: "Low", label: "Low" },
      ],
    };
    const wrapper = mountFilters({ searchTerm: "alpha", advancedFilters: [rank] });
    const summary = wrapper.getComponent(RunaiAdvancedFilterSummary);
    summary.vm.$emit("select", rank);
    await wrapper.vm.$nextTick();
    wrapper.findAllComponents(QCheckbox)[1].vm.$emit("update:modelValue", ["High", "Low"]);
    await wrapper.vm.$nextTick();
    expect(rank.selectedValues).toEqual(["High"]);
    wrapper.getComponent(RunaiPageFilterWrapper).vm.$emit("close");
    await wrapper.vm.$nextTick();
    expect(wrapper.emitted("update-filter-by")).toBeUndefined();
    summary.vm.$emit("select", rank);
    await wrapper.vm.$nextTick();
    expect(wrapper.findAllComponents(QCheckbox)[0].props("modelValue")).toEqual(["High"]);
    wrapper.findAllComponents(QCheckbox)[1].vm.$emit("update:modelValue", ["Low"]);
    await wrapper.vm.$nextTick();
    wrapper.getComponent(RunaiPageFilterWrapper).vm.$emit("save");
    expect(wrapper.emitted("update-filter-by")?.[0][0]).toMatchObject({
      searchTerm: "alpha",
      advancedFilters: [{ name: "rank", selectedValues: ["Low"] }],
    });
    expect(rank.selectedValues).toEqual(["High"]);
    summary.vm.$emit("remove", "rank");
    expect(wrapper.emitted("update-filter-by")?.[1][0]).toEqual({ searchTerm: "alpha", advancedFilters: [] });
    wrapper.unmount();
  });
  it("offers only Department when ranks are unavailable and leaves empty terms unapplied", async () => {
    const wrapper = mountFilters(undefined, false);
    expect(wrapper.findAll(".q-item--clickable").map((item) => item.text())).toEqual(["Department"]);
    await wrapper.get(".q-item--clickable").trigger("click");
    wrapper.getComponent(RunaiPageFilterWrapper).vm.$emit("save");
    expect(wrapper.emitted("update-filter-by")).toBeUndefined();
    wrapper.unmount();
  });
});
