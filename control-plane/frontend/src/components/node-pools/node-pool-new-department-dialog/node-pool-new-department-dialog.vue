<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-base-dialog
    :model-value="true"
    custom-class="scheduling-new-department"
    size="width-lg"
    :is-modal-persistent="saving"
    :hide-x="saving"
    @close="!saving && emit('close')"
  >
    <template #header>New Department</template>
    <template #body>
      <q-form ref="form" class="new-department-form" @submit.prevent="createDepartment">
        <fieldset :disabled="fieldsDisabled" class="department-fields">
          <p>The department will be created under the following cluster and can use all of its node pools</p>
          <q-input
            :model-value="clusterName || clusterId"
            label="Cluster"
            readonly
            stack-label
            class="department-cluster"
          />
          <scheduling-department-name-fields
            v-model:department-name="name"
            v-model:description="description"
            :cluster-id="clusterId"
          />
          <org-unit-administrators-fields
            ref="administratorsRef"
            title="Set the department administrator(s)"
            :tooltip="EOrgUnitFormTooltip.DepartmentAdministrators"
            default-role="Department administrator"
            :disable="fieldsDisabled"
          />
        </fieldset>
      </q-form>
    </template>
    <template #footer-left>
      <p v-if="saveError" class="text-negative department-save-error" role="alert">{{ saveError }}</p>
    </template>
    <template #footer
      ><runai-btn
        aid="cancel-new-department"
        flat
        color="primary"
        label="Cancel"
        :disable="saving"
        @click="emit('close')" /><runai-btn
        color="primary"
        :label="createdDepartment ? 'Retry role assignments' : 'Create department'"
        aid="create-department"
        :loading="saving"
        :disable="!canCreateDepartment"
        @click="createDepartment"
    /></template>
  </runai-base-dialog>
</template>
<script setup lang="ts">
import { computed, ref } from "vue";
import { QForm, useQuasar } from "quasar";
import { alertUtil } from "@/utils/alert.util";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { usePermissionStore } from "@/stores/permissions.store/index";
import { Action, ResourceType, ScopeType } from "@/swagger-models/authorization-client/index";
import type { Department } from "@/swagger-models/org-unit-service-client/index";
import { extractErrorMessages } from "@/utils/error-alert.util/index";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog/index";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { SchedulingDepartmentNameFields } from "@/components/node-pools/scheduling-department-name-fields";
import { OrgUnitAdministratorsFields } from "@/components/org-unit/v2/common/org-unit-administrators-fields";
import { EOrgUnitFormTooltip } from "@/models/org-unit-forms.model";
const props = defineProps<{ clusterId: string; clusterName?: string }>();
const emit = defineEmits<{ close: []; created: [department: Department]; saving: [value: boolean] }>();
const $q = useQuasar();
const name = ref(""),
  description = ref("");
const saving = ref(false);
const saveError = ref("");
const createdDepartment = ref<Department>();
const fieldsDisabled = computed(() => saving.value || !!createdDepartment.value);
const permissionStore = usePermissionStore();
const canCreateDepartment = computed(() => permissionStore.hasPermission(ResourceType.Department, Action.Create));
const administratorsRef = ref<InstanceType<typeof OrgUnitAdministratorsFields>>();
const form = ref<QForm>();
async function createDepartment() {
  if (saving.value || !canCreateDepartment.value) return;
  saving.value = true;
  emit("saving", true);
  saveError.value = "";
  try {
    if (!createdDepartment.value && !(await form.value?.validate())) return;
    const administrators = administratorsRef.value?.validate();
    if (administrators && !administrators.valid) {
      if (administrators.message) saveError.value = administrators.message;
      return;
    }
    if (!createdDepartment.value) {
      createdDepartment.value = await orgUnitService.createDepartment({
        name: name.value.trim(),
        description: description.value,
        clusterId: props.clusterId,
        resources: [],
      });
      emit("created", createdDepartment.value);
    }
    await administratorsRef.value?.save({
      scopeId: createdDepartment.value.id,
      scopeType: ScopeType.Department,
      clusterId: props.clusterId,
    });
    emit("close");
  } catch (error) {
    if (createdDepartment.value) {
      const message = extractErrorMessages(error, "Could not complete department creation.").join(" ");
      saveError.value = `Department created, but role assignments are incomplete. Retry to finish. ${message}`;
    } else {
      $q.notify(alertUtil.getError(`Department ${name.value.trim()} couldn't be created`));
    }
  } finally {
    saving.value = false;
    emit("saving", false);
  }
}
</script>
<style lang="scss">
.q-card.scheduling-new-department {
  display: flex;
  flex-direction: column;
  width: 800px !important;
  max-width: calc(100vw - 40px) !important;
}
.scheduling-new-department > .q-separator {
  display: none;
}
.q-card.scheduling-new-department > .runai-base-dialog-header {
  box-shadow: inset 0 -1px var(--border-subtle);
  height: 64px;
  min-height: 64px;
  padding: 16px;
}
.scheduling-new-department .runai-base-dialog-header b {
  font-weight: 500;
  line-height: 28px;
}
.scheduling-new-department .runai-base-dialog-close-btn {
  height: 32px;
  width: 32px;
  min-width: 32px;
  min-height: 32px;
  padding: 0;
}
.scheduling-new-department .runai-base-dialog-close-btn .q-icon {
  width: 20px;
  height: 20px;
  font-size: 20px;
}
.scheduling-new-department .runai-base-dialog-body {
  min-height: 0;
  padding: 26px 16px;
}
.q-card.scheduling-new-department > .runai-base-dialog-footer {
  box-shadow: inset 0 1px var(--border-subtle);
  min-height: 70px;
  padding: 16px;
}
.scheduling-new-department .department-fields > p {
  display: flex;
  align-items: center;
  min-height: 32px;
  margin: 0;
}
.new-department-form .department-cluster {
  width: 655px;
  max-width: 100%;
  margin-top: 26px;
  margin-bottom: 26px;
  padding-bottom: 24px;
}
.scheduling-new-department .department-cluster .q-field__control {
  height: auto;
}
.scheduling-new-department .department-cluster .q-field__control-container {
  display: flex;
  flex-direction: column;
  padding: 0;
}
.scheduling-new-department .department-cluster .q-field__label {
  position: static;
  order: 0;
  height: auto;
  min-height: 17px;
  transform: none;
  font-size: 12px;
  line-height: 16.8px;
}
.scheduling-new-department .department-cluster .q-field__native {
  order: 1;
  min-height: 31px;
  margin-top: 3px;
  padding: 0 0 10px;
  line-height: 21px;
}
.scheduling-new-department .scheduling-department-name-fields .q-field__native {
  line-height: 21px;
}
.scheduling-new-department .scheduling-department-name-fields textarea.q-field__native {
  overflow-y: auto !important;
}
.scheduling-new-department .q-field__messages,
.scheduling-new-department .q-field__counter {
  min-height: 17px;
  font-size: 12px;
  line-height: 16.8px;
  letter-spacing: 0;
}
.new-department-form .department-cluster .q-field__control:before {
  border-bottom-style: solid;
}
.new-department-form .department-cluster .q-field__label,
.new-department-form .department-cluster .q-field__native {
  color: var(--text-disabled);
}
.new-department-form .administrator-title {
  height: 20px;
  line-height: 20px;
}
.new-department-form .administrator-title > .row {
  height: 20px;
  flex-wrap: nowrap;
  align-items: center;
}
.new-department-form .administrator-title .q-btn {
  height: 32px;
  width: 32px;
  min-height: 32px;
  padding: 0;
}
.new-department-form .administrator-title .q-icon {
  font-size: 20px;
}
.new-department-form .runai-section-title {
  font-weight: 400;
}
.department-fields {
  border: 0;
  padding: 0;
  margin: 0;
  min-width: 0;
}
.department-save-error {
  color: var(--b-text-sub);
  font-size: 12px;
  margin: 0;
}
.scheduling-new-department .remove-administrator .q-icon {
  width: 20px;
  height: 20px;
  font-size: 20px;
}

.q-card.scheduling-new-department {
  --q-negative: var(--text-error);
  overflow: hidden;
  border-radius: 5px;
  box-shadow: var(--shadow-s);
  letter-spacing: 0;
}
.scheduling-new-department .runai-base-dialog-header h6 {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 20px;
  font-weight: 500;
  line-height: 28px;
  letter-spacing: 0;
}
.scheduling-new-department .q-field__native {
  font-size: 14px;
  line-height: 21px;
  letter-spacing: 0;
}
.scheduling-new-department .q-field__label {
  letter-spacing: 0;
}
.q-card.scheduling-new-department .new-department-form .department-cluster .q-field__label,
.q-card.scheduling-new-department .new-department-form .department-cluster .q-field__native {
  color: var(--text-disabled);
}
.scheduling-new-department .q-field:not(.q-field--error) .q-field__native::placeholder {
  color: rgba(0, 0, 0, 0.6);
}
.scheduling-new-department .q-field:not(.q-field--error) .q-field__control:before {
  border-color: #bdbdbd;
}
.scheduling-new-department .administrator-title > .row > span {
  min-width: 224px;
  margin-right: 10px;
  line-height: 21px;
}
.scheduling-new-department .administrator-rule .subject-select,
.scheduling-new-department .administrator-rule .role-select {
  width: 250px;
}
.scheduling-new-department .subject-select + .role-select {
  margin-block-start: 40px;
}
.scheduling-new-department .administrator-rule .subject-addable-multi-select {
  width: 100%;
}
.scheduling-new-department .runai-base-dialog-footer .footer {
  gap: 16px;
}
.scheduling-new-department .runai-base-dialog-footer .footer .q-btn,
.scheduling-new-department .add-administrator {
  height: 38px;
  min-height: 38px;
  padding: 10px;
  border-radius: 3px;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  letter-spacing: 0;
}
.scheduling-new-department .runai-base-dialog-footer .footer .q-btn + .q-btn {
  margin-left: 0;
  padding: 10px;
}
.q-card.scheduling-new-department .runai-base-dialog-footer .footer [aid="cancel-new-department"] {
  min-width: 74px;
}
.q-card.scheduling-new-department .runai-base-dialog-footer .footer [aid="create-department"] {
  min-width: 164px;
  box-shadow: var(--shadow-s);

  &::before {
    box-shadow: none;
  }
}
.scheduling-new-department .add-administrator {
  min-width: 160px;
}

.scheduling-new-department .administrator-rule .q-field:not(.q-field--float) .q-field__label {
  height: auto;
  min-height: 21px;
  font-size: 14px;
  line-height: 21px;
}
.scheduling-new-department .administrator-rule .q-field--float .q-field__label {
  top: 0;
  height: auto;
  min-height: 17px;
  transform: none;
  font-size: 12px;
  line-height: 16.8px;
}
</style>
