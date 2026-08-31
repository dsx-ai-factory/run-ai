/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { defineComponent, type ComponentPublicInstance } from "vue";

import {
  provideWizardPreview,
  useWizardPreview,
  useWizardPreviewRequired,
  type IWizardPreview,
} from "@/composables/org-unit/use-wizard-preview.composable";

const PreviewHost = defineComponent({
  setup() {
    const preview = provideWizardPreview();
    return { preview };
  },
  template: "<div />",
});

const mountPreview = () => {
  const wrapper = mount(PreviewHost);
  const preview = (wrapper.vm as ComponentPublicInstance<{ preview: IWizardPreview }>).preview;
  return { preview, wrapper };
};

describe("use-wizard-preview", () => {
  let preview: IWizardPreview;
  let wrapper: VueWrapper<ComponentPublicInstance<{ preview: IWizardPreview }>>;

  beforeEach(() => {
    const mounted = mountPreview();
    preview = mounted.preview;
    wrapper = mounted.wrapper as VueWrapper<ComponentPublicInstance<{ preview: IWizardPreview }>>;
  });

  afterEach(() => {
    wrapper.unmount();
  });

  describe("provideWizardPreview", () => {
    it("should initialize with empty state", () => {
      expect(preview.state.value).toEqual({});
    });

    it("should return an object with state, updatePreview, and clearPreview", () => {
      expect(preview.state).toBeDefined();
      expect(typeof preview.updatePreview).toBe("function");
      expect(typeof preview.clearPreview).toBe("function");
    });
  });

  describe("updatePreview", () => {
    it("should update gpuQuota preview state", () => {
      const gpuQuotaData = {
        nodePoolName: "default",
        gpuQuota: 10,
        defaultNodePools: ["default", "pool1"],
      };

      preview.updatePreview("gpuQuota", gpuQuotaData);

      expect(preview.state.value.gpuQuota).toEqual(gpuQuotaData);
    });

    it("should update schedulingRules preview state", () => {
      const schedulingRulesData = {
        schedulingRules: {
          interactiveJobTimeLimitSeconds: 3600,
          trainingJobTimeLimitSeconds: 7200,
        },
      };

      preview.updatePreview("schedulingRules", schedulingRulesData);

      expect(preview.state.value.schedulingRules).toEqual(schedulingRulesData);
    });

    it("should preserve existing state when updating a different key", () => {
      const gpuQuotaData = {
        nodePoolName: "default",
        gpuQuota: 10,
      };
      const schedulingRulesData = {
        schedulingRules: {
          interactiveJobTimeLimitSeconds: 3600,
        },
      };

      preview.updatePreview("gpuQuota", gpuQuotaData);
      preview.updatePreview("schedulingRules", schedulingRulesData);

      expect(preview.state.value.gpuQuota).toEqual(gpuQuotaData);
      expect(preview.state.value.schedulingRules).toEqual(schedulingRulesData);
    });

    it("should overwrite existing state for the same key", () => {
      const initialData = {
        nodePoolName: "default",
        gpuQuota: 10,
      };
      const updatedData = {
        nodePoolName: "pool1",
        gpuQuota: 20,
      };

      preview.updatePreview("gpuQuota", initialData);
      preview.updatePreview("gpuQuota", updatedData);

      expect(preview.state.value.gpuQuota).toEqual(updatedData);
    });
  });

  describe("clearPreview", () => {
    it("should clear a specific key from state", () => {
      preview.updatePreview("gpuQuota", { nodePoolName: "default", gpuQuota: 10 });
      preview.updatePreview("schedulingRules", { schedulingRules: null });

      preview.clearPreview("gpuQuota");

      expect(preview.state.value.gpuQuota).toBeUndefined();
      expect(preview.state.value.schedulingRules).toEqual({ schedulingRules: null });
    });

    it("should clear all state when no key is provided", () => {
      preview.updatePreview("gpuQuota", { nodePoolName: "default", gpuQuota: 10 });
      preview.updatePreview("schedulingRules", { schedulingRules: null });

      preview.clearPreview();

      expect(preview.state.value).toEqual({});
    });

    it("should handle clearing a key that does not exist", () => {
      preview.updatePreview("gpuQuota", { nodePoolName: "default", gpuQuota: 10 });

      preview.clearPreview("schedulingRules");

      expect(preview.state.value.gpuQuota).toEqual({ nodePoolName: "default", gpuQuota: 10 });
    });
  });

  describe("useWizardPreview (null-object pattern)", () => {
    it("should return a no-op implementation when used outside provider context", () => {
      const NoProviderComponent = defineComponent({
        setup() {
          const preview = useWizardPreview();
          return { preview };
        },
        template: "<div />",
      });

      const noProviderWrapper = mount(NoProviderComponent);
      const noProviderPreview = (noProviderWrapper.vm as ComponentPublicInstance<{ preview: IWizardPreview }>).preview;

      // Should have empty state
      expect(noProviderPreview.state.value).toEqual({});

      // updatePreview should be a no-op (not throw)
      expect(() => noProviderPreview.updatePreview("gpuQuota", { nodePoolName: "test", gpuQuota: 5 })).not.toThrow();

      // State should remain empty after update attempt
      expect(noProviderPreview.state.value).toEqual({});

      // clearPreview should be a no-op (not throw)
      expect(() => noProviderPreview.clearPreview()).not.toThrow();

      noProviderWrapper.unmount();
    });
  });

  describe("useWizardPreviewRequired", () => {
    it("should throw when used outside provider context", () => {
      const NoProviderComponent = defineComponent({
        setup() {
          // This should throw
          useWizardPreviewRequired();
          return {};
        },
        template: "<div />",
      });

      // Vue wraps setup errors, so we need to catch the error during mount
      expect(() => mount(NoProviderComponent, { global: { config: { errorHandler: () => {} } } })).toThrow(
        "useWizardPreviewRequired must be used within provideWizardPreview",
      );
    });

    it("should return the preview when used within provider context", () => {
      const ChildComponent = defineComponent({
        setup() {
          const preview = useWizardPreviewRequired();
          return { preview };
        },
        template: "<div />",
      });

      const ParentComponent = defineComponent({
        components: { ChildComponent },
        setup() {
          provideWizardPreview();
          return {};
        },
        template: "<ChildComponent />",
      });

      const parentWrapper = mount(ParentComponent);
      const childInstance = parentWrapper.findComponent(ChildComponent);
      const childPreview = (childInstance.vm as ComponentPublicInstance<{ preview: IWizardPreview }>).preview;

      expect(childPreview).toBeDefined();
      expect(childPreview.state).toBeDefined();

      parentWrapper.unmount();
    });
  });

  describe("reactivity", () => {
    it("should be reactive when state changes", async () => {
      expect(preview.state.value.gpuQuota).toBeUndefined();

      preview.updatePreview("gpuQuota", { nodePoolName: "default", gpuQuota: 10 });

      // Vue reactivity should update immediately in sync context
      expect(preview.state.value.gpuQuota?.gpuQuota).toBe(10);

      preview.updatePreview("gpuQuota", { nodePoolName: "default", gpuQuota: 20 });

      expect(preview.state.value.gpuQuota?.gpuQuota).toBe(20);
    });
  });
});
