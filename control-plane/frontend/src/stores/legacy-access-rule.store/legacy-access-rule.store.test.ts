import type { Mock } from "vitest";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useLegacyAccessRuleStore } from "./legacy-access-rule.store";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";

vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: {
    getAccessRulesCount: vi.fn(),
    getAccessRulesV2: vi.fn(),
  },
}));

const mockGetCount = accessRuleService.getAccessRulesCount as Mock;
const mockGetRules = accessRuleService.getAccessRulesV2 as Mock;

describe("useLegacyAccessRuleStore", () => {
  let store: ReturnType<typeof useLegacyAccessRuleStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    store = useLegacyAccessRuleStore();
    vi.clearAllMocks();
  });

  describe("initial state", () => {
    it("starts with empty legacyRows and loading false", () => {
      expect(store.legacyRows).toEqual([]);
      expect(store.loading).toBe(false);
      expect(store.hasLegacyAccessRules).toBe(false);
    });
  });

  describe("loadLegacyAccessRules", () => {
    it("populates rows for legacy roles that have access rules", async () => {
      mockGetCount.mockResolvedValue(3);
      mockGetRules.mockResolvedValue([{ roleId: 42 }]);

      await store.loadLegacyAccessRules();

      expect(store.legacyRows.length).toBeGreaterThan(0);
      store.legacyRows.forEach((row) => {
        expect(row.accessRuleCount).toBe(3);
        expect(row.roleId).toBe(42);
      });
    });

    it("excludes legacy roles with zero access rules", async () => {
      mockGetCount.mockResolvedValue(0);
      mockGetRules.mockResolvedValue([]);

      await store.loadLegacyAccessRules();

      expect(store.legacyRows).toEqual([]);
      expect(store.hasLegacyAccessRules).toBe(false);
    });

    it("sets loading true during fetch and false after", async () => {
      let capturedLoading = false;
      mockGetCount.mockImplementation(() => {
        capturedLoading = store.loading;
        return Promise.resolve(1);
      });
      mockGetRules.mockResolvedValue([{ roleId: 1 }]);

      await store.loadLegacyAccessRules();

      expect(capturedLoading).toBe(true);
      expect(store.loading).toBe(false);
    });

    it("pre-populates newRole using the replacement from ROLE_STAGE_CONFIG", async () => {
      mockGetCount.mockResolvedValue(5);
      mockGetRules.mockResolvedValue([{ roleId: 10 }]);

      await store.loadLegacyAccessRules();

      const l1Row = store.legacyRows.find((r) => r.legacyRole === "L1 researcher");
      expect(l1Row?.newRole).toBe("AI practitioner");
    });

    it("applies ROLE_NAME_API_MAP when pre-populating newRole", async () => {
      mockGetCount.mockResolvedValue(2);
      mockGetRules.mockResolvedValue([{ roleId: 7 }]);

      await store.loadLegacyAccessRules();

      const dataSourceRow = store.legacyRows.find((r) => r.legacyRole === "Data source administrator");
      // "Data & storage administrator" → "Data and storage administrator" via ROLE_NAME_API_MAP
      expect(dataSourceRow?.newRole).toBe("Data and storage administrator");
    });

    it("leaves newRole undefined for legacy roles without a replacement", async () => {
      mockGetCount.mockResolvedValue(1);
      mockGetRules.mockResolvedValue([{ roleId: 99 }]);

      await store.loadLegacyAccessRules();

      const envRow = store.legacyRows.find((r) => r.legacyRole === "Environment administrator");
      expect(envRow?.newRole).toBeUndefined();
    });

    it("uses getAccessRulesCount (not row length) for accessRuleCount", async () => {
      mockGetCount.mockResolvedValue(150);
      mockGetRules.mockResolvedValue([{ roleId: 5 }]); // only 1 returned (limit=1)

      await store.loadLegacyAccessRules();

      const row = store.legacyRows[0];
      expect(row.accessRuleCount).toBe(150);
    });

    it("excludes rows when getAccessRulesV2 returns no roleId", async () => {
      mockGetCount.mockResolvedValue(3);
      mockGetRules.mockResolvedValue([]);

      await store.loadLegacyAccessRules();

      // count > 0 but no roleId resolved → row is dropped to prevent migrating roleId=0
      expect(store.legacyRows).toEqual([]);
    });

    it("resets loading to false when API throws", async () => {
      mockGetCount.mockRejectedValue(new Error("network error"));
      mockGetRules.mockResolvedValue([]);

      await expect(store.loadLegacyAccessRules()).rejects.toThrow("network error");

      expect(store.loading).toBe(false);
    });

    it("updates hasLegacyAccessRules correctly", async () => {
      mockGetCount.mockResolvedValue(4);
      mockGetRules.mockResolvedValue([{ roleId: 1 }]);

      await store.loadLegacyAccessRules();

      expect(store.hasLegacyAccessRules).toBe(true);
    });
  });
});
