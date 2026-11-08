<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <section class="department-section" aria-label="Department scheduling">
    <div class="department-toolbar row items-center q-gutter-sm">
      <span class="text-b-subtitle1">Departments</span>
      <q-badge v-if="readonly" color="grey">View only</q-badge>
      <runai-select
        v-if="ranked"
        class="rank-filter"
        :model-value="rankFilter ?? null"
        :options="rankOptions"
        label="Rank"
        emit-value
        dense
        @update:model-value="handleRankFilter"
      />
      <q-space />
      <runai-general-search
        :search-term="searchTerm ?? ''"
        tooltip-text="Search departments by name"
        @update-search-term="search"
      />
      <runai-btn
        aid="refresh-departments"
        flat
        icon="ph ph-arrow-clockwise"
        label="Refresh"
        :disable="loading"
        @click="emit('refresh')"
      />
    </div>
    <div v-if="draft.mode === 'guaranteed'" class="row items-center q-gutter-md q-pa-md">
      <span>Weight within rank</span>
      <runai-select
        :model-value="draft.automaticWeights ? 'automatic' : 'custom'"
        :options="weightOptions"
        emit-value
        dense
        :disable="readonly"
        @update:model-value="emit('automatic-weights', $event === 'automatic')"
      />
    </div>
    <runai-btn
      aid="move-departments"
      v-if="ranked && selected.size"
      flat
      :label="`Move ${selected.size} departments to rank`"
      :disable="readonly"
      @click="showMove = true"
    />
    <p
      v-if="
        draft.mode === 'guaranteed' && totalsComplete !== false && capacity != null && totals.all?.deservedSum > capacity
      "
      class="q-pa-md text-warning"
      role="status"
    >
      Assigned quota exceeds node pool capacity.
    </p>
    <p v-if="totalsComplete === false" class="q-px-md text-b-subtitle2" role="status">
      Weight percentages are unavailable because you cannot view every department in this node pool.
    </p>
    <p v-if="!automatic && !rawWeightsAvailable" class="q-px-md text-b-subtitle2" role="status">
      Raw weights are unavailable while the saved mode uses automatic weighting. Percentages remain unknown. Apply a
      value to add a draft weight.
    </p>
    <q-table
      flat
      square
      class="scheduling-table"
      :rows="displayRows"
      :columns="columns"
      row-key="key"
      :loading="loading"
      :pagination="{ rowsPerPage: 0 }"
      hide-pagination
      :rows-per-page-options="[0]"
      :table-style="{ minWidth: draft.mode === 'guaranteed' ? '680px' : '500px' }"
    >
      <template #body="{ row }">
        <q-tr v-if="row.group" class="rank-header">
          <q-td :colspan="columns.length">
            <runai-btn
              :aid="`rank-${row.rank}`"
              flat
              :icon="expanded.has(row.rank) ? 'ph ph-caret-up' : 'ph ph-caret-down'"
              :label="`${row.name} (${row.count} departments)`"
              :aria-expanded="expanded.has(row.rank)"
              @click="toggle(row.rank)"
            />
            <runai-btn
              v-if="expanded.has(row.rank) && rows.filter((item) => item.rank === row.rank).length < row.count"
              :aid="`load-rank-${row.rank}`"
              flat
              label="Load more"
              :disable="loading"
              @click="emit('load-rank', row.rank)"
            />
          </q-td>
        </q-tr>
        <q-tr v-else :key="row.key">
          <q-td
            ><q-checkbox
              v-if="ranked"
              :model-value="selected.has(row.id)"
              :disable="readonly || canEditDepartment?.(row.id) === false"
              :aria-label="`Select ${row.name}`"
              @update:model-value="select(row.id, $event)"
          /></q-td>
          <q-td class="department-name">{{ row.name }}</q-td>
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
          <department-resource-editor
            :id="row.id"
            kind="weight"
            :value="row.weight"
            :display-value="percentage(row)"
            :disabled="readonly || automatic || canEditDepartment?.(row.id) === false"
            :disabled-tooltip="
              automatic ? 'Weights are automatically calculated from guaranteed GPU quotas. Select Custom to edit.' : ''
            "
            @apply="emit('edit', { id: row.id, weight: $event })"
          />
        </q-tr>
      </template>
      <template #no-data
        ><div class="q-pa-lg" role="status">
          {{
            loading
              ? "Loading departments…"
              : hasError
                ? "Department data is unavailable."
                : searchTerm || rankFilter
                  ? "No departments match your filters."
                  : "No departments in this node pool."
          }}
        </div></template
      >
    </q-table>
    <div v-if="hasMore" class="q-pa-md text-center">
      <runai-btn
        aid="load-more-departments"
        flat
        label="Load more departments"
        :loading="loading"
        @click="emit('load-more')"
      />
    </div>
    <move-to-rank-modal
      v-if="showMove"
      :selected-count="selected.size"
      :entity-type="EQuotaEntity.department"
      @close="showMove = false"
      @move="move"
    />
  </section>
</template>
<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { QTableColumn } from "quasar";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiGeneralSearch } from "@/components/common/runai-general-search";
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
    draft: SchedulingDraft;
    aggregates: Record<string, SchedulingTotals>;
    ranks: string[];
    rankCounts?: Record<string, number>;
    capacity?: number;
    readonly: boolean;
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
  { totalsComplete: true, rawWeightsAvailable: true },
);
const emit = defineEmits<{
  edit: [DepartmentSchedulingEdit];
  search: [string];
  "filter-rank": [string | undefined];
  "load-more": [];
  "load-rank": [string];
  refresh: [];
  newDepartment: [];
  advancedSettings: [];
  "automatic-weights": [boolean];
}>();
const baseline = ref<Record<string, DepartmentSchedulingEdit>>({});
const knownRows = ref<Record<string, DepartmentNodePoolResources>>({});
watch(
  () => props.rows,
  (rows) => {
    for (const row of rows) {
      knownRows.value[row.orgUnitId] = row;
      // Page replacement must not replace the original values underlying pending deltas.
      baseline.value[row.orgUnitId] = {
        id: row.orgUnitId,
        weight: props.rawWeightsAvailable ? (row.resources.gpu?.overQuotaWeight ?? undefined) : undefined,
        quota: row.resources.gpu?.deserved ?? undefined,
        limit: row.resources.gpu?.limit,
        rank: row.rank,
      };
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
  [ranked, () => props.rankCounts?.[ENTITY_PRIORITY_DEFAULT_VALUE], () => props.rows.length],
  () => {
    if (
      ranked.value &&
      (props.rankCounts?.[ENTITY_PRIORITY_DEFAULT_VALUE] ?? 0) > 0 &&
      !props.rows.some((row) => row.rank === ENTITY_PRIORITY_DEFAULT_VALUE)
    )
      emit("load-rank", ENTITY_PRIORITY_DEFAULT_VALUE);
  },
  { immediate: true },
);
const rankOptions = [{ label: "All ranks", value: "" }, ...ENTITY_PRIORITY_SELECT_OPTIONS];
const weightOptions = [
  { label: "Automatic", value: "automatic" },
  { label: "Custom", value: "custom" },
];
const columns = computed<QTableColumn[]>(() => [
  { name: "selection", label: "", field: "id", align: "left" },
  { name: "department", label: ranked.value ? "Rank / Department" : "Department", field: "name", align: "left" },
  ...(props.draft.mode === "guaranteed"
    ? [{ name: "quota", label: "Guaranteed GPUs", field: "quota", align: "left" as const }]
    : []),
  { name: "weight", label: ranked.value ? "Weight within rank" : "Weight", field: "weight", align: "left" },
]);
const effectiveRows = computed(() => {
  const visible = new Map(props.rows.map((row) => [row.orgUnitId, row]));
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
    .filter((row) => !ranked.value || !props.rankFilter || row.rank === props.rankFilter);
});
const displayRows = computed(() => {
  if (!ranked.value) return effectiveRows.value;
  const ranks = [
    ...new Set([...props.ranks, ...effectiveRows.value.map((row) => row.rank).filter((rank): rank is string => !!rank)]),
  ];
  return ranks.flatMap((rank) => {
    const rows = effectiveRows.value.filter((row) => row.rank === rank);
    let count = props.searchTerm ? rows.length : (props.rankCounts?.[rank] ?? rows.length);
    if (!props.searchTerm && props.rankCounts?.[rank] !== undefined)
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
        name: entityPriorityMap[rank] ?? rank,
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
function search(value: string) {
  emit("search", value);
}
function handleRankFilter(value: unknown) {
  emit("filter-rank", typeof value === "string" && value ? value : undefined);
}
async function quotaContext(row: DepartmentSchedulingEdit) {
  const result = await props.loadQuotaContext?.(row.id);
  return {
    ...result,
    capacity: props.capacity,
    totalQuota:
      props.totalsComplete !== false && totals.value.all?.deservedSum >= 0 ? totals.value.all.deservedSum : undefined,
  };
}
</script>
<style scoped lang="scss">
.department-section {
  margin-top: 32px;
  border-top: 1px solid var(--border-subtle);
}
.department-toolbar {
  padding: 10px 16px;
  min-height: 60px;
}
.rank-filter {
  min-width: 150px;
}
.scheduling-table :deep(th) {
  position: sticky;
  top: 0;
  background: var(--bg-primary-default, white);
  z-index: 1;
  font-weight: 500;
}
.scheduling-table :deep(td) {
  height: 60px;
  border-bottom: 1px solid var(--border-subtle);
}
.scheduling-table :deep(.q-table__middle) {
  max-height: 540px;
}
.department-name {
  min-width: 220px;
}
.rank-header {
  background: var(--bg-secondary-default, #fafafa);
}
</style>
