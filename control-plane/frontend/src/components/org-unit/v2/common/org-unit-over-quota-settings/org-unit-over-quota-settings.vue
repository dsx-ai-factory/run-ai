<template>
  <section class="org-unit-over-quota-settings">
    <div class="row no-wrap">
      <div class="col-auto row no-wrap items-center">
        <q-toggle
          v-model="localIsOverQuotaEnabled"
          color="primary"
          left-label
          :label="toggleLabel"
          aid="over-quota-toggle"
          @update:model-value="emitIsOverQuotaEnabled"
        />
        <runai-tooltip
          tooltip-text="Note: Workloads running over quota may stop at any time"
          tooltip-position="left"
          width="400px"
        />
      </div>

      <div class="col-auto" v-if="localIsOverQuotaEnabled">
        <div class="over-quota-container">
          <q-separator vertical class="q-ml-md q-mr-lg" />

          <org-unit-over-quota-limit-input
            v-model="localLimit"
            class="full-width"
            :deserved="deserved"
            :entity-type="entityType"
          />
          <runai-tooltip :tooltip-text="tooltipText" tooltip-position="right" />
        </div>
      </div>
    </div>
    <slot v-if="localIsOverQuotaEnabled"></slot>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { OrgUnitOverQuotaLimitInput } from "@/components/org-unit/v2/common/org-unit-over-quota-limit-input";
import { EQuotaEntity } from "@/models/resource.model";

export interface Props {
  limit: number | null;
  isOverQuotaEnabled: boolean;
  deserved: number;
  entityType?: EQuotaEntity;
  isSingleNodePool: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  entityType: EQuotaEntity.project,
});

const emit = defineEmits<{
  (e: "update:limit", value: number | null): void;
  (e: "update:is-over-quota-enabled", value: boolean): void;
}>();

const localLimit = computed({
  get: () => props.limit,
  set: (val: number | null) => emit("update:limit", val),
});

const localIsOverQuotaEnabled = computed({
  get: () => props.isOverQuotaEnabled,
  set: (val: boolean) => emit("update:is-over-quota-enabled", val),
});

const entityLabel = computed(() => {
  return props.entityType === "department" ? "department’s" : "project’s";
});

const toggleLabel = computed(() => {
  if (props.isSingleNodePool) {
    return `Allow workloads to exceed the ${props.entityType} quota`;
  } else {
    return `Allow workloads to exceed the ${entityLabel.value} quota in the selected node pool`;
  }
});

const emitIsOverQuotaEnabled = (value: boolean) => {
  emit("update:is-over-quota-enabled", value);
};

const tooltipText = computed(() => {
  return `The maximum GPU device allocation the ${props.entityType} can get from this node pool, representing the maximum sum of assigned quota and over-quota GPUs`;
});
</script>

<style scoped>
.over-quota-container {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  width: 220px;
}
</style>
