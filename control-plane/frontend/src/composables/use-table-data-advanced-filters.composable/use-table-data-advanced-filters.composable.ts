/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { computed, type ComputedRef, onUnmounted, ref, type Ref, watch } from "vue";
import { useQuasar } from "quasar";
import { useRoute } from "vue-router";
//model
import { CLUSTER_COLUMN_FILTER_NAME, type IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { type EIntervalLabels, EIntervalTime } from "@/models/interval.model";
import type { EAdvancedIndexPages } from "../helpers/use-table-data.mapper";
import { useAdvancedTableDataIndexMapper } from "../helpers/use-table-data.mapper";
//service
import { filterService } from "@/services/filter.service/filter.service";
//utils
import { deepCopy } from "@/utils/common.util";
import { intervalUtil } from "@/utils/interval.util";
import { ErrorAlert } from "@/utils/error-alert.util";
import { useClusterStore } from "@/stores/cluster.store";

/**
 * Composable function to manage table data and filters.
 *
 * @template T - The type of the table data.
 * @param {CallableFunction} loadListFunc - Function to load the list of entities.
 * @param {Array<ITableColumn>} columns - The columns of the table.
 * @param {EIndexPages} indexPage - The index page identifier.
 * @param {Function} getEntityById - Function to get an entity by its id, for last created entity handling.
 * @param {CallableFunction} [secondaryLoadFunc] - Optional function for loading additional info from other sources without delaying table loading.
 * @returns {Object} - An object containing methods and reactive properties for managing table data.
 */
export function useTableDataAdvancedFilters<T>(
  loadListFunc: CallableFunction,
  columns: ComputedRef<Array<ITableColumn>>,
  indexPage: EAdvancedIndexPages,
  getEntityById?: (id: string) => Promise<T>,
  secondaryLoadFunc?: CallableFunction,
) {
  const $q = useQuasar();
  const clusterStore = useClusterStore();
  const $route = useRoute();

  const tableData: Ref<T[]> = ref([]);
  const filterBy: Ref<IFilterBy> = ref({});
  const loadingTableData: Ref<boolean> = ref(true);
  const loadingError: Ref<boolean> = ref(false);
  const decelerateRefreshRateTimeoutId: Ref<null | number> = ref(null);
  const lastCreatedEntity: Ref<null | T> = ref(null);
  const pinnedEntityIds: Ref<Set<string>> = ref(new Set());
  // Kept outside tableData: a pinned row stops matching the active filter the moment it
  // starts deleting, and it has to stay on screen rather than silently disappear.
  const pinnedEntities: Ref<Map<string, T>> = ref(new Map()) as Ref<Map<string, T>>;
  const isFirstLoad = ref(true);
  const clusterId = computed(() => clusterStore.selectedClusterId);
  const selectedRows: Ref<T[]> = ref([]);

  const tableDataIndex = useAdvancedTableDataIndexMapper.get(indexPage);
  if (!tableDataIndex) {
    throw new Error(`Missing table data index configuration for page "${indexPage}"`);
  }
  const isAdvancedSearch = !!tableDataIndex.isAdvancedSearch;
  const intervalLabel: EIntervalLabels | undefined = tableDataIndex.intervalLabel;
  const intervalLabelPostCreation: EIntervalLabels | undefined = tableDataIndex.intervalLabelPostCreation;

  const listEntities = async (softLoad = false): Promise<T[]> => {
    try {
      const loadFunc = getLoadFunction();
      const currentFilterBy = deepCopy(filterBy.value);
      if (currentFilterBy.sortBy === CLUSTER_COLUMN_FILTER_NAME) {
        //we need to change the sortBy to clusterId for the API call
        currentFilterBy.sortBy = "clusterId";
      }
      isFirstLoad.value = false;
      const data = await loadFunc(currentFilterBy, softLoad);

      loadingError.value = false;
      return isAdvancedSearch ? data : filterBySearchTerm(data);
    } catch (error) {
      handleErrorResponse(error);
      return [];
    }
  };

  const handleErrorResponse = (error: unknown) => {
    const errorAlert = new ErrorAlert({
      generalMessage: ErrorAlert.failedLoadingMessage(tableDataIndex.entityText),
    });
    $q.notify(errorAlert.getNotification(error));
    loadingError.value = true;
  };

  const updateLastCreatedEntity = async (refreshedList: T[], externalLastCreatedId?: string) => {
    const lastCreatedId = externalLastCreatedId
      ? externalLastCreatedId
      : $route.query?.createdEntityId?.toString() || null;

    if (lastCreatedId) {
      const matchedEntity = refreshedList.find((e) => tableDataIndex.getEntityId(e) === lastCreatedId);
      if (matchedEntity !== undefined) {
        lastCreatedEntity.value = matchedEntity;
      } else if (getEntityById) {
        lastCreatedEntity.value = (await getEntityById(lastCreatedId)) || null;
      }
    }
  };

  const updatePinnedEntities = async (refreshedList: T[]) => {
    if (!pinnedEntityIds.value.size) return;

    const refreshed = new Map<string, T>();
    for (const id of pinnedEntityIds.value) {
      const matchedEntity = refreshedList.find((entity: T) => tableDataIndex.getEntityId(entity) === id);
      if (matchedEntity !== undefined) {
        refreshed.set(id, matchedEntity);
        continue;
      }
      if (!getEntityById) continue;
      try {
        const fetchedEntity = await getEntityById(id);
        if (fetchedEntity) refreshed.set(id, fetchedEntity);
      } catch {
        // gone from the backend, so it should stop being pinned
      }
    }
    pinnedEntityIds.value = new Set(refreshed.keys());
    pinnedEntities.value = refreshed;
  };

  const refreshList = async (withLoading = false, externalLastCreatedId?: string, softLoad = false) => {
    if (withLoading) loadingTableData.value = true;
    const refreshedList: T[] = await listEntities(softLoad);

    await updateLastCreatedEntity(refreshedList, externalLastCreatedId);
    await updatePinnedEntities(refreshedList);

    tableData.value = refreshedList;
    if (withLoading) loadingTableData.value = false;
    if (softLoad) return;
    updateCount();
  };

  const loadFilters = () => {
    const defaultFilters = filterService.getDefaultFilters(
      tableDataIndex.sortByDefault,
      columns.value,
      tableDataIndex.isTenantScope,
    );

    filterBy.value = filterService.loadFilters(window.location, tableDataIndex.storageFilterKey, defaultFilters);

    if (filterBy.value.clusterUuid) {
      filterBy.value.clusterUuid = clusterId.value;
    }
  };

  const setFilterBy = async (
    newFilter: IFilterBy,
    keyChanged: string | null = null,
    forceLoad = true,
    clearTopRow = true,
  ) => {
    filterBy.value = newFilter;
    const preventRefresh = !forceLoad || keyChanged === "displayedColumns";
    saveFiltersToStorage();
    if (preventRefresh) return;
    if (clearTopRow) clearLastCreated();
    await refreshList(true);
  };

  const updateCount = async () => {
    if (tableDataIndex.getEntityCount) {
      const filters = filterService.mapColumnsFilterToFilterParams(
        filterBy.value,
        tableDataIndex.validFilterFields,
        isAdvancedSearch,
      );
      const count = await tableDataIndex.getEntityCount(filters.filterBy, filters.search);
      updateFilterByKey("rowsNumber", count);
    }
  };

  const resetSelectedRows = () => (selectedRows.value = []);
  const selectedRowsAmount = computed(() => selectedRows.value.length);
  const selectedRow = computed(() => selectedRows.value[0] || null);
  /**
   * Updates a specific key in the filter criteria.
   *
   * @param {keyof IFilterBy} key - The key to update.
   * @param {IFilterBy[keyof IFilterBy]} value - The new value for the key.
   */
  const updateFilterByKey = <T extends keyof IFilterBy>(key: T, value: IFilterBy[T]) => {
    filterBy.value = { ...filterBy.value, [key]: value } as IFilterBy;
    saveFiltersToStorage();
  };

  /**
   * Updates a cell content in the table.
   *
   * @param {T} entity - The entity to update.
   * @param {keyof T} key - The key of the entity to update.
   * @param {T[keyof T]} newValue - The new value to update the entity with.
   */
  const updateCellContent = <K extends keyof T>(entity: T, key: K, newValue: T[K]) => {
    if (lastCreatedEntity.value && isLastCreatedEntity(entity)) {
      lastCreatedEntity.value = { ...lastCreatedEntity.value, [key]: deepCopy(newValue) } as T;
    } else {
      const rowIndex = getRowIndex(entity);
      const row = tableData.value.at(rowIndex);
      if (!row) throw new Error(`Invalid key: ${String(key)}`);
      tableData.value.splice(rowIndex, 1, { ...row, [key]: deepCopy(newValue) } as T);
    }
  };

  /**
   * Removes a row from the table.
   *
   * @param {T} entity - The entity to remove.
   */
  const removeRow = (entity: T) => {
    const rowIndex = getRowIndex(entity);
    tableData.value.splice(rowIndex, 1);
    clearLastCreated();
  };

  /**
   * Clears the filters and optionally sets default filters.
   *
   * @param {IFilterBy} [forcedDefaultFilter] - The default filter to set.
   */
  const clearFilters = (forcedDefaultFilter?: IFilterBy) => {
    if (forcedDefaultFilter) {
      setFilterBy(forcedDefaultFilter);
      return;
    }
    setFilterBy({ ...filterBy.value, advancedFilters: [], searchTerm: "" });
  };

  /**
   * Initializes the table filter with specific filter values.
   *
   * @param {IFilterBy} [specificFilterVals={}] - Specific filter values to set.
   */
  const initTableFilter = (specificFilterVals: IFilterBy = {}) => {
    loadFilters();
    setFilterBy({ ...filterBy.value, ...specificFilterVals }, null, true, false);
    if (intervalLabel) setRefreshRate();
  };

  const setRefreshRate = () => {
    if (!intervalLabel) return;
    if (!intervalLabelPostCreation || !$route.query.createdEntityId) {
      intervalUtil.startInterval(intervalLabel, refreshList);
      return;
    }
    intervalUtil.startInterval(intervalLabelPostCreation, refreshList);
    decelerateRefreshRateTimeoutId.value = window.setTimeout(() => {
      clearRefreshInterval();
      intervalUtil.startInterval(intervalLabel, refreshList);
    }, EIntervalTime.OneMinute);
  };

  const clearLastCreated = () => {
    if ($route.query.createdEntityId) delete $route.query.createdEntityId;
    lastCreatedEntity.value = null;
    pinnedEntityIds.value = new Set();
    pinnedEntities.value = new Map();
  };

  const clearRefreshInterval = () => {
    intervalLabel && intervalUtil.stopInterval(intervalLabel);
    intervalLabelPostCreation && intervalUtil.stopInterval(intervalLabelPostCreation);
  };

  const clearDecelerateTimeout = () => {
    if (decelerateRefreshRateTimeoutId.value) window.clearTimeout(decelerateRefreshRateTimeoutId.value);
  };

  onUnmounted(() => {
    clearRefreshInterval();
    clearDecelerateTimeout();
  });

  /**
   * Gets the row index of an entity.
   *
   * @param {T} entity - The entity to find.
   * @returns {number} - The row index of the entity.
   */
  const getRowIndex = (entity: T): number => {
    const rowIndex = tableData.value.findIndex(
      (e) => tableDataIndex.getEntityId(e) === tableDataIndex.getEntityId(entity),
    );
    if (rowIndex < 0 || rowIndex >= tableData.value.length) throw new Error("Invalid row index");
    return rowIndex;
  };

  const filterBySearchTerm = (data: T[]): T[] => {
    return filterService.filterBySearchTerm(
      data,
      filterBy.value.searchTerm || "",
      filterBy.value.displayedColumns || [],
      tableDataIndex.allEntityColumns,
    );
  };

  const isLastCreatedEntity = (entity: T): boolean => {
    if (!lastCreatedEntity.value) return false;
    return tableDataIndex.getEntityId(entity) === tableDataIndex.getEntityId(lastCreatedEntity.value);
  };

  const isPinned = (entity: T): boolean => pinnedEntityIds.value.has(tableDataIndex.getEntityId(entity));

  /**
   * Registers entities the user just acted on so they pin above the table body.
   *
   * Extends lastCreatedEntity from a single row to many while keeping the same
   * contract: pin what *this user* just did, never derive the pin from entity
   * state, or one user's action would pin rows on everyone else's screen.
   */
  const pinEntities = (entityIds: string[]) => {
    pinnedEntityIds.value = new Set([...pinnedEntityIds.value, ...entityIds]);

    // Seed from what is on screen so the rows stay visible until the next refresh.
    const seeded = new Map(pinnedEntities.value);
    for (const entity of tableData.value) {
      const id = tableDataIndex.getEntityId(entity);
      if (pinnedEntityIds.value.has(id)) seeded.set(id, entity);
    }
    pinnedEntities.value = seeded;
  };

  const pinnedRows: ComputedRef<T[]> = computed(() => [
    ...(lastCreatedEntity.value && !isPinned(lastCreatedEntity.value) ? [lastCreatedEntity.value] : []),
    ...pinnedEntities.value.values(),
  ]);

  const saveFiltersToStorage = () => {
    filterService.saveFilters(tableDataIndex.storageFilterKey, filterBy.value);
  };

  /**
   * Gets the appropriate load function.
   *
   * @returns {CallableFunction} - The load function.
   */
  const getLoadFunction = (): CallableFunction => {
    return secondaryLoadFunc && !isFirstLoad.value ? secondaryLoadFunc : loadListFunc;
  };

  const unwatch = watch(tableData, () => {
    if (secondaryLoadFunc) {
      refreshList(false, undefined, true);
    }
    unwatch();
  });

  return {
    initTableFilter,
    setFilterBy,
    updateFilterByKey,
    clearFilters,
    refreshList,
    tableData,
    filterBy,
    loadingTableData,
    loadingError,
    lastCreatedEntity,
    clearLastCreated,
    updateCellContent,
    removeRow,
    columns,
    clusterId,
    selectedRows,
    resetSelectedRows,
    selectedRowsAmount,
    selectedRow,
    pinnedRows,
    pinEntities,
  };
}
