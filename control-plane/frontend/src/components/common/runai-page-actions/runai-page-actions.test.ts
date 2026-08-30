import { mount, type VueWrapper } from "@vue/test-utils";
import { RunaiPageActions } from ".";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import type { Component } from "vue";
import { runaiPageActionsMocks } from "./runai-page-actions.mock";
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import type { ActionMenuItem } from "@/components/common/runai-page-actions/model";
import { beforeEach, describe, expect, vi, it } from "vitest";

interface IRunaiPageActionsProps {
  advancedFilters?: boolean;
  columns: ITableColumn[];
  filters: IFilterBy;
  selectedRowsAmount?: number;
  showPrimaryBtn?: boolean;
  showSearch?: boolean;
  primaryBtnLabel?: string;
  primaryBtnIcon?: string;
  primaryBtnDisable?: boolean;
  primaryTooltipText?: string;
  filterInputLabel?: string;
  actionMenuItems?: ActionMenuItem[];
  hideExportCsv?: boolean;
  showDetailsBtn?: boolean;
  emptyMessage?: string;
  drawerDarkMode?: boolean;
  disableActions?: boolean;
  supportRefresh?: boolean;
  hideAllClusterFilterOptions?: boolean;
  hideClusterFilter?: boolean;
  chipText?: string;
}
const DEFAULT_PROPS = {
  filters: runaiPageActionsMocks.filters,
  columns: runaiPageActionsMocks.columns,
  primaryBtnLabel: "new environment",
};
const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("RunaiPageActions", () => {
  let runaiPageActionsCmp: VueWrapper<InstanceType<typeof RunaiPageActions>>;
  const factory = (props?: IRunaiPageActionsProps) => {
    runaiPageActionsCmp = mount(RunaiPageActions, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          clickOutside: vi.fn(),
          permission: vi.fn(),
        },
        mocks: {
          $route: {
            query: "",
            meta: {
              resourceType: "environments",
            },
          },
        },
      },
      props: {
        ...DEFAULT_PROPS,
        ...props,
      },
    });
  };
  beforeEach(() => {
    factory();
  });
  describe("rendering behavior", () => {
    it("should load the component", () => {
      expect(runaiPageActionsCmp.html()).toContain("new environment");
    });

    it("Should not render the action buttons when there are no selected rows", () => {
      expect(runaiPageActionsCmp.find(".selected-rows-actions").exists()).toBe(false);
    });

    it("Should render the action buttons when there are selected rows", async () => {
      await runaiPageActionsCmp.setProps({ selectedRowsAmount: 1 });
      expect(runaiPageActionsCmp.find(".selected-rows-actions").exists()).toBe(true);
    });

    it("Should render the primary button when showPrimaryBtn is true", async () => {
      await runaiPageActionsCmp.setProps({ showPrimaryBtn: true });
      expect(runaiPageActionsCmp.find("[aid=primary-cta-btn]").exists()).toBe(true);
    });

    it("Should not render the primary button when showPrimaryBtn is false", async () => {
      await runaiPageActionsCmp.setProps({ showPrimaryBtn: false });
      expect(runaiPageActionsCmp.find("[aid=primary-cta-btn]").exists()).toBe(false);
    });

    it("Should disable the primary button when primaryBtnDisable is true", async () => {
      await runaiPageActionsCmp.setProps({ primaryBtnDisable: true });
      const primaryBtn = runaiPageActionsCmp.find(".primary-cta-btn");
      expect(primaryBtn.html()).toContain("disabled");
    });

    it("Should not disable the primary button when primaryBtnDisable is false", async () => {
      await runaiPageActionsCmp.setProps({ primaryBtnDisable: false });
      const primaryBtn = runaiPageActionsCmp.find(".primary-cta-btn");
      expect(primaryBtn.html()).not.toContain("disabled");
    });

    it("Should render the search input when showSearch is true", async () => {
      await runaiPageActionsCmp.setProps({ showSearch: true });
      expect(runaiPageActionsCmp.find(".runai-general-search").exists()).toBe(true);
    });

    it("Should not render the search input when showSearch is false", async () => {
      await runaiPageActionsCmp.setProps({ showSearch: false });
      expect(runaiPageActionsCmp.find(".runai-general-search").exists()).toBe(false);
    });

    it("Should render advanced filters when advancedFilters is true", async () => {
      await runaiPageActionsCmp.setProps({ advancedFilters: true });
      expect(runaiPageActionsCmp.find(".advanced-filters").exists()).toBe(true);
    });
  });

  describe("csv tooltip behavior", () => {
    it("should not include tooltip on CSV option when rowsNumber is undefined", () => {
      factory({ columns: runaiPageActionsMocks.columns, filters: { ...runaiPageActionsMocks.filters } });
      const vm = runaiPageActionsCmp.vm as unknown as { filteredActionMenuItems: ActionMenuItem[] };
      const csvItem = vm.filteredActionMenuItems.find((item) => item.label === "Download as CSV");
      expect(csvItem?.tooltip).toBeUndefined();
    });

    it("should not include tooltip on CSV option when rowsNumber is below limit", () => {
      factory({
        columns: runaiPageActionsMocks.columns,
        filters: { ...runaiPageActionsMocks.filters, rowsNumber: 100 },
      });
      const vm = runaiPageActionsCmp.vm as unknown as { filteredActionMenuItems: ActionMenuItem[] };
      const csvItem = vm.filteredActionMenuItems.find((item) => item.label === "Download as CSV");
      expect(csvItem?.tooltip).toBeUndefined();
    });

    it("should include tooltip on CSV option when rowsNumber reaches the limit", () => {
      factory({
        columns: runaiPageActionsMocks.columns,
        filters: { ...runaiPageActionsMocks.filters, rowsNumber: 20000 },
      });
      const vm = runaiPageActionsCmp.vm as unknown as { filteredActionMenuItems: ActionMenuItem[] };
      const csvItem = vm.filteredActionMenuItems.find((item) => item.label === "Download as CSV");
      expect(csvItem?.tooltip).toBe("Export to CSV is limited to 20,000 rows");
    });

    it("should include tooltip on CSV option when rowsNumber exceeds the limit", () => {
      factory({
        columns: runaiPageActionsMocks.columns,
        filters: { ...runaiPageActionsMocks.filters, rowsNumber: 50000 },
      });
      const vm = runaiPageActionsCmp.vm as unknown as { filteredActionMenuItems: ActionMenuItem[] };
      const csvItem = vm.filteredActionMenuItems.find((item) => item.label === "Download as CSV");
      expect(csvItem?.tooltip).toBe("Export to CSV is limited to 20,000 rows");
    });
  });

  describe("emits behavior", () => {
    it("Should emit 'create-entity' when the primary button is clicked", async () => {
      const primaryBtn = runaiPageActionsCmp.find(".primary-cta-btn");
      await primaryBtn.trigger("click");
      expect(runaiPageActionsCmp.emitted("create-entity")).toBeTruthy();
    });

    it("Should emit 'filters-changed' when the search input is changed", async () => {
      const searchInput: Component = runaiPageActionsCmp.findComponent(".runai-general-search");
      await searchInput.vm.$emit("update-search-term", "new-filter");
      expect(runaiPageActionsCmp.emitted("filters-changed")).toBeTruthy();
    });

    it("Should emit 'selected-actions-close' when the close button is clicked", async () => {
      await runaiPageActionsCmp.setProps({ selectedRowsAmount: 1 });
      const closeBtn = runaiPageActionsCmp.find("[aid='close-actions-btn']");
      await closeBtn.trigger("click");
      expect(runaiPageActionsCmp.emitted("selected-actions-close")).toBeTruthy();
    });
  });
});
