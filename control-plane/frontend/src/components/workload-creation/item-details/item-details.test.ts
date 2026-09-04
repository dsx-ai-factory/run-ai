/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import ItemDetails from "./item-details.vue";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import { itemDetailsMock } from "./item-details.mock";
import { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";

describe("item-details.vue", () => {
  const entity = WorkloadCreationDrawerEntities.Environment;

  const factory = (props = {}) => {
    return mount(ItemDetails, {
      plugins: [Quasar],
      props: { item: itemDetailsMock, entity, ...props },
    });
  };

  it("renders item details correctly", () => {
    const wrapper = factory();

    expect(wrapper.findAll(".details-title").length).toBe(6);
    expect(wrapper.findAll(".content-text").length).toBe(10);
    expect(wrapper.findAll(".border-bottom").length).toBe(5);
  });

  it("renders icon when detail.icon is present", () => {
    const wrapper = factory();
    expect(wrapper.findAllComponents({ name: "runai-svg-icon" }).length).toBe(2);
  });

  it("renders space correctly", () => {
    const wrapper = factory();

    expect(wrapper.findAll("[aid='spacing-div']").length).toBe(1);
  });
});
