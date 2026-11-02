/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import UserDisplayNameCol from "../user-display-name-col.vue";
import { ECustomCell, type ITableColumn } from "@/models/table.model";

const column: ITableColumn = {
  name: "subject",
  label: "Performed by",
  field: "subject",
  customCell: ECustomCell.USER_DISPLAY_NAME_COL,
  customCellConfig: {
    [ECustomCell.USER_DISPLAY_NAME_COL]: {
      skipWhen: (row) => (row as { subjectType: string }).subjectType !== "user",
    },
  },
};
const render = (subject: string, subjectType: string) =>
  mount(UserDisplayNameCol, {
    props: { row: { subject, subjectType }, column },
    global: { plugins: [Quasar], stubs: { RunaiUserDisplayName: true } },
  });

describe("user-display-name-col actor fallback", () => {
  it("keeps service-account names readable on hover without looking them up as users", () => {
    const subject = "deployment-automation-".repeat(15);
    const wrapper = render(subject, "service-account");
    expect(wrapper.text()).toBe(subject);
    expect(wrapper.get("span").attributes("title")).toBe(subject);
    expect(wrapper.findComponent({ name: "RunaiUserDisplayName" }).exists()).toBe(false);
  });

  it("continues to use the user-display component for user actors", () => {
    const wrapper = render("admin@example.com", "user");
    expect(wrapper.getComponent({ name: "RunaiUserDisplayName" }).props("username")).toBe("admin@example.com");
  });

  it("uses a dash for an empty subject", () => {
    expect(render("", "app").text()).toBe("-");
  });
});
