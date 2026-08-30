import { describe, it, expect, vi, beforeEach } from "vitest";
import { shallowMount } from "@vue/test-utils";
import { Quasar } from "quasar";
import BrandingSection from "./branding-section.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { ResourceType, Action } from "@/swagger-models/authorization-client";

vi.mock("@/stores/permissions.store");
vi.mock("@/stores/auth.store", () => ({
  useAuthStore: vi.fn(() => ({ logo: "" })),
}));
vi.mock("@/services/control-plane/branding-settings.service/branding-settings.service", () => ({
  brandingSettingsService: { getLogo: vi.fn().mockResolvedValue(undefined) },
}));

const makePermissionMock = (overrides: Record<string, ReturnType<typeof vi.fn>> = {}) => ({
  isReadOnly: vi.fn().mockReturnValue(false),
  hasSomePermittedActions: vi.fn().mockReturnValue(false),
  hasPermission: vi.fn().mockReturnValue(false),
  ...overrides,
});

const mountComponent = (readonly = false) => {
  return shallowMount(BrandingSection, {
    global: {
      plugins: [Quasar],
      stubs: {
        RunaiExpansionItem: { template: "<div><slot /></div>" },
        RunaiTooltip: true,
        SettingWrapperBox: true,
        SettingReadonlyInput: {
          template: `<div><slot :readonly="${readonly}" /></div>`,
        },
        QTooltip: true,
        QFile: true,
      },
    },
  });
};

const withLogo = async (wrapper: ReturnType<typeof mountComponent>) => {
  (wrapper.vm as unknown as Record<string, unknown>).currentLogoDataUrl = "data:image/png;base64,abc123";
  await wrapper.vm.$nextTick();
};

describe("branding-section", () => {
  beforeEach(() => {
    vi.mocked(usePermissionStore).mockReturnValue(
      makePermissionMock() as unknown as ReturnType<typeof usePermissionStore>,
    );
  });

  describe("canEditBranding", () => {
    it("is true when hasSomePermittedActions returns true", () => {
      vi.mocked(usePermissionStore).mockReturnValue(
        makePermissionMock({
          hasSomePermittedActions: vi.fn().mockReturnValue(true),
        }) as unknown as ReturnType<typeof usePermissionStore>,
      );
      const wrapper = mountComponent();
      expect(wrapper.vm.canEditBranding).toBe(true);
    });

    it("is false when hasSomePermittedActions returns false", () => {
      const wrapper = mountComponent();
      expect(wrapper.vm.canEditBranding).toBe(false);
    });

    it("checks BrandingSettings with Create and Update actions", () => {
      const mock = makePermissionMock();
      vi.mocked(usePermissionStore).mockReturnValue(mock as unknown as ReturnType<typeof usePermissionStore>);
      const wrapper = mountComponent();
      // Access to ensure the call happened
      void wrapper.vm.canEditBranding;
      expect(mock.hasSomePermittedActions).toHaveBeenCalledWith(ResourceType.BrandingSettings, [
        Action.Create,
        Action.Update,
      ]);
    });
  });

  describe("canDeleteBranding", () => {
    it("is true when hasPermission returns true", () => {
      vi.mocked(usePermissionStore).mockReturnValue(
        makePermissionMock({ hasPermission: vi.fn().mockReturnValue(true) }) as unknown as ReturnType<
          typeof usePermissionStore
        >,
      );
      const wrapper = mountComponent();
      expect(wrapper.vm.canDeleteBranding).toBe(true);
    });

    it("is false when hasPermission returns false", () => {
      const wrapper = mountComponent();
      expect(wrapper.vm.canDeleteBranding).toBe(false);
    });

    it("checks BrandingSettings with Delete action", () => {
      const mock = makePermissionMock();
      vi.mocked(usePermissionStore).mockReturnValue(mock as unknown as ReturnType<typeof usePermissionStore>);
      const wrapper = mountComponent();
      void wrapper.vm.canDeleteBranding;
      expect(mock.hasPermission).toHaveBeenCalledWith(ResourceType.BrandingSettings, Action.Delete);
    });
  });

  describe("logo action buttons", () => {
    it("hides both buttons when user has no write permissions (e.g. Delete-only)", async () => {
      const wrapper = mountComponent(false);
      await withLogo(wrapper);
      expect(wrapper.find('[aid="branding-edit-logo-btn"]').exists()).toBe(false);
      expect(wrapper.find('[aid="branding-remove-logo-btn"]').exists()).toBe(false);
    });

    it("shows only edit button when user has Create/Update but not Delete", async () => {
      vi.mocked(usePermissionStore).mockReturnValue(
        makePermissionMock({
          hasSomePermittedActions: vi.fn().mockReturnValue(true),
        }) as unknown as ReturnType<typeof usePermissionStore>,
      );
      const wrapper = mountComponent(false);
      await withLogo(wrapper);
      expect(wrapper.find('[aid="branding-edit-logo-btn"]').exists()).toBe(true);
      expect(wrapper.find('[aid="branding-remove-logo-btn"]').exists()).toBe(false);
    });

    it("shows only remove button when user has Delete only", async () => {
      vi.mocked(usePermissionStore).mockReturnValue(
        makePermissionMock({ hasPermission: vi.fn().mockReturnValue(true) }) as unknown as ReturnType<
          typeof usePermissionStore
        >,
      );
      const wrapper = mountComponent(false);
      await withLogo(wrapper);
      expect(wrapper.find('[aid="branding-edit-logo-btn"]').exists()).toBe(false);
      expect(wrapper.find('[aid="branding-remove-logo-btn"]').exists()).toBe(true);
    });

    it("shows both buttons when user has Create/Update and Delete", async () => {
      vi.mocked(usePermissionStore).mockReturnValue(
        makePermissionMock({
          hasSomePermittedActions: vi.fn().mockReturnValue(true),
          hasPermission: vi.fn().mockReturnValue(true),
        }) as unknown as ReturnType<typeof usePermissionStore>,
      );
      const wrapper = mountComponent(false);
      await withLogo(wrapper);
      expect(wrapper.find('[aid="branding-edit-logo-btn"]').exists()).toBe(true);
      expect(wrapper.find('[aid="branding-remove-logo-btn"]').exists()).toBe(true);
    });

    it("shows both buttons disabled when user is read-only", async () => {
      const wrapper = mountComponent(true);
      await withLogo(wrapper);
      expect(wrapper.find('[aid="branding-edit-logo-btn"]').exists()).toBe(true);
      expect(wrapper.find('[aid="branding-remove-logo-btn"]').exists()).toBe(true);
      expect(wrapper.find('[aid="branding-edit-logo-btn"]').attributes("disable")).toBe("true");
      expect(wrapper.find('[aid="branding-remove-logo-btn"]').attributes("disable")).toBe("true");
    });
  });

  describe("+ LOGO button (no existing logo)", () => {
    it("is hidden when user has no edit permission and is not read-only", async () => {
      const wrapper = mountComponent(false);
      await wrapper.vm.$nextTick();
      expect(wrapper.find('[aid="add-logo-btn"]').exists()).toBe(false);
    });

    it("is shown when user has Create/Update", async () => {
      vi.mocked(usePermissionStore).mockReturnValue(
        makePermissionMock({
          hasSomePermittedActions: vi.fn().mockReturnValue(true),
        }) as unknown as ReturnType<typeof usePermissionStore>,
      );
      const wrapper = mountComponent(false);
      await wrapper.vm.$nextTick();
      expect(wrapper.find('[aid="add-logo-btn"]').exists()).toBe(true);
    });

    it("is shown disabled when user is read-only", async () => {
      const wrapper = mountComponent(true);
      await wrapper.vm.$nextTick();
      const btn = wrapper.find('[aid="add-logo-btn"]');
      expect(btn.exists()).toBe(true);
      expect(btn.attributes("disable")).toBe("true");
    });
  });

  describe("zero branding-settings permission", () => {
    it("renders no action buttons when user has zero branding-settings actions and no logo", async () => {
      const wrapper = mountComponent(false);
      await wrapper.vm.$nextTick();
      expect(wrapper.find('[aid="add-logo-btn"]').exists()).toBe(false);
      expect(wrapper.find('[aid="branding-edit-logo-btn"]').exists()).toBe(false);
      expect(wrapper.find('[aid="branding-remove-logo-btn"]').exists()).toBe(false);
    });

    it("renders no action buttons when user has zero branding-settings actions and a logo exists", async () => {
      const wrapper = mountComponent(false);
      await withLogo(wrapper);
      expect(wrapper.find('[aid="add-logo-btn"]').exists()).toBe(false);
      expect(wrapper.find('[aid="branding-edit-logo-btn"]').exists()).toBe(false);
      expect(wrapper.find('[aid="branding-remove-logo-btn"]').exists()).toBe(false);
    });
  });
});
