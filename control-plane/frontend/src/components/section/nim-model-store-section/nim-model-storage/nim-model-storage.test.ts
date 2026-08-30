import { describe, it, expect, vi } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import NimModelStorage from "./nim-model-storage.vue";
import { mockDataSources, unsupportedDatasources } from "./nim-model-storage.mock";
import type { AssetIdAndKind, DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";
import { errorMessages } from "@/common/error-message.constant";
import { type ICardListItem, TCardCmpName } from "@/components/common/runai-card-list";

// stores
import { useClusterStore } from "@/stores/cluster.store";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("nim-model-storage.vue", () => {
  const clusterStore = useClusterStore(pinia);
  const defaultProps = {
    dataSources: mockDataSources,
    selectedDataSources: [],
    clusterId: "test-cluster-id",
    loading: false,
    required: false,
  };

  function getMountedComponent(
    props: {
      dataSources?: Array<DatasourceListResponseEntry>;
      selectedDataSources?: Array<AssetIdAndKind>;
      clusterId?: string;
      loading?: boolean;
      required?: boolean;
    } = {},
    isClusterVersionSupported = true,
  ): VueWrapper<InstanceType<typeof NimModelStorage>> {
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(isClusterVersionSupported);

    return mount(NimModelStorage, {
      global: {
        plugins: [Quasar],
        directives: {
          clickOutside: vi.fn(),
        },
      },
      provide: {
        policy: {},
      },
      props: {
        ...defaultProps,
        ...props,
      },
    });
  }

  it("renders the component correctly", () => {
    const wrapper = getMountedComponent();
    expect(wrapper.find(".gallery-message").exists()).toBe(false);
  });

  it("renders the component correctly", async () => {
    const wrapper = getMountedComponent({ required: true });
    expect(wrapper.find(".gallery-message").exists()).toBe(true);
  });

  it("computes `shouldBeOpened` correctly depends", () => {
    const datasource = mockDataSources[0];
    let wrapper = getMountedComponent({ selectedDataSources: [{ id: datasource.meta.id, kind: datasource.meta.kind }] });
    expect(wrapper.vm.shouldBeOpened).toBe(true); // selectedDataSources.length > 0

    wrapper = getMountedComponent({ required: false });
    expect(wrapper.vm.shouldBeOpened).toBe(false);

    wrapper = getMountedComponent({ required: true });
    expect(wrapper.vm.shouldBeOpened).toBe(true);
  });

  it("filters `dataSources` correctly", () => {
    const wrapper = getMountedComponent({ dataSources: [...mockDataSources, ...unsupportedDatasources] });
    expect(wrapper.vm.filteredDataSources).toEqual(mockDataSources);
  });

  it("emits `datasource-changed` when a data source is changed", async () => {
    const datasource = mockDataSources[0];
    const wrapper = getMountedComponent();

    const dataSourceGallery = wrapper.findComponent({ name: "data-source-gallery" });
    await dataSourceGallery.vm.$emit("datasource-changed", [{ id: datasource.meta.id, kind: datasource.meta.kind }]);

    expect(wrapper.emitted()["datasource-changed"]).toBeTruthy();
    expect(wrapper.emitted()["datasource-changed"][0]).toEqual([
      [{ id: datasource.meta.id, kind: datasource.meta.kind }],
    ]);
  });

  it("emits `create-new` when a new data source is created", async () => {
    const wrapper = getMountedComponent();
    const dataSourceGallery = wrapper.findComponent({ name: "data-source-gallery" });
    await dataSourceGallery.vm.$emit("create-new");

    expect(wrapper.emitted()["create-new"]).toBeTruthy();
  });

  it("renders `q-field` for required validation when `required` is true", () => {
    let wrapper = getMountedComponent({ required: true });
    expect(wrapper.find(".hidden-field").exists()).toBe(true);

    wrapper = getMountedComponent({ required: false });
    expect(wrapper.find(".hidden-field").exists()).toBe(false);
  });

  it("validates `isRequired` correctly", () => {
    const wrapper = getMountedComponent();
    const datasource = mockDataSources[0];
    const validInput: Array<ICardListItem> = [
      { id: datasource.meta.id, cardName: TCardCmpName.DATA_SOURCE, data: datasource },
    ];
    const invalidInput: Array<ICardListItem> = [];

    expect(wrapper.vm.isRequired(validInput)).toBe(true);
    expect(wrapper.vm.isRequired(invalidInput)).toBe(errorMessages.SELECT_DATA_SOURCE);
  });
});
