/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ENodePoolSchedulingMode as ESchedulingMode } from "@/models/node-pool.model";
import { flushPromises, mount } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createTestingPinia } from "@pinia/testing";
import { QCheckbox, QToggle, QResizeObserver, Quasar } from "quasar";
import type { SchedulingDraft } from "@/models/node-pool-scheduling.model";
import type { DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client/index";
import { SchedulingMoveToRankDialog as MoveToRankModal } from "@/components/node-pools/scheduling-move-to-rank-dialog";
import { RunaiPageActions } from "@/components/common/runai-page-actions/index";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", async (importOriginal) => {
  const original = await importOriginal<{ orgUnitService: typeof orgUnitService }>();
  return { ...original, orgUnitService: { ...original.orgUnitService, simulateNodePoolFairShare: vi.fn() } };
});
const simulateNodePool = orgUnitService.simulateNodePoolFairShare;
import Table from "@/components/node-pools/node-pool-department-table/node-pool-department-table.vue";
import { SchedulingFilters } from "@/components/node-pools/scheduling-filters";
import { DepartmentResourceEditor } from "@/components/node-pools/department-resource-editor";

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
          mode: ESchedulingMode.Ranked,
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
          mode: ESchedulingMode.Ranked,
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
const draft: SchedulingDraft = {
  mode: ESchedulingMode.Proportional,
  automaticWeights: false,
  cpuQuota: false,
  departments: {},
};

function renderSimulation(capacity?: number) {
  return mount(Table, {
    props: {
      poolId: "17",
      rows,
      draft,
      capacity,
      aggregates: { all: { weightsSum: 2, deservedSum: 3 } },
      ranks: [],
      readonly: true,
      loading: false,
    },
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      directives: { "click-outside": {}, permission: {} },
    },
  });
}

describe("simulation capacity", () => {
  const hint = "Simulation needs the node pool’s GPU capacity.";
  let wrapper: ReturnType<typeof renderSimulation>;

  beforeEach(() => {
    vi.useFakeTimers();
    vi.mocked(simulateNodePool).mockReset().mockResolvedValue({ orgUnits: [] });
  });
  afterEach(() => {
    wrapper.unmount();
    vi.useRealTimers();
  });

  it("explains the disabled switch when capacity is unknown and skips simulation requests", async () => {
    wrapper = renderSimulation();
    const toggle = wrapper.get('[role="switch"]');
    expect(toggle.isVisible()).toBe(true);
    expect(toggle.attributes("aria-checked")).toBe("false");
    expect(toggle.attributes("aria-disabled")).toBe("true");
    const description = wrapper.get(`[id="${toggle.attributes("aria-describedby")}"]`);
    expect(description.text()).toBe(hint);
    expect(description.isVisible()).toBe(true);
    expect(toggle.element.closest("th")?.contains(description.element)).toBe(true);
    await toggle.trigger("click");
    await vi.advanceTimersByTimeAsync(500);
    expect(wrapper.find(".demand-slider").exists()).toBe(false);
    expect(simulateNodePool).not.toHaveBeenCalled();
  });

  it("enables simulation for a confirmed zero only with department read permission", async () => {
    wrapper = renderSimulation();
    await wrapper.setProps({ capacity: 0, canView: false });
    const toggle = wrapper.get('[role="switch"]');
    expect(toggle.attributes("aria-disabled")).toBe("true");
    expect(toggle.attributes("aria-describedby")).toBeUndefined();
    expect(wrapper.text()).not.toContain(hint);
    await toggle.trigger("click");
    await vi.advanceTimersByTimeAsync(500);
    expect(simulateNodePool).not.toHaveBeenCalled();

    await wrapper.setProps({ canView: true });
    expect(toggle.attributes("aria-disabled")).toBeUndefined();
    expect(toggle.attributes("aria-checked")).toBe("false");
    await toggle.trigger("click");
    await vi.advanceTimersByTimeAsync(500);
    expect(toggle.attributes("aria-checked")).toBe("true");
    expect(vi.mocked(simulateNodePool).mock.lastCall?.[1].nodePool?.capacity?.gpu).toBe(0);
  });

  it("shows a loading track until the selected mode has allocations", async () => {
    const result = {
      orgUnits: [
        {
          id: "a",
          fairShare: { gpu: 4 },
          guaranteed: { gpu: 0 },
          overQuota: { gpu: 4 },
          unmet: { gpu: 6 },
          demand: { gpu: 10 },
          weightShareInRank: 1,
        },
      ],
    };
    vi.mocked(simulateNodePool).mockResolvedValue(result);
    wrapper = renderSimulation(40);
    await wrapper.get('[role="switch"]').trigger("click");
    await vi.advanceTimersByTimeAsync(300);
    expect(wrapper.find(".simulation-values").exists()).toBe(true);
    await wrapper.setProps({ draft: { ...draft, mode: ESchedulingMode.Ranked } });
    expect(wrapper.find('.result-bar[aria-busy="true"] .q-skeleton').exists()).toBe(true);
    expect(wrapper.find(".simulation-values").exists()).toBe(false);
    await vi.advanceTimersByTimeAsync(0);
    expect(wrapper.find(".simulation-values").exists()).toBe(true);
    await wrapper.setProps({ draft });
    expect(wrapper.find(".q-skeleton").exists()).toBe(false);
    expect(wrapper.find('.result-bar[aria-busy="true"] .simulation-values').exists()).toBe(true);
  });

  it("shows all 51 matching departments in a rank without a paging action", async () => {
    wrapper = renderSimulation(40);
    const members = Array.from({ length: 51 }, (_, i) => ({
      ...rows[0],
      orgUnitId: String(i),
      orgUnitName: `demo-${i}`,
      rank: "Medium",
    }));
    await wrapper.setProps({
      rows: [...members, rows[0]],
      ranks: ["Medium"],
      searchTerm: "demo-",
      draft: { ...draft, mode: ESchedulingMode.Ranked },
    });
    expect(wrapper.text()).toContain("Medium (51 departments)");
    await wrapper.get('[aid="rank-Medium"]').trigger("click");
    expect(wrapper.findAll(".department-name")).toHaveLength(51);
    expect(wrapper.text()).not.toContain("Load more");
  });

  it("turns simulation off when all ranks collapse or a search hides every department", async () => {
    wrapper = renderSimulation(40);
    await wrapper.setProps({ draft: { ...draft, mode: ESchedulingMode.Ranked } });
    const toggle = wrapper.get('[role="switch"]');
    await toggle.trigger("click");
    await wrapper.get('[aid="rank-MediumLow"]').trigger("click");
    expect(toggle.attributes("aria-disabled")).toBe("true");
    expect(toggle.attributes("aria-checked")).toBe("false");
    await wrapper.get('[aid="rank-MediumLow"]').trigger("click");
    await toggle.trigger("click");
    await wrapper.setProps({ searchTerm: "no match" });
    expect(toggle.attributes("aria-disabled")).toBe("true");
    expect(toggle.attributes("aria-checked")).toBe("false");
    expect(wrapper.find(".simulation-notice").exists()).toBe(false);
    await vi.advanceTimersByTimeAsync(500);
    expect(simulateNodePool).not.toHaveBeenCalled();
  });

  it("cancels a queued simulation if capacity disappears before the request is sent", async () => {
    wrapper = renderSimulation(8);
    await wrapper.get('[role="switch"]').trigger("click");
    expect(wrapper.find(".demand-slider").exists()).toBe(true);
    await wrapper.setProps({ capacity: undefined });
    await vi.advanceTimersByTimeAsync(500);
    expect(simulateNodePool).not.toHaveBeenCalled();
  });

  it("restores the normal Result view on capacity loss and keeps demand for an explicit restart", async () => {
    wrapper = renderSimulation(40);
    const toggle = wrapper.get('[role="switch"]');
    await toggle.trigger("click");
    await wrapper.get('input[aria-label="Simulated GPU demand for Alpha"]').setValue("13");
    await vi.advanceTimersByTimeAsync(500);
    expect(simulateNodePool).toHaveBeenCalledTimes(1);
    expect(vi.mocked(simulateNodePool).mock.lastCall?.[1].orgUnits[0].demand?.gpu).toBe(13);

    await wrapper.setProps({ capacity: undefined });
    expect(toggle.attributes("aria-checked")).toBe("false");
    expect(toggle.attributes("aria-disabled")).toBe("true");
    expect(wrapper.get(`[id="${toggle.attributes("aria-describedby")}"]`).text()).toBe(hint);
    expect(wrapper.find(".demand-slider").exists()).toBe(false);
    expect(wrapper.get(".result-bar").classes()).not.toContain("simulation-bar");
    expect(wrapper.find(".quota-marker").exists()).toBe(true);
    await vi.advanceTimersByTimeAsync(500);
    expect(simulateNodePool).toHaveBeenCalledTimes(1);

    await wrapper.setProps({ capacity: 32 });
    expect(toggle.attributes("aria-checked")).toBe("false");
    expect(toggle.attributes("aria-disabled")).toBeUndefined();
    expect(toggle.attributes("aria-describedby")).toBeUndefined();
    expect(wrapper.text()).not.toContain(hint);
    await vi.advanceTimersByTimeAsync(500);
    expect(simulateNodePool).toHaveBeenCalledTimes(1);
    await toggle.trigger("click");
    expect((wrapper.get(".demand-slider").element as HTMLInputElement).value).toBe("13");
    await vi.advanceTimersByTimeAsync(500);
    expect(vi.mocked(simulateNodePool).mock.lastCall?.[1].orgUnits[0].demand?.gpu).toBe(13);
    expect(simulateNodePool).toHaveBeenCalledTimes(2);
  });
});

describe("scheduling department table", () => {
  it("shows rank filters and row selection only for ranked scheduling modes", async () => {
    const wrapper = mount(Table, {
      props: { rows, draft, aggregates: {}, ranks: ["MediumLow"], readonly: false, canView: true, loading: false },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
      },
    });
    try {
      expect(wrapper.findComponent(SchedulingFilters).props("ranked")).toBe(false);
      expect(wrapper.find('[aria-label="Select Alpha"]').exists()).toBe(false);
      await wrapper.setProps({ draft: { ...draft, mode: ESchedulingMode.Ranked } });
      expect(wrapper.findComponent(SchedulingFilters).props("ranked")).toBe(true);
      expect(wrapper.find('[aria-label="Select Alpha"]').exists()).toBe(true);
    } finally {
      wrapper.unmount();
    }
  });
  it("clamps demand and immediately corrects the native handle without changing the pool scale", async () => {
    const wrapper = mount(Table, {
      props: {
        rows,
        draft: {
          ...draft,
          mode: ESchedulingMode.Guaranteed,
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
        draft: {
          ...draft,
          mode: ESchedulingMode.Guaranteed,
          maxAllocation: true,
          departments: { a: { id: "a", limit: 12 } },
        },
      });
      expect((input.element as HTMLInputElement).value).toBe("12");
    } finally {
      wrapper.unmount();
    }
  });
  it("commits integer demand on change, keeps drag previews local, and preserves demand across modes", async () => {
    vi.useFakeTimers();
    const transport = vi
      .mocked(simulateNodePool)
      .mockReset()
      .mockResolvedValue({
        orgUnits: [
          {
            id: "a",
            fairShare: { gpu: 10.4 },
            guaranteed: { gpu: 3.2 },
            overQuota: { gpu: 7.2 },
            unmet: { gpu: 22.1 },
            demand: { gpu: 32.5 },
            weightShareInRank: 0.25,
          },
        ],
      });
    const currentDraft: SchedulingDraft = {
      ...draft,
      mode: ESchedulingMode.Guaranteed,
      maxAllocation: true,
      departments: { a: { id: "a", limit: 40.8 } },
    };
    const wrapper = mount(Table, {
      props: {
        poolId: "17",
        rows,
        draft: currentDraft,
        capacity: 65,
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
      wrapper.findComponent(QToggle).vm.$emit("update:modelValue", true);
      await wrapper.vm.$nextTick();
      await vi.advanceTimersByTimeAsync(300);
      expect(transport).toHaveBeenCalledTimes(1);
      expect(transport.mock.calls[0][1].orgUnits[0].demand).toBeUndefined();
      const input = wrapper.get('input[type="range"]');
      expect(input.attributes("step")).toBe("1");
      expect(input.attributes("max")).toBe("65");
      expect(input.attributes("aria-valuemax")).toBe("40");
      expect((input.element as HTMLInputElement).value).toBe("33");
      (input.element as HTMLInputElement).value = "45.7";
      await input.trigger("input");
      expect((input.element as HTMLInputElement).value).toBe("40");
      await vi.advanceTimersByTimeAsync(600);
      expect(transport).toHaveBeenCalledTimes(1);
      await input.trigger("change");
      await vi.advanceTimersByTimeAsync(300);
      expect(transport.mock.calls.at(-1)?.[1].orgUnits[0].demand).toEqual({ gpu: 40 });
      // A native keyboard change also commits without requiring a drag/input event.
      (input.element as HTMLInputElement).value = "39";
      await input.trigger("change");
      await vi.advanceTimersByTimeAsync(300);
      expect(transport.mock.calls.at(-1)?.[1].orgUnits[0].demand).toEqual({ gpu: 39 });
      for (const mode of [ESchedulingMode.Proportional, ESchedulingMode.Ranked, ESchedulingMode.Guaranteed] as const) {
        await wrapper.setProps({ draft: { ...currentDraft, mode } });
        await vi.advanceTimersByTimeAsync(300);
        expect(wrapper.findComponent(QToggle).props("modelValue")).toBe(true);
        expect((wrapper.get('input[type="range"]').element as HTMLInputElement).value).toBe("39");
        expect(transport.mock.calls.at(-1)?.[1].orgUnits[0].demand).toEqual({ gpu: 39 });
      }
    } finally {
      wrapper.unmount();
      vi.useRealTimers();
    }
  });
  it("uses consistent integer simulation totals in labels and tooltip while retaining fractional geometry and settings", async () => {
    vi.useFakeTimers();
    vi.mocked(simulateNodePool)
      .mockReset()
      .mockResolvedValue({
        orgUnits: [
          {
            id: "a",
            fairShare: { gpu: 1.4 },
            guaranteed: { gpu: 0.6 },
            overQuota: { gpu: 0.8 },
            unmet: { gpu: 1.4 },
            demand: { gpu: 2.8 },
            weightShareInRank: 0.12345,
          },
        ],
      });
    const wrapper = mount(Table, {
      props: {
        poolId: "17",
        rows,
        draft: {
          ...draft,
          mode: ESchedulingMode.Guaranteed,
          maxAllocation: true,
          departments: { a: { id: "a", quota: 0.6, limit: 3.7 } },
        },
        capacity: 5,
        aggregates: {},
        ranks: ["MediumLow"],
        readonly: false,
        loading: false,
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        directives: { "click-outside": {}, permission: {} },
        stubs: { QTooltip: { template: '<div class="rendered-tooltip"><slot /></div>' } },
      },
    });
    try {
      wrapper.findComponent(QToggle).vm.$emit("update:modelValue", true);
      await wrapper.vm.$nextTick();
      await vi.advanceTimersByTimeAsync(300);
      expect(wrapper.get(".allocated").text()).toBe("1");
      expect(wrapper.get(".unmet").text()).toBe("2");
      expect(parseFloat((wrapper.get(".allocated").element as HTMLElement).style.width)).toBeCloseTo(28, 8);
      expect(parseFloat((wrapper.get(".unmet").element as HTMLElement).style.width)).toBeCloseTo(28, 8);
      const entries = wrapper.findAll(".result-tooltip-description").map((entry) => entry.text());
      expect(entries).toEqual([
        "Guaranteed GPU allocation 0.6",
        "Weight within rank 12.345%",
        "Max GPU allocation 3.7",
        "Expected guaranteed GPU allocation 1",
        "Expected over-quota GPU allocation 0",
        "Unmet GPU demand 2",
        "Simulated real-time GPU demand 3",
      ]);
      const allocationMarkers = () =>
        wrapper
          .findAll(".result-tooltip-entry")
          .filter((entry) => entry.text().includes("GPU allocation") && !entry.text().startsWith("Max"));
      for (const entry of allocationMarkers()) {
        expect(entry.get(".result-tooltip-swatch").classes()).toContain(
          entry.text().includes("over-quota") ? "swatch-over-quota" : "swatch-guaranteed",
        );
      }
      for (const mode of [ESchedulingMode.Proportional, ESchedulingMode.Ranked]) {
        await wrapper.setProps({ draft: { ...draft, mode } });
        await vi.advanceTimersByTimeAsync(300);
        const [expected] = allocationMarkers();
        expect(expected.text()).toContain("Expected GPU allocation");
        expect(expected.get(".result-tooltip-swatch").classes()).toContain("swatch-guaranteed");
      }
    } finally {
      wrapper.unmount();
      vi.useRealTimers();
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
          draft: { ...draft, mode: ESchedulingMode.Guaranteed, automaticWeights: automatic },
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
    emits: ["apply", "reset", "before-show"],
  };
  it.each([ESchedulingMode.Proportional, ESchedulingMode.Ranked] as const)(
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
    [ESchedulingMode.Proportional, "6.09%"],
    [ESchedulingMode.Ranked, "46.67%"],
    [ESchedulingMode.Guaranteed, "46.67%"],
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
      editor.findComponent(weightPopup).vm.$emit("before-show");
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
    editor.findComponent(weightPopup).vm.$emit("before-show");
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
      emits: ["apply", "reset", "before-show"],
    };
    const currentDraft: SchedulingDraft = {
      ...draft,
      mode: ESchedulingMode.Guaranteed,
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
    quota.findComponent(resourcePopup).vm.$emit("before-show");
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
  it.each([ESchedulingMode.Proportional, ESchedulingMode.Ranked] as const)(
    "enforces maximum allocation only while Guaranteed quota is active, starting from %s",
    async (mode) => {
      const resourcePopup = { template: '<td><slot name="input" /></td>', emits: ["apply", "reset", "before-show"] };
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
      await wrapper.setProps({ draft: { ...currentDraft, mode: ESchedulingMode.Guaranteed } });
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
      mode: ESchedulingMode.Ranked,
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
    await wrapper.setProps({ draft: { ...currentDraft, mode: ESchedulingMode.Proportional } });
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
      mode: ESchedulingMode.Guaranteed,
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
  it("counts draft rank moves and shows all matching departments", async () => {
    const wrapper = mount(Table, {
      props: {
        rows: [...rows, { ...rows[0], id: "resource-b", orgUnitId: "b", orgUnitName: "Beta", rank: "High" }],
        draft: { ...draft, mode: ESchedulingMode.Ranked, departments: { a: { id: "a", rank: "High" } } },
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
    const transport = vi.mocked(simulateNodePool).mockResolvedValue({
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
      },
    });
    expect(wrapper.findComponent(QToggle).props("disable")).toBe(false);
    wrapper.findComponent(QToggle).vm.$emit("update:modelValue", true);
    await wrapper.vm.$nextTick();
    expect(wrapper.text()).not.toContain("Calculating simulation");
    await vi.advanceTimersByTimeAsync(300);
    expect((wrapper.get(".demand-slider").element as HTMLInputElement).value).toBe("72");
    expect(wrapper.get(".allocated").text()).toBe("12");
    expect(wrapper.get(".unmet").attributes("style")).toContain("41.666");
    expect(wrapper.get(".demand-slider").attributes("disabled")).toBeUndefined();
    expect(wrapper.text()).toContain("25%");
    expect(wrapper.text()).not.toContain("mock results");
    expect(wrapper.text()).not.toContain("Updating simulation");
    expect(wrapper.text()).not.toContain("Previous results shown");
    expect(wrapper.text()).not.toContain("Calculating simulation");
    expect(wrapper.text()).not.toContain("Simulation results are not available yet");
    transport.mockRejectedValue({ response: { status: 501 } });
    await wrapper.setProps({ draft: { ...draft, departments: { a: { id: "a", weight: 8 } } } });
    await vi.advanceTimersByTimeAsync(300);
    expect(wrapper.find(".allocated").exists()).toBe(false);
    expect(wrapper.text()).toContain("Simulation failed. Your changes are preserved. Try again.");
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
        draft: { ...draft, mode: ESchedulingMode.Ranked },
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
        mode: ESchedulingMode.Ranked,
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
        draft: { ...draft, mode: ESchedulingMode.Guaranteed, automaticWeights: true },
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
        draft: { ...draft, mode: ESchedulingMode.Ranked },
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
    expect(wrapper.text()).not.toContain("cannot view every department");
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
    expect(wrapper.text()).not.toContain("Raw weights are unavailable");
    expect(wrapper.text()).toContain("Unknown");
    expect(wrapper.findComponent({ name: "EditableCellPopup" }).props("displayValue")).toBe("Unknown");
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
