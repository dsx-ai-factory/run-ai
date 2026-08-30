<template>
  <runai-base-dialog
    class="access-rule-management-modal"
    model-value
    @close="handleClose"
    size="width-lg"
    no-padding
    :loading="initialLoader"
  >
    <template #header>{{ header }} </template>
    <template #body>
      <access-rule-management
        :form-options="modalOptions"
        @access-rules-created="onAccessRulesCreated"
        @access-rule-deleted="onAccessRuleDeleted"
      />
    </template>

    <template #footer-left>
      <q-field class="col-4 form-hint-translate-y-100" :model-value="displayFormHint" :rules="[isFormIncomplete]" />
    </template>
    <template #footer>
      <q-btn label="Close" color="primary" @click="handleClose" />
    </template>
  </runai-base-dialog>
</template>
<script setup lang="ts">
import { ref, computed } from "vue";
// Components
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { AccessRuleManagement } from "@/components/rbac/access-rule/access-rule-management";

// Models
import type { IAccessRuleManagementModalOptions } from "@/models/access-rule.model";
import type { AccessRulesBatchCreateResponse } from "@/swagger-models/authorization-client";
import { errorMessages } from "@/common/error-message.constant";

// Utils
import { accessRuleUtil } from "@/utils/rbac.util/access-rule.util/access-rule.util";

// Props
const props = defineProps<{
  modalOptions: IAccessRuleManagementModalOptions;
}>();

// Emits
const emit = defineEmits<{
  close: [];
  "access-rules-created": [createdAccessRules: AccessRulesBatchCreateResponse];
  "access-rule-deleted": [index: number];
}>();

// Refs
const initialLoader = ref(false);
const displayFormHint = ref(false);

// Computed
const header = computed(() => accessRuleUtil.getAccessRuleModalHeader(props.modalOptions));

// Methods
function handleClose(): void {
  emit("close");
}

function isFormIncomplete(val: boolean): boolean | string {
  return !val ? true : errorMessages.FORM_INCOMPLETE;
}

function onAccessRulesCreated(createdAccessRules: AccessRulesBatchCreateResponse): void {
  emit("access-rules-created", createdAccessRules);
}

function onAccessRuleDeleted(index: number): void {
  emit("access-rule-deleted", index);
}
</script>

<style scoped lang="scss"></style>
