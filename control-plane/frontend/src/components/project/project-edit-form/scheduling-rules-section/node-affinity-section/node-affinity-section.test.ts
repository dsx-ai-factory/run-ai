import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import NodeAffinitySection from "./node-affinity-section.vue";
import { EWorkloadNodeAffinity } from "@/models/project.model";
import type { INodeAffinitySelectOption } from "@/models/project.model";
import type { NodeTypesPerWorkload } from "@/swagger-models/org-unit-service-client";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { RunaiAddableMultiSelect } from "@/components/common/runai-addable-multi-select/";
import type { ISelectOption } from "@/models/global.model";

// ============================================================================
// Mock Setup
// ============================================================================

vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

vi.mock("@/components/common/runai-addable-multi-select/", async () => {
  const { defineComponent, h } = await import("vue");

  const RunaiAddableMultiSelect = defineComponent({
    name: "RunaiAddableMultiSelect",
    props: {
      options: { type: Array, default: () => [] },
      selectedOptions: { type: Array, default: () => [] },
      loading: { type: Boolean, default: false },
      disable: { type: Boolean, default: false },
      required: { type: Boolean, default: false },
      supportAddingNewValue: { type: Boolean, default: true },
      label: { type: String, default: "" },
      newOptionLabel: { type: String, default: "" },
    },
    setup() {
      return () => h("div");
    },
  });

  return { RunaiAddableMultiSelect };
});

vi.mock("@/components/common/runai-select", async () => {
  const { defineComponent, h } = await import("vue");

  return {
    RunaiSelect: defineComponent({
      name: "RunaiSelect",
      props: {
        modelValue: { default: null },
        options: { default: () => [] },
        disable: { default: false },
        rules: { default: () => [] },
        label: { default: "" },
      },
      setup() {
        return () => h("div");
      },
    }),
  };
});

vi.mock("@/components/project/project-edit-form/cards/", async () => {
  const { defineComponent, h } = await import("vue");

  return {
    WrapperCard: defineComponent({
      name: "WrapperCard",
      setup(_, { slots }) {
        return () => h("div", [slots.body?.(), slots.actions?.()]);
      },
    }),
    GrayCard: defineComponent({
      name: "GrayCard",
      props: { hideCloseButton: { default: false } },
      setup(_, { slots }) {
        return () => h("div", slots.default?.());
      },
    }),
  };
});

const mockNodeTypes = [
  { id: "nt-1", name: "GPU-A100" },
  { id: "nt-2", name: "GPU-V100" },
  { id: "nt-3", name: "CPU-Only" },
];

// ============================================================================
// Test Suite
// ============================================================================

describe("NodeAffinitySection", () => {
  let wrapper: VueWrapper;

  const createWorkload = (overrides: Partial<INodeAffinitySelectOption> = {}): INodeAffinitySelectOption => ({
    label: "",
    value: "",
    loading: false,
    selectedTypes: [],
    ...overrides,
  });

  const mountComponent = async (
    props: {
      workloads?: INodeAffinitySelectOption[];
      clusterId?: string;
      parentNodeTypes?: NodeTypesPerWorkload;
      readOnly?: boolean;
    } = {},
  ) => {
    wrapper = mount(NodeAffinitySection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        workloads: props.workloads ?? [createWorkload()],
        clusterId: props.clusterId ?? "cluster-1",
        parentNodeTypes: props.parentNodeTypes,
        readOnly: props.readOnly ?? false,
      },
    });

    await flushPromises();
  };

  const getMultiSelect = () => wrapper.findComponent(RunaiAddableMultiSelect);

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(orgUnitService.getNodeTypes).mockResolvedValue(mockNodeTypes);
  });

  // ==========================================================================
  // Parent Node Types Filtering
  // ==========================================================================

  describe("Parent Node Types Filtering", () => {
    it("shows all node type options when no parent restriction", async () => {
      const workload = createWorkload({
        value: EWorkloadNodeAffinity.Train,
        label: "Training",
      });

      await mountComponent({ workloads: [workload] });

      const multiSelect = getMultiSelect();
      const options = multiSelect.props("options");
      expect(options).toHaveLength(3);
      expect(options?.map((o: ISelectOption) => o.value)).toEqual(["nt-1", "nt-2", "nt-3"]);
    });

    it("filters training workload options by parent training node types", async () => {
      const workload = createWorkload({
        value: EWorkloadNodeAffinity.Train,
        label: "Training",
      });

      const parentNodeTypes: NodeTypesPerWorkload = {
        training: ["nt-1", "nt-3"],
      };

      await mountComponent({ workloads: [workload], parentNodeTypes });

      const multiSelect = getMultiSelect();
      const options = multiSelect.props("options");
      expect(options).toHaveLength(2);
      expect(options?.map((o: ISelectOption) => o.value)).toEqual(["nt-1", "nt-3"]);
    });

    it("filters workspace workload options by parent workspace node types", async () => {
      const workload = createWorkload({
        value: EWorkloadNodeAffinity.Interactive,
        label: "Workspace",
      });

      const parentNodeTypes: NodeTypesPerWorkload = {
        workspace: ["nt-2"],
      };

      await mountComponent({ workloads: [workload], parentNodeTypes });

      const multiSelect = getMultiSelect();
      const options = multiSelect.props("options");
      expect(options).toHaveLength(1);
      expect(options?.[0]?.value).toBe("nt-2");
    });

    it("returns all options for unselected workload type even with parent restriction", async () => {
      const workload = createWorkload(); // no value set yet

      const parentNodeTypes: NodeTypesPerWorkload = {
        training: ["nt-1"],
        workspace: ["nt-2"],
      };

      await mountComponent({ workloads: [workload], parentNodeTypes });

      const multiSelect = getMultiSelect();
      const options = multiSelect.props("options");
      // No workload type selected yet, so all options should show
      expect(options).toHaveLength(3);
    });

    it("disables adding new values when parent has node types", async () => {
      const workload = createWorkload({
        value: EWorkloadNodeAffinity.Train,
        label: "Training",
      });

      const parentNodeTypes: NodeTypesPerWorkload = {
        training: ["nt-1"],
      };

      await mountComponent({ workloads: [workload], parentNodeTypes });

      const multiSelect = getMultiSelect();
      expect(multiSelect.props("supportAddingNewValue")).toBe(false);
    });

    it("allows adding new values when no parent node types", async () => {
      const workload = createWorkload({
        value: EWorkloadNodeAffinity.Train,
        label: "Training",
      });

      await mountComponent({ workloads: [workload] });

      const multiSelect = getMultiSelect();
      expect(multiSelect.props("supportAddingNewValue")).toBe(true);
    });

    it("allows adding new values when parent node types is empty", async () => {
      const workload = createWorkload({
        value: EWorkloadNodeAffinity.Train,
        label: "Training",
      });

      const parentNodeTypes: NodeTypesPerWorkload = {};

      await mountComponent({ workloads: [workload], parentNodeTypes });

      const multiSelect = getMultiSelect();
      expect(multiSelect.props("supportAddingNewValue")).toBe(true);
    });
  });
});
