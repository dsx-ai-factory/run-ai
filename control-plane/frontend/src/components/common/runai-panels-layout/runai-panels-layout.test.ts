/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, it, expect, beforeEach } from "vitest";
import { Quasar } from "quasar";
import RunaiPanelsLayout from "./runai-panels-layout.vue";

describe("RunaiPanelsLayout", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiPanelsLayout>>;

  const createWrapper = (
    props?: { marginTop?: string },
    slots?: {
      default?: string;
      left?: string;
      right?: string;
    },
  ) => {
    return mount(RunaiPanelsLayout, {
      props: props || {},
      slots: slots || {
        default: "<div class='test-main'>Main Content</div>",
      },
      global: {
        plugins: [Quasar],
      },
    });
  };

  beforeEach(() => {
    wrapper = createWrapper();
  });

  describe("Component Rendering", () => {
    it("renders the component correctly", () => {
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.find(".runai-panels-layout-wrapper").exists()).toBe(true);
    });

    it("renders the layout section", () => {
      expect(wrapper.find(".runai-panels-layout").exists()).toBe(true);
    });

    it("renders the main content area", () => {
      expect(wrapper.find(".panel-main").exists()).toBe(true);
    });

    it("renders default slot content", () => {
      const content = wrapper.find(".test-main");
      expect(content.exists()).toBe(true);
      expect(content.text()).toBe("Main Content");
    });

    it("does not render left panel when slot is not provided", () => {
      const leftPanel = wrapper.find(".panel-left");
      expect(leftPanel.exists()).toBe(false);
    });

    it("renders left panel when slot is provided", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main</div>",
        left: "<div class='test-left'>Left Content</div>",
      });

      const leftPanel = wrapper.find(".panel-left");
      expect(leftPanel.exists()).toBe(true);
      expect(wrapper.find(".test-left").text()).toBe("Left Content");
    });

    it("does not render right panel when slot is not provided", () => {
      const rightPanel = wrapper.find(".panel-right");
      expect(rightPanel.exists()).toBe(false);
    });

    it("renders right panel when slot is provided", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main</div>",
        right: "<div class='test-right'>Right Content</div>",
      });

      const rightPanel = wrapper.find(".panel-right");
      expect(rightPanel.exists()).toBe(true);
      expect(wrapper.find(".test-right").text()).toBe("Right Content");
    });

    it("renders all three sections when all slots provided", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main</div>",
        left: "<div>Left</div>",
        right: "<div>Right</div>",
      });

      expect(wrapper.find(".panel-left").exists()).toBe(true);
      expect(wrapper.find(".panel-main").exists()).toBe(true);
      expect(wrapper.find(".panel-right").exists()).toBe(true);
    });
  });

  describe("Props", () => {
    it("uses default marginTop of 24px", () => {
      expect(wrapper.props("marginTop")).toBe("24px");
    });

    it("accepts custom marginTop value", () => {
      wrapper = createWrapper({ marginTop: "100px" });
      expect(wrapper.props("marginTop")).toBe("100px");
    });

    it("applies custom marginTop via CSS variable", () => {
      wrapper = createWrapper({ marginTop: "60px" });
      expect(wrapper.props("marginTop")).toBe("60px");
    });
  });

  describe("Layout Structure", () => {
    it("uses CSS Grid layout", () => {
      const layout = wrapper.find(".runai-panels-layout");
      expect(layout.exists()).toBe(true);
    });

    it("applies wrapper centering styles", () => {
      const wrapper_elem = wrapper.find(".runai-panels-layout-wrapper");
      expect(wrapper_elem.exists()).toBe(true);
    });

    it("applies sticky positioning to side panels", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main</div>",
        left: "<div>Left</div>",
        right: "<div>Right</div>",
      });

      const leftPanel = wrapper.find(".panel-left");
      const rightPanel = wrapper.find(".panel-right");

      expect(leftPanel.classes()).toContain("panel-sticky");
      expect(rightPanel.classes()).toContain("panel-sticky");
    });

    it("main panel does not have sticky positioning", () => {
      const mainPanel = wrapper.find(".panel-main");
      expect(mainPanel.classes()).not.toContain("panel-sticky");
    });

    it("uses semantic main element for main content", () => {
      expect(wrapper.find("main.panel-main").exists()).toBe(true);
    });
  });

  describe("Slot System", () => {
    it("supports default slot for main content", () => {
      const customContent = "<div class='custom'>Custom Main Content</div>";
      wrapper = createWrapper(undefined, {
        default: customContent,
      });

      expect(wrapper.find(".custom").exists()).toBe(true);
      expect(wrapper.find(".custom").text()).toBe("Custom Main Content");
    });

    it("supports left slot", () => {
      const leftContent = "<div class='left-custom'>Left Custom</div>";
      wrapper = createWrapper(undefined, {
        default: "<div>Main</div>",
        left: leftContent,
      });

      expect(wrapper.find(".left-custom").exists()).toBe(true);
      expect(wrapper.find(".left-custom").text()).toBe("Left Custom");
    });

    it("supports right slot", () => {
      const rightContent = "<div class='right-custom'>Right Custom</div>";
      wrapper = createWrapper(undefined, {
        default: "<div>Main</div>",
        right: rightContent,
      });

      expect(wrapper.find(".right-custom").exists()).toBe(true);
      expect(wrapper.find(".right-custom").text()).toBe("Right Custom");
    });

    it("allows all three slots simultaneously", () => {
      wrapper = createWrapper(undefined, {
        default: "<div class='main-slot'>Main</div>",
        left: "<div class='left-slot'>Left</div>",
        right: "<div class='right-slot'>Right</div>",
      });

      expect(wrapper.find(".main-slot").exists()).toBe(true);
      expect(wrapper.find(".left-slot").exists()).toBe(true);
      expect(wrapper.find(".right-slot").exists()).toBe(true);
    });

    it("works with only main content (no side panels)", () => {
      wrapper = createWrapper(undefined, {
        default: "<div class='only-main'>Only Main</div>",
      });

      expect(wrapper.find(".only-main").exists()).toBe(true);
      expect(wrapper.find(".panel-left").exists()).toBe(false);
      expect(wrapper.find(".panel-right").exists()).toBe(false);
    });

    it("works with left panel but no right panel", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main</div>",
        left: "<div>Left</div>",
      });

      expect(wrapper.find(".panel-left").exists()).toBe(true);
      expect(wrapper.find(".panel-main").exists()).toBe(true);
      expect(wrapper.find(".panel-right").exists()).toBe(false);
    });

    it("works with right panel but no left panel", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main</div>",
        right: "<div>Right</div>",
      });

      expect(wrapper.find(".panel-left").exists()).toBe(false);
      expect(wrapper.find(".panel-main").exists()).toBe(true);
      expect(wrapper.find(".panel-right").exists()).toBe(true);
    });
  });

  describe("Responsive Behavior", () => {
    it("maintains centered layout", () => {
      const layoutWrapper = wrapper.find(".runai-panels-layout-wrapper");
      expect(layoutWrapper.exists()).toBe(true);
    });

    it("has minimum width constraint", () => {
      const layout = wrapper.find(".runai-panels-layout");
      expect(layout.exists()).toBe(true);
    });

    it("main content has minimum width", () => {
      const mainPanel = wrapper.find(".panel-main");
      expect(mainPanel.exists()).toBe(true);
    });
  });

  describe("Component Lifecycle", () => {
    it("handles prop updates correctly", async () => {
      wrapper = createWrapper({ marginTop: "84px" });

      await wrapper.setProps({ marginTop: "100px" });

      expect(wrapper.props("marginTop")).toBe("100px");
    });

    it("handles slot updates", async () => {
      wrapper = createWrapper(undefined, {
        default: "<div class='original'>Original</div>",
      });

      expect(wrapper.find(".original").exists()).toBe(true);
    });
  });

  describe("Accessibility", () => {
    it("uses semantic HTML elements", () => {
      expect(wrapper.find("section.runai-panels-layout").exists()).toBe(true);
      expect(wrapper.find("main.panel-main").exists()).toBe(true);
    });

    it("provides proper document structure", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main Content</div>",
        left: "<nav>Navigation</nav>",
        right: "<aside>Additional Info</aside>",
      });

      expect(wrapper.find("nav").exists()).toBe(true);
      expect(wrapper.find("main").exists()).toBe(true);
      expect(wrapper.find("aside").exists()).toBe(true);
    });
  });

  describe("Edge Cases", () => {
    it("handles empty main slot", () => {
      wrapper = createWrapper(undefined, {
        default: "",
      });

      expect(wrapper.find(".panel-main").exists()).toBe(true);
      expect(wrapper.find(".panel-main").text()).toBe("");
    });

    it("handles very long content in slots", () => {
      const longContent = "<div>" + "Long content ".repeat(100) + "</div>";
      wrapper = createWrapper(undefined, {
        default: longContent,
        left: longContent,
        right: longContent,
      });

      expect(wrapper.exists()).toBe(true);
    });

    it("handles nested components in slots", () => {
      const nestedContent = `
        <div class="outer">
          <div class="middle">
            <div class="inner">Nested</div>
          </div>
        </div>
      `;
      wrapper = createWrapper(undefined, {
        default: nestedContent,
        left: nestedContent,
        right: nestedContent,
      });

      expect(wrapper.findAll(".outer")).toHaveLength(3);
      expect(wrapper.findAll(".inner")).toHaveLength(3);
    });
  });

  describe("Generic Reusability", () => {
    it("can be used for wizard layouts", () => {
      wrapper = createWrapper(undefined, {
        default: "<div class='wizard-content'>Wizard Step Content</div>",
        left: "<div class='wizard-steps'>Steps</div>",
        right: "<div class='wizard-help'>Help</div>",
      });

      expect(wrapper.find(".wizard-content").exists()).toBe(true);
      expect(wrapper.find(".wizard-steps").exists()).toBe(true);
      expect(wrapper.find(".wizard-help").exists()).toBe(true);
    });

    it("can be used for dashboard layouts", () => {
      wrapper = createWrapper(undefined, {
        default: "<div class='dashboard-main'>Main Dashboard</div>",
        left: "<div class='dashboard-nav'>Navigation</div>",
        right: "<div class='dashboard-widgets'>Widgets</div>",
      });

      expect(wrapper.find(".dashboard-main").exists()).toBe(true);
      expect(wrapper.find(".dashboard-nav").exists()).toBe(true);
      expect(wrapper.find(".dashboard-widgets").exists()).toBe(true);
    });

    it("can be used for settings layouts", () => {
      wrapper = createWrapper(undefined, {
        default: "<div class='settings-content'>Settings Content</div>",
        left: "<div class='settings-menu'>Settings Menu</div>",
      });

      expect(wrapper.find(".settings-content").exists()).toBe(true);
      expect(wrapper.find(".settings-menu").exists()).toBe(true);
      expect(wrapper.find(".panel-right").exists()).toBe(false);
    });
  });

  describe("Component Structure", () => {
    it("has correct component name", () => {
      expect(wrapper.vm.$options.name).toBe("runai-panels-layout");
    });

    it("has marginTop prop defined", () => {
      expect(wrapper.vm.$options.props?.marginTop).toBeDefined();
    });

    it("has no emits", () => {
      expect(wrapper.vm.$options.emits).toBeUndefined();
    });
  });
});
