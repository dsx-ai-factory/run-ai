import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import ToggleSetting from "./toggle-setting.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { useSettingStore } from "@/stores/setting.store";
import { SettingKeys, type Setting } from "@/models/setting.model";
import { Action, ResourceType, ResourceTypeGroupId, type Permission } from "@/swagger-models/authorization-client";

function permission(resourceType: ResourceType, actions: Action[]): Permission {
  return {
    resourceType,
    displayName: resourceType,
    groupId: ResourceTypeGroupId.Organization,
    actions,
  };
}

function buildSetting(overrides: Partial<Setting> = {}): Setting {
  return {
    key: SettingKeys.AllowSwap,
    label: "Allow swap",
    description: "Allow swap",
    value: "false",
    ...overrides,
  } as Setting;
}

function createWrapper({
  permissions = [],
  props = {},
}: {
  permissions?: Permission[];
  props?: Record<string, unknown>;
} = {}) {
  const pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
  usePermissionStore(pinia).setPermissions(permissions);

  const settingStore = useSettingStore(pinia);
  settingStore.updateSetting = vi.fn().mockResolvedValue(undefined);

  const wrapper = mount(ToggleSetting, {
    global: {
      plugins: [Quasar, pinia],
      stubs: {
        QToggle: {
          name: "QToggle",
          props: ["modelValue", "disable", "aid"],
          emits: ["update:model-value"],
          template:
            '<button type="button" :data-disabled="disable" :data-aid="aid" @click="$emit(\'update:model-value\', !modelValue)"><slot /></button>',
        },
        QTooltip: true,
        QSeparator: true,
        NotificationAlertBar: true,
        RunaiStageBadge: true,
      },
      mocks: { $q: { notify: vi.fn() } },
    },
    props: {
      setting: buildSetting(),
      ...props,
    },
  });

  return { wrapper, settingStore };
}

describe("toggle-setting.vue", () => {
  describe("disable state derived from permission store", () => {
    it("disables the toggle when the resource is READ-only", () => {
      const { wrapper } = createWrapper({
        permissions: [permission(ResourceType.SecuritySettings, [Action.Read])],
        props: { resourceType: "SecuritySettings" },
      });

      const toggle = wrapper.findComponent({ name: "QToggle" });
      expect(toggle.props("disable")).toBe(true);
      expect(wrapper.vm.isReadOnly).toBe(true);
    });

    it("enables the toggle when the resource has CRUD", () => {
      const { wrapper } = createWrapper({
        permissions: [permission(ResourceType.SecuritySettings, [Action.Read, Action.Update])],
        props: { resourceType: "SecuritySettings" },
      });

      const toggle = wrapper.findComponent({ name: "QToggle" });
      expect(toggle.props("disable")).toBe(false);
      expect(wrapper.vm.isReadOnly).toBe(false);
    });

    it("defaults resourceType to 'Settings' so it checks ResourceType.Settings", () => {
      const { wrapper } = createWrapper({
        permissions: [permission(ResourceType.Settings, [Action.Read])],
      });

      expect(wrapper.vm.isReadOnly).toBe(true);
      expect(wrapper.findComponent({ name: "QToggle" }).props("disable")).toBe(true);
    });

    it("forwards the resourceType to the wrapping setting-readonly-input", () => {
      const { wrapper } = createWrapper({
        permissions: [permission(ResourceType.BrandingSettings, [Action.Read])],
        props: { resourceType: "BrandingSettings" },
      });

      const wrapperReadonly = wrapper.findComponent({ name: "setting-readonly-input" });
      expect(wrapperReadonly.props("resourceType")).toBe("BrandingSettings");
    });
  });

  describe("disabled prop", () => {
    it("disables the toggle even when the user has CRUD permission", () => {
      const { wrapper } = createWrapper({
        permissions: [permission(ResourceType.SecuritySettings, [Action.Read, Action.Update])],
        props: { resourceType: "SecuritySettings", disabled: true },
      });

      expect(wrapper.findComponent({ name: "QToggle" }).props("disable")).toBe(true);
    });
  });

  describe("update behavior", () => {
    it("emits update-setting with key and new value on toggle", async () => {
      const { wrapper } = createWrapper({
        permissions: [permission(ResourceType.SecuritySettings, [Action.Read, Action.Update])],
        props: { resourceType: "SecuritySettings" },
      });

      await wrapper.findComponent({ name: "QToggle" }).vm.$emit("update:model-value", true);

      const emitted = wrapper.emitted("update-setting");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]).toEqual([SettingKeys.AllowSwap, true]);
    });

    it("calls settingStore.updateSetting by default (emitOnly=false)", async () => {
      const { wrapper, settingStore } = createWrapper({
        permissions: [permission(ResourceType.SecuritySettings, [Action.Read, Action.Update])],
        props: { resourceType: "SecuritySettings" },
      });

      await wrapper.findComponent({ name: "QToggle" }).vm.$emit("update:model-value", true);

      expect(settingStore.updateSetting).toHaveBeenCalledWith(SettingKeys.AllowSwap, true);
    });

    it("does not call settingStore.updateSetting when emitOnly is true", async () => {
      const { wrapper, settingStore } = createWrapper({
        permissions: [permission(ResourceType.SecuritySettings, [Action.Read, Action.Update])],
        props: { resourceType: "SecuritySettings", emitOnly: true },
      });

      await wrapper.findComponent({ name: "QToggle" }).vm.$emit("update:model-value", true);

      expect(settingStore.updateSetting).not.toHaveBeenCalled();
      expect(wrapper.emitted("update-setting")).toBeTruthy();
    });
  });
});
