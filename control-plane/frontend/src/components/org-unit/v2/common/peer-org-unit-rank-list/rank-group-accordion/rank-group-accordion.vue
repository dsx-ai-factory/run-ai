<template>
  <q-table
    flat
    :rows="rows"
    :columns="columns"
    row-key="rank"
    hide-pagination
    hide-bottom
    :rows-per-page-options="[0]"
    :pagination="{ rowsPerPage: 0 }"
    class="rank-group-accordion"
  >
    <template v-slot:header="props">
      <q-tr :props="props" class="text-b-body1-md">
        <q-th class="col-width-trigger" />

        <q-th v-for="col in props.cols" :key="col.name" :props="props" :class="col.headerClasses">
          {{ col.label }}
        </q-th>
      </q-tr>
    </template>

    <template v-slot:body="props">
      <q-tr
        :props="props"
        class="cursor-pointer"
        @click="handleRowClick(props.row.rank)"
        @keydown.enter.prevent="handleRowClick(props.row.rank)"
        @keydown.space.prevent="handleRowClick(props.row.rank)"
        tabindex="0"
        role="row"
        :aria-expanded="isRowExpanded(props.row.rank)"
        :aria-label="`${getRankLabel(props.row.rank)} - ${props.row.count} ${entityType}${props.row.count !== 1 ? 's' : ''}`"
      >
        <q-td auto-width>
          <q-icon
            :name="isRowExpanded(props.row.rank) ? 'ph-bold ph-caret-up' : 'ph-bold ph-caret-down'"
            color="black-54"
            size="16px"
          />
        </q-td>
        <!--  Rank and count -->
        <q-td key="rank" :props="props">
          <div class="q-py-md row items-center no-wrap text-b-body1 rank-count-header">
            <span>
              {{ getRankLabel(props.row.rank) }} (<span class="rank-count">{{ props.row.count }}</span> {{ entityType
              }}{{ props.row.count !== 1 ? "s" : "" }})
            </span>
            <q-icon
              v-if="shouldShowStar(props.row.rank)"
              name="ph-fill ph-star"
              color="black-54"
              size="16px"
              class="q-ml-sm"
            />
          </div>
        </q-td>

        <!--  Weight -->
        <q-td key="weight" :props="props">
          <slot name="weight" :row="props.row" />
        </q-td>
        <!--  Distribution -->
        <q-td key="distribution" :props="props">
          <slot name="distribution" :row="props.row" />
        </q-td>
      </q-tr>

      <!--  Expanded content row-->
      <q-tr v-show="isRowExpanded(props.row.rank)" :props="props">
        <q-td colspan="100%" class="no-padding">
          <slot name="expanded-content" :row="props.row"> </slot>
        </q-td>
      </q-tr>
    </template>
  </q-table>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { EEntityPriorityLabel } from "@/models/org-unit.model";
import { entityPriorityMap } from "@/models/org-unit.model";
import { EQuotaEntity } from "@/models/resource.model";

export interface RankGroupRow {
  rank: string;
  count: number;
}

export interface Props {
  rows: RankGroupRow[];
  entityType?: EQuotaEntity;
  expandedRank?: string | null; // Externally controlled expanded rank
  currentOrgUnitRank?: string | null; // Rank containing the current org unit (for star indicator)
}

const props = withDefaults(defineProps<Props>(), {
  entityType: EQuotaEntity.project,
  rows: () => [],
  expandedRank: null,
  currentOrgUnitRank: null,
});

// Show star on collapsed rank row that contains the current org unit
const shouldShowStar = (rank: string) => {
  return props.currentOrgUnitRank === rank && props.expandedRank !== rank;
};

const emit = defineEmits<{
  (e: "rank-toggled", payload: { rank: string; isExpanded: boolean }): void;
}>();

const entityLabel = computed(() => (props.entityType === "department" ? "Department" : "Project"));

const getRankLabel = (rankValue: string): EEntityPriorityLabel => {
  return entityPriorityMap[rankValue] ?? rankValue;
};

const handleRowClick = (rank: string) => {
  // If clicking on the already expanded row, collapse it
  if (props.expandedRank === rank) {
    emit("rank-toggled", { rank, isExpanded: false });
  } else {
    // Otherwise, expand the clicked row (and collapse any other)
    emit("rank-toggled", { rank, isExpanded: true });
  }
};

const isRowExpanded = (rank: string) => {
  return props.expandedRank === rank;
};

const columns = computed(() => [
  {
    name: "rank",
    label: `Rank / ${entityLabel.value}`,
    field: "rank",
    align: "left" as const,
    sortable: false,
    // Remove inline style, use class in template instead
    classes: "col-width-rank",
    headerClasses: "col-width-rank",
  },
  {
    name: "weight",
    label: "Weight",
    field: "weight",
    align: "left" as const,
    sortable: false,
    classes: "col-width-weight",
    headerClasses: "col-width-weight",
  },
  {
    name: "distribution",
    label: "Weight distribution",
    field: "distribution",
    align: "left" as const,
    sortable: false,
    classes: "col-width-distribution",
    headerClasses: "col-width-distribution",
  },
]);
</script>

<style scoped lang="scss">
thead .col-width-trigger {
  width: 50px; // Fixed width for Chevron (Parent) and Checkbox (Child)
  min-width: 50px;
}

thead .col-width-rank {
  width: 40%; // Or a fixed pixel value like 300px if preferred
  padding-left: 71px;
}

thead .col-width-weight {
  width: 140px;
}

thead .col-width-distribution {
  // Let this take remaining space or define a width
  width: auto;

  padding-left: 20px;
}

.rank-group-accordion {
  // Smooth chevron rotation
  :deep(.q-icon) {
    transition: transform 0.2s ease;
  }

  // Count badge subtle pulse animation when value changes
  .rank-count {
    transition: transform 0.15s ease;
  }

  // Expanded content smooth reveal
  :deep(.no-padding) {
    animation: expand-in 0.25s ease-out;
  }

  .rank-count-header {
    padding-left: 58px;
  }
}

@keyframes expand-in {
  from {
    opacity: 0;
    transform: translateY(-4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
