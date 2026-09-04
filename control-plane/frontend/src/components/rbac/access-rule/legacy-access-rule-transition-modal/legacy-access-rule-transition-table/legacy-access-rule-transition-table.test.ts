/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import LegacyAccessRuleTransitionTable from "./legacy-access-rule-transition-table.vue";
import type { ILegacyRoleTransitionRow } from "@/models/legacy-access-rule-transition.model";
import { ECustomCell, type ITableColumn } from "@/models/table.model";
import type { RoleV1 } from "@/swagger-models/authorization-client/api";

vi.mock("vue-router", () => ({
  useRouter: vi.fn().mockReturnValue({ resolve: vi.fn().mockReturnValue({ href: "/roles/1" }) }),
  useRoute: vi.fn().mockReturnValue({ name: "", query: {}, matched: [], params: {}, path: "", fullPath: "", hash: "" }),
}));

const mockRow = (overrides: Partial<ILegacyRoleTransitionRow> = {}): ILegacyRoleTransitionRow => ({
  id: "L1 researcher",
  legacyRole: "L1 researcher",
  roleId: 10,
  accessRuleCount: 3,
  newRole: "AI practitioner",
  ...overrides,
});

const legacyRole = { id: 99, name: "L1 researcher" } as RoleV1;
const newRole1 = { id: 1, name: "AI practitioner" } as RoleV1;
const newRole2 = { id: 2, name: "Project administrator" } as RoleV1;

interface ITransitionTableVm {
  roleOptions: string[];
  columns: ITableColumn[];
}

describe("legacy-access-rule-transition-table.vue", () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let wrapper: VueWrapper<any>;

  const createComponent = (
    rows: ILegacyRoleTransitionRow[] = [mockRow()],
    availableRoles = [legacyRole, newRole1, newRole2],
  ) => {
    wrapper = mount(LegacyAccessRuleTransitionTable, {
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        stubs: {
          RunaiTable: {
            name: "RunaiTable",
            template: "<div class='stub-runai-table'></div>",
            props: ["rows", "columns", "loading", "filterBy", "selected", "getRowKey"],
            emits: ["update:selected"],
          },
        },
      },
      props: { rows, availableRoles, selected: [] },
    });
  };

  const vm = (): ITransitionTableVm => wrapper.vm as unknown as ITransitionTableVm;

  beforeEach(() => {
    createComponent();
  });

  afterEach(() => {
    wrapper.unmount();
    vi.clearAllMocks();
  });

  describe("roleOptions computed", () => {
    it("excludes roles with Legacy stage", () => {
      expect(vm().roleOptions).not.toContain("L1 researcher");
    });

    it("includes roles not in ROLE_STAGE_CONFIG (unrecognized roles are allowed)", () => {
      createComponent([], [{ id: 5, name: "Custom Role" } as RoleV1]);
      expect(vm().roleOptions).toContain("Custom Role");
    });

    it("includes New-stage roles from ROLE_STAGE_CONFIG", () => {
      createComponent(
        [],
        [{ id: 1, name: "AI practitioner" } as RoleV1, { id: 2, name: "Project administrator" } as RoleV1],
      );
      expect(vm().roleOptions).toContain("AI practitioner");
      expect(vm().roleOptions).toContain("Project administrator");
    });
  });

  describe("columns computed", () => {
    it("attaches customCellConfig only to the newRole column", () => {
      const newRoleCol = vm().columns.find((c) => c.name === "newRole");
      const legacyCol = vm().columns.find((c) => c.name === "legacyRole");

      expect(newRoleCol?.customCellConfig?.[ECustomCell.NEW_ROLE_COL]).toBeDefined();
      expect(legacyCol?.customCellConfig).toBeUndefined();
    });

    it("passes roleOptions into NEW_ROLE_COL config", () => {
      const newRoleCol = vm().columns.find((c) => c.name === "newRole");
      const config = newRoleCol?.customCellConfig?.[ECustomCell.NEW_ROLE_COL];
      expect(config?.roleOptions).toEqual(vm().roleOptions);
    });

    it("getReviewRoleId resolves the correct id for a role name", () => {
      const newRoleCol = vm().columns.find((c) => c.name === "newRole");
      const config = newRoleCol?.customCellConfig?.[ECustomCell.NEW_ROLE_COL];
      expect(config?.getReviewRoleId("AI practitioner")).toBe(1);
    });

    it("getReviewRoleId returns undefined for unknown role name", () => {
      const newRoleCol = vm().columns.find((c) => c.name === "newRole");
      const config = newRoleCol?.customCellConfig?.[ECustomCell.NEW_ROLE_COL];
      expect(config?.getReviewRoleId("Does not exist")).toBeUndefined();
    });
  });

  describe("update:selected passthrough", () => {
    it("re-emits update:selected from the inner table", async () => {
      createComponent();
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const tableStub = wrapper.findComponent({ name: "RunaiTable" }) as VueWrapper<any>;
      await tableStub.vm.$emit("update:selected", [mockRow()]);
      expect(wrapper.emitted("update:selected")?.[0][0]).toEqual([mockRow()]);
    });
  });
});
