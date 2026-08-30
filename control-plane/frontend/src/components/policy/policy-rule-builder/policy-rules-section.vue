<template>
  <runai-expansion-item
    class="policy-rules-section"
    label="Rules"
    default-opened
    :body-no-padding="true"
    :section-invalid="savedEntries.length === 0"
    :subheader="rulesSummary"
  >
    <section class="rules-content">
      <div v-if="$slots['mode-toggle']" class="q-mb-md"><slot name="mode-toggle" /></div>
      <div class="text-black-87 q-mb-md">
        Add rules to set the limits and suggested defaults that will be shown during workload submission
      </div>

      <q-linear-progress v-if="loadingMinimal" animation-speed="100" indeterminate class="q-mb-md" />

      <div v-if="savedEntries.length" class="saved-rules q-mb-sm">
        <template v-for="(entry, i) in savedEntries" :key="entry.draft.id">
          <div v-if="editingIndex === i && openDraft" class="rule-edit-wrap">
            <policy-rule-summary :rule="entry.draft" :index="i" @edit="cancelRule" @delete="deleteRule(i)" />
            <policy-rule-card
              :ref="setOpenCard"
              :mapping="mapping"
              :initial-draft="openDraft"
              :existing-rule-keys="existingRuleKeys"
              is-edit
              embedded
              @save="onRuleSave"
              @cancel="cancelRule"
            />
          </div>
          <policy-rule-summary
            v-else
            :rule="entry.draft"
            :index="i"
            @edit="attemptEditRule(i)"
            @delete="attemptDeleteRule(i)"
          />
        </template>
      </div>

      <policy-rule-card
        v-if="openDraft && editingIndex === null"
        :ref="setOpenCard"
        :mapping="mapping"
        :initial-draft="openDraft"
        :existing-rule-keys="existingRuleKeys"
        @save="onRuleSave"
        @cancel="cancelRule"
      />

      <div class="row items-center q-mt-sm">
        <runai-btn
          v-if="!openDraft"
          :disable="!canEditPolicies"
          flat
          color="primary"
          class="btn-icon-sm"
          icon="ph-bold ph-plus"
          label="Rule"
          aid="add-rule-btn"
          @click="attemptNewRule"
        />
        <runai-badge-mono
          v-if="syncBadgeLabel"
          class="q-ml-auto"
          show-badge-group
          :outlined-label="syncBadgeLabel"
          filled-label=""
          status="success"
          aid="rules-sync-badge"
        />
      </div>

      <div
        v-if="validationRevealed && displayedValidation"
        class="text-negative text-caption q-mt-sm"
        aid="rules-validation-error"
      >
        {{ displayedValidation }}
      </div>
    </section>
  </runai-expansion-item>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, type ComponentPublicInstance } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiBtn } from "@/components/common/runai-btn";
import PolicyRuleSummary from "./policy-rule-summary.vue";
import PolicyRuleCard from "./policy-rule-card.vue";
import { RunaiBadgeMono } from "@/components/common/runai-badge-mono";
import { usePolicyMapping } from "@/composables/use-policy-mapping.composable";
import { gvkPolicyUtil } from "@/utils/gvk-policy.util";
import { deepCopy } from "@/utils/common.util";
import { errorMessages } from "@/common/error-message.constant";
import type { GVK, IRuleDraft, PolicyRule } from "@/models/gvk-policy.model";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { usePermissionStore } from "@/stores/permissions.store";

interface ISavedRuleEntry {
  draft: IRuleDraft;
  rule: PolicyRule;
}

export interface Props {
  gvk: GVK;
  modelValue: PolicyRule[];
  syncBadgeLabel?: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (event: "update:modelValue", rules: PolicyRule[]): void;
}>();

const permissionStore = usePermissionStore();
// v-permission can't reach runai-btn (non-element root). Backend-verified: create needs Policies/CREATE, update needs Policies/UPDATE.
const canEditPolicies = computed<boolean>(
  () =>
    permissionStore.hasPermission(ResourceType.Policies, Action.Create) ||
    permissionStore.hasPermission(ResourceType.Policies, Action.Update),
);

const mapping = usePolicyMapping();
const { loadingMinimal } = mapping;

const savedEntries = ref<ISavedRuleEntry[]>([]);
const openDraft = ref<IRuleDraft | null>(null);
const editingIndex = ref<number | null>(null);
const existingRuleKeys = computed<string[]>(() =>
  savedEntries.value
    .filter((_, index) => index !== editingIndex.value)
    .map((entry) => gvkPolicyUtil.ruleIdentityKey(entry.rule)),
);
const validationRevealed = ref<boolean>(false);

onMounted(async () => {
  savedEntries.value = (props.modelValue || []).map((rule) => ({
    rule,
    draft: gvkPolicyUtil.apiRuleToDraft(rule),
  }));
  await mapping.loadMinimal(props.gvk);
});

function emitRules(): void {
  emit(
    "update:modelValue",
    savedEntries.value.map((entry) => entry.rule),
  );
}

const openCard = ref<InstanceType<typeof PolicyRuleCard> | null>(null);
function setOpenCard(el: Element | ComponentPublicInstance | null): void {
  openCard.value = el as InstanceType<typeof PolicyRuleCard> | null;
}

async function guardOpenRule(): Promise<boolean> {
  if (!openDraft.value || !openCard.value) return true;
  return await openCard.value.attemptLeave();
}
async function attemptEditRule(index: number): Promise<void> {
  if (await guardOpenRule()) editRule(index);
}
async function attemptDeleteRule(index: number): Promise<void> {
  if (await guardOpenRule()) deleteRule(index);
}
async function attemptNewRule(): Promise<void> {
  if (await guardOpenRule()) openNewRule();
}

function openNewRule(): void {
  editingIndex.value = null;
  openDraft.value = gvkPolicyUtil.createEmptyRuleDraft();
}

function editRule(index: number): void {
  editingIndex.value = index;
  openDraft.value = deepCopy(savedEntries.value[index].draft);
}

function deleteRule(index: number): void {
  if (editingIndex.value !== null) {
    if (editingIndex.value === index) cancelRule();
    else if (index < editingIndex.value) editingIndex.value -= 1;
  }
  savedEntries.value.splice(index, 1);
  emitRules();
}

function cancelRule(): void {
  openDraft.value = null;
  editingIndex.value = null;
}

function onRuleSave(payload: { draft: IRuleDraft; rule: PolicyRule }): void {
  const entry: ISavedRuleEntry = { draft: payload.draft, rule: payload.rule };
  if (editingIndex.value !== null) {
    savedEntries.value.splice(editingIndex.value, 1, entry);
  } else {
    savedEntries.value.push(entry);
  }
  emitRules();
  openDraft.value = null;
  editingIndex.value = null;
}

const rulesSummary = computed<string>(() => {
  const count = savedEntries.value.length;
  if (count === 0) return "None";
  return `${count} ${count === 1 ? "rule" : "rules"}`;
});

const rulesValidationValue = computed<string>(() => {
  if (openDraft.value !== null) return errorMessages.SAVE_OR_CANCEL_OPEN_RULE;
  if (savedEntries.value.length === 0) return errorMessages.ADD_AT_LEAST_ONE_RULE;
  return "";
});

const displayedValidation = computed<string>(() =>
  rulesValidationValue.value === errorMessages.SAVE_OR_CANCEL_OPEN_RULE ? "" : rulesValidationValue.value,
);

function revealValidation(): void {
  validationRevealed.value = true;
}

function isRuleDirty(): boolean {
  return openDraft.value !== null && !!openCard.value && openCard.value.isDirty();
}
async function saveOpenRule(): Promise<boolean> {
  if (!openCard.value) return true;
  return await openCard.value.commit();
}
function discardOpenRule(): void {
  cancelRule();
}

defineExpose({ rulesValidationValue, revealValidation, isRuleDirty, saveOpenRule, discardOpenRule });
</script>

<style lang="scss" scoped>
.rules-content {
  padding: var(--space-l, 20px) var(--space-m, 16px);
  min-height: 80px;
}

// Rows stack flush sharing one border line via the -1px pull below; only outer corners round.
.saved-rules > .policy-rule-summary,
.saved-rules > .rule-edit-wrap {
  border-radius: 0;
}
.saved-rules > * + * {
  margin-top: -1px;
}
.saved-rules > :first-child {
  border-top-left-radius: var(--b-border-radius-radius-xs);
  border-top-right-radius: var(--b-border-radius-radius-xs);
}
.saved-rules > :last-child {
  border-bottom-left-radius: var(--b-border-radius-radius-xs);
  border-bottom-right-radius: var(--b-border-radius-radius-xs);
}
// Detach the inline editor from the next row, cancelling the -1px pull.
.saved-rules > .rule-edit-wrap {
  margin-bottom: var(--space-m, 16px);
}
.saved-rules > .rule-edit-wrap + * {
  margin-top: 0;
}

.rule-edit-wrap {
  border: 1px solid var(--border-subtle);
  background: var(--bg-primary-default);

  > .policy-rule-summary {
    border: none;
    border-bottom: 1px solid var(--border-subtle);
    border-radius: 0;
  }
}
</style>
