<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <div class="node-pool-edit-form-v2">
    <runai-wizard-layout
      :steps="steps"
      :active-step="activeStep"
      @step-click="onStepClick"
      width="1000px"
      margin-top="26px"
      left-width="200px"
      right-width="200px"
      hide-right-on-small-screen
    >
      <runai-form-wrapper ref="form" :form-state="draft" full-width greedy @leave-page="emit('leave-approved')">
        <runai-expansion-wrapper no-first-top-padding>
          <runai-expansion-item label="Cluster" default-opened disable-closing>
            <runai-section-title title="Select the cluster under which to create this node pool" />
            <div class="cluster-field">
              <runai-cluster-dropdown
                :cluster-id="draft.clusterId"
                :disabled="locked"
                @update:cluster-id="draft.clusterId = $event"
              />
            </div>
          </runai-expansion-item>
          <runai-expansion-item label="Node pool name & description" default-opened disable-closing>
            <runai-name-validation
              v-model="draft.name"
              class="name-field"
              aid="node-pool-name"
              :debounce="0"
              dense
              :disable="locked"
              :rules="[validateName, uniqueName, nameConflictRule]"
              :maxlength="40"
            />
            <!-- maxlength is the shared character cap; the rule also enforces the API UTF-8 byte limit. -->
            <runai-description-input
              v-model="draft.description"
              class="description-field q-mt-lg"
              custom-aid="node-pool-description"
              dense
              :disable="locked"
              :debounce="0"
              :rules="[validateDescription]"
              no-error-icon
            />
          </runai-expansion-item>
          <runai-expansion-item label="Label" default-opened disable-closing :loading="loading">
            <runai-section-title title="Set the label for the nodes that belong to this node pool" />
            <q-banner v-if="loadError" class="bg-red-1 text-negative q-mb-md" role="alert">
              {{ loadError }}
              <template #action
                ><runai-btn aid="retry-node-pools" label="Retry" flat color="primary" @click="loadPools"
              /></template>
            </q-banner>
            <q-select
              v-model="draft.labelKey"
              class="label-field"
              aid="node-pool-label-key"
              dense
              label="Key"
              :options="keyOptions"
              option-label="value"
              use-input
              fill-input
              hide-selected
              emit-value
              map-options
              input-debounce="0"
              :disable="locked || !draft.clusterId || !!loadError"
              :rules="[validateLabelKey]"
              no-error-icon
              @filter="filterKeys"
              @new-value="addKey"
            >
              <template #option="scope">
                <q-item v-bind="scope.itemProps">
                  <q-item-section>{{ scope.opt.label }}</q-item-section>
                </q-item>
              </template>
            </q-select>
            <q-select
              v-if="draft.labelKey"
              v-model="draft.labelValue"
              class="label-field q-mt-lg"
              aid="node-pool-label-value"
              dense
              label="Value"
              :options="valueOptions"
              option-label="value"
              use-input
              fill-input
              hide-selected
              emit-value
              map-options
              input-debounce="0"
              :disable="locked || !!loadError"
              :rules="[validateLabelValue, uniqueLabel]"
              reactive-rules
              no-error-icon
              @filter="filterValues"
              @new-value="addValue"
            >
              <template #option="scope">
                <q-item v-bind="scope.itemProps">
                  <q-item-section>{{ scope.opt.label }}</q-item-section>
                  <q-tooltip v-if="scope.opt.disable">{{ nodePoolErrors.NODE_POOL_LABEL_IN_USE }}</q-tooltip>
                </q-item>
              </template>
            </q-select>
          </runai-expansion-item>
        </runai-expansion-wrapper>
      </runai-form-wrapper>
      <template #content-footer>
        <div class="creation-footer">
          <div class="row justify-end q-gutter-md">
            <runai-btn
              aid="create-node-pool-close"
              :label="created ? 'Close' : 'Create node pool & close'"
              flat
              color="primary"
              :loading="submitting && destination === ENodePoolFormDestination.Close"
              :disable="busy"
              @click="close"
            />
            <runai-btn
              aid="create-node-pool-continue"
              label="Create node pool & continue"
              color="primary"
              :loading="submitting && destination === ENodePoolFormDestination.Continue"
              :disable="busy"
              @click="submit(ENodePoolFormDestination.Continue)"
            />
          </div>
        </div>
      </template>
      <template #right-sidebar>
        <node-pool-wizard-summary :pool="created ?? draft" :cluster-name="clusterName" />
      </template>
    </runai-wizard-layout>
    <unsaved-changes-dialog
      v-if="unsavedGuard.dialog.isOpen.value"
      :model-value="true"
      @update:model-value="unsavedGuard.cancelPending()"
      @save="saveStepChanges"
      @discard="unsavedGuard.confirmAction('discard')"
    />
  </div>
</template>
<script setup lang="ts">
import { computed, onScopeDispose, ref, watch } from "vue";
import { useQuasar } from "quasar";
import NodePoolWizardSummary from "@/components/node-pools/node-pool-wizard-summary/node-pool-wizard-summary.vue";
import RunaiWizardLayout from "@/components/common/runai-wizard-layout/runai-wizard-layout.vue";
import RunaiFormWrapper from "@/components/common/runai-form-wrapper/runai-form-wrapper.vue";
import RunaiExpansionWrapper from "@/components/common/runai-expansion-wrapper/runai-expansion-wrapper.vue";
import RunaiExpansionItem from "@/components/common/runai-expansion-item/runai-expansion-item.vue";
import RunaiClusterDropdown from "@/components/common/runai-cluster-selection/runai-cluster-dropdown/runai-cluster-dropdown.vue";
import RunaiDescriptionInput from "@/components/common/runai-description-input/runai-description-input.vue";
import RunaiNameValidation from "@/components/common/runai-name-validation/runai-name-validation.vue";
import RunaiSectionTitle from "@/components/common/runai-section-title/runai-section-title.vue";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { useClusterStore } from "@/stores/cluster.store";
import { useNodePoolStore } from "@/stores/node-pool.store";
import { alertUtil } from "@/utils/alert.util";
import { deepCopy } from "@/utils/common.util";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import { useNodePoolForm } from "@/composables/node-pool/use-node-pool-form.composable";
import {
  labelOptions,
  validateName,
  validateDescription,
  validateLabelKey,
  validateLabelValue,
} from "@/utils/node-pool.util/node-pool.util";
import { errorMessages as nodePoolErrors, errorMessages } from "@/common/error-message.constant";
import { type INodePoolFormDraft } from "@/components/node-pools/node-pool-edit-form-v2/node-pool-edit-form-v2.model";

import { ENodePoolFormDestination } from "./node-pool-edit-form-v2.model";
import { useWizardLayout } from "@/composables/use-wizard-layout.composable";
import { provideUnsavedChangesGuard } from "@/composables/org-unit/use-unsaved-changes-guard.composable";
import { UnsavedChangesDialog } from "@/components/org-unit/v2/common/unsaved-changes-dialog";
const props = withDefaults(defineProps<{ initialClusterId?: string }>(), { initialClusterId: "" });
const emit = defineEmits<{
  "leave-approved": [];
  "busy-change": [busy: boolean];
  created: [pool: Nodepool, draft: INodePoolFormDraft, destination: ENodePoolFormDestination];
}>();
const { draft, pools, loading, loadError, submitError, submitting, created, nameConflict, loadPools, create } =
  useNodePoolForm(props.initialClusterId);
const clusterStore = useClusterStore();
const nodePoolStore = useNodePoolStore();
const $q = useQuasar();
const form = ref<InstanceType<typeof RunaiFormWrapper>>();
const nameConflictRule = computed(() => (nameConflict.value ? () => errorMessages.NAME_ALREADY_EXIST : () => true));
const validating = ref(false);
const destination = ref<ENodePoolFormDestination>(ENodePoolFormDestination.Continue);
// Disabling a Quasar field clears its validation errors, so lock only once saving starts.
const locked = computed(() => submitting.value || !!created.value);
const busy = computed(() => validating.value || submitting.value || loading.value || !!loadError.value);
watch(
  () => validating.value || submitting.value,
  (value) => emit("busy-change", value),
  { flush: "sync" },
);
const clusterName = computed(() => clusterStore.clusterList.find((c) => c.uuid === draft.clusterId)?.name);
const { steps, activeStep, handleStepClick, markStepComplete } = useWizardLayout(
  [
    { name: "identity", title: "Cluster & name", done: false },
    { name: "scheduling", title: "Scheduling mode", done: false, disabled: true },
    { name: "topology", title: "Topology & placement", done: false, disabled: true },
  ],
  "identity",
);
watch(created, (pool) => {
  if (pool) markStepComplete("identity");
});
const unsavedGuard = provideUnsavedChangesGuard();
const initialDraft = deepCopy(draft);
const unregisterGuard = unsavedGuard.registerStepGuard("identity", {
  isDirty: () => !!form.value?.isDirty,
  save: async () => {
    await submit(ENodePoolFormDestination.Continue);
    if (!created.value) throw new Error(errorMessages.NODE_POOL_SAVE_FAILED);
  },
  discard: () => {
    Object.assign(draft, deepCopy(initialDraft));
    form.value?.initFirstState();
    form.value?.updateDirtyState();
  },
});
onScopeDispose(unregisterGuard);
async function onStepClick(payload: Parameters<typeof handleStepClick>[0]) {
  if (
    payload.name === activeStep.value ||
    busy.value ||
    !steps.value.some((step) => step.name === payload.name && !step.disabled)
  )
    return;
  await unsavedGuard.requestAction({ stepName: activeStep.value, onProceed: () => handleStepClick(payload) });
}
async function saveStepChanges() {
  try {
    await unsavedGuard.confirmAction("save");
  } catch {
    /* Validation keeps the user on the current step. */
  }
}
const keySearch = ref("");
const valueSearch = ref("");
const unavailableValues = computed(() =>
  pools.value
    .filter((p) => p.labelKey === draft.labelKey)
    .map((p) => p.labelValue)
    .filter((v): v is string => typeof v === "string"),
);
const keyOptions = computed(() =>
  labelOptions(
    [...pools.value.map((p) => p.labelKey).filter((v): v is string => typeof v === "string")],
    keySearch.value,
    [],
    "key",
  ),
);
const valueOptions = computed(() => labelOptions([], valueSearch.value, unavailableValues.value));
function filterKeys(value: string, update: (fn: () => void) => void) {
  update(() => {
    keySearch.value = value;
  });
}
function filterValues(value: string, update: (fn: () => void) => void) {
  update(() => {
    valueSearch.value = value;
  });
}
function addKey(value: string, done: (value: string, mode: "add-unique") => void) {
  done(value, "add-unique");
}
function addValue(value: string, done: (value: string, mode: "add-unique") => void) {
  if (!unavailableValues.value.includes(value)) done(value, "add-unique");
}
function uniqueLabel(value: string) {
  return !unavailableValues.value.includes(value) || nodePoolErrors.NODE_POOL_LABEL_IN_USE;
}
function uniqueName(value: string) {
  return !pools.value.some((p) => p.name === value) || errorMessages.NAME_ALREADY_EXIST;
}
function close() {
  if (busy.value) return;
  if (created.value) emit("created", created.value, deepCopy(draft), ENodePoolFormDestination.Close);
  else void submit(ENodePoolFormDestination.Close);
}
async function submit(target: ENodePoolFormDestination) {
  if (busy.value) return;
  destination.value = target;
  if (!created.value) {
    validating.value = true;
    let valid: boolean;
    try {
      valid = !!(await form.value?.validate());
    } finally {
      validating.value = false;
    }
    if (!valid) return;
  }
  const result = await create();
  if (created.value) {
    nodePoolStore.removeNodePoolsByClusterId(draft.clusterId);
    form.value?.initFirstState();
    form.value?.updateDirtyState();
  }
  if (!result) {
    if (submitError.value) $q.notify(alertUtil.getError(`Node pool ${draft.name} couldn't be created`));
    return;
  }
  $q.notify(alertUtil.getSuccess(`Node pool ${draft.name} created`));
  emit("created", result, deepCopy(draft), target);
}
</script>
<style scoped lang="scss">
.node-pool-edit-form-v2 {
  padding: 0 20px 26px;
}
.cluster-field {
  max-width: 655px;
  margin-top: 26px;
}
.name-field {
  max-width: 510px;
}
.description-field {
  max-width: 702px;
}
.label-field {
  max-width: 380px;
  margin-top: 26px;
}
.creation-footer {
  width: 100%;
  margin-top: 40px;
}
</style>
