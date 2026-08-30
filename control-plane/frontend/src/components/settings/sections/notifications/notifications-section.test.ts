import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, it, expect, beforeEach, vi } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import NotificationsSection from "./notifications-section.vue";
import { notificationService } from "@/services/control-plane/notifications.service/notification.service";
import type {
  DetailedNotificationStateSlack,
  NotificationChannelMeta,
  NotificationStateStatus,
} from "@/swagger-models/notifications-proxy-client";
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType, ResourceTypeGroupId, type Permission } from "@/swagger-models/authorization-client";

vi.mock("@/services/control-plane/notifications.service/notification.service");

const mockNotificationService = vi.mocked(notificationService);

const testTenantId = 1;

const testTenant = {
  id: testTenantId,
};

describe("NotificationsSection - Slack Display Logic", () => {
  let wrapper: VueWrapper<InstanceType<typeof NotificationsSection>>;

  const mockNotificationChannels: NotificationChannelMeta[] = [
    {
      name: `slack-${testTenantId}`,
      type: "slack",
    },
  ];

  const mockSlackNotificationState: DetailedNotificationStateSlack = {
    enabled: true,
    settings: {
      workspaceName: "Test Workspace",
    },
  };

  const mockEmailNotificationState: NotificationStateStatus = {
    enabled: false,
  };

  const createWrapper = (options = {}) => {
    const pinia = createTestingPinia({
      createSpy: vi.fn,
    });

    const authStore = useAuthStore(pinia);
    authStore.$patch({
      tenant: testTenant,
    });

    return mount(NotificationsSection, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $q: {
            notify: vi.fn(),
          },
        },
        stubs: {
          EmailNotifications: {
            template: '<div class="email-notifications-stub"></div>',
            props: ["setting"],
          },
          SystemNotifications: {
            template: '<div class="system-notifications-stub"></div>',
            props: ["setting"],
          },
          ToggleSetting: {
            template: '<div class="toggle-setting-stub"></div>',
            props: ["setting"],
          },
          RunaiExpansionItem: {
            template: '<div class="expansion-item-stub"><slot /></div>',
          },
          SlackNotifications: {
            template: '<div class="slack-notifications-stub"></div>',
            props: ["hasSlackIntegration"],
          },
        },
      },
      ...options,
    });
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
    mockNotificationService.getNotificationChannels.mockResolvedValue(mockNotificationChannels);
    mockNotificationService.getDetailedSlackNotificationState.mockResolvedValue(mockSlackNotificationState);
    mockNotificationService.getEmailNotificationState.mockResolvedValue(mockEmailNotificationState);
  });

  describe("Slack Display Logic", () => {
    it("shows Slack integration card when Slack is connected", async () => {
      wrapper = createWrapper();

      // Wait for the component to finish its async operations
      await wrapper.vm.$nextTick();
      await new Promise((resolve) => setTimeout(resolve, 0));

      expect(wrapper.vm.shouldShowSlackSection).toBe(true);
      expect(wrapper.vm.isSlackEnabled).toBe(true);
      expect(wrapper.vm.subheader).toContain("Slack notifications");
    });

    it("shows connect button when Slack is not connected", async () => {
      const mockStateWithoutSlack: DetailedNotificationStateSlack = {
        enabled: false,
        settings: {
          workspaceName: "",
        },
      };
      mockNotificationService.getDetailedSlackNotificationState.mockResolvedValue(mockStateWithoutSlack);
      mockNotificationService.getNotificationChannels.mockResolvedValue([]);

      wrapper = createWrapper();

      // Wait for the component to finish its async operations
      await wrapper.vm.$nextTick();
      await new Promise((resolve) => setTimeout(resolve, 0));

      expect(wrapper.vm.shouldShowSlackSection).toBe(true);
      expect(wrapper.vm.isSlackEnabled).toBe(false);
      expect(wrapper.vm.subheader).not.toContain("Slack notifications");
    });

    it("does not show Slack section when notification channels are not loaded", async () => {
      mockNotificationService.getDetailedSlackNotificationState.mockImplementation(
        () =>
          new Promise<DetailedNotificationStateSlack>(() => {
            // Never resolves - this is intentional for testing
          }),
      );

      wrapper = createWrapper();
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.shouldShowSlackSection).toBe(false);
    });

    it("correctly identifies tenant Slack channel name", async () => {
      wrapper = createWrapper();
      await wrapper.vm.$nextTick();
      await new Promise((resolve) => setTimeout(resolve, 0));

      expect(wrapper.vm.tenantSlackChannelName).toBe(`slack-${testTenantId}`);
    });
  });
});

describe("NotificationsSection - Permission wiring", () => {
  function permission(resourceType: ResourceType, actions: Action[]): Permission {
    return {
      resourceType,
      displayName: resourceType,
      groupId: ResourceTypeGroupId.Organization,
      actions,
    };
  }

  function createWrapperWithPermissions(permissions: Permission[]) {
    const pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
    usePermissionStore(pinia).setPermissions(permissions);
    const authStore = useAuthStore(pinia);
    authStore.$patch({ tenant: testTenant });
    const appStore = useAppStore(pinia);
    appStore.$patch({ config: { isMultiTenant: true } });

    mockNotificationService.getNotificationChannels.mockResolvedValue([]);
    mockNotificationService.getDetailedSlackNotificationState.mockResolvedValue({
      enabled: false,
      settings: { workspaceName: "" },
    });
    mockNotificationService.getEmailNotificationState.mockResolvedValue({ enabled: false });

    return mount(NotificationsSection, {
      global: {
        plugins: [Quasar, pinia],
        mocks: { $q: { notify: vi.fn() } },
        stubs: {
          EmailNotifications: true,
          SystemNotifications: true,
          SlackNotifications: true,
          RunaiExpansionItem: { template: "<div><slot /></div>" },
          ToggleSetting: {
            name: "ToggleSetting",
            props: {
              setting: { type: Object, default: () => ({}) },
              resourceType: { type: String, default: "Settings" },
              emitOnly: { type: Boolean, default: false },
            },
            template: '<div class="toggle-setting-stub"></div>',
          },
        },
      },
    });
  }

  beforeEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  it("passes resourceType='NotificationSettings' to the email-enable toggle", async () => {
    const wrapper = createWrapperWithPermissions([
      permission(ResourceType.NotificationSettings, [Action.Read, Action.Update]),
    ]);

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await wrapper.vm.$nextTick();

    const toggle = wrapper.findComponent({ name: "ToggleSetting" });
    expect(toggle.exists()).toBe(true);
    expect(toggle.props("resourceType")).toBe("NotificationSettings");
  });

  it("never passes resourceType='Settings' to the email-enable toggle (NotificationSettings owns it)", async () => {
    const wrapper = createWrapperWithPermissions([
      permission(ResourceType.NotificationSettings, [Action.Read, Action.Update]),
    ]);

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await wrapper.vm.$nextTick();

    const toggle = wrapper.findComponent({ name: "ToggleSetting" });
    expect(toggle.props("resourceType")).not.toBe("Settings");
  });

  it("emits-only on the toggle so the section owns the update flow (skips settingStore.updateSetting)", async () => {
    const wrapper = createWrapperWithPermissions([
      permission(ResourceType.NotificationSettings, [Action.Read, Action.Update]),
    ]);

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await wrapper.vm.$nextTick();

    const toggle = wrapper.findComponent({ name: "ToggleSetting" });
    expect(toggle.props("emitOnly")).toBe(true);
  });
});
