<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-card flat bordered class="q-mt-sm peer-org-unit-rank-list">
    <runai-page-actions
      :columns="columns"
      :filters="filterBy"
      :selected-rows-amount="selectedOrgUnits.length"
      @filters-changed="handleFiltersChanged"
      @selected-actions-close="resetSelection"
      :show-primary-btn="false"
      show-search
      hide-export-csv
      advanced-filters
      hide-cluster-filter
      hide-columns-btn
      hide-search-text
      :search-tooltip-text="searchTooltipText"
    >
      <template #selected-rows-actions>
        <span class="move-to-rank-btn-wrapper">
          <runai-action-button btn-action="moveToRank" @click="openMoveToRankModal" />
        </span>
      </template>
    </runai-page-actions>

    <q-linear-progress animation-speed="100" indeterminate size="xs" v-if="isLoadingCounts" />

    <rank-group-accordion
      v-if="!isLoadingCounts && rankGroups.length > 0"
      :rows="rankGroups"
      :entity-type="entityType"
      :expanded-rank="expandedRank"
      :current-org-unit-rank="effectiveCurrentOrgUnitRank"
      @rank-toggled="handleRankToggle"
    >
      <template #expanded-content="{ row }">
        <q-linear-progress animation-speed="100" indeterminate size="xs" v-if="isLoadingOrgUnits" />
        <rank-org-unit-table
          :org-units="getOrgUnitsForRank(row.rank)"
          :entity-type="entityType"
          :loading="isLoadingOrgUnits && expandedRank === row.rank"
          :node-pool-id="props.nodePoolId"
          :pending-weight-changes="weightChangesMap"
          :total-weight-in-rank="adjustedRankTotalWeights.get(row.rank) ?? 0"
          :is-over-quota-priority-enabled="isOverQuotaPriorityEnabled"
          :current-org-unit-id="props.currentOrgUnitId"
          :parent-filter-by="getFilterByForRank(row.rank)"
          v-model:selected="selectedOrgUnits"
          @weight-changed="handleWeightChanged"
          @update-filters="handleTableFiltersUpdate"
        />
      </template>
    </rank-group-accordion>

    <runai-table-no-data
      v-else-if="!isLoadingCounts"
      :entity-name="entityType"
      :filter-by="filterBy"
      :show-create-btn="false"
      :show-filter-icon-and-button="false"
    />

    <move-to-rank-modal
      v-if="isMoveToRankModalOpen"
      :selected-count="selectedOrgUnits.length"
      :entity-type="entityType"
      :current-rank="expandedRank ?? undefined"
      @close="closeMoveToRankModal"
      @move="handleMoveToRank"
    />
  </q-card>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RankGroupAccordion } from "./rank-group-accordion";
import { RankOrgUnitTable } from "./rank-org-unit-table";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { MoveToRankModal } from "./move-to-rank-modal";
import type { IFilterBy } from "@/models/filter.model";
import {
  peerProjectRankColumns,
  peerDepartmentRankColumns,
  EPeerOrgUnitRankColumnName,
} from "@/table-models/org-unit.table-model";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { ProjectNodePoolResources, DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client";
import {
  ProjectNodePoolResourcesFilterSortFields,
  DepartmentNodePoolResourcesFilterSortFields,
} from "@/swagger-models/org-unit-service-client";

// Union type to support both project and department node pool resources
export type OrgUnitNodePoolResources = ProjectNodePoolResources | DepartmentNodePoolResources;
import { filterUtil } from "@/utils/filter.util/filter.util";
import { filterService } from "@/services/filter.service/filter.service";
import { EEntityPriorityValue } from "@/models/org-unit.model";
import { EQuotaEntity } from "@/models/resource.model";
import { ErrorAlert } from "@/utils/error-alert.util";
import { useQuasar } from "quasar";
import { useSettingStore } from "@/stores/setting.store";

export interface Props {
  entityType?: EQuotaEntity;
  parentId?: string; // Parent org unit ID (e.g., Department ID for projects)
  nodePoolId: string; // Selected node pool ID
  clusterId: string;
  currentOrgUnitId?: string; // Current org unit ID being edited (for star indicator)
  currentOrgUnitRank?: string; // Rank of current org unit (for auto-expansion)
}

const $q = useQuasar();
const settingStore = useSettingStore();

const isOverQuotaPriorityEnabled = computed(() => settingStore.isOverQuotaPriorityEnabled);

const props = withDefaults(defineProps<Props>(), {
  entityType: EQuotaEntity.project,
  currentOrgUnitId: undefined,
  currentOrgUnitRank: undefined,
});

const emit = defineEmits<{
  (
    e: "current-entity-weight-changed",
    payload: { orgUnitId: string; nodePoolId: string; weight: number; originalWeight: number | null },
  ): void;
  (e: "pending-changes-update", hasPendingChanges: boolean): void;
}>();

// Get columns from table model based on entity type (for filters)
const columns = computed(() => {
  return props.entityType === EQuotaEntity.project ? peerProjectRankColumns : peerDepartmentRankColumns;
});

const searchTooltipText = computed(() => {
  return props.entityType === EQuotaEntity.project ? "Search projects" : "Search departments";
});

// Get filter sort fields based on entity type
const FilterSortFields = computed(() => {
  return props.entityType === EQuotaEntity.project
    ? ProjectNodePoolResourcesFilterSortFields
    : DepartmentNodePoolResourcesFilterSortFields;
});

// Define filters for runai-page-actions with advanced filters
const filterBy = ref<IFilterBy>(
  filterService.getEmptyFilterByModel({
    sortBy:
      props.entityType === EQuotaEntity.project
        ? ProjectNodePoolResourcesFilterSortFields.OrgUnitName
        : DepartmentNodePoolResourcesFilterSortFields.OrgUnitName,
    rowsPerPage: 10,
    page: 1,
    advancedFilters: [], // Required for advanced filters to work properly
  }),
);

const rankCounts = ref<Map<string, number>>(new Map());
const expandedRankOrgUnits = ref<OrgUnitNodePoolResources[]>([]);
const expandedRank = ref<string | null>(null);
const isLoadingCounts = ref(false);
const isLoadingOrgUnits = ref(false);
const selectedOrgUnits = ref<OrgUnitNodePoolResources[]>([]);
const isMoveToRankModalOpen = ref(false);

// Cache all loaded org units (keyed by orgUnitId)
const allLoadedOrgUnits = ref<Map<string, OrgUnitNodePoolResources>>(new Map());

// Cache total weight per rank (keyed by rank) - server values
const rankTotalWeights = ref<Map<string, number>>(new Map());

// Adjusted total weight per rank considering pending changes (rank moves + weight changes)
const adjustedRankTotalWeights = computed(() => {
  const adjusted = new Map<string, number>();

  // Start with server-fetched totals
  rankTotalWeights.value.forEach((total, rank) => {
    adjusted.set(rank, total);
  });

  // Apply pending changes
  allLoadedOrgUnits.value.forEach((orgUnit) => {
    const pendingChange = pendingOrgUnitChanges.value.get(orgUnit.orgUnitId);
    const originalRank = orgUnit.rank;
    const originalWeight = orgUnit.resources?.gpu?.overQuotaWeight ?? 0;

    if (pendingChange) {
      const newRank = pendingChange.newRank ?? originalRank;
      const newWeight = pendingChange.newWeight ?? originalWeight;

      // If rank changed, move weight between ranks
      if (pendingChange.newRank && pendingChange.newRank !== originalRank) {
        // Subtract original weight from old rank
        const oldRankTotal = adjusted.get(originalRank) ?? 0;
        adjusted.set(originalRank, Math.max(0, oldRankTotal - originalWeight));

        // Add new weight to new rank
        const newRankTotal = adjusted.get(newRank) ?? 0;
        adjusted.set(newRank, newRankTotal + newWeight);
      } else if (pendingChange.newWeight !== undefined) {
        // Only weight changed, adjust the current rank's total
        const rankTotal = adjusted.get(originalRank) ?? 0;
        const weightDelta = newWeight - originalWeight;
        adjusted.set(originalRank, Math.max(0, rankTotal + weightDelta));
      }
    }
  });

  return adjusted;
});

// Pending changes (rank and/or weight): Map<orgUnitId, { rank?, weight? }>
export interface PendingOrgUnitChange {
  originalRank?: string;
  newRank?: string;
  originalWeight?: number | null;
  newWeight?: number;
}
const pendingOrgUnitChanges = ref<Map<string, PendingOrgUnitChange>>(new Map());

// Get effective rank for current org unit (considering pending changes)
// This ensures the star indicator moves when the current entity is moved to a new rank
const effectiveCurrentOrgUnitRank = computed(() => {
  if (!props.currentOrgUnitId) return props.currentOrgUnitRank;
  const pendingChange = pendingOrgUnitChanges.value.get(props.currentOrgUnitId);
  return pendingChange?.newRank ?? props.currentOrgUnitRank;
});

// Extract weight changes for rank-org-unit-table (to maintain compatibility)
const weightChangesMap = computed(() => {
  const weightChanges = new Map<string, { newWeight: number; originalWeight: number | null }>();
  pendingOrgUnitChanges.value.forEach((change, orgUnitId) => {
    if (change.newWeight !== undefined) {
      weightChanges.set(orgUnitId, {
        newWeight: change.newWeight,
        originalWeight: change.originalWeight ?? null,
      });
    }
  });
  return weightChanges;
});

const handleFiltersChanged = (newFilters: IFilterBy) => {
  filterBy.value = newFilters;
  // Reload counts with new filters
  loadRankCounts();
  // If a rank is expanded, reload its org units with new filters
  if (expandedRank.value) {
    loadOrgUnitsForRank(expandedRank.value);
  }
};

// Build filters using mapColumnsFilterToFilterParams
const buildFilters = (rank?: string) => {
  const fields = FilterSortFields.value;
  const paginationFilters = filterService.mapColumnsFilterToFilterParams(
    filterBy.value,
    [fields.OrgUnitName, EPeerOrgUnitRankColumnName.Rank],
    true,
  );

  const filterByArray: string[] = [
    filterUtil.getEqualsFilterString(fields.NodePoolId, props.nodePoolId),
    filterUtil.getEqualsFilterString(fields.ClusterId, props.clusterId),
    ...(paginationFilters.filterBy || []),
  ];

  // Add parentId filter only if provided (used for projects)
  if (props.parentId && props.entityType === EQuotaEntity.project) {
    filterByArray.push(
      filterUtil.getEqualsFilterString(ProjectNodePoolResourcesFilterSortFields.ParentId, props.parentId),
    );
  }

  // Add rank filter if provided
  if (rank) {
    filterByArray.push(filterUtil.getEqualsFilterString("rank", rank));
  }

  return {
    ...paginationFilters,
    filterBy: filterByArray,
  };
};

// All possible ranks in priority order
const ALL_RANKS = [
  EEntityPriorityValue.Highest,
  EEntityPriorityValue.VeryHigh,
  EEntityPriorityValue.High,
  EEntityPriorityValue.MediumHigh,
  EEntityPriorityValue.Medium,
  EEntityPriorityValue.MediumLow,
  EEntityPriorityValue.Low,
  EEntityPriorityValue.VeryLow,
  EEntityPriorityValue.Lowest,
];

// Load rank counts using count API (more efficient than fetching all resources)
const loadRankCounts = async () => {
  if (!props.nodePoolId || !props.clusterId) {
    return;
  }

  isLoadingCounts.value = true;
  try {
    // Make parallel count requests for each rank using buildFilters
    const countPromises = ALL_RANKS.map(async (rank) => {
      const filters = buildFilters(rank);
      const count =
        props.entityType === EQuotaEntity.project
          ? await orgUnitService.countProjectsNodePoolResources(filters.filterBy, filters.search)
          : await orgUnitService.countDepartmentNodePoolResources(filters.filterBy, filters.search);
      return { rank, count };
    });

    const results = await Promise.all(countPromises);

    // Build counts map (only ranks with count > 0 will be displayed)
    const counts = new Map<string, number>();
    results.forEach(({ rank, count }) => {
      if (count > 0) {
        counts.set(rank, count);
      }
    });

    rankCounts.value = counts;
  } catch (error) {
    const errorAlert = new ErrorAlert({
      generalMessage: "Failed to load rank counts",
    });
    $q.notify(errorAlert.getNotification(error));
    rankCounts.value = new Map();
  } finally {
    isLoadingCounts.value = false;
  }
};

// Build filters for a specific rank (reusable)
const buildRankFilters = (rank: string): string[] => {
  const fields = FilterSortFields.value;
  const filters = [
    filterUtil.getEqualsFilterString(fields.ClusterId, props.clusterId),
    filterUtil.getEqualsFilterString("rank", rank),
  ];

  if (props.parentId) {
    // For projects, use ParentId; for departments, this is the cluster ID (already in filters)
    if (props.entityType === EQuotaEntity.project) {
      filters.push(filterUtil.getEqualsFilterString(ProjectNodePoolResourcesFilterSortFields.ParentId, props.parentId));
    }
  }

  return filters;
};

// Load total weight for a specific rank (called when rank is expanded)
const loadTotalWeightForRank = async (rank: string) => {
  if (!props.nodePoolId || !props.clusterId) {
    return;
  }

  try {
    const filterBy = buildRankFilters(rank);
    const response =
      props.entityType === EQuotaEntity.project
        ? await orgUnitService.sumProjectNodePoolResources(props.nodePoolId, filterBy)
        : await orgUnitService.sumDepartmentNodePoolResources(props.nodePoolId, filterBy);
    rankTotalWeights.value.set(rank, response.weightsSum ?? 0);
  } catch (error) {
    console.error(`Failed to load total weight for rank ${rank}:`, error);
    rankTotalWeights.value.set(rank, 0);
  }
};

// Helper to fetch node pool resources based on entity type
const fetchNodePoolResources = (
  filterBy: string[],
  sortField: string,
  offset?: number,
  limit?: number,
  search?: string,
) => {
  return props.entityType === EQuotaEntity.project
    ? orgUnitService.getProjectNodePoolResources(
        filterBy,
        sortField as ProjectNodePoolResourcesFilterSortFields,
        "desc",
        offset,
        limit,
        search,
      )
    : orgUnitService.getDepartmentNodePoolResources(
        filterBy,
        sortField as DepartmentNodePoolResourcesFilterSortFields,
        "desc",
        offset,
        limit,
        search,
      );
};

// Get the sort field based on entity type and over-quota priority setting
const getSortField = () => {
  const isProject = props.entityType === EQuotaEntity.project;
  if (isOverQuotaPriorityEnabled.value) {
    return isProject
      ? ProjectNodePoolResourcesFilterSortFields.GpuOverQuotaWeight
      : DepartmentNodePoolResourcesFilterSortFields.GpuOverQuotaWeight;
  }
  return isProject
    ? ProjectNodePoolResourcesFilterSortFields.GpuDeserved
    : DepartmentNodePoolResourcesFilterSortFields.GpuDeserved;
};

// Get the orgUnitId filter field based on entity type
const getOrgUnitIdField = () =>
  props.entityType === EQuotaEntity.project
    ? ProjectNodePoolResourcesFilterSortFields.OrgUnitId
    : DepartmentNodePoolResourcesFilterSortFields.OrgUnitId;

// Load org units for a specific rank (called when rank is expanded)
// Ensures the current entity is always visible (pinned at top) to avoid pagination issues
const loadOrgUnitsForRank = async (rank: string) => {
  if (!props.nodePoolId || !props.clusterId) {
    return;
  }

  isLoadingOrgUnits.value = true;
  try {
    const filters = buildFilters(rank);
    const sortField = getSortField();
    const shouldPinCurrentEntity = props.currentOrgUnitId && effectiveCurrentOrgUnitRank.value === rank;

    let orgUnits: OrgUnitNodePoolResources[] = [];

    if (shouldPinCurrentEntity && props.currentOrgUnitId) {
      // Fetch current entity (pinned) and others (paginated) in parallel
      const orgUnitIdField = getOrgUnitIdField();
      const currentOrgUnitId = props.currentOrgUnitId;
      const [currentResponse, paginatedResponse] = await Promise.all([
        // Fetch the current entity
        fetchNodePoolResources(
          [...filters.filterBy, filterUtil.getEqualsFilterString(orgUnitIdField, currentOrgUnitId)],
          sortField,
          0,
          1,
          filters.search,
        ),
        // Fetch paginated entities excluding the current one
        fetchNodePoolResources(
          [...filters.filterBy, filterUtil.getNotEqualsFilterString(orgUnitIdField, currentOrgUnitId)],
          sortField,
          filters.offset,
          filters.limit,
          filters.search,
        ),
      ]);

      const currentEntity = currentResponse.nodePoolResources?.[0];
      const paginatedEntities = paginatedResponse.nodePoolResources ?? [];
      orgUnits = currentEntity ? [currentEntity, ...paginatedEntities] : paginatedEntities;
    } else {
      // No pinning needed, fetch normally
      const response = await fetchNodePoolResources(
        filters.filterBy,
        sortField,
        filters.offset,
        filters.limit,
        filters.search,
      );
      orgUnits = response.nodePoolResources ?? [];
    }

    expandedRankOrgUnits.value = orgUnits;

    // Cache all loaded org units
    orgUnits.forEach((orgUnit) => {
      allLoadedOrgUnits.value.set(orgUnit.orgUnitId, orgUnit);
    });
  } catch (error) {
    console.error(`Failed to load org units for rank ${rank}:`, error);
    expandedRankOrgUnits.value = [];
  } finally {
    isLoadingOrgUnits.value = false;
  }
};

// Load rank data (org units + total weight) - reusable function
const loadRankData = async (rank: string, shouldSelectCurrentOrgUnit = false) => {
  await Promise.all([loadOrgUnitsForRank(rank), loadTotalWeightForRank(rank)]);
  if (shouldSelectCurrentOrgUnit) {
    autoSelectCurrentOrgUnit();
  }
};

// Auto-select the current org unit (the one being edited) so user is aware they can move ranks
const autoSelectCurrentOrgUnit = () => {
  if (!props.currentOrgUnitId) return;

  const alreadySelected = selectedOrgUnits.value.some((orgUnit) => orgUnit.orgUnitId === props.currentOrgUnitId);
  if (alreadySelected) return;

  const currentOrgUnit = expandedRankOrgUnits.value.find((orgUnit) => orgUnit.orgUnitId === props.currentOrgUnitId);
  if (currentOrgUnit) {
    selectedOrgUnits.value = [currentOrgUnit, ...selectedOrgUnits.value];
  }
};

// Helper: Get the effective rank for an org unit (considering pending changes)
const getEffectiveRank = (orgUnitId: string, originalRank?: string): string | undefined => {
  const pendingChange = pendingOrgUnitChanges.value.get(orgUnitId);
  return pendingChange?.newRank ?? originalRank;
};

// Adjusted rank counts considering pending changes (used for pagination and display)
const adjustedRankCounts = computed(() => {
  const adjusted = new Map(rankCounts.value);

  // Adjust counts based on pending rank changes
  pendingOrgUnitChanges.value.forEach((change) => {
    const { originalRank, newRank } = change;

    if (originalRank && newRank && originalRank !== newRank) {
      // Decrease count in original rank
      const originalCount = adjusted.get(originalRank) || 0;
      adjusted.set(originalRank, Math.max(0, originalCount - 1));

      // Increase count in new rank
      const newCount = adjusted.get(newRank) || 0;
      adjusted.set(newRank, newCount + 1);
    }
  });

  return adjusted;
});

// Get filterBy for a specific rank (includes rowsNumber for pagination)
const getFilterByForRank = (rank: string): IFilterBy => {
  return {
    ...filterBy.value,
    rowsNumber: adjustedRankCounts.value.get(rank) ?? 0,
  };
};

// Group org units by rank and count them (using adjusted counts)
const rankGroups = computed(() => {
  return ALL_RANKS.filter((rank) => {
    const count = adjustedRankCounts.value.get(rank) || 0;
    return count > 0;
  }).map((rank) => ({
    rank,
    count: adjustedRankCounts.value.get(rank) || 0,
  }));
});

// Helper: Check if an org unit matches the current search/filter
const matchesSearch = (orgUnit: OrgUnitNodePoolResources): boolean => {
  // Check global search term
  const searchTerm = filterBy.value.searchTerm?.toLowerCase();
  if (!searchTerm) return true;

  // Check against name (add other fields if needed)
  return orgUnit.orgUnitName?.toLowerCase().includes(searchTerm) ?? false;
};

// Get org units for a specific rank (only returns data for the currently expanded rank)
// Includes org units moved into this rank and excludes those moved out
// Always places the current entity first in the list
const getOrgUnitsForRank = (rank: string) => {
  if (rank !== expandedRank.value) {
    return [];
  }

  // 1. Start with the server-fetched, filtered list for this rank
  // This list respects the search term applied on the backend
  const baseList = expandedRankOrgUnits.value;
  const result: OrgUnitNodePoolResources[] = [];
  const processedIds = new Set<string>();
  let currentEntity: OrgUnitNodePoolResources | null = null;

  // 2. Process server items: Keep them unless they moved OUT
  baseList.forEach((orgUnit) => {
    const effectiveRank = getEffectiveRank(orgUnit.orgUnitId, orgUnit.rank);
    // If effective rank is still this rank, keep it
    if (effectiveRank === rank) {
      // Check if this is the current entity - save it for later
      if (orgUnit.orgUnitId === props.currentOrgUnitId) {
        currentEntity = orgUnit;
      } else {
        result.push(orgUnit);
      }
    }
    processedIds.add(orgUnit.orgUnitId);
  });

  // 3. Process Pending Changes: Add items moved INTO this rank
  pendingOrgUnitChanges.value.forEach((change, orgUnitId) => {
    // Only check if newRank matches current rank AND we haven't processed it yet
    if (change.newRank === rank && !processedIds.has(orgUnitId)) {
      const orgUnit = allLoadedOrgUnits.value.get(orgUnitId);

      // Critical: Since these items come from cache (potentially from a different search context),
      // we MUST validate they match the current client-side search term.
      if (orgUnit && matchesSearch(orgUnit)) {
        // Check if this is the current entity - save it for later
        if (orgUnitId === props.currentOrgUnitId) {
          currentEntity = orgUnit;
        } else {
          result.push(orgUnit);
        }
        processedIds.add(orgUnitId);
      }
    }
  });

  // 4. Always place current entity first if it belongs to this rank
  if (currentEntity) {
    return [currentEntity, ...result];
  }

  return result;
};

// Handle rank expand/collapse
const handleRankToggle = async (payload: { rank: string; isExpanded: boolean }) => {
  resetSelection();

  if (payload.isExpanded) {
    // Expanding - load org units and total weight for this rank
    expandedRank.value = payload.rank;
    await loadRankData(payload.rank);
  } else {
    // Collapsing - clear expanded data
    expandedRank.value = null;
    expandedRankOrgUnits.value = [];
  }
};

// Selection management
const resetSelection = () => {
  selectedOrgUnits.value = [];
};

const openMoveToRankModal = () => {
  isMoveToRankModalOpen.value = true;
};

const closeMoveToRankModal = () => {
  isMoveToRankModalOpen.value = false;
};

const handleMoveToRank = async (newRank: string) => {
  // Add to pending changes for each selected org unit
  selectedOrgUnits.value.forEach((orgUnit) => {
    if (orgUnit.rank) {
      const existingChange = pendingOrgUnitChanges.value.get(orgUnit.orgUnitId);

      // If there's already a pending change, preserve the original rank
      const originalRank = existingChange?.originalRank ?? orgUnit.rank;

      // If moving back to original rank and no weight changes, remove the pending change entirely
      if (originalRank === newRank && !existingChange?.newWeight) {
        pendingOrgUnitChanges.value.delete(orgUnit.orgUnitId);
      } else {
        // Otherwise, update or create the pending change
        pendingOrgUnitChanges.value.set(orgUnit.orgUnitId, {
          ...existingChange,
          originalRank,
          newRank,
        });
      }
    }
  });

  // Reset selection and close modal
  resetSelection();
  closeMoveToRankModal();

  // Auto-expand the destination rank to show moved items
  if (expandedRank.value !== newRank) {
    expandedRank.value = newRank;
    await loadRankData(newRank);
  }
};

// Handle weight change from editable weight column
const handleWeightChanged = (payload: {
  orgUnitId: string;
  nodePoolId: string;
  weight: number;
  originalWeight: number | null;
}) => {
  const existingChange = pendingOrgUnitChanges.value.get(payload.orgUnitId);

  // If there's already a pending change, preserve the original weight
  const originalWeight = existingChange?.originalWeight ?? payload.originalWeight;

  // If weight is back to original and no rank changes, remove the pending change entirely
  if (originalWeight === payload.weight && !existingChange?.newRank) {
    pendingOrgUnitChanges.value.delete(payload.orgUnitId);
  } else {
    // Notify parent if this is the current entity being edited
    if (payload.orgUnitId === props.currentOrgUnitId) {
      emit("current-entity-weight-changed", payload);
    }
    // Otherwise, update or create the pending change
    pendingOrgUnitChanges.value.set(payload.orgUnitId, {
      ...existingChange,
      originalWeight,
      newWeight: payload.weight,
    });
  }
};

// Handle filter updates from child table (pagination changes)
const handleTableFiltersUpdate = async (newFilterBy: IFilterBy) => {
  // Update parent's filterBy with new pagination values
  filterBy.value = {
    ...filterBy.value,
    page: newFilterBy.page,
    rowsPerPage: newFilterBy.rowsPerPage,
  };

  // Refetch data for the expanded rank
  if (expandedRank.value) {
    await loadOrgUnitsForRank(expandedRank.value);
  }
};

// Reset all component state
const resetState = () => {
  expandedRank.value = null;
  expandedRankOrgUnits.value = [];
  allLoadedOrgUnits.value.clear();
  pendingOrgUnitChanges.value.clear();
  rankTotalWeights.value.clear();
  selectedOrgUnits.value = [];
};

// Notify parent when pending changes status changes
watch(
  () => pendingOrgUnitChanges.value.size,
  (newSize, oldSize) => {
    const hasPending = newSize > 0;
    const hadPending = oldSize > 0;
    if (hasPending !== hadPending) {
      emit("pending-changes-update", hasPending);
    }
  },
);

const initializeRankList = async () => {
  await loadRankCounts();

  // Auto-expand the rank that contains the current org unit and auto-select it
  if (props.currentOrgUnitRank && rankCounts.value.has(props.currentOrgUnitRank)) {
    expandedRank.value = props.currentOrgUnitRank;
    await loadRankData(props.currentOrgUnitRank, true);
  }
};

watch(
  () => [props.parentId, props.nodePoolId, props.clusterId],
  async () => {
    resetState();
    await initializeRankList();
  },
  { immediate: true },
);

const refetch = async () => {
  await loadRankCounts();
  if (expandedRank.value) {
    await loadRankData(expandedRank.value);
  }
};

// Expose methods and data for parent component
defineExpose({
  getPendingOrgUnitChanges: () => pendingOrgUnitChanges.value,
  hasPendingChanges: () => pendingOrgUnitChanges.value.size > 0,
  clearPendingChanges: () => {
    pendingOrgUnitChanges.value.clear();
    selectedOrgUnits.value = [];
  },
  refetch,
});
</script>

<style lang="scss">
.peer-org-unit-rank-list {
  .right-actions-container {
    // in this specific component, we need to adjust the padding.
    padding-right: 10px;
  }
}
</style>
