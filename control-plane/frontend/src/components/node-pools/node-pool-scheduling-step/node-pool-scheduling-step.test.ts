/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { mount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, it, vi } from "vitest";
import { createTestingPinia } from "@pinia/testing";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { NodePoolSchedulingSettings as Settings } from "@/components/node-pools/node-pool-scheduling-settings";
import { NodePoolDepartmentTable as Table } from "@/components/node-pools/node-pool-department-table";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiSegmentedButton } from "@/components/common/runai-segmented-button";
import NodePoolSchedulingStep from "./node-pool-scheduling-step.vue";
import { NodePoolNewDepartmentDialog as NewDepartment } from "@/components/node-pools/node-pool-new-department-dialog";
import RunaiWizardLayout from "@/components/common/runai-wizard-layout/runai-wizard-layout.vue";
import NodePoolWizardSummary from "../node-pool-wizard-summary/node-pool-wizard-summary.vue";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
const { createNodePoolDraft } = nodePoolUtil;
import { DepartmentResourceEditor } from "../department-resource-editor";
import type { Nodepool } from "@/swagger-models/cluster-service-client";

const saveMocks = vi.hoisted(() => ({ patchPool: vi.fn(), patchDepartments: vi.fn() }));
vi.mock("@/services/control-plane/cluster.service/cluster.service", async (original) => {
  const actual = await original<{ clusterService: typeof clusterService }>();
  return { ...actual, clusterService: { ...actual.clusterService, patchNodepool: saveMocks.patchPool } };
});
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", async (original) => {
  const actual = await original<{ orgUnitService: typeof orgUnitService }>();
  return {
    ...actual,
    orgUnitService: { ...actual.orgUnitService, batchPatchDepartmentNodePoolResources: saveMocks.patchDepartments },
  };
});

const pool = {
  ...createNodePoolDraft("cluster"),
  id: "pool-1",
  name: "training",
  clusterName: "Tokyo",
  networkTopologyName: "Existing fabric",
} as Nodepool;

function render(input = pool, stubs = {}) {
  return mount(NodePoolSchedulingStep, {
    props: { pool: input },
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      directives: { "click-outside": {}, permission: {} },
      stubs,
    },
  });
}

describe("scheduling step foundation", () => {
  it.each(["canEditDepartments", "canViewDepartments"] as const)(
    "finishes a pending department request when %s is revoked",
    async (permission) => {
      const wrapper = render(pool, {
        NodePoolNewDepartmentDialog: {
          emits: ["saving"],
          template: "<div />",
        },
      });
      try {
        await wrapper.setProps({ canSavePool: true });
        wrapper.findComponent(Table).vm.$emit("new-department");
        await flushPromises();
        const department = wrapper.findComponent(NewDepartment);
        department.vm.$emit("saving", true);
        await flushPromises();
        expect(await requestToLeave()).toBe(false);
        await wrapper.setProps({ [permission]: false });
        expect(wrapper.findComponent(NewDepartment).exists()).toBe(true);
        department.vm.$emit("saving", false);
        await flushPromises();
        expect(wrapper.emitted("saving")?.at(-1)).toEqual([false]);
        expect(wrapper.findComponent(NewDepartment).exists()).toBe(false);
        await wrapper.setProps({ [permission]: true });
        expect(wrapper.get('[aid="scheduling-save-continue"]').attributes("disabled")).toBeUndefined();
      } finally {
        wrapper.unmount();
      }
    },
  );

  it.each([
    { quota: false, ranks: false },
    { quota: false, ranks: true },
    { quota: true, ranks: true },
  ])("explains permission-disabled mode cards for the actual saved mode: %j", async (saved) => {
    const wrapper = render({
      ...pool,
      schedulingCapabilities: { adjustableWeights: true, cpuQuota: false, limits: false, ...saved },
    });
    try {
      await wrapper.setProps({ canEditDepartments: false });
      const options = wrapper.findComponent(RunaiSegmentedButton).props("options");
      expect(options.map((option) => option.disabledTooltip)).toEqual([
        "You do not have permission to edit departments.",
        "You do not have permission to edit departments.",
        "You do not have permission to edit departments.",
      ]);
      await wrapper.setProps({ canEditDepartments: true });
      const enabledOptions = wrapper.findComponent(RunaiSegmentedButton).props("options");
      expect(enabledOptions[0].disabledTooltip).toBe(
        saved.quota ? "This can’t be modified once a node pool has been created" : undefined,
      );
    } finally {
      wrapper.unmount();
    }
  });

  it.each(["ranked", "proportional"])(
    "requires cluster-wide update permission when saving a weights-off Ranked pool as %s",
    async (mode) => {
      const wrapper = render({
        ...pool,
        schedulingCapabilities: { quota: false, ranks: true, adjustableWeights: false, cpuQuota: false, limits: false },
      });
      try {
        await wrapper.setProps({ canSavePool: true, canResetResources: false });
        await wrapper.get(`[aid="mode-${mode}"]`).trigger("click");
        const save = wrapper.get('[aid="scheduling-save-close"]');
        expect(wrapper.get('[role="alert"]').text()).toContain(
          "Using adjustable weights without GPU quota requires update permission for all departments and projects in this cluster.",
        );
        expect(save.attributes("disabled")).toBeDefined();
        saveMocks.patchPool.mockClear();
        await save.trigger("click");
        await flushPromises();
        expect(saveMocks.patchPool).not.toHaveBeenCalled();

        await wrapper.setProps({ canResetResources: true });
        expect(wrapper.find('[role="alert"]').exists()).toBe(false);
        expect(save.attributes("disabled")).toBeUndefined();
      } finally {
        wrapper.unmount();
      }
    },
  );

  it("keeps saved GPU limits when switching from Ranked to Proportional without reset permission", async () => {
    const wrapper = render({
      ...pool,
      schedulingCapabilities: { quota: false, ranks: true, adjustableWeights: true, cpuQuota: false, limits: true },
    });
    try {
      await wrapper.setProps({ canSavePool: true, canResetResources: false });
      const table = wrapper.findComponent(Table);
      const summary = wrapper.findComponent(NodePoolWizardSummary);
      expect(table.props("draft").maxAllocation).toBe(true);
      expect(summary.text()).toContain("Ranked");
      await wrapper.get('[aid="mode-proportional"]').trigger("click");
      expect(table.props("draft").maxAllocation).toBe(true);
      expect(summary.text()).toContain("Proportional");
      expect(summary.props("pool").schedulingCapabilities?.limits).toBe(true);
      expect(wrapper.get('[aid="scheduling-save-close"]').attributes("disabled")).toBeUndefined();
    } finally {
      wrapper.unmount();
    }
  });

  it("keeps all table columns when Advanced settings opens and returns focus after Escape or Close", async () => {
    const wrapper = render({
      ...pool,
      schedulingCapabilities: { quota: true, ranks: true, adjustableWeights: false, cpuQuota: false, limits: true },
    });
    document.body.appendChild(wrapper.element);
    try {
      const trigger = wrapper.get('[aid="advanced-settings"]');
      await trigger.trigger("click");
      const panel = wrapper.get(".table-settings-layout .scheduling-settings");
      expect(trigger.attributes("aria-expanded")).toBe("true");
      expect(panel.element.parentElement?.id).toBe(trigger.attributes("aria-controls"));
      expect(wrapper.find(".column-result").exists()).toBe(true);
      expect(wrapper.find(".column-limit").exists()).toBe(true);
      expect(wrapper.find(".column-weight").exists()).toBe(true);
      const close = panel.get('[aria-label="Hide Advanced settings"]');
      expect(document.activeElement).toBe(close.element);
      await close.trigger("keydown", { key: "Escape" });
      expect(wrapper.find(".scheduling-settings").exists()).toBe(false);
      expect(wrapper.find(".column-result").exists()).toBe(true);
      expect(wrapper.find(".column-limit").exists()).toBe(true);
      expect(document.activeElement).toBe(trigger.element);

      await trigger.trigger("click");
      await wrapper.get('[aria-label="Hide Advanced settings"]').trigger("click");
      expect(trigger.attributes("aria-expanded")).toBe("false");
      expect(document.activeElement).toBe(trigger.element);
    } finally {
      wrapper.element.remove();
      wrapper.unmount();
    }
  });

  it("reveals a table control covered by Advanced settings when it receives focus", async () => {
    const wrapper = render();
    document.body.appendChild(wrapper.element);
    try {
      await wrapper.setProps({
        resources: { gpus: 8 },
        rows: [
          {
            id: "resource-1",
            orgUnitId: "1",
            orgUnitName: "Research",
            rank: "MediumLow",
            resources: { gpu: { overQuotaWeight: 2 } },
            clusterId: pool.clusterId,
            nodePoolId: pool.id,
            nodePoolName: pool.name,
          },
        ],
      });
      await wrapper.get('[aid="advanced-settings"]').trigger("click");
      const panel = wrapper.get(".department-settings-panel").element;
      const toggle = wrapper.get<HTMLElement>('[aid="simulate-demand"]').element;
      vi.spyOn(panel, "getBoundingClientRect").mockReturnValue(new DOMRect(620, 0, 380, 400));
      vi.spyOn(toggle, "getBoundingClientRect").mockReturnValue(new DOMRect(800, 20, 30, 30));
      toggle.focus();
      await flushPromises();
      expect(wrapper.find(".department-settings-panel").exists()).toBe(false);
      expect(document.activeElement).toBe(toggle);
    } finally {
      wrapper.element.remove();
      wrapper.unmount();
    }
  });

  it("previews Custom weights across all pages before saving an Automatic pool", async () => {
    const savedPool = {
      ...pool,
      schedulingCapabilities: { quota: true, ranks: true, adjustableWeights: false, cpuQuota: false, limits: false },
    };
    const peers = [-1, 0, 4].map((quota, i) => ({
      id: "resource-" + i,
      orgUnitId: String(i),
      orgUnitName: "Department " + i,
      clusterId: savedPool.clusterId,
      nodePoolId: savedPool.id,
      nodePoolName: savedPool.name,
      rank: "MediumLow",
      resources: { gpu: { deserved: quota, overQuotaWeight: quota } },
    }));
    const wrapper = render(savedPool);
    const table = wrapper.getComponent(Table);
    const weights = () =>
      table.findAllComponents(DepartmentResourceEditor).filter((cell) => cell.props("kind") === "weight");
    const labels = () => weights().map((cell) => cell.props("displayValue"));
    const pages = { nextOffset: 2 as number | undefined, loadingAll: false, loadNext: vi.fn(), loadAll: vi.fn() };
    vi.stubGlobal(
      "IntersectionObserver",
      class {
        observe() {}
        unobserve() {}
        disconnect() {}
      },
    );
    pages.loadAll.mockResolvedValue(true);
    try {
      await wrapper.setProps({
        canSavePool: true,
        rows: peers.slice(0, 2),
        baselineRows: peers.slice(0, 2),
        departmentPages: pages,
        ranks: ["MediumLow"],
        rankCounts: { MediumLow: 3 },
        aggregates: { MediumLow: { weightsSum: 4, deservedSum: -1, automaticWeightsSum: 4, rawWeightsSum: undefined } },
      });
      await wrapper.get('[aid="advanced-settings"]').trigger("click");
      await wrapper.get('[aid="radio-option-Custom"]').trigger("click");
      expect(labels()).toEqual(["Unknown", "Unknown"]);
      expect(pages.loadAll).toHaveBeenCalledOnce();
      await wrapper.setProps({ rows: peers, baselineRows: peers, departmentPages: { ...pages, nextOffset: undefined } });
      expect(labels()).toEqual(["33.28%", "0.17%", "66.56%"]);
      expect(weights()[0].props("value")).toBe(2);
      expect(weights()[0].props("previewWeightPercentage")?.(4)).toBeCloseTo(49.94);
      weights()[0].vm.$emit("apply", 4);
      await wrapper.vm.$nextTick();
      expect(labels()).toEqual(["49.94%", "0.12%", "49.94%"]);
      await wrapper.get('[aid="radio-option-Automatic"]').trigger("click");
      expect(labels()).toEqual(["0%", "0.2%", "99.8%"]);
      await wrapper.get('[aid="radio-option-Custom"]').trigger("click");
      expect(labels()).toEqual(["33.28%", "0.17%", "66.56%"]);
      expect(table.props("draft").departments["0"]?.weight).toBeUndefined();
    } finally {
      wrapper.unmount();
      vi.unstubAllGlobals();
    }
  });

  it("seeds pending Automatic quotas as weights and preserves subsequent manual edits", async () => {
    const wrapper = render({
      ...pool,
      schedulingCapabilities: { quota: true, ranks: true, adjustableWeights: false, cpuQuota: false, limits: false },
    });
    try {
      await wrapper.setProps({
        canSavePool: true,
        canResetResources: true,
        canViewDepartments: true,
        canEditDepartments: true,
      });
      const table = wrapper.findComponent(Table);
      table.vm.$emit("edit", { id: "a", quota: 24.5 });
      await wrapper.get('[aid="advanced-settings"]').trigger("click");
      wrapper.findComponent(Settings).vm.$emit("patch", { automaticWeights: false });
      await wrapper.vm.$nextTick();
      expect(table.props("draft").departments.a).toMatchObject({ weight: 24.5, weightFromQuota: true });
      expect(wrapper.get('[aid="scheduling-save-close"]').attributes("disabled")).toBeUndefined();
      table.vm.$emit("edit", { id: "a", weight: 6 });
      await wrapper.vm.$nextTick();
      expect(table.props("draft").departments.a).toEqual({ id: "a", quota: 24.5, weight: 6 });
    } finally {
      wrapper.unmount();
    }
  });
  it("requires cluster-wide permission to initialize custom weights in Guaranteed mode", async () => {
    const wrapper = render({
      ...pool,
      schedulingCapabilities: { quota: true, ranks: true, adjustableWeights: false, cpuQuota: false, limits: false },
    });
    try {
      await wrapper.setProps({ canSavePool: true, canResetResources: false });
      await wrapper.get('[aid="advanced-settings"]').trigger("click");
      wrapper.findComponent(Settings).vm.$emit("patch", { automaticWeights: false });
      await wrapper.vm.$nextTick();
      expect(wrapper.findComponent(NodePoolWizardSummary).props("pool").schedulingCapabilities?.adjustableWeights).toBe(
        true,
      );
      expect(wrapper.get('[aid="scheduling-save-close"]').attributes("disabled")).toBeDefined();
      expect(wrapper.get('[role="alert"]').text()).toContain("Switching weights to Custom requires update permission");
      await wrapper.setProps({ canResetResources: true });
      expect(wrapper.get('[aid="scheduling-save-close"]').attributes("disabled")).toBeUndefined();
    } finally {
      wrapper.unmount();
    }
  });

  it("identifies unsaved pool settings when a changed retry fails after partial success", async () => {
    const wrapper = render();
    await wrapper.setProps({ canSavePool: true });
    saveMocks.patchPool.mockReset().mockImplementation(async (_id, payload) => ({ ...pool, ...payload }));
    saveMocks.patchDepartments
      .mockReset()
      .mockResolvedValue({ succeeded: [], totalFailed: 1, failed: [{ message: "department conflict" }] });
    wrapper.findComponent(Table).vm.$emit("edit", { id: "a", weight: 5 });
    await wrapper.get('[aid="scheduling-save-close"]').trigger("click");
    await flushPromises();
    expect(wrapper.get('[role="alert"]').text()).toContain("Pool settings were saved");
    await wrapper.get('[aid="mode-ranked"]').trigger("click");
    saveMocks.patchPool.mockRejectedValueOnce(new Error("latest pool update failed"));
    await wrapper.get('[aid="scheduling-save-close"]').trigger("click");
    await flushPromises();
    expect(wrapper.get('[role="alert"]').text()).toContain("Current pool settings were not saved");
    expect(wrapper.get('[aid="mode-ranked"]').attributes("aria-checked")).toBe("true");
    expect(wrapper.findComponent(Table).props("draft").departments.a.weight).toBe(5);
    wrapper.unmount();
  });
  it("expands the whole mode group on hover and dismisses descriptions with Escape without selecting", async () => {
    const wrapper = render();
    const cards = wrapper.get('[aid="scheduling-mode-options"]');
    await wrapper.get('[aid="scheduling-mode-hover-target"]').trigger("mouseenter");
    expect(wrapper.findAll(".mode-explanation")).toHaveLength(3);
    expect(wrapper.get('[aid="mode-proportional"]').attributes("aria-checked")).toBe("true");
    await cards.trigger("keydown", { key: "Escape" });
    expect(wrapper.findAll(".mode-explanation").filter((item) => item.isVisible())).toHaveLength(0);
    await wrapper.get('[aid="mode-proportional"] .runai-segment__label').trigger("mouseover");
    expect(wrapper.findAll(".mode-explanation").filter((item) => item.isVisible())).toHaveLength(0);
    expect(wrapper.get('[aid="mode-proportional"]').attributes("aria-checked")).toBe("true");
    wrapper.unmount();
  });
  it("warns before committing Guaranteed quota when hovered or selected", async () => {
    const wrapper = render();
    const cards = wrapper.get('[aid="scheduling-mode-options"]');
    await wrapper.get('[aid="mode-guaranteed"]').trigger("mouseover");
    expect(wrapper.find(".mode-guarantee-warning").text()).toBe("Mode can't be changed after saving");
    await wrapper.get('[aid="mode-guaranteed"]').trigger("click");
    await cards.trigger("mouseleave");
    expect(wrapper.find(".mode-guarantee-warning").text()).toBe("Mode can't be changed after saving");
    wrapper.unmount();
  });
  it("keeps descriptions open through a brief pointer exit, then closes after a sustained exit", async () => {
    vi.useFakeTimers();
    const wrapper = render();
    try {
      const target = wrapper.get('[aid="scheduling-mode-hover-target"]');
      const expanded = () => wrapper.get('[aid="scheduling-mode-options"]').classes("mode-cards--expanded");
      await target.trigger("mouseenter");
      expect(expanded()).toBe(true);
      await target.trigger("mouseleave");
      await vi.advanceTimersByTimeAsync(80);
      expect(expanded()).toBe(true);
      await target.trigger("mouseenter");
      await vi.advanceTimersByTimeAsync(700);
      expect(expanded()).toBe(true);
      await target.trigger("mouseleave");
      await vi.advanceTimersByTimeAsync(300);
      expect(expanded()).toBe(false);
      expect(wrapper.get('[aid="mode-proportional"]').attributes("aria-checked")).toBe("true");
    } finally {
      wrapper.unmount();
      vi.useRealTimers();
    }
  });
  it("dismisses immediately with Escape during the pointer-exit grace period and allows a fresh hover", async () => {
    vi.useFakeTimers();
    const wrapper = render();
    try {
      const target = wrapper.get('[aid="scheduling-mode-hover-target"]');
      const expanded = () => wrapper.get('[aid="scheduling-mode-options"]').classes("mode-cards--expanded");
      await target.trigger("mouseenter");
      await target.trigger("mouseleave");
      await wrapper.get('[aid="scheduling-mode-options"]').trigger("keydown", { key: "Escape" });
      expect(expanded()).toBe(false);
      await target.trigger("mouseenter");
      await vi.advanceTimersByTimeAsync(700);
      expect(expanded()).toBe(true);
    } finally {
      wrapper.unmount();
      vi.useRealTimers();
    }
  });
  it("defaults a first Guaranteed selection to Automatic and retains an explicit Custom choice", async () => {
    const wrapper = render();
    try {
      await wrapper.get('[aid="mode-guaranteed"]').trigger("click");
      expect(wrapper.findComponent(Table).props("draft").automaticWeights).toBe(true);

      await wrapper.get('[aid="advanced-settings"]').trigger("click");
      wrapper.findComponent(Settings).vm.$emit("patch", { automaticWeights: false });
      await wrapper.get('[aid="mode-ranked"]').trigger("click");
      await wrapper.get('[aid="mode-guaranteed"]').trigger("click");
      expect(wrapper.findComponent(Table).props("draft").automaticWeights).toBe(false);
    } finally {
      wrapper.unmount();
    }
  });

  it("uses the API fairshare defaults when enabling previously unconfigured history", async () => {
    saveMocks.patchPool.mockReset().mockImplementation(async (_id, payload) => ({ ...pool, ...payload }));
    saveMocks.patchDepartments.mockReset().mockResolvedValue({ succeeded: [], failed: [], totalFailed: 0 });
    const wrapper = render({ ...pool, schedulingConfiguration: { timeBasedFairShare: { enabled: false } } });
    try {
      await wrapper.setProps({ canSavePool: true });
      await wrapper.get('[aid="advanced-settings"]').trigger("click");
      const settings = wrapper.findComponent(Settings);
      expect(settings.props("draft").historicalWeight).toBe(1);
      expect(settings.props("draft").historicalWindow).toEqual({ Days: 7, Hours: 0, Minutes: 0, Seconds: 0 });
      settings.vm.$emit("patch", { timeFairshare: true });
      await wrapper.get('[aid="scheduling-save-close"]').trigger("click");
      await flushPromises();
      expect(saveMocks.patchPool.mock.calls.at(-1)?.[1].schedulingConfiguration.timeBasedFairShare).toMatchObject({
        enabled: true,
        historicalUsageWeight: 1,
        windowDuration: "7d",
      });
    } finally {
      wrapper.unmount();
    }
  });

  it("preserves explicitly saved zero historical weight and zero window", () => {
    const wrapper = render({
      ...pool,
      schedulingConfiguration: { timeBasedFairShare: { enabled: true, historicalUsageWeight: 0, windowDuration: "0" } },
    });
    const draft = wrapper.findComponent(Table).props("draft");
    expect(draft.historicalWeight).toBe(0);
    expect(draft.historicalWindow).toEqual({ Days: 0, Hours: 0, Minutes: 0, Seconds: 0 });
    wrapper.unmount();
  });

  it("loads fractional windows and weekly runtimes into the visible duration fields", () => {
    const wrapper = render({
      ...pool,
      schedulingConfiguration: {
        minGuaranteedRuntime: "1w",
        timeBasedFairShare: { enabled: true, windowDuration: "0.5m" },
      },
    });
    const draft = wrapper.findComponent(Table).props("draft");
    expect(draft.minimumRuntime).toEqual({ Days: 7, Hours: 0, Minutes: 0, Seconds: 0 });
    expect(draft.historicalWindow).toEqual({ Days: 0, Hours: 0, Minutes: 0.5, Seconds: 0 });
    wrapper.unmount();
  });
  it.each([
    { maxAllocation: true },
    { timeFairshare: true },
    { historicalWeight: 2 },
    { historicalWindow: { Days: 2 } },
    { minimumRuntime: { Minutes: 5 } },
  ])("warns before leaving when only advanced settings changed: %j", async (patch) => {
    const wrapper = render();
    await flushPromises();
    await wrapper.get('[aid="advanced-settings"]').trigger("click");
    wrapper.findComponent(Settings).vm.$emit("patch", patch);
    await flushPromises();
    const leave = requestToLeave();
    await flushPromises();
    expect(document.querySelector('[aid="leave-page-modal-stay-btn"]')).not.toBeNull();
    (document.querySelector('[aid="leave-page-modal-stay-btn"]') as HTMLButtonElement).click();
    expect(await leave).toBe(false);
    wrapper.unmount();
  });
  it("renders the department controls in the live document, outside native template elements", () => {
    const wrapper = render();
    expect(wrapper.findComponent(RunaiWizardLayout).props("responsive")).toBe(true);
    expect(wrapper.find("template").exists()).toBe(false);
    expect(wrapper.get('[aid="new-department"]').element.closest("template")).toBeNull();
    wrapper.unmount();
  });
  it("previews every mode in the summary without mutating the saved pool or topology", async () => {
    const wrapper = render();
    const before = JSON.stringify(pool);
    const summary = wrapper.findComponent(NodePoolWizardSummary);
    expect(wrapper.get('[aid="mode-proportional"]').attributes("aria-checked")).toBe("true");
    await wrapper.get('[aid="mode-ranked"]').trigger("click");
    expect(summary.text()).toContain("Ranked");
    expect(summary.props("pool").schedulingCapabilities).toEqual({
      adjustableWeights: true,
      quota: false,
      ranks: true,
      cpuQuota: false,
      limits: false,
    });
    expect(summary.text()).toContain("Existing fabric");
    expect(wrapper.findAll(".mode-check")).toHaveLength(1);
    await wrapper.get('[aid="mode-guaranteed"]').trigger("click");
    expect(summary.text()).toContain("Guaranteed quota");
    expect(summary.props("pool").schedulingCapabilities).toEqual({
      adjustableWeights: false,
      quota: true,
      ranks: true,
      cpuQuota: false,
      limits: false,
    });
    await wrapper.get('[aid="mode-proportional"]').trigger("click");
    expect(summary.text()).toContain("Proportional");
    expect(JSON.stringify(pool)).toBe(before);
    wrapper.unmount();
  });

  it.each([
    { limits: false, quota: false, ranks: false, adjustableWeights: false, cpuQuota: false },
    { limits: false, quota: false, ranks: true, adjustableWeights: false, cpuQuota: false },
    { limits: false, quota: true, ranks: false, adjustableWeights: true, cpuQuota: false },
  ])("uses the shared Step 1 classification of saved capabilities: %j", (schedulingCapabilities) => {
    const wrapper = render({
      ...pool,
      schedulingCapabilities,
    });
    const expected = schedulingCapabilities.quota
      ? "Guaranteed quota"
      : schedulingCapabilities.ranks
        ? "Ranked"
        : "Custom";
    expect(wrapper.findAll('[role="radio"][aria-checked="true"]')).toHaveLength(expected === "Custom" ? 0 : 1);
    expect(wrapper.findComponent(NodePoolWizardSummary).text()).toContain(expected);
    wrapper.unmount();
  });

  it("retains saved custom weights and CPU quota until another mode is selected", () => {
    const capabilities = { limits: false, quota: true, ranks: true, adjustableWeights: true, cpuQuota: true };
    const wrapper = render({ ...pool, schedulingCapabilities: capabilities });
    expect(wrapper.get('[aid="mode-guaranteed"]').attributes("aria-checked")).toBe("true");
    expect(wrapper.findComponent(NodePoolWizardSummary).props("pool").schedulingCapabilities).toEqual(capabilities);
    wrapper.unmount();
  });

  it("distinguishes an empty pool, an unknown capacity, and a measured zero with nodes", async () => {
    const wrapper = render();
    expect(wrapper.get('[aid="pool-gpu-capacity"]').text()).toBe("Unknown");
    await wrapper.setProps({ resources: { nodes: 2 } });
    expect(wrapper.get('[aid="pool-gpu-capacity"]').text()).toBe("Unknown");
    await wrapper.setProps({ resources: { nodes: 0, gpus: 0 } });
    expect(wrapper.get('[aid="pool-gpu-capacity"]').text()).toBe("Empty");
    expect(wrapper.getComponent(Table).props("capacity")).toBe(0);
    await wrapper.setProps({ resources: { nodes: 2, gpus: 0 } });
    expect(wrapper.get('[aid="pool-gpu-capacity"]').text()).toBe("0 GPUs");
    wrapper.unmount();
  });

  it("saves selected mode and navigates only after successful persistence", async () => {
    const wrapper = render();
    await wrapper.setProps({ canSavePool: true });
    await wrapper.get('[aid="mode-ranked"]').trigger("click");
    saveMocks.patchPool.mockImplementation(async (_id, payload) => ({ ...pool, ...payload }));
    await wrapper.get('[aid="scheduling-save-close"]').trigger("click");
    await flushPromises();
    expect(saveMocks.patchPool.mock.calls.at(-1)?.[1].schedulingCapabilities.ranks).toBe(true);
    expect(wrapper.emitted("saved")).toHaveLength(1);
    wrapper.unmount();
  });
  it("supports keyboard mode selection and keeps save actions disabled without pool permission", async () => {
    const wrapper = render();
    await wrapper.get('[aid="mode-proportional"]').trigger("keydown", { key: "ArrowRight" });
    expect(wrapper.get('[aid="mode-ranked"]').attributes("aria-checked")).toBe("true");
    expect(wrapper.get('[aid="scheduling-save-close"]').attributes("disabled")).toBeDefined();
    expect(wrapper.get('[aid="scheduling-save-continue"]').attributes("disabled")).toBeDefined();
    expect(wrapper.text()).not.toContain("Topology & placement is not available yet.");
    expect(wrapper.find(".preview-notice").exists()).toBe(false);
    expect(
      wrapper
        .findAll(".mode-icon")
        .map((icon) => icon.attributes("class"))
        .join(" "),
    ).toContain("ph-chart-pie-slice");
    await flushPromises();
    wrapper.unmount();
  });
  it("continues only after the pool and department changes finish saving", async () => {
    const wrapper = render();
    await wrapper.setProps({ canSavePool: true });
    let finishDepartments!: (value: unknown) => void;
    saveMocks.patchPool.mockReset().mockImplementation(async (_id, payload) => ({ ...pool, ...payload }));
    saveMocks.patchDepartments.mockReset().mockImplementation(
      () =>
        new Promise((resolve) => {
          finishDepartments = resolve;
        }),
    );
    wrapper.findComponent(Table).vm.$emit("edit", { id: "a", weight: 5 });
    await wrapper.get('[aid="scheduling-save-continue"]').trigger("click");
    await flushPromises();
    expect(wrapper.emitted("saved")).toBeUndefined();
    expect(wrapper.get('[aid="scheduling-save-continue"]').attributes("disabled")).toBeDefined();
    await wrapper.get('[aid="wizard-step-item-topology"]').trigger("click");
    expect(wrapper.emitted("open-topology")).toBeUndefined();
    finishDepartments({ succeeded: [{ orgUnitId: "a", nodePoolId: "pool-1" }], failed: [], totalFailed: 0 });
    await flushPromises();
    expect(wrapper.emitted("saved")?.[0]?.[1]).toBe("continue");
    wrapper.unmount();
  });

  it("stays on scheduling when the pool save fails before continuing", async () => {
    const wrapper = render();
    await wrapper.setProps({ canSavePool: true });
    saveMocks.patchPool.mockReset().mockRejectedValue(new Error("pool update failed"));
    saveMocks.patchDepartments.mockReset();
    await wrapper.get('[aid="mode-ranked"]').trigger("click");
    await wrapper.get('[aid="scheduling-save-continue"]').trigger("click");
    await flushPromises();
    expect(wrapper.emitted("saved")).toBeUndefined();
    expect(wrapper.emitted("pool-saved")).toBeUndefined();
    expect(saveMocks.patchDepartments).not.toHaveBeenCalled();
    expect(wrapper.get('[role="alert"]').text()).toContain("pool update failed");
    expect(wrapper.get('[aid="mode-ranked"]').attributes("aria-checked")).toBe("true");
    wrapper.unmount();
  });

  it("keeps the scheduling draft after partial failure and continues only when retry completes", async () => {
    const wrapper = render();
    await wrapper.setProps({ canSavePool: true });
    saveMocks.patchPool.mockReset().mockImplementation(async (_id, payload) => ({ ...pool, ...payload }));
    saveMocks.patchDepartments
      .mockReset()
      .mockResolvedValueOnce({ succeeded: [], failed: [{ message: "department conflict" }], totalFailed: 1 })
      .mockResolvedValueOnce({ succeeded: [{ orgUnitId: "a", nodePoolId: "pool-1" }], failed: [], totalFailed: 0 });
    wrapper.findComponent(Table).vm.$emit("edit", { id: "a", weight: 5 });
    await wrapper.get('[aid="scheduling-save-continue"]').trigger("click");
    await flushPromises();
    expect(wrapper.emitted("saved")).toBeUndefined();
    expect(wrapper.get('[role="alert"]').text()).toContain("remain unsaved");
    expect(wrapper.findComponent(Table).props("draft").departments.a.weight).toBe(5);
    await wrapper.get('[aid="scheduling-save-continue"]').trigger("click");
    await flushPromises();
    expect(wrapper.emitted("saved")?.[0]?.[1]).toBe("continue");
    expect(saveMocks.patchPool).toHaveBeenCalledTimes(1);
    wrapper.unmount();
  });

  it.each(["canEditDepartments", "canViewDepartments"] as const)(
    "lets a pool editor continue without writing when %s is missing",
    async (permission) => {
      const wrapper = render();
      await wrapper.setProps({ canSavePool: true, [permission]: false });
      saveMocks.patchPool.mockClear();
      saveMocks.patchDepartments.mockClear();
      const next = wrapper.get('[aid="scheduling-continue"]');
      expect(next.text()).toBe("Continue");
      await next.trigger("click");
      expect(wrapper.emitted("open-topology")).toHaveLength(1);
      expect(saveMocks.patchPool).not.toHaveBeenCalled();
      expect(saveMocks.patchDepartments).not.toHaveBeenCalled();
      wrapper.unmount();
    },
  );

  it("navigates through the stepper without saving a dirty or invalid draft", async () => {
    const wrapper = render();
    await wrapper.setProps({ canSavePool: true });
    saveMocks.patchPool.mockClear();
    saveMocks.patchDepartments.mockClear();
    wrapper.findComponent(Table).vm.$emit("edit", { id: "a", weight: 1.5 });
    await wrapper.vm.$nextTick();
    expect(wrapper.get('[aid="scheduling-save-continue"]').attributes("disabled")).toBeDefined();
    await wrapper.get('[aid="wizard-step-item-identity"]').trigger("click");
    expect(wrapper.emitted("open-identity")).toHaveLength(1);
    expect(wrapper.emitted("open-topology")).toBeUndefined();
    await wrapper.get('[aid="wizard-step-item-topology"]').trigger("click");
    expect(wrapper.emitted("open-topology")).toHaveLength(1);
    expect(saveMocks.patchPool).not.toHaveBeenCalled();
    expect(saveMocks.patchDepartments).not.toHaveBeenCalled();
    wrapper.unmount();
  });

  it("preserves applied edits when changing mode without a discard dialog", async () => {
    const wrapper = render();
    const tableUid = wrapper.findComponent(Table).vm.$.uid;
    wrapper.findComponent(Table).vm.$emit("edit", { id: "a", weight: 5 });
    await wrapper.vm.$nextTick();
    await wrapper.get('[aid="mode-ranked"]').trigger("click");
    expect(wrapper.findComponent(RunaiBaseDialog).exists()).toBe(false);
    expect(wrapper.findComponent(Table).props("draft").departments.a.weight).toBe(5);
    expect(wrapper.get('[aid="mode-ranked"]').attributes("aria-checked")).toBe("true");
    for (const aid of ["mode-guaranteed", "mode-proportional", "mode-ranked"]) {
      await wrapper.get(`[aid="${aid}"]`).trigger("click");
      expect(wrapper.findComponent(Table).vm.$.uid).toBe(tableUid);
      expect(wrapper.findComponent(Table).props("draft").departments.a.weight).toBe(5);
    }
    wrapper.unmount();
  });
  it("shows permission denial instead of an empty department table", async () => {
    const wrapper = render();
    await wrapper.setProps({ canViewDepartments: false });
    expect(wrapper.text()).toContain("permissions to view departments");
    expect(wrapper.findComponent(Table).props("canView")).toBe(false);
    wrapper.unmount();
  });
  it("clears the server rank filter only when Proportional actually takes effect", async () => {
    const wrapper = render({
      ...pool,
      schedulingCapabilities: { limits: false, adjustableWeights: true, quota: false, ranks: true, cpuQuota: false },
    });
    await wrapper.setProps({ rankFilter: "High" });
    wrapper.findComponent(Table).vm.$emit("edit", { id: "a", weight: 5 });
    await wrapper.vm.$nextTick();
    await wrapper.get('[aid="mode-proportional"]').trigger("click");
    await flushPromises();
    expect(wrapper.emitted("filter-rank")).toEqual([[undefined]]);
    wrapper.unmount();
  });
});
