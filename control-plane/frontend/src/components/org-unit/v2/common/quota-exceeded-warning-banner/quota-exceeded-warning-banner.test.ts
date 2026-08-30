import { mount } from "@vue/test-utils";
import { describe, it, expect } from "vitest";
import { Quasar } from "quasar";
import QuotaExceededWarningBanner from "./quota-exceeded-warning-banner.vue";

describe("QuotaExceededWarningBanner", () => {
  const mountComponent = (props: { excessAmount: number; parentEntityType?: string; hideExcessAmount?: boolean }) => {
    return mount(QuotaExceededWarningBanner, {
      props,
      global: {
        plugins: [Quasar],
      },
    });
  };

  describe("Rendering", () => {
    it("renders the component correctly", () => {
      const wrapper = mountComponent({ excessAmount: 24 });
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.find(".quota-exceeded-warning-banner").exists()).toBe(true);
    });

    it("displays the warning icon", () => {
      const wrapper = mountComponent({ excessAmount: 10 });
      const icon = wrapper.find(".q-icon");
      expect(icon.exists()).toBe(true);
      expect(icon.classes()).toContain("ph-warning-circle");
    });

    it("displays the excess amount in the message", () => {
      const wrapper = mountComponent({ excessAmount: 24 });
      const text = wrapper.text();
      expect(text).toContain("24");
      expect(text).toContain("exceeds");
      expect(text).toContain("GPUs");
    });

    it("has orange warning styling", () => {
      const wrapper = mountComponent({ excessAmount: 15 });
      const banner = wrapper.find(".quota-exceeded-warning-banner");
      expect(banner.classes()).toContain("bg-orange-1");
      expect(banner.classes()).toContain("text-b-body1");
    });
  });

  describe("Content", () => {
    it("shows default parent entity type as 'department'", () => {
      const wrapper = mountComponent({ excessAmount: 10 });
      const text = wrapper.text();
      expect(text).toContain("department");
    });

    it("shows custom parent entity type when provided", () => {
      const wrapper = mountComponent({ excessAmount: 10, parentEntityType: "cluster" });
      const text = wrapper.text();
      expect(text).toContain("cluster");
      expect(text).not.toContain("department");
    });

    it("shows correct message for single GPU excess", () => {
      const wrapper = mountComponent({ excessAmount: 1 });
      const text = wrapper.text();
      expect(text).toContain("exceeds the department's quota by");
      expect(text).toContain("1");
      expect(text).toContain("GPUs");
    });

    it("shows correct message for multiple GPUs excess", () => {
      const wrapper = mountComponent({ excessAmount: 50 });
      const text = wrapper.text();
      expect(text).toContain("exceeds the department's quota by");
      expect(text).toContain("50");
      expect(text).toContain("GPUs");
    });
  });

  describe("Reactivity", () => {
    it("updates the displayed amount when prop changes", async () => {
      const wrapper = mountComponent({ excessAmount: 10 });
      expect(wrapper.text()).toContain("10");

      await wrapper.setProps({ excessAmount: 25 });
      expect(wrapper.text()).toContain("25");
      expect(wrapper.text()).not.toContain("10");
    });

    it("updates parent entity type when prop changes", async () => {
      const wrapper = mountComponent({ excessAmount: 10, parentEntityType: "department" });
      expect(wrapper.text()).toContain("department");

      await wrapper.setProps({ parentEntityType: "cluster" });
      expect(wrapper.text()).toContain("cluster");
      expect(wrapper.text()).not.toContain("department");
    });
  });

  describe("Hide Excess Amount", () => {
    it("hides excess amount when hideExcessAmount is true", () => {
      const wrapper = mountComponent({
        excessAmount: 24,
        parentEntityType: "node pool",
        hideExcessAmount: true,
      });
      const text = wrapper.text();
      expect(text).toContain("exceeds the node pool's available GPU's");
      expect(text).not.toContain("24");
      expect(text).not.toContain("by");
    });

    it("shows excess amount when hideExcessAmount is false", () => {
      const wrapper = mountComponent({
        excessAmount: 24,
        parentEntityType: "node pool",
        hideExcessAmount: false,
      });
      const text = wrapper.text();
      expect(text).toContain("exceeds the node pool's available GPU's by 24 GPUs");
    });

    it("defaults to showing excess amount when hideExcessAmount is not provided", () => {
      const wrapper = mountComponent({
        excessAmount: 15,
        parentEntityType: "department",
      });
      const text = wrapper.text();
      expect(text).toContain("by 15 GPUs");
    });
  });
});
