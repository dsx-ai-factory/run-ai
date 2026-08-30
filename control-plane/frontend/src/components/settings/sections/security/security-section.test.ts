import { describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import SecuritySection from "./security-section.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { useSecuritySettingsStore } from "@/stores/security-settings.store";
import { useIdpsStore } from "@/stores/idps.store";
import { Action, ResourceType, ResourceTypeGroupId, type Permission } from "@/swagger-models/authorization-client";

vi.mock("@/services/infra/idle-timer.service/idle-timer.service", () => ({
  IdleTimer: {
    getInstance: () => ({ updateTimeout: vi.fn() }),
  },
}));

function permission(resourceType: ResourceType, actions: Action[]): Permission {
  return {
    resourceType,
    displayName: resourceType,
    groupId: ResourceTypeGroupId.Organization,
    actions,
  };
}

async function createWrapper(permissions: Permission[]) {
  const pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
  usePermissionStore(pinia).setPermissions(permissions);

  const securitySettingsStore = useSecuritySettingsStore(pinia);
  securitySettingsStore.getSettings = vi.fn().mockResolvedValue({
    browserSessionTimeout: { value: { idleTimeout: 30 } },
    logoutRedirectUri: { value: { uri: "" } },
  });
  securitySettingsStore.updateSetting = vi.fn().mockResolvedValue(undefined);

  const idpStore = useIdpsStore(pinia);
  Object.assign(idpStore, { currentIdp: undefined, idpMappers: [] });

  const wrapper = mount(SecuritySection, {
    global: {
      plugins: [Quasar, pinia],
      stubs: {
        RunaiExpansionItem: { template: "<div><slot /></div>" },
        RunaiTooltip: true,
        SsoSettings: true,
        SettingsUrlField: {
          name: "SettingsUrlField",
          props: ["resourceType", "serverUrl"],
          template: "<div></div>",
        },
        QInput: {
          name: "QInput",
          props: ["modelValue", "disable", "label"],
          template: '<input :data-disabled="disable" />',
        },
      },
      mocks: { $q: { notify: vi.fn() } },
    },
  });
  await flushPromises();
  return wrapper;
}

describe("security-section.vue", () => {
  it("exposes securityResourceType as 'SecuritySettings'", async () => {
    const wrapper = await createWrapper([]);

    expect(wrapper.vm.securityResourceType).toBe("SecuritySettings");
  });

  it("forwards resourceType='SecuritySettings' to settings-url-field", async () => {
    const wrapper = await createWrapper([permission(ResourceType.SecuritySettings, [Action.Read])]);

    const urlField = wrapper.findComponent({ name: "SettingsUrlField" });
    expect(urlField.props("resourceType")).toBe("SecuritySettings");
  });

  it("disables the session timeout input when SecuritySettings is READ-only", async () => {
    const wrapper = await createWrapper([permission(ResourceType.SecuritySettings, [Action.Read])]);

    const input = wrapper.findComponent({ name: "QInput" });
    expect(input.props("disable")).toBe(true);
  });

  it("enables the session timeout input when SecuritySettings has CRUD", async () => {
    const wrapper = await createWrapper([permission(ResourceType.SecuritySettings, [Action.Read, Action.Update])]);

    const input = wrapper.findComponent({ name: "QInput" });
    expect(input.props("disable")).toBe(false);
  });

  it("does not disable the session timeout input from Settings perms (gated only by SecuritySettings)", async () => {
    const wrapper = await createWrapper([permission(ResourceType.Settings, [Action.Read])]);

    const input = wrapper.findComponent({ name: "QInput" });
    expect(input.props("disable")).toBe(false);
  });

  it("wraps the session timeout input in a setting-readonly-input scoped to SecuritySettings", async () => {
    const wrapper = await createWrapper([permission(ResourceType.SecuritySettings, [Action.Read])]);

    const readonlyWrappers = wrapper.findAllComponents({ name: "setting-readonly-input" });
    expect(readonlyWrappers.some((w) => w.props("resourceType") === "SecuritySettings")).toBe(true);
  });
});
