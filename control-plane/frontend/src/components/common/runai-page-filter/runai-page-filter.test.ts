/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { RunaiPageFilter } from ".";
import type { IFilterModel, IFilterOption } from "@/models/filter.model";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);
describe("runai-page-filter", () => {
  it("should add none existing filters to the filtersModel", async () => {
    const cmp = mount(RunaiPageFilter, {
      global: {
        plugins: [Quasar],
      },
      props: {
        initialFiltersModel: [],
        filterOptions: [
          {
            field: "name",
            label: "Department",
            name: "department-name",
          },
        ],
      },
    });

    let filter: IFilterModel = {
      field: "name",
      term: "dep",
      name: "department-name",
    };
    cmp.vm.saveFilter(filter);
    expect(cmp.vm.filtersModel[0]).toEqual(filter);

    filter = { field: "assignedGpus", term: "3", name: "assigned-gpus" };
    cmp.vm.saveFilter(filter);
    expect(cmp.vm.filtersModel.length).toBe(2);
  });

  it("should find and update an existing filter in filtersModel", () => {
    const cmp = mount(RunaiPageFilter, {
      global: {
        plugins: [Quasar],
      },
      props: {
        initialFiltersModel: [
          { field: "name", term: "dep", name: "department-name" },
          { field: "assignedGpus", term: "1", name: "assigned-gpus" },
        ],
        filterOptions: [
          {
            field: "name",
            label: "Department",
            name: "department-name",
          },
          {
            field: "assignedGpus",
            label: "Assigned GPUs",
            name: "assigned-gpus",
          },
        ],
      },
    });

    const filter: IFilterModel = {
      field: "name",
      term: "dep1",
      name: "department-name",
    };
    cmp.vm.saveFilter(filter);
    expect(cmp.vm.filtersModel[0]).toEqual(filter);
    expect(cmp.vm.filtersModel.length).toBe(2);
  });

  it("should remove a specific filter from the filtersModel", () => {
    const cmp = mount(RunaiPageFilter, {
      global: {
        plugins: [Quasar],
      },
      props: {
        initialFiltersModel: [
          { field: "name", term: "dep", name: "department-name" },
          { field: "assignedGpus", term: "1", name: "assigned-gpus" },
        ],
        filterOptions: [
          {
            field: "name",
            label: "Department",
            name: "department-name",
          },
          {
            field: "assignedGpus",
            label: "Assigned GPUs",
            name: "assigned-gpus",
          },
        ],
      },
    });
    const filterToBeRemoved = "department-name";
    cmp.vm.removeFilter(filterToBeRemoved);
    expect(cmp.vm.filtersModel.find((fm: IFilterModel) => fm.name === filterToBeRemoved)).toBeFalsy();
  });

  describe("getAutocompleteOptionsFn", () => {
    it("should return undefined for filter options without autocompleteOptionsFn", () => {
      const cmp = mount(RunaiPageFilter, {
        global: { plugins: [Quasar] },
        props: {
          initialFiltersModel: [],
          filterOptions: [
            { field: "gpuCount", label: "GPU Count", name: "gpu-count" },
            { field: "createdAt", label: "Created", name: "created-at" },
          ] as IFilterOption[],
        },
      });

      expect(cmp.vm.getAutocompleteOptionsFn("gpu-count")).toBeUndefined();
      expect(cmp.vm.getAutocompleteOptionsFn("created-at")).toBeUndefined();
    });

    it("should return undefined for unknown filter name", () => {
      const cmp = mount(RunaiPageFilter, {
        global: { plugins: [Quasar] },
        props: {
          initialFiltersModel: [],
          filterOptions: [{ field: "name", label: "Name", name: "name" }] as IFilterOption[],
        },
      });

      expect(cmp.vm.getAutocompleteOptionsFn("nonexistent")).toBeUndefined();
    });

    it("should return a wrapped function for options with autocompleteOptionsFn", () => {
      const mockAutocompleteFn = vi.fn().mockResolvedValue(["result-1", "result-2"]);
      const cmp = mount(RunaiPageFilter, {
        global: { plugins: [Quasar] },
        props: {
          initialFiltersModel: [],
          filterOptions: [
            {
              field: "name",
              label: "Workload",
              name: "workload-name",
              autocompleteOptionsFn: mockAutocompleteFn,
            },
          ] as IFilterOption[],
          filterBy: { clusterUuid: "cluster-xyz" },
        },
      });

      const wrappedFn = cmp.vm.getAutocompleteOptionsFn("workload-name");
      expect(wrappedFn).toBeDefined();
    });

    it("should pass clusterUuid from filterBy to the original autocomplete function", async () => {
      const mockAutocompleteFn = vi.fn().mockResolvedValue(["alpha", "beta"]);
      const cmp = mount(RunaiPageFilter, {
        global: { plugins: [Quasar] },
        props: {
          initialFiltersModel: [],
          filterOptions: [
            {
              field: "name",
              label: "Workload",
              name: "workload-name",
              autocompleteOptionsFn: mockAutocompleteFn,
            },
          ] as IFilterOption[],
          filterBy: { clusterUuid: "cluster-xyz" },
        },
      });

      const wrappedFn = cmp.vm.getAutocompleteOptionsFn("workload-name");
      expect(wrappedFn).toBeDefined();
      // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
      await wrappedFn!("search-term");

      expect(mockAutocompleteFn).toHaveBeenCalledWith("search-term", "cluster-xyz");
    });

    it("should still work when filterBy has no clusterUuid", async () => {
      const mockAutocompleteFn = vi.fn().mockResolvedValue([]);
      const cmp = mount(RunaiPageFilter, {
        global: { plugins: [Quasar] },
        props: {
          initialFiltersModel: [],
          filterOptions: [
            {
              field: "name",
              label: "Name",
              name: "name",
              autocompleteOptionsFn: mockAutocompleteFn,
            },
          ] as IFilterOption[],
          filterBy: {},
        },
      });

      const wrappedFn = cmp.vm.getAutocompleteOptionsFn("name");
      expect(wrappedFn).toBeDefined();
      // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
      await wrappedFn!("query");

      expect(mockAutocompleteFn).toHaveBeenCalledWith("query", undefined);
    });
  });

  describe("filter operations without autocomplete (regression)", () => {
    it("should save and update filters that have no autocompleteOptionsFn", () => {
      const cmp = mount(RunaiPageFilter, {
        global: { plugins: [Quasar] },
        props: {
          initialFiltersModel: [],
          filterOptions: [{ field: "gpuCount", label: "GPU Count", name: "gpu-count" }] as IFilterOption[],
        },
      });

      const filter: IFilterModel = { field: "gpuCount", term: "4", name: "gpu-count" };
      cmp.vm.saveFilter(filter);

      expect(cmp.vm.filtersModel).toHaveLength(1);
      expect(cmp.vm.filtersModel[0]).toEqual(filter);
    });

    it("should handle mixed filters with and without autocomplete", () => {
      const mockAutocompleteFn = vi.fn().mockResolvedValue([]);
      const cmp = mount(RunaiPageFilter, {
        global: { plugins: [Quasar] },
        props: {
          initialFiltersModel: [],
          filterOptions: [
            { field: "name", label: "Name", name: "name", autocompleteOptionsFn: mockAutocompleteFn },
            { field: "gpuCount", label: "GPU Count", name: "gpu-count" },
            { field: "createdAt", label: "Created", name: "created-at" },
          ] as IFilterOption[],
        },
      });

      cmp.vm.saveFilter({ field: "name", term: "train", name: "name" });
      cmp.vm.saveFilter({ field: "gpuCount", term: "2", name: "gpu-count" });
      cmp.vm.saveFilter({ field: "createdAt", term: "2024-01", name: "created-at" });

      expect(cmp.vm.filtersModel).toHaveLength(3);
      expect(cmp.vm.getAutocompleteOptionsFn("name")).toBeDefined();
      expect(cmp.vm.getAutocompleteOptionsFn("gpu-count")).toBeUndefined();
      expect(cmp.vm.getAutocompleteOptionsFn("created-at")).toBeUndefined();
    });
  });
});
