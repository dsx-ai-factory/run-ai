/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { nextTick } from "vue";
import type { ComponentPublicInstance } from "vue";

// Component under test
import ProjectEditFormV2 from "./project-edit-form-v2.vue";

// Child components for assertions
import { CpuQuotaStep } from "@/components/org-unit/v2/common/cpu-quota-step";
import { ScopeNameStep } from "@/components/org-unit/v2/project/steps/scope-name-step";
import { UnsavedChangesDialog } from "@/components/org-unit/v2/common/unsaved-changes-dialog";
import { OrgUnitCreationWarningDialog } from "@/components/org-unit/v2/common/org-unit-creation-warning-dialog";
import { ExploreNewExperienceDialog } from "@/components/org-unit/v2/common/explore-new-experience-dialog";
import { OrgUnitFormFooter } from "@/components/org-unit/v2/org-unit-form-footer";
import { CpuQuotaInput } from "@/components/org-unit/v2/common/cpu-quota-input";

// Storage keys
import { STORAGE_KEYS } from "@/models/storage.model";

// Types
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";
import type { Department } from "@/swagger-models/org-unit-service-client";

// Models
import { SettingKeys } from "@/models/setting.model";
import { ESaveOrgUnitAction } from "@/models/org-unit-forms.model";
import { EQuotaInputState } from "@/models/org-unit.model";

// Mocks
import {
  mockProject,
  mockDepartment,
  mockMultiNodePoolProject,
  mockMultiNodePoolDepartment,
  mockProjectExceedingQuota,
  mockProjectUnderAllocated,
  mockProjectWithinLimits,
  mockDepartmentWithLimit,
  mockCpuEnabledNodePools,
  mockNoCpuEnabledNodePools,
  mockNoGpuQuotaEnabledNodePools,
  mockFullCapabilitiesNodePools,
  createMockRouter,
  createMockRoute,
  RunaiFormWrapperStub,
  createStepClickPayload,
  ETestStepNames,
} from "./project-edit-form-v2.mock";

// ============================================================================
// Mock Setup
// ============================================================================

// Mutable mock route for dynamic stepName query parameter per test
const mockRouteQuery: { stepName?: string } = {};

// Mock vue-router's useRoute to provide route.query.stepName
vi.mock("vue-router", () => ({
  useRoute: () => ({
    query: mockRouteQuery,
    params: {},
    name: "project-new-v2",
    meta: {},
  }),
  useRouter: () => ({
    push: vi.fn(),
    back: vi.fn(),
    replace: vi.fn(),
  }),
}));

vi.mock("@/utils/storage.util", () => ({
  storageUtil: {
    get: vi.fn(),
    save: vi.fn(),
    remove: vi.fn(),
  },
}));

import { storageUtil } from "@/utils/storage.util";

vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: {
    getAccessRulesCount: vi.fn().mockResolvedValue(0),
  },
}));

/**
 * Helper to set the stepName query parameter for testing initial step navigation
 */
const setMockRouteStepName = (stepName?: string) => {
  if (stepName) {
    mockRouteQuery.stepName = stepName;
  } else {
    delete mockRouteQuery.stepName;
  }
};

// ============================================================================
// Test Suite
// ============================================================================

describe("ProjectEditFormV2", () => {
  let mockRouter: ReturnType<typeof createMockRouter>;
  let mockRoute: ReturnType<typeof createMockRoute>;

  beforeEach(() => {
    mockRouter = createMockRouter();
    mockRoute = createMockRoute();
    // Reset mock route query for each test
    setMockRouteStepName(undefined);
    vi.clearAllMocks();
  });

  // ==========================================================================
  // Test Utilities
  // ==========================================================================

  interface CreateWrapperOptions {
    props?: Record<string, unknown>;
    nodePoolsMinimal?: typeof mockCpuEnabledNodePools;
    isLimitQuotaOverSubscriptionEnabled?: boolean;
    useGpuStub?: boolean;
  }

  const GpuQuotaStepStub = {
    name: "GpuQuotaStep",
    template: '<div class="gpu-quota-step-stub"></div>',
    emits: ["save-gpu-quota"],
  };

  const createWrapper = (options: CreateWrapperOptions = {}) => {
    const {
      props = {},
      nodePoolsMinimal = mockCpuEnabledNodePools,
      isLimitQuotaOverSubscriptionEnabled = false,
      useGpuStub = false,
    } = options;

    const settingsMap = new Map();
    settingsMap.set(SettingKeys.CPUResourcesQuota, { value: true }); // Keep for backward compatibility

    return mount(ProjectEditFormV2, {
      props: {
        project: mockProject,
        department: mockDepartment as Department,
        isNewProject: true,
        nodePoolsMinimal,
        ...props,
      },
      global: {
        plugins: [
          Quasar,
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              Setting: {
                settings: settingsMap,
                limitQuotaOverSubscription: { value: isLimitQuotaOverSubscriptionEnabled },
              },
            },
          }),
        ],
        mocks: {
          $route: mockRoute,
          $router: mockRouter,
        },
        stubs: {
          RunaiFormWrapper: RunaiFormWrapperStub,
          ...(useGpuStub
            ? {
                GpuQuotaStep: GpuQuotaStepStub,
                GpuQuotaInput: true,
              }
            : {}),
        },
      },
    });
  };

  const getWizardLayout = (wrapper: VueWrapper<ComponentPublicInstance>) =>
    wrapper.findComponent({ name: "RunaiWizardLayout" });

  const getWizardSteps = (wrapper: VueWrapper<ComponentPublicInstance>) =>
    getWizardLayout(wrapper).props("steps") as IRunaiWizardLayoutStep[];

  const navigateToStep = async (wrapper: VueWrapper<ComponentPublicInstance>, stepName: string) => {
    const wizardLayout = getWizardLayout(wrapper);
    wizardLayout.vm.$emit("step-click", createStepClickPayload(stepName));
    await flushPromises();
  };

  const makeCpuQuotaStepDirty = async (wrapper: VueWrapper<ComponentPublicInstance>) => {
    const cpuQuotaStep = wrapper.findComponent(CpuQuotaStep);
    const cpuInput = cpuQuotaStep.findComponent(CpuQuotaInput);
    cpuInput.vm.$emit("update:model-value", 5);
    await nextTick();
  };

  // ==========================================================================
  // Basic Rendering Tests
  // ==========================================================================

  describe("Rendering", () => {
    it("renders the component correctly", () => {
      const wrapper = createWrapper();

      expect(wrapper.exists()).toBe(true);
      expect(getWizardLayout(wrapper).exists()).toBe(true);
    });

    it("renders ScopeNameStep as the initial step", async () => {
      const wrapper = createWrapper();
      await nextTick();

      expect(wrapper.findComponent(ScopeNameStep).exists()).toBe(true);
      expect(wrapper.findComponent(CpuQuotaStep).exists()).toBe(false);
    });

    it("renders the wizard summary sidebar", async () => {
      const wrapper = createWrapper();
      await nextTick();

      expect(wrapper.text()).toContain("Resources");
      expect(wrapper.text()).toContain("Scheduling rules");
      expect(wrapper.text()).toContain("Access rules");
    });

    it("renders the form footer", async () => {
      const wrapper = createWrapper();
      await nextTick();

      expect(wrapper.findComponent(OrgUnitFormFooter).exists()).toBe(true);
    });
  });

  // ==========================================================================
  // CPU Quota Step Visibility Tests
  // ==========================================================================

  describe("CPU Quota Step Visibility", () => {
    it("includes CPU quota step when at least one node pool has CPU quota enabled", async () => {
      const wrapper = createWrapper({ nodePoolsMinimal: mockCpuEnabledNodePools });
      await nextTick();

      const steps = getWizardSteps(wrapper);
      expect(steps.some((s) => s.name === ETestStepNames.CPUQuota)).toBe(true);
    });

    it("excludes CPU quota step when no node pools have CPU quota enabled", async () => {
      const wrapper = createWrapper({ nodePoolsMinimal: mockNoCpuEnabledNodePools });
      await nextTick();

      const steps = getWizardSteps(wrapper);
      expect(steps.some((s) => s.name === ETestStepNames.CPUQuota)).toBe(false);
    });
  });

  // ==========================================================================
  // GPU Quota Step Visibility Tests
  // ==========================================================================

  describe("GPU Quota Step Visibility", () => {
    it("includes GPU quota step when at least one node pool has quota enabled", async () => {
      const wrapper = createWrapper({ nodePoolsMinimal: mockFullCapabilitiesNodePools });
      await nextTick();

      const steps = getWizardSteps(wrapper);
      expect(steps.some((s) => s.name === ETestStepNames.GpuQuota)).toBe(true);
    });

    it("excludes GPU quota step when no node pools have quota enabled", async () => {
      const wrapper = createWrapper({ nodePoolsMinimal: mockNoGpuQuotaEnabledNodePools });
      await nextTick();

      const steps = getWizardSteps(wrapper);
      expect(steps.some((s) => s.name === ETestStepNames.GpuQuota)).toBe(false);
    });

    it("excludes GPU quota step when nodePoolsMinimal is empty", async () => {
      const wrapper = createWrapper({ nodePoolsMinimal: [] });
      await nextTick();

      const steps = getWizardSteps(wrapper);
      expect(steps.some((s) => s.name === ETestStepNames.GpuQuota)).toBe(false);
    });
  });

  // ==========================================================================
  // Initial Step from Query Parameter Tests
  // ==========================================================================

  describe("Initial Step from Query Parameter", () => {
    it("starts at first step when no stepName query param is provided", async () => {
      setMockRouteStepName(undefined);
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      const wizardLayout = getWizardLayout(wrapper);
      expect(wizardLayout.props("activeStep")).toBe(ETestStepNames.ScopeName);
    });

    it("navigates to specified step when stepName query param is provided", async () => {
      setMockRouteStepName(ETestStepNames.CPUQuota);
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      const wizardLayout = getWizardLayout(wrapper);
      expect(wizardLayout.props("activeStep")).toBe(ETestStepNames.CPUQuota);
    });

    it("navigates to scheduling-rules step when stepName is scheduling-rules", async () => {
      setMockRouteStepName(ETestStepNames.SchedulingRules);
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      const wizardLayout = getWizardLayout(wrapper);
      expect(wizardLayout.props("activeStep")).toBe(ETestStepNames.SchedulingRules);
    });

    it("marks first step as complete when navigating directly to another step", async () => {
      setMockRouteStepName(ETestStepNames.CPUQuota);
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      const steps = getWizardSteps(wrapper);
      const firstStep = steps.find((s) => s.name === ETestStepNames.ScopeName);
      expect(firstStep?.done).toBe(true);
    });

    it("does not mark first step as complete when starting from first step", async () => {
      setMockRouteStepName(undefined);
      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const steps = getWizardSteps(wrapper);
      const firstStep = steps.find((s) => s.name === ETestStepNames.ScopeName);
      expect(firstStep?.done).toBe(false);
    });
  });

  // ==========================================================================
  // Wizard Steps Configuration Tests
  // ==========================================================================

  describe("Wizard Steps Configuration", () => {
    it("includes all expected steps when CPU node pools exist", async () => {
      const wrapper = createWrapper({ nodePoolsMinimal: mockCpuEnabledNodePools });
      await nextTick();

      const stepNames = getWizardSteps(wrapper).map((s) => s.name);

      expect(stepNames).toContain(ETestStepNames.ScopeName);
      expect(stepNames).toContain(ETestStepNames.GpuQuota);
      expect(stepNames).toContain(ETestStepNames.OverQuota);
      expect(stepNames).toContain(ETestStepNames.CPUQuota);
      expect(stepNames).toContain(ETestStepNames.SchedulingRules);
      // Note: AccessRules step visibility depends on permissions (hasPermission check)
      // It may or may not be visible depending on the permission store mock
    });

    it("sets correct step titles", async () => {
      const wrapper = createWrapper();
      await nextTick();

      const steps = getWizardSteps(wrapper);

      expect(steps.find((s) => s.name === ETestStepNames.ScopeName)?.title).toBe("Department & name");
      expect(steps.find((s) => s.name === ETestStepNames.CPUQuota)?.title).toBe("CPU quota");
    });

    it("enables all steps in edit mode", async () => {
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      const steps = getWizardSteps(wrapper);
      steps.forEach((step) => {
        expect(step.disabled).toBe(false);
      });
    });

    it("disables non-first steps in create mode", async () => {
      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const steps = getWizardSteps(wrapper);

      expect(steps[0].disabled).toBe(false);
      expect(steps[1].disabled).toBe(true);
    });
  });

  // ==========================================================================
  // Unsaved Changes Dialog Tests
  // ==========================================================================

  describe("Unsaved Changes Dialog", () => {
    it("shows dialog when navigating away from dirty CPU quota step", async () => {
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);
      expect(wrapper.findComponent(CpuQuotaStep).exists()).toBe(true);

      await makeCpuQuotaStepDirty(wrapper);
      expect(wrapper.findComponent(CpuQuotaStep).vm.isDirty).toBe(true);

      await navigateToStep(wrapper, ETestStepNames.ScopeName);

      const dialog = wrapper.findComponent(UnsavedChangesDialog);
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("modelValue")).toBe(true);
    });

    it("navigates without dialog when step is clean", async () => {
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);
      expect(wrapper.findComponent(CpuQuotaStep).exists()).toBe(true);

      // Navigate away without making changes
      await navigateToStep(wrapper, ETestStepNames.ScopeName);

      const dialog = wrapper.findComponent(UnsavedChangesDialog);
      expect(dialog.props("modelValue")).toBe(false);
      expect(wrapper.findComponent(ScopeNameStep).exists()).toBe(true);
    });

    it("discards changes and navigates when Discard is clicked", async () => {
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);
      await makeCpuQuotaStepDirty(wrapper);
      await navigateToStep(wrapper, ETestStepNames.ScopeName);

      const dialog = wrapper.findComponent(UnsavedChangesDialog);
      dialog.vm.$emit("discard");
      await flushPromises();

      expect(wrapper.findComponent(ScopeNameStep).exists()).toBe(true);
      expect(wrapper.findComponent(CpuQuotaStep).exists()).toBe(false);
    });

    it("saves changes and navigates when Save is clicked", async () => {
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);
      await makeCpuQuotaStepDirty(wrapper);
      await navigateToStep(wrapper, ETestStepNames.ScopeName);

      const dialog = wrapper.findComponent(UnsavedChangesDialog);
      dialog.vm.$emit("save");
      await flushPromises();

      expect(wrapper.findComponent(ScopeNameStep).exists()).toBe(true);
    });

    it("shows dialog when switching node pools with unsaved changes", async () => {
      const wrapper = createWrapper({
        props: {
          project: mockMultiNodePoolProject,
          department: mockMultiNodePoolDepartment,
          isNewProject: false,
        },
      });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);
      await makeCpuQuotaStepDirty(wrapper);

      const nodePoolTabs = wrapper.findComponent({ name: "NodePoolTabs" });
      // Use pool ID "2" instead of name "pool2" - tabs now track by ID
      nodePoolTabs.vm.$emit("update:model-value", "2");
      await flushPromises();

      const dialog = wrapper.findComponent(UnsavedChangesDialog);
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("modelValue")).toBe(true);
    });

    it("blocks Continue action and shows dialog when current step is dirty", async () => {
      // Use multi-node pool project to ensure footer is rendered
      const wrapper = createWrapper({
        props: {
          isNewProject: false,
          project: mockMultiNodePoolProject,
          department: mockMultiNodePoolDepartment,
        },
      });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);
      await makeCpuQuotaStepDirty(wrapper);

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      footer.vm.$emit("next");
      await flushPromises();

      const dialog = wrapper.findComponent(UnsavedChangesDialog);
      expect(dialog.props("modelValue")).toBe(true);
      expect(wrapper.findComponent(CpuQuotaStep).exists()).toBe(true);
    });

    it("defers close action until unsaved changes decision is made", async () => {
      // Use multi-node pool project to ensure footer is rendered
      const wrapper = createWrapper({
        props: {
          isNewProject: false,
          project: mockMultiNodePoolProject,
          department: mockMultiNodePoolDepartment,
        },
      });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);
      await makeCpuQuotaStepDirty(wrapper);

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      footer.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("on-cancel")).toBeUndefined();

      const dialog = wrapper.findComponent(UnsavedChangesDialog);
      dialog.vm.$emit("discard");
      await flushPromises();

      expect(wrapper.emitted("on-cancel")).toBeTruthy();
    });
  });

  // ==========================================================================
  // Footer Actions Tests
  // ==========================================================================

  describe("Footer Actions", () => {
    it("emits on-cancel when close is clicked", async () => {
      const wrapper = createWrapper();
      await nextTick();

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      footer.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("on-cancel")).toBeTruthy();
    });

    it("emits on-save with Close action on submit-and-close in edit mode", async () => {
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      footer.vm.$emit("submit-and-close");
      await flushPromises();

      const emitted = wrapper.emitted("on-save");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]?.[0]).toMatchObject({
        action: ESaveOrgUnitAction.Close,
      });
    });

    it("emits on-save with Continue action on submit-and-continue in edit mode", async () => {
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      footer.vm.$emit("submit-and-continue");
      await flushPromises();

      const emitted = wrapper.emitted("on-save");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]?.[0]).toMatchObject({
        action: ESaveOrgUnitAction.Continue,
      });
    });

    it("resets submittingAction after save completes", async () => {
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      footer.vm.$emit("submit-and-close");
      await flushPromises();

      const emitted = wrapper.emitted("on-save");
      const payload = emitted?.[0]?.[0] as { onComplete: () => void };

      expect(footer.props("submittingAction")).toBe("close");

      payload.onComplete();
      await nextTick();

      expect(footer.props("submittingAction")).toBe(null);
    });
  });

  // ==========================================================================
  // Warning Dialog Tests
  // ==========================================================================

  describe("Warning Dialog", () => {
    it("renders warning dialog component", async () => {
      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const warningDialog = wrapper.findComponent(OrgUnitCreationWarningDialog);
      expect(warningDialog.exists()).toBe(true);
      expect(warningDialog.props("modelValue")).toBe(false);
    });

    it("does not show warning dialog in edit mode", async () => {
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      footer.vm.$emit("submit-and-close");
      await flushPromises();

      const warningDialog = wrapper.findComponent(OrgUnitCreationWarningDialog);
      expect(warningDialog.props("modelValue")).toBe(false);
      expect(wrapper.emitted("on-save")).toBeTruthy();
    });

    it("emits on-save when create-and-close event is received", async () => {
      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const warningDialog = wrapper.findComponent(OrgUnitCreationWarningDialog);
      warningDialog.vm.$emit("create-and-close");
      await flushPromises();

      const emitted = wrapper.emitted("on-save");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]?.[0]).toMatchObject({
        action: ESaveOrgUnitAction.Close,
      });
    });

    it("has correct warning dialog props", async () => {
      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const warningDialog = wrapper.findComponent(OrgUnitCreationWarningDialog);
      expect(warningDialog.props("title")).toBe("Create Project & Close?");
      expect(warningDialog.props("continueEditingLabel")).toBe("Continue Editing");
      expect(warningDialog.props("createAndCloseLabel")).toBe("Create Project & Close");
    });
  });

  // ==========================================================================
  // Scope Update Tests
  // ==========================================================================

  describe("Scope Update", () => {
    it("relays update-scope event from ScopeNameStep", async () => {
      const wrapper = createWrapper();
      await nextTick();

      const scopeNameStep = wrapper.findComponent(ScopeNameStep);
      const scopePayload = { clusterId: "new-cluster", departmentId: "new-dept" };
      scopeNameStep.vm.$emit("update-scope", scopePayload);
      await flushPromises();

      const emitted = wrapper.emitted("update-scope");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]?.[0]).toEqual(scopePayload);
    });
  });

  // ==========================================================================
  // Data Passing Tests
  // ==========================================================================

  describe("Data Passing", () => {
    it("passes correct project data to ScopeNameStep", async () => {
      const wrapper = createWrapper();
      await nextTick();

      const scopeNameStep = wrapper.findComponent(ScopeNameStep);
      expect(scopeNameStep.props("isNewProject")).toBe(true);
      expect(scopeNameStep.props("clusterId")).toBe(mockProject.clusterId);
      expect(scopeNameStep.props("parentId")).toBe(mockProject.parentId);
    });

    it("passes department data to ScopeNameStep", async () => {
      const wrapper = createWrapper();
      await nextTick();

      const scopeNameStep = wrapper.findComponent(ScopeNameStep);
      const department = scopeNameStep.props("department") as { name: string } | null;
      expect(department).toBeTruthy();
      expect(department?.name).toBe(mockDepartment.name);
    });

    it("passes correct props to footer in create mode", async () => {
      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      expect(footer.props("isNewEntity")).toBe(true);
      expect(footer.props("entityType")).toBe("Project");
    });

    it("passes correct props to footer in edit mode", async () => {
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      expect(footer.props("isNewEntity")).toBe(false);
    });
  });

  // ==========================================================================
  // CPU Quota Validation State Tests
  // ==========================================================================

  describe("CPU Quota Validation States", () => {
    it("shows default state when quota is within limits", async () => {
      const wrapper = createWrapper({
        props: {
          project: mockProjectWithinLimits,
          department: mockDepartmentWithLimit,
          isNewProject: false,
        },
      });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);

      const cpuQuotaStep = wrapper.findComponent(CpuQuotaStep);
      const cpuInput = cpuQuotaStep.findComponent(CpuQuotaInput);

      expect(cpuInput.props("state")).toBe(EQuotaInputState.Default);
    });

    it("shows warning state when quota exceeds department limit (over-subscription disabled)", async () => {
      const wrapper = createWrapper({
        props: {
          project: mockProjectExceedingQuota,
          department: mockDepartmentWithLimit,
          isNewProject: false,
        },
        isLimitQuotaOverSubscriptionEnabled: false,
      });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);

      const cpuQuotaStep = wrapper.findComponent(CpuQuotaStep);
      const cpuInput = cpuQuotaStep.findComponent(CpuQuotaInput);

      expect(cpuInput.props("state")).toBe(EQuotaInputState.Warning);
    });

    it("shows error state when quota exceeds department limit (over-subscription enabled)", async () => {
      const wrapper = createWrapper({
        props: {
          project: mockProjectExceedingQuota,
          department: mockDepartmentWithLimit,
          isNewProject: false,
        },
        isLimitQuotaOverSubscriptionEnabled: true,
      });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);

      const cpuQuotaStep = wrapper.findComponent(CpuQuotaStep);
      const cpuInput = cpuQuotaStep.findComponent(CpuQuotaInput);

      expect(cpuInput.props("state")).toBe(EQuotaInputState.Error);
    });

    it("shows warning state when quota is under allocated non-preemptible (over-subscription disabled)", async () => {
      const wrapper = createWrapper({
        props: {
          project: mockProjectUnderAllocated,
          department: mockDepartmentWithLimit,
          isNewProject: false,
        },
        isLimitQuotaOverSubscriptionEnabled: false,
      });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);

      const cpuQuotaStep = wrapper.findComponent(CpuQuotaStep);
      const cpuInput = cpuQuotaStep.findComponent(CpuQuotaInput);

      expect(cpuInput.props("state")).toBe(EQuotaInputState.Warning);
    });

    it("shows error state when quota is under allocated non-preemptible (over-subscription enabled)", async () => {
      const wrapper = createWrapper({
        props: {
          project: mockProjectUnderAllocated,
          department: mockDepartmentWithLimit,
          isNewProject: false,
        },
        isLimitQuotaOverSubscriptionEnabled: true,
      });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);

      const cpuQuotaStep = wrapper.findComponent(CpuQuotaStep);
      const cpuInput = cpuQuotaStep.findComponent(CpuQuotaInput);

      expect(cpuInput.props("state")).toBe(EQuotaInputState.Error);
    });

    it("displays warning message when quota exceeds department limit", async () => {
      const wrapper = createWrapper({
        props: {
          project: mockProjectExceedingQuota,
          department: mockDepartmentWithLimit,
          isNewProject: false,
        },
        isLimitQuotaOverSubscriptionEnabled: false,
      });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);

      const cpuQuotaStep = wrapper.findComponent(CpuQuotaStep);
      const cpuInput = cpuQuotaStep.findComponent(CpuQuotaInput);

      expect(cpuInput.props("message")).toContain("exceeds the available department quota");
    });

    it("displays warning message when quota is under allocated usage", async () => {
      const wrapper = createWrapper({
        props: {
          project: mockProjectUnderAllocated,
          department: mockDepartmentWithLimit,
          isNewProject: false,
        },
        isLimitQuotaOverSubscriptionEnabled: false,
      });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);

      const cpuQuotaStep = wrapper.findComponent(CpuQuotaStep);
      const cpuInput = cpuQuotaStep.findComponent(CpuQuotaInput);

      expect(cpuInput.props("message")).toContain("less than the usage by non-preemptive workloads");
    });

    it("displays recommended value message when quota is within limits", async () => {
      const wrapper = createWrapper({
        props: {
          project: mockProjectWithinLimits,
          department: mockDepartmentWithLimit,
          isNewProject: false,
        },
      });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.CPUQuota);

      const cpuQuotaStep = wrapper.findComponent(CpuQuotaStep);
      const cpuInput = cpuQuotaStep.findComponent(CpuQuotaInput);
      const message = cpuInput.props("message") as string;

      // Should contain recommended range info
      expect(message).toContain("Recommended:");
    });
  });

  // ==========================================================================
  // GPU Quota Step Tests
  // ==========================================================================

  describe("GPU Quota Step", () => {
    it("renders the GPU quota step when navigating to the GPU step", async () => {
      const wrapper = createWrapper({ useGpuStub: true, props: { isNewProject: false } });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.GpuQuota);
      await flushPromises();

      expect(wrapper.find(".gpu-quota-step-stub").exists()).toBe(true);
    });

    it("relays save-gpu-quota event and emits on-save when no peer changes", async () => {
      const wrapper = createWrapper({ useGpuStub: true, props: { isNewProject: false } });
      await nextTick();

      await navigateToStep(wrapper, ETestStepNames.GpuQuota);
      await flushPromises();

      const gpuQuotaStep = wrapper.findComponent({ name: "GpuQuotaStep" });
      gpuQuotaStep.vm.$emit("save-gpu-quota", {
        nodePoolName: "default",
        gpuQuota: 40,
        peerProjectChanges: [],
      });
      await flushPromises();

      const emitted = wrapper.emitted("on-save");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]?.[0]).toMatchObject({
        action: ESaveOrgUnitAction.Save,
      });
    });
  });

  // ==========================================================================
  // Explore New Experience Dialog (Guided Tour) Tests
  // ==========================================================================

  describe("Explore New Experience Dialog (Guided Tour)", () => {
    beforeEach(() => {
      vi.mocked(storageUtil.get).mockReturnValue(null);
    });

    it("renders explore dialog component", async () => {
      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const exploreDialog = wrapper.findComponent(ExploreNewExperienceDialog);
      expect(exploreDialog.exists()).toBe(true);
    });

    it("does not show explore dialog for existing projects", async () => {
      const wrapper = createWrapper({ props: { isNewProject: false } });
      await nextTick();

      const exploreDialog = wrapper.findComponent(ExploreNewExperienceDialog);
      expect(exploreDialog.props("modelValue")).toBe(false);
    });

    it("does not show explore dialog when tour was already seen", async () => {
      vi.mocked(storageUtil.get).mockImplementation((key: string) => {
        if (key === STORAGE_KEYS.PROJECT_TOUR_SEEN) return true;
        return null;
      });

      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const exploreDialog = wrapper.findComponent(ExploreNewExperienceDialog);
      expect(exploreDialog.props("modelValue")).toBe(false);
    });

    it("saves storage key when skip is clicked", async () => {
      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const exploreDialog = wrapper.findComponent(ExploreNewExperienceDialog);
      exploreDialog.vm.$emit("skip");
      await flushPromises();

      expect(storageUtil.save).toHaveBeenCalledWith(STORAGE_KEYS.PROJECT_TOUR_SEEN, true);
    });

    it("saves storage key when start tour is clicked", async () => {
      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const exploreDialog = wrapper.findComponent(ExploreNewExperienceDialog);
      exploreDialog.vm.$emit("start-tour");
      await flushPromises();

      expect(storageUtil.save).toHaveBeenCalledWith(STORAGE_KEYS.PROJECT_TOUR_SEEN, true);
    });

    it("closes dialog when skip is clicked", async () => {
      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const exploreDialog = wrapper.findComponent(ExploreNewExperienceDialog);
      exploreDialog.vm.$emit("skip");
      await flushPromises();

      expect(exploreDialog.props("modelValue")).toBe(false);
    });

    it("closes dialog when start tour is clicked", async () => {
      const wrapper = createWrapper({ props: { isNewProject: true } });
      await nextTick();

      const exploreDialog = wrapper.findComponent(ExploreNewExperienceDialog);
      exploreDialog.vm.$emit("start-tour");
      await flushPromises();

      expect(exploreDialog.props("modelValue")).toBe(false);
    });
  });
});
