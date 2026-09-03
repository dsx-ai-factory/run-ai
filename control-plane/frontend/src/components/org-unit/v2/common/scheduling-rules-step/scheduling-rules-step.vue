<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <scheduling-rules-section
      v-model:scheduling-rules="localSchedulingRules"
      v-model:node-types="localNodeTypes"
      :cluster-id="clusterId"
      :entity="entityType"
      :parent-scheduling-rules="parentSchedulingRules"
      :parent-node-types="parentNodeTypes"
      :loading="loading"
      :read-only="readOnly"
      default-opened
      disable-closing
      :show-save-button="!isSingleNodePool"
      prevent-top-padding-when-open
      hide-loading-skeleton
      @save="handleSave"
    />

    <!-- Single node pool: show footer-style buttons (like first step) -->
    <q-card-section v-if="isSingleNodePool && !readOnly" class="row justify-end q-gutter-x-sm q-px-none">
      <q-btn flat color="primary" label="Close" @click="emit('close')" />
      <q-btn color="primary" label="Save & Continue" :loading="loading" @click="handleSaveAndContinue" />
    </q-card-section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount, onMounted, nextTick } from "vue";
import { QBtn, QCardSection } from "quasar";
import type { NodeTypesPerWorkload, SchedulingRules } from "@/swagger-models/org-unit-service-client";
import { SchedulingRulesSection } from "@/components/project/project-edit-form/scheduling-rules-section";
import type { EQuotaEntity } from "@/models/resource.model";
import { deepCopy } from "@/utils/common.util";
import { toString } from "@/utils/string.util";
import { useUnsavedChangesGuard } from "@/composables/org-unit/use-unsaved-changes-guard.composable";
import { useWizardPreview } from "@/composables/org-unit/use-wizard-preview.composable";
import { orgUnitUtil } from "@/utils/org-unit.util";
import { EStepNames } from "@/models/org-unit-forms.model";

export interface Props {
  entityType: EQuotaEntity;
  clusterId: string;
  schedulingRules?: SchedulingRules | null;
  nodeTypes?: NodeTypesPerWorkload;
  parentSchedulingRules?: SchedulingRules | null;
  parentNodeTypes?: NodeTypesPerWorkload;
  loading?: boolean;
  readOnly?: boolean;
  /** When true, hides the inner save button and shows footer-style buttons instead */
  isSingleNodePool?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  schedulingRules: null,
  nodeTypes: undefined,
  parentSchedulingRules: null,
  parentNodeTypes: undefined,
  loading: false,
  readOnly: false,
  isSingleNodePool: false,
});

const emit = defineEmits<{
  (
    e: "save-scheduling-rules",
    payload: { schedulingRules: SchedulingRules | null; nodeTypes?: NodeTypesPerWorkload; navigateAfterSave?: boolean },
  ): void;
  (e: "close"): void;
}>();

const STEP_NAME = EStepNames.SchedulingRules;

const cloneSchedulingRules = (rules?: SchedulingRules | null) => (rules ? deepCopy(rules) : null);
const cloneNodeTypes = (types?: NodeTypesPerWorkload) => (types ? deepCopy(types) : undefined);

const localSchedulingRules = ref<SchedulingRules | null>(null);
const localNodeTypes = ref<NodeTypesPerWorkload | undefined>(cloneNodeTypes(props.nodeTypes));

const savedSchedulingRules = ref<SchedulingRules | null>(null);
const savedNodeTypes = ref<NodeTypesPerWorkload | undefined>(cloneNodeTypes(props.nodeTypes));

const getInitializedSchedulingRules = (rules?: SchedulingRules | null): SchedulingRules | null => {
  const entityRules = (rules ? deepCopy(rules) : {}) as SchedulingRules;
  const parentRules = (props.parentSchedulingRules ? deepCopy(props.parentSchedulingRules) : {}) as SchedulingRules;
  const merged = orgUnitUtil.initSchedulingRules(entityRules, parentRules);
  return Object.keys(merged).length > 0 ? merged : null;
};

const syncSchedulingRulesFromProps = () => {
  const mergedRules = getInitializedSchedulingRules(props.schedulingRules);
  localSchedulingRules.value = cloneSchedulingRules(mergedRules);
  savedSchedulingRules.value = cloneSchedulingRules(mergedRules);
};

syncSchedulingRulesFromProps();

// Normalize scheduling rules by removing keys with null/undefined values
// This ensures that { field: null } is treated the same as {} or null
const normalizeSchedulingRules = (rules: SchedulingRules | null): SchedulingRules | null => {
  if (!rules) return null;

  const normalized = Object.entries(rules).reduce((acc, [key, value]) => {
    if (value !== null && value !== undefined) {
      acc[key as keyof SchedulingRules] = value;
    }
    return acc;
  }, {} as Partial<SchedulingRules>);

  return Object.keys(normalized).length > 0 ? (normalized as SchedulingRules) : null;
};

const serializeState = () =>
  toString({
    rules: normalizeSchedulingRules(localSchedulingRules.value),
    nodeTypes: localNodeTypes.value,
  });

const serializeSavedState = () =>
  toString({
    rules: normalizeSchedulingRules(savedSchedulingRules.value),
    nodeTypes: savedNodeTypes.value,
  });

const isDirty = computed(() => serializeState() !== serializeSavedState());

const unsavedGuard = useUnsavedChangesGuard();
const wizardPreview = useWizardPreview();

// Update wizard preview when scheduling rules or node types change
watch(
  [localSchedulingRules, localNodeTypes],
  ([rules, types]) => {
    wizardPreview?.updatePreview("schedulingRules", {
      schedulingRules: rules ? deepCopy(rules) : null,
      nodeTypes: types ? deepCopy(types) : undefined,
    });
  },
  { immediate: true, deep: true },
);

const handleSave = (navigateAfterSave = false) => {
  emit("save-scheduling-rules", {
    schedulingRules: localSchedulingRules.value ? deepCopy(localSchedulingRules.value) : null,
    nodeTypes: localNodeTypes.value ? deepCopy(localNodeTypes.value) : undefined,
    navigateAfterSave,
  });
  savedSchedulingRules.value = cloneSchedulingRules(localSchedulingRules.value);
  savedNodeTypes.value = cloneNodeTypes(localNodeTypes.value);
};

const handleSaveAndContinue = () => {
  handleSave(true);
};

const handleDiscard = () => {
  localSchedulingRules.value = cloneSchedulingRules(savedSchedulingRules.value);
  localNodeTypes.value = cloneNodeTypes(savedNodeTypes.value);
};

const unregisterGuard = unsavedGuard.registerStepGuard(STEP_NAME, {
  isDirty: () => isDirty.value,
  save: handleSave,
  discard: handleDiscard,
});

// Re-snapshot saved state after child component initialization.
// When scheduling-rules-section mounts, its watchers may re-emit node types with a
// slightly different shape (e.g. extra keys). Capturing the post-init state prevents
// false-positive dirty detection.
onMounted(async () => {
  await nextTick();
  savedSchedulingRules.value = cloneSchedulingRules(localSchedulingRules.value);
  savedNodeTypes.value = cloneNodeTypes(localNodeTypes.value);
});

onBeforeUnmount(() => {
  unregisterGuard();
  wizardPreview?.clearPreview("schedulingRules");
});

defineExpose({
  isDirty,
  save: handleSave,
  discardChanges: handleDiscard,
});
</script>

<style scoped></style>
