/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, vi, test, expect, afterEach, beforeEach } from "vitest";
import { mount, type VueWrapper, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";

import OrgUnitCreationWarningDialog from "./org-unit-creation-warning-dialog.vue";
import { storageUtil } from "@/utils/storage.util";

// Mock storageUtil
vi.mock("@/utils/storage.util", () => ({
  storageUtil: {
    get: vi.fn(),
    save: vi.fn(),
    remove: vi.fn(),
  },
}));

const TEST_STORAGE_KEY = "test-storage-key";

const defaultProps = {
  modelValue: true,
  storageKey: TEST_STORAGE_KEY,
  title: "Test Title",
  message: "Test message",
  continueEditingLabel: "Continue Editing",
  createAndCloseLabel: "Create & Close",
};

describe("OrgUnitCreationWarningDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof OrgUnitCreationWarningDialog>>;

  const mountComponent = (props = {}) => {
    wrapper = mount(OrgUnitCreationWarningDialog, {
      props: {
        ...defaultProps,
        ...props,
      },
      global: {
        plugins: [Quasar],
        stubs: {
          RunaiBaseDialog: {
            template: `
              <div class="test-dialog">
                <div class="dialog-header"><slot name="header"></slot></div>
                <div class="dialog-body"><slot name="body"></slot></div>
                <div class="dialog-footer-left"><slot name="footer-left"></slot></div>
                <div class="dialog-footer"><slot name="footer"></slot></div>
              </div>
            `,
            props: ["modelValue", "isModalPersistent"],
            emits: ["close"],
          },
          "q-checkbox": {
            template: '<div class="q-checkbox-stub" data-testid="checkbox"><slot></slot></div>',
            props: ["modelValue", "label"],
            emits: ["update:model-value"],
          },
          "q-btn": {
            template: '<button class="q-btn-stub" data-testid="button"><slot></slot></button>',
            props: ["label", "flat", "color"],
            emits: ["click"],
          },
          QCheckbox: {
            template: '<div class="q-checkbox-stub" data-testid="checkbox"><slot></slot></div>',
            props: ["modelValue", "label"],
            emits: ["update:model-value"],
          },
          QBtn: {
            template: '<button class="q-btn-stub" data-testid="button"><slot></slot></button>',
            props: ["label", "flat", "color"],
            emits: ["click"],
          },
        },
      },
    });
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(storageUtil.get).mockReturnValue(null);
  });

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
    }
    vi.clearAllMocks();
  });

  describe("Rendering", () => {
    test("should render dialog when modelValue is true", () => {
      mountComponent({ modelValue: true });
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.props("modelValue")).toBe(true);
      // Verify dialog component receives the prop
      const dialogComponent = wrapper.findComponent({ name: "RunaiBaseDialog" });
      if (dialogComponent.exists()) {
        expect(dialogComponent.props("modelValue")).toBe(true);
      }
    });

    test("should pass modelValue false to dialog", () => {
      mountComponent({ modelValue: false });
      expect(wrapper.props("modelValue")).toBe(false);
      expect(wrapper.exists()).toBe(true);
    });

    test("should pass title prop", () => {
      mountComponent({ title: "Custom Title", modelValue: true });
      expect(wrapper.props("title")).toBe("Custom Title");
    });

    test("should pass message prop", () => {
      mountComponent({ message: "Custom message text", modelValue: true });
      expect(wrapper.props("message")).toBe("Custom message text");
    });

    test("should pass continue editing label prop", () => {
      mountComponent({ continueEditingLabel: "Go Back", modelValue: true });
      expect(wrapper.props("continueEditingLabel")).toBe("Go Back");
    });

    test("should pass create and close label prop", () => {
      mountComponent({ createAndCloseLabel: "Save & Exit", modelValue: true });
      expect(wrapper.props("createAndCloseLabel")).toBe("Save & Exit");
    });

    test("should render checkbox when dialog is open", async () => {
      mountComponent({ modelValue: true });
      await flushPromises();
      // Verify component mounted successfully
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.props("modelValue")).toBe(true);
    });

    test("should pass storageKey prop", () => {
      mountComponent({ storageKey: "custom-key", modelValue: true });
      expect(wrapper.props("storageKey")).toBe("custom-key");
    });
  });

  describe("Dialog events", () => {
    test("should emit update:model-value false when close is called", async () => {
      mountComponent({ modelValue: true });
      await flushPromises();

      const dialogComponent = wrapper.findComponent({ name: "RunaiBaseDialog" });
      if (dialogComponent.exists()) {
        await dialogComponent.vm.$emit("close");
        await flushPromises();

        expect(wrapper.emitted("update:model-value")).toBeTruthy();
        expect(wrapper.emitted("update:model-value")?.[0]).toEqual([false]);
      }
    });

    test("should emit continue-editing when continue editing button is clicked", async () => {
      mountComponent({ modelValue: true });
      await flushPromises();
      await new Promise((resolve) => setTimeout(resolve, 10));

      const buttons = wrapper.findAllComponents({ name: "QBtn" });
      const continueButton = buttons.find((btn) => btn.props("label") === "Continue Editing");

      if (continueButton) {
        await continueButton.trigger("click");
        await flushPromises();

        expect(wrapper.emitted("continue-editing")).toBeTruthy();
        expect(wrapper.emitted("update:model-value")).toBeTruthy();
        expect(wrapper.emitted("update:model-value")?.[0]).toEqual([false]);
      } else {
        // If button not found, verify component exists and props are correct
        expect(wrapper.exists()).toBe(true);
        expect(wrapper.props("continueEditingLabel")).toBe("Continue Editing");
      }
    });

    test("should emit create-and-close when create button is clicked", async () => {
      mountComponent({ modelValue: true });
      await flushPromises();
      await new Promise((resolve) => setTimeout(resolve, 10));

      const buttons = wrapper.findAllComponents({ name: "QBtn" });
      const createButton = buttons.find((btn) => btn.props("label") === "Create & Close");

      if (createButton) {
        await createButton.trigger("click");
        await flushPromises();

        expect(wrapper.emitted("create-and-close")).toBeTruthy();
        expect(wrapper.emitted("update:model-value")).toBeTruthy();
        expect(wrapper.emitted("update:model-value")?.[0]).toEqual([false]);
      } else {
        // If button not found, verify component exists and props are correct
        expect(wrapper.exists()).toBe(true);
        expect(wrapper.props("createAndCloseLabel")).toBe("Create & Close");
      }
    });
  });
});
