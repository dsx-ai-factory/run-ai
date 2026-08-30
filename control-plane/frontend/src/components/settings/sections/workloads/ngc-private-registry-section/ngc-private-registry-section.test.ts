import { describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import NgcPrivateRegistrySection from "./ngc-private-registry-section.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { useAuthStore } from "@/stores/auth.store";
import { catalogService } from "@/services/control-plane/catalog.service/catalog.service";
import { Action, ResourceType, ResourceTypeGroupId, type Permission } from "@/swagger-models/authorization-client";

vi.mock("@/services/control-plane/catalog.service/catalog.service", () => ({
  catalogService: {
    getRegistries: vi.fn().mockResolvedValue([]),
    createRegistry: vi.fn(),
    updateRegistry: vi.fn(),
    deleteRegistry: vi.fn(),
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
  const authStore = useAuthStore(pinia);
  Object.assign(authStore, { tenant: { id: 1 } });

  const wrapper = mount(NgcPrivateRegistrySection, {
    global: {
      plugins: [Quasar, pinia],
      stubs: {
        RunaiSection: { template: "<div><slot /><slot name='title-right-content' /></div>" },
        RunaiStageBadge: true,
        NgcPrivateRegistryBox: true,
        QBtn: {
          name: "QBtn",
          props: ["disable", "label"],
          template: '<button :data-disabled="disable">{{ label }}</button>',
        },
      },
      mocks: { $q: { notify: vi.fn() } },
    },
  });
  await flushPromises();
  return wrapper;
}

describe("ngc-private-registry-section.vue", () => {
  describe("isReadOnly computed", () => {
    it("returns true when Registries is READ-only", async () => {
      const wrapper = await createWrapper([permission(ResourceType.Registries, [Action.Read])]);

      expect(wrapper.vm.isReadOnly).toBe(true);
    });

    it("returns false when Registries has CRUD", async () => {
      const wrapper = await createWrapper([
        permission(ResourceType.Registries, [Action.Read, Action.Update, Action.Create]),
      ]);

      expect(wrapper.vm.isReadOnly).toBe(false);
    });

    it("returns false (default) when Registries is absent — visibility is gated by the parent", async () => {
      const wrapper = await createWrapper([permission(ResourceType.Settings, [Action.Read])]);

      expect(wrapper.vm.isReadOnly).toBe(false);
    });
  });

  describe("add-registry button", () => {
    it("is disabled when Registries is READ-only", async () => {
      const wrapper = await createWrapper([permission(ResourceType.Registries, [Action.Read])]);

      const btn = wrapper.findComponent({ name: "QBtn" });
      expect(btn.props("disable")).toBe(true);
    });

    it("is enabled when Registries has CRUD", async () => {
      const wrapper = await createWrapper([permission(ResourceType.Registries, [Action.Read, Action.Create])]);

      const btn = wrapper.findComponent({ name: "QBtn" });
      expect(btn.props("disable")).toBe(false);
    });
  });

  describe("setting-readonly-input wrapper", () => {
    it("uses resourceType='Registries' for the add-registry control", async () => {
      const wrapper = await createWrapper([permission(ResourceType.Registries, [Action.Read])]);

      const wrappers = wrapper.findAllComponents({ name: "setting-readonly-input" });
      expect(wrappers.some((w) => w.props("resourceType") === "Registries")).toBe(true);
    });

    it("does not use resourceType='Settings' anywhere (Registries is independent)", async () => {
      const wrapper = await createWrapper([permission(ResourceType.Registries, [Action.Read])]);

      const wrappers = wrapper.findAllComponents({ name: "setting-readonly-input" });
      expect(wrappers.every((w) => w.props("resourceType") !== "Settings")).toBe(true);
    });
  });

  it("calls catalogService.getRegistries on mount", async () => {
    await createWrapper([permission(ResourceType.Registries, [Action.Read])]);

    expect(catalogService.getRegistries).toHaveBeenCalled();
  });
});
