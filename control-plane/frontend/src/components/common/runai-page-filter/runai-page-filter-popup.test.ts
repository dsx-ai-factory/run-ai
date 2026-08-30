import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import RunaiPageFilterPopup from "./runai-page-filter-popup.vue";
import type { IFilterModel } from "@/models/filter.model";

describe("runai-page-filter-popup", () => {
  it("should emit an edited filter with the inputValue", async () => {
    const cmp = mount(RunaiPageFilterPopup, {
      global: {
        plugins: [Quasar],
      },
      props: {
        title: "Department",
        filter: { field: "name", term: "dep" } as IFilterModel,
      },
    });

    cmp.setData({ inputValue: "dep1" });
    cmp.vm.saveFilter();
    expect(cmp.emitted().save[0]).toBeTruthy();
    expect(cmp.emitted().save[0]).toEqual([{ field: "name", term: "dep1" }]);
  });
});
