import { describe, expect, test } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import RoleWithBadgeCol from "../role-with-badge-col.vue";
import { RoleStageBadge } from "@/components/common/role-stage-badge";
import { ECustomCell, type ITableColumn } from "@/models/table.model";
import { ERoleStage } from "@/models/roles.model";

const createColumn = (overrides: Partial<ITableColumn> = {}): ITableColumn => ({
  name: "roleName",
  label: "Role",
  field: (row: { roleName: string }) => row.roleName,
  align: "left",
  sortable: true,
  customCell: ECustomCell.ROLE_WITH_BADGE_COL,
  ...overrides,
});

const mountComponent = (props: { row: { roleName: string }; column: ITableColumn }) => {
  return mount(RoleWithBadgeCol, {
    global: { plugins: [Quasar] },
    props,
  });
};

describe("RoleWithBadgeCol", () => {
  describe("link vs text rendering", () => {
    test("renders a clickable link by default (no config)", () => {
      const wrapper = mountComponent({
        row: { roleName: "Admin" },
        column: createColumn(),
      });
      expect(wrapper.find(".btn-link").exists()).toBe(true);
    });

    test("renders text when isTextOnly is true", () => {
      const wrapper = mountComponent({
        row: { roleName: "Admin" },
        column: createColumn({
          customCellConfig: {
            [ECustomCell.ROLE_WITH_BADGE_COL]: { isTextOnly: true },
          },
        }),
      });
      expect(wrapper.find(".btn-link").exists()).toBe(false);
      expect(wrapper.find("span").text()).toBe("Admin");
    });

    test("renders link when isTextOnly is false", () => {
      const wrapper = mountComponent({
        row: { roleName: "Admin" },
        column: createColumn({
          customCellConfig: {
            [ECustomCell.ROLE_WITH_BADGE_COL]: { isTextOnly: false },
          },
        }),
      });
      expect(wrapper.find(".btn-link").exists()).toBe(true);
      expect(wrapper.find(".btn-link").text()).toBe("Admin");
    });
  });

  describe("badge visibility", () => {
    test("shows Legacy badge for legacy role", () => {
      const wrapper = mountComponent({
        row: { roleName: "L1 researcher" },
        column: createColumn(),
      });
      const badge = wrapper.findComponent(RoleStageBadge);
      expect(badge.exists()).toBe(true);
      expect(badge.props("roleName")).toBe("L1 researcher");
    });

    test("passes hiddenStages to badge from config", () => {
      const wrapper = mountComponent({
        row: { roleName: "AI practitioner" },
        column: createColumn({
          customCellConfig: {
            [ECustomCell.ROLE_WITH_BADGE_COL]: { hiddenStages: [ERoleStage.New] },
          },
        }),
      });
      const badge = wrapper.findComponent(RoleStageBadge);
      expect(badge.exists()).toBe(true);
      expect(badge.props("hiddenStages")).toEqual([ERoleStage.New]);
    });

    test("does not pass hiddenStages when not configured", () => {
      const wrapper = mountComponent({
        row: { roleName: "L1 researcher" },
        column: createColumn(),
      });
      const badge = wrapper.findComponent(RoleStageBadge);
      expect(badge.props("hiddenStages")).toBeUndefined();
    });
  });

  describe("events", () => {
    test("emits custom-cell-event with row on link click", async () => {
      const row = { roleName: "Admin" };
      const wrapper = mountComponent({
        row,
        column: createColumn(),
      });
      await wrapper.find(".btn-link").trigger("click");
      expect(wrapper.emitted("custom-cell-event")?.[0]).toEqual([row]);
    });

    test("does not render clickable link in text-only mode", () => {
      const wrapper = mountComponent({
        row: { roleName: "Admin" },
        column: createColumn({
          customCellConfig: {
            [ECustomCell.ROLE_WITH_BADGE_COL]: { isTextOnly: true },
          },
        }),
      });
      expect(wrapper.find(".btn-link").exists()).toBe(false);
    });
  });
});
