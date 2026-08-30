import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import SystemNotifications from "./system-notifications.vue";
import { Action, ResourceType, ResourceTypeGroupId, type Permission } from "@/swagger-models/authorization-client";
import { usePermissionStore } from "@/stores/permissions.store";
import { useSettingStore } from "@/stores/setting.store";
import { SettingKeys, type IAdminMessage } from "@/models/setting.model";

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: {
    getYamlSubmissionSetting: () => Promise.resolve(null),
    getApplicationSubmissionSetting: () => Promise.resolve(null),
  },
}));

const adminMessage: IAdminMessage = {
  id: "admin-message-id",
  msg: "<p>Maintenance window tonight</p>",
  enableDismiss: true,
  createdAt: new Date("2026-04-30T00:00:00Z"),
};

function permission(resourceType: ResourceType, actions: Action[]): Permission {
  return {
    resourceType,
    displayName: resourceType,
    groupId: ResourceTypeGroupId.Organization,
    actions,
  };
}

function createWrapper(permissions: Permission[]) {
  const pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
  const permissionStore = usePermissionStore(pinia);
  permissionStore.setPermissions(permissions);

  const settingStore = useSettingStore(pinia);
  settingStore.adminMessage = adminMessage;

  return mount(SystemNotifications, {
    global: {
      plugins: [Quasar, pinia],
      stubs: {
        RunaiTooltip: true,
        RunaiTextEditor: true,
        SettingWrapperBox: {
          template: "<div><slot /></div>",
        },
      },
    },
    props: {
      setting: {
        key: SettingKeys.AdminMessage,
        value: adminMessage,
        label: "Admin message",
        description: "Write a message that will be displayed to all users",
      },
    },
  });
}

describe("SystemNotifications permissions", () => {
  it("allows editing admin message with settings update and notification settings read", () => {
    const wrapper = createWrapper([
      permission(ResourceType.Settings, [Action.Read, Action.Update]),
      permission(ResourceType.NotificationSettings, [Action.Read]),
    ]);

    expect(wrapper.vm.isReadOnly).toBe(false);
    expect(wrapper.vm.readEditMessage).toBe("Edit Message");
  });

  it("keeps admin message read-only when only notification settings can be updated", () => {
    const wrapper = createWrapper([
      permission(ResourceType.Settings, [Action.Read]),
      permission(ResourceType.NotificationSettings, [Action.Read, Action.Update]),
    ]);

    expect(wrapper.vm.isReadOnly).toBe(true);
    expect(wrapper.vm.readEditMessage).toBe("View Message");
  });

  it("keeps admin message read-only without settings permissions", () => {
    const wrapper = createWrapper([permission(ResourceType.NotificationSettings, [Action.Read, Action.Update])]);

    expect(wrapper.vm.isReadOnly).toBe(true);
    expect(wrapper.vm.readEditMessage).toBe("View Message");
  });
});
