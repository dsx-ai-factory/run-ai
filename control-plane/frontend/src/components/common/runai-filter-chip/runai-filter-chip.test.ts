/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import RunaiFilterChip from "./runai-filter-chip.vue";
import { beforeEach, describe, expect, test } from "vitest";
import { Quasar } from "quasar";
import { unitTestUtil } from "@/utils/unit-test.util";

describe("runai-filter-chip.vue", () => {
  let runaiFilterChipWrapper: VueWrapper<InstanceType<typeof RunaiFilterChip>>;

  const demoDataProps = {
    title: "name",
    value: "special-test-name",
    display: false,
    removable: true,
  };

  function mountComponent(props = demoDataProps) {
    runaiFilterChipWrapper = mount(RunaiFilterChip, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  }

  async function clickOnChip() {
    return runaiFilterChipWrapper.find(".runai-filter-chip").trigger("click");
  }

  beforeEach(async () => {
    mountComponent();
  });

  test("component is loading properly", () => {
    expect(runaiFilterChipWrapper.classes()).toContain("runai-filter-chip");
  });

  describe("component emits", () => {
    test("component will emit 'save' when apply button is clicked", async () => {
      await clickOnChip();
      const popUpEl = unitTestUtil.getMenuElement();
      expect(popUpEl.find("[aid=apply-btn]").exists()).toBe(true);
      await popUpEl.find("[aid=apply-btn]").trigger("click");
      expect(runaiFilterChipWrapper.emitted()).toHaveProperty("save");
    });
  });

  describe("pop-up behaviour", () => {
    test("clicking on the chip opens a popup", async () => {
      await clickOnChip();
      const popUpEl = unitTestUtil.getMenuElement();
      expect(popUpEl.classes()).toContain("page-filter-popup");
    });
  });
});
