import { mount, shallowMount } from "@vue/test-utils";
import ComputeResourceSection from "./compute-resource-section.vue";
import { beforeEach, describe, expect, test, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { mockComputeResourceSectionData } from "./compute-resource-section.mock";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { usePermissionStore } from "@/stores/permissions.store";
import type { IComputeSectionData, IWorkloadComputeSectionOptions } from "./compute-resource-section.models";
import { EWorkloadFormType } from "@/models/workload.model";
import type { TTestEmit } from "@/models/unit-test.model";
import type { ComputeAsset } from "@/swagger-models/assets-service-client";
import { deepCopy } from "@/utils/common.util";
import { DEFAULT_THRESHOLD_VALUE } from "./auto-scale-section/threshold-metric-section/threshold-metric-section.model";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);
const permissionStore = usePermissionStore(pinia);

describe("compute-resource-section.vue", () => {
  let computeResourceSectionCmp: VueWrapper<InstanceType<typeof ComputeResourceSection>>;
  const props = {
    entityType: EWorkloadFormType.Workspace,
    loading: false as boolean,
    computeResources: mockComputeResourceSectionData.computeResources,
    nodeAffinity: [] as string[],
    computeResourceData: {
      computeResourceId: null,
      nodeType: null,
    } as IComputeSectionData,
  };

  beforeEach(() => {
    computeResourceSectionCmp = mount(ComputeResourceSection, {
      global: {
        plugins: [Quasar],
        directives: {
          clickOutside: vi.fn(),
        },
      },
      props,
      provide: {
        policy: {},
        isGPUOverProvisioningEnabled: false,
      },
    });
    vi.mocked(permissionStore.hasPermission).mockImplementation(() => true);
  });

  test("component is loading properly", () => {
    expect(computeResourceSectionCmp.html()).toContain("Compute resource");
    expect(computeResourceSectionCmp.classes()).toContain("compute-resource-section");
  });

  test("section summary should be 'None' when computeResourceId is falsy", () => {
    expect(computeResourceSectionCmp.vm.computeResourceSummary).toBe("None");
  });

  test("section summary should display the compute resource name when there is a selected compute resource", async () => {
    const selected: string = mockComputeResourceSectionData.computeResources[0].meta.id;
    const computeResourceData: IComputeSectionData = {
      computeResourceId: selected,
      nodeType: null,
    };
    await computeResourceSectionCmp.setProps({ computeResourceData });
    expect(computeResourceSectionCmp.vm.computeResourceSummary).toBe(
      mockComputeResourceSectionData.computeResources[0].meta.name,
    );
  });

  test("the component will emit the updated selected compute resource correctly when a new selection is made", async () => {
    const selectedCompute = mockComputeResourceSectionData.computeResources[1];
    // clicking the 2nd compute resource card (notice the selector):
    await computeResourceSectionCmp.find(`[aid=${selectedCompute.meta.name}-card]`).trigger("click");
    const expectedEmit: TTestEmit<IComputeSectionData> = [
      [{ computeResourceId: selectedCompute.meta.id, nodeType: null }],
    ];
    expect(computeResourceSectionCmp.emitted("compute-resource-data-changed")).toEqual(expectedEmit);
  });

  test("the component will emit 'null' when the selected compute resource is clicked again", async () => {
    const selectedCompute = mockComputeResourceSectionData.computeResources[1];
    // setting a selected compute resource:
    const computeResourceData: IComputeSectionData = {
      computeResourceId: selectedCompute.meta.id,
      nodeType: null,
    };
    await computeResourceSectionCmp.setProps({
      computeResourceData,
    });

    // clicking the right compute resource card (notice the selector):
    await computeResourceSectionCmp.find(`[aid=${selectedCompute.meta.name}-card]`).trigger("click");
    const expectedEmit: TTestEmit<IComputeSectionData> = [[{ computeResourceId: null, nodeType: null }]];
    expect(computeResourceSectionCmp.emitted("compute-resource-data-changed")).toEqual(expectedEmit);
  });

  test("the component will emit 'create-new' when compute resource creation button is clicked", async () => {
    expect(computeResourceSectionCmp.emitted("create-new")).toBeFalsy();
    await computeResourceSectionCmp.find("[aid=add-new-compute-resource-btn]").trigger("click");
    expect(computeResourceSectionCmp.emitted("create-new")).toBeTruthy();
  });

  test("the component will render the input for the node affinity when the nodeAffinity prop is empty", async () => {
    // Note: the default value of the nodeAffinity prop is an empty array ([])
    expect(computeResourceSectionCmp.find("[aid=node-affinity-input]").exists()).toBe(true);
    expect(computeResourceSectionCmp.find("[aid=node-affinity-select]").exists()).toBe(false);
  });

  test("the component will render the select for the node affinity when the nodeAffinity prop is empty", async () => {
    await computeResourceSectionCmp.setProps({
      nodeAffinity: ["test-aff-1", "test-aff-2"],
    });
    expect(computeResourceSectionCmp.find("[aid=node-affinity-input]").exists()).toBe(false);
    expect(computeResourceSectionCmp.find("[aid=node-affinity-select]").exists()).toBe(true);
  });

  test("the component will emit the correct node-affinity value when the INPUT value has changed", async () => {
    await computeResourceSectionCmp.find("[aid=node-affinity-input]").setValue("test-affinity");
    const expectedEmit: TTestEmit<IComputeSectionData> = [[{ computeResourceId: null, nodeType: "test-affinity" }]];
    expect(computeResourceSectionCmp.emitted("compute-resource-data-changed")).toEqual(expectedEmit);
  });

  test("the component will emit the correct node-affinity value when the SELECT value has changed", async () => {
    await computeResourceSectionCmp.setProps({
      nodeAffinity: ["test-aff-1", "test-aff-2"],
    });
    await (computeResourceSectionCmp.findComponent(".node-affinity-select") as VueWrapper).vm.$emit(
      "update:model-value",
      "test-aff-1",
    );
    const expectedEmit: TTestEmit<IComputeSectionData> = [[{ computeResourceId: null, nodeType: "test-aff-1" }]];
    expect(computeResourceSectionCmp.emitted("compute-resource-data-changed")).toEqual(expectedEmit);
  });

  test("sectionInvalid will return true if there are node-affinities and none was selected, and will return false if node-affinity is selected", async () => {
    const computeResourceId = "id-1-test"; // Validation also requires a compute to be selected
    const computeResourceData: IComputeSectionData = {
      computeResourceId,
      nodeType: null,
    };
    await computeResourceSectionCmp.setProps({
      nodeAffinity: ["test-aff-1", "test-aff-2"],
      computeResourceData,
    });

    expect(computeResourceSectionCmp.vm.sectionInvalid).toBe(true); // Note: by default none is selected
    // Setting a selected and then expecting the validation to pass
    await computeResourceSectionCmp.setProps({
      computeResourceData: {
        computeResourceId,
        nodeType: "test-aff-1",
      },
    });

    expect(computeResourceSectionCmp.vm.sectionInvalid).toBe(false);
  });

  test("section should display the compute resource describer details when entity type is inference", () => {
    const firstCompute: ComputeAsset = mockComputeResourceSectionData.computeResources[1];

    const props = {
      entityType: EWorkloadFormType.Inference,
      loading: false as boolean,
      computeResources: mockComputeResourceSectionData.computeResources,
      nodeAffinity: [] as string[],
      computeResourceData: {
        computeResourceId: firstCompute.meta.id,
        nodeType: null,
        autoScaleData: {
          minReplicas: 1,
          maxReplicas: 1,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: undefined,
        },
      } as IComputeSectionData,
      sectionOptions: {
        autoScale: true,
      } as IWorkloadComputeSectionOptions,
    };

    computeResourceSectionCmp = mount(ComputeResourceSection, {
      global: {
        plugins: [Quasar],
        directives: {
          clickOutside: vi.fn(),
        },
      },
      props,
      provide: {
        policy: {},
        isGPUOverProvisioningEnabled: false,
      },
    });

    vi.mocked(permissionStore.hasPermission).mockImplementation(() => true);

    expect(computeResourceSectionCmp.find(".compute-resource-describer").exists()).toBe(true);
    expect(computeResourceSectionCmp.find(".up-to-section").exists()).toBe(false);
  });

  test("section should display the compute resource describer details including up to when entity type is inference", () => {
    const firstCompute: ComputeAsset = mockComputeResourceSectionData.computeResources[1];

    const props = {
      entityType: EWorkloadFormType.Inference,
      loading: false as boolean,
      computeResources: mockComputeResourceSectionData.computeResources,
      nodeAffinity: [] as string[],
      computeResourceData: {
        computeResourceId: firstCompute.meta.id,
        nodeType: null,
        autoScaleData: {
          minReplicas: 1,
          maxReplicas: 4,
          thresholdMetric: "throughput",
          thresholdValue: DEFAULT_THRESHOLD_VALUE,
          scaleToZeroRetentionSeconds: undefined,
        },
      } as IComputeSectionData,
      sectionOptions: {
        autoScale: true,
      } as IWorkloadComputeSectionOptions,
    };

    computeResourceSectionCmp = mount(ComputeResourceSection, {
      global: {
        plugins: [Quasar],
        directives: {
          clickOutside: vi.fn(),
        },
      },
      props,
      provide: {
        policy: {},
        isGPUOverProvisioningEnabled: false,
      },
    });

    vi.mocked(permissionStore.hasPermission).mockImplementation(() => true);

    expect(computeResourceSectionCmp.find(".compute-resource-describer").exists()).toBe(true);
    expect(computeResourceSectionCmp.find(".up-to-section").exists()).toBe(true);
  });

  describe("getCardVisibilityState", () => {
    const DEFAULT_PROPS = {
      entityType: EWorkloadFormType.Inference,
      loading: false,
      nodeAffinity: [],
      computeResourceData: {
        computeResourceId: null,
        nodeType: null,
        autoScaleData: {
          minReplicas: 1,
          maxReplicas: 4,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: undefined,
        },
      },
      sectionOptions: {
        autoScale: true,
      },
    };

    function factory(
      computeResources: ComputeAsset[],
      props = {},
    ): VueWrapper<InstanceType<typeof ComputeResourceSection>> {
      const wrapper = shallowMount(ComputeResourceSection, {
        global: {
          plugins: [Quasar],
        },
        props: {
          ...DEFAULT_PROPS,
          computeResources: computeResources,
          ...props,
        },
      });

      wrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
      return wrapper;
    }

    test("should return default visibility state when there are no compliance issues and no memory requirement", () => {
      const firstCompute = deepCopy(mockComputeResourceSectionData.computeResources[0]);
      const secondCompute = deepCopy(mockComputeResourceSectionData.computeResources[1]);
      secondCompute.spec = {
        ...secondCompute.spec,
        gpuDevicesRequest: 1,
        gpuMemoryRequest: "16G",
        gpuRequestType: "memory",
      };
      const wrapper = factory([]);

      expect(wrapper.vm.getCardVisibilityState(firstCompute)).toEqual({
        disabled: false,
        tooltip: "",
        showDisabledInfo: false,
      });

      expect(wrapper.vm.getCardVisibilityState(secondCompute)).toEqual({
        disabled: false,
        tooltip: "",
        showDisabledInfo: false,
      });
    });

    test("should disable the card with appropriate tooltip when there are general compliance issues", () => {
      const computeResourceBase = deepCopy(mockComputeResourceSectionData.computeResources[0]);
      const computeResource = {
        ...computeResourceBase,
        compliance: {
          compliance: false,
          reason: [{ details: "Some compliance issue" }],
        },
      };

      const wrapper = factory([]);
      const result = wrapper.vm.getCardVisibilityState(computeResource);
      expect(result.disabled).toBe(true);
      expect(result.tooltip).toBe("Some compliance issue");
      expect(result.showDisabledInfo).toBe(false);
    });

    test("should disable the card with appropriate tooltip when there are compliance issues to specific field", () => {
      const computeResourceBase = deepCopy(mockComputeResourceSectionData.computeResources[0]);
      const computeResource = {
        ...computeResourceBase,
        compliance: {
          compliance: false,
          reason: [{ field: "memory", details: "Some compliance issue" }],
        },
      };

      const wrapper = factory([]);
      const result = wrapper.vm.getCardVisibilityState(computeResource);
      expect(result.disabled).toBe(true);
      expect(result.tooltip).not.toBe("Some compliance issue");
      expect(result.tooltip).toBeTruthy();
      expect(result.showDisabledInfo).toBe(true);
    });
  });
});
