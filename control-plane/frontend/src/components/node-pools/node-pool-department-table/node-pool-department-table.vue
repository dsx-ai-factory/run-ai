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
      :columns="[]"
      :filters="{ ...filters, searchTerm }"
      advanced-filters
      search-tooltip-text="Search departments"
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
      <template #filters>
        <scheduling-filters :filter-by="{ ...filters, searchTerm }" :ranked="ranked" @update-filter-by="updateFilters" />
      </template>
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
          ref="settingsTrigger"
          aid="advanced-settings"
          flat
          stack
          size="sm"
          icon="ph-bold ph-gear-six"
          :label="props.searchTerm || noMatchingDepartments ? 'Show settings' : 'Advanced settings'"
          :disable="readonly"
          :aria-expanded="!!settingsOpen"
          :aria-controls="settingsOpen ? settingsPanelId : undefined"
          @click="emit('advanced-settings')"
      /></template>
    </runai-page-actions>
    <div class="table-settings-layout" :class="{ 'table-settings-layout--open': panelVisible }">
      <div
        v-if="panelVisible"
        ref="settingsPanel"
        :id="settingsPanelId"
        class="department-settings-panel"
        @keydown.esc.stop="emit('close-settings')"
      >
        <slot name="settings" />
      </div>
      <div
        ref="tableContent"
        class="department-table-content"
        :style="preserveTableSpace ? { minHeight: `${unfilteredHeight}px` } : undefined"
        @focusin="revealFocusedControl"
      >
        <q-resize-observer @resize="rememberTableHeight" />
        <q-table
          flat
          square
          class="scheduling-table"
          :rows="canView && !resolvingFilters ? displayRows : []"
          :columns="columns"
          row-key="key"
          :loading="loading || resolvingFilters"
          :pagination="{ rowsPerPage: 0 }"
          hide-pagination
          :rows-per-page-options="[0]"
          :style="{ '--scheduling-table-min-width': maxAllocation ? '998px' : '960px' }"
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
                  <template v-if="col.name === 'result'">
                    <div class="simulation-control">
                      Simulate workload demand
                      <q-toggle
                        v-model="simulation"
                        dense
                        size="35px"
                        :disable="!canView || capacity == null || !hasVisibleDepartments"
                        :aria-describedby="capacity == null ? capacityHintId : undefined"
                        aid="simulate-demand"
                        aria-label="Simulate workload demand"
                      />
                    </div>
                    <div v-if="capacity == null" :id="capacityHintId" class="simulation-capacity-hint">
                      Simulation needs the node pool’s GPU capacity.
                    </div>
                  </template>
                </template>
              </q-th>
            </q-tr></template
          >
          <template #body="{ row }">
            <q-tr v-if="row.pageEnd" :key="row.key" class="department-page-end" aria-hidden="true">
              <q-td :colspan="columns.length"
                ><div
                  v-intersection.once="
                    (entry: IntersectionObserverEntry) => {
                      if (entry.isIntersecting) showNextPage(row.rank);
                      return true;
                    }
                  "
              /></q-td>
            </q-tr>
            <q-tr v-else-if="row.group" :key="row.key" class="rank-header">
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
                  <span :class="{ 'rank-label-quota': draft.mode === ESchedulingMode.Guaranteed }"
                    >{{ row.name }} ({{ row.count }} {{ row.count === 1 ? "department" : "departments" }})</span
                  >
                </div></q-td
              >
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
                v-if="draft.mode === ESchedulingMode.Guaranteed"
                :id="row.id"
                :kind="EResourceEditorKind.Quota"
                :value="row.quota"
                :limit="maxAllocation ? row.limit : undefined"
                :disabled="readonly || canEditDepartment?.(row.id) === false"
                :load-context="() => quotaContext(row)"
                @apply="emit('edit', { id: row.id, quota: $event })"
              />
              <q-td class="department-name"><runai-ellipsis-text-with-tooltip :text="row.name" /></q-td>
              <department-resource-editor
                :id="row.id"
                :kind="EResourceEditorKind.Weight"
                :title="ranked ? 'Weight within rank' : 'Weight'"
                :value="row.weight"
                :display-value="percentage(row)"
                :load-context="
                  !automatic && totalsComplete && hasMore && weightTotal(row) == null ? loadWeightPeers : undefined
                "
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
                v-if="maxAllocation"
                :id="row.id"
                :kind="EResourceEditorKind.Limit"
                :value="row.limit ?? -1"
                :quota="draft.mode === ESchedulingMode.Guaranteed ? row.quota : undefined"
                :disabled="readonly || canEditDepartment?.(row.id) === false"
                @apply="emit('edit', { id: row.id, limit: $event })"
              />
              <q-td class="result-cell">
                <div
                  class="result-bar"
                  :class="{ 'simulation-bar': simulation }"
                  :aria-busy="simulation && simulationStatus === 'loading'"
                >
                  <simulation-result-bar
                    v-if="simulation && simulationResults?.[row.id]"
                    :allocated="simulationResults[row.id].allocated"
                    :unmet="simulationResults[row.id].unmet"
                    :capacity="capacity ?? 0"
                    :demand="simulationResults[row.id].demand"
                    :guaranteed="
                      draft.mode === ESchedulingMode.Guaranteed
                        ? simulationResults[row.id].guaranteed
                        : simulationResults[row.id].allocated
                    "
                  />
                  <q-skeleton
                    v-else-if="simulation && simulationStatus === 'loading'"
                    type="rect"
                    width="100%"
                    height="100%"
                    aria-hidden="true"
                  />
                  <div v-else-if="percentage(row) !== 'Unknown'" class="simulation-segments">
                    <span
                      class="quota-marker"
                      aria-hidden="true"
                      :style="{
                        left: `${Math.min(100, Math.max(0, rowWeightPercentage(row) ?? 0))}%`,
                      }"
                    />
                  </div>
                  <q-tooltip class="tooltip-dark scheduling-result-tooltip" max-width="380px">
                    <div class="result-tooltip-title">{{ row.name }}</div>
                    <div class="result-tooltip-legend">
                      <div v-for="entry in resultLegend(row)" :key="entry.label" class="result-tooltip-entry">
                        <span
                          v-if="entry.kind === 'weight' || entry.kind === 'maximum'"
                          class="result-tooltip-line"
                          :class="`marker-${entry.kind}`"
                          aria-hidden="true"
                        />
                        <span v-else class="result-tooltip-swatch" :class="`swatch-${entry.kind}`" aria-hidden="true" />
                        <span class="result-tooltip-description"
                          ><span>{{ entry.label }}</span
                          >{{ " " }}<strong>{{ entry.value }}</strong></span
                        >
                      </div>
                    </div>
                  </q-tooltip>
                  <input
                    v-if="simulation"
                    @input="previewDemand(row, $event)"
                    @change="setDemand(row, $event)"
                    type="range"
                    step="1"
                    min="0"
                    :max="capacity ?? 0"
                    :value="sliderDemand(row)"
                    :aria-valuemax="maximumDemand(row)"
                    :aria-valuetext="demandText(row)"
                    :disabled="!canView || capacity == null"
                    :aria-label="`Simulated GPU demand for ${row.name}`"
                    class="demand-slider"
                  />
                  <q-chip
                    v-if="simulation"
                    aria-hidden="true"
                    class="demand-chip"
                    :style="{ '--demand-fraction': capacity ? sliderDemand(row) / capacity : 0 }"
                  >
                    <strong>{{ sliderDemand(row) }}</strong
                    >{{ sliderDemand(row) === 1 ? " GPU" : " GPUs" }}
                  </q-chip>
                </div>
              </q-td>
            </q-tr>
          </template>
          <template #no-data
            ><div v-if="!canView || (!loading && !resolvingFilters && !hasError)" class="scheduling-empty" role="status">
              {{
                !canView
                  ? "You’re missing permissions to view departments"
                  : searchTerm || rankFilter || filters.advancedFilters?.length
                    ? "No departments match your filters"
                    : "No departments yet"
              }}
            </div></template
          >
        </q-table>
        <div v-if="simulation && simulationMessage" class="simulation-notice" role="alert">
          <span>{{ simulationMessage }}</span>
          <runai-btn v-if="canRetrySimulation" flat label="Retry" aid="retry-simulation" @click="retrySimulation" />
        </div>
      </div>
    </div>
    <move-to-rank-modal
      v-if="showMove"
      :selected-count="selected.size"
      :current-rank="currentRank"
      :entity-type="EQuotaEntity.department"
      @close="showMove = false"
      @move="move"
    />
  </section>
</template>
<script setup lang="ts">
import { EResourceEditorKind } from "@/models/node-pool-scheduling.model";
import { ENodePoolSchedulingMode as ESchedulingMode, isSupportedSchedulingMode } from "@/models/node-pool.model";

import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar/index";
import { SimulationResultBar } from "@/components/node-pools/simulation-result-bar";
import { SchedulingFilters } from "@/components/node-pools/scheduling-filters";
import { computed, nextTick, ref, useId, watch } from "vue";
import { QBtn, Intersection as vIntersection } from "quasar";
import { RunaiPageActions } from "@/components/common/runai-page-actions/index";
import { RunaiEllipsisTextWithTooltip } from "@/components/common/runai-ellipsis-text-with-tooltip/index";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button/index";
import type { IFilterBy } from "@/models/filter.model";
import { matchesSchedulingFilters } from "@/components/node-pools/scheduling-filters/scheduling-filters.util";
import { useNodePoolSimulation } from "@/composables/node-pool/use-node-pool-simulation.composable";
import { defaultSimulationDemand } from "@/utils/node-pool-scheduling.util/node-pool-simulation.util";
import { integerSimulationValues } from "@/utils/node-pool-scheduling.util/node-pool-simulation-display.util";
import { formatUtil } from "@/utils/format.util/format.util";
import { completePeerSchedulingTotals } from "@/utils/node-pool-scheduling.util/node-pool-scheduling-totals.util";
import {
  departmentTableColumns,
  schedulingColumnOrder,
  rankedDepartmentRows,
  previewDepartmentWeightPercentage,
} from "@/components/node-pools/node-pool-department-table/node-pool-department-table.util";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { SchedulingMoveToRankDialog as MoveToRankModal } from "@/components/node-pools/scheduling-move-to-rank-dialog";
import { ENTITY_PRIORITY_DEFAULT_VALUE } from "@/models/org-unit.model";
import { EQuotaEntity } from "@/models/resource.model";
import type { DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client/index";
import { DepartmentResourceEditor } from "@/components/node-pools/department-resource-editor";
import {
  DEPARTMENT_PAGE_SIZE,
  schedulingTotalsWithEdits,
  effectiveDepartmentEdit,
  automaticWeightFromQuota,
  weightFromGuaranteedQuota,
  weightPercentage,
  type SchedulingDraft,
  type IDepartmentPages,
  type IDepartmentSchedulingEdit,
  type ISchedulingTotals,
  type IDepartmentValidationContext,
} from "@/models/node-pool-scheduling.model";
const props = withDefaults(
  defineProps<{
    poolId?: string;
    rows: DepartmentNodePoolResources[];
    departmentPages?: IDepartmentPages;
    baselineRows?: DepartmentNodePoolResources[];
    draft: SchedulingDraft;
    aggregates: Record<string, ISchedulingTotals>;
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
    hasError?: boolean;
    searchTerm?: string;
    rankFilter?: string;
    loadQuotaContext?: (id: string) => Promise<IDepartmentValidationContext>;
  }>(),
  { totalsComplete: true, rawWeightsAvailable: true, canView: true },
);
const emit = defineEmits<{
  edit: [IDepartmentSchedulingEdit];
  search: [string];
  "filter-rank": [string | undefined];
  refresh: [];
  "new-department": [];
  "advanced-settings": [];
  "close-settings": [];
  "automatic-weights": [boolean];
}>();
const initializingCustomWeights = computed(
  () =>
    props.draft.mode === ESchedulingMode.Guaranteed &&
    !props.draft.automaticWeights &&
    props.draft.capabilities?.quota === true &&
    props.draft.capabilities.adjustableWeights === false &&
    !props.rawWeightsAvailable,
);
const knownRows = ref<Record<string, DepartmentNodePoolResources>>({});
watch(
  [() => props.baselineRows, () => props.rows],
  ([snapshot, rows]) => {
    for (const row of snapshot ?? []) knownRows.value[row.orgUnitId] = row;
    for (const row of rows) knownRows.value[row.orgUnitId] = row;
  },
  { immediate: true },
);
const baseline = computed(() => {
  const departments: Record<string, IDepartmentSchedulingEdit> = {};
  for (const row of props.baselineRows ?? Object.values(knownRows.value)) {
    const gpu = row.resources.gpu;
    departments[row.orgUnitId] = {
      id: row.orgUnitId,
      weight: initializingCustomWeights.value
        ? gpu?.deserved == null
          ? undefined
          : weightFromGuaranteedQuota(gpu.deserved)
        : props.rawWeightsAvailable
          ? (gpu?.overQuotaWeight ?? undefined)
          : undefined,
      quota: gpu?.deserved ?? undefined,
      limit: gpu?.limit,
      rank: row.rank,
    };
  }
  return departments;
});
const selected = ref(new Set<string>());
const expanded = ref(new Set<string>([ENTITY_PRIORITY_DEFAULT_VALUE]));
const showMove = ref(false);
const maxAllocation = computed(() => isSupportedSchedulingMode(props.draft.mode) && !!props.draft.maxAllocation);
const ranked = computed(
  () => props.draft.mode === ESchedulingMode.Ranked || props.draft.mode === ESchedulingMode.Guaranteed,
);
const automatic = computed(() => props.draft.mode === ESchedulingMode.Guaranteed && props.draft.automaticWeights);
const hasMore = computed(() => props.departmentPages?.nextOffset != null);
const completeEffectiveRows = computed(() =>
  hasMore.value
    ? undefined
    : props.baselineRows?.map((row) => effectiveDepartmentEdit(baseline.value, props.draft.departments, row.orgUnitId)),
);
const peerTotals = computed(() =>
  props.totalsComplete && completeEffectiveRows.value
    ? completePeerSchedulingTotals(completeEffectiveRows.value)
    : undefined,
);
const totals = computed(() =>
  schedulingTotalsWithEdits(props.aggregates, baseline.value, props.draft.departments, automatic.value),
);
const quotaExcess = computed(() => {
  const capacity = props.capacity;
  const assigned = totals.value.all;
  if (
    props.draft.mode !== ESchedulingMode.Guaranteed ||
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

const filters = ref<IFilterBy>({ advancedFilters: [], displayedColumns: [] });
const unfilteredHeight = ref(0);
const preserveTableSpace = computed(
  () => props.loading || !!props.searchTerm || !!props.rankFilter || !!filters.value.advancedFilters?.length,
);
function rememberTableHeight(size: { height: number }) {
  if (!preserveTableSpace.value) unfilteredHeight.value = size.height;
}
function updateFilters(value: IFilterBy) {
  filters.value = value;
  if (value.searchTerm !== props.searchTerm) emit("search", value.searchTerm ?? "");
}
const panelVisible = computed(() => props.canView && props.settingsOpen);
const columns = computed(() => departmentTableColumns(props.draft.mode, maxAllocation.value));
const simulation = ref(false);
const capacityHintId = useId();
const settingsPanelId = useId();
const settingsTrigger = ref<QBtn>();
const tableContent = ref<HTMLElement>();
const settingsPanel = ref<HTMLElement>();
function revealFocusedControl(event: FocusEvent) {
  if (!settingsPanel.value || !(event.target instanceof HTMLElement)) return;
  const panel = settingsPanel.value.getBoundingClientRect();
  const control = event.target.getBoundingClientRect();
  if (
    control.right > panel.left &&
    control.left < panel.right &&
    control.bottom > panel.top &&
    control.top < panel.bottom
  )
    emit("close-settings");
}
watch(
  () => props.settingsOpen,
  (open) => {
    if (!open && !tableContent.value?.contains(document.activeElement)) settingsTrigger.value?.$el.focus();
  },
  { flush: "post" },
);
const demand = ref<Record<string, number>>({});
const demandPreview = ref<Record<string, number>>({});
const currentRank = computed(() => {
  const ranks = new Set([...selected.value].map((id) => props.draft.departments[id]?.rank ?? baseline.value[id]?.rank));
  return ranks.size === 1 ? [...ranks][0] : undefined;
});
const allSelected = computed(
  () =>
    !hasMore.value && effectiveRows.value.length > 0 && effectiveRows.value.every((row) => selected.value.has(row.id)),
);
async function selectAll(value: boolean) {
  if (hasMore.value && !(await props.departmentPages?.loadAll())) return;
  await nextTick();
  if (props.readonly) return;
  const next = new Set(selected.value);
  for (const row of effectiveRows.value) {
    if (props.canEditDepartment?.(row.id) === false) continue;
    if (value) next.add(row.id);
    else next.delete(row.id);
  }
  selected.value = next;
}
const headerTooltips = computed<Record<string, string>>(() => ({
  selection: "The numbers indicate the order in which the Scheduler considers resource allocation in real time.",
  quota:
    "These GPUs are guaranteed to the department and allocated before any sharing whenever its workloads request them. Cluster conditions such as fragmentation can occasionally prevent the full quota from being allocated. Across all departments, quota cannot exceed the node pool capacity.",
  department:
    "The highest rank receives over-quota capacity first. Unused capacity flows down to the next rank, where it is distributed proportionally by weight within the rank. Select departments to move them to another rank.",
  weight: automatic.value
    ? "Automatic weight is derived from the department’s Guaranteed GPU quota. Zero-quota departments receive a small minimum weight. Custom weights can be set in Advanced settings."
    : ranked.value
      ? "The weight determines each department’s proportional share of the over-quota resources within its rank. Unused capacity flows down to the next rank, where it is distributed by weight."
      : "The weight determines each department’s proportional share of the resources.",
  limit:
    props.draft.mode === ESchedulingMode.Guaranteed
      ? "This is the upper limit of GPUs a department can consume from a node pool, combining assigned quota and over-quota."
      : "The maximum number of GPUs a department can use from this node pool.",
  result:
    "View a visual summary of each department’s resource settings. Turn on Simulate workload demand to see how GPUs would be allocated based on real-time demand.",
}));
const order = (name: string) => schedulingColumnOrder(name, props.draft.mode);
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
      ...effectiveDepartmentEdit(baseline.value, props.draft.departments, row.orgUnitId),
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
const filtering = computed(() => !!(props.searchTerm || props.rankFilter || filters.value.advancedFilters?.length));
const resolvingFilters = computed(() => filtering.value && hasMore.value && !props.hasError);
const noMatchingDepartments = computed(
  () =>
    !!(
      props.canView &&
      !props.loading &&
      !props.hasError &&
      !resolvingFilters.value &&
      (props.searchTerm || props.rankFilter || filters.value.advancedFilters?.length) &&
      !effectiveRows.value.length
    ),
);
const matchingRows = computed(() =>
  ranked.value
    ? rankedDepartmentRows(effectiveRows.value, {
        ranks: props.ranks,
        expanded: expanded.value,
        rankCounts: props.rankCounts,
        filtered: !!props.searchTerm || !!filters.value.advancedFilters?.length,
        rankFilter: props.rankFilter,
        baseline: baseline.value,
        edits: props.draft.departments,
      })
    : effectiveRows.value,
);
// Keep the DOM paged even when a search, bulk selection or simulation has needed every peer.
const shownByRank = ref<Record<string, number>>({});
watch([filtering, () => props.searchTerm, () => filters.value, ranked], () => {
  shownByRank.value = {};
});
const displayRows = computed(() => {
  if (!props.departmentPages) return matchingRows.value;
  type Row = (typeof matchingRows.value)[number];
  type PageEnd = { key: string; pageEnd: true; rank: string };
  const result: (Row | PageEnd)[] = [];
  let rank = "all",
    shown = 0,
    count = 0;
  function finishGroup() {
    if (props.hasError || props.loading || props.departmentPages?.loadingAll) return;
    if (ranked.value && !expanded.value.has(rank)) return;
    const limit = shownByRank.value[rank] ?? DEPARTMENT_PAGE_SIZE;
    if ((shown < count && (shown >= limit || hasMore.value)) || (!ranked.value && hasMore.value))
      result.push({
        key: `page-end-${rank}-${shown}-${props.departmentPages?.nextOffset ?? props.rows.length}`,
        pageEnd: true,
        rank,
      });
  }
  for (const row of matchingRows.value) {
    if ("group" in row) {
      finishGroup();
      rank = row.rank;
      count = row.count;
      shown = 0;
      result.push(row);
    } else if (shown < (shownByRank.value[rank] ?? DEPARTMENT_PAGE_SIZE)) {
      result.push(row);
      shown++;
    }
  }
  if (!ranked.value) count = matchingRows.value.length;
  finishGroup();
  return result;
});
function showNextPage(rank: string) {
  const limit = shownByRank.value[rank] ?? DEPARTMENT_PAGE_SIZE;
  const cached = effectiveRows.value.filter((row) => rank === "all" || row.rank === rank).length;
  if (cached >= limit) shownByRank.value = { ...shownByRank.value, [rank]: limit + DEPARTMENT_PAGE_SIZE };
  if (cached <= limit && hasMore.value) void props.departmentPages?.loadNext();
}
watch(
  () =>
    hasMore.value &&
    !props.loading &&
    !props.hasError &&
    (filtering.value ||
      simulation.value ||
      ((automatic.value || initializingCustomWeights.value) && props.totalsComplete) ||
      Object.keys(props.draft.departments).some((id) => !baseline.value[id])),
  (needsAll) => {
    if (needsAll) void props.departmentPages?.loadAll();
  },
  { immediate: true },
);
watch(
  () => [props.capacity, maxAllocation.value, baseline.value, props.draft.departments],
  () => {
    if (props.capacity == null) {
      simulation.value = false;
      return;
    }
    for (const values of [demand.value, demandPreview.value])
      for (const id of Object.keys(values)) {
        const row = effectiveDepartmentEdit(baseline.value, props.draft.departments, id);
        values[id] = clampDemand(row, values[id]);
      }
  },
  { deep: true, flush: "sync" },
);
const visibleDepartmentIds = computed(() => matchingRows.value.flatMap((row) => ("id" in row ? [row.id] : [])));
const hasVisibleDepartments = computed(() => visibleDepartmentIds.value.length > 0);
watch([hasVisibleDepartments, () => props.loading], ([visible, loading]) => {
  if (!visible && !loading) simulation.value = false;
});
const initialSimulationWeights = computed(() =>
  initializingCustomWeights.value
    ? Object.fromEntries(Object.values(baseline.value).map((row) => [row.id, row.weight]))
    : undefined,
);
const scenario = computed(() =>
  simulation.value &&
  !hasMore.value &&
  hasVisibleDepartments.value &&
  props.canView &&
  props.poolId &&
  props.capacity != null
    ? {
        poolId: props.poolId,
        capacity: props.capacity,
        draft: props.draft,
        initialWeights: initialSimulationWeights.value,
        ids: visibleDepartmentIds.value,
        demand: demand.value,
      }
    : undefined,
);
const {
  results: simulationResults,
  status: simulationStatus,
  message: simulationMessage,
  retry: retrySimulation,
  canRetry: canRetrySimulation,
} = useNodePoolSimulation(scenario);
type LegendMarker = "guaranteed" | "over-quota" | "weight" | "maximum" | "unmet" | "demand";
function resultLegend(row: IDepartmentSchedulingEdit) {
  const entries: { label: string; value: string | number; kind: LegendMarker }[] = [];
  if (props.draft.mode === ESchedulingMode.Guaranteed)
    entries.push({
      label: "Guaranteed GPU allocation",
      value: row.quota === -1 ? "Unlimited" : (row.quota ?? "Unknown"),
      kind: "guaranteed",
    });
  const weightShare = rowWeightPercentage(row);
  entries.push({
    label: ranked.value ? "Weight within rank" : "Weight",
    value:
      weightShare == null
        ? "Unknown"
        : `${formatUtil.roundToDecimal(weightShare, Math.max(0, 5 - Math.floor(Math.log10(Math.abs(weightShare) || 1))))}%`,
    kind: "weight",
  });
  if (maxAllocation.value)
    entries.push({
      label: "Max GPU allocation",
      value: row.limit == null || row.limit === -1 ? "Unlimited" : row.limit,
      kind: "maximum",
    });
  const result = simulation.value ? simulationResults.value?.[row.id] : undefined;
  if (result) {
    const display = integerSimulationValues(result);
    const guaranteed = props.draft.mode === ESchedulingMode.Guaranteed;
    if (guaranteed)
      entries.push({
        label: "Expected guaranteed GPU allocation",
        value: display.guaranteed,
        kind: "guaranteed",
      });
    entries.push({
      label: guaranteed ? "Expected over-quota GPU allocation" : "Expected GPU allocation",
      value: guaranteed ? display.overQuota : display.allocated,
      kind: guaranteed ? "over-quota" : "guaranteed",
    });
    entries.push({ label: "Unmet GPU demand", value: display.unmet, kind: "unmet" });
    entries.push({
      label: "Simulated real-time GPU demand",
      value: display.demand,
      kind: "demand",
    });
  }
  return entries.map((entry) => ({
    ...entry,
    value: typeof entry.value === "number" ? formatUtil.roundToDecimal(entry.value) : entry.value,
  }));
}
function weightTotal(row: IDepartmentSchedulingEdit) {
  const key = ranked.value ? (row.rank ?? "") : "all";
  if (automatic.value || initializingCustomWeights.value) {
    // The server's quota sum omits the zero-quota scheduler floor and unlimited defaults.
    if (hasMore.value) return undefined;
    if (peerTotals.value)
      return automatic.value ? peerTotals.value[key]?.automaticWeightsSum : peerTotals.value[key]?.rawWeightsSum;
  }
  return totals.value[key]?.weightsSum ?? peerTotals.value?.[key]?.rawWeightsSum;
}
async function loadWeightPeers() {
  if (!(await props.departmentPages?.loadAll())) throw new Error("Department weights are unavailable.");
  return {};
}
function previewWeightPercentage(row: IDepartmentSchedulingEdit, weight: number) {
  if (automatic.value || props.totalsComplete === false) return undefined;
  return previewDepartmentWeightPercentage(row, weight, weightTotal(row), completeEffectiveRows.value, ranked.value);
}
function rowWeightPercentage(row: IDepartmentSchedulingEdit): number | undefined {
  const simulated = simulation.value ? simulationResults.value?.[row.id] : undefined;
  if (simulated) return simulated.weightPercentage;
  if (props.totalsComplete === false) return undefined;
  const value = automatic.value ? automaticWeightFromQuota(row.quota) : row.weight;
  return value == null ? undefined : weightPercentage(value, weightTotal(row));
}
function percentage(row: IDepartmentSchedulingEdit) {
  const value = rowWeightPercentage(row);
  if (value == null || !Number.isFinite(value)) return "Unknown";
  if (value > 0 && value < 0.1) return "<0.1%";
  return `${formatUtil.roundToDecimal(value, automatic.value ? 1 : 2)}%`;
}
function maximumDemand(row: IDepartmentSchedulingEdit) {
  const capacity = Number.isFinite(props.capacity) ? Math.max(0, props.capacity ?? 0) : 0;
  return Math.floor(
    maxAllocation.value && row.limit != null && row.limit >= 0 ? Math.min(capacity, row.limit) : capacity,
  );
}
function clampDemand(row: IDepartmentSchedulingEdit, value: number) {
  return Math.min(maximumDemand(row), Math.max(0, Number.isFinite(value) ? Math.round(value) : 0));
}
function sliderDemand(row: IDepartmentSchedulingEdit) {
  return clampDemand(
    row,
    demandPreview.value[row.id] ??
      demand.value[row.id] ??
      simulationResults.value?.[row.id]?.demand ??
      defaultSimulationDemand(props.capacity ?? 0),
  );
}
function demandText(row: IDepartmentSchedulingEdit) {
  const value = sliderDemand(row);
  return `${value} ${value === 1 ? "GPU" : "GPUs"}`;
}
function previewDemand(row: IDepartmentSchedulingEdit, event: Event) {
  const input = event.target as HTMLInputElement;
  const value = clampDemand(row, Number(input.value));
  input.value = String(value);
  demandPreview.value[row.id] = value;
}
function setDemand(row: IDepartmentSchedulingEdit, event: Event) {
  const input = event.target as HTMLInputElement;
  const value = clampDemand(row, Number(input.value));
  input.value = String(value);
  demand.value[row.id] = value;
  delete demandPreview.value[row.id];
}
function toggle(rank: string) {
  const next = new Set(expanded.value);
  if (next.has(rank)) next.delete(rank);
  else next.add(rank);
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
async function quotaContext(row: IDepartmentSchedulingEdit) {
  if (props.totalsComplete && hasMore.value && props.aggregates.all?.deservedSum < 0) {
    if (!(await props.departmentPages?.loadAll())) throw new Error("Department quotas are unavailable.");
    await nextTick();
  }
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
  --selection-width: 56px;
  --department-width: 220px;
  --resource-width: 150px;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
}
.department-settings-panel {
  grid-area: 1 / 1;
  place-self: start end;
  width: min(380px, 100%);
  z-index: 1;
  background: var(--bg-primary-default);
}
@container department-section (max-width: 520px) {
  .table-settings-layout--open .department-table-content {
    grid-row: 2;
  }
}
.department-table-content {
  grid-area: 1 / 1;
  z-index: 0;
  align-self: start;
  min-width: 0;
  overflow: hidden;
  padding-bottom: 40px;
}
.department-section {
  container: department-section / inline-size;
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
  color: var(--text-sub);
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
  width: 78px;
  padding: 0 10px;
  line-height: 20px;
}
.scheduling-actions :deep(.main-actions) {
  height: 40px;
}
.scheduling-actions :deep(.right-actions-container > .q-btn) {
  height: 40px;
  min-height: 40px;
  padding: 0 10px;
  line-height: 14px;
  color: var(--text-sub);
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
  width: 63px;
}
.scheduling-actions :deep([aid="advanced-settings"]) {
  width: 121px;
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
  vertical-align: top;
  line-height: 16.8px;
  text-align: left;
  font-size: 12px;
  background: var(--table-bg);
  white-space: normal;
  font-weight: 500;
  color: var(--b-text-main);
  border-bottom: 0;
  box-shadow: inset 0 -1px var(--border-subtle);
}
.scheduling-table :deep(td) {
  height: 60px;
  font-size: 14px;
  line-height: 21px;
  padding: 5px 10px;
  border-bottom: 0;
  box-shadow: inset 0 -1px var(--border-subtle);
  color: var(--b-text-main);
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
  width: var(--selection-width);
  padding: 0 10px;
  color: var(--text-sub);
}
.scheduling-table :deep(.column-department) {
  width: var(--department-width);
}
.scheduling-table :deep(.column-quota),
.scheduling-table :deep(.column-weight),
.scheduling-table :deep(.column-limit) {
  width: var(--resource-width);
}
.scheduling-table :deep(.column-result) {
  min-width: 220px;
}
.column-order {
  font-weight: 700;
  line-height: 18px;
  margin-bottom: 3px;
  color: var(--b-text-main);
}
.order-quota {
  color: $blue-5;
}
.order-department {
  color: $purple-10;
}
.order-limit {
  color: $orange-10;
}
.column-label {
  min-height: 17px;
  display: flex;
  align-items: center;
  gap: 7px;
}
.scheduling-table :deep(th.column-department:not(:has(.column-order))) {
  vertical-align: bottom;
}
.scheduling-table :deep(th.column-department:not(:has(.column-order))) .column-label {
  min-height: 21px;
}
.column-label .q-icon {
  color: var(--text-sub);
}
.simulation-control {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 400;
  line-height: 18px;
  margin-top: 5px;
  white-space: nowrap;
}
.simulation-control :deep(.q-toggle__inner) {
  padding-top: 3px;
  padding-bottom: 0;
}
.simulation-capacity-hint {
  color: var(--b-text-sub);
  font-weight: 400;
  margin-top: 4px;
}
.simulation-control :deep(.q-toggle__inner--truthy .q-toggle__track) {
  background: var(--interactive-primary-default-bg);
  opacity: 0.5;
}
.simulation-control :deep(.q-toggle__inner--truthy .q-toggle__thumb) {
  left: 13px;
}
.simulation-control :deep(.q-toggle__inner--truthy .q-toggle__thumb::after) {
  box-shadow: var(--shadow-s);
}
.scheduling-table :deep(.department-page-end td) {
  height: 1px;
  padding: 0;
  box-shadow: none;
}
.department-page-end div {
  height: 1px;
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
  padding: 0 15px;
  font-size: 14px;
  border-radius: 3px;
}
.scheduling-table :deep(.cell--disabled .cell-trigger) {
  color: var(--text-disabled);
  border-color: var(--interactive-secondary-disabled-border);
}
.scheduling-table :deep(.unlimited-allocation .cell-trigger) {
  color: var(--text-disabled);
  border-color: var(--interactive-secondary-disabled-border);
}
.selected-department {
  background: var(--bg-primary-pressed);
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
  border: 1px solid var(--border-default);
  border-radius: 3px;
  overflow: visible;
  font-size: 12px;
  line-height: 16.8px;
  font-weight: 500;
}
.quota-marker {
  border-inline: 1px solid var(--text-inverse);
  background: repeating-linear-gradient(to bottom, var(--b-text-main) 0 3px, var(--text-inverse) 3px 6px);
  position: absolute;
  top: 50%;
  width: 4px;
  height: 38px;
  transform: translate(-3px, -50%);
}
.simulation-segments {
  position: relative;
  display: flex;
  width: 100%;
  height: 100%;
  overflow: hidden;
  border-radius: 2px;
}
.simulation-bar {
  container-type: inline-size;
}
.demand-chip {
  position: absolute;
  bottom: calc(100% + 4px);
  left: 0;
  visibility: hidden;
  pointer-events: none;
  margin: 0;
  padding: 0 10px;
  height: 21px;
  border-radius: 999px;
  background: var(--b-text-main);
  color: var(--text-inverse);
  font-size: 12px;
  font-weight: 400;
  // The native thumb center travels from 5px inside the bar to 5px from its end.
  // Center the chip on that point, keeping it within the bar at either end.
  transform: translateX(clamp(0px, calc(5px + var(--demand-fraction) * (100cqw - 10px) - 50%), calc(100cqw - 100%)));
}
.demand-chip strong {
  margin-right: 5px;
}
.demand-slider:enabled:is(:hover, :active, :focus-visible) + .demand-chip {
  visibility: visible;
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
  border: 1px solid var(--text-inverse);
  border-radius: 50%;
  background: var(--b-text-main);
  box-shadow: var(--shadow-s);
}
.demand-slider::-moz-range-thumb {
  width: 18px;
  height: 18px;
  border: 1px solid var(--text-inverse);
  border-radius: 50%;
  background: var(--b-text-main);
}
.scheduling-empty {
  width: 100%;
  min-height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--b-text-main);
  font-size: 14px;
  font-weight: 400;
  line-height: 1.4;
}
.scheduling-table :deep(.q-table__bottom:has(.scheduling-empty)) {
  padding: 0;
}
.simulation-notice {
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
:global(.q-tooltip.scheduling-result-tooltip) {
  background: var(--b-text-main);
  color: var(--text-inverse);
  padding: 10px;
  border-radius: 3px;
  font-size: 12px;
  line-height: 16.8px;
  box-shadow: var(--shadow-m);
}
.result-tooltip-title {
  margin-bottom: 7px;
  font-weight: 400;
  min-height: 17px;
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
  min-height: 17px;
}
.result-tooltip-entry strong {
  font-weight: 800;
}
.result-tooltip-description {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.result-tooltip-entry:first-child:has(> .swatch-guaranteed) {
  min-height: 20px;
}
.result-tooltip-entry:first-child > .swatch-guaranteed {
  margin-right: 8px;
}
.result-tooltip-line {
  width: 20px;
  height: 20px;
  overflow: hidden;
  position: relative;
  flex-shrink: 0;
}
.result-tooltip-line::before {
  content: "";
  position: absolute;
  left: 4px;
  top: -9px;
  width: 4px;
  height: 38px;
  border-inline: 1px solid var(--text-inverse);
  background: repeating-linear-gradient(to bottom, var(--b-text-main) 0 3px, var(--text-inverse) 3px 6px);
}
.result-tooltip-line.marker-maximum::before {
  background: repeating-linear-gradient(to bottom, #{$orange-7} 0 3px, var(--text-inverse) 3px 6px);
}
.result-tooltip-swatch {
  width: 12px;
  height: 12px;
  border-radius: 3px;
  flex-shrink: 0;
}
.swatch-guaranteed {
  background: $blue-5;
}
.swatch-over-quota {
  background: $blue-3;
}
.swatch-unmet {
  background: $amber-3;
}
.swatch-demand {
  background: var(--b-text-main);
  border: 2px solid var(--text-inverse);
  border-radius: 50%;
}
</style>
