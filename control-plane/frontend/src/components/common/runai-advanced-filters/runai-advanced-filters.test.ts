/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import RunaiAdvancedFilters from "./runai-advanced-filters.vue";
import {
  EColumnFilterType,
  EFilterOperator,
  type IAdvancedFilterOption,
  type IFreeTextFilterModel,
  type INumericFilterModel,
  type IDateFilterModel,
  type IEnumBasedFilterModel,
  type IFilterBy,
  EDateFilterOperator,
} from "@/models/filter.model";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

const CLUSTER_UUID = "cluster-abc-123";

function createFilterBy(overrides: Partial<IFilterBy> = {}): IFilterBy {
  return { clusterUuid: CLUSTER_UUID, ...overrides };
}

describe("runai-advanced-filters", () => {
  describe("selectFilter - autocomplete injection", () => {
    it("should inject autocompleteOptionsFn for FreeText filter with autocomplete", () => {
      const mockAutocompleteFn = vi.fn().mockResolvedValue(["r1"]);
      const filterOptions: IAdvancedFilterOption[] = [
        {
          field: "name",
          label: "Workload",
          name: "workloadName",
          type: EColumnFilterType.FreeText,
          autocompleteOptionsFn: mockAutocompleteFn,
        },
      ];

      const cmp = mount(RunaiAdvancedFilters, {
        global: { plugins: [Quasar] },
        props: { filterOptions, filterBy: createFilterBy(), initialFiltersModel: [] },
      });

      const freeTextFilter: IFreeTextFilterModel = {
        name: "workloadName",
        label: "Workload",
        type: EColumnFilterType.FreeText,
        term: "",
        value: EFilterOperator.Contains,
      };

      cmp.vm.selectFilter(freeTextFilter);

      expect(cmp.vm.selectedFilter).toBeTruthy();
      expect((cmp.vm.selectedFilter as IFreeTextFilterModel).autocompleteOptionsFn).toBeDefined();
    });

    it("should pass clusterUuid from filterBy when calling injected autocomplete", async () => {
      const mockAutocompleteFn = vi.fn().mockResolvedValue(["result"]);
      const filterOptions: IAdvancedFilterOption[] = [
        {
          field: "name",
          label: "Workload",
          name: "workloadName",
          type: EColumnFilterType.FreeText,
          autocompleteOptionsFn: mockAutocompleteFn,
        },
      ];

      const cmp = mount(RunaiAdvancedFilters, {
        global: { plugins: [Quasar] },
        props: { filterOptions, filterBy: createFilterBy(), initialFiltersModel: [] },
      });

      const filter: IFreeTextFilterModel = {
        name: "workloadName",
        label: "Workload",
        type: EColumnFilterType.FreeText,
        term: "",
        value: EFilterOperator.Contains,
      };

      cmp.vm.selectFilter(filter);
      const injected = (cmp.vm.selectedFilter as IFreeTextFilterModel).autocompleteOptionsFn;
      expect(injected).toBeDefined();
      // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
      await injected!("search-text");

      expect(mockAutocompleteFn).toHaveBeenCalledWith("search-text", CLUSTER_UUID);
    });

    it("should NOT inject autocompleteOptionsFn for FreeText filter without autocomplete", () => {
      const filterOptions: IAdvancedFilterOption[] = [
        {
          field: "nodeName",
          label: "Node",
          name: "nodeName",
          type: EColumnFilterType.FreeText,
        },
      ];

      const cmp = mount(RunaiAdvancedFilters, {
        global: { plugins: [Quasar] },
        props: { filterOptions, filterBy: createFilterBy(), initialFiltersModel: [] },
      });

      const filter: IFreeTextFilterModel = {
        name: "nodeName",
        label: "Node",
        type: EColumnFilterType.FreeText,
        term: "",
        value: EFilterOperator.Contains,
      };

      cmp.vm.selectFilter(filter);

      expect((cmp.vm.selectedFilter as IFreeTextFilterModel).autocompleteOptionsFn).toBeUndefined();
    });

    it("should NOT inject autocompleteOptionsFn for Numeric filter", () => {
      const filterOptions: IAdvancedFilterOption[] = [
        {
          field: "gpuCount",
          label: "GPU Count",
          name: "gpuCount",
          type: EColumnFilterType.Numeric,
        },
      ];

      const cmp = mount(RunaiAdvancedFilters, {
        global: { plugins: [Quasar] },
        props: { filterOptions, filterBy: createFilterBy(), initialFiltersModel: [] },
      });

      const filter: INumericFilterModel = {
        name: "gpuCount",
        label: "GPU Count",
        type: EColumnFilterType.Numeric,
        minValue: 0,
        maxValue: 10,
        operator: EFilterOperator.GreaterThanOrEqual,
      };

      cmp.vm.selectFilter(filter);

      expect(cmp.vm.selectedFilter).toBeTruthy();
      expect((cmp.vm.selectedFilter as IFreeTextFilterModel).autocompleteOptionsFn).toBeUndefined();
    });

    it("should NOT inject autocompleteOptionsFn for Date filter", () => {
      const filterOptions: IAdvancedFilterOption[] = [
        {
          field: "createdAt",
          label: "Created",
          name: "createdAt",
          type: EColumnFilterType.Date,
        },
      ];

      const cmp = mount(RunaiAdvancedFilters, {
        global: { plugins: [Quasar] },
        props: { filterOptions, filterBy: createFilterBy(), initialFiltersModel: [] },
      });

      const filter: IDateFilterModel = {
        name: "createdAt",
        label: "Created",
        type: EColumnFilterType.Date,
        value: EDateFilterOperator.After,
        date: "2024-01-01",
      };

      cmp.vm.selectFilter(filter);

      expect(cmp.vm.selectedFilter).toBeTruthy();
      expect((cmp.vm.selectedFilter as IFreeTextFilterModel).autocompleteOptionsFn).toBeUndefined();
    });

    it("should NOT inject autocompleteOptionsFn for EnumBased filter", () => {
      const filterOptions: IAdvancedFilterOption[] = [
        {
          field: "status",
          label: "Status",
          name: "status",
          type: EColumnFilterType.EnumBased,
          selectOptions: [
            { label: "Running", value: "Running" },
            { label: "Failed", value: "Failed" },
          ],
        },
      ];

      const cmp = mount(RunaiAdvancedFilters, {
        global: { plugins: [Quasar] },
        props: { filterOptions, filterBy: createFilterBy(), initialFiltersModel: [] },
      });

      const filter: IEnumBasedFilterModel = {
        name: "status",
        label: "Status",
        type: EColumnFilterType.EnumBased,
        selectOptions: [
          { label: "Running", value: "Running" },
          { label: "Failed", value: "Failed" },
        ],
        selectedValues: [],
      };

      cmp.vm.selectFilter(filter);

      expect(cmp.vm.selectedFilter).toBeTruthy();
      expect((cmp.vm.selectedFilter as IFreeTextFilterModel).autocompleteOptionsFn).toBeUndefined();
    });

    it("should clear selectedFilter when null is passed", () => {
      const filterOptions: IAdvancedFilterOption[] = [
        { field: "name", label: "Name", name: "name", type: EColumnFilterType.FreeText },
      ];

      const cmp = mount(RunaiAdvancedFilters, {
        global: { plugins: [Quasar] },
        props: { filterOptions, filterBy: createFilterBy(), initialFiltersModel: [] },
      });

      cmp.vm.selectFilter(null);

      expect(cmp.vm.selectedFilter).toBeNull();
    });
  });

  describe("mixed filter options (regression)", () => {
    it("should correctly handle filter options with and without autocomplete side by side", () => {
      const mockAutocompleteFn = vi.fn().mockResolvedValue(["suggestion"]);
      const filterOptions: IAdvancedFilterOption[] = [
        {
          field: "name",
          label: "Workload",
          name: "workloadName",
          type: EColumnFilterType.FreeText,
          autocompleteOptionsFn: mockAutocompleteFn,
        },
        {
          field: "gpuCount",
          label: "GPU Count",
          name: "gpuCount",
          type: EColumnFilterType.Numeric,
        },
        {
          field: "status",
          label: "Status",
          name: "status",
          type: EColumnFilterType.EnumBased,
          selectOptions: [{ label: "Running", value: "Running" }],
        },
        {
          field: "createdAt",
          label: "Created",
          name: "createdAt",
          type: EColumnFilterType.Date,
        },
        {
          field: "nodeName",
          label: "Node",
          name: "nodeName",
          type: EColumnFilterType.FreeText,
        },
      ];

      const cmp = mount(RunaiAdvancedFilters, {
        global: { plugins: [Quasar] },
        props: { filterOptions, filterBy: createFilterBy(), initialFiltersModel: [] },
      });

      const freeTextWithAutocomplete: IFreeTextFilterModel = {
        name: "workloadName",
        label: "Workload",
        type: EColumnFilterType.FreeText,
        term: "",
        value: EFilterOperator.Contains,
      };
      cmp.vm.selectFilter(freeTextWithAutocomplete);
      expect((cmp.vm.selectedFilter as IFreeTextFilterModel).autocompleteOptionsFn).toBeDefined();

      const numericFilter: INumericFilterModel = {
        name: "gpuCount",
        label: "GPU Count",
        type: EColumnFilterType.Numeric,
        minValue: 0,
        maxValue: 8,
        operator: EFilterOperator.Equals,
      };
      cmp.vm.selectFilter(numericFilter);
      expect((cmp.vm.selectedFilter as IFreeTextFilterModel).autocompleteOptionsFn).toBeUndefined();

      const freeTextWithout: IFreeTextFilterModel = {
        name: "nodeName",
        label: "Node",
        type: EColumnFilterType.FreeText,
        term: "",
        value: EFilterOperator.Contains,
      };
      cmp.vm.selectFilter(freeTextWithout);
      expect((cmp.vm.selectedFilter as IFreeTextFilterModel).autocompleteOptionsFn).toBeUndefined();
    });
  });
});
