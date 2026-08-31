<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="gvk-policy-yaml-section"
    label="Rules"
    default-opened
    :body-no-padding="true"
    :section-invalid="!hasRules"
  >
    <section class="gvk-yaml-content">
      <div v-if="$slots['mode-toggle']" class="q-mb-md"><slot name="mode-toggle" /></div>
      <section v-if="hasRules" class="existing-policy-section">
        <div class="row items-center justify-between q-mb-md">
          <div>A policy YAML has already been configured for the selected scope and workload type</div>
          <runai-btn
            :disable="!canEditPolicies"
            color="primary"
            flat
            aid="edit-gvk-policy-btn"
            label="Edit Policy"
            @click="openEditor"
          />
        </div>
        <div class="policy-overview">
          <runai-editor :key="editorKey" read-only :code="policyYamlPreview" height="400px" />
        </div>
        <div v-if="syncBadgeLabel" class="row q-mt-md">
          <runai-badge-mono
            class="q-ml-auto"
            show-badge-group
            :outlined-label="syncBadgeLabel"
            filled-label=""
            status="success"
            aid="gvk-yaml-sync-badge"
          />
        </div>
      </section>

      <section v-else class="no-policy-section">
        <div class="q-mb-md">Add a raw YAML configuration file to define policy rules</div>
        <div>
          <runai-btn
            :disable="!canEditPolicies"
            color="primary"
            flat
            class="btn-icon-sm"
            icon="ph-bold ph-plus"
            aid="add-gvk-policy-btn"
            label="Policy YAML"
            @click="openEditor"
          />
        </div>
      </section>

      <div class="row">
        <q-field class="col-4 form-hint no-padding" :model-value="!hasRules" :rules="[isYamlConfigured]"></q-field>
      </div>
    </section>

    <runai-editor-modal
      aid="gvk-runai-editor-modal"
      v-if="isEditorOpen"
      header-label="YAML Editor"
      :invalid-message="errorMessages.ADD_VALID_POLICY_YAML"
      :initial-code="policyYamlPreview"
      :invalid="yamlInvalid"
      @submit="applyYaml"
      @changed="yamlInvalid = false"
      @cancel="closeEditor"
    >
      <template #subheader>
        <span>Type or paste a YAML policy with a list of rules</span>
      </template>
    </runai-editor-modal>
  </runai-expansion-item>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiEditor } from "@/components/common/runai-editor";
import { RunaiEditorModal } from "@/components/common/runai-editor-modal";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiBadgeMono } from "@/components/common/runai-badge-mono";
import { yamlUtil } from "@/utils/yaml.util";
import { errorMessages } from "@/common/error-message.constant";
import type { PolicyRule } from "@/models/gvk-policy.model";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { usePermissionStore } from "@/stores/permissions.store";

const EMPTY_RULES_SKELETON = `- field:
  specSelector:
  enforcements:
    - `;

export interface Props {
  modelValue: PolicyRule[];
  syncBadgeLabel?: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (event: "update:modelValue", rules: PolicyRule[]): void;
}>();

const permissionStore = usePermissionStore();
// v-permission can't be used here: runai-btn's root is a non-element node, so the directive never reaches the button.
// Backend-verified: create needs Policies/CREATE, update needs Policies/UPDATE — gate on either.
const canEditPolicies = computed<boolean>(
  () =>
    permissionStore.hasPermission(ResourceType.Policies, Action.Create) ||
    permissionStore.hasPermission(ResourceType.Policies, Action.Update),
);

const isEditorOpen = ref<boolean>(false);
const yamlInvalid = ref<boolean>(false);
const editorKey = ref<number>(201);

const hasRules = computed<boolean>(() => props.modelValue.length > 0);
const policyYamlPreview = computed<string>(() => {
  if (!hasRules.value) return EMPTY_RULES_SKELETON;
  return yamlUtil.objectToYaml(props.modelValue)?.trimStart() || "";
});

function isYamlConfigured(val: boolean): boolean | string {
  return !val ? true : errorMessages.ADD_VALID_POLICY_YAML;
}
function openEditor(): void {
  yamlInvalid.value = false;
  isEditorOpen.value = true;
}
function closeEditor(): void {
  isEditorOpen.value = false;
  yamlInvalid.value = false;
}
function applyYaml(yaml: string): void {
  const rules = parseRules(yaml);
  if (rules === null) {
    yamlInvalid.value = true;
    return;
  }
  emit("update:modelValue", rules);
  editorKey.value++;
  closeEditor();
}
function parseRules(yaml: string): PolicyRule[] | null {
  try {
    const parsed = yamlUtil.yamlToObject(yaml) as unknown;
    const rules = Array.isArray(parsed) ? parsed : (parsed as { rules?: unknown })?.rules;
    if (!Array.isArray(rules)) return null;
    const valid = rules.every(
      (rule) =>
        !!rule &&
        typeof rule === "object" &&
        typeof (rule as PolicyRule).field === "string" &&
        typeof (rule as PolicyRule).specSelector === "string" &&
        Array.isArray((rule as PolicyRule).enforcements),
    );
    return valid ? (rules as PolicyRule[]) : null;
  } catch {
    return null;
  }
}
</script>

<style lang="scss" scoped>
.gvk-yaml-content {
  padding: var(--space-l, 20px) var(--space-m, 16px);
}
</style>
