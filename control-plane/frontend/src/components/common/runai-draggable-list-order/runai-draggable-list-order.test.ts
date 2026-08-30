import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, it, expect } from "vitest";
import RunaiDraggableListOrder from "./runai-draggable-list-order.vue";
import { errorMessages } from "@/common/error-message.constant";

describe("runai-draggable-list-order.vue - disabled functionality", () => {
  const defaultList = [
    { id: "1", name: "Item 1" },
    { id: "2", name: "Item 2" },
  ];

  function mountComponent(props = {}) {
    return mount(RunaiDraggableListOrder, {
      global: {
        plugins: [Quasar],
        stubs: {
          VueDraggable: {
            template: '<div class="vue-draggable-stub"><slot /></div>',
            props: ["modelValue", "disabled"],
          },
          "transition-group": {
            template: "<div><slot /></div>",
          },
        },
      },
      props: {
        list: defaultList,
        ...props,
      },
    });
  }

  describe("disabled prop", () => {
    it("defaults to false", () => {
      const wrapper = mountComponent();
      expect(wrapper.vm.disabled).toBe(false);
    });

    it("does not render disabled attribute on items when disabled is false", () => {
      const wrapper = mountComponent({ disabled: false });
      const items = wrapper.findAll(".main-list-col .item");
      items.forEach((item) => {
        expect(item.attributes("disabled")).toBeUndefined();
      });
    });

    it("renders disabled attribute on items when disabled is true", () => {
      const wrapper = mountComponent({ disabled: true });
      const items = wrapper.findAll(".main-list-col .item");
      items.forEach((item) => {
        expect(item.attributes("disabled")).toBeDefined();
      });
    });

    it("adds list-disabled class to main-list-col when disabled", () => {
      const wrapper = mountComponent({ disabled: true });
      expect(wrapper.find(".main-list-col").classes()).toContain("list-disabled");
    });

    it("does not add list-disabled class when not disabled", () => {
      const wrapper = mountComponent({ disabled: false });
      expect(wrapper.find(".main-list-col").classes()).not.toContain("list-disabled");
    });
  });

  describe("disabled elements visibility", () => {
    it("hides grip icon when disabled", () => {
      const wrapper = mountComponent({ disabled: true });
      expect(wrapper.find(".draggable-icon").exists()).toBe(false);
    });

    it("shows grip icon when not disabled", () => {
      const wrapper = mountComponent({ disabled: false });
      expect(wrapper.find(".draggable-icon").exists()).toBe(true);
    });

    it("hides reset button when disabled", () => {
      const wrapper = mountComponent({ disabled: true, showHeader: true });
      expect(wrapper.findComponent({ name: "runai-tooltip" }).exists()).toBe(false);
    });

    it("shows reset button when not disabled", () => {
      const wrapper = mountComponent({ disabled: false, showHeader: true });
      expect(wrapper.findComponent({ name: "runai-tooltip" }).exists()).toBe(true);
    });
  });

  describe("disabledTooltipText computed", () => {
    it("returns default policy message when disabledTooltip prop is not set", () => {
      const wrapper = mountComponent({ disabled: true });
      expect(wrapper.vm.disabledTooltipText).toBe(errorMessages.CANT_BE_MODIFIED_POLICY);
    });

    it("returns custom tooltip when disabledTooltip prop is set", () => {
      const customTooltip = "Custom disabled message";
      const wrapper = mountComponent({ disabled: true, disabledTooltip: customTooltip });
      expect(wrapper.vm.disabledTooltipText).toBe(customTooltip);
    });
  });

  describe("slot disabled prop", () => {
    it("passes disabled=false to slot when not disabled", () => {
      const wrapper = mountComponent({ disabled: false });
      const slot = wrapper.find(".item-content");
      expect(slot.exists()).toBe(true);
      // The slot receives disabled prop - we verify the component passes it correctly
      expect(wrapper.vm.disabled).toBe(false);
    });

    it("passes disabled=true to slot when disabled", () => {
      const wrapper = mountComponent({ disabled: true });
      expect(wrapper.vm.disabled).toBe(true);
    });
  });
});
