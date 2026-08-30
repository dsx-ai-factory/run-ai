import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, expect, it, beforeEach, vi } from "vitest";
import { Quasar } from "quasar";
import { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import WorkloadCreationDrawerContent from "./workload-creation-drawer-content.vue";
import { workloadCreationDrawerContenatMock } from "./workload-creation-drawer-content.mock";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("workload-creation-drawer-content", () => {
  let wrapper: VueWrapper<InstanceType<typeof WorkloadCreationDrawerContent>>;

  const factory = (props = {}) => {
    return mount(WorkloadCreationDrawerContent, {
      global: {
        plugins: [Quasar],
      },
      props: {
        entity: WorkloadCreationDrawerEntities.Environment,
        items: workloadCreationDrawerContenatMock.items,
        displayedItemId: "",
        loadingError: false,
        isMultiSelect: false,
        selectedItems: [],
        createdEntityId: null,
        ...props,
      },
    });
  };

  beforeEach(() => {
    wrapper = factory();
  });

  it("renders correctly", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("filters items based on search input", async () => {
    const input = wrapper.find("[aid='search-input-drawer']");
    await input.setValue(workloadCreationDrawerContenatMock.items[0].name);
    expect(wrapper.vm.filteredItems).toEqual([workloadCreationDrawerContenatMock.items[0]]);
  });

  it("sorts items based on selected sort option", async () => {
    await wrapper.setData({ sortBy: "name" });
    expect(wrapper.vm.filteredItems).toEqual([
      workloadCreationDrawerContenatMock.items[0],
      workloadCreationDrawerContenatMock.items[2],
      workloadCreationDrawerContenatMock.items[3],
      workloadCreationDrawerContenatMock.items[1],
    ]);
  });

  it('emits "create-new" event when create button is clicked', async () => {
    const button = wrapper.find(".add-new-btn");
    await button.trigger("click");
    expect(wrapper.emitted("create-new")).toBeTruthy();
  });

  it('emits "select" event when an item is clicked', async () => {
    const item = wrapper.findComponent({ name: "workload-drawer-item" });
    await item.vm.$emit("item-clicked", {
      id: workloadCreationDrawerContenatMock.items[2].id,
      name: workloadCreationDrawerContenatMock.items[2].name,
    });
    expect(wrapper.emitted("select")).toBeTruthy();
  });

  it("shows no data message when there are no filtered items", async () => {
    await wrapper.setData({ filter: "Non-existent item" });
    expect(wrapper.find(".list-no-data").exists()).toBe(true);
  });

  it("shows loading error message when loadingError is true", () => {
    wrapper = factory({ loadingError: true });

    expect(wrapper.find("[aid='loading-error']").text()).toContain(
      `For some reason environments can't be displayed right now.`,
    );
  });

  it('emits "multi-select" event when apply button is clicked in multi-select mode', async () => {
    wrapper = factory({ isMultiSelect: true, selectedItems: [{ id: "1", name: "Item 1" }] });
    const button = wrapper.find("[aid='multi-select-btn']");
    await button.trigger("click");
    expect(wrapper.emitted("multi-select")).toBeTruthy();
  });
});
