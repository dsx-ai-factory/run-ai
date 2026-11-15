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
import NodePoolSchedulingStep from "./node-pool-scheduling-step.vue";
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

function render(input = pool) {
  return mount(NodePoolSchedulingStep, {
    props: { pool: input },
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      directives: { "click-outside": {}, permission: {} },
    },
  });
}

describe("scheduling step foundation", () => {
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
      adjustableWeights: true,
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
  it("supports keyboard mode selection and keeps unfinished save actions disabled", async () => {
    const wrapper = render();
    await wrapper.get('[aid="mode-proportional"]').trigger("keydown", { key: "ArrowRight" });
    expect(wrapper.get('[aid="mode-ranked"]').attributes("aria-checked")).toBe("true");
    expect(wrapper.get('[aid="scheduling-save-close"]').attributes("disabled")).toBeDefined();
    expect(wrapper.get('[aid="scheduling-save-continue"]').attributes("disabled")).toBeDefined();
    expect(wrapper.text()).toContain("Topology & placement is not available yet.");
    await flushPromises();
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
