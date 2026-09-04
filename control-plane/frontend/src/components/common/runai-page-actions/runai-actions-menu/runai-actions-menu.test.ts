/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";

import RunaiActionsMenu from "./runai-actions-menu.vue";

function mountMenu(largeIcon = false) {
  return mount(RunaiActionsMenu, {
    global: {
      stubs: {
        RunaiDropdown: true,
      },
    },
    props: {
      actionItems: [],
      largeIcon,
    },
  });
}

describe("RunaiActionsMenu", () => {
  it("keeps the default 20px icon class outside chart contexts", () => {
    const dropdown = mountMenu().getComponent({ name: "RunaiDropdown" });

    expect(dropdown.props("customClasses")).toContain("runai-icon-btn--default");
    expect(dropdown.props("customClasses")).not.toContain("runai-icon-btn--large");
  });

  it("supports the 24px icon class for chart menus", () => {
    const dropdown = mountMenu(true).getComponent({ name: "RunaiDropdown" });

    expect(dropdown.props("customClasses")).toContain("runai-icon-btn--large");
    expect(dropdown.props("customClasses")).not.toContain("runai-icon-btn--default");
  });
});
