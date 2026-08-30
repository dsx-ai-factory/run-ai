import { mount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { describe, expect, it, vi } from "vitest";
import { AssetKind } from "@/swagger-models/assets-service-client";
import type { TUIDataSourceInstancesKinds } from "@/models/data-source.model";
import DataSourceDropdown from "./data-source-dropdown.vue";

function mountDropdown(
  allowedOptions?: Set<TUIDataSourceInstancesKinds>,
): VueWrapper<InstanceType<typeof DataSourceDropdown>> {
  return mount(DataSourceDropdown, {
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
    },
    props: allowedOptions === undefined ? {} : { allowedOptions },
  });
}

describe("data-source-dropdown", () => {
  it("shows all options when no allowed-options filter is provided", () => {
    const wrapper = mountDropdown();

    expect(wrapper.vm.options).toHaveLength(wrapper.vm.allOptions.length);
  });

  it("shows no options when policy explicitly allows no data source types", () => {
    const wrapper = mountDropdown(new Set<TUIDataSourceInstancesKinds>());

    expect(wrapper.vm.options).toEqual([]);
  });

  it("shows only the options included in the allowed-options filter", () => {
    const wrapper = mountDropdown(new Set<TUIDataSourceInstancesKinds>([AssetKind.Nfs]));

    expect(wrapper.vm.options.map((option) => option.id)).toEqual([AssetKind.Nfs]);
  });
});
