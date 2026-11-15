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
          <department-name-section
            compact
            auto-focus
            v-model:department-name="name"
            v-model:description="description"
            :cluster-id="clusterId"
            :read-only="false"
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
              icon="ph ph-x"
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
import { computed, onMounted, ref } from "vue";
import { useAccessRuleRoles } from "@/composables/access-rule/use-access-rule-roles.composable";
import { useAuthStore } from "@/stores/auth.store";
import { errorMessages } from "@/common/error-message.constant";
import { QForm, useQuasar } from "quasar";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { usePermissionStore } from "@/stores/permissions.store";
import type { SubjectType } from "@/swagger-models/authorization-client";
import { Action, ResourceType, ScopeType } from "@/swagger-models/authorization-client";
import type { Department } from "@/swagger-models/org-unit-service-client";
import { extractErrorMessages } from "@/utils/error-alert.util";
import { alertUtil } from "@/utils/alert.util";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { RunaiSectionTitle } from "@/components/common/runai-section-title";
import { RunaiSelect } from "@/components/common/runai-select";
import DepartmentNameSection from "@/components/department/department-edit-form/department-name-section/department-name-section.vue";
import { SubjectAddableMultiSelect } from "@/components/rbac/common/subject-addable-multi-select";
import { subjectTypeOptions, subjectTypeOptionsWithSSO } from "@/models/access-rule.model";
import { SubjectTypeExtended } from "@/swagger-models/authorization-client";
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
const $q = useQuasar();
const authStore = useAuthStore();
const subjectOptions = computed(() => (authStore.isSSO ? subjectTypeOptionsWithSSO : subjectTypeOptions));
const subject = ref<ISelectOption | null>(null);
const subjects = ref<ISelectOption[]>([]);
const { roleOptions, canAssignRoles, loadRoles, validateRole, loading: rolesLoading } = useAccessRuleRoles();
const role = ref<ISelectOption | null>(null);
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
    $q.notify(alertUtil.getSuccess(`Department ${createdDepartment.value.name} created`));
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
.scheduling-new-department {
  width: 800px !important;
  max-width: calc(100vw - 40px) !important;
}
.scheduling-new-department .runai-base-dialog-header {
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
  min-height: 32px;
}
.scheduling-new-department .runai-base-dialog-body {
  padding: 26px 16px;
}
.scheduling-new-department .runai-base-dialog-footer {
  min-height: 70px;
  padding: 16px;
}
.new-department-form .department-cluster {
  width: 85%;
  margin-top: 31px;
  margin-bottom: 26px;
  padding-bottom: 24px;
}
.new-department-form .department-cluster .q-field__control:before {
  border-bottom-style: solid;
}
.new-department-form .department-cluster .q-field__label,
.new-department-form .department-cluster .q-field__native {
  color: var(--text-disabled, rgba(0, 0, 0, 0.4));
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
.new-department-form .runai-expansion-item {
  border: 0;
  box-shadow: none;
  margin: 0 0 26px;
}
.new-department-form .runai-section-title {
  font-weight: 400;
}
.administrator-rule {
  position: relative;
  border: 1px solid var(--border-subtle);
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
  top: 20px;
  right: 10px;
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
</style>

<style scoped>
.remove-administrator :deep(.q-icon) {
  font-size: 20px;
}
</style>
