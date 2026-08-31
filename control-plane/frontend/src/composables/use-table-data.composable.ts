/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { CLUSTER_COLUMN_FILTER_NAME, type IAssetsFilter } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { filterService } from "@/services/filter.service/filter.service";
import { computed, onUnmounted, ref, type Ref } from "vue";
import { useQuasar } from "quasar";
import { alertUtil } from "@/utils/alert.util";
import { type EIndexPages } from "./helpers/use-table-data.mapper";
import { useTableDataIndexMapper } from "./helpers/use-table-data.mapper";
import { useRoute } from "vue-router";
import { deepCopy } from "@/utils/common.util";
import { type EIntervalLabels, EIntervalTime } from "@/models/interval.model";
import { intervalUtil } from "@/utils/interval.util";
import { useClusterStore } from "@/stores/cluster.store";

/**
 * Composable function to manage table data and filters.
 *
 * @template T - The type of the table data.
 * @param {CallableFunction} loadListFunc - Function to load the list of entities.
 * @param {Array<ITableColumn>} tableColumns - Array of table columns.
 * @param {EIndexPages} indexPage - The index page identifier.
 * @param {CallableFunction} [secondaryLoadFunc] - Optional function for loading additional info from other sources without delaying table loading.
 * @returns {Object} - An object containing methods and reactive properties for managing table data.
 */
export function useTableData<T>(
  loadListFunc: CallableFunction,
  tableColumns: Array<ITableColumn>,
  indexPage: EIndexPages,
  secondaryLoadFunc?: CallableFunction,
) {
  const clusterStore = useClusterStore();
  const $q = useQuasar();

  const $route = useRoute();

  const tableData: Ref<T[]> = ref([]);
  const filterBy: Ref<IAssetsFilter> = ref({});
  const loadingTableData: Ref<boolean> = ref(false);
  const loadingError: Ref<boolean> = ref(false);
  const decelerateRefreshRateTimeoutId: Ref<null | number> = ref(null);
  const columns = ref(tableColumns);
  const lastCreatedEntity: Ref<null | T> = ref(null);
  const isFirstLoad = ref(true);
  const clusterId = computed(() => clusterStore.selectedClusterId);
  const tableDataIndex = useTableDataIndexMapper[indexPage];
  //interval
  const intervalLabel: EIntervalLabels | undefined = tableDataIndex.intervalLabel;
  const intervalLabelPostCreation: EIntervalLabels | undefined = tableDataIndex.intervalLabelPostCreation;

  const listEntities = async (withLoading = false): Promise<T[]> => {
    try {
      if (withLoading) loadingTableData.value = true;

      const loadFunc = getLoadFunc();
      if (isFirstLoad.value) {
        isFirstLoad.value = false;
      }

      const apiFilter = deepCopy(filterBy.value);
      if (!apiFilter.includeDescendants) delete apiFilter.clusterUuid;

      return await loadFunc(apiFilter);
    } catch (e: unknown) {
      console.error(e);
      $q.notify(alertUtil.getError(`Failed to load ${tableDataIndex.entityText}`));
      loadingError.value = true;
      throw e;
    } finally {
      if (withLoading) loadingTableData.value = false;
    }
  };

  const refreshList = async (withLoading = false, externalLastCreatedId?: string) => {
    const refreshedList: T[] = await listEntities(withLoading);

    const lastCreatedId = externalLastCreatedId
      ? externalLastCreatedId
      : $route.query.createdEntityId?.toString() || null;
    // if createdEntityId exists set the lastCreatedEntity and remove it from the list
    if (lastCreatedId) {
      const entityIndexInList: number = refreshedList.findIndex((e) => {
        const id = tableDataIndex.getEntityId(e);
        return id === lastCreatedId;
      });

      if (entityIndexInList !== -1) {
        lastCreatedEntity.value = refreshedList[entityIndexInList];
      } else {
        clearLastCreated();
      }
    }

    tableData.value = filterService.filterListByTableFilters(refreshedList, filterBy.value, columns.value);
  };

  const loadFilters = () => {
    const defaultFilters: IAssetsFilter = filterService.getDefaultFilters(tableDataIndex.sortByDefault, columns.value);

    filterBy.value = filterService.loadFilters(window.location, tableDataIndex.storageFilterKey, defaultFilters);

    if (filterBy.value.clusterUuid) {
      filterBy.value.clusterUuid = clusterId.value;

      const clusterFilterColExists = filterBy.value.columnFilters?.find((c) => c.name === CLUSTER_COLUMN_FILTER_NAME);
      if (clusterFilterColExists) {
        clusterFilterColExists.term = clusterId.value;
        return;
      }

      const clusterCol = columns.value.find((c) => c.name === CLUSTER_COLUMN_FILTER_NAME);
      if (!clusterCol) return;

      filterBy.value.columnFilters = [
        ...(filterBy.value.columnFilters || []),
        {
          field: clusterCol.field,
          name: CLUSTER_COLUMN_FILTER_NAME,
          term: clusterId.value,
        },
      ];
    }
  };

  const setFilterBy = async (
    newFilter: IAssetsFilter,
    keyChanged: null | string = null,
    forceLoad = true,
    clearTopRow = true,
  ) => {
    filterBy.value = newFilter;
    const preventRefresh = !forceLoad || keyChanged === "displayedColumns";
    filterService.saveFilters(tableDataIndex.storageFilterKey, filterBy.value);

    if (preventRefresh) return;
    if (clearTopRow) clearLastCreated();
    refreshList(true);
  };

  /*
   * @param entity - the entity to update
   * @param key - the key of the entity to update
   * @param newValue - the new value to update the entity with
   * use this function to update a cell content in the table, instead of reloading the whole table
   */
  const updateCellContent = <K extends keyof T>(entity: T, key: K, newValue: T[K]) => {
    if (lastCreatedEntity.value && _isLastCreatedEntity(entity)) {
      lastCreatedEntity.value[key] = deepCopy(newValue) as NonNullable<T>[K];
    } else {
      const rowIndex = _getRowIndex(entity);

      const row: T = tableData.value[rowIndex];

      if (!row) {
        throw new Error(`Invalid key: ${String(key)}`);
      }

      tableData.value.splice(rowIndex, 1, { ...row, [key]: deepCopy(newValue) } as T);
    }
  };

  /*
   * @param entity - the entity to remove from the table
   * use this function to remove a row from the table, instead of reloading the whole table
   * always remove the last created entity from the table upon deletion
   */
  const removeRow = (entity: T) => {
    if (!_isLastCreatedEntity(entity)) {
      const rowIndex = _getRowIndex(entity);
      tableData.value.splice(rowIndex, 1);
    }
    clearLastCreated();
  };

  const clearFilters = (forcedDefaultFilter?: IAssetsFilter) => {
    if (forcedDefaultFilter) {
      setFilterBy(forcedDefaultFilter);
      return;
    }
    setFilterBy({
      ...filterBy.value,
      columnFilters: [],
      searchTerm: "",
    });
  };

  // this func is essentialy the init func of the composable, its here for the purpose of passing special filters which are unique to specific apis (such as usageInfo: true etc...)
  const initTableFilter = (specificFilterVals: IAssetsFilter = {}) => {
    loadFilters();
    setFilterBy({ ...filterBy.value, ...specificFilterVals }, null, true, false);
    if (intervalLabel) _setRefreshRate();

    // we want to trigger refresh in order to get the full table data
    if (secondaryLoadFunc) refreshList();
  };

  const _setRefreshRate = () => {
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
  };

  const clearRefreshInterval = () => {
    intervalLabel && intervalUtil.stopInterval(intervalLabel);
    intervalLabelPostCreation && intervalUtil.stopInterval(intervalLabelPostCreation);
  };

  const clearDecelerateTimeout = () => {
    decelerateRefreshRateTimeoutId.value && window.clearTimeout(decelerateRefreshRateTimeoutId.value);
  };

  //columns
  const setColumns = (newColumns: Array<ITableColumn>) => {
    columns.value = newColumns;
  };

  onUnmounted(() => {
    clearRefreshInterval();
    clearDecelerateTimeout();
  });

  /*
   * helper functions
   */
  const _getRowIndex = (entity: T): number => {
    const rowIndex = tableData.value.findIndex((e: T) => {
      const id = tableDataIndex.getEntityId(e);
      return id === tableDataIndex.getEntityId(entity);
    });
    if (rowIndex < 0 || rowIndex >= tableData.value.length) {
      throw new Error("Invalid row index");
    }
    return rowIndex;
  };

  const _isLastCreatedEntity = (entity: T): boolean => {
    if (!lastCreatedEntity.value) return false;
    return tableDataIndex.getEntityId(entity) === tableDataIndex.getEntityId(lastCreatedEntity.value);
  };

  // Use secondaryLoadFunc for subsequent loads if provided
  const getLoadFunc = (): CallableFunction => {
    const shouldUseSecondaryLoadFunc = secondaryLoadFunc && !isFirstLoad.value;
    return shouldUseSecondaryLoadFunc ? secondaryLoadFunc : loadListFunc;
  };

  return {
    initTableFilter,
    setFilterBy,
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
    setColumns,
    columns,
  };
}
