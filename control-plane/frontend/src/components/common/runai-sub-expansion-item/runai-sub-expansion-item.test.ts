import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";

import RunaiSubExpansionItem from "./runai-sub-expansion-item.vue";

describe("runai-sub-expansion-item.vue", () => {
  const getMountedComponent = (props = {}): VueWrapper<InstanceType<typeof RunaiSubExpansionItem>> => {
    return mount(RunaiSubExpansionItem, {
      global: {
        plugins: [Quasar],
      },
      props: {
        label: "Test Label",
        ...props,
      },
    });
  };

  it("renders label correctly", () => {
    const wrapper = getMountedComponent();
    expect(wrapper.find(".runai-sub-expansion-header").text()).toBe("Test Label");
  });

  it("renders closed by default if defaultOpened is false", () => {
    const wrapper = getMountedComponent({ defaultOpened: false });
    expect(wrapper.vm.isOpen).toBe(false);
    expect(wrapper.find(".information-bar-container").exists()).toBe(true);
    expect(wrapper.find(".information-bar-container-text").exists()).toBe(false);
  });

  it("opens the expansion item when defaultOpened is true", () => {
    const wrapper = getMountedComponent({ defaultOpened: true });
    expect(wrapper.vm.isOpen).toBe(true);
    expect(wrapper.find(".information-bar-container-text").exists()).toBe(false);
  });

  it("toggles the isOpen state when header is clicked", async () => {
    const wrapper = getMountedComponent({ defaultOpened: false });
    const header = wrapper.find(".runai-sub-expansion-header");

    expect(wrapper.vm.isOpen).toBe(false);
    expect(wrapper.find(".information-bar-container-text").exists()).toBe(false);

    await header.trigger("click");
    expect(wrapper.vm.isOpen).toBe(true);
    expect(wrapper.find(".information-bar-container-text").exists()).toBe(false);

    await header.trigger("click");
    expect(wrapper.vm.isOpen).toBe(false);
    expect(wrapper.find(".information-bar-container-text").exists()).toBe(false);
  });

  it("renders NotificationAlertBar if informationBar is provided and isOpen is false", () => {
    const wrapper = getMountedComponent({
      defaultOpened: false,
      informationBar: "Information Bar Text",
    });

    expect(wrapper.find(".information-bar-container-text").text()).toBe("Information Bar Text");
  });

  it("does not render NotificationAlertBar when isOpen is true", async () => {
    const informationBar = "Information Bar Text";
    const wrapper = getMountedComponent({
      defaultOpened: true,
      informationBar,
    });

    expect(wrapper.find(".runai-sub-expansion-header").text()).toBe("Test Label");
    expect(wrapper.find(".information-bar-container-text").exists()).toBe(false);
  });

  it("sets the correct aid attribute on the header", () => {
    const wrapper = getMountedComponent({ aid: "unique-id" });
    const header = wrapper.find(".runai-sub-expansion-header");
    expect(header.attributes("aid")).toBe("unique-id");
  });
});
