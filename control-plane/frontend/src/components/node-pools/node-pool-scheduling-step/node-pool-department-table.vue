<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <section class="department-section" aria-label="Department scheduling">
    <q-banner
      v-if="
        draft.mode === 'guaranteed' && totalsComplete !== false && capacity != null && totals.all?.deservedSum > capacity
      "
      class="quota-warning"
      role="status"
      >The total assigned quota exceeds the node pool’s total capacity. Update departments to stay within the
      limit.</q-banner
    >
    <runai-page-actions
      class="scheduling-actions"
      :class="{ 'actions-readonly': readonly }"
      :columns="filterColumns"
      :filters="{ ...filters, searchTerm }"
      advanced-filters
      search-tooltip-text="Search departments"
      filter-menu-width="208px"
      hide-cluster-filter
      hide-columns-btn
      hide-export-csv
      :show-search="canView"
      :support-refresh="canView"
      :primary-btn-disable="readonly"
      :selected-rows-amount="selected.size"
      :disable-actions="readonly"
      @filters-changed="updateFilters"
      @refresh-data="emit('refresh')"
      @selected-actions-close="selected = new Set()"
    >
      <template #primary-btn
        ><runai-btn
          aid="new-department"
          color="primary"
          icon="ph ph-plus"
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
          icon="ph ph-gear-six"
          :label="noMatchingDepartments ? 'Show settings' : 'Advanced settings'"
          :disable="readonly"
          @click="emit('advanced-settings')"
      /></template>
    </runai-page-actions>
    <div class="table-settings-layout" :class="{ 'settings-open': settingsOpen }">
      <div class="department-table-content">
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
          :style="{ '--scheduling-table-min-width': draft.maxAllocation ? '1000px' : '960px' }"
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
                  <q-icon v-else name="ph-fill ph-question" size="18px"
                    ><q-tooltip>{{ headerTooltips.selection }}</q-tooltip></q-icon
                  >
                </template>
                <template v-else>
                  <div v-if="order(col.name)" :class="`column-order order-${col.name}`">{{ order(col.name) }}</div>
                  <div class="column-label">
                    {{ col.label }}
                    <q-icon v-if="col.name !== 'department' || ranked" name="ph ph-question" size="16px"
                      ><q-tooltip max-width="340px">{{ headerTooltips[col.name] }}</q-tooltip></q-icon
                    >
                  </div>
                  <div v-if="col.name === 'result'" class="simulation-control">
                    Simulate workload demand
                    <q-toggle
                      v-model="simulation"
                      dense
                      size="30px"
                      :disable="readonly"
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
                    :icon="expanded.has(row.rank) ? 'ph ph-caret-up' : 'ph ph-caret-down'"
                    :aid="`rank-${row.rank}`"
                    :aria-label="`${expanded.has(row.rank) ? 'Collapse' : 'Expand'} ${row.name}`"
                    :aria-expanded="expanded.has(row.rank)"
                    @click="toggle(row.rank)"
                  />
                  <span :class="{ 'rank-label-quota': draft.mode === 'guaranteed' }"
                    >{{ row.name }} ({{ row.count }} {{ row.count === 1 ? "department" : "departments" }})</span
                  >
                  <runai-btn
                    v-if="expanded.has(row.rank) && rows.filter((item) => item.rank === row.rank).length < row.count"
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
                :limit="row.limit"
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
                :quota="row.quota"
                :disabled="readonly || canEditDepartment?.(row.id) === false"
                @apply="emit('edit', { id: row.id, limit: $event })"
              />
              <q-td class="result-cell">
                <div class="result-bar" :class="{ 'simulation-bar': simulation }">
                  <template v-if="simulation && simulationResults?.[row.id]">
                    <span
                      class="allocated"
                      :style="{ width: `${(simulationResults[row.id].allocated / (capacity || 1)) * 100}%` }"
                      >{{ simulationResults[row.id].allocated }}</span
                    >
                    <span class="unmet">{{ simulationResults[row.id].unmet }}</span>
                  </template>
                  <span
                    v-else-if="percentage(row) !== 'Unknown'"
                    class="quota-marker"
                    :style="{
                      left: `${Math.min(100, Math.max(0, parseFloat(percentage(row))))}%`,
                    }"
                  />
                  <q-tooltip max-width="320px">
                    <strong>{{ row.name }}</strong
                    ><br />{{ ranked ? "Weight within rank" : "Weight" }}: {{ percentage(row) }}
                    <template v-if="draft.mode === 'guaranteed'"
                      ><br />Guaranteed GPU allocation:
                      {{ row.quota === -1 ? "Unlimited" : (row.quota ?? "Unknown") }}</template
                    >
                    <template v-if="draft.maxAllocation"
                      ><br />Max GPU allocation:
                      {{ row.limit == null || row.limit === -1 ? "Unlimited" : row.limit }}</template
                    >
                    <template v-if="simulation"
                      ><br />Simulation preview — mock results
                      <template v-if="simulationResults?.[row.id]">
                        <br />{{
                          draft.mode === "guaranteed" ? "Expected over-quota GPU allocation" : "Expected GPU allocation"
                        }}:
                        {{
                          draft.mode === "guaranteed"
                            ? simulationResults[row.id].overQuota
                            : simulationResults[row.id].allocated
                        }}
                        <template v-if="simulationResults[row.id].guaranteed"
                          ><br />Expected guaranteed GPU allocation: {{ simulationResults[row.id].guaranteed }}</template
                        >
                        <br />Unmet GPU demand: {{ simulationResults[row.id].unmet }} <br />Simulated real-time GPU
                        demand: {{ demand[row.id] ?? capacity }} <br />Authored example scenario; not a scheduler
                        prediction.
                      </template>
                      <template v-else><br />Mock preview not available for these values.</template>
                    </template>
                  </q-tooltip>
                  <input
                    v-if="simulation"
                    @input="demand[row.id] = Number(($event.target as HTMLInputElement).value)"
                    type="range"
                    min="0"
                    :max="capacity ?? 0"
                    :value="demand[row.id] ?? capacity"
                    :disabled="readonly || capacity == null"
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
                        ? "No departments match your filters."
                        : "No departments yet"
              }}
            </div></template
          >
        </q-table>
        <div v-if="simulation" class="simulation-notice" role="status">
          Simulation preview — mock results
          <q-icon name="ph ph-info"
            ><q-tooltip>{{
              simulationResults
                ? "Authored example scenario; not a scheduler prediction."
                : "Mock preview not available for these values. Demand controls are a preview only."
            }}</q-tooltip></q-icon
          >
        </div>
        <div v-if="totalsComplete === false" class="data-notice" role="status">
          Weight percentages are unavailable because you cannot view every department in this node pool.
        </div>
        <div v-if="!automatic && !rawWeightsAvailable" class="data-notice" role="status">
          Raw weights are unavailable while the saved mode uses automatic weighting. Percentages remain unknown. Apply a
          value to add a draft weight.
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
      <slot name="settings" />
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
import { computed, ref, watch } from "vue";
import type { QTableColumn } from "quasar";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { EColumnFilterType, EFilterOperator, type IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { matchesSchedulingFilters } from "./node-pool-scheduling-filters";
import {
  getMockSchedulingResult,
  MOCK_SCHEDULING_TEN_IDS,
  MOCK_GUARANTEED_OVERCOMMIT_QUOTAS,
} from "./node-pool-scheduling-simulation.mock";
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
const totals = computed(() =>
  schedulingTotalsWithEdits(props.aggregates, baseline.value, props.draft.departments, automatic.value),
);
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
const simulationResults = computed(() => {
  if (
    !props.draft.mode ||
    props.draft.automaticWeights ||
    props.draft.cpuQuota ||
    props.draft.timeFairshare ||
    (props.draft.historicalWeight != null && props.draft.historicalWeight !== 0) ||
    Object.values(props.draft.historicalWindow ?? {}).some((value) => value != null && value !== 0) ||
    Object.values(props.draft.minimumRuntime ?? {}).some((value) => value != null && value !== 0) ||
    Object.keys(props.draft.departments).length ||
    Object.values(baseline.value).some(
      (row) => row.weight !== 2 || row.rank !== "MediumLow" || (row.limit != null && row.limit !== -1),
    )
  )
    return undefined;
  const ids = Object.keys(baseline.value);
  const zeroQuota = Object.values(baseline.value).every((row) => row.quota === 0);
  const guaranteedOvercommit =
    props.draft.mode === "guaranteed" &&
    ids.length === 10 &&
    MOCK_SCHEDULING_TEN_IDS.every((id) => baseline.value[id]?.quota === MOCK_GUARANTEED_OVERCOMMIT_QUOTAS[id]);
  if (!zeroQuota && !guaranteedOvercommit) return undefined;
  return getMockSchedulingResult(
    ids,
    props.capacity,
    demand.value,
    guaranteedOvercommit ? "guaranteed-overcommit" : "zero-quota",
  );
});
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
function percentage(row: DepartmentSchedulingEdit) {
  if (props.totalsComplete === false) return "Unknown";
  const value = automatic.value ? Math.max(0, row.quota ?? 0) : row.weight;
  const percent =
    value == null
      ? undefined
      : weightPercentage(value, totals.value[ranked.value ? (row.rank ?? "") : "all"]?.weightsSum);
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
.settings-open .scheduling-table :deep(th.column-selection) {
  border-right: 1px solid var(--border-subtle);
}
.scheduling-table :deep(.column-department) {
  width: 150px;
}
.settings-open .rank-row .rank-label-quota {
  margin-left: 150px;
}
.scheduling-table :deep(table) {
  min-width: var(--scheduling-table-min-width);
  table-layout: fixed;
  width: 100%;
}
.scheduling-table :deep(th) {
  height: 60px;
  padding: 6px 10px;
  text-align: left;
  font-size: 12px;
  background: white;
  white-space: normal;
  font-weight: 400;
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
  margin-bottom: 4px;
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
  margin-top: 5px;
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
  flex: 1;
  text-align: center;
  height: 100%;
}
.demand-slider {
  appearance: none;
  position: absolute;
  inset: 0;
  margin: 0;
  width: 100%;
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
  background: #242219;
  box-shadow: 0 1px 3px #0006;
}
.demand-slider::-moz-range-thumb {
  width: 18px;
  height: 18px;
  border: 1px solid white;
  border-radius: 50%;
  background: #242219;
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
  background: #fff8e1;
  color: #805b00;
  font-size: 12px;
}
</style>
