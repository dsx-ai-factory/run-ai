/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, it } from "vitest";
import Table from "./node-pool-department-table.vue";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import NodePoolSchedulingStep from "./node-pool-scheduling-step.vue";
import NodePoolWizardSummary from "../node-pool-wizard-summary/node-pool-wizard-summary.vue";
import { createNodePoolDraft } from "../node-pool-create-flow/node-pool-create-flow.model";
import type { Nodepool } from "@/swagger-models/cluster-service-client";

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
    global: { plugins: [Quasar], directives: { "click-outside": {} } },
  });
}

describe("scheduling step foundation", () => {
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
    });
    await wrapper.get('[aid="mode-proportional"]').trigger("click");
    expect(summary.text()).toContain("Proportional");
    expect(JSON.stringify(pool)).toBe(before);
    wrapper.unmount();
  });

  it.each([
    { quota: false, ranks: false, adjustableWeights: false, cpuQuota: false },
    { quota: false, ranks: true, adjustableWeights: false, cpuQuota: false },
    { quota: true, ranks: false, adjustableWeights: true, cpuQuota: false },
  ])("does not pretend a custom saved configuration is a supported preset: %j", (schedulingCapabilities) => {
    const wrapper = render({
      ...pool,
      schedulingCapabilities,
    });
    expect(wrapper.findAll('[role="radio"][aria-checked="true"]')).toHaveLength(0);
    expect(wrapper.findComponent(NodePoolWizardSummary).text()).toContain("Custom");
    wrapper.unmount();
  });

  it("retains saved custom weights and CPU quota until another mode is selected", () => {
    const capabilities = { quota: true, ranks: true, adjustableWeights: true, cpuQuota: true };
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

  it("supports keyboard mode selection and keeps unfinished save actions disabled", async () => {
    const wrapper = render();
    await wrapper.get('[aid="mode-proportional"]').trigger("keydown", { key: "ArrowRight" });
    expect(wrapper.get('[aid="mode-ranked"]').attributes("aria-checked")).toBe("true");
    expect(wrapper.get('[aid="scheduling-save-close"]').attributes("disabled")).toBeDefined();
    expect(wrapper.get('[aid="scheduling-save-continue"]').attributes("disabled")).toBeDefined();
    expect(wrapper.text()).toContain("Changes are not saved in this preview.");
    await flushPromises();
    wrapper.unmount();
  });
  it("keeps applied edits on Stay and discards them on Change mode", async () => {
    const wrapper = render();
    wrapper.findComponent(Table).vm.$emit("edit", { id: "a", weight: 5 });
    await wrapper.vm.$nextTick();
    await wrapper.get('[aid="mode-ranked"]').trigger("click");
    expect(wrapper.findComponent(RunaiBaseDialog).exists()).toBe(true);
    expect(wrapper.findComponent(Table).props("draft").departments.a.weight).toBe(5);
    wrapper.findComponent(RunaiBaseDialog).vm.$emit("close");
    await wrapper.vm.$nextTick();
    expect(wrapper.findComponent(Table).props("draft").departments.a.weight).toBe(5);
    await wrapper.get('[aid="mode-ranked"]').trigger("click");
    (document.querySelector('[aid="confirm-scheduling-mode"]') as HTMLButtonElement).click();
    await flushPromises();
    expect(wrapper.findComponent(Table).props("draft").departments).toEqual({});
    expect(wrapper.get('[aid="mode-ranked"]').attributes("aria-checked")).toBe("true");
    wrapper.unmount();
  });
  it("shows permission denial instead of an empty department table", async () => {
    const wrapper = render();
    await wrapper.setProps({ canViewDepartments: false });
    expect(wrapper.text()).toContain("permission to view departments");
    expect(wrapper.findComponent(Table).exists()).toBe(false);
    wrapper.unmount();
  });
});
