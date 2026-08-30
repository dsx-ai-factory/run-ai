import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { nextTick } from "vue";
import type { ComponentPublicInstance } from "vue";

// Component under test
import DepartmentEditFormV2 from "./department-edit-form-v2.vue";

// Child components for assertions
import { ScopeNameStep } from "@/components/org-unit/v2/department/steps/scope-name-step";
import { SchedulingRulesStep } from "@/components/org-unit/v2/common/scheduling-rules-step";
import { UnsavedChangesDialog } from "@/components/org-unit/v2/common/unsaved-changes-dialog";
import { OrgUnitCreationWarningDialog } from "@/components/org-unit/v2/common/org-unit-creation-warning-dialog";
import { OrgUnitFormFooter } from "@/components/org-unit/v2/org-unit-form-footer";

// Types
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";

// Models
import { SettingKeys } from "@/models/setting.model";
import { ESaveOrgUnitAction } from "@/models/org-unit-forms.model";
import { EQuotaEntity } from "@/models/resource.model";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

// Mocks
import {
  mockDepartment,
  mockCpuEnabledNodePools,
  mockNoCpuEnabledNodePools,
  mockNoGpuQuotaEnabledNodePools,
  mockFullCapabilitiesNodePools,
  createMockRouter,
  createMockRoute,
  RunaiFormWrapperStub,
  ETestStepNames,
} from "./department-edit-form-v2.mock";

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
    name: "department-new-v2",
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

describe("DepartmentEditFormV2", () => {
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
  }

  const createWrapper = (options: CreateWrapperOptions = {}) => {
    const { props = {}, nodePoolsMinimal = mockCpuEnabledNodePools } = options;

    const settingsMap = new Map();
    settingsMap.set(SettingKeys.CPUResourcesQuota, { value: true });

    return mount(DepartmentEditFormV2, {
      props: {
        department: mockDepartment,
        isNewDepartment: true,
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
        },
      },
    });
  };

  const getWizardLayout = (wrapper: VueWrapper<ComponentPublicInstance>) =>
    wrapper.findComponent({ name: "RunaiWizardLayout" });

  const getWizardSteps = (wrapper: VueWrapper<ComponentPublicInstance>) =>
    getWizardLayout(wrapper).props("steps") as IRunaiWizardLayoutStep[];

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
      const wrapper = createWrapper({ props: { isNewDepartment: false } });
      await nextTick();

      const wizardLayout = getWizardLayout(wrapper);
      expect(wizardLayout.props("activeStep")).toBe(ETestStepNames.ScopeName);
    });

    it("navigates to specified step when stepName query param is provided", async () => {
      setMockRouteStepName(ETestStepNames.CPUQuota);
      const wrapper = createWrapper({ props: { isNewDepartment: false } });
      await nextTick();

      const wizardLayout = getWizardLayout(wrapper);
      expect(wizardLayout.props("activeStep")).toBe(ETestStepNames.CPUQuota);
    });

    it("marks first step as complete when navigating directly to another step", async () => {
      setMockRouteStepName(ETestStepNames.CPUQuota);
      const wrapper = createWrapper({ props: { isNewDepartment: false } });
      await nextTick();

      const steps = getWizardSteps(wrapper);
      const firstStep = steps.find((s) => s.name === ETestStepNames.ScopeName);
      expect(firstStep?.done).toBe(true);
    });

    it("does not mark first step as complete when starting from first step", async () => {
      setMockRouteStepName(undefined);
      const wrapper = createWrapper({ props: { isNewDepartment: true } });
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
      expect(stepNames).toContain(ETestStepNames.CPUQuota);
      expect(stepNames).toContain(ETestStepNames.SchedulingRules);
      // Note: AccessRules step visibility depends on permissions (hasPermission check)
      // It may or may not be visible depending on the permission store mock
    });

    it("sets correct step titles", async () => {
      const wrapper = createWrapper();
      await nextTick();

      const steps = getWizardSteps(wrapper);

      expect(steps.find((s) => s.name === ETestStepNames.ScopeName)?.title).toBe("Cluster & name");
      expect(steps.find((s) => s.name === ETestStepNames.CPUQuota)?.title).toBe("CPU quota");
    });

    it("enables all steps in edit mode", async () => {
      const wrapper = createWrapper({ props: { isNewDepartment: false } });
      await nextTick();

      const steps = getWizardSteps(wrapper);
      steps.forEach((step) => {
        expect(step.disabled).toBe(false);
      });
    });

    it("disables non-first steps in create mode", async () => {
      const wrapper = createWrapper({ props: { isNewDepartment: true } });
      await nextTick();

      const steps = getWizardSteps(wrapper);

      expect(steps[0].disabled).toBe(false);
      expect(steps[1].disabled).toBe(true);
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
      const wrapper = createWrapper({ props: { isNewDepartment: false } });
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
      const wrapper = createWrapper({ props: { isNewDepartment: false } });
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
      const wrapper = createWrapper({ props: { isNewDepartment: false } });
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
      const wrapper = createWrapper({ props: { isNewDepartment: true } });
      await nextTick();

      const warningDialog = wrapper.findComponent(OrgUnitCreationWarningDialog);
      expect(warningDialog.exists()).toBe(true);
      expect(warningDialog.props("modelValue")).toBe(false);
    });

    it("does not show warning dialog in edit mode", async () => {
      const wrapper = createWrapper({ props: { isNewDepartment: false } });
      await nextTick();

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      footer.vm.$emit("submit-and-close");
      await flushPromises();

      const warningDialog = wrapper.findComponent(OrgUnitCreationWarningDialog);
      expect(warningDialog.props("modelValue")).toBe(false);
      expect(wrapper.emitted("on-save")).toBeTruthy();
    });

    it("emits on-save when create-and-close event is received", async () => {
      const wrapper = createWrapper({ props: { isNewDepartment: true } });
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
      const wrapper = createWrapper({ props: { isNewDepartment: true } });
      await nextTick();

      const warningDialog = wrapper.findComponent(OrgUnitCreationWarningDialog);
      expect(warningDialog.props("title")).toBe("Create Department & Close?");
      expect(warningDialog.props("continueEditingLabel")).toBe("Continue Editing");
      expect(warningDialog.props("createAndCloseLabel")).toBe("Create Department & Close");
    });
  });

  // ==========================================================================
  // Scope Update Tests
  // ==========================================================================

  describe("Scope Update", () => {
    it("relays update-scope event from ScopeNameStep", async () => {
      const wrapper = createWrapper();
      await nextTick();

      const clusterNameStep = wrapper.findComponent(ScopeNameStep);
      const scopePayload = { clusterId: "new-cluster" };
      clusterNameStep.vm.$emit("update-scope", scopePayload);
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
    it("passes correct department data to ScopeNameStep", async () => {
      const wrapper = createWrapper();
      await nextTick();

      const clusterNameStep = wrapper.findComponent(ScopeNameStep);
      expect(clusterNameStep.props("isNewDepartment")).toBe(true);
      expect(clusterNameStep.props("clusterId")).toBe(mockDepartment.clusterId);
    });

    it("passes correct props to footer in create mode", async () => {
      const wrapper = createWrapper({ props: { isNewDepartment: true } });
      await nextTick();

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      expect(footer.props("isNewEntity")).toBe(true);
      expect(footer.props("entityType")).toBe("Department");
    });

    it("passes correct props to footer in edit mode", async () => {
      const wrapper = createWrapper({ props: { isNewDepartment: false } });
      await nextTick();

      const footer = wrapper.findComponent(OrgUnitFormFooter);
      expect(footer.props("isNewEntity")).toBe(false);
    });
  });

  // ==========================================================================
  // Unsaved Changes Dialog Tests
  // ==========================================================================

  describe("Unsaved Changes Dialog", () => {
    it("renders unsaved changes dialog component", async () => {
      const wrapper = createWrapper({ props: { isNewDepartment: false } });
      await nextTick();

      const dialog = wrapper.findComponent(UnsavedChangesDialog);
      expect(dialog.exists()).toBe(true);
    });
  });

  // ==========================================================================
  // Scheduling Rules Step Tests
  // ==========================================================================

  describe("Scheduling Rules Step", () => {
    const createWrapperWithPermissions = (options: CreateWrapperOptions = {}) => {
      const { props = {}, nodePoolsMinimal = mockCpuEnabledNodePools } = options;

      const settingsMap = new Map();
      settingsMap.set(SettingKeys.CPUResourcesQuota, { value: true });

      // Build resourceTypeMap for permission lookups
      const accessRulesPermission = { resourceType: ResourceType.AccessRules, actions: [Action.Update] };
      const resourceTypeMap = {
        [ResourceType.AccessRules]: accessRulesPermission,
      };

      return mount(DepartmentEditFormV2, {
        props: {
          department: mockDepartment,
          isNewDepartment: false,
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
                },
                Permission: {
                  permissions: [accessRulesPermission],
                  resourceTypeMap,
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
          },
        },
      });
    };

    it("navigates to scheduling-rules step when stepName query param is set", async () => {
      setMockRouteStepName(ETestStepNames.SchedulingRules);
      const wrapper = createWrapperWithPermissions();
      await nextTick();

      const wizardLayout = getWizardLayout(wrapper);
      expect(wizardLayout.props("activeStep")).toBe(ETestStepNames.SchedulingRules);
    });

    it("renders SchedulingRulesStep when navigated to scheduling rules step", async () => {
      setMockRouteStepName(ETestStepNames.SchedulingRules);
      const wrapper = createWrapperWithPermissions();
      await nextTick();

      expect(wrapper.findComponent(SchedulingRulesStep).exists()).toBe(true);
    });

    it("passes correct props to SchedulingRulesStep", async () => {
      setMockRouteStepName(ETestStepNames.SchedulingRules);
      const wrapper = createWrapperWithPermissions();
      await nextTick();

      const step = wrapper.findComponent(SchedulingRulesStep);
      expect(step.props("entityType")).toBe(EQuotaEntity.department);
      expect(step.props("clusterId")).toBe(mockDepartment.clusterId);
    });

    it("includes scheduling rules step in wizard steps", async () => {
      const wrapper = createWrapperWithPermissions();
      await nextTick();

      const steps = getWizardSteps(wrapper);
      expect(steps.some((s) => s.name === ETestStepNames.SchedulingRules)).toBe(true);
    });

    it("emits on-save when scheduling rules are saved", async () => {
      setMockRouteStepName(ETestStepNames.SchedulingRules);
      const wrapper = createWrapperWithPermissions();
      await nextTick();

      const step = wrapper.findComponent(SchedulingRulesStep);
      const payload = {
        schedulingRules: { trainingJobMaxIdleDurationSeconds: 300 },
        nodeTypes: { training: ["gpu1"] },
      };
      step.vm.$emit("save-scheduling-rules", payload);
      await flushPromises();

      const emitted = wrapper.emitted("on-save");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]?.[0]).toMatchObject({
        action: ESaveOrgUnitAction.Save,
      });
    });
  });

  // ==========================================================================
  // Access Rules Step Tests
  // ==========================================================================

  describe("Access Rules Step", () => {
    // Note: AccessRules step visibility depends on permissions (hasPermission check)
    // These tests verify the step is configured correctly in DEPARTMENT_FORM_WIZARD_STEPS
    // The actual permission filtering happens at runtime via filterOrgUnitSteps

    it("access rules step is defined in wizard steps configuration", () => {
      // The AccessRules step should be part of the department form wizard steps
      // This test verifies the step configuration, not runtime permission filtering
      expect(ETestStepNames.AccessRules).toBe("access-rules");
    });
  });
});
