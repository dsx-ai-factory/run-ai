<template>
  <div class="org-unit-quota-table">
    <div class="row items-center q-py-sm">
      <div class="text-b-subtitle2 q-mr-auto">
        You can adjust other {{ entityType }}s' quotas to free up additional resources.
      </div>
      <runai-general-search
        :search-term="filterBy.searchTerm || ''"
        @update-search-term="updateSearchTerm"
        :tooltip-text="searchTooltipText"
        :show-search-text="false"
      />
    </div>
    <runai-table
      no-shadow
      :columns="columnsWithConfig"
      :rows="rows"
      :loading="loading"
      :filter-by="filterBy"
      :get-row-key="(row: IQuotaTableRow) => row.id"
      :rows-per-page-options="[50]"
      rounded-corners
      disable-selection
      @update-filters="handleUpdateFilters"
    >
      <template #no-data>
        <runai-table-no-data
          :entity-name="entityType"
          :filter-by="filterBy"
          :show-create-btn="false"
          :show-filter-icon-and-button="false"
          :custom-message="emptyStateMessage"
        />
      </template>
    </runai-table>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RunaiTable } from "@/components/common/runai-table";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";

import { RunaiGeneralSearch } from "@/components/common/runai-general-search";
import {
  projectQuotaColumns,
  departmentQuotaColumns,
  type QuotaTableEntityType,
  type QuotaTableFetchContext,
  type IQuotaTableRow,
} from "@/table-models/org-unit.table-model";
import type { ITableColumn } from "@/models/table.model";
import { ECustomCell } from "@/models/table.model";
import type { IFilterBy, IPaginationFilter } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import {
  ProjectNodePoolResourcesFilterSortFields,
  DepartmentNodePoolResourcesFilterSortFields,
  type ProjectNodePoolResources,
  type DepartmentNodePoolResources,
} from "@/swagger-models/org-unit-service-client";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { EQuotaEntity, RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC } from "@/models/resource.model";

export interface Props {
  entityType: QuotaTableEntityType;
  fetchContext: QuotaTableFetchContext;
  availableQuota: number | null;
  currentEntityId: string; // ID of current project/department to exclude
  currentEntityQuota: number | null; // Current entity's GPU quota (the main project being edited)
  totalQuota: number | null; // Total quota for calculating distribution
  pendingChanges?: Map<string, { nodePoolName: string; newValue: number | null; originalValue: number | null }>;
  parentName?: string;
  // Initial total assigned quota from API (deservedSum for departments, projectsAggregatedResources for projects)
  // This is needed because rows are paginated and we can't calculate the total from rows alone
  serverTotalAssigned: number | null;
  // True when the "default" department has unlimited quota (-1)
  isDefaultDepartmentUnlimited?: boolean;
}
const props = defineProps<Props>();

const emit = defineEmits<{
  (
    e: "quota-changed",
    payload: {
      entityId: string;
      entityName: string;
      nodePoolName: string;
      nodePoolId: string;
      gpuQuota: number | null;
      originalValue: number | null;
      originalLimit: number | null;
      originalOverQuotaWeight: number | null;
    },
  ): void;
}>();

// Get columns based on entity type
const baseColumns = computed<ITableColumn[]>(() => {
  if (props.entityType === "project") {
    return projectQuotaColumns;
  } else {
    return departmentQuotaColumns;
  }
});

// State
const rawNodePoolResources = ref<Array<ProjectNodePoolResources | DepartmentNodePoolResources>>([]);
const loading = ref<boolean>(false);

const filterBy = ref<IFilterBy>(
  filterService.getEmptyFilterByModel({
    sortBy: ProjectNodePoolResourcesFilterSortFields.OrgUnitName,
    rowsPerPage: 50,
    page: 1,
  }),
);

const buildFilters = (): IPaginationFilter => {
  const nodePoolName = (props.fetchContext as { nodePoolName: string }).nodePoolName;

  const paginationFilters = filterService.mapColumnsFilterToFilterParams(filterBy.value, undefined, true);

  // Build base filters for node pool
  const baseFilters: string[] = [];

  if (props.entityType === "project") {
    const projectContext = props.fetchContext as { departmentId: string };
    baseFilters.push(
      filterUtil.getEqualsFilterString(ProjectNodePoolResourcesFilterSortFields.ParentId, projectContext.departmentId),
    );
    baseFilters.push(
      filterUtil.getEqualsFilterString(ProjectNodePoolResourcesFilterSortFields.NodePoolName, nodePoolName),
    );
  } else {
    // Department context: filter by clusterId and nodePoolName
    baseFilters.push(
      filterUtil.getEqualsFilterString(
        DepartmentNodePoolResourcesFilterSortFields.ClusterId,
        props.fetchContext.clusterId,
      ),
    );
    baseFilters.push(
      filterUtil.getEqualsFilterString(DepartmentNodePoolResourcesFilterSortFields.NodePoolName, nodePoolName),
    );
  }

  return {
    ...paginationFilters,
    filterBy: [...baseFilters, ...(paginationFilters.filterBy || [])],
  };
};

// Fetch data based on entity type
const fetchData = async () => {
  loading.value = true;
  try {
    await fetchCount();

    // Fetch data based on entity type
    const filters = buildFilters();
    let fetchedData: Array<ProjectNodePoolResources | DepartmentNodePoolResources> = [];

    if (props.entityType === "project") {
      const response = await orgUnitService.getProjectNodePoolResources(
        filters.filterBy,
        filters.sortBy as ProjectNodePoolResourcesFilterSortFields,
        filters.sortOrder,
        filters.offset,
        filters.limit,
        filters.search,
      );
      fetchedData = response.nodePoolResources;
    } else {
      const response = await orgUnitService.getDepartmentNodePoolResources(
        filters.filterBy,
        filters.sortBy as DepartmentNodePoolResourcesFilterSortFields,
        filters.sortOrder,
        filters.offset,
        filters.limit,
        filters.search,
      );
      fetchedData = response.nodePoolResources;
    }

    rawNodePoolResources.value = fetchedData;
  } catch (err) {
    console.error(err);
    rawNodePoolResources.value = [];
  } finally {
    loading.value = false;
  }
};

// Fetch count based on entity type
const fetchCount = async () => {
  try {
    const filters = buildFilters();
    let totalCount: number;

    if (props.entityType === "project") {
      totalCount = await orgUnitService.countProjectsNodePoolResources(filters.filterBy);
    } else {
      totalCount = await orgUnitService.countDepartmentNodePoolResources(filters.filterBy);
    }

    // Subtract 1 for current entity
    filterBy.value.rowsNumber = totalCount - 1;
  } catch (error) {
    console.error(error);
    filterBy.value.rowsNumber = 0;
  }
};

// Process node pool resources to add gpuQuota and quotaDistribution
const processNodePoolResources = (
  resources: Array<ProjectNodePoolResources | DepartmentNodePoolResources>,
): IQuotaTableRow[] => {
  const nodePoolName = (props.fetchContext as { nodePoolName: string }).nodePoolName;

  return resources
    .filter((resource) => resource.orgUnitId !== props.currentEntityId)
    .map((resource) => {
      const serverGpuQuota = resource.resources?.gpu?.deserved ?? null;
      const serverGpuLimit = resource.resources?.gpu?.limit ?? null;
      const serverOverQuotaWeight = resource.resources?.gpu?.overQuotaWeight ?? null;
      // For projects, get allocated non-preemptible GPUs
      const allocatedNonPreemptibleGpu =
        (resource as ProjectNodePoolResources)?.quotaStatus?.allocatedNonPreemptible?.gpu ?? undefined;
      // Check if there's a pending change for this entity
      const pendingChange = props.pendingChanges?.get(resource.orgUnitId);
      const gpuQuota =
        pendingChange && pendingChange.nodePoolName === nodePoolName ? pendingChange.newValue : serverGpuQuota;

      // Calculate quota distribution for visualization
      const totalQuota = props.totalQuota ?? 0;
      const quotaDistribution = totalQuota > 0 ? ((gpuQuota ?? 0) / totalQuota) * 100 : 0;

      return {
        id: resource.orgUnitId,
        name: resource.orgUnitName,
        gpuQuota,
        quotaDistribution,
        serverGpuQuota, // Store original server value for parent calculations
        serverGpuLimit, // Store original limit to preserve unlimited/custom limits
        serverOverQuotaWeight, // Store original over-quota weight to preserve settings
        allocatedNonPreemptibleGpu, // for projects only
      };
    });
};

// Compute rows from raw node pool resources - will react to pendingChanges
const rows = computed<IQuotaTableRow[]>(() => {
  return processNodePoolResources(rawNodePoolResources.value);
});

const handleUpdateFilters = (newFilterBy: IFilterBy) => {
  filterBy.value = newFilterBy;
  fetchData();
};

const updateSearchTerm = (searchTerm: string) => {
  filterBy.value = { ...filterBy.value, searchTerm };
  fetchData();
};

// For departments: unlimited (-1) quotas should be treated as 0 in calculations
const toQuotaValue = (value: number | null, isDepartment: boolean): number => {
  if (isDepartment && value === RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC) return 0;
  return value ?? 0;
};

// Calculate total assigned quota based on:
// - serverTotalAssigned: sum of OTHER entities' quotas (excluding current entity)
// - currentEntityQuota: current entity's quota (from main input)
// - pendingDelta: changes made in the table
// We can't calculate from rows because they're paginated (only shows current page)
const totalAssignedQuota = computed(() => {
  // Server total for other entities (excluding current entity)
  const serverTotal = props.serverTotalAssigned ?? 0;

  // Current entity's quota (from main input)
  const currentQuota = props.currentEntityQuota ?? 0;

  // Calculate delta from pending peer changes (changes made in the table)
  let pendingDelta = 0;
  if (props.pendingChanges) {
    const isDepartment = props.entityType === EQuotaEntity.department;
    const nodePoolName = (props.fetchContext as { nodePoolName: string }).nodePoolName;
    for (const change of props.pendingChanges.values()) {
      if (change.nodePoolName === nodePoolName) {
        const original = toQuotaValue(change.originalValue, isDepartment);
        const newVal = toQuotaValue(change.newValue, isDepartment);
        pendingDelta += newVal - original;
      }
    }
  }

  return serverTotal + currentQuota + pendingDelta;
});

// Empty state message based on filter state
const emptyStateMessage = computed(() => {
  if (filterBy.value.searchTerm) {
    return `No ${props.entityType}s match your filters`;
  }
  return `No other ${props.entityType}s yet`;
});

const searchTooltipText = computed(() => {
  return props.entityType === EQuotaEntity.project ? "Search projects" : "Search departments";
});

// Inject dynamic config into columns
const columnsWithConfig = computed<ITableColumn[]>(() => {
  const nodePoolName = (props.fetchContext as { nodePoolName: string }).nodePoolName;

  // Extract nodePoolId from the first entity's resource (same for all rows)
  const nodePoolId = rawNodePoolResources.value.length > 0 ? rawNodePoolResources.value[0].nodePoolId : "";

  return baseColumns.value.map((column) => {
    // Clone the column to avoid mutating the original
    const clonedColumn = { ...column };

    // If this is the GPU quota column, inject the custom cell config
    if (clonedColumn.customCell === ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL) {
      clonedColumn.customCellConfig = {
        [ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL]: {
          availableQuota: props.availableQuota,
          totalQuota: props.totalQuota,
          nodePoolName,
          nodePoolId,
          entityType: props.entityType,
          isDefaultDepartmentUnlimited: props.isDefaultDepartmentUnlimited,
          onSave: (payload: {
            projectId: string;
            entityName: string;
            nodePoolName: string;
            nodePoolId: string;
            gpuQuota: number | null;
            originalValue: number | null;
            originalLimit: number | null;
            originalOverQuotaWeight: number | null;
          }) => {
            emit("quota-changed", {
              entityId: payload.projectId,
              entityName: payload.entityName,
              nodePoolName: payload.nodePoolName,
              nodePoolId: payload.nodePoolId,
              gpuQuota: payload.gpuQuota,
              originalValue: payload.originalValue,
              originalLimit: payload.originalLimit,
              originalOverQuotaWeight: payload.originalOverQuotaWeight,
            });
          },
        },
      };
    }

    // If this is the quota distribution column, inject the custom cell config
    if (clonedColumn.customCell === ECustomCell.QUOTA_DISTRIBUTION_COL) {
      clonedColumn.customCellConfig = {
        [ECustomCell.QUOTA_DISTRIBUTION_COL]: {
          entityType: props.entityType,
          totalQuota: props.totalQuota,
          isUnlimited: props.totalQuota === null || props.totalQuota === undefined,
          isDefaultDepartmentUnlimited: props.isDefaultDepartmentUnlimited ?? false,
          parentName: props.parentName,
          totalAssignedQuota: totalAssignedQuota.value,
          nodePoolName,
          currentEntityId: props.currentEntityId,
        },
      };
    }

    return clonedColumn;
  });
});

onMounted(() => {
  fetchData();
});

// Expose refetch method and totalAssignedQuota - allows parent to trigger data refresh and access computed totals
defineExpose({
  refetch: fetchData,
  totalAssignedQuota,
});
</script>
