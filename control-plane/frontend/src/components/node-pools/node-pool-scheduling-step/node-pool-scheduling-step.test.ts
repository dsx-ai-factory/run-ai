/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, it, vi } from "vitest";
import { createTestingPinia } from "@pinia/testing";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import Settings from "./node-pool-scheduling-settings.vue";
import Table from "./node-pool-department-table.vue";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiSegmentedButton } from "@/components/common/runai-segmented-button";
import NodePoolSchedulingStep from "./node-pool-scheduling-step.vue";
import NewDepartment from "./node-pool-new-department-dialog.vue";
import RunaiWizardLayout from "@/components/common/runai-wizard-layout/runai-wizard-layout.vue";
import NodePoolWizardSummary from "../node-pool-wizard-summary/node-pool-wizard-summary.vue";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
const { createNodePoolDraft } = nodePoolUtil;
import type { Nodepool } from "@/swagger-models/cluster-service-client";

const saveMocks = vi.hoisted(() => ({ patchPool: vi.fn(), patchDepartments: vi.fn() }));
vi.mock("./node-pool-scheduling.service", async (original) => ({
  ...(await original<object>()),
  patchSchedulingPool: saveMocks.patchPool,
  patchSchedulingDepartments: saveMocks.patchDepartments,
}));

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
        saved.quota ? "An existing Guaranteed quota pool cannot switch to a lower scheduling mode." : undefined,
      );
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

  it("distinguishes unknown capacity from a measured zero", async () => {
    const wrapper = render();
    expect(wrapper.get('[aid="pool-gpu-capacity"]').text()).toBe("Unknown");
    await wrapper.setProps({ resources: { gpus: 0 } });
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

  it("saves before entering topology from the step panel and keeps identity unavailable", async () => {
    const wrapper = render();
    await wrapper.setProps({ canSavePool: true });
    saveMocks.patchPool.mockReset().mockImplementation(async (_id, payload) => ({ ...pool, ...payload }));
    await wrapper.get('[aid="wizard-step-item-identity"]').trigger("click");
    expect(saveMocks.patchPool).not.toHaveBeenCalled();
    await wrapper.get('[aid="wizard-step-item-topology"]').trigger("click");
    await flushPromises();
    expect(saveMocks.patchPool).toHaveBeenCalledTimes(1);
    expect(wrapper.emitted("saved")?.[0]?.[1]).toBe("continue");
    wrapper.unmount();
  });

  it("prevents continuing or using the topology step when scheduling validation fails", async () => {
    const wrapper = render();
    await wrapper.setProps({ canSavePool: true });
    saveMocks.patchPool.mockReset();
    wrapper.findComponent(Table).vm.$emit("edit", { id: "a", weight: 1.5 });
    await wrapper.vm.$nextTick();
    expect(wrapper.get('[aid="scheduling-save-continue"]').attributes("disabled")).toBeDefined();
    await wrapper.get('[aid="wizard-step-item-topology"]').trigger("click");
    expect(saveMocks.patchPool).not.toHaveBeenCalled();
    expect(wrapper.emitted("saved")).toBeUndefined();
    wrapper.unmount();
  });

  it("preserves applied edits when changing mode without a discard dialog", async () => {
    const wrapper = render();
    wrapper.findComponent(Table).vm.$emit("edit", { id: "a", weight: 5 });
    await wrapper.vm.$nextTick();
    await wrapper.get('[aid="mode-ranked"]').trigger("click");
    expect(wrapper.findComponent(RunaiBaseDialog).exists()).toBe(false);
    expect(wrapper.findComponent(Table).props("draft").departments.a.weight).toBe(5);
    expect(wrapper.get('[aid="mode-ranked"]').attributes("aria-checked")).toBe("true");
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
