/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// services
import { storageUtil } from "@/utils/storage.util";
import { urlService } from "@/services/url.service/url.service";
import { getFieldContent } from "@/utils/table-format.util";
import { arrayToObjectMap, deepCopy } from "@/utils/common.util";

// models
import {
  EColumnFilterType,
  EFilterOperator,
  ESortOrder,
  type IAdvancedFilterOption,
  type IAssetsFilter,
  type IFilterBy,
  type IFilterModel,
  type IFilterOption,
  type IFreeTextFilterModel,
  type IPaginationFilter,
  type IAdvancedFilterModel,
  type IEnumBasedFilterModel,
  type IEnumBasedFilterSelectOption,
  type IDateFilterModel,
  EDateFilterOperator,
  type INumericFilterModel,
  numericFilterSupportedOperators,
  DEFAULT_PAGE_SIZE,
  CLUSTER_COLUMN_FILTER_NAME,
} from "@/models/filter.model";
import { ETableFilters, type ITableColumn, MULTI_CLUSTER_SUPPORTED_TABLE_FILTERS } from "@/models/table.model";

import { dateUtil } from "@/utils/date.util";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { CLUSTER_ID_FILTER } from "@/common/storage.constant";
import { AUDIT_APP_SUBJECT_TYPE, AUDIT_SERVICE_ACCOUNT_SUBJECT_TYPE_VALUES } from "@/models/audit.model";
import { DATE_FORMATS } from "@/constants/date-formats";
import type { ISelectOption } from "@/models/global.model";

export const filterService = {
  filterByColumns,
  getAvailableFilterOptions,
  getFilterOptions,
  getEmptyFilterByModel,
  loadFilters,
  saveFilters,
  filterBySearchTerm,
  getDefaultFilters,
  filterBySearchTermAndByColumns,
  mapColumnsFilterToFilterParams,
  setColumnFilter,
  filterListByTableFilters,
  getAdvancedFilterOptions,
  setColumnAdvancedFilter,
  removeColumnFilter,
  removeAdvancedFilter,
  getClusterUuidQueryFromColumnFilter,
  setClusterUuidForMultiClusterTableFilters,
  clearAllTablesFilters,
};

/**
 * @entities -  an array of departments/projects to be filtered.
 * @filters - an array of filters , each filter has a name, field and term
 * @entityColumnsModel - an array of tableColumns , in some fields the displayed data needs to be formatted, and the specific format function is on the column object.
 */
function filterByColumns<T>(
  entities: Array<T>,
  filters: Array<IFilterModel>,
  entityColumnsModel: Array<ITableColumn>,
): Array<T> {
  const columnsMap = arrayToObjectMap(entityColumnsModel, "name");
  return entities.filter((entity: T): boolean => {
    return filters.every((filter: IFilterModel): boolean => {
      const currCol: ITableColumn = columnsMap[filter.name];
      if (!currCol) {
        console.warn(`Column ${filter.name} not found in columns model`);
        return true;
      }
      // ignoring Column filter under specific circumstances
      if (currCol.ignoreFilter && currCol.ignoreFilter(entity)) return true;

      // the next line is for the case when reloading the
      // page and field is a function the stringify remove it,
      // so we need to return the field
      filter.field ??= currCol.field;
      let currField: string | number | object =
        typeof filter.field === "function" ? filter.field(entity) : entity[filter.field as keyof object];
      if (currCol.format) {
        if (currCol.filterKey) currField = currCol.format(currField, entity)[currCol.filterKey];
        else currField = currCol.format(currField, entity);
      }
      switch (typeof currField) {
        case "string":
          return currField.toLowerCase().includes(filter.term.toLowerCase());
        case "number":
          return currField.toString().includes(filter.term);
        case "object":
          if (Array.isArray(currField)) {
            return currField.some((item: string) => item.includes(filter.term));
          } else {
            return false;
          }
        default:
          return false;
      }
    });
  });
}

/**
 * @items -  an array of departments/projects to be filtered.
 * @term - a term to search through all the columns.
 * @displayedColumns - an array of the displayed columns in the table.
 * @columns - the full columns list
 */
function filterBySearchTerm<T>(
  items: Array<T>,
  term: string,
  displayedColumns: Array<string>,
  columns: Array<ITableColumn>,
): Array<T> {
  const columnsToFilter: Array<ITableColumn> = _getColumnsToFilter(displayedColumns, columns);
  return items.filter((item: T) => {
    const itemValues = _getSearchableValues<T>(item, columnsToFilter);
    return itemValues.some((value: string | number | boolean | Record<string, string>): boolean => {
      if (!value) return false;
      switch (typeof value) {
        case "string":
          return value.toLowerCase().includes(term.toLowerCase());
        case "number":
          return value === +term;
        case "boolean":
          return value === !!term;
        case "object": {
          if (value["filterKey"]) {
            // this serves for the cases where the column format returns an object and needs to know under which key the value is
            return value[value.filterKey].toLowerCase().includes(term.toLowerCase());
          }
          return false;
        }
      }
    });
  });
}

function _getSearchableValues<T>(item: T, columns: Array<ITableColumn>): Array<string | number | boolean> {
  const filterValues: Array<string | number | boolean> = columns.map((col: ITableColumn) => getFieldContent(col, item));
  return filterValues.flat();
}

function getAvailableFilterOptions<T extends IFilterOption | IAdvancedFilterOption>(
  activeFilters: Array<IFilterModel | IAdvancedFilterModel>,
  filterOptions: Array<T>,
): Array<T> {
  const filtersModelNamesMap: Map<string, string> = activeFilters.reduce(
    (acc, fm: IFilterModel | IAdvancedFilterModel) => {
      acc.set(fm.name, fm.name);
      return acc;
    },
    new Map(),
  );
  return filterOptions.filter((fo: T) => !filtersModelNamesMap.get(fo.name));
}

function getEmptyFilterByModel(options: IFilterBy): IFilterBy {
  const defaultFilterOptions = {
    sortBy: options.sortBy,
    descending: false,
    page: 1,
    rowsPerPage: DEFAULT_PAGE_SIZE,
    columnFilters: [],
    searchTerm: "",
  } as IFilterBy;

  return { ...defaultFilterOptions, ...options };
}

function saveFilters(tableName: ETableFilters, filters: IFilterBy): void {
  const searchParams: string = urlService.toSearchParams(filters);

  if (searchParams) {
    const url = window.location.href.split("?")[0];
    window.history.replaceState(window.history.state, "", `${url}?${searchParams}`);
  }

  storageUtil.save(tableName, filters);
}

function loadFilters(
  location: Location,
  tableName: ETableFilters,
  defaultFilterOptions: IFilterBy,
  forceCleanup = false,
): IFilterBy {
  const defaultsFilters: IFilterBy = getEmptyFilterByModel(defaultFilterOptions);

  if (forceCleanup) return { ...defaultsFilters };

  const storageFiltering: IFilterBy | null = storageUtil.get<IFilterBy | null>(tableName);
  const urlParamsFilters: IAssetsFilter = urlService.getFiltersObj(location);

  const displayedColumns: Array<string> = storageFiltering?.displayedColumns || defaultsFilters.displayedColumns || [];
  return urlParamsFilters && Object.keys(urlParamsFilters).length !== 0
    ? {
        ...defaultsFilters,
        ...storageFiltering,
        ...urlParamsFilters,
        displayedColumns,
      }
    : { ...defaultsFilters, ...storageFiltering };
}

function getDefaultFilters(sortBy: string, columns: Array<ITableColumn>, isTenantScope = false): IFilterBy {
  const defaultClusterId = storageUtil.get<string>(CLUSTER_ID_FILTER) || "";
  return {
    sortBy: sortBy,
    displayedColumns: columns.filter((col: ITableColumn) => col.display).map((col: ITableColumn) => col.name),
    ...(isTenantScope ? {} : { clusterUuid: defaultClusterId }),
  };
}

// Destructured the column object to a filter option object.
function getFilterOptions(columns: Array<ITableColumn>): Array<IFilterOption> {
  return columns.map(
    ({ field, label, name, filter }: ITableColumn): IFilterOption => ({
      field,
      label,
      name,
      ...(filter?.autocompleteOptionsFn && { autocompleteOptionsFn: filter.autocompleteOptionsFn }),
    }),
  );
}

function getAdvancedFilterOptions(columns: Array<ITableColumn>): Array<IAdvancedFilterOption> {
  return columns.map(
    ({ field, label, name, filter }: ITableColumn): IAdvancedFilterOption => ({
      field,
      label,
      name,
      type: filter?.type || EColumnFilterType.FreeText,
      ...(filter?.selectOptions && { selectOptions: filter.selectOptions }),
      ...(filter?.excludeOperators && { excludeOperators: filter.excludeOperators }),
      ...(filter?.inputType && { inputType: filter.inputType }),
      ...(filter?.maxWidth && { maxWidth: filter.maxWidth }),
      ...(filter?.autocompleteOptionsFn && { autocompleteOptionsFn: filter.autocompleteOptionsFn }),
    }),
  );
}

function _getColumnsToFilter(displayedColumns: Array<string>, columns: Array<ITableColumn>): Array<ITableColumn> {
  const displayedColumnsMap: Set<string> = new Set(displayedColumns);
  return columns.filter((col: ITableColumn) => displayedColumnsMap.has(col.name));
}

function filterBySearchTermAndByColumns<T>(
  entities: Array<T>,
  filterBy: IFilterBy,
  columns: Array<ITableColumn>,
): Array<T> {
  let entitiesFiltered = [...entities];

  if (filterBy.searchTerm && filterBy.displayedColumns) {
    entitiesFiltered = filterBySearchTerm(entities, filterBy.searchTerm, filterBy.displayedColumns, columns);
  }

  if (filterBy.columnFilters && filterBy.columnFilters.length) {
    entitiesFiltered = filterByColumns(entitiesFiltered, filterBy.columnFilters, columns);
  }
  return entitiesFiltered;
}

export function validateAdvancedFiltersFilterBy<T extends IFilterBy>(filterBy: T, validFields: string[]): void {
  // Validate sortBy
  if (filterBy.sortBy && !validFields.includes(filterBy.sortBy)) {
    filterBy.sortBy = undefined;
  }

  // Validate advancedFilters
  if (filterBy.advancedFilters) {
    filterBy.advancedFilters = filterBy.advancedFilters?.filter((filter) => validFields.includes(filter.name));
  }
}
/**
 * Maps advanced filter columns to filter parameters.
 *
 * @param {IFilterBy} filter - The filter object containing various filter criteria. For example: { clusterId: '123', searchTerm: 'runai', advancedFilters: [{ name: 'status', type: 'enum', selectedValues: ['running'] }] }
 * @param validFields - An array of valid fields that the advanced filters
 * @returns {string[]} An array of filter strings to be used in the query. for example: ['name=runai', 'clusterId=123']
 */
function mapAdvancedFilterColumnsToFilterParams(filter: IFilterBy, validFields?: string[]): string[] {
  const filterBy: string[] = [];
  // some api's use clusterUuid and some use clusterId, so we need to check which one to use
  if (validFields?.includes("clusterUuid") && filter.clusterUuid) {
    filterBy.push(filterUtil.getEqualsFilterString("clusterUuid", filter.clusterUuid));
  } else if (filter.clusterUuid) {
    filterBy.push(filterUtil.getEqualsFilterString("clusterId", filter.clusterUuid));
  }

  if (filter.advancedFilters) {
    filter.advancedFilters.forEach((advancedFilter: IAdvancedFilterModel) => {
      const filterParams = mapAdvancedFilterToFilterParams(advancedFilter);
      if (filterParams) {
        filterBy.push(...filterParams);
      }
    });
  }

  return filterBy;
}

function mapAdvancedFilterToFilterParams(advancedFilter: IAdvancedFilterModel): string[] | null {
  switch (advancedFilter.type) {
    case EColumnFilterType.FreeText:
      return _mapFreeTextFilterToFilterParams(advancedFilter as IFreeTextFilterModel);
    case EColumnFilterType.EnumBased:
      return _mapEnumBasedFilterToFilterParams(advancedFilter as IEnumBasedFilterModel);
    case EColumnFilterType.Date:
      return _mapDateFilterToFilterParams(advancedFilter as IDateFilterModel);
    case EColumnFilterType.Numeric:
      return _mapNumericFilterToFilterParams(advancedFilter as INumericFilterModel);
    default:
      return null;
  }
}

function _mapFreeTextFilterToFilterParams(filter: IFreeTextFilterModel): string[] {
  const term = filter.name === "source_ip" ? _escapeSourceIpTerm(filter.term) : filter.term;
  return [`${filter.name}${filter.value}${term}`];
}

// open api generated code uses ',' as the clause delimiter, so commas inside a source_ip value
// would split it. The solution is to escape the commas with ';' and swap them back in the service.
function _escapeSourceIpTerm(term: string): string {
  return term.replace(/,/g, ";");
}

function _mapEnumBasedFilterToFilterParams(filter: IEnumBasedFilterModel): string[] {
  if (!Array.isArray(filter.selectedValues) || filter.selectedValues.length === 0) return [];

  if (filter.name === "subject_type") {
    return _mapSubjectTypeFilterToFilterParams(filter);
  }

  const excludedOptions: IEnumBasedFilterSelectOption[] = filter.selectOptions.filter(
    (option: IEnumBasedFilterSelectOption) => !filter.selectedValues.includes(option.value as string),
  );
  return excludedOptions.flatMap((option: IEnumBasedFilterSelectOption) =>
    (option.matchValues ?? [String(option.value)]).map(
      (value: string) => `${filter.name}${EFilterOperator.NotEquals}${value}`,
    ),
  );
}

function _mapSubjectTypeFilterToFilterParams(filter: IEnumBasedFilterModel): string[] {
  const includedValues: ISelectOption[] = filter.selectOptions.filter((option: ISelectOption) =>
    filter.selectedValues.includes(option.value as string),
  );

  if (includedValues.length === 0) return [];

  const values: string[] = includedValues.flatMap((option: ISelectOption) => {
    const raw: string = String(option.value);
    return raw === AUDIT_APP_SUBJECT_TYPE ? AUDIT_SERVICE_ACCOUNT_SUBJECT_TYPE_VALUES : [raw];
  });

  return [`${filter.name}${EFilterOperator.Equals}${values.join("||")}`];
}

function _mapDateFilterToFilterParams(filter: IDateFilterModel): string[] {
  const filterDate: Date | null = dateUtil.parseToDate(filter.date, DATE_FORMATS.STANDARD_DATE);

  // If date parsing fails, return empty array to prevent crash
  if (!filterDate || isNaN(filterDate.getTime())) {
    console.error(`Failed to parse date: ${filter.date}`);
    return [];
  }

  const startOfDay: Date = dateUtil.getStartOfDay(filterDate);
  const endOfDay: Date = dateUtil.getEndOfDay(filterDate);

  const startOfDayString: string = startOfDay.toISOString();
  const endOfDayString: string = endOfDay.toISOString();

  switch (filter.value) {
    case EDateFilterOperator.Before:
      return [`${filter.name}${EFilterOperator.LessThanOrEqual}${startOfDayString}`];
    case EDateFilterOperator.After:
      return [`${filter.name}${EFilterOperator.GreaterThanOrEqual}${endOfDayString}`];
    case EDateFilterOperator.On:
      return [
        `${filter.name}${EFilterOperator.GreaterThanOrEqual}${startOfDayString}`,
        `${filter.name}${EFilterOperator.LessThanOrEqual}${endOfDayString}`,
      ];

    default:
      throw new Error(`Unsupported date filter value: ${filter.value}`);
  }
}

function _mapNumericFilterToFilterParams(filter: INumericFilterModel): string[] {
  const supportedOperators = new Set<string>(numericFilterSupportedOperators);
  if (!supportedOperators.has(filter.operator)) {
    throw new Error(`Unsupported numeric filter operator: ${filter.operator}`);
  }
  if (filter.operator === EFilterOperator.Range) {
    return [
      `${filter.name}${EFilterOperator.GreaterThanOrEqual}${filter.minValue}`,
      `${filter.name}${EFilterOperator.LessThanOrEqual}${filter.maxValue}`,
    ];
  }

  return [`${filter.name}${filter.operator}${filter.minValue}`];
}

/**
 * Maps advanced filter columns to filter parameters.
 *
 * @param {IFilterBy} filter - The filter object containing various filter criteria. For example: { clusterId: '123', searchTerm: 'runai', advancedFilters: [{ name: 'status', type: 'enum', selectedValues: ['running'] }] }
 * @param {string[]} [validFields] - An array of valid fields that the advanced filters can be applied on. if some of the advanced filters are not valid, they will be removed.
 * @param isAdvancedSearch
 * @returns {string[]} An array of filter strings to be used in the query. for example: ['name=runai', 'clusterId=123']
 */
function mapColumnsFilterToFilterParams(
  filter: IFilterBy,
  validFields?: string[],
  isAdvancedSearch?: boolean,
): IPaginationFilter {
  const sortOrder: ESortOrder = filter.descending ? ESortOrder.Desc : ESortOrder.Asc;

  if (validFields) {
    validateAdvancedFiltersFilterBy(filter, validFields);
  }

  const filterBy: string[] = mapAdvancedFilterColumnsToFilterParams(filter, validFields);
  const sortBy =
    filter.sortBy === CLUSTER_COLUMN_FILTER_NAME && validFields?.includes("clusterId") ? "clusterId" : filter.sortBy;

  const search = isAdvancedSearch && filter.searchTerm ? filter.searchTerm : undefined;
  const offset: number = ((filter.page || 0) - 1) * (filter.rowsPerPage || 0);
  return {
    offset: offset,
    limit: filter.rowsPerPage || DEFAULT_PAGE_SIZE,
    sortOrder: sortOrder,
    sortBy: sortBy,
    filterBy: filterBy,
    search: search,
  };
}

enum EColumnFilterKey {
  ColumnFilters = "columnFilters",
  AdvancedFilters = "advancedFilters",
}

function _initializeFilter<T extends IFilterModel | IAdvancedFilterModel>(
  filters: IFilterBy,
  filter: T,
  filterKey: EColumnFilterKey,
): void {
  (filters[filterKey] as T[]) = [filter];
}

function _updateFilter<T extends IFilterModel | IAdvancedFilterModel>(
  filters: IFilterBy,
  filter: T,
  filterKey: EColumnFilterKey,
  index: number,
): void {
  (filters[filterKey] as T[])[index] = filter;
}

function _addFilter<T extends IFilterModel | IAdvancedFilterModel>(
  filters: IFilterBy,
  filter: T,
  filterKey: EColumnFilterKey,
): void {
  (filters[filterKey] as T[]).push(filter);
}

function _setFilter<T extends IFilterModel | IAdvancedFilterModel>(
  filters: IFilterBy,
  filter: T,
  filterKey: EColumnFilterKey,
  filterStorageKey: ETableFilters,
): void {
  if (!filters[filterKey]) {
    _initializeFilter(filters, filter, filterKey);
  } else {
    const index = (filters[filterKey] as T[]).findIndex((colFilter: T) => colFilter.name === filter.name);
    if (index !== -1) {
      _updateFilter(filters, filter, filterKey, index);
    } else {
      _addFilter(filters, filter, filterKey);
    }
  }
  filterService.saveFilters(filterStorageKey, filters);
}

function setColumnFilter(filters: IFilterBy, term: string, name: string, filterStorageKey: ETableFilters) {
  const filter: IFilterModel = {
    term: term,
    name: name,
    field: () => "",
  };
  _setFilter(filters, filter, EColumnFilterKey.ColumnFilters, filterStorageKey);
}

function removeColumnFilter(filters: IFilterBy, name: string, filterStorageKey: ETableFilters) {
  if (!filters.columnFilters) return;
  filters.columnFilters = filters.columnFilters.filter((filter: IFilterModel) => filter.name !== name);
  filterService.saveFilters(filterStorageKey, filters);
}

function removeAdvancedFilter(filters: IFilterBy, name: string, filterStorageKey: ETableFilters) {
  if (!filters.advancedFilters) return;
  filters.advancedFilters = filters.advancedFilters.filter((filter: IAdvancedFilterModel) => filter.name !== name);
  filterService.saveFilters(filterStorageKey, filters);
}

function setColumnAdvancedFilter(filters: IFilterBy, filter: IAdvancedFilterModel, filterStorageKey: ETableFilters) {
  _setFilter(filters, filter, EColumnFilterKey.AdvancedFilters, filterStorageKey);
}

function filterListByTableFilters<T>(list: T[], filterBy: IFilterBy, columns: ITableColumn[]): T[] {
  if (!filterBy || !filterBy.displayedColumns) return list;

  let filteredList: T[] = deepCopy(list);
  if (filterBy.searchTerm) {
    filteredList = filterBySearchTerm<T>(filteredList, filterBy.searchTerm, filterBy.displayedColumns, columns);
  }

  if (filterBy.columnFilters && filterBy.columnFilters.length) {
    filteredList = filterService.filterByColumns(filteredList, filterBy.columnFilters, columns);
  }
  return filteredList;
}

function getClusterUuidQueryFromColumnFilter(filterBy: IFilterBy): { selectedClusterId?: string } {
  const clusterFilter = filterBy.clusterUuid;
  const query = clusterFilter ? { selectedClusterId: clusterFilter } : {};
  return query;
}

/**
 * Sets the cluster UUID for all multi-cluster supported table filters storage.
 * @param {string} clusterUuid - The UUID of the cluster to be set in the filters.
 */
function setClusterUuidForMultiClusterTableFilters(clusterUuid: string): void {
  MULTI_CLUSTER_SUPPORTED_TABLE_FILTERS.forEach((tableFilter) => {
    const storedFilter = storageUtil.get<IFilterBy>(tableFilter);
    if (storedFilter) {
      storageUtil.save(tableFilter, { ...storedFilter, clusterUuid });
    }
  });
}

function clearAllTablesFilters(): void {
  Object.keys(ETableFilters).forEach((tableFilter) => {
    storageUtil.remove(ETableFilters[tableFilter as keyof typeof ETableFilters]);
  });
}
