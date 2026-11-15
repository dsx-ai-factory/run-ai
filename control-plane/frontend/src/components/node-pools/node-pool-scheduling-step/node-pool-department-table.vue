<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <section class="department-section" aria-label="Department scheduling">
    <notification-alert-bar
      v-if="quotaExcess != null"
      class="quota-warning"
      role="status"
      :variant="ENotificationBarVariant.Warning"
    >
      The total assigned quota exceeds the node pool's total by {{ quotaExcess }} GPUs. Departments may not receive their
      full guaranteed quota.
    </notification-alert-bar>
    <runai-page-actions
      class="scheduling-actions"
      :class="{ 'actions-readonly': readonly }"
      :columns="filterColumns"
      :filters="{ ...filters, searchTerm }"
      advanced-filters
      search-tooltip-text="Search departments"
      filter-menu-width="208px"
      filter-menu-max-height="280px"
      hide-cluster-filter
      hide-columns-btn
      hide-export-csv
      :show-search="canView"
      :support-refresh="canView"
      :primary-btn-disable="readonly"
      :selected-rows-amount="ranked ? selected.size : 0"
      :disable-actions="readonly"
      @filters-changed="updateFilters"
      @refresh-data="emit('refresh')"
      @selected-actions-close="selected = new Set()"
    >
      <template #primary-btn
        ><runai-btn
          aid="new-department"
          color="primary"
          icon="ph-bold ph-plus"
          label="New department"
          :disable="readonly"
          @click="emit('new-department')"
      /></template>
      <template #selected-rows-actions
        ><runai-action-button btn-action="moveToRank" aid="move-departments" @click="showMove = true"
      /></template>
      <template #right-actions
        ><q-btn
          v-if="canView"
          aid="advanced-settings"
          flat
          stack
          size="sm"
          icon="ph-bold ph-gear-six"
          :label="props.searchTerm || noMatchingDepartments ? 'Show settings' : 'Advanced settings'"
          :disable="readonly"
          @click="emit('advanced-settings')"
          ><q-menu
            :model-value="settingsOpen"
            no-parent-event
            anchor="bottom right"
            self="top right"
            :offset="[0, 8]"
            max-height="70vh"
            class="scheduling-settings-popup"
            @update:model-value="!$event && emit('close-settings')"
          >
            <slot name="settings" /> </q-menu></q-btn
      ></template>
    </runai-page-actions>
    <div class="table-settings-layout">
      <div
        class="department-table-content"
        :style="preserveTableSpace ? { minHeight: `${unfilteredHeight}px` } : undefined"
      >
        <q-resize-observer @resize="rememberTableHeight" />
        <q-table
          flat
          square
          class="scheduling-table"
          :rows="canView ? displayRows : []"
          :columns="columns"
          row-key="key"
          :loading="loading"
          :pagination="{ rowsPerPage: 0 }"
          hide-pagination
          :rows-per-page-options="[0]"
          :style="{ '--scheduling-table-min-width': draft.maxAllocation ? '998px' : '960px' }"
        >
          <template #header="{ cols }"
            ><q-tr>
              <q-th v-for="col in cols" :key="col.name" :class="`column-${col.name}`">
                <template v-if="col.name === 'selection'">
                  <q-checkbox
                    v-if="ranked && selected.size"
                    :model-value="allSelected"
                    :disable="readonly"
                    aria-label="Select all departments"
                    @update:model-value="selectAll"
                  />
                  <q-icon v-else name="ph-fill ph-question" size="20px"
                    ><q-tooltip>{{ headerTooltips.selection }}</q-tooltip></q-icon
                  >
                </template>
                <template v-else>
                  <div v-if="order(col.name)" :class="`column-order order-${col.name}`">{{ order(col.name) }}</div>
                  <div class="column-label">
                    {{ col.label }}
                    <q-icon v-if="col.name !== 'department' || ranked" name="ph-bold ph-question" size="16px"
                      ><q-tooltip max-width="340px">{{ headerTooltips[col.name] }}</q-tooltip></q-icon
                    >
                  </div>
                  <div v-if="col.name === 'result'" class="simulation-control">
                    Simulate workload demand
                    <q-toggle
                      v-model="simulation"
                      dense
                      size="30px"
                      :disable="!canView"
                      aid="simulate-demand"
                      aria-label="Simulate workload demand"
                    />
                  </div>
                </template>
              </q-th> </q-tr
          ></template>
          <template #body="{ row }">
            <q-tr v-if="row.group" class="rank-header">
              <q-td :colspan="columns.length"
                ><div class="rank-row">
                  <q-btn
                    flat
                    round
                    dense
                    :icon="expanded.has(row.rank) ? 'ph-bold ph-caret-up' : 'ph-bold ph-caret-down'"
                    :aid="`rank-${row.rank}`"
                    :aria-label="`${expanded.has(row.rank) ? 'Collapse' : 'Expand'} ${row.name}`"
                    :aria-expanded="expanded.has(row.rank)"
                    @click="toggle(row.rank)"
                  />
                  <span :class="{ 'rank-label-quota': draft.mode === 'guaranteed' }"
                    >{{ row.name }} ({{ row.count }} {{ row.count === 1 ? "department" : "departments" }})</span
                  >
                  <runai-btn
                    v-if="
                      expanded.has(row.rank) && effectiveRows.filter((item) => item.rank === row.rank).length < row.count
                    "
                    :aid="`load-rank-${row.rank}`"
                    flat
                    label="Load more"
                    :disable="loading"
                    @click="emit('load-rank', row.rank)"
                  /></div
              ></q-td>
            </q-tr>
            <q-tr v-else :key="row.key" :class="{ 'selected-department': selected.has(row.id) }">
              <q-td class="column-selection"
                ><q-checkbox
                  v-if="ranked"
                  :model-value="selected.has(row.id)"
                  :disable="readonly || canEditDepartment?.(row.id) === false"
                  :aria-label="`Select ${row.name}`"
                  @update:model-value="select(row.id, $event)"
              /></q-td>
              <department-resource-editor
                v-if="draft.mode === 'guaranteed'"
                :id="row.id"
                kind="quota"
                :value="row.quota"
                :limit="draft.maxAllocation ? row.limit : undefined"
                :disabled="readonly || canEditDepartment?.(row.id) === false"
                :load-context="() => quotaContext(row)"
                @apply="emit('edit', { id: row.id, quota: $event })"
              />
              <q-td class="department-name">{{ row.name }}</q-td>
              <department-resource-editor
                :id="row.id"
                kind="weight"
                :title="ranked ? 'Weight within rank' : 'Weight'"
                :value="row.weight"
                :display-value="percentage(row)"
                :preview-weight-percentage="(weight) => previewWeightPercentage(row, weight)"
                :disabled="readonly || automatic || canEditDepartment?.(row.id) === false"
                :disabled-tooltip="
                  automatic
                    ? 'The department’s weight is proportional to its assigned GPU quota. To set a custom value, select Custom in Advanced settings.'
                    : ''
                "
                @apply="emit('edit', { id: row.id, weight: $event })"
              />
              <department-resource-editor
                v-if="draft.maxAllocation"
                :id="row.id"
                kind="limit"
                :value="row.limit ?? -1"
                :quota="draft.mode === 'guaranteed' ? row.quota : undefined"
                :disabled="readonly || canEditDepartment?.(row.id) === false"
                @apply="emit('edit', { id: row.id, limit: $event })"
              />
              <q-td class="result-cell">
                <div class="result-bar" :class="{ 'simulation-bar': simulation }">
                  <template v-if="simulation && simulationResults?.[row.id]">
                    <span
                      class="allocated"
                      :style="{
                        width: `${Math.min(100, (simulationResults[row.id].allocated / (capacity || 1)) * 100)}%`,
                        background: allocationBackground(row),
                      }"
                      >{{ Number(simulationResults[row.id].allocated.toFixed(2)) }}</span
                    >
                    <span
                      class="unmet"
                      :style="{ width: `${Math.min(100, (simulationResults[row.id].unmet / (capacity || 1)) * 100)}%` }"
                      >{{ Number(simulationResults[row.id].unmet.toFixed(2)) }}</span
                    >
                  </template>
                  <span
                    v-else-if="percentage(row) !== 'Unknown'"
                    class="quota-marker"
                    :style="{
                      left: `${Math.min(100, Math.max(0, parseFloat(percentage(row))))}%`,
                    }"
                  />
                  <q-tooltip class="tooltip-dark scheduling-result-tooltip" max-width="380px">
                    <div class="result-tooltip-title">{{ row.name }}</div>
                    <div class="result-tooltip-legend">
                      <div v-for="entry in resultLegend(row)" :key="entry.label" class="result-tooltip-entry">
                        <span v-if="entry.image" class="result-tooltip-line" aria-hidden="true">
                          <img :src="entry.image" alt="" width="4" height="38" />
                        </span>
                        <span v-else class="result-tooltip-swatch" :class="`swatch-${entry.kind}`" aria-hidden="true" />
                        <span
                          >{{ entry.label }} <strong>{{ entry.value }}</strong></span
                        >
                      </div>
                    </div>
                    <div v-if="simulation && !simulationResults?.[row.id]" class="result-tooltip-unavailable">
                      {{ simulationMessage || "Simulation results are not available yet." }}
                    </div>
                  </q-tooltip>
                  <input
                    v-if="simulation"
                    @input="demand[row.id] = Number(($event.target as HTMLInputElement).value)"
                    type="range"
                    step="any"
                    min="0"
                    :max="capacity ?? 0"
                    :value="
                      demand[row.id] ?? simulationResults?.[row.id]?.demand ?? defaultSimulationDemand(capacity ?? 0)
                    "
                    :disabled="!canView || capacity == null"
                    :aria-label="`Simulated GPU demand for ${row.name}`"
                    class="demand-slider"
                  />
                </div>
              </q-td>
            </q-tr>
          </template>
          <template #no-data
            ><div class="scheduling-empty" role="status">
              {{
                !canView
                  ? "You’re missing permissions to view departments"
                  : loading
                    ? "Loading departments…"
                    : hasError
                      ? "Department data is unavailable."
                      : searchTerm || rankFilter || filters.advancedFilters?.length
                        ? "No departments match your filters"
                        : "No departments yet"
              }}
            </div></template
          >
        </q-table>
        <div v-if="simulation" class="simulation-notice" role="status" aria-live="polite">
          <span v-if="preview">Simulation preview — mock results. Authored examples, not scheduler predictions.</span>
          <span v-if="simulationStatus === 'loading'">Calculating simulation…</span>
          <span v-else-if="simulationMessage">{{ simulationMessage }}</span>
          <runai-btn
            v-if="simulationStatus === 'error' || simulationStatus === 'unavailable'"
            flat
            label="Retry"
            aid="retry-simulation"
            @click="retrySimulation"
          />
          <span v-if="draft.timeFairshare">Time-based fairshare is not included in this simulation.</span>
        </div>
        <div v-if="totalsComplete === false && !simulationResults" class="data-notice" role="status">
          Weight percentages are unavailable because you cannot view every department in this node pool.
        </div>
        <div v-if="!automatic && !rawWeightsAvailable" class="data-notice" role="status">
          Raw weights are unavailable while the saved mode uses automatic weighting. Percentages require a complete set
          of custom weights.
        </div>
        <div v-if="hasMore" class="q-pa-md text-center">
          <runai-btn
            aid="load-more-departments"
            flat
            label="Load more departments"
            :loading="loading"
            @click="emit('load-more')"
          />
        </div>
      </div>
    </div>
    <move-to-rank-modal
      v-if="showMove"
      :selected-count="selected.size"
      descending
      compact
      show-unsupported-no-over-quota
      :current-rank="currentRank"
      :entity-type="EQuotaEntity.department"
      @close="showMove = false"
      @move="move"
    />
  </section>
</template>
<script setup lang="ts">
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";
import weightLegend from "@/assets/icons/scheduling/weight-legend.png";
import maximumLegend from "@/assets/icons/scheduling/maximum-legend.png";
import { computed, inject, ref, watch } from "vue";
import type { QTableColumn } from "quasar";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { EColumnFilterType, EFilterOperator, type IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { matchesSchedulingFilters } from "./node-pool-scheduling-filters";
import { useNodePoolSimulation } from "./use-node-pool-simulation.composable";
import { simulateNodePool, simulationPreviewKey } from "./node-pool-simulation.service";
import { defaultSimulationDemand } from "./node-pool-simulation.util";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import MoveToRankModal from "@/components/org-unit/v2/common/peer-org-unit-rank-list/move-to-rank-modal/move-to-rank-modal.vue";
import {
  ENTITY_PRIORITY_DEFAULT_VALUE,
  ENTITY_PRIORITY_SELECT_OPTIONS,
  entityPriorityMap,
} from "@/models/org-unit.model";
import { EQuotaEntity } from "@/models/resource.model";
import type { DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client";
import DepartmentResourceEditor from "./department-resource-editor.vue";
import {
  schedulingTotalsWithEdits,
  weightPercentage,
  type SchedulingDraft,
  type DepartmentSchedulingEdit,
  type SchedulingTotals,
  type DepartmentValidationContext,
} from "./node-pool-scheduling.model";
const props = withDefaults(
  defineProps<{
    poolId?: string;
    rows: DepartmentNodePoolResources[];
    baselineRows?: DepartmentNodePoolResources[];
    draft: SchedulingDraft;
    aggregates: Record<string, SchedulingTotals>;
    ranks: string[];
    rankCounts?: Record<string, number>;
    capacity?: number;
    readonly: boolean;
    canView?: boolean;
    settingsOpen?: boolean;
    totalsComplete?: boolean;
    rawWeightsAvailable?: boolean;
    canEditDepartment?: (id: string) => boolean;
    loading: boolean;
    hasMore?: boolean;
    hasError?: boolean;
    searchTerm?: string;
    rankFilter?: string;
    loadQuotaContext?: (id: string) => Promise<DepartmentValidationContext>;
  }>(),
  { totalsComplete: true, rawWeightsAvailable: true, canView: true },
);
const emit = defineEmits<{
  edit: [DepartmentSchedulingEdit];
  search: [string];
  "filter-rank": [string | undefined];
  "load-more": [];
  "load-rank": [string];
  refresh: [];
  "new-department": [];
  "advanced-settings": [];
  "close-settings": [];
  "automatic-weights": [boolean];
}>();
const baseline = ref<Record<string, DepartmentSchedulingEdit>>({});
const knownRows = ref<Record<string, DepartmentNodePoolResources>>({});
function baselineRecord(row: DepartmentNodePoolResources): DepartmentSchedulingEdit {
  return {
    id: row.orgUnitId,
    weight: props.rawWeightsAvailable ? (row.resources.gpu?.overQuotaWeight ?? undefined) : undefined,
    quota: row.resources.gpu?.deserved ?? undefined,
    limit: row.resources.gpu?.limit,
    rank: row.rank,
  };
}
watch(
  () => props.baselineRows,
  (rows) => {
    if (!rows) return;
    baseline.value = Object.fromEntries(rows.map((row) => [row.orgUnitId, baselineRecord(row)]));
    for (const row of rows) knownRows.value[row.orgUnitId] = row;
  },
  { immediate: true },
);
watch(
  () => props.rows,
  (rows) => {
    for (const row of rows) {
      knownRows.value[row.orgUnitId] = row;
      if (!props.baselineRows) baseline.value[row.orgUnitId] = baselineRecord(row);
    }
  },
  { immediate: true },
);
const selected = ref(new Set<string>());
const expanded = ref(new Set<string>([ENTITY_PRIORITY_DEFAULT_VALUE]));
const showMove = ref(false);
const ranked = computed(() => props.draft.mode === "ranked" || props.draft.mode === "guaranteed");
const automatic = computed(() => props.draft.mode === "guaranteed" && props.draft.automaticWeights);
const completeCustomWeightTotals = computed(() => {
  const sums = new Map<string, number | undefined>();
  if (automatic.value || props.totalsComplete === false || !props.baselineRows) return sums;
  // A saved Automatic pool exposes no raw weights. Drafts can establish a total only when every peer is known.
  for (const source of props.baselineRows) {
    const row = { ...baselineRecord(source), ...props.draft.departments[source.orgUnitId] };
    for (const key of ["all", row.rank ?? ""]) {
      if (!sums.has(key)) sums.set(key, 0);
      const current = sums.get(key);
      const weight = row.weight;
      if (weight == null || !Number.isFinite(weight) || !Number.isInteger(weight) || weight < 0)
        sums.set(key, undefined);
      else if (current !== undefined) sums.set(key, current + weight);
    }
  }
  return sums;
});
const totals = computed(() =>
  schedulingTotalsWithEdits(props.aggregates, baseline.value, props.draft.departments, automatic.value),
);
const quotaExcess = computed(() => {
  const capacity = props.capacity;
  const assigned = totals.value.all;
  if (
    props.draft.mode !== "guaranteed" ||
    !props.canView ||
    props.totalsComplete === false ||
    capacity == null ||
    !Number.isFinite(capacity) ||
    capacity < 0 ||
    !assigned ||
    !Number.isFinite(assigned.deservedSum) ||
    assigned.deservedSum <= capacity
  )
    return undefined;
  const finiteTotal = assigned.finiteQuotaSum ?? assigned.deservedSum;
  return Number.isFinite(finiteTotal) && finiteTotal > capacity ? finiteTotal - capacity : undefined;
});

watch(
  [
    ranked,
    () => props.rankCounts?.[ENTITY_PRIORITY_DEFAULT_VALUE],
    () => props.rows.length,
    () => props.rankFilter,
    () => props.searchTerm,
  ],
  () => {
    if (
      ranked.value &&
      !props.rankFilter &&
      (props.rankCounts?.[ENTITY_PRIORITY_DEFAULT_VALUE] ?? 0) > 0 &&
      !props.rows.some((row) => row.rank === ENTITY_PRIORITY_DEFAULT_VALUE)
    )
      emit("load-rank", ENTITY_PRIORITY_DEFAULT_VALUE);
  },
  { immediate: true },
);
const filters = ref<IFilterBy>({ advancedFilters: [], displayedColumns: [] });
const unfilteredHeight = ref(0);
const preserveTableSpace = computed(
  () => props.loading || !!props.searchTerm || !!props.rankFilter || !!filters.value.advancedFilters?.length,
);
function rememberTableHeight(size: { height: number }) {
  if (!preserveTableSpace.value) unfilteredHeight.value = size.height;
}
const filterColumns = computed<ITableColumn[]>(() => [
  {
    name: "department",
    label: "Department",
    field: "orgUnitName",
    filter: {
      type: EColumnFilterType.FreeText,
      maxWidth: "282px",
      operatorMenuWidth: "208px",
      operatorLabels: {
        [EFilterOperator.Contains]: "Contains",
        [EFilterOperator.DoesNotContain]: "Does not contain",
        [EFilterOperator.Equals]: "Equals",
        [EFilterOperator.NotEquals]: "Does not equal",
      },
      excludeOperators: [EFilterOperator.StartsWith, EFilterOperator.EndsWith],
    },
  },
  ...(ranked.value
    ? [
        {
          name: "rank",
          label: "Rank",
          field: "rank",
          filter: {
            type: EColumnFilterType.EnumBased,
            maxWidth: "208px",
            selectOptions: ENTITY_PRIORITY_SELECT_OPTIONS,
            inputType: "checkbox" as const,
          },
        },
      ]
    : []),
]);
function updateFilters(value: IFilterBy) {
  filters.value = value;
  if (value.searchTerm !== props.searchTerm) emit("search", value.searchTerm ?? "");
}
const columns = computed<QTableColumn[]>(() => [
  { name: "selection", label: "", field: "id", align: "left" },
  ...(props.draft.mode === "guaranteed"
    ? [{ name: "quota", label: "Guaranteed GPUs", field: "quota", align: "left" as const }]
    : []),
  { name: "department", label: ranked.value ? "Rank" : "Department", field: "name", align: "left" },
  { name: "weight", label: ranked.value ? "Weight within rank" : "Weight", field: "weight", align: "left" },
  ...(props.draft.maxAllocation
    ? [{ name: "limit", label: "Max GPU allocation", field: "limit", align: "left" as const }]
    : []),
  { name: "result", label: "Result", field: "id", align: "left" },
]);
const simulation = ref(false);
const demand = ref<Record<string, number>>({});
const currentRank = computed(() => {
  const ranks = new Set([...selected.value].map((id) => props.draft.departments[id]?.rank ?? baseline.value[id]?.rank));
  return ranks.size === 1 ? [...ranks][0] : undefined;
});
const allSelected = computed(
  () => effectiveRows.value.length > 0 && effectiveRows.value.every((row) => selected.value.has(row.id)),
);
function selectAll(value: boolean) {
  for (const row of effectiveRows.value) if (props.canEditDepartment?.(row.id) !== false) select(row.id, value);
}
const headerTooltips = computed<Record<string, string>>(() => ({
  selection: "The numbers indicate the order in which the Scheduler considers resource allocation in real time.",
  quota:
    "These GPUs are guaranteed to the department and allocated before any sharing whenever its workloads request them. Cluster conditions such as fragmentation can occasionally prevent the full quota from being allocated. Across all departments, quota cannot exceed the node pool capacity.",
  department:
    "The highest rank receives over-quota capacity first. Unused capacity flows down to the next rank, where it is distributed proportionally by weight within the rank. Select departments to move them to another rank.",
  weight: ranked.value
    ? "The weight determines each department’s proportional share of the over-quota resources within its rank. Unused capacity flows down to the next rank, where it is distributed by weight."
    : "The weight determines each department’s proportional share of the resources.",
  limit:
    "This is the upper limit of GPUs a department can consume from a node pool, combining assigned quota and over-quota.",
  result:
    "View a visual summary of each department’s resource settings. Turn on Simulate workload demand to see how GPUs would be allocated based on real-time demand.",
}));
function rankHeading(rank: string) {
  const label = entityPriorityMap[rank] ?? rank;
  return label.charAt(0) + label.slice(1).toLowerCase();
}
function order(name: string) {
  if (name === "quota") return "#1";
  if (name === "department") return ranked.value ? (props.draft.mode === "guaranteed" ? "#2" : "#1") : "";
  if (name === "weight") return props.draft.mode === "guaranteed" ? "#3" : ranked.value ? "#2" : "#1";
  if (name === "limit") return props.draft.mode === "guaranteed" ? "#4" : ranked.value ? "#3" : "#2";
  return "";
}
const effectiveRows = computed(() => {
  const source = filters.value.advancedFilters?.length ? (props.baselineRows ?? props.rows) : props.rows;
  const visible = new Map(source.map((row) => [row.orgUnitId, row]));
  // Keep moved rows visible even when their original rank was the active server filter.
  for (const [id, edit] of Object.entries(props.draft.departments))
    if (
      edit.rank &&
      knownRows.value[id] &&
      (!props.searchTerm || knownRows.value[id].orgUnitName.toLowerCase().includes(props.searchTerm.toLowerCase()))
    )
      visible.set(id, knownRows.value[id]);
  return [...visible.values()]
    .map((row) => ({
      key: row.orgUnitId,
      ...baseline.value[row.orgUnitId],
      ...props.draft.departments[row.orgUnitId],
      id: row.orgUnitId,
      name: row.orgUnitName,
    }))
    .filter(
      (row) =>
        (!ranked.value || !props.rankFilter || row.rank === props.rankFilter) &&
        (!props.searchTerm || row.name.toLowerCase().includes(props.searchTerm.toLowerCase())) &&
        matchesSchedulingFilters(row, filters.value.advancedFilters ?? []),
    );
});
const noMatchingDepartments = computed(
  () =>
    !!(
      props.canView &&
      !props.loading &&
      !props.hasError &&
      (props.searchTerm || props.rankFilter || filters.value.advancedFilters?.length) &&
      !effectiveRows.value.length
    ),
);
const displayRows = computed(() => {
  if (!ranked.value) return effectiveRows.value;
  const ranks = [
    ...new Set([...props.ranks, ...effectiveRows.value.map((row) => row.rank).filter((rank): rank is string => !!rank)]),
  ];
  return ranks.flatMap((rank) => {
    const rows = effectiveRows.value.filter((row) => row.rank === rank);
    let count =
      props.searchTerm || filters.value.advancedFilters?.length
        ? rows.length
        : (props.rankCounts?.[rank] ?? rows.length);
    if (!props.searchTerm && !filters.value.advancedFilters?.length && props.rankCounts?.[rank] !== undefined)
      for (const [id, edit] of Object.entries(props.draft.departments)) {
        const original = baseline.value[id];
        if (edit.rank && original?.rank && edit.rank !== original.rank)
          count += (edit.rank === rank ? 1 : 0) - (original.rank === rank ? 1 : 0);
      }
    if (!count && !rows.length) return [];
    if (props.rankFilter && props.rankFilter !== rank) return [];
    return [
      {
        key: `rank-${rank}`,
        group: true,
        rank,
        name: rankHeading(rank),
        count: Math.max(count, rows.length),
      },
      ...(expanded.value.has(rank) ? rows : []),
    ];
  });
});
const preview = inject(simulationPreviewKey, undefined);
const scenario = computed(() =>
  simulation.value && props.canView && props.poolId && props.capacity != null
    ? {
        poolId: props.poolId,
        capacity: props.capacity,
        draft: props.draft,
        ids: displayRows.value.flatMap((row) => ("id" in row ? [row.id] : [])),
        demand: demand.value,
      }
    : undefined,
);
const {
  results: simulationResults,
  status: simulationStatus,
  message: simulationMessage,
  retry: retrySimulation,
} = useNodePoolSimulation(scenario, preview?.transport ?? simulateNodePool);
function allocationBackground(row: DepartmentSchedulingEdit) {
  const result = simulationResults.value?.[row.id];
  if (props.draft.mode !== "guaranteed" || !result?.allocated) return "#42a5f5";
  const guaranteed = (result.guaranteed / result.allocated) * 100;
  return `linear-gradient(to right, #42a5f5 ${guaranteed}%, #90caf9 ${guaranteed}%)`;
}
function resultLegend(row: DepartmentSchedulingEdit) {
  const entries: { label: string; value: string | number; kind: string; image?: string }[] = [];
  if (props.draft.mode === "guaranteed")
    entries.push({
      label: "Guaranteed GPU allocation",
      value: row.quota === -1 ? "Unlimited" : (row.quota ?? "Unknown"),
      kind: "guaranteed",
    });
  entries.push({
    label: ranked.value ? "Weight within rank" : "Weight",
    value: percentage(row),
    kind: "weight",
    image: weightLegend,
  });
  if (props.draft.maxAllocation)
    entries.push({
      label: "Max GPU allocation",
      value: row.limit == null || row.limit === -1 ? "Unlimited" : row.limit,
      kind: "maximum",
      image: maximumLegend,
    });
  const result = simulation.value ? simulationResults.value?.[row.id] : undefined;
  if (result) {
    const guaranteed = props.draft.mode === "guaranteed";
    if (guaranteed)
      entries.push({ label: "Expected guaranteed GPU allocation", value: result.guaranteed, kind: "guaranteed" });
    entries.push({
      label: guaranteed ? "Expected over-quota GPU allocation" : "Expected GPU allocation",
      value: guaranteed ? result.overQuota : result.allocated,
      kind: guaranteed ? "over-quota" : "guaranteed",
    });
    entries.push({ label: "Unmet GPU demand", value: result.unmet, kind: "unmet" });
    entries.push({
      label: "Simulated real-time GPU demand",
      value: result.demand,
      kind: "demand",
    });
  }
  return entries.map((entry) => ({
    ...entry,
    value: typeof entry.value === "number" ? Number(entry.value.toFixed(2)) : entry.value,
  }));
}
function weightTotal(row: DepartmentSchedulingEdit) {
  const key = ranked.value ? (row.rank ?? "") : "all";
  return totals.value[key]?.weightsSum ?? completeCustomWeightTotals.value.get(key);
}
function previewWeightPercentage(row: DepartmentSchedulingEdit, weight: number) {
  if (automatic.value || props.totalsComplete === false) return undefined;
  const total = weightTotal(row);
  if (total != null && Number.isFinite(total) && row.weight != null && row.weight >= 0 && total >= row.weight)
    return weightPercentage(weight, total - row.weight + weight);
  if (!props.baselineRows) return undefined;
  // The proposed value can complete a Custom scenario whose applied row weight is still unknown.
  let otherWeights = 0;
  for (const source of props.baselineRows) {
    const peer = { ...baselineRecord(source), ...props.draft.departments[source.orgUnitId] };
    if (peer.id === row.id || (ranked.value && peer.rank !== row.rank)) continue;
    if (peer.weight == null || !Number.isFinite(peer.weight) || !Number.isInteger(peer.weight) || peer.weight < 0)
      return undefined;
    otherWeights += peer.weight;
  }
  return weightPercentage(weight, otherWeights + weight);
}
function percentage(row: DepartmentSchedulingEdit) {
  const simulated = simulation.value ? simulationResults.value?.[row.id] : undefined;
  if (simulated) return `${Number(simulated.weightPercentage.toFixed(2))}%`;
  if (props.totalsComplete === false) return "Unknown";
  const value = automatic.value ? Math.max(0, row.quota ?? 0) : row.weight;
  const percent = value == null ? undefined : weightPercentage(value, weightTotal(row));
  return percent == null ? "Unknown" : `${Number(percent.toFixed(2))}%`;
}
function toggle(rank: string) {
  const next = new Set(expanded.value);
  if (next.has(rank)) next.delete(rank);
  else {
    next.add(rank);
    emit("load-rank", rank);
  }
  expanded.value = next;
}
function select(id: string, value: boolean) {
  const next = new Set(selected.value);
  if (value) next.add(id);
  else next.delete(id);
  selected.value = next;
}
function move(rank: string) {
  for (const id of selected.value) emit("edit", { id, rank });
  expanded.value = new Set([...expanded.value, rank]);
  showMove.value = false;
  selected.value = new Set();
}
async function quotaContext(row: DepartmentSchedulingEdit) {
  const result = await props.loadQuotaContext?.(row.id);
  return {
    ...result,
    capacity: props.capacity,
    totalsComplete: props.totalsComplete !== false,
    finiteQuotaSum: props.totalsComplete !== false ? totals.value.all?.finiteQuotaSum : undefined,
    unlimitedQuotaCount: props.totalsComplete !== false ? totals.value.all?.unlimitedQuotaCount : undefined,
    totalQuota:
      props.totalsComplete !== false && totals.value.all?.deservedSum >= 0 ? totals.value.all.deservedSum : undefined,
  };
}
</script>
<style scoped lang="scss">
.table-settings-layout {
  display: flex;
  align-items: stretch;
}
.department-table-content {
  min-width: 0;
  flex: 1 1 0;
  overflow: hidden;
  padding-bottom: 40px;
}
.department-section {
  min-width: 0;
}
.scheduling-actions {
  padding-right: 16px;
  border-top: 1px solid var(--border-subtle);
  min-height: 60px;
  height: 60px;
}
.scheduling-actions :deep([aid="new-department"]) {
  height: 38px;
  min-height: 38px;
  width: 174px;
  padding: 0 10px;
  font-weight: 500;
  line-height: 20px;
  letter-spacing: 0;
}
.scheduling-actions :deep([aid="new-department"] .q-icon) {
  width: 20px;
  height: 20px;
  font-size: 20px;
  margin-right: 10px;
}
.scheduling-actions :deep(.runai-general-search .icon-container) {
  height: 40px;
  padding: 0 10px;
  gap: 3px;
}
.scheduling-actions :deep(.runai-general-search .icon-container i) {
  font-size: 20px;
  line-height: 20px;
}
.scheduling-actions :deep(.runai-general-search .icon-container span) {
  font-size: 10px;
  line-height: 14px;
}
.scheduling-actions :deep(.runai-general-search:not(.focus):not(.filled)) {
  width: 59px;
}
.scheduling-actions :deep(.left-actions) {
  margin-right: 26px;
  flex-shrink: 0;
}
.scheduling-actions :deep(.add-filter-btn) {
  margin-left: 0;
  height: 38px;
  min-height: 38px;
  padding: 0 10px;
}
.scheduling-actions :deep(.main-actions) {
  height: 40px;
}
.scheduling-actions :deep(.right-actions-container > .q-btn) {
  height: 40px;
  min-height: 40px;
  padding: 0 10px;
}
.scheduling-actions :deep(.right-actions-container > .q-btn .q-btn__content) {
  gap: 3px;
}
.scheduling-actions :deep(.right-actions-container > .q-btn .q-icon) {
  font-size: 20px;
  width: 20px;
  height: 20px;
}
.scheduling-actions :deep(.right-actions-container > .q-btn .q-btn__content > span.block) {
  font-size: 10px;
  line-height: 14px;
}
.scheduling-actions :deep([aid="refresh-data-button"]) {
  margin-inline: 16px;
}
.scheduling-actions :deep([aid="advanced-settings"]) {
  width: 128px;
}
.scheduling-actions :deep(.right-actions-container) {
  flex-shrink: 0;
}
.actions-readonly :deep(.advanced-filters) {
  pointer-events: none;
  opacity: 0.5;
}
.scheduling-actions :deep(.q-btn__content) {
  white-space: nowrap;
}
.scheduling-actions :deep(.right-actions-container .q-btn__label) {
  font-size: 10px;
}
.scheduling-table {
  min-width: 0;
  width: 100%;
  max-width: 100%;
}
.scheduling-table :deep(table) {
  min-width: var(--scheduling-table-min-width);
  table-layout: fixed;
  width: 100%;
}
.scheduling-table :deep(th) {
  height: 60px;
  padding: 10px;
  vertical-align: bottom;
  line-height: 16.8px;
  text-align: left;
  font-size: 12px;
  background: white;
  white-space: normal;
  font-weight: 500;
}
.scheduling-table :deep(td) {
  height: 60px;
  font-size: 14px;
  line-height: 19.5px;
  padding: 4px 10px;
  border-bottom: 1px solid var(--border-subtle);
}
.scheduling-table :deep(.q-table__middle) {
  overflow-x: auto;
  min-width: 0;
  max-width: 100%;
}
.scheduling-table :deep(th.column-selection) {
  vertical-align: middle;
  text-align: center;
  border-right: 1px solid var(--border-subtle);
}
.scheduling-table :deep(.column-selection) {
  width: 56px;
  padding: 0 10px;
  color: var(--icon-default);
}
.scheduling-table :deep(.column-department) {
  width: 220px;
}
.scheduling-table :deep(.column-quota),
.scheduling-table :deep(.column-weight),
.scheduling-table :deep(.column-limit) {
  width: 150px;
}
.scheduling-table :deep(.column-result) {
  min-width: 220px;
}
.column-order {
  font-weight: 500;
  margin-bottom: 3px;
}
.order-quota {
  color: #2196f3;
}
.order-department {
  color: #673ab7;
}
.order-limit {
  color: #ff6d00;
}
.column-label {
  min-height: 21px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.column-label .q-icon {
  color: var(--icon-default);
}
.simulation-control {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 400;
  margin-top: 3px;
  white-space: nowrap;
}
.rank-header td {
  height: 60px !important;
  font-size: 12px;
  line-height: 16.8px;
}
.rank-row {
  display: flex;
  align-items: center;
  gap: 28px;
}
.rank-row > .q-btn {
  flex-shrink: 0;
  color: var(--icon-default);
}
.rank-row .rank-label-quota {
  margin-left: 150px;
}
.scheduling-table :deep(.q-checkbox__inner) {
  font-size: 40px;
}
.scheduling-table :deep(.cell-trigger) {
  height: 50px;
  width: 100px;
  justify-content: flex-start;
  padding: 0 14px;
  font-size: 14px;
}
.scheduling-table :deep(.cell--disabled .cell-trigger) {
  color: var(--text-disabled);
}
.selected-department {
  background: #eee;
}
.selected-department :deep(.cell-trigger) {
  border-color: var(--border-pressed);
}
.result-bar {
  display: flex;
  align-items: center;
  position: relative;
  width: 100%;
  height: 18px;
  border: 1px solid #bdbdbd;
  border-radius: 3px;
  overflow: visible;
  font-size: 12px;
  line-height: 16px;
  font-weight: 500;
}
.result-bar:not(.simulation-bar) {
  height: 18px;
}
.quota-marker {
  position: absolute;
  height: 100%;
  border-left: 1px dashed #212121;
}
.allocated {
  background: #90caf9;
  height: 100%;
  text-align: center;
  overflow: hidden;
  border-right: 1px dashed #212121;
}
.unmet {
  background: #ffe082;
  flex: 0 0 auto;
  text-align: center;
  height: 100%;
}
.demand-slider {
  appearance: none;
  position: absolute;
  top: 0;
  left: -5px;
  margin: 0;
  width: calc(100% + 10px);
  height: 100%;
  background: transparent;
  cursor: pointer;
}
.demand-slider::-webkit-slider-thumb {
  appearance: none;
  width: 20px;
  height: 20px;
  border: 1px solid white;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.87);
  box-shadow: 0 2px 2px #0004;
}
.demand-slider::-moz-range-thumb {
  width: 18px;
  height: 18px;
  border: 1px solid white;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.87);
}
.scheduling-empty {
  width: 100%;
  min-height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--b-text-main);
}
.simulation-notice,
.data-notice {
  color: var(--b-text-sub);
  padding: 8px 16px;
  font-size: 12px;
}
.quota-warning {
  border-radius: 0;
}
.scheduling-table :deep(.result-cell) {
  padding-right: 20px;
}
.allocated,
.unmet {
  text-shadow:
    -1px 0 white,
    1px 0 white,
    0 -1px white,
    0 1px white;
}
:global(.q-tooltip.scheduling-result-tooltip) {
  background: rgba(0, 0, 0, 0.87);
  color: white;
  padding: 10px;
  border-radius: 3px;
  font-size: 12px;
  line-height: 16.8px;
  box-shadow: -1px 0 7px #00000026;
}
.result-tooltip-title {
  margin-bottom: 7px;
  font-weight: 400;
}
.result-tooltip-legend {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.result-tooltip-entry {
  display: flex;
  align-items: center;
  gap: 7px;
  white-space: nowrap;
}
.result-tooltip-entry strong {
  font-weight: 800;
  margin-left: 5px;
}
.result-tooltip-line {
  width: 20px;
  height: 20px;
  overflow: hidden;
  position: relative;
  flex-shrink: 0;
}
.result-tooltip-line img {
  position: absolute;
  left: 4px;
  top: -9px;
}
.result-tooltip-swatch {
  width: 12px;
  height: 12px;
  border-radius: 3px;
  flex-shrink: 0;
}
.swatch-guaranteed {
  background: #42a5f5;
}
.swatch-over-quota {
  background: #90caf9;
}
.swatch-unmet {
  background: #ffe082;
}
.swatch-demand {
  background: rgba(0, 0, 0, 0.87);
  border: 2px solid white;
  border-radius: 50%;
}
.result-tooltip-unavailable {
  margin-top: 7px;
}
</style>
