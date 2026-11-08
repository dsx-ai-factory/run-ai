/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createTestingPinia } from "@pinia/testing";
import { QCheckbox, QToggle, Quasar } from "quasar";
import type { SchedulingDraft } from "./node-pool-scheduling.model";
import type { DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client";
import MoveToRankModal from "@/components/org-unit/v2/common/peer-org-unit-rank-list/move-to-rank-modal/move-to-rank-modal.vue";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
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
  it("renders the exact Guaranteed Automatic profile with truthful quota-derived percentages only", async () => {
    const profile = Array.from({ length: 5 }, (_, index) => ({
      ...rows[0],
      id: `fixture-auto-${index + 1}`,
      orgUnitId: `preview-department-${index + 1}`,
      orgUnitName: `department ${index + 1}`,
      resources: { gpu: { deserved: 20, overQuotaWeight: 2, limit: -1 } },
    }));
    const automatic = { ...draft, mode: "guaranteed" as const, automaticWeights: true };
    const wrapper = mount(Table, {
      props: {
        rows: profile,
        baselineRows: profile,
        draft: automatic,
        capacity: 144,
        aggregates: { all: { weightsSum: 10, deservedSum: 100 }, MediumLow: { weightsSum: 10, deservedSum: 100 } },
        ranks: ["MediumLow"],
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    wrapper.findComponent(QToggle).vm.$emit("update:modelValue", true);
    await wrapper.vm.$nextTick();
    expect(wrapper.findAll(".allocated")).toHaveLength(5);
    expect(wrapper.text()).toContain("20%");
    await wrapper.setProps({ draft: { ...automatic, automaticWeights: false } });
    expect(wrapper.findAll(".allocated")).toHaveLength(0);
    const nearby = profile.map((row, index) =>
      index ? row : { ...row, resources: { gpu: { ...row.resources.gpu, deserved: 21 } } },
    );
    await wrapper.setProps({ rows: nearby, baselineRows: nearby, draft: automatic });
    expect(wrapper.findAll(".allocated")).toHaveLength(0);
    const live = profile.map((row, index) => ({ ...row, orgUnitId: `live-auto-${index + 1}` }));
    await wrapper.setProps({ rows: live, baselineRows: live });
    expect(wrapper.findAll(".allocated")).toHaveLength(0);
    wrapper.unmount();
  });
  it("renders only the exact authored overcommit profile and clears bars for nearby quotas, modes and live IDs", async () => {
    const profile = Array.from({ length: 10 }, (_, index) => ({
      ...rows[0],
      id: `fixture-resource-${index + 1}`,
      orgUnitId: `preview-department-${index + 1}`,
      orgUnitName: `department ${index + 1}`,
      resources: { gpu: { deserved: index === 0 ? 24 : 16, overQuotaWeight: 2, limit: -1 } },
    }));
    const guaranteed = { ...draft, mode: "guaranteed" as const };
    const wrapper = mount(Table, {
      props: {
        rows: profile,
        baselineRows: profile,
        draft: guaranteed,
        capacity: 144,
        aggregates: { all: { weightsSum: 20, deservedSum: 168 }, MediumLow: { weightsSum: 20, deservedSum: 168 } },
        ranks: ["MediumLow"],
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    wrapper.findComponent(QToggle).vm.$emit("update:modelValue", true);
    await wrapper.vm.$nextTick();
    expect(wrapper.findAll(".allocated")).toHaveLength(10);
    expect(wrapper.get(".quota-warning").text()).toContain("exceeds");
    expect(wrapper.get(".quota-warning").text()).toContain("by 24 GPUs");
    expect(wrapper.text()).toContain("Simulation preview — mock results");
    const nearby = profile.map((row, index) =>
      index ? row : { ...row, resources: { gpu: { ...row.resources.gpu, deserved: 25 } } },
    );
    await wrapper.setProps({ rows: nearby, baselineRows: nearby });
    expect(wrapper.findAll(".allocated")).toHaveLength(0);
    await wrapper.setProps({ rows: profile, baselineRows: profile, draft: { ...guaranteed, mode: "ranked" } });
    expect(wrapper.findAll(".allocated")).toHaveLength(0);
    const live = profile.map((row, index) => ({ ...row, orgUnitId: `live-department-${index + 1}` }));
    await wrapper.setProps({ rows: live, baselineRows: live, draft: guaranteed });
    expect(wrapper.findAll(".allocated")).toHaveLength(0);
    wrapper.unmount();
  });
  it("uses the reference zero-state copy and uses Show settings during search, including matching rows", async () => {
    const wrapper = mount(Table, {
      props: { rows: [], draft, aggregates: {}, ranks: [], readonly: false, loading: false },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    expect(wrapper.get(".scheduling-empty").text()).toBe("No departments yet");
    expect(wrapper.get('[aid="advanced-settings"]').text()).toBe("Advanced settings");
    await wrapper.setProps({ searchTerm: "missing" });
    expect(wrapper.get(".scheduling-empty").text()).toBe("No departments match your filters");
    expect(wrapper.get('[aid="advanced-settings"]').text()).toBe("Show settings");
    await wrapper.setProps({ rows, searchTerm: "Alpha" });
    expect(wrapper.find(".scheduling-empty").exists()).toBe(false);
    expect(wrapper.get('[aid="advanced-settings"]').text()).toBe("Show settings");
    wrapper.unmount();
  });
  it.each([
    { automaticWeights: true },
    { cpuQuota: true },
    { timeFairshare: true },
    { historicalWeight: 1 },
    { historicalWindow: { Days: 1 } },
    { minimumRuntime: { Minutes: 1 } },
  ])("clears authored results when advanced inputs no longer match: %j", async (patch) => {
    const fixtureRows = Array.from({ length: 5 }, (_, index) => ({
      ...rows[0],
      orgUnitId: `preview-department-${index + 1}`,
      orgUnitName: `department ${index + 1}`,
      resources: { gpu: { deserved: 0, overQuotaWeight: 2, limit: -1 } },
    }));
    const wrapper = mount(Table, {
      props: {
        rows: fixtureRows,
        baselineRows: fixtureRows,
        draft,
        capacity: 144,
        aggregates: { all: { weightsSum: 10, deservedSum: 0 } },
        ranks: [],
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    wrapper.findComponent(QToggle).vm.$emit("update:modelValue", true);
    await wrapper.vm.$nextTick();
    expect(wrapper.findAll(".allocated")).toHaveLength(5);
    await wrapper.setProps({ draft: { ...draft, ...patch } });
    expect(wrapper.findAll(".allocated")).toHaveLength(0);
    wrapper.unmount();
  });
  it("initializes keyboard demand controls at capacity and clears fixture results after any configuration edit", async () => {
    const fixtureRows = Array.from({ length: 5 }, (_, index) => ({
      ...rows[0],
      orgUnitId: `preview-department-${index + 1}`,
      orgUnitName: `department ${index + 1}`,
      resources: { gpu: { deserved: 0, overQuotaWeight: 2, limit: -1 } },
    }));
    const wrapper = mount(Table, {
      props: {
        rows: fixtureRows,
        baselineRows: fixtureRows,
        draft,
        capacity: 144,
        aggregates: { all: { weightsSum: 10, deservedSum: 0 } },
        ranks: [],
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    wrapper.findComponent(QToggle).vm.$emit("update:modelValue", true);
    await wrapper.vm.$nextTick();
    expect(wrapper.findAll(".allocated")).toHaveLength(5);
    expect((wrapper.get('input[type="range"]').element as HTMLInputElement).value).toBe("144");
    await wrapper.setProps({
      draft: { ...draft, departments: { "preview-department-1": { id: "preview-department-1", weight: 4 } } },
    });
    expect(wrapper.findAll(".allocated")).toHaveLength(0);
    expect(wrapper.text()).toContain("Simulation preview — mock results");
    wrapper.unmount();
  });
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
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    expect(wrapper.text()).toContain("20%");
    expect(wrapper.get(".quota-marker").attributes("style")).toContain("left: 20%");
    await wrapper.setProps({ draft: { ...draft, departments: { a: { id: "a", weight: 5 } } } });
    expect(wrapper.text()).toContain("38.46%");
    expect(wrapper.get(".quota-marker").attributes("style")).toContain("left: 38.46%");
    await wrapper.setProps({ rows: [] });
    await wrapper.setProps({ rows });
    expect(wrapper.text()).toContain("38.46%");
  });
  it("shows unknown totals and read-only cells", () => {
    const wrapper = mount(Table, {
      props: { rows, draft, aggregates: {}, ranks: [], readonly: true, loading: false },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    expect(wrapper.text()).toContain("Unknown");
    expect(wrapper.find('[aria-disabled="true"]').exists()).toBe(true);
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
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
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
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    expect(wrapper.text()).toContain("25%");
    expect(wrapper.find('[aria-disabled="true"]').exists()).toBe(true);
    wrapper.unmount();
  });
  it("filters ranks through shared controls and hides percentages for restricted visibility", async () => {
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
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    wrapper.findComponent(RunaiPageActions).vm.$emit("filters-changed", {
      advancedFilters: [{ name: "rank", type: "EnumBased", selectedValues: ["High"], selectOptions: [], label: "Rank" }],
    });
    await wrapper.vm.$nextTick();
    expect(wrapper.text()).not.toContain("Alpha");
    expect(wrapper.text()).toContain("cannot view every department");
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
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
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
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
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
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
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
