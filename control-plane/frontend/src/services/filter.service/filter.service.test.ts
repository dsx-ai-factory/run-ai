/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi } from "vitest";
import { filterService } from "./filter.service";
import type { ITableColumn } from "@/models/table.model";
import { ETableFilters } from "@/models/table.model";
import type {
  IFilterBy,
  IFreeTextFilterModel,
  IPaginationFilter,
  IDateFilterModel,
  IEnumBasedFilterModel,
} from "@/models/filter.model";
import { EColumnFilterType, EFilterOperator, ESortOrder, EDateFilterOperator } from "@/models/filter.model";
import { WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import {
  AUDIT_APP_SUBJECT_TYPE,
  AUDIT_LEGACY_SERVICE_ACCOUNT_SUBJECT_TYPE,
  AUDIT_SERVICE_ACCOUNT_SUBJECT_TYPE,
  AUDIT_USER_SUBJECT_TYPE,
} from "@/models/audit.model";

interface IMockList {
  id: string;
  name: string;
  clusterUuid: string;
  resourcesId: number;
  projectsGpus: null | number;
  projects: Array<{ id: string; name: string }>;
  resources: {
    gpu: {
      deserved: number;
      overQuotaWeight: null;
      maxAllowed: number;
    };
  };
}

const mockList: Array<IMockList> = [
  {
    id: "63f4a274cfa144a74d65b7b4",
    name: "Lavonne Owen",
    clusterUuid: "01fc8553-5947-4d0e-9d07-796a3ed6498c",
    resourcesId: 90,
    projectsGpus: null,
    projects: [
      {
        id: "8eb93469-d468-4cfd-8957-aa9551c31f08",
        name: "Liquidoc",
      },
      {
        id: "1b8387cc-9d16-4abb-9e71-aec078e6fdac",
        name: "Enthaze",
      },
    ],
    resources: {
      gpu: {
        deserved: 3,
        overQuotaWeight: null,
        maxAllowed: 0,
      },
    },
  },
  {
    id: "63f4a2741c73c949936a64dc",
    name: "Potts Pratt",
    clusterUuid: "ca66453d-6d8f-4533-aa4d-c46130ead91c",
    resourcesId: 27,
    projectsGpus: 2,
    projects: [
      {
        id: "0423e22c-1968-4584-aff1-83bfca5d63b3",
        name: "Bristo",
      },
      {
        id: "f3496e61-137b-4a66-a789-f82a50d7765c",
        name: "Maximind",
      },
    ],
    resources: {
      gpu: {
        deserved: 4,
        overQuotaWeight: null,
        maxAllowed: 0,
      },
    },
  },
  {
    id: "63f4a274e180d1a233bb2df9",
    name: "Bettie Pratt",
    clusterUuid: "6fecd1f6-0e20-44ab-b168-814e4fd0a331",
    resourcesId: 8,
    projectsGpus: null,
    projects: [
      {
        id: "751e9c71-29e0-41e4-87a8-d36bd857af28",
        name: "Ersum",
      },
      {
        id: "a9e27f52-a804-4dde-9dfa-9517c3eb845f",
        name: "Voratak",
      },
    ],
    resources: {
      gpu: {
        deserved: 3,
        overQuotaWeight: null,
        maxAllowed: 0,
      },
    },
  },
];

const columns: Array<ITableColumn> = [
  {
    name: "id",
    label: "Id",
    field: (row: IMockList) => row.id,
    sortable: true,
    align: "left",
  },
  {
    name: "name",
    label: "Name",
    field: (row: IMockList) => row.name,
    sortable: true,
    align: "left",
  },
  {
    name: "clusterUuid",
    label: "Cluster Uuid",
    field: (row: IMockList) => row.clusterUuid,
    sortable: true,
    align: "left",
  },
  {
    name: "resourcesId",
    label: "Resources Id",
    field: (row: IMockList) => row.resourcesId,
    sortable: true,
    align: "left",
  },
  {
    name: "projectsGpus",
    label: "Projects Gpus",
    field: (row: IMockList) => row.projectsGpus,
    sortable: true,
    align: "left",
  },
  {
    name: "projects",
    label: "projects",
    field: (row: IMockList) => row.projects,
    sortable: true,
    align: "left",
    format: (projects: Array<{ id: string; name: string }>): Array<string> => {
      return projects.map((project: { id: string; name: string }) => project.name);
    },
  },
  {
    name: "id",
    label: "Id",
    field: (row: IMockList) => row.resources,
    sortable: true,
    align: "left",
    format: (resources: { cpu: { deserved: number } }) =>
      resources.cpu.deserved != null && resources.cpu.deserved >= 0 ? resources.cpu.deserved / 1000 : "-",
  },
];

describe("FilterService", () => {
  describe("FilterService::filterByColumns", () => {
    it("Should return an array filtered by specific columns", () => {
      const filters = [
        { field: "name", term: "Pratt", name: "name" },
        { field: "projectsGpus", term: "2", name: "projectsGpus" },
      ];

      const filteredDepartments = filterService.filterByColumns(mockList, filters, columns);
      expect(filteredDepartments.length).toEqual(1);
    });

    it("Should return an empty array when search in columns not found", () => {
      const filters = [{ field: "name", term: "Popo", name: "name" }];

      const filteredDepartments = filterService.filterByColumns(mockList, filters, columns);
      expect(filteredDepartments.length).toEqual(0);
    });
  });

  describe("FilterService::filterBySearchTerm2", () => {
    const displayedColumns = ["name", "clusterUuid", "projectsGpus", "projects"];
    it("Should return an array filtered by a search term", () => {
      const searchTerm = "Bristo";
      const filteredDepartments: Array<IMockList> = filterService.filterBySearchTerm<IMockList>(
        mockList,
        searchTerm,
        displayedColumns,
        columns,
      );

      expect(filteredDepartments.length).toEqual(1);
    });

    it("Should return an array an empty array when a search term is not found", () => {
      const searchTerm = "Popo";
      const filteredDepartments: Array<IMockList> = filterService.filterBySearchTerm<IMockList>(
        mockList,
        searchTerm,
        displayedColumns,
        columns,
      );

      expect(filteredDepartments.length).toEqual(0);
    });
  });
  describe("FilterService::enum filter options with matchValues", () => {
    function buildStatusFilter(selectedValues: string[]): IFilterBy {
      return {
        page: 1,
        rowsPerPage: 20,
        displayedColumns: ["status"],
        advancedFilters: [
          {
            name: "status",
            label: "Status",
            type: EColumnFilterType.EnumBased,
            inputType: "checkbox",
            selectedValues,
            selectOptions: [
              { label: "Ready", value: "Ready" },
              { label: "Failed", value: "Failed" },
              { label: "Deleting", value: "Deleting", matchValues: ["Deleting", "PendingDelete"] },
            ],
          },
        ],
      } as IFilterBy;
    }

    // Deleting and PendingDelete render as one label in the grid, so selecting that
    // option must not exclude PendingDelete rows the user can plainly see.
    it("keeps every value an option stands for when that option is selected", () => {
      const result = filterService.mapColumnsFilterToFilterParams(buildStatusFilter(["Deleting"]));

      expect(result.filterBy).toEqual(["status!=Ready", "status!=Failed"]);
    });

    it("excludes every value an option stands for when that option is not selected", () => {
      const result = filterService.mapColumnsFilterToFilterParams(buildStatusFilter(["Ready"]));

      expect(result.filterBy).toEqual(["status!=Failed", "status!=Deleting", "status!=PendingDelete"]);
    });

    it("falls back to the option value when it stands for nothing else", () => {
      const result = filterService.mapColumnsFilterToFilterParams(buildStatusFilter(["Ready", "Deleting"]));

      expect(result.filterBy).toEqual(["status!=Failed"]);
    });
  });

  describe("FilterService::mapColumnsFilterToFilterParams", () => {
    describe("mapColumnsFilterToFilterParams", () => {
      it("should map column filters to filter params correctly", () => {
        const filterBy: IFilterBy = {
          sortBy: "name",
          descending: false,
          page: 1,
          rowsPerPage: 20,
          columnFilters: [],
          searchTerm: "test",
          displayedColumns: [
            "name",
            "type",
            "phase",
            "projectName",
            "departmentName",
            "nodepool",
            "runningPodsVsRequestedPods",
            "requestedGPU",
            "allocatedGPU",
            "requestedGPUMemory",
            "allocatedGPUMemory",
          ],
          advancedFilters: [
            {
              term: "train",
              value: EFilterOperator.Contains,
              label: "Workload",
              name: "name",
              type: EColumnFilterType.FreeText,
            },
          ],
          rowsNumber: 0,
        };

        const result: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result).toEqual({
          offset: 0,
          limit: 20,
          sortOrder: ESortOrder.Asc,
          sortBy: "name",
          filterBy: ["name=@train"],
        });
      });
      it("should be able to search free text and contains", () => {
        const filterBy: IFilterBy = {
          sortBy: "name",
          descending: false,
          page: 1,
          rowsPerPage: 20,
          columnFilters: [],
          searchTerm: "test",
          advancedFilters: [
            {
              term: "train",
              value: EFilterOperator.Contains,
              label: "Workload",
              name: "name",
              type: EColumnFilterType.FreeText,
            },
          ],
          rowsNumber: 0,
        };

        const result: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result).toEqual({
          offset: 0,
          limit: 20,
          sortOrder: ESortOrder.Asc,
          sortBy: "name",
          filterBy: ["name=@train"],
        });
      });
      it("should escape commas as ';' inside a source_ip free-text term", () => {
        const filterBy: IFilterBy = {
          sortBy: "timestamp",
          descending: false,
          page: 1,
          rowsPerPage: 20,
          columnFilters: [],
          advancedFilters: [
            {
              term: "1.1.1.1, 2.2.2.2",
              value: EFilterOperator.Equals,
              label: "Source IP",
              name: "source_ip",
              type: EColumnFilterType.FreeText,
            },
          ],
          rowsNumber: 0,
        };

        const result: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result.filterBy).toEqual(["source_ip==1.1.1.1; 2.2.2.2"]);
      });
      it("should map advanced column filters to filter params correctly", () => {
        const filterBy: IFilterBy = {
          sortBy: "name",
          descending: false,
          page: 1,
          rowsPerPage: 10,
          columnFilters: [],
          searchTerm: "training",
          rowsNumber: 3,
        };

        const result: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result).toEqual({
          offset: 0,
          limit: 10,
          sortOrder: ESortOrder.Asc,
          search: undefined,
          sortBy: "name",
          filterBy: [],
        });
      });
    });

    describe("EnumBased filters", () => {
      const buildFilterBy = (advancedFilter: IEnumBasedFilterModel): IFilterBy => ({
        sortBy: "subject_type",
        descending: false,
        page: 1,
        rowsPerPage: 20,
        advancedFilters: [advancedFilter],
      });

      const subjectTypeOptions = [
        { label: "User", value: AUDIT_USER_SUBJECT_TYPE },
        { label: "Service account", value: AUDIT_APP_SUBJECT_TYPE },
      ];

      it("should emit a positive == filter for a single selected enum value", () => {
        const filterBy = buildFilterBy({
          name: "subject_type",
          label: "Subject Type",
          type: EColumnFilterType.EnumBased,
          selectOptions: subjectTypeOptions,
          selectedValues: [AUDIT_USER_SUBJECT_TYPE],
        });

        const result = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result.filterBy).toEqual([`subject_type==${AUDIT_USER_SUBJECT_TYPE}`]);
      });

      it("should expand 'App' subject_type to its legacy variants joined by ||", () => {
        const filterBy = buildFilterBy({
          name: "subject_type",
          label: "Subject Type",
          type: EColumnFilterType.EnumBased,
          selectOptions: subjectTypeOptions,
          selectedValues: [AUDIT_APP_SUBJECT_TYPE],
        });

        const result = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result.filterBy).toEqual([
          `subject_type==${AUDIT_APP_SUBJECT_TYPE}||${AUDIT_SERVICE_ACCOUNT_SUBJECT_TYPE}||${AUDIT_LEGACY_SERVICE_ACCOUNT_SUBJECT_TYPE}`,
        ]);
      });

      it("should combine User and Service account selections in one == clause", () => {
        const filterBy = buildFilterBy({
          name: "subject_type",
          label: "Subject Type",
          type: EColumnFilterType.EnumBased,
          selectOptions: subjectTypeOptions,
          selectedValues: [AUDIT_USER_SUBJECT_TYPE, AUDIT_APP_SUBJECT_TYPE],
        });

        const result = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result.filterBy).toEqual([
          `subject_type==${AUDIT_USER_SUBJECT_TYPE}||${AUDIT_APP_SUBJECT_TYPE}||${AUDIT_SERVICE_ACCOUNT_SUBJECT_TYPE}||${AUDIT_LEGACY_SERVICE_ACCOUNT_SUBJECT_TYPE}`,
        ]);
      });

      it("should not emit a filter when no values are selected", () => {
        const filterBy = buildFilterBy({
          name: "subject_type",
          label: "Subject Type",
          type: EColumnFilterType.EnumBased,
          selectOptions: subjectTypeOptions,
          selectedValues: [],
        });

        const result = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result.filterBy).toEqual([]);
      });

      it("should emit a != clause per unselected value for non-subject_type fields", () => {
        const filterBy = buildFilterBy({
          name: "result",
          label: "Status",
          type: EColumnFilterType.EnumBased,
          selectOptions: [
            { label: "Succeeded", value: "Succeeded" },
            { label: "Failed", value: "Failed" },
          ],
          selectedValues: ["Succeeded"],
        });

        const result = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result.filterBy).toEqual(["result!=Failed"]);
      });

      it("should drop subject_type values that are not in the allow-list", () => {
        const filterBy = buildFilterBy({
          name: "subject_type",
          label: "Subject Type",
          type: EColumnFilterType.EnumBased,
          selectOptions: subjectTypeOptions,
          selectedValues: ["Hacker", AUDIT_USER_SUBJECT_TYPE],
        });

        const result = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result.filterBy).toEqual([`subject_type==${AUDIT_USER_SUBJECT_TYPE}`]);
      });

      it("should emit nothing when every subject_type value is rejected", () => {
        const filterBy = buildFilterBy({
          name: "subject_type",
          label: "Subject Type",
          type: EColumnFilterType.EnumBased,
          selectOptions: subjectTypeOptions,
          selectedValues: ["Hacker", "Other"],
        });

        const result = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result.filterBy).toEqual([]);
      });

      it("should return [] when selectedValues is not an array", () => {
        const filterBy = buildFilterBy({
          name: "result",
          label: "Status",
          type: EColumnFilterType.EnumBased,
          selectOptions: [{ label: "Succeeded", value: "Succeeded" }],
          selectedValues: undefined as unknown as string[],
        });

        const result = filterService.mapColumnsFilterToFilterParams(filterBy);

        expect(result.filterBy).toEqual([]);
      });
    });
  });
  describe("FilterService::setColumnFilter", () => {
    it("should set column filter correctly", () => {
      const saveFiltersFilterSpy = vi.spyOn(filterService, "saveFilters");
      const filterBy: IFilterBy = {};

      filterService.setColumnFilter(filterBy, "name", "train", ETableFilters.PROJECT);
      expect(filterBy.columnFilters).toEqual([{ term: "name", name: "train", field: expect.anything() }]);
      expect(saveFiltersFilterSpy).toHaveBeenCalledOnce();
    });

    it("should set advanced column filter correctly", () => {
      const saveFiltersFilterSpy = vi.spyOn(filterService, "saveFilters");
      const filterBy: IFilterBy = {};
      const nodePoolColumnFilter: IFreeTextFilterModel = {
        name: WorkloadSortFilterFields.Nodepool,
        term: "default",
        value: EFilterOperator.Equals,
        label: "Node pool",
        type: EColumnFilterType.FreeText,
      };
      filterService.setColumnAdvancedFilter(filterBy, nodePoolColumnFilter, ETableFilters.NODE);
      expect(filterBy.advancedFilters).toEqual([
        {
          name: "nodepool",
          term: "default",
          value: "==",
          label: "Node pool",
          type: "FreeText",
        },
      ]);
      expect(saveFiltersFilterSpy).toHaveBeenCalledOnce();
    });

    it("should replace existing filter with new one", () => {
      const saveFiltersFilterSpy = vi.spyOn(filterService, "saveFilters");
      const filterBy: IFilterBy = {
        advancedFilters: [
          {
            name: "nodepool",
            term: "default",
            value: "==",
            label: "Node pool",
            type: "FreeText",
          } as IFreeTextFilterModel,
        ],
      };
      const nodePoolColumnFilter: IFreeTextFilterModel = {
        name: WorkloadSortFilterFields.Nodepool,
        term: "node pool a",
        value: EFilterOperator.Contains,
        label: "Node pool",
        type: EColumnFilterType.FreeText,
      };
      filterService.setColumnAdvancedFilter(filterBy, nodePoolColumnFilter, ETableFilters.NODE);
      expect(filterBy.advancedFilters).toEqual([
        {
          name: "nodepool",
          term: "node pool a",
          value: EFilterOperator.Contains,
          label: "Node pool",
          type: "FreeText",
        },
      ]);
      expect(saveFiltersFilterSpy).toHaveBeenCalledOnce();
    });
  });
  describe("FilterService::removeColumnFilter", () => {
    it("should remove column filter correctly", () => {
      const saveFiltersFilterSpy = vi.spyOn(filterService, "saveFilters");
      const filterBy: IFilterBy = {
        columnFilters: [
          {
            term: "train",
            name: "name",
            field: expect.anything(),
          },
        ],
      };

      filterService.removeColumnFilter(filterBy, "name", ETableFilters.PROJECT);
      expect(filterBy.columnFilters).toEqual([]);
      expect(saveFiltersFilterSpy).toHaveBeenCalledOnce();
    });
    it("should remove column filter correctly", () => {
      const saveFiltersFilterSpy = vi.spyOn(filterService, "saveFilters");
      const filterBy: IFilterBy = {
        advancedFilters: [
          {
            term: "train",
            name: "name",
            value: EFilterOperator.Contains,
            label: "Workload",
            type: EColumnFilterType.FreeText,
          },
        ],
      };

      filterService.removeAdvancedFilter(filterBy, "name", ETableFilters.PROJECT);
      expect(filterBy.advancedFilters).toEqual([]);
      expect(saveFiltersFilterSpy).toHaveBeenCalledOnce();
    });
    it("should not remove column filter correctly", () => {
      const filterBy: IFilterBy = {
        columnFilters: [
          {
            term: "train",
            name: "Name",
            field: expect.anything(),
          },
        ],
      };

      filterService.removeColumnFilter(filterBy, "name", ETableFilters.NODE);
      expect(filterBy.columnFilters).toEqual([
        {
          term: "train",
          name: "Name",
          field: expect.anything(),
        },
      ]);
    });
  });

  describe("FilterService::mapColumnsFilterToFilterParams with Date Filters", () => {
    it("should parse DD/MM/YYYY date format correctly", () => {
      const filterBy: IFilterBy = {
        sortBy: "createdAt",
        descending: false,
        page: 1,
        rowsPerPage: 20,
        columnFilters: [],
        searchTerm: "",
        advancedFilters: [
          {
            name: "createdAt",
            label: "Created At",
            type: EColumnFilterType.Date,
            date: "26/11/2025",
            value: EDateFilterOperator.On,
          } as IDateFilterModel,
        ],
        rowsNumber: 0,
      };

      const result: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy);

      expect(result.filterBy).toBeDefined();
      expect(result.filterBy).toHaveLength(2);
      expect(result.filterBy?.[0]).toContain("createdAt>=");
      expect(result.filterBy?.[1]).toContain("createdAt<=");
      // Verify the dates are valid ISO strings
      expect(result.filterBy?.[0]).toMatch(/createdAt>=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
      expect(result.filterBy?.[1]).toMatch(/createdAt<=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
    });

    it("should handle 'before' date filter correctly", () => {
      const filterBy: IFilterBy = {
        sortBy: "createdAt",
        descending: false,
        page: 1,
        rowsPerPage: 20,
        advancedFilters: [
          {
            name: "createdAt",
            label: "Created At",
            type: EColumnFilterType.Date,
            date: "01/01/2025",
            value: EDateFilterOperator.Before,
          } as IDateFilterModel,
        ],
      };

      const result: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy);

      expect(result.filterBy).toBeDefined();
      expect(result.filterBy).toHaveLength(1);
      expect(result.filterBy?.[0]).toContain("createdAt<=");
      // Check it's a valid ISO date format (timezone may vary)
      expect(result.filterBy?.[0]).toMatch(/createdAt<=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
    });

    it("should handle 'after' date filter correctly", () => {
      const filterBy: IFilterBy = {
        sortBy: "createdAt",
        descending: false,
        page: 1,
        rowsPerPage: 20,
        advancedFilters: [
          {
            name: "createdAt",
            label: "Created At",
            type: EColumnFilterType.Date,
            date: "15/06/2024",
            value: EDateFilterOperator.After,
          } as IDateFilterModel,
        ],
      };

      const result: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy);

      expect(result.filterBy).toBeDefined();
      expect(result.filterBy).toHaveLength(1);
      expect(result.filterBy?.[0]).toContain("createdAt>=");
      // Check it's a valid ISO date format (timezone may vary)
      expect(result.filterBy?.[0]).toMatch(/createdAt>=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
    });

    it("should return empty array for invalid date format", () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      const filterBy: IFilterBy = {
        sortBy: "createdAt",
        descending: false,
        page: 1,
        rowsPerPage: 20,
        advancedFilters: [
          {
            name: "createdAt",
            label: "Created At",
            type: EColumnFilterType.Date,
            date: "invalid-date",
            value: EDateFilterOperator.On,
          } as IDateFilterModel,
        ],
      };

      const result: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy);

      // Invalid date should be gracefully handled by returning empty filter
      expect(result.filterBy).toBeDefined();
      expect(result.filterBy).toHaveLength(0);
      errorSpy.mockRestore();
    });

    it("should return empty array for null/undefined date", () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      const filterBy: IFilterBy = {
        sortBy: "createdAt",
        descending: false,
        page: 1,
        rowsPerPage: 20,
        advancedFilters: [
          {
            name: "createdAt",
            label: "Created At",
            type: EColumnFilterType.Date,
            date: "",
            value: EDateFilterOperator.On,
          } as IDateFilterModel,
        ],
      };

      const result: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy);

      expect(result.filterBy).toBeDefined();
      expect(result.filterBy).toHaveLength(0);
      errorSpy.mockRestore();
    });
  });
});
