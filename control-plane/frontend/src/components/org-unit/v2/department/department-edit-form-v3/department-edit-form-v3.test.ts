/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flushPromises, mount } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { createTestingPinia } from "@pinia/testing";
import { Notify, Quasar } from "quasar";
import { ResourceType } from "@/swagger-models/authorization-client";
import { storageUtil } from "@/utils/storage.util";
import { usePermissionStore } from "@/stores/permissions.store";
import DepartmentEditFormV3 from "./department-edit-form-v3.vue";
import { ScopeNameStep } from "@/components/org-unit/v2/department/steps/scope-name-step";
import { SchedulingRulesStep } from "@/components/org-unit/v2/common/scheduling-rules-step";
import { AccessRulesStep } from "@/components/org-unit/v2/common/access-rules-step";
import { NodePoolsPriorityList } from "@/components/org-unit/v2/common/node-pools-priority-list";
import { UnsavedChangesDialog } from "@/components/org-unit/v2/common/unsaved-changes-dialog";
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { OrgUnitFormFooter } from "@/components/org-unit/v2/org-unit-form-footer";
import {
  EOrgUnitFormTooltip,
  EStepNames,
  ESaveOrgUnitAction,
  type ISaveOrgUnitPayload,
} from "@/models/org-unit-forms.model";
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";
import type { Department } from "@/swagger-models/org-unit-service-client";
import {
  mockCpuDisabledNodePool,
  mockCpuEnabledNodePool,
  mockDepartment,
  RunaiFormWrapperStub,
} from "@/components/org-unit/v2/department/department-edit-form-v2/department-edit-form-v2.mock";

const mockRouteQuery: { stepName?: string } = {};

vi.mock("vue-router", () => ({
  useRoute: () => ({ query: mockRouteQuery, params: {}, name: "department-new", meta: {} }),
  useRouter: () => ({ push: vi.fn(), back: vi.fn(), replace: vi.fn() }),
}));

vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: { isDepartmentNameUnique: vi.fn().mockResolvedValue(true) },
}));

vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: { getAccessRulesCount: vi.fn().mockResolvedValue(3) },
}));

const twoPoolDepartment = {
  ...mockDepartment,
  defaultNodePools: [mockCpuEnabledNodePool.name, mockCpuDisabledNodePool.name],
} as Department;

const mountForm = (props: Record<string, unknown> = {}, stubs: Record<string, unknown> = {}) => {
  const pinia = createTestingPinia({ createSpy: vi.fn });
  vi.mocked(usePermissionStore(pinia).hasPermission).mockReturnValue(true);

  return mount(DepartmentEditFormV3, {
    props: {
      department: twoPoolDepartment,
      isNewDepartment: true,
      nodePoolsMinimal: [mockCpuEnabledNodePool, mockCpuDisabledNodePool],
      ...props,
    },
    global: {
      plugins: [Quasar, pinia],
      mocks: { $route: { query: {} }, $router: { push: vi.fn() } },
      stubs: {
        RunaiFormWrapper: RunaiFormWrapperStub,
        SchedulingRulesStep: true,
        AccessRulesStep: true,
        ...stubs,
      },
    },
  });
};

const wizardSteps = (wrapper: ReturnType<typeof mountForm>) =>
  wrapper.findComponent({ name: "RunaiWizardLayout" }).props("steps") as IRunaiWizardLayoutStep[];

const mountAtStep = async (stepName: EStepNames, props: Record<string, unknown> = {}) => {
  mockRouteQuery.stepName = stepName;
  const wrapper = mountForm(props);
  await flushPromises();
  return wrapper;
};

const clickStep = async (wrapper: ReturnType<typeof mountForm>, stepName: EStepNames) => {
  wrapper
    .findComponent({ name: "RunaiWizardLayout" })
    .vm.$emit("step-click", { step: { name: stepName }, name: stepName });
  await flushPromises();
};

describe("DepartmentEditFormV3", () => {
  beforeEach(() => {
    delete mockRouteQuery.stepName;
    vi.clearAllMocks();
  });

  it("shows the six steps with Steps 2–6 disabled until Step 1 is done", async () => {
    const wrapper = mountForm();
    await flushPromises();

    const steps = wizardSteps(wrapper);
    expect(steps.map((step) => step.title)).toEqual([
      "Cluster & name",
      "GPU quota",
      "CPU quota",
      "Projects",
      "Scheduling rules",
      "Access rules",
    ]);
    expect(steps.slice(1).every((step) => step.disabled)).toBe(true);
    expect(wrapper.findComponent(ScopeNameStep).exists()).toBe(true);
  });

  it("switches to the narrow layout below the design's 1680px wide frame", async () => {
    const wrapper = mountForm();
    await flushPromises();

    const layout = wrapper.findComponent({ name: "RunaiWizardLayout" });
    expect(layout.props()).toMatchObject({
      width: "1201px",
      collapseLeftOnSmallScreen: true,
      hideRightOnSmallScreen: true,
      collapseLeftBreakpoint: 1680,
      hideRightBreakpoint: 1680,
    });
  });

  it("hides the CPU quota step when no node pool has CPU quota on", async () => {
    const wrapper = mountForm({ nodePoolsMinimal: [mockCpuDisabledNodePool] });
    await flushPromises();

    expect(wizardSteps(wrapper).map((step) => step.name)).not.toContain(EStepNames.CPUQuota);
  });

  it("keeps the GPU quota step when no node pool has quota on", async () => {
    const noQuotaPool = {
      ...mockCpuDisabledNodePool,
      schedulingCapabilities: { ...mockCpuDisabledNodePool.schedulingCapabilities, quota: false },
    };
    const wrapper = mountForm({ nodePoolsMinimal: [noQuotaPool] });
    await flushPromises();

    expect(wizardSteps(wrapper).map((step) => step.name)).toContain(EStepNames.GpuQuota);
  });

  it("opens every step in edit mode", async () => {
    const wrapper = mountForm({ isNewDepartment: false });
    await flushPromises();

    expect(wizardSteps(wrapper).every((step) => !step.disabled)).toBe(true);
  });

  it.each([
    { stepName: EStepNames.GpuQuota, nodePools: twoPoolDepartment.defaultNodePools },
    { stepName: EStepNames.CPUQuota, nodePools: [mockCpuEnabledNodePool.name] },
  ])("lists the step's node pools with the first one selected in $stepName", async ({ stepName, nodePools }) => {
    const wrapper = await mountAtStep(stepName, { isNewDepartment: false });

    const panel = wrapper.findComponent(NodePoolsPriorityList);
    expect(panel.props("nodePools")).toEqual(nodePools);
    expect(panel.props("modelValue")).toBe(mockCpuEnabledNodePool.name);
  });

  describe("node pool reorder", () => {
    const reordered = [mockCpuDisabledNodePool.name, mockCpuEnabledNodePool.name];

    const reorderThenOpenSchedulingRules = async (isNewDepartment: boolean) => {
      const wrapper = await mountAtStep(EStepNames.GpuQuota, { isNewDepartment });

      wrapper.findComponent(NodePoolsPriorityList).vm.$emit("update:node-pools", reordered);
      await flushPromises();
      await clickStep(wrapper, EStepNames.SchedulingRules);
      return wrapper;
    };

    it("asks to save an edited department's new order before leaving the step, then saves it", async () => {
      const wrapper = await reorderThenOpenSchedulingRules(false);

      expect(wrapper.findComponent(UnsavedChangesDialog).props("modelValue")).toBe(true);
      expect(wrapper.findComponent(SchedulingRulesStep).exists()).toBe(false);

      wrapper.findComponent(UnsavedChangesDialog).vm.$emit("save");
      await flushPromises();

      const [payload] = wrapper.emitted("on-save")?.[0] as [ISaveOrgUnitPayload<Department>];
      expect(payload.action).toBe(ESaveOrgUnitAction.Save);
      expect(payload.department?.defaultNodePools).toEqual(reordered);
      expect(wrapper.findComponent(SchedulingRulesStep).exists()).toBe(false);

      payload.onComplete?.();
      await flushPromises();
      expect(wrapper.findComponent(SchedulingRulesStep).exists()).toBe(true);
    });

    it("stays on the step with the order unsaved when the save fails", async () => {
      const wrapper = await reorderThenOpenSchedulingRules(false);

      wrapper.findComponent(UnsavedChangesDialog).vm.$emit("save");
      await flushPromises();
      const [payload] = wrapper.emitted("on-save")?.[0] as [ISaveOrgUnitPayload<Department>];
      payload.onError?.();
      await flushPromises();

      expect(wrapper.findComponent(SchedulingRulesStep).exists()).toBe(false);
      expect(wrapper.findComponent(NodePoolsPriorityList).props("nodePools")).toEqual(reordered);
      expect(wrapper.findComponent(RunaiFormWrapper).props("initDirty")).toBe(true);
    });

    it("sends one save when Save changes is clicked again or the dialog is closed while saving", async () => {
      const wrapper = await reorderThenOpenSchedulingRules(false);
      const dialog = wrapper.findComponent(UnsavedChangesDialog);

      dialog.vm.$emit("save");
      dialog.vm.$emit("save");
      await flushPromises();
      dialog.vm.$emit("update:model-value", false);
      await flushPromises();

      expect(wrapper.emitted("on-save")).toHaveLength(1);
      const [payload] = wrapper.emitted("on-save")?.[0] as [ISaveOrgUnitPayload<Department>];
      payload.onComplete?.();
      await flushPromises();
      expect(wrapper.findComponent(SchedulingRulesStep).exists()).toBe(true);
    });

    it("locks reordering while a save is in progress", async () => {
      const wrapper = await mountAtStep(EStepNames.GpuQuota, { isNewDepartment: false, isDepartmentSubmitting: true });

      expect(wrapper.findComponent(NodePoolsPriorityList).props("disableReorder")).toBe(true);
    });

    it("restores the saved order when the user discards it", async () => {
      const wrapper = await reorderThenOpenSchedulingRules(false);

      wrapper.findComponent(UnsavedChangesDialog).vm.$emit("discard");
      await flushPromises();

      expect(wrapper.emitted("on-save")).toBeUndefined();
      expect(wrapper.findComponent(SchedulingRulesStep).exists()).toBe(true);

      await clickStep(wrapper, EStepNames.GpuQuota);
      expect(wrapper.findComponent(NodePoolsPriorityList).props("nodePools")).toEqual(
        twoPoolDepartment.defaultNodePools,
      );
    });

    it("keeps the selected node pool and marks the form dirty until the order is restored", async () => {
      const wrapper = await mountAtStep(EStepNames.GpuQuota, { isNewDepartment: false });
      const isDirty = () => wrapper.findComponent(RunaiFormWrapper).props("initDirty");
      const panel = () => wrapper.findComponent(NodePoolsPriorityList);
      expect(isDirty()).toBe(false);

      panel().vm.$emit("update:node-pools", reordered);
      await flushPromises();
      expect(panel().props("modelValue")).toBe(mockCpuEnabledNodePool.name);
      expect(isDirty()).toBe(true);

      panel().vm.$emit("update:node-pools", twoPoolDepartment.defaultNodePools);
      await flushPromises();
      expect(isDirty()).toBe(false);
    });

    it("keeps node pools without CPU quota in place when reordering in the CPU quota step", async () => {
      const h100Pool = { ...mockCpuEnabledNodePool, id: "pool-3", name: "h100-pool" };
      const wrapper = await mountAtStep(EStepNames.CPUQuota, {
        isNewDepartment: false,
        department: {
          ...mockDepartment,
          defaultNodePools: [mockCpuEnabledNodePool.name, mockCpuDisabledNodePool.name, h100Pool.name],
        },
        nodePoolsMinimal: [mockCpuEnabledNodePool, mockCpuDisabledNodePool, h100Pool],
      });

      wrapper
        .findComponent(NodePoolsPriorityList)
        .vm.$emit("update:node-pools", [h100Pool.name, mockCpuEnabledNodePool.name]);
      await flushPromises();
      expect(wrapper.findComponent(NodePoolsPriorityList).props("nodePools")).toEqual([
        h100Pool.name,
        mockCpuEnabledNodePool.name,
      ]);

      await clickStep(wrapper, EStepNames.GpuQuota);
      wrapper.findComponent(UnsavedChangesDialog).vm.$emit("save");
      await flushPromises();

      const [payload] = wrapper.emitted("on-save")?.[0] as [ISaveOrgUnitPayload<Department>];
      expect(payload.department?.defaultNodePools).toEqual([
        h100Pool.name,
        mockCpuDisabledNodePool.name,
        mockCpuEnabledNodePool.name,
      ]);
    });

    it("disables reordering with an explanation in view mode", async () => {
      const wrapper = await mountAtStep(EStepNames.GpuQuota, { isNewDepartment: false, readOnly: true });

      expect(wrapper.findComponent(NodePoolsPriorityList).props()).toMatchObject({
        disableReorder: true,
        disableReorderTooltip: EOrgUnitFormTooltip.NodePoolsPriorityOrderReadOnly,
      });
    });
  });

  describe("step from the URL", () => {
    it.each([
      {
        name: "opens a valid step of an existing department",
        stepName: EStepNames.CPUQuota,
        isNew: false,
        want: EStepNames.CPUQuota,
      },
      {
        name: "sends today's GPU over quota link to GPU quota",
        stepName: EStepNames.OverQuota,
        isNew: false,
        want: EStepNames.GpuQuota,
      },
      {
        name: "falls back to the first step for an unknown step",
        stepName: "no-such-step",
        isNew: false,
        want: EStepNames.ScopeName,
      },
      {
        name: "falls back to the first step while a new department's other steps are locked",
        stepName: EStepNames.GpuQuota,
        isNew: true,
        want: EStepNames.ScopeName,
      },
      {
        name: "doesn't mark Step 1 as done for a new department",
        stepName: EStepNames.ScopeName,
        isNew: true,
        want: EStepNames.ScopeName,
      },
    ])("$name", async ({ stepName, isNew, want }) => {
      const wrapper = await mountAtStep(stepName as EStepNames, { isNewDepartment: isNew });

      expect(wrapper.findComponent({ name: "RunaiWizardLayout" }).props("activeStep")).toBe(want);
      if (isNew) expect(wizardSteps(wrapper)[0].done).toBe(false);
    });
  });

  it("moves past Step 1 only once the department is saved", async () => {
    const wrapper = mountForm({ isNewDepartment: false }, { ScopeNameStep: true });
    await flushPromises();

    wrapper.findComponent(OrgUnitFormFooter).vm.$emit("submit-and-continue");
    await flushPromises();
    expect(wrapper.findComponent({ name: "RunaiWizardLayout" }).props("activeStep")).toBe(EStepNames.ScopeName);

    const [payload] = wrapper.emitted("on-save")?.[0] as [ISaveOrgUnitPayload<Department>];
    payload.onComplete?.();
    await flushPromises();

    expect(wrapper.findComponent({ name: "RunaiWizardLayout" }).props("activeStep")).toBe(EStepNames.GpuQuota);
  });

  describe("failed saves", () => {
    it.each(["submit-and-continue", "submit-and-close"])("stops the busy state after a failed %s", async (event) => {
      vi.spyOn(storageUtil, "get").mockReturnValue(true);
      const wrapper = mountForm({}, { ScopeNameStep: true });
      await flushPromises();

      wrapper.findComponent(OrgUnitFormFooter).vm.$emit(event);
      await flushPromises();
      expect(wrapper.findComponent(OrgUnitFormFooter).props("submittingAction")).not.toBeNull();

      const [payload] = wrapper.emitted("on-save")?.[0] as [ISaveOrgUnitPayload<Department>];
      payload.onError?.();
      await flushPromises();

      expect(wrapper.findComponent(OrgUnitFormFooter).props("submittingAction")).toBeNull();
      expect(wrapper.findComponent({ name: "RunaiWizardLayout" }).props("activeStep")).toBe(EStepNames.ScopeName);
    });
  });

  it("saves and closes when Scheduling rules is the last step", async () => {
    const pinia = createTestingPinia({ createSpy: vi.fn });
    vi.mocked(usePermissionStore(pinia).hasPermission).mockImplementation(
      (resource) => resource !== ResourceType.AccessRules,
    );
    mockRouteQuery.stepName = EStepNames.SchedulingRules;
    const wrapper = mount(DepartmentEditFormV3, {
      props: { department: twoPoolDepartment, isNewDepartment: false, nodePoolsMinimal: [mockCpuEnabledNodePool] },
      global: {
        plugins: [[Quasar, { plugins: { Notify } }], pinia],
        stubs: { RunaiFormWrapper: RunaiFormWrapperStub, SchedulingRulesStep: true },
      },
    });
    await flushPromises();

    const step = wrapper.findComponent(SchedulingRulesStep);
    expect(step.props("isLastStep")).toBe(true);
    step.vm.$emit("save-scheduling-rules", { schedulingRules: null, navigateAfterSave: true });
    await flushPromises();
    const [payload] = wrapper.emitted("on-save")?.[0] as [ISaveOrgUnitPayload<Department>];

    expect(payload.action).toBe(ESaveOrgUnitAction.Close);
    expect(payload.onComplete).toBeUndefined();
  });

  it("renders today's access rules step and the footer in the last step", async () => {
    const wrapper = await mountAtStep(EStepNames.AccessRules, { isNewDepartment: false });

    expect(wrapper.findComponent(AccessRulesStep).exists()).toBe(true);
    expect(wrapper.findComponent(OrgUnitFormFooter).props("isLastStep")).toBe(true);
  });
});
