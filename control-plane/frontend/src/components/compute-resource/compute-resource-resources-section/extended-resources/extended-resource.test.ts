import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import ExtendedResources from "./extended-resources.vue";
import { afterEach, describe, expect, test, vi } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import type { ExtendedResource } from "@/swagger-models/assets-service-client";
const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

describe("compute-resource-resources-section.vue", () => {
  let extendedResourcesCmp: VueWrapper<InstanceType<typeof ExtendedResources>>;

  const factory = (extendedResources: ExtendedResource[] = []) => {
    extendedResourcesCmp = mount(ExtendedResources, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        extendedResources,
      },
    });
  };

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    factory();
    expect(extendedResourcesCmp.html()).toContain("Set extended resource(s)");
    expect(extendedResourcesCmp.classes()).toContain("extended-resources");
  });

  test("emits the 'update:model-value' event with correct payload when add() is called", async () => {
    factory();
    await extendedResourcesCmp.find(".add-button").trigger("click");
    const expectedEmits: ExtendedResource[][][] = [[[{ resource: "", quantity: "" }]]];
    expect(extendedResourcesCmp.emitted("update:model-value")).toEqual(expectedEmits);
  });

  test("emits the 'update:model-value' event with correct payload when remove() is called", async () => {
    const extendedResources = [
      { resource: "CPU", quantity: "1" },
      { resource: "Memory", quantity: "512Mi" },
    ];
    factory(extendedResources);
    expect(extendedResourcesCmp.vm.extendedResources.length).toBe(2);
    await extendedResourcesCmp.findAll(".close-button")[0].trigger("click"); // deleting the first

    const expectedEmits: ExtendedResource[][][] = [[[{ resource: "Memory", quantity: "512Mi" }]]];
    expect(extendedResourcesCmp.emitted("update:model-value")).toEqual(expectedEmits);
  });

  test("updates the model correctly when updateModel() is called on a resource change", async () => {
    factory([{ resource: "", quantity: "" }]); // adding a single extended resource
    const resourceInput = extendedResourcesCmp.findAllComponents(".input-name")[0] as VueWrapper; // selecting the first one
    await resourceInput.vm.$emit("update:model-value", "runai-test/gpu_1");

    const expectedEmits: ExtendedResource[][][] = [[[{ resource: "runai-test/gpu_1", quantity: "" }]]];
    expect(extendedResourcesCmp.emitted("update:model-value")).toEqual(expectedEmits);
  });

  test("updates the model correctly when updateModel() is called on a quantity change", async () => {
    factory([{ resource: "", quantity: "" }]); // adding a single extended resource
    const quantityInput = extendedResourcesCmp.findAllComponents(".quantity-input")[0] as VueWrapper; // selecting the first one
    await quantityInput.vm.$emit("update:model-value", "10m");

    const expectedEmits: ExtendedResource[][][] = [[[{ resource: "", quantity: "10m" }]]];
    expect(extendedResourcesCmp.emitted("update:model-value")).toEqual(expectedEmits);
  });
});
