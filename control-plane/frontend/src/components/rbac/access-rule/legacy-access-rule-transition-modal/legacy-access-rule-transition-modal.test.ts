import type { Mock } from "vitest";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import LegacyAccessRuleTransitionModal from "./legacy-access-rule-transition-modal.vue";
import type { ILegacyRoleTransitionRow } from "@/models/legacy-access-rule-transition.model";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { useLegacyAccessRuleStore } from "@/stores/legacy-access-rule.store";
import type { RoleV1 } from "@/swagger-models/authorization-client/api";

vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: {
    migrateDeprecatedAccessRules: vi.fn(),
  },
}));

vi.mock("vue-router", () => ({
  useRouter: vi.fn().mockReturnValue({ resolve: vi.fn().mockReturnValue({ href: "/roles/1" }) }),
  useRoute: vi.fn().mockReturnValue({ name: "", query: {}, matched: [], params: {}, path: "", fullPath: "", hash: "" }),
}));

const mockMigrate = accessRuleService.migrateDeprecatedAccessRules as Mock;

const mockRoles = [
  { id: 1, name: "AI practitioner" },
  { id: 2, name: "Project administrator" },
  { id: 3, name: "Data and storage administrator" },
] as RoleV1[];

const mockRow = (overrides: Partial<ILegacyRoleTransitionRow> = {}): ILegacyRoleTransitionRow => ({
  id: "L1 researcher",
  legacyRole: "L1 researcher",
  roleId: 10,
  accessRuleCount: 5,
  newRole: "AI practitioner",
  ...overrides,
});

interface IModalVm {
  selectedRows: ILegacyRoleTransitionRow[];
  resolvedRows: ILegacyRoleTransitionRow[];
  transitioning: boolean;
  actionButtonLabel: string;
  showLimitWarning: boolean;
  isActionDisabled: boolean;
  onTransition: () => Promise<void>;
  closeDialog: () => void;
}

describe("legacy-access-rule-transition-modal.vue", () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let wrapper: VueWrapper<any>;
  let store: ReturnType<typeof useLegacyAccessRuleStore>;

  const createComponent = (rows: ILegacyRoleTransitionRow[] = [mockRow()]) => {
    const pinia = createTestingPinia({ createSpy: vi.fn });
    store = useLegacyAccessRuleStore(pinia);
    store.legacyRows = rows;
    store.loading = false;

    wrapper = mount(LegacyAccessRuleTransitionModal, {
      global: {
        plugins: [Quasar, pinia],
        stubs: {
          LegacyAccessRuleTransitionTable: {
            template: "<div class='stub-table'></div>",
            props: ["rows", "availableRoles", "loading", "selected"],
            emits: ["update:selected"],
          },
          RunaiBaseDialog: {
            template: "<div><slot name='header'></slot><slot name='body'></slot><slot name='footer'></slot></div>",
          },
          NotificationAlertBar: { template: "<div class='stub-notification-bar'></div>" },
        },
      },
      props: { availableRoles: mockRoles },
    });
    wrapper.vm.$q.notify = vi.fn();
  };

  const vm = (): IModalVm => wrapper.vm as unknown as IModalVm;

  beforeEach(() => {
    createComponent();
  });

  afterEach(() => {
    wrapper.unmount();
    vi.clearAllMocks();
  });

  describe("action button label", () => {
    it("shows zero count when no row is selected", () => {
      expect(vm().actionButtonLabel).toBe("Transition 0 access rules");
    });

    it("reflects the selected row's accessRuleCount", async () => {
      vm().selectedRows = [mockRow({ accessRuleCount: 7 })];
      await wrapper.vm.$nextTick();
      expect(vm().actionButtonLabel).toBe("Transition 7 access rules");
    });

    it("caps count at 100 when row has more than 100 rules", async () => {
      vm().selectedRows = [mockRow({ accessRuleCount: 150 })];
      await wrapper.vm.$nextTick();
      expect(vm().actionButtonLabel).toBe("Transition 100 access rules");
    });
  });

  describe("limit warning", () => {
    it("is hidden when row has <= 100 rules", async () => {
      vm().selectedRows = [mockRow({ accessRuleCount: 100 })];
      await wrapper.vm.$nextTick();
      expect(vm().showLimitWarning).toBe(false);
    });

    it("is shown when row has > 100 rules", async () => {
      vm().selectedRows = [mockRow({ accessRuleCount: 101 })];
      await wrapper.vm.$nextTick();
      expect(vm().showLimitWarning).toBe(true);
    });
  });

  describe("isActionDisabled", () => {
    it("is true when no row is selected", () => {
      expect(vm().isActionDisabled).toBe(true);
    });

    it("is true when selected row has no newRole", async () => {
      vm().selectedRows = [mockRow({ newRole: undefined })];
      await wrapper.vm.$nextTick();
      expect(vm().isActionDisabled).toBe(true);
    });

    it("is false when a row with a newRole is selected", async () => {
      vm().selectedRows = [mockRow()];
      await wrapper.vm.$nextTick();
      expect(vm().isActionDisabled).toBe(false);
    });
  });

  describe("watch — newRole normalization", () => {
    it("keeps newRole when it exists in availableRoles", async () => {
      store.legacyRows = [mockRow({ newRole: "AI practitioner" })];
      await flushPromises();
      expect(vm().resolvedRows[0].newRole).toBe("AI practitioner");
    });

    it("clears newRole when it no longer exists in availableRoles", async () => {
      store.legacyRows = [mockRow({ newRole: "Deprecated role" })];
      await flushPromises();
      expect(vm().resolvedRows[0].newRole).toBeUndefined();
    });

    it("resolves ROLE_NAME_API_MAP alias (& → and) before checking availability", async () => {
      store.legacyRows = [mockRow({ newRole: "Data & storage administrator" })];
      await flushPromises();
      expect(vm().resolvedRows[0].newRole).toBe("Data and storage administrator");
    });
  });

  describe("onTransition — success", () => {
    it("calls migrateDeprecatedAccessRules with correct role IDs", async () => {
      mockMigrate.mockResolvedValue({
        succeeded: { deleted: ["r1", "r2", "r3"], created: [], preExisting: [] },
        failed: [],
      });
      store.loadLegacyAccessRules = vi.fn().mockResolvedValue(undefined);
      vm().selectedRows = [mockRow({ roleId: 10, newRole: "AI practitioner" })];

      await vm().onTransition();
      await flushPromises();

      expect(mockMigrate).toHaveBeenCalledWith(10, 1);
    });

    it("shows success toast with deleted/total count", async () => {
      mockMigrate.mockResolvedValue({ succeeded: { deleted: ["r1", "r2"], created: [], preExisting: [] }, failed: [] });
      store.loadLegacyAccessRules = vi.fn().mockResolvedValue(undefined);
      vm().selectedRows = [mockRow({ accessRuleCount: 5, newRole: "AI practitioner" })];

      await vm().onTransition();
      await flushPromises();

      expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(
        expect.objectContaining({ message: "2/5 access rules transitioned" }),
      );
    });

    it("clears selectedRows after successful transition", async () => {
      mockMigrate.mockResolvedValue({ succeeded: { deleted: ["r1"], created: [], preExisting: [] }, failed: [] });
      store.loadLegacyAccessRules = vi.fn().mockResolvedValue(undefined);
      vm().selectedRows = [mockRow()];

      await vm().onTransition();
      await flushPromises();

      expect(vm().selectedRows).toEqual([]);
    });

    it("emits close when legacyRows is empty after transition", async () => {
      mockMigrate.mockResolvedValue({ succeeded: { deleted: ["r1"], created: [], preExisting: [] }, failed: [] });
      store.loadLegacyAccessRules = vi.fn().mockImplementation(() => {
        store.legacyRows = [];
        return Promise.resolve();
      });
      vm().selectedRows = [mockRow()];

      await vm().onTransition();
      await flushPromises();

      expect(wrapper.emitted("close")).toBeTruthy();
    });

    it("does not emit close when legacyRows still has rows after transition", async () => {
      mockMigrate.mockResolvedValue({ succeeded: { deleted: ["r1"], created: [], preExisting: [] }, failed: [] });
      store.loadLegacyAccessRules = vi.fn().mockImplementation(() => {
        store.legacyRows = [mockRow({ id: "L2 researcher" })];
        return Promise.resolve();
      });
      vm().selectedRows = [mockRow()];

      await vm().onTransition();
      await flushPromises();

      expect(wrapper.emitted("close")).toBeFalsy();
    });
  });

  describe("onTransition — >100 cap", () => {
    it("action button remains enabled when accessRuleCount exceeds 100", async () => {
      vm().selectedRows = [mockRow({ accessRuleCount: 150, newRole: "AI practitioner" })];
      await wrapper.vm.$nextTick();
      expect(vm().isActionDisabled).toBe(false);
    });

    it("label caps at 100 even when accessRuleCount exceeds 100", async () => {
      vm().selectedRows = [mockRow({ accessRuleCount: 150 })];
      await wrapper.vm.$nextTick();
      expect(vm().actionButtonLabel).toBe("Transition 100 access rules");
    });
  });

  describe("onTransition — partial failure", () => {
    it("shows deleted/total count when only some rules are migrated", async () => {
      mockMigrate.mockResolvedValue({
        succeeded: { deleted: ["r1", "r2"], created: [], preExisting: [] },
        failed: ["r3", "r4", "r5"],
      });
      store.loadLegacyAccessRules = vi.fn().mockResolvedValue(undefined);
      vm().selectedRows = [mockRow({ accessRuleCount: 5, newRole: "AI practitioner" })];

      await vm().onTransition();
      await flushPromises();

      expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(
        expect.objectContaining({ message: "2/5 access rules transitioned" }),
      );
    });
  });

  describe("onTransition — store reload failure", () => {
    it("resets transitioning to false when loadLegacyAccessRules throws after successful migration", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      mockMigrate.mockResolvedValue({ succeeded: { deleted: ["r1"], created: [], preExisting: [] }, failed: [] });
      store.loadLegacyAccessRules = vi.fn().mockRejectedValue(new Error("reload failed"));
      vm().selectedRows = [mockRow()];

      await vm().onTransition();
      await flushPromises();

      expect(vm().transitioning).toBe(false);
      expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(expect.objectContaining({ type: "negative" }));
      errorSpy.mockRestore();
    });
  });

  describe("onTransition — error", () => {
    it("shows error notification on API failure", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      mockMigrate.mockRejectedValue(new Error("Network error"));
      vm().selectedRows = [mockRow()];

      await vm().onTransition();
      await flushPromises();

      expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(expect.objectContaining({ type: "negative" }));
      errorSpy.mockRestore();
    });

    it("resets transitioning to false even when API throws", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      mockMigrate.mockRejectedValue(new Error("fail"));
      vm().selectedRows = [mockRow()];

      await vm().onTransition();
      await flushPromises();

      expect(vm().transitioning).toBe(false);
      errorSpy.mockRestore();
    });
  });

  describe("closeDialog", () => {
    it("emits close event", () => {
      vm().closeDialog();
      expect(wrapper.emitted("close")).toBeTruthy();
    });
  });
});
