import { mount, type VueWrapper } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { defineComponent, h, nextTick, reactive } from "vue";
import type { ComponentPublicInstance } from "vue";

// Component under test
import SchedulingRulesStep from "./scheduling-rules-step.vue";

// Child components (Imported for type referencing in findComponent)
import { SchedulingRulesSection } from "@/components/project/project-edit-form/scheduling-rules-section";

// Types
import type { SchedulingRules, NodeTypesPerWorkload } from "@/swagger-models/org-unit-service-client";
import { EQuotaEntity } from "@/models/resource.model";

// Mocks & Utils
import { provideUnsavedChangesGuard } from "@/composables/org-unit/use-unsaved-changes-guard.composable";
import { deepCopy } from "@/utils/common.util";

// ============================================================================
// Mock Setup
// ============================================================================

/**
 * We use async import inside the factory to solve the Vitest hoisting error.
 * This allows us to define the stub entirely within the mock factory
 * while still accessing Vue utilities.
 */
vi.mock("@/components/project/project-edit-form/scheduling-rules-section", async () => {
  const { defineComponent, h } = await import("vue");

  const SchedulingRulesSectionStub = defineComponent({
    name: "SchedulingRulesSectionStub",
    emits: ["update:schedulingRules", "update:nodeTypes", "save", "is-section-invalid"],
    props: {
      schedulingRules: { type: Object, default: null },
      nodeTypes: { type: Object, default: undefined },
      clusterId: { type: String, default: "" },
      entity: { type: String, default: "" },
      parentSchedulingRules: { type: Object, default: null },
      parentNodeTypes: { type: Object, default: undefined },
      loading: { type: Boolean, default: false },
      readOnly: { type: Boolean, default: false },
      defaultOpened: { type: Boolean, default: false },
      disableClosing: { type: Boolean, default: false },
      showSaveButton: { type: Boolean, default: false },
      preventTopPaddingWhenOpen: { type: Boolean, default: false },
    },
    setup(_, { slots }) {
      return () => h("div", slots.default ? slots.default() : []);
    },
  });

  return {
    SchedulingRulesSection: SchedulingRulesSectionStub,
  };
});

// ============================================================================
// Test Suite
// ============================================================================

describe("SchedulingRulesStep", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // ==========================================================================
  // Test Utilities
  // ==========================================================================

  interface IStepProps {
    entityType: EQuotaEntity;
    clusterId: string;
    schedulingRules?: SchedulingRules | null;
    nodeTypes?: NodeTypesPerWorkload;
    parentSchedulingRules?: SchedulingRules | null;
    parentNodeTypes?: NodeTypesPerWorkload;
    loading?: boolean;
    readOnly?: boolean;
  }

  interface CreateWrapperOptions {
    props?: Partial<IStepProps>;
  }

  type SavePayload = {
    schedulingRules: SchedulingRules | null;
    nodeTypes?: NodeTypesPerWorkload;
  };

  const defaultNodeTypes: NodeTypesPerWorkload = {
    workspace: ["4500007", "4500008"],
  };

  const defaultSchedulingRules: SchedulingRules = {
    trainingJobMaxIdleDurationSeconds: 300,
  };

  const parentSchedulingRulesDefault: SchedulingRules = {
    trainingJobMaxIdleDurationSeconds: 600,
  };

  const createDefaultProps = (overrides: Partial<IStepProps> = {}): IStepProps => ({
    entityType: EQuotaEntity.project,
    clusterId: "cluster-1",
    schedulingRules: deepCopy(defaultSchedulingRules),
    nodeTypes: deepCopy(defaultNodeTypes),
    parentSchedulingRules: deepCopy(parentSchedulingRulesDefault),
    loading: false,
    readOnly: false,
    ...overrides,
  });

  const createWrapper = (options: CreateWrapperOptions = {}) => {
    const stepProps = reactive<IStepProps>(createDefaultProps(options.props));

    // We use a TestHost to handle the `provideUnsavedChangesGuard` dependency
    // and to capture events emitted by the step easily.
    const TestHost = defineComponent({
      components: { SchedulingRulesStep },
      emits: ["save-scheduling-rules", "validation-change"],
      setup(_, { emit, expose }) {
        provideUnsavedChangesGuard();
        expose({ stepProps });

        return () =>
          h(SchedulingRulesStep, {
            ...stepProps,
            onSaveSchedulingRules: (payload: unknown) => emit("save-scheduling-rules", payload),
            onValidationChange: (value: boolean) => emit("validation-change", value),
          });
      },
    });

    const wrapper = mount(TestHost);

    // Expose the reactive props from the VM so tests can modify them
    const hostVm = wrapper.vm as unknown as { stepProps: IStepProps };

    return {
      wrapper,
      stepProps: hostVm.stepProps,
    };
  };

  // Helpers to locate specific components
  const getSchedulingRulesSection = (wrapper: VueWrapper<ComponentPublicInstance>) =>
    wrapper.findComponent(SchedulingRulesSection);

  type StepExpose = {
    isDirty: { value: boolean };
    discardChanges: () => void;
  };

  const getStepExpose = (wrapper: VueWrapper<ComponentPublicInstance>): StepExpose => {
    const component = wrapper.findComponent(SchedulingRulesStep);
    return component.vm.$.exposed as StepExpose;
  };

  // ==========================================================================
  // Initialization & Rendering Tests
  // ==========================================================================

  describe("Initialization & Rendering", () => {
    it("renders with correct entity type for project", () => {
      const { wrapper } = createWrapper({
        props: { entityType: EQuotaEntity.project },
      });

      const section = getSchedulingRulesSection(wrapper);
      expect(section.props("entity")).toBe(EQuotaEntity.project);
    });

    it("renders with correct entity type for department", () => {
      const { wrapper } = createWrapper({
        props: {
          entityType: EQuotaEntity.department,
          parentSchedulingRules: null,
        },
      });

      const section = getSchedulingRulesSection(wrapper);
      expect(section.props("entity")).toBe(EQuotaEntity.department);
    });

    it("merges parent scheduling rules when entity has none", () => {
      const parentRules: SchedulingRules = {
        trainingJobMaxIdleDurationSeconds: 900,
      };

      const { wrapper } = createWrapper({
        props: {
          schedulingRules: null,
          parentSchedulingRules: parentRules,
        },
      });

      const section = getSchedulingRulesSection(wrapper);
      expect(section.props("schedulingRules")).toEqual(parentRules);
    });

    it("passes cluster id correctly", () => {
      const { wrapper } = createWrapper({
        props: { clusterId: "test-cluster-123" },
      });

      const section = getSchedulingRulesSection(wrapper);
      expect(section.props("clusterId")).toBe("test-cluster-123");
    });

    it("passes parentNodeTypes to scheduling rules section", () => {
      const parentNodeTypes: NodeTypesPerWorkload = {
        training: ["nt-1"],
        workspace: ["nt-2"],
      };

      const { wrapper } = createWrapper({
        props: { parentNodeTypes },
      });

      const section = getSchedulingRulesSection(wrapper);
      expect(section.props("parentNodeTypes")).toEqual(parentNodeTypes);
    });
  });

  // ==========================================================================
  // State Management & Actions Tests
  // ==========================================================================

  describe("State Management & Actions", () => {
    it("emits save event with cloned payload and resets dirty flag", async () => {
      const { wrapper } = createWrapper({
        props: {
          schedulingRules: null,
          nodeTypes: undefined,
        },
      });
      // Let onMounted re-snapshot complete before making changes
      await nextTick();
      const section = getSchedulingRulesSection(wrapper);
      const step = getStepExpose(wrapper);

      const updatedRules: SchedulingRules = {
        trainingJobMaxIdleDurationSeconds: 1200,
      };

      // 1. Simulate change
      await section.vm.$emit("update:schedulingRules", updatedRules);
      await nextTick();
      expect(step.isDirty.value).toBe(true);

      // 2. Simulate save
      await section.vm.$emit("save");
      await nextTick();

      // 3. Assertions
      const emitted = wrapper.emitted("save-scheduling-rules");
      expect(emitted).toBeTruthy();
      const payload = emitted?.[0]?.[0] as SavePayload;
      expect(payload).toEqual({
        schedulingRules: updatedRules,
        nodeTypes: undefined,
        navigateAfterSave: false,
      });

      // Ensure payload is a clone, not reference
      expect(payload.schedulingRules).not.toBe(updatedRules);
      expect(step.isDirty.value).toBe(false);
    });

    it("reverts local changes when discard is invoked", async () => {
      const { wrapper } = createWrapper({
        props: {
          schedulingRules: null,
          nodeTypes: undefined,
        },
      });
      // Let onMounted re-snapshot complete before making changes
      await nextTick();
      const section = getSchedulingRulesSection(wrapper);
      const step = getStepExpose(wrapper);

      // 1. Make dirty
      await section.vm.$emit("update:schedulingRules", {
        trainingJobMaxIdleDurationSeconds: 50,
      });
      await nextTick();
      expect(step.isDirty.value).toBe(true);

      // 2. Discard
      step.discardChanges();
      await nextTick();

      // 3. Assertions
      expect(step.isDirty.value).toBe(false);
      // When schedulingRules is null and parentSchedulingRules is provided, merged value equals parent
      expect(section.props("schedulingRules")).toEqual(parentSchedulingRulesDefault);
    });

    it("includes node types in save payload and clones them", async () => {
      const updatedNodeTypes: NodeTypesPerWorkload = {
        training: ["4500007", "4500008"],
      };

      const { wrapper } = createWrapper({
        props: {
          schedulingRules: null,
          nodeTypes: undefined,
        },
      });
      // Let onMounted re-snapshot complete before making changes
      await nextTick();

      const section = getSchedulingRulesSection(wrapper);
      const step = getStepExpose(wrapper);

      await section.vm.$emit("update:nodeTypes", updatedNodeTypes);
      await nextTick();
      expect(step.isDirty.value).toBe(true);

      await section.vm.$emit("save");
      await nextTick();

      const payload = wrapper.emitted("save-scheduling-rules")?.[0]?.[0] as SavePayload;
      expect(payload.nodeTypes).toEqual(updatedNodeTypes);
      expect(payload.nodeTypes).not.toBe(updatedNodeTypes);
    });
  });
});
