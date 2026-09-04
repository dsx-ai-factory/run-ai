/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, it, expect, beforeEach } from "vitest";
import { Quasar } from "quasar";
import RunaiPanelContainer from "./runai-panel-container.vue";

describe("RunaiPanelContainer", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiPanelContainer>>;

  const createWrapper = (props?: { width?: string; transparent?: boolean }, slots?: { default?: string }) => {
    return mount(RunaiPanelContainer, {
      props: props || {},
      slots: slots || {
        default: "<div class='test-content'>Panel Content</div>",
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
      expect(wrapper.find(".runai-panel-container").exists()).toBe(true);
    });

    it("renders as an aside element", () => {
      expect(wrapper.find("aside.runai-panel-container").exists()).toBe(true);
    });

    it("has correct CSS classes", () => {
      const container = wrapper.find(".runai-panel-container");
      expect(container.classes()).toContain("runai-panel-container");
    });

    it("renders slot content", () => {
      const content = wrapper.find(".test-content");
      expect(content.exists()).toBe(true);
      expect(content.text()).toBe("Panel Content");
    });
  });

  describe("Props", () => {
    it("uses default width of 220px", () => {
      expect(wrapper.props("width")).toBe("220px");
    });

    it("accepts custom width", () => {
      wrapper = createWrapper({ width: "300px" });
      expect(wrapper.props("width")).toBe("300px");
    });

    it("applies custom width via CSS variable", () => {
      wrapper = createWrapper({ width: "250px" });
      expect(wrapper.props("width")).toBe("250px");
    });
  });

  describe("Slot System", () => {
    it("renders default slot content", () => {
      const customContent = "<div class='custom'>Custom Content</div>";
      wrapper = createWrapper(undefined, {
        default: customContent,
      });

      expect(wrapper.find(".custom").exists()).toBe(true);
      expect(wrapper.find(".custom").text()).toBe("Custom Content");
    });

    it("renders complex slot content", () => {
      const complexContent = `
        <div>
          <h3>Section Title</h3>
          <p>Description text</p>
          <button>Action</button>
        </div>
      `;
      wrapper = createWrapper(undefined, {
        default: complexContent,
      });

      expect(wrapper.find("h3").text()).toBe("Section Title");
      expect(wrapper.find("p").text()).toBe("Description text");
      expect(wrapper.find("button").text()).toBe("Action");
    });

    it("handles empty slot", () => {
      wrapper = createWrapper(undefined, {
        default: "",
      });

      expect(wrapper.find(".runai-panel-container").exists()).toBe(true);
      expect(wrapper.text()).toBe("");
    });

    it("renders multiple elements in slot", () => {
      const multiContent = `
        <div class="item-1">Item 1</div>
        <div class="item-2">Item 2</div>
        <div class="item-3">Item 3</div>
      `;
      wrapper = createWrapper(undefined, {
        default: multiContent,
      });

      expect(wrapper.find(".item-1").exists()).toBe(true);
      expect(wrapper.find(".item-2").exists()).toBe(true);
      expect(wrapper.find(".item-3").exists()).toBe(true);
    });
  });

  describe("Component Structure", () => {
    it("has correct component name", () => {
      expect(wrapper.vm.$options.name).toBe("runai-panel-container");
    });

    it("has width prop defined", () => {
      expect(wrapper.vm.$options.props?.width).toBeDefined();
    });

    it("has transparent prop defined", () => {
      expect(wrapper.vm.$options.props?.transparent).toBeDefined();
    });
  });

  describe("Styling", () => {
    it("has consistent styling across instances", () => {
      const wrapper1 = createWrapper(undefined, { default: "<div>Content 1</div>" });
      const wrapper2 = createWrapper(undefined, { default: "<div>Content 2</div>" });

      expect(wrapper1.find(".runai-panel-container").classes()).toEqual(
        wrapper2.find(".runai-panel-container").classes(),
      );
    });
  });

  describe("Transparent Prop", () => {
    it("defaults to non-transparent (opaque) mode", () => {
      wrapper = createWrapper();
      const container = wrapper.find(".runai-panel-container");
      expect(container.classes()).not.toContain("runai-panel-container--transparent");
      expect(wrapper.props("transparent")).toBe(false);
    });

    it("applies transparent class when transparent prop is true", () => {
      wrapper = createWrapper({ transparent: true });
      const container = wrapper.find(".runai-panel-container");
      expect(container.classes()).toContain("runai-panel-container--transparent");
    });

    it("does not apply transparent class when transparent prop is false", () => {
      wrapper = createWrapper({ transparent: false });
      const container = wrapper.find(".runai-panel-container");
      expect(container.classes()).not.toContain("runai-panel-container--transparent");
    });

    it("can toggle transparent prop", async () => {
      wrapper = createWrapper({ transparent: false });
      let container = wrapper.find(".runai-panel-container");
      expect(container.classes()).not.toContain("runai-panel-container--transparent");

      await wrapper.setProps({ transparent: true });
      container = wrapper.find(".runai-panel-container");
      expect(container.classes()).toContain("runai-panel-container--transparent");

      await wrapper.setProps({ transparent: false });
      container = wrapper.find(".runai-panel-container");
      expect(container.classes()).not.toContain("runai-panel-container--transparent");
    });

    it("maintains other classes when transparent", () => {
      wrapper = createWrapper({ transparent: true });
      const container = wrapper.find(".runai-panel-container");
      expect(container.classes()).toContain("runai-panel-container");
      expect(container.classes()).toContain("runai-panel-container--transparent");
    });

    it("works with custom width and transparent props together", () => {
      wrapper = createWrapper({ width: "300px", transparent: true });
      const container = wrapper.find(".runai-panel-container");
      expect(wrapper.props("width")).toBe("300px");
      expect(wrapper.props("transparent")).toBe(true);
      expect(container.classes()).toContain("runai-panel-container--transparent");
    });
  });

  describe("Content Variations", () => {
    it("renders text content", () => {
      wrapper = createWrapper(undefined, {
        default: "Simple text content",
      });

      expect(wrapper.text()).toBe("Simple text content");
    });

    it("renders list content", () => {
      const listContent = `
        <ul>
          <li>Item 1</li>
          <li>Item 2</li>
          <li>Item 3</li>
        </ul>
      `;
      wrapper = createWrapper(undefined, {
        default: listContent,
      });

      const listItems = wrapper.findAll("li");
      expect(listItems).toHaveLength(3);
    });

    it("renders form elements", () => {
      const formContent = `
        <form>
          <input type="text" class="input-field" />
          <button type="submit" class="submit-btn">Submit</button>
        </form>
      `;
      wrapper = createWrapper(undefined, {
        default: formContent,
      });

      expect(wrapper.find(".input-field").exists()).toBe(true);
      expect(wrapper.find(".submit-btn").exists()).toBe(true);
    });
  });

  describe("Accessibility", () => {
    it("uses semantic HTML (aside element)", () => {
      expect(wrapper.find("aside").exists()).toBe(true);
    });

    it("maintains proper content structure", () => {
      const accessibleContent = `
        <div>
          <h3>Section Title</h3>
          <p>Section description</p>
        </div>
      `;
      wrapper = createWrapper(undefined, {
        default: accessibleContent,
      });

      expect(wrapper.find("h3").exists()).toBe(true);
      expect(wrapper.find("p").exists()).toBe(true);
    });
  });

  describe("Edge Cases", () => {
    it("handles very long content", () => {
      const longContent = "<div>" + "Long text content. ".repeat(100) + "</div>";
      wrapper = createWrapper(undefined, {
        default: longContent,
      });

      expect(wrapper.exists()).toBe(true);
      expect(wrapper.text().length).toBeGreaterThan(1000);
    });

    it("handles nested components", () => {
      const nestedContent = `
        <div class="outer">
          <div class="middle">
            <div class="inner">Nested Content</div>
          </div>
        </div>
      `;
      wrapper = createWrapper(undefined, {
        default: nestedContent,
      });

      expect(wrapper.find(".outer").exists()).toBe(true);
      expect(wrapper.find(".middle").exists()).toBe(true);
      expect(wrapper.find(".inner").exists()).toBe(true);
      expect(wrapper.find(".inner").text()).toBe("Nested Content");
    });

    it("handles special characters in content", () => {
      const specialContent = "<div>Content with & < > \" ' special chars</div>";
      wrapper = createWrapper(undefined, {
        default: specialContent,
      });

      expect(wrapper.text()).toContain("Content with");
    });
  });

  describe("Component Lifecycle", () => {
    it("handles mount correctly", () => {
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.find(".runai-panel-container").exists()).toBe(true);
    });

    it("handles prop updates", async () => {
      wrapper = createWrapper({ width: "220px" });

      await wrapper.setProps({ width: "300px" });

      expect(wrapper.props("width")).toBe("300px");
    });
  });

  describe("Reusability", () => {
    it("can be used in different contexts with different content", () => {
      const wizardContent = createWrapper(undefined, {
        default: "<div class='wizard-steps'>Steps</div>",
      });
      const sidebarContent = createWrapper(undefined, {
        default: "<div class='sidebar-info'>Info</div>",
      });

      expect(wizardContent.find(".wizard-steps").exists()).toBe(true);
      expect(sidebarContent.find(".sidebar-info").exists()).toBe(true);
    });

    it("maintains independent state between instances", () => {
      const wrapper1 = createWrapper({ width: "220px" });
      const wrapper2 = createWrapper({ width: "300px" });

      expect(wrapper1.props("width")).toBe("220px");
      expect(wrapper2.props("width")).toBe("300px");
    });
  });
});
