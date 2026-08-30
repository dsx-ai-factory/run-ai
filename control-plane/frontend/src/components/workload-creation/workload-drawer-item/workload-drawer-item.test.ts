import { mount } from "@vue/test-utils";
import WorkloadDrawerItem from "./workload-drawer-item.vue";
import type { IDrawerListItem } from "@/models/workload-creation-drawer-model";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import { AssetKind } from "@/swagger-models/assets-service-client";

describe("workload-drawer-item.vue", () => {
  const item: IDrawerListItem = {
    id: "1",
    name: "some-name",
    description: "lorem lorel lorem lorem",
    kind: AssetKind.Compute,
    policy: { compliance: true },
    icons: [{ name: "icon", icon: "s3" }],
    createdAt: "2021-01-01",
    lastUsed: "2021-01-01",
    disabled: false,
    details: [
      {
        title: "title",
        content: ["value1"],
      },
    ],
  };

  it("renders item name and description", () => {
    const wrapper = mount(WorkloadDrawerItem, {
      global: {
        plugins: [Quasar],
      },
      props: { item },
    });
    expect(wrapper.find(".item-label").text()).toBe(item.name);
    expect(wrapper.find(".item-description").text()).toBe(item.description);
  });

  it("emits item-clicked event when item is clicked", async () => {
    const wrapper = mount(WorkloadDrawerItem, {
      global: {
        plugins: [Quasar],
      },
      props: { item },
    });
    await wrapper.trigger("click");
    expect(wrapper.emitted("item-clicked")).toBeTruthy();
    expect(wrapper.emitted("item-clicked")?.[0]).toEqual([{ id: item.id, name: item.name, kind: AssetKind.Compute }]);
  });

  it("shows policy info button when policy reason is present", () => {
    const wrapper = mount(WorkloadDrawerItem, {
      global: {
        plugins: [Quasar],
      },
      props: {
        item: { ...item, policy: { compliance: false, reason: [{ field: "bla", details: "this aint working" }] } },
      },
    });
    expect(wrapper.find("[aid='workload-drawer-item-policy-button']").exists()).toBe(true);
  });

  it("emits policy-clicked event when policy button is clicked", async () => {
    const wrapper = mount(WorkloadDrawerItem, {
      global: {
        plugins: [Quasar],
      },
      props: {
        item: { ...item, policy: { compliance: false, reason: [{ field: "bla", details: "this aint working" }] } },
      },
    });
    await wrapper.find("[aid='workload-drawer-item-policy-button']").trigger("click");
    expect(wrapper.emitted("policy-clicked")).toBeTruthy();
    expect(wrapper.emitted("policy-clicked")?.[0]).toEqual([item.id]);
  });

  it("emits details-clicked event when details button is clicked", async () => {
    const wrapper = mount(WorkloadDrawerItem, {
      global: {
        plugins: [Quasar],
      },
      props: { item },
    });
    await wrapper.find("[aid='workload-drawer-item-details-button']").trigger("click");
    expect(wrapper.emitted("details-clicked")).toBeTruthy();
    expect(wrapper.emitted("details-clicked")?.[0]).toEqual([item.id]);
  });

  it("applies disabled class when item is disabled", () => {
    const disabledItem = { ...item, disabled: true };
    const wrapper = mount(WorkloadDrawerItem, {
      global: {
        plugins: [Quasar],
      },
      props: { item: disabledItem },
    });
    expect(wrapper.find(".disabled").exists()).toBe(true);
  });
});
