/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { describe, it, expect, vi } from "vitest";
import { defineComponent, h, nextTick } from "vue";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import CpuQuotaStep from "./cpu-quota-step.vue";
import { CpuQuotaInput } from "@/components/org-unit/v2/common/cpu-quota-input";
import {
  mockProject,
  mockDepartment,
  mockDepartmentResources,
  mockDepartmentNodePoolQuotaStatuses,
  mockNodePoolTotals,
  mockSubordinateResources,
} from "./cpu-quota-step.mock";
import { EQuotaInputState } from "@/models/org-unit.model";
import { provideUnsavedChangesGuard } from "@/composables/org-unit/use-unsaved-changes-guard.composable";
import { EQuotaEntity } from "@/models/resource.model";
import type { QuotaStatusNodePool, Resources, AggregatedResources } from "@/swagger-models/org-unit-service-client";

describe("CpuQuotaStep", () => {
  const createWrapper = (props = {}, isOverSubscriptionEnabled = false) => {
    const settingsMap = new Map();

    const stepProps = {
      entityType: EQuotaEntity.project,
      resources: mockProject.resources as Resources[],
      defaultNodePools: mockProject.defaultNodePools,
      parentResources: mockDepartment.resources as Resources[],
      parentDefaultNodePools: undefined as string[] | undefined,
      nodePoolQuotaStatuses: mockProject.status.nodePoolQuotaStatuses as QuotaStatusNodePool[],
      ...props,
    };

    const TestHost = defineComponent({
      components: { CpuQuotaStep },
      emits: ["save-cpu-quota", "close"],
      setup(_, { emit }) {
        provideUnsavedChangesGuard();
        return () =>
          h(CpuQuotaStep, {
            ...stepProps,
            onSaveCpuQuota: (payload: unknown) => emit("save-cpu-quota", payload),
            onClose: () => emit("close"),
          });
      },
    });

    return mount(TestHost, {
      global: {
        plugins: [
          Quasar,
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              Setting: {
                settings: settingsMap,
                limitQuotaOverSubscription: { value: isOverSubscriptionEnabled },
              },
            },
          }),
        ],
      },
    });
  };

  it("renders correctly", () => {
    const wrapper = createWrapper();
    expect(wrapper.findComponent(CpuQuotaInput).exists()).toBe(true);
  });

  it("initializes cpu quota from resources", async () => {
    const wrapper = createWrapper();
    await nextTick();
    // 1000m = 1 core
    expect(wrapper.findComponent(CpuQuotaInput).props("modelValue")).toBe(1);
  });

  it("emits save event with correct payload when save is clicked", async () => {
    // Use multiple node pools to ensure save button is rendered
    const multiNodePoolResources = [
      {
        nodePool: { name: "pool1", id: "1" },
        cpu: { deserved: 1000 },
      },
      {
        nodePool: { name: "pool2", id: "2" },
        cpu: { deserved: 5000 },
      },
    ] as Resources[];

    const wrapper = createWrapper({
      resources: multiNodePoolResources,
      defaultNodePools: ["pool1", "pool2"],
    });
    await nextTick();

    // Emit save event from org-unit-step-wrapper
    const stepWrapper = wrapper.findComponent({ name: "OrgUnitStepWrapper" });
    stepWrapper.vm.$emit("save");
    await nextTick();

    expect(wrapper.emitted("save-cpu-quota")).toBeTruthy();
    expect(wrapper.emitted("save-cpu-quota")?.[0]).toEqual([
      {
        nodePoolName: "pool1",
        cpuQuota: 1000,
        memoryQuota: null,
        memoryUnit: "MB",
        navigateAfterSave: false,
      },
    ]);
  });

  it("updates cpu quota when node pool selection changes", async () => {
    const multiNodePoolResources = [
      {
        nodePool: { name: "pool1", id: "1" },
        cpu: { deserved: 1000 },
      },
      {
        nodePool: { name: "pool2", id: "2" },
        cpu: { deserved: 5000 },
      },
    ] as Resources[];

    const wrapper = createWrapper({
      resources: multiNodePoolResources,
      defaultNodePools: ["pool1", "pool2"],
    });
    await nextTick();

    // Initial check (pool1)
    expect(wrapper.findComponent(CpuQuotaInput).props("modelValue")).toBe(1);

    // Switch node pool via stubbed component event (using ID "2" for pool2)
    await wrapper.findComponent({ name: "NodePoolTabs" }).vm.$emit("update:model-value", "2");
    await nextTick();

    // Check new value (5000m = 5 cores)
    expect(wrapper.findComponent(CpuQuotaInput).props("modelValue")).toBe(5);
  });

  it("validates quota against parent limit (Warning)", async () => {
    const wrapper = createWrapper();
    await nextTick();

    // Parent limit is 2 cores (mockDepartment)
    // Set input to 3 cores
    wrapper.findComponent(CpuQuotaInput).vm.$emit("update:model-value", 3);
    await nextTick();

    expect(wrapper.findComponent(CpuQuotaInput).props("state")).toBe(EQuotaInputState.Warning);
  });

  it("validates quota against parent limit (Error when oversubscription disabled)", async () => {
    const wrapper = createWrapper({}, true);
    await nextTick();

    // Parent limit is 2 cores
    // Set input to 3 cores
    wrapper.findComponent(CpuQuotaInput).vm.$emit("update:model-value", 3);
    await nextTick();

    expect(wrapper.findComponent(CpuQuotaInput).props("state")).toBe(EQuotaInputState.Error);
  });

  it("shows warning when quota is less than allocated non-preemptible", async () => {
    const wrapper = createWrapper();
    await nextTick();

    // Allocated is 0.5 cores (500m)
    // Set input to 0.1 cores
    wrapper.findComponent(CpuQuotaInput).vm.$emit("update:model-value", 0.1);
    await nextTick();

    expect(wrapper.findComponent(CpuQuotaInput).props("state")).toBe(EQuotaInputState.Warning);
  });

  it("shows error when quota is less than allocated non-preemptible (Error when oversubscription disabled)", async () => {
    const wrapper = createWrapper({}, true);
    await nextTick();

    // Allocated is 0.5 cores
    // Set input to 0.1 cores
    wrapper.findComponent(CpuQuotaInput).vm.$emit("update:model-value", 0.1);
    await nextTick();

    expect(wrapper.findComponent(CpuQuotaInput).props("state")).toBe(EQuotaInputState.Error);
  });

  it("initializes quota to null when cpu deserved is missing", async () => {
    const missingCpuResources = [
      {
        nodePool: { name: "default", id: "1" },
        // no cpu property
      },
    ] as Resources[];

    const wrapper = createWrapper({
      resources: missingCpuResources,
      defaultNodePools: ["default"],
    });
    await nextTick();

    expect(wrapper.findComponent(CpuQuotaInput).props("modelValue")).toBeNull();
  });

  it("displays recommended value message when state is default", async () => {
    const wrapper = createWrapper();
    await nextTick();

    // allocatedNonPreemptible is 0.5 (from nodePoolQuotaStatuses)
    // parentLimit is 2 (parentResources)
    // Message uses <br> as separator between parts
    const expectedMessage =
      "Recommended: From 0.50 to 2.<br> Non-preemptive workload usage: 0.50 cores.<br> Department's quota: 2 cores";

    // Ensure state is default (input 1 is between 0.5 and 2)
    expect(wrapper.findComponent(CpuQuotaInput).props("state")).toBe(EQuotaInputState.Default);
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toBe(expectedMessage);
  });

  it("displays only available quota message when usage is 0 and limit exists", async () => {
    const nodePoolQuotaStatusesZeroUsage = [
      { nodePoolName: "default", allocatedNonPreemptible: { cpu: 0 } },
    ] as QuotaStatusNodePool[];

    const wrapper = createWrapper({
      nodePoolQuotaStatuses: nodePoolQuotaStatusesZeroUsage,
    });
    await nextTick();

    // allocatedNonPreemptible is 0
    // parentLimit is 2 (parentResources)
    // Message includes recommended value and parent quota
    const expectedMessage = "Recommended: Up to 2 cores.<br> Department's quota: 2 cores";

    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toBe(expectedMessage);
  });

  it("displays only usage message when usage > 0 and limit is null", async () => {
    const parentResourcesUnlimited = [
      {
        nodePool: { name: "default", id: "1" },
        cpu: { deserved: null }, // Unlimited
      },
    ] as Resources[];

    const wrapper = createWrapper({ parentResources: parentResourcesUnlimited });
    await nextTick();

    // allocatedNonPreemptible is 0.5 (from mockProject.status.nodePoolQuotaStatuses)
    const expectedMessage = "Non-preemptive workload usage: 0.50 cores";

    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toBe(expectedMessage);
  });

  it("displays nothing when usage is 0 and limit is null", async () => {
    const nodePoolQuotaStatusesZeroUsage = [
      { nodePoolName: "default", allocatedNonPreemptible: { cpu: 0 } },
    ] as QuotaStatusNodePool[];

    const parentResourcesUnlimited = [
      {
        nodePool: { name: "default", id: "1" },
        cpu: { deserved: null },
      },
    ] as Resources[];

    const wrapper = createWrapper({
      nodePoolQuotaStatuses: nodePoolQuotaStatusesZeroUsage,
      parentResources: parentResourcesUnlimited,
    });
    await nextTick();

    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toBe("");
  });
});

describe("CpuQuotaStep - Department Context", () => {
  const createDepartmentWrapper = (props = {}, isOverSubscriptionEnabled = false) => {
    const settingsMap = new Map();

    const stepProps = {
      entityType: EQuotaEntity.department,
      resources: mockDepartmentResources as Resources[],
      defaultNodePools: ["default"],
      nodePoolQuotaStatuses: mockDepartmentNodePoolQuotaStatuses as QuotaStatusNodePool[],
      nodePoolTotals: mockNodePoolTotals,
      subordinateResources: mockSubordinateResources as AggregatedResources[],
      ...props,
    };

    const TestHost = defineComponent({
      components: { CpuQuotaStep },
      emits: ["save-cpu-quota", "close"],
      setup(_, { emit }) {
        provideUnsavedChangesGuard();
        return () =>
          h(CpuQuotaStep, {
            ...stepProps,
            onSaveCpuQuota: (payload: unknown) => emit("save-cpu-quota", payload),
            onClose: () => emit("close"),
          });
      },
    });

    return mount(TestHost, {
      global: {
        plugins: [
          Quasar,
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              Setting: {
                settings: settingsMap,
                limitQuotaOverSubscription: { value: isOverSubscriptionEnabled },
              },
            },
          }),
        ],
      },
    });
  };

  it("renders correctly for department", () => {
    const wrapper = createDepartmentWrapper();
    expect(wrapper.findComponent(CpuQuotaInput).exists()).toBe(true);
  });

  it("uses nodePoolTotals for parentLimit validation", async () => {
    const wrapper = createDepartmentWrapper();
    await nextTick();

    // nodePoolTotals.default.cpuCores = 10
    // allocatedNonPreemptible = 1 core (1000 millicores)
    // Message should use "Node pool's total" label
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toContain("Node pool's total: 10 cores");
  });

  it("shows warning when department quota exceeds node pool total", async () => {
    const highQuotaResources = [
      {
        nodePool: { name: "default", id: "1" },
        cpu: { deserved: 15000 }, // 15 cores - exceeds node pool total of 10
        memory: { deserved: 1024, units: "MB" },
      },
    ] as Resources[];

    const wrapper = createDepartmentWrapper({
      resources: highQuotaResources,
    });
    await nextTick();

    expect(wrapper.findComponent(CpuQuotaInput).props("state")).toBe(EQuotaInputState.Warning);
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toContain(
      "The department's quota exceeds the node pool's total",
    );
  });

  it("shows warning when department quota is less than subordinate projects", async () => {
    const lowQuotaResources = [
      {
        nodePool: { name: "default", id: "1" },
        cpu: { deserved: 2000 }, // 2 cores - less than subordinate projects' 3 cores
        memory: { deserved: 1024, units: "MB" },
      },
    ] as Resources[];

    const wrapper = createDepartmentWrapper({
      resources: lowQuotaResources,
    });
    await nextTick();

    expect(wrapper.findComponent(CpuQuotaInput).props("state")).toBe(EQuotaInputState.Warning);
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toContain(
      "less than the quota assigned to its subordinate projects",
    );
  });

  it("shows warning when department quota is less than non-preemptible usage", async () => {
    const lowQuotaResources = [
      {
        nodePool: { name: "default", id: "1" },
        cpu: { deserved: 500 }, // 0.5 cores - less than allocated non-preemptible of 1 core
        memory: { deserved: 1024, units: "MB" },
      },
    ] as Resources[];

    const wrapper = createDepartmentWrapper({
      resources: lowQuotaResources,
    });
    await nextTick();

    expect(wrapper.findComponent(CpuQuotaInput).props("state")).toBe(EQuotaInputState.Warning);
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toContain(
      "less than the usage by non-preemptive workloads",
    );
  });

  it("handles empty nodePoolTotals gracefully - no exceeds total validation", async () => {
    const wrapper = createDepartmentWrapper({
      nodePoolTotals: {},
    });
    await nextTick();

    // Should not show "exceeds node pool's total" validation
    // Because parentLimit is null when nodePoolTotals is empty
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).not.toContain("exceeds the node pool's total");
  });

  it("handles null nodePoolTotals gracefully - no exceeds total validation", async () => {
    const wrapper = createDepartmentWrapper({
      nodePoolTotals: null,
    });
    await nextTick();

    // Should not show "exceeds node pool's total" validation
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).not.toContain("exceeds the node pool's total");
  });

  it("shows recommended text with node pool total for department", async () => {
    const wrapper = createDepartmentWrapper();
    await nextTick();

    // Should use the new instructive text format for departments
    // allocatedNonPreemptible = 1 core, nodePoolTotal = 10 cores
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toContain("Recommended: From 1 to 10");
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toContain("Non-preemptive workload usage: 1 core");
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toContain("Node pool's total: 10 cores");
  });

  it("shows only node pool total when no usage exists", async () => {
    const wrapper = createDepartmentWrapper({
      nodePoolQuotaStatuses: [{ nodePoolName: "default", allocatedNonPreemptible: { cpu: 0 } }],
    });
    await nextTick();

    // Message includes recommended value and node pool total
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toBe(
      "Recommended: Up to 10 cores.<br> Node pool's total: 10 cores",
    );
  });

  it("shows only usage when nodePoolTotals is empty", async () => {
    const wrapper = createDepartmentWrapper({
      nodePoolTotals: {},
      subordinateResources: [], // No subordinate resources either
    });
    await nextTick();

    // allocatedNonPreemptible = 1 core
    expect(wrapper.findComponent(CpuQuotaInput).props("message")).toBe("Non-preemptive workload usage: 1 core");
  });
});
