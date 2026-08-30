import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import WorkloadCreationDrawer from "./workload-creation-drawer.vue";
import { WorkloadCreationDrawerEntities, ExtendedSections } from "@/models/workload-creation-drawer-model";
import { describe, expect, it, afterEach, vi, afterAll, beforeEach } from "vitest";
import { Quasar } from "quasar";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";

import { computeResourceService } from "@/services/control-plane/compute-resource.service/compute-resource.service";
import { workloadCreationDrawerMock } from "./workload-creation-drawer.mock";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { AssetKind } from "@/swagger-models/assets-service-client";

vi.mock("@/services/control-plane/compute-resource.service/compute-resource.service");

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("workload-creation-drawer.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof WorkloadCreationDrawer>>;

  const computeListSpy = vi
    .spyOn(computeResourceService, "list")
    .mockResolvedValue(workloadCreationDrawerMock.computeList);
  const dataListSpy = vi.spyOn(dataSourceService, "list").mockResolvedValue(workloadCreationDrawerMock.dataSourceList);

  const createComponent = (propsData = {}) => {
    wrapper = mount(WorkloadCreationDrawer, {
      plugins: [Quasar],
      props: {
        isOpen: true,
        complyTo: {},
        entity: WorkloadCreationDrawerEntities.Compute,
        ...propsData,
      },
      global: {
        stubs: {
          QDrawer: true,
        },
      },
    });
  };

  beforeEach(() => {
    createComponent();
  });

  afterEach(() => {
    wrapper.unmount();
  });

  afterAll(() => {
    computeListSpy.mockClear();
    dataListSpy.mockClear();
  });

  it("emits close event when closeDrawer is called", async () => {
    wrapper.vm.closeDrawer();
    expect(wrapper.emitted().close).toBeTruthy();
  });

  it("opens policy section correctly", async () => {
    wrapper.vm.openPolicySection("item1");
    expect(wrapper.vm.itemId).toBe("item1");
    expect(wrapper.vm.displayedSection).toBe(ExtendedSections.POLICY);
  });

  it("opens details section correctly", async () => {
    wrapper.vm.openDetailsSection("item2");
    expect(wrapper.vm.itemId).toBe("item2");
    expect(wrapper.vm.displayedSection).toBe(ExtendedSections.DETAILS);
  });

  it("resets selected items when resetSelectedItems is called", async () => {
    wrapper.vm.selectedItems = [{ id: "item1", name: "Item 1", kind: AssetKind.Git }];
    wrapper.vm.resetSelectedItems();
    expect(wrapper.vm.selectedItems).toEqual([]);
  });

  it("loads entities correctly", async () => {
    await wrapper.vm.loadEntities();
    expect(wrapper.vm.loadedItems[WorkloadCreationDrawerEntities.Compute]).toEqual(
      workloadCreationDrawerMock.computesItems,
    );
  });

  it("handles selection correctly", async () => {
    wrapper.vm.onSelection({ id: "item1", name: "Item 1", kind: AssetKind.Git });
    expect(wrapper.emitted().select).toBeTruthy();
  });

  it("handles multi-selection correctly", async () => {
    createComponent({ entity: WorkloadCreationDrawerEntities.DataSource });
    wrapper.vm.onSelection({ id: "item1", name: "Item 1", kind: AssetKind.Git });
    wrapper.vm.onSelection({ id: "item2", name: "Item 2", kind: AssetKind.Git });
    wrapper.vm.onMultiSelection();
    expect(wrapper.emitted()["select-multiple"]).toBeTruthy();
  });
});
