<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-form ref="ruleForm" greedy class="policy-rule-card q-pa-md" :class="{ 'policy-rule-card--embedded': embedded }">
    <!-- no RunAI wrapper: inner q-form needed for isolated rule validation -->
    <template v-if="!embedded">
      <div class="text-black-87">{{ isEdit ? "Edit rule" : "New rule" }}</div>
      <q-separator class="rule-divider q-mt-md q-mb-lg" />
    </template>

    <div class="step">
      <div class="text-black-87 q-mb-md">Select a field from the spec</div>
      <div class="text-b-body1 text-b-label field-label">Field</div>
      <runai-select
        aid="rule-field-select"
        class="field-select-control select-italic-placeholder"
        dense
        filterable
        :no-option-text="NO_RESULTS_FOUND_LABEL"
        :model-value="draft.field"
        :options="fieldOptions"
        @update:model-value="onFieldChange"
        emit-value
        placeholder="Select..."
        :rules="[isFieldSelected]"
      />
    </div>

    <template v-if="draft.field">
      <q-separator class="rule-step-divider q-mt-md q-mb-lg" />
      <div class="step">
        <div class="text-black-87 q-mb-md">Set the spec selector and its parameters</div>
        <div class="text-b-body1 text-b-label field-label">Spec selector</div>
        <runai-select
          aid="rule-spec-selector-select"
          class="rule-control select-italic-placeholder"
          dense
          filterable
          :no-option-text="NO_RESULTS_FOUND_LABEL"
          :model-value="draft.specSelector"
          :options="specSelectorOptions"
          @update:model-value="onSpecSelectorChange"
          emit-value
          placeholder="Select..."
          :rules="[isSpecSelectorSelected]"
        />
        <rule-selector-params
          v-if="fieldInfo && fieldInfo.specSelectorParams.length"
          :params="fieldInfo.specSelectorParams"
          :values="draft.selectorParams"
          @update="updateSelectorParam"
        />
      </div>
    </template>

    <template v-if="draft.specSelector && isItemized">
      <q-separator class="rule-step-divider q-mt-md q-mb-lg" />
      <div class="step">
        <div class="text-black-87 q-mb-md">Set the item selector and its parameters</div>
        <div class="text-b-body1 text-b-label field-label">Item selector</div>
        <runai-select
          aid="rule-item-selector-select"
          class="rule-control select-italic-placeholder"
          dense
          filterable
          :no-option-text="NO_RESULTS_FOUND_LABEL"
          :model-value="draft.itemSelector"
          :options="itemSelectorOptions"
          @update:model-value="onItemSelectorChange"
          emit-value
          placeholder="Select..."
          :rules="[isItemSelectorSelected]"
        />
        <rule-selector-params
          v-if="fieldInfo && fieldInfo.itemSelectorParams.length"
          :params="fieldInfo.itemSelectorParams"
          :values="draft.selectorParams"
          @update="updateSelectorParam"
        />
      </div>
    </template>

    <q-linear-progress v-if="loadingFieldInfo" animation-speed="100" indeterminate class="q-my-sm" />

    <template v-if="showEnforcements && fieldInfo">
      <q-separator class="rule-step-divider q-mt-md q-mb-lg" />
      <div class="step">
        <rule-enforcement-section
          :enforcements="draft.enforcements"
          :field-info="fieldInfo"
          :show-validation="showFormError"
          @update:enforcement="updateEnforcement"
          @add="addEnforcement"
          @remove="removeEnforcement"
        />
      </div>
    </template>

    <div v-if="ruleError" class="text-negative text-caption q-mt-sm" aid="rule-card-error">{{ ruleError }}</div>

    <div class="rule-footer row justify-end items-center">
      <div v-if="footerMessage" class="text-negative text-caption q-mr-auto" aid="rule-footer-error">
        {{ footerMessage }}
      </div>
      <runai-btn aid="rule-cancel-btn" flat color="primary" label="cancel" class="q-mr-sm" @click="emit('cancel')" />
      <runai-btn aid="rule-save-btn" flat color="primary" label="save rule" @click="commit" />
    </div>
  </q-form>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiSelect } from "@/components/common/runai-select";
import RuleSelectorParams from "./rule-selector-params.vue";
import RuleEnforcementSection from "./rule-enforcement-section.vue";
import { NO_RESULTS_FOUND_LABEL } from "./policy-rule-builder.constant";
import type { EEnforcementType, IEnforcementDraft, IFieldInfo, IRuleDraft, PolicyRule } from "@/models/gvk-policy.model";
import type { PolicyMapping } from "@/composables/use-policy-mapping.composable";
import type { ISelectOption } from "@/models/global.model";
import { gvkPolicyUtil } from "@/utils/gvk-policy.util";
import { deepCopy } from "@/utils/common.util";
import { isNotEmpty } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { ErrorAlert } from "@/utils/error-alert.util";
import { QForm, useQuasar } from "quasar";

export interface Props {
  mapping: PolicyMapping;
  initialDraft: IRuleDraft;
  isEdit?: boolean;
  embedded?: boolean;
  existingRuleKeys?: string[];
}

const props = withDefaults(defineProps<Props>(), {
  isEdit: false,
  embedded: false,
  existingRuleKeys: () => [],
});

const emit = defineEmits<{
  (event: "save", payload: { draft: IRuleDraft; rule: PolicyRule }): void;
  (event: "cancel"): void;
}>();

const $q = useQuasar();
const ruleForm = ref<QForm | null>(null);
const draft = ref<IRuleDraft>(deepCopy(props.initialDraft));
const fieldInfo = ref<IFieldInfo | null>(null);
const loadingFieldInfo = ref<boolean>(false);
const ruleError = ref<string>("");
const showFormError = ref<boolean>(false);
const lastValidationPassed = ref<boolean>(false);
const footerError = computed<string>(() => {
  if (!showFormError.value) return "";
  return lastValidationPassed.value ? errorMessages.UNSAVED_RULE_CHANGES : errorMessages.FORM_INCOMPLETE;
});
const duplicateError = ref<string>("");
const footerMessage = computed<string>(() => duplicateError.value || footerError.value);

onMounted(async () => {
  if (draft.value.field && draft.value.specSelector) {
    if (isItemized.value && !draft.value.itemSelector && singleItemSelector.value) {
      draft.value.itemSelector = singleItemSelector.value;
    }
    await resolveFieldInfo();
    return;
  }
  if (draft.value.field && singleSpecSelector.value) {
    await onSpecSelectorChange(singleSpecSelector.value);
  }
});

function handleDocumentClick(event: MouseEvent): void {
  const target = event.target;
  if (target instanceof Element && target.closest(".policy-rule-card, .q-menu, .q-dialog")) return;
  void runValidation().then((valid) => {
    lastValidationPassed.value = valid;
    showFormError.value = !valid;
  });
}
onMounted(() => document.addEventListener("click", handleDocumentClick));
onBeforeUnmount(() => document.removeEventListener("click", handleDocumentClick));

const fieldOptions = computed<ISelectOption[]>(() =>
  props.mapping.fields().map((field) => ({ label: field, value: field, aid: `rule-field-option-${field}` })),
);
const specSelectorOptions = computed<ISelectOption[]>(() =>
  props.mapping.specSelectorsForField(draft.value.field).map((name) => ({
    label: name,
    value: name,
    aid: `rule-spec-option-${name}`,
  })),
);
const itemSelectorOptions = computed<ISelectOption[]>(() =>
  props.mapping
    .itemSelectorsFor(draft.value.field, draft.value.specSelector)
    .map((name) => ({ label: name, value: name, aid: `rule-item-option-${name}` })),
);
const isItemized = computed<boolean>(() => itemSelectorOptions.value.length > 0);
const singleSpecSelector = computed<string | null>(() =>
  specSelectorOptions.value.length === 1 ? (specSelectorOptions.value[0].value as string) : null,
);
const singleItemSelector = computed<string | null>(() =>
  itemSelectorOptions.value.length === 1 ? (itemSelectorOptions.value[0].value as string) : null,
);
const showEnforcements = computed<boolean>(() => {
  if (!draft.value.field || !draft.value.specSelector) return false;
  if (isItemized.value && !draft.value.itemSelector) return false;
  return true;
});

// `preselectParams` fills single-option selector params on user-driven selector changes; left off on
// mount so editing a saved rule never silently changes it (which would also mark the form dirty).
async function resolveFieldInfo(preselectParams = false): Promise<void> {
  if (!draft.value.field || !draft.value.specSelector) {
    fieldInfo.value = null;
    return;
  }
  loadingFieldInfo.value = true;
  try {
    fieldInfo.value = await props.mapping.resolveFieldInfo(
      draft.value.field,
      draft.value.specSelector,
      draft.value.itemSelector,
    );
    if (preselectParams && fieldInfo.value) {
      draft.value.selectorParams = gvkPolicyUtil.fillSingleOptionSelectorParams(
        fieldInfo.value,
        draft.value.selectorParams,
      );
    }
  } catch (error: unknown) {
    fieldInfo.value = null;
    const errorAlert = new ErrorAlert({ generalMessage: ErrorAlert.failedLoadingMessage("field info") });
    $q.notify(errorAlert.getNotification(error));
  } finally {
    loadingFieldInfo.value = false;
  }
}
async function prefetchFieldInfo(): Promise<void> {
  if (!draft.value.field) return;
  loadingFieldInfo.value = true;
  try {
    await props.mapping.loadFieldInfo(draft.value.field);
  } catch (error: unknown) {
    const errorAlert = new ErrorAlert({ generalMessage: ErrorAlert.failedLoadingMessage("field info") });
    $q.notify(errorAlert.getNotification(error));
  } finally {
    loadingFieldInfo.value = false;
  }
}
async function onFieldChange(field: string): Promise<void> {
  draft.value.field = field;
  draft.value.specSelector = null;
  draft.value.itemSelector = null;
  draft.value.selectorParams = {};
  draft.value.enforcements = [gvkPolicyUtil.createEmptyEnforcement()];
  fieldInfo.value = null;
  ruleError.value = "";
  showFormError.value = false;
  duplicateError.value = "";
  await prefetchFieldInfo();
  if (singleSpecSelector.value) await onSpecSelectorChange(singleSpecSelector.value);
}
async function onSpecSelectorChange(specSelector: string): Promise<void> {
  draft.value.specSelector = specSelector;
  draft.value.itemSelector = null;
  draft.value.selectorParams = {};
  draft.value.enforcements = [gvkPolicyUtil.createEmptyEnforcement()];
  ruleError.value = "";
  showFormError.value = false;
  duplicateError.value = "";
  if (!isItemized.value) {
    await resolveFieldInfo(true);
  } else if (singleItemSelector.value) {
    draft.value.itemSelector = singleItemSelector.value;
    await resolveFieldInfo(true);
  } else {
    await resolveFieldInfo(true);
  }
}
async function onItemSelectorChange(itemSelector: string): Promise<void> {
  draft.value.itemSelector = itemSelector;
  draft.value.enforcements = [gvkPolicyUtil.createEmptyEnforcement()];
  ruleError.value = "";
  showFormError.value = false;
  duplicateError.value = "";
  await resolveFieldInfo(true);
}
function updateSelectorParam(name: string, value: string): void {
  ruleError.value = "";
  showFormError.value = false;
  duplicateError.value = "";
  draft.value.selectorParams = { ...draft.value.selectorParams, [name]: value };
}
function updateEnforcement(index: number, value: IEnforcementDraft): void {
  showFormError.value = false;
  draft.value.enforcements.splice(index, 1, value);
}
function addEnforcement(): void {
  showFormError.value = false;
  draft.value.enforcements.push(gvkPolicyUtil.createEmptyEnforcement());
}
function removeEnforcement(index: number): void {
  showFormError.value = false;
  draft.value.enforcements.splice(index, 1);
}
function isFieldSelected(val: string | null): boolean | string {
  return (!!val && isNotEmpty(val)) || errorMessages.SELECT_A_VALUE;
}
function isSpecSelectorSelected(val: string | null): boolean | string {
  return (!!val && isNotEmpty(val)) || errorMessages.SELECT_A_VALUE;
}
function isItemSelectorSelected(val: string | null): boolean | string {
  return (!!val && isNotEmpty(val)) || errorMessages.SELECT_A_VALUE;
}
function validateEnforcements(): string | null {
  if (draft.value.enforcements.length === 0) return errorMessages.ADD_AT_LEAST_ONE_ENFORCEMENT;
  const seen = new Set<EEnforcementType>();
  for (const enforcement of draft.value.enforcements) {
    if (enforcement.type) {
      if (!gvkPolicyUtil.enforcementMayRepeat(enforcement.type)) {
        if (seen.has(enforcement.type)) return errorMessages.DUPLICATE_ENFORCEMENT;
        seen.add(enforcement.type);
      }
    }
  }
  return (
    gvkPolicyUtil.enforcementCombinationError(draft.value.enforcements) ||
    (fieldInfo.value ? gvkPolicyUtil.setDefaultCombinationError(draft.value.enforcements, fieldInfo.value) : null)
  );
}
function hasEnforcementValueError(): boolean {
  return draft.value.enforcements.some(
    (enforcement) => !enforcement.type || gvkPolicyUtil.enforcementValueError(enforcement) !== null,
  );
}
async function runValidation(): Promise<boolean> {
  ruleError.value = "";
  if (!ruleForm.value) return false;
  const formValid = await ruleForm.value.validate();
  if (!formValid) return false;
  if (!fieldInfo.value) {
    ruleError.value = errorMessages.ENFORCEMENT_FIELD_NOT_READY;
    return false;
  }
  const enforcementError = validateEnforcements();
  if (enforcementError) {
    ruleError.value = enforcementError;
    return false;
  }
  if (hasEnforcementValueError()) return false;
  return true;
}
async function commit(): Promise<boolean> {
  duplicateError.value = "";
  const valid = await runValidation();
  lastValidationPassed.value = valid;
  showFormError.value = !valid;
  if (!valid || !fieldInfo.value) return false;
  const rule = gvkPolicyUtil.ruleDraftToApi(draft.value, fieldInfo.value);
  if (props.existingRuleKeys.includes(gvkPolicyUtil.ruleIdentityKey(rule))) {
    duplicateError.value = errorMessages.DUPLICATE_RULE;
    return false;
  }
  emit("save", { draft: deepCopy(draft.value), rule });
  return true;
}

function isDirty(): boolean {
  return JSON.stringify(draft.value) !== JSON.stringify(props.initialDraft);
}

async function attemptLeave(): Promise<boolean> {
  if (!isDirty()) return true;
  lastValidationPassed.value = await runValidation();
  showFormError.value = true;
  return false;
}

defineExpose({ commit, isDirty, attemptLeave });
</script>

<style lang="scss" scoped>
.policy-rule-card {
  border: 1px solid var(--border-subtle);
  border-radius: var(--b-border-radius-radius-xs);
  background: var(--bg-primary-default);

  &--embedded {
    border: none;
    border-radius: 0;
  }
}
.rule-divider {
  background-color: var(--border-subtle);
  margin-left: -16px;
  margin-right: -16px;
}
.rule-step-divider {
  background-color: var(--border-subtle);
}
.field-label {
  margin-bottom: -8px;
}
.rule-control {
  width: 256px;
  max-width: 100%;
}
.field-select-control {
  width: 400px;
  max-width: 100%;
}
.rule-footer {
  height: 48px;
  margin: 16px -16px -16px;
  padding: 0 16px;
  border-top: 1px solid var(--border-subtle);
}
</style>
