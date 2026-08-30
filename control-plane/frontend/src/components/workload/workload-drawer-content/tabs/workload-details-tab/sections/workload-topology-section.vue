<template>
  <div class="workload-topology-section column">
    <runai-details-panel-row v-if="topology.name" label="Topology name">{{ topology.name }}</runai-details-panel-row>

    <runai-details-panel-row v-if="actualPlacementDisplay" label="Actual placement">
      <div class="row items-center b-gap-2xs">
        <span class="text-b-mono mono-badge">{{ actualPlacementDisplay.level }}</span>
        <span class="text-b-mono mono-badge">[{{ actualPlacementDisplay.value }}]</span>
      </div>
    </runai-details-panel-row>

    <runai-details-panel-row v-if="levels.length > 0" label="Topology levels">
      <div class="levels column">
        <div v-for="level in levels" :key="level" class="topology-level-row row items-center b-gap-2xs">
          <span class="text-b-mono mono-badge">{{ level }}</span>

          <span
            v-if="level === topology.requiredPlacement"
            class="constraint-badge constraint-badge--outline text-b-body1"
          >
            Required
          </span>

          <span
            v-if="level === topology.preferredPlacement"
            class="constraint-badge constraint-badge--outline text-b-body1"
          >
            Preferred
          </span>

          <span
            v-if="topology.actualPlacement && level === topology.actualPlacement.level"
            class="constraint-badge constraint-badge--filled text-b-body1"
          >
            Actual
          </span>

          <span
            v-if="constraintStatusByLevel.get(level)"
            :aid="`topology-constraint-status-${level}`"
            class="constraint-icon"
          >
            <q-icon
              :name="constraintStatusByLevel.get(level)?.icon || ''"
              :class="constraintStatusByLevel.get(level)?.colorClass || ''"
              size="14px"
            />
            <q-tooltip max-width="300px">{{ constraintStatusByLevel.get(level)?.tooltip }}</q-tooltip>
          </span>
        </div>
      </div>
    </runai-details-panel-row>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { WorkloadTopology } from "@/swagger-models/workloads-client";
import { RunaiDetailsPanelRow } from "@/components/common/runai-details-panel";

defineOptions({ name: "workload-topology-section" });

interface ConstraintStatus {
  icon: string;
  colorClass: string;
  tooltip: string;
}

enum EConstraintType {
  Required = "required",
  Preferred = "preferred",
}

enum EConstraintState {
  Met = "met",
  Unmet = "unmet",
}

const CONSTRAINT_STATUS: Record<EConstraintType, Record<EConstraintState, ConstraintStatus>> = {
  [EConstraintType.Required]: {
    [EConstraintState.Met]: {
      icon: "ph-bold ph-check-circle",
      colorClass: "constraint-icon--success",
      tooltip:
        "The required topology constraint was met. Pods are placed according to the required topology, ensuring optimal pod-to-pod communication and performance.",
    },
    [EConstraintState.Unmet]: {
      icon: "ph-bold ph-warning-circle",
      colorClass: "constraint-icon--warning",
      tooltip: "The required topology constraint was not met. The workload may not be placed as expected.",
    },
  },
  [EConstraintType.Preferred]: {
    [EConstraintState.Met]: {
      icon: "ph-bold ph-check-circle",
      colorClass: "constraint-icon--success",
      tooltip:
        "The preferred topology constraint was met. Pods are placed according to the preferred topology, ensuring optimal pod-to-pod communication and performance.",
    },
    [EConstraintState.Unmet]: {
      icon: "ph-bold ph-warning-circle",
      colorClass: "constraint-icon--warning",
      tooltip:
        "The preferred topology constraint was not met. Pods may be placed farther apart than preferred, which can impact communication latency and performance. Consider adjusting the topology constraints.",
    },
  },
};

export interface Props {
  topology: WorkloadTopology;
  levels: string[];
}

const props = defineProps<Props>();

const actualPlacementDisplay = computed(() => {
  const ap = props.topology.actualPlacement;
  if (!ap?.level || !ap?.value) return null;
  return { level: ap.level, value: ap.value };
});

const constraintStatusByLevel = computed<Map<string, ConstraintStatus>>(() => {
  const map = new Map<string, ConstraintStatus>();
  const actualLevel = props.topology.actualPlacement?.level;
  if (!actualLevel) return map;

  // Levels are ordered broadest → most specific (e.g., region → zone → rack → hostname).
  // A constraint is met if actual placement is at the same level or more specific (higher index).
  const actualIdx = props.levels.indexOf(actualLevel);

  for (const level of props.levels) {
    const constraintType: EConstraintType | null =
      level === props.topology.requiredPlacement
        ? EConstraintType.Required
        : level === props.topology.preferredPlacement
          ? EConstraintType.Preferred
          : null;
    if (!constraintType) continue;

    const constraintIdx = props.levels.indexOf(level);
    const isMet = constraintIdx !== -1 && actualIdx !== -1 && actualIdx >= constraintIdx;
    const state = isMet ? EConstraintState.Met : EConstraintState.Unmet;
    map.set(level, CONSTRAINT_STATUS[constraintType][state]);
  }
  return map;
});
</script>

<style scoped lang="scss">
.workload-topology-section {
  gap: var(--space-xs);

  .levels {
    gap: var(--space-2xs);
  }

  .mono-badge {
    background: var(--bg-secondary-default, #{$body-background-color});
    border-radius: var(--space-2xs);
    padding: var(--space-3xs) var(--space-xs);
  }

  .constraint-badge {
    border-radius: var(--space-3xs);
    padding: 0 var(--space-xs);
    text-align: center;
  }

  .constraint-badge--outline {
    color: var(--b-text-sub, #{$black-54});
    border: 1px solid var(--border-strong);
  }

  .constraint-badge--filled {
    background-color: var(--feedback-default-bg, #{$grey-14});
    color: var(--text-inverse, #ffffff);
  }

  .constraint-icon {
    display: inline-flex;
    align-items: center;
  }

  .constraint-icon--success {
    color: var(--text-success);
  }

  .constraint-icon--warning {
    color: var(--icon-warning);
  }
}
</style>
