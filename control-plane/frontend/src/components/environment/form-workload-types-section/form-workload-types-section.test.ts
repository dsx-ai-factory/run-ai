import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import FormWorkloadTypesSection from "./form-workload-types-section.vue";
import { WorkloadTypesSelection } from "./workload-types-selection";
import { EWorkloadName } from "@/models/workload.model";
import type { WorkloadSupportedTypes } from "@/swagger-models/assets-service-client";
import { useAppStore } from "@/stores/app.store";
import type { ITheme } from "@/models/theme.model";
import type { IWorkloadTypesOption } from "./workload-types-selection/workload-types-section.model";

const defaultProps = {
  workloadTypesSectionModel: {
    workspace: true,
    training: false,
    inference: false,
    distributed: false,
  } as WorkloadSupportedTypes,
  lockedWorkloadTypes: false,
  clusterId: "cluster-1",
};

const pinia = createTestingPinia({ createSpy: vi.fn });

describe("form-workload-types-section.vue", () => {
  beforeEach(() => {
    setActivePinia(pinia);
    useAppStore().$patch({ theme: { isWhiteLabel: false } as ITheme });
  });

  function mountComponent(
    props: Partial<typeof defaultProps> = {},
  ): VueWrapper<InstanceType<typeof FormWorkloadTypesSection>> {
    return mount(FormWorkloadTypesSection, {
      global: {
        plugins: [Quasar, pinia],
        stubs: {
          RunaiExpansionItem: { template: "<div><slot /></div>" },
          DistributedSelection: true,
          DistributedFrameworkSelectionSection: true,
        },
      },
      props: { ...defaultProps, ...props },
    });
  }

  it("renders workload types section", () => {
    const wrapper = mountComponent();
    expect(wrapper.find(".form-workload-types-section").exists()).toBe(true);
  });

  it("passes workload type options to WorkloadTypesSelection", () => {
    const wrapper = mountComponent();
    const selection = wrapper.findComponent(WorkloadTypesSelection);
    expect(selection.exists()).toBe(true);
    const options = selection.props("options") as IWorkloadTypesOption[];
    expect(Array.isArray(options)).toBe(true);
    expect(options.some((o) => o.value === EWorkloadName.Workspace)).toBe(true);
    expect(options.some((o) => o.value === EWorkloadName.Inference)).toBe(true);
  });

  describe("lockedWorkloadTypes with distributed inference", () => {
    it("disables training and inference with distributed-inference tooltips when locked and model is distributed+inference", () => {
      const wrapper = mountComponent({
        lockedWorkloadTypes: true,
        workloadTypesSectionModel: {
          workspace: false,
          training: false,
          inference: true,
          distributed: true,
        } as WorkloadSupportedTypes,
      });
      const selection = wrapper.findComponent(WorkloadTypesSelection);
      const options = selection.props("options") as Array<{
        value: string;
        disable?: boolean;
        toolTip?: string;
        toolTipWidth?: string;
      }>;
      const trainingOption = options.find((o) => o.value === EWorkloadName.Training);
      const inferenceOption = options.find((o) => o.value === EWorkloadName.Inference);
      const workspaceOption = options.find((o) => o.value === EWorkloadName.Workspace);
      expect(workspaceOption?.disable).toBe(true);
      expect(workspaceOption?.toolTip).toBe(
        "This can't be modified when an environment is created via a distributed workload.",
      );
      expect(trainingOption?.disable).toBe(true);
      expect(trainingOption?.toolTip).toBe(
        "This can't be modified when an environment is created via a distributed workload.",
      );
      expect(inferenceOption?.disable).toBe(true);
      expect(inferenceOption?.toolTip).toBe(
        "This can't be modified when an environment is created via a distributed workload.",
      );
      expect(inferenceOption?.toolTipWidth).toBe("300px");
    });

    it("does not disable training/inference with distributed-inference tooltips when locked but model is not distributed+inference", () => {
      const wrapper = mountComponent({
        lockedWorkloadTypes: true,
        workloadTypesSectionModel: {
          workspace: false,
          training: true,
          inference: false,
          distributed: false,
        } as WorkloadSupportedTypes,
      });
      const selection = wrapper.findComponent(WorkloadTypesSelection);
      const options = selection.props("options") as Array<{ value: string; disable?: boolean; toolTip?: string }>;
      const trainingOption = options.find((o) => o.value === EWorkloadName.Training);
      expect(trainingOption?.disable).toBe(true);
      expect(trainingOption?.toolTip).toBe(
        "This option must be checked when an environment is created via creation of training",
      );
      const inferenceOption = options.find((o) => o.value === EWorkloadName.Inference);
      expect(inferenceOption?.disable ?? false).toBe(false);
    });
  });

  describe("updateWorkloadTypes", () => {
    it("emits distFramework undefined when distributed and inference only selected", async () => {
      const wrapper = mountComponent({
        workloadTypesSectionModel: {
          workspace: false,
          training: false,
          inference: true,
          distributed: true,
          distFramework: "PyTorch",
        } as WorkloadSupportedTypes,
      });
      wrapper.vm.updateWorkloadTypes({
        [EWorkloadName.Workspace]: false,
        [EWorkloadName.Training]: false,
        [EWorkloadName.Inference]: true,
        [EWorkloadName.Distributed]: false,
      });
      const emitted = wrapper.emitted("update-workload-types-section-model");
      expect(emitted).toHaveLength(1);
      const payload = emitted?.[0]?.[0] as WorkloadSupportedTypes;
      expect(payload.distributed).toBe(true);
      expect(payload.inference).toBe(true);
      expect(payload.training).toBe(false);
      expect(payload.distFramework).toBeUndefined();
    });

    it("when distributed and multiple selected, emits only the newly selected workload type", async () => {
      const wrapper = mountComponent({
        workloadTypesSectionModel: {
          workspace: false,
          training: true,
          inference: false,
          distributed: true,
        } as WorkloadSupportedTypes,
      });
      wrapper.vm.updateWorkloadTypes({
        [EWorkloadName.Workspace]: false,
        [EWorkloadName.Training]: true,
        [EWorkloadName.Inference]: true,
        [EWorkloadName.Distributed]: false,
      });
      const emitted = wrapper.emitted("update-workload-types-section-model");
      const payload = emitted?.[0]?.[0] as WorkloadSupportedTypes;
      expect(payload.inference).toBe(true);
      expect(payload.training).toBe(false);
    });
  });

  describe("workload architecture states", () => {
    it("omits the empty framework suffix from the distributed summary", () => {
      const wrapper = mountComponent({
        workloadTypesSectionModel: {
          workspace: false,
          training: true,
          inference: false,
          distributed: true,
          distFramework: undefined,
        } as WorkloadSupportedTypes,
      });

      expect(wrapper.vm.summary).toBe("Distributed - Training");
    });
  });
});
