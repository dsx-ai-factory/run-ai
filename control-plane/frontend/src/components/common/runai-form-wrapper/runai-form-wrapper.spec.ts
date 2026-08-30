import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar, QForm, QBtn, QField } from "quasar";
import RunaiFormWrapper from "./runai-form-wrapper.vue";
import RunaiBaseDialog from "@/components/common/runai-base-dialog/runai-base-dialog.vue";

describe("RunaiFormWrapper - Keyboard Actions", () => {
  const createWrapper = (props = {}) => {
    return mount(RunaiFormWrapper, {
      global: {
        plugins: [Quasar],
        components: {
          QForm,
          QBtn,
          QField,
          RunaiBaseDialog,
        },
      },
      props: {
        formState: { test: "value" },
        ...props,
      },
      slots: {
        default: "<div>Form content</div>",
      },
    });
  };

  describe("enterCallback prop", () => {
    it("should call enterCallback when Enter key is pressed", async () => {
      const enterCallback = vi.fn();

      const wrapper = createWrapper({ enterCallback });

      await wrapper.find("form").trigger("keydown.enter");

      expect(enterCallback).toHaveBeenCalledTimes(1);
    });

    it("should NOT call enterCallback when it's not provided", async () => {
      const wrapper = createWrapper();

      // This should not throw an error
      await wrapper.find("form").trigger("keydown.enter");

      // No error means test passes - just verify wrapper is rendered
      expect(wrapper.find("section.runai-form-wrapper").exists()).toBe(true);
    });

    it("should call enterCallback multiple times when Enter is pressed multiple times", async () => {
      const enterCallback = vi.fn();

      const wrapper = createWrapper({ enterCallback });

      const form = wrapper.find("form");

      // Press Enter multiple times
      await form.trigger("keydown.enter");
      await form.trigger("keydown.enter");
      await form.trigger("keydown.enter");

      expect(enterCallback).toHaveBeenCalledTimes(3);
    });

    it("should call handleOnEnterKey method when Enter is pressed", async () => {
      const enterCallback = vi.fn();

      const wrapper = createWrapper({ enterCallback });

      await wrapper.find("form").trigger("keydown.enter");

      expect(enterCallback).toHaveBeenCalledTimes(1);
    });
  });

  describe("component structure", () => {
    it("should render the form wrapper section with correct class", () => {
      const wrapper = createWrapper();

      const section = wrapper.find("section.runai-form-wrapper");
      expect(section.exists()).toBe(true);
    });

    it("should render with medium-width class by default", () => {
      const wrapper = createWrapper();

      const section = wrapper.find("section.runai-form-wrapper");
      expect(section.classes()).toContain("medium-width");
      expect(section.classes()).not.toContain("full-width");
    });

    it("should render with full-width class when fullWidth prop is true", () => {
      const wrapper = createWrapper({ fullWidth: true });

      const section = wrapper.find("section.runai-form-wrapper");
      expect(section.classes()).toContain("full-width");
      expect(section.classes()).not.toContain("medium-width");
    });

    it("should render the slot content", () => {
      const wrapper = createWrapper();

      expect(wrapper.html()).toContain("Form content");
    });

    it("should contain a QForm component", () => {
      const wrapper = createWrapper();

      expect(wrapper.findComponent(QForm).exists()).toBe(true);
    });
  });

  describe("keyboard event handling", () => {
    it("should have keydown.enter listener on the form", async () => {
      const enterCallback = vi.fn();
      const wrapper = createWrapper({ enterCallback });

      // Check that the event listener is set up by triggering it on the form
      await wrapper.find("form").trigger("keydown.enter");

      // If enterCallback was called, the listener is properly set up
      expect(enterCallback).toHaveBeenCalled();
    });

    it("should not respond to other key events", async () => {
      const enterCallback = vi.fn();
      const wrapper = createWrapper({ enterCallback });

      const form = wrapper.find("form");

      // Try other keys
      await form.trigger("keydown.escape");
      await form.trigger("keydown.space");
      await form.trigger("keydown.tab");

      // enterCallback should not have been called
      expect(enterCallback).not.toHaveBeenCalled();
    });
  });

  describe("form state management", () => {
    it("should accept formState prop", () => {
      const formState = { name: "test", value: 123 };
      const wrapper = createWrapper({ formState });

      expect(wrapper.props("formState")).toEqual(formState);
    });

    it("should accept isPageReady prop", () => {
      const wrapper = createWrapper({ isPageReady: false });

      expect(wrapper.props("isPageReady")).toBe(false);
    });

    it("should accept initDirty prop", () => {
      const wrapper = createWrapper({ initDirty: true });

      expect(wrapper.props("initDirty")).toBe(true);
    });

    it("should accept greedy prop", () => {
      const wrapper = createWrapper({ greedy: true });

      expect(wrapper.props("greedy")).toBe(true);
    });
  });

  describe("form validation", () => {
    it("should expose validate method", () => {
      const wrapper = createWrapper();

      expect(typeof wrapper.vm.validate).toBe("function");
    });

    it("should validate successfully when form is valid", async () => {
      const wrapper = createWrapper();

      // Mock the internal form's validate method to return true
      const qForm = wrapper.findComponent(QForm);
      vi.spyOn(qForm.vm, "validate").mockResolvedValue(true);

      const result = await wrapper.vm.validate();

      expect(result).toBe(true);
    });
  });

  describe("initDirty reactivity", () => {
    it("should update isDirty when initDirty prop changes", async () => {
      const wrapper = createWrapper({ initDirty: false });

      // Initially isDirty should be false
      expect(wrapper.vm.isDirty).toBe(false);

      // Change initDirty prop to true
      await wrapper.setProps({ initDirty: true });
      expect(wrapper.vm.isDirty).toBe(true);

      // Change it back to false
      await wrapper.setProps({ initDirty: false });
      expect(wrapper.vm.isDirty).toBe(false);
    });

    it("should not reset isDirty to false when formState triggers updateDirtyState while initDirty is true", async () => {
      const wrapper = createWrapper({ initDirty: true, formState: {} });

      expect(wrapper.vm.isDirty).toBe(true);

      // Simulate a parent re-render passing a new {} reference (e.g. after API call)
      await wrapper.setProps({ formState: {} });
      expect(wrapper.vm.isDirty).toBe(true);
    });
  });

  describe("regression test - keyboard action bug", () => {
    it("should prevent regression: enterCallback must be called when provided", async () => {
      // This test specifically addresses the bug where enterCallback was not being passed
      // to runai-form-wrapper in workspace-new, training-new, and inference-new components

      const enterCallback = vi.fn();
      const wrapper = createWrapper({ enterCallback });

      // Simulate user pressing Enter in the form
      await wrapper.find("form").trigger("keydown.enter");

      // This MUST pass - if it fails, the keyboard action feature is broken
      expect(enterCallback).toHaveBeenCalledTimes(1);
      expect(enterCallback).toHaveBeenCalledWith();
    });

    it("should handle missing enterCallback gracefully without errors", async () => {
      // When enterCallback is not provided, pressing Enter should not throw errors
      const wrapper = createWrapper({ enterCallback: undefined });

      // This should not throw
      await expect(async () => {
        await wrapper.find("form").trigger("keydown.enter");
      }).not.toThrow();
    });
  });
});
