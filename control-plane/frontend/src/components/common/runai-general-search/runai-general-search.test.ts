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

  afterEach(() => {
    runaiGeneralSearchCmp.unmount();
    vi.useRealTimers();
  });

  test("only emits the latest term after typing pauses", async () => {
    vi.useFakeTimers();
    await runaiGeneralSearchCmp.get("input").setValue("first");
    await runaiGeneralSearchCmp.get("input").setValue("latest");
    expect(runaiGeneralSearchCmp.emitted("update-search-term")).toBeUndefined();
    vi.advanceTimersByTime(300);
    expect(runaiGeneralSearchCmp.emitted("update-search-term")).toEqual([["latest"]]);
  });

  test("clearing cancels pending updates instead of restoring the old term", async () => {
    vi.useFakeTimers();
    await runaiGeneralSearchCmp.setProps({ searchTerm: "active" });
    await runaiGeneralSearchCmp.get("input").setValue("pending");
    await runaiGeneralSearchCmp.get('[aria-label="Clear search"]').trigger("click");
    vi.advanceTimersByTime(400);
    expect(runaiGeneralSearchCmp.emitted("update-search-term")).toEqual([[""]]);
  });

  test("an external reset cancels an uncommitted search", async () => {
    vi.useFakeTimers();
    await runaiGeneralSearchCmp.setProps({ searchTerm: "active" });
    await runaiGeneralSearchCmp.get("input").setValue("pending");
    await runaiGeneralSearchCmp.setProps({ searchTerm: "" });
    vi.advanceTimersByTime(400);
    expect(runaiGeneralSearchCmp.emitted("update-search-term")).toBeUndefined();
  });

  test("unmounting cancels the pending search timer", async () => {
    vi.useFakeTimers();
    await runaiGeneralSearchCmp.get("input").setValue("pending");
    runaiGeneralSearchCmp.unmount();
    expect(vi.getTimerCount()).toBe(0);
  });

  test("component is loading properly", () => {
    expect(runaiGeneralSearchCmp.html()).toContain("search");
  });
});
