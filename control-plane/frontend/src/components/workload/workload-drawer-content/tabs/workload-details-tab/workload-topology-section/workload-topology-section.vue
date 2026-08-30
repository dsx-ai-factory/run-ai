<template>
  <div class="workload-topology-section">
    <div class="text-b-body1">Topology name: {{ topology.name }}</div>

    <div v-if="actualPlacementDisplay" class="row items-center b-gap-3xs text-b-body1">
      <span>Actual placement:</span>
      <span class="mono-badge text-b-mono">{{ actualPlacementDisplay.level }}</span>
      <span class="mono-badge text-b-mono">[{{ actualPlacementDisplay.value }}]</span>
    </div>

    <div v-if="levels.length > 0" class="column b-gap-2xs">
      <div class="text-b-body1">Topology levels:</div>
      <div v-for="level in levels" :key="level" class="topology-level-row row items-center b-gap-2xs">
        <span class="mono-badge text-b-mono">{{ level }}</span>

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

        <span v-if="getConstraintStatus(level)" :aid="`topology-constraint-status-${level}`" class="constraint-icon">
          <q-icon
            :name="getConstraintStatus(level)?.icon || ''"
            :class="getConstraintStatus(level)?.colorClass || ''"
            size="14px"
          />
          <q-tooltip max-width="300px">{{ getConstraintStatus(level)?.tooltip }}</q-tooltip>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { WorkloadTopology } from "@/swagger-models/workloads-client";

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

function getConstraintStatus(level: string): ConstraintStatus | null {
  const actualLevel = props.topology.actualPlacement?.level;

  const constraintType: EConstraintType | null =
    level === props.topology.requiredPlacement
      ? EConstraintType.Required
      : level === props.topology.preferredPlacement
        ? EConstraintType.Preferred
        : null;

  if (!constraintType) return null;
  if (!actualLevel) return null;

  // Levels are ordered broadest → most specific (e.g., region → zone → rack → hostname).
  // A constraint is met if actual placement is at the same level or more specific (higher index).
  const constraintIdx = props.levels.indexOf(level);
  const actualIdx = props.levels.indexOf(actualLevel);
  const isMet = constraintIdx !== -1 && actualIdx !== -1 && actualIdx >= constraintIdx;
  const state = isMet ? EConstraintState.Met : EConstraintState.Unmet;
  return CONSTRAINT_STATUS[constraintType][state];
}
</script>

<style scoped lang="scss">
.workload-topology-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.mono-badge {
  background-color: var(--bg-secondary-default);
  border-radius: var(--space-2xs);
  padding: var(--space-3xs) var(--space-xs);
}

.constraint-badge {
  border-radius: var(--space-3xs);
  padding: 0 var(--space-xs);
  text-align: center;
}

.constraint-badge--outline {
  color: var(--b-text-sub);
  border: 1px solid var(--border-strong);
}

.constraint-badge--filled {
  background-color: var(--feedback-default-bg);
  color: var(--feedback-text);
}

.constraint-icon {
  display: inline-flex;
  align-items: center;
}

.constraint-icon--success {
  color: var(--text-success);
  font-size: 14px;
}

.constraint-icon--warning {
  color: var(--icon-warning);
  font-size: 14px;
}
</style>
