import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import SettingReadonlyInput from "./setting-readonly-input.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType, ResourceTypeGroupId, type Permission } from "@/swagger-models/authorization-client";

function permission(resourceType: ResourceType, actions: Action[]): Permission {
  return {
    resourceType,
    displayName: resourceType,
    groupId: ResourceTypeGroupId.Organization,
    actions,
  };
}

function createWrapper(permissions: Permission[], props: Record<string, unknown> = {}) {
  const pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
  usePermissionStore(pinia).setPermissions(permissions);

  return mount(SettingReadonlyInput, {
    global: {
      plugins: [Quasar, pinia],
      stubs: { QTooltip: true },
    },
    props,
    slots: {
      default: `<input data-test="child-input" />`,
    },
  });
}

describe("setting-readonly-input.vue", () => {
  describe("isReadOnly derived from permission store", () => {
    it("is read-only when the store reports READ-only for the given resourceType", () => {
      const wrapper = createWrapper([permission(ResourceType.SecuritySettings, [Action.Read])], {
        resourceType: "SecuritySettings",
      });

      expect(wrapper.vm.isReadOnly).toBe(true);
      expect(wrapper.classes()).toContain("readonly-wrapper");
    });

    it("is not read-only when the store reports CRUD for the given resourceType", () => {
      const wrapper = createWrapper([permission(ResourceType.SecuritySettings, [Action.Read, Action.Update])], {
        resourceType: "SecuritySettings",
      });

      expect(wrapper.vm.isReadOnly).toBe(false);
      expect(wrapper.classes()).not.toContain("readonly-wrapper");
    });

    it("is not read-only when the resourceType is missing entirely", () => {
      const wrapper = createWrapper([], { resourceType: "BrandingSettings" });

      expect(wrapper.vm.isReadOnly).toBe(false);
    });

    it("defaults resourceType to 'Settings' so default mounts check ResourceType.Settings", () => {
      const wrapper = createWrapper([permission(ResourceType.Settings, [Action.Read])]);

      expect(wrapper.vm.isReadOnly).toBe(true);
    });
  });

  describe("explicit readOnly prop overrides the store", () => {
    it("forces read-only=true even when the store would say false", () => {
      const wrapper = createWrapper([permission(ResourceType.SecuritySettings, [Action.Read, Action.Update])], {
        resourceType: "SecuritySettings",
        readOnly: true,
      });

      expect(wrapper.vm.isReadOnly).toBe(true);
      expect(wrapper.classes()).toContain("readonly-wrapper");
    });

    it("forces read-only=false even when the store would say true", () => {
      const wrapper = createWrapper([permission(ResourceType.SecuritySettings, [Action.Read])], {
        resourceType: "SecuritySettings",
        readOnly: false,
      });

      expect(wrapper.vm.isReadOnly).toBe(false);
      expect(wrapper.classes()).not.toContain("readonly-wrapper");
    });
  });

  describe("slot binding", () => {
    it("exposes the readonly state to slot consumers", () => {
      const pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
      usePermissionStore(pinia).setPermissions([permission(ResourceType.SecuritySettings, [Action.Read])]);

      const wrapper = mount(SettingReadonlyInput, {
        global: { plugins: [Quasar, pinia], stubs: { QTooltip: true } },
        props: { resourceType: "SecuritySettings" },
        slots: {
          default: `<template #default="{ readonly }"><span data-test="ro">{{ readonly }}</span></template>`,
        },
      });

      expect(wrapper.get('[data-test="ro"]').text()).toBe("true");
    });
  });

  describe("tooltip", () => {
    it("renders the tooltip when read-only and displayTooltip is true (default)", () => {
      const wrapper = createWrapper([permission(ResourceType.SecuritySettings, [Action.Read])], {
        resourceType: "SecuritySettings",
      });

      expect(wrapper.findComponent({ name: "QTooltip" }).exists()).toBe(true);
    });

    it("suppresses the tooltip when displayTooltip is false", () => {
      const wrapper = createWrapper([permission(ResourceType.SecuritySettings, [Action.Read])], {
        resourceType: "SecuritySettings",
        displayTooltip: false,
      });

      expect(wrapper.findComponent({ name: "QTooltip" }).exists()).toBe(false);
    });

    it("does not render the tooltip when not read-only", () => {
      const wrapper = createWrapper([permission(ResourceType.SecuritySettings, [Action.Read, Action.Update])], {
        resourceType: "SecuritySettings",
      });

      expect(wrapper.findComponent({ name: "QTooltip" }).exists()).toBe(false);
    });
  });
});
