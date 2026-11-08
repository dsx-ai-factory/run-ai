/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { QCheckbox, Quasar } from "quasar";
import type { SchedulingDraft } from "./node-pool-scheduling.model";
import type { DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client";
import MoveToRankModal from "@/components/org-unit/v2/common/peer-org-unit-rank-list/move-to-rank-modal/move-to-rank-modal.vue";
import { RunaiSelect } from "@/components/common/runai-select";
import Table from "./node-pool-department-table.vue";
const rows: DepartmentNodePoolResources[] = [
  {
    id: "resource-a",
    clusterId: "c",
    nodePoolId: "p",
    nodePoolName: "pool",
    orgUnitId: "a",
    orgUnitName: "Alpha",
    rank: "MediumLow",
    resources: { gpu: { overQuotaWeight: 2, deserved: 3 } },
  },
];
const draft: SchedulingDraft = { mode: "proportional", automaticWeights: false, cpuQuota: false, departments: {} };
describe("scheduling department table", () => {
  it("displays a whole-pool percentage and preserves pending edits when pages change", async () => {
    const wrapper = mount(Table, {
      props: {
        rows,
        draft,
        aggregates: { all: { weightsSum: 10, deservedSum: 3 } },
        ranks: [],
        readonly: false,
        loading: false,
      },
      global: { plugins: [Quasar], directives: { "click-outside": {} } },
    });
    expect(wrapper.text()).toContain("20%");
    await wrapper.setProps({ draft: { ...draft, departments: { a: { id: "a", weight: 5 } } } });
    expect(wrapper.text()).toContain("38.46%");
    await wrapper.setProps({ rows: [] });
    await wrapper.setProps({ rows });
    expect(wrapper.text()).toContain("38.46%");
  });
  it("shows unknown totals and read-only cells", () => {
    const wrapper = mount(Table, {
      props: { rows, draft, aggregates: {}, ranks: [], readonly: true, loading: false },
      global: { plugins: [Quasar], directives: { "click-outside": {} } },
    });
    expect(wrapper.text()).toContain("Unknown");
    expect(wrapper.text()).toContain("View only");
  });
  it("selects departments across ranks and moves them together", async () => {
    const wrapper = mount(Table, {
      props: {
        rows: [...rows, { ...rows[0], id: "resource-b", orgUnitId: "b", orgUnitName: "Beta", rank: "High" }],
        draft: { ...draft, mode: "ranked" },
        aggregates: {},
        ranks: ["High", "MediumLow"],
        readonly: false,
        loading: false,
      },
      global: { plugins: [Quasar], directives: { "click-outside": {} } },
    });
    expect(wrapper.text()).toContain("Alpha");
    expect(wrapper.text()).not.toContain("Beta");
    wrapper.findComponent(QCheckbox).vm.$emit("update:modelValue", true);
    await wrapper.get('[aid="rank-High"]').trigger("click");
    const beta = wrapper.findAllComponents(QCheckbox).find((item) => item.attributes("aria-label") === "Select Beta");
    expect(beta).toBeDefined();
    beta?.vm.$emit("update:modelValue", true);
    await wrapper.vm.$nextTick();
    await wrapper.get('[aid="move-departments"]').trigger("click");
    wrapper.findComponent(MoveToRankModal).vm.$emit("move", "Highest");
    expect(wrapper.emitted("edit")).toEqual([[{ id: "a", rank: "Highest" }], [{ id: "b", rank: "Highest" }]]);
    await wrapper.setProps({
      draft: {
        ...draft,
        mode: "ranked",
        departments: { a: { id: "a", rank: "Highest" }, b: { id: "b", rank: "Highest" } },
      },
    });
    expect(wrapper.text()).toContain("Highest (2 departments)");
    expect(wrapper.text()).toContain("Alpha");
    expect(wrapper.text()).toContain("Beta");
    wrapper.unmount();
  });
  it("uses quotas for automatic percentages and disables automatic weight editing", () => {
    const wrapper = mount(Table, {
      props: {
        rows,
        draft: { ...draft, mode: "guaranteed", automaticWeights: true },
        aggregates: { MediumLow: { weightsSum: 2, deservedSum: 12 } },
        ranks: ["MediumLow"],
        readonly: false,
        loading: false,
      },
      global: { plugins: [Quasar], directives: { "click-outside": {} } },
    });
    expect(wrapper.text()).toContain("25%");
    expect(wrapper.find('[aria-disabled="true"]').exists()).toBe(true);
    wrapper.unmount();
  });
  it("emits a server rank filter and hides percentages for restricted visibility", async () => {
    const wrapper = mount(Table, {
      props: {
        rows,
        draft: { ...draft, mode: "ranked" },
        aggregates: { MediumLow: { weightsSum: 10, deservedSum: 3 } },
        ranks: ["MediumLow"],
        totalsComplete: false,
        readonly: true,
        loading: false,
      },
      global: { plugins: [Quasar], directives: { "click-outside": {} } },
    });
    wrapper.findComponent(RunaiSelect).vm.$emit("update:modelValue", "High");
    expect(wrapper.emitted("filter-rank")).toEqual([["High"]]);
    expect(wrapper.text()).toContain("cannot view every department");
    expect(wrapper.text()).toContain("Unknown");
    wrapper.unmount();
  });
  it("does not present an automatic quota as a saved adjustable weight", () => {
    const wrapper = mount(Table, {
      props: {
        rows,
        draft,
        aggregates: { all: { weightsSum: 3, deservedSum: 3, rawWeightsSum: undefined, automaticWeightsSum: 3 } },
        ranks: [],
        rawWeightsAvailable: false,
        readonly: false,
        loading: false,
      },
      global: { plugins: [Quasar], directives: { "click-outside": {} } },
    });
    expect(wrapper.text()).toContain("Raw weights are unavailable");
    expect(wrapper.text()).toContain("Unknown");
    expect(wrapper.findComponent({ name: "EditableCellPopup" }).props("displayValue")).toBe("Unknown");
    wrapper.unmount();
  });
  it("does not prefetch the default rank under another active rank filter", () => {
    const wrapper = mount(Table, {
      props: {
        rows: [],
        draft: { ...draft, mode: "ranked" },
        aggregates: {},
        ranks: ["High", "MediumLow"],
        rankCounts: { High: 60, MediumLow: 10 },
        rankFilter: "High",
        searchTerm: "match",
        readonly: false,
        loading: false,
      },
      global: { plugins: [Quasar], directives: { "click-outside": {} } },
    });
    expect(wrapper.emitted("load-rank")).toBeUndefined();
    wrapper.unmount();
  });
  it("rebases hidden edits from the same complete snapshot as refreshed totals", async () => {
    const wrapper = mount(Table, {
      props: {
        rows,
        draft: { ...draft, departments: { a: { id: "a", weight: 5 } } },
        aggregates: { all: { weightsSum: 10, deservedSum: 3 } },
        ranks: [],
        readonly: false,
        loading: false,
      },
      global: { plugins: [Quasar], directives: { "click-outside": {} } },
    });
    const beta = {
      ...rows[0],
      id: "rb",
      orgUnitId: "b",
      orgUnitName: "Beta",
      resources: { gpu: { overQuotaWeight: 8, deserved: 0 } },
    };
    await wrapper.setProps({ rows: [beta], searchTerm: "Beta" });
    await wrapper.setProps({
      aggregates: { all: { weightsSum: 12, deservedSum: 3 } },
      baselineRows: [{ ...rows[0], resources: { gpu: { overQuotaWeight: 4, deserved: 3 } } }, beta],
    } as never);
    expect(wrapper.text()).toContain("61.54%");
    await wrapper.setProps({
      rows: [...rows.map((row) => ({ ...row, resources: { gpu: { overQuotaWeight: 4, deserved: 3 } } })), beta],
      searchTerm: "",
    });
    expect(wrapper.text()).toContain("61.54%");
    wrapper.unmount();
  });
});
