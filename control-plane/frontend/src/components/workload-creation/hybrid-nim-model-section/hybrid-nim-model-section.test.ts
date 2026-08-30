import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import HybridNimModelSection from "./hybrid-nim-model-section.vue";
import { storageUtil } from "@/utils/storage.util";
import { STORAGE_KEYS } from "@/models/storage.model";
import { useClusterStore } from "@/stores/cluster.store";
import type { IUISuperWorkloadSpec } from "@/models/workload.model";
import type { IScopeModel } from "@/models/global.model";
import { createPinia, setActivePinia } from "pinia";
import { ENimEnvVariableNames } from "@/models/inference.model";
import { EModelSource } from "./model-source-selector/model-source-selector.model";

// Mock the modules
vi.mock("@/utils/storage.util", () => ({
  storageUtil: {
    get: vi.fn(),
    save: vi.fn(),
  },
}));

describe("hybrid-nim-model-section.vue - showInfoBox logic", () => {
  let pinia: ReturnType<typeof createPinia>;
  let clusterStore: ReturnType<typeof useClusterStore>;

  const mockWorkloadSpec: IUISuperWorkloadSpec = {
    image: "",
    environmentVariables: [],
  } as IUISuperWorkloadSpec;

  const mockScopeModel: IScopeModel = {
    projectId: 123,
    departmentId: "test-dept",
  } as IScopeModel;

  const mockClusterId = "test-cluster-id";

  const createWrapper = (supportsDockerRegistry = false, forceHideInfoBox = false) => {
    return mount(HybridNimModelSection, {
      props: {
        workloadSpec: mockWorkloadSpec,
        clusterId: mockClusterId,
        scopeModel: mockScopeModel,
        sectionDisabled: false,
        supportsDockerRegistry,
        forceHideInfoBox,
      },
      global: {
        plugins: [pinia],
        stubs: {
          HybridSectionWrapper: true,
          ModelSourceSelector: true,
          NimModelAndAccess: true,
          HuggingFaceModelAndAccess: true,
          ServingEndpointSection: true,
          NimInfoBox: true,
        },
      },
    });
  };

  beforeEach(() => {
    vi.clearAllMocks();

    pinia = createPinia();
    setActivePinia(pinia);

    clusterStore = useClusterStore();

    clusterStore.isClusterVersionSufficient = vi.fn();
  });

  describe("when forceHideInfoBox is true", () => {
    it("should always set showInfoBox to false regardless of storage", () => {
      vi.mocked(storageUtil.get).mockReturnValue(undefined);

      const wrapper = createWrapper(false, true);

      expect(wrapper.vm.showInfoBox).toBe(false);
      expect(storageUtil.get).not.toHaveBeenCalled();
    });
  });

  describe("when supportsDockerRegistry is true", () => {
    it("should always set showInfoBox to false regardless of storage value (storage is true)", () => {
      vi.mocked(storageUtil.get).mockReturnValue(true);

      const wrapper = createWrapper(true);

      expect(wrapper.vm.showInfoBox).toBe(false);
      expect(storageUtil.get).not.toHaveBeenCalled();
    });

    it("should always set showInfoBox to false regardless of storage value (storage is false)", () => {
      vi.mocked(storageUtil.get).mockReturnValue(false);

      const wrapper = createWrapper(true);

      expect(wrapper.vm.showInfoBox).toBe(false);
      expect(storageUtil.get).not.toHaveBeenCalled();
    });

    it("should always set showInfoBox to false regardless of storage value (storage is undefined)", () => {
      vi.mocked(storageUtil.get).mockReturnValue(undefined);

      const wrapper = createWrapper(true);

      expect(wrapper.vm.showInfoBox).toBe(false);
      expect(storageUtil.get).not.toHaveBeenCalled();
    });
  });

  describe("when supportsDockerRegistry is false", () => {
    it("should set showInfoBox to false when NIM_INFO_BOX_CLOSED is true in storage", () => {
      vi.mocked(storageUtil.get).mockReturnValue(true);

      const wrapper = createWrapper(false);

      expect(wrapper.vm.showInfoBox).toBe(false);
      expect(storageUtil.get).toHaveBeenCalledWith(STORAGE_KEYS.NIM_INFO_BOX_CLOSED);
    });

    it("should set showInfoBox to true when NIM_INFO_BOX_CLOSED is false in storage", () => {
      vi.mocked(storageUtil.get).mockReturnValue(false);

      const wrapper = createWrapper(false);

      expect(wrapper.vm.showInfoBox).toBe(true);
      expect(storageUtil.get).toHaveBeenCalledWith(STORAGE_KEYS.NIM_INFO_BOX_CLOSED);
    });

    it("should set showInfoBox to true when NIM_INFO_BOX_CLOSED is not in storage", () => {
      vi.mocked(storageUtil.get).mockReturnValue(undefined);

      const wrapper = createWrapper(false);

      expect(wrapper.vm.showInfoBox).toBe(true);
      expect(storageUtil.get).toHaveBeenCalledWith(STORAGE_KEYS.NIM_INFO_BOX_CLOSED);
    });
  });
});

describe("hybrid-nim-model-section.vue - onModelSourceChanged", () => {
  let pinia: ReturnType<typeof createPinia>;

  const mockScopeModel: IScopeModel = {
    projectId: 123,
    departmentId: "test-dept",
  } as IScopeModel;

  const mockClusterId = "test-cluster-id";

  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    const clusterStore = useClusterStore();
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
  });

  it("removes NGC_API_KEY when switching from NIM to Hugging Face", async () => {
    const workloadSpec = {
      image: "nvcr.io/nim/meta/llama-3-1-8b-instruct:1.0.0",
      environmentVariables: [
        {
          name: ENimEnvVariableNames.NgcApiKey,
          secret: { name: "ngc-secret", key: ENimEnvVariableNames.NgcApiKey },
        },
        { name: ENimEnvVariableNames.NimServerPort, value: "8080" },
      ],
    } as IUISuperWorkloadSpec;

    const wrapper = mount(HybridNimModelSection, {
      props: {
        workloadSpec,
        clusterId: mockClusterId,
        scopeModel: mockScopeModel,
      },
      global: {
        plugins: [pinia],
        stubs: {
          HybridSectionWrapper: true,
          ModelSourceSelector: true,
          NimModelAndAccess: true,
          HuggingFaceModelAndAccess: true,
          ServingEndpointSection: true,
          NimInfoBox: true,
        },
      },
    });

    wrapper.vm.onModelSourceChanged(EModelSource.HuggingFace);

    const emitted = wrapper.emitted("update") as Array<[IUISuperWorkloadSpec]> | undefined;
    expect(emitted?.[0]?.[0].environmentVariables?.some((e) => e.name === ENimEnvVariableNames.NgcApiKey)).toBe(false);
  });
});
