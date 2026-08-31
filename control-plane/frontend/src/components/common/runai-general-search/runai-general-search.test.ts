/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe } from "vitest";
import { RunaiGeneralSearch } from ".";
describe("runai-general-search.vue", () => {
  let runaiGeneralSearchCmp: VueWrapper<InstanceType<typeof RunaiGeneralSearch>>;

  beforeEach(() => {
    runaiGeneralSearchCmp = mount(RunaiGeneralSearch, {
      global: {
        plugins: [Quasar],
        directives: {
          clickOutside: vi.fn(),
          permission: vi.fn(),
        },
      },
      props: {
        searchTerm: "",
      },
    });
  });

  test("component is loading properly", () => {
    expect(runaiGeneralSearchCmp.html()).toContain("search");
  });
});
