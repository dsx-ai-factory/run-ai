<template>
  <section class="access-rules-step">
    <runai-expansion-item label="Access Rules" default-opened prevent-top-padding-when-open disable-closing>
      <access-rule-management
        ref="accessRuleManagementRef"
        :form-options="props.formOptions"
        no-inner-padding
        @access-rules-created="onAccessRulesCreated"
        @access-rule-deleted="onAccessRuleDeleted"
      />
    </runai-expansion-item>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeUnmount } from "vue";
import { AccessRuleManagement } from "@/components/rbac/access-rule/access-rule-management";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { useUnsavedChangesGuard } from "@/composables/org-unit/use-unsaved-changes-guard.composable";
import type { IAccessRuleManagementFormOptions } from "@/models/access-rule.model";
import type { AccessRulesBatchCreateResponse } from "@/swagger-models/authorization-client";
import { EStepNames } from "@/models/org-unit-forms.model";

const STEP_NAME = EStepNames.AccessRules;

const props = defineProps<{
  formOptions: IAccessRuleManagementFormOptions;
}>();

const emit = defineEmits<{
  "access-rules-created": [createdAccessRules: AccessRulesBatchCreateResponse];
  "access-rule-deleted": [];
}>();

const unsavedGuard = useUnsavedChangesGuard();
const accessRuleManagementRef = ref<InstanceType<typeof AccessRuleManagement> | null>(null);

const isDirty = computed(() => {
  const management = accessRuleManagementRef.value;
  if (!management) return false;
  return management.isFormOpen && management.isFormDirty;
});

const handleSave = async () => {
  await accessRuleManagementRef.value?.submitForm();
};

const handleDiscard = () => {
  accessRuleManagementRef.value?.closeForm();
};

const unregisterGuard = unsavedGuard.registerStepGuard(STEP_NAME, {
  isDirty: () => isDirty.value,
  save: handleSave,
  discard: handleDiscard,
});

onBeforeUnmount(() => {
  unregisterGuard();
});

function onAccessRulesCreated(createdAccessRules: AccessRulesBatchCreateResponse): void {
  emit("access-rules-created", createdAccessRules);
}

function onAccessRuleDeleted(): void {
  emit("access-rule-deleted");
}

defineExpose({
  isDirty,
  save: handleSave,
  discardChanges: handleDiscard,
});
</script>
