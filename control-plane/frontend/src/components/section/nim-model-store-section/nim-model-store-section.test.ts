import { describe, it, expect, vi } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import NimModelStoreSection from "./nim-model-store-section.vue";
import type { DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";
import type { INimModelStoreSection } from "./nim-model-store-section.model";

import { mockDataSources } from "./nim-model-store-section.mock";

import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("nim-model-store-section.vue", () => {
  const defaultProps = {
    dataSources: mockDataSources,
    clusterId: "17264c0b-fcb1-4f9c-99dc-56303e19da62",
    modelStoreData: {
      selectedDataSources: [],
    },
  };

  function getMountedComponent(props?: {
    modelStoreData?: INimModelStoreSection;
    clusterId?: string;
    dataSources?: Array<DatasourceListResponseEntry>;
    createdDataSourceId?: string;
    disableClosing?: boolean;
    defaultOpen?: boolean;
  }): VueWrapper<InstanceType<typeof NimModelStoreSection>> {
    return mount(NimModelStoreSection, {
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
    expect(wrapper.findComponent({ name: "runai-expansion-item" }).exists()).toBe(true);
    expect(wrapper.findComponent({ name: "nim-model-storage" }).exists()).toBe(true);
    expect(wrapper.html()).toContain("Model store");
  });

  it("computes `summary` correctly", () => {
    let wrapper = getMountedComponent({
      modelStoreData: { selectedDataSources: [] },
    });

    expect(wrapper.vm.summary).toBe("None");

    wrapper = getMountedComponent({
      modelStoreData: {
        selectedDataSources: [
          {
            id: mockDataSources[0].meta.id,
            kind: mockDataSources[0].meta.kind,
          },
        ],
      },
    });

    expect(wrapper.vm.summary).toBe(mockDataSources[0].meta.name);
  });
});
