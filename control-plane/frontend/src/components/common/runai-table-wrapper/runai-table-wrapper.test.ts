import { describe, expect, it, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { RunaiTableWrapper } from "./index";
import { setActivePinia, createPinia } from "pinia";

describe("runai-table-wrapper", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  const mockFilters = {
    sortBy: "date-time",
    descending: true,
    page: 1,
    rowsPerPage: 50,
    columnFilters: [],
    searchTerm: "",
    forceLoad: false,
  };

  it("Should contains slots and sticky class", async () => {
    const actionSlot = "Action slot";
    const tableSlot = "Table slot";
    const cmp = mount(RunaiTableWrapper, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        actions: actionSlot,
        table: tableSlot,
      },
      props: {
        filtersObject: mockFilters,
        sticky: true,
      },
    });

    expect(cmp.html()).toContain(actionSlot);
    expect(cmp.html()).toContain(tableSlot);
    expect(cmp.html()).toContain("runai-table-wrapper");
    expect(cmp.html()).toContain("runai-sticky");
  });

  it("Should not contains sticky class", async () => {
    const actionSlot = "Action slot";
    const tableSlot = "Table slot";
    const cmp = mount(RunaiTableWrapper, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        actions: actionSlot,
        table: tableSlot,
      },
      props: {
        filtersObject: mockFilters,
        sticky: false,
      },
    });
    expect(cmp.html()).toContain("runai-table-wrapper");
    expect(cmp.html()).not.toContain("runai-sticky");
  });
});
