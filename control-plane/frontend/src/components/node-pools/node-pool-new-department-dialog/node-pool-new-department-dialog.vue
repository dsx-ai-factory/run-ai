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
        <fieldset :disabled="saving || !!createdDepartment" class="department-fields">
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
          <runai-section-title
            flat
            class="administrator-title"
            title="Set the department administrator(s)"
            tooltip-text="Department administrators can manage the department and its projects. Adding them here enables an access rule scoped to the department. It can be edited later in the department."
          />
          <div v-if="administrator" class="administrator-rule">
            <runai-btn
              icon-button
              aid="remove-department-administrator"
              icon="ph-bold ph-x"
              size="sm"
              tooltip="Remove administrator"
              class="remove-administrator"
              aria-label="Remove administrator"
              @click="administrator = false"
            />
            <runai-select
              v-model="subject"
              :options="subjectOptions"
              :rules="[requiredSubject]"
              hide-bottom-space
              label="Subject"
              class="subject-select"
              @update:model-value="subjects = []"
            />
            <subject-addable-multi-select
              v-if="subject && subject.value !== SubjectTypeExtended.AllAuthenticatedUsers"
              :subject-type="subject"
              :selected-subjects-options="subjects"
              @update:selected-subjects="subjects = $event"
              @add-subject="subjects = [...subjects, $event]"
            />
            <runai-select
              v-model="role"
              :options="roleOptions"
              :rules="[validateRole]"
              :loading="rolesLoading"
              :disable="!canAssignRoles || rolesLoading || !!rolesError"
              label="Role"
              class="role-select"
            />
            <p v-if="rolesError" role="alert" class="text-negative">
              {{ rolesError }} <runai-btn aid="retry-department-roles" flat label="Retry" @click="loadRoleOptions" />
            </p>
          </div>
          <runai-btn
            v-else
            flat
            color="primary"
            icon="ph-bold ph-plus"
            label="Administrator"
            class="add-administrator"
            aid="add-department-administrator"
            :disable="!canAssignRoles"
            @click="administrator = true"
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
import { computed, onMounted, ref, watch } from "vue";
import { useAccessRuleRoles } from "@/composables/node-pool/use-access-rule-roles.composable";
import { useAuthStore } from "@/stores/auth.store/index";
import { errorMessages } from "@/common/error-message.constant";
import { QForm } from "quasar";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { usePermissionStore } from "@/stores/permissions.store/index";
import type { SubjectType } from "@/swagger-models/authorization-client/index";
import { Action, ResourceType, ScopeType } from "@/swagger-models/authorization-client/index";
import type { Department } from "@/swagger-models/org-unit-service-client/index";
import { extractErrorMessages } from "@/utils/error-alert.util/index";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog/index";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { RunaiSectionTitle } from "@/components/common/runai-section-title/index";
import { RunaiSelect } from "@/components/common/runai-select/index";
import { SchedulingDepartmentNameFields } from "@/components/node-pools/scheduling-department-name-fields";
import { SubjectAddableMultiSelect } from "@/components/rbac/common/subject-addable-multi-select/index";
import { subjectTypeOptions, subjectTypeOptionsWithSSO } from "@/models/access-rule.model/index";
import { SubjectTypeExtended } from "@/swagger-models/authorization-client/index";
import type { ISelectOption } from "@/models/global.model";
const props = defineProps<{ clusterId: string; clusterName?: string }>();
const emit = defineEmits<{ close: []; created: [department: Department]; saving: [value: boolean] }>();
const name = ref(""),
  description = ref(""),
  administrator = ref(false);
const saving = ref(false);
const saveError = ref("");
const createdDepartment = ref<Department>();
const completedSubjects = new Set<string>();
const permissionStore = usePermissionStore();
const canCreateDepartment = computed(() => permissionStore.hasPermission(ResourceType.Department, Action.Create));
const authStore = useAuthStore();
const subjectOptions = computed(() => (authStore.isSSO ? subjectTypeOptionsWithSSO : subjectTypeOptions));
const subject = ref<ISelectOption | null>(null);
const subjects = ref<ISelectOption[]>([]);
const { roleOptions, canAssignRoles, loadRoles, validateRole, loading: rolesLoading } = useAccessRuleRoles();
const role = ref<ISelectOption | null>(null);
watch(
  roleOptions,
  (options) => {
    if (!role.value) role.value = options.find((option) => option.label === "Department administrator") ?? null;
  },
  { immediate: true },
);
const rolesError = ref("");
const requiredSubject = (value: ISelectOption | null) => !!value || errorMessages.SUBJECT_NOT_EMPTY;
async function loadRoleOptions() {
  rolesError.value = "";
  try {
    await loadRoles();
  } catch {
    rolesError.value = "Could not load roles.";
  }
}
onMounted(loadRoleOptions);
const form = ref<QForm>();
async function createDepartment() {
  if (saving.value || !canCreateDepartment.value) return;
  saving.value = true;
  emit("saving", true);
  saveError.value = "";
  try {
    if (!createdDepartment.value && !(await form.value?.validate())) return;
    if (
      administrator.value &&
      (!canAssignRoles.value || rolesLoading.value || rolesError.value || validateRole(role.value) !== true)
    )
      return;
    if (
      administrator.value &&
      subject.value?.value !== SubjectTypeExtended.AllAuthenticatedUsers &&
      !subjects.value.length
    ) {
      saveError.value = "Select at least one subject.";
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
    if (administrator.value && role.value && subject.value) {
      const scope = {
        scopeId: createdDepartment.value.id,
        scopeType: ScopeType.Department,
        clusterId: props.clusterId,
        roleId: Number(role.value.value),
      };
      if (subject.value.value === SubjectTypeExtended.AllAuthenticatedUsers) {
        await accessRuleService.createAuthenticatedUsersAccessRule(scope);
      } else {
        for (const selected of subjects.value) {
          const subjectId = String(selected.value);
          if (completedSubjects.has(subjectId)) continue;
          await accessRuleService.createAccessRule({
            ...scope,
            subjectId,
            subjectType: subject.value.value as SubjectType,
          });
          completedSubjects.add(subjectId);
        }
      }
    }
    emit("close");
  } catch (error) {
    const message = extractErrorMessages(error, "Could not complete department creation.").join(" ");
    saveError.value = createdDepartment.value
      ? `Department created, but role assignments are incomplete. Retry to finish. ${message}`
      : message;
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
.scheduling-new-department .scheduling-department-name-fields .q-textarea .q-field__control-container {
  padding: 0;
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
.administrator-rule {
  position: relative;
  box-shadow: inset 0 0 0 1px var(--border-subtle);
  padding: 16px;
  margin-top: 26px;
}
.administrator-rule .subject-select,
.administrator-rule .role-select {
  width: 380px;
  max-width: 100%;
}
.administrator-rule .subject-addable-multi-select {
  margin: 26px 0;
  width: 90%;
}
.administrator-rule .remove-administrator {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 32px;
  height: 32px;
  min-width: 32px;
  min-height: 32px;
  padding: 0;
}
.add-administrator {
  margin-top: 26px;
  height: 38px;
  min-height: 38px;
  padding: 0 10px;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  letter-spacing: 0;
}
.add-administrator .q-icon {
  width: 20px;
  height: 20px;
  font-size: 20px;
  margin-right: 10px;
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
