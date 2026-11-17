/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { flushPromises, mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createTestingPinia } from "@pinia/testing";
import { QCheckbox, QToggle, QResizeObserver, Quasar } from "quasar";
import type { SchedulingDraft } from "./node-pool-scheduling.model";
import type { DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client";
import MoveToRankModal from "@/components/org-unit/v2/common/peer-org-unit-rank-list/move-to-rank-modal/move-to-rank-modal.vue";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { simulationPreviewKey } from "./node-pool-simulation.service";
import Table from "./node-pool-department-table.vue";
import DepartmentResourceEditor from "./department-resource-editor.vue";

describe("rank group identity after filtering", () => {
  it("does not retain duplicate group rows when an edited destination survives a multi-rank filter", async () => {
    const peers: DepartmentNodePoolResources[] = Array.from({ length: 12 }, (_, index) => ({
      id: "resource-" + index,
      clusterId: "c",
      nodePoolId: "p",
      nodePoolName: "Pool",
      orgUnitId: "department-" + index,
      orgUnitName: "Department " + index,
      rank: index < 4 ? "High" : index < 8 ? "MediumLow" : "Low",
      resources: { gpu: { overQuotaWeight: 2, deserved: 0 } },
    }));
    const wrapper = mount(Table, {
      props: {
        rows: peers,
        baselineRows: peers,
        aggregates: {},
        draft: {
          mode: "ranked",
          automaticWeights: false,
          cpuQuota: false,
          departments: {},
        },
        rankCounts: { High: 4, MediumLow: 4, Low: 4 },
        ranks: ["High", "MediumLow", "Low", "Lowest"],
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    try {
      for (const rank of ["High", "Low"]) {
        await wrapper.get('[aid="rank-' + rank + '"]').trigger("click");
      }
      expect(wrapper.findAll(".rank-header")).toHaveLength(3);
      await wrapper.setProps({
        draft: {
          mode: "ranked",
          automaticWeights: false,
          cpuQuota: false,
          departments: {
            "department-0": { id: "department-0", rank: "Lowest" },
            "department-4": { id: "department-4", rank: "Lowest" },
          },
        },
      });
      await wrapper.get('[aid="rank-Lowest"]').trigger("click");
      expect(wrapper.findAll(".rank-header")).toHaveLength(4);
      wrapper.findComponent(RunaiPageActions).vm.$emit("filters-changed", {
        advancedFilters: [
          { name: "rank", type: "EnumBased", selectedValues: ["High", "Lowest"], selectOptions: [], label: "Rank" },
        ],
      });
      await wrapper.vm.$nextTick();
      expect(wrapper.findAll(".rank-header").map((group) => group.text())).toEqual([
        "High (3 departments)",
        "Lowest (2 departments)",
      ]);
      expect(wrapper.findAll(".department-name")).toHaveLength(5);
    } finally {
      wrapper.unmount();
    }
  });
});
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
  it("clamps demand and immediately corrects the native handle without changing the pool scale", async () => {
    const wrapper = mount(Table, {
      props: {
        rows,
        draft: {
          ...draft,
          mode: "guaranteed",
          automaticWeights: false,
          maxAllocation: true,
          departments: { a: { id: "a", limit: 50 } },
        },
        capacity: 64,
        aggregates: {},
        ranks: ["MediumLow"],
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    try {
      await wrapper.findComponent(QToggle).vm.$emit("update:modelValue", true);
      await wrapper.vm.$nextTick();
      const input = wrapper.get('input[type="range"]');
      await input.setValue("64");
      expect((input.element as HTMLInputElement).value).toBe("50");
      expect(input.attributes("max")).toBe("64");
      await input.setValue("64");
      expect((input.element as HTMLInputElement).value).toBe("50");
      await wrapper.setProps({
        draft: { ...draft, mode: "guaranteed", maxAllocation: true, departments: { a: { id: "a", limit: 12 } } },
      });
      expect((input.element as HTMLInputElement).value).toBe("12");
    } finally {
      wrapper.unmount();
    }
  });
  it.each([
    { value: 0.001, automatic: true, display: "<0.1%", marker: "0.01%" },
    { value: 0, automatic: true, display: "0%", marker: "0%" },
    { value: 4.996, automatic: true, display: "50%", marker: "49.96%" },
    { value: 4.996, automatic: false, display: "49.96%", marker: "49.96%" },
  ])(
    "formats weight $value with automatic=$automatic without rounding geometry",
    ({ value, automatic, display, marker }) => {
      const tiny = { ...rows[0], resources: { gpu: { deserved: value, overQuotaWeight: value } } };
      const wrapper = mount(Table, {
        props: {
          rows: [tiny],
          baselineRows: [tiny],
          draft: { ...draft, mode: "guaranteed", automaticWeights: automatic },
          aggregates: { MediumLow: { weightsSum: 10, automaticWeightsSum: 10, deservedSum: 10 } },
          ranks: ["MediumLow"],
          readonly: false,
          loading: false,
        },
        global: {
          plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
          directives: { "click-outside": {}, permission: {} },
        },
      });
      try {
        const editor = wrapper.findAllComponents(DepartmentResourceEditor).find((e) => e.props("kind") === "weight");
        expect(editor?.props("displayValue")).toBe(display);
        expect(parseFloat((wrapper.get(".quota-marker").element as HTMLElement).style.left)).toBeCloseTo(
          parseFloat(marker),
          8,
        );
      } finally {
        wrapper.unmount();
      }
    },
  );
  const weightPopup = {
    template: '<td><span class="applied-weight">{{ displayValue }}</span><slot name="input" /></td>',
    props: ["displayValue", "isSubmitDisabled"],
    emits: ["apply", "reset", "opened"],
  };
  it.each(["proportional", "ranked"] as const)(
    "never shows maximum-allocation columns or editors in %s with stale draft limits",
    (mode) => {
      const wrapper = mount(Table, {
        props: {
          rows,
          draft: { ...draft, mode, maxAllocation: true, departments: { a: { id: "a", limit: 7 } } },
          aggregates: {},
          ranks: ["MediumLow"],
          readonly: false,
          loading: false,
        },
        global: {
          plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
          directives: { "click-outside": {}, permission: {} },
        },
      });
      expect(wrapper.get("thead").text()).not.toContain("Max GPU allocation");
      expect(
        wrapper.findAllComponents(DepartmentResourceEditor).filter((editor) => editor.props("kind") === "limit"),
      ).toHaveLength(0);
      wrapper.unmount();
    },
  );
  it("preserves table space while filtering and releases it when cleared", async () => {
    const wrapper = mount(Table, {
      props: { rows, draft, aggregates: {}, ranks: [], readonly: false, loading: false },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    wrapper.findComponent(QResizeObserver).vm.$emit("resize", { height: 420, width: 1000 });
    await wrapper.setProps({ searchTerm: "missing", rows: [] });
    expect(wrapper.get(".department-table-content").attributes("style")).toContain("min-height: 420px");
    wrapper.findComponent(QResizeObserver).vm.$emit("resize", { height: 100, width: 1000 });
    await wrapper.vm.$nextTick();
    expect(wrapper.get(".department-table-content").attributes("style")).toContain("420px");
    await wrapper.setProps({ searchTerm: "", rows });
    expect(wrapper.get(".department-table-content").attributes("style") ?? "").not.toContain("min-height");
    wrapper.unmount();
  });
  it.each([
    ["proportional", "6.09%"],
    ["ranked", "46.67%"],
    ["guaranteed", "46.67%"],
  ] as const)(
    "previews the proposed weight share in %s without changing applied table values",
    async (mode, proposed) => {
      const peers = [
        rows[0],
        { ...rows[0], orgUnitId: "b", orgUnitName: "Beta", resources: { gpu: { overQuotaWeight: 1 } } },
        { ...rows[0], orgUnitId: "c", orgUnitName: "Gamma", resources: { gpu: { overQuotaWeight: 7 } } },
        { ...rows[0], orgUnitId: "d", orgUnitName: "Delta", resources: { gpu: { overQuotaWeight: 0 } } },
        {
          ...rows[0],
          orgUnitId: "h",
          orgUnitName: "Higher",
          rank: "High",
          resources: { gpu: { overQuotaWeight: 100 } },
        },
      ];
      const currentDraft: SchedulingDraft = { ...draft, mode };
      const wrapper = mount(Table, {
        props: {
          rows: peers,
          baselineRows: peers,
          draft: currentDraft,
          aggregates: {
            all: { weightsSum: 110, deservedSum: 3 },
            MediumLow: { weightsSum: 10, deservedSum: 3 },
            High: { weightsSum: 100, deservedSum: 0 },
          },
          ranks: ["High", "MediumLow"],
          readonly: false,
          loading: false,
        },
        global: {
          plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
          directives: { "click-outside": {}, permission: {} },
          stubs: { EditableCellPopup: weightPopup },
        },
      });
      const editor = wrapper
        .findAllComponents(DepartmentResourceEditor)
        .find((item) => item.props("kind") === "weight" && item.props("id") === "a");
      if (!editor) throw new Error("Weight editor is unavailable.");
      const applied = editor.get(".applied-weight").text();
      const markerBefore = wrapper.findAll(".quota-marker").map((marker) => marker.attributes("style"));
      editor.findComponent(weightPopup).vm.$emit("opened");
      await editor.get("input").setValue("7");
      expect(editor.get(".weight-percentage").text()).toBe(`(${proposed})`);
      expect(editor.get(".applied-weight").text()).toBe(applied);
      expect(wrapper.findAll(".quota-marker").map((marker) => marker.attributes("style"))).toEqual(markerBefore);
      expect(wrapper.emitted("edit")).toBeUndefined();
      editor.findComponent(weightPopup).vm.$emit("reset");
      await wrapper.vm.$nextTick();
      expect((editor.get("input").element as HTMLInputElement).value).toBe("2");
      expect(editor.get(".weight-percentage").text()).toBe(`(${applied})`);
      await editor.get("input").setValue("7");
      editor.findComponent(weightPopup).vm.$emit("apply");
      expect(wrapper.emitted("edit")).toEqual([[{ id: "a", weight: 7 }]]);
      await wrapper.setProps({ draft: { ...currentDraft, departments: { a: { id: "a", weight: 7 } } } });
      expect(editor.get(".applied-weight").text()).toBe(proposed);
      wrapper.unmount();
    },
  );
  it("previews zero weights and avoids a stale percentage for invalid input", async () => {
    const zero = { ...rows[0], resources: { gpu: { overQuotaWeight: 0 } } };
    const wrapper = mount(Table, {
      props: {
        rows: [zero],
        baselineRows: [zero],
        draft,
        aggregates: { all: { weightsSum: 0, deservedSum: 0 } },
        ranks: [],
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
        stubs: { EditableCellPopup: weightPopup },
      },
    });
    const editor = wrapper.findComponent(DepartmentResourceEditor);
    expect(editor.get(".weight-percentage").text()).toBe("(0%)");
    await editor.get("input").setValue("5");
    expect(editor.get(".weight-percentage").text()).toBe("(100%)");
    for (const invalid of ["", "-1", "0.5"]) {
      await editor.get("input").setValue(invalid);
      expect(editor.find(".weight-percentage").exists()).toBe(false);
      editor.findComponent(weightPopup).vm.$emit("apply");
    }
    expect(wrapper.emitted("edit")).toBeUndefined();
    wrapper.unmount();
  });
  it("previews the final unknown custom weight only when every other peer is known with full visibility", async () => {
    const peers = [rows[0], { ...rows[0], orgUnitId: "b", orgUnitName: "Beta" }];
    const currentDraft: SchedulingDraft = { ...draft, departments: { b: { id: "b", weight: 3 } } };
    const wrapper = mount(Table, {
      props: {
        rows: peers,
        baselineRows: peers,
        draft: currentDraft,
        aggregates: { all: { weightsSum: 6, deservedSum: 6, rawWeightsSum: undefined } },
        ranks: [],
        rawWeightsAvailable: false,
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
        stubs: { EditableCellPopup: weightPopup },
      },
    });
    const editor = wrapper.findAllComponents(DepartmentResourceEditor).find((item) => item.props("id") === "a");
    if (!editor) throw new Error("Weight editor is unavailable.");
    editor.findComponent(weightPopup).vm.$emit("opened");
    await editor.get("input").setValue("7");
    expect(editor.get(".weight-percentage").text()).toBe("(70%)");
    expect(editor.get(".applied-weight").text()).toBe("Unknown");
    await wrapper.setProps({ totalsComplete: false });
    expect(editor.get(".weight-percentage").text()).toBe("(Unknown)");
    await wrapper.setProps({ totalsComplete: true, baselineRows: undefined });
    expect(editor.get(".weight-percentage").text()).toBe("(Unknown)");
    await wrapper.setProps({ baselineRows: peers, draft });
    expect(editor.get(".weight-percentage").text()).toBe("(Unknown)");
    expect(wrapper.emitted("edit")).toBeUndefined();
    wrapper.unmount();
  });
  it("allows quota above a preserved limit only after maximum allocation is disabled", async () => {
    const resourcePopup = {
      template: '<td><slot name="input" /></td>',
      emits: ["apply", "reset", "opened"],
    };
    const currentDraft: SchedulingDraft = {
      ...draft,
      mode: "guaranteed",
      maxAllocation: true,
      departments: { a: { id: "a", quota: 20, limit: 30 } },
    };
    const wrapper = mount(Table, {
      props: {
        rows,
        draft: currentDraft,
        aggregates: { all: { weightsSum: 2, deservedSum: 3 } },
        ranks: ["MediumLow"],
        capacity: 100,
        readonly: false,
        loading: false,
        loadQuotaContext: async () => ({ childQuotaSum: 0 }),
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
        stubs: { EditableCellPopup: resourcePopup },
      },
    });
    const quota = wrapper.findAllComponents(DepartmentResourceEditor).find((editor) => editor.props("kind") === "quota");
    if (!quota) throw new Error("Quota editor is unavailable.");
    quota.findComponent(resourcePopup).vm.$emit("opened");
    await flushPromises();
    await quota.get("input").setValue("40");
    quota.findComponent(resourcePopup).vm.$emit("apply");
    expect(wrapper.emitted("edit")).toBeUndefined();
    await wrapper.setProps({ draft: { ...currentDraft, maxAllocation: false } });
    quota.findComponent(resourcePopup).vm.$emit("apply");
    expect(wrapper.emitted("edit")).toEqual([[{ id: "a", quota: 40 }]]);
    expect(wrapper.props("draft").departments.a.limit).toBe(30);
    await wrapper.setProps({ draft: currentDraft });
    quota.findComponent(resourcePopup).vm.$emit("apply");
    expect(wrapper.emitted("edit")).toHaveLength(1);
    wrapper.unmount();
  });
  it.each(["proportional", "ranked"] as const)(
    "enforces maximum allocation only while Guaranteed quota is active, starting from %s",
    async (mode) => {
      const resourcePopup = { template: '<td><slot name="input" /></td>', emits: ["apply", "reset", "opened"] };
      const currentDraft: SchedulingDraft = {
        ...draft,
        mode,
        maxAllocation: true,
        departments: { a: { id: "a", quota: 20, limit: 30 } },
      };
      const wrapper = mount(Table, {
        props: {
          rows,
          draft: currentDraft,
          aggregates: {},
          ranks: ["MediumLow"],
          readonly: false,
          loading: false,
        },
        global: {
          plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
          directives: { "click-outside": {}, permission: {} },
          stubs: { EditableCellPopup: resourcePopup },
        },
      });
      expect(
        wrapper.findAllComponents(DepartmentResourceEditor).some((editor) => editor.props("kind") === "limit"),
      ).toBe(false);
      await wrapper.setProps({ draft: { ...currentDraft, mode: "guaranteed" } });
      const maximum = wrapper
        .findAllComponents(DepartmentResourceEditor)
        .find((editor) => editor.props("kind") === "limit");
      if (!maximum) throw new Error("Maximum allocation editor is unavailable.");
      await maximum.get("input").setValue("10");
      maximum.findComponent(resourcePopup).vm.$emit("apply");
      expect(wrapper.emitted("edit")).toBeUndefined();
      await maximum.get("input").setValue("25");
      maximum.findComponent(resourcePopup).vm.$emit("apply");
      expect(wrapper.emitted("edit")).toEqual([[{ id: "a", limit: 25 }]]);
      await wrapper.setProps({ draft: currentDraft });
      expect(
        wrapper.findAllComponents(DepartmentResourceEditor).some((editor) => editor.props("kind") === "limit"),
      ).toBe(false);
      expect(wrapper.props("draft").departments.a.quota).toBe(20);
      wrapper.unmount();
    },
  );
  it("hides the rank action in Proportional and restores the selection when returning to Ranked", async () => {
    const currentDraft: SchedulingDraft = {
      ...draft,
      mode: "ranked",
      departments: { a: { id: "a", rank: "MediumLow" } },
    };
    const wrapper = mount(Table, {
      props: {
        rows,
        draft: currentDraft,
        aggregates: {},
        ranks: ["MediumLow"],
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    const alpha = wrapper.findAllComponents(QCheckbox).find((item) => item.attributes("aria-label") === "Select Alpha");
    if (!alpha) throw new Error("Department selection is unavailable.");
    alpha.vm.$emit("update:modelValue", true);
    await wrapper.vm.$nextTick();
    expect(wrapper.get('[aid="move-departments"]').isVisible()).toBe(true);
    await wrapper.setProps({ draft: { ...currentDraft, mode: "proportional" } });
    expect(wrapper.find('[aid="move-departments"]').exists()).toBe(false);
    expect(wrapper.find(".selected-rows-actions").exists()).toBe(false);
    expect(wrapper.get('[aid="advanced-settings"]').isVisible()).toBe(true);
    expect(wrapper.props("draft").departments.a.rank).toBe("MediumLow");
    await wrapper.setProps({ draft: currentDraft });
    expect(wrapper.get('[aid="move-departments"]').isVisible()).toBe(true);
    expect(wrapper.get(".selected-rows-actions").text()).toContain("1 selected");
    const restored = wrapper
      .findAllComponents(QCheckbox)
      .find((item) => item.attributes("aria-label") === "Select Alpha");
    expect(restored?.props("modelValue")).toBe(true);
    wrapper.unmount();
  });
  it("recovers custom percentages only from a complete set of known draft peer weights", async () => {
    const peers = [
      rows[0],
      { ...rows[0], orgUnitId: "b", orgUnitName: "Beta" },
      { ...rows[0], orgUnitId: "c", orgUnitName: "Gamma" },
      { ...rows[0], orgUnitId: "d", orgUnitName: "Delta" },
    ];
    const currentDraft: SchedulingDraft = {
      ...draft,
      mode: "guaranteed",
      departments: { a: { id: "a", weight: 0 }, b: { id: "b", weight: 1 }, c: { id: "c", weight: 2 } },
    };
    const wrapper = mount(Table, {
      props: {
        rows: peers,
        baselineRows: peers,
        draft: currentDraft,
        aggregates: { MediumLow: { weightsSum: 12, deservedSum: 12, rawWeightsSum: undefined } },
        ranks: ["MediumLow"],
        rawWeightsAvailable: false,
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    const displayedWeights = () =>
      wrapper
        .findAllComponents(DepartmentResourceEditor)
        .filter((editor) => editor.props("kind") === "weight")
        .map((editor) => editor.get('[role="button"]').text());
    expect(displayedWeights()).toEqual(["Unknown", "Unknown", "Unknown", "Unknown"]);
    const completeDraft = { ...currentDraft, departments: { ...currentDraft.departments, d: { id: "d", weight: 7 } } };
    await wrapper.setProps({ draft: completeDraft });
    expect(displayedWeights()).toEqual(["0%", "10%", "20%", "70%"]);
    await wrapper.setProps({ rows: [peers[1]], searchTerm: "Beta" });
    expect(displayedWeights()).toEqual(["10%"]);
    await wrapper.setProps({
      draft: {
        ...completeDraft,
        departments: { ...completeDraft.departments, d: { id: "d", weight: 7, rank: "High" } },
      },
    });
    expect(displayedWeights()).toEqual(["33.33%"]);
    await wrapper.setProps({ totalsComplete: false });
    expect(displayedWeights()).toEqual(["Unknown"]);
    await wrapper.setProps({ totalsComplete: true, baselineRows: undefined });
    expect(displayedWeights()).toEqual(["Unknown"]);
    wrapper.unmount();
  });
  it("counts draft rank moves before offering to load more departments", async () => {
    const wrapper = mount(Table, {
      props: {
        rows: [...rows, { ...rows[0], id: "resource-b", orgUnitId: "b", orgUnitName: "Beta", rank: "High" }],
        draft: { ...draft, mode: "ranked", departments: { a: { id: "a", rank: "High" } } },
        aggregates: {},
        ranks: ["High", "MediumLow"],
        rankCounts: { High: 1, MediumLow: 1 },
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    await wrapper.get('[aid="rank-High"]').trigger("click");
    expect(wrapper.text()).toContain("High (2 departments)");
    expect(wrapper.text()).toContain("Alpha");
    expect(wrapper.text()).toContain("Beta");
    expect(wrapper.find('[aid="load-rank-High"]').exists()).toBe(false);
    await wrapper.setProps({ rankCounts: { High: 3, MediumLow: 1 } });
    expect(wrapper.text()).toContain("High (4 departments)");
    await wrapper.get('[aid="load-rank-High"]').trigger("click");
    expect(wrapper.emitted("load-rank")).toEqual([["High"], ["High"]]);
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
  it("renders API results, defaults demand to half capacity, and preserves drafts on errors", async () => {
    vi.useFakeTimers();
    const transport = vi.fn().mockResolvedValue({
      capacity: { gpu: 144 },
      orgUnits: [
        {
          id: "a",
          fairShare: { gpu: 12 },
          guaranteed: { gpu: 0 },
          overQuota: { gpu: 12 },
          unmet: { gpu: 60 },
          demand: { gpu: 72 },
          weightShareInRank: 0.25,
        },
      ],
    });
    const wrapper = mount(Table, {
      props: { poolId: "17", rows, draft, capacity: 144, aggregates: {}, ranks: [], readonly: true, loading: false },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
        provide: { [simulationPreviewKey as symbol]: { transport } },
      },
    });
    expect(wrapper.findComponent(QToggle).props("disable")).toBe(false);
    wrapper.findComponent(QToggle).vm.$emit("update:modelValue", true);
    await wrapper.vm.$nextTick();
    await vi.advanceTimersByTimeAsync(300);
    expect((wrapper.get(".demand-slider").element as HTMLInputElement).value).toBe("72");
    expect(wrapper.get(".allocated").text()).toBe("12");
    expect(wrapper.get(".unmet").attributes("style")).toContain("41.666");
    expect(wrapper.get(".demand-slider").attributes("disabled")).toBeUndefined();
    expect(wrapper.text()).toContain("25%");
    expect(wrapper.text()).toContain("mock results");
    transport.mockRejectedValue({ response: { status: 501 } });
    await wrapper.setProps({ draft: { ...draft, departments: { a: { id: "a", weight: 8 } } } });
    await vi.advanceTimersByTimeAsync(300);
    expect(wrapper.find(".allocated").exists()).toBe(false);
    expect(wrapper.text()).toContain("not available on this backend");
    expect(wrapper.props("draft").departments.a.weight).toBe(8);
    wrapper.unmount();
    vi.useRealTimers();
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
    expect(parseFloat((wrapper.get(".quota-marker").element as HTMLElement).style.left)).toBeCloseTo(38.4615, 3);
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
