/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IFilterBy } from "@/models/filter.model";

export const urlService = {
  getFiltersObj,
  toSearchParams,
  getBackendBaseUrl,
  updateQueryParams,
};

function getBackendBaseUrl(): string {
  return `${window.location.origin}/`;
}

function getFiltersObj(location: Location): IFilterBy {
  const searchParamsStr: string = location.href.split("?")[1];
  const searchParams = new URLSearchParams(searchParamsStr);

  const query: string | null = searchParams.get("query");
  const page: number | null = Number(searchParams.get("page"));
  const rowsPerPage: number | null = Number(searchParams.get("rowsPerPage"));
  const sortBy: string | null = searchParams.get("sortBy");
  const descending: string | null = searchParams.get("descending");
  const columnFilters: string | null = searchParams.get("columnFilters");
  const advancedFilters: string | null = searchParams.get("advancedFilters");
  const searchTerm: string | null = searchParams.get("searchTerm");
  const clusterUuid: string | null = searchParams.get("clusterUuid");
  const retFilters: IFilterBy = {};
  query && (retFilters.query = query);
  page && (retFilters.page = page);
  rowsPerPage && (retFilters.rowsPerPage = rowsPerPage);
  sortBy && (retFilters.sortBy = sortBy);
  columnFilters && (retFilters.columnFilters = JSON.parse(columnFilters));
  advancedFilters && (retFilters.advancedFilters = JSON.parse(advancedFilters));
  searchTerm && (retFilters.searchTerm = searchTerm);
  clusterUuid && (retFilters.clusterUuid = clusterUuid);

  if (descending !== null) {
    retFilters.descending = JSON.parse(descending);
  }

  return retFilters;
}

function toSearchParams(filterBy: IFilterBy): string {
  const searchParams = new URLSearchParams();
  const excludedKeys = ["", undefined, "displayedColumns"];
  const objectKeys = ["columnFilters", "advancedFilters"];

  for (const [key, value] of Object.entries(filterBy)) {
    if (excludedKeys.includes(value) || value === "") continue;

    if (objectKeys.includes(key)) {
      if (Array.isArray(value) && value.length > 0) {
        searchParams.set(key, JSON.stringify(value));
      }
    } else {
      searchParams.set(key, `${value}`);
    }
  }

  return decodeURIComponent(searchParams.toString());
}

function updateQueryParams(params: { [key: string]: string | null }): void {
  const url = new URL(window.location.href);

  for (const key in params) {
    if (params[key] !== null) {
      url.searchParams.set(key, params[key] as string);
    } else {
      url.searchParams.delete(key);
    }
  }

  history.replaceState(history.state, "", url.toString());
}
